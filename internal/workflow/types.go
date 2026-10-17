// Package workflow compiles closed workflow definitions and deterministically
// derives workflow state from immutable owner facts. It performs no I/O and
// grants no authority.
package workflow

import "time"

const (
	DefinitionMediaType       = "application/vnd.open-sdd.workflow-definition+json;version=1"
	DefinitionSchemaID        = "urn:open-sdd:workflow-definition:1"
	CanonicalizerID           = "sdd.workflow.canonical-json"
	CanonicalizerVersion      = "1"
	CompilerContractVersion   = "sdd.workflow.compiler/1"
	definitionSchemaVersion   = "1"
	semanticDigestDomain      = "sdd.workflow.semantic.v1\x00"
	compiledDigestDomain      = "sdd.workflow.compiled.v1\x00"
	evaluationDigestDomain    = "sdd.workflow.evaluation.v1\x00"
	maximumStableIDBytes      = 128
	maximumTypeReferenceBytes = 256
)

// SourceCoordinate identifies a location using foundation-compatible, one-based
// coordinates. An absent range is represented by zero line and column values.
type SourceCoordinate struct {
	Path        string `json:"path,omitempty"`
	Field       string `json:"field,omitempty"`
	StartLine   uint64 `json:"startLine,omitempty"`
	StartColumn uint64 `json:"startColumn,omitempty"`
	EndLine     uint64 `json:"endLine,omitempty"`
	EndColumn   uint64 `json:"endColumn,omitempty"`
	Unit        string `json:"unit,omitempty"`
}

// DefinitionSource is the trusted transport wrapper around exact definition
// bytes. ExactDigest is an optional lowercase SHA-256 claim over Bytes.
type DefinitionSource struct {
	MediaType   string
	SchemaID    string
	SourceID    string
	Version     string
	ExactDigest string
	Coordinate  SourceCoordinate
	Bytes       []byte
}

// Definition is the closed version-1 graph document.
type Definition struct {
	SchemaVersion      string              `json:"schemaVersion"`
	DefinitionID       string              `json:"definitionId"`
	SemanticVersion    string              `json:"semanticVersion"`
	Canonicalizer      CanonicalizerRef    `json:"canonicalizer"`
	Description        string              `json:"description,omitempty"`
	FactTypes          []FactType          `json:"factTypes"`
	Nodes              []NodeDefinition    `json:"nodes"`
	Edges              []EdgeDefinition    `json:"edges"`
	MandatoryGateIDs   []string            `json:"mandatoryGateIds"`
	CompatibilityRules []CompatibilityRule `json:"compatibilityRules"`
}

type CanonicalizerRef struct {
	ID               string `json:"id"`
	Version          string `json:"version"`
	DefinitionDigest string `json:"definitionDigest"`
}

type FactKind string

const (
	FactBool     FactKind = "bool"
	FactEnum     FactKind = "enum"
	FactInteger  FactKind = "integer"
	FactPresence FactKind = "presence"
)

type FactType struct {
	Key        string           `json:"key"`
	Kind       FactKind         `json:"kind"`
	EnumValues []string         `json:"enumValues,omitempty"`
	Minimum    *int64           `json:"minimum,omitempty"`
	Maximum    *int64           `json:"maximum,omitempty"`
	Owner      string           `json:"owner"`
	Source     SourceCoordinate `json:"source"`
}

type NodeKind string

const (
	NodeAuthored   NodeKind = "authored"
	NodeAssessment NodeKind = "assessment"
	NodeExecution  NodeKind = "execution"
	NodeGate       NodeKind = "gate"
	NodeHandoff    NodeKind = "handoff"
)

type NodeDefinition struct {
	ID                     string           `json:"id"`
	Version                string           `json:"version"`
	Kind                   NodeKind         `json:"kind"`
	Description            string           `json:"description,omitempty"`
	Inputs                 []PortDefinition `json:"inputs"`
	Outputs                []PortDefinition `json:"outputs"`
	ApplicabilityPredicate *Expression      `json:"applicabilityPredicate,omitempty"`
	SkipPredicate          *Expression      `json:"skipPredicate,omitempty"`
	RequiredGateIDs        []string         `json:"requiredGateIds"`
	ProfileBinding         string           `json:"profileBinding"`
	Action                 ActionDefinition `json:"action"`
	Source                 SourceCoordinate `json:"source"`
}

type PortDirection string

const (
	PortInput  PortDirection = "input"
	PortOutput PortDirection = "output"
)

type Cardinality string

const (
	CardinalityOne        Cardinality = "one"
	CardinalityZeroOrOne  Cardinality = "zero-or-one"
	CardinalityZeroOrMany Cardinality = "zero-or-many"
	CardinalityAtLeastOne Cardinality = "at-least-one"
)

type PortDefinition struct {
	Name        string           `json:"name"`
	Direction   PortDirection    `json:"direction"`
	TypeRef     string           `json:"typeRef"`
	Cardinality Cardinality      `json:"cardinality"`
	Source      SourceCoordinate `json:"source"`
}

type EdgeRelation string

const (
	EdgeDependency EdgeRelation = "dependency"
	EdgeFreshness  EdgeRelation = "freshness"
)

type PortReference struct {
	NodeID   string `json:"nodeId"`
	PortName string `json:"portName"`
}

