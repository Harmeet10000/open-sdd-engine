# Implementation Tasks: sdd-completion-lifecycle

## Provisional scope and execution rules

These are proposed implementation tasks derived from the provisional
requirements and design. They are not approval, implemented work, passing
tests, or permission to publish a protocol version. Every checkbox is
intentionally unchecked.

Tasks are vertical slices through the deep completion interface and, where
named, `Engine.Execute`. The local `dependsOn` field contains only `CL-TASK-*`
IDs. Upstream specifications, interface-ledger entries, company decisions and
consumer tasks are listed separately under **External dependencies** so the
local DAG does not pretend to schedule another owner's work.

A deterministic fake may prove a seam during an early slice. It cannot satisfy
production durability, provider-execution, reviewer-authority or native
platform acceptance. No task may implement artifact parsing/digests, workflow
routing, process supervision/claims, workspace recovery or a second protocol
envelope.

## Tasks

### CL-TASK-001 — Walk a finish request from Engine.Execute to one exact proposal

- [ ] CL-TASK-001 — Walk a finish request from Engine.Execute to one exact proposal
- **Requirements:** CL-REQ-001, CL-REQ-008, CL-REQ-012
- **dependsOn:** []
- **External dependencies:** `IC-01`, `IC-02`, `IC-05`, `IC-07`, `IC-10`, `IC-11`, `IC-12`; `sdd-engine-foundation/FND-003` shared Outcome, `FND-007` execution/error rail and `FND-012` CLI stream/exit behavior; `sdd-workflow-graph/WG-TASK-007` completion handoff; `sdd-workspace-transactions/WS-TASK-001` coherent snapshot and `WS-TASK-002` exact-plan/fake apply seam. The fake workspace proves shape only.
- **Owned files/modules:** `internal/completion/evaluate.go`, `decision.go`, `gate.go`, a completion handler adapter consumed by the foundation-owned engine dispatch, focused completion/engine contract tests and `testdata/completion/walking/`; no public protocol schema publication. Foundation owns the `Engine.Execute` command dispatch and CLI registration.
- **Proposed verification:** `go test ./internal/completion ./internal/engine -run 'TestFinishWalkingSlice|TestFinishBlockedNoMutation'`; fixture `CL-FIX-003` with a minimal missing gate, plus an accepted synthetic fact set that yields exact bytes/effects but is applied only by a deterministic fake.
- **Done when:** one public finish request uses one coherent supplied snapshot, returns ordered blockers with no mutation when a gate is absent, and returns one inspectable exact proposal when all synthetic facts pass; the outcome does not claim `completed` until the fake owner reports committed, and the task records that real durability remains blocked.

### CL-TASK-002 — Select and hash policy-bounded candidate read sets

- [ ] CL-TASK-002 — Select and hash policy-bounded candidate read sets
- **Requirements:** CL-REQ-002, CL-REQ-003, CL-REQ-016
- **dependsOn:** [CL-TASK-001]
- **External dependencies:** `IC-01`, `IC-10`, `IC-11`; `sdd-workspace-transactions/WS-TASK-001` ordered presence/absence/raw observations and generation coherence, `WS-TASK-005` containment/file identity; `sdd-artifact-validation/AV-TASK-004` canonical framing reuse; policy-owner decision for ignored/generated/submodule rules and bookkeeping exclusions.
- **Owned files/modules:** `internal/completion/candidate.go`, `readset.go`, `candidate_digest.go`, candidate property/fuzz tests and `testdata/completion/candidate/`.
- **Proposed verification:** `go test ./internal/completion -run 'TestCandidate|TestReadSet|TestNoGit' -count=2` and `go test ./internal/completion -run=FuzzCandidateReadSet`; fixtures `CL-FIX-004`–`CL-FIX-006`, `CL-FIX-010`, `CL-FIX-021` and candidate digest vectors with shuffled map insertion.
- **Done when:** tracked/untracked/ignored/generated/configuration/dependency/submodule/external inputs, membership sets and required absences participate deterministically; policy-owned exclusions cannot hide a declared input; no-Git manifests fabricate no commit; administrative IDs/revisions/timestamps do not change the content digest while decision-bearing provenance and content do.

