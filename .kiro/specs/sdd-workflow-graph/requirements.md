# Requirements Document: sdd-workflow-graph

## Status and scope

This is a proposed, reviewable requirements draft. It is not evidence of
acceptance or implementation. It defines the generic declarative graph/compiler
and deterministic state evaluation seam used by `sdd`. It does not define the
Bug Fix or Quick Spec journey contents, durable file transactions, artifact
parsing, process execution, claims/leases, or completion/DONE generation.

The graph is a deep module: callers provide a pinned definition and an
immutable fact snapshot; the graph returns compiled identity, typed state
facts, diagnostics, and proposed next actions. It returns no authorization and
performs no effects. All requirements and acceptance criteria have stable IDs
so that fixtures and later design/tasks can refer to them without using section
numbers as identity.

## Glossary

- **Definition**: a versioned, immutable graph document containing nodes, ports,
  edges, predicates, policies, and profile bindings.
- **Dependency edge**: a relation requiring a successful source result before a
  target may proceed. It is not a freshness relation.
- **Freshness edge**: a relation requiring a current source digest/validity fact
  for a target input. It does not by itself assert that the source succeeded.
- **Predicate result**: `true`, `false`, or `unknown`; unknown is not false and
  is never coerced to pass.
- **Ready**: a derived state whose required dependencies, inputs, gates, and
  predicates are satisfied and whose action has not begun.
- **Running**: a durable attempt has begun. **Claimed** is a separate external
  execution fact and does not mean running or authorize execution.
- **Attempt**: an immutable record of one node transition request/result. A
  retry creates a new attempt and preserves prior attempts.
- **Initiative**: a stable parent with independently tracked child leaves and
  aggregate status. It is not a second definition hierarchy.
- **Fact snapshot**: the immutable, revision-bound values supplied by workspace,
  artifact-validation, policy, execution, and completion owners.

## Ownership and interface ledger

The graph consumes and produces small, versioned interfaces. `Engine.Execute`
and the public envelope, source-coordinate units, diagnostics, and exit-code
mapping remain owned by `sdd-engine-foundation`. Workspace owns snapshots,
revisions, locks, idempotency, effects, journaling, and recovery. Artifact
validation owns parsing, IDs, canonicalization, digest calculation, and
validation results. Completion owns candidate-bound gates, evidence acceptance,
DONE, receipts, promotion, archive, and completion lifecycle. Bounded execution
owns process supervision and claims/leases. Guided workflows own specialized
journey definitions and content policy.

| Ledger ID | Interface | Definition owner | Graph responsibility | Producer/consumer and blocker |
|---|---|---|---|---|
| WG-C-001 | immutable fact snapshot, selection coordinates, revision tokens | workspace-transactions (foundation read-only seam is upstream) | consume opaque snapshot ID, scoped revisions, selected work; never reread | workspace produces; graph consumes. Blocked until snapshot contract/version is reviewed. Cross-ledger IC-01. |
| WG-C-002 | pinned graph definition and composition identity | workflow-graph | define and compile stable nodes/ports/edges/predicates | guided workflows produce specialized definitions; graph produces compiled identity. Cross-ledger IC-04. |
| WG-C-003 | action evaluation/next-action result | workflow-graph | return state, blockers, structured action proposal, expected/action revisions | engine consumes and renders; execution/completion may consume facts. Blocked on foundation envelope extension. Cross-ledger IC-05. |
| WG-C-004 | artifact/evidence validity fact | artifact-validation and completion | consume status, semantic/derivation digest, validator/evidence provenance | validation/completion produce; graph never parses or trusts agent claims. Blocked on versioned fact schemas. Cross-ledger IC-07 and IC-11. |
| WG-C-005 | effective policy and assessment facts | trusted policy authority; artifact-validation normalizes inputs | apply precedence and floor, preserve `unassessed`/`unknown` | policy/validation produce; graph evaluates. Authority source remains an external decision. Cross-ledger IC-06 and IC-10. |
| WG-C-006 | workflow attempt, external claim, and execution facts | workflow owns action-attempt semantics; workspace persists; bounded-execution owns execution attempts/claims | derive workflow attempt transitions from accepted submissions; consume external execution/claim facts without acquiring/expiring/adopting leases | engine and declared provider supply facts. Only execution-dependent routes are blocked by absent execution contracts; structural authoring is not. Candidate/evidence linkage uses IC-09/IC-11. |
| WG-C-007 | mutation/replay/cancellation outcome | workspace-transactions | request an opaque transition/effect plan with expected revisions; interpret committed/replayed/conflict/recovery-required | workspace applies; graph does not write. Blocked on workspace exact-plan/idempotency contract. Cross-ledger IC-02 and IC-03. |
| WG-C-008 | completion gate/evidence outcome | completion-lifecycle | consume gate result and recommend completion action; never generate DONE | completion produces; graph consumes. Blocked until completion contract exists. Cross-ledger IC-12. |
| WG-C-009 | public command/result envelope | foundation | add discriminated graph result through reviewed minor schema amendment | foundation owns `Engine.Execute`, source, schema, and rendering. Blocked on one shared version allocation; this is a foundation seam outside the twelve domain-interface rows. |

