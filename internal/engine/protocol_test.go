package engine_test

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"sdd.dev/sdd/internal/engine"
)

func TestCommandsUseClosedSchemaShapes(t *testing.T) {
	t.Parallel()

	start, err := engine.NewStartDryRunCommand("add export", "quick-spec", "/repo")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		command engine.Command
		want    string
	}{
		{name: "status", command: engine.NewStatusCommand("work-1", "/repo"), want: `{"type":"status","workId":"work-1","root":"/repo"}`},
		{name: "version", command: engine.NewVersionCommand(), want: `{"type":"version"}`},
		{name: "doctor", command: engine.NewDoctorCommand("/repo"), want: `{"type":"doctor","root":"/repo"}`},
		{name: "start", command: start, want: `{"type":"start","dryRun":true,"intent":"add export","workflow":"quick-spec","root":"/repo"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := json.Marshal(test.command)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, got, []byte(test.want))
		})
	}
}

func TestStartCommandRequiresIntent(t *testing.T) {
	t.Parallel()

	_, err := engine.NewStartDryRunCommand("  ", "", "")
	if !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("expected invalid protocol value, got %v", err)
	}
}

func TestOutcomeDefensivelyCopiesMutableInputs(t *testing.T) {
	t.Parallel()

	metadata := map[string]any{
		"nested": map[string]any{"value": "original"},
		"items":  []any{"first"},
	}
	diagnostic := mustDiagnostic(t, engine.DiagnosticInput{
		Code:     "DIAG_USAGE_ERROR",
		Severity: engine.SeverityError,
		Message:  "Invalid input",
		Source:   mustSource(t),
		Metadata: metadata,
	})
	args := []string{"next"}
	flags := map[string]engine.FlagValue{"work": engine.StringFlag("work-1")}
	action, err := engine.NewNextAction(engine.NextActionInput{
		Command: "next",
		Args:    args,
		Flags:   flags,
		Reason:  "Continue",
	})
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := []engine.Diagnostic{diagnostic}
	actions := []engine.NextAction{action}
	outcome := mustOutcome(t, engine.OutcomeInput{
		Command:          engine.CommandUnknown,
		RequestedCommand: "bad",
		Diagnostics:      diagnostics,
		NextActions:      actions,
		ExitCode:         engine.ExitUsage,
	})

	metadata["nested"].(map[string]any)["value"] = "mutated"
	metadata["items"].([]any)[0] = "mutated"
	args[0] = "mutated"
	flags["work"] = engine.StringFlag("mutated")
	diagnostics[0] = engine.Diagnostic{}
	actions[0] = engine.NextAction{}

	gotMetadata := outcome.Diagnostics()[0].Metadata()
	gotMetadata["nested"].(map[string]any)["value"] = "accessor mutation"
	gotArgs := outcome.NextActions()[0].Args()
	gotArgs[0] = "accessor mutation"

	againMetadata := outcome.Diagnostics()[0].Metadata()
	if got := againMetadata["nested"].(map[string]any)["value"]; got != "original" {
		t.Fatalf("nested metadata aliases caller storage: got %v", got)
	}
	if got := outcome.NextActions()[0].Args()[0]; got != "next" {
		t.Fatalf("next-action args alias caller storage: got %q", got)
	}
}

func TestResultConstructorsDefensivelyCopyCollections(t *testing.T) {
	t.Parallel()

	work := mustWork(t, "work-1", "one")
	items := []engine.WorkRef{work}
	status := mustStatusResult(t, items)
	items[0] = mustWork(t, "work-2", "two")
	statusValue, ok := status.Status()
	if !ok || statusValue.WorkItems()[0].ID() != "work-1" {
		t.Fatal("status result did not retain its immutable input snapshot")
	}

	buildInputs := map[string]string{"sourceDateEpoch": "0"}
	version := engine.NewVersionResult(engine.VersionResultInput{
		Version:     "0.1.0",
		Commit:      "abc",
		BuildTime:   "unknown",
		GoVersion:   "go1.24.0",
		BuildInputs: buildInputs,
	})
	buildInputs["sourceDateEpoch"] = "99"
	versionValue, ok := version.Version()
	if !ok || versionValue.BuildInputs()["sourceDateEpoch"] != "0" {
		t.Fatal("version result aliased build inputs")
	}

	checks := []map[string]any{{"name": "workspace", "ok": true}}
	doctor, err := engine.NewDoctorResult(checks)
	if err != nil {
		t.Fatal(err)
	}
	checks[0]["name"] = "mutated"
	doctorValue, ok := doctor.Doctor()
	if !ok || doctorValue.Checks()[0]["name"] != "workspace" {
		t.Fatal("doctor result aliased check objects")
	}
}

func TestOutcomeRejectsInvalidStateCombinations(t *testing.T) {
	t.Parallel()

	workspace := mustWorkspace(t)
	status := mustStatusResult(t, nil)
	diagnostic := mustDiagnostic(t, engine.DiagnosticInput{
		Code:     "DIAG_USAGE_ERROR",
		Severity: engine.SeverityError,
		Message:  "Invalid input",
		Source:   mustSource(t),
	})
	tests := []struct {
		name  string
		input engine.OutcomeInput
	}{
		{name: "unknown exit", input: engine.OutcomeInput{Command: engine.CommandStatus, Workspace: &workspace, Result: &status, ExitCode: 99}},
		{name: "success without result", input: engine.OutcomeInput{Command: engine.CommandStatus, Workspace: &workspace, ExitCode: engine.ExitSuccess}},
		{name: "failure without diagnostic", input: engine.OutcomeInput{Command: engine.CommandStatus, ExitCode: engine.ExitWorkspace}},
		{name: "mismatched result", input: engine.OutcomeInput{Command: engine.CommandVersion, Result: &status, ExitCode: engine.ExitSuccess}},
		{name: "unknown without requested command", input: engine.OutcomeInput{Command: engine.CommandUnknown, Diagnostics: []engine.Diagnostic{diagnostic}, ExitCode: engine.ExitUsage}},
		{name: "unknown with result", input: engine.OutcomeInput{Command: engine.CommandUnknown, RequestedCommand: "bad", Result: &status, Diagnostics: []engine.Diagnostic{diagnostic}, ExitCode: engine.ExitUsage}},
		{name: "requested command on known command", input: engine.OutcomeInput{Command: engine.CommandStatus, RequestedCommand: "status", Workspace: &workspace, Result: &status, ExitCode: engine.ExitSuccess}},
		{name: "workspace command without workspace", input: engine.OutcomeInput{Command: engine.CommandStatus, Result: &status, ExitCode: engine.ExitSuccess}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := engine.NewOutcome(test.input)
			if !errors.Is(err, engine.ErrInvalidProtocolValue) {
				t.Fatalf("expected invalid protocol value, got %v", err)
			}
		})
	}
}

func TestEmptyCollectionsMarshalAsArraysAndObjects(t *testing.T) {
	t.Parallel()

	workspace := mustWorkspace(t)
	result := mustStatusResult(t, nil)
	action, err := engine.NewNextAction(engine.NextActionInput{
		Command: "status",
		Reason:  "Refresh status",
	})
	if err != nil {
		t.Fatal(err)
	}
	outcome := mustOutcome(t, engine.OutcomeInput{
		Command:     engine.CommandStatus,
		Workspace:   &workspace,
		Result:      &result,
		NextActions: []engine.NextAction{action},
		ExitCode:    engine.ExitSuccess,
	})
	encoded, err := json.Marshal(outcome)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(encoded, &value); err != nil {
		t.Fatal(err)
	}
	if value["diagnostics"] == nil || value["nextActions"] == nil {
		t.Fatal("outcome collections must not marshal as null")
	}
	actionValue := value["nextActions"].([]any)[0].(map[string]any)
	if actionValue["args"] == nil || actionValue["flags"] == nil {
		t.Fatal("next-action collections must not marshal as null")
	}
}

func TestPlanValueValidation(t *testing.T) {
	t.Parallel()

	content := "workId: work-1\n"
	validDigest := "sha256:133452e2e3d639d26930125c30cbfe15eb23364cb6555c78b63c00842f69f750"
	operation, err := engine.NewPlanOperation(engine.PlanOperationInput{
		Kind:             engine.OperationCreateFile,
		Path:             ".sdd/work/new/work.yaml",
		Content:          &content,
		ContentDigest:    validDigest,
		ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	content = "mutated"
	if got := *operation.Content(); got != "workId: work-1\n" {
		t.Fatalf("operation content aliases input: got %q", got)
	}
	precondition, err := engine.NewPlanPrecondition(engine.PlanPreconditionInput{
		Kind:             engine.PreconditionAbsent,
		Path:             ".sdd/work/new/work.yaml",
		ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := engine.NewMutationPlan(engine.MutationPlanInput{
		PlanID: "plan-1", WorkspaceID: "ws-1", WorkID: "work-1",
		RevisionScope: engine.RevisionWorkspace, ExpectedRevision: 4,
		InputDigest: validDigest, PolicyDigest: validDigest,
		WorkflowDefinitionDigest: validDigest, ValidatorDefinitionDigest: validDigest,
		Operations: []engine.PlanOperation{operation}, Preconditions: []engine.PlanPrecondition{precondition},
		Digest: validDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Operations()) != 1 || len(plan.Preconditions()) != 1 {
		t.Fatal("valid plan did not retain its operations and preconditions")
	}

	_, err = engine.NewPlanOperation(engine.PlanOperationInput{Kind: engine.OperationCreateFile, Path: "file"})
	if !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("invalid operation error = %v", err)
	}
	_, err = engine.NewPlanPrecondition(engine.PlanPreconditionInput{Kind: engine.PreconditionDigest, Path: "file"})
	if !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("invalid precondition error = %v", err)
	}
	_, err = engine.NewMutationPlan(engine.MutationPlanInput{})
	if !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("invalid plan error = %v", err)
	}
}

func TestStructuredSourceAndFlagValidation(t *testing.T) {
	t.Parallel()

	start, err := engine.NewPosition(3, 7, engine.PositionUTF8Byte)
	if err != nil {
		t.Fatal(err)
	}
	end, err := engine.NewPosition(3, 12, engine.PositionUTF8Byte)
	if err != nil {
		t.Fatal(err)
	}
	rangeValue, err := engine.NewSourceRange("spec.md", start, &end)
	if err != nil {
		t.Fatal(err)
	}
	source, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{Kind: "artifact", Range: &rangeValue})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(encoded) || source.Range().End().Column() != 12 {
		t.Fatal("structured source range was not preserved")
	}

	if _, err := engine.NewPosition(0, 1, engine.PositionUTF8Byte); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("zero position error = %v", err)
	}
	if _, err := engine.NewSourceRange("spec.md", engine.Position{}, nil); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("zero start position error = %v", err)
	}
	if _, err := engine.NumberFlag(math.Inf(1)); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("infinite flag error = %v", err)
	}

	// Zero work references never reach the wire: the constructor rejects the
	// whole result instead of silently dropping observed data.
	if _, err := engine.NewStatusResult([]engine.WorkRef{{}, mustWork(t, "work-1", "one")}); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("zero work reference error = %v", err)
	}
}

func TestProtocolAccessorsAndMarshalVariants(t *testing.T) {
	t.Parallel()

	statusCommand := engine.NewStatusCommand("work-1", "/repo")
	if statusCommand.Kind() != engine.CommandStatus || statusCommand.WorkID() != "work-1" || statusCommand.Root() != "/repo" {
		t.Fatal("status command accessors lost input")
	}
	startCommand, err := engine.NewStartDryRunCommand("intent", "quick-spec", "/repo")
	if err != nil {
		t.Fatal(err)
	}
	if startCommand.Intent() != "intent" || startCommand.Workflow() != "quick-spec" || !startCommand.DryRun() {
		t.Fatal("start command accessors lost input")
	}

	workspace := mustWorkspace(t)
	if workspace.Root() != "/repo" || workspace.FormatVersion() != 1 {
		t.Fatal("workspace accessors lost input")
	}
	work, err := engine.NewWorkRef(engine.WorkRefInput{ID: "work-1", Slug: "slug", Workflow: "bug-fix", Revision: 7})
	if err != nil {
		t.Fatal(err)
	}
	if work.Slug() != "slug" || work.Workflow() != "bug-fix" || work.Revision() != 7 {
		t.Fatal("work accessors lost input")
	}
	marshalValidJSON(t, work)

	startPosition, err := engine.NewPosition(2, 3, engine.PositionUnicodeScalar)
	if err != nil {
		t.Fatal(err)
	}
	rangeValue, err := engine.NewSourceRange("requirements.md", startPosition, nil)
	if err != nil {
		t.Fatal(err)
	}
	if startPosition.Line() != 2 || startPosition.Unit() != engine.PositionUnicodeScalar || rangeValue.Path() != "requirements.md" || rangeValue.Start().Column() != 3 {
		t.Fatal("source range accessors lost input")
	}
	source, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{
		Kind: "artifact", Path: "requirements.md", Field: "REQ-001", ArtifactID: "artifact-1", Range: &rangeValue,
	})
	if err != nil {
		t.Fatal(err)
	}
	if source.Kind() != "artifact" || source.Path() != "requirements.md" || source.Field() != "REQ-001" || source.ArtifactID() != "artifact-1" {
		t.Fatal("diagnostic source accessors lost input")
	}
	diagnostic := mustDiagnostic(t, engine.DiagnosticInput{
		Code: "DIAG_USAGE_ERROR", Severity: engine.SeverityWarning, Message: "message", Source: source,
		Remediation: "fix it", Retryable: true,
	})
	if diagnostic.Code() != "DIAG_USAGE_ERROR" || diagnostic.Severity() != engine.SeverityWarning || diagnostic.Message() != "message" || diagnostic.Source().Kind() != "artifact" || diagnostic.Remediation() != "fix it" || !diagnostic.Retryable() {
		t.Fatal("diagnostic accessors lost input")
	}
	marshalValidJSON(t, diagnostic)

	numberFlag, err := engine.NumberFlag(1.25)
	if err != nil {
		t.Fatal(err)
	}
	flagValues := []engine.FlagValue{
		engine.StringFlag("value"), engine.BoolFlag(false), engine.IntegerFlag(42), numberFlag,
	}
	for _, flag := range flagValues {
		if flag.Value() == nil {
			t.Fatal("initialized flag returned nil")
		}
		marshalValidJSON(t, flag)
	}
	revision := uint64(8)
	action, err := engine.NewNextAction(engine.NextActionInput{
		Command: "next", Args: []string{"arg"}, Flags: map[string]engine.FlagValue{"count": engine.IntegerFlag(1)},
		Reason: "continue", ExpectedRevision: &revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if action.Command() != "next" || action.Reason() != "continue" || action.Flags()["count"].Value() != int64(1) || *action.ExpectedRevision() != 8 {
		t.Fatal("next-action accessors lost input")
	}

	versionResult := engine.NewVersionResult(engine.VersionResultInput{
		Version: "1.0.0", Commit: "abc", BuildTime: "unknown", GoVersion: "go1.24", BuildInputs: map[string]string{"epoch": "0"},
	})
	version, ok := versionResult.Version()
	if !ok || version.Version() != "1.0.0" || version.Commit() != "abc" || version.BuildTime() != "unknown" || version.GoVersion() != "go1.24" || version.ProtocolSchemaVersion() != engine.SchemaVersion {
		t.Fatal("version accessors lost input")
	}
	marshalValidJSON(t, version)

	doctorResult, err := engine.NewDoctorResult([]map[string]any{{"name": "workspace", "ok": true}})
	if err != nil {
		t.Fatal(err)
	}
	doctor, _ := doctorResult.Doctor()
	marshalValidJSON(t, doctor)

	plan := completePlan(t, ".sdd/work/slug/work.yaml", "workId: work-1\n")
	if plan.PlanID() != "plan-1" || plan.WorkspaceID() != "ws-1" || plan.WorkID() != "work-1" || plan.RevisionScope() != engine.RevisionWorkspace || plan.ExpectedRevision() != 4 || plan.InputDigest() == "" || plan.PolicyDigest() == "" || plan.WorkflowDefinitionDigest() == "" || plan.ValidatorDefinitionDigest() == "" || plan.Digest() == "" || plan.IdempotencyKey() != "idem-1" || plan.Authorizable() {
		t.Fatal("plan accessors lost input")
	}
	operation := plan.Operations()[0]
	if operation.Kind() != engine.OperationCreateFile || operation.Path() == "" || operation.ContentDigest() == "" || operation.ExpectedDigest() != "" || operation.ExpectedRevision() != 4 || operation.Authorizable() {
		t.Fatal("operation accessors lost input")
	}
	precondition := plan.Preconditions()[0]
	if precondition.Kind() != engine.PreconditionAbsent || precondition.Path() == "" || precondition.ExpectedRevision() != 4 || precondition.ExpectedDigest() != "" {
		t.Fatal("precondition accessors lost input")
	}
	marshalValidJSON(t, operation)
	marshalValidJSON(t, precondition)
	marshalValidJSON(t, plan)
	startResult := engine.NewStartDryRunResult(plan)
	startValue, ok := startResult.StartDryRun()
	if !ok || startValue.Plan().PlanID() != "plan-1" {
		t.Fatal("start dry-run result lost plan")
	}
	marshalValidJSON(t, startValue)
	for _, result := range []engine.Result{versionResult, doctorResult, startResult} {
		marshalValidJSON(t, result)
	}

	statusResult := mustStatusResult(t, []engine.WorkRef{work})
	outcome := mustOutcome(t, engine.OutcomeInput{
		Command: engine.CommandStatus, Workspace: &workspace, Work: &work, Result: &statusResult,
		Diagnostics: []engine.Diagnostic{diagnostic}, NextActions: []engine.NextAction{action}, ExitCode: engine.ExitSuccess,
	})
	if outcome.SchemaVersion() != engine.SchemaVersion || outcome.Command() != engine.CommandStatus || outcome.RequestedCommand() != "" || outcome.Workspace().Root() != "/repo" || outcome.Work().ID() != "work-1" || outcome.Result().Kind() != engine.ResultStatus || outcome.ExitCode() != engine.ExitSuccess {
		t.Fatal("outcome accessors lost input")
	}
}

func TestRedactedCoversEveryResultVariant(t *testing.T) {
	t.Parallel()

	replace := func(value string) string { return strings.ReplaceAll(value, "secret", "[x]") }
	workspace, err := engine.NewWorkspaceRef("/secret/repo", 1)
	if err != nil {
		t.Fatal(err)
	}
	work, err := engine.NewWorkRef(engine.WorkRefInput{ID: "secret-id", Slug: "secret-slug", Workflow: "secret-flow", Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	position, _ := engine.NewPosition(1, 1, engine.PositionUTF8Byte)
	rangeValue, _ := engine.NewSourceRange("secret-range", position, nil)
	source, _ := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{Kind: "file", Path: "secret-path", Field: "secret-field", ArtifactID: "secret-artifact", Range: &rangeValue})
	diagnostic := mustDiagnostic(t, engine.DiagnosticInput{
		Code: "DIAG_USAGE_ERROR", Severity: engine.SeverityWarning, Message: "secret-message", Source: source,
		Remediation: "secret-remediation", Metadata: map[string]any{"value": "secret", "items": []any{"secret", 7}},
	})
	revision := uint64(1)
	action, err := engine.NewNextAction(engine.NextActionInput{
		Command: "secret-command", Args: []string{"secret-arg"}, Flags: map[string]engine.FlagValue{"value": engine.StringFlag("secret-flag")},
		Reason: "secret-reason", ExpectedRevision: &revision,
	})
	if err != nil {
		t.Fatal(err)
	}

	statusResult := mustStatusResult(t, []engine.WorkRef{work})
	statusOutcome := mustOutcome(t, engine.OutcomeInput{
		Command: engine.CommandStatus, Workspace: &workspace, Work: &work, Result: &statusResult,
		Diagnostics: []engine.Diagnostic{diagnostic}, NextActions: []engine.NextAction{action}, ExitCode: engine.ExitSuccess,
	})
	redactedStatus := statusOutcome.Redacted(replace)
	encoded, err := json.Marshal(redactedStatus)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret") || !strings.Contains(string(encoded), "[x]") {
		t.Fatalf("status outcome was not fully redacted: %s", encoded)
	}
	if !strings.Contains(statusOutcome.Workspace().Root(), "secret") {
		t.Fatal("redaction mutated original status outcome")
	}

	versionResult := engine.NewVersionResult(engine.VersionResultInput{
		Version: "secret-version", Commit: "secret-commit", BuildTime: "secret-time", GoVersion: "secret-go", BuildInputs: map[string]string{"env": "secret-value"},
	})
	versionOutcome := mustOutcome(t, engine.OutcomeInput{Command: engine.CommandVersion, Result: &versionResult, ExitCode: engine.ExitSuccess})
	assertNoSecret(t, versionOutcome.Redacted(replace))

	doctorResult, err := engine.NewDoctorResult([]map[string]any{{"message": "secret", "nested": map[string]any{"value": "secret"}}})
	if err != nil {
		t.Fatal(err)
	}
	doctorOutcome := mustOutcome(t, engine.OutcomeInput{Command: engine.CommandDoctor, Workspace: &workspace, Result: &doctorResult, ExitCode: engine.ExitSuccess})
	assertNoSecret(t, doctorOutcome.Redacted(replace))

	plan := completePlan(t, "secret-path", "secret-content")
	startResult := engine.NewStartDryRunResult(plan)
	startOutcome := mustOutcome(t, engine.OutcomeInput{Command: engine.CommandStart, Workspace: &workspace, Result: &startResult, ExitCode: engine.ExitSuccess})
	assertNoSecret(t, startOutcome.Redacted(replace))

	unknownOutcome := mustOutcome(t, engine.OutcomeInput{
		Command: engine.CommandUnknown, RequestedCommand: "secret-command", Diagnostics: []engine.Diagnostic{diagnostic}, ExitCode: engine.ExitUsage,
	})
	if got := unknownOutcome.Redacted(replace).RequestedCommand(); strings.Contains(got, "secret") {
		t.Fatalf("requested command not redacted: %q", got)
	}
	if got := unknownOutcome.Redacted(nil).RequestedCommand(); got != "secret-command" {
		t.Fatalf("nil redactor changed outcome: %q", got)
	}
}

func TestInvalidProtocolValueBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "workspace root", run: func() error { _, err := engine.NewWorkspaceRef("", 1); return err }},
		{name: "workspace version", run: func() error { _, err := engine.NewWorkspaceRef("/repo", 0); return err }},
		{name: "work ID", run: func() error { _, err := engine.NewWorkRef(engine.WorkRefInput{Slug: "slug"}); return err }},
		{name: "position unit", run: func() error { _, err := engine.NewPosition(1, 1, "grapheme"); return err }},
		{name: "range path", run: func() error {
			position, _ := engine.NewPosition(1, 1, engine.PositionUTF8Byte)
			_, err := engine.NewSourceRange("", position, nil)
			return err
		}},
		{name: "source kind", run: func() error { _, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{}); return err }},
		{name: "diagnostic code", run: func() error {
			_, err := engine.NewDiagnostic(engine.DiagnosticInput{Code: "bad", Severity: engine.SeverityError, Message: "message", Source: mustSource(t)})
			return err
		}},
		{name: "diagnostic severity", run: func() error {
			_, err := engine.NewDiagnostic(engine.DiagnosticInput{Code: "DIAG_BAD", Severity: "fatal", Message: "message", Source: mustSource(t)})
			return err
		}},
		{name: "diagnostic message", run: func() error {
			_, err := engine.NewDiagnostic(engine.DiagnosticInput{Code: "DIAG_BAD", Severity: engine.SeverityError, Source: mustSource(t)})
			return err
		}},
		{name: "diagnostic source", run: func() error {
			_, err := engine.NewDiagnostic(engine.DiagnosticInput{Code: "DIAG_BAD", Severity: engine.SeverityError, Message: "message"})
			return err
		}},
		{name: "diagnostic metadata", run: func() error {
			_, err := engine.NewDiagnostic(engine.DiagnosticInput{Code: "DIAG_BAD", Severity: engine.SeverityError, Message: "message", Source: mustSource(t), Metadata: map[string]any{"bad": make(chan int)}})
			return err
		}},
		{name: "NaN flag", run: func() error { _, err := engine.NumberFlag(math.NaN()); return err }},
		{name: "action command", run: func() error {
			_, err := engine.NewNextAction(engine.NextActionInput{Command: "two words", Reason: "reason"})
			return err
		}},
		{name: "action reason", run: func() error { _, err := engine.NewNextAction(engine.NextActionInput{Command: "next"}); return err }},
		{name: "action flag", run: func() error {
			_, err := engine.NewNextAction(engine.NextActionInput{Command: "next", Reason: "reason", Flags: map[string]engine.FlagValue{"": {}}})
			return err
		}},
		{name: "outcome command", run: func() error {
			_, err := engine.NewOutcome(engine.OutcomeInput{Command: "future", ExitCode: engine.ExitUsage, Diagnostics: []engine.Diagnostic{mustDiagnostic(t, engine.DiagnosticInput{Code: "DIAG_BAD", Severity: engine.SeverityError, Message: "bad", Source: mustSource(t)})}})
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.run(); !errors.Is(err, engine.ErrInvalidProtocolValue) {
				t.Fatalf("error = %v", err)
			}
		})
	}

	if _, err := json.Marshal(engine.Command{}); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("zero command marshal error = %v", err)
	}
	if _, err := json.Marshal(engine.FlagValue{}); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("zero flag marshal error = %v", err)
	}
	if _, err := json.Marshal(engine.Result{}); !errors.Is(err, engine.ErrInvalidProtocolValue) {
		t.Fatalf("zero result marshal error = %v", err)
	}
}

func TestOperationAndPreconditionVariants(t *testing.T) {
	t.Parallel()

	digest := protocolTestDigest
	content := "new"
	operations := []engine.PlanOperationInput{
		{Kind: engine.OperationCreateFile, Path: "new", Content: &content, ContentDigest: digest},
		{Kind: engine.OperationUpdateFile, Path: "existing", Content: &content, ContentDigest: digest, ExpectedDigest: digest},
		{Kind: engine.OperationDeleteFile, Path: "old", ExpectedDigest: digest},
	}
	for _, input := range operations {
		operation, err := engine.NewPlanOperation(input)
		if err != nil {
			t.Fatalf("valid %s operation: %v", input.Kind, err)
		}
		marshalValidJSON(t, operation)
	}
	invalidOperations := []engine.PlanOperationInput{
		{Kind: engine.OperationCreateFile, Path: "new", Content: &content, ContentDigest: digest, ExpectedDigest: digest},
		{Kind: engine.OperationUpdateFile, Path: "existing", Content: &content, ContentDigest: digest},
		{Kind: engine.OperationDeleteFile, Path: "old", Content: &content, ExpectedDigest: digest},
		{Kind: "future", Path: "file"},
		{Kind: engine.OperationCreateFile},
		{Kind: engine.OperationCreateFile, Path: "../escape", Content: &content, ContentDigest: digest},
		{Kind: engine.OperationCreateFile, Path: "/absolute", Content: &content, ContentDigest: digest},
		{Kind: engine.OperationCreateFile, Path: `dir\escape`, Content: &content, ContentDigest: digest},
		{Kind: engine.OperationCreateFile, Path: "C:/absolute", Content: &content, ContentDigest: digest},
		{Kind: engine.OperationCreateFile, Path: "dir/../escape", Content: &content, ContentDigest: digest},
	}
	for _, input := range invalidOperations {
		if _, err := engine.NewPlanOperation(input); !errors.Is(err, engine.ErrInvalidProtocolValue) {
			t.Fatalf("invalid operation %+v error = %v", input, err)
		}
	}

	preconditions := []engine.PlanPreconditionInput{
		{Kind: engine.PreconditionRevision, Path: "file", ExpectedRevision: 1},
		{Kind: engine.PreconditionDigest, Path: "file", ExpectedDigest: digest},
		{Kind: engine.PreconditionAbsent, Path: "file"},
	}
	for _, input := range preconditions {
		value, err := engine.NewPlanPrecondition(input)
		if err != nil {
			t.Fatalf("valid %s precondition: %v", input.Kind, err)
		}
		marshalValidJSON(t, value)
	}
	invalidPreconditions := []engine.PlanPreconditionInput{
		{Kind: engine.PreconditionRevision, Path: "file", ExpectedDigest: digest},
		{Kind: engine.PreconditionDigest, Path: "file", ExpectedDigest: "bad"},
		{Kind: "future", Path: "file"},
		{Kind: engine.PreconditionAbsent},
		{Kind: engine.PreconditionAbsent, Path: "../escape"},
		{Kind: engine.PreconditionAbsent, Path: "/absolute"},
		{Kind: engine.PreconditionAbsent, Path: `dir\escape`},
	}
	for _, input := range invalidPreconditions {
		if _, err := engine.NewPlanPrecondition(input); !errors.Is(err, engine.ErrInvalidProtocolValue) {
			t.Fatalf("invalid precondition %+v error = %v", input, err)
		}
	}

	valid := completePlanInput(t)
	invalidPlans := []engine.MutationPlanInput{
		func() engine.MutationPlanInput { value := valid; value.RevisionScope = "future"; return value }(),
		func() engine.MutationPlanInput { value := valid; value.InputDigest = "bad"; return value }(),
		func() engine.MutationPlanInput { value := valid; value.Operations = nil; return value }(),
		func() engine.MutationPlanInput { value := valid; value.Preconditions = nil; return value }(),
	}
	for _, input := range invalidPlans {
		if _, err := engine.NewMutationPlan(input); !errors.Is(err, engine.ErrInvalidProtocolValue) {
			t.Fatalf("invalid plan error = %v", err)
		}
	}
}

const protocolTestDigest = "sha256:133452e2e3d639d26930125c30cbfe15eb23364cb6555c78b63c00842f69f750"

func completePlan(t *testing.T, path, contentValue string) engine.MutationPlan {
	t.Helper()
	input := completePlanInput(t)
	operation, err := engine.NewPlanOperation(engine.PlanOperationInput{
		Kind: engine.OperationCreateFile, Path: path, Content: &contentValue,
		ContentDigest: protocolTestDigest, ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	precondition, err := engine.NewPlanPrecondition(engine.PlanPreconditionInput{
		Kind: engine.PreconditionAbsent, Path: path, ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	input.Operations = []engine.PlanOperation{operation}
	input.Preconditions = []engine.PlanPrecondition{precondition}
	plan, err := engine.NewMutationPlan(input)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func completePlanInput(t *testing.T) engine.MutationPlanInput {
	t.Helper()
	content := "content"
	operation, err := engine.NewPlanOperation(engine.PlanOperationInput{
		Kind: engine.OperationCreateFile, Path: "file", Content: &content, ContentDigest: protocolTestDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	precondition, err := engine.NewPlanPrecondition(engine.PlanPreconditionInput{Kind: engine.PreconditionAbsent, Path: "file"})
	if err != nil {
		t.Fatal(err)
	}
	return engine.MutationPlanInput{
		PlanID: "plan-1", WorkspaceID: "ws-1", WorkID: "work-1", RevisionScope: engine.RevisionWorkspace,
		ExpectedRevision: 4, InputDigest: protocolTestDigest, PolicyDigest: protocolTestDigest,
		WorkflowDefinitionDigest: protocolTestDigest, ValidatorDefinitionDigest: protocolTestDigest,
		Operations: []engine.PlanOperation{operation}, Preconditions: []engine.PlanPrecondition{precondition},
		Digest: protocolTestDigest, IdempotencyKey: "idem-1",
	}
}

func marshalValidJSON(t *testing.T, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(encoded) {
		t.Fatalf("invalid JSON: %s", encoded)
	}
}

func assertNoSecret(t *testing.T, outcome engine.Outcome) {
	t.Helper()
	encoded, err := json.Marshal(outcome)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret") {
		t.Fatalf("redacted outcome leaked secret: %s", encoded)
	}
}

func mustWorkspace(t *testing.T) engine.WorkspaceRef {
	t.Helper()
	value, err := engine.NewWorkspaceRef("/repo", 1)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustWork(t *testing.T, id, slug string) engine.WorkRef {
	t.Helper()
	value, err := engine.NewWorkRef(engine.WorkRefInput{ID: id, Slug: slug, Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustStatusResult(t *testing.T, workItems []engine.WorkRef) engine.Result {
	t.Helper()
	value, err := engine.NewStatusResult(workItems)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustSource(t *testing.T) engine.DiagnosticSource {
	t.Helper()
	value, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{Kind: "command"})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustDiagnostic(t *testing.T, input engine.DiagnosticInput) engine.Diagnostic {
	t.Helper()
	value, err := engine.NewDiagnostic(input)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustOutcome(t *testing.T, input engine.OutcomeInput) engine.Outcome {
	t.Helper()
	value, err := engine.NewOutcome(input)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func assertJSONEqual(t *testing.T, got, want []byte) {
	t.Helper()
	var gotValue, wantValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("decode got JSON: %v", err)
	}
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("decode wanted JSON: %v", err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("JSON mismatch\ngot:  %s\nwant: %s", got, want)
	}
}