### CL-TASK-003 — Accept only discriminated, candidate-bound evidence

- [ ] CL-TASK-003 — Accept only discriminated, candidate-bound evidence
- **Requirements:** CL-REQ-004, CL-REQ-006, CL-REQ-017
- **dependsOn:** [CL-TASK-002]
- **External dependencies:** `IC-09`, `IC-10`, `IC-11`; `sdd-artifact-validation/AV-TASK-007` normalized descriptor/policy facts, `AV-TASK-008` staged validation results, `AV-TASK-009` provider invocation/receipt interpretation and `AV-TASK-010` freshness eligibility; `DEC-01`, `DEC-03`, `DEC-04`; future bounded-execution provider for production evidence.
- **Owned files/modules:** `internal/completion/evidence.go`, `evidence_decode.go`, `freshness.go`, `replay.go`, union/schema consumer fixtures and `testdata/completion/evidence/`.
- **Proposed verification:** `go test ./internal/completion -run 'TestEvidence|TestVerificationState|TestEvidenceReplay'`; fixtures `CL-FIX-001`, `CL-FIX-008`, `CL-FIX-009`, `CL-FIX-020` and `CL-FIX-025`, with inactive-union payload, fake exit-code-for-manual-work and truncated-provider negative cases.
- **Done when:** executed/manual/proxy/review evidence are structurally exclusive; wrong subject/environment/input/contract, missing provider, timeout/cancellation/truncation, expired trust and unavailable references cannot pass; permitted manual evidence retains actor, method, steps, environment, attachments and limitations without fabricated process fields.

### CL-TASK-004 — Evaluate regression, baseline, flaky and bounded-alternative proof

- [ ] CL-TASK-004 — Evaluate regression, baseline, flaky and bounded-alternative proof
- **Requirements:** CL-REQ-005
- **dependsOn:** [CL-TASK-003]
- **External dependencies:** `IC-07`, `IC-09`, `IC-11`; `sdd-artifact-validation/AV-TASK-005` guided profile envelopes and `AV-TASK-009` provider receipts; `sdd-guided-workflows/GW-006` regression-contract consumer definition and `GW-008` journey verification consumer; `DEC-04` production provider decision.
- **Owned files/modules:** `internal/completion/regression.go`, `baseline.go`, regression gate tests and `testdata/completion/regression/`.
- **Proposed verification:** `go test ./internal/completion -run 'TestRegression|TestKnownBaseline|TestFlaky|TestProxy'`; fixtures `CL-FIX-012` and `CL-FIX-025`, plus distinct-before/after, known-failure fingerprint and not-reproduced negative cases.
- **Done when:** failing-before and passing-after bind the correct distinct subjects; accepted baseline failures remain failures; every flaky attempt and policy threshold is visible; intermittent/environment-specific/not-reproduced/unsafe states are preserved; an alternative satisfies only a profile that explicitly permits its class and limitations.

### CL-TASK-005 — Enforce reviewer independence and scoped waiver authority

- [ ] CL-TASK-005 — Enforce reviewer independence and scoped waiver authority
- **Requirements:** CL-REQ-007
- **dependsOn:** [CL-TASK-003]
- **External dependencies:** `IC-10`, `IC-11`; `sdd-artifact-validation/AV-TASK-007` normalized policy/authority facts; `DEC-01` trusted actors/independence and `DEC-06` emergency-waiver policy; guided `GW-009` supplies journey-specific emergency facts but not completion authority.
- **Owned files/modules:** `internal/completion/approval.go`, `waiver.go`, `independence.go`, authority tests and `testdata/completion/authority/`.
- **Proposed verification:** `go test ./internal/completion -run 'TestApproval|TestIndependence|TestWaiver'`; fixtures `CL-FIX-002`, `CL-FIX-007` and `CL-FIX-013`, including role-string, Boolean approval, self-review, wrong scope/candidate and revoked issuer cases.
- **Done when:** approvals bind reviewer/issuer, work/candidate, artifact/policy/definition and gate scope; independence is policy-evaluated; waivers remain `WAIVED`, name eligible diagnostics, expiry and retrospective gates; integrity and authentic authority cannot be waived.