The root cross-spec interface ledger uses IC-01..IC-12. The graph-side links
are: IC-01 → WG-C-001 (snapshot); IC-02 → WG-C-007 (mutation plan);
IC-03 → WG-C-007 (idempotency/replay); IC-04 → WG-C-002 (compiled graph);
IC-05 → WG-C-003 (action evaluation); IC-06 → WG-C-005 (composition/risk);
IC-07 → WG-C-004 (typed artifact identity); IC-08 → completion/artifact
owners (proposed delta, consumed only as opaque promotion input); IC-09 →
WG-C-004 (validation request/result); IC-10 → WG-C-005 (effective policy);
IC-11 → WG-C-006/WG-C-008 (verification subject and candidate evidence);
IC-12 → WG-C-008 (completion proposal/receipt). IC-08 is deliberately not
implemented by the graph.

### Non-goals and exclusions

The graph does not own `.sdd` paths, staging directories, rename/cleanup,
filesystem atomicity, journal recovery, lock ownership, idempotency storage,
artifact Markdown/YAML parsing, semantic digest implementation, subprocesses,
resource limits, scheduling, agent invocation, completion checklists, DONE,
receipts, promotion, archive, Git mutation, automatic merge, or a competing
command/result envelope. It may describe typed intentions and consume facts
from those owners, but it never directly applies them.

## Requirements

### WG-REQ-001 — Definition language and identity

1. A definition SHALL contain a stable ID, explicit schema version, immutable
   source identity, semantic definition version, and exact content digest.
2. A node SHALL declare a stable ID, kind, version, source coordinate, input
   and output ports, required gates/predicates, and profile binding. A port
   SHALL declare name, direction, type reference, explicit cardinality, and
   source coordinate.
3. An edge SHALL declare a stable ID, source node/port, target node/port,
   relation (`dependency` or `freshness`), optional predicate, and coordinate.
4. Definitions SHALL be closed data. They MAY contain inert descriptions,
   examples, prompt/reference text, and remediation prose when the schema marks
   those values as data. Such text SHALL never be interpreted as an instruction
   or execution authority. Only the bounded predicate expression grammar may be
   evaluated; executable code, shell callbacks, network references, dynamic
   plugins, and untyped expressions SHALL be rejected.
5. Cardinality SHALL be one of `one`, `zero-or-one`, `zero-or-many`, or
   `at-least-one`. Missing cardinality on a required input or output SHALL be a
   compile error; it SHALL never silently mean `zero-or-one`. Any legacy default
   requires a versioned compatibility rule and fixture.
6. A graph digest SHALL canonicalize only the graph's declared semantic fields
   under the pinned canonicalizer/profile identity. Source coordinates and
   presentation-only whitespace may be excluded only where that profile says so;
   behavior-changing predicates, ports, edges, policy bindings, or unknown
   potentially meaningful fields SHALL change identity.

Acceptance criteria: WG-AC-001 (valid typed definition compiles with stable
identity); WG-AC-002 (missing required cardinality fails rather than becoming
optional); WG-AC-003 (embedded command/callback is rejected); WG-AC-004
(cosmetic permitted change preserves semantic identity while predicate/edge
change invalidates it). Fixtures: WG-F-001, WG-F-002, WG-F-003, WG-F-004.

### WG-REQ-002 — Compilation and bounded validation

1. Compilation SHALL reject duplicate node IDs, edge IDs, port names, and
   ambiguous stable identities; every diagnostic SHALL include a foundation
   source coordinate and structured remediation.
2. Compilation SHALL reject missing node/port references, missing producers,
   incompatible port types, invalid cardinality combinations, malformed
   predicates, and unsupported definition/schema versions.
3. Compilation SHALL detect self, direct, and indirect cycles in dependency
   edges and in the combined dependency/freshness graph whenever the cycle can
   prevent readiness. It SHALL report the exact edge types and cycle path. A
   retry or rejection event SHALL not be represented as a new graph edge.
4. Compilation SHALL detect unreachable required nodes/gates and impossible
   completion routes. It SHALL distinguish static impossibility (provable from
   definition) from runtime blocking caused by missing/unknown facts.
5. Node/edge counts, predicate recursion, expression size, and evaluation
   complexity SHALL have enforceable upper bounds. Exhaustion SHALL be a
   deterministic compile diagnostic, not an unbounded runtime operation.
6. Compile failures SHALL occur before the first state transition and SHALL be
   distinguishable from runtime `unknown`, stale, blocked, conflict, or policy
   results.

Acceptance criteria: WG-AC-005 through WG-AC-010 cover duplicate/missing
references, type mismatch, dependency and readiness-cycle paths, unreachable
mandatory gate, impossible route, and each bound. Fixtures: WG-F-005 through
WG-F-012 (WG-F-012 runs separate impossible-route and budget-exhaustion cases).

### WG-REQ-003 — Predicates, reachability, gates, and inputs

1. Predicate evaluation SHALL return `true`, `false`, or `unknown` with the
   missing fact/reason and source coordinate. `unknown` SHALL not be coerced to
   `false`, `passed`, or `quick`.
