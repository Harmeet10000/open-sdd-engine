package engine_test

import (
	"encoding/json"
	"strings"
	"testing"

	"sdd.dev/sdd/internal/engine"
)

// TestRedactionLeavesOriginalOutcomeExact proves the presentation copy never
// aliases the authoritative outcome: redacting paths, content, and source
// ranges must leave digests and originals verifiable.
func TestRedactionLeavesOriginalOutcomeExact(t *testing.T) {
	t.Parallel()

	replace := func(value string) string { return strings.ReplaceAll(value, "secret", "[x]") }
	position, err := engine.NewPosition(1, 1, engine.PositionUTF8Byte)
	if err != nil {
		t.Fatal(err)
	}
	rangeValue, err := engine.NewSourceRange("secret-range", position, nil)
	if err != nil {
		t.Fatal(err)
	}
	source, err := engine.NewDiagnosticSource(engine.DiagnosticSourceInput{Kind: "file", Range: &rangeValue})
	if err != nil {
		t.Fatal(err)
	}
	diagnostic := mustDiagnostic(t, engine.DiagnosticInput{
		Code: "DIAG_USAGE_ERROR", Severity: engine.SeverityError, Message: "secret-message", Source: source,
	})
	workspace := mustWorkspace(t)
	plan := completePlan(t, "secret-path", "secret-content")
	startResult := engine.NewStartDryRunResult(plan)
	original, err := engine.NewOutcome(engine.OutcomeInput{
		Command: engine.CommandStart, Workspace: &workspace, Result: &startResult,
		Diagnostics: []engine.Diagnostic{diagnostic}, ExitCode: engine.ExitSuccess,
	})
	if err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	redacted := original.Redacted(replace)
	redactedBytes, err := json.Marshal(redacted)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(redactedBytes), "secret") {
		t.Fatalf("redacted outcome leaked: %s", redactedBytes)
	}
	after, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("redaction mutated the authoritative outcome:\nbefore %s\nafter  %s", before, after)
	}
}

// TestConstructorsIsolateInputSlices proves built plans do not share slice
// backing arrays with caller inputs: replacing caller elements must not
// affect the plan, and successive accessors must observe stable contents.
func TestConstructorsIsolateInputSlices(t *testing.T) {
	t.Parallel()

	content := "workId: work-1\n"
	operation, err := engine.NewPlanOperation(engine.PlanOperationInput{
		Kind: engine.OperationCreateFile, Path: "file", Content: &content,
		ContentDigest: protocolTestDigest, ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	precondition, err := engine.NewPlanPrecondition(engine.PlanPreconditionInput{
		Kind: engine.PreconditionAbsent, Path: "file",
	})
	if err != nil {
		t.Fatal(err)
	}
	operations := make([]engine.PlanOperation, 0, 1)
	operations = append(operations, operation)
	preconditions := []engine.PlanPrecondition{precondition}
	plan, err := engine.NewMutationPlan(engine.MutationPlanInput{
		PlanID: "plan-1", WorkspaceID: "ws-1", WorkID: "work-1", RevisionScope: engine.RevisionWorkspace,
		InputDigest: protocolTestDigest, PolicyDigest: protocolTestDigest,
		WorkflowDefinitionDigest: protocolTestDigest, ValidatorDefinitionDigest: protocolTestDigest,
		Operations: operations, Preconditions: preconditions, Digest: protocolTestDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	operations[0] = engine.PlanOperation{}
	preconditions[0] = engine.PlanPrecondition{}
	content = "mutated"

	result := engine.NewStartDryRunResult(plan)
	if got := plan.Operations(); len(got) != 1 || got[0].Path() != "file" || *got[0].Content() != "workId: work-1\n" {
		t.Fatalf("plan aliases caller operation storage: %#v", got)
	}
	if got := plan.Preconditions(); len(got) != 1 || got[0].Path() != "file" {
		t.Fatalf("plan aliases caller precondition storage: %#v", got)
	}
	startValue, ok := result.StartDryRun()
	if !ok {
		t.Fatal("start-dry-run result lost its plan")
	}
	if got := startValue.Plan().Operations(); len(got) != 1 || *got[0].Content() != "workId: work-1\n" {
		t.Fatalf("result aliases plan operation storage: %#v", got)
	}
}
