package workflow

import (
	"fmt"
	"slices"
	"sort"
)

type TruthValue string

const (
	TruthTrue    TruthValue = "true"
	TruthFalse   TruthValue = "false"
	TruthUnknown TruthValue = "unknown"
)

type UnknownReason struct {
	FactKey        string           `json:"factKey"`
	Owner          string           `json:"owner"`
	Reason         string           `json:"reason"`
	ExpectedDigest string           `json:"expectedDigest,omitempty"`
	CurrentDigest  string           `json:"currentDigest,omitempty"`
	Source         SourceCoordinate `json:"source"`
}

type PredicateResult struct {
	Value    TruthValue      `json:"value"`
	Unknowns []UnknownReason `json:"unknowns"`
}

type FactValidity string

const (
	ValidityCurrent     FactValidity = "current"
	ValidityStale       FactValidity = "stale"
	ValidityUnknown     FactValidity = "unknown"
	ValidityRevoked     FactValidity = "revoked"
	ValidityUnavailable FactValidity = "unavailable"
)

// FactRecord is an opaque, owner-produced predicate fact. Workflow compares
// its declared subjects and validity but never derives authenticity from Value.
type FactRecord struct {
	Key             string       `json:"key"`
	Owner           string       `json:"owner"`
	SubjectRef      string       `json:"subjectRef"`
	SubjectRevision uint64       `json:"subjectRevision"`
	SemanticDigest  string       `json:"semanticDigest,omitempty"`
	ExpectedDigest  string       `json:"expectedDigest,omitempty"`
	Validity        FactValidity `json:"validity"`
	Present         bool         `json:"present"`
	Value           Scalar       `json:"value"`
}

// OwnerObservation distinguishes observed absence from an observation that
// never completed. It is used only by the present predicate.
type OwnerObservation struct {
	Owner    string `json:"owner"`
	Complete bool   `json:"complete"`
}

type factSnapshot struct {
	byKey       map[string]FactRecord
	observed    map[string]bool
	definitions map[string]FactType
}

func buildFactSnapshot(definition Definition, facts []FactRecord, observations []OwnerObservation) (factSnapshot, []Diagnostic) {
	result := factSnapshot{
		byKey:       make(map[string]FactRecord, len(facts)),
		observed:    make(map[string]bool, len(observations)),
		definitions: make(map[string]FactType, len(definition.FactTypes)),
	}
	for _, factType := range definition.FactTypes {
		result.definitions[factType.Key] = factType
	}
	var diagnostics []Diagnostic
	for _, observation := range observations {
		if !validStableID(observation.Owner) {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_FACT_OWNER_INVALID", "owner observation has an invalid owner", SourceCoordinate{}, "Use the declared stable owner ID."))
			continue
		}
		if previous, exists := result.observed[observation.Owner]; exists && previous != observation.Complete {
			diagnostics = append(diagnostics, diagnostic(
				"WORKFLOW_FACT_CONTRADICTORY",
				"owner observation completeness is contradictory",
				SourceCoordinate{},
				"Provide one observation-completeness fact per owner.",
				MetadataEntry{Key: "owner", Value: observation.Owner},
			))
			continue
		}
		result.observed[observation.Owner] = observation.Complete
	}
	for _, fact := range facts {
		factType, declared := result.definitions[fact.Key]
		if !declared {
			diagnostics = append(diagnostics, diagnostic(
				"WORKFLOW_FACT_UNKNOWN_TYPE",
				"runtime fact is not declared by the pinned definition",
				SourceCoordinate{},
				"Remove the fact or migrate to a definition that declares it.",
				MetadataEntry{Key: "factKey", Value: fact.Key},
			))
			continue
		}
		if fact.Owner != factType.Owner {
			diagnostics = append(diagnostics, diagnostic(
				"WORKFLOW_FACT_OWNER_MISMATCH",
				"runtime fact owner does not match the definition",
				factType.Source,
				"Obtain the fact from the pinned owner.",
				MetadataEntry{Key: "actualOwner", Value: fact.Owner},
				MetadataEntry{Key: "expectedOwner", Value: factType.Owner},
				MetadataEntry{Key: "factKey", Value: fact.Key},
			))
			continue
		}
		if fact.SubjectRef == "" || fact.SemanticDigest == "" || !validFactValidity(fact.Validity) {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_FACT_INVALID", "runtime fact subject or validity is invalid", factType.Source, "Provide a subject-bound fact with a supported validity state.", MetadataEntry{Key: "factKey", Value: fact.Key}))
			continue
		}
		if fact.Present {
			if (factType.Kind == FactPresence && scalarSetCount(fact.Value) != 0) || (factType.Kind != FactPresence && !scalarMatches(fact.Value, factType)) {
				diagnostics = append(diagnostics, diagnostic("WORKFLOW_FACT_TYPE_MISMATCH", "runtime fact value does not match its declared type", factType.Source, "Provide a typed fact value from the declared owner.", MetadataEntry{Key: "factKey", Value: fact.Key}))
				continue
			}
		} else if scalarSetCount(fact.Value) != 0 {
			diagnostics = append(diagnostics, diagnostic("WORKFLOW_FACT_TYPE_MISMATCH", "absent runtime fact carries a value", factType.Source, "Remove the value or mark the fact present.", MetadataEntry{Key: "factKey", Value: fact.Key}))
			continue
		}
		if previous, exists := result.byKey[fact.Key]; exists {
			code := "WORKFLOW_FACT_DUPLICATE"
			message := "runtime fact is duplicated"
			if !factRecordsEqual(previous, fact) {
				code = "WORKFLOW_FACT_CONTRADICTORY"
				message = "runtime facts for the same key are contradictory"
			}
			diagnostics = append(diagnostics, diagnostic(code, message, factType.Source, "Provide one immutable fact per key and subject.", MetadataEntry{Key: "factKey", Value: fact.Key}))
			continue
		}
		result.byKey[fact.Key] = cloneFactRecord(fact)
	}
	sortDiagnostics(diagnostics)
	return result, diagnostics
}

