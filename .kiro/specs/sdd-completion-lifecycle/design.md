# Design Document: sdd-completion-lifecycle

## Draft status and design stance

This is a provisional design derived from the current completion requirements
and the provisional prerequisite contracts. It is not user approval, a public
schema allocation, an implementation claim, or evidence that any fixture has
run. Every interface below remains subject to the blockers in
`docs/kiro-input/prerequisite-review.md`.

Completion is a concrete, in-process deep module. It owns the question “may
this exact work item be accepted for this exact candidate under this exact
policy?” and returns either deterministic blockers or one exact completion
proposal. It never parses Markdown, executes a process, acquires a claim,
writes a file, advances a revision, or repairs a prepared transaction.

The central invariant is:

```text
one coherent observation
  + owner-produced normalized facts
  + explicit authority observations
  -> one pure completion decision
  -> one exact, preconditioned workspace plan
  -> committed/replayed workspace outcome before “completed” is reported
```

Physical files are not promised to change atomically. Workspace may expose
intermediate replacements after `PREPARED`; all engine readers then report
`recovery-required` until workspace commits or quarantines the recorded intent.

## Scope and explicit ownership exclusions

Completion owns:

- candidate read-set selection rules and content identity;
- candidate-bound evidence acceptance, replay eligibility and freshness;
- reviewer approvals, independence rules and scoped waivers;
- deterministic completion-gate evaluation;
- completion lifecycle decisions;
- the accepted fact set from which both `done.md` and the receipt are rendered;
- acceptance of pure three-way comparison facts;
- exact DONE, receipt, lifecycle, promotion and archive effect intentions;
- archive, reopen and supersession domain semantics;
- completion-specific privacy, redaction and retention dispositions.

Completion explicitly does not own:

- raw snapshot acquisition, path containment, revisions, locks, plan
  application, idempotency storage, journals or crash recovery
  (`sdd-workspace-transactions`);
- artifact parsing, common envelopes, IDs, source ranges, canonicalization,
  blob/semantic/derivation digests or pure delta comparison
  (`sdd-artifact-validation`);
- generic graph compilation, routing, freshness propagation or next-action
  selection (`sdd-workflow-graph`);
- subprocesses, process trees, resource enforcement, execution claims/leases,
  worktrees or provider isolation (`sdd-bounded-execution` and the declared
  provider);
- Bug Fix/Quick Spec topology or journey-specific profiles
  (`sdd-guided-workflows`);
- CLI presentation, protocol envelope/version allocation or exit taxonomy
  (`sdd-engine-foundation`).

An agent assertion, local role string, editable receipt file, lease, or
pre-existing `done.md` grants none of these authorities.

## Requirement-to-design traceability

| Requirement | Design realization |
|---|---|
| CL-REQ-001 | Closed work/attempt state model and transition/authority table |
| CL-REQ-002 | `CandidateReadRequest`, membership/absence observations and `CandidateSubject` |
| CL-REQ-003 | policy-owned exclusions, content-only candidate digest and acyclic output graph |
| CL-REQ-004 | closed verification states and discriminated evidence records |
| CL-REQ-005 | `RegressionEvidence` with before/after, baseline, flaky and proxy facts |
| CL-REQ-006 | evidence bindings, trust/freshness evaluation and replay key |
| CL-REQ-007 | `Approval`, `Waiver` and policy-owned authority/independence rules |
| CL-REQ-008 | ordered gate descriptors/results and fail-closed evaluation passes |
| CL-REQ-009 | profile-shaped DONE view rendered only from accepted facts |
| CL-REQ-010 | digest/symbol-bearing `CodeReference` and `GuidelineReference` |
| CL-REQ-011 | preallocated IDs, canonical receipt body and no-self-hash scheme |
| CL-REQ-012 | exact workspace plan handoff, apply-outcome interpretation and recovery-required behavior |
| CL-REQ-013 | pure IC-08 `DeltaResult` consumption and completion-owned acceptance decision |
| CL-REQ-014 | archive preservation, collision and immutable discovery semantics |
| CL-REQ-015 | workspace idempotency/recovery consumption and explicit partial-scope disposition |
| CL-REQ-016 | closure-based staleness, read-only current findings, reopen and supersession |
| CL-REQ-017 | redacted evidence references, retention class/availability and renderer parity |

## Module seams and dependency direction

The completion implementation is concrete. It has no parser, repository,
policy-provider, filesystem or process interfaces. Its small transformation
surface is:

```go
package completion

func SelectCandidate(input SubjectSelectionInput) (CandidateReadRequest, []Diagnostic)

func Evaluate(input EvaluationInput) Decision
```

`SelectCandidate` compiles trusted policy and typed work/artifact facts into a
bounded observation request. `Evaluate` consumes the corresponding coherent
snapshot plus all owner-produced facts and returns a closed decision tagged
`blocked`, `accepted`, `stale`, `conflict`, or `recovery-required`. An accepted
decision contains the exact completion proposal; it is not a commit result.

Operational failure while the engine obtains a snapshot, validates provider
proof, or calls workspace remains an ordinary Go `error` translated once by
the engine. Missing, stale, rejected or insufficient domain facts are values in
`Decision`, not a generic `Result[T]` or Boolean.

