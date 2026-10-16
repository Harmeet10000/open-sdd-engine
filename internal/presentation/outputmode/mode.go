// Package outputmode resolves the presentation mode for a CLI invocation.
//
// Resolution is deliberately independent from terminal libraries. Callers
// observe the environment at the executable boundary, construct a Request,
// and pass the resulting Mode to a renderer. This keeps terminal heuristics
// out of the engine and makes non-interactive behavior deterministic.
package outputmode

import (
	"errors"
	"fmt"
	"strings"
)

// Mode identifies one presentation adapter.
type Mode string

const (
	// ModeText renders concise, stable human-readable output.
	ModeText Mode = "text"
	// ModeJSON renders the versioned machine protocol.
	ModeJSON Mode = "json"
	// ModeTUI renders a stateful terminal interface.
	ModeTUI Mode = "tui"
)

// Interaction expresses an explicit caller preference. Auto leaves the
// decision to the machine and terminal observations in Request.
type Interaction uint8

const (
	InteractionAuto Interaction = iota
	InteractionDisabled
	InteractionEnabled
)

// Request contains facts already observed by the executable adapter.
//
// ExplicitOutput has highest precedence. Interaction comes next. Machine and
// redirected-stream facts then suppress interactive rendering. Terminal
// capability is consulted only when no stronger signal exists.
type Request struct {
	ExplicitOutput string
	Interaction    Interaction
	MachineCaller  bool
	StdinTerminal  bool
	StdoutTerminal bool
}

// Resolve selects a presentation mode without reading process-global state.
func Resolve(request Request) (Mode, error) {
	if request.ExplicitOutput != "" {
		mode, err := parseExplicit(request.ExplicitOutput)
		if err != nil {
			return "", err
		}
		if mode == ModeTUI {
			if request.Interaction == InteractionDisabled {
				return "", errors.New("TUI output conflicts with non-interactive mode")
			}
			if request.MachineCaller || !request.StdinTerminal || !request.StdoutTerminal {
				return "", errors.New("TUI output requires terminal input and output from a human caller")
			}
		}
		return mode, nil
	}

	switch request.Interaction {
	case InteractionDisabled:
		if request.MachineCaller {
			return ModeJSON, nil
		}
		return ModeText, nil
	case InteractionEnabled:
		if request.MachineCaller || !request.StdinTerminal || !request.StdoutTerminal {
			return "", errors.New("interactive output requires terminal input and output from a human caller")
		}
		return ModeTUI, nil
	case InteractionAuto:
		// Continue with observations below.
	default:
		return "", fmt.Errorf("unknown interaction preference %d", request.Interaction)
	}

	if request.MachineCaller {
		return ModeJSON, nil
	}
	if !request.StdinTerminal || !request.StdoutTerminal {
		return ModeText, nil
	}
	return ModeText, nil
}

func parseExplicit(value string) (Mode, error) {
	mode := Mode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case ModeJSON, ModeText:
		return mode, nil
	case ModeTUI:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported output mode %q: expected json, text, or tui", value)
	}
}
