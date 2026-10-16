# ADR-0005: Use Cobra for CLI parsing, without Viper

- Status: Accepted
- Date: 2026-10-17
- Scope: `sdd` command tree, flag parsing, help generation (FND-012)

## Context

Foundation needs subcommand dispatch (`version`, `doctor`, `status`, `start`,
`help`), per-command flags, `--help` generation, and usage errors with stable
exit codes. The engine and renderers are already tested behind
`Engine.Execute`; the CLI layer must stay a thin parse/compose/translate
shell that never decides workflow policy.

## Decision

Use `github.com/spf13/cobra` pinned to `v1.10.2` for command dispatch and
flag parsing only:

- `SilenceUsage` and `SilenceErrors` are always set; all failures render
  through the engine outcome envelope, never through Cobra's error printer.
- `--help`/`-h`/`help` short-circuit to the foundation-owned help text before
  Cobra dispatches, so help output is byte-stable and workspace-independent.
- Unknown input is classified into `Engine.UnknownCommand`, never trusted as
  a command.
- `RunE` functions never return errors and never call `os.Exit`; they render
  an outcome and record its exit code for `main` to return.

Viper is explicitly NOT adopted. Environment-variable and config-file
layering would invent a precedence the specification leaves open (foundation
open question: env/flag interaction). Flags are the only configuration
source until that decision is recorded. No `viper.BindPFlag`, no env prefix,
no config file lookup.

## Why this dependency

- Command/flag parsing with `--help` generation is substantial,
  user-visible complexity a local parser would reproduce worse.
- Pure Go, no CGO; small transitive surface (pflag only).
- MIT-family license (Apache 2.0) with long maintenance evidence.
- The API permits full control of output streams (`SetOut`/`SetErr`) and
  error silencing, which the test fixtures require.

## Rejected alternatives

- Standard-library `flag` only: no subcommand tree; reimplementing dispatch,
  help, and usage errors by hand around four commands that will grow.
- Cobra + Viper now: invents env/config precedence the spec has not decided.
- Bubble Tea in this slice: no inspect/recovery/monitor screen exists yet;
  non-interactive behavior must be proven first.

## Consequences and controls

- `go.mod`/`go.sum` pin Cobra and pflag; CI uses `-mod=readonly`.
- CLI fixtures execute the command tree programmatically with buffer
  streams: TTY, redirected, JSON, malformed-input, and redaction cases.
- `govulncheck` remains a release signal for reachable vulnerabilities.

## Exit strategy

Replacement requires only an argv parser that produces the same command
values, the same help bytes, and the same exit codes through the existing
fixtures. No engine type, outcome, or protocol schema changes to swap it.