2. A false skip predicate makes a node/edge applicable; a true skip predicate
   excludes only the explicitly scoped route. An unknown skip predicate blocks
   the affected decision unless a pinned policy explicitly defines a safe
   disabled disposition.
3. A mandatory gate cannot be removed by a conditional route. A skipped node
   SHALL not waive a required output or downstream mandatory gate unless the
   definition explicitly declares that route not applicable and compilation
   proves the remaining route is complete.
4. For every applicable input, the graph SHALL enforce cardinality. A source
   in `failed`, `unknown`, `stale`, or merely `skipped` state SHALL not satisfy
   a required dependency unless the edge's declared conditional semantics make
   that source not applicable. A successful dependency requires the source to
   have completed with its required output and current validity facts.
5. Static impossibility (for example mutually exclusive required predicates on
   every route) SHALL be a compile error; a runtime unavailable assessment,
   expired policy, disabled conditional, or absent optional fact SHALL be a
   runtime diagnostic with a safe next action.
6. Reachability evaluation SHALL include active predecessors, required outputs,
   gates, and both relation types. It SHALL not equate “predecessor terminal”
   with “dependency satisfied.”

Acceptance criteria: WG-AC-011 through WG-AC-016 cover tri-state predicates,
conditional gates, successful dependency semantics, cardinality, static versus
runtime failure, and route reachability. Fixtures: WG-F-013 through WG-F-018.

### WG-REQ-004 — Legal runtime states and transition table

The graph SHALL evaluate node state using the following normative table. The
graph records facts and proposes transitions; it does not claim or run work.

| Current/derived state | Event/input | Authority and preconditions | Result/facts | Revision/diagnostic |
|---|---|---|---|---|
| `unstarted` | evaluation | snapshot and compiled definition available | `ready` if all required inputs/gates/predicates are satisfied; otherwise `blocked`, `stale`, or `skipped` | no mutation; missing/unknown fact is structured blocker |
| `ready` | action proposal | no authorization implied; expected action/work revision captured | typed action with action ID/revision and required submission schema | no mutation; stale expected revision returns conflict/restart |
| `ready` | accepted authored/structural submission | workspace records the typed submission and validation result; no process provider is required for this node kind | `completed` when the required structural/policy fact passes, otherwise `failed`/`blocked` with the validation result | workspace/work revision consequence is consumed, not allocated by graph; absence of an execution provider does not block a valid draft submission |
| `ready` | accepted execution-backed action | workspace transition accepted and the execution provider records an attempt | `running` with immutable attempt ID | provider availability is required only for this node/profile, not for authored/structural validation |
| `running` | durable execution/result fact | execution provider or trusted assessment supplies result for same attempt/candidate | `completed`, `failed`, `cancelled`, `stale`, or `blocked` as defined below | result includes evidence/provenance; mismatched subject is stale/conflict |
| `running` | claim observation | bounded-execution owns claim/lease and freshness | graph reports `claimed` as a separate fact, never a replacement state | claim expiry/revocation is a blocker; graph never adopts/releases |
| any nonterminal | source/policy/definition digest change | current snapshot compared to pinned expected digest | `stale` when a freshness edge is invalid; re-evaluation may yield `ready` after new facts | no automatic topology rewrite; diagnostic identifies expected/current digest |
| any applicable node | explicit skip fact | predicate is true and policy permits skip; required outputs/gates remain checked | `skipped` with reason and route scope | missing authority/unknown predicate blocks rather than skipping |
| `failed` | policy-authorized retry | prior attempt retained; new expected snapshot and idempotency key | new `running` attempt; previous remains immutable | attempt number increments; no topology revision |
| `running` | cancellation outcome | workspace/execution provider reports whether durable intent was reached | `cancelled` before durable transition, or recovery/committed fact after it | graph mirrors owner outcome; it does not roll back effects |
| `failed` | attempt limit reached | policy says no further retry without intervention | terminal `failed(exhausted)`; downstream ordinary dependencies remain `blocked` | actionable policy/intervention diagnostic; no infinite retry |
| `completed` | evaluation | required outputs and validity remain current | remains completed or becomes stale if freshness invalidates | later edits create new candidate/attempt, not history rewrite |

Acceptance criteria: WG-AC-017 (ready is derived and distinct), WG-AC-018
(running and claimed are not conflated), WG-AC-019 (all transition outcomes and
authority are recorded), WG-AC-020 (failed exhaustion is terminal at node and
downstream blocks), WG-AC-021 (cancellation mirrors durable owner outcome).
Fixtures: WG-F-019 through WG-F-025.

### WG-REQ-005 — Static topology, attempts, and reconciliation

1. The compiled topology and pinned definition SHALL be immutable for a work
   revision. Runtime attempts, results, rejection, retry, cancellation, and
   reconciliation facts SHALL be separate records.
2. An attempt identity SHALL include work ID/revision, node ID/version,
   candidate/input digests, action ID/revision, and monotonically scoped
   attempt number. Prior attempts SHALL never be overwritten.
3. Retry, rejection, and recovery SHALL revisit the same topology. A runtime
   reconciliation (including policy exception, manual override, or execution
   retry) SHALL NOT automatically create a topology/definition revision.
