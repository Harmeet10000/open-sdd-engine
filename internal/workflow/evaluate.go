package workflow

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"sdd.dev/sdd/internal/protocol/canonicaljson"
)

type NodeState string

const (
	StateUnstarted NodeState = "unstarted"
	StateReady     NodeState = "ready"
	StateRunning   NodeState = "running"
	StateBlocked   NodeState = "blocked"
	StateStale     NodeState = "stale"
	StateFailed    NodeState = "failed"
	StateCancelled NodeState = "cancelled"
	StateSkipped   NodeState = "skipped"
	StateCompleted NodeState = "completed"
)

type FailureDisposition string

const (
	FailureRetryEligible FailureDisposition = "retry-eligible"
	FailureExhausted     FailureDisposition = "exhausted"
)

type OutputFact struct {
	PortName       string       `json:"portName"`
	TypeRef        string       `json:"typeRef"`
	ValueRef       string       `json:"valueRef"`
	SemanticDigest string       `json:"semanticDigest"`
	ExpectedDigest string       `json:"expectedDigest,omitempty"`
	Validity       FactValidity `json:"validity"`
}

// NodeResultFact is immutable history supplied by the workspace/execution
// owners. Ready and blocked are derived and therefore are not accepted here.
type NodeResultFact struct {
	NodeID             string             `json:"nodeId"`
	NodeVersion        string             `json:"nodeVersion"`
	State              NodeState          `json:"state"`
	FailureDisposition FailureDisposition `json:"failureDisposition,omitempty"`
	AttemptID          string             `json:"attemptId,omitempty"`
	CandidateDigest    string             `json:"candidateDigest,omitempty"`
	Outputs            []OutputFact       `json:"outputs"`
}

type SnapshotRef struct {
	SnapshotID        string `json:"snapshotId"`
	WorkspaceID       string `json:"workspaceId"`
	WorkspaceRevision uint64 `json:"workspaceRevision"`
	WorkID            string `json:"workId"`
	WorkRevision      uint64 `json:"workRevision"`
	ReadSetDigest     string `json:"readSetDigest"`
}

type EvaluationInput struct {
	Snapshot             SnapshotRef         `json:"snapshot"`
	PinnedCompiledDigest string              `json:"pinnedCompiledDigest"`
	PolicyDigest         string              `json:"policyDigest"`
	CompositionID        string              `json:"compositionId"`
	CompositionRevision  uint64              `json:"compositionRevision"`
	EvaluationInstant    EvaluationInstant   `json:"evaluationInstant"`
	Facts                []FactRecord        `json:"facts"`
	OwnerObservations    []OwnerObservation  `json:"ownerObservations"`
	NodeResults          []NodeResultFact    `json:"nodeResults"`
	SkipAuthorizations   []SkipAuthorization `json:"skipAuthorizations"`
}

// SkipAuthorization is an opaque policy fact for one explicit skip. A true
// predicate scopes the skip; it does not manufacture policy authority.
type SkipAuthorization struct {
	NodeID       string `json:"nodeId"`
	PolicyDigest string `json:"policyDigest"`
	Authorized   bool   `json:"authorized"`
}

type EvaluateOptions struct {
	IncludeAlternatives bool
}

type Blocker struct {
	Code           string           `json:"code"`
	NodeID         string           `json:"nodeId"`
	SubjectID      string           `json:"subjectId,omitempty"`
	Reason         string           `json:"reason"`
	ExpectedDigest string           `json:"expectedDigest,omitempty"`
	CurrentDigest  string           `json:"currentDigest,omitempty"`
	Remediation    string           `json:"remediation"`
	Source         SourceCoordinate `json:"source"`
}

type NodeEvaluation struct {
	NodeID             string             `json:"nodeId"`
	NodeVersion        string             `json:"nodeVersion"`
	State              NodeState          `json:"state"`
	FailureDisposition FailureDisposition `json:"failureDisposition,omitempty"`
	Claimed            bool               `json:"claimed"`
	Blockers           []Blocker          `json:"blockers"`
}

