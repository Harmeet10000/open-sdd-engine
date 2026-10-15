# Implementation Tasks: sdd-artifact-validation

## Provisional scope and dependency rules

These are unchecked implementation tasks derived from the provisional design.
They do not approve the requirements/design, claim implementation, or authorize
edits outside the files named by each future task. `dependsOn` contains only
local AV task IDs. Cross-spec requirements, tasks and IC ledger entries are
listed separately and are not smuggled into the local DAG.

Production execution evidence remains blocked on a real execution-provider
contract. Candidate binding remains blocked on ratification of
`sdd-completion-lifecycle`'s `CandidateSubject`. Deterministic fakes may prove
request/result shape but cannot close either production gate.

## Tasks

### AV-TASK-001 — Deliver the read-only artifact check walking slice

- [ ] AV-TASK-001 — Deliver the read-only artifact check walking slice
- **Requirements:** AV-REQ-001, AV-REQ-003, AV-REQ-010, AV-REQ-014, AV-REQ-019, AV-REQ-021
- **dependsOn:** []
- **External dependencies:** foundation-owned AM-02 allocation of `CheckCommand`/`CheckResult` and registry entry; `sdd-engine-foundation` FND-001, FND-003, FND-007 and FND-012; `sdd-workspace-transactions` WS-REQ-002; IC-01, IC-07.
- **Owned files/modules:** `internal/artifact/parse.go`, `internal/artifact/document.go`, `internal/validation/run.go`, the AV handler registered into the foundation extension seam, `protocol/schema/artifact-normalized.schema.json`, focused package/engine tests and `testdata/artifact-validation/walking/`. Foundation owns `internal/engine` public command dispatch, CLI command registration and `CheckResult` envelope schema.
- **Observable outcome:** after the foundation extension is allocated, the AV handler for `Engine.Execute(CheckCommand)` reads one bounded snapshot, parses one ordinary typed artifact, runs syntax/schema stages, returns AV payload facts through the foundation outcome in human/JSON modes, supplies one safe next action on failure, and changes no authoritative bytes/revisions.
- **Verification:** proposed `go test ./internal/artifact ./internal/validation ./internal/engine ./internal/presentation/... -run 'ArtifactCheckWalking|CheckReadOnly|CheckHumanJSON'` and offline schema validation of valid/invalid walking fixtures.
- **Fault/rollback checks:** inject snapshot read failure and cancellation before normalization; assert no provider call, plan, allocation, journal, cache write or repair. Before AM-02 allocation, this slice remains blocked and the closed foundation `1.0` schema remains unchanged.
- **Done when:** the valid fixture yields a typed computed projection, malformed input yields a stable structured diagnostic/exit, human and JSON facts agree, and before/after workspace manifests are identical.

### AV-TASK-002 — Harden YAML and Markdown parsing before alias expansion

- [ ] AV-TASK-002 — Harden YAML and Markdown parsing before alias expansion
- **Requirements:** AV-REQ-001, AV-REQ-002, AV-REQ-003
- **dependsOn:** [AV-TASK-001]
- **External dependencies:** `sdd-engine-foundation` FND-002; IC-07; parser dependency admission remains subject to `.kiro/steering/go-engineering.md`.
- **Owned files/modules:** `internal/artifact/frontmatter.go`, `internal/artifact/markdown.go`, `internal/artifact/budget.go`, parser diagnostics, fuzz tests and `testdata/artifact-validation/parser/`.
- **Observable outcome:** byte/token/depth/anchor/alias/per-alias/expanded-node/expanded-byte budgets are enforced from a token/anchor graph before materialization; UTF-8, duplicate keys, control characters, fences, BOM and LF/CRLF are classified deterministically with exact source ranges.
- **Verification:** proposed `go test ./internal/artifact -run 'Parser|AliasBudget|DuplicateKey|Fence'`, `go test ./internal/artifact -fuzz FuzzStructuredArtifact -fuzztime=30s`, and fixture checks for AV-FIX-001 through AV-FIX-004.
- **Fault/rollback checks:** use saturating-count overflow, cyclic aliases, one-byte-over-limit inputs and cancellation; assert no partial typed AST escapes and allocation remains under the declared hard test budget.
- **Done when:** AV-FIX-001 through AV-FIX-004 have explicit assertions, the exponential alias fixture rejects before expansion, and fenced requirement-like headings never enter the requirement index.