4. A new topology revision is required only for an explicit definition change
   or accepted migration/recomposition. It must preserve or explicitly map
   compatible history and produce a new pinned digest.
5. Attempt exhaustion SHALL yield the normative `failed(exhausted)` node state;
   a dependent node is `blocked` because its dependency is not successful.
   Policy intervention may authorize a new attempt without changing topology.
6. Roll-forward after interruption SHALL use the durable attempt/transaction
   outcome supplied by workspace/execution. The graph SHALL not reconstruct a
   committed result by rereading mutable files or inventing a second effect.

Acceptance criteria: WG-AC-022 through WG-AC-026 cover immutable topology,
attempt identity/history, retry without cycles, no automatic topology revision,
exhaustion, and interrupted roll-forward. Fixtures: WG-F-026 through WG-F-030.

### WG-REQ-006 — Deterministic evaluation and next actions

1. Identical fact snapshot, pinned definitions, policy, assessments, and
   attempt history SHALL produce byte-stable state facts, diagnostics, and
   recommendation ordering.
2. The default result SHALL contain one useful human action or an explicit
   blocked/completed result. Machine consumers may request additional eligible
   alternatives. Equivalent actions tie-break by stable node ID, action ID,
   then source coordinate.
3. A blocked result SHALL name unsatisfied dependencies, missing/unknown gates,
   expected/current revisions or digests, and one safe remediation. A stale
   result SHALL identify the invalidated edge and expected fresh input.
4. A recommendation SHALL be a typed proposal only; it is not authorization,
   claim acquisition, effect application, or completion acceptance.
5. When no graph action remains, the graph SHALL return completion-gate facts
   and a structured handoff to completion. It SHALL not generate `done.md`, a
   receipt, or a completion checklist as authoritative output.

Acceptance criteria: WG-AC-027 through WG-AC-031 cover deterministic bytes,
one human action, tie-breaking, blocked/stale explanation, and completion
handoff ownership. Fixtures: WG-F-031 through WG-F-036.

### WG-REQ-007 — Composition, policy precedence, and assessments

1. Composition SHALL resolve journey, rigor, artifact/validation profiles,
   graph version, validator bindings, and exact definition/policy digests from
   typed inputs. The graph SHALL not infer authority from prose.
2. Precedence SHALL be trusted organization-required policy, then project
   narrowing policy, then work rationale. A missing, expired, revoked, stale,
   or unverifiable required policy fails closed.
3. An explicit preset may increase rigor but cannot lower the effective policy
   floor. A low-risk flag cannot override a trusted critical policy fact.
4. External risk/materiality/diagnosis assessments SHALL include issuer,
   evidence, uncertainty, captured source revision, and assessment status
   (`unassessed`, `insufficient`, `assessed`). Null or missing is unassessed.
   The graph consumes the fact and never turns a heuristic into proof.
5. A policy or composition change is a runtime stale/blocking fact until an
   explicit proposal/acceptance/application transition creates a new work
   composition revision. The graph preserves journey and valid history when
   policy permits; changing journey requires explicit migration.

Acceptance criteria: WG-AC-032 through WG-AC-036 cover precedence, missing
authority, policy floor, typed unknown assessment, emergency reconciliation,
and invalid escalation. Fixtures: WG-F-037 through WG-F-042.

### WG-REQ-008 — Definition pinning, compatibility, and freshness

1. Active work SHALL pin graph definition, composition, policy, artifact
   profiles, and validator semantic-contract versions/digests. Offline lookup
   uses only those pins; missing pins fail closed.
2. Updating a named preset SHALL not rewrite an active work. It returns a
   `WORKFLOW_UPDATE_AVAILABLE` diagnostic with migration data.
3. Compatibility is explicit and versioned. A node/predicate/profile change is
   not backward-compatible merely because the display name or declared patch
   version is unchanged. Adding a node is compatible only when the versioned
   compatibility rule proves existing accepted paths remain valid.
4. Freshness edges consume source semantic/derivation/evidence validity facts
   from their owners. The graph propagates invalidation only along declared
   freshness edges and includes the source digest, derivation inputs, and
   expected digest in the stale fact.
5. A node SHALL never require its own output/receipt to become ready. A bundle
   may be submitted for validation before acceptance; verification evidence is
   produced by verification and is not a prerequisite for beginning it.
6. A semantic no-op formatting change preserves freshness only when the owning
   canonicalizer says it is a no-op. Meaningful wording, code, ordering,
   unknown-field, policy, validator, or dependency changes conservatively stale
   affected consumers.

Acceptance criteria: WG-AC-037 through WG-AC-042 cover offline pins, named
update, compatibility, owner-provided freshness, no self-output cycle, and
semantic/no-op invalidation. Fixtures: WG-F-043 through WG-F-047.

### WG-REQ-009 — Generic submissions and validation facts

1. The submission seam SHALL carry typed artifact/assessment facts with status
   `draft`, `submitted`, `rejected`, `accepted`, or `stale`, source coordinate,
   artifact ID, work/action revision, blob/semantic/derivation digests, and
   idempotency/replay reference where applicable.