type ActionProposal struct {
	ActionID                  string                `json:"actionId"`
	ActionRevision            string                `json:"actionRevision"`
	NodeID                    string                `json:"nodeId"`
	NodeVersion               string                `json:"nodeVersion"`
	Kind                      ActionKind            `json:"kind"`
	WorkID                    string                `json:"workId"`
	SnapshotID                string                `json:"snapshotId"`
	ReadSetDigest             string                `json:"readSetDigest"`
	ExpectedWorkspaceRevision uint64                `json:"expectedWorkspaceRevision"`
	ExpectedWorkRevision      uint64                `json:"expectedWorkRevision"`
	DefinitionRef             CompiledDefinitionRef `json:"definitionRef"`
	CompositionID             string                `json:"compositionId"`
	CompositionRevision       uint64                `json:"compositionRevision"`
	PolicyDigest              string                `json:"policyDigest"`
	SubmissionTypeRef         string                `json:"submissionTypeRef,omitempty"`
	ValidatorBinding          string                `json:"validatorBinding,omitempty"`
	ExecutionBinding          string                `json:"executionBinding,omitempty"`
	InputRefs                 []ActionInputRef      `json:"inputRefs"`
	ReasonCode                string                `json:"reasonCode"`
	Source                    SourceCoordinate      `json:"source"`
}

type ActionInputRef struct {
	Kind            string `json:"kind"`
	ID              string `json:"id"`
	SubjectRef      string `json:"subjectRef"`
	SubjectRevision uint64 `json:"subjectRevision,omitempty"`
	Digest          string `json:"digest,omitempty"`
}

type CompletionHandoff struct {
	WorkID              string                `json:"workId"`
	WorkRevision        uint64                `json:"workRevision"`
	DefinitionRef       CompiledDefinitionRef `json:"definitionRef"`
	CompositionID       string                `json:"compositionId"`
	CompositionRevision uint64                `json:"compositionRevision"`
	PolicyDigest        string                `json:"policyDigest"`
}

type EvaluationStatus string

const (
	EvaluationActionable    EvaluationStatus = "actionable"
	EvaluationBlocked       EvaluationStatus = "blocked"
	EvaluationGraphComplete EvaluationStatus = "graph-complete"
)

type Evaluation struct {
	Status            EvaluationStatus      `json:"status"`
	DefinitionRef     CompiledDefinitionRef `json:"definitionRef"`
	Snapshot          SnapshotRef           `json:"snapshot"`
	StateDigest       string                `json:"stateDigest"`
	Nodes             []NodeEvaluation      `json:"nodes"`
	PrimaryAction     *ActionProposal       `json:"primaryAction,omitempty"`
	Alternatives      []ActionProposal      `json:"alternatives"`
	Blockers          []Blocker             `json:"blockers"`
	CompletionHandoff *CompletionHandoff    `json:"completionHandoff,omitempty"`
	Diagnostics       []Diagnostic          `json:"diagnostics"`
}

