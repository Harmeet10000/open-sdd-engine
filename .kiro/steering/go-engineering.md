---
inclusion: auto
name: sdd-go-engineering
description: Apply when specifying, designing, implementing, or reviewing Go packages, data layouts and transformations, workflow configuration, errors, concurrency, context cancellation, interfaces, dependencies, performance, or tests.
---

# Go engineering policy: data-driven/data-oriented design and railway-style flow

Within this project, **DDP means data-driven programming with data-oriented design principles**, in the game-engine sense: understand the data, its lifetime, ownership, access patterns, transformations, and scale before choosing abstractions. DDP does not mean Domain-Driven Design here.

Railway-Oriented Programming (ROP) is used as a reasoning model for explicit success/failure composition. It is not a request to reproduce F# discriminated unions or install a generic monad framework in Go.

Data-driven and data-oriented are related but not identical:

- **Data-driven behavior** moves supported variability into validated, versioned data such as workflow graphs and policy profiles.
- **Data-oriented implementation** shapes in-memory and on-disk representations around the transformations the engine actually performs, avoiding pointer-rich object graphs and cache-unfriendly traversal when evidence shows it matters.

Correctness, determinism, and clarity come before micro-optimization. This CLI is I/O-heavy, not a frame-budgeted game loop; use data-oriented thinking immediately, but apply low-level layout optimization only after representative benchmarks and profiles.

## Stable vocabulary and plain data

Use these product terms consistently in code, schemas, tests, diagnostics, and documentation:

- workspace;
- work item;
- workflow, journey, rigor, artifact profile, and preset;
- artifact and artifact revision;
- action, claim, and lease;
- snapshot;
- diagnostic;
- mutation plan and precondition;
- transaction and journal event;
- evidence, verification, and completion receipt;
- context pack and provenance.

Do not replace these terms with generic `manager`, `service`, `processor`, `handler`, `record`, `item`, or `data` names. A type should reveal the data it represents or the transformation it performs.

## Capability boundaries around transformations

- Organize internal packages by coherent capability: `workspace`, `workflow`, `artifact`, `validation`, `completion`, and `contextpack`.
- Keep closely changing behavior together even if it includes parsing, policy, and orchestration. Do not scatter one capability across global controller/service/repository/model layers.
- `engine` coordinates a use case through one deep seam. It does not become a bag of business logic or a service locator.
- A package exposes the smallest behavior callers need and hides representations, storage layout, and algorithms.
- Preserve invariants at decode, normalize, compile, and mutation boundaries. A plain data model does not mean invalid states may flow unchecked.
- Prefer immutable snapshot-like inputs and returned results for decisions. Confine mutable state to explicit builders, indexes, and transaction/application scopes.
- Use stable IDs for identity and content/revision digests for concurrency. Do not confuse a mutable display slug with identity.
- Do not manufacture object-oriented or enterprise-DDD artifacts. There is no default `entities/`, `valueobjects/`, `aggregates/`, `domainservices/`, or `repositories/` tree.
- Name packages and functions for the data and operation: `workflow.Compile`, `validation.Run`, `completion.Evaluate`, or `artifact.Normalize`, not `WorkflowManager.Process`.

## Data-oriented model for this CLI

Start by writing down the dominant transformations and their access patterns:

```text
bytes/files
  -> decoded records
  -> normalized records
  -> indexed snapshot
  -> compiled workflow/artifact graph
  -> diagnostics and next actions
  -> mutation plan
  -> journal events and projections
  -> Outcome
```

Then shape data for those passes:

