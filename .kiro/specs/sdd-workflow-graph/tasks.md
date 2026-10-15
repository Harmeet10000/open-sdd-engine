# Implementation Tasks: sdd-workflow-graph

## Draft status, scope, and execution rule

This is a provisional implementation plan derived from the proposed
`requirements.md` and `design.md`. It is not evidence that the requirements,
design, public schema amendment, implementation, fixtures, or integrations have
been accepted or completed. Every implementation checkbox remains unchecked.

The workflow graph is one deep, pure module. Its supported implementation
surface is `workflow.Compile`, `workflow.ResolveComposition`, and
`workflow.Evaluate`, coordinated through `Engine.Execute`. Tasks below may
define graph-owned values, compilation, composition, evaluation, attempts,
initiatives, diagnostics, and typed proposals. They must not implement:

- workspace snapshot acquisition, revision allocation, locks, idempotency,
  staging, exact-plan application, journaling, cleanup, or recovery;
- artifact Markdown/YAML parsing, artifact ID allocation, canonicalization,
  artifact digests, traceability, or validator execution;
- process supervision, scheduling, claims, leases, resource enforcement, or
  provider result fabrication;
- completion gate acceptance, candidate observation, `done.md`, receipts,
  promotion, archive, or lifecycle effects;
- Bug Fix or Quick Spec topology and specialized content policy; or
- a second public command/result envelope or locally allocated protocol version.

The first task is a walking slice through the public engine seam using
deterministic owner-fact fakes. Later tasks deepen that slice by observable
behavior rather than building all records, repositories, or handlers first.
Fixture-based tests are not release evidence for an external owner contract:
durability, process execution, authority, and completion remain blocked until
their real producer/consumer conformance suites pass.

## Tasks

### WG-TASK-001 — Walk one definition from compile to one engine proposal

- [ ] WG-TASK-001 — Walk one definition from compile to one engine proposal
- **Requirements:** WG-REQ-001, WG-REQ-004, WG-REQ-006, WG-REQ-011
- **dependsOn:** []
- **External dependencies:** `sdd-engine-foundation` REQ-002/REQ-003/REQ-008/REQ-009 deliver the existing `Engine.Execute`, `Outcome`, `Diagnostic`, `NextAction`, source-coordinate, rendering, and exit-taxonomy seams; IC-01 / `sdd-workspace-transactions` WS-REQ-002 delivers the immutable snapshot/revision value consumed by the fake-backed slice. Public graph result publication remains blocked on AM-02 / WG-A-001 and is not satisfied by this task.
- **Owned files/modules:** `internal/workflow/compile.go`, `internal/workflow/evaluate.go`, `internal/workflow/action.go`, `internal/engine/workflow.go`, package-local tests, and `testdata/workflow-graph/walking/`; no workspace or provider adapter.
- **Observable outcome:** compile a minimal valid typed DAG, evaluate an unstarted authored node to `ready`, select exactly one structured proposal, and carry the same facts through `Engine.Execute` to human and JSON rendering without mutation or authorization.
- **Primary fixtures:** WG-F-001, WG-F-019, WG-F-032, WG-F-058.
- **Proposed verification:** `go test ./internal/workflow ./internal/engine -run 'TestWorkflowWalkingSlice|TestWorkflowEngineWalkingSlice'` and validate the emitted provisional payload against the locally generated schema fixture only after foundation exposes the shared extension test hook.
- **Rollback/recovery checks:** snapshot, work, and action revisions are captured but never advanced; cancellation before or during pure evaluation returns without files, locks, claims, attempts, or recovery records.
- **Done when:** the same fake snapshot and definition yield byte-stable state/proposal facts through the engine seam, default human output exposes one useful action, JSON uses structured arguments, and the test proves that no owner effect method was called. This does not complete WG-F-058 production conformance until the foundation-owned extension is ratified.

### WG-TASK-002 — Deepen the walking slice with the closed definition and identity contract

- [ ] WG-TASK-002 — Deepen the walking slice with the closed definition and identity contract
- **Requirements:** WG-REQ-001
- **dependsOn:** [WG-TASK-001]
- **External dependencies:** IC-04 is graph-owned; foundation source-coordinate units from REQ-002 are consumed unchanged. The first serialization/schema ID and compiler-contract version require the design open decision and must not be inferred from artifact-validation's parser or canonicalizer.
- **Owned files/modules:** `workflow/schema/definition.schema.json`, `internal/workflow/definition.go`, `internal/workflow/decode.go`, `internal/workflow/predicate.go`, `internal/workflow/identity.go`, schema fixtures under `testdata/workflow-graph/definition/`, and focused fuzz seeds.
- **Observable outcome:** decode one declared graph serialization into closed typed nodes, ports, edges, action descriptors, predicates, cardinalities, coordinates, and declared extensions; compute distinct exact, semantic, and compiled identities without a self-hash; keep descriptions/examples inert.
- **Primary fixtures:** WG-F-002, WG-F-003, WG-F-004.
- **Proposed verification:** `go test ./internal/workflow -run 'TestDefinitionContract|TestDefinitionIdentity|TestInertDefinitionText'` and `go test ./internal/workflow -fuzz=FuzzDefinitionDecode -fuzztime=30s`; validate every valid/invalid definition document with the repository's approved Draft-07 schema validator once selected.
- **Rollback/recovery checks:** decoder or digest failure returns diagnostics and no `CompiledGraph`; no parser fallback, file write, network lookup, callback, shell, or partial runtime graph is available to callers.
- **Done when:** required cardinality cannot default, executable/untyped expressions are rejected while inert prose remains data, permitted presentation-only changes preserve only the identities allowed by the pinned canonicalizer, and any semantic/unknown critical change alters identity or fails closed.