### AV-TASK-003 — Integrate allocation-once artifact identity and scoped revisions

- [ ] AV-TASK-003 — Integrate allocation-once artifact identity and scoped revisions
- **Requirements:** AV-REQ-004, AV-REQ-005, AV-REQ-006
- **dependsOn:** [AV-TASK-001]
- **External dependencies:** `sdd-workspace-transactions` WS-REQ-001, WS-REQ-003 and WS-REQ-004; IC-01 `RawSnapshot`/artifact-identity bridge, IC-02 and IC-07. Durable completion is blocked until the workspace allocation/replay implementation exists.
- **Owned files/modules:** `internal/artifact/identity.go`, identity projection tests, `internal/engine/check_identity_test.go` and `testdata/artifact-validation/identity/`.
- **Observable outcome:** normalization distinguishes proposed new IDs, persisted allocation records, tombstones and missing-record integrity errors; uniqueness uses workspace/work/namespace scope; source scope/owner make integer revisions comparable only within one owner.
- **Verification:** proposed `go test ./internal/artifact ./internal/engine -run 'ArtifactIdentity|ScopedRevision|Tombstone|Move'` with AV-FIX-007 through AV-FIX-008 and workspace conformance handoff fixtures.
- **Fault/rollback checks:** inject collision, missing allocation, failed pre-PREPARED create and replay; assert parsing never writes, failed creation publishes no revision, replay does not reallocate, and move retains ID while workspace advances artifact revision.
- **Done when:** proposed versus allocated identity is visible in typed results and only a workspace exact transaction can establish allocation and the initial/persisted revision.

### AV-TASK-004 — Implement the lossless canonical codec and three digest roles

- [ ] AV-TASK-004 — Implement the lossless canonical codec and three digest roles
- **Requirements:** AV-REQ-008, AV-REQ-009, AV-REQ-010
- **dependsOn:** [AV-TASK-001]
- **External dependencies:** AM-04; `sdd-engine-foundation` FND-006 and FND-016 consume only the agreed framing and must not retain a second generic artifact digest implementation; IC-07.
- **Owned files/modules:** `internal/artifact/canonical.go`, `internal/artifact/digest.go`, codec property tests and `testdata/artifact-validation/canonical/`.
- **Observable outcome:** `AV-C14N-1` preserves null/absent, types, ordering, exact Unicode/code/whitespace where meaningful and retained unknowns; blob, semantic and derivation digests use distinct exact preimages and deterministic ordered closure.
- **Verification:** proposed `go test ./internal/artifact -run 'Canonical|Digest|MeaningRoundTrip|Vector'`; independently recompute published vectors with `sha256sum`; exercise AV-FIX-005 through AV-FIX-006, AV-FIX-010 and AV-FIX-012.
- **Fault/rollback checks:** reject unsupported/lossy values, cyclic derivation closure, ambiguous ordering and integer/length overflow without emitting a digest; assert no input mutation.
- **Done when:** codec round-trip/injectivity properties pass for the typed model, real vectors match, LF/CRLF changes blob but not allowed semantic meaning, and normative/code/order/unknown changes invalidate conservatively.

### AV-TASK-005 — Publish generic trusted profile machinery

- [ ] AV-TASK-005 — Publish generic trusted profile machinery
- **Requirements:** AV-REQ-003, AV-REQ-008, AV-REQ-018
- **dependsOn:** [AV-TASK-002, AV-TASK-004]
- **External dependencies:** AM-02 and AM-06; `sdd-guided-workflows` GW-004–GW-007 and GW-012 own guided content definitions/consumer fixtures; IC-06 and IC-07.
- **Owned files/modules:** `internal/artifact/profile.go`, `protocol/schema/artifact-profile.schema.json`, built-in lightweight/ordinary/protocol profile resources, generic extension decoder, schema fixtures and profile conformance tests. Guided owns its own profile resources and guided fixture content.
- **Observable outcome:** name plus definition digest selects a closed profile; generic typed extensions declare order/text/null/unknown semantics without redefining the common envelope; an enabled `ScopePriority` extension accepts only `must|should|could|wont` plus its required baseline/workaround/exclusion fields; guided-provided definitions pass through the same parser; deterministic protocol rules do not claim external certification.
- **Verification:** proposed `go test ./internal/artifact -run 'Profile|GuidedExtension|ScopePriority|ProtocolRules'` plus offline schema checks and AV-FIX-004, AV-FIX-021 through AV-FIX-022.
- **Fault/rollback checks:** change a definition under the same name, add a new-version-only extension and inject an unknown critical field; assert explicit revalidation/rejection and no fallback to the newer profile.
- **Done when:** each enabled kind resolves one trusted typed schema, old profiles reject newer extensions, malformed or unscoped priority records fail structurally, retained unknowns affect meaning, and guided has no duplicate common-envelope parser or profile-resource ownership overlap.

