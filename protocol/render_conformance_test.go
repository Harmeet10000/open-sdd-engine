package protocol

import (
	"bytes"
	"encoding/json"
	"testing"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/presentation/render"
)

func TestTypedCommandsConformToPublishedSchema(t *testing.T) {
	t.Parallel()

	start, err := engine.NewStartDryRunCommand("add export", "quick-spec", "/repo")
	if err != nil {
		t.Fatal(err)
	}
	schema := compileFoundationSchema(t, "#/definitions/Command")
	commands := []engine.Command{
		engine.NewStatusCommand("work-1", "/repo"),
		engine.NewVersionCommand(),
		engine.NewDoctorCommand("/repo"),
		start,
	}
	for _, command := range commands {
		command := command
		t.Run(string(command.Kind()), func(t *testing.T) {
			t.Parallel()
			encoded, err := json.Marshal(command)
			if err != nil {
				t.Fatal(err)
			}
			if err := schema.Validate(decodeJSONBytes(t, encoded)); err != nil {
				t.Fatalf("typed command rejected by published schema: %v\n%s", err, encoded)
			}
		})
	}
}

func TestRenderedOutcomesConformToPublishedSchema(t *testing.T) {
	t.Parallel()

	schema := compileFoundationSchema(t, "")
	outcomes := []engine.Outcome{
		statusOutcome(t),
		versionOutcome(t),
		doctorOutcome(t),
		unknownOutcome(t),
		startDryRunOutcome(t),
	}
	for _, outcome := range outcomes {
		outcome := outcome
		t.Run(string(outcome.Command()), func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			input := render.NewInput(outcome, render.NewRedactor())
			if err := render.WriteJSON(&output, input); err != nil {
				t.Fatal(err)
			}
			if err := schema.Validate(decodeJSONBytes(t, output.Bytes())); err != nil {
				t.Fatalf("rendered outcome rejected by published schema: %v\n%s", err, output.Bytes())
			}
		})
	}
}

func statusOutcome(t *testing.T) engine.Outcome {
	t.Helper()
	workspace := conformanceWorkspace(t)
	result := engine.NewStatusResult(nil)
	return conformanceOutcome(t, engine.OutcomeInput{
		Command:   engine.CommandStatus,
		Workspace: &workspace,
		Result:    &result,
		ExitCode:  engine.ExitSuccess,
	})
}

func versionOutcome(t *testing.T) engine.Outcome {
	t.Helper()
	result := engine.NewVersionResult(engine.VersionResultInput{
		Version:     "0.1.0-dev",
		Commit:      "unknown",
		BuildTime:   "unknown",
		GoVersion:   "go1.24.0",
		BuildInputs: map[string]string{"sourceDateEpoch": "0"},
	})
	return conformanceOutcome(t, engine.OutcomeInput{
		Command:  engine.CommandVersion,
		Result:   &result,
		ExitCode: engine.ExitSuccess,
	})
}

func doctorOutcome(t *testing.T) engine.Outcome {
	t.Helper()
	doctor, err := engine.NewDoctorResult([]map[string]any{{"name": "workspace", "ok": true}})
	if err != nil {
		t.Fatal(err)
	}
	workspace := conformanceWorkspace(t)
	return conformanceOutcome(t, engine.OutcomeInput{
		Command:   engine.CommandDoctor,
		Workspace: &workspace,
		Result:    &doctor,
		ExitCode:  engine.ExitSuccess,
	})
}

func unknownOutcome(t *testing.T) engine.Outcome {
	t.Helper()
	source, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{Kind: "command"})
	if err != nil {
		t.Fatal(err)
	}
	diagnostic, err := engine.NewDiagnostic(engine.DiagnosticInput{
		Code:     "DIAG_COMMAND_UNKNOWN",
		Severity: engine.SeverityError,
		Message:  "Unknown command",
		Source:   source,
	})
	if err != nil {
		t.Fatal(err)
	}
	return conformanceOutcome(t, engine.OutcomeInput{
		Command:          engine.CommandUnknown,
		RequestedCommand: "bogus",
		Diagnostics:      []engine.Diagnostic{diagnostic},
		ExitCode:         engine.ExitUsage,
	})
}

func startDryRunOutcome(t *testing.T) engine.Outcome {
	t.Helper()
	const (
		contentDigest   = "sha256:133452e2e3d639d26930125c30cbfe15eb23364cb6555c78b63c00842f69f750"
		inputDigest     = "sha256:b727d8da7e2cbd385d9e474234dda18a9677913b37f40a8074b410d2efbce58c"
		policyDigest    = "sha256:6d96b4f675f734210734c22f9d1f78eb21ee03316fe04068cf64f1ebdf97f3a6"
		workflowDigest  = "sha256:44e103af0242e49bc680a1ee30f96c847814cdf79401b1f6f10b1b7275bc623d"
		validatorDigest = "sha256:e25d7eb5474afb9a12bf09813d869079e847bf64c9d79c12b68939870af944a6"
		planDigest      = "sha256:58463e47efe0a8dcc69122060fa085f0924a66a4fc6d6bd9af78158eacf99a1d"
	)
	content := "workId: work-1\n"
	operation, err := engine.NewPlanOperation(engine.PlanOperationInput{
		Kind:             engine.OperationCreateFile,
		Path:             ".sdd/work/new/work.yaml",
		Content:          &content,
		ContentDigest:    contentDigest,
		ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
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
		PlanID:                    "plan-1",
		WorkspaceID:               "ws-1",
		WorkID:                    "work-1",
		RevisionScope:             engine.RevisionWorkspace,
		ExpectedRevision:          4,
		InputDigest:               inputDigest,
		PolicyDigest:              policyDigest,
		WorkflowDefinitionDigest:  workflowDigest,
		ValidatorDefinitionDigest: validatorDigest,
		Operations:                []engine.PlanOperation{operation},
		Preconditions:             []engine.PlanPrecondition{precondition},
		Digest:                    planDigest,
		Authorizable:              false,
	})
	if err != nil {
		t.Fatal(err)
	}
	result := engine.NewStartDryRunResult(plan)
	source, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{Kind: "plan"})
	if err != nil {
		t.Fatal(err)
	}
	diagnostic, err := engine.NewDiagnostic(engine.DiagnosticInput{
		Code:     "DIAG_PLAN_NOT_AUTHORIZABLE",
		Severity: engine.SeverityWarning,
		Message:  "Preview only; durable application is not implemented in foundation",
		Source:   source,
	})
	if err != nil {
		t.Fatal(err)
	}
	workspace := conformanceWorkspace(t)
	return conformanceOutcome(t, engine.OutcomeInput{
		Command:     engine.CommandStart,
		Workspace:   &workspace,
		Result:      &result,
		Diagnostics: []engine.Diagnostic{diagnostic},
		ExitCode:    engine.ExitSuccess,
	})
}

func conformanceWorkspace(t *testing.T) engine.WorkspaceRef {
	t.Helper()
	workspace, err := engine.NewWorkspaceRef("/repo", 1)
	if err != nil {
		t.Fatal(err)
	}
	return workspace
}

func conformanceOutcome(t *testing.T, input engine.OutcomeInput) engine.Outcome {
	t.Helper()
	outcome, err := engine.NewOutcome(input)
	if err != nil {
		t.Fatal(err)
	}
	return outcome
}
