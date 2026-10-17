package workflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"sort"
	"strings"

	"sdd.dev/sdd/internal/protocol/canonicaljson"
)

var (
	stableIDPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,127}$`)
	typeRefPattern  = regexp.MustCompile(`^[a-z][a-z0-9._-]*/[a-z][a-z0-9._-]*@[1-9][0-9]*$`)
)

type compiledNode struct {
	definition     NodeDefinition
	actionRevision string
	incoming       []int
	outgoing       []int
}

type compiledEdge struct {
	definition  EdgeDefinition
	sourceIndex int
	targetIndex int
}

// CompiledGraph is an immutable graph value. Its process-local dense indexes
// are never serialized as identity or exposed to callers.
type CompiledGraph struct {
	ref        CompiledDefinitionRef
	definition Definition
	limits     CompileLimits
	nodes      []compiledNode
	edges      []compiledEdge
	topology   []int
}

func (g CompiledGraph) Ref() CompiledDefinitionRef { return g.ref }

func (g CompiledGraph) DefinitionID() string { return g.ref.Definition.DefinitionID }

func (g CompiledGraph) NodeIDs() []string {
	result := make([]string, len(g.nodes))
	for index := range g.nodes {
		result[index] = g.nodes[index].definition.ID
	}
	return result
}

func (g CompiledGraph) valid() bool {
	return g.ref.CompiledDigest != "" && len(g.nodes) != 0
}

// Compile decodes, validates, proves, normalizes, and identifies one closed
// graph definition. Any diagnostic means no runnable graph is returned.
func Compile(ctx context.Context, source DefinitionSource, requested CompileLimits) (CompiledGraph, []Diagnostic) {
	if ctx == nil {
		return CompiledGraph{}, []Diagnostic{diagnostic(
			"WORKFLOW_CONTEXT_INVALID",
			"compile requires a non-nil context",
			source.Coordinate,
			"Pass the request context to workflow.Compile.",
		)}
	}
	limits, limitDiagnostics := normalizeLimits(requested, source.Coordinate)
	if len(limitDiagnostics) != 0 {
		return CompiledGraph{}, limitDiagnostics
	}
	if err := ctx.Err(); err != nil {
		return CompiledGraph{}, []Diagnostic{cancelledDiagnostic(source.Coordinate, err)}
	}
	if uint64(len(source.Bytes)) > limits.DefinitionBytes {
		return CompiledGraph{}, []Diagnostic{limitDiagnostic(
			source.Coordinate,
			"definition-bytes",
			uint64(len(source.Bytes)),
			limits.DefinitionBytes,
		)}
	}

	diagnostics := validateDefinitionSource(source)
	if len(diagnostics) != 0 {
		sortDiagnostics(diagnostics)
		return CompiledGraph{}, diagnostics
	}

	exactDigest := digestBytes(source.Bytes)
	if source.ExactDigest != "" && source.ExactDigest != exactDigest {
		return CompiledGraph{}, []Diagnostic{diagnostic(
			"WORKFLOW_DEFINITION_DIGEST_MISMATCH",
			"definition bytes do not match the exact digest claim",
			source.Coordinate,
			"Reload the pinned definition from its authoritative source.",
			MetadataEntry{Key: "actual", Value: exactDigest},
			MetadataEntry{Key: "expected", Value: source.ExactDigest},
		)}
	}

	definition, decodeDiagnostics := decodeDefinition(source)
	if len(decodeDiagnostics) != 0 {
		return CompiledGraph{}, decodeDiagnostics
	}
	compiler := graphCompiler{
		ctx:        ctx,
		source:     source,
		definition: definition,
		limits:     limits,
		nodeByID:   make(map[string]int, len(definition.Nodes)),
	}
	graph := compiler.compile(exactDigest)
	if len(compiler.diagnostics) != 0 {
		sortDiagnostics(compiler.diagnostics)
		return CompiledGraph{}, slices.Clone(compiler.diagnostics)
	}
	return graph, nil
}

type graphCompiler struct {
	ctx            context.Context
	source         DefinitionSource
	definition     Definition
	limits         CompileLimits
	diagnostics    []Diagnostic
	nodeByID       map[string]int
	predicateNodes uint64
	proofSteps     uint64
	portCount      uint64
	cancelled      bool
}

func (c *graphCompiler) compile(exactDigest string) CompiledGraph {
	c.validateHeader()
	if c.stop() {
		return CompiledGraph{}
	}
	c.validateFacts()
	if c.stop() {
		return CompiledGraph{}
	}
	c.validateNodes()
	if c.stop() {
		return CompiledGraph{}
	}

	normalized := normalizeDefinition(c.definition)
	nodes := make([]compiledNode, len(normalized.Nodes))
	for index := range normalized.Nodes {
		nodes[index].definition = normalized.Nodes[index]
		c.nodeByID[normalized.Nodes[index].ID] = index
	}
	edges := c.validateEdges(normalized, nodes)
	c.validateMandatoryGates(normalized, nodes, edges)
	if c.stop() {
		return CompiledGraph{}
	}

	topology := c.topologicalOrder(nodes, edges)
	if c.stop() {
		return CompiledGraph{}
	}
	semanticDigest, err := semanticDefinitionDigest(normalized)
	if err != nil {
		c.add(diagnostic(
			"WORKFLOW_DEFINITION_INVALID",
			"definition cannot be canonically identified",
			c.source.Coordinate,
			"Remove unsupported representation values from the definition.",
			MetadataEntry{Key: "cause", Value: err.Error()},
		))
		return CompiledGraph{}
	}
	limitsDigest, err := digestCanonical(c.limits)
	if err != nil {
		c.add(internalCompileDiagnostic(c.source.Coordinate, err))
		return CompiledGraph{}
	}
	compiledProjection := struct {
		SemanticDigest string        `json:"semanticDigest"`
		Compiler       string        `json:"compiler"`
		Limits         CompileLimits `json:"limits"`
		Topology       []string      `json:"topology"`
	}{
		SemanticDigest: semanticDigest,
		Compiler:       CompilerContractVersion,
		Limits:         c.limits,
		Topology:       nodeIDsAt(nodes, topology),
	}
	compiledBytes, err := canonicaljson.Marshal(compiledProjection)
	if err != nil {
		c.add(internalCompileDiagnostic(c.source.Coordinate, err))
		return CompiledGraph{}
	}
	compiledDigest := digestWithDomain(compiledDigestDomain, compiledBytes)
	for index := range nodes {
		actionRevision, revisionErr := actionRevision(nodes[index].definition, compiledDigest)
		if revisionErr != nil {
			c.add(internalCompileDiagnostic(nodes[index].definition.Action.Source, revisionErr))
			return CompiledGraph{}
		}
		nodes[index].actionRevision = actionRevision
	}

	return CompiledGraph{
		ref: CompiledDefinitionRef{
			Definition: DefinitionRef{
				DefinitionID:                normalized.DefinitionID,
				SchemaVersion:               normalized.SchemaVersion,
				SemanticVersion:             normalized.SemanticVersion,
				SourceID:                    c.source.SourceID,
				SourceVersion:               c.source.Version,
				ExactDigest:                 exactDigest,
				CanonicalizerID:             normalized.Canonicalizer.ID,
				CanonicalizerVersion:        normalized.Canonicalizer.Version,
				CanonicalizerDefinitionHash: normalized.Canonicalizer.DefinitionDigest,
				SemanticDigest:              semanticDigest,
			},
			CompilerVersion: CompilerContractVersion,
			LimitsDigest:    limitsDigest,
			CompiledDigest:  compiledDigest,
		},
		definition: normalized,
		limits:     c.limits,
		nodes:      nodes,
		edges:      edges,
		topology:   topology,
	}
}

func decodeDefinition(source DefinitionSource) (Definition, []Diagnostic) {
	decoder := json.NewDecoder(bytes.NewReader(source.Bytes))
	decoder.DisallowUnknownFields()
	var definition Definition
	if err := decoder.Decode(&definition); err != nil {
		return Definition{}, []Diagnostic{diagnostic(
			"WORKFLOW_DEFINITION_INVALID",
			"definition is not valid closed JSON",
			source.Coordinate,
			"Correct the JSON document and remove unknown fields.",
			MetadataEntry{Key: "cause", Value: err.Error()},
		)}
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return Definition{}, []Diagnostic{diagnostic(
			"WORKFLOW_DEFINITION_INVALID",
			"definition contains trailing data",
			source.Coordinate,
			"Provide exactly one JSON definition document.",
			MetadataEntry{Key: "cause", Value: err.Error()},
		)}
	}
	return definition, nil
}

func validateDefinitionSource(source DefinitionSource) []Diagnostic {
	var diagnostics []Diagnostic
	if source.MediaType != DefinitionMediaType {
		diagnostics = append(diagnostics, diagnostic(
			"WORKFLOW_DEFINITION_MEDIA_UNSUPPORTED",
			"workflow definition media type is unsupported",
			source.Coordinate,
			"Use the accepted version-1 JSON media type.",
			MetadataEntry{Key: "actual", Value: source.MediaType},
			MetadataEntry{Key: "expected", Value: DefinitionMediaType},
		))
	}
	if source.SchemaID != DefinitionSchemaID {
		diagnostics = append(diagnostics, diagnostic(
			"WORKFLOW_DEFINITION_SCHEMA_UNSUPPORTED",
			"workflow definition schema is unsupported",
			source.Coordinate,
			"Use the accepted version-1 schema identifier.",
			MetadataEntry{Key: "actual", Value: source.SchemaID},
			MetadataEntry{Key: "expected", Value: DefinitionSchemaID},
		))
	}
	if !validStableID(source.SourceID) || strings.TrimSpace(source.Version) == "" {
		diagnostics = append(diagnostics, diagnostic(
			"WORKFLOW_DEFINITION_SOURCE_INVALID",
			"definition source identity and version are required",
			source.Coordinate,
			"Provide a stable lowercase source ID and immutable source version.",
		))
	}
	if len(source.Bytes) == 0 {
		diagnostics = append(diagnostics, diagnostic(
			"WORKFLOW_DEFINITION_INVALID",
			"definition source is empty",
			source.Coordinate,
			"Provide one closed JSON definition document.",
		))
	}
	if source.ExactDigest != "" && !validSHA256(source.ExactDigest) {
		diagnostics = append(diagnostics, diagnostic(
			"WORKFLOW_DEFINITION_DIGEST_INVALID",
			"exact digest claim is not lowercase SHA-256",
			source.Coordinate,
			"Provide a 64-character lowercase hexadecimal SHA-256 digest.",
		))
	}
	return diagnostics
}

func (c *graphCompiler) validateHeader() {
	definition := c.definition
	if definition.SchemaVersion != definitionSchemaVersion {
		c.add(diagnostic(
			"WORKFLOW_DEFINITION_SCHEMA_UNSUPPORTED",
			"definition schema version is unsupported",
			c.source.Coordinate,
			"Migrate the definition to the accepted schema version.",
			MetadataEntry{Key: "actual", Value: definition.SchemaVersion},
			MetadataEntry{Key: "expected", Value: definitionSchemaVersion},
		))
	}
	if !validStableID(definition.DefinitionID) || strings.TrimSpace(definition.SemanticVersion) == "" {
		c.add(diagnostic(
			"WORKFLOW_DEFINITION_INVALID",
			"definition ID or semantic version is invalid",
			c.source.Coordinate,
			"Use a stable lowercase definition ID and a non-empty semantic version.",
		))
	}
	if definition.Canonicalizer.ID != CanonicalizerID || definition.Canonicalizer.Version != CanonicalizerVersion || !validSHA256(definition.Canonicalizer.DefinitionDigest) {
		c.add(diagnostic(
			"WORKFLOW_CANONICALIZER_UNSUPPORTED",
			"definition does not pin the accepted canonicalizer contract",
			c.source.Coordinate,
			"Pin the accepted canonicalizer ID, version, and definition digest.",
		))
	}
	if len(definition.Nodes) == 0 {
		c.add(diagnostic(
			"WORKFLOW_DEFINITION_INVALID",
			"definition contains no nodes",
			c.source.Coordinate,
			"Declare at least one workflow node.",
		))
	}
	if uint64(len(definition.Nodes)) > c.limits.Nodes {
		c.add(limitDiagnostic(c.source.Coordinate, "nodes", uint64(len(definition.Nodes)), c.limits.Nodes))
	}
	if uint64(len(definition.Edges)) > c.limits.Edges {
		c.add(limitDiagnostic(c.source.Coordinate, "edges", uint64(len(definition.Edges)), c.limits.Edges))
	}
}

func (c *graphCompiler) validateFacts() {
	seen := make(map[string]SourceCoordinate, len(c.definition.FactTypes))
	for index := range c.definition.FactTypes {
		factType := c.definition.FactTypes[index]
		if c.cancelledNow() {
			return
		}
		if !validStableID(factType.Key) || !validStableID(factType.Owner) {
			c.add(diagnostic("WORKFLOW_FACT_TYPE_INVALID", "fact key or owner is invalid", factType.Source, "Use stable lowercase fact and owner IDs."))
		}
		if previous, exists := seen[factType.Key]; exists {
			c.add(duplicateDiagnostic("fact", factType.Key, previous, factType.Source))
		} else {
			seen[factType.Key] = factType.Source
		}
		switch factType.Kind {
		case FactBool, FactPresence:
			if len(factType.EnumValues) != 0 || factType.Minimum != nil || factType.Maximum != nil {
				c.add(diagnostic("WORKFLOW_FACT_TYPE_INVALID", "boolean or presence fact has incompatible constraints", factType.Source, "Remove enum and integer constraints."))
			}
		case FactEnum:
			if len(factType.EnumValues) == 0 {
				c.add(diagnostic("WORKFLOW_FACT_TYPE_INVALID", "enum fact has no values", factType.Source, "Declare at least one unique enum value."))
			}
			if hasDuplicateStrings(factType.EnumValues) {
				c.add(diagnostic("WORKFLOW_FACT_TYPE_INVALID", "enum fact contains duplicate values", factType.Source, "Remove duplicate enum values."))
			}
		case FactInteger:
			if factType.Minimum == nil || factType.Maximum == nil || *factType.Minimum > *factType.Maximum {
				c.add(diagnostic("WORKFLOW_FACT_TYPE_INVALID", "integer fact requires a valid bounded range", factType.Source, "Declare minimum and maximum integer values."))
			}
		default:
			c.add(diagnostic("WORKFLOW_FACT_TYPE_INVALID", "fact kind is unsupported", factType.Source, "Use bool, enum, integer, or presence."))
		}
	}
}

func (c *graphCompiler) validateNodes() {
	seenNodes := make(map[string]SourceCoordinate, len(c.definition.Nodes))
	seenActions := make(map[string]SourceCoordinate, len(c.definition.Nodes))
	factTypes := make(map[string]FactType, len(c.definition.FactTypes))
	for _, factType := range c.definition.FactTypes {
		factTypes[factType.Key] = factType
	}
	for index := range c.definition.Nodes {
		if c.cancelledNow() {
			return
		}
		node := &c.definition.Nodes[index]
		if !validStableID(node.ID) || strings.TrimSpace(node.Version) == "" || !validNodeKind(node.Kind) {
			c.add(diagnostic("WORKFLOW_NODE_INVALID", "node identity, version, or kind is invalid", node.Source, "Use a stable ID, non-empty version, and supported node kind."))
		}
		if previous, exists := seenNodes[node.ID]; exists {
			c.add(duplicateDiagnostic("node", node.ID, previous, node.Source))
		} else {
			seenNodes[node.ID] = node.Source
		}
		if !validStableID(node.Action.ID) || !validActionKind(node.Action.Kind) {
			c.add(diagnostic("WORKFLOW_ACTION_INVALID", "node action identity or kind is invalid", node.Action.Source, "Use a stable action ID and supported action kind."))
		}
		if previous, exists := seenActions[node.Action.ID]; exists {
			c.add(duplicateDiagnostic("action", node.Action.ID, previous, node.Action.Source))
		} else {
			seenActions[node.Action.ID] = node.Action.Source
		}
		if node.Kind == NodeExecution && node.Action.ExecutionBinding == "" {
			c.add(diagnostic("WORKFLOW_ACTION_INVALID", "execution node has no execution binding", node.Action.Source, "Declare a typed execution binding reference."))
		}
		if node.Kind != NodeExecution && node.Action.ExecutionBinding != "" {
			c.add(diagnostic("WORKFLOW_ACTION_INVALID", "non-execution node carries an execution binding", node.Action.Source, "Remove the execution binding."))
		}
		c.validatePorts(node)
		c.validateExpression(node.ApplicabilityPredicate, factTypes, 1)
		c.validateExpression(node.SkipPredicate, factTypes, 1)
	}
	if c.portCount > c.limits.Ports {
		c.add(limitDiagnostic(c.source.Coordinate, "ports", c.portCount, c.limits.Ports))
	}
	if c.predicateNodes > c.limits.PredicateNodes {
		c.add(limitDiagnostic(c.source.Coordinate, "predicate-nodes", c.predicateNodes, c.limits.PredicateNodes))
	}
}

func (c *graphCompiler) validatePorts(node *NodeDefinition) {
	seen := make(map[string]SourceCoordinate, len(node.Inputs)+len(node.Outputs))
	validate := func(port PortDefinition, expected PortDirection) {
		c.portCount++
		if !validStableID(port.Name) || port.Direction != expected || !validTypeRef(port.TypeRef) || !validCardinality(port.Cardinality) {
			c.add(diagnostic("WORKFLOW_PORT_INVALID", "port name, direction, type, or cardinality is invalid", port.Source, "Declare an explicit supported port contract."))
		}
		if previous, exists := seen[port.Name]; exists {
			c.add(duplicateDiagnostic("port", node.ID+"."+port.Name, previous, port.Source))
		} else {
			seen[port.Name] = port.Source
		}
	}
	for _, port := range node.Inputs {
		validate(port, PortInput)
	}
	for _, port := range node.Outputs {
		validate(port, PortOutput)
	}
}

func (c *graphCompiler) validateExpression(expression *Expression, factTypes map[string]FactType, depth uint64) {
	if expression == nil || c.cancelledNow() {
		return
	}
	c.predicateNodes++
	if depth > c.limits.PredicateDepth {
		c.add(limitDiagnostic(expression.Source, "predicate-depth", depth, c.limits.PredicateDepth))
		return
	}
	factType, factExists := factTypes[expression.FactKey]
	switch expression.Kind {
	case ExpressionConst:
		if expression.Const == nil || expression.FactKey != "" || expression.Value != nil || len(expression.Values) != 0 || expression.Operator != "" || expression.Operand != nil || len(expression.Operands) != 0 {
			c.invalidExpression(*expression, "const requires exactly one Boolean literal")
		}
	case ExpressionBoolFact:
		if !factExists || factType.Kind != FactBool || expressionHasNonFactFields(expression) {
			c.invalidExpression(*expression, "bool-fact requires a declared Boolean fact")
		}
	case ExpressionPresent:
		if !factExists || expressionHasNonFactFields(expression) {
			c.invalidExpression(*expression, "present requires a declared fact")
		}
	case ExpressionEqual, ExpressionNotEqual:
		if !factExists || expression.Value == nil || !scalarMatches(*expression.Value, factType) || expression.Const != nil || len(expression.Values) != 0 || expression.Operator != "" || expression.Operand != nil || len(expression.Operands) != 0 {
			c.invalidExpression(*expression, "equality requires one typed literal matching the declared fact")
		}
	case ExpressionIn:
		if !factExists || len(expression.Values) == 0 || expression.Const != nil || expression.Value != nil || expression.Operator != "" || expression.Operand != nil || len(expression.Operands) != 0 {
			c.invalidExpression(*expression, "in requires a declared fact and at least one literal")
			break
		}
		for _, value := range expression.Values {
			if !scalarMatches(value, factType) {
				c.invalidExpression(*expression, "in contains a literal that does not match the fact type")
				break
			}
		}
	case ExpressionCompare:
		if !factExists || factType.Kind != FactInteger || expression.Value == nil || !scalarMatches(*expression.Value, factType) || !validCompareOperator(expression.Operator) || expression.Const != nil || len(expression.Values) != 0 || expression.Operand != nil || len(expression.Operands) != 0 {
			c.invalidExpression(*expression, "compare requires a bounded integer fact, integer literal, and supported operator")
		}
	case ExpressionNot:
		if expression.Operand == nil || expression.FactKey != "" || expression.Const != nil || expression.Value != nil || len(expression.Values) != 0 || expression.Operator != "" || len(expression.Operands) != 0 {
			c.invalidExpression(*expression, "not requires exactly one operand")
			break
		}
		c.validateExpression(expression.Operand, factTypes, depth+1)
	case ExpressionAll, ExpressionAny:
		if len(expression.Operands) == 0 || expression.FactKey != "" || expression.Const != nil || expression.Value != nil || len(expression.Values) != 0 || expression.Operator != "" || expression.Operand != nil {
			c.invalidExpression(*expression, "all and any require at least one operand")
			break
		}
		for index := range expression.Operands {
			c.validateExpression(&expression.Operands[index], factTypes, depth+1)
		}
	default:
		c.invalidExpression(*expression, "predicate operator is unsupported")
	}
}

func (c *graphCompiler) invalidExpression(expression Expression, reason string) {
	c.add(diagnostic("WORKFLOW_PREDICATE_INVALID", reason, expression.Source, "Use only the accepted closed predicate grammar."))
}

func (c *graphCompiler) validateEdges(definition Definition, nodes []compiledNode) []compiledEdge {
	seen := make(map[string]SourceCoordinate, len(definition.Edges))
	edges := make([]compiledEdge, 0, len(definition.Edges))
	factTypes := make(map[string]FactType, len(definition.FactTypes))
	for _, factType := range definition.FactTypes {
		factTypes[factType.Key] = factType
	}
	for _, edge := range definition.Edges {
		if c.cancelledNow() {
			return nil
		}
		if !validStableID(edge.ID) || !validEdgeRelation(edge.Relation) {
			c.add(diagnostic("WORKFLOW_EDGE_INVALID", "edge identity or relation is invalid", edge.Coordinate, "Use a stable edge ID and dependency or freshness relation."))
		}
		if previous, exists := seen[edge.ID]; exists {
			c.add(duplicateDiagnostic("edge", edge.ID, previous, edge.Coordinate))
		} else {
			seen[edge.ID] = edge.Coordinate
		}
		sourceIndex, sourceExists := c.nodeByID[edge.SourceRef.NodeID]
		targetIndex, targetExists := c.nodeByID[edge.TargetRef.NodeID]
		if !sourceExists || !targetExists {
			c.add(diagnostic(
				"WORKFLOW_REFERENCE_MISSING",
				"edge references a missing node",
				edge.Coordinate,
				"Correct the edge node references.",
				MetadataEntry{Key: "sourceNode", Value: edge.SourceRef.NodeID},
				MetadataEntry{Key: "targetNode", Value: edge.TargetRef.NodeID},
			))
			continue
		}
		sourcePort, sourceFound := outputPort(nodes[sourceIndex].definition, edge.SourceRef.PortName)
		targetPort, targetFound := inputPort(nodes[targetIndex].definition, edge.TargetRef.PortName)
		if !sourceFound || !targetFound {
			c.add(diagnostic("WORKFLOW_REFERENCE_MISSING", "edge references a missing source output or target input", edge.Coordinate, "Correct the edge port references."))
			continue
		}
		if sourcePort.TypeRef != targetPort.TypeRef {
			c.add(diagnostic(
				"WORKFLOW_TYPE_MISMATCH",
				"edge connects incompatible port types",
				edge.Coordinate,
				"Connect ports with the same versioned type reference.",
				MetadataEntry{Key: "sourceType", Value: sourcePort.TypeRef},
				MetadataEntry{Key: "targetType", Value: targetPort.TypeRef},
			))
		}
		c.validateExpression(edge.Predicate, factTypes, 1)
		compiledIndex := len(edges)
		edges = append(edges, compiledEdge{definition: edge, sourceIndex: sourceIndex, targetIndex: targetIndex})
		nodes[sourceIndex].outgoing = append(nodes[sourceIndex].outgoing, compiledIndex)
		nodes[targetIndex].incoming = append(nodes[targetIndex].incoming, compiledIndex)
	}
	for nodeIndex := range nodes {
		for _, port := range nodes[nodeIndex].definition.Inputs {
			count := 0
			for _, edgeIndex := range nodes[nodeIndex].incoming {
				if edges[edgeIndex].definition.TargetRef.PortName == port.Name && edges[edgeIndex].definition.Relation == EdgeDependency && !predicateAlwaysFalse(edges[edgeIndex].definition.Predicate) {
					count++
				}
			}
			if (port.Cardinality == CardinalityOne || port.Cardinality == CardinalityAtLeastOne) && count == 0 {
				c.add(diagnostic("WORKFLOW_PRODUCER_MISSING", "required input has no dependency producer", port.Source, "Connect at least one compatible dependency output."))
			}
			if (port.Cardinality == CardinalityOne || port.Cardinality == CardinalityZeroOrOne) && count > 1 {
				c.add(diagnostic("WORKFLOW_CARDINALITY_INVALID", "single-valued input has multiple possible producers", port.Source, "Reduce the input to one possible dependency producer."))
			}
		}
	}
	return edges
}

func (c *graphCompiler) validateMandatoryGates(definition Definition, nodes []compiledNode, edges []compiledEdge) {
	seen := make(map[string]struct{}, len(definition.MandatoryGateIDs))
	for _, gateID := range definition.MandatoryGateIDs {
		if _, exists := seen[gateID]; exists {
			c.add(diagnostic("WORKFLOW_GATE_INVALID", "mandatory gate is listed more than once", c.source.Coordinate, "List each mandatory gate once.", MetadataEntry{Key: "gateId", Value: gateID}))
			continue
		}
		seen[gateID] = struct{}{}
		index, exists := c.nodeByID[gateID]
		if !exists || definition.Nodes[index].Kind != NodeGate {
			c.add(diagnostic("WORKFLOW_GATE_UNREACHABLE", "mandatory gate does not reference a gate node", c.source.Coordinate, "Declare the referenced node as a gate.", MetadataEntry{Key: "gateId", Value: gateID}))
		}
	}
	for _, node := range definition.Nodes {
		for _, gateID := range node.RequiredGateIDs {
			index, exists := c.nodeByID[gateID]
			if !exists || definition.Nodes[index].Kind != NodeGate {
				c.add(diagnostic("WORKFLOW_REFERENCE_MISSING", "node requires a missing gate", node.Source, "Reference an existing gate node.", MetadataEntry{Key: "gateId", Value: gateID}, MetadataEntry{Key: "nodeId", Value: node.ID}))
				continue
			}
			if !pathExists(index, c.nodeByID[node.ID], nodes, edges) {
				c.add(diagnostic("WORKFLOW_GATE_UNREACHABLE", "required gate is not a predecessor of the guarded node", node.Source, "Connect the gate to the guarded node through a possible dependency or freshness path.", MetadataEntry{Key: "gateId", Value: gateID}, MetadataEntry{Key: "nodeId", Value: node.ID}))
			}
		}
	}
}

func pathExists(source, target int, nodes []compiledNode, edges []compiledEdge) bool {
	if source == target {
		return false
	}
	visited := make([]bool, len(nodes))
	visited[source] = true
	queue := []int{source}
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		for _, edgeIndex := range nodes[current].outgoing {
			edge := edges[edgeIndex]
			if predicateAlwaysFalse(edge.definition.Predicate) || visited[edge.targetIndex] {
				continue
			}
			if edge.targetIndex == target {
				return true
			}
			visited[edge.targetIndex] = true
			queue = append(queue, edge.targetIndex)
		}
	}
	return false
}

func (c *graphCompiler) topologicalOrder(nodes []compiledNode, edges []compiledEdge) []int {
	indegree := make([]int, len(nodes))
	for _, edge := range edges {
		if predicateAlwaysFalse(edge.definition.Predicate) {
			continue
		}
		indegree[edge.targetIndex]++
	}
	ready := make([]int, 0, len(nodes))
	for index := range nodes {
		if indegree[index] == 0 {
			ready = append(ready, index)
		}
	}
	sort.Slice(ready, func(i, j int) bool { return nodes[ready[i]].definition.ID < nodes[ready[j]].definition.ID })
	result := make([]int, 0, len(nodes))
	for len(ready) != 0 {
		if c.cancelledNow() {
			return nil
		}
		current := ready[0]
		ready = ready[1:]
		result = append(result, current)
		for _, edgeIndex := range nodes[current].outgoing {
			edge := edges[edgeIndex]
			if predicateAlwaysFalse(edge.definition.Predicate) {
				continue
			}
			c.proofSteps++
			if c.proofSteps > c.limits.ProofSteps {
				c.add(limitDiagnostic(edge.definition.Coordinate, "proof-steps", c.proofSteps, c.limits.ProofSteps))
				return nil
			}
			indegree[edge.targetIndex]--
			if indegree[edge.targetIndex] == 0 {
				ready = append(ready, edge.targetIndex)
				sort.Slice(ready, func(i, j int) bool { return nodes[ready[i]].definition.ID < nodes[ready[j]].definition.ID })
			}
		}
	}
	if len(result) != len(nodes) {
		cycle := findCycle(nodes, edges)
		metadata := make([]MetadataEntry, 0, len(cycle)+1)
		relations := make(map[string]EdgeRelation, len(edges))
		for _, edge := range edges {
			relations[edge.definition.ID] = edge.definition.Relation
		}
		for index, edgeID := range cycle {
			metadata = append(metadata, MetadataEntry{Key: fmt.Sprintf("edge.%04d", index), Value: edgeID + ":" + string(relations[edgeID])})
		}
		code := "WORKFLOW_READINESS_CYCLE"
		if cycleIsDependencyOnly(cycle, edges) {
			code = "WORKFLOW_DEPENDENCY_CYCLE"
		}
		c.add(diagnostic(code, "workflow contains a possible readiness cycle", c.source.Coordinate, "Remove or prove mutual exclusion for the reported cycle.", metadata...))
	}
	return result
}

func findCycle(nodes []compiledNode, edges []compiledEdge) []string {
	state := make([]uint8, len(nodes))
	stackNodes := make([]int, 0, len(nodes))
	stackEdges := make([]int, 0, len(nodes))
	var cycle []string
	var visit func(int) bool
	visit = func(node int) bool {
		state[node] = 1
		stackNodes = append(stackNodes, node)
		for _, edgeIndex := range nodes[node].outgoing {
			edge := edges[edgeIndex]
			if predicateAlwaysFalse(edge.definition.Predicate) {
				continue
			}
			target := edge.targetIndex
			if state[target] == 0 {
				stackEdges = append(stackEdges, edgeIndex)
				if visit(target) {
					return true
				}
				stackEdges = stackEdges[:len(stackEdges)-1]
				continue
			}
			if state[target] == 1 {
				start := slices.Index(stackNodes, target)
				for _, pathEdge := range stackEdges[start:] {
					cycle = append(cycle, edges[pathEdge].definition.ID)
				}
				cycle = append(cycle, edge.definition.ID)
				return true
			}
		}
		stackNodes = stackNodes[:len(stackNodes)-1]
		state[node] = 2
		return false
	}
	for index := range nodes {
		if state[index] == 0 && visit(index) {
			break
		}
	}
	return cycle
}

func cycleIsDependencyOnly(cycle []string, edges []compiledEdge) bool {
	byID := make(map[string]EdgeRelation, len(edges))
	for _, edge := range edges {
		byID[edge.definition.ID] = edge.definition.Relation
	}
	for _, edgeID := range cycle {
		if byID[edgeID] != EdgeDependency {
			return false
		}
	}
	return true
}

func (c *graphCompiler) add(value Diagnostic) {
	if uint64(len(c.diagnostics)) >= c.limits.Diagnostics {
		return
	}
	c.diagnostics = append(c.diagnostics, value)
}

func (c *graphCompiler) stop() bool {
	return c.cancelledNow() || len(c.diagnostics) != 0
}

func (c *graphCompiler) cancelledNow() bool {
	if c.cancelled {
		return true
	}
	if err := c.ctx.Err(); err != nil {
		c.cancelled = true
		c.add(cancelledDiagnostic(c.source.Coordinate, err))
		return true
	}
	return false
}

func normalizeLimits(requested CompileLimits, source SourceCoordinate) (CompileLimits, []Diagnostic) {
	result := requested
	var diagnostics []Diagnostic
	normalize := func(name string, value *uint64, hard uint64) {
		if *value == 0 {
			*value = hard
			return
		}
		if *value > hard {
			diagnostics = append(diagnostics, diagnostic(
				"WORKFLOW_COMPILE_LIMIT_INVALID",
				"requested compile limit exceeds the accepted hard ceiling",
				source,
				"Use the hard ceiling or a lower deployment limit.",
				MetadataEntry{Key: "hardMaximum", Value: fmt.Sprint(hard)},
				MetadataEntry{Key: "requested", Value: fmt.Sprint(*value)},
				MetadataEntry{Key: "resource", Value: name},
			))
		}
	}
	normalize("definition-bytes", &result.DefinitionBytes, hardCompileLimits.DefinitionBytes)
	normalize("nodes", &result.Nodes, hardCompileLimits.Nodes)
	normalize("edges", &result.Edges, hardCompileLimits.Edges)
	normalize("ports", &result.Ports, hardCompileLimits.Ports)
	normalize("predicate-nodes", &result.PredicateNodes, hardCompileLimits.PredicateNodes)
	normalize("predicate-depth", &result.PredicateDepth, hardCompileLimits.PredicateDepth)
	normalize("proof-region-atoms", &result.ProofRegionAtoms, hardCompileLimits.ProofRegionAtoms)
	normalize("proof-steps", &result.ProofSteps, hardCompileLimits.ProofSteps)
	normalize("diagnostics", &result.Diagnostics, hardCompileLimits.Diagnostics)
	normalize("runtime-facts", &result.RuntimeFacts, hardCompileLimits.RuntimeFacts)
	normalize("evaluation-steps", &result.EvaluationSteps, hardCompileLimits.EvaluationSteps)
	sortDiagnostics(diagnostics)
	return result, diagnostics
}

func normalizeDefinition(input Definition) Definition {
	result := input
	result.FactTypes = slices.Clone(input.FactTypes)
	for index := range result.FactTypes {
		result.FactTypes[index].EnumValues = slices.Clone(result.FactTypes[index].EnumValues)
		sort.Strings(result.FactTypes[index].EnumValues)
	}
	sort.Slice(result.FactTypes, func(i, j int) bool { return result.FactTypes[i].Key < result.FactTypes[j].Key })
	result.Nodes = slices.Clone(input.Nodes)
	for index := range result.Nodes {
		node := &result.Nodes[index]
		node.Inputs = slices.Clone(node.Inputs)
		node.Outputs = slices.Clone(node.Outputs)
		node.RequiredGateIDs = slices.Clone(node.RequiredGateIDs)
		node.ApplicabilityPredicate = cloneExpression(node.ApplicabilityPredicate)
		node.SkipPredicate = cloneExpression(node.SkipPredicate)
		sort.Slice(node.Inputs, func(i, j int) bool { return node.Inputs[i].Name < node.Inputs[j].Name })
		sort.Slice(node.Outputs, func(i, j int) bool { return node.Outputs[i].Name < node.Outputs[j].Name })
		sort.Strings(node.RequiredGateIDs)
	}
	sort.Slice(result.Nodes, func(i, j int) bool { return result.Nodes[i].ID < result.Nodes[j].ID })
	result.Edges = slices.Clone(input.Edges)
	for index := range result.Edges {
		result.Edges[index].Predicate = cloneExpression(result.Edges[index].Predicate)
	}
	sort.Slice(result.Edges, func(i, j int) bool { return result.Edges[i].ID < result.Edges[j].ID })
	result.MandatoryGateIDs = slices.Clone(input.MandatoryGateIDs)
	sort.Strings(result.MandatoryGateIDs)
	result.CompatibilityRules = slices.Clone(input.CompatibilityRules)
	sort.Slice(result.CompatibilityRules, func(i, j int) bool { return result.CompatibilityRules[i].ID < result.CompatibilityRules[j].ID })
	return result
}

func actionRevision(node NodeDefinition, compiledDigest string) (string, error) {
	action := node.Action
	action.Source = SourceCoordinate{}
	action.RemediationTemplate = ""
	projection := struct {
		CompiledDigest string           `json:"compiledDigest"`
		NodeID         string           `json:"nodeId"`
		NodeVersion    string           `json:"nodeVersion"`
		Action         ActionDefinition `json:"action"`
	}{CompiledDigest: compiledDigest, NodeID: node.ID, NodeVersion: node.Version, Action: action}
	encoded, err := canonicaljson.Marshal(projection)
	if err != nil {
		return "", err
	}
	return digestWithDomain("sdd.workflow.action.v1\x00", encoded), nil
}

func cloneExpression(expression *Expression) *Expression {
	if expression == nil {
		return nil
	}
	result := *expression
	if expression.Value != nil {
		value := *expression.Value
		result.Value = &value
	}
	result.Values = slices.Clone(expression.Values)
	result.Operand = cloneExpression(expression.Operand)
	result.Operands = slices.Clone(expression.Operands)
	for index := range result.Operands {
		cloned := cloneExpression(&result.Operands[index])
		result.Operands[index] = *cloned
	}
	return &result
}

func semanticDefinitionDigest(definition Definition) (string, error) {
	semantic := definition
	semantic.Description = ""
	for index := range semantic.FactTypes {
		semantic.FactTypes[index].Source = SourceCoordinate{}
	}
	for index := range semantic.Nodes {
		node := &semantic.Nodes[index]
		node.Description = ""
		node.Source = SourceCoordinate{}
		node.Action.Source = SourceCoordinate{}
		node.Action.RemediationTemplate = ""
		for portIndex := range node.Inputs {
			node.Inputs[portIndex].Source = SourceCoordinate{}
		}
		for portIndex := range node.Outputs {
			node.Outputs[portIndex].Source = SourceCoordinate{}
		}
		clearExpressionCoordinates(node.ApplicabilityPredicate)
		clearExpressionCoordinates(node.SkipPredicate)
	}
	for index := range semantic.Edges {
		semantic.Edges[index].Coordinate = SourceCoordinate{}
		clearExpressionCoordinates(semantic.Edges[index].Predicate)
	}
	encoded, err := canonicaljson.Marshal(semantic)
	if err != nil {
		return "", err
	}
	return digestWithDomain(semanticDigestDomain, encoded), nil
}

func clearExpressionCoordinates(expression *Expression) {
	if expression == nil {
		return
	}
	expression.Source = SourceCoordinate{}
	clearExpressionCoordinates(expression.Operand)
	for index := range expression.Operands {
		clearExpressionCoordinates(&expression.Operands[index])
	}
}

func digestCanonical(value any) (string, error) {
	encoded, err := canonicaljson.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(encoded), nil
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func digestWithDomain(domain string, value []byte) string {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(domain))
	_, _ = hasher.Write(value)
	return hex.EncodeToString(hasher.Sum(nil))
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validStableID(value string) bool {
	return stableIDPattern.MatchString(value) && len(value) <= maximumStableIDBytes
}

func validTypeRef(value string) bool {
	return len(value) <= maximumTypeReferenceBytes && typeRefPattern.MatchString(value)
}

func validNodeKind(value NodeKind) bool {
	return value == NodeAuthored || value == NodeAssessment || value == NodeExecution || value == NodeGate || value == NodeHandoff
}

func validActionKind(value ActionKind) bool {
	return value == ActionSubmit || value == ActionAssess || value == ActionExecute || value == ActionAcceptComposition || value == ActionHandoffCompletion
}

func validCardinality(value Cardinality) bool {
	return value == CardinalityOne || value == CardinalityZeroOrOne || value == CardinalityZeroOrMany || value == CardinalityAtLeastOne
}

func validEdgeRelation(value EdgeRelation) bool {
	return value == EdgeDependency || value == EdgeFreshness
}

func validCompareOperator(value CompareOperator) bool {
	return value == CompareLess || value == CompareLessOrEqual || value == CompareGreater || value == CompareGreaterOrEqual
}

func expressionHasNonFactFields(expression *Expression) bool {
	return expression.Const != nil || expression.Value != nil || len(expression.Values) != 0 || expression.Operator != "" || expression.Operand != nil || len(expression.Operands) != 0
}

func scalarMatches(value Scalar, factType FactType) bool {
	set := 0
	if value.Bool != nil {
		set++
	}
	if value.Enum != nil {
		set++
	}
	if value.Integer != nil {
		set++
	}
	if set != 1 {
		return false
	}
	switch factType.Kind {
	case FactBool:
		return value.Bool != nil
	case FactEnum:
		return value.Enum != nil && slices.Contains(factType.EnumValues, *value.Enum)
	case FactInteger:
		return value.Integer != nil && factType.Minimum != nil && factType.Maximum != nil && *value.Integer >= *factType.Minimum && *value.Integer <= *factType.Maximum
	default:
		return false
	}
}

func outputPort(node NodeDefinition, name string) (PortDefinition, bool) {
	for _, port := range node.Outputs {
		if port.Name == name {
			return port, true
		}
	}
	return PortDefinition{}, false
}

func inputPort(node NodeDefinition, name string) (PortDefinition, bool) {
	for _, port := range node.Inputs {
		if port.Name == name {
			return port, true
		}
	}
	return PortDefinition{}, false
}

func hasDuplicateStrings(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func predicateAlwaysFalse(expression *Expression) bool {
	return expression != nil && expression.Kind == ExpressionConst && expression.Const != nil && !*expression.Const
}

func nodeIDsAt(nodes []compiledNode, indexes []int) []string {
	result := make([]string, len(indexes))
	for index, nodeIndex := range indexes {
		result[index] = nodes[nodeIndex].definition.ID
	}
	return result
}

func duplicateDiagnostic(kind, id string, first, second SourceCoordinate) Diagnostic {
	return diagnostic(
		"WORKFLOW_DUPLICATE_ID",
		"workflow definition contains a duplicate "+kind+" identity",
		second,
		"Give every "+kind+" a unique stable identity.",
		MetadataEntry{Key: "firstPath", Value: first.Path},
		MetadataEntry{Key: "id", Value: id},
		MetadataEntry{Key: "kind", Value: kind},
	)
}

func limitDiagnostic(source SourceCoordinate, resource string, actual, maximum uint64) Diagnostic {
	return diagnostic(
		"WORKFLOW_COMPILE_LIMIT",
		"workflow definition exceeds a bounded compiler resource",
		source,
		"Reduce the definition or select a reviewed compiler-contract revision.",
		MetadataEntry{Key: "actual", Value: fmt.Sprint(actual)},
		MetadataEntry{Key: "maximum", Value: fmt.Sprint(maximum)},
		MetadataEntry{Key: "resource", Value: resource},
	)
}

func cancelledDiagnostic(source SourceCoordinate, err error) Diagnostic {
	return diagnostic(
		"WORKFLOW_CANCELLED",
		"workflow operation was cancelled",
		source,
		"Retry with a live request context.",
		MetadataEntry{Key: "cause", Value: err.Error()},
	)
}

func internalCompileDiagnostic(source SourceCoordinate, err error) Diagnostic {
	return diagnostic(
		"WORKFLOW_DEFINITION_INVALID",
		"workflow definition could not be canonically compiled",
		source,
		"Correct the definition and retry.",
		MetadataEntry{Key: "cause", Value: err.Error()},
	)
}