### AV-TASK-006 — Validate traceability and emit pure pinned-base delta facts

- [ ] AV-TASK-006 — Validate traceability and emit pure pinned-base delta facts
- **Requirements:** AV-REQ-006, AV-REQ-011, AV-REQ-012
- **dependsOn:** [AV-TASK-003, AV-TASK-004, AV-TASK-005]
- **External dependencies:** `sdd-workflow-graph` WG-REQ-001, WG-REQ-002 and WG-REQ-009 supply opaque compiled/routing facts; provisional `sdd-completion-lifecycle` CL-REQ-013 consumes conflict facts; IC-04, IC-05 and IC-08.
- **Owned files/modules:** `internal/artifact/trace.go`, `internal/artifact/delta.go`, trace/delta schemas, pure property tests and `testdata/artifact-validation/trace-delta/`.
- **Observable outcome:** exact scoped entity references yield forward/reverse coverage; add/modify/remove require resolvable pinned identities; three-way comparison returns ordered base-resolution/change/conflict facts and never compiles workflow topology, resolves conflicts or writes.
- **Verification:** proposed `go test ./internal/artifact -run 'Traceability|Delta|PinnedBase|NoWorkflowCompiler'` with AV-FIX-008 through AV-FIX-012 and AV-FIX-024.
- **Fault/rollback checks:** missing/ambiguous/removed targets, changed base, delete-modify, rename/ID collision and cancellation leave snapshot bytes unchanged and return all safely known conflict facts.
- **Done when:** completion can consume a deterministic `DeltaResult` without reaching into parser internals, while authorization/promotion decisions remain absent from the result.

### AV-TASK-007 — Normalize descriptors, policy and typed assessments

- [ ] AV-TASK-007 — Normalize descriptors, policy and typed assessments
- **Requirements:** AV-REQ-007, AV-REQ-013, AV-REQ-015, AV-REQ-017, AV-REQ-018
- **dependsOn:** [AV-TASK-004, AV-TASK-005]
- **External dependencies:** AM-01 and AM-10; `sdd-engine-foundation` FND-001/FND-003 severity contract; `sdd-workflow-graph` WG-REQ-007/WG-REQ-008; IC-06, IC-09 and IC-10; company DEC-01.
- **Owned files/modules:** `internal/validation/descriptor.go`, `policy.go`, `assessment.go`, their protocol schemas, trusted-config fixtures and package tests.
- **Observable outcome:** semantic contract, binary provenance, trust tier, compatibility claim, foundation severity and blocking policy remain distinct; policy precedence and authority observations fail closed; assessments expose unassessed/insufficient/assessed without asserting prose truth or universal signatures.
- **Verification:** proposed `go test ./internal/validation -run 'Descriptor|Policy|Assessment|Authority'` and offline schemas for AV-FIX-013 through AV-FIX-015, AV-FIX-021 through AV-FIX-022 and AV-FIX-025.
- **Fault/rollback checks:** forged name, heuristic-as-proof, missing org policy, expired/revoked issuer, self-declared patch compatibility and critical severity all retain their exact blocking/trust semantics.
- **Done when:** all descriptor fields participate in descriptor identity, behavior fields participate in semantic-contract identity, and no name/version/signature/digest alone establishes current authority.

### AV-TASK-008 — Run the deterministic staged validation pipeline