// Evaluate deterministically derives graph state and mutation-neutral action
// proposals from one immutable input. It performs no I/O.
func Evaluate(ctx context.Context, graph CompiledGraph, input EvaluationInput, options EvaluateOptions) Evaluation {
	base := Evaluation{
		DefinitionRef: graph.ref,
		Snapshot:      input.Snapshot,
		Nodes:         []NodeEvaluation{},
		Alternatives:  []ActionProposal{},
		Blockers:      []Blocker{},
		Diagnostics:   []Diagnostic{},
	}
	if ctx == nil {
		base.Status = EvaluationBlocked
		base.Diagnostics = append(base.Diagnostics, diagnostic("WORKFLOW_CONTEXT_INVALID", "evaluation requires a non-nil context", SourceCoordinate{}, "Pass the request context to workflow.Evaluate."))
		return finishEvaluation(base)
	}
	if err := ctx.Err(); err != nil {
		base.Status = EvaluationBlocked
		base.Diagnostics = append(base.Diagnostics, cancelledDiagnostic(SourceCoordinate{}, err))
		return finishEvaluation(base)
	}
	if !graph.valid() {
		base.Status = EvaluationBlocked
		base.Diagnostics = append(base.Diagnostics, diagnostic("WORKFLOW_DEFINITION_INVALID", "evaluation requires a successfully compiled graph", SourceCoordinate{}, "Compile and pin a valid workflow definition."))
		return finishEvaluation(base)
	}
	if input.PinnedCompiledDigest == "" || input.PinnedCompiledDigest != graph.ref.CompiledDigest {
		base.Status = EvaluationBlocked
		base.Diagnostics = append(base.Diagnostics, diagnostic(
			"WORKFLOW_DEFINITION_MISSING",
			"pinned compiled definition does not match the supplied graph",
			SourceCoordinate{},
			"Load the exact pinned compiled definition or explicitly migrate the work.",
			MetadataEntry{Key: "actual", Value: graph.ref.CompiledDigest},
			MetadataEntry{Key: "expected", Value: input.PinnedCompiledDigest},
		))
		return finishEvaluation(base)
	}
	if input.Snapshot.SnapshotID == "" || input.Snapshot.WorkspaceID == "" || input.Snapshot.WorkID == "" || input.Snapshot.ReadSetDigest == "" {
		base.Status = EvaluationBlocked
		base.Diagnostics = append(base.Diagnostics, diagnostic("WORKFLOW_SNAPSHOT_INVALID", "evaluation snapshot coordinates are incomplete", SourceCoordinate{}, "Capture one coherent workspace snapshot before evaluation."))
		return finishEvaluation(base)
	}
	if uint64(len(input.Facts)) > graphLimits(graph).RuntimeFacts {
		base.Status = EvaluationBlocked
		base.Diagnostics = append(base.Diagnostics, diagnostic(
			"WORKFLOW_EVALUATION_LIMIT",
			"runtime fact count exceeds the compiled limit",
			SourceCoordinate{},
			"Reduce the fact snapshot to facts declared by this workflow.",
			MetadataEntry{Key: "actual", Value: fmt.Sprint(len(input.Facts))},
			MetadataEntry{Key: "maximum", Value: fmt.Sprint(graphLimits(graph).RuntimeFacts)},
		))
		return finishEvaluation(base)
	}

	facts, factDiagnostics := buildFactSnapshot(graph.definition, input.Facts, input.OwnerObservations)
	if len(factDiagnostics) != 0 {
		base.Status = EvaluationBlocked
		base.Diagnostics = append(base.Diagnostics, factDiagnostics...)
		return finishEvaluation(base)
	}
	resultFacts, resultDiagnostics := normalizeNodeResults(graph, input.NodeResults)
	if len(resultDiagnostics) != 0 {
		base.Status = EvaluationBlocked
		base.Diagnostics = append(base.Diagnostics, resultDiagnostics...)
		return finishEvaluation(base)
	}

	states := make([]NodeEvaluation, len(graph.nodes))
	proposals := make([]ActionProposal, 0)
	steps := uint64(0)
	for _, nodeIndex := range graph.topology {
		if err := ctx.Err(); err != nil {
			base.Diagnostics = append(base.Diagnostics, cancelledDiagnostic(graph.nodes[nodeIndex].definition.Source, err))
			break
		}
		steps++
		if steps > graphLimits(graph).EvaluationSteps {
			base.Diagnostics = append(base.Diagnostics, diagnostic("WORKFLOW_EVALUATION_LIMIT", "evaluation exceeded its deterministic step budget", graph.nodes[nodeIndex].definition.Source, "Reduce workflow or fact complexity."))
			break
		}
		nodeState, proposal, usedSteps := evaluateNode(graph, nodeIndex, states, resultFacts, facts, input)
		steps += usedSteps
		if steps > graphLimits(graph).EvaluationSteps {
			base.Diagnostics = append(base.Diagnostics, diagnostic("WORKFLOW_EVALUATION_LIMIT", "evaluation exceeded its deterministic step budget", graph.nodes[nodeIndex].definition.Source, "Reduce workflow or fact complexity."))
			break
		}
		states[nodeIndex] = nodeState
		if proposal != nil {
			proposals = append(proposals, *proposal)
		}
	}
	if len(base.Diagnostics) != 0 {
		base.Status = EvaluationBlocked
		base.Nodes = stableNodeEvaluations(states)
		return finishEvaluation(base)
	}

	base.Nodes = stableNodeEvaluations(states)
	for _, state := range base.Nodes {
		base.Blockers = append(base.Blockers, state.Blockers...)
	}
	sortBlockers(base.Blockers)
	sort.Slice(proposals, func(i, j int) bool { return proposalLess(proposals[i], proposals[j]) })
	if len(proposals) != 0 {
		base.Status = EvaluationActionable
		primary := proposals[0]
		base.PrimaryAction = &primary
		if options.IncludeAlternatives {
			base.Alternatives = slices.Clone(proposals[1:])
		}
		return finishEvaluation(base)
	}
	if graphIsComplete(base.Nodes) {
		base.Status = EvaluationGraphComplete
		base.CompletionHandoff = &CompletionHandoff{
			WorkID:              input.Snapshot.WorkID,
			WorkRevision:        input.Snapshot.WorkRevision,
			DefinitionRef:       graph.ref,
			CompositionID:       input.CompositionID,
			CompositionRevision: input.CompositionRevision,
			PolicyDigest:        input.PolicyDigest,
		}
		return finishEvaluation(base)
	}
	base.Status = EvaluationBlocked
	return finishEvaluation(base)
}

