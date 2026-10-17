package workflow

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

const testDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestCompileAndEvaluateWalkingSlice(t *testing.T) {
	t.Parallel()
	definition := validDefinition()
	graph := mustCompile(t, definition)

	input := validEvaluationInput(graph)
	first := Evaluate(context.Background(), graph, input, EvaluateOptions{IncludeAlternatives: true})
	if first.Status != EvaluationActionable {
		t.Fatalf("status = %q, want %q; diagnostics = %#v", first.Status, EvaluationActionable, first.Diagnostics)
	}
	if first.PrimaryAction == nil || first.PrimaryAction.NodeID != "draft" || first.PrimaryAction.ActionID != "submit-draft" {
		t.Fatalf("primary action = %#v, want draft submission", first.PrimaryAction)
	}
	if len(first.Alternatives) != 0 {
		t.Fatalf("alternatives = %#v, want none", first.Alternatives)
	}
	assertNodeState(t, first.Nodes, "draft", StateReady)
	assertNodeState(t, first.Nodes, "review", StateBlocked)

	input.NodeResults = []NodeResultFact{{
		NodeID:      "draft",
		NodeVersion: "1",
		State:       StateCompleted,
		AttemptID:   "attempt-1",
		Outputs: []OutputFact{{
			PortName:       "spec",
			TypeRef:        "artifact/spec@1",
			ValueRef:       "artifact-1",
			SemanticDigest: testDigest,
			Validity:       ValidityCurrent,
		}},
	}}
	second := Evaluate(context.Background(), graph, input, EvaluateOptions{})
	if second.PrimaryAction == nil || second.PrimaryAction.NodeID != "review" {
		t.Fatalf("primary action = %#v, want review", second.PrimaryAction)
	}
	assertNodeState(t, second.Nodes, "draft", StateCompleted)
	assertNodeState(t, second.Nodes, "review", StateReady)

	input.NodeResults = append(input.NodeResults, NodeResultFact{
		NodeID:      "review",
		NodeVersion: "1",
		State:       StateCompleted,
		AttemptID:   "attempt-2",
		Outputs:     []OutputFact{},
	})
	complete := Evaluate(context.Background(), graph, input, EvaluateOptions{})
	if complete.Status != EvaluationGraphComplete || complete.CompletionHandoff == nil {
		t.Fatalf("complete evaluation = %#v", complete)
	}
	if complete.PrimaryAction != nil {
		t.Fatalf("graph-complete evaluation invented action %#v", complete.PrimaryAction)
	}
}

func TestCompileIdentitiesSeparateExactSemanticAndBehaviorChanges(t *testing.T) {
	t.Parallel()
	firstDefinition := validDefinition()
	firstSource := sourceFor(t, firstDefinition)
	first, diagnostics := Compile(context.Background(), firstSource, CompileLimits{})
	if len(diagnostics) != 0 {
		t.Fatalf("first diagnostics = %#v", diagnostics)
	}

	cosmetic := validDefinition()
	cosmetic.Description = "presentation changed"
	cosmetic.Nodes[0].Description = "still inert"
	cosmetic.Nodes[0].Source.Path = "elsewhere.json"
	slices.Reverse(cosmetic.Nodes)
	secondSource := sourceFor(t, cosmetic)
	second, diagnostics := Compile(context.Background(), secondSource, CompileLimits{})
	if len(diagnostics) != 0 {
		t.Fatalf("second diagnostics = %#v", diagnostics)
	}
	if first.Ref().Definition.ExactDigest == second.Ref().Definition.ExactDigest {
		t.Fatal("exact digest did not change for different bytes")
	}
	if first.Ref().Definition.SemanticDigest != second.Ref().Definition.SemanticDigest {
		t.Fatalf("semantic digest changed for presentation/reorder: %q != %q", first.Ref().Definition.SemanticDigest, second.Ref().Definition.SemanticDigest)
	}
	if first.Ref().CompiledDigest != second.Ref().CompiledDigest {
		t.Fatalf("compiled digest changed for equivalent definition: %q != %q", first.Ref().CompiledDigest, second.Ref().CompiledDigest)
	}

	behavior := validDefinition()
	value := true
	behavior.Nodes[0].SkipPredicate = &Expression{Kind: ExpressionConst, Const: &value}
	third := mustCompile(t, behavior)
	if first.Ref().Definition.SemanticDigest == third.Ref().Definition.SemanticDigest {
		t.Fatal("semantic digest did not change for predicate change")
	}
}