### CL-TASK-006 — Complete deterministic gate evaluation and profile dispositions

- [ ] CL-TASK-006 — Complete deterministic gate evaluation and profile dispositions
- **Requirements:** CL-REQ-008, CL-REQ-009, CL-REQ-015
- **dependsOn:** [CL-TASK-003, CL-TASK-004, CL-TASK-005]
- **External dependencies:** `IC-04`–`IC-11`; `sdd-workflow-graph/WG-TASK-007` deterministic graph-complete handoff and `WG-TASK-009` pin/freshness facts; `sdd-artifact-validation/AV-TASK-006` traceability/delta facts, `AV-TASK-008` validation results and `AV-TASK-011` convergence findings; `sdd-guided-workflows/GW-006`, `GW-008` and `GW-015` journey-specific gate inputs.
- **Owned files/modules:** `internal/completion/gate.go`, `profile.go`, `coverage.go`, `changed_paths.go`, `necessity.go`, ordering/property tests and `testdata/completion/gates/`.
- **Proposed verification:** `go test ./internal/completion -run 'TestGate|TestProfileDisposition|TestScopePriority|TestPartialScope' -count=2`; fixtures `CL-FIX-001`–`CL-FIX-003`, `CL-FIX-011`, `CL-FIX-012` and `CL-FIX-024`, including randomized input ordering.
- **Done when:** all safe independent gates evaluate in `(orderKey, gateId)` order; unknown/missing/stale facts never pass; documentation/research/manual/generated/feature-flag profiles use explicit applicability; changed paths, traceability, guidelines, convergence, migration, review, promotion, necessity and requested verification are covered; an incomplete Must or missing baseline-linked deferred/excluded disposition blocks acceptance, and a feature flag cannot stand in for unimplemented Must scope.

### CL-TASK-007 — Render DONE and the immutable receipt from one accepted fact set

- [ ] CL-TASK-007 — Render DONE and the immutable receipt from one accepted fact set
- **Requirements:** CL-REQ-009, CL-REQ-010, CL-REQ-011, CL-REQ-017
- **dependsOn:** [CL-TASK-006]
- **External dependencies:** `IC-07`, `IC-10`, `IC-11`, `IC-12`; `AM-04` artifact digest ownership; `sdd-artifact-validation/AV-TASK-004` canonical codec framing and `AV-TASK-011` reference/convergence facts; preallocated ID/clock source from foundation composition; `DEC-03` retention policy.
- **Owned files/modules:** `internal/completion/facts.go`, `done.go`, `receipt.go`, `reference.go`, `receipt_digest.go`, renderer golden/property tests and `testdata/completion/records/`.
- **Proposed verification:** `go test ./internal/completion -run 'TestDone|TestReceipt|TestScopePrioritySummary|TestReference|TestDigestGraph' -count=2`; fixtures `CL-FIX-006`, `CL-FIX-009`, `CL-FIX-011`, `CL-FIX-022`, `CL-FIX-024` and `CL-FIX-025`; assert the DONE gate rows, scope-priority summary and receipt gate results have one source and no self/plan-digest cycle.
- **Done when:** preallocated receipt/attempt/plan/transaction IDs make the digest graph acyclic; DONE contains receipt ID but not receipt digest; receipt contains DONE digest and preallocated plan/transaction IDs but not final plan/receipt digest; the applicable scope baseline and category/disposition summary are rendered from accepted facts; code/guideline references bind symbols/content; secrets are absent; historical bytes are immutable.

### CL-TASK-008 — Integrate exact plan application, replay and recovery-required outcomes