```text
presentation -> engine
                  |
                  +-> workspace: coherent raw snapshot; apply/recover outcome
                  +-> artifact-validation: artifacts, policy, validation,
                  |                       traceability and pure delta facts
                  +-> workflow: composition and graph-complete handoff facts
                  +-> completion: candidate selection + pure decision/proposal
                  +-> execution/reviewer providers: already-produced facts
```

The engine moves values between owner modules. Completion may consume
owner-defined value contracts without importing workspace, parser, provider or
platform implementations. Returned slices and byte buffers are copied or
ownership-transferred; maps are lookup indexes only and never determine hashes,
diagnostic order or output order.

## Plain Go values and immutable input

The declarations below are design-level shapes, not a public Go SDK.

```go
type EvaluationInput struct {
    Request          CompletionRequest
    Snapshot         RawCompletionSnapshot
    CandidateRequest CandidateReadRequest
    Artifacts        []ArtifactFact
    Traceability     TraceabilityFact
    Validation       []ValidationFact
    Composition      CompositionFact
    GraphHandoff     GraphCompletionHandoff
    Evidence         []Evidence
    Approvals        []Approval
    Waivers          []Waiver
    DeltaResults     []DeltaResult
    Policy           CompletionPolicy
    Observations     CurrentObservations
    Allocations      PreallocatedIDs
    Limits           Limits
}

type CompletionRequest struct {
    Operation       Operation // finish | archive | reopen | supersede | abandon
    WorkspaceID     string
    WorkID          string
    ExpectedRevision uint64
    Idempotency     IdempotencyCoordinates
    TargetWorkID    string // required only for supersede when policy requires it
}

type PreallocatedIDs struct {
    CandidateID  string
    AttemptID    string
    ReceiptID    string
    PlanID       string
    TransactionID string
    EventIDs     []string
}
```

IDs and time are inputs so evaluation is pure and repeatable. `Snapshot` is one
coherent workspace observation, not a bag of paths that completion may reopen.
`CurrentObservations` contains the explicit evaluation instant, policy and
issuer revocation observations, evidence-reference availability and transaction
recovery state. Missing observations remain missing; zero values never mean
trusted, current or passed.

The module performs stable passes:

```text
validate input schemas/ownership
 -> derive and verify CandidateSubject
 -> classify evidence, approvals and waivers
 -> evaluate independent gates in profile order
 -> evaluate lifecycle operation and promotion/archive facts
 -> freeze AcceptedCompletionFacts, or deterministic blockers
 -> render DONE body and receipt body from that same fact set
 -> construct exact effects/preconditions and workspace plan
 -> return Decision without applying anything
```

Cancellation is checked by the engine before snapshot/provider/apply effects.
Pure evaluation accepts a cancellation observation only to avoid publishing a
decision after cancellation; it owns no child cleanup and starts no goroutine.

## Candidate observation and content identity

### Bounded read request

`SelectCandidate` returns a request whose selectors come only from the trusted
completion policy, accepted artifacts, validator contracts and declared build
inputs:

```text
CandidateReadRequest {
  schemaVersion,
  workspaceScope,
  selectors[] {
    selectorId, root/path, matchKind, requiredPresence,
    roles[], includeRuleId, exclusionRuleIds[], enumerationLimit
  },
  requiredExternalInputs[],
  environmentSelectors[], toolchainSelectors[],
  exclusionPolicy: {policyId, version, digest, rules[]},
  observerContract: {schemaVersion, stableReadRequirements, limits},
  requestDigest
}
```

Each selector is bounded. A selector that can match membership requests a
deterministic ordered membership observation, not only reads of paths that
already existed. That is what lets a new relevant untracked, ignored or
generated file stale a candidate. Required absence is represented as an
observation and enters the digest; absence is never an omitted slice element.

Exclusion resolution precedence is fail-closed:

1. mandatory declared implementation/build/configuration/dependency inputs are
   included;
2. policy-owned include rules are applied;
3. only policy-owned, versioned bookkeeping exclusions may remove a remaining
   selector;
4. an excluded path that is also a declared input is included and diagnosed as
   an exclusion collision;
5. repository prose, candidate bytes, agents and command payloads cannot add or
   alter exclusion rules.

Typical excluded administrative outputs are DONE, completion receipts,
journal frames and transaction metadata. An organization may choose different
paths, but the resolved rules and policy digest are fixed before workspace
observation and rechecked before apply.

### CandidateSubject

Workspace returns raw ordered observations. Completion verifies that response
against `requestDigest` and derives:

```text
CandidateSubject {
  schemaVersion,
  candidateId,                 // preallocated administrative identity
  candidateDigest,             // digest of CandidateContentCore
  workspaceId, workId,
  workRevisionAtObservation,
  snapshotRef,
  observationMode: git | no-git,
  entries[] CandidateEntry,
  membership[] MembershipObservation,
  absence[] AbsenceObservation,
  configurationDigest,
  generatedInputDigest,
  dependencyLockDigest,
  toolchainDigest,
  environmentObservationDigest,
  exclusionPolicyRef,
  gitProvenance?, noGitManifest?,
  observedAt, observerProvenance
}

CandidateEntry {
  locator, role,
  workspaceClass: tracked | untracked | ignored | generated |
                  dependency | submodule | external,
  presence: present | absent,
  blobOrContentDigest?, size?, mode?, linkObservation?,
  inclusionRuleId, inclusionReason,
  provenance: {sourceKind, sourceContract, sourceContentIdentity}
}
```