type EdgeDefinition struct {
	ID         string           `json:"id"`
	SourceRef  PortReference    `json:"source"`
	TargetRef  PortReference    `json:"target"`
	Relation   EdgeRelation     `json:"relation"`
	Predicate  *Expression      `json:"predicate,omitempty"`
	Coordinate SourceCoordinate `json:"coordinate"`
}

type ActionKind string

const (
	ActionSubmit            ActionKind = "submit"
	ActionAssess            ActionKind = "assess"
	ActionExecute           ActionKind = "execute"
	ActionAcceptComposition ActionKind = "accept-composition"
	ActionHandoffCompletion ActionKind = "handoff-completion"
)

type ActionDefinition struct {
	ID                  string           `json:"id"`
	Kind                ActionKind       `json:"kind"`
	SubmissionTypeRef   string           `json:"submissionTypeRef,omitempty"`
	ValidatorBinding    string           `json:"validatorBinding,omitempty"`
	ExecutionBinding    string           `json:"executionBinding,omitempty"`
	RemediationTemplate string           `json:"remediationTemplate,omitempty"`
	Source              SourceCoordinate `json:"source"`
}

type CompatibilityRule struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type ExpressionKind string

const (
	ExpressionConst    ExpressionKind = "const"
	ExpressionBoolFact ExpressionKind = "bool-fact"
	ExpressionPresent  ExpressionKind = "present"
	ExpressionEqual    ExpressionKind = "equal"
	ExpressionNotEqual ExpressionKind = "not-equal"
	ExpressionIn       ExpressionKind = "in"
	ExpressionCompare  ExpressionKind = "compare"
	ExpressionNot      ExpressionKind = "not"
	ExpressionAll      ExpressionKind = "all"
	ExpressionAny      ExpressionKind = "any"
)

type CompareOperator string

const (
	CompareLess           CompareOperator = "lt"
	CompareLessOrEqual    CompareOperator = "le"
	CompareGreater        CompareOperator = "gt"
	CompareGreaterOrEqual CompareOperator = "ge"
)

// Scalar is a closed predicate literal. Exactly one pointer must be non-nil.
type Scalar struct {
	Bool    *bool   `json:"bool,omitempty"`
	Enum    *string `json:"enum,omitempty"`
	Integer *int64  `json:"integer,omitempty"`
}

// Expression is closed data, never executable code.
type Expression struct {
	Kind     ExpressionKind   `json:"kind"`
	FactKey  string           `json:"factKey,omitempty"`
	Const    *bool            `json:"const,omitempty"`
	Value    *Scalar          `json:"value,omitempty"`
	Values   []Scalar         `json:"values,omitempty"`
	Operator CompareOperator  `json:"operator,omitempty"`
	Operand  *Expression      `json:"operand,omitempty"`
	Operands []Expression     `json:"operands,omitempty"`
	Source   SourceCoordinate `json:"source"`
}

// CompileLimits are part of compiled identity. Zero selects the accepted hard
// ceiling; a caller may lower but never raise a value.
type CompileLimits struct {
	DefinitionBytes  uint64 `json:"definitionBytes"`
	Nodes            uint64 `json:"nodes"`
	Edges            uint64 `json:"edges"`
	Ports            uint64 `json:"ports"`
	PredicateNodes   uint64 `json:"predicateNodes"`
	PredicateDepth   uint64 `json:"predicateDepth"`
	ProofRegionAtoms uint64 `json:"proofRegionAtoms"`
	ProofSteps       uint64 `json:"proofSteps"`
	Diagnostics      uint64 `json:"diagnostics"`
	RuntimeFacts     uint64 `json:"runtimeFacts"`
	EvaluationSteps  uint64 `json:"evaluationSteps"`
}

var hardCompileLimits = CompileLimits{
	DefinitionBytes:  2 << 20,
	Nodes:            2_048,
	Edges:            8_192,
	Ports:            16_384,
	PredicateNodes:   32_768,
	PredicateDepth:   32,
	ProofRegionAtoms: 16,
	ProofSteps:       2_000_000,
	Diagnostics:      4_096,
	RuntimeFacts:     65_536,
	EvaluationSteps:  1_000_000,
}

// AcceptedCompileLimits returns the immutable safety ceilings accepted by
// ADR-0007. Callers cannot mutate compiler global state through the result.
func AcceptedCompileLimits() CompileLimits { return hardCompileLimits }

type DefinitionRef struct {
	DefinitionID                string `json:"definitionId"`
	SchemaVersion               string `json:"schemaVersion"`
	SemanticVersion             string `json:"semanticVersion"`
	SourceID                    string `json:"sourceId"`
	SourceVersion               string `json:"sourceVersion"`
	ExactDigest                 string `json:"exactDigest"`
	CanonicalizerID             string `json:"canonicalizerId"`
	CanonicalizerVersion        string `json:"canonicalizerVersion"`
	CanonicalizerDefinitionHash string `json:"canonicalizerDefinitionDigest"`
	SemanticDigest              string `json:"semanticDigest"`
}

type CompiledDefinitionRef struct {
	Definition      DefinitionRef `json:"definition"`
	CompilerVersion string        `json:"compilerContractVersion"`
	LimitsDigest    string        `json:"compileLimitsDigest"`
	CompiledDigest  string        `json:"compiledDigest"`
}

// EvaluationInstant is explicit input for expiry checks; the workflow package
// never reads the wall clock.
type EvaluationInstant struct {
	At time.Time `json:"at"`
}