### WG-TASK-003 — Compile references, types, cycles, routes, and bounded proof obligations

- [ ] WG-TASK-003 — Compile references, types, cycles, routes, and bounded proof obligations
- **Requirements:** WG-REQ-001, WG-REQ-002, WG-REQ-003, WG-REQ-008
- **dependsOn:** [WG-TASK-002]
- **External dependencies:** foundation REQ-002 supplies diagnostic coordinates/severity; no workspace, artifact parser, validator, or execution provider is required for this pure slice. Provisional hard limits and the compiler-contract identity require the corpus/benchmark review recorded in the design open decisions.
- **Owned files/modules:** `internal/workflow/compile.go`, `internal/workflow/index.go`, `internal/workflow/proof.go`, `internal/workflow/limits.go`, `internal/workflow/diagnostic.go`, compile fixtures under `testdata/workflow-graph/compile/`, and compiler fuzz tests.
- **Observable outcome:** one bounded deterministic compiler resolves stable references to runtime-local indexes, verifies port direction/type/cardinality, separates dependency and freshness relations, reports canonical dependency or mixed-readiness cycle paths, proves mandatory routes/gates, and emits no partial runnable graph on any failure.
- **Primary fixtures:** WG-F-005, WG-F-006, WG-F-007, WG-F-008, WG-F-009, WG-F-010, WG-F-011, WG-F-012.
- **Proposed verification:** `go test ./internal/workflow -run 'TestCompileIntegrity|TestCompileCycles|TestCompileRouteProof|TestCompileLimits'` and `go test ./internal/workflow -fuzz=FuzzCompileDefinition -fuzztime=60s`.
- **Rollback/recovery checks:** cancellation is checked between bounded proof regions and before large allocations; limit exhaustion produces `WORKFLOW_COMPILE_LIMIT` or `WORKFLOW_COMPILE_ANALYSIS_LIMIT`, never an unbounded solver fallback or runtime punt.
- **Done when:** duplicate/missing/type errors identify exact coordinates, each prohibited cycle reports edge IDs and relation types, disconnected or impossible mandatory routes fail statically, budget exhaustion is deterministic, and dense indexes never enter persisted/public identity.

### WG-TASK-004 — Evaluate tri-state predicates, reachability, gates, and cardinality

- [ ] WG-TASK-004 — Evaluate tri-state predicates, reachability, gates, and cardinality
- **Requirements:** WG-REQ-003, WG-REQ-004, WG-REQ-006
- **dependsOn:** [WG-TASK-003]
- **External dependencies:** IC-01 provides one immutable snapshot; IC-04 provides the compiled graph; IC-07/IC-09 and IC-10 provide typed validity, validation, policy, and assessment facts. Until those producer schemas are ratified, tests use explicitly test-only value fakes and cannot establish authority conformance.
- **Owned files/modules:** `internal/workflow/facts.go`, `internal/workflow/predicate_eval.go`, `internal/workflow/readiness.go`, `internal/workflow/state.go`, and `testdata/workflow-graph/evaluation/`.
- **Observable outcome:** evaluate strong-Kleene predicates and route applicability over copied/sorted immutable facts; distinguish compile impossibility from runtime unknown; require successful, current outputs rather than any terminal predecessor; enforce each applicable port's cardinality and mandatory gate.
- **Primary fixtures:** WG-F-013, WG-F-014, WG-F-015, WG-F-016, WG-F-017, WG-F-018.
- **Proposed verification:** `go test ./internal/workflow -run 'TestPredicateTruthTables|TestReachabilityAndGates|TestRuntimeCardinality|TestStaticVersusRuntimeBlock'`.
- **Rollback/recovery checks:** evaluation is read-only; duplicate/contradictory owner facts fail instead of using last-element-wins; an unknown skip predicate blocks and never authorizes a skip or mutation.
- **Done when:** every unknown result retains owner, missing fact/reason, subject/digest where known, and source coordinate; failed/stale/skipped predecessors cannot satisfy ordinary dependencies; required outputs/gates cannot be conditionally waived without a statically complete not-applicable route.

### WG-TASK-005 — Derive legal transitions while consuming external execution and cancellation facts

