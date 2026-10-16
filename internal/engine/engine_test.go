package engine_test

import (
	"context"
	"errors"
	"testing"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/testkit"
)

func executeEngine(t *testing.T, files map[string]testkit.MemFile) *engine.Engine {
	t.Helper()
	return executeEngineWithPolicy(t, files, []byte("floor: strict\n"), testkit.NewSequenceIDSource("work"))
}

func executeEngineWithPolicy(
	t *testing.T,
	files map[string]testkit.MemFile,
	floor []byte,
	ids *testkit.SequenceIDSource,
) *engine.Engine {
	t.Helper()
	policy, err := engine.NewPolicySnapshot(engine.PolicySnapshotInput{Source: "test-policy", Floor: floor})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := engine.NewEngine(engine.Dependencies{
		Facts:   testkit.NewMemFileFacts(files),
		IDs:     ids,
		Policy:  testkit.NewFixedPolicySnapshotter(policy),
		Workdir: "/repo",
		Build:   engine.BuildMetadata{Version: "0.1.0-dev", Commit: "abc", BuildTime: "unknown", GoVersion: "go1.24.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return eng
}

func validLayoutFiles() map[string]testkit.MemFile {
	return map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml":       {Bytes: []byte(discoveryWorkspace)},
		"/repo/.sdd/work/apple/work.yaml": {Bytes: []byte(discoveryWorkApple)},
	}
}

func outcomeDiagnostic(t *testing.T, outcome engine.Outcome) engine.Diagnostic {
	t.Helper()
	diagnostics := outcome.Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(diagnostics))
	}
	return diagnostics[0]
}

func TestExecuteVersionSucceedsOutsideWorkspace(t *testing.T) {
	t.Parallel()

	eng := executeEngine(t, nil)
	outcome := eng.Execute(context.Background(), engine.NewVersionCommand())
	if !outcome.OK() || outcome.ExitCode() != engine.ExitSuccess {
		t.Fatalf("version outcome: %#v", outcome)
	}
	result := outcome.Result()
	if result == nil || result.Kind() != engine.ResultVersion {
		t.Fatalf("version result kind: %#v", result)
	}
	version, ok := result.Version()
	if !ok || version.Version() != "0.1.0-dev" || version.Commit() != "abc" || version.ProtocolSchemaVersion() != engine.SchemaVersion {
		t.Fatalf("version fields: %#v", version)
	}
}