- Represent a workspace read as an immutable `Snapshot` containing normalized records plus explicit indexes. Pure evaluation reads the snapshot; it does not reopen files opportunistically.
- Keep canonical ordered slices for deterministic traversal. Build maps from stable ID to slice index for lookup; do not rely on map iteration order.
- Persist stable external IDs. Dense integer indexes are permitted only inside a compiled snapshot and are rebuilt rather than stored as identity.
- Prefer value slices and compact enums/discriminators for high-count homogeneous records when that keeps ownership clear. Avoid one interface value or heap pointer per graph node by default.
- Separate **hot data** repeatedly traversed by routing/validation—IDs, kinds, dependency ranges, status, digests—from **cold data** such as full prose, source excerpts, and remediation text. Load or retain cold payloads only when the action needs them.
- Store graph edges in deterministic adjacency slices or compact ranges after compilation. Reject missing IDs, duplicates, cycles where prohibited, and out-of-range indexes before evaluation.
- Batch passes when they share the same input: parse once, normalize once, compute reusable digests/indexes once, then run pure evaluators over the compiled representation.
- Keep effect requests as plain, inspectable data. `MutationPlan` describes ordered operations and preconditions; platform adapters execute it only after authorization.
- Prefer explicit state-transition tables or compiled workflow data over webs of objects calling each other.
- Keep ownership obvious. If a function retains a slice, map, byte buffer, or AST, it copies it or documents transfer; callers must not be able to mutate an accepted snapshot through an alias.

### Performance restraint

- Do not introduce ECS, archetypes, sparse sets, struct-of-arrays, arenas, object pools, unsafe code, manual allocators, or custom binary formats because they are associated with game-engine DOD.
- First establish representative corpora and budgets: small repository, large monorepo, thousands of artifacts, deep dependency graph, malformed input, and cold/warm projection paths.
- Use Go benchmarks, CPU/heap profiles, allocation counts, and execution traces to identify the actual limiting pass.
- Choose array-of-structs for clarity by default. Consider struct-of-arrays only when a measured hot pass reads a small subset of fields across many records and the improvement is material.
- Reducing pointers can reduce GC scan work, but index-based representations trade type safety and update complexity for locality. Make that trade only with before/after evidence.
- Record the dataset, benchmark, profile, memory effect, and deletion/rollback criterion for every specialized layout.

The goal is not “maximum cache locality everywhere.” The goal is a small number of explicit, deterministic transformations over data shaped for the way this engine actually reads it.

## Real ports and adapters

- Define an interface in the package that consumes it, not beside its implementation.
- Introduce an interface only for a real variation or effect boundary: filesystem transaction, process supervision, Git observation/mutation, clock/ID, or a replaceable knowledge index.
- Production adapters return concrete types where callers do not need substitution.
- Do not create interfaces solely so a mocking framework can generate mocks. Prefer deterministic fakes that model the relevant behavior and faults.
- Keep interfaces narrow and behavioral. Avoid a generic filesystem interface that mirrors dozens of `os` functions.
- Platform adapters report facts and perform already-authorized effects; they do not decide workflow or policy.
- The dependency direction points inward from CLI/platform composition to the engine and capabilities. Domain packages never import CLI renderers or concrete platform adapters.

## Railway-style command flow in idiomatic Go

Model each use case as a visible sequence:

```text
decode -> normalize -> discover/snapshot -> decide -> plan
       -> authorize -> apply -> verify -> render outcome
```

Read-only commands stop before authorize/apply. Mutating commands produce an inspectable plan before applying effects.

Use ordinary Go signatures:

```go
func BuildPlan(ctx context.Context, snap Snapshot, cmd StartCommand) (MutationPlan, error)
```

Do not introduce `Result[T]`, `Either[L, R]`, fluent bind chains, exception emulation, or reflection-heavy pipeline frameworks unless a measured codebase problem proves that native Go is inadequate.

### Error-flow rules

- Return errors; do not encode an operational failure as a zero value, boolean, log line, or panic.
- Add actionable context with `%w` while preserving identity. Branch with `errors.Is` and `errors.As`, never substring matching.
- Expected, caller-actionable categories get stable sentinel or typed errors only when a caller can handle them differently.
- Translate internal errors once at the engine/presentation boundary into stable diagnostics and exit codes. Do not let package-specific error text become protocol.
- A function either handles an error or returns it with context. It does not log and return the same error, which would duplicate messages and leak layer details.
- Do not wrap an error if exposing its identity would accidentally make an implementation detail part of the package contract. Translation may be safer at a boundary.
- Use `errors.Join` only when multiple independent failures all matter, such as cleanup or a batch of independent checks. Do not use it to hide execution order or partial mutation.
- Short-circuit on failed preconditions, stale revisions, broken authority, persistence faults, and unsafe uncertainty.
- Aggregate independent validation diagnostics when seeing all findings is useful and does not make later checks misleading or unsafe.
- `panic` is for impossible programmer/invariant defects, not bad files, user input, process failure, cancellation, or incompatible versions. A top-level recovery guard may protect terminal cleanliness, but the defect remains a failure and is observable in debug evidence.