func validFactValidity(value FactValidity) bool {
	return value == ValidityCurrent || value == ValidityStale || value == ValidityUnknown || value == ValidityRevoked || value == ValidityUnavailable
}

func scalarSetCount(value Scalar) int {
	count := 0
	if value.Bool != nil {
		count++
	}
	if value.Enum != nil {
		count++
	}
	if value.Integer != nil {
		count++
	}
	return count
}

func factRecordsEqual(left, right FactRecord) bool {
	return left.Key == right.Key &&
		left.Owner == right.Owner &&
		left.SubjectRef == right.SubjectRef &&
		left.SubjectRevision == right.SubjectRevision &&
		left.SemanticDigest == right.SemanticDigest &&
		left.ExpectedDigest == right.ExpectedDigest &&
		left.Validity == right.Validity &&
		left.Present == right.Present &&
		scalarsEqual(left.Value, right.Value)
}

func cloneFactRecord(fact FactRecord) FactRecord {
	result := fact
	result.Value = cloneScalar(fact.Value)
	return result
}

func cloneScalar(value Scalar) Scalar {
	result := value
	if value.Bool != nil {
		copyValue := *value.Bool
		result.Bool = &copyValue
	}
	if value.Enum != nil {
		copyValue := *value.Enum
		result.Enum = &copyValue
	}
	if value.Integer != nil {
		copyValue := *value.Integer
		result.Integer = &copyValue
	}
	return result
}

func evaluatePredicate(expression *Expression, facts factSnapshot) PredicateResult {
	if expression == nil {
		return PredicateResult{Value: TruthTrue, Unknowns: []UnknownReason{}}
	}
	switch expression.Kind {
	case ExpressionConst:
		if *expression.Const {
			return knownPredicate(TruthTrue)
		}
		return knownPredicate(TruthFalse)
	case ExpressionBoolFact:
		fact, unknown := facts.currentFact(expression.FactKey, expression.Source)
		if unknown != nil {
			return unknownPredicate(*unknown)
		}
		if !fact.Present || fact.Value.Bool == nil {
			return facts.missing(expression.FactKey, expression.Source, "boolean fact has no observed value")
		}
		if *fact.Value.Bool {
			return knownPredicate(TruthTrue)
		}
		return knownPredicate(TruthFalse)
	case ExpressionPresent:
		factType := facts.definitions[expression.FactKey]
		fact, exists := facts.byKey[expression.FactKey]
		if exists {
			if _, unknown := facts.currentFact(expression.FactKey, expression.Source); unknown != nil {
				return unknownPredicate(*unknown)
			}
			if fact.Present {
				return knownPredicate(TruthTrue)
			}
			if facts.observed[factType.Owner] {
				return knownPredicate(TruthFalse)
			}
		}
		if facts.observed[factType.Owner] {
			return knownPredicate(TruthFalse)
		}
		return facts.missing(expression.FactKey, expression.Source, "owner observation is incomplete")
	case ExpressionEqual, ExpressionNotEqual, ExpressionIn, ExpressionCompare:
		fact, unknown := facts.currentFact(expression.FactKey, expression.Source)
		if unknown != nil {
			return unknownPredicate(*unknown)
		}
		if !fact.Present {
			return facts.missing(expression.FactKey, expression.Source, "fact value is absent")
		}
		var matched bool
		switch expression.Kind {
		case ExpressionEqual, ExpressionNotEqual:
			matched = scalarsEqual(fact.Value, *expression.Value)
			if expression.Kind == ExpressionNotEqual {
				matched = !matched
			}
		case ExpressionIn:
			matched = slices.ContainsFunc(expression.Values, func(value Scalar) bool { return scalarsEqual(fact.Value, value) })
		case ExpressionCompare:
			matched = compareIntegers(*fact.Value.Integer, *expression.Value.Integer, expression.Operator)
		default:
			return unknownPredicate(UnknownReason{FactKey: expression.FactKey, Reason: "compiled comparison kind is invalid", Source: expression.Source})
		}
		if matched {
			return knownPredicate(TruthTrue)
		}
		return knownPredicate(TruthFalse)
	case ExpressionNot:
		result := evaluatePredicate(expression.Operand, facts)
		switch result.Value {
		case TruthTrue:
			result.Value = TruthFalse
		case TruthFalse:
			result.Value = TruthTrue
		case TruthUnknown:
			// Strong Kleene negation preserves unknown and its reasons.
		}
		return result
	case ExpressionAll:
		unknowns := make([]UnknownReason, 0)
		for index := range expression.Operands {
			result := evaluatePredicate(&expression.Operands[index], facts)
			if result.Value == TruthFalse {
				return knownPredicate(TruthFalse)
			}
			if result.Value == TruthUnknown {
				unknowns = append(unknowns, result.Unknowns...)
			}
		}
		if len(unknowns) != 0 {
			return PredicateResult{Value: TruthUnknown, Unknowns: normalizeUnknowns(unknowns)}
		}
		return knownPredicate(TruthTrue)
	case ExpressionAny:
		unknowns := make([]UnknownReason, 0)
		for index := range expression.Operands {
			result := evaluatePredicate(&expression.Operands[index], facts)
			if result.Value == TruthTrue {
				return knownPredicate(TruthTrue)
			}
			if result.Value == TruthUnknown {
				unknowns = append(unknowns, result.Unknowns...)
			}
		}
		if len(unknowns) != 0 {
			return PredicateResult{Value: TruthUnknown, Unknowns: normalizeUnknowns(unknowns)}
		}
		return knownPredicate(TruthFalse)
	default:
		return unknownPredicate(UnknownReason{FactKey: expression.FactKey, Reason: fmt.Sprintf("compiled predicate kind %q is invalid", expression.Kind), Source: expression.Source})
	}
}

