package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
)

var diagnosticCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

type ExitCode int

const (
	ExitSuccess      ExitCode = 0
	ExitValidation   ExitCode = 1
	ExitUsage        ExitCode = 2
	ExitWorkspace    ExitCode = 3
	ExitConflict     ExitCode = 4
	ExitExternalTool ExitCode = 5
	ExitPolicy       ExitCode = 6
	ExitInterrupted  ExitCode = 130
)

func (c ExitCode) valid() bool {
	switch c {
	case ExitSuccess, ExitValidation, ExitUsage, ExitWorkspace,
		ExitConflict, ExitExternalTool, ExitPolicy, ExitInterrupted:
		return true
	default:
		return false
	}
}

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

func (s Severity) valid() bool {
	switch s {
	case SeverityInfo, SeverityWarning, SeverityError, SeverityCritical:
		return true
	default:
		return false
	}
}

type PositionUnit string

const (
	PositionUTF8Byte      PositionUnit = "utf8-byte"
	PositionUnicodeScalar PositionUnit = "unicode-scalar"
)

type Position struct {
	line   uint64
	column uint64
	unit   PositionUnit
}

func NewPosition(line, column uint64, unit PositionUnit) (Position, error) {
	if line == 0 || column == 0 {
		return Position{}, fmt.Errorf("%w: source positions are one-based", ErrInvalidProtocolValue)
	}
	if unit != PositionUTF8Byte && unit != PositionUnicodeScalar {
		return Position{}, fmt.Errorf("%w: unsupported position unit %q", ErrInvalidProtocolValue, unit)
	}
	return Position{line: line, column: column, unit: unit}, nil
}

func (p Position) Line() uint64       { return p.line }
func (p Position) Column() uint64     { return p.column }
func (p Position) Unit() PositionUnit { return p.unit }

func (p Position) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Line   uint64       `json:"line"`
		Column uint64       `json:"column"`
		Unit   PositionUnit `json:"unit"`
	}{Line: p.line, Column: p.column, Unit: p.unit})
}

type SourceRange struct {
	path  string
	start Position
	end   *Position
}

func NewSourceRange(path string, start Position, end *Position) (SourceRange, error) {
	if path == "" {
		return SourceRange{}, fmt.Errorf("%w: source range path is required", ErrInvalidProtocolValue)
	}
	var endCopy *Position
	if end != nil {
		copyValue := *end
		endCopy = &copyValue
	}
	return SourceRange{path: path, start: start, end: endCopy}, nil
}

func (r SourceRange) Path() string    { return r.path }
func (r SourceRange) Start() Position { return r.start }
func (r SourceRange) End() *Position  { return clonePointer(r.end) }

func (r SourceRange) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Path  string    `json:"path"`
		Start Position  `json:"start"`
		End   *Position `json:"end,omitempty"`
	}{Path: r.path, Start: r.start, End: r.end})
}

type DiagnosticSource struct {
	kind       string
	path       string
	field      string
	artifactID string
	rangeValue *SourceRange
}

type DiagnosticSourceInput struct {
	Kind       string
	Path       string
	Field      string
	ArtifactID string
	Range      *SourceRange
}

func NewDiagnosticSource(input DiagnosticSourceInput) (DiagnosticSource, error) {
	if strings.TrimSpace(input.Kind) == "" {
		return DiagnosticSource{}, fmt.Errorf("%w: diagnostic source kind is required", ErrInvalidProtocolValue)
	}
	return DiagnosticSource{
		kind:       input.Kind,
		path:       input.Path,
		field:      input.Field,
		artifactID: input.ArtifactID,
		rangeValue: clonePointer(input.Range),
	}, nil
}

func (s DiagnosticSource) Kind() string        { return s.kind }
func (s DiagnosticSource) Path() string        { return s.path }
func (s DiagnosticSource) Field() string       { return s.field }
func (s DiagnosticSource) ArtifactID() string  { return s.artifactID }
func (s DiagnosticSource) Range() *SourceRange { return cloneSourceRange(s.rangeValue) }

