# Requirements Document: sdd-artifact-validation

## Introduction

This specification defines the typed artifact and validation interface used by
`sdd`. It owns structured Markdown/YAML parsing, profile-governed artifact
envelopes, allocation-once artifact identity, canonicalization, blob/semantic/
derivation digests, traceability, pure delta analysis, validator descriptors,
structured assessments and normalized policy inputs. It treats content and
agent claims as untrusted until a declared authority and validator contract
make them usable facts.

It does not own durable workspace writes, locks, journals, promotion or
conflict authorization, completion decisions, process supervision, Git
mutation, guided-journey policy, or remote policy distribution. Workspace
transactions apply authorized effects; completion decides whether evidence
satisfies a completion gate; guided workflows define specialized content
profiles; and an execution provider owns process lifetime and isolation.

This is a requirements document only. It deliberately leaves implementation
library choices and unaccepted trust/platform decisions as explicit blockers.

## Glossary and contract ownership

- **Artifact**: a typed, revisioned document with a common envelope and a
  profile-defined content extension.
- **Artifact identity**: the immutable allocation record for an artifact. It is
  not recomputed from edited content, a filename, a heading or a line number.
- **Artifact revision**: the monotonic revision of one artifact identity.
- **Source revision**: the workspace/work snapshot revision from which an
  artifact was read or produced.
- **Blob digest**: SHA-256 of exact source bytes, including line endings and a
  BOM when present.
- **Semantic digest**: SHA-256 of the versioned canonical typed representation.
- **Derivation digest**: SHA-256 of the ordered semantic input closure and the
  contracts/policy that derived a result.
- **Candidate identity**: the exact implementation snapshot supplied to a
  validator or completion consumer. It is distinct from an artifact digest.
- **Profile**: a versioned, trusted extension contract for a kind of artifact.
  Guided profiles define kinds such as `bug-report`, `diagnosis`,
  `regression-contract` and `verification`; this module validates their
  envelope and delegates journey-specific meaning to the profile owner.
- **Validator semantic contract**: the versioned behavior and applicability
  promise of a validator, separate from its binary/build provenance.
- **Receipt**: provider-produced evidence of one validator invocation and its
  result. A receipt is not completion and does not authorize a mutation.
- **Pure assessment**: a read-only evaluation returning findings/proposals;
  applying a proposal is a new workspace transaction.

## Common envelope

Every normalized artifact has this shape (profile content is typed, not an
unconstrained escape hatch). This is a computed in-memory/result projection,
not a requirement to place a file's own blob digest inside the bytes it hashes.
Raw authoring input supplies declared identity/content/profile fields; parsing
computes observations and digest fields. Supplied digest claims are checked
against computed values and never trusted as their own evidence:

```text
ArtifactEnvelope {
  schemaVersion: string,
  kind: profile-defined string,
  id: immutable allocated ArtifactID,
  workId: immutable WorkID,
  artifactRevision: non-negative integer allocated by workspace,
  sourceRevision: non-negative integer from the foundation workspace/work
    revision contract,
  status: draft | submitted | accepted | rejected | stale | archived | removed,
  blobDigest: sha256 digest of exact source bytes,
  semanticDigest: digest of canonical typed representation,
  derivationDigest: digest of declared semantic inputs (when derived),
  source: {
    path: workspace-relative canonical path,
    range: {path, start: {line, column, unit}, end?: {line, column, unit}},
    // Foundation units: utf8-byte | unicode-scalar; no competing rangeUnit field.
    provenance: typed producer/source observation
  },
  content: profile-defined typed value
}
```

`status` is a submission/lifecycle fact, not a claim that content is valid or
that a work item is complete. Parsing validity, structural validity,
submission validity, trusted acceptance, and completion eligibility are
separate result fields with separate diagnostics. A supplied `accepted` or
`passed` value is untrusted content unless an authorized validator/authority
produces the corresponding fact.

## Requirements

### Parsing and typed model

**AV-REQ-001 — Versioned structured parsing.** The Parser SHALL parse the
frontmatter and Markdown body into a typed AST under an explicitly selected
grammar version and profile version. It SHALL reject malformed UTF-8, duplicate
YAML keys, invalid envelope types, disallowed control characters, unknown
critical envelope fields and unsupported schema/profile versions with stable
diagnostics. A BOM and LF/CRLF are accepted as specified by the canonicalizer;
their raw bytes remain available for blob identity.

