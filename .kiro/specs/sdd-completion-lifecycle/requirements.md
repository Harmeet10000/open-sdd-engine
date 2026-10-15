# Requirements Document: sdd-completion-lifecycle

## Introduction

This specification defines the completion owner for `sdd`. It evaluates whether
a work item has satisfied its profile and policy gates, accepts candidate-bound
evidence and approvals, renders the proof-bearing `done.md` and immutable
completion receipt, prepares exact promotion/lifecycle/archive effects, and
hands those effects to workspace transactions for recoverable application.

Completion owns the decision and proposal. It does not parse Markdown or
implement artifact digests, compile the workflow graph, supervise processes,
apply filesystem effects, or invent reviewer authority. Artifact-validation
supplies normalized artifacts, deltas and validation facts; workflow supplies
routing/composition facts; the execution provider supplies process evidence;
workspace supplies immutable snapshots, revision/transaction/idempotency and
recovery; the engine coordinates them through one public seam.

This document is a requirements contract only. It contains no implementation
tasks, no design commitment to a particular storage library, and no claim that
the fixtures have been executed.

## Glossary and ownership

- **Work item**: an immutable work identity with mutable administrative
  revision, lifecycle state and display slug.
- **Candidate**: the implementation read set evaluated for completion, including
  relevant tracked/untracked files, configuration, generated inputs,
  dependencies and execution environment facts.
- **Candidate identity**: a content-addressed, versioned description of the
  candidate read set and its observation context. It is not a Git HEAD and is
  not the work's administrative revision.
- **Administrative revision**: the workspace/work revision used for optimistic
  concurrency and lifecycle events. Recording a receipt advances this only as
  an administrative mutation; it does not change the candidate identity.
- **Evidence**: a typed result from an executed check, review, observation or
  approved manual/proxy procedure, bound to its subject and authority.
- **Gate**: a profile/policy-defined predicate over artifacts, evidence,
  approvals, candidate facts and authoritative observations.
- **Completion attempt**: one immutable evaluation with an attempt ID and
  idempotency/replay coordinates. A failed or cancelled attempt is not itself
  a terminal work state.
- **DONE**: the human-readable completion record generated from accepted facts.
  Its narrative cannot grant acceptance.
- **Completion receipt**: the machine-readable, immutable acceptance record
  produced by the completion owner after all gates pass.
- **Promotion**: applying an authorized semantic change from a pinned base to
  the current canonical accepted target. Artifact-validation computes delta
  facts; completion decides whether the promotion is acceptable; workspace
  applies the exact plan.
- **Trusted authority**: an issuer recognized by the active policy and bound to
  the relevant scope, candidate, gate and expiry/revocation observations.

## Contract boundaries

Completion consumes the shared artifact envelope from artifact-validation,
including `schemaVersion`, immutable artifact/work IDs, integer
`artifactRevision` and `sourceRevision`, status, source coordinates, and
blob/semantic/derivation digests. It does not redefine that envelope.

Completion consumes graph state and policy/definition digests from workflow,
but it alone decides whether completion gates, evidence and approvals are
satisfied. It emits plain typed effects and preconditions to workspace; it
does not acquire locks, write files, recover journals, or re-plan during apply.

## Candidate and receipt identity model

The candidate subject has at least:

```text
CandidateSubject {
  candidateId: immutable allocation for this observed candidate,
  candidateDigest: content identity of versioned read-set/environment/exclusion policy,
  workspaceId, workId, workRevisionAtObservation,
  readSet: ordered entries {
    repository-relative path or declared external input,
    kind: tracked | untracked | ignored | generated | dependency | submodule,
    byte/content digest, mode/link observation, inclusion reason
  },
  configurationDigest, generated-input-digest, dependency-lock-digest,
  toolchainDigest, environmentObservationDigest, gitObservation?,
  noGitManifest?, subjectSchemaVersion
}
```

The candidate digest excludes observation IDs, administrative revisions and
incidental timestamps; those are provenance and concurrency coordinates, not
content identity. Its versioned encoding is unambiguous and domain-separated.
The selected read-set includes membership/absence observations needed to detect
new relevant untracked files, not only digests of previously existing files.
Completion owns subject selection/identity rules; workspace supplies raw
coherent observations without interpreting completion policy.

Git observations are useful provenance only. A dirty tree, ignored file that a
policy declares relevant, submodule state, generated output, configuration,
lockfile or dependency change is part of the candidate when the declared
read-set policy includes it. A no-Git repository uses a deterministic manifest
of the selected workspace and explicit tool/environment observations; it never
fabricates a commit or tree reference.

DONE, receipts, journals, transaction metadata and other bookkeeping are
excluded from the candidate only by a policy-owned, versioned exclusion list
that is resolved before observation. An agent or candidate content cannot
modify that list to hide implementation changes. A bookkeeping file that is
also a declared implementation input is not excluded. The completion receipt
does not hash or require itself, and recording it cannot invalidate the
candidate solely by adding administrative output.

