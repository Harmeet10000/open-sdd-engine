# Implementation Tasks: sdd-guided-workflows

## Scope and prerequisites

These tasks implement the two maintained journey compositions through the foundation `Engine.Execute` seam. They depend on contracts from `sdd-engine-foundation`, `sdd-workspace-transactions`, `sdd-workflow-graph`, `sdd-artifact-validation`, and `sdd-completion-lifecycle`. Until the latter four real specs conform, use deterministic fakes for draft work but do not claim durable transaction, authoritative validation, graph, or completion integration.

Task IDs below are stable identifiers. `dependsOn` is the complete DAG; waves are derived, not hand-maintained. Each task is a vertical observable slice, not a layer batch.

## Tasks

### GW-001 — Freeze composition, policy, and definition-digest contract

- [ ] GW-001 — Freeze composition, policy, and definition-digest contract
- **Requirements:** 1, 13, 17, 18, 21
- **dependsOn:** []
- Define journey/rigor/profile/preset values, policy precedence, fail-closed missing/expired/revoked behavior, immutable work ID versus slug/revision, exact workflow/policy/schema/validator definition digests, routing evidence, escalation persistence, and an optional pinned scope-priority profile/baseline reference. MoSCoW is explicitly orthogonal to rigor, severity, readiness, and dependency order.
- **Done when:** composition fixtures show explicit override cannot lower policy floor, escalation preserves work identity/artifact IDs, and a priority baseline cannot silently reclassify a Must.

### GW-002 — Compile graph node, edge, and gate schema

- [ ] GW-002 — Compile graph node, edge, and gate schema
- **Requirements:** 1, 2, 8
- **dependsOn:** [GW-001, GW-018]
- Define versioned graph records with stable node IDs, typed inputs/outputs, gate predicates, profile-selected optional design/policy-review nodes, and rejection/revision transitions. Reject unknown critical fields, dangling refs, duplicate IDs, unsafe cycles, and `implementation-ready -> finish`.
- **Done when:** graph compilation produces deterministic adjacency and diagnostics for invalid Bug Fix/Quick Spec topologies.

### GW-003 — Implement composition routing and escalation vertical slice

- [ ] GW-003 — Implement composition routing and escalation vertical slice
- **Requirements:** 1, 13, 17, 18
- **dependsOn:** [GW-001, GW-002]
- Resolve preset/risk/profile from trusted policy plus evidence; retain Bug Fix for critical bug risk; emit structured escalation and create a new revision preserving accepted answers/history/valid artifacts. Require typed exception fields and trust-boundary metadata.
- **Done when:** critical auth, migration, trust boundary, concurrency/crash recovery, privacy, and policy-floor fixtures escalate without lowering or losing artifacts.

### GW-004 — Implement Bug Fix report and reproduction slice

- [ ] GW-004 — Implement Bug Fix report and reproduction slice
- **Requirements:** 2, 3, 4
- **dependsOn:** [GW-002, GW-003]
- Add typed bug-report and reproduction artifacts with stable IDs, source/provenance, environment/snapshot, attempts, closed statuses, and explicit unassessed disposition. Block diagnosis for unassessed outcomes until approved safe proxy/simulation/incident evidence/exception exists.
- **Done when:** reproduced, intermittent, environment-specific, not-reproduced, unsafe-to-reproduce, unsafe proxy, and symptom-only fixtures produce correct next actions and diagnostics.

### GW-005 — Implement investigation, diagnosis, and assessment evidence slice

- [ ] GW-005 — Implement investigation, diagnosis, and assessment evidence slice
- **Requirements:** 2, 5, 6
- **dependsOn:** [GW-004]
- Track observations, competing hypotheses, discriminating evidence, effects, mechanism, scopes, accepted-spec relationship, confidence, provenance, and uncertainty. Keep deterministic structural/policy checks separate from advisory root-cause assessment.
- **Done when:** passing reproduction without discriminating evidence is rejected; temporary effects require retain/remove/block disposition.