- [ ] CL-TASK-008 — Integrate exact plan application, replay and recovery-required outcomes
- **Requirements:** CL-REQ-012, CL-REQ-015
- **dependsOn:** [CL-TASK-007]
- **External dependencies:** `IC-02`, `IC-03`, `IC-10`, `IC-12`; `AM-03`; `sdd-workspace-transactions/WS-TASK-002` exact commit/replay walking slice, `WS-TASK-003` revision/precondition safety, `WS-TASK-004` prepared recovery, `WS-TASK-006` idempotency/disclosure and `WS-TASK-011` extension effects; `sdd-engine-foundation/FND-014` transaction integration record.
- **Owned files/modules:** `internal/completion/proposal.go`, completion-to-workspace plan assembly, the completion apply-result interpreter consumed by foundation-owned engine orchestration, consumer contract tests and `testdata/completion/apply/`. Workspace retains all staging/journal/recovery implementation and foundation retains `Engine.Execute` dispatch.
- **Proposed verification:** `go test ./internal/completion ./internal/engine -run 'TestFinishApply|TestFinishReplay|TestFinishRecovery'`; fixtures `CL-FIX-007`, `CL-FIX-017`–`CL-FIX-020` and `CL-FIX-026`, driven through the real workspace conformance adapter when available and explicitly fake-labelled otherwise.
- **Done when:** completion hands the exact owner-defined plan/content/preconditions to workspace; committed/replayed alone report completed; pre-PREPARED drift conflicts without re-plan; PREPARED/uncertain/quarantined states report recovery-required and preserve transaction identity; concurrent/same-key retries never duplicate receipt/archive effects.

### CL-TASK-009 — Decide pure three-way promotion without a second parser or merge engine

- [ ] CL-TASK-009 — Decide pure three-way promotion without a second parser or merge engine
- **Requirements:** CL-REQ-013
- **dependsOn:** [CL-TASK-006, CL-TASK-007]
- **External dependencies:** `IC-07`, `IC-08`, `IC-12`; `CL-AM-04`; `sdd-artifact-validation/AV-TASK-006` pinned-base `DeltaResult` plus the proposed exact conflict-free target value/bytes; `sdd-workspace-transactions/WS-TASK-003` target preconditions and `WS-TASK-011` approved move/metadata effects.
- **Owned files/modules:** `internal/completion/promotion.go`, promotion decision tests and `testdata/completion/promotion/`; no Markdown parser, canonicalizer or generic merge code.
- **Proposed verification:** `go test ./internal/completion -run 'TestPromotion|TestDisjointPromotion|TestPromotionConflict'`; fixtures `CL-FIX-014`, `CL-FIX-015` and the base/current part of `CL-FIX-026`.
- **Done when:** unchanged/disjoint facts are accepted only with preserved unrelated content and an artifact-owner-produced schema-valid target; same-entity, delete/modify, rename, ID, base and invalid-reference conflicts block; reviewer resolution requires a fresh proposal/base/current observation/approval/plan; no ours/theirs rule exists.

### CL-TASK-010 — Deliver archive, reopen, supersede and abandonment lifecycle slices

- [ ] CL-TASK-010 — Deliver archive, reopen, supersede and abandonment lifecycle slices
- **Requirements:** CL-REQ-001, CL-REQ-014, CL-REQ-015, CL-REQ-016
- **dependsOn:** [CL-TASK-007, CL-TASK-008, CL-TASK-009]
- **External dependencies:** `IC-01`–`IC-03`, `IC-10`, `IC-12`; `sdd-workspace-transactions/WS-TASK-003` tombstones/revisions, `WS-TASK-004` recovery, `WS-TASK-011` move/metadata extension; `DEC-03` retention remains separate from archive.
- **Owned files/modules:** `internal/completion/lifecycle.go`, `archive.go`, `reopen.go`, `supersede.go`, operation engine slices and `testdata/completion/lifecycle/`.
- **Proposed verification:** `go test ./internal/completion ./internal/engine -run 'TestLifecycle|TestArchive|TestReopen|TestSupersede'`; fixtures `CL-FIX-016`, `CL-FIX-021`, `CL-FIX-023` and `CL-FIX-024`.
- **Done when:** illegal/direct transitions reject; failed attempts do not terminalize work; archive collisions preserve existing bytes and archive deletes nothing; reopen creates a new attempt/candidate while retaining the historical receipt; supersession links a later work identity; abandoned/completed/archived/superseded remain distinct.