## Requirements

### Lifecycle and state

**CL-REQ-001 — Completion lifecycle.** The completion owner SHALL implement the
following logical states and distinguish work state from attempt/evidence
state:

| Work state | Meaning | Legal completion-owner transitions |
|---|---|---|
| `open` | Work exists but required authoring/implementation is incomplete | `open -> verifying` only when the graph says prerequisites are ready; `open -> abandoned` under explicit authority |
| `verifying` | Candidate-bound checks are planned/running or being reconciled | `verifying -> review` after required evidence; `verifying -> open` on failed/cancelled attempt; `verifying -> stale` on subject/policy drift |
| `review` | Evidence/approvals and gates are evaluated | `review -> finish-pending` when all gates pass; `review -> open|stale` when a gate fails or inputs drift |
| `finish-pending` | Exact completion/promotion/archive plan exists but is not committed | `finish-pending -> completed` only after workspace commit; `finish-pending -> stale|open` on precondition conflict; recovery may resume the prepared plan |
| `completed` | Accepted receipt and DONE are durably committed for one candidate | `completed -> reopened` only under an explicit policy-authorized reopen; `completed -> superseded` when a later work item supersedes it; `completed -> archived` under explicit archive authority |
| `reopened` | A new active attempt follows a historical completion | `reopened -> verifying` or `reopened -> abandoned`; historical receipt remains immutable |
| `abandoned` | This attempt was intentionally stopped without claiming delivery | `abandoned -> reopened` only if policy supports it and a new attempt is recorded; `abandoned -> archived` preserves its non-delivered disposition |
| `archived` | Closed record relocated to archive | No in-place mutation; correction uses erratum, reopen or superseding work according to policy |
| `superseded` | A later work item replaces the accepted scope | Historical record remains readable; `superseded -> archived` preserves supersession; no relabeling as incomplete or done |
| `stale` | Required candidate/policy/evidence/definition input changed | `stale -> verifying` only after a new candidate/evidence evaluation |

There is no direct `ready -> completed` transition. A failed check, cancelled
request, unavailable provider or rejected review is an attempt result, not an
automatic terminal work state. The completion owner SHALL preserve attempt and
failure history while returning one structured safe next action.

Acceptance criteria: CL-AC-001 (illegal direct completion is rejected),
CL-AC-002 (failed/cancelled attempt returns to an appropriate active state and
retains history), CL-AC-003 (reopen/supersede preserves the prior immutable
receipt), CL-AC-004 (archive is distinct from completion and deletion).

### Candidate identity and self-invalidating proof

**CL-REQ-002 — Candidate read-set identity.** Completion SHALL derive candidate
identity from the actual implementation read set and explicit environment
inputs, not from Git HEAD alone. The read set SHALL cover relevant tracked,
untracked, ignored, generated, configuration, dependency-lock, submodule and
external-input facts according to a pinned policy. For each included path the
candidate records canonical relative path, bytes/content digest, file-kind and
link/mode observation. It SHALL record toolchain/build/test environment
identity and whether the observation came from Git or a deterministic no-Git
manifest.

The observer SHALL report missing/unreadable paths, unstable observations,
submodule changes and unsupported platform facts as explicit diagnostics. A
dirty or no-Git candidate is valid when its complete policy-selected read set
is observable; it is not silently treated as clean.

Acceptance criteria: CL-AC-005 (dirty implementation file changes candidate
identity), CL-AC-006 (relevant untracked/ignored/generated/config/dependency
input is included), CL-AC-007 (submodule state is bound), CL-AC-008 (no-Git
manifest is deterministic and contains no fabricated commit), CL-AC-009
(unstable or missing required read-set observation blocks completion).

**CL-REQ-003 — Administrative revision and exclusion policy.** Candidate
identity SHALL be distinct from workspace/work administrative revision. The
completion evaluator SHALL bind evidence to both where policy requires, but
shall not treat a receipt, DONE file, journal event or transaction record as an
implementation change when it is in the policy-owned bookkeeping exclusion
set. Exclusions are resolved and digested before candidate evaluation, cannot
be supplied or changed by an agent to conceal a changed implementation file,
and are rechecked at apply time. An excluded path that becomes a declared
input is included and can stale evidence.

The completion digest graph SHALL be acyclic: the receipt body may reference
the canonical DONE body digest and evidence/approval digests, while DONE may
reference a receipt ID (not its final digest). A receipt's final digest is
computed over a canonical body that excludes a field containing that digest;
the receipt cannot require or authorize itself.

Acceptance criteria: CL-AC-010 (writing DONE/receipt advances administrative
state without staling the same candidate), CL-AC-011 (agent-modified exclusion
rules are rejected), CL-AC-012 (implementation edit hidden under a bookkeeping
path stales candidate), CL-AC-013 (receipt/DONE hash graph has no self-cycle).

