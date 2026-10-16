package render_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/presentation/render"
)

func TestHumanAndJSONRenderTheSamePreparedOutcome(t *testing.T) {
	t.Parallel()

	workspace := mustWorkspace(t, "/repo")
	result := engine.NewStatusResult(nil)
	action, err := engine.NewNextAction(engine.NextActionInput{
		Command: "next",
		Args:    []string{"two words", "it's-safe"},
		Flags: map[string]engine.FlagValue{
			"work": engine.StringFlag("work-1"),
			"json": engine.BoolFlag(true),
		},
		Reason: "Continue",
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
	input := render.NewInput(outcome, render.NewRedactor())

	var humanOut, humanErr, jsonOut bytes.Buffer
	if err := render.WriteHuman(&humanOut, &humanErr, input); err != nil {
		t.Fatal(err)
	}
	if err := render.WriteJSON(&jsonOut, input); err != nil {
		t.Fatal(err)
	}
	if got, want := humanOut.String(), "No active work items.\nNext: sdd next 'two words' 'it'\"'\"'s-safe' --json --work work-1\n"; got != want {
		t.Fatalf("human stdout\ngot:  %q\nwant: %q", got, want)
	}
	if humanErr.Len() != 0 {
		t.Fatalf("successful human rendering wrote stderr: %q", humanErr.String())
	}
	if !strings.Contains(jsonOut.String(), `"command":"status"`) || !strings.Contains(jsonOut.String(), `"nextActions":[`) {
		t.Fatalf("JSON rendering did not contain shared outcome: %s", jsonOut.String())
	}
}

func TestRedactionIsAppliedOnceToBothRenderers(t *testing.T) {
	t.Parallel()

	secret := "company-secret-value"
	source, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{
		Kind: "file",
		Path: "/repo/" + secret + "/spec.md",
	})
	if err != nil {
		t.Fatal(err)
	}
	diagnostic, err := engine.NewDiagnostic(engine.DiagnosticInput{
		Code:        "DIAG_USAGE_ERROR",
		Severity:    engine.SeverityError,
		Message:     "token=abc123 Bearer ey.secret " + secret + " https://user:pass@example.test",
		Source:      source,
		Remediation: "remove password=hunter2",
		Metadata:    map[string]any{"nested": []any{"api_key=qwerty", secret}},
	})
	if err != nil {
		t.Fatal(err)
	}
	outcome := mustOutcome(t, engine.OutcomeInput{
		Command:          engine.CommandUnknown,
		RequestedCommand: "bad-" + secret,
		Diagnostics:      []engine.Diagnostic{diagnostic},
		ExitCode:         engine.ExitUsage,
	})
	input := render.NewInput(outcome, render.NewRedactor(secret))

	var humanOut, humanErr, jsonOut bytes.Buffer
	if err := render.WriteHuman(&humanOut, &humanErr, input); err != nil {
		t.Fatal(err)
	}
	if err := render.WriteJSON(&jsonOut, input); err != nil {
		t.Fatal(err)
	}
	combined := humanOut.String() + humanErr.String() + jsonOut.String()
	for _, leaked := range []string{secret, "abc123", "ey.secret", "hunter2", "qwerty", "user:pass"} {
		if strings.Contains(combined, leaked) {
			t.Fatalf("rendered output leaked %q: %s", leaked, combined)
		}
	}
	if count := strings.Count(combined, "[REDACTED]"); count < 8 {
		t.Fatalf("expected redaction markers, got %d in %s", count, combined)
	}
	if got := outcome.Diagnostics()[0].Message(); !strings.Contains(got, secret) {
		t.Fatal("preparing renderer input mutated the authoritative outcome")
	}
}

func TestRendererWriteErrorsAreReturned(t *testing.T) {
	t.Parallel()

	outcome := successfulStatus(t)
	input := render.NewInput(outcome, render.NewRedactor())
	failure := errors.New("transport closed")

	if err := render.WriteJSON(errorWriter{err: failure}, input); !errors.Is(err, render.ErrWrite) || !errors.Is(err, failure) {
		t.Fatalf("JSON write error = %v", err)
	}
	if err := render.WriteJSON(shortWriter{}, input); !errors.Is(err, render.ErrWrite) || !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("JSON short write error = %v", err)
	}
	if err := render.WriteHuman(errorWriter{err: failure}, io.Discard, input); !errors.Is(err, render.ErrWrite) {
		t.Fatalf("human stdout error = %v", err)
	}

	diagnostic := mustDiagnostic(t)
	failed := mustOutcome(t, engine.OutcomeInput{
		Command:          engine.CommandUnknown,
		RequestedCommand: "bad",
		Diagnostics:      []engine.Diagnostic{diagnostic},
		ExitCode:         engine.ExitUsage,
	})
	failedInput := render.NewInput(failed, render.NewRedactor())
	if err := render.WriteHuman(io.Discard, errorWriter{err: failure}, failedInput); !errors.Is(err, render.ErrWrite) {
		t.Fatalf("human stderr error = %v", err)
	}
}