func TestExecuteMapsEveryFailureCategory(t *testing.T) {
	t.Parallel()

	ambiguous := map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml":       {Bytes: []byte(discoveryWorkspace)},
		"/repo/.sdd/work/apple/work.yaml": {Bytes: []byte(discoveryWorkApple)},
		"/repo/.sdd/work/zebra/work.yaml": {Bytes: []byte(discoveryWorkZebra)},
	}
	incompatible := map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml": {Bytes: []byte("formatVersion: 2\nworkspaceId: ws-1\nrevision: 1\n")},
	}

	tests := []struct {
		name         string
		files        map[string]testkit.MemFile
		command      engine.Command
		cancel       bool
		policyFault  error
		idFault      error
		wantCode     string
		wantExit     engine.ExitCode
		wantSeverity engine.Severity
	}{
		{name: "malformed command", files: validLayoutFiles(), command: engine.Command{}, wantCode: "DIAG_USAGE_ERROR", wantExit: engine.ExitUsage, wantSeverity: engine.SeverityError},
		{name: "cancelled version", files: validLayoutFiles(), command: engine.NewVersionCommand(), cancel: true, wantCode: "DIAG_INTERRUPTED", wantExit: engine.ExitInterrupted, wantSeverity: engine.SeverityError},
		{name: "cancelled status", files: validLayoutFiles(), command: engine.NewStatusCommand("", ""), cancel: true, wantCode: "DIAG_INTERRUPTED", wantExit: engine.ExitInterrupted, wantSeverity: engine.SeverityError},
		{name: "missing workspace", files: nil, command: engine.NewStatusCommand("", ""), wantCode: "DIAG_WORKSPACE_NOT_FOUND", wantExit: engine.ExitWorkspace, wantSeverity: engine.SeverityError},
		{name: "incompatible workspace", files: incompatible, command: engine.NewStatusCommand("", ""), wantCode: "DIAG_WORKSPACE_VERSION_INCOMPATIBLE", wantExit: engine.ExitWorkspace, wantSeverity: engine.SeverityError},
		{name: "ambiguous selection", files: ambiguous, command: engine.NewStatusCommand("", ""), wantCode: "DIAG_WORKSPACE_SELECTION_AMBIGUOUS", wantExit: engine.ExitWorkspace, wantSeverity: engine.SeverityError},
		{name: "missing work item", files: validLayoutFiles(), command: engine.NewStatusCommand("work-missing", ""), wantCode: "DIAG_WORK_ITEM_NOT_FOUND", wantExit: engine.ExitWorkspace, wantSeverity: engine.SeverityError},
		{name: "policy unavailable", files: validLayoutFiles(), command: engine.NewDoctorCommand(""), policyFault: errors.New("vault sealed"), wantCode: "WORKFLOW_POLICY_UNAVAILABLE", wantExit: engine.ExitPolicy, wantSeverity: engine.SeverityCritical},
		{name: "identity fault", files: validLayoutFiles(), command: mustStart(t), idFault: errors.New("namespace exhausted"), wantCode: "DIAG_WORKSPACE_VERSION_INCOMPATIBLE", wantExit: engine.ExitWorkspace, wantSeverity: engine.SeverityError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ids := testkit.NewSequenceIDSource("work")
			if test.idFault != nil {
				ids.InjectFault(test.idFault)
			}
			policy, err := engine.NewPolicySnapshot(engine.PolicySnapshotInput{Source: "test-policy", Floor: []byte("floor: strict\n")})
			if err != nil {
				t.Fatal(err)
			}
			snapshotter := testkit.NewFixedPolicySnapshotter(policy)
			if test.policyFault != nil {
				snapshotter.InjectFault(test.policyFault)
			}
			eng, err := engine.NewEngine(engine.Dependencies{
				Facts:   testkit.NewMemFileFacts(test.files),
				IDs:     ids,
				Policy:  snapshotter,
				Workdir: "/repo",
				Build:   engine.BuildMetadata{Version: "0.1.0-dev"},
			})
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if test.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			outcome := eng.Execute(ctx, test.command)
			if outcome.OK() {
				t.Fatalf("failure case succeeded: %#v", outcome)
			}
			if outcome.ExitCode() != test.wantExit {
				t.Fatalf("exit code = %d want %d", outcome.ExitCode(), test.wantExit)
			}
			diagnostic := outcomeDiagnostic(t, outcome)
			if diagnostic.Code() != test.wantCode || diagnostic.Severity() != test.wantSeverity {
				t.Fatalf("diagnostic = %s/%s want %s/%s", diagnostic.Code(), diagnostic.Severity(), test.wantCode, test.wantSeverity)
			}
			if diagnostic.Message() == "" || diagnostic.Remediation() == "" {
				t.Fatal("failure diagnostic lacks message or remediation")
			}
		})
	}
}

func TestExecuteUnknownCommandPreservesRequest(t *testing.T) {
	t.Parallel()

	eng := executeEngine(t, nil)
	outcome := eng.UnknownCommand("bogus")
	if outcome.Command() != engine.CommandUnknown || outcome.RequestedCommand() != "bogus" {
		t.Fatalf("unknown outcome: %#v", outcome)
	}
	if outcome.OK() || outcome.ExitCode() != engine.ExitUsage {
		t.Fatalf("unknown outcome exit: %#v", outcome)
	}
	if code := outcomeDiagnostic(t, outcome).Code(); code != "DIAG_COMMAND_UNKNOWN" {
		t.Fatalf("unknown code = %q", code)
	}
	if outcome.Result() != nil {
		t.Fatal("unknown outcome must carry a null result")
	}

	empty := eng.UnknownCommand("  ")
	if empty.RequestedCommand() == "" {
		t.Fatal("empty requested command was not coerced")
	}
}