Acceptance criteria: AC-001 (valid envelope yields AST plus source ranges),
AC-002 (duplicate keys yield `VALIDATION_YAML_DUPLICATE_KEY`), AC-003
(malformed UTF-8/control characters fail before semantic validation), AC-004
(schema/profile version mismatch is explicit and non-retryable).

**AV-REQ-002 — Bounded YAML expansion.** Alias and anchor handling SHALL detect
cycles and enforce limits *before* unbounded materialization: input bytes,
tokens, nesting depth, anchor count, alias count, per-alias expansion, total
expanded nodes/bytes and aggregate document size. Limits are versioned profile
inputs with an enforced maximum; exhaustion yields
`VALIDATION_YAML_LIMIT` or `VALIDATION_YAML_ALIAS_CYCLE`. Aliased values are
resolved for type checking, but aliases do not create separate artifact
identity or silently bypass limits.

Acceptance criteria: AC-005 (bounded alias resolves to one typed value),
AC-006 (cyclic alias fails), AC-007 (exponential alias expansion fails before
allocation exceeds the configured budget), AC-008 (nesting and size exhaustion
are deterministic).

**AV-REQ-003 — Common envelope and profile extensions.** The Parser SHALL
require the common envelope above and validate `content` against the active
profile. Profile-defined kinds SHALL include, where enabled, the guided kinds
`bug-report`, `reproduction`, `investigation`, `diagnosis`,
`regression-contract`, `design`, `tasks`, `verification` and `policy-review`,
as well as ordinary requirements/scenarios/design/tasks/evidence. A profile may
add fields but cannot redefine envelope identity, digest or status semantics.
Source provenance is retained through every parse/normalize/validate result.

Acceptance criteria: AC-009 (guided envelope uses `schemaVersion`,
`artifactRevision`, `sourceRevision` and all three digest roles), AC-010
(unsupported kind/profile extension fails explicitly), AC-011 (source range
uses foundation units and is not inferred from a display filename), AC-012
(parsing, structural, submission and trusted-acceptance results are distinct).

**AV-REQ-004 — Status and revision semantics.** `artifactRevision` SHALL
be allocated by workspace for authoritative content, metadata, path or lifecycle
changes according to WS-REQ-003. Rename/move preserves identity, not necessarily
artifact revision; pure observation never allocates a revision. `sourceRevision`
binds the artifact to its observed workspace/work snapshot, with explicit scope
and owner identity alongside the integer so it cannot be compared across scopes.
Legal status transitions are profile/policy supplied
and at minimum support `draft -> submitted -> accepted` and
`submitted -> rejected|stale`; `archived` and `removed` preserve history.
This module reports transition validity; completion owns whether an accepted
artifact satisfies a gate.

Acceptance criteria: AC-013 (same ID edited under a new artifact revision),
AC-014 (a stale source revision is reported without being treated as accepted),
AC-015 (agent-supplied accepted status cannot create a trusted-acceptance fact).

### Stable identity and provenance

**AV-REQ-005 — Allocation-once stable IDs.** Each artifact kind has a
profile-registered namespace. An ID is allocated once in the workspace/work
scope and persisted as part of the authorized artifact-creation transaction
before subsequent edits. Draft parsing may propose a new ID without creating
an allocation record; a missing record for a previously persisted artifact is
an integrity error. The artifact envelope/tombstone can carry the allocation
record; a second identity database is not required. An
implementation MAY use a deterministic content-derived value only for that
initial allocation, but it MUST persist the allocation and MUST NOT recompute
identity when content changes. Human labels are display data, not identity.
IDs use a profile-registered syntax; the module does not require UUID-only IDs
where an accepted profile uses another collision-resistant syntax.

An allocated `ArtifactEnvelope` SHALL NOT be emitted for a draft proposal. A
read-only normalization result is a tagged `ProposedArtifact` with a proposed
ID and no artifact revision/allocation provenance, or an `AllocatedArtifact`
containing the complete allocated envelope. Revision zero MUST NOT encode the
proposed state.

Acceptance criteria: AC-016 (editing content preserves ID and changes semantic
digest), AC-017 (moving/renaming a file preserves ID), AC-018 (a duplicate ID
in one work scope fails even when source coordinates differ), AC-019 (a
missing/corrupt allocation record for persisted content fails rather than
silently reallocating; a new draft may propose an ID without writing state).

**AV-REQ-006 — Scope, collisions, tombstones and references.** ID uniqueness
is checked in the declared workspace/work namespace, not merely a current
revision or basename. Removal creates a tombstone retaining ID, work scope,
last revision and removal provenance. Cross-work/repository references carry
the target scope and resolved target identity; unavailable targets are
diagnostics, never arbitrary basename matches. A document ID is distinct from
requirement/scenario/task IDs contained in its content.

