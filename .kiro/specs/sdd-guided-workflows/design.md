# Design Document: sdd-guided-workflows

## Overview

Bug Fix and Quick Spec are maintained data-driven journey definitions consumed through `Engine.Execute`. They do not add commands or embed a model. A resolved composition is:

```text
journey topology + policy/profile-derived rigor + artifact/validation profiles + trusted definitions
```

The composition is persisted under the canonical `.sdd/work/<slug>/work.yaml` record with immutable work ID, revision, and exact definition/policy/validator digests. A slug is display data; it is not identity.

## Package seams and dependency direction

`workflow` compiles and routes graph definitions; `artifact` parses, normalizes, and digests typed artifacts; `validation` consumes artifacts and trusted policy to execute validators; `completion` evaluates gates and plans completion outputs; the workspace transaction applies authorized effects. `engine` coordinates immutable snapshots and plans. The direction is:

```text
artifact (parse/normalize/value)
          ^
validation (consume/diagnose) <- workflow (compile/route)
          ^                         ^
          +---------- engine -------+
                         |
                    completion
```

Artifact does not import validation. Validation does not own parsing. Engine does not reopen files during pure passes and no package uses mutable global state. Effects are plain plans at the workspace transaction seam. The named prerequisite specs own durable implementation details; this spec defines their required interfaces only.

## Composition and policy

Preset selection is a convenience. Resolution loads required organization policy from its trusted versioned source, then project narrowing policy, then work metadata. Missing, expired, revoked, stale, or unverifiable required policy fails closed. A local CLI and mutable local policy files are not tamper-resistant against the same user; an approval/exception claim must identify a trusted CI or other declared trust boundary.

Resolution computes:

```text
Composition {
  workID, revision, preset, journey, rigor,
  artifactProfiles, schemaVersion,
  workflowDefinitionDigest, policyDigest, validatorDefinitionDigest,
  routingReasons, evidenceRefs,
  scopePriorityProfileRef?, scopeBaselineRef?
}
```

The definition digest covers exact versioned workflow graph, policy profile, artifact schemas, and validator definitions. It is invalid if only names match. Explicit selection may increase rigor; it cannot lower the policy floor. Risk composition may escalate rigor while retaining Bug Fix or Quick Spec journey unless policy explicitly requires a new topology.

Risk input is a typed external assessment with `unassessed|insufficient|assessed`, evidence references and author/approval provenance. The pure resolver takes that assessment and trusted policy explicitly; it never receives an implicit null that classifies every start as critical. Unassessed work may gather evidence but cannot enter implementation/approval. For assessed input, take the maximum requested/risk/organization rigor, union mandatory artifact/check profiles, and produce deterministic reasons. Quick Spec exceeding its quick-risk ceiling must record an accepted escalation before further implementation; retained authoring topology does not imply retained quick rigor.

MoSCoW is an orthogonal scope-baseline value. It is not rigor, severity,
dependency readiness, approval, or a graph edge. The guided composition stores
the selected priority profile and exact baseline reference/digest, while the
artifact profile validates individual `ScopePriority` records and completion
consumes the accepted baseline. The resolver may report an over-cap or
unverified budget, but it never silently reclassifies a category or promotes a
feature flag into completion evidence.

## Bug Fix graph

The canonical graph is:

```text
report
  -> reproduce/observe
  -> investigate
  -> diagnose
  -> define-regression-contract
  -> conditional-design
  -> tasks
  -> implement
  -> verify
  -> policy-review
  -> finish
```

`conditional-design` is a policy-selected node. The ordinary graph does not allow `diagnose -> implement`, nor `contract -> finish`. A scoped preauthorized emergency route may enter `implement` early for a mitigation, but it records the exception and remains unable to finish until the ordinary evidence gates are reconciled.

### Bug Fix artifacts and gates

