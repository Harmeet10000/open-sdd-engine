# Go project structure for `sdd`

**Status:** Proposed architecture input  
**Decision source:** `docs/cli-product-plan.md`, revision 6  
**Audience:** maintainers, Kiro spec authors, reviewers, and agent integrators  

## 1. Architecture choice

Use a hybrid of:

- **one deep application module** for the supported in-process seam;
- **capability-oriented internal modules** for locality of business rules;
- **ports and adapters only at real I/O seams**;
- **language-neutral schemas and files** for public compatibility.

Do not use frontend Feature-Sliced Design layers such as `pages`, `widgets`, and `features`. Do not organize the core as global `controllers/`, `services/`, `repositories/`, and `models/` folders. Both shapes scatter one workflow behavior across many folders.

The governing dependency direction is:

```text
CLI / JSON request
       |
       v
  Engine.Execute                 one supported deep interface
       |
       +--> workspace            selection, snapshots, transactions, journal
       +--> workflow             profiles, artifact/action graph, routing
       +--> artifact             typed Markdown, identity, digests, freshness
       +--> validation           diagnostics, policy, traceability
       +--> completion           receipts, DONE gate, promotion, archive
       +--> contextpack          provenance-aware progressive disclosure
       +--> integration          agent renderers and managed files
       |
       v
 effect plans / internal ports
       |
       +--> local filesystem adapter
       +--> OS process-supervisor adapter
       +--> Git CLI adapter
       +--> system clock/ID adapter
       +--> SQLite projection adapter (later)
```

The modules above are not independently deployed services. They are packages inside one executable and may collaborate directly through private types. An interface is introduced only where a real alternate implementation exists, normally production plus a deterministic test adapter.

## 2. Recommended repository tree