Acceptance criteria: AC-020 (same basename in two directories is allowed when
IDs and canonical source coordinates differ), AC-021 (ambiguous target fails
with `VALIDATION_ID_COLLISION` or `VALIDATION_TRACEABILITY_INVALID`), AC-022
(removed target is not silently treated as present), AC-023 (duplicate entity
ID and duplicate document ID are diagnosed independently).

**AV-REQ-007 — Provenance and trust separation.** Producer identity, source
observation, validator identity, build/binary digest and declared semantic
compatibility SHALL be recorded separately. Content may report an author's
claim, but only a trusted policy-selected authority can establish issuer
authenticity. This module SHALL NOT mandate cryptographic signatures for every
local assessment. Where a policy requires a signature, attestation or trusted
CI origin, the policy names the verifier, key/source, scope, expiry and
revocation behavior; otherwise integrity (digest) is not confused with issuer
authority.

Acceptance criteria: AC-024 (agent `passed`/`accepted` is provenance only),
AC-025 (forged validator name is rejected), AC-026 (unsigned local assessment
is handled according to declared policy rather than universally rejected),
AC-027 (expired or revoked authority cannot establish a current trusted fact).

### Canonicalization and digest roles

**AV-REQ-008 — Versioned conservative canonicalization.** The Canonicalizer
SHALL identify its schema/canonicalizer version in the semantic digest domain.
It SHALL preserve null versus absent, map ordering where order is meaningful,
ordered lists, normative wording, Markdown structure, code blocks, executable
examples and table content whenever the active profile treats them as meaning.
It SHALL NOT apply Unicode NFC or whitespace normalization universally. A
profile may define normalization for a particular field only when that rule is
explicit and semantics-preserving. Unknown fields are retained and included in
semantic identity by default; only explicitly declared bookkeeping fields may
be excluded.

Acceptance criteria: AC-028 (key reordering preserves identity only where keys
are declared unordered), AC-029 (normative wording/code/example/order changes
invalidate semantic identity), AC-030 (unknown potentially meaningful field
changes invalidate identity), AC-031 (a permitted BOM/line-ending cosmetic
change preserves semantic identity but changes blob identity), AC-032 (Unicode
or whitespace-sensitive literals are not rewritten by a global normalizer).

**AV-REQ-009 — Three digest roles and derivation closure.** Blob identity is
`SHA-256(exact source bytes)`; semantic identity is
`SHA-256(encode(schemaVersion, canonicalizerVersion, canonical typed value))`;
derivation identity is
`SHA-256(encode(workflow/profile identity, effective policy identity, ordered
input semantic digests, validator semantic-contract identities, explicit
external observations))`. `encode` is a versioned, domain-separated,
unambiguous encoding with explicit field boundaries, not raw string
concatenation. Derivation references include current semantic
digests, not IDs alone. Bookkeeping/self-digests and incidental evaluation
timestamps are excluded unless the profile declares them meaningful. A move
preserves blob/semantic digest only when bytes/meaning actually remain
unchanged; location-sensitive derivation may change.

Acceptance criteria: AC-033 (same ID with changed content changes semantic
digest), AC-034 (changing a referenced artifact's semantic digest changes the
derivation digest), AC-035 (self-reference/cyclic closure fails with
`VALIDATION_GRAPH_CYCLIC`), AC-036 (path move with unchanged bytes preserves
blob digest), AC-037 (a behaviorally meaningful code move is not assumed
irrelevant merely because content text is equal).

**AV-REQ-010 — Deterministic evaluation inputs.** Any evaluation that depends
on time, expiry, revocation, filesystem observation, network response or other
external fact SHALL receive that fact as an explicit versioned input. Pure
normalization and delta evaluation SHALL not consult a clock or network. The
same input snapshot, policy, profile and validator contracts produce the same
ordered result and diagnostics.

Acceptance criteria: AC-038 (fixed observation time yields reproducible output),
AC-039 (missing external fact is `unavailable`/blocking where policy requires,
not a silent default), AC-040 (same inputs yield byte-identical diagnostics).

### Traceability and pure deltas

**AV-REQ-011 — Typed traceability.** The module SHALL validate typed edges for
requirements→scenarios, requirements→design, design→tasks, tasks→code,
tasks→checks and tasks→evidence, with reverse coverage derivable. An edge
stores source/target ID and scope, relation, optional/waiver disposition and
derivation/provenance. A URL, path or code line is a locator, not proof of a
relationship. Structural completeness (edge exists) is separate from semantic
validity (target exists, status/profile permits the relation). Removal warnings
must distinguish a permitted cascade/waiver from an invalid reference.