| Artifact | Produced by | Required distinction |
|---|---|---|
| `bug-report` | report | symptom, impact, environment, source; no cause assertion |
| `reproduction` | reproduce/observe | closed status, attempts, safe evidence, snapshot |
| `investigation` | investigate | observations, hypotheses, evidence, effects |
| `diagnosis` | diagnose | mechanism, discriminating evidence, scope, uncertainty, spec relationship |
| `regression-contract` | define-regression-contract | observable oracle, failing-before expectation, passing-after expectation, negative/safety cases |
| `design` | conditional-design | accepted boundary/decision when policy requires |
| `tasks` | tasks | implementation/verification/rollback work and traceability |
| `verification` | verify | independent executed checks, result, environment, output digest, candidate snapshot digest |
| `policy-review` record | policy-review | trusted profile gates and decision evidence |

Every artifact has this common envelope: `kind`, `schemaVersion`, immutable `id`, `workId`, `artifactRevision`, `sourceRevision`, `status`, `blobDigest`, `semanticDigest`, `derivationDigest`, `source` (structured path/range/provenance), and typed `content`. Status transitions are explicit (`draft -> submitted -> accepted` or `rejected/stale`); an agent cannot omit provenance or manufacture an accepted status.

The typed content contracts require at minimum:

```text
BugReport.content = {expected, observed, impact, environment, firstKnownVersion, claim}
Reproduction.content = {status, attempts[], stepsOrProxy, environment, snapshotDigest, evidence[]}
Investigation.content = {observations[], hypotheses[], discriminatingEvidence[], effects[]}
Diagnosis.content = {mechanism, affectedScope, unaffectedScope, confidence, uncertainty[], specRelationship, evidence[]}
RegressionContract.content = {requirementRefs[], oracle, failingBefore, passingAfterExpected, negativeCases[], safetyCases[]}
Design.content = {decisionRefs[], boundaries[], alternatives[], acceptedBy, evidence[]}
Tasks.content = {taskIds[], dependsOn, requirementRefs[], designRefs[], implementationCoverage[], verificationCoverage[], rollbackCleanup[]}
Verification.content = {candidateSnapshotDigest, checks[], result, environment, outputDigests[], validatorsExecuted[]}
```

`passingAfterExpected` is an expectation in a contract; only `Verification.content.result` from checks executed after implementation can establish verification. Every evidence item carries a content digest and provenance, and every source range uses the foundation units (`utf8-byte` or `unicode-scalar`).

The regression contract is defined before implementation. It describes what must be observed, but does not claim that a check has run. Verification is a separate artifact produced only after implementation and bound to a candidate snapshot; an agent's `passed` field or a contract's passing-after column cannot satisfy it.

Reproduction statuses are `reproduced`, `intermittent`, `environment-specific`, `not-reproduced`, and `unsafe-to-reproduce`. `not-reproduced` records an unsuccessful observation; `unsafe-to-reproduce` records a safety disposition. Neither implies a pass or diagnosis. A safe proxy, simulation, incident evidence, or exception must be approved and recorded before diagnosis; an unsafe proxy cannot be invented by relabeling a missing test. Risk/materiality assessment has its own explicit `unassessed` status.

Diagnosis is an assessment with evidence/provenance/confidence/uncertainty, not a deterministic root-cause classifier. Structural/policy validators are deterministic. Heuristics can suggest hypotheses but never satisfy a mandatory diagnosis or verification gate.

### Bug Fix transition contract

The graph below is the static DAG. Runtime retries, rejection loops, and emergency reconciliation are events recorded against a node revision; they do not add graph edges or create a second topology.