Configuration files are entries with a configuration role; dependencies and
submodules carry resolved identities required by policy. A submodule pointer
alone is insufficient when the policy requires resolved content. A no-Git
candidate contains a canonical ordered manifest and never fabricates a commit.
Git branch, HEAD and dirty status are provenance only; the selected bytes,
membership and absences remain the subject.

### Candidate digest core

The content digest is:

```text
SHA-256(
  "sdd.completion-candidate-content/1\x00" ||
  CanonicalEncode(CandidateContentCore)
)
```

`CandidateContentCore` contains the subject schema/canonicalizer versions,
observation mode, ordered entries, membership and absence observations,
configuration/generated/dependency/toolchain/environment content identities,
resolved exclusion-policy digest, observer semantic-contract identity, and
non-administrative provenance needed to interpret those facts.

It deliberately excludes candidate/snapshot/attempt/workspace/work/plan/
transaction/event IDs, administrative revisions, retry coordinates, mutable
display slugs, journal offsets and incidental observation timestamps. Those
remain on `CandidateSubject` and the receipt for scope, concurrency and audit.
Consequently identical selected content can have the same content digest in a
different administrative observation, while copying its receipt to another
work still fails because the receipt separately binds workspace/work IDs,
revision, candidate ID and policy.

Provenance is not discarded: source class, semantic observer contract,
inclusion rule and immutable source-content identity are in the digest core;
provider IDs, locators and observed-at metadata remain explicit subject fields.
An unstable, unreadable, missing-required, over-budget or unsupported platform
observation blocks candidate construction rather than producing a partial
digest.

## Evidence, verification and regression values

### Closed evidence union

Evidence is a discriminated record with a common binding and exactly one
variant. The wire schema uses `oneOf`; Go uses a closed tag with concrete value
fields and rejects populated inactive variants.

```text
Evidence {
  schemaVersion, evidenceId, kind,
  binding: {
    workId, actionOrGateScope, candidateId, candidateDigest,
    artifactInputs[], policyDigest, profileDigest,
    environmentDigest, validatorSemanticContractDigest?
  },
  authority: {issuer, authenticityBasis, scope, trustObservation},
  resultDigest, redaction, retention, issuedAt, expiresAt?,
  executed? | manual? | proxy? | review?
}

ExecutedEvidence {
  executableIdentity, argv[], cwd, environmentIdentity,
  providerReceiptRef, providerImplementationProvenance,
  executionDisposition, exitOrSignal?, outputDigests[],
  truncation, enforcedLimits, cleanupObservation
}

ManualEvidence {
  approvedMethodId, actor, orderedSteps[], observations[],
  environmentIdentity, attachments[], limitations[]
}

ProxyEvidence {
  approvedProxyMethodId, actorOrProvider, targetRelationship,
  orderedSteps[], observations[], attachments[], limitations[], safetyReviewRef
}

ReviewEvidence {
  reviewMethodId, reviewer, reviewedSubjects[], findings[], dispositions[],
  limitations[]
}
```

Manual/proxy/review variants contain no fabricated executable, exit code or
provider invocation. An editable local file can carry the record but cannot
establish issuer authenticity by itself. The active policy says which issuer
facts and evidence classes may satisfy each gate.

### Verification states

| State | Meaning | Can satisfy a pass-required gate? |
|---|---|---:|
| `planned` | check is described but has not begun | no |
| `not-run` | no attempt occurred | no |
| `running` | provider/manual procedure is incomplete | no |
| `passed` | interpretable accepted result proves the declared condition | yes, subject to authority/freshness |
| `failed` | check ran and the condition was negative | no |
| `inconclusive` | result cannot establish either outcome | no |
| `unavailable` | provider/reference/capability was not available | no |
| `stale` | a bound identity or trust observation changed | no |
| `waived` | a separate valid waiver disposes an eligible gate | only as `WAIVED`, never `PASS` |

Provider dispositions `partial`, `timed-out`, `cancelled`, `truncated`,
`malformed` and `unverifiable` map to `inconclusive` or `unavailable` with the
original disposition retained. They never map to `failed` or `passed` merely to
fit a Boolean API.

### Regression and baseline evidence

```text
RegressionEvidence {
  contractRef,
  failingBefore: {expectedFailure, candidate, environment, oracle, observation},
  passingAfter:  {candidate, environment, oracle, evidenceRef},
  negativeAndSafetyEvidence[],
  reproductionDisposition,
  baseline?: {fingerprint, acceptedFailures[], comparisonResult},
  flakiness?: {attempts[], passRate, policyThreshold, thresholdResult},
  alternative?: {kind: manual | proxy | incident, evidenceRef, limitations[]}
}
```

Before and after subjects must be distinct where the contract requires a
failing-before/passing-after demonstration. A known failing baseline preserves
its fingerprint and proves only the policy-defined “no new matching failure”
relationship; it does not turn the full check into `passed`. Retrying a flaky
check records every attempt and applies the pinned policy threshold. The
`not-reproduced`, `intermittent`, `environment-specific` and
`unsafe-to-reproduce` dispositions remain visible in the accepted fact set.

## Authority, approvals, waivers and replay