- [ ] AV-TASK-008 — Run the deterministic staged validation pipeline
- **Requirements:** AV-REQ-010, AV-REQ-013, AV-REQ-014, AV-REQ-015, AV-REQ-021
- **dependsOn:** [AV-TASK-002, AV-TASK-006, AV-TASK-007]
- **External dependencies:** `sdd-engine-foundation` FND-007 and FND-012; `sdd-workflow-graph` WG-REQ-009; IC-05, IC-09 and IC-10.
- **Owned files/modules:** `internal/validation/pipeline.go`, `diagnostic.go`, `order.go`, stage result schemas, engine integration tests and `testdata/artifact-validation/pipeline/`.
- **Observable outcome:** syntax, schema, artifact-graph, traceability, deterministic-policy, execution-backed and heuristic stages produce fixed `pass|fail|error|skipped|unavailable` results, explicit dependencies and byte-stable diagnostic order from explicit time/external observations.
- **Verification:** proposed `go test ./internal/validation ./internal/engine -run 'Pipeline|StageStatus|DeterministicOrder|ExplicitObservation'` with repeated/shuffled-map runs.
- **Fault/rollback checks:** fatal syntax/schema errors prevent unsafe dependent stages while independent safe facts remain; timeout/cancellation/uninterpretable input never becomes fail/pass or downgraded warning.
- **Done when:** identical complete requests yield byte-identical results and every skipped/unavailable stage names a non-success reason and blocking disposition.

### AV-TASK-009 — Integrate exact argv provider invocation and receipt interpretation

- [ ] AV-TASK-009 — Integrate exact argv provider invocation and receipt interpretation
- **Requirements:** AV-REQ-007, AV-REQ-014, AV-REQ-016, AV-REQ-021
- **dependsOn:** [AV-TASK-008]
- **External dependencies:** provisional `sdd-completion-lifecycle` CL-REQ-002, CL-REQ-004 and CL-REQ-006; IC-09 and IC-11; DEC-04; future bounded-execution provider is required for production proof.
- **Owned files/modules:** `internal/validation/provider.go`, `invocation.go`, `receipt.go`, the consumer-owned provider port, deterministic fake contract tests, redaction tests and `testdata/artifact-validation/provider/`.
- **Observable outcome:** invocation binds approved executable plus real argument array, cwd, allowlisted environment, limits, artifact inputs and a distinct candidate subject; provider owns supervision/scratch and reports enforcement/cleanup; receipts preserve timeout/cancellation/signal/truncation/redaction without granting completion authority.
- **Verification:** proposed `go test ./internal/validation -run 'Invocation|Argv|ProviderReceipt|Redaction|Cancellation'` for AV-FIX-016 through AV-FIX-020 and AV-FIX-023; later run the same contract suite against each real native provider adapter.
- **Fault/rollback checks:** embedded Markdown commands, space-bearing argv, denied environment, absent provider, secret stderr, timeout, cancellation, malformed receipt and unexpected authoritative input change. Assert no shell parsing, orphan claim, authoritative write or fake production pass.
- **Done when:** fake tests prove only request/interpretation shape; production acceptance remains explicitly blocked until a real provider proves descendant cleanup, resource enforcement, scratch isolation and receipt provenance.

### AV-TASK-010 — Enforce conservative freshness and cache eligibility

- [ ] AV-TASK-010 — Enforce conservative freshness and cache eligibility
- **Requirements:** AV-REQ-007, AV-REQ-009, AV-REQ-010, AV-REQ-013, AV-REQ-020
- **dependsOn:** [AV-TASK-004, AV-TASK-007, AV-TASK-008, AV-TASK-009]
- **External dependencies:** AM-10; provisional `sdd-completion-lifecycle` CL-REQ-003, CL-REQ-006 and CL-REQ-016; IC-09–IC-11; DEC-03 controls retention but not eligibility.
- **Owned files/modules:** `internal/validation/cache.go`, cache-key codec fixtures and eligibility tests. No persistent cache adapter is owned by this task.
- **Observable outcome:** candidate digest, ordered artifact digests, source revision scope/owner, policy/profile/contract, implementation/environment/trust/expiry observations determine eligibility; changed binary reuses only through an explicit trusted compatibility edge; lookup never upgrades trust.
- **Verification:** proposed `go test ./internal/validation -run 'CacheEligibility|Compatibility|Freshness'` with AV-FIX-020 through AV-FIX-023 and repeated deterministic key generation.
- **Fault/rollback checks:** alter each key input independently, revoke trust, expire evidence, change security/semantic contract and present a patch label without proof; each produces a miss/stale fact while the old immutable result remains unchanged.
- **Done when:** the allow case requires exact equality plus any necessary compatibility proof, all unknowns deny reuse, and no retention/deletion policy is silently chosen.