| Node | Typed inputs | Output/status | Skip/retry branch |
|---|---|---|---|
| `report` | intent, workspace snapshot | `bug-report` / submitted | malformed -> revise report |
| `reproduce/observe` | bug-report, safe probe policy | `reproduction` / observed status | intermittent may retry within policy; not-reproduced/unsafe blocks or requests bounded evidence |
| `investigate` | reproduction, snapshot | investigation hypotheses/effects / open or ready | add evidence or dispose probe; no silent pass |
| `diagnose` | investigation, evidence | diagnosis assessment / supported or unsupported | unsupported -> investigate; emergency exception may defer |
| `define-regression-contract` | diagnosis or scoped mitigation, requirements | contract / defined | invalid -> revise; contract never counts as verification |
| `conditional-design` | contract, rigor/profile | accepted design or explicit skip / skipped when not required | required design rejection -> revise |
| `tasks` | contract, design if present | accepted task DAG / ready | traceability failure -> revise |
| `implement` | accepted tasks, expected revision | candidate snapshot / implemented | stale claim -> replan; emergency path is scoped |
| `verify` | accepted implementation, candidate snapshot, contract | fresh verification artifact / passed or failed | failed -> implement/replan; stale candidate -> rerun |
| `policy-review` | all evidence, profile, trusted policy | review decision / accepted or blocked | optional only when profile says skip; blocked -> remediate |
| `finish` | accepted gates and candidate | completion request / done or blocked | blocked -> named gate; completion owner writes outputs atomically |

## Emergency repair

An emergency exception is a typed authorization:

```text
actorID, grantScope, authorizationEvidence, policySnapshotDigest,
workID, startingRevision, startingSnapshotDigest, inputDigest,
issuedAt, expiresAt, revocationState, preauthorizationChecks
```

It is bound to the exact mitigation plan digest. `owner` text and `approved: true` are not authority. The engine may permit a scoped implementation and immediate verification before complete diagnosis only when policy preauthorizes it. Before finish it must obtain retrospective reproduction/observation, diagnosis, regression contract, accepted-spec convergence, independent verification, and policy review. Deployment state and SDD completion state remain distinct.

## Quick Spec authoring graph and shared lifecycle

The authoring graph is:

```text
intent
  -> discover-material-unknowns
  -> assess-materiality-and-risk
  -> answer-clarifications
  -> prioritize-delivery-scope
  -> generate-spec-bundle
  -> validate-bundle
  -> approval
  -> implementation-ready
```

The final node hands off to the shared implementation lifecycle:

```text
implementation-ready -> implement -> verify -> policy-review -> finish
```

`implementation-ready -> finish` is invalid. A rejected bundle returns to authoring with structured diagnostics and a new candidate revision; a revised bundle must be validated and reapproved. An approval is bound to work ID, revision, bundle content/derivation digests, policy digest, and plan digest. No implicit acceptance occurs.

Clarification discovery uses trusted repository/accepted-spec/configuration/read-only evidence first. A material question changes behavior, compatibility, migration/destructive behavior, trust/privacy/authorization, architecture, task ordering, verification, or rollout/rollback. Zero questions is a valid closed state.

### Scope-priority baseline

`prioritize-delivery-scope` produces a versioned baseline rather than a
backlog label list:

```text
ScopeBaseline {
  scopeId, revision, predecessorDigest?, policyRef,
  capacity?, capacityUnit?, conversionRuleRef?, mustCap?,
  decisions[] ScopePriority {
    subjectRef, category: must | should | could | wont,
    rationale, decisionEvidenceRefs[], workaround?, deferredTo?,
    estimate?, capacityUnit?, decidedBy, decidedAt
  },
  baselineDigest
}
```

Each `subjectRef` occurs exactly once in a baseline; category inheritance and
duplicate estimates are rejected rather than guessed. The default profile
calculates a 60% Must cap only when units are comparable, conversion rules are
pinned, shared work is not double-counted, and usable capacity is declared. It
returns `scope-budget-unverified` when they are absent and
`scope-replan-required` when Must work exceeds the cap. Those are planning
outcomes, not a hidden rewrite: only the policy-authorized scope decision can
change a Must. A viable manual workaround normally requires `should`; a `wont`
requires an exclusion/reconsideration record and produces no implementation
commitment. A revision binds the predecessor digest and expected scope revision;
implemented excluded scope must be surfaced as a scope change, never credited
silently.

Feature flags remain rollout records. Their cohort, fallback, observability,
owner, and removal/review condition may be included in a scope decision, but
they cannot satisfy an unimplemented Must or substitute for validation.

