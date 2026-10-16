// Package engine owns the deep application seam and its closed protocol
// values. Presentation and process-exit policy remain outside this package.
package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const SchemaVersion = "1.0"

var ErrInvalidProtocolValue = errors.New("invalid protocol value")

// CommandKind is the closed foundation command discriminator.
type CommandKind string

const (
	CommandStatus  CommandKind = "status"
	CommandVersion CommandKind = "version"
	CommandDoctor  CommandKind = "doctor"
	CommandStart   CommandKind = "start"
	CommandUnknown CommandKind = "unknown"
)

// Command is an immutable, closed command value. Constructors are the only
// way to create a valid command; accessors never expose mutable storage.
type Command struct {
	kind     CommandKind
	workID   string
	root     string
	intent   string
	workflow string
	dryRun   bool
}

func NewStatusCommand(workID, root string) Command {
	return Command{kind: CommandStatus, workID: workID, root: root}
}

func NewVersionCommand() Command {
	return Command{kind: CommandVersion}
}

func NewDoctorCommand(root string) Command {
	return Command{kind: CommandDoctor, root: root}
}

func NewStartDryRunCommand(intent, workflow, root string) (Command, error) {
	if strings.TrimSpace(intent) == "" {
		return Command{}, fmt.Errorf("%w: start intent is required", ErrInvalidProtocolValue)
	}
	return Command{
		kind:     CommandStart,
		intent:   intent,
		workflow: workflow,
		root:     root,
		dryRun:   true,
	}, nil
}

func (c Command) Kind() CommandKind { return c.kind }
func (c Command) WorkID() string    { return c.workID }
func (c Command) Root() string      { return c.root }
func (c Command) Intent() string    { return c.intent }
func (c Command) Workflow() string  { return c.workflow }
func (c Command) DryRun() bool      { return c.dryRun }

func (c Command) MarshalJSON() ([]byte, error) {
	switch c.kind {
	case CommandStatus:
		return json.Marshal(struct {
			Type   CommandKind `json:"type"`
			WorkID string      `json:"workId,omitempty"`
			Root   string      `json:"root,omitempty"`
		}{Type: c.kind, WorkID: c.workID, Root: c.root})
	case CommandVersion:
		return json.Marshal(struct {
			Type CommandKind `json:"type"`
		}{Type: c.kind})
	case CommandDoctor:
		return json.Marshal(struct {
			Type CommandKind `json:"type"`
			Root string      `json:"root,omitempty"`
		}{Type: c.kind, Root: c.root})
	case CommandStart:
		return json.Marshal(struct {
			Type     CommandKind `json:"type"`
			DryRun   bool        `json:"dryRun"`
			Intent   string      `json:"intent"`
			Workflow string      `json:"workflow,omitempty"`
			Root     string      `json:"root,omitempty"`
		}{
			Type:     c.kind,
			DryRun:   c.dryRun,
			Intent:   c.intent,
			Workflow: c.workflow,
			Root:     c.root,
		})
	default:
		return nil, fmt.Errorf("%w: unsupported command kind %q", ErrInvalidProtocolValue, c.kind)
	}
}
