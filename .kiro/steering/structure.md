---
inclusion: always
---

# Project structure and dependency direction

The repository uses capability-oriented Go packages inside a hexagonal I/O shape.

## Stable shape

```text
cmd/sdd/                 composition, CLI parsing, rendering, exit only
internal/engine/         one deep Execute(ctx, Command) Outcome seam
internal/workspace/      discovery, snapshots, locks, transactions, journal
internal/workflow/       journeys, rigor, presets, graph, actions, routing
internal/artifact/       typed Markdown, IDs, deltas, digests, freshness
internal/validation/     diagnostics, policy, traceability, convergence
internal/completion/     evidence gates, receipts, promotion, archive
internal/contextpack/    authority-filtered progressive context
internal/integration/    thin host adapters and managed files
internal/effect/         effect-plan application and internal ports
internal/platform/       OS/filesystem/process/Git/SQLite adapters
internal/presentation/   CLI plus human/JSON rendering
internal/testkit/        hermetic workspaces and fault injection
protocol/schema/         normative public JSON Schemas
workflow/builtin/        versioned built-in workflow resources
migrations/              workspace and legacy-format migrations
testdata/                golden repos, malformed docs, fault scenarios
```

## Rules

1. Business decisions enter through `internal/engine`; CLI handlers do not decide workflow behavior.
2. Capability packages do not import platform implementations.
3. Platform adapters execute authorized effects but do not decide policy.
4. A consuming capability owns its port; do not create a global interfaces package.
5. Introduce an interface only for a real alternative, normally production plus deterministic test adapter.
6. No `pkg/`, `utils`, `common`, `helpers`, generic `services`, or generic `repositories` in v1.
7. Agent integrations cannot alter canonical workflow semantics.
8. SQLite remains confined to the optional context-index adapter.
9. Tests primarily cross `Engine.Execute`; package tests cover pure laws and platform faults.
10. Do not create empty packages merely to make the tree resemble the target architecture.
11. Within capability packages, prefer explicit decode → normalize → index/compile → evaluate → plan passes over pointer-linked objects calling one another.
12. Persist stable IDs; dense indexes and hot/cold layouts are runtime optimizations admitted only with deterministic tests and, for specialized layouts, profiling evidence.
13. Keep journey topology, rigor policy, and artifact/validation profiles orthogonal; named presets compose them without weakening the risk-derived floor.

The complete target tree and rationale are in `docs/architecture/go-project-structure.md`. Load that document when drafting architecture or tasks, not for every ordinary interaction.