### Quick Spec transition contract

The authoring graph and the shared implementation graph are static DAGs. Retry/rejection/reapproval are runtime events that create candidate revisions; they do not turn `implementation-ready` into a completion edge.

| Node | Typed inputs | Output/status | Skip/retry branch |
|---|---|---|---|
| `intent` | intent text, workspace snapshot | work identity / open | invalid input -> usage diagnostic |
| `discover-material-unknowns` | resolved snapshot, accepted specs | evidence set / discovered | no unknown -> continue |
| `assess-materiality-and-risk` | unresolved questions, evidence | assessment with provenance/uncertainty / assessed or unassessed | unassessed -> bounded external assessment action |
| `answer-clarifications` | material questions, evidence | answers / closed or blocked | revise answers; zero questions is valid |
| `prioritize-delivery-scope` | accepted answers, trusted policy, estimates/capacity where available | pinned scope baseline / ready, unverified, or scope-replan-required | revise scope decision; no automatic Must downgrade |
| `generate-spec-bundle` | answers, immutable context packet | requirements/design/tasks candidate / generated | budget overflow -> digest-only references and remediation |
| `validate-bundle` | candidate bundle, candidate snapshot, trusted policy | validation result / accepted-for-review or rejected | rejected/stale -> candidate revision and revalidate |
| `approval` | validated bundle, plan/digests, approval claim | accepted bundle / approved or denied | denial/revision -> authoring; same key changed -> conflict |
| `implementation-ready` | accepted bundle, current revision | lifecycle handoff / ready | changed definitions/revision -> stale and revalidate |
| `implement` | accepted tasks, expected revision | candidate snapshot / implemented | stale -> replan |
| `verify` | implementation, candidate snapshot, requirements | verification artifact / passed or failed | failed/dirty -> rerun or replan |
| `policy-review` | verification and profile gates | review / accepted or blocked | profile may skip only if policy says so |
| `finish` | accepted lifecycle gates | completion request / done or blocked | completion owner validates then atomically emits outputs |

## Guided command payload contracts

Foundation owns the common envelope, command discriminator, diagnostics, source ranges, and exit mapping. Guided workflows extend the shared command support without adding a preset-specific command family. Mutating payloads carry an expected workspace/work revision and idempotency key; action submissions carry `actionId` and `actionRevision`. Read-only `next` may be called with only an explicit work ID or selection and returns the expected/action revisions to use for a later submission. A `claim` is optional for read-only commands and required when trusted policy selects a lease/approval authority. All revisions and digests are checked again at application/transition time.

These guided command discriminators publish a new minor protocol schema (`1.1`) referencing unchanged foundation value types. They do not mutate the closed `1.0` schema or create a second envelope. Version/capability output declares supported protocol versions; an older binary rejects unsupported versions/commands with a structured usage diagnostic rather than accepting an incomplete payload.

```text
StartCommand {
  type: "start", dryRun, intent, workflow?, root?, workspaceId,
  expectedWorkspaceRevision, idempotencyKey? // required when dryRun=false
}
NextCommand {
  type: "next", root?, workId? // absent workId requires one unambiguous active item
}
SubmitCommand {
  type: "submit", root?, workspaceId, workId, actionId, actionRevision, expectedRevision,
  idempotencyKey, payload: ArtifactSubmission | DraftBundle | MaterialityRiskAssessment | Approval,
  claim?
}
ArtifactSubmission { kind: "artifact", artifact, inputDigest, policyDigest,
  workflowDefinitionDigest, validatorDefinitionDigest }
DraftBundle { kind: "draft-bundle", requirements, design, tasks, inputDigest,
  policyDigest, workflowDefinitionDigest, validatorDefinitionDigest }
MaterialityRiskAssessment { kind: "materiality-risk", assessment, evidence[], provenance,
  confidence, uncertainty[] }
Approval { kind: "approval", targetArtifactSetDigest, approvalEvidence, claim? }
FinishCommand {
  type: "finish", root?, workspaceId, workId, expectedRevision, actionId, actionRevision,
  idempotencyKey, candidateSnapshotDigest, claim?
}
```