Acceptance criteria: AC-041 (missing target emits source ID, target ID,
relation and source range), AC-042 (optional/waived edge is explicit), AC-043
(removed requirement referenced by task is a blocking invalid reference unless
an authorized disposition exists), AC-044 (human/agent semantic assessment is
not substituted for structural link validation).

**AV-REQ-012 — Artifact deltas and pinned bases.** The module SHALL parse and
validate pure `add`, `modify` and `remove` delta operations. A modify names a
pinned base artifact revision and base semantic/blob observations; a remove
retains a tombstone; every operation includes ID, kind, source/content and
expected identity as applicable. A three-way comparison consumes the pinned
base, proposed change and current target, and returns deterministic change and
conflict facts without writing files, resolving authorization, or promoting
content. “Previous revision” is insufficient unless it resolves to the pinned
base identity and digest.

Acceptance criteria: AC-045 (missing/mismatched base digest is a conflict
fact), AC-046 (same proposed delta is stable across runs), AC-047 (delta
evaluation does not mutate the workspace), AC-048 (conflicting base/current/
proposal changes are reported for completion rather than auto-resolved),
AC-049 (cross-work comparison carries explicit scope).

### Validators, pipeline and assessments

**AV-REQ-013 — Validator descriptors.** A trusted descriptor SHALL contain
stable validator ID, semantic-contract version/digest, implementation
provenance, accepted input/profile schemas, applicability, deterministic
ordering/dependencies, foundation severity (`info|warning|error|critical` pending
the explicit AM-01 reconciliation), trust tier
(`trusted|untrusted|heuristic`), required capabilities and structured output
schema. Descriptors are loaded from trusted configuration/policy, never from
artifact prose. A binary/build change may preserve freshness only when policy
and contract comparison establish semantic compatibility; a contract,
behavioral, security or trust change invalidates affected evidence.

Acceptance criteria: AC-050 (descriptor digest covers all descriptor fields),
AC-051 (unknown descriptor is rejected), AC-052 (heuristic result cannot meet
a mandatory proof gate), AC-053 (semantic contract change invalidates cached
evidence even when a self-declared patch is “compatible”).

**AV-REQ-014 — Validation pipeline results.** Validation SHALL expose separate
syntax, schema, graph, traceability, deterministic-policy, execution-backed and
heuristic stages. Each result is `pass`, `fail`, `error`, `skipped` or
`unavailable` with reason and provenance. Mandatory unavailable or
uninterpretable checks fail closed; timeout, cancellation, wrong snapshot and
truncated output are not `fail` or `pass`. Diagnostics retain stage ordering
and are not downgraded to warning to make a command appear successful.

Acceptance criteria: AC-054 (fatal syntax/schema errors short-circuit unsafe
later stages), AC-055 (mandatory unavailable check blocks), AC-056 (wrong
candidate or stale input is rejected), AC-057 (truncated/malformed validator
output is an error), AC-058 (read-only validation produces no workspace
mutation).

**AV-REQ-015 — Assessment and authority inputs.** Risk, diagnosis and
materiality assessments SHALL be typed records with evidence references,
uncertainty, author/issuer provenance, approval facts where supplied, and
explicit `insufficient`/`unassessed` states. The module validates structure,
provenance and evidence integrity; it does not claim that root-cause prose is
true. Confidence from a heuristic is not proof. Expiry/revocation is checked
against explicit current observations where the owning policy requires it.

Acceptance criteria: AC-059 (empty evidence produces explicit insufficient
state), AC-060 (heuristic confidence cannot establish a mandatory diagnosis),
AC-061 (assessment provenance is distinct from validator provenance), AC-062
(expired/revoked approval is not current authority).

**AV-REQ-016 — Validator invocation seam and inert content.** The validation
interface describes an invocation as `{validatorContract, artifactInputs,
implementationCandidate, policy, requestedLimits, environmentAllowlist}` and
returns a typed result/receipt reference. Artifact input identity and
implementation-candidate identity are separate. Commands in Markdown,
reproduction steps, examples and code blocks are inert data; only a separately
authorized typed invocation may call an approved provider, using an executable
plus an argv array that preserves spaces within one argument. The validator
module declares requested limits and interprets provider-reported enforcement;
the provider owns process groups, descendants, pipes, resource limits,
isolation and cancellation. No unrestricted shell runner is introduced here.