2. The graph requests validation and consumes a validation result containing
   validator ID, semantic-contract/definition digest, candidate/input digest,
   result status, evidence references, and producer provenance. It SHALL not
   parse artifacts or trust an agent-provided `accepted`/`passed` field.
3. Mandatory evidence that is missing, stale, cancelled, unverifiable,
   truncated, or empty is not a pass. Such a fact is a blocker or unavailable
   result according to the foundation/external-tool taxonomy.
4. A rejected artifact remains in attempt history with reasons. Retry creates
   a new candidate/attempt as authorized; it does not delete or rewrite the
   rejection.
5. Same-key same-content duplicate submission behavior SHALL be the workspace
   idempotency replay result. Same-key different-content is a conflict. A
   second simultaneous submission is not universally rejected: it conflicts
   only when its expected revision/preconditions lose the workspace race;
   an exact in-flight/replayed duplicate follows the workspace contract.

Acceptance criteria: WG-AC-043 through WG-AC-047 cover status/facts, untrusted
claims, unavailable mandatory evidence, rejection history, positive draft
submission, invalid acceptance self-dependency, and workspace
replay/concurrency. Fixtures: WG-F-048 through WG-F-053.

### WG-REQ-010 — Initiatives and compound progress

1. An initiative parent ID and each leaf ID SHALL remain stable across retries,
   resume, and revisions. A new initiative gets a new parent ID.
2. Leaf progress SHALL be tracked independently. Parent status is a deterministic
   aggregate over leaf states and does not imply atomic rollback or cancellation.
3. Dependencies between leaves SHALL be dependency edges; freshness edges are
   separate and require digest validity. A terminal failed/cancelled leaf does
   not satisfy an ordinary dependency merely because it is terminal.
4. Failure or cancellation of one leaf SHALL not automatically cancel unrelated
   leaves. A dependent leaf becomes blocked or conditionally skipped according
   to its declared edge/predicate and records the reason.
5. Resume facts SHALL be persisted through workspace transactions. A cursor may
   optimize display, but evaluation SHALL consider every eligible unresolved
   leaf/dependency, preserve completed leaves, and never replay successful
   effects merely because a cursor points earlier.
6. A compound operation with partial success retries only eligible failed or
   unresolved leaves. Atomic bundle acceptance remains one workspace transaction
   and its replay/cancellation outcome comes from workspace.

Acceptance criteria: WG-AC-048 through WG-AC-053 cover identity, aggregate
status, relation separation, cancellation, all-leaf resume, and partial
retry/atomic acceptance. Fixtures: WG-F-054 through WG-F-057.

### WG-REQ-011 — Engine integration and public contract

1. `start` accepts a preset and records routing/composition evidence without
   changing a pinned definition. `next` is read-only and returns expected
   workspace/work/action revisions plus one proposal. `submit` validates those
   coordinates and delegates mutation to workspace. `finish` is a completion
   handoff, not graph-owned completion.
2. Results SHALL pass through `Engine.Execute` and the foundation envelope.
   Graph extensions require one reviewed minor schema allocation; they SHALL
   not invent a competing command spelling, envelope, result shape, or
   protocol version.
3. Every diagnostic SHALL use foundation source coordinates and stable
   severity. Human and JSON output consume the same facts and preserve stdout/
   stderr separation, structured arguments, redaction, and non-interactive
   behavior.
4. Graph outcomes SHALL use the foundation exit taxonomy: `0` success,
   `1` validation/verification failure, `2` usage, `3` workspace/config/schema,
   `4` conflict/stale revision/claim, `5` required external-tool failure,
   `6` policy/permission denial, and `130` interrupted after cleanup.
   Read-only `next` may successfully report a blocker with exit `0`; an
   attempted prohibited mutation returns the applicable nonzero category.
5. Recovery commands expose state and an owner-provided safe action. The graph
   does not auto-fix, clean staging, release locks, adopt leases, or apply a
   plan.

Acceptance criteria: WG-AC-054 through WG-AC-058 cover command seam, versioned
envelope, coordinates/rendering, full exit mapping, explicit recovery, and
changed-candidate staleness. Fixtures: WG-F-058 through WG-F-061.

### WG-REQ-012 — Rebuildable graph facts

1. Canonical graph facts are the pinned definition/composition, submission and
   validation facts, attempt records, leaf progress, and workspace references.
   Caches/projections/selection views are rebuildable and are not a second
   authority.
2. Rebuilding graph state from the same canonical facts and snapshot SHALL
   produce identical state, diagnostics, digests, and ordered actions.
3. Cache/projection loss or unavailable derived storage SHALL produce a
   rebuild/read-only diagnostic, not changed correctness or silent mutation.

Acceptance criteria: WG-AC-059 through WG-AC-061 cover authority, deterministic
rebuild, and cache loss. Fixtures: WG-F-062 through WG-F-063.

## Conformance fixture catalog and coverage matrix

Every fixture declares definition/fact preconditions, input/action, expected
observable result, and forbidden side effects in the test plan. The fixture IDs
below are normative handles; they are not claims that tests have run.