### Verification, evidence and baseline

**CL-REQ-004 — Verification state and proof.** Each required check SHALL have a
typed state: `planned`, `not-run`, `running`, `passed`, `failed`,
`inconclusive`, `unavailable`, `stale` or `waived`. A checkbox, expected
result, command string, agent assertion or existence of an output file is not
execution evidence. `not-run`, `partial`, `cancelled`, `timed-out`,
`unavailable`, `inconclusive` and `unverifiable` cannot satisfy a gate that
requires a pass. A waiver is a separate policy-authorized fact and never an
implicit pass.

Completion SHALL consume validation/provider receipts rather than reimplement
process supervision. It SHALL reject evidence with a missing subject, wrong
candidate, wrong artifact input, missing provider result, truncated output or
uninterpretable status.

Acceptance criteria: CL-AC-014 (all tasks checked with no execution evidence
blocks), CL-AC-015 (fake pass/expected result does not satisfy verification),
CL-AC-016 (partial/timeout/cancelled/unavailable evidence remains non-pass),
CL-AC-017 (waived gate names scope, authority and expiry).

**CL-REQ-005 — Baseline and regression evidence.** A regression contract SHALL
distinguish failing-before expectation, passing-after observation, oracle,
negative/safety cases and the candidate/environment used for each. A known
failing baseline records its failure fingerprint and proves that the result
candidate adds no new matching failure; it does not convert the whole command
to `passed`. Flaky checks record attempts, pass rate and policy threshold;
retries alone are not a clean pass. Intermittent, environment-specific,
not-reproduced and unsafe reproduction states remain explicit. An approved
proxy/manual/incident packet records its limitations and authority; inability
to reproduce is not a passing regression check.

Acceptance criteria: CL-AC-018 (failing-before plus passing-after bind to
distinct, correct candidates), CL-AC-019 (known baseline is not relabeled pass),
CL-AC-020 (flaky threshold is policy-driven), CL-AC-021 (unsafe reproduction
requires an approved bounded alternative), CL-AC-022 (not-reproduced does not
pass the regression gate).

**CL-REQ-006 — Evidence authority and freshness.** Every evidence item SHALL
bind the check to work/action scope, candidate identity, artifact input
digests, command/method and authorized argv/cwd/environment identity, validator
semantic-contract and implementation provenance, result/output digests,
provider receipt, execution time, and explicit external observations. The
active policy SHALL define issuer trust, expiry and revocation observations.
Evidence is a discriminated record: executed checks carry executable/argv and
provider fields; manual/proxy/review evidence carries the approved method,
actor, steps, observations and limitations. Do not fabricate an exit code or
provider invocation for manual work. A manual/proxy disposition can satisfy
only a gate whose trusted profile explicitly permits that evidence class.
Locally editable files alone are not tamper-resistant authority. Independently
verifiable CI or external receipts identify their issuer, subject, policy and
verification route; unavailable/redacted evidence is not silently accepted.

Replay of evidence is allowed only when all bound identities and validity
observations still match. Completion SHALL preserve redaction metadata and
must not store secrets merely to make a receipt auditable.

Acceptance criteria: CL-AC-023 (evidence from another candidate/environment is
rejected), CL-AC-024 (CI receipt without authenticatable issuer is unavailable),
CL-AC-025 (expired/revoked evidence is stale), CL-AC-026 (redacted output
retains interpretable result and provenance), CL-AC-027 (evidence replay with
matching identities is deterministic).

### Approvals, waivers and deterministic gates

**CL-REQ-007 — Approval and waiver authority.** A reviewer decision SHALL bind
to reviewer identity/issuer, work and candidate identity, artifact/spec
revision and digest set, policy/definition digest, gate scope, decision,
issued/expiry times, revocation observation, and the approval evidence. A
display name, role string, `approved: true`, agent claim or lease is not
reviewer authority. Independence rules are policy inputs; when required, the
author/implementer cannot satisfy the independent reviewer gate. Emergency
waivers name exact diagnostics/scope, rationale, authority, expiry and
retrospective gates. Policy may waive only explicitly eligible gate classes;
candidate binding, receipt integrity and authentic waiver authority are
non-waivable integrity invariants. A waiver never converts an untrusted claim
into an authentic issuer or a not-run check into an executed pass.

Acceptance criteria: CL-AC-028 (forged reviewer identity is rejected),
CL-AC-029 (self-approval is rejected where independence is required),
CL-AC-030 (approval for another candidate/policy/gate scope is rejected),
CL-AC-031 (expired/revoked approval blocks), CL-AC-032 (emergency waiver
requires bounded retrospective obligations).