func evaluateNode(
	graph CompiledGraph,
	nodeIndex int,
	states []NodeEvaluation,
	results map[string]NodeResultFact,
	facts factSnapshot,
	input EvaluationInput,
) (NodeEvaluation, *ActionProposal, uint64) {
	node := graph.nodes[nodeIndex].definition
	evaluation := NodeEvaluation{NodeID: node.ID, NodeVersion: node.Version, State: StateUnstarted, Blockers: []Blocker{}}
	steps := uint64(1)
	if result, exists := results[node.ID]; exists {
		evaluation.State = result.State
		evaluation.FailureDisposition = result.FailureDisposition
		if result.State == StateCompleted && outputsStale(result.Outputs) {
			evaluation.State = StateStale
			evaluation.Blockers = append(evaluation.Blockers, Blocker{Code: "WORKFLOW_OUTPUT_STALE", NodeID: node.ID, SubjectID: node.ID, Reason: "a completed output is no longer current", Remediation: "Regenerate or revalidate the stale output.", Source: node.Source})
		}
		return evaluation, nil, steps
	}

	applicable := evaluatePredicate(node.ApplicabilityPredicate, facts)
	steps += uint64(len(applicable.Unknowns) + 1)
	if applicable.Value == TruthUnknown {
		evaluation.State = StateBlocked
		evaluation.Blockers = append(evaluation.Blockers, predicateBlockers(node.ID, applicable.Unknowns)...)
		return evaluation, nil, steps
	}
	if applicable.Value == TruthFalse {
		if graph.mandatoryGate(node.ID) {
			evaluation.State = StateBlocked
			evaluation.Blockers = append(evaluation.Blockers, Blocker{Code: "WORKFLOW_GATE_BLOCKED", NodeID: node.ID, SubjectID: node.ID, Reason: "mandatory gate is not applicable under current facts", Remediation: "Resolve the route or migrate the workflow; mandatory gates cannot disappear at runtime.", Source: node.Source})
			return evaluation, nil, steps
		}
		evaluation.State = StateSkipped
		return evaluation, nil, steps
	}
	skip := knownPredicate(TruthFalse)
	if node.SkipPredicate != nil {
		skip = evaluatePredicate(node.SkipPredicate, facts)
	}
	steps += uint64(len(skip.Unknowns) + 1)
	if skip.Value == TruthUnknown {
		evaluation.State = StateBlocked
		evaluation.Blockers = append(evaluation.Blockers, predicateBlockers(node.ID, skip.Unknowns)...)
		return evaluation, nil, steps
	}
	if skip.Value == TruthTrue {
		if graph.mandatoryGate(node.ID) {
			evaluation.State = StateBlocked
			evaluation.Blockers = append(evaluation.Blockers, Blocker{Code: "WORKFLOW_GATE_BLOCKED", NodeID: node.ID, SubjectID: node.ID, Reason: "mandatory gate cannot be skipped", Remediation: "Complete the mandatory gate with trusted evidence.", Source: node.Source})
			return evaluation, nil, steps
		}
		if !skipAuthorized(node.ID, input.PolicyDigest, input.SkipAuthorizations) {
			evaluation.State = StateBlocked
			evaluation.Blockers = append(evaluation.Blockers, Blocker{Code: "WORKFLOW_SKIP_UNAUTHORIZED", NodeID: node.ID, SubjectID: node.ID, Reason: "skip predicate is true but no matching policy authorization exists", Remediation: "Provide a current scoped skip authorization or complete the node.", Source: node.Source})
			return evaluation, nil, steps
		}
		evaluation.State = StateSkipped
		return evaluation, nil, steps
	}

	for _, gateID := range node.RequiredGateIDs {
		gateIndex := graph.nodeIndex(gateID)
		if gateIndex < 0 || states[gateIndex].State != StateCompleted {
			evaluation.Blockers = append(evaluation.Blockers, Blocker{Code: "WORKFLOW_GATE_BLOCKED", NodeID: node.ID, SubjectID: gateID, Reason: "required gate has not completed", Remediation: "Complete the required gate with trusted evidence.", Source: node.Source})
		}
	}

	inputCounts := make(map[string]int, len(node.Inputs))
	for _, edgeIndex := range graph.nodes[nodeIndex].incoming {
		edge := graph.edges[edgeIndex]
		predicate := evaluatePredicate(edge.definition.Predicate, facts)
		steps += uint64(len(predicate.Unknowns) + 1)
		if predicate.Value == TruthFalse {
			continue
		}
		if predicate.Value == TruthUnknown {
			evaluation.Blockers = append(evaluation.Blockers, predicateBlockers(node.ID, predicate.Unknowns)...)
			continue
		}
		sourceState := states[edge.sourceIndex]
		if sourceState.State != StateCompleted {
			code := "WORKFLOW_DEPENDENCY_BLOCKED"
			if sourceState.State == StateStale || edge.definition.Relation == EdgeFreshness {
				code = "WORKFLOW_FRESHNESS_STALE"
			}
			evaluation.Blockers = append(evaluation.Blockers, Blocker{Code: code, NodeID: node.ID, SubjectID: edge.definition.ID, Reason: "source node has not produced a current successful output", Remediation: "Complete or refresh the source node.", Source: edge.definition.Coordinate})
			continue
		}
		result := results[graph.nodes[edge.sourceIndex].definition.ID]
		outputs := matchingCurrentOutputs(result.Outputs, edge.definition.SourceRef.PortName)
		if len(outputs) == 0 {
			evaluation.Blockers = append(evaluation.Blockers, Blocker{Code: "WORKFLOW_OUTPUT_MISSING", NodeID: node.ID, SubjectID: edge.definition.ID, Reason: "completed source lacks the required current output", Remediation: "Regenerate or revalidate the source output.", Source: edge.definition.Coordinate})
			continue
		}
		inputCounts[edge.definition.TargetRef.PortName] += len(outputs)
	}
	for _, port := range node.Inputs {
		count := inputCounts[port.Name]
		if !cardinalityAccepts(port.Cardinality, count) {
			evaluation.Blockers = append(evaluation.Blockers, Blocker{
				Code:        "WORKFLOW_CARDINALITY_BLOCKED",
				NodeID:      node.ID,
				SubjectID:   port.Name,
				Reason:      fmt.Sprintf("input cardinality %s received %d current values", port.Cardinality, count),
				Remediation: "Produce the required number of current input values.",
				Source:      port.Source,
			})
		}
	}
	if len(evaluation.Blockers) != 0 {
		evaluation.State = StateBlocked
		sortBlockers(evaluation.Blockers)
		return evaluation, nil, steps
	}

	evaluation.State = StateReady
	proposal := newActionProposal(graph, nodeIndex, input)
	return evaluation, &proposal, steps
}