The `SubmitCommand.payload` discriminator supports ordinary Bug Fix artifacts and makes a Quick Spec draft bundle atomic (requirements/design/tasks together), while assessments and approvals use the same shared submission seam. The selected action defines allowed artifact kinds; a draft-bundle action rejects three separate artifact submissions. Approval evidence is mandatory whenever policy requires approval; a lease claim is not itself reviewer identity or approval. Same idempotency key plus identical canonical content replays the prior outcome; the same key with changed content is a conflict. An uncertain transition returns recovery-required with the prepared action/transaction. The command/result schemas must discriminate these payloads without relying on human messages. The normal CLI captures revisions and an idempotency key internally, exposes that key in plan/outcome for recovery, and never silently replaces it on an uncertain retry. Agent-facing payloads expose these coordinates explicitly.

The guided minor schema defines result kinds `work-started`, `start-dry-run`, `next-action`, `submission`, and `completion`, tied respectively to start, next, submit, and finish. `next-action` carries workspace/work revisions, action ID/revision, resolved policy/definition digests, input references, allowed submission schema and blockers. It may truthfully return `blocked` with exit `0` because observation succeeded; an attempted prohibited mutation returns the relevant nonzero gate category. `submission` includes before/after revisions, replay status, artifact IDs/digests and receipt reference; `completion` includes final candidate digest and DONE/receipt paths. Real integration is blocked until the versioned schemas and positive/negative fixtures for these result types exist.

## Draft bundle and authoritative validation

The generation packet includes accepted intent, clarification records, accepted specs, trusted policy snapshot, targeted evidence/provenance, artifact schemas and stable-ID rules, the pinned scope baseline and its digest, canonical output paths, expected workspace/work revision, byte/token budgets, validator definition digests, and completion conditions. The core never invokes an LLM.

Each context reference includes stable identity, captured source revision, content digest, normalized locator, authority classification, sensitivity and requiredness. References resolve to the captured content, never silently to current mutable bytes. Byte budgets are enforced deterministically; a token budget names its estimator/version and is advisory unless that tokenizer is available. Optional context may be omitted with recorded reasons. Required context that cannot fit is replaced by resolvable bounded references; if even the required manifest cannot fit, return a budget diagnostic without emitting a misleading complete packet or advancing state. Repository prose is evidence, not instructions or a grant to run tools.

Bundle submission is one transaction containing requirements, design, and tasks. Each artifact includes stable ID, work/source revision, blob/semantic/derivation digest. Design references requirements derivation; tasks reference requirements/design IDs and derivations, have a deterministic dependency DAG, and cover implementation and verification.

Validation derives mandatory validators from trusted policy and the selected profile, not from an agent payload. It runs each validator against the candidate snapshot and records:

```text
validatorsExecuted[] = {
  validatorID, definitionDigest, candidateDigest,
  status (passed|failed|cancelled|stale|missing), evidenceRefs, executedAt
}
```

`validatorsExecuted` is the sole canonical field name; `fieldValidators` is not an alias. Agent claims are provenance only. Empty, missing, stale, cancelled, or unverifiable required evidence is never `passed`.

Same submission key plus identical canonical content returns the original result. The same key with changed content returns conflict. A crash/timeout with uncertain commit returns recovery-required and leaves the prepared transaction inspectable. No artifact is partially advanced.

## Freshness and digest rules

Artifact canonicalization is typed and versioned. It preserves meaningful Markdown whitespace/code/order and unknown fields; semantic no-op formatting can retain dependent freshness, while behavior-changing or uncertain unknown changes invalidate conservatively. Blob, semantic, and derivation digests are separate. Requirements changes stale design and affected tasks; design changes stale affected tasks. Workflow, policy, or validator definition changes stale any result whose derivation binds the old definition.