func TestCompileRejectsClosedSchemaAndStructuralErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Definition)
		code string
	}{
		{
			name: "missing cardinality",
			edit: func(definition *Definition) { definition.Nodes[1].Inputs[0].Cardinality = "" },
			code: "WORKFLOW_PORT_INVALID",
		},
		{
			name: "type mismatch",
			edit: func(definition *Definition) { definition.Nodes[1].Inputs[0].TypeRef = "artifact/other@1" },
			code: "WORKFLOW_TYPE_MISMATCH",
		},
		{
			name: "duplicate node",
			edit: func(definition *Definition) { definition.Nodes[1].ID = definition.Nodes[0].ID },
			code: "WORKFLOW_DUPLICATE_ID",
		},
		{
			name: "cycle",
			edit: func(definition *Definition) {
				definition.Nodes[0].Inputs = []PortDefinition{{Name: "accepted", Direction: PortInput, TypeRef: "artifact/accepted@1", Cardinality: CardinalityOne}}
				definition.Nodes[1].Outputs = []PortDefinition{{Name: "accepted", Direction: PortOutput, TypeRef: "artifact/accepted@1", Cardinality: CardinalityOne}}
				definition.Edges = append(definition.Edges, EdgeDefinition{ID: "review-to-draft", SourceRef: PortReference{NodeID: "review", PortName: "accepted"}, TargetRef: PortReference{NodeID: "draft", PortName: "accepted"}, Relation: EdgeDependency})
			},
			code: "WORKFLOW_DEPENDENCY_CYCLE",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			definition := validDefinition()
			test.edit(&definition)
			graph, diagnostics := Compile(context.Background(), sourceFor(t, definition), CompileLimits{})
			if graph.valid() {
				t.Fatal("invalid definition produced a runnable graph")
			}
			assertDiagnostic(t, diagnostics, test.code)
		})
	}
}

func TestCompileRejectsUnknownJSONAndDigestMismatch(t *testing.T) {
	t.Parallel()
	source := sourceFor(t, validDefinition())
	source.Bytes = append(source.Bytes[:len(source.Bytes)-1], []byte(`,"command":"rm"}`)...)
	_, diagnostics := Compile(context.Background(), source, CompileLimits{})
	assertDiagnostic(t, diagnostics, "WORKFLOW_DEFINITION_INVALID")

	source = sourceFor(t, validDefinition())
	source.ExactDigest = strings.Repeat("b", 64)
	_, diagnostics = Compile(context.Background(), source, CompileLimits{})
	assertDiagnostic(t, diagnostics, "WORKFLOW_DEFINITION_DIGEST_MISMATCH")
}

func TestCompileHonorsCancellationAndLimits(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	graph, diagnostics := Compile(ctx, sourceFor(t, validDefinition()), CompileLimits{})
	if graph.valid() {
		t.Fatal("cancelled compilation returned graph")
	}
	assertDiagnostic(t, diagnostics, "WORKFLOW_CANCELLED")

	limits := CompileLimits{Nodes: 1}
	graph, diagnostics = Compile(context.Background(), sourceFor(t, validDefinition()), limits)
	if graph.valid() {
		t.Fatal("over-limit compilation returned graph")
	}
	assertDiagnostic(t, diagnostics, "WORKFLOW_COMPILE_LIMIT")

	limits = CompileLimits{Nodes: AcceptedCompileLimits().Nodes + 1}
	_, diagnostics = Compile(context.Background(), sourceFor(t, validDefinition()), limits)
	assertDiagnostic(t, diagnostics, "WORKFLOW_COMPILE_LIMIT_INVALID")
}