func TestFailureUsesStderrAndPreservesStructuredJSON(t *testing.T) {
	t.Parallel()

	diagnostic := mustDiagnostic(t)
	action, err := engine.NewNextAction(engine.NextActionInput{
		Command: "status",
		Reason:  "Retry with a valid workspace",
	})
	if err != nil {
		t.Fatal(err)
	}
	outcome := mustOutcome(t, engine.OutcomeInput{
		Command:          engine.CommandUnknown,
		RequestedCommand: "bogus",
		Diagnostics:      []engine.Diagnostic{diagnostic},
		NextActions:      []engine.NextAction{action},
		ExitCode:         engine.ExitUsage,
	})
	input := render.NewInput(outcome, render.NewRedactor())
	var stdout, stderr, jsonOutput bytes.Buffer
	if err := render.WriteHuman(&stdout, &stderr, input); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("failed human outcome wrote stdout: %q", stdout.String())
	}
	if got := stderr.String(); !strings.Contains(got, "error [DIAG_COMMAND_UNKNOWN]") || !strings.Contains(got, "Next: sdd status") {
		t.Fatalf("human failure missing diagnostic or action: %q", got)
	}
	if err := render.WriteJSON(&jsonOutput, input); err != nil {
		t.Fatal(err)
	}
	if strings.Count(jsonOutput.String(), "\n") != 1 || !strings.Contains(jsonOutput.String(), `"result":null`) {
		t.Fatalf("machine output is not one complete envelope: %q", jsonOutput.String())
	}
}