| Fixture IDs | Required case and expected result |
|---|---|
| ID | Preconditions and action | Expected observable result | Forbidden side effect |
|---|---|---|---|
| WG-F-001 | Valid typed DAG with explicit IDs/types/cardinalities; compile it | Compiles with stable definition digest and coordinates | No files, locks, claims, or effects |
| WG-F-002 | Required input omits cardinality; compile | Compile diagnostic, never optional default | No runnable graph |
| WG-F-003 | Description contains an inert prompt/example; predicate contains shell/callback | Inert text is data; executable/untyped expression is rejected | No prompt execution or shell invocation |
| WG-F-004 | Change permitted formatting, then a predicate/edge; compile both | Only permitted formatting preserves semantic digest; behavior change changes it | No silent digest reuse |
| WG-F-005 | Two nodes share an ID; compile | Duplicate-ID diagnostic with both source coordinates | No partial compiled definition |
| WG-F-006 | Two edges or ports share an ID/name; compile | Duplicate edge/port diagnostic | No arbitrary winner |
| WG-F-007 | Edge names missing source node/output; compile | Missing-producer/reference diagnostic | No inferred producer |
| WG-F-008 | Edge connects incompatible port profiles; compile | Type/profile diagnostic | No coercion or runtime attempt |
| WG-F-009 | Self, direct, and indirect dependency cycles; compile | Each exact dependency cycle path is reported | No retry edge is synthesized |
| WG-F-010 | Freshness-only or mixed cycle prevents all members becoming ready; compile | Readiness-cycle diagnostic names freshness/dependency edges | No infinite readiness evaluation |
| WG-F-011 | Required gate has no path from an initial route; compile | Disconnected mandatory-gate diagnostic | No conditional bypass |
| WG-F-012 | Every completion route contains mutually exclusive required predicates; compile | Static impossible-route diagnostic before transitions | No runtime “unknown” substitution |
| WG-F-013 | Runtime predicate input is absent; evaluate | `unknown` with missing fact/source and remediation | No false/pass/quick coercion |
| WG-F-014 | Conditional skip excludes a required gate; compile/evaluate | Gate remains mandatory or definition is rejected | No policy enforcement bypass |
| WG-F-015 | Skipped route omits a required output; evaluate | Blocked/invalid route diagnostic unless explicitly not-applicable | No downstream waiver |
| WG-F-016 | Predecessor is terminal failed; evaluate dependent | Dependent remains blocked; no success fact | No terminal-is-success shortcut |
| WG-F-017 | Required input has zero or too many values; evaluate | Cardinality diagnostic and blocked node | No silent optional default |
| WG-F-018 | Compare static contradiction with missing runtime policy; evaluate | First is compile error; second is runtime blocked/unknown | No same diagnostic category |
| WG-F-019 | Unstarted node has all successful inputs/gates; evaluate | Derived `ready` state and typed action proposal | No mutation or authorization |
| WG-F-020 | Accepted execution action has attempt; separate claim fact arrives | `running` plus separate `claimed` observation | Graph does not acquire/expire/adopt lease |
| WG-F-021 | Authored/structural validation submission, no execution provider; submit | Required structural/policy result can complete or reject the node | Provider absence does not block every draft |
| WG-F-022 | Execution-backed node receives same-candidate provider result; evaluate | Completed/failed with evidence and attempt identity | No result accepted for another candidate |
| WG-F-023 | Node reaches policy attempt limit; retry without intervention | Terminal `failed(exhausted)`; ordinary dependents blocked | No infinite retry or topology rewrite |
| WG-F-024 | Cancel before workspace durable intent; submit cancellation | Cancelled/no authoritative graph mutation with owner outcome | No partial effect or silent continuation |
| WG-F-025 | Cancel after durable intent or uncertain reply; recover/evaluate | Committed/recovery-required fact is surfaced; no rollback invented | No duplicate effect |
| WG-F-026 | Retry same node/candidate twice; inspect records | Distinct immutable attempt identities and preserved history | No overwrite of prior attempt |
| WG-F-027 | Rejected submission then authorized retry; evaluate | New attempt/action while rejection remains historical | No deletion of rejection |
| WG-F-028 | Runtime reconciliation/exception changes facts only; evaluate | Same pinned topology, new runtime fact/recommendation | No automatic topology revision |
| WG-F-029 | Explicit definition migration accepted; compile new revision | New pinned digest/version with explicit history mapping | No silent active-work rewrite |
| WG-F-030 | Interrupted action has durable owner outcome; re-evaluate | Roll-forward uses recorded outcome and remains deterministic | No reread/replan to invent effect |
| WG-F-031 | Same snapshot/definition/policy/history evaluated twice | Byte-stable states, diagnostics, and order | No wall-clock/random ordering |
| WG-F-032 | Multiple eligible actions; default next | Exactly one useful human action plus structured alternatives if requested | No shell-only action string |
| WG-F-033 | Equivalent actions differ only by stable IDs/coordinates | Documented deterministic tie-break result | No map/insertion-order dependence |
| WG-F-034 | Required dependency/gate missing or unknown; next | Blocker names fact, revision/digest, and safe remediation | No authorization or auto-fix |
| WG-F-035 | Freshness source digest differs; next | Stale node identifies edge, expected/current digest, regeneration action | No use of stale output |
| WG-F-036 | No graph action remains; evaluate | Completion-gate handoff only | No graph-generated DONE/receipt/checklist |
| WG-F-037 | Organization policy, project narrowing policy, work rationale conflict | Organization floor wins deterministically | No local override of required policy |
| WG-F-038 | Required policy missing/expired/revoked/unverifiable | Closed blocked/policy diagnostic | No fallback to project/rationale |
| WG-F-039 | Explicit low-risk flag for trusted critical change | Floor violation rejected | No rigor reduction |
| WG-F-040 | Null/heuristic assessment without sufficient evidence | `unassessed`/`insufficient` fact and next action | No authoritative risk inference |
| WG-F-041 | Emergency reconciliation fact with valid scope/authority; evaluate | Existing topology and history retained; ordinary reconciliation gates remain | No automatic topology revision or skipped gate |
| WG-F-042 | Escalation lacks authority/scope/expiry/revocation evidence | Escalation rejected with policy diagnostic | No profile/rigor change |
| WG-F-043 | Active work references missing pinned workflow/profile/validator | Closed missing-definition diagnostic | No fallback by display name |
| WG-F-044 | Definition edited under same display name; next | Update-available/stale result requiring explicit migration | No silent adoption |
| WG-F-045 | Definition adds node or changes predicate under version claims | Compatibility result follows explicit rule; behavior change stale/rejected | No self-declared compatibility trust |
| WG-F-046 | Upstream semantic/derivation/evidence digest changes | Only declared freshness descendants stale with provenance | No whole-repository invalidation or missed descendant |
| WG-F-047 | Verifier requires its own receipt/output to become ready; compile | Self-output readiness cycle rejected | No circular verification proof |
| WG-F-048 | Valid unaccepted draft submitted to validation; submit | Submission enters validation/structural evaluation | No prior acceptance requirement |
| WG-F-049 | Graph definition says draft must already be accepted to begin acceptance; compile | Invalid acceptance self-dependency rejected | No inverted rejection oracle |
| WG-F-050 | Mandatory validation evidence is absent/forged/stale/truncated | Blocked/unavailable, not passed | No trust in agent `passed`/`accepted` field |
| WG-F-051 | Rejected artifact retried with new content/key | New attempt and reasons preserve old rejection | No history replacement |
| WG-F-052 | Same key and identical canonical content replayed | Workspace replay result returned | No second effect |
| WG-F-053 | Same key changed content or lost expected revision concurrently | Conflict/revision outcome from workspace | No universal duplicate rejection or overwrite |
| WG-F-054 | Initiative has stable parent/leaf IDs; retry one leaf | IDs and aggregate status remain stable | No new hierarchy or parent identity |
| WG-F-055 | One leaf fails/cancels while dependent/unrelated leaves exist | Dependent blocks/conditional-skips; unrelated leaf remains eligible; parent aggregates | No automatic cancellation of unrelated leaf |
| WG-F-056 | Three leaves completed, one failed; resume with cursor before first leaf | Completed leaves preserved; all eligible unresolved leaves considered | No replay of successful effects |
| WG-F-057 | Partial compound operation and atomic draft bundle | Only failed/unresolved leaf retries; bundle is one workspace transaction/replay | No whole-compound retry or partial bundle commit |
| WG-F-058 | Graph action sent through `Engine.Execute` | Foundation envelope/result discriminator and revisions are valid | No competing envelope/schema |
| WG-F-059 | Diagnostic uses path/field/range and each exit category | Coordinates, severities, and `0/1/2/3/4/5/6/130` map correctly | No message-selected exit code |
| WG-F-060 | Read-only next sees blocker; recovery state is pending | Observation may exit 0; recovery offers owner action and does not mutate | No cleanup, lock release, lease adoption, or plan apply |
| WG-F-061 | Candidate/workspace changes after proposal before acceptance | Action becomes stale/conflict with expected/current revisions | No acceptance against changed candidate |
| WG-F-062 | Rebuild from identical canonical facts and snapshot | Identical state, diagnostics, digests, and ordered actions | No cache as authority |
| WG-F-063 | Delete cache/projection/selection view; evaluate | Rebuild/read-only diagnostic; correctness unchanged | No silent authoritative mutation |