**CL-REQ-008 — Deterministic completion gates.** Completion SHALL evaluate all
independent required gates in deterministic profile order to produce actionable
diagnostics; it may stop an unsafe effect plan immediately, but it SHALL not
hide other known blockers. Unknown, missing or stale evidence never passes by
default. The evaluator SHALL recheck candidate, policy, approval, evidence,
definition, baseline and work revision facts against the evaluation snapshot,
then provide workspace with exact apply preconditions. A read-only blocked
observation may succeed as a command while the completion decision remains
blocked; the public outcome uses foundation diagnostics and exit categories.

At minimum the built-in gate set covers valid/fresh required artifacts and
tasks, requirement/scenario implementation and verification coverage, changed
path mapping, same-candidate verification, required test/lint/type/build/
migration/performance/security checks, project guidelines, reconciled design
discoveries, migration/rollback evidence, independent review, no blocking
diagnostics/expired claims/prepared transactions, conflict-free promotion,
necessity of new dependencies/interfaces/configuration/processes, and all
requested verification steps having snapshot-bound interpretable results.

When the trusted profile supplies a MoSCoW scope baseline, completion SHALL
also validate its pinned identity and policy digest. Every applicable `must`
subject must be satisfied against that exact baseline. A `should` or `could`
may be absent only with its explicit deferred disposition and required
workaround/follow-up; a `wont` is an excluded-scope record, not delivered scope
or an automatically failing task. Capacity erosion and a proposed Must change
return `scope-replan-required` until the authorized scope owner commits a new
baseline. A feature flag is rollout evidence only and cannot satisfy an
unimplemented Must.

Acceptance criteria: CL-AC-033 (all blockers are deterministic and actionable),
CL-AC-034 (missing/unknown evidence cannot pass), CL-AC-035 (changed candidate
between evaluation and apply is rejected), CL-AC-036 (all required gates are
reported without unsafe application), CL-AC-037 (read-only blocked status does
not falsely claim completion), CL-AC-080 (incomplete Must blocks completion),
CL-AC-081 (deferred Should/Could and Won't-have records remain visible and do
not masquerade as completed scope).

### DONE, references and receipt

**CL-REQ-009 — DONE content.** Completion SHALL generate `done.md` from
accepted facts and the trusted profile. Required sections are tailored by
profile but cover, where applicable: delivered and explicitly undelivered
scope; high-level summary; requirements/scenario coverage; design decisions;
repository-relative code references; project-guideline references; planned and
actual execution order; verification evidence; deviations, waivers, residual
risk and follow-ups; security/privacy/migration/operations; and exact
acceptance/reproduction instructions. The generated `## DONE` section SHALL
show gate/status/evidence rows, acceptance authority, and a separate MoSCoW
scope summary when a scope baseline applies: Must delivered/blocking, Should
and Could delivered/deferred with their follow-ups, and Won't-have excluded
scope. An agent may draft
narrative inputs but cannot write its own acceptance stamp or turn a planned
result into a pass. No `--force-done` exists.

Documentation-only, research/spike, generated-output, feature-flag and manual
verification profiles may tailor required sections, but a tailored omission is
an explicit profile disposition, not filler or an untracked bypass.

Acceptance criteria: CL-AC-038 (DONE includes delivered versus deferred scope),
CL-AC-039 (planned and actual order expose deviations/parallel groups),
CL-AC-040 (gate table and receipt derive from the same accepted gate facts), CL-AC-041 (profile-specific
not-applicable sections are explicit), CL-AC-042 (agent narrative cannot stamp
acceptance), CL-AC-082 (DONE preserves MoSCoW category and disposition without
claiming excluded/deferred work was delivered).

**CL-REQ-010 — Durable references.** Code references SHALL use repository-
relative canonical path, symbol/heading/test/schema pointer when available,
candidate or immutable revision/content digest, owning task/requirement and an
optional line range. Line numbers are locator conveniences, not identity.
Guideline references SHALL identify source path, section/anchor, content
semantic digest/version and applicable clause, with disposition `followed`,
`not-applicable` or `exception-approved`. Completion SHALL report stale or
unresolvable code/guideline references and cannot pretend a line-only reference
survived an edit.

Acceptance criteria: CL-AC-043 (historical code reference resolves against its
pinned revision and symbol; changed current bytes still trigger freshness
checks), CL-AC-044 (stale code reference blocks when required),
CL-AC-045 (guideline version/clause and disposition are recorded), CL-AC-046
(missing guideline source is explicit, not silently accepted).

**CL-REQ-011 — Acyclic immutable completion receipt.** The receipt SHALL use a
versioned schema and include immutable receipt/attempt IDs, work ID and
administrative revision, candidate subject identity, result snapshot,
accepted artifact/spec digest set, policy/profile/definition digests, gate
results, evidence and approval references, waiver references, issuer, event/
transaction linkage, and timestamps as metadata. It SHALL include the
canonical DONE body digest and may include a completion digest calculated over
the receipt body excluding the completion-digest field itself. The receipt
shall never contain its own final digest as an input and shall never be a
prerequisite for its own acceptance. It links to a preallocated transaction/plan
ID, not the final digest of a full plan containing the receipt's own bytes.
The workspace commit summary binds that ID, the final plan digest and receipt
digest after those values are computable. Historical receipts are immutable;
correction is an erratum, reopen or superseding work item.