### AV-TASK-011 — Deliver bounded read-only convergence assessment

- [ ] AV-TASK-011 — Deliver bounded read-only convergence assessment
- **Requirements:** AV-REQ-010, AV-REQ-011, AV-REQ-019, AV-REQ-021
- **dependsOn:** [AV-TASK-006, AV-TASK-008, AV-TASK-010]
- **External dependencies:** `sdd-workspace-transactions` WS-REQ-002/WS-REQ-004; IC-01, IC-02 and IC-10. Applying remediation remains a new workspace plan.
- **Owned files/modules:** `internal/validation/convergence.go`, `budget.go`, remediation value schema, engine tests and `testdata/artifact-validation/convergence/`.
- **Observable outcome:** bounded artifacts/edges/depth/bytes/output traversal returns deterministic redacted findings, exact source locations and inert remediation proposals; cycles/missing facts/budget exhaustion are explicit and the command performs no provider call or authoritative mutation.
- **Verification:** proposed `go test ./internal/validation ./internal/engine -run 'Convergence|AssessmentBudget|ReadOnly|Redaction'` using AV-FIX-026 and before/after filesystem/revision manifests.
- **Fault/rollback checks:** one-over-limit traversal, missing external facts, cycles, secrets and cancellation preserve entries/bytes/modes/revisions and return no applicable mutation plan.
- **Done when:** AV-FIX-026 is deterministic across repeated runs and every proposal requires a separate normal plan/authorize/apply path.

### AV-TASK-012 — Freeze public schemas, diagnostics and Engine outcomes

- [ ] AV-TASK-012 — Freeze public schemas, diagnostics and Engine outcomes
- **Requirements:** AV-REQ-001, AV-REQ-003, AV-REQ-014, AV-REQ-016, AV-REQ-021
- **dependsOn:** [AV-TASK-008, AV-TASK-009, AV-TASK-011]
- **External dependencies:** AM-01 and AM-02 must be ratified; `sdd-engine-foundation` FND-001, FND-002, FND-003, FND-012 and FND-016; guided GW-018; IC-07, IC-09 and IC-10.
- **Owned files/modules:** AV-owned payload schemas under the foundation-allocated extension directory, `protocol/fixtures/{valid,invalid}/artifact-validation/`, validation diagnostic catalog, human/JSON focused render tests and compatibility notes. Foundation owns the registry, outer envelope, command/result discriminators and extension fixture inventory.
- **Observable outcome:** all AV payload records validate offline against the foundation-allocated registry; diagnostics use foundation source objects/range units/four severities/exits; machine payload variants are deterministic; human/JSON retain the same facts without parsing message text.
- **Verification:** proposed repository schema command, `go test ./internal/engine ./internal/presentation/... -run 'ValidationOutcome|DiagnosticParity|Stream'`, and invalid fixtures for unknown versions/variants/nullability/source-unit/severity.
- **Fault/rollback checks:** old foundation `1.0` fixtures remain byte/meaning compatible, unsupported AV version rejects explicitly, partial stdout never receives a second envelope, and protocol allocation can be reverted without changing internal typed semantics.
- **Done when:** a reviewer can resolve every `$ref` offline, every example validates with recomputed digests, AV contributes no public envelope/discriminator, and no competing exit/source/severity vocabulary exists.

### AV-TASK-013 — Complete adversarial, fuzz and real-adapter conformance