### CL-TASK-011 — Derive post-verification staleness and durable-reference findings

- [ ] CL-TASK-011 — Derive post-verification staleness and durable-reference findings
- **Requirements:** CL-REQ-003, CL-REQ-010, CL-REQ-016
- **dependsOn:** [CL-TASK-002, CL-TASK-003, CL-TASK-006, CL-TASK-010]
- **External dependencies:** `IC-01`, `IC-04`, `IC-07`, `IC-09`–`IC-11`; `sdd-artifact-validation/AV-TASK-010` cache/freshness and `AV-TASK-011` read-only convergence/reference facts; `sdd-workflow-graph/WG-TASK-009` edge-scoped freshness; workspace coherent current observation from `WS-TASK-001`.
- **Owned files/modules:** `internal/completion/staleness.go`, `reference_freshness.go`, read-only status integration tests and `testdata/completion/staleness/`.
- **Proposed verification:** `go test ./internal/completion ./internal/engine -run 'TestStaleness|TestReferenceFreshness|TestStatusNoMutation'`; fixtures `CL-FIX-004`, `CL-FIX-006`, `CL-FIX-021`–`CL-FIX-023` and `CL-FIX-026`.
- **Done when:** every relevant code/config/generated/dependency/artifact/policy/definition/baseline/guideline/evidence change stales the correct closure; excluded bookkeeping alone does not; line-only references cannot survive changed identity; status derives current staleness without writing a transition; pending exact plans invalidate on bound drift.

### CL-TASK-012 — Enforce evidence minimization, redaction and retention availability

- [ ] CL-TASK-012 — Enforce evidence minimization, redaction and retention availability
- **Requirements:** CL-REQ-006, CL-REQ-017
- **dependsOn:** [CL-TASK-003, CL-TASK-007, CL-TASK-010]
- **External dependencies:** `IC-09`–`IC-12`; `sdd-artifact-validation/AV-TASK-009` pre-persistence provider redaction and `AV-TASK-010` availability/freshness; `sdd-workspace-transactions/WS-TASK-008` journal redaction/reference retention; `DEC-03` company retention/privacy decision.
- **Owned files/modules:** `internal/completion/privacy.go`, `retention.go`, redacted evidence-reference rendering tests and `testdata/completion/privacy/`. No deletion/compaction adapter is owned.
- **Proposed verification:** `go test ./internal/completion -run 'TestRedaction|TestRetention|TestReferenceAvailability'`; fixtures `CL-FIX-009`, `CL-FIX-020` and `CL-FIX-025`, plus seeded token/environment/secret corpora asserted absent from DONE, receipt, diagnostics and debug-safe values.
- **Done when:** only policy-permitted excerpts or digests persist; result interpretability survives valid redaction; retention class and available/redacted/expired/missing state are explicit; missing required references block deterministically; archive never triggers evidence deletion.

### CL-TASK-013 — Publish completion outcomes only through the shared protocol extension

- [ ] CL-TASK-013 — Publish completion outcomes only through the shared protocol extension
- **Requirements:** CL-REQ-008, CL-REQ-009, CL-REQ-011, CL-REQ-012, CL-REQ-014, CL-REQ-017
- **dependsOn:** [CL-TASK-001, CL-TASK-007, CL-TASK-008, CL-TASK-009, CL-TASK-010, CL-TASK-011, CL-TASK-012]
- **External dependencies:** `AM-01`, `AM-02`, `AM-03`, `DEC-05`; `sdd-engine-foundation/FND-001`, `FND-002`, `FND-003`, `FND-012`, `FND-016`; `sdd-artifact-validation/AV-TASK-012`; `sdd-workflow-graph/WG-TASK-012`; guided `GW-018`. Foundation owns the actual discriminator/version/schema allocation.
- **Owned files/modules:** completion nested-result encoder/decoder, diagnostic catalog contribution, `internal/presentation` consumer render tests and provisional positive/negative completion payload fixtures in a completion test area; final `protocol/schema` ownership remains with foundation after ratification.
- **Proposed verification:** run the repository's foundation-owned offline draft-07 schema suite after allocation, plus `go test ./internal/engine ./internal/presentation/... -run 'TestCompletionOutcome|TestCompletionStreams|TestCompletionDiagnostics'`; exercise blocked, proposal-ready, committed, replayed, stale/conflict and recovery-required variants in human/JSON modes.
- **Done when:** one centrally allocated extension preserves closed foundation `1.0` fixtures, rejects unknown/invalid variants, reuses source units/severity/exit taxonomy, emits exactly one JSON document on stdout, keeps diagnostics/progress on stderr and gives human/JSON the same blocker/subject/next-action facts; no document claims a `1.1` allocation without the foundation decision.