### GW-006 — Implement regression-contract and conditional-design slice

- [ ] GW-006 — Implement regression-contract and conditional-design slice
- **Requirements:** 2, 5, 15
- **dependsOn:** [GW-005]
- Define the observable regression contract before implementation, with failing-before expectation, passing-after expectation, negative/safety cases, and traceability. Resolve design requiredness from policy/profile, not agent selection.
- **Done when:** contract exists without claiming verification, and a policy-required design blocks implementation until accepted.

### GW-007 — Implement Bug Fix tasks and shared implementation handoff

- [ ] GW-007 — Implement Bug Fix tasks and shared implementation handoff
- **Requirements:** 2, 15
- **dependsOn:** [GW-006]
- Produce tasks with requirement/diagnosis/contract/design links, deterministic dependency DAG, rollback/cleanup work, and verification coverage. Hand off to shared `implement -> verify -> policy-review -> finish` lifecycle; no direct contract/diagnosis-to-finish path.
- **Done when:** implement is unavailable without required gates, while emergency path is explicitly scoped and still reconciles before finish.

### GW-008 — Implement independent verification artifact and policy-review gate

- [ ] GW-008 — Implement independent verification artifact and policy-review gate
- **Requirements:** 2, 5, 14, 15, 19
- **dependsOn:** [GW-007]
- Run/record verification against candidate snapshot after implementation, with commands/checks, result, environment, output digest, candidate digest, and provenance. Execute required policy/profile gates and reject stale/dirty candidate changes. Keep verification artifact separate from regression contract.
- **Done when:** forged agent `passed`, missing result, stale candidate, and dirty post-verification code cannot satisfy verify; policy-review diagnostics are stable.

### GW-009 — Implement typed emergency authorization and reconciliation

- [ ] GW-009 — Implement typed emergency authorization and reconciliation
- **Requirements:** 7, 15
- **dependsOn:** [GW-003, GW-008]
- Validate identified actor, grant scope, policy snapshot, expiry/revocation, starting snapshot/revision, exact plan digest, and preauthorization. Permit only scoped mitigation/immediate verification; track deployment separately; require retrospective gates before finish.
- **Done when:** owner-string/Boolean, expired/revoked/forged exceptions and plan mismatch fail closed; authorized emergency fixture cannot finish early.

### GW-010 — Implement Quick Spec material-unknown discovery slice

- [ ] GW-010 — Implement Quick Spec material-unknown discovery slice
- **Requirements:** 8, 9, 13, 19
- **dependsOn:** [GW-002, GW-003]
- Gather trusted repository/spec/config/read-only evidence, classify material questions with stable IDs and evidence/provenance, and close zero-question paths. Emit an explicit `assess-materiality-and-risk` action when evidence cannot determine risk/materiality; keep heuristics advisory where policy requires human/agent assessment.
- **Done when:** zero, one, several, already-known, and unresolved material questions plus external assessment produce correct structured next actions.

### GW-011 — Implement generation packet and host-agent fallback

- [ ] GW-011 — Implement generation packet and host-agent fallback
- **Requirements:** 8, 9, 10, 19, 21
- **dependsOn:** [GW-010]
- Emit `generate-spec-bundle` packet with intent, answers/evidence, accepted specs, trusted policy snapshot, pinned scope-baseline reference/digest, resolved snapshot reference, content digests for each context item, schemas/IDs, canonical paths, expected workspace/work/action revisions, budgets, provenance, validator definitions, and completion conditions. Return bounded digest-only references plus structured remediation on budget overflow; return JSON/manual instructions when no host agent exists; invoke no model.
- **Done when:** packet is bounded, offline, schema-valid, carries the immutable scope baseline, and does not claim generated artifacts are accepted.

### GW-012 — Implement bundle parsing, canonicalization, and traceability