Handoff topic coverage: definition language → WG-REQ-001/WG-F-001..004;
compilation → WG-REQ-002/WG-F-005..012; reachability/gates → WG-REQ-003/
WG-F-013..018; runtime transition table → WG-REQ-004/WG-F-019..025; topology
versus attempts → WG-REQ-005/WG-F-026..030; deterministic next action →
WG-REQ-006/WG-F-031..036; composition → WG-REQ-007/WG-F-037..042; definition
pinning → WG-REQ-008/WG-F-043..047; escalation is included in WG-REQ-007/
WG-F-041..042; freshness → WG-REQ-008/WG-F-045..047; submissions → WG-REQ-009/
WG-F-048..053; initiatives/compound progress → WG-REQ-010/WG-F-054..057;
engine integration → WG-REQ-011/WG-F-058..061; rebuildability → WG-REQ-012/
WG-F-062..063.

Adversarial coverage is explicit: cycle, readiness-cycle, disconnected gate,
impossible predicate, missing producer, duplicate ID, type mismatch, and bound
exhaustion (WG-F-005..012); unknown predicate, conditional skip, skipped
required output, failed predecessor, cardinality, and static/runtime distinction
(WG-F-013..018); unknown risk, stale policy, missing pin, overridden floor,
emergency reconciliation, invalid escalation, same-name definition, changed
candidate, and upstream semantic/cosmetic edits (WG-F-037..047 and WG-F-061);
verification self-output deadlock, valid draft submission, and invalid
acceptance self-dependency (WG-F-047..049); missing/forged evidence,
retry/replay/concurrency, stale next, and deterministic tie-break (WG-F-031..035
and WG-F-050..053); and compound partial/resume/cancellation consistency
(WG-F-054..057).