Acceptance criteria: CL-AC-047 (receipt has all subject/policy/evidence bindings),
CL-AC-048 (receipt/DONE digest dependencies are acyclic), CL-AC-049 (editing
accepted DONE or completion inputs makes current completion stale), CL-AC-050
(historical receipt bytes and references are not rewritten).

### Exact plan, promotion, archive and retry

**CL-REQ-012 — Atomic completion proposal.** After gates pass, completion SHALL
produce one exact proposal containing DONE, receipt, lifecycle, promotion and
archive effects (where selected), with ordered operations, exact new bytes or
content digests, target identities, pinned candidate/policy/definition
observations, expected administrative revisions, all read/write preconditions,
authorization facts, plan digest and idempotency coordinates. Workspace owns
authorization/application, durable commit intent and recovery. Completion SHALL
not write any effect during evaluation, silently re-plan at apply, or claim
completion before the transaction commits.

If candidate, policy, baseline, definition, target or work revision changes
between evaluation and apply, workspace rejects the exact plan with a conflict/
stale result. A new authorized mutation may record a new attempt; read-only
evaluation does not write one. Engine readers never accept a partial
DONE/receipt/archive/promotion set as completed. Physical files may be
intermediate after PREPARED; workspace exposes recovery-required until the
recorded intent is committed or quarantined, not a fictitious cross-file atomic
filesystem write.

Acceptance criteria: CL-AC-051 (pure evaluation leaves authoritative files
unchanged), CL-AC-052 (plan contains complete preconditions and digest),
CL-AC-053 (pre-apply drift rejects exact plan), CL-AC-054 (readers accept only
the coherent committed result or report recovery-required), CL-AC-055 (completion is not
reported before durable commit).

**CL-REQ-013 — Three-way promotion.** Promotion SHALL compare a pinned base
canonical target, the proposed semantic change and the current canonical target
using stable artifact identity and artifact-validation normalized delta facts.
It SHALL distinguish clean/disjoint merges, same-entity conflicting changes,
delete-versus-edit, rename/path changes, ID collisions, changed base and
changes made by another work item. It SHALL never overwrite another accepted
change or resolve conflicts by ours/theirs/default ordering. A reviewer
resolution is a new proposal with a new base/current observation, policy and
approval; promotion authorization is not delegated to the parser.

Acceptance criteria: CL-AC-056 (disjoint changes merge only with preserved
unrelated content), CL-AC-057 (same-requirement conflict blocks), CL-AC-058
(delete/modify conflict blocks), CL-AC-059 (rename/ID collision blocks),
CL-AC-060 (resolved conflict requires a fresh authorized plan).

**CL-REQ-014 — Archive and historical discovery.** Archive SHALL preserve work
ID, artifact IDs, candidate/provenance/evidence/approval references, receipt
and lifecycle history. It SHALL distinguish `completed`, `archived`,
`abandoned` and `superseded`; moving a record is not permission to delete
evidence, erase a branch, or change accepted bytes. Destination collisions,
existing archive identities and incompatible archive format are explicit
conflicts. Status/discovery can find completed work after relocation through
immutable IDs and archive metadata. Retention/deletion is a separate,
policy-authorized operation and cannot be smuggled into finish.

Acceptance criteria: CL-AC-061 (archive collision does not overwrite),
CL-AC-062 (archived receipt/evidence remains addressable), CL-AC-063 (archive
does not delete Git branches or evidence), CL-AC-064 (retention/deletion is not
implied by archive).

**CL-REQ-015 — Idempotent finish, crash recovery and partial scope.** Finish
SHALL use workspace idempotency semantics: same principal/workspace/operation
namespace and same normalized request fingerprint returns the committed
receipt/result; the same key with changed candidate, policy, plan or content is
a conflict. A lost response, uncertain commit or crash returns recovery state
and an inspectable prepared transaction; retry distinguishes committed replay
from a new attempt and never guesses success. Concurrent duplicate finish
requests converge through the workspace transaction contract.

No-change and documentation-only completion are valid only when the trusted
profile defines the applicable gates. Partially delivered scope records actual
delivery and creates a traceable follow-up work ID or superseding disposition;
it cannot leave an unchecked task under a DONE stamp. Deferred requirements are
removed/superseded through an authorized revision, not silently ignored.

Acceptance criteria: CL-AC-065 (same-key retry returns original receipt),
CL-AC-066 (same-key changed content conflicts), CL-AC-067 (lost response after
commit replays committed result), CL-AC-068 (crash at each effect boundary is
recoverable without partial authoritative completion), CL-AC-069 (concurrent
duplicate finish does not duplicate receipts/archive), CL-AC-070 (partial or
deferred scope is explicit and not marked done).