```text
Approval {
  schemaVersion, approvalId,
  issuer, reviewerIdentity, authenticityBasis,
  workId, candidateId, candidateDigest,
  artifactDigestSet, policyDigest, definitionDigest,
  gateScopes[], decision,
  issuedAt, expiresAt?, revocationObservation,
  evidenceRef, authorImplementerRelationships[]
}

Waiver {
  schemaVersion, waiverId,
  issuer, authenticityBasis,
  workId, candidateId, candidateDigest,
  eligibleGateScopes[], diagnosticScopes[], rationale,
  policyDigest, issuedAt, expiresAt, revocationObservation,
  retrospectiveGateIds[], evidenceRef
}
```

`CompletionPolicy` supplies issuer allowlists/trust rules, reviewer
independence predicates, evidence-class permissions, gate waiver eligibility,
expiry/revocation rules and retrospective obligations. Completion evaluates
these rules from explicit observations. A display name, role string, lease or
`approved: true` field is never an authority fact.

Candidate binding, receipt integrity and authentic waiver authority are
non-waivable. A waiver produces `WAIVED` for its exact eligible scope; it does
not change evidence state to passed or authenticate an untrusted issuer.

Evidence replay uses a canonical replay key over candidate content identity,
artifact inputs, method/argv/environment, policy/profile, validator semantic
contract, authority and relevant expiry/revocation observations. Matching the
key makes the acceptance decision deterministic, but workspace idempotency is
still the owner of action replay. A previously committed receipt remains a
historical fact even if a new attempt now finds evidence expired; it is not
replayed as a current pass.

## Deterministic gate evaluator

The policy supplies an ordered, versioned gate profile:

```text
GateDescriptor {
  gateId, orderKey, class, subjectSelector,
  requiredEvidenceClasses[], requiredCheckStates[],
  approvalRule?, waiverEligibility,
  applicability, failureCode, remediationTemplate
}

GateResult {
  gateId, orderKey,
  status: pass | fail | blocked | waived | not-applicable,
  subjectRefs[], evidenceRefs[], approvalRefs[], waiverRef?,
  policyRuleRefs[], diagnostics[], freshnessBindings[]
}
```

Independent gates are all evaluated in `(orderKey, gateId)` order so one
missing check does not hide other actionable blockers. Unsafe plan generation
stops after the decision is known to be non-accepting, but already-known gate
facts remain reported. `unknown`, missing, stale, unavailable or uninterpretable
facts fail closed.

The built-in profile vocabulary covers:

| Gate class | Required owner-produced facts |
|---|---|
| artifact/task freshness | accepted artifacts, derivation closure, task state |
| coverage/traceability | requirement/scenario implementation and verification edges |
| changed paths | every selected path mapped to task/generated/incidental disposition |
| verification | same-candidate checks for selected test/lint/type/build/migration/performance/security classes |
| guidelines | current digest/clause and followed/not-applicable/exception disposition |
| discoveries/convergence | read-only findings and authorized reconciliation facts |
| migration/rollback | rollout, rollback/irreversibility, compatibility and operator evidence |
| review/authority | authentic current approvals and independence |
| workspace safety | no blocking diagnostics, expired required claims or unresolved prepared transactions |
| promotion | conflict-free accepted three-way comparison |
| necessity | new dependency/interface/config/process/extension rationale and simplification disposition |
| requested verification | every requested step has an interpretable candidate-bound result |

Applicability is a trusted profile result, not filler. Documentation-only,
research/spike, generated-output, feature-flag and manual-verification profiles
can mark a gate `not-applicable` only with the defining rule in the result.

The evaluator returns either ordered blockers and one safe next action, or an
immutable `AcceptedCompletionFacts`. It never infers acceptance from a file's
existence or from a graph node merely being terminal.

## Lifecycle and operation transitions

Work lifecycle is separate from attempt, verification and transaction state.

| Current work state | Operation/event | Required authority and facts | Proposed next state | Forbidden shortcut |
|---|---|---|---|---|
| `open` | begin finish | graph prerequisites ready, coherent candidate | `verifying` | direct `completed` |
| `verifying` | evidence accepted | required evidence current | `review` | unavailable check as pass |
| `verifying` | failure/cancel | recorded attempt outcome | `open` | terminalize work automatically |
| `verifying` | bound input drift | current observation | `stale` | reuse stale evidence |
| `review` | all gates accepted | current policy/approvals | `finish-pending` | existing DONE as authority |
| `review` | gate negative/drift | named result | `open` or `stale` | hide blocker |
| `finish-pending` | workspace `committed`/authorized replay | exact plan and commit summary | `completed` | report completed at PREPARED |
| `finish-pending` | pre-PREPARED conflict | owner outcome | `stale` or `open` | silent re-plan |
| `finish-pending` | prepared/incomplete outcome | owner outcome | remains pending; report `recovery-required` | guessed rollback/success |
| `completed` | authorized reopen | new attempt allocation, historical receipt retained | `reopened` | rewrite old receipt |
| `completed` | authorized supersession | later work identity and relationship | `superseded` | relabel incomplete |
| `completed` | authorized archive | collision-free exact archive plan | `archived` | deletion/branch removal |
| `reopened` | begin new evaluation | new candidate/attempt | `verifying` | reuse historical pass blindly |
| `open`/`reopened` | authorized abandon | rationale/disposition | `abandoned` | claim delivery |
| `abandoned` | authorized reopen | policy permits, new attempt | `reopened` | erase abandonment |
| `abandoned`/`superseded` | archive | preservation plan | `archived` | delete evidence |
| `archived` | correction | erratum, reopen or superseding work | separate new fact | in-place mutation |