Receipts record validator contract ID/digest, implementation provenance,
artifact input digests, candidate identity, policy/profile/environment identity,
argv/cwd authorization, output digest, exit/signal/timeout/cancellation,
redaction outcome, explicit evaluation time and external observations. Fakes
or mocks are development checks only, not production proof.

Acceptance criteria: AC-063 (Markdown command is preserved but never executed),
AC-064 (argv `['tool','arg with spaces']` is not split), AC-065 (missing
provider capability yields unavailable/fail-closed), AC-066 (receipt binds the
implementation candidate rather than only the artifact digest), AC-067
(secret output is redacted before persistence), AC-068 (receipt records a
timeout/cancellation distinctly from validation failure). Pure assessment
never writes authoritative files. Execution-backed checks may use an explicitly
authorized isolated scratch area owned by their provider; any unexpected change
to authoritative candidate inputs makes the receipt invalid, not a clean pass.

### Policy, profiles and read-only limits

**AV-REQ-017 — Normalized policy inputs.** Policy input is a typed, normalized
artifact with schema/profile version, pinned identity/digest, precedence
inputs, required validators/capabilities, approval/exception facts, issuer,
issued/expiry times and revocation status. Organization authority, project
narrowing and local policy precedence are explicit inputs. Missing, expired,
revoked, unverifiable or stale required policy fails closed. This module
validates policy data; it does not distribute policy or invent organizational
identity.

Acceptance criteria: AC-069 (policy references to missing artifact/validator
fail), AC-070 (local policy cannot replace a required organization policy),
AC-071 (policy identity participates in derivation/cache identity), AC-072
(approval fact has explicit issuer/scope/expiry/revocation data).

**AV-REQ-018 — Profile contracts.** Profiles are versioned and identified by
name plus definition digest. Lightweight profiles may be dependency-free;
ordinary profiles support requirement/scenario/design/task/evidence
traceability; protocol profiles validate normative keywords, syntax,
state/failure semantics, security/privacy/operations and conformance
structure; guided profiles validate the common envelope and their typed
content. Local protocol linting SHALL NOT claim IETF publication,
certification or standards conformance beyond its declared rules.

A profile MAY enable a `ScopePriority` extension for requirements, scenarios,
tasks, or initiative slices. Where enabled, it SHALL use the closed values
`must|should|could|wont`, bind the subject to an explicit delivery-scope
baseline, retain rationale/evidence, and validate category-specific fields:
`should` requires a bounded workaround; `wont` requires an exclusion rationale
and reconsideration target; capacity/estimate fields are either unit-scoped or
absent. Artifact validation owns this structural/canonical check only. It
SHALL NOT infer a category from prose, choose a capacity cap, reorder workflow
dependencies, or decide whether a delivery scope may finish.

Acceptance criteria: AC-073 (same profile name with changed definition digest
requires explicit revalidation), AC-074 (old profile rejects an extension
introduced in a newer version), AC-075 (protocol profile checks its declared
deterministic rules beyond section presence; semantic judgments outside those
rules remain explicit reviewer assessments).

**AV-REQ-019 — Read-only bounded assessment.** Validation/convergence assessment
SHALL return findings, exact source locations, redacted evidence and bounded
remediation proposals without modifying artifacts, projections or external
systems. Traversal, reference count, depth, bytes and output size are bounded;
cycles and missing facts are explicit diagnostics. Applying a proposal is a
new plan through workspace transactions. Output ordering and redaction are
deterministic.

Acceptance criteria: AC-076 (assessment leaves bytes/modes/entries unchanged),
AC-077 (cyclic/excessive traversal exhausts a budget), AC-078 (missing
repository fact is not silently defaulted), AC-079 (same snapshot produces
same findings/source ranges).

**AV-REQ-020 — Cache and freshness identity.** A cached validation result is
eligible only when candidate identity, ordered artifact input semantic/blob
digests, source revision, policy/profile definition digests, validator semantic
contract and relevant implementation/environment/trust/expiry observations
match. A validator binary patch may reuse evidence only under the explicit
compatibility rule; a contract, policy, profile, candidate, input, revocation,
expiry or relevant environment change invalidates it. Cache lookup never
elevates a stale or untrusted result to accepted.

Acceptance criteria: AC-080 (candidate change invalidates cache), AC-081
(policy/profile/validator contract change invalidates dependent cache),
AC-082 (expiry/revocation observation invalidates authority-bound evidence),
AC-083 (compatible implementation patch is reusable only with recorded policy
proof of semantic preservation).