### Post-verification changes, privacy and observability

**CL-REQ-016 — Post-verification edits and reopening.** Any relevant code,
configuration, generated input, dependency, accepted artifact, policy,
workflow/validator contract, baseline, guideline or required evidence change
after verification SHALL stale affected evidence/approval and invalidate a
pending completion proposal. An edit is irrelevant only when the declared
digest graph proves it is outside the candidate/gate closure. Post-completion
edits do not rewrite the historical receipt; they produce a derived current
staleness finding and require new verification or an explicit policy-authorized
reopen. Read-only status does not persist a lifecycle transition. A new plan
records any authorized reopen/supersession separately from historical truth.
Reopen preserves the prior receipt and candidate as historical evidence while
creating a new administrative attempt; supersession creates a traceable later
work identity when policy requires it.

Acceptance criteria: CL-AC-071 (one-line relevant code edit stales evidence),
CL-AC-072 (irrelevant excluded bookkeeping write does not stale candidate),
CL-AC-073 (post-completion edit preserves historical receipt and reports current
staleness without read-only writes), CL-AC-074 (reopen creates a new attempt without rewriting the
old one), CL-AC-075 (changed guideline reference is detected).

**CL-REQ-017 — Privacy, redaction and observable outcomes.** Completion SHALL
minimize stored evidence, redact secrets before persistence/rendering, retain
only policy-permitted output excerpts or digests, and record retention class
and reference availability. It SHALL distinguish auditability from storing
entire logs. Human and JSON outcomes SHALL report the same blocked gate,
affected candidate/evidence/subject and one safe next action; human text is not
the authority. Missing/expired/redacted/unavailable references are explicit
facts with policy-driven behavior, never silent success.

Acceptance criteria: CL-AC-076 (secret is absent from stored/rendered receipt),
CL-AC-077 (minimal evidence remains independently traceable), CL-AC-078
(missing retained reference is a deterministic blocker), CL-AC-079 (human/JSON
outcomes preserve identical structured blocker facts).

## Lifecycle and authority rules

The following table is normative for completion-specific authority. A physical
actor may perform multiple roles only when policy allows and the receipt records
each role; “agent” is never an authority by declaration.

| Decision/effect | Required authority | Completion obligation | Owner of effect |
|---|---|---|---|
| Candidate observation | workspace snapshotter plus policy-selected observer | Bind read set, environment and observation revision | Workspace/observer |
| Verification result | declared execution provider, trusted CI or policy-approved manual actor | Check subject, validator contract, output/result and freshness | Provider/validation |
| Gate evaluation | completion evaluator using pinned profile/policy | Evaluate all required gates deterministically | Completion |
| Review/approval | issuer recognized by policy, independent where required | Bind reviewer, candidate, gate scope, expiry/revocation | Completion consumes fact |
| Waiver/exception | policy-authorized actor and exact scope | Require expiry and retrospective obligations | Completion consumes fact |
| Promotion authorization | policy-selected reviewer/authority | Resolve three-way facts in a new exact plan | Completion proposes; workspace applies |
| Files/journal/lifecycle/archive write | workspace transaction authority | Supply exact effects and preconditions; never write directly | Workspace |
| Process limits/isolation | execution provider | Consume requested limits and report enforcement | Provider |

Authority authenticity mechanisms (trusted CI, signature, attestation, local
review or another declared source) are policy inputs. This spec does not
mandate a universal signature scheme or accept a self-asserted role string.

## Completion receipt and DONE dependency graph

The following dependency direction prevents proof from invalidating itself:

```text
candidate/read-set + artifacts + policy + definitions + evidence + approvals
       -> gate results and promotion facts
       -> canonical DONE body -> doneDocumentDigest
       -> canonical receipt body (contains doneDocumentDigest)
       -> receiptDigest (computed outside the body)
       -> workspace completion plan / event
```

`done.md` may contain the stable `receiptId` and a completion digest over its
own canonical body; it MUST NOT require the final receipt digest as an input to
that same body. The receipt may reference `doneDocumentDigest` and evidence
digests but not its own final digest. Neither output is accepted merely because
it exists on disk.

## Public result and diagnostic obligations

Completion extends the foundation envelope through the reviewed guided
`completion` result kind. It uses foundation source units and exit categories:
`0` for successful observation (including a read-only blocked report), `1` for
validation/verification failure, `2` malformed input, `3` workspace/config/
schema error, `4` stale/conflict/recovery, `5` required external provider
failure, `6` policy/permission denial and `130` interruption after known
cleanup. It does not create a competing command/result envelope or exit code.

