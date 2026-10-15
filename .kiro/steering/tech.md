---
inclusion: always
---

# Technology and runtime constraints

## Production stack

- Language: Go.
- Deliverable: signed native `sdd` binaries for explicitly supported OS/architecture pairs.
- Public compatibility surfaces: CLI behavior, JSON protocol, and `.sdd/` workspace format.
- Internal Go packages are not a public SDK in v1.
- Built-in schemas/templates are versioned resources embedded in the binary and exportable for inspection.

## Migration constraint

The existing Bun/TypeScript code is a temporary executable oracle. Freeze feature development there, capture hermetic fixtures, port behavior through language-neutral JSON contracts, and remove the TypeScript runtime from the shipped product after Go parity. Do not build a permanent Go/TypeScript hybrid.

## Persistence

- Markdown/YAML plus an append-only journal are authoritative.
- Materialized state and SQLite/FTS indexes are rebuildable projections.
- SQLite is introduced only after the core workflow is stable; its Go driver requires a separate pure-Go versus CGO build decision.
- Obsidian is an optional authoring/view adapter, not authority or access control.
- Hosted libSQL/Turso is deferred and may only project accepted repository knowledge.

## Engineering constraints

- Prefer the Go standard library.
- Every dependency must satisfy a concrete missing capability and record upgrade/security/deletion cost.
- Process execution uses executable plus argument array, never shell strings.
- Unix process groups and Windows Job Objects are distinct process-tree adapters.
- Atomic visibility and crash durability are separate guarantees and must be tested per supported platform/filesystem.
- Network access is off by default.
- Generated code is allowed only for public protocol/schema types with checked-in reproducibility tests.

## Required CI signals

- formatting and static analysis;
- unit, interface, integration, race, and targeted fuzz tests;
- native builds and empty-directory artifact smoke tests;
- protocol/schema conformance;
- vulnerability analysis and release SBOM;
- Linux, macOS, and Windows path/process/recovery fixtures for every claimed platform.