func (s DiagnosticSource) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind       string       `json:"kind"`
		Path       string       `json:"path,omitempty"`
		Field      string       `json:"field,omitempty"`
		ArtifactID string       `json:"artifactId,omitempty"`
		Range      *SourceRange `json:"range,omitempty"`
	}{
		Kind:       s.kind,
		Path:       s.path,
		Field:      s.field,
		ArtifactID: s.artifactID,
		Range:      s.rangeValue,
	})
}

type Diagnostic struct {
	code        string
	severity    Severity
	message     string
	source      DiagnosticSource
	remediation string
	metadata    map[string]any
	retryable   bool
}

type DiagnosticInput struct {
	Code        string
	Severity    Severity
	Message     string
	Source      DiagnosticSource
	Remediation string
	Metadata    map[string]any
	Retryable   bool
}

func NewDiagnostic(input DiagnosticInput) (Diagnostic, error) {
	if !diagnosticCodePattern.MatchString(input.Code) {
		return Diagnostic{}, fmt.Errorf("%w: malformed diagnostic code %q", ErrInvalidProtocolValue, input.Code)
	}
	if !input.Severity.valid() {
		return Diagnostic{}, fmt.Errorf("%w: unsupported severity %q", ErrInvalidProtocolValue, input.Severity)
	}
	if strings.TrimSpace(input.Message) == "" {
		return Diagnostic{}, fmt.Errorf("%w: diagnostic message is required", ErrInvalidProtocolValue)
	}
	if strings.TrimSpace(input.Source.kind) == "" {
		return Diagnostic{}, fmt.Errorf("%w: diagnostic source is required", ErrInvalidProtocolValue)
	}
	metadata, err := cloneJSONObject(input.Metadata)
	if err != nil {
		return Diagnostic{}, fmt.Errorf("%w: diagnostic metadata: %w", ErrInvalidProtocolValue, err)
	}
	return Diagnostic{
		code:        input.Code,
		severity:    input.Severity,
		message:     input.Message,
		source:      input.Source,
		remediation: input.Remediation,
		metadata:    metadata,
		retryable:   input.Retryable,
	}, nil
}

func (d Diagnostic) Code() string             { return d.code }
func (d Diagnostic) Severity() Severity       { return d.severity }
func (d Diagnostic) Message() string          { return d.message }
func (d Diagnostic) Source() DiagnosticSource { return d.source }
func (d Diagnostic) Remediation() string      { return d.remediation }
func (d Diagnostic) Metadata() map[string]any { return cloneStoredJSONObject(d.metadata) }
func (d Diagnostic) Retryable() bool          { return d.retryable }

func (d Diagnostic) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code        string           `json:"code"`
		Severity    Severity         `json:"severity"`
		Message     string           `json:"message"`
		Source      DiagnosticSource `json:"source"`
		Remediation string           `json:"remediation,omitempty"`
		Metadata    map[string]any   `json:"metadata,omitempty"`
		Retryable   bool             `json:"retryable"`
	}{
		Code:        d.code,
		Severity:    d.severity,
		Message:     d.message,
		Source:      d.source,
		Remediation: d.remediation,
		Metadata:    d.metadata,
		Retryable:   d.retryable,
	})
}

type FlagKind uint8

const (
	flagString FlagKind = iota + 1
	flagBool
	flagInteger
	flagNumber
)

type FlagValue struct {
	kind        FlagKind
	stringValue string
	boolValue   bool
	intValue    int64
	numberValue float64
}

func StringFlag(value string) FlagValue { return FlagValue{kind: flagString, stringValue: value} }
func BoolFlag(value bool) FlagValue     { return FlagValue{kind: flagBool, boolValue: value} }
func IntegerFlag(value int64) FlagValue { return FlagValue{kind: flagInteger, intValue: value} }