- [ ] WG-TASK-005 — Derive legal transitions while consuming external execution and cancellation facts
- **Requirements:** WG-REQ-004, WG-REQ-006, WG-REQ-011
- **dependsOn:** [WG-TASK-004]
- **External dependencies:** IC-03 / WS-REQ-008 supplies replay identity; IC-09 / AV-REQ-014 and AV-REQ-016 supplies structural or execution-backed validation facts; future `sdd-bounded-execution` must publish versioned `ExecutionAttemptFact`, `ExecutionResultFact`, and `ClaimObservation` deliverables. Claims, leases, descendant cleanup, and process execution remain external.
- **Owned files/modules:** `internal/workflow/transition.go`, `internal/workflow/execution_fact.go`, `internal/workflow/cancellation.go`, transition fixtures under `testdata/workflow-graph/transitions/`, and deterministic provider fakes used only in tests.
- **Observable outcome:** authored/structural submissions can complete from matching trusted validation facts without a process provider; execution-backed nodes require matching external attempt/result facts; `claimed` stays a separate observation; exhaustion and cancellation mirror owner facts.
- **Primary fixtures:** WG-F-020, WG-F-021, WG-F-022, WG-F-023, WG-F-024, WG-F-025.
- **Proposed verification:** `go test ./internal/workflow -run 'TestTransitionTable|TestAuthoredNodeWithoutProvider|TestExecutionFactBinding|TestCancellationOwnerOutcome|TestAttemptExhaustion'`.
- **Rollback/recovery checks:** pre-intent cancellation has no authoritative graph mutation; post-intent uncertainty surfaces committed/recovery-required owner state; workflow never kills a process, releases/adopts a claim, deletes staging, rolls back, or retries an effect.
- **Done when:** the normative transition rows are observable through `Evaluate`, mismatched candidates/attempts never satisfy a node, exhaustion blocks ordinary dependents, and cancellation cannot create a state the workspace/provider did not report.

### WG-TASK-006 — Preserve immutable attempts across retry, reconciliation, migration, and roll-forward

- [ ] WG-TASK-006 — Preserve immutable attempts across retry, reconciliation, migration, and roll-forward
- **Requirements:** WG-REQ-004, WG-REQ-005, WG-REQ-008
- **dependsOn:** [WG-TASK-005]
- **External dependencies:** IC-02/IC-03 and WS-REQ-003/WS-REQ-006/WS-REQ-008 deliver persisted revision, transaction, replay, cancellation, and recovery outcomes; workspace allocates/persists attempt IDs and revisions but does not decide workflow semantics. Explicit definition migration requires a reviewed compatibility contract and accepted history mapping.
- **Owned files/modules:** `internal/workflow/attempt.go`, `internal/workflow/reconcile.go`, `internal/workflow/compatibility.go`, and `testdata/workflow-graph/attempts/`.
- **Observable outcome:** validate the complete attempt subject tuple, derive monotonic scoped attempt numbers, retain rejected/failed/cancelled attempts, retry the same immutable topology, and consume durable roll-forward or explicit migration facts without rereading mutable files.
- **Primary fixtures:** WG-F-026, WG-F-027, WG-F-028, WG-F-029, WG-F-030.
- **Proposed verification:** `go test ./internal/workflow -run 'TestAttemptIdentityAndHistory|TestRetryKeepsTopology|TestReconciliationFacts|TestDefinitionMigrationMapping|TestInterruptedRollForward'`.
- **Rollback/recovery checks:** missing/duplicate attempt numbers, tuple changes, or ambiguous durable outcomes are integrity blockers; no prior attempt is overwritten and no retry/rejection/recovery event becomes a graph edge.
- **Done when:** retries create distinct immutable attempts, runtime exception/reconciliation facts leave topology unchanged, explicit migration produces a new pin/digest with history mapping, and interruption resolution uses only the owner-recorded outcome.

### WG-TASK-007 — Make next-action, blockers, freshness explanations, and completion handoff deterministic

- [ ] WG-TASK-007 — Make next-action, blockers, freshness explanations, and completion handoff deterministic
- **Requirements:** WG-REQ-003, WG-REQ-006, WG-REQ-011
- **dependsOn:** [WG-TASK-004, WG-TASK-006]
- **External dependencies:** foundation REQ-002/REQ-008 owns structured `NextAction` rendering; IC-12 / `sdd-completion-lifecycle` CL-REQ-008/CL-REQ-012 must publish the completion gate/handoff consumer contract. Workflow emits a handoff only and cannot satisfy completion acceptance.
- **Owned files/modules:** `internal/workflow/action.go`, `internal/workflow/blocker.go`, `internal/workflow/handoff.go`, canonical result encoding tests, and `testdata/workflow-graph/actions/`.
- **Observable outcome:** identical inputs produce byte-stable states, diagnostics, state digest, primary action, alternatives, and blocker order; default evaluation returns one proposal or one safe owner-provided remediation; graph-complete returns only a completion handoff.
- **Primary fixtures:** WG-F-031, WG-F-033, WG-F-034, WG-F-035, WG-F-036.
- **Proposed verification:** `go test ./internal/workflow -run 'TestEvaluationDeterminism|TestActionTieBreak|TestBlockedAndStaleExplanation|TestCompletionHandoffOnly' -count=100`.
- **Rollback/recovery checks:** shuffle input slice/map insertion order and rerun; actions contain no grant, claim, shell command, staging path, or DONE content; read-only blockers never invoke a repair action.
- **Done when:** tie-breaking is node ID, action ID, then source coordinate; blocker/stale results expose expected/current revision or digest and safe remediation; no graph action remains means a typed completion handoff rather than generated completion artifacts.