```text
.
├── cmd/
│   └── sdd/
│       ├── main.go                 composition root only
│       ├── root.go                 global flags and command dispatch
│       └── version.go              version/build display
│
├── internal/
│   ├── engine/
│   │   ├── engine.go               Engine construction
│   │   ├── execute.go              Execute(ctx, Command) Outcome
│   │   ├── command.go              generated/closed command union helpers
│   │   ├── outcome.go              stable result and diagnostic envelope
│   │   └── execute_test.go         interface-level behavior tests
│   │
│   ├── workspace/
│   │   ├── discover.go             canonical root and workspace identity
│   │   ├── select.go               explicit work selection and revisions
│   │   ├── snapshot.go             immutable input snapshot
│   │   ├── plan.go                 MutationPlan and preconditions
│   │   ├── transaction.go          prepare/apply/recover protocol
│   │   ├── journal.go              append-only authoritative events
│   │   ├── projection.go           rebuildable materialized state
│   │   ├── lock.go                 lock policy and owner metadata
│   │   └── migrate.go              workspace-version migrations
│   │
│   ├── workflow/
│   │   ├── schema.go               workflow schema model/compiler
│   │   ├── graph.go                dependency/freshness graph
│   │   ├── action.go               next-action calculation
│   │   ├── route.go                evidence-based risk routing
│   │   ├── initiative.go           shallow slice DAG
│   │   ├── cursor.go               durable compound-workflow progress
│   │   ├── journey.go              general/bug-fix/quick-spec topology
│   │   ├── rigor.go                quick/standard/critical policy floor
│   │   └── preset.go               named maintained compositions
│   │
│   ├── artifact/
│   │   ├── document.go             typed Markdown document model
│   │   ├── frontmatter.go          typed metadata and authority
│   │   ├── requirement.go          requirements and scenarios
│   │   ├── decision.go             decisions and complexity rationale
│   │   ├── task.go                 task model and dependencies
│   │   ├── evidence.go             evidence and source references
│   │   ├── delta.go                structural delta operations
│   │   ├── digest.go               blob/semantic/derivation digests
│   │   ├── parse.go                Markdown AST to typed artifacts
│   │   └── render.go               typed artifacts to canonical Markdown
│   │
│   ├── validation/
│   │   ├── validator.go            validator registry and execution order
│   │   ├── diagnostic.go           stable codes, severity, source ranges
│   │   ├── policy.go               organization/project policy merge
│   │   ├── traceability.go         requirement/design/task/evidence graph
│   │   ├── convergence.go          read-only implementation comparison
│   │   ├── protocol.go             protocol-spec profile checks
│   │   └── profiles.go             validation profiles by risk
│   │
│   ├── completion/
│   │   ├── gate.go                 policy-driven completion decision
│   │   ├── receipt.go              snapshot-bound machine receipt
│   │   ├── done.go                 human completion record
│   │   ├── evidence.go             result and verification matching
│   │   ├── promotion.go            three-way accepted-spec promotion
│   │   └── archive.go              immutable historical lifecycle
│   │
│   ├── contextpack/
│   │   ├── select.go               smallest sufficient evidence selection
│   │   ├── authority.go            authority/status/time filtering
│   │   ├── provenance.go           provenance and trust labels
│   │   ├── budget.go               deterministic size budgets
│   │   └── index.go                projection interface, not authority
│   │
│   ├── integration/
│   │   ├── manifest.go             managed-file ownership and hashes
│   │   ├── capability.go           host capability handshake
│   │   ├── generic.go              AGENTS.md/SKILL.md adapter
│   │   ├── codex.go                thin Codex renderer
│   │   ├── claude.go               thin Claude renderer
│   │   └── kiro.go                 thin Kiro renderer/import adapter
│   │
│   ├── effect/
│   │   ├── apply.go                authorize and apply MutationPlan
│   │   ├── filesystem.go           internal transaction port
│   │   ├── process.go              internal process-supervisor port
│   │   ├── git.go                  internal Git observation/mutation port
│   │   ├── clock.go                clock and ID port
│   │   └── fake/                   deterministic test adapters
│   │
│   ├── platform/
│   │   ├── filesystem/
│   │   │   ├── local.go            atomic visibility/durability implementation
│   │   │   ├── lock_unix.go        Unix lock implementation
│   │   │   └── lock_windows.go     Windows lock implementation
│   │   ├── process/
│   │   │   ├── supervisor.go       shared policy and output caps
│   │   │   ├── tree_unix.go        process groups/signals
│   │   │   └── tree_windows.go     Job Objects/Windows termination
│   │   ├── gitcli/
│   │   │   └── git.go              argv-only Git adapter
│   │   ├── system/
│   │   │   └── clock.go            wall/monotonic time and IDs
│   │   └── sqlite/
│   │       └── index.go            optional Phase 5 read model
│   │
│   ├── presentation/
│   │   ├── cli/
│   │   │   ├── init.go
│   │   │   ├── start.go
│   │   │   ├── next.go
│   │   │   ├── status.go
│   │   │   ├── check.go
│   │   │   ├── finish.go
│   │   │   └── doctor.go
│   │   ├── human/
│   │   │   └── render.go
│   │   └── jsonout/
│   │       └── render.go
│   │
│   └── testkit/
│       ├── workspace.go            hermetic workspace builder
│       ├── fault.go                write/process failure injection
│       ├── golden.go               fixture normalization and comparison
│       └── clock.go                deterministic time/ID support
│
├── protocol/
│   ├── schema/
│   │   ├── command.schema.json
│   │   ├── outcome.schema.json
│   │   ├── diagnostic.schema.json
│   │   ├── mutation-plan.schema.json
│   │   ├── receipt.schema.json
│   │   └── completion.schema.json
│   ├── fixtures/
│   │   ├── valid/
│   │   └── invalid/
│   └── VERSION
│
├── workflow/
│   └── builtin/
│       ├── quick-change/
│       ├── standard-change/
│       ├── critical-change/
│       ├── bug-fix/
│       ├── quick-spec/
│       └── protocol-change/
│
├── migrations/
│   ├── workspace/
│   └── import/
│       ├── kiro/
│       ├── openspec/
│       └── spec-kit/
│
├── testdata/
│   ├── golden-repositories/
│   ├── markdown/
│   ├── legacy-typescript/
│   └── fault-scenarios/
│
├── docs/
│   ├── tutorials/
│   ├── how-to/
│   ├── reference/                  generated contract reference
│   ├── explanation/
│   ├── architecture/               ADRs and stable architecture
│   └── research/                   dated, non-normative evidence
│
├── .kiro/
│   └── steering/                   Kiro-only authoring context, not product authority
├── go.mod
├── go.sum
├── Makefile                        thin aliases only; CI runs underlying commands
├── AGENTS.md
└── README.md
```

The tree is a target shape, not an instruction to create every empty directory. Add a package only when a vertical slice needs it. Empty architecture is not progress.

## 3. Data-oriented implementation inside the packages

The package tree establishes ownership; it does not prescribe an object graph. Inside those packages, model the engine as explicit transformations over snapshots and compiled records:

```text
workspace bytes
  -> decoded artifact/workflow records
  -> normalized, validated records
  -> immutable snapshot with ordered slices and ID indexes
  -> compiled graph with runtime-local dense indexes
  -> pure routing/validation/completion passes
  -> diagnostics, next actions, or MutationPlan
  -> authorized platform effects
  -> journal events and rebuildable projection
```