- [ ] AV-TASK-013 — Complete adversarial, fuzz and real-adapter conformance
- **Requirements:** AV-REQ-001, AV-REQ-002, AV-REQ-005, AV-REQ-006, AV-REQ-008, AV-REQ-009, AV-REQ-011, AV-REQ-012, AV-REQ-013, AV-REQ-014, AV-REQ-016, AV-REQ-017, AV-REQ-018, AV-REQ-019, AV-REQ-020, AV-REQ-021
- **dependsOn:** [AV-TASK-002, AV-TASK-003, AV-TASK-005, AV-TASK-006, AV-TASK-009, AV-TASK-010, AV-TASK-011, AV-TASK-012]
- **External dependencies:** IC-01, IC-07–IC-11; real workspace and provider adapters for durability/execution claims; FND-015; guided GW-017. Missing adapters leave the corresponding production assertion blocked rather than skipped-as-pass.
- **Owned files/modules:** `testdata/artifact-validation/AV-FIX-001..026/`, package fuzz/property suites, engine black-box suite, provider adapter contract suite and CI test manifest.
- **Observable outcome:** every AV fixture has immutable input, expected structured result/exit, authoritative-state manifest and forbidden-side-effect assertion; parser/delta fuzzers minimize regressions; platform provider checks are labeled real versus fake.
- **Verification:** proposed `go test ./...`, `go test -race ./...`, targeted fuzz seeds for parser/delta/codec, offline schema validation, and native provider contract jobs where supported.
- **Fault/rollback checks:** cancellation and one-before/after boundaries cover parse allocation, provider start/output/cleanup and snapshot drift; a missing native capability blocks release evidence instead of passing through a skip.
- **Done when:** AV-FIX-001 through AV-FIX-026 are each referenced by an executable assertion, failures identify stable codes rather than English, and no test touches the developer workspace/home.

### AV-TASK-014 — Reconcile consumers and run the provisional release audit

- [ ] AV-TASK-014 — Reconcile consumers and run the provisional release audit
- **Requirements:** AV-REQ-001, AV-REQ-002, AV-REQ-003, AV-REQ-004, AV-REQ-005, AV-REQ-006, AV-REQ-007, AV-REQ-008, AV-REQ-009, AV-REQ-010, AV-REQ-011, AV-REQ-012, AV-REQ-013, AV-REQ-014, AV-REQ-015, AV-REQ-016, AV-REQ-017, AV-REQ-018, AV-REQ-019, AV-REQ-020, AV-REQ-021
- **dependsOn:** [AV-TASK-013]
- **External dependencies:** FND-016; guided GW-012, GW-013, GW-017 and GW-019; provisional completion CL-REQ-002/CL-REQ-006/CL-REQ-013; IC-01–IC-12; AM-01, AM-02, AM-04, AM-06, AM-07 and AM-10.
- **Owned files/modules:** AV consumer conformance tests, cross-spec compatibility report and release-gate manifest only. Consumer implementations remain owned by their specs.
- **Observable outcome:** foundation, workspace, graph, guided and completion consumers use one envelope/parser/digest implementation; all produced/consumed fields have owner, version, freshness/trust rule and positive/negative fixture; unresolved external gates remain visibly blocked.
- **Verification:** proposed task-DAG/ID/reference checker, schema/example validation, `go list -deps` import-direction check, full Engine black-box suite and explicit manual review of every IC/amendment row.
- **Fault/rollback checks:** remove each producer/fake authority in turn; consumers report unavailable/incompatible rather than compiling a substitute parser/graph/provider or silently accepting evidence.
- **Done when:** every AV requirement and fixture maps below, local DAG has no unknown/cycle, consumer amendments are filed with their exact owning tasks, and no durable/execution/completion claim depends only on a fake.

## Requirement coverage

| Requirement | Tasks |
|---|---|
| AV-REQ-001 | AV-TASK-001, 002, 012, 014 |
| AV-REQ-002 | AV-TASK-002, 013, 014 |
| AV-REQ-003 | AV-TASK-001, 002, 005, 012, 014 |
| AV-REQ-004 | AV-TASK-003, 014 |
| AV-REQ-005 | AV-TASK-003, 013, 014 |
| AV-REQ-006 | AV-TASK-003, 006, 013, 014 |
| AV-REQ-007 | AV-TASK-007, 009, 010, 014 |
| AV-REQ-008 | AV-TASK-004, 005, 013, 014 |
| AV-REQ-009 | AV-TASK-004, 010, 013, 014 |
| AV-REQ-010 | AV-TASK-001, 004, 008, 010, 011, 014 |
| AV-REQ-011 | AV-TASK-006, 011, 013, 014 |
| AV-REQ-012 | AV-TASK-006, 013, 014 |
| AV-REQ-013 | AV-TASK-007, 008, 010, 013, 014 |
| AV-REQ-014 | AV-TASK-001, 008, 009, 012, 013, 014 |
| AV-REQ-015 | AV-TASK-007, 008, 014 |
| AV-REQ-016 | AV-TASK-009, 012, 013, 014 |
| AV-REQ-017 | AV-TASK-007, 013, 014 |
| AV-REQ-018 | AV-TASK-005, 007, 013, 014 |
| AV-REQ-019 | AV-TASK-001, 011, 013, 014 |
| AV-REQ-020 | AV-TASK-010, 013, 014 |
| AV-REQ-021 | AV-TASK-001, 008, 009, 011, 012, 013, 014 |