### WG-TASK-008 — Resolve composition, policy floors, assessments, and escalation as a three-fact lifecycle

- [ ] WG-TASK-008 — Resolve composition, policy floors, assessments, and escalation as a three-fact lifecycle
- **Requirements:** WG-REQ-005, WG-REQ-007, WG-REQ-008
- **dependsOn:** [WG-TASK-002, WG-TASK-004]
- **External dependencies:** IC-06 is graph-owned; IC-10 / AV-REQ-015/AV-REQ-017 supplies normalized assessment and effective-policy inputs; WS-REQ-003/WS-REQ-004 supplies application revision/exact-plan outcomes. DEC-01/DEC-03/DEC-06 must define trusted issuers, revocation, retention, independence, and waiver authority before production authority claims.
- **Owned files/modules:** `internal/workflow/composition.go`, `internal/workflow/policy.go`, `internal/workflow/escalation.go`, and `testdata/workflow-graph/composition/`.
- **Observable outcome:** resolve journey, effective rigor, profiles, validators, graph and exact policy/definition digests from typed inputs; enforce organization floor > project narrowing > work rationale; keep `unassessed`/`insufficient` explicit; separate escalation proposal, acceptance, and workspace application.
- **Primary fixtures:** WG-F-037, WG-F-038, WG-F-039, WG-F-040, WG-F-041, WG-F-042.
- **Proposed verification:** `go test ./internal/workflow -run 'TestCompositionPrecedence|TestPolicyFailsClosed|TestAssessmentStates|TestEscalationLifecycle|TestEmergencyReconciliation'`.
- **Rollback/recovery checks:** proposal alone never changes active composition, acceptance never persists it, application requires the exact accepted proposal and current revisions, and emergency reconciliation never rewrites topology or removes later gates.
- **Done when:** lower-precedence input can add but not remove requirements, missing/expired/revoked/unverifiable required policy blocks, low-risk input cannot lower a critical floor, invalid escalation authority changes nothing, and history/work identity survive permitted escalation.

### WG-TASK-009 — Enforce pins, explicit compatibility, and edge-scoped freshness

- [ ] WG-TASK-009 — Enforce pins, explicit compatibility, and edge-scoped freshness
- **Requirements:** WG-REQ-002, WG-REQ-003, WG-REQ-005, WG-REQ-008
- **dependsOn:** [WG-TASK-003, WG-TASK-004, WG-TASK-008]
- **External dependencies:** IC-07 / AV-REQ-008/AV-REQ-009/AV-REQ-020 supplies opaque semantic/derivation/evidence validity facts and canonicalizer compatibility; IC-10 supplies policy validity; foundation/workspace supply offline pinned definitions. Workflow must not recompute artifact or evidence digests.
- **Owned files/modules:** `internal/workflow/pin.go`, `internal/workflow/freshness.go`, `internal/workflow/compatibility.go`, reverse-adjacency evaluation tests, and `testdata/workflow-graph/freshness/`.
- **Observable outcome:** require the complete exact pin set offline, report same-name definition updates without adoption, apply a versioned compatibility/history rule, propagate owner-reported invalidation only through applicable freshness edges, and reject direct/indirect self-output readiness cycles.
- **Primary fixtures:** WG-F-043, WG-F-044, WG-F-045, WG-F-046, WG-F-047.
- **Proposed verification:** `go test ./internal/workflow -run 'TestPinSetFailsClosed|TestDefinitionUpdateAvailable|TestCompatibilityRules|TestFreshnessPropagation|TestSelfOutputReadinessCycle'`.
- **Rollback/recovery checks:** no display-name/newest-version/network fallback; a stale fact records source edge, derivation inputs, expected/current digest, and regeneration action; dependency edges are reevaluated for success but never silently promoted to freshness relations.
- **Done when:** missing pins close the route, behavior-changing same-name edits require explicit migration, cosmetic equivalence is accepted only from the owning canonicalizer, and unrelated nodes remain current when no declared freshness path reaches them.

### WG-TASK-010 — Submit typed facts through workspace replay and validation contracts