A read-only status can derive “currently stale” from current observations
without persisting a lifecycle transition. Only a separately authorized plan
records reopen, supersession, abandonment or archive.

## Durable code and guideline references

```text
CodeReference {
  workspaceRelativePath,
  symbolOrPointer?, testOrSchemaPointer?, optionalLineRange?,
  candidateDigest, blobOrImmutableRevisionDigest,
  taskIds[], requirementIds[]
}

GuidelineReference {
  workspaceRelativePath, sectionAnchor,
  semanticDigest, declaredVersion?, applicableClause,
  disposition: followed | not-applicable | exception-approved,
  exceptionRef?
}
```

Line ranges are locators only. Completion verifies path plus symbol/pointer and
immutable content identity. A moved symbol may be found through owner-produced
artifact/code-reference facts, but changed current bytes still participate in
candidate freshness. Missing or stale required guideline/reference facts
block; they are never silently dropped from DONE.

## One accepted fact source for DONE and receipt

After every required gate passes, completion freezes:

```text
AcceptedCompletionFacts {
  schemaVersion, attemptId,
  workScope, candidateSubject,
  artifactDigestSet, policyRef, profileRef, definitionRefs,
  scopeBaselineRef?, scopeBaselineDigest?,
  scopePrioritySummary?: {
    must: { delivered[], blocking[] },
    should: { delivered[], deferred[] },
    could: { delivered[], deferred[] },
    wont: { excluded[] }
  },
  gateResults[], evidenceRefs[], approvalRefs[], waiverRefs[],
  deliveredScope[], undeliveredScope[], followUpWorkRefs[],
  codeReferences[], guidelineReferences[],
  plannedOrder[], actualExecutionGroups[],
  promotionFacts[], lifecycleDecision,
  securityPrivacyMigrationOperations,
  deviations[], residualRisks[], reproductionInstructions[],
  retentionDispositions[]
}
```

Both renderers consume this exact immutable value. DONE is a human view; the
receipt is the machine view. Neither reparses the other, and neither builds a
second gate table. This guarantees that a gate cannot say `WAIVED` in the
receipt but `PASS` in DONE.

DONE has profile-selected sections for summary, delivered/undelivered scope,
requirements/scenarios, decisions, durable references, planned/actual order,
verification, deviations/waivers/residual risk/follow-ups,
security/privacy/migration/operations, acceptance/reproduction and the final
gate table. Omitted sections carry an explicit profile disposition. Narrative
inputs may come from an agent, but the final acceptance stamp and gate rows are
engine-rendered from accepted facts.

When a trusted profile supplies a scope baseline, the fact builder consumes the
artifact-validation-owned, profile-validated `ScopePriority` values and the
pinned baseline digest. It records every applicable Must as delivered or
blocking; it records a Should/Could deferral with its authorized disposition
and follow-up; and it records Won't-have subjects as exclusions. Completion
does not estimate capacity, infer a priority, change a category, or interpret
a feature flag as implementation. A changed Must or capacity-based reduction
requires a newly authorized baseline before a fresh evaluation can pass.

## Acyclic receipt and plan linkage

All administrative IDs are preallocated and supplied in `EvaluationInput`.
The dependency order is:

```text
candidate/artifacts/policy/definitions/evidence/approvals
  -> AcceptedCompletionFacts
  -> canonical DONE body containing receiptId (not receiptDigest)
  -> doneDocumentDigest
  -> canonical ReceiptBody containing doneDocumentDigest,
     planId and transactionId (not planDigest or receiptDigest)
  -> receiptDigest = H(ReceiptBody)
  -> exact workspace PlanCore containing DONE/receipt bytes and digests
  -> planDigest = H(PlanCore)
  -> workspace committed summary binds planId + planDigest + receiptDigest
```

`ReceiptBody` contains receipt/attempt IDs, work/revision coordinates,
`CandidateSubject`, accepted artifacts, policy/profile/definition digests,
scope-baseline reference/digest and priority summary when applicable, gate
results, evidence/approval/waiver refs, issuer, preallocated event/transaction
IDs and audit timestamps. Its canonical encoding excludes any
field holding `receiptDigest`. DONE never contains the final receipt or plan
digest. The receipt never contains the final plan digest. Thus neither output
hashes itself or needs a later byte rewrite.

Historical bytes are immutable. A correction is a separately linked erratum,
reopen or superseding work item.

## Exact plan handoff and recovery semantics

An accepted decision returns one `CompletionProposal`:

```text
CompletionProposal {
  proposalSchemaVersion,
  acceptedFactsDigest,
  done: {path, exactBytes, blobDigest, semanticBodyDigest},
  receipt: {path, exactBytes, bodyDigest, receiptDigest},
  promotionEffects[], lifecycleEffects[], archiveEffects[],
  sourceSnapshotBinding,
  candidate/policy/definition/baseline/currentTarget bindings,
  expected workspace/work/artifact revisions,
  authorizationScope,
  idempotencyCoordinates,
  workspacePlan: {preallocated planId, exact owner-defined PlanCore, planDigest}
}
```

Completion uses the workspace-owned effect vocabulary and canonical plan codec;
it does not define a competing plan hash. Engine hands the exact plan, content
blobs, authorization fact and idempotency coordinates to
`workspace.Apply`. Workspace rechecks raw observations, revisions,
authorization, exclusion policy and target containment before `PREPARED` and
never asks completion to re-plan inside apply.