func TestHumanSummariesCoverEveryResultVariant(t *testing.T) {
	t.Parallel()

	workspace := mustWorkspace(t, "/repo")
	version := engine.NewVersionResult(engine.VersionResultInput{
		Version: "0.1.0-dev", Commit: "unknown", BuildTime: "unknown", GoVersion: "go1.24.0",
	})
	doctor, err := engine.NewDoctorResult([]map[string]any{{"name": "workspace", "ok": true}})
	if err != nil {
		t.Fatal(err)
	}
	content := "workId: work-1\n"
	operation, err := engine.NewPlanOperation(engine.PlanOperationInput{
		Kind:             engine.OperationCreateFile,
		Path:             ".sdd/work/new/work.yaml",
		Content:          &content,
		ContentDigest:    "sha256:133452e2e3d639d26930125c30cbfe15eb23364cb6555c78b63c00842f69f750",
		ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	precondition, err := engine.NewPlanPrecondition(engine.PlanPreconditionInput{
		Kind: engine.PreconditionAbsent, Path: ".sdd/work/new/work.yaml", ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := engine.NewMutationPlan(engine.MutationPlanInput{
		PlanID: "plan-1", WorkspaceID: "ws-1", WorkID: "work-1",
		RevisionScope: engine.RevisionWorkspace, ExpectedRevision: 4,
		InputDigest:               "sha256:b727d8da7e2cbd385d9e474234dda18a9677913b37f40a8074b410d2efbce58c",
		PolicyDigest:              "sha256:6d96b4f675f734210734c22f9d1f78eb21ee03316fe04068cf64f1ebdf97f3a6",
		WorkflowDefinitionDigest:  "sha256:44e103af0242e49bc680a1ee30f96c847814cdf79401b1f6f10b1b7275bc623d",
		ValidatorDefinitionDigest: "sha256:e25d7eb5474afb9a12bf09813d869079e847bf64c9d79c12b68939870af944a6",
		Operations:                []engine.PlanOperation{operation}, Preconditions: []engine.PlanPrecondition{precondition},
		Digest: "sha256:58463e47efe0a8dcc69122060fa085f0924a66a4fc6d6bd9af78158eacf99a1d",
	})
	if err != nil {
		t.Fatal(err)
	}
	start := engine.NewStartDryRunResult(plan)
	status := engine.NewStatusResult(nil)

	cases := []struct {
		name    string
		outcome engine.Outcome
		want    string
	}{
		{name: "version", outcome: mustOutcome(t, engine.OutcomeInput{Command: engine.CommandVersion, Result: &version, ExitCode: engine.ExitSuccess}), want: "sdd 0.1.0-dev (unknown)\n"},
		{name: "doctor", outcome: mustOutcome(t, engine.OutcomeInput{Command: engine.CommandDoctor, Workspace: &workspace, Result: &doctor, ExitCode: engine.ExitSuccess}), want: "Checked 1 system condition(s).\n"},
		{name: "start", outcome: mustOutcome(t, engine.OutcomeInput{Command: engine.CommandStart, Workspace: &workspace, Result: &start, ExitCode: engine.ExitSuccess}), want: "Prepared plan plan-1 with 1 operation(s).\n"},
		{name: "status", outcome: mustOutcome(t, engine.OutcomeInput{Command: engine.CommandStatus, Workspace: &workspace, Result: &status, ExitCode: engine.ExitSuccess}), want: "No active work items.\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			input := render.NewInput(tc.outcome, render.NewRedactor())
			if got := input.Outcome().Command(); got != tc.outcome.Command() {
				t.Fatalf("renderer input lost outcome command: got %q want %q", got, tc.outcome.Command())
			}
			var stdout, stderr bytes.Buffer
			if err := render.WriteHuman(&stdout, &stderr, input); err != nil {
				t.Fatal(err)
			}
			if stdout.String() != tc.want {
				t.Fatalf("human summary\ngot:  %q\nwant: %q", stdout.String(), tc.want)
			}
		})
	}
}

func TestRedactionCoversJSONCredentialForms(t *testing.T) {
	t.Parallel()

	source, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{Kind: "file"})
	if err != nil {
		t.Fatal(err)
	}
	diagnostic, err := engine.NewDiagnostic(engine.DiagnosticInput{
		Code:     "DIAG_USAGE_ERROR",
		Severity: engine.SeverityError,
		Message:  `auth failed for {"token":"abc123"} with password="alpha beta"`,
		Source:   source,
		Metadata: map[string]any{
			"/repo/token=hunter2": "path-label",
			"password":            "s3cr3t-value",
			"nested":              map[string]any{"api_key": "qwerty"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	outcome := mustOutcome(t, engine.OutcomeInput{
		Command:          engine.CommandUnknown,
		RequestedCommand: "bad",
		Diagnostics:      []engine.Diagnostic{diagnostic},
		ExitCode:         engine.ExitUsage,
	})
	input := render.NewInput(outcome, render.NewRedactor())

	var jsonOut bytes.Buffer
	if err := render.WriteJSON(&jsonOut, input); err != nil {
		t.Fatal(err)
	}
	rendered := jsonOut.String()
	for _, leaked := range []string{"abc123", "alpha", "beta", "hunter2", "qwerty", "s3cr3t-value"} {
		if strings.Contains(rendered, leaked) {
			t.Fatalf("JSON credential form leaked %q: %s", leaked, rendered)
		}
	}
	if !strings.Contains(rendered, "[REDACTED]") {
		t.Fatalf("no redaction markers in %s", rendered)
	}
	// The authoritative outcome still carries exact metadata keys and values.
	original := outcome.Diagnostics()[0].Metadata()
	if _, ok := original["/repo/token=hunter2"]; !ok {
		t.Fatalf("redaction mutated authoritative metadata keys: %v", original)
	}
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }

type shortWriter struct{}

func (shortWriter) Write(payload []byte) (int, error) { return len(payload) - 1, nil }

func successfulStatus(t *testing.T) engine.Outcome {
	t.Helper()
	workspace := mustWorkspace(t, "/repo")
	result := engine.NewStatusResult(nil)
	return mustOutcome(t, engine.OutcomeInput{
		Command:   engine.CommandStatus,
		Workspace: &workspace,
		Result:    &result,
		ExitCode:  engine.ExitSuccess,
	})
}

func mustWorkspace(t *testing.T, root string) engine.WorkspaceRef {
	t.Helper()
	value, err := engine.NewWorkspaceRef(root, 1)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustDiagnostic(t *testing.T) engine.Diagnostic {
	t.Helper()
	source, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{Kind: "command"})
	if err != nil {
		t.Fatal(err)
	}
	value, err := engine.NewDiagnostic(engine.DiagnosticInput{
		Code:        "DIAG_COMMAND_UNKNOWN",
		Severity:    engine.SeverityError,
		Message:     "Unknown command",
		Source:      source,
		Remediation: "Run sdd --help",
	})
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