- [ ] WG-TASK-010 — Submit typed facts through workspace replay and validation contracts
- **Requirements:** WG-REQ-004, WG-REQ-005, WG-REQ-009, WG-REQ-011
- **dependsOn:** [WG-TASK-005, WG-TASK-006, WG-TASK-009]
- **External dependencies:** IC-07/IC-09 / AV-REQ-003/AV-REQ-004/AV-REQ-014/AV-REQ-016 deliver normalized submission and validation facts; IC-02/IC-03 / WS-REQ-004/WS-REQ-006/WS-REQ-008 deliver exact transition application, replay, conflict, and recovery outcomes. Artifact parsing/digests/validator execution and workspace persistence remain owner implementations.
- **Owned files/modules:** `internal/workflow/submission.go`, `internal/workflow/validation_fact.go`, `internal/engine/workflow_submit.go`, test-only artifact/workspace adapters, and `testdata/workflow-graph/submissions/`.
- **Observable outcome:** bind a submission and validation result to artifact, action, candidate/input, validator contract, policy, snapshot, and revisions; emit a mutation-neutral `TransitionIntent`; interpret committed/replayed/conflict/recovery outcomes without a second parser or idempotency store.
- **Primary fixtures:** WG-F-048, WG-F-049, WG-F-050, WG-F-051, WG-F-052, WG-F-053.
- **Proposed verification:** `go test ./internal/workflow ./internal/engine -run 'TestDraftSubmission|TestAcceptanceSelfDependency|TestValidationFactBinding|TestRejectedRetryHistory|TestSubmissionReplayAndConflict'`.
- **Rollback/recovery checks:** valid unaccepted drafts may enter validation; missing/forged/stale/truncated evidence blocks; same-key/same-fingerprint replays the workspace result; changed fingerprint or lost expected revision conflicts; uncertain outcome exposes recovery and never performs a second effect.
- **Done when:** agent `accepted`/`passed` claims have no authority, rejection history remains immutable, a new candidate gets a new attempt, and the engine delegates the exact transition to workspace with expected coordinates instead of staging or writing locally.

### WG-TASK-011 — Evaluate initiatives and partial compound progress without cursor authority

- [ ] WG-TASK-011 — Evaluate initiatives and partial compound progress without cursor authority
- **Requirements:** WG-REQ-003, WG-REQ-005, WG-REQ-006, WG-REQ-010
- **dependsOn:** [WG-TASK-005, WG-TASK-006, WG-TASK-009]
- **External dependencies:** IC-01/IC-02/IC-03 and WS-REQ-002/WS-REQ-006/WS-REQ-008 supply canonical leaf progress, transaction, replay, and recovery facts. Workspace owns persistence and atomic bundle outcome; graph owns only stable aggregation and eligibility.
- **Owned files/modules:** `internal/workflow/initiative.go`, `internal/workflow/aggregate.go`, `internal/workflow/resume.go`, and `testdata/workflow-graph/initiatives/`.
- **Observable outcome:** preserve stable parent/leaf IDs, aggregate status deterministically from every applicable leaf, keep dependency/freshness meanings separate, allow unrelated leaves to proceed, scan all unresolved leaves on resume, and propose retry only for eligible failed/unresolved leaves.
- **Primary fixtures:** WG-F-054, WG-F-055, WG-F-056, WG-F-057.
- **Proposed verification:** `go test ./internal/workflow -run 'TestInitiativeIdentityAndAggregate|TestLeafFailureIsolation|TestResumeScansAllLeaves|TestPartialRetryAndAtomicBundle'`.
- **Rollback/recovery checks:** a display cursor cannot hide an eligible leaf or replay a completed effect; one failed/cancelled leaf cannot cancel an unrelated leaf; an atomic bundle is consumed as one workspace outcome and is never decomposed by graph recovery.
- **Done when:** aggregate status follows the documented precedence and exposes counts/blockers, retry retains parent/leaf identities, completed leaves remain complete, dependents block or conditionally skip for an explicit reason, and only unresolved eligible leaves are proposed.

### WG-TASK-012 — Publish graph results through the foundation protocol and verify owner-bound recovery behavior

- [ ] WG-TASK-012 — Publish graph results through the foundation protocol and verify owner-bound recovery behavior
- **Requirements:** WG-REQ-006, WG-REQ-007, WG-REQ-008, WG-REQ-009, WG-REQ-010, WG-REQ-011
- **dependsOn:** [WG-TASK-007, WG-TASK-008, WG-TASK-009, WG-TASK-010, WG-TASK-011]
- **External dependencies:** AM-02 / WG-A-001 must deliver one foundation-owned minor schema allocation for graph result variants under REQ-002/REQ-003/REQ-009/REQ-010; IC-01/02/03 must be published by workspace; IC-07/09 by artifact-validation; IC-11/12 by completion; bounded execution must publish provider/claim fact schemas. No local version number may be assigned.
- **Owned files/modules:** graph-owned nested schema definitions under the foundation-selected `protocol/schema/` extension location, `internal/engine/workflow.go`, graph engine consumer tests, public valid/invalid fixtures, and `testdata/workflow-graph/engine/`; foundation retains renderer and envelope implementation ownership.
- **Observable outcome:** `start`, read-only `next`, `submit`, and `finish` handoff use the shared envelope, source coordinates, severities, structured actions, revisions, redaction, and full exit taxonomy; recovery reports only owner-provided safe actions and changed candidates invalidate proposals.
- **Primary fixtures:** WG-F-059, WG-F-060, WG-F-061.
- **Proposed verification:** `go test ./internal/engine ./internal/presentation/... -run 'TestWorkflowProtocolConformance|TestWorkflowExitTaxonomy|TestReadOnlyRecoveryReport|TestWorkflowCandidateStaleness'` plus the foundation-selected offline JSON Schema validation command over every graph valid/invalid fixture.
- **Rollback/recovery checks:** `next` stops before plan/authorize/apply; read-only recovery does not clean staging, append events, release locks, adopt leases, or apply a plan; submit rechecks expected workspace/work/action and candidate coordinates through the owner seam.
- **Done when:** human and JSON views consume one outcome, every exit category `0/1/2/3/4/5/6/130` is selected from typed facts rather than message text, stdout/stderr and redaction rules hold, unknown result/version negotiation fails safely, and no production conformance claim is made for an unavailable external owner.