func NumberFlag(value float64) (FlagValue, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return FlagValue{}, fmt.Errorf("%w: flag number must be finite", ErrInvalidProtocolValue)
	}
	return FlagValue{kind: flagNumber, numberValue: value}, nil
}

func (v FlagValue) Value() any {
	switch v.kind {
	case flagString:
		return v.stringValue
	case flagBool:
		return v.boolValue
	case flagInteger:
		return v.intValue
	case flagNumber:
		return v.numberValue
	default:
		return nil
	}
}

func (v FlagValue) MarshalJSON() ([]byte, error) {
	value := v.Value()
	if value == nil {
		return nil, fmt.Errorf("%w: uninitialized flag value", ErrInvalidProtocolValue)
	}
	return json.Marshal(value)
}

type NextAction struct {
	command          string
	args             []string
	flags            map[string]FlagValue
	reason           string
	expectedRevision *uint64
}

type NextActionInput struct {
	Command          string
	Args             []string
	Flags            map[string]FlagValue
	Reason           string
	ExpectedRevision *uint64
}

func NewNextAction(input NextActionInput) (NextAction, error) {
	if strings.TrimSpace(input.Command) == "" || strings.ContainsAny(input.Command, " \t\r\n") {
		return NextAction{}, fmt.Errorf("%w: next-action command must be one executable token", ErrInvalidProtocolValue)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return NextAction{}, fmt.Errorf("%w: next-action reason is required", ErrInvalidProtocolValue)
	}
	flags := make(map[string]FlagValue, len(input.Flags))
	for name, value := range input.Flags {
		if strings.TrimSpace(name) == "" || value.Value() == nil {
			return NextAction{}, fmt.Errorf("%w: invalid next-action flag %q", ErrInvalidProtocolValue, name)
		}
		flags[name] = value
	}
	return NextAction{
		command:          input.Command,
		args:             append([]string{}, input.Args...),
		flags:            flags,
		reason:           input.Reason,
		expectedRevision: clonePointer(input.ExpectedRevision),
	}, nil
}

func (a NextAction) Command() string             { return a.command }
func (a NextAction) Args() []string              { return append([]string{}, a.args...) }
func (a NextAction) Flags() map[string]FlagValue { return cloneMap(a.flags) }
func (a NextAction) Reason() string              { return a.reason }
func (a NextAction) ExpectedRevision() *uint64   { return clonePointer(a.expectedRevision) }

func (a NextAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Command          string               `json:"command"`
		Args             []string             `json:"args"`
		Flags            map[string]FlagValue `json:"flags"`
		Reason           string               `json:"reason"`
		ExpectedRevision *uint64              `json:"expectedRevision,omitempty"`
	}{
		Command:          a.command,
		Args:             nonNilSlice(a.args),
		Flags:            nonNilMap(a.flags),
		Reason:           a.reason,
		ExpectedRevision: a.expectedRevision,
	})
}

func cloneJSONObject(value map[string]any) (map[string]any, error) {
	if value == nil {
		return map[string]any{}, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var cloned map[string]any
	if err := decoder.Decode(&cloned); err != nil {
		return nil, err
	}
	return cloned, nil
}

func cloneStoredJSONObject(value map[string]any) map[string]any {
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		cloned[key] = cloneStoredJSONValue(item)
	}
	return cloned
}

func cloneStoredJSONValue(value any) any {
	switch typed := value.(type) {
	case []any:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = cloneStoredJSONValue(item)
		}
		return items
	case map[string]any:
		return cloneStoredJSONObject(typed)
	default:
		return typed
	}
}

func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func cloneMap[K comparable, V any](value map[K]V) map[K]V {
	cloned := make(map[K]V, len(value))
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}

func nonNilSlice[T any](value []T) []T {
	if value == nil {
		return []T{}
	}
	return value
}

func nonNilMap[K comparable, V any](value map[K]V) map[K]V {
	if value == nil {
		return map[K]V{}
	}
	return value
}
