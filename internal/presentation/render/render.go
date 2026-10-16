// Package render turns one prepared, redacted engine outcome into human or
// JSON output. It owns no command or workflow decisions.
package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"sdd.dev/sdd/internal/engine"
)

var ErrWrite = errors.New("render output write failed")

// Input is the one immutable value accepted by every output renderer.
// Preparing it performs redaction once so renderers cannot diverge.
type Input struct {
	outcome engine.Outcome
}

func NewInput(outcome engine.Outcome, redactor Redactor) Input {
	return Input{outcome: outcome.Redacted(redactor.Redact)}
}

func (i Input) Outcome() engine.Outcome { return i.outcome }

func WriteJSON(writer io.Writer, input Input) error {
	encoded, err := json.Marshal(input.outcome)
	if err != nil {
		return fmt.Errorf("marshal JSON outcome: %w", err)
	}
	encoded = append(encoded, '\n')
	if err := writePayload(writer, encoded); err != nil {
		return fmt.Errorf("%w: JSON: %w", ErrWrite, err)
	}
	return nil
}

// WriteHuman writes primary payload to stdout and actionable diagnostics to
// stderr. It returns the first transport error and never calls os.Exit.
func WriteHuman(stdout, stderr io.Writer, input Input) error {
	primary, diagnostics := humanOutput(input.outcome)
	if primary != "" {
		if err := writePayload(stdout, []byte(primary)); err != nil {
			return fmt.Errorf("%w: human stdout: %w", ErrWrite, err)
		}
	}
	if diagnostics != "" {
		if err := writePayload(stderr, []byte(diagnostics)); err != nil {
			return fmt.Errorf("%w: human stderr: %w", ErrWrite, err)
		}
	}
	return nil
}

func humanOutput(outcome engine.Outcome) (string, string) {
	var stdout strings.Builder
	var stderr strings.Builder

	if outcome.OK() {
		stdout.WriteString(successSummary(outcome.Result()))
	} else {
		stderr.WriteString("Command failed.\n")
	}

	for _, diagnostic := range outcome.Diagnostics() {
		fmt.Fprintf(
			&stderr,
			"%s [%s]: %s\n",
			diagnostic.Severity(),
			diagnostic.Code(),
			diagnostic.Message(),
		)
		if diagnostic.Remediation() != "" {
			fmt.Fprintf(&stderr, "Remediation: %s\n", diagnostic.Remediation())
		}
	}

	actions := outcome.NextActions()
	if len(actions) > 0 {
		target := &stdout
		if !outcome.OK() {
			target = &stderr
		}
		fmt.Fprintf(target, "Next: %s\n", formatAction(actions[0]))
	}

	return stdout.String(), stderr.String()
}

func successSummary(result *engine.Result) string {
	if result == nil {
		return "Command completed.\n"
	}
	switch result.Kind() {
	case engine.ResultStatus:
		status, _ := result.Status()
		count := len(status.WorkItems())
		if count == 0 {
			return "No active work items.\n"
		}
		return fmt.Sprintf("Found %d active work item(s).\n", count)
	case engine.ResultVersion:
		version, _ := result.Version()
		return fmt.Sprintf("sdd %s (%s)\n", version.Version(), version.Commit())
	case engine.ResultDoctor:
		doctor, _ := result.Doctor()
		return fmt.Sprintf("Checked %d system condition(s).\n", len(doctor.Checks()))
	case engine.ResultStartDryRun:
		start, _ := result.StartDryRun()
		plan := start.Plan()
		return fmt.Sprintf("Prepared plan %s with %d operation(s).\n", plan.PlanID(), len(plan.Operations()))
	default:
		return "Command completed.\n"
	}
}

func formatAction(action engine.NextAction) string {
	parts := []string{"sdd", quoteArgument(action.Command())}
	for _, arg := range action.Args() {
		parts = append(parts, quoteArgument(arg))
	}
	flags := action.Flags()
	names := make([]string, 0, len(flags))
	for name := range flags {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		flagName := "--" + strings.TrimLeft(name, "-")
		value := flags[name].Value()
		if enabled, ok := value.(bool); ok {
			if enabled {
				parts = append(parts, flagName)
			} else {
				parts = append(parts, flagName+"=false")
			}
			continue
		}
		parts = append(parts, flagName, quoteArgument(fmt.Sprint(value)))
	}
	return strings.Join(parts, " ")
}

func quoteArgument(value string) string {
	if value != "" && !strings.ContainsAny(value, " \t\r\n'\"\\$`;&|<>()*?[]{}!#~") {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func writePayload(writer io.Writer, payload []byte) error {
	if writer == nil {
		return errors.New("writer is nil")
	}
	written, err := writer.Write(payload)
	if err != nil {
		return err
	}
	if written != len(payload) {
		return io.ErrShortWrite
	}
	return nil
}