### WG-TASK-013 — Rebuild identical graph state from canonical facts after projection loss

- [ ] WG-TASK-013 — Rebuild identical graph state from canonical facts after projection loss
- **Requirements:** WG-REQ-005, WG-REQ-006, WG-REQ-008, WG-REQ-009, WG-REQ-010, WG-REQ-012
- **dependsOn:** [WG-TASK-006, WG-TASK-007, WG-TASK-008, WG-TASK-009, WG-TASK-010, WG-TASK-011]
- **External dependencies:** IC-01 / WS-REQ-002 and WS-REQ-007/WS-REQ-010 supply canonical snapshot/journal references and distinguish disposable projections; artifact-validation and completion supply their canonical owner facts. Workflow does not own projection persistence or read-only repair.
- **Owned files/modules:** `internal/workflow/rebuild.go` only if a named pure rebuild entry is needed behind `Evaluate`, canonical encoding tests, and `testdata/workflow-graph/rebuild/`; no cache database or repair adapter.
- **Observable outcome:** reconstruct graph state solely from pins/composition, immutable attempts, owner submission/validation facts, initiative facts, and workspace references; compare byte-identical state, diagnostics, digests, and ordered actions before and after every derived view is removed.
- **Primary fixtures:** WG-F-062, WG-F-063.
- **Proposed verification:** `go test ./internal/workflow -run 'TestCanonicalFactRebuild|TestProjectionLossDoesNotChangeEvaluation' -count=100`.
- **Rollback/recovery checks:** absent/corrupt derived storage returns `WORKFLOW_CACHE_REBUILD_REQUIRED` or an owner-provided read-only diagnostic; evaluation never silently mutates, repairs workspace state, allocates revisions, or treats a cache as authority.
- **Done when:** deletion of dense indexes, selection views, and optional caches changes neither correctness nor output bytes, and any unavailable external projection is reported without changing canonical facts.

### WG-TASK-014 — Run the adversarial corpus and final ownership, traceability, and DAG audit

- [ ] WG-TASK-014 — Run the adversarial corpus and final ownership, traceability, and DAG audit
- **Requirements:** WG-REQ-001, WG-REQ-002, WG-REQ-003, WG-REQ-004, WG-REQ-005, WG-REQ-006, WG-REQ-007, WG-REQ-008, WG-REQ-009, WG-REQ-010, WG-REQ-011, WG-REQ-012
- **dependsOn:** [WG-TASK-012, WG-TASK-013]
- **External dependencies:** every graph-produced or graph-consumed contract among IC-01 through IC-12 must have its owner-published schema and consumer conformance result (IC-08 remains outside graph implementation); AM-01/AM-02 and WG-A-001 through WG-A-005 must be resolved or recorded as release blockers. Fake-based evidence remains labeled test-only.
- **Owned files/modules:** `testdata/workflow-graph/manifest.json`, `internal/workflow/conformance_test.go`, engine consumer-contract tests, definition/predicate/fact fuzz targets, and the workflow-graph traceability audit record; no producer implementation for an external contract.
- **Observable outcome:** one manifest assigns each normative fixture to exactly one primary slice, records preconditions/action/expected result/forbidden side effects, verifies every requirement and acceptance ID is traceable, and runs the complete malicious/boundary corpus through package or engine interfaces as appropriate.
- **Audit corpus (not a second primary assignment):** WG-F-001, WG-F-002, WG-F-003, WG-F-004, WG-F-005, WG-F-006, WG-F-007, WG-F-008, WG-F-009, WG-F-010, WG-F-011, WG-F-012, WG-F-013, WG-F-014, WG-F-015, WG-F-016, WG-F-017, WG-F-018, WG-F-019, WG-F-020, WG-F-021, WG-F-022, WG-F-023, WG-F-024, WG-F-025, WG-F-026, WG-F-027, WG-F-028, WG-F-029, WG-F-030, WG-F-031, WG-F-032, WG-F-033, WG-F-034, WG-F-035, WG-F-036, WG-F-037, WG-F-038, WG-F-039, WG-F-040, WG-F-041, WG-F-042, WG-F-043, WG-F-044, WG-F-045, WG-F-046, WG-F-047, WG-F-048, WG-F-049, WG-F-050, WG-F-051, WG-F-052, WG-F-053, WG-F-054, WG-F-055, WG-F-056, WG-F-057, WG-F-058, WG-F-059, WG-F-060, WG-F-061, WG-F-062, WG-F-063.
- **Proposed verification:** `go test ./internal/workflow ./internal/engine ./internal/presentation/... -run 'TestWorkflowFixtureManifestCoverage|TestWorkflowRequirementTraceability|TestWorkflowOwnerBoundaries|TestWorkflowProtocolConformance'`, `go test -race ./internal/workflow ./internal/engine`, and the approved bounded fuzz jobs for definition decoding, compilation, predicates, and owner-fact reconciliation.
- **Rollback/recovery checks:** inject cancellation between pure passes and owner boundaries; assert no forbidden file/process/claim/completion action; retain exact failed input and seed for reproduction without persisting secrets.
- **Done when:** the manifest has no duplicate/missing fixture ownership, all requirements and public variants have observable coverage or an explicit external blocker, the local task DAG is acyclic, package imports respect dependency direction, schema/examples validate under the approved tool, and the review record distinguishes commands actually run from tests still proposed.