**AV-REQ-021 — Diagnostics and exit mapping.** Every diagnostic has a stable
`VALIDATION_<CATEGORY>_<DESCRIPTION>` code, foundation-compatible severity,
structured source/range, remediation, retryability and relevant provenance.
The public outcome uses foundation's exit taxonomy (`0`, `1`, `2`, `3`, `4`,
`5`, `6`, `130`); this module does not invent a competing exit-code set.
Diagnostic ordering is deterministic and machine consumers match codes and
metadata, never message text.

Acceptance criteria: AC-084 (malformed input maps to usage/schema category),
AC-085 (execution/provider failure maps to the external-tool category when
applicable), AC-086 (conflict/stale source maps to foundation conflict),
AC-087 (human and JSON diagnostics retain identical structured facts).

## Adversarial conformance fixtures

Fixture IDs are unique across this document. Each fixture specifies the
precondition, action, observable result, authoritative state change and
forbidden side effect; fixture text is a test contract, not evidence that a
test has already run.

| Fixture | Preconditions and action | Expected observable result / forbidden side effect |
|---|---|---|
| AV-FIX-001 | Parse duplicate YAML key | `VALIDATION_YAML_DUPLICATE_KEY`; no artifact allocated |
| AV-FIX-002 | Parse bounded alias and cyclic alias documents | bounded value succeeds; cycle fails before unbounded allocation |
| AV-FIX-003 | Parse malformed UTF-8, control character, oversized/deep input | deterministic parse/limit diagnostic; no partial AST accepted |
| AV-FIX-004 | Put a requirement-like heading in a fenced example | heading is code content, not a requirement edge |
| AV-FIX-005 | Parse same content with LF, CRLF and permitted BOM | semantic digest equal; blob digest differs where bytes differ |
| AV-FIX-006 | Change normative wording, code/example, meaningful ordering or unknown field | semantic digest changes |
| AV-FIX-007 | Allocate ID, edit content, then move/rename path | ID remains; revision/digest rules apply; no reallocation |
| AV-FIX-008 | Present duplicate entity/document IDs and ambiguous target basenames | explicit collision/traceability diagnostic; no arbitrary link |
| AV-FIX-009 | Reference removed requirement from a task | blocking invalid reference unless explicit authorized waiver/cascade |
| AV-FIX-010 | Move code or change a path with unchanged bytes | blob identity may remain; semantic/behavioral relevance is not inferred from location alone |
| AV-FIX-011 | Resolve unavailable external/cross-work target | unavailable diagnostic; no network retry hidden in pure evaluation |
| AV-FIX-012 | Construct cyclic derivation inputs | `VALIDATION_GRAPH_CYCLIC`; no derivation digest emitted |
| AV-FIX-013 | Submit agent `passed`/`accepted` claims | claims remain untrusted provenance; trusted result absent |
| AV-FIX-014 | Request forged/untrusted validator name | `VALIDATION_VALIDATOR_UNKNOWN`; no provider invocation |
| AV-FIX-015 | Return heuristic confidence as proof | heuristic result retained; mandatory gate remains unsatisfied |
| AV-FIX-016 | Put shell command/reproduction command in Markdown prose | content is inert; no process starts; authorized typed argv remains possible |
| AV-FIX-017 | Validator missing required provider, times out, is cancelled or emits truncated/malformed output | unavailable/error receipt; never pass; no orphaned process is owned by this module |
| AV-FIX-018 | Validator writes secret to stderr or requests disallowed environment | redacted receipt and environment-denied result; secret is absent from stored output |
| AV-FIX-019 | Supply receipt missing candidate/input/contract/output fields | malformed receipt diagnostic; no trusted evidence |
| AV-FIX-020 | Change implementation candidate after validation | stale candidate diagnostic; cache and evidence rejected |
| AV-FIX-021 | Upgrade profile under same name with new digest | stale/revalidation-required result; old extension rules remain in force |
| AV-FIX-022 | Compare compatible validator patch with semantic/security contract change | only policy-proven compatible patch may reuse; changed contract invalidates |
| AV-FIX-023 | Remove process adapter entirely | mandatory execution result unavailable/fails closed; no fake production pass |
| AV-FIX-024 | Evaluate delta against changed base/current target | pinned-base mismatch and three-way conflict facts; no file mutation or auto-promotion |
| AV-FIX-025 | Evaluate policy with expired/revoked/missing issuer | policy fails closed with issuer/expiry/revocation diagnostic |
| AV-FIX-026 | Run convergence assessment with missing facts, cycles and redaction candidates | bounded deterministic findings; no defaults, writes or leaked secrets |