- [ ] GW-012 — Implement bundle parsing, canonicalization, and traceability
- **Requirements:** 10, 11, 12, 21
- **dependsOn:** [GW-011]
- Consume the artifact-validation typed values for requirements/design/tasks and scope-priority records; compute no competing parser/digest implementation; validate stable IDs, source/work revision, requirement/design/task links, deterministic task DAG, implementation/verification coverage, closed MoSCoW values and baseline references.
- **Done when:** malformed, duplicate, missing-link, semantic-no-op, behavior-changing, unknown-field, unscoped/invalid-priority, and changed-definition fixtures yield deterministic freshness/traceability diagnostics.

### GW-013 — Implement authoritative validator execution

- [ ] GW-013 — Implement authoritative validator execution
- **Requirements:** 11, 14
- **dependsOn:** [GW-012]
- Derive mandatory validator IDs from trusted policy/profile and candidate, execute each against immutable snapshot, and record canonical `validatorsExecuted` with validator/definition/candidate digests, status, evidence refs, and execution provenance. Agent claims are never authority.
- **Done when:** empty, missing, stale, cancelled, unverifiable, forged, and changed-definition validator evidence fails closed; valid full execution passes.

### GW-014 — Implement atomic draft submission, approval, retry, and recovery

- [ ] GW-014 — Implement atomic draft submission, approval, retry, and recovery
- **Requirements:** 11, 12, 13, 14, 21
- **dependsOn:** [GW-012, GW-013, GW-003, GW-020]
- Submit all three artifacts and the pinned scope-baseline reference as one transaction; reject all on any failure; bind approval to revision/digests/plan/baseline; same-key same-content replays; same-key changed content conflicts; uncertain crash returns recovery-required. Revisions require revalidation and reapproval.
- **Done when:** partial/malformed/stale/rejected/revised/approved/uncertain submissions preserve prior state and never partially advance; a changed priority baseline invalidates the old approval.

### GW-015 — Implement Quick Spec shared-lifecycle handoff

- [ ] GW-015 — Implement Quick Spec shared-lifecycle handoff
- **Requirements:** 8, 13, 16, 21
- **dependsOn:** [GW-014, GW-008]
- Transition only accepted bundle to `implementation-ready`, then shared implement/verify/policy-review/finish. Rejected/revised bundles return to authoring; no direct finish transition. Escalation changes profile/gates while preserving work ID/artifacts. The active baseline makes Must scope a completion floor; Should/Could are deferred only through their recorded dispositions.
- **Done when:** an accepted bundle cannot finish before implementation and independent verification, a Must cannot disappear through a status edit, and all rejection/reapproval routes are explicit.

### GW-016 — Implement completion input validation and atomic outputs

- [ ] GW-016 — Implement completion input validation and atomic outputs
- **Requirements:** 6, 7, 15, 16, 21
- **dependsOn:** [GW-008, GW-009, GW-015]
- Have the shared completion owner validate all journey-specific gates, freshness, candidate digest, policy review when selected, uncertainty/effect disposition, accepted-spec convergence, and the pinned MoSCoW baseline before one transaction creates `done.md`, machine receipt, and lifecycle events. Guided handlers request this operation; existing output files cannot satisfy inputs.
- **Done when:** completion failure creates no partial done/receipt/events; success produces all atomically; post-verification edit blocks; DONE distinguishes delivered, deferred, and Won't-have scope.

### GW-017 — Add full journey conformance fixtures and renderers

- [ ] GW-017 — Add full journey conformance fixtures and renderers
- **Requirements:** 14, 19, 20, 21
- **dependsOn:** [GW-004, GW-005, GW-006, GW-008, GW-009, GW-010, GW-011, GW-013, GW-014, GW-015, GW-016, GW-020]
- Exercise full Bug Fix and Quick Spec paths end-to-end, all required failure/uncertainty/recovery/authority/security and MoSCoW scope-governance cases, stdout/stderr and JSON equivalence, cancellation, and redaction. Keep fixtures in `testdata`; executable tests live in packages.
- **Done when:** each required fixture has a stable code, exit category, structured remediation, and no reliance on English parsing.