## Mapping from the previous draft

The former unnumbered Requirements 1–17 map to WG-REQ-001/002/003/004,
WG-REQ-005, WG-REQ-006, WG-REQ-007, WG-REQ-008, WG-REQ-007, WG-REQ-008,
WG-REQ-009, WG-REQ-010, WG-REQ-011, **removed and delegated to workspace**,
WG-REQ-012, WG-REQ-011, and the fixture catalog respectively. Former
Requirement 17.19 is intentionally corrected: an unaccepted draft may be
submitted to validation; only a graph that requires prior acceptance in order
to begin acceptance is rejected. The old Requirement 14 staging/rename/
cleanup rules are removed because workspace-transactions owns them.

## Dependencies, blockers, and amendment proposals

Drafting dependency direction is foundation → workspace-transactions →
workflow-graph → artifact-validation/completion/guided-workflows; runtime fact
flow is workspace snapshot + policy + artifact/validation + execution facts →
graph evaluation → engine outcome/proposal → workspace/completion consumer.
These are distinct from module import/effect dependencies: graph imports only
versioned value contracts; workspace does not import graph or completion policy;
engine coordinates values.

Requirements approval is blocked on WG-C-001, WG-C-004, WG-C-006, WG-C-007,
WG-C-008, and WG-C-009 being versioned with their owner specs. The following
amendments are proposed rather than silently applied:

| Amendment | Owner/path | Proposed change | Compatibility/tests |
|---|---|---|---|
| WG-A-001 | foundation REQ-002/009/010 (cross-ledger IC-09) | Allocate one reviewed minor envelope extension for graph result/action fields and reuse `0/1/2/3/4/5/6/130`, source coordinates, and severity values | preserve 1.0 fixtures; WG-F-058..060; blocks integration |
| WG-A-002 | workspace-transactions WS-REQ-002/003/004/006/008/009 (cross-ledger IC-01/02/03) | Publish snapshot/revision, exact-plan, replay, cancellation and recovery outcome types consumed opaquely here | workspace conformance plus WG-F-019..030/WG-F-050..057; blocks design |
| WG-A-003 | artifact-validation requirements | Publish validation fact, semantic/derivation freshness and assessment schemas | WG-F-034..049; blocks validation integration |
| WG-A-004 | completion-lifecycle requirements | Publish completion gate/receipt handoff without graph-owned DONE | WG-F-029..033 and completion fixtures; blocks finish integration |
| WG-A-005 | bounded-execution later spec | Publish claim/attempt/result facts; graph consumes leases and never manages them | WG-F-019..023; blocks execution integration |

## References

- `docs/kiro-input/next-prerequisite-specs.md`, sections 2–4 and 7
- `docs/kiro-input/sdd-cli-spec-brief.md`, especially sections 6, 8, 10, 11.2–11.3
- `docs/architecture/guided-workflows.md`
- `docs/cli-product-plan.md`, sections 4.3–4.5, 6.1–6.5, 6.9–6.13, 8.5–8.6,
  12.2–12.3, and 16.4
- `.kiro/specs/sdd-engine-foundation/requirements.md`
- `.kiro/specs/sdd-workspace-transactions/requirements.md`
- `.kiro/specs/sdd-guided-workflows/requirements.md`
- `.kiro/steering/spec-authoring.md`, `.kiro/steering/cli-ux.md`,
  `.kiro/steering/quality-and-safety.md`

## Open decisions retained for design review

- Exact graph expression language and its maximum node/edge, depth, and cost
  budgets.
- One shared minor schema allocation and exact result discriminators for graph
  actions across foundation and guided workflows.
- Workspace snapshot/revision and idempotency record shapes, including recovery
  outcomes and cancellation at durable commit intent.
- Trusted policy/assessment authority and retention/redaction rules.
- Artifact-validation canonicalizer and freshness contract, including unknown
  field treatment and validator semantic compatibility.
- Completion gate/receipt handoff and the bounded-execution claim/attempt
  provider.

## For the chosen ones

The graph’s most important invariant is not “it is a DAG”; it is that every
edge has a *proof obligation*. A dependency edge asks, “did the predecessor
successfully produce the required value?” A freshness edge asks, “does the
consumer still refer to the same valid value?” Treating either question as
“the predecessor is terminal” creates a graph that looks deterministic while
quietly allowing failed work to authorize downstream work. Keeping those
relations separate is what makes retries, partial initiatives, and candidate
changes auditable rather than merely convenient.
