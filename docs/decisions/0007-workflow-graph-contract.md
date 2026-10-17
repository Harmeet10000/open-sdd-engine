# ADR-0007: Accept the pure workflow graph contract

- Status: Accepted
- Date: 2026-10-17
- Scope: closed workflow definitions, bounded compilation, and deterministic evaluation
- Source conversation: `01a1278b-94cb-7c63-a18c-89e44c073dc1`

## Context

Every built-in and organization-defined journey needs the same answer to three
questions: whether its topology is safe, which nodes are ready from the current
facts, and what action may be proposed next. If commands or guided workflows
answer those questions independently, retries become graph mutations,
validation and completion authority leak into routing, and apparently identical
work can produce different actions from map order, wall-clock reads, or stale
files.

The requirements, design, and task decomposition in
`.kiro/specs/sdd-workflow-graph/` have been reviewed as one contract. This ADR
accepts the pure graph module for implementation while preserving public
protocol publication and real producer/consumer integrations as explicit
release gates.

## Decision

Accept `WG-REQ-001` through `WG-REQ-012`, the design's deep-module seam, and
the `WG-TASK-001` through `WG-TASK-014` implementation DAG with these binding
baselines:

1. The external module interface is three context-aware pure operations:
   `Compile`, `ResolveComposition`, and `Evaluate`. `context.Context` is the
   first parameter so bounded proof regions and large deterministic passes can
   honor caller cancellation. Context values never affect graph semantics or
   identity.
2. Version 1 graph definitions use closed JSON with media type
   `application/vnd.open-sdd.workflow-definition+json;version=1` and schema ID
   `urn:open-sdd:workflow-definition:1`. There is no implicit YAML fallback.
   A later encoding requires a distinct decoder, exact-byte identity rule, and
   migration fixtures.
3. Graph semantic canonicalization uses
   `sdd.workflow.canonical-json/1`; compilation uses
   `sdd.workflow.compiler/1`. These identify graph behavior only and must not
   reuse artifact-validation canonicalizer identities.
4. The hard limits in the accepted design are conservative safety ceilings.
   A deployment profile may lower them. Changing a hard ceiling changes the
   compiler-contract identity and requires compatibility and adversarial
   fixtures; benchmark results may inform such a reviewed change but do not
   permit an unbounded fallback.
5. Predicates are the closed typed tree in the design and use strong Kleene
   three-valued logic. Missing, stale, unverifiable, revoked, or wrong-subject
   inputs remain `unknown`; they never become false, passed, skipped, or quick.
6. The compiled topology is immutable for a work revision. Attempts, retries,
   rejection, recovery, reconciliation, and cancellation are facts over that
   topology, not edges or implicit definition revisions.
7. The graph consumes owner facts as immutable values and emits states,
   diagnostics, composition decisions, transition intentions, and action or
   completion-handoff proposals. It never reads the workspace, parses an
   artifact, runs a validator/process, manages claims, applies effects, accepts
   completion, or generates DONE/receipt content.
8. Pure compilation and evaluation may be implemented and verified now.
   Engine publication, durable submission, real validation/execution facts,
   and completion handoff conformance remain blocked until their owning
   contracts are ratified. Test-only facts must remain visibly test-only.
9. The published foundation protocol remains `1.0`. The graph must not assign
   a public minor version or add result variants until the foundation-owned
   extension inventory and schemas are accepted.

## Consequences

- `internal/workflow` becomes one deep, in-process module. Its interface is
  also its primary test surface; graph algorithms and canonical encoding do
  not receive speculative interfaces.
- Identical definitions and immutable fact inputs produce byte-stable
  identities, states, diagnostics, and proposal order on every supported
  platform.
- Workspace, artifact validation, completion, bounded execution, presentation,
  and guided workflows remain adapters or peer owners coordinated by the
  engine; none is reimplemented inside the graph.
- The implementation may initially prove the compiler and authored-node paths
  with in-memory typed values. That evidence cannot be relabeled as durability,
  policy authority, process supervision, or public protocol conformance.

## Release gates that remain open

- Foundation allocation and publication of the shared graph result extension.
- Ratified artifact-validation submission, validity, assessment, and validator
  result contracts plus real producer/consumer evidence.
- Ratified completion candidate/gate/handoff contracts.
- Bounded-execution attempt, result, claim, and cancellation contracts.
- Company policy/assessment/waiver issuers, revocation, retention, privacy, and
  reviewer-independence decisions.
- Corpus and benchmark evidence for production deployment profiles; it may
  justify a reviewed limit change but does not reopen boundedness.

## Rejected alternatives

- A workflow interface without `context.Context`: it would make the promised
  cancellation checks impossible to propagate from `Engine.Execute`.
- Executable predicates or plugins: they turn compilation into code execution
  and make closed identity, offline use, and bounded proof impossible.
- Treating every terminal predecessor as successful: failure, cancellation,
  staleness, and skip would silently authorize downstream work.
- Letting retries modify topology: history would cease to be comparable and
  a recovery event could change the workflow being recovered.
- Publishing graph result variants before foundation allocation: it would
  create a second protocol whose compatibility the engine cannot guarantee.

## For the chosen ones

The compiler's real product is not an adjacency list; it is a reusable proof
artifact whose identity binds the normalized topology, proof rules, and hard
resource budget. A graph compiled with a larger proof budget is not necessarily
the same trusted object: the larger compiler may admit a definition the smaller
one must reject. Including limits in compiled identity prevents a cache or
organization boundary from laundering “compiled successfully elsewhere” into
local acceptance.