func normalizeNodeResults(graph CompiledGraph, values []NodeResultFact) (map[string]NodeResultFact, []Diagnostic) {
	result := make(map[string]NodeResultFact, len(values))
	var diagnostics []Diagnostic
	for _, value := range values {
		index := graph.nodeIndex(value.NodeID)
		if index < 0 {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_RESULT_UNKNOWN_NODE", "node result references an unknown node", SourceCoordinate{}, "Remove the fact or migrate the definition.", MetadataEntry{Key: "nodeId", Value: value.NodeID}))
			continue
		}
		node := graph.nodes[index].definition
		if value.NodeVersion != node.Version {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_RESULT_SUBJECT_MISMATCH", "node result version does not match the compiled node", node.Source, "Use a result for the exact pinned node version.", MetadataEntry{Key: "nodeId", Value: value.NodeID}))
			continue
		}
		if !acceptedResultState(value.State) {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_RESULT_STATE_INVALID", "node result contains a derived or unsupported state", node.Source, "Provide only immutable owner result states.", MetadataEntry{Key: "nodeId", Value: value.NodeID}, MetadataEntry{Key: "state", Value: string(value.State)}))
			continue
		}
		if value.State == StateFailed {
			if value.FailureDisposition != FailureRetryEligible && value.FailureDisposition != FailureExhausted {
				diagnostics = append(diagnostics, diagnostic("WORKFLOW_RESULT_STATE_INVALID", "failed node result lacks a valid failure disposition", node.Source, "Declare retry-eligible or exhausted for the failed attempt.", MetadataEntry{Key: "nodeId", Value: value.NodeID}))
				continue
			}
		} else if value.FailureDisposition != "" {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_RESULT_STATE_INVALID", "non-failed node result carries a failure disposition", node.Source, "Remove the failure disposition from non-failed results.", MetadataEntry{Key: "nodeId", Value: value.NodeID}))
			continue
		}
		if value.AttemptID == "" && value.State != StateStale {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_RESULT_SUBJECT_MISMATCH", "node result is not bound to an immutable attempt", node.Source, "Provide the owning workflow attempt identity.", MetadataEntry{Key: "nodeId", Value: value.NodeID}))
			continue
		}
		if previous, exists := result[value.NodeID]; exists {
			code := "WORKFLOW_RESULT_DUPLICATE"
			message := "node result is duplicated"
			if !nodeResultsEqual(previous, value) {
				code = "WORKFLOW_RESULT_CONTRADICTORY"
				message = "node results for the same node are contradictory"
			}
			diagnostics = append(diagnostics, diagnostic(code, message, node.Source, "Provide one current immutable result per node.", MetadataEntry{Key: "nodeId", Value: value.NodeID}))
			continue
		}
		copyValue := value
		copyValue.Outputs = slices.Clone(value.Outputs)
		sort.Slice(copyValue.Outputs, func(i, j int) bool {
			if copyValue.Outputs[i].PortName != copyValue.Outputs[j].PortName {
				return copyValue.Outputs[i].PortName < copyValue.Outputs[j].PortName
			}
			return copyValue.Outputs[i].ValueRef < copyValue.Outputs[j].ValueRef
		})
		outputCounts := make(map[string]int, len(node.Outputs))
		for _, output := range copyValue.Outputs {
			port, exists := outputPort(node, output.PortName)
			if !exists || port.TypeRef != output.TypeRef || output.ValueRef == "" || output.SemanticDigest == "" || !validFactValidity(output.Validity) {
				diagnostics = append(diagnostics, diagnostic("WORKFLOW_OUTPUT_INVALID", "node result contains an output outside the compiled port contract", node.Source, "Provide output facts matching the exact compiled node ports.", MetadataEntry{Key: "nodeId", Value: node.ID}, MetadataEntry{Key: "portName", Value: output.PortName}))
				continue
			}
			outputCounts[output.PortName]++
		}
		if copyValue.State == StateCompleted {
			for _, port := range node.Outputs {
				if !cardinalityAccepts(port.Cardinality, outputCounts[port.Name]) {
					diagnostics = append(diagnostics, diagnostic("WORKFLOW_OUTPUT_CARDINALITY_INVALID", "completed node output does not satisfy its declared cardinality", port.Source, "Provide the required completed output facts.", MetadataEntry{Key: "nodeId", Value: node.ID}, MetadataEntry{Key: "portName", Value: port.Name}))
				}
			}
		}
		result[value.NodeID] = copyValue
	}
	sortDiagnostics(diagnostics)
	return result, diagnostics
}