## Local dependency DAG

The `dependsOn` fields above are authoritative and contain local task IDs only.
The intended partial order is:

```text
WG-TASK-001 -> WG-TASK-002 -> WG-TASK-003 -> WG-TASK-004 -> WG-TASK-005
                                                        |          |
                                                        |          +-> WG-TASK-006
                                                        |                  |
                                                        +-> WG-TASK-008    +-> WG-TASK-007
                                                                 |        |
                                                                 +-> WG-TASK-009
                                                                          |
                                                   +----------------------+----------------+
                                                   |                      |                |
                                             WG-TASK-010            WG-TASK-011      WG-TASK-013
                                                   |                      |                |
                                                   +----------> WG-TASK-012 <---------------+
                                                                      |
                                                                WG-TASK-014
```

This illustration is non-normative; implementations and schedulers must derive
order from the per-task `dependsOn` fields and reject an unknown ID,
self-dependency, duplicate edge, or cycle.

## Cross-spec blockers and release gates

| Blocker | Exact external deliverable | Tasks affected | Honest status rule |
|---|---|---|---|
| WG-BLK-001 | AM-02 / WG-A-001: foundation-owned reviewed minor extension for graph command/result variants, diagnostics, negotiation, and full exit mapping | WG-TASK-001, WG-TASK-012, WG-TASK-014 | Internal/fake engine tests may proceed; public protocol conformance and release remain blocked. |
| WG-BLK-002 | IC-01/IC-02/IC-03 and WG-A-002: workspace immutable snapshot/revisions, exact-plan/transition outcome, replay, cancellation, and recovery schemas plus real conformance evidence | WG-TASK-004 through WG-TASK-014 | Pure evaluation may proceed with values/fakes; no durability, replay, cancellation, or recovery acceptance claim. |
| WG-BLK-003 | IC-07/IC-09 and WG-A-003: artifact-validation submission, validity, semantic/derivation/evidence, assessment, and validation request/result schemas plus validator conformance | WG-TASK-004, WG-TASK-008, WG-TASK-009, WG-TASK-010, WG-TASK-012 through WG-TASK-014 | Structural shapes may be tested; no parser, digest, authority, validator, or evidence acceptance claim. |
| WG-BLK-004 | IC-11/IC-12 and WG-A-004: completion candidate/evidence/gate and completion-handoff schemas plus completion consumer conformance | WG-TASK-007, WG-TASK-012, WG-TASK-014 | Handoff shape may be tested; no DONE, receipt, promotion, archive, or completion claim. |
| WG-BLK-005 | WG-A-005: bounded-execution versioned execution-attempt/result/claim facts and real process-provider contract tests | WG-TASK-005, WG-TASK-012, WG-TASK-014 | Authored/structural paths remain implementable; execution routes stay unavailable and fail closed in production. |
| WG-BLK-006 | DEC-01/DEC-03/DEC-06: trusted policy/assessment/reviewer/waiver issuers, revocation, retention, privacy, and independence decisions | WG-TASK-008, WG-TASK-010, WG-TASK-012, WG-TASK-014 | Explicit injectable facts support tests; no organization-governed authority claim. |
| WG-BLK-007 | Approved graph serialization/schema ID, compiler-contract version, and corpus-backed hard-limit review | WG-TASK-002, WG-TASK-003, WG-TASK-009, WG-TASK-012, WG-TASK-014 | Provisional implementation/fixtures must record the selected draft identity and cannot freeze compatibility. |

## Exact amendments required in existing guided consumer tasks