Completion diagnostics SHALL identify the gate/attempt, work and candidate
subject, expected/current revision or digest, policy/authority source,
retryability, remediation and safe next action. Examples include stable
completion-registered codes such as `COMPLETION_GATE_BLOCKED`,
`COMPLETION_CANDIDATE_STALE`, `COMPLETION_EVIDENCE_UNAVAILABLE`,
`COMPLETION_APPROVAL_EXPIRED`, `COMPLETION_PROMOTION_CONFLICT`,
`COMPLETION_RECOVERY_REQUIRED` and `COMPLETION_SCOPE_DEFERRED`; exact catalog
registration remains a foundation/guided schema amendment blocker.

## Adversarial conformance fixtures

Fixtures are contracts for black-box tests; none is claimed executed here.
Every fixture names the observable outcome and forbidden side effect.

| Fixture | Preconditions and action | Expected observable outcome / forbidden side effect |
|---|---|---|
| CL-FIX-001 | Mark all tasks checked but provide no execution evidence | gate blocked; no DONE, receipt or completed state |
| CL-FIX-002 | Submit fake pass, forged reviewer or forbidden self-approval | authority/evidence rejection; no gate satisfaction |
| CL-FIX-003 | Omit one mandatory gate | missing-gate blocker names gate; no implicit pass |
| CL-FIX-004 | Edit dirty/untracked implementation after verification | candidate digest changes; affected evidence/approval stale |
| CL-FIX-005 | Re-submit candidate with agent-modified exclusion rules | policy/exclusion mismatch blocks; implementation file cannot be hidden |
| CL-FIX-006 | Write administrative receipt/DONE after evaluation | admin revision may advance; candidate does not self-stale solely from bookkeeping |
| CL-FIX-007 | Revoke policy before finish apply | exact plan rejected with policy/revision conflict; no partial outputs |
| CL-FIX-008 | Verify another environment/candidate | candidate-bound evidence rejected as stale/wrong subject |
| CL-FIX-009 | Present unauthenticated CI receipt, unavailable/redacted output or no provider | unavailable/fail-closed result; never pass |
| CL-FIX-010 | Finish a no-Git repository | deterministic manifest candidate accepted only if policy gates pass; no fake commit |
| CL-FIX-011 | Finish docs-only work | profile decides code gates not applicable; DONE records scope and rationale |
| CL-FIX-012 | Use flaky test, unsafe reproduction or approved proxy | explicit attempts/limits/authority; no automatic pass |
| CL-FIX-013 | Use expired emergency waiver | waiver rejected; retrospective gate remains blocking |
| CL-FIX-014 | Concurrent promotions from same base with disjoint changes | disjoint merge preserves both changes under fresh authorized plan |
| CL-FIX-015 | Concurrent same-requirement, delete/modify or rename/ID collision | three-way conflict; no ours/theirs overwrite |
| CL-FIX-016 | Archive when destination already exists | archive conflict; existing record untouched |
| CL-FIX-017 | Crash during each completion effect | workspace recovery yields all-old, all-new or inspectable prepared state; no guessed success |
| CL-FIX-018 | Lose response after commit then retry finish | committed receipt replayed; no duplicate receipt/archive |
| CL-FIX-019 | Concurrent duplicate finish requests | one committed result and authorized replay, or explicit conflict; no duplicate effects |
| CL-FIX-020 | Retry after evidence expiry | expired attempt is not replayed as pass; committed prior result is distinguished from new attempt |
| CL-FIX-021 | Edit code after completion | historical receipt immutable; current work stale/open and requires new evidence |
| CL-FIX-022 | Edit/move guideline or line-only code reference | stale reference diagnostic; digest/symbol identity required |
| CL-FIX-023 | Reopen completed work | prior receipt retained; new attempt/candidate created; no history rewrite |
| CL-FIX-024 | Leave an incomplete Must, or a deferred Should/Could/Won't record without its required scope disposition, under DONE | incomplete Must blocks; deferred/excluded scope requires an explicit baseline-linked follow-up/supersession/disposition and remains visible |
| CL-FIX-025 | Run manual verification with environment and screenshots | accepted only when policy permits and actor/steps/evidence are bound |
| CL-FIX-026 | Change candidate/policy/base between evaluation and apply | exact plan conflicts; no silent re-plan or partial completion |

## Coverage matrix