## Fixture coverage

| Fixtures | Tasks |
|---|---|
| AV-FIX-001 through AV-FIX-004 | AV-TASK-002, AV-TASK-013 |
| AV-FIX-005 through AV-FIX-006 | AV-TASK-004, AV-TASK-013 |
| AV-FIX-007 through AV-FIX-008 | AV-TASK-003, AV-TASK-013 |
| AV-FIX-009 through AV-FIX-012 | AV-TASK-006 (AV-FIX-010/AV-FIX-012 also exercise digests), AV-TASK-013 |
| AV-FIX-013 through AV-FIX-015 | AV-TASK-007, AV-TASK-013 |
| AV-FIX-016 through AV-FIX-020 | AV-TASK-009, AV-TASK-013 |
| AV-FIX-021 through AV-FIX-023 | AV-TASK-005, AV-TASK-007, AV-TASK-009, AV-TASK-010, AV-TASK-013 |
| AV-FIX-024 | AV-TASK-006, AV-TASK-013 |
| AV-FIX-025 | AV-TASK-007, AV-TASK-013 |
| AV-FIX-026 | AV-TASK-011, AV-TASK-013 |

## Local dependency audit

The local DAG is intentionally acyclic:

```text
001 -> {002,003,004}
{002,004} -> 005
{003,004,005} -> 006
{004,005} -> 007
{002,006,007} -> 008 -> 009
{004,007,008,009} -> 010
{006,008,010} -> 011
{008,009,011} -> 012
{002,003,005,006,009,010,011,012} -> 013 -> 014
```

An automated checker in AV-TASK-014 must derive this graph from `dependsOn`,
reject duplicate/unknown/self IDs and cycles, and ignore cross-spec references
when evaluating the local DAG.

## Exact consumer amendment list

- **Foundation FND-001/FND-003/FND-016:** reconcile AM-01 severity/blocking and
  centrally allocate AV schema/result variants under AM-02.
- **Foundation FND-006:** retain foundation-native canonicalization only;
  replace generic artifact canonicalization/digests with the AV-TASK-004
  implementation and cross-contract vectors (AM-04).
- **Guided GW-004–GW-007:** publish only journey-specific content schemas and
  assertions; consume the AV common envelope (AM-06).
- **Guided GW-012:** replace parser/digest ownership with AV-TASK-001/004/005/006
  consumer conformance and bundle traceability.
- **Guided GW-008/GW-013:** consume AV-TASK-009 receipts and the ratified
  completion candidate; keep journey-specific regression/gate assertions and
  preserve the no-fake-production-proof block (AM-07).
- **Guided GW-001:** consume AV-TASK-007/010's separated semantic/binary
  compatibility instead of patch-label freshness (AM-10).
- **Completion CL-REQ-013 implementation task (not yet published):** consume
  AV-TASK-006 `DeltaResult`; never import parser internals or resolve conflicts
  inside AV.

## Unresolved blockers

- AM-01/AM-02 central protocol decisions are required before AV-TASK-012 can
  freeze a public schema.
- The candidate subject/exclusion digest in completion requirements is
  provisional; AV-TASK-009/010 may implement an adapter at the seam but cannot
  freeze IC-11 or claim accepted evidence yet.
- The workspace allocation/revision adapter and real bounded-execution provider
  have no accepted implementation tasks in this batch. Fake tests are useful
  but do not close durability, isolation, cancellation or evidence gates.
- Parser dependency, issuer/authenticity policy, cache/evidence retention and
  platform support remain explicit decisions; no task chooses them by default.

No checkbox above may be checked by drafting this file.

## For the chosen ones

The vertical split is deliberately not parser/model/validator/storage. Each
slice closes an observable invariant at the engine seam: a bounded parse, an
allocation binding, a digest law, a pure conflict fact, or a candidate-bound
receipt. That shape makes counterfeit authority visible. A horizontal “build
all models” phase can look complete while no caller has yet proved that a
proposed ID stays proposed, a fake stays a fake, or a read-only result stays
read-only.