These are consumer-task amendments, not implementation owned by this spec. They
must be reviewed in `.kiro/specs/sdd-guided-workflows/tasks.md`; this plan does
not silently edit or mark them complete.

- **GW-001:** retain Bug Fix/Quick Spec composition inputs and journey-specific assertions, but consume graph-owned `ResolveComposition`, `PinSet`, policy-floor, and definition-identity contracts instead of freezing a second generic composition/digest model.
- **GW-002:** replace generic node/edge/gate compiler implementation with authored maintained journey definitions and consumer conformance fixtures against IC-04 / `workflow.Compile`; remove deterministic-adjacency/compiler ownership from guided workflows.
- **GW-003:** retain journey-specific routing and preservation assertions, but consume IC-06 proposal/acceptance/application results; do not implement another precedence/escalation engine.
- **GW-004 through GW-007 and GW-010 through GW-011:** own specialized Bug Fix/Quick Spec content and actions only; consume the common artifact envelope and graph state/action contracts rather than introducing parser, generic graph, or transaction behavior.
- **GW-008:** retain journey-specific regression/verification/policy-review assertions; consume IC-09 execution evidence and IC-11 completion evidence acceptance. A deterministic fake remains test-only and cannot prove production execution.
- **GW-009:** retain emergency journey constraints and retrospective gates; consume graph composition/reconciliation plus policy authority facts, without rewriting topology or owning exact-plan application.
- **GW-012:** replace bundle parser/canonicalizer/digest implementation with Quick Spec profile/traceability consumer integration against IC-07 and artifact-validation; guided workflows own bundle-specific required links, not the common parser.
- **GW-013:** replace authoritative validator execution ownership with selection of journey-required validators and consumer conformance against IC-09; artifact-validation interprets results and bounded execution owns processes.
- **GW-014:** retain atomic three-artifact bundle semantics and journey retry assertions, but delegate parsing/validation to artifact-validation and transaction/idempotency/recovery to IC-02/IC-03 workspace contracts.
- **GW-015:** consume IC-05 graph state/action output and IC-12 completion handoff; retain only the Quick Spec journey transition assertions and rejection/reapproval route.
- **GW-016:** retain journey-specific gate inputs, but completion owns gate acceptance, DONE/receipt/promotion/archive proposal and workspace owns application; guided handlers request and consume those results.
- **GW-018:** contribute guided command payloads to the one foundation-owned AM-02 extension inventory; remove independent minor-version allocation and consume the shared graph action/submission coordinates.
- **GW-019:** gate guided release on real prerequisite producer/consumer conformance, including this spec's fixture manifest; fake-only graph, validation, workspace, completion, or execution tests must remain explicit blockers.

## Final acceptance checklist

- [ ] Every task heading, checkbox label, and `dependsOn` reference uses one unique `WG-TASK-NNN` ID, and the derived local DAG is acyclic.
- [ ] Every `dependsOn` entry is local; every external contract appears under **External dependencies** or the blocker table with its exact IC/spec/amendment deliverable.
- [ ] WG-REQ-001 through WG-REQ-012 each map to one or more observable task outcomes and proposed verification commands.
- [ ] WG-F-001 through WG-F-063 each have exactly one **Primary fixtures** assignment, and the final manifest records preconditions, action, expected result, and forbidden side effect.
- [ ] The definition schema and all public valid/invalid result fixtures pass the approved offline schema validator; no task invents a protocol version.
- [ ] Compilation/evaluation output is deterministic under input-order shuffling, repeated runs, cancellation, cache loss, and platform-independent stable-ID ordering.
- [ ] All read-only graph operations prove no workspace write, event append, revision allocation, repair, lock/claim/lease mutation, process launch, or completion effect.
- [ ] Real workspace, artifact-validation, completion, and bounded-execution consumer-contract tests are distinguished from deterministic fakes; unavailable real contracts remain release blockers.
- [ ] Human and JSON output consume the same foundation outcome, preserve stdout/stderr and redaction rules, and use the complete exit taxonomy without parsing messages.
- [ ] Import and ownership review finds no workspace transaction, Markdown/YAML artifact parser, validator runner, completion engine, process supervisor, lease manager, or guided journey topology inside the graph module.
- [ ] The exact guided task amendments above are reviewed before any guided prerequisite-conformance task is claimed complete.
- [ ] The implementation review records only commands actually run and their results; unchecked or blocked evidence is never described as acceptance.

## For the chosen ones

The task order deliberately makes the first walking slice *less complete but
more truthful*: it proves the deep seam before all owner contracts exist, then
replaces test-only facts with producer conformance one contract at a time. The
hidden failure mode in graph engines is not usually a bad topological sort; it
is authority accretion. A parser, replay cache, lease helper, or DONE shortcut
added “temporarily” lets the graph manufacture evidence its inputs were meant
to prove. Keeping every external fact opaque and every proposal effect-free is
what allows the compiler and evaluator to become deep without becoming a
second control plane.