func acceptedResultState(state NodeState) bool {
	return state == StateRunning || state == StateFailed || state == StateCancelled || state == StateStale || state == StateCompleted
}

func nodeResultsEqual(left, right NodeResultFact) bool {
	left.Outputs = slices.Clone(left.Outputs)
	right.Outputs = slices.Clone(right.Outputs)
	sort.Slice(left.Outputs, func(i, j int) bool { return outputLess(left.Outputs[i], left.Outputs[j]) })
	sort.Slice(right.Outputs, func(i, j int) bool { return outputLess(right.Outputs[i], right.Outputs[j]) })
	return left.NodeID == right.NodeID && left.NodeVersion == right.NodeVersion && left.State == right.State && left.FailureDisposition == right.FailureDisposition && left.AttemptID == right.AttemptID && left.CandidateDigest == right.CandidateDigest && slices.Equal(left.Outputs, right.Outputs)
}

func outputLess(left, right OutputFact) bool {
	if left.PortName != right.PortName {
		return left.PortName < right.PortName
	}
	return left.ValueRef < right.ValueRef
}

func outputsStale(outputs []OutputFact) bool {
	for _, output := range outputs {
		if output.Validity != ValidityCurrent || (output.ExpectedDigest != "" && output.ExpectedDigest != output.SemanticDigest) {
			return true
		}
	}
	return false
}

func matchingCurrentOutputs(outputs []OutputFact, portName string) []OutputFact {
	result := make([]OutputFact, 0)
	for _, output := range outputs {
		if output.PortName == portName && output.Validity == ValidityCurrent && (output.ExpectedDigest == "" || output.ExpectedDigest == output.SemanticDigest) {
			result = append(result, output)
		}
	}
	return result
}

func cardinalityAccepts(cardinality Cardinality, count int) bool {
	switch cardinality {
	case CardinalityOne:
		return count == 1
	case CardinalityZeroOrOne:
		return count <= 1
	case CardinalityZeroOrMany:
		return true
	case CardinalityAtLeastOne:
		return count >= 1
	default:
		return false
	}
}