The engine interprets workspace outcomes as follows:

| Workspace outcome | Completion/public meaning |
|---|---|
| `committed` | report `completed`; publish committed receipt/location |
| `replayed-committed` | return original committed receipt/result after disclosure check |
| `conflict` before PREPARED | report stale/conflict; a new observation/evaluation/authorization is required |
| `prepared`/`recovering`/uncertain reply | report `recovery-required` with transaction ID and owner-provided action |
| `quarantined` | report `recovery-required`; preserve exact intent/evidence and require authorized resolution |
| failed before PREPARED | no completion; retain attempt outcome according to the authorized follow-up plan |

No engine reader accepts a partial DONE/receipt/archive/promotion set as
completed. Read-only `status` reports recovery-required and does not recover.
After `PREPARED`, recovery rolls forward or quarantines the already-recorded
workspace intent; it does not rerun completion gates, acquire replacement
authority, invent rollback, or generate new bytes under the old plan ID.

Same namespace/key plus the same workspace request fingerprint replays the
owner outcome. Changed candidate, policy, content, plan or authority under the
same key conflicts. Evidence expiry on a new attempt does not invalidate an
already committed historical receipt, and a lost response does not create a
second receipt.

## Pure three-way promotion consumption

Artifact-validation owns IC-08 and returns `DeltaResult` for pinned
`base + proposal + current` values. Completion consumes:

- resolved scoped target IDs and complete base identity;
- ordered proposed/current entity changes and preserved entities;
- `base-missing`, `base-mismatch`, `current-changed`, `same-entity-change`,
  `delete-modify`, `rename-collision`, `id-collision` and
  `invalid-reference` facts;
- schema/trace validation facts, delta digest and source coordinates;
- for an allowed disjoint case, owner-produced exact merged target bytes or a
  canonical target value/digest that the artifact owner can render without a
  second parser.

Completion, not the parser, decides whether policy permits promotion. An
unchanged current target or structurally disjoint changes can be accepted only
when unrelated content is preserved and the resulting target is schema-valid.
Every semantic conflict blocks. Completion never chooses ours/theirs or
default ordering. A reviewer resolution is a new artifact proposal against a
fresh base/current observation, with new approval and a new exact plan.

IC-08 does not yet finalize the exact conflict-free merged-target field. Until
that owner amendment is accepted, completion design can test comparison
consumption with typed fixtures but cannot claim production promotion output.

## Archive, reopen, supersede and partial scope

Archive is a distinct authorized plan. It preserves immutable work/artifact/
receipt identities, lifecycle history, candidate/provenance/evidence/approval
references and discovery metadata. Destination absence, archive format and
identity/index observations are exact preconditions. A collision blocks; no
overwrite, branch deletion, evidence deletion or retention action is implied.

Reopen preserves the historical receipt/candidate and allocates a new attempt,
administrative revision and candidate observation. Supersede links the old
work and a later immutable work ID with exact accepted/replaced scope. Neither
rewrites the earlier record. Post-completion edits create a derived current
staleness finding; a read-only command does not persist reopened state.

No-change, documentation-only, research and manual-verification work may finish
only through a profile that explicitly defines the relevant gates. Partial
delivery freezes delivered and undelivered scope and requires either a
traceable follow-up work ID or an authorized superseding/removal revision.
Unchecked required work cannot remain beneath a DONE stamp.

## Privacy, redaction, retention and reference availability

Evidence storage follows minimization, not “save every log.” Each retained
reference records:

```text
EvidenceReference {
  referenceId, contentDigest, subjectDigest,
  issuer/provenance, redactionProfile,
  retentionClass, availability: available | redacted | expired | missing,
  verificationRoute, boundedExcerpt?, externalLocator?
}
```

Secrets are removed before persistence and rendering. Environment values are
represented by approved names/value-reference digests, not copied into the
receipt. Bounded excerpts are retained only when policy permits; otherwise the
receipt carries result/output digests and an independently usable verification
route. Redaction must preserve whether a result is interpretable. If required
evidence becomes unavailable, policy decides whether the current attempt is
blocked, stale or requires re-attestation; it never silently passes.

Retention policy is an explicit external decision (`DEC-03`). Completion
records the selected class and expiry/reference-availability facts but does not
delete or compact evidence. Archive is never a retention operation.

## Engine.Execute flows

### Finish

```text
decode Finish command through foundation envelope
 -> workspace Snapshot using CandidateReadRequest
 -> artifact-validation parse/normalize/validate/delta facts
 -> workflow graph-complete handoff and composition facts
 -> completion.Evaluate (pure)
 -> if blocked: render ordered gates + one safe next action; no mutation
 -> if accepted: expose exact plan for dry-run/authorization
 -> workspace Apply(exact plan, authorization, idempotency)
 -> interpret committed/replay/conflict/recovery-required owner outcome
 -> render human and JSON from one Outcome
```

### Archive, reopen and supersede

These commands use the same snapshot → pure decision → exact plan → authorize
→ workspace apply rail. Archive never piggybacks deletion. Reopen/supersede
always create new facts linked to immutable history. Status performs only the
snapshot and pure evaluation portions.

## Public result and diagnostic extension proposal

Foundation owns the closed envelope and extension allocation. Completion
proposes, but does not assign a version to, one nested result variant:

