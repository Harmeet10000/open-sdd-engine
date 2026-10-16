// Package help owns the static command overview text. It performs no
// parsing and assigns no flag semantics; option handling belongs to the CLI
// composition root. The text names only commands the engine implements, so
// rendering help can never advertise a capability that does not exist.
package help

// Text is the complete help overview. It contains no ANSI escapes, no
// terminal-width formatting, and no prose that a machine must parse.
const Text = `sdd is the deterministic engine for spec-driven development.

Usage:
  sdd [command] [flags]

Commands:
  version     Print build metadata and the protocol schema version
  doctor      Check executable, workspace, schema, and policy status
  status      Report zero, one, or a selected work item (read-only)
  start       Preview a work-creation plan without applying it (dry-run only)

Flags:
  --json      Emit exactly one protocol envelope on stdout
  --root      Inspect only the given workspace root (no parent fallback)

Run "sdd [command] --help" for command details.
`

// VersionHelp details the version command.
const VersionHelp = `Print build metadata and the protocol schema version.

Usage:
  sdd version [flags]

Works from any directory without a workspace, network, or configuration.
`

// DoctorHelp details the doctor command.
const DoctorHelp = `Check executable, workspace, schema, and policy status.

Usage:
  sdd doctor [flags]

Reports executable metadata, current-directory access, workspace discovery,
canonical layout, schema compatibility, and policy source status. Exits 3
when no workspace, the workspace is malformed, or required policy is
unavailable. Never mutates.
`

// StatusHelp details the status command.
const StatusHelp = `Report zero, one, or a selected work item.

Usage:
  sdd status [flags]

Flags:
  --work      Select a work item by immutable ID

With several active work items and no --work flag, status reports the
candidates instead of guessing. Read-only.
`

// StartHelp details the start command.
const StartHelp = `Preview a work-creation plan without applying it.

Usage:
  sdd start --dry-run "intent" [flags]

Flags:
  --dry-run   Required. Only preview mode is implemented in foundation
  --workflow  Journey the new work item belongs to

The preview never writes. Durable application arrives with workspace
transactions.
`

// Topic returns the help text for a command name, or false when the name is
// not a documented topic.
func Topic(name string) (string, bool) {
	switch name {
	case "version":
		return VersionHelp, true
	case "doctor":
		return DoctorHelp, true
	case "status":
		return StatusHelp, true
	case "start":
		return StartHelp, true
	default:
		return "", false
	}
}