func TestExecuteStatusReportsSelection(t *testing.T) {
	t.Parallel()

	eng := executeEngine(t, validLayoutFiles())
	outcome := eng.Execute(context.Background(), engine.NewStatusCommand("", ""))
	if !outcome.OK() {
		t.Fatalf("status outcome: %#v", outcomeDiagnostic(t, outcome))
	}
	status, ok := outcome.Result().Status()
	if !ok || len(status.WorkItems()) != 1 || status.WorkItems()[0].ID() != "work-b" {
		t.Fatalf("implicit selection: %#v", status)
	}
	if outcome.Workspace() == nil || outcome.Workspace().Root() != "/repo" {
		t.Fatalf("status workspace: %#v", outcome.Workspace())
	}

	explicit := eng.Execute(context.Background(), engine.NewStatusCommand("work-b", ""))
	selected, _ := explicit.Result().Status()
	if len(selected.WorkItems()) != 1 {
		t.Fatalf("explicit selection: %#v", selected)
	}

	empty := executeEngine(t, map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml": {Bytes: []byte(discoveryWorkspace)},
	})
	zero := empty.Execute(context.Background(), engine.NewStatusCommand("", ""))
	if !zero.OK() {
		t.Fatalf("zero-work status: %#v", outcomeDiagnostic(t, zero))
	}
	items, _ := zero.Result().Status()
	if len(items.WorkItems()) != 0 {
		t.Fatalf("zero-work status items: %#v", items)
	}
}

func TestExecuteDoctorReportsChecksAndPartialFailures(t *testing.T) {
	t.Parallel()

	eng := executeEngine(t, validLayoutFiles())
	outcome := eng.Execute(context.Background(), engine.NewDoctorCommand(""))
	if !outcome.OK() {
		t.Fatalf("doctor outcome: %#v", outcomeDiagnostic(t, outcome))
	}
	doctor, ok := outcome.Result().Doctor()
	if !ok || len(doctor.Checks()) != 5 {
		t.Fatalf("doctor checks: %#v", doctor)
	}
	for _, check := range doctor.Checks() {
		if check["ok"] != true || check["name"] == nil {
			t.Fatalf("passing check is malformed: %v", check)
		}
	}

	broken := executeEngine(t, nil)
	failed := broken.Execute(context.Background(), engine.NewDoctorCommand(""))
	if failed.OK() || failed.ExitCode() != engine.ExitWorkspace {
		t.Fatalf("doctor without workspace: %#v", failed)
	}
	partial, ok := failed.Result().Doctor()
	if !ok || len(partial.Checks()) == 0 {
		t.Fatalf("failed doctor lost its completed checks: %#v", failed.Result())
	}
	if code := outcomeDiagnostic(t, failed).Code(); code != "DIAG_WORKSPACE_NOT_FOUND" {
		t.Fatalf("doctor failure code = %q", code)
	}
}