```text
CompletionResult {
  kind: completion,
  operation: finish | archive | reopen | supersede | abandon,
  status: blocked | proposal-ready | committed | replayed |
          stale | conflict | recovery-required,
  workId, workRevision,
  attemptId?, candidateRef?, acceptedFactsDigest?,
  gateResults[], blockers[],
  planRef?, transactionRef?, receiptRef?, archiveRef?,
  nextAction?
}
```

This is an AM-02 foundation-extension proposal only. It reuses foundation
`Outcome`, `Diagnostic`, `NextAction`, source-coordinate units, severities and
exit codes. It does not claim guided's provisional `1.1`, mutate foundation
`1.0`, or allocate another envelope.

Proposed completion catalog entries are:

| Code | Meaning | Exit category when an attempted mutation fails |
|---|---|---:|
| `COMPLETION_GATE_BLOCKED` | required gate not satisfied | 1 |
| `COMPLETION_CANDIDATE_STALE` | subject/input changed | 4 |
| `COMPLETION_EVIDENCE_UNAVAILABLE` | mandatory evidence/provider reference unavailable | 5 when provider failure; otherwise policy-defined 1/6 |
| `COMPLETION_APPROVAL_EXPIRED` | current approval/issuer fact invalid | 6 |
| `COMPLETION_PROMOTION_CONFLICT` | three-way facts require new proposal | 4 |
| `COMPLETION_RECOVERY_REQUIRED` | workspace owns prepared/quarantined recovery | 4 |
| `COMPLETION_SCOPE_DEFERRED` | required scope lacks authorized follow-up/supersession | 1 |

AM-01 must reconcile severity separately from blocking behavior before catalog
publication. A read-only observation may return exit `0` while its nested
decision is blocked. Malformed input uses `2`, workspace/schema `3`, external
provider failure `5`, policy denial `6`, and interruption after known cleanup
`130`. Completion does not invent a ninth category.

Human and JSON renderers consume the same structured result. Both name the
blocked gate, candidate/evidence subject and one safe next action. Human prose
is never the authority.

## Trust, failure and resource model

- Repository content, generated narrative, screenshots, tool output and local
  metadata are untrusted until a policy-selected authority makes the relevant
  fact usable.
- Candidate/evidence/approval/policy/definition facts are all compared against
  one immutable evaluation snapshot. Volatile revocation and availability
  observations are explicit inputs and are rechecked where owner contracts
  require.
- Candidate entries, selectors, evidence, references, gates, deltas, effects,
  rendered bytes and diagnostic output have declared limits. Limit exhaustion
  blocks before plan creation; output is never silently truncated into a pass.
- Independent safe gate findings are aggregated. Unsafe provider invocation or
  effect planning is short-circuited.
- Operational errors preserve identity with `errors.Is`/`errors.As` and are
  translated at the engine seam. Completion does not log and return the same
  error or match error strings.
- A fake workspace/provider/reviewer adapter proves integration shape only.
  It cannot satisfy production durability, execution or authority claims.
- Platform support comes from workspace/provider native conformance. Completion
  contains no `runtime.GOOS` branches and makes no filesystem guarantee.

## Conformance fixture design

All fixture IDs are contracts, not executed-test claims. Tests primarily cross
`Engine.Execute`; package tests cover pure candidate/digest/gate/receipt laws,
and owner adapter suites cover durability, execution and authority.

| Fixtures | Designed test surface and expected result |
|---|---|
| CL-FIX-001, CL-FIX-002, CL-FIX-003 | checked tasks, fake pass, forged/self approval and omitted gate remain blocked; no proposal |
| CL-FIX-004, CL-FIX-005, CL-FIX-006 | membership/content edit stales; exclusion tamper blocks; administrative outputs do not self-stale |
| CL-FIX-007, CL-FIX-008, CL-FIX-009 | policy drift rejects exact plan; wrong candidate and unauthenticated/unavailable evidence fail closed |
| CL-FIX-010, CL-FIX-011, CL-FIX-012, CL-FIX-013 | deterministic no-Git, profile-tailored docs, flaky/proxy evidence and expired waiver retain exact dispositions |
| CL-FIX-014, CL-FIX-015 | pure disjoint facts preserve unrelated entities; semantic/delete/rename/ID conflicts block |
| CL-FIX-016 | archive destination collision leaves existing record unchanged |
| CL-FIX-017, CL-FIX-018, CL-FIX-019, CL-FIX-020 | every effect fault exposes workspace recovery/replay; duplicates do not duplicate outputs; new expired attempt differs from committed history |
| CL-FIX-021, CL-FIX-022, CL-FIX-023 | post-completion edit derives staleness; durable references detect drift; reopen preserves old receipt |
| CL-FIX-024, CL-FIX-025 | unfinished scope requires follow-up/supersession; permitted manual evidence binds actor/steps/environment/limitations |
| CL-FIX-026 | candidate/policy/base drift rejects exact plan with no re-plan or partial accepted state |

Canonical digest tests use fixed vectors, shuffled Go map insertion, repeated
evaluation and Unicode/path edge cases. Property tests assert that changing any
decision-bearing field changes its owning digest, while changing only excluded
administrative metadata does not. Fuzz tests target evidence unions, read-set
selectors, reference locators and receipt decoding. Fault injection remains in
workspace/provider suites; completion consumer tests assert the reported owner
outcome rather than reimplementing the fault.

## Shared-contract ledger and amendment proposals

