# Open SDD Engine

This repository is the product research, protocol design, and migration workspace for `sdd`: an agent-neutral CLI that turns a software-change intent into a deterministic, reviewable path from specification to verified completion.

The production CLI will be implemented in Go and distributed as signed native binaries. The existing Bun/TypeScript implementation is a temporary prototype and parity oracle; it is not the target runtime and should not receive new product features.

## Product shape

The ordinary user journey is intentionally small:

```text
sdd init
sdd start "intent"
sdd next
sdd status
sdd check
sdd finish
```

The engine hides workflow selection, artifact graphs, snapshot revisions, transaction recovery, validation profiles, context budgets, and host-adapter details. Human output is concise and actionable; `--json` is a stable machine interface for any agent.

Two guided presets use the same commands:

```text
sdd start "login fails after token refresh" --workflow bug-fix
sdd start "add CSV export" --workflow quick-spec
```

Bug Fix structures report, reproduction, investigation, diagnosis, regression evidence, resolution, and verification. Quick Spec resolves material clarifications and then produces a linked requirements/design/tasks draft bundle. Presets never lower the independently selected risk rigor.

## Current authority

- [CLI product plan](docs/cli-product-plan.md) — consolidated product decisions and delivery plan.
- [Go project structure](docs/architecture/go-project-structure.md) — target module boundaries and dependency rules.
- [Guided workflows](docs/architecture/guided-workflows.md) — Bug Fix and Quick Spec behavior, escalation, diagnostics, and conformance fixtures.
- [Markdown corpus audit](docs/research/markdown-corpus-audit.md) — disposition of the original research files.
- [OKF storage research](docs/research/okf-storage-and-protocol-specs.md) — Markdown authority versus SQLite/Obsidian/Turso projections.
- [Kiro input pack](docs/kiro-input/README.md) — bounded files and prompts for generating implementation specs.

Historical notes, vendored framework material, and research reports are evidence, not product authority. Where they conflict, follow accepted ADRs and the current product plan.

## Engineering references

- [Go library selection](docs/research/go-library-selection-2026-10-15.md) — primary-source evaluation of Cobra, Bubble Tea, parser, terminal, test, SQLite, and release dependencies. It is research evidence until an owning spec or ADR admits a dependency.
- [Project-local Go skills](.agents/skills/README.md) — curated agent guidance copied with source provenance; these files do not add runtime dependencies.

## Draft the implementation specs in Kiro

Start Kiro at the repository root so it discovers `.kiro/steering/`. Then follow [the Kiro input instructions](docs/kiro-input/README.md). Create bounded specs in dependency order rather than one repository-wide mega-spec. The first spec is `sdd-engine-foundation` and begins with requirements only.

## Target architecture

The Go executable uses one deep application seam—`Engine.Execute(ctx, Command) Outcome`—with capability-oriented internal packages and ports only at real filesystem, process, Git, clock/ID, and later knowledge-index boundaries. Public compatibility surfaces are the CLI, versioned JSON schemas, and the `.sdd/` workspace format; internal Go packages are not a public SDK in v1.

Do not create the full target folder tree up front. Each vertical slice should add only the packages it needs and remain independently testable through the engine boundary.

## Prototype status

The TypeScript workspace is useful for extracting black-box behavior fixtures. Its advertised typecheck is not currently green in this checkout: dependency resolution is incomplete and the source references a missing Cognee module. Repair only what is needed to capture deterministic fixtures, then port read-only behavior before mutation and recovery behavior.

## License

MIT