## Acceptance-criterion traceability

The ranges in this table are inclusive closed ranges of the canonical IDs shown
(for example, `AC-001`–`AC-004` means exactly those four criterion IDs). Every
criterion in a range is asserted by the listed fixture set through the named
task's executable verification surface; a fixture may prove more than one
criterion but cannot silently retire one.

| Criteria | Fixtures | Primary task verification surface |
|---|---|---|
| AC-001–AC-004 | AV-FIX-001, AV-FIX-003, AV-FIX-004 | AV-TASK-002 |
| AC-005–AC-008 | AV-FIX-002, AV-FIX-003 | AV-TASK-002 |
| AC-009–AC-012 | AV-FIX-004, AV-FIX-007, AV-FIX-021 | AV-TASK-005 |
| AC-013–AC-015 | AV-FIX-007, AV-FIX-020 | AV-TASK-003 |
| AC-016–AC-019 | AV-FIX-007, AV-FIX-008 | AV-TASK-003 |
| AC-020–AC-023 | AV-FIX-008, AV-FIX-009, AV-FIX-011 | AV-TASK-006 |
| AC-024–AC-027 | AV-FIX-013, AV-FIX-014, AV-FIX-025 | AV-TASK-007 |
| AC-028–AC-032 | AV-FIX-005, AV-FIX-006 | AV-TASK-004 |
| AC-033–AC-037 | AV-FIX-005, AV-FIX-006, AV-FIX-010, AV-FIX-012 | AV-TASK-004 |
| AC-038–AC-040 | AV-FIX-011, AV-FIX-026 | AV-TASK-008 |
| AC-041–AC-044 | AV-FIX-008, AV-FIX-009, AV-FIX-011 | AV-TASK-006 |
| AC-045–AC-049 | AV-FIX-024 | AV-TASK-006 |
| AC-050–AC-053 | AV-FIX-014, AV-FIX-015, AV-FIX-022 | AV-TASK-007 |
| AC-054–AC-058 | AV-FIX-017, AV-FIX-019, AV-FIX-020 | AV-TASK-008 |
| AC-059–AC-062 | AV-FIX-013, AV-FIX-015, AV-FIX-025 | AV-TASK-007 |
| AC-063–AC-068 | AV-FIX-016, AV-FIX-017, AV-FIX-018, AV-FIX-019 | AV-TASK-009 |
| AC-069–AC-072 | AV-FIX-021, AV-FIX-025 | AV-TASK-007 |
| AC-073–AC-075 | AV-FIX-004, AV-FIX-021, AV-FIX-022 | AV-TASK-005 |
| AC-076–AC-079 | AV-FIX-026 | AV-TASK-011 |
| AC-080–AC-083 | AV-FIX-020, AV-FIX-021, AV-FIX-022 | AV-TASK-010 |
| AC-084–AC-087 | AV-FIX-001, AV-FIX-017, AV-FIX-020 | AV-TASK-012 |

## Coverage matrix and identifier migration

The following matrix maps every mandatory topic from handoff section 8 to this
document and its conformance fixtures. “Owned” means this spec defines the
input/result contract; application, promotion, and process effects remain with
their named owner.

| Handoff topic | Owning requirements | Fixtures |
|---|---|---|
| Parsing | AV-REQ-001, AV-REQ-002 | AV-FIX-001–AV-FIX-004 |
| Typed model/envelope | AV-REQ-003, AV-REQ-004 | AV-FIX-007, AV-FIX-019–AV-FIX-021 |
| Stable IDs | AV-REQ-005, AV-REQ-006 | AV-FIX-007–AV-FIX-009 |
| Canonicalization | AV-REQ-008 | AV-FIX-004–AV-FIX-006, AV-FIX-010 |
| Digest roles/closure | AV-REQ-009, AV-REQ-010 | AV-FIX-005, AV-FIX-006, AV-FIX-010–AV-FIX-012 |
| Provenance/implementation identity | AV-REQ-007, AV-REQ-020 | AV-FIX-013–AV-FIX-015, AV-FIX-020–AV-FIX-022 |
| Traceability | AV-REQ-011 | AV-FIX-008, AV-FIX-009, AV-FIX-011 |
| Artifact deltas | AV-REQ-012 | AV-FIX-024 |
| Validator descriptors | AV-REQ-013 | AV-FIX-014, AV-FIX-015, AV-FIX-021–AV-FIX-023 |
| Validation pipeline | AV-REQ-014 | AV-FIX-017–AV-FIX-020, AV-FIX-023 |
| Authority/assessments | AV-REQ-007, AV-REQ-015 | AV-FIX-013, AV-FIX-015, AV-FIX-025 |
| Execution seam | AV-REQ-016 | AV-FIX-016–AV-FIX-020, AV-FIX-023 |
| Policy data | AV-REQ-017 | AV-FIX-021, AV-FIX-025 |
| Profiles | AV-REQ-003, AV-REQ-018 | AV-FIX-004, AV-FIX-021, AV-FIX-022 |
| Read-only assessment/limits | AV-REQ-019 | AV-FIX-026 |