### CL-TASK-014 — Run all completion fixtures through package, engine and owner contract suites

- [ ] CL-TASK-014 — Run all completion fixtures through package, engine and owner contract suites
- **Requirements:** CL-REQ-001, CL-REQ-002, CL-REQ-003, CL-REQ-004, CL-REQ-005, CL-REQ-006, CL-REQ-007, CL-REQ-008, CL-REQ-009, CL-REQ-010, CL-REQ-011, CL-REQ-012, CL-REQ-013, CL-REQ-014, CL-REQ-015, CL-REQ-016, CL-REQ-017
- **dependsOn:** [CL-TASK-008, CL-TASK-009, CL-TASK-010, CL-TASK-011, CL-TASK-012, CL-TASK-013]
- **External dependencies:** `IC-01`–`IC-12`; `sdd-engine-foundation/FND-015` conformance/build harness; `sdd-workspace-transactions/WS-TASK-013` fault/state-model suite; `sdd-artifact-validation/AV-TASK-013` parser/provider conformance; `sdd-workflow-graph/WG-TASK-014` graph corpus; guided `GW-017`; real adapters and approved portions of `DEC-01`–`DEC-06`. Missing real adapters remain blockers, not skipped passes.
- **Owned files/modules:** `testdata/completion/CL-FIX-001..026/`, completion package law/property/fuzz tests, Engine black-box suite, owner-consumer contract manifest and CI test manifest.
- **Proposed verification:** `go test ./internal/completion ./internal/engine ./internal/presentation/... -count=2`, targeted fuzz/property jobs, workspace fault-injection contract tests and native owner adapter jobs; run every `CL-FIX-001`–`CL-FIX-026` expected/forbidden outcome.
- **Done when:** every fixture has an executable expected structured result and forbidden-side-effect assertion; repeated runs are byte-stable; no fake satisfies production authority/execution/durability; cancellations and faults preserve owner semantics; all 17 requirements are traceable to passing evidence or an explicit unresolved external blocker.

### CL-TASK-015 — Reconcile consumers and close the provisional completion audit

- [ ] CL-TASK-015 — Reconcile consumers and close the provisional completion audit
- **Requirements:** CL-REQ-001, CL-REQ-002, CL-REQ-003, CL-REQ-004, CL-REQ-005, CL-REQ-006, CL-REQ-007, CL-REQ-008, CL-REQ-009, CL-REQ-010, CL-REQ-011, CL-REQ-012, CL-REQ-013, CL-REQ-014, CL-REQ-015, CL-REQ-016, CL-REQ-017
- **dependsOn:** [CL-TASK-014]
- **External dependencies:** accepted dispositions for `AM-01`–`AM-03`, `AM-07`, `AM-09`, `AM-10`, `DEC-01`–`DEC-06`, `CL-AM-01`–`CL-AM-06`; `sdd-workspace-transactions/WS-TASK-014`, `sdd-artifact-validation/AV-TASK-014`, `sdd-workflow-graph/WG-TASK-014`, foundation `FND-016`, guided `GW-019`. Consumer files are changed only in owner-approved follow-ups.
- **Owned files/modules:** completion conformance/traceability report, interface-ledger evidence, schema compatibility report and exact consumer-amendment manifest; no silent edits to foundation, workspace, artifact, workflow, guided or bounded-execution implementations.
- **Proposed verification:** mechanically check requirement/task/fixture coverage, schema/example references, diagnostic registration, local DAG, module imports and owner boundaries; rerun CL conformance and each available owner-consumer contract suite, recording commands and actual outcomes only.
- **Done when:** reviewers can trace every requirement and fixture to concrete evidence or a named unresolved blocker; IC-01–IC-12 ownership has no duplicate implementation; all required amendments have explicit owner dispositions; no provisional design/task checkbox is used as evidence of acceptance or implementation.