This is the useful part of ROP in Go: every stage has an explicit input, success value, and failure path; failure bypasses unsafe downstream effects; boundary translation stays centralized.

## Outcome versus error

Keep three concepts distinct:

1. **Successful outcome:** the command ran and the requested condition holds.
2. **Domain-negative outcome:** execution was valid but validation, verification, or a policy condition did not hold; return structured diagnostics and the designated non-zero exit category.
3. **Operational failure:** the engine could not reliably determine or apply the result because input, environment, conflict, external tool, persistence, cancellation, or an internal defect prevented it.

Never label “not run,” “timed out,” “cancelled,” “stale,” or “could not parse” as a failed verification result. Those are distinct states because retry and safety behavior differs.

## Context, cancellation, and process lifetime

- Pass `context.Context` as the first parameter of request-scoped functions. Do not store it in a struct.
- Propagate the caller context; do not replace it with `context.Background()` in a command path.
- Check cancellation before expensive computation and before every irreversible boundary, without sprinkling meaningless checks into tiny pure functions.
- A timeout is policy supplied by the caller or workflow; leaf packages do not invent hidden deadlines.
- Cancellation of a verification process includes descendant cleanup, pipe closure, bounded wait, and an outcome that distinguishes interruption from verification failure.
- Never start a goroutine without a named owner, termination condition, and wait/cleanup path.

## Data-driven behavior, code-enforced invariants

Workflow variability belongs in versioned declarative resources; safety invariants remain in Go.

- Built-in workflow data defines artifact/action nodes, dependencies, validators, profile parameters, and permitted transitions.
- A pure compiler parses and validates workflow data into a closed runtime graph before execution.
- Configuration is untrusted input. Reject unknown critical fields, duplicate stable IDs, cycles, missing references, impossible gates, unsafe paths, and incompatible schema versions.
- Declarative workflow files are not scripts. No embedded shell, arbitrary Go templates, dynamic code loading, implicit network access, or Turing-complete expression language.
- Data may choose among engine capabilities that code already authorizes; it cannot create a new capability or bypass policy.
- Defaults and merge precedence are explicit and testable. The resolved configuration can be inspected without mutation.
- Preserve the exact workflow version/digest used by a work item so later upgrades cannot silently change its meaning.

The rule is: **supported policy variability in data; integrity, authority, compilation, and effects in code**. A data-driven system is not one where arbitrary data can invent behavior.

## Go API and package practices

- Prefer the standard library until a dependency removes substantial, demonstrated complexity.
- Keep package names short, singular, and domain-specific. Avoid `util`, `common`, `base`, `helpers`, and stuttering names.
- Constructors validate mandatory dependencies and invariants. Do not create constructors for plain values whose useful zero value is already valid.
- Make the zero value useful where doing so does not weaken invariants.
- Accept interfaces at the narrow consuming boundary; return concrete types by default.
- Avoid global mutable state, hidden `init` registration, ambient current-directory mutation, and environment reads below the composition/configuration boundary.
- Copy or document ownership of slices, maps, byte buffers, and mutable ASTs crossing package boundaries. Do not leak aliases that let callers violate invariants.
- Keep serialization records separate from compiled runtime data when protocol/storage constraints or access patterns differ. Conversion occurs at a named boundary and is tested.
- Do not expose internal Go packages as the compatibility contract. The CLI, JSON Schemas, and `.sdd/` format are versioned surfaces.
- Keep build tags and OS-specific files at real platform seams. Shared domain logic must not branch on `runtime.GOOS`.
- All subprocess invocation uses executable plus argument slice, never a shell command string.

## Determinism and concurrency