Escalation creates a new revision of the same immutable work ID, retaining accepted clarification answers, history, artifact IDs, and valid artifacts. It recomputes composition/profile/digests and records reasons/evidence. It never silently rewrites the previous accepted revision.

## Completion lifecycle

The shared `completion` owner evaluates inputs before writing outputs. It requires accepted/fresh artifacts, implementation state, independent verification against an unchanged candidate snapshot, policy review where selected, residual uncertainty/effect disposition, relevant workflow gates, and the pinned scope baseline. Every applicable Must must be satisfied or changed through an authorized new baseline; Should/Could deferrals must retain their explicit workaround/follow-up and Wont remains visible as excluded scope. Only after all checks pass does that owner generate `done.md`, a machine completion receipt, and lifecycle events atomically. Guided handlers request the completion operation; they do not write outputs directly. Pre-existing `done.md` or receipt files are not proof of completion and cannot satisfy a gate. Post-verification edits change the candidate digest and invalidate the gate.

## Public schemas and diagnostics

Guided results use the foundation envelope and valid schema references. Artifact schemas are discriminated by stable kind and version and use structured source/provenance. Diagnostics are registered as:

```text
BUG_REPRODUCTION_REQUIRED
BUG_REPRODUCTION_UNASSESSED
BUG_DIAGNOSIS_UNSUPPORTED
BUG_REGRESSION_REQUIRED
BUG_VERIFICATION_REQUIRED
BUG_SPEC_CONFLICT_UNRESOLVED
BUG_TEMPORARY_PROBE_UNRESOLVED
QUICK_SPEC_CLARIFICATION_REQUIRED
QUICK_SPEC_BUNDLE_INCOMPLETE
QUICK_SPEC_BUNDLE_REJECTED
QUICK_SPEC_TRACEABILITY_INVALID
QUICK_SPEC_BUNDLE_STALE
QUICK_SPEC_APPROVAL_REQUIRED
WORKFLOW_ESCALATION_REQUIRED
WORKFLOW_POLICY_UNAVAILABLE
WORKFLOW_RECOVERY_REQUIRED
```

Each has stable severity, exit category, source, remediation, and retryability. Human output remains concise; machine output carries IDs, revisions, digests, structured args, and evidence refs.

For attempted transitions, missing reproduction/diagnosis/regression/verification, unresolved spec conflict or effects, incomplete/rejected/untraceable bundles and unresolved clarification use exit `1`; stale bundles/revisions and prepared recovery use `4`; missing/invalid authority, required approval or unresolved policy escalation use `6`. Required validator process failure uses `5`, malformed payload `2`, corrupt workspace/schema `3`, and cancellation `130`. Read-only `next`/`status` may successfully report these blockers without claiming the gated action succeeded. The checked-in catalog must supply exact source/remediation metadata and engine tests for every registered code before integration.

## Fixtures and verification strategy

Conformance must execute full journeys end-to-end and include all five reproduction states; unsafe proxy rejection; intermittent evidence; symptom fixed with unsupported diagnosis; implementation/spec defect; critical auth risk retaining Bug Fix with critical rigor; emergency repair and retrospective gate; retained probe; zero/one/several questions; unavailable host/manual packet; partial/malformed/duplicate/untraceable/stale bundles; same-key changed content; crash uncertainty; forged validator claims; changed definitions; policy removal/expiry/revocation; dirty post-verification code; escalation preservation; completion atomicity; stdout/stderr and cancellation.

## Dependencies and non-goals

This design consumes `sdd-engine-foundation`, `sdd-workspace-transactions`, `sdd-workflow-graph`, `sdd-artifact-validation`, and `sdd-completion-lifecycle`. Fakes can validate the graph and packet seams before those specs land, but durable integration and release claims are blocked until their real contracts conform. No automatic Git commit/PR/merge, embedded model, Bun claim, or arbitrary performance assertion is part of this design.

## References

- `requirements.md`
- `docs/architecture/guided-workflows.md`
- `docs/kiro-input/sdd-cli-spec-brief.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/cli-ux.md`