### GW-018 — Freeze guided command payload contracts

- [ ] GW-018 — Freeze guided command payload contracts
- **Requirements:** 8, 10, 11, 13, 16, 19, 21
- **dependsOn:** [GW-001]
- Before implementation slices, contribute guided payload definitions to the foundation-owned shared extension inventory: `StartCommand`, read-only `NextCommand`, common `SubmitCommand`, and `FinishCommand` result variants. `SubmitCommand` supports ordinary Bug Fix artifacts, an atomic draft bundle, materiality/risk assessments, a scope-baseline decision, and approvals. Mutations carry workspace/work expected revisions; action submissions carry action ID/revision; retries carry idempotency keys. Claims follow policy but are not approval evidence. Bind transitions to content/plan/candidate/baseline digests and recovery results; test version negotiation and both journey submission paths.
- **Done when:** same-key replay, changed-content or changed-baseline conflict, stale action revision, missing claim, and uncertain-transition fixtures are structurally specified without relying on human messages.

### GW-020 — Implement governed MoSCoW scope-baseline planning

- [ ] GW-020 — Implement governed MoSCoW scope-baseline planning
- **Requirements:** 10, 16, 19, 20, 21
- **dependsOn:** [GW-010, GW-011, GW-012]
- Consume profile-validated `ScopePriority` records and trusted scope policy to produce a deterministic baseline: one closed category per stable subject/scope reference, rationale/evidence, required workaround or exclusion data, comparable non-double-counted capacity totals where available, pinned unit/conversion rule, predecessor digest/revision, and a baseline digest. Default policy enforces the 60% Must cap only with declared comparable units; otherwise return explicit `scope-budget-unverified`. When capacity contracts, produce a `scope-replan-required` proposal that considers Could then Should deferrals without mutating the current baseline. Preserve a Must until an authorized baseline revision changes it. Model feature flags only as rollout metadata with cohort/fallback/observability/owner/removal conditions.
- **Done when:** manual-workaround, must-inflation, mixed-unit/shared-estimate, duplicate-subject, unknown-capacity, Could/Should deferral, unauthorized Must demotion, implemented-Won't-have visibility, and feature-flag-not-completion fixtures produce deterministic structured decisions; default human output stays concise and JSON carries the complete baseline facts.

### GW-019 — Guided release gate and prerequisite conformance review

- [ ] GW-019 — Guided release gate and prerequisite conformance review
- **Requirements:** 1–21
- **dependsOn:** [GW-017, GW-018, GW-020]
- Verify graph acyclicity/reference resolution, cross-document schema/example consistency, package dependency direction, canonical layout, definition binding, and prerequisite release gates. Record blocked integration claims when a prerequisite is not accepted.
- **Done when:** a reviewer can trace every requirement to an observable task/fixture or an explicitly blocked prerequisite contract.

## Dependency DAG and derived waves

The `dependsOn` fields above are authoritative. The scheduler derives waves at runtime and rejects unknown IDs, duplicate IDs, self-dependencies, or cycles; no hand-written wave list is normative.

## Verification checklist

- No task references an absent ID; the DAG is acyclic and stable.
- Bug Fix graph contains independent `regression-contract`, post-implementation `verify`, `policy-review`, and finish gate.
- Quick Spec authoring explicitly hands off to shared implementation lifecycle; rejection/revision/reapproval are represented.
- Authoritative validator IDs derive from trusted policy/profile and are recorded as `validatorsExecuted`.
- Completion validates inputs before atomic `done.md`/receipt/events generation.
- Canonical product layout and immutable identity are used in all examples.
- Required fixtures cover stale writers, same-key changed requests, crash uncertainty, dirty candidate, exception authority, policy removal, changed definitions, unsafe proxy, escalation preservation, forged validators, and end-to-end journeys.

## References

- `requirements.md`
- `design.md`
- `docs/architecture/guided-workflows.md`
- `docs/kiro-input/sdd-cli-spec-brief.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/cli-ux.md`