## Requirement coverage matrix

| Requirement | Primary tasks |
|---|---|
| CL-REQ-001 | CL-TASK-001, CL-TASK-010, CL-TASK-014 |
| CL-REQ-002 | CL-TASK-002, CL-TASK-014 |
| CL-REQ-003 | CL-TASK-002, CL-TASK-011, CL-TASK-014 |
| CL-REQ-004 | CL-TASK-003, CL-TASK-014 |
| CL-REQ-005 | CL-TASK-004, CL-TASK-014 |
| CL-REQ-006 | CL-TASK-003, CL-TASK-012, CL-TASK-014 |
| CL-REQ-007 | CL-TASK-005, CL-TASK-014 |
| CL-REQ-008 | CL-TASK-001, CL-TASK-006, CL-TASK-013, CL-TASK-014 |
| CL-REQ-009 | CL-TASK-006, CL-TASK-007, CL-TASK-013, CL-TASK-014 |
| CL-REQ-010 | CL-TASK-007, CL-TASK-011, CL-TASK-014 |
| CL-REQ-011 | CL-TASK-007, CL-TASK-013, CL-TASK-014 |
| CL-REQ-012 | CL-TASK-001, CL-TASK-008, CL-TASK-013, CL-TASK-014 |
| CL-REQ-013 | CL-TASK-009, CL-TASK-014 |
| CL-REQ-014 | CL-TASK-010, CL-TASK-013, CL-TASK-014 |
| CL-REQ-015 | CL-TASK-006, CL-TASK-008, CL-TASK-010, CL-TASK-014 |
| CL-REQ-016 | CL-TASK-002, CL-TASK-010, CL-TASK-011, CL-TASK-014 |
| CL-REQ-017 | CL-TASK-003, CL-TASK-007, CL-TASK-012, CL-TASK-013, CL-TASK-014 |

## Fixture coverage matrix

| Fixtures | Primary tasks |
|---|---|
| CL-FIX-001, CL-FIX-002, CL-FIX-003 | CL-TASK-001, CL-TASK-003, CL-TASK-005, CL-TASK-006 |
| CL-FIX-004, CL-FIX-005, CL-FIX-006 | CL-TASK-002, CL-TASK-007, CL-TASK-011 |
| CL-FIX-007, CL-FIX-008, CL-FIX-009 | CL-TASK-003, CL-TASK-005, CL-TASK-008, CL-TASK-012 |
| CL-FIX-010, CL-FIX-011, CL-FIX-012, CL-FIX-013 | CL-TASK-002, CL-TASK-004, CL-TASK-005, CL-TASK-006 |
| CL-FIX-014, CL-FIX-015 | CL-TASK-009 |
| CL-FIX-016 | CL-TASK-010 |
| CL-FIX-017, CL-FIX-018, CL-FIX-019, CL-FIX-020 | CL-TASK-003, CL-TASK-008, CL-TASK-012 |
| CL-FIX-021, CL-FIX-022, CL-FIX-023 | CL-TASK-002, CL-TASK-010, CL-TASK-011 |
| CL-FIX-024, CL-FIX-025 | CL-TASK-003, CL-TASK-004, CL-TASK-006, CL-TASK-007, CL-TASK-010 |
| CL-FIX-026 | CL-TASK-008, CL-TASK-009, CL-TASK-011 |

CL-TASK-014 reruns every fixture through the integrated conformance surface;
the matrix above names the first owned slice that makes each behavior concrete.

## Local dependency DAG and derived scheduling

The `dependsOn` fields are the only local scheduling source. A scheduler must
reject unknown IDs, duplicates, self-dependencies and cycles and derive waves
rather than trusting a hand-maintained wave list. External dependencies do not
become hidden local edges; when unavailable, their production claim remains
blocked while pure/fake-labelled work may proceed where explicitly stated.