The old generated labels are mapped for review only; they are not stable IDs:

| Old label family | Replacement |
|---|---|
| `P01–P10`, `S01–S10`, `E01–E10` | AV-REQ-001/002/003/014/016 as applicable; use AC IDs, not old labels |
| `A01–A10` | AV-REQ-003/004/015/017 as applicable |
| `ID01–ID10` | AV-REQ-005/006 |
| `C01–C10` | AV-REQ-008 |
| `D01–D10` | AV-REQ-009/010 |
| `T01–T10` | AV-REQ-011 |
| `DELTA01–DELTA10` | AV-REQ-012 |
| `V01–V10` | AV-REQ-013 |
| `PRF01–PRF10` | AV-REQ-018 |
| `R01–R10`, `DC01–DC10`, `EPI01–EPI10`, `PU01–PU10`, `EDT01–EDT10`, `AF01–AF30` | AV-REQ-019/020/021 and AV-FIX-001–026; each old case must be migrated to a fixture or explicitly retired during review |

## Interfaces and ownership

The validation interface consumes immutable snapshot values and returns typed
facts. It never imports a second parser/digest implementation from guided
workflows, and it never writes authoritative state.

```text
Parse(bytes, GrammarProfile) -> ParsedArtifact | diagnostics
Normalize(ParsedArtifact, CanonicalizerProfile, IdentityObservation) -> NormalizationResult
Validate(snapshot, AllocatedArtifact, PolicySnapshot) -> ValidationResult
EvaluateDelta(base, proposal, current, CanonicalizerProfile) -> DeltaResult
DescribeValidator(TrustedConfig) -> ValidatorDescriptor
```

Workspace supplies source bytes, path/range observations and revisions;
completion supplies candidate-bound consumers and owns gate decisions;
workflow graph consumes validation facts for routing; guided workflows supply
profile definitions; the execution provider owns process lifecycle; context
indexing consumes canonical state as a projection. Any contract amendment
needed to foundation or guided workflows must be proposed there before design
or implementation, rather than being silently changed here.

## Explicit blockers and unresolved decisions

The following are acceptance blockers, not implementation assumptions:

1. Foundation/guided maintain the envelope names used above; any change to
   `schemaVersion`, source-range units, digest algorithm/domain separation or
   status vocabulary requires a versioned amendment and fixtures.
2. A trusted policy must select the concrete YAML/Markdown parsers and their
   exact grammar/limit behavior. No library is approved by this requirements
   file.
3. Policy owners must define which authenticity mechanisms (trusted CI,
   signature, attestation or local review) are required for each assessment or
   approval class. Local signatures are not assumed to be tamper-resistant.
4. The process/execution provider must define enforceable resource/isolation
   limits and the receipt transport before execution-backed checks can claim
   production proof.
5. Completion must define the implementation-candidate identity record and
   promotion/approval semantics consumed by AV-REQ-016 and AV-REQ-020.
6. The exact cache retention/redaction policy and validator semantic-contract
   compatibility matrix require owner approval.

## Non-goals

- durable file writes, journals, locks, recovery or transaction application;
- completion/DONE/archive/promotion authorization;
- process supervision, descendant cleanup or sandbox implementation;
- Git commit/merge/push or automatic conflict resolution;
- remote organization policy distribution or issuer provisioning;
- treating prose, agent claims, fake execution or a file's existence as proof.

## References

- `docs/kiro-input/next-prerequisite-specs.md`, sections 2–4 and 8
- `docs/cli-product-plan.md`, sections 5.3, 5.5, 5.7, 8.2–8.4, 10.2, 11.2, 11.4
- `.kiro/specs/sdd-engine-foundation/requirements.md`
- `.kiro/specs/sdd-guided-workflows/requirements.md`
- `.kiro/specs/sdd-guided-workflows/design.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/quality-and-safety.md`
