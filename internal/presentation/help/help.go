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