func (f factSnapshot) currentFact(key string, source SourceCoordinate) (FactRecord, *UnknownReason) {
	factType := f.definitions[key]
	fact, exists := f.byKey[key]
	if !exists {
		reason := UnknownReason{FactKey: key, Owner: factType.Owner, Reason: "fact is missing", Source: source}
		return FactRecord{}, &reason
	}
	if fact.Validity != ValidityCurrent {
		reason := UnknownReason{
			FactKey:        key,
			Owner:          factType.Owner,
			Reason:         "fact validity is " + string(fact.Validity),
			ExpectedDigest: fact.ExpectedDigest,
			CurrentDigest:  fact.SemanticDigest,
			Source:         source,
		}
		return FactRecord{}, &reason
	}
	if fact.ExpectedDigest != "" && fact.SemanticDigest != fact.ExpectedDigest {
		reason := UnknownReason{
			FactKey:        key,
			Owner:          factType.Owner,
			Reason:         "fact digest does not match its expected subject",
			ExpectedDigest: fact.ExpectedDigest,
			CurrentDigest:  fact.SemanticDigest,
			Source:         source,
		}
		return FactRecord{}, &reason
	}
	return fact, nil
}

func (f factSnapshot) missing(key string, source SourceCoordinate, reason string) PredicateResult {
	factType := f.definitions[key]
	return unknownPredicate(UnknownReason{FactKey: key, Owner: factType.Owner, Reason: reason, Source: source})
}

func knownPredicate(value TruthValue) PredicateResult {
	return PredicateResult{Value: value, Unknowns: []UnknownReason{}}
}

func unknownPredicate(reason UnknownReason) PredicateResult {
	return PredicateResult{Value: TruthUnknown, Unknowns: []UnknownReason{reason}}
}

func normalizeUnknowns(values []UnknownReason) []UnknownReason {
	result := append([]UnknownReason(nil), values...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].FactKey != result[j].FactKey {
			return result[i].FactKey < result[j].FactKey
		}
		if result[i].Reason != result[j].Reason {
			return result[i].Reason < result[j].Reason
		}
		return result[i].Source.Path < result[j].Source.Path
	})
	return slices.CompactFunc(result, func(left, right UnknownReason) bool { return left == right })
}

func scalarsEqual(left, right Scalar) bool {
	if scalarSetCount(left) == 0 && scalarSetCount(right) == 0 {
		return true
	}
	if left.Bool != nil && right.Bool != nil {
		return *left.Bool == *right.Bool
	}
	if left.Enum != nil && right.Enum != nil {
		return *left.Enum == *right.Enum
	}
	if left.Integer != nil && right.Integer != nil {
		return *left.Integer == *right.Integer
	}
	return false
}

func compareIntegers(left, right int64, operator CompareOperator) bool {
	switch operator {
	case CompareLess:
		return left < right
	case CompareLessOrEqual:
		return left <= right
	case CompareGreater:
		return left > right
	case CompareGreaterOrEqual:
		return left >= right
	default:
		return false
	}
}