## Required consumer amendments

These are exact follow-up proposals; this task file does not edit their owners.

| Consumer task | Required amendment after owner approval |
|---|---|
| `sdd-guided-workflows/GW-008` | Consume AV provider results plus CL discriminated evidence/candidate acceptance; retain only Bug Fix journey-specific verification assertions. |
| `sdd-guided-workflows/GW-009` | Emit scoped emergency/retrospective facts for CL waiver/gate consumption; do not grant completion authority. |
| `sdd-guided-workflows/GW-015` | Hand accepted Quick Spec lifecycle facts to CL; remove any direct finish transition. |
| `sdd-guided-workflows/GW-016` | Replace guided-owned validation/output writing with CL gate/proposal consumption and workspace apply; keep specialized profile tests. |
| `sdd-guided-workflows/GW-018` | Remove independent protocol `1.1` allocation and contribute guided/completion payloads to AM-02's foundation-owned extension inventory. |
| `sdd-engine-foundation/FND-001`, `FND-002`, `FND-014`, `FND-016` | Allocate the completion result/catalog/effect seam centrally, preserve `1.0`, and consume owner-defined snapshot/plan/recovery values without a second implementation. |
| `sdd-artifact-validation/AV-TASK-006` | Publish the IC-08 exact conflict-free target value/bytes and digest in addition to pure conflict facts. |
| `sdd-artifact-validation/AV-TASK-009`, `AV-TASK-010` | Bind provider/cache facts to ratified `CandidateSubject` and CL evidence acceptance without making AV the gate owner. |
| `sdd-workspace-transactions/WS-TASK-002`, `WS-TASK-004`, `WS-TASK-011` | Accept the exact CL proposal/effects, preserve PREPARED recovery semantics and expose commit-summary linkage for plan/receipt digests; never import completion policy. |

## Final acceptance checklist

- [ ] The project owner has reviewed the requirements and design; no draft is described as accepted merely because it exists.
- [ ] Every CL requirement and fixture maps to executable evidence or an explicit external blocker.
- [ ] Candidate digest vectors prove membership/absence and policy-owned exclusions without administrative self-invalidation.
- [ ] Evidence unions and verification states cannot encode manual/proxy work as fake execution or non-pass as pass.
- [ ] DONE and receipt use one accepted fact set and the digest/link graph is acyclic.
- [ ] Only workspace commit/replay yields completed; PREPARED/uncertain/quarantined yields recovery-required.
- [ ] Promotion consumes IC-08 facts and contains no second parser, ours/theirs rule or silent overwrite.
- [ ] Archive, reopen, supersede, retention and deletion remain separate operations.
- [ ] Human/JSON outputs share structured facts, stream placement and foundation exit categories.
- [ ] Protocol publication records a foundation-owned allocation; no task invents `1.1`.
- [ ] Fake adapters are labelled and cannot satisfy production durability, execution or authority gates.
- [ ] The local `CL-TASK-*` DAG is acyclic and contains no undefined dependency.
- [ ] All implementation checkboxes remain unchecked until implementation and verification actually exist.

## Unresolved blockers

- `DEC-01`–`DEC-06` and `AM-01`–`AM-03`/`AM-07`/`AM-09`/`AM-10` remain provisional.
- The final IC-01/02/03 workspace schemas and IC-08 conflict-free target output are not accepted.
- The production execution provider and reviewer/policy authenticity sources are not selected.
- The native OS/filesystem support matrix and real adapter evidence are not selected.
- Retention/deletion periods and reference-availability policy are not selected.
- Until those decisions land, no completion schema freeze, production evidence, durable finish or organizational approval claim is truthful.

## For the chosen ones

The safest implementation order deliberately builds a blocked finish before a
successful one. A blocked result exercises the deepest semantic seam—candidate,
gate and next-action reasoning—without borrowing durability or authority from
fakes. Only after the accepted fact set and acyclic bytes are stable should the
workspace plan be connected. That ordering makes a fake transaction adapter a
temporary test tool instead of an accidental source of product truth.