Use stable string/UUID identities in files and protocols. A compiled snapshot may map those identities to dense integer positions for deterministic graph traversal, but those positions are ephemeral and must never leak into persisted identity. Keep frequently traversed graph metadata separate from cold Markdown prose where that makes the pass clearer or profiling proves it material. Prefer array-of-structs and ordinary slices first; ECS, struct-of-arrays, arenas, pooling, unsafe code, or a custom binary representation require representative benchmark/profile evidence and an explicit removal criterion.

Data-driven workflow files may select nodes, transitions, validators, and maintained engine capabilities. They cannot introduce code, shell execution, network authority, or an unregistered effect. The workflow compiler is the trust boundary: it rejects unknown critical fields, duplicate identifiers, bad references, invalid cycles, unsafe paths, and incompatible versions before any evaluator sees the graph.

## 4. Dependency rules

1. `cmd/sdd` imports presentation and composition code only. It contains no workflow decision.
2. Presentation imports `internal/engine`; the engine never imports presentation.
3. Capability modules may use shared protocol/value types, but they do not import platform implementations.
4. A capability owns the internal port it consumes. Platform adapters implement those ports at the composition root.
5. `internal/platform` cannot decide product policy. It reports facts and performs already-authorized effects.
6. SQLite imports are confined to the context index adapter. Deleting the database must not remove authoritative work.
7. Agent integrations render/parse host files and capabilities. They cannot change workflow semantics.
8. Legacy importers produce canonical proposals; they never create a second source of truth.
9. No `pkg/` directory in v1. Public packages require two proven external consumers and their own compatibility decision.
10. No `utils`, `common`, `helpers`, or generic `service` package. Put knowledge with the capability that owns it; extract only after real duplication reveals a stable concept.

## 5. Real ports, not interface theater

The likely real internal ports are:

- **Workspace transaction:** local durable filesystem in production; in-memory/fault-injecting adapter in tests.
- **Process supervisor:** Unix/Windows implementations plus deterministic fake runner.
- **Git:** argv-only Git implementation plus fake repository for plan-level tests.
- **Clock/ID:** system implementation plus fixed deterministic test implementation.
- **Knowledge index:** SQLite implementation plus in-memory query projection once the Phase 5 need is proven.

Do not add ports for Markdown parsers, graph algorithms, policy evaluation, or digest calculation. Those are in-process logic and can be tested directly. Do not expose a generic filesystem interface with every `os` operation; expose the higher-leverage workspace transaction behavior the engine needs.

## 6. Slice shape

A product slice cuts through the modules without inventing another public layer. For example, `finish` uses:

```text
presentation/cli/finish.go
  -> engine.Execute(FinishCommand)
  -> workspace snapshot
  -> validation completion profile
  -> completion gate + receipt + accepted-spec promotion
  -> effect MutationPlan authorization/apply
  -> workspace journal/projection
  -> Outcome rendered by human/jsonout
```

The behavior is tested first through `Engine.Execute`. Focused package tests exist for parsers, graph laws, and platform fault behavior, but callers never orchestrate those internals themselves.

## 7. Testing placement

- Keep package-local tests beside their implementation.
- Put black-box engine fixtures under `testdata/` and drive them through `Engine.Execute`.
- Use build-tagged or OS-specific platform tests for Windows and Unix behavior.
- Run race tests for concurrent claims, locks, projection rebuilds, and cancellation bookkeeping.
- Fuzz Markdown/frontmatter parsing, path handling, protocol decoding, delta application, and recovery manifests.
- Run destructive fault tests only in directories created by the test harness.
- Compare normalized Go outcomes with frozen TypeScript fixtures only during migration; delete the cross-language harness afterward.

## 8. What remains deliberately outside the executable

- LLM/model selection and invocation;
- general-purpose agent scheduling;
- web research providers;
- remote extension marketplace;
- issue tracker synchronization;
- hosted knowledge service;
- organization policy server;
- arbitrary in-process plugins.

Those concerns integrate through files, the JSON protocol, explicit subprocesses, or later governed services. Keeping them outside is what lets the engine interface remain deep.

## For the chosen ones

Hexagonal architecture is often weakened by putting an interface in front of every package. That reverses its value: callers learn more names while behavior stays shallow. Here, the valuable hexagon is asymmetric. There is one broad inward command seam and only a few outward effect seams. Most domain packages are intentionally concrete. The architecture earns its keep when replacing the local filesystem with a fault-injecting adapter tests the entire transaction protocol without forcing artifact, workflow, and completion callers to learn filesystem mechanics.
