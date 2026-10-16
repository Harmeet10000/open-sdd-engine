# ADR 0002: Quality gates and optional CLI presentation adapters

- Status: accepted
- Date: 2026-10-16
- Owners: Open SDD Engine maintainers
- Related: `docs/research/go-library-selection-2026-10-15.md`,
  `docs/architecture/go-project-structure.md`

## Context

Open SDD Engine is intended to become a company-wide protocol tool used by
humans, CI, and autonomous agents. Its contribution system therefore needs to
detect contract drift, data races, platform assumptions, dependency risk, and
irreproducible repository state before merge. Its terminal UX must also remain
pleasant without allowing a TUI framework to become a second application
engine.

The repository contains a frozen TypeScript prototype and the replacement Go
engine. The Go module declares Go 1.24 as its minimum supported version. A
top-level `vendor/` directory contains research evidence, not Go vendored
dependencies, so every Go command must use explicit module mode.

## Decision

### Local and hosted verification share one contract

The Makefile is the executable source of truth for checks. GitHub Actions
parallelizes those targets but does not invent different validation semantics.
The local `make ci` target includes:

1. `gofmt` cleanliness;
2. `go mod tidy -diff` and checksum verification;
3. package-graph loading and `go vet`;
4. shuffled tests, race tests, and a 70% repository coverage floor;
5. golangci-lint with an explicit repository policy;
6. gosec source analysis; and
7. actionlint validation of workflow files.

`make vuln` remains separate because govulncheck consults a network-maintained
database. CI and a weekly scheduled workflow always execute it.

Tool versions are pinned in the Makefile and workflows. Third-party GitHub
Actions are pinned to full commit SHAs. Workflows receive read-only repository
permissions unless a future job documents why it needs more. Linux jobs pin
the Ubuntu 24.04 runner generation so a moving `ubuntu-latest` alias cannot
silently change the protected build environment.

### CI validates compatibility, not only the maintainer's machine

Pull requests run tests on the minimum Go version and the current Go version.
They also run current-Go tests on Linux, macOS, and Windows, with the race
detector on Linux. A single `Required checks` aggregation job is the stable
branch-protection interface; the internal matrix can evolve without silently
removing protection.

### Commit policy is repository-native

`internal/tools/commitlint` validates Conventional Commit subjects for both
local commits and pull-request titles. It invokes Git with an argument vector,
never a shell command. This avoids introducing Node solely for commit policy
and ensures local and hosted validation cannot disagree because of two parser
implementations.

### Output-mode policy is a pure boundary decision

`internal/presentation/outputmode` resolves output mode from an explicit value,
an interaction preference, a known-machine signal, and pre-observed stream
capabilities. It reads no process-global state and imports no terminal or TUI
framework.

The precedence is:

```text
explicit --output value
    > explicit interaction preference
    > known machine caller or redirected stream
    > terminal capability heuristic
```

Explicit JSON and text always win. TUI selection is additionally constrained:
non-interactive mode, a known machine caller, or redirected input/output cannot
start a TUI, even if it requests one. This is capability validation, not a
lower-precedence output decision.

The default on a human TTY remains concise text. A TUI starts only when a user
explicitly requests an interactive experience. This avoids changing behavior
when a process happens to gain or lose a pseudo-terminal.

### Runtime dependency timing is capability-driven

- Cobra is added with the first real `cmd/sdd` command tree.
- `x/term` is added when the executable observes terminal capabilities.
- Bubble Tea, Bubbles, and Lip Gloss are added with the first stateful inspect,
  recovery, evidence, or bounded-execution screen.
- Goldmark and YAML v4 are added by artifact/workflow parsing specs.
- SQLite remains a disposable projection and is deferred to the context-index
  spec.
- GoReleaser is deferred to release hardening.

The TUI will receive immutable view models and may submit typed commands. It
will not read workspace files, evaluate workflow policy, validate evidence,
construct recovery actions, or determine exit codes.

## Consequences

- A contributor can reproduce merge gates with one local command.
- Platform and minimum-version failures appear before release work begins.
- The commit policy adds no runtime or Node dependency.
- Bubble Tea remains an optional presentation adapter rather than a foundation.
- Some checks are intentionally expensive; CI runs them concurrently while
  local developers can run focused targets during iteration.
- The coverage floor is a ratchet baseline, not evidence that critical paths
  are adequately tested. Capability specs may require stricter package-level,
  property, fuzz, fault-injection, or platform evidence.

## Revisit triggers

Revisit this decision when the minimum Go version changes, a release pipeline
needs signed provenance, the first real TUI is specified, or the aggregate
coverage metric begins hiding an under-tested safety-critical capability.
