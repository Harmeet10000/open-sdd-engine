package render

import (
	"regexp"
	"sort"
	"strings"
)

const redactionMarker = "[REDACTED]"

var (
	bearerPattern = regexp.MustCompile(`(?i)(\bbearer\s+)[A-Za-z0-9._~+/=-]+`)
	// Credential assignments come in quoted and bare shapes. Quoted
	// patterns run first and preserve their delimiters so redacted JSON
	// stays parseable; the bare pattern never consumes quotes.
	quotedDoubleCredentialPattern = regexp.MustCompile(`(?i)(\b(?:password|passwd|token|secret|api[-_]?key)["']?\s*[:=]\s*")([^"]*)(")`)
	quotedSingleCredentialPattern = regexp.MustCompile(`(?i)(\b(?:password|passwd|token|secret|api[-_]?key)["']?\s*[:=]\s*')([^']*)(')`)
	bareCredentialPattern         = regexp.MustCompile(`(?i)(\b(?:password|passwd|token|secret|api[-_]?key)\s*[:=]\s*)([^\s,;\"'}]+)`)
	urlCredentialPattern          = regexp.MustCompile(`([A-Za-z][A-Za-z0-9+.-]*://)[^/@\s]+@`)
)

// Redactor removes configured literal values and common credential forms.
// Literals are applied longest-first so overlapping secrets do not leak a
// suffix through replacement order.
type Redactor struct {
	literals []string
}

func NewRedactor(literals ...string) Redactor {
	unique := make(map[string]struct{}, len(literals))
	for _, literal := range literals {
		if literal != "" {
			unique[literal] = struct{}{}
		}
	}
	values := make([]string, 0, len(unique))
	for literal := range unique {
		values = append(values, literal)
	}
	sort.Slice(values, func(left, right int) bool {
		if len(values[left]) == len(values[right]) {
			return values[left] < values[right]
		}
		return len(values[left]) > len(values[right])
	})
	return Redactor{literals: values}
}

func (r Redactor) Redact(value string) string {
	if value == "" {
		return ""
	}
	redacted := value
	for _, literal := range r.literals {
		redacted = strings.ReplaceAll(redacted, literal, redactionMarker)
	}
	redacted = bearerPattern.ReplaceAllString(redacted, `${1}`+redactionMarker)
	redacted = quotedDoubleCredentialPattern.ReplaceAllString(redacted, `$1`+redactionMarker+`$3`)
	redacted = quotedSingleCredentialPattern.ReplaceAllString(redacted, `$1`+redactionMarker+`$3`)
	redacted = bareCredentialPattern.ReplaceAllString(redacted, `${1}`+redactionMarker)
	redacted = urlCredentialPattern.ReplaceAllString(redacted, `${1}`+redactionMarker+`@`)
	return redacted
}