func predicateBlockers(nodeID string, unknowns []UnknownReason) []Blocker {
	result := make([]Blocker, len(unknowns))
	for index, unknown := range unknowns {
		result[index] = Blocker{
			Code:           "WORKFLOW_FACT_UNKNOWN",
			NodeID:         nodeID,
			SubjectID:      unknown.FactKey,
			Reason:         unknown.Reason,
			ExpectedDigest: unknown.ExpectedDigest,
			CurrentDigest:  unknown.CurrentDigest,
			Remediation:    "Provide a current trusted fact from " + unknown.Owner + ".",
			Source:         unknown.Source,
		}
	}
	return result
}

func newActionProposal(graph CompiledGraph, nodeIndex int, input EvaluationInput) ActionProposal {
	node := graph.nodes[nodeIndex].definition
	return ActionProposal{
		ActionID:                  node.Action.ID,
		ActionRevision:            graph.nodes[nodeIndex].actionRevision,
		NodeID:                    node.ID,
		NodeVersion:               node.Version,
		Kind:                      node.Action.Kind,
		WorkID:                    input.Snapshot.WorkID,
		SnapshotID:                input.Snapshot.SnapshotID,
		ReadSetDigest:             input.Snapshot.ReadSetDigest,
		ExpectedWorkspaceRevision: input.Snapshot.WorkspaceRevision,
		ExpectedWorkRevision:      input.Snapshot.WorkRevision,
		DefinitionRef:             graph.ref,
		CompositionID:             input.CompositionID,
		CompositionRevision:       input.CompositionRevision,
		PolicyDigest:              input.PolicyDigest,
		SubmissionTypeRef:         node.Action.SubmissionTypeRef,
		ValidatorBinding:          node.Action.ValidatorBinding,
		ExecutionBinding:          node.Action.ExecutionBinding,
		InputRefs:                 proposalInputRefs(graph, nodeIndex, input),
		ReasonCode:                "WORKFLOW_NODE_READY",
		Source:                    node.Action.Source,
	}
}

func proposalInputRefs(graph CompiledGraph, nodeIndex int, input EvaluationInput) []ActionInputRef {
	node := graph.nodes[nodeIndex].definition
	factKeys := make(map[string]struct{})
	collectExpressionFactKeys(node.ApplicabilityPredicate, factKeys)
	collectExpressionFactKeys(node.SkipPredicate, factKeys)
	for _, edgeIndex := range graph.nodes[nodeIndex].incoming {
		collectExpressionFactKeys(graph.edges[edgeIndex].definition.Predicate, factKeys)
	}
	refs := make([]ActionInputRef, 0, len(factKeys)+len(graph.nodes[nodeIndex].incoming))
	for _, fact := range input.Facts {
		if _, used := factKeys[fact.Key]; !used {
			continue
		}
		refs = append(refs, ActionInputRef{Kind: "fact", ID: fact.Key, SubjectRef: fact.SubjectRef, SubjectRevision: fact.SubjectRevision, Digest: fact.SemanticDigest})
	}
	results := make(map[string]NodeResultFact, len(input.NodeResults))
	for _, result := range input.NodeResults {
		results[result.NodeID] = result
	}
	for _, edgeIndex := range graph.nodes[nodeIndex].incoming {
		edge := graph.edges[edgeIndex]
		result := results[graph.nodes[edge.sourceIndex].definition.ID]
		for _, output := range result.Outputs {
			if output.PortName != edge.definition.SourceRef.PortName {
				continue
			}
			refs = append(refs, ActionInputRef{
				Kind:       "node-output",
				ID:         result.NodeID + "." + output.PortName,
				SubjectRef: output.ValueRef,
				Digest:     output.SemanticDigest,
			})
		}
	}
	for _, gateID := range node.RequiredGateIDs {
		result := results[gateID]
		refs = append(refs, ActionInputRef{Kind: "gate-result", ID: gateID, SubjectRef: result.AttemptID, Digest: result.CandidateDigest})
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Kind != refs[j].Kind {
			return refs[i].Kind < refs[j].Kind
		}
		if refs[i].ID != refs[j].ID {
			return refs[i].ID < refs[j].ID
		}
		return refs[i].SubjectRef < refs[j].SubjectRef
	})
	return refs
}