- Sort map-derived output before rendering, hashing, snapshotting, or testing.
- Inject time and IDs where they affect persisted or compared output.
- Hash canonical representations with the representation type and schema version in the digest domain.
- Concurrency is earned by measured latency or independent work. Sequential is the default for mutation and recovery logic.
- Parallel validation preserves deterministic diagnostic ordering and obeys cancellation.
- Protect shared state with ownership first, synchronization second. Do not use channels merely to appear idiomatic.
- Run race tests for claims, locks, projection rebuilds, cancellation bookkeeping, and any parallel validation introduced.
- Never rely on process-local locking alone for workspace correctness across processes.

## Testing policy

- Drive use-case behavior primarily through `Engine.Execute`; test pure package laws locally and platform faults at adapter boundaries.
- Prefer table tests when cases share behavior, not when a table obscures a state machine or causal sequence.
- Use hermetic temporary workspaces created by `internal/testkit`; never target the developer's repository or home directory.
- Use deterministic fakes for clock/ID, process, Git, and transaction failure injection.
- Test the failure immediately before and after every persistence boundary: prepare write, file sync, rename/replace, directory sync where supported, journal append, projection update, and recovery.
- Fuzz parsers, frontmatter, relative paths, protocol decoding, workflow compilation, delta application, and recovery manifests.
- Test public JSON against schemas and invalid fixtures. Golden files require an explicit review path; regeneration is not the assertion.
- Keep platform claims honest with OS-specific tests for path semantics, file replacement, locking, signals/process groups, and Windows Job Objects.
- A test that skips because a capability is absent cannot prove that capability for a release artifact.

## Dependency admission

Before adding a module, record:

1. the concrete capability missing from the standard library/current code;
2. why a small local implementation would be riskier;
3. transitive dependency, CGO, cross-compilation, licensing, vulnerability, and maintenance impact;
4. how the dependency is isolated and could be replaced;
5. the tests that protect behavior during an upgrade.

Do not add a library only to obtain `Result[T]`, dependency injection, repositories, service containers, logging facades, UUIDs, configuration merging, or filesystem abstractions before the need is proven.

## Review rejection criteria

Reject a Go change when it:

- adds a generic layer with no invariant or information hiding;
- lets transport/storage types dictate every runtime representation without examining access patterns;
- creates a pointer-rich object graph or per-record interface dispatch where a simple indexed data pass is clearer;
- adds ECS, struct-of-arrays, pooling, arenas, `unsafe`, or a custom format without representative profiling evidence and a deletion criterion;
- persists dense runtime indexes as stable identity;
- creates an interface on the implementation side only for mocking;
- handles errors by message matching, logging-and-returning, panic, or silent fallback;
- turns workflow YAML into executable code or an authority bypass;
- launches unowned goroutines or loses cancellation;
- introduces nondeterministic rendering, hashes, diagnostics, or fixtures;
- mutates during a read-only command;
- mixes planning and effect application so exact authorization is impossible;
- requires callers to understand several internal packages to perform one supported use case.

## Reference basis

- Richard Fabian's [Data-Oriented Design](https://www.dataorienteddesign.com/dodbook/dodmain.html) supplies the game-industry framing: design around data, access patterns, and transformations rather than object taxonomies.
- Go's official [garbage-collector guide](https://go.dev/doc/gc-guide) explains pointer scanning, allocation costs, index-versus-pointer trade-offs, and why optimization must follow profiling.
- Go's official [diagnostics guide](https://go.dev/doc/diagnostics) defines the profiling evidence used before specialized layouts are admitted.
- Go's official [Code Review Comments](https://go.dev/wiki/CodeReviewComments) guide context propagation, error-flow indentation, interface ownership, and naming.
- Go's official [errors package](https://pkg.go.dev/errors) defines wrapping-tree inspection and joining behavior.
- Rob Pike's [Errors are values](https://go.dev/blog/errors-are-values) motivates composing behavior with explicit error values instead of treating checks as boilerplate.
- The repository's binding package/dependency decisions are in `docs/architecture/go-project-structure.md` and `.kiro/steering/structure.md`.