| Handoff section 9 topic | Requirements | Fixtures |
|---|---|---|
| Lifecycle states/transitions | CL-REQ-001 | CL-FIX-001, CL-FIX-016 through CL-FIX-024 |
| Verification subject/read set | CL-REQ-002 | CL-FIX-004, CL-FIX-008, CL-FIX-010 through CL-FIX-013 |
| No self-invalidating proof | CL-REQ-003 | CL-FIX-005, CL-FIX-006, CL-FIX-017, CL-FIX-021 |
| Verification states | CL-REQ-004 | CL-FIX-001 through CL-FIX-003, CL-FIX-009, CL-FIX-020 |
| Baseline/regression evidence | CL-REQ-005 | CL-FIX-011 through CL-FIX-013, CL-FIX-025 |
| Evidence authority | CL-REQ-006 | CL-FIX-002, CL-FIX-008 through CL-FIX-010, CL-FIX-017 through CL-FIX-020 |
| Approvals/waivers | CL-REQ-007 | CL-FIX-002, CL-FIX-003, CL-FIX-007, CL-FIX-013 |
| Deterministic gate evaluation | CL-REQ-008 | CL-FIX-001 through CL-FIX-003, CL-FIX-007, CL-FIX-024, CL-FIX-026 |
| DONE content | CL-REQ-009 | CL-FIX-011, CL-FIX-012, CL-FIX-024 |
| Durable code/guideline references | CL-REQ-010 | CL-FIX-021, CL-FIX-022 |
| Completion receipt/no hash cycle | CL-REQ-011 | CL-FIX-006, CL-FIX-017 through CL-FIX-021 |
| Atomic completion proposal | CL-REQ-012 | CL-FIX-007, CL-FIX-017, CL-FIX-026 |
| Three-way promotion | CL-REQ-013 | CL-FIX-014, CL-FIX-015, CL-FIX-026 |
| Archive | CL-REQ-014 | CL-FIX-016, CL-FIX-018, CL-FIX-023 |
| Retry/crash/partial scope | CL-REQ-015 | CL-FIX-017 through CL-FIX-020, CL-FIX-024 |
| Post-verification changes/reopen | CL-REQ-016 | CL-FIX-004, CL-FIX-021 through CL-FIX-023 |
| Privacy/observability | CL-REQ-017 | CL-FIX-009, CL-FIX-022, CL-FIX-025, CL-FIX-026 |

## Explicit ownership exclusions

- Artifact parsing, common envelope, artifact IDs and blob/semantic/derivation
  computation belong to artifact-validation.
- Generic graph compilation, dependency routing and next-action evaluation
  belong to workflow. Execution claims/leases belong to bounded-execution;
  workflow and completion only consume their declared facts.
- Raw bytes, snapshot acquisition, administrative/work revisions, locks,
  idempotency records, exact effects, journals and crash recovery belong to
  workspace transactions.
- Process groups, descendant cleanup, sandboxing, resource enforcement and
  provider execution belong to the declared execution provider.
- Guided workflows own Bug Fix/Quick Spec journey definitions and specialized
  profile rules; completion consumes their typed facts.

## Acceptance blockers and unresolved decisions

These are explicit release/design blockers, not assumptions hidden in this
requirements document:

1. Completion must finalize the subject/read-set selection and no-Git manifest
   schema, supported submodule/ignored/generated-file rules and policy-owned
   exclusions. Workspace must supply coherent raw observations satisfying that
   read request; it does not own candidate semantics.
2. Foundation/guided must allocate the completion result schema/version and
   exact diagnostic catalog, including severity/blocking metadata without
   losing foundation's `critical` semantics.
3. Policy owners must define trusted reviewer/CI/issuer sources, independence
   rules, authenticity mechanism, retention and revocation/expiry observations.
4. The execution provider must define the receipt transport, resource/isolation
   guarantees and what evidence can be independently verified after redaction.
5. Artifact-validation must finalize the delta result fields and canonical
   three-way comparison contract consumed by CL-REQ-013.
6. Workflow must supply versioned composition/handoff facts; guided profiles
   supply journey-specific gates. Completion owns completion lifecycle and
   deferred-scope acceptance rules, not a second graph engine.
7. Workspace must ratify the exact idempotency namespace/fingerprint/replay
   disclosure and prepared-transaction recovery states used by CL-REQ-012 and
   CL-REQ-015.
8. The project must decide retention/deletion policy separately from archive;
   no completion contract assumes evidence may be deleted.

## Non-goals

- parsing or canonicalizing artifacts;
- compiling or routing the workflow graph;
- process execution, supervision, sandboxing or tool installation;
- direct file/Git/database writes, locks, journals or transaction recovery;
- automatic Git commit, branch, merge, PR, push or ours/theirs resolution;
- universal cryptographic signatures, remote issuer provisioning or policy
  distribution;
- treating checkboxes, prose, file existence, agent claims or fake execution
  as completion proof.

## References

- `docs/kiro-input/next-prerequisite-specs.md`, sections 2–4 and 9
- `docs/kiro-input/prerequisite-review.md`, IC-01 through IC-12
- `docs/cli-product-plan.md`, sections 5.1–5.5, 6.8–6.11 and 8.2
- `.kiro/specs/sdd-engine-foundation/requirements.md`
- `.kiro/specs/sdd-guided-workflows/requirements.md`
- `.kiro/specs/sdd-guided-workflows/design.md`
- `.kiro/specs/sdd-workspace-transactions/requirements.md`
- `.kiro/specs/sdd-workflow-graph/requirements.md`
- `.kiro/specs/sdd-artifact-validation/requirements.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/quality-and-safety.md`
