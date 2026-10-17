package workflow

import "sort"

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

type MetadataEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Diagnostic is the graph-owned nested diagnostic. The engine adapter maps it
// to the foundation diagnostic without selecting behavior from Message.
type Diagnostic struct {
	Code        string           `json:"code"`
	Severity    Severity         `json:"severity"`
	Message     string           `json:"message"`
	Source      SourceCoordinate `json:"source"`
	Remediation string           `json:"remediation,omitempty"`
	Metadata    []MetadataEntry  `json:"metadata"`
	Retryable   bool             `json:"retryable"`
}

func diagnostic(code, message string, source SourceCoordinate, remediation string, metadata ...MetadataEntry) Diagnostic {
	result := append([]MetadataEntry(nil), metadata...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		return result[i].Value < result[j].Value
	})
	return Diagnostic{
		Code:        code,
		Severity:    SeverityError,
		Message:     message,
		Source:      source,
		Remediation: remediation,
		Metadata:    result,
	}
}

func sortDiagnostics(values []Diagnostic) {
	sort.SliceStable(values, func(i, j int) bool {
		left, right := values[i], values[j]
		if left.Source.Path != right.Source.Path {
			return left.Source.Path < right.Source.Path
		}
		if left.Source.StartLine != right.Source.StartLine {
			return left.Source.StartLine < right.Source.StartLine
		}
		if left.Source.StartColumn != right.Source.StartColumn {
			return left.Source.StartColumn < right.Source.StartColumn
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		if left.Message != right.Message {
			return left.Message < right.Message
		}
		for index := 0; index < len(left.Metadata) && index < len(right.Metadata); index++ {
			if left.Metadata[index].Key != right.Metadata[index].Key {
				return left.Metadata[index].Key < right.Metadata[index].Key
			}
			if left.Metadata[index].Value != right.Metadata[index].Value {
				return left.Metadata[index].Value < right.Metadata[index].Value
			}
		}
		return len(left.Metadata) < len(right.Metadata)
	})
}