func collectExpressionFactKeys(expression *Expression, result map[string]struct{}) {
	if expression == nil {
		return
	}
	if expression.FactKey != "" {
		result[expression.FactKey] = struct{}{}
	}
	collectExpressionFactKeys(expression.Operand, result)
	for index := range expression.Operands {
		collectExpressionFactKeys(&expression.Operands[index], result)
	}
}

func (g CompiledGraph) nodeIndex(id string) int {
	index, found := sort.Find(len(g.nodes), func(index int) int { return compareString(id, g.nodes[index].definition.ID) })
	if !found {
		return -1
	}
	return index
}

func (g CompiledGraph) mandatoryGate(id string) bool {
	return slices.Contains(g.definition.MandatoryGateIDs, id)
}

func skipAuthorized(nodeID, policyDigest string, values []SkipAuthorization) bool {
	authorized := false
	for _, value := range values {
		if value.NodeID != nodeID {
			continue
		}
		if value.PolicyDigest != policyDigest || !value.Authorized {
			return false
		}
		authorized = true
	}
	return authorized
}

func compareString(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func graphLimits(graph CompiledGraph) CompileLimits {
	return graph.limits
}

func stableNodeEvaluations(values []NodeEvaluation) []NodeEvaluation {
	result := append([]NodeEvaluation(nil), values...)
	for index := range result {
		result[index].Blockers = append([]Blocker(nil), result[index].Blockers...)
		sortBlockers(result[index].Blockers)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].NodeID < result[j].NodeID })
	return result
}

func sortBlockers(values []Blocker) {
	sort.SliceStable(values, func(i, j int) bool {
		left, right := values[i], values[j]
		if left.NodeID != right.NodeID {
			return left.NodeID < right.NodeID
		}
		if left.SubjectID != right.SubjectID {
			return left.SubjectID < right.SubjectID
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		if left.Source.Path != right.Source.Path {
			return left.Source.Path < right.Source.Path
		}
		if left.Source.StartLine != right.Source.StartLine {
			return left.Source.StartLine < right.Source.StartLine
		}
		return left.Source.StartColumn < right.Source.StartColumn
	})
}

func proposalLess(left, right ActionProposal) bool {
	if left.NodeID != right.NodeID {
		return left.NodeID < right.NodeID
	}
	if left.ActionID != right.ActionID {
		return left.ActionID < right.ActionID
	}
	if left.Source.Path != right.Source.Path {
		return left.Source.Path < right.Source.Path
	}
	if left.Source.StartLine != right.Source.StartLine {
		return left.Source.StartLine < right.Source.StartLine
	}
	return left.Source.StartColumn < right.Source.StartColumn
}

func graphIsComplete(nodes []NodeEvaluation) bool {
	if len(nodes) == 0 {
		return false
	}
	for _, node := range nodes {
		if node.State != StateCompleted && node.State != StateSkipped {
			return false
		}
	}
	return true
}

func finishEvaluation(value Evaluation) Evaluation {
	sortDiagnostics(value.Diagnostics)
	sortBlockers(value.Blockers)
	projection := struct {
		Status            EvaluationStatus      `json:"status"`
		DefinitionRef     CompiledDefinitionRef `json:"definitionRef"`
		Snapshot          SnapshotRef           `json:"snapshot"`
		Nodes             []NodeEvaluation      `json:"nodes"`
		PrimaryAction     *ActionProposal       `json:"primaryAction,omitempty"`
		Alternatives      []ActionProposal      `json:"alternatives"`
		Blockers          []Blocker             `json:"blockers"`
		CompletionHandoff *CompletionHandoff    `json:"completionHandoff,omitempty"`
		Diagnostics       []Diagnostic          `json:"diagnostics"`
	}{
		Status:            value.Status,
		DefinitionRef:     value.DefinitionRef,
		Snapshot:          value.Snapshot,
		Nodes:             value.Nodes,
		PrimaryAction:     value.PrimaryAction,
		Alternatives:      value.Alternatives,
		Blockers:          value.Blockers,
		CompletionHandoff: value.CompletionHandoff,
		Diagnostics:       value.Diagnostics,
	}
	encoded, err := canonicaljson.Marshal(projection)
	if err != nil {
		value.Diagnostics = append(value.Diagnostics, diagnostic("WORKFLOW_INTERNAL_ENCODING", "evaluation could not be canonically encoded", SourceCoordinate{}, "Report this invariant failure.", MetadataEntry{Key: "cause", Value: err.Error()}))
		value.StateDigest = ""
		return value
	}
	value.StateDigest = digestWithDomain(evaluationDigestDomain, encoded)
	return value
}