func TestPredicatesUseStrongKleeneLogic(t *testing.T) {
	t.Parallel()
	definition := validDefinition()
	definition.FactTypes = []FactType{
		{Key: "enabled", Kind: FactBool, Owner: "policy"},
		{Key: "risk", Kind: FactEnum, EnumValues: []string{"critical", "quick"}, Owner: "policy"},
	}
	snapshot, diagnostics := buildFactSnapshot(definition, nil, nil)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	missing := Expression{Kind: ExpressionBoolFact, FactKey: "enabled"}
	falseValue := false
	falseExpr := Expression{Kind: ExpressionConst, Const: &falseValue}
	trueValue := true
	trueExpr := Expression{Kind: ExpressionConst, Const: &trueValue}

	all := Expression{Kind: ExpressionAll, Operands: []Expression{missing, falseExpr}}
	if result := evaluatePredicate(&all, snapshot); result.Value != TruthFalse {
		t.Fatalf("unknown AND false = %q, want false", result.Value)
	}
	any := Expression{Kind: ExpressionAny, Operands: []Expression{missing, trueExpr}}
	if result := evaluatePredicate(&any, snapshot); result.Value != TruthTrue {
		t.Fatalf("unknown OR true = %q, want true", result.Value)
	}
	not := Expression{Kind: ExpressionNot, Operand: &missing}
	if result := evaluatePredicate(&not, snapshot); result.Value != TruthUnknown || len(result.Unknowns) != 1 {
		t.Fatalf("NOT unknown = %#v", result)
	}
}

func TestUnknownSkipBlocksAndNeverSkips(t *testing.T) {
	t.Parallel()
	definition := validDefinition()
	definition.FactTypes = []FactType{{Key: "may-skip", Kind: FactBool, Owner: "policy"}}
	definition.Nodes[0].SkipPredicate = &Expression{Kind: ExpressionBoolFact, FactKey: "may-skip"}
	graph := mustCompile(t, definition)
	evaluation := Evaluate(context.Background(), graph, validEvaluationInput(graph), EvaluateOptions{})
	assertNodeState(t, evaluation.Nodes, "draft", StateBlocked)
	if evaluation.PrimaryAction != nil && evaluation.PrimaryAction.NodeID == "draft" {
		t.Fatalf("unknown skip proposed draft action: %#v", evaluation.PrimaryAction)
	}
	if len(evaluation.Blockers) == 0 || evaluation.Blockers[0].Code != "WORKFLOW_FACT_UNKNOWN" {
		t.Fatalf("blockers = %#v", evaluation.Blockers)
	}
}

func TestExplicitSkipRequiresScopedPolicyAuthority(t *testing.T) {
	t.Parallel()
	definition := validDefinition()
	trueValue := true
	definition.Nodes[0].SkipPredicate = &Expression{Kind: ExpressionConst, Const: &trueValue}
	graph := mustCompile(t, definition)
	input := validEvaluationInput(graph)

	unauthorized := Evaluate(context.Background(), graph, input, EvaluateOptions{})
	assertNodeState(t, unauthorized.Nodes, "draft", StateBlocked)
	if len(unauthorized.Blockers) == 0 || unauthorized.Blockers[0].Code != "WORKFLOW_SKIP_UNAUTHORIZED" {
		t.Fatalf("unauthorized blockers = %#v", unauthorized.Blockers)
	}

	input.SkipAuthorizations = []SkipAuthorization{{NodeID: "draft", PolicyDigest: input.PolicyDigest, Authorized: true}}
	authorized := Evaluate(context.Background(), graph, input, EvaluateOptions{})
	assertNodeState(t, authorized.Nodes, "draft", StateSkipped)
}

func TestMandatoryGateCannotBeSkipped(t *testing.T) {
	t.Parallel()
	definition := validDefinition()
	trueValue := true
	definition.Nodes[0].Kind = NodeGate
	definition.Nodes[0].SkipPredicate = &Expression{Kind: ExpressionConst, Const: &trueValue}
	definition.MandatoryGateIDs = []string{"draft"}
	graph := mustCompile(t, definition)
	input := validEvaluationInput(graph)
	input.SkipAuthorizations = []SkipAuthorization{{NodeID: "draft", PolicyDigest: input.PolicyDigest, Authorized: true}}

	evaluation := Evaluate(context.Background(), graph, input, EvaluateOptions{})
	assertNodeState(t, evaluation.Nodes, "draft", StateBlocked)
	if len(evaluation.Blockers) == 0 || evaluation.Blockers[0].Code != "WORKFLOW_GATE_BLOCKED" {
		t.Fatalf("mandatory-gate blockers = %#v", evaluation.Blockers)
	}
}