func TestExecuteStartDryRunPreviewsWithoutApplying(t *testing.T) {
	t.Parallel()

	facts := testkit.NewMemFileFacts(validLayoutFiles())
	policy, err := engine.NewPolicySnapshot(engine.PolicySnapshotInput{Source: "test-policy", Floor: []byte("floor: strict\n")})
	if err != nil {
		t.Fatal(err)
	}
	engineWithFacts, err := engine.NewEngine(engine.Dependencies{
		Facts:   facts,
		IDs:     testkit.NewSequenceIDSource("work"),
		Policy:  testkit.NewFixedPolicySnapshotter(policy),
		Workdir: "/repo",
		Build:   engine.BuildMetadata{Version: "0.1.0-dev"},
	})
	if err != nil {
		t.Fatal(err)
	}
	outcome := engineWithFacts.Execute(context.Background(), mustStart(t))
	if !outcome.OK() {
		t.Fatalf("dry-run outcome: %#v", outcomeDiagnostic(t, outcome))
	}
	result := outcome.Result()
	if result == nil || result.Kind() != engine.ResultStartDryRun {
		t.Fatalf("dry-run result kind: %#v", result)
	}
	start, _ := result.StartDryRun()
	plan := start.Plan()
	if len(plan.Operations()) != 1 || len(plan.Preconditions()) != 1 {
		t.Fatal("dry-run plan is not meaningful")
	}
	if plan.Authorizable() {
		t.Fatal("foundation preview must not be authorizable")
	}
	if plan.PlanID() == plan.WorkID() {
		t.Fatal("plan and work identities must differ")
	}
	if plan.ExpectedRevision() != 11 || plan.WorkspaceID() != "ws-1" {
		t.Fatalf("plan scope is wrong: %#v", plan)
	}
	if plan.PolicyDigest() != policy.Digest() {
		t.Fatal("plan does not bind the observed policy digest")
	}
	if verifyErr := engine.VerifyPlanDigest(plan); verifyErr != nil {
		t.Fatalf("returned plan failed verification: %v", verifyErr)
	}
	diagnostics := outcome.Diagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code() != "DIAG_PLAN_NOT_AUTHORIZABLE" {
		t.Fatalf("dry-run diagnostics: %v", diagnostics)
	}
	for name, file := range validLayoutFiles() {
		content, readErr := facts.ReadFile(context.Background(), name)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(content) != string(file.Bytes) {
			t.Fatalf("dry-run mutated %s", name)
		}
	}
	after, err := facts.ListDir(context.Background(), "/repo/.sdd/work")
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 1 || after[0].Name() != "apple" {
		t.Fatalf("dry-run applied effects: %v", after)
	}
}

func TestExecuteAmbiguousFailureCarriesCandidates(t *testing.T) {
	t.Parallel()

	eng := executeEngine(t, map[string]testkit.MemFile{
		"/repo/.sdd/workspace.yaml":       {Bytes: []byte(discoveryWorkspace)},
		"/repo/.sdd/work/apple/work.yaml": {Bytes: []byte(discoveryWorkApple)},
		"/repo/.sdd/work/zebra/work.yaml": {Bytes: []byte(discoveryWorkZebra)},
	})
	outcome := eng.Execute(context.Background(), engine.NewStatusCommand("", ""))
	diagnostic := outcomeDiagnostic(t, outcome)
	metadata := diagnostic.Metadata()
	candidates, ok := metadata["candidates"].([]any)
	if !ok || len(candidates) != 2 {
		t.Fatalf("ambiguous metadata: %v", metadata)
	}
	actions := outcome.NextActions()
	if len(actions) != 1 || actions[0].Command() != "status" {
		t.Fatalf("ambiguous next actions: %v", actions)
	}
	if actions[0].Flags()["work"].Value() != candidates[0] {
		t.Fatalf("next action does not target the first candidate: %v", actions[0].Flags())
	}
}

func TestNewEngineRejectsIncompleteDependencies(t *testing.T) {
	t.Parallel()

	policy, err := engine.NewPolicySnapshot(engine.PolicySnapshotInput{Source: "test-policy"})
	if err != nil {
		t.Fatal(err)
	}
	complete := engine.Dependencies{
		Facts:   testkit.NewMemFileFacts(nil),
		IDs:     testkit.NewSequenceIDSource("work"),
		Policy:  testkit.NewFixedPolicySnapshotter(policy),
		Workdir: "/repo",
		Build:   engine.BuildMetadata{Version: "0.1.0-dev"},
	}
	broken := []func(*engine.Dependencies){
		func(d *engine.Dependencies) { d.Facts = nil },
		func(d *engine.Dependencies) { d.IDs = nil },
		func(d *engine.Dependencies) { d.Policy = nil },
		func(d *engine.Dependencies) { d.Workdir = "" },
		func(d *engine.Dependencies) { d.Build.Version = "" },
	}
	for index, mutate := range broken {
		deps := complete
		mutate(&deps)
		if _, err := engine.NewEngine(deps); !errors.Is(err, engine.ErrInvalidProtocolValue) {
			t.Fatalf("incomplete dependencies %d error = %v", index, err)
		}
	}
}

func mustStart(t *testing.T) engine.Command {
	t.Helper()
	start, err := engine.NewStartDryRunCommand("add export", "quick-spec", "")
	if err != nil {
		t.Fatal(err)
	}
	return start
}