| Contract | Concrete completion consumption/output | Provisional dependency |
|---|---|---|
| IC-01 | coherent raw snapshot, revisions, membership/absence/file observations | workspace `Snapshot`; completion owns selector meaning |
| IC-02 | exact owner-defined plan/effects and apply/recovery outcome | workspace plan codec/effect extension not yet accepted |
| IC-03 | idempotency namespace/fingerprint/replay disclosure | workspace is sole durable owner |
| IC-04 | pinned graph/composition/definition facts | workflow producer |
| IC-05 | graph-complete handoff/blockers | workflow producer; terminal node is not completion |
| IC-06 | journey/rigor/profile/policy-floor facts | workflow producer; guided supplies specialized definitions |
| IC-07 | normalized artifacts and digest set | artifact-validation producer |
| IC-08 | pure base/proposal/current comparison and exact conflict-free target output | artifact-validation amendment required |
| IC-09 | validation result/provider receipt references | artifact-validation/provider producer |
| IC-10 | normalized effective policy and authority observations | AV normalizes; completion enforces its gates |
| IC-11 | `CandidateReadRequest`, `CandidateSubject`, accepted evidence/approval binding | completion owns; workspace/provider supply observations |
| IC-12 | accepted facts, DONE/receipt, promotion/lifecycle/archive exact proposal | completion produces; workspace applies |

Required proposals, not silent source edits:

1. **CL-AM-01 / AM-02 / DEC-05 — foundation result extension.** Foundation
   must allocate the `CompletionResult` discriminator and schemas in the one
   shared extension inventory, preserve closed `1.0` fixtures and test unknown
   negotiation. This design assigns no version.
2. **CL-AM-02 / AM-01 — diagnostic catalog.** Foundation must reconcile
   severity vocabulary while preserving `critical` and keeping blocking a
   separate rule before completion codes publish.
3. **CL-AM-03 / AM-03 — exact effect vocabulary.** Workspace/foundation must
   ratify required move/metadata/event representations and plan linkage.
   Completion does not disguise archive as unrestricted update or expose
   internal journal housekeeping as a domain effect.
4. **CL-AM-04 / IC-08 — conflict-free target output.** Artifact-validation
   must extend its pure `DeltaResult` with the exact normalized/renderable
   conflict-free target or equivalent owner-produced bytes/digest used by the
   plan. Completion will not implement another parser/merge algorithm.
5. **CL-AM-05 / AM-07, AM-09 — guided consumption.** Guided GW-008 and GW-016
   must consume CL candidate/evidence/gate/proposal facts; guided retains only
   journey-specific regression/profile assertions and cannot write completion
   outputs.
6. **CL-AM-06 / AM-10 — compatibility identity.** Consumers must distinguish
   semantic-contract identity from exact implementation/source provenance;
   trust/security changes may stale even when semantics are claimed equal.

## Open decisions and implementation blockers

1. `DEC-01`: trusted policy, reviewer/CI issuers, authenticity and reviewer
   independence are not selected. Production approval remains blocked.
2. `DEC-02`: supported OS/filesystem durability is not selected. Workspace
   integration cannot claim native durability without real runners.
3. `DEC-03`: evidence/reference/idempotency retention and destructive deletion
   policy are not selected. Completion records dispositions but deletes
   nothing.
4. `DEC-04`: production execution-provider transport, isolation and
   independently verifiable receipt are not selected. Mandatory execution
   gates remain unavailable, never fake-passed.
5. `DEC-05`: the common protocol/effect extension and diagnostic reconciliation
   are not allocated. No `1.1` claim is made here.
6. `DEC-06`: emergency waiver authority/eligibility/retrospective policy is not
   selected. No Boolean bypass is designed.
7. IC-08 must publish exact conflict-free target content/digest; IC-01/02/03
   must publish final snapshot, plan, idempotency and recovery value schemas.
8. Policy must finalize read-set selectors, submodule/ignored/generated rules,
   bookkeeping exclusions and evidence/reference retention classes.

These block interoperable schema freeze or production claims, not provisional
pure module design and deterministic fixture work.

## References

- `requirements.md`
- `docs/kiro-input/prerequisite-review.md`
- `docs/kiro-input/next-prerequisite-specs.md`, sections 2–4, 9 and 10
- `docs/cli-product-plan.md`, sections 5.1–5.5, 5.8, 6.6–6.11, 8.2, 8.6 and 10.5
- `.kiro/specs/sdd-engine-foundation/{requirements,design,tasks}.md`
- `.kiro/specs/sdd-workspace-transactions/requirements.md`
- `.kiro/specs/sdd-workflow-graph/requirements.md`
- `.kiro/specs/sdd-artifact-validation/{requirements,design,tasks}.md`
- `.kiro/specs/sdd-guided-workflows/{requirements,design,tasks}.md`
- `.kiro/steering/spec-authoring.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/quality-and-safety.md`

## For the chosen ones

The non-obvious property is that candidate content identity and administrative
scope must be bound **together but hashed separately**. If work IDs and
revisions enter the content digest, recording completion can stale its own
proof. If they disappear from the receipt, the proof can be copied to another
work item. The two-layer subject solves both: a portable content digest answers
“were the same decision-bearing inputs observed?”, while the outer receipt
answers “who accepted those inputs for which work, revision, policy and gate?”
That separation also makes idempotent replay honest without confusing replay
coordinates with engineering evidence.