func TestCompileRejectsIgnoredPredicateFieldsAndInvalidIntegerBounds(t *testing.T) {
	t.Parallel()
	definition := validDefinition()
	definition.FactTypes = []FactType{{Key: "enabled", Kind: FactBool, Owner: "policy"}}
	trueValue := true
	definition.Nodes[0].SkipPredicate = &Expression{
		Kind:    ExpressionBoolFact,
		FactKey: "enabled",
		Operand: &Expression{Kind: ExpressionConst, Const: &trueValue},
	}
	_, diagnostics := Compile(context.Background(), sourceFor(t, definition), CompileLimits{})
	assertDiagnostic(t, diagnostics, "WORKFLOW_PREDICATE_INVALID")

	definition = validDefinition()
	definition.FactTypes = []FactType{{Key: "risk", Kind: FactInteger, Owner: "policy"}}
	value := int64(3)
	definition.Nodes[0].ApplicabilityPredicate = &Expression{Kind: ExpressionCompare, FactKey: "risk", Operator: CompareGreater, Value: &Scalar{Integer: &value}}
	_, diagnostics = Compile(context.Background(), sourceFor(t, definition), CompileLimits{})
	assertDiagnostic(t, diagnostics, "WORKFLOW_FACT_TYPE_INVALID")
}

func TestFailedOrStalePredecessorCannotSatisfyDependency(t *testing.T) {
	t.Parallel()
	graph := mustCompile(t, validDefinition())
	for _, state := range []NodeState{StateFailed, StateCancelled, StateStale} {
		state := state
		t.Run(string(state), func(t *testing.T) {
			t.Parallel()
			input := validEvaluationInput(graph)
			result := NodeResultFact{NodeID: "draft", NodeVersion: "1", State: state, AttemptID: "attempt-1", Outputs: []OutputFact{}}
			if state == StateFailed {
				result.FailureDisposition = FailureRetryEligible
			}
			input.NodeResults = []NodeResultFact{result}
			evaluation := Evaluate(context.Background(), graph, input, EvaluateOptions{})
			assertNodeState(t, evaluation.Nodes, "review", StateBlocked)
			if evaluation.PrimaryAction != nil && evaluation.PrimaryAction.NodeID == "review" {
				t.Fatalf("%s predecessor authorized review", state)
			}
		})
	}
}

func TestEvaluateIsDeterministicAcrossInputOrder(t *testing.T) {
	t.Parallel()
	definition := validDefinition()
	definition.FactTypes = []FactType{
		{Key: "alpha", Kind: FactBool, Owner: "policy"},
		{Key: "beta", Kind: FactBool, Owner: "policy"},
	}
	trueValue := true
	definition.Nodes[0].ApplicabilityPredicate = &Expression{
		Kind: ExpressionAll,
		Operands: []Expression{
			{Kind: ExpressionBoolFact, FactKey: "alpha"},
			{Kind: ExpressionBoolFact, FactKey: "beta"},
		},
	}
	graph := mustCompile(t, definition)
	input := validEvaluationInput(graph)
	input.Facts = []FactRecord{
		{Key: "beta", Owner: "policy", SubjectRef: "work-1", SemanticDigest: testDigest, Validity: ValidityCurrent, Present: true, Value: Scalar{Bool: &trueValue}},
		{Key: "alpha", Owner: "policy", SubjectRef: "work-1", SemanticDigest: testDigest, Validity: ValidityCurrent, Present: true, Value: Scalar{Bool: &trueValue}},
	}
	first := Evaluate(context.Background(), graph, input, EvaluateOptions{IncludeAlternatives: true})
	if first.PrimaryAction == nil || len(first.PrimaryAction.InputRefs) != 2 {
		t.Fatalf("proposal input refs = %#v, want both predicate facts", first.PrimaryAction)
	}
	if first.PrimaryAction.SnapshotID != input.Snapshot.SnapshotID || first.PrimaryAction.ReadSetDigest != input.Snapshot.ReadSetDigest {
		t.Fatalf("proposal is not bound to exact snapshot: %#v", first.PrimaryAction)
	}
	slices.Reverse(input.Facts)
	second := Evaluate(context.Background(), graph, input, EvaluateOptions{IncludeAlternatives: true})
	if first.StateDigest != second.StateDigest {
		first.StateDigest = ""
		second.StateDigest = ""
		firstDebug, _ := json.Marshal(first)
		secondDebug, _ := json.Marshal(second)
		t.Fatalf("state digest depends on fact order:\n%s\n%s", firstDebug, secondDebug)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("evaluation bytes differ:\n%s\n%s", firstJSON, secondJSON)
	}
}

func validDefinition() Definition {
	return Definition{
		SchemaVersion:   definitionSchemaVersion,
		DefinitionID:    "walking",
		SemanticVersion: "1.0.0",
		Canonicalizer: CanonicalizerRef{
			ID:               CanonicalizerID,
			Version:          CanonicalizerVersion,
			DefinitionDigest: testDigest,
		},
		FactTypes: []FactType{},
		Nodes: []NodeDefinition{
			{
				ID:              "draft",
				Version:         "1",
				Kind:            NodeAuthored,
				Inputs:          []PortDefinition{},
				Outputs:         []PortDefinition{{Name: "spec", Direction: PortOutput, TypeRef: "artifact/spec@1", Cardinality: CardinalityOne}},
				RequiredGateIDs: []string{},
				ProfileBinding:  "standard",
				Action:          ActionDefinition{ID: "submit-draft", Kind: ActionSubmit, SubmissionTypeRef: "artifact/spec@1"},
			},
			{
				ID:              "review",
				Version:         "1",
				Kind:            NodeAssessment,
				Inputs:          []PortDefinition{{Name: "spec", Direction: PortInput, TypeRef: "artifact/spec@1", Cardinality: CardinalityOne}},
				Outputs:         []PortDefinition{},
				RequiredGateIDs: []string{},
				ProfileBinding:  "standard",
				Action:          ActionDefinition{ID: "assess-draft", Kind: ActionAssess, SubmissionTypeRef: "assessment/review@1"},
			},
		},
		Edges: []EdgeDefinition{{
			ID:        "draft-to-review",
			SourceRef: PortReference{NodeID: "draft", PortName: "spec"},
			TargetRef: PortReference{NodeID: "review", PortName: "spec"},
			Relation:  EdgeDependency,
		}},
		MandatoryGateIDs:   []string{},
		CompatibilityRules: []CompatibilityRule{},
	}
}

func sourceFor(t *testing.T, definition Definition) DefinitionSource {
	t.Helper()
	content, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	return DefinitionSource{
		MediaType:  DefinitionMediaType,
		SchemaID:   DefinitionSchemaID,
		SourceID:   "test-source",
		Version:    "1",
		Coordinate: SourceCoordinate{Path: "workflow.json", StartLine: 1, StartColumn: 1, Unit: "utf8-byte"},
		Bytes:      content,
	}
}

func mustCompile(t *testing.T, definition Definition) CompiledGraph {
	t.Helper()
	graph, diagnostics := Compile(context.Background(), sourceFor(t, definition), CompileLimits{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	return graph
}

func validEvaluationInput(graph CompiledGraph) EvaluationInput {
	return EvaluationInput{
		Snapshot: SnapshotRef{
			SnapshotID:        "snapshot-1",
			WorkspaceID:       "workspace-1",
			WorkspaceRevision: 3,
			WorkID:            "work-1",
			WorkRevision:      5,
			ReadSetDigest:     testDigest,
		},
		PinnedCompiledDigest: graph.Ref().CompiledDigest,
		PolicyDigest:         testDigest,
		CompositionID:        "composition-1",
		CompositionRevision:  2,
		Facts:                []FactRecord{},
		OwnerObservations:    []OwnerObservation{},
		NodeResults:          []NodeResultFact{},
	}
}

func assertNodeState(t *testing.T, values []NodeEvaluation, nodeID string, expected NodeState) {
	t.Helper()
	for _, value := range values {
		if value.NodeID == nodeID {
			if value.State != expected {
				t.Fatalf("node %s state = %q, want %q; blockers = %#v", nodeID, value.State, expected, value.Blockers)
			}
			return
		}
	}
	t.Fatalf("node %s not found in %#v", nodeID, values)
}

func assertDiagnostic(t *testing.T, values []Diagnostic, code string) {
	t.Helper()
	for _, value := range values {
		if value.Code == code {
			return
		}
	}
	t.Fatalf("diagnostic %s not found in %#v", code, values)
}
