# Design Document: sdd-artifact-validation

## Draft status and scope

This is a provisional design derived from the reviewed requirements. It is not
an approval, implementation record, or schema allocation. In particular,
`sdd-completion-lifecycle/requirements.md` is still under root review, so the
`CandidateSubject` consumed here is a provisional external value contract. The
design can be implemented behind that seam, but production evidence acceptance
cannot be claimed until completion ratifies the subject schema.

Artifact-validation owns two deep modules:

- `artifact`: bounded structured parsing, typed profile projection, stable
  artifact identity checks, canonicalization, digests, traceability values and
  pure three-way delta facts;
- `validation`: trusted descriptor and policy normalization, deterministic
  validation stages, typed assessments, provider invocation interpretation,
  cache eligibility and read-only convergence findings.

The modules never allocate a durable revision, write an artifact, authorize a
promotion, decide a completion gate, compile a workflow topology, or supervise
a process. Workspace allocates artifact revisions and commits identity records;
completion resolves promotion conflicts and accepts evidence; workflow owns its
single graph compiler; a declared provider owns process and scratch lifetime.

## Dependency direction and deep interfaces

```text
presentation -> engine -> artifact <- validation
                    |          |
                    |          +-> execution provider port (optional effect)
                    +-> workspace snapshot / exact-plan interface
                    +-> workflow compiled-definition facts
                    +-> completion candidate and gate consumers
```

`artifact` is concrete in-process logic, not a parser port. `validation` owns
the only interface it consumes that represents a real variable effect:

```go
type Provider interface {
    Invoke(ctx context.Context, invocation Invocation) (ProviderReceipt, error)
}
```

There is no parser interface, validator registry service, artifact repository,
generic filesystem abstraction, `Result[T]`, or global interface package.
Callers primarily cross `Engine.Execute`; focused package tests exercise pure
laws at the same smaller interfaces used by the engine.

The small internal transformation surface is:

```go
artifact.Parse(source RawSource, grammar GrammarContract) (ParsedArtifact, []Diagnostic)
artifact.Normalize(parsed ParsedArtifact, profile ProfileDefinition,
    identity IdentityObservation) (NormalizationResult, []Diagnostic)
artifact.CompareDelta(base ResolvedArtifact, proposal ArtifactDelta,
    current ResolvedArtifact, profile ProfileDefinition) DeltaResult

validation.Run(ctx context.Context, request ValidationRequest,
    provider Provider) ValidationResult
validation.Assess(request AssessmentRequest) AssessmentResult
validation.CacheEligibility(entry CacheEntry, request ValidationRequest) CacheDecision
```

`Provider` may be absent. A nil/absent provider is a typed capability fact, not
a panic and not permission to substitute a fake. Pure stages remain usable;
mandatory execution-backed stages return `unavailable` and fail closed.

## Data flow, ownership and lifetimes

The dominant pass is explicit and single-read:

```text
workspace RawSnapshot
  -> lexical budgets and pre-alias YAML scan
  -> Markdown/frontmatter AST with source coordinates
  -> profile-typed value plus identity observation
  -> normalized artifact and ordered indexes
  -> canonical bytes and digest set
  -> syntax/schema/graph/trace/policy facts
  -> optional authorized provider invocation
  -> ordered validation result / cache record / convergence findings
```

The snapshot owns raw byte buffers. `ParsedArtifact` either copies retained
slices or has a documented transfer of ownership; normalized values never
retain mutable aliases into parser buffers. Canonical ordered slices are the
iteration source. Maps are lookup indexes only and are never iterated for
diagnostics, hashing or rendering without a declared sort.

Evaluation never reopens a file, reads environment variables, consults a clock,
or performs a network request. Time, revocation, filesystem and network facts
are `Observation` values in the request. The engine requests a new workspace
snapshot when a new observation is required.

## Common artifact value and source coordinates

The normalized projection is the IC-07 value. It is not the raw authoring file
shape and therefore does not require a file to contain its own blob digest.

```go
type ArtifactEnvelope struct {
    SchemaVersion     string
    Profile           ProfileRef
    Kind              string
    ID                ArtifactID
    WorkID            WorkID
    ArtifactRevision  uint64
    SourceRevision    uint64
    Status            ArtifactStatus
    BlobDigest        Digest
    SemanticDigest    Digest
    DerivationDigest  *Digest
    Source            SourceRef
    Content           ProfileValue
    RetainedExtension []ExtensionField
}

// ArtifactEnvelope exists only for an allocated artifact. A proposed ID is not
// represented by a zero revision or a partially populated envelope.
type NormalizationResult interface{ isNormalizationResult() }

type ProposedArtifact struct {
    ProposedID ArtifactID
    Namespace  string
    Scope      ArtifactScope
    Source     SourceRef
    Content    ProfileValue
    // No ArtifactRevision, allocation provenance, or authoritative envelope.
}

type AllocatedArtifact struct {
    Envelope ArtifactEnvelope
}

type SourceRef struct {
    Scope      SourceScope // workspace | work | external
    OwnerID    string      // workspace ID, work ID, or declared external owner
    Path       string      // normalized workspace-relative path, or typed external locator
    Range      SourceRange
    Provenance Provenance
}

type SourceRange struct {
    Path  string
    Start Position
    End   *Position
}

type Position struct {
    Line, Column uint64 // one-based
    Unit SourceUnit     // utf8-byte | unicode-scalar
}
```

`artifactRevision` and `sourceRevision` remain non-negative integers for the
foundation wire contract. Their ownership is unambiguous without changing
their scalar representation:

- `artifactRevision` is owned by `source.ownerId + artifact.id` and allocated
  by workspace only inside an authoritative transaction;
- `sourceRevision` is interpreted only inside `source.scope + source.ownerId`;
  comparing it across owners or scopes is invalid;
- pure observation, parsing, normalization and validation never increment
  either integer;
- a path move preserves artifact identity but workspace advances the artifact
  revision as required by WS-REQ-003, even if blob and semantic digests remain
  unchanged.

The raw authoring schema contains declared envelope fields and content. It does
not contain computed `blobDigest`, `semanticDigest`, or `derivationDigest` as
authoritative fields. Import formats may carry `declaredDigestClaims`; these
are compared with computed observations and retained as provenance, never used
as their own proof.

Parsing validity, structural validity, submission validity, trusted acceptance
and completion eligibility are separate tagged fields in results. `status:
accepted` in author content is only a claimed lifecycle value until a trusted
workspace/policy observation establishes the fact.

## Schema bundle and typed profile extensions

Artifact-validation owns versioned *payload* schemas and references foundation
definitions rather than copying them. Foundation alone owns the public command,
result-envelope, discriminator, schema registry location and extension-ID
allocation under AM-02. This module contributes the following payload schema
definitions only after that allocation; it MUST NOT publish a competing public
envelope or result discriminator:

| Schema ID | Owner and purpose |
|---|---|
| `sdd://artifact/authoring/1` | Raw declared envelope and profile content; no self blob hash |
| `sdd://artifact/normalized/1` | Computed IC-07 projection above |
| `sdd://artifact/profile-definition/1` | Trusted field grammar, meaning/order rules, limits and kinds |
| `sdd://artifact/delta/1` | IC-08 add/modify/remove and pinned-base observations |
| `sdd://validation/policy/1` | Normalized IC-10 policy and authority facts |
| `sdd://validation/descriptor/1` | Validator semantic contract and implementation provenance |
| `sdd://validation/request/1` | Ordered stages, artifact inputs and separate candidate subject |
| `sdd://validation/result/1` | Stage results, diagnostics and receipt references |
| `sdd://validation/provider-receipt/1` | Provider facts; not completion evidence by existence alone |
| `sdd://validation/assessment/1` | Risk/diagnosis/materiality with explicit unknown states |

A `ProfileDefinition` is trusted configuration, never artifact prose:

```text
ProfileDefinition {
  schemaVersion, name, definitionDigest,
  supportedKinds[], contentSchemasByKind,
  fieldSemantics[] {
    jsonPointer, valueType,
    ordering: unordered-map | ordered-map | ordered-list,
    textMode: exact | line-ending-insensitive | named-profile-rule,
    nullability, unknownChildPolicy
  },
  ID namespaces and syntax,
  grammar/parser budget within engine hard maxima,
  legal status transitions,
  traceability relations,
  deterministic validator requirements,
  scopePriority?: {
    enabledForKinds[], closedValues: must | should | could | wont,
    deliveryScopeRefRequired: true,
    requiredFieldsByValue: {
      should: [workaround], wont: [exclusionRationale, reconsiderationTarget]
    },
    estimateUnitRules
  }
}
```

Profile content is decoded into a closed tagged `ProfileValue`, with a retained
typed extension tree for permitted non-critical fields. An extension declares
its value type, ordering and text semantics. It cannot redefine common
identity, revisions, source units, digest roles, status meaning, authority or
execution. Unknown critical envelope/profile fields are rejected. Permitted
unknown non-critical content is retained and is meaning-bearing by default, so
it participates in semantic identity.

`ScopePriority` is a profile extension rather than a common-envelope field:
not every artifact participates in delivery planning. Where it is enabled, the
parser checks the closed category, delivery-scope reference, required
workaround/exclusion fields, and unit-bearing estimates, then includes the
decision record in semantic identity. It does not decide that an item is a
Must, calculate capacity, prioritize a technically blocked action, or treat a
feature flag as a priority value; guided composition and completion own those
decisions.

Built-in definition data owned here covers lightweight, ordinary and protocol
profiles. Guided workflow owns the schemas and profile resources for `bug-report`, `reproduction`,
`investigation`, `diagnosis`, `regression-contract`, `design`, `tasks`,
`verification` and `policy-review`; artifact-validation validates them through
the trusted `ProfileDefinition` seam. It does not publish those resources or
hard-code a second guided parser. Protocol
profiles encode only declared deterministic rules and cannot claim IETF or
other standards certification.

## Parser and pre-expansion resource budgets

Parsing is a bounded sequence with early returns:

1. Check total raw byte limit and BOM policy, then validate UTF-8 and reject
   forbidden controls before constructing semantic nodes.
2. Lex frontmatter into tokens while counting bytes, tokens, indentation depth,
   anchors, aliases and scalar sizes. Duplicate map keys are rejected at this
   stage with both source ranges.
3. Build an anchor dependency graph containing token ranges and estimated
   expansion weights, without cloning aliased subtrees.
4. Detect anchor cycles and compute a saturating upper bound for expanded
   nodes/bytes. Reject when per-alias, total expansion or hard engine maxima are
   exceeded. Overflow itself is `VALIDATION_YAML_LIMIT`.
5. Materialize aliases only after the bound succeeds, debiting an allocation
   budget during construction as a second guard.
6. Parse Markdown into an AST that preserves fence boundaries, exact code/table
   content and source ranges. Requirement-like text inside a fence remains a
   code node.
7. Project the AST through the selected profile and reject unsupported
   versions/types before graph, policy or provider work begins.

Profile limits may narrow engine hard maxima but cannot raise them. Diagnostics
report the limit name, configured maximum, observed lower bound, grammar
version and source object; they do not echo secret-bearing input.

## Allocation-once identity

Identity checking consumes a workspace-produced observation; parsing never
persists or allocates identity:

```text
IdentityObservation =
  NewDraft { proposedId, namespace, scope }
  | Allocated { id, namespace, workspaceId, workId, allocationRevision,
                currentArtifactRevision, currentPath }
  | Tombstoned { id, lastArtifactRevision, removalProvenance }
  | MissingRecord { claimedId, previouslyPersistedEvidence }
```

For `NewDraft`, the ID source may propose a collision-resistant ID. The result
is the tagged `ProposedArtifact` value, never `ArtifactEnvelope`; no
`AllocationRecord` exists and a read-only command cannot turn it into
`allocated`. An artifact-creation plan contains the
proposed ID and authoritative file/allocation effects under one exact workspace
transaction. Workspace allocates the first artifact revision at commit. A
failed pre-PREPARED transaction publishes neither record nor revision; replay
returns the original record/revision.

For `Allocated`, edits and moves must match the persisted binding. For
`MissingRecord`, content previously known as persisted is an integrity error;
the parser never silently reuses the claimed ID. Tombstones reserve the ID and
prevent reuse. Uniqueness indexes use `(workspaceId, workId, namespace, id)`;
cross-work references carry both source and target scopes. Display heading,
basename, line and slug are never identity.

## Exact canonical encoding and digest graph

### Canonical value

Canonicalization first creates a typed value. Every retained distinction has a
representation: null versus absent, integer versus string, ordered versus
unordered collections, exact code bytes, profile-sensitive whitespace and
unknown retained extensions. A canonicalization rule is admitted only when the
profile names the affected field and fixtures prove it meaning-preserving.

The binary-safe `AV-C14N-1` encoding is defined recursively over UTF-8 bytes:

```text
N0:                         null
B1:0 | B1:1                boolean
I<len>:<base-10 integer>    integer; zero is 0; no leading zero; no + sign
S<len>:<UTF-8 bytes>        string; no Unicode normalization
L<count>:<value>...         ordered list
M<count>:<key><value>...    unordered map, keys are S values sorted by raw UTF-8 bytes
O<count>:<key><value>...    meaning-bearing ordered map, pairs remain in declared order
```

`<len>` and `<count>` are shortest unsigned ASCII decimal with no leading zero.
Keys are strings, duplicate keys are impossible at this stage, integers are
arbitrary precision in the canonical model, and floats are unsupported unless
a later profile defines an exact lossless representation. Concatenated values
are unambiguous because every scalar is length-delimited and every container
declares child count. Canonical decode followed by encode must reproduce the
same bytes, and parse-to-canonical-to-typed-decode must preserve the complete
profile meaning value. If the implementation cannot represent an input
distinction losslessly, normalization fails instead of choosing a digest.

The semantic canonical value contains `schemaVersion`, exact profile identity,
kind, artifact/work IDs, status, typed content and retained meaning-bearing
extensions. It excludes computed source observations, integer revisions and
all three digest fields. This lets an unchanged move preserve semantic identity
while the workspace still advances the path-sensitive artifact revision.

### Digest framing

All digests are lower-case `sha256:<64 hex>`. Hash preimages use:

```text
Digest(role, fields...) = SHA-256(ASCII "AVD1" || E(role) || E(field1) ...)
```

where `E` is `AV-C14N-1`. Roles are distinct strings:
`artifact-semantic`, `artifact-derivation`, `profile-definition`,
`validator-contract`, `validation-cache-key`, and `artifact-delta`. Blob
identity alone is `SHA-256(exact source bytes)` and is not `AVD1` framed.

Semantic digest fields are, in order: envelope schema version,
canonicalizer version, profile name, profile definition digest and canonical
semantic value. Derivation fields are, in order: derivation schema version,
workflow/profile identity, effective policy identity, ordered input records
`{scope,id,semanticDigest}`, ordered validator semantic-contract identities and
ordered explicit external observations. Ordering is semantic and declared;
callers cannot pass a map and inherit Go map order. Self-digests, raw file
digest claims, incidental evaluation time and output bookkeeping are excluded.

Meaning-sensitive unknowns or uncertain compatibility invalidate
conservatively. A location-sensitive external observation may change
derivation identity after a move even when blob and semantic digests do not.

### Reproducible test vectors

These are real SHA-256 values over the exact bytes shown, not symbolic example
hashes:

| Role | Exact preimage bytes | SHA-256 |
|---|---|---|
| blob, CRLF | `---\r\nschemaVersion: 1.0\r\n---\r\n` | `sha256:2aeadd20d4462ba55b4f93f26e411036818c17d43f4d2999643f0946e335d0b8` |
| blob, LF | `---\nschemaVersion: 1.0\n---\n` | `sha256:3df5f8f15ea0005a1658e15895e9421f8247afe810c56a329a009876ac28689d` |
| semantic framing | `AVD1S17:artifact-semanticS3:1.0S9:av-c14n-1M2:S2:idS5:REQ-1S4:kindS11:requirement` | `sha256:b8bb7b6aa2d570aa74ae858eb2fd2929cc56eacca58395f948abd2e91cb63d41` |

The semantic framing vector is a codec vector for the shown typed map; full
artifact fixture vectors additionally include the profile name/definition
digest required above. Schema examples must use recomputed full values, not
copy the abbreviated vector into a different preimage.

## Traceability without a second graph compiler

Traceability builds an `ArtifactIndex` from normalized artifact entity IDs and
validates a closed relation vocabulary. It does not compile workflow nodes,
predicates, gates or action reachability. Workflow supplies only an opaque
`CompiledDefinitionRef` and typed facts where a profile relation requires one.

```text
TraceEdge {
  source: ScopedEntityRef,
  target: ScopedEntityRef,
  relation,
  disposition: required | optional | waived,
  waiverRef?, sourceRange, provenance, derivationDigest
}
```

Passes are deterministic: index entities, diagnose duplicate scoped IDs,
resolve exact targets, validate relation/profile/status, then derive reverse
coverage. Missing, removed and ambiguous targets retain the source/target IDs,
scope, relation and source range. A path, URL or line number remains a locator;
it never becomes relationship proof. The traceability index may detect cycles
only for artifact relations whose profile prohibits them. It never substitutes
for `workflow.Compile`.

## Pure deltas and three-way facts

`ArtifactDelta` is a closed tagged value:

```text
Add    {targetScope, proposedArtifact, expectedAbsent}
Modify {targetScope, targetId, pinnedBaseRevision, pinnedBaseBlobDigest,
        pinnedBaseSemanticDigest, proposedArtifact}
Remove {targetScope, targetId, pinnedBaseRevision, pinnedBaseBlobDigest,
        pinnedBaseSemanticDigest, tombstoneProposal}
```

Before comparison, `BaseResolver` resolves the complete pinned tuple
`(scope, targetId, artifactRevision, blobDigest, semanticDigest)` from the
supplied immutable snapshot. “Previous revision” is not a valid base. Failure
to resolve is returned as a fact, not hidden as an operational reread.

```text
DeltaResult {
  schemaVersion, baseResolution,
  orderedEntityChanges[], preservedEntities[],
  conflictFacts[] {
    kind: base-missing | base-mismatch | current-changed | same-entity-change |
          delete-modify | rename-collision | id-collision | invalid-reference,
    scopedEntity, base?, proposal?, current?, sourceRanges[]
  },
  diagnostics[], deltaDigest
}
```

The module may classify disjointness and exact conflicting facts. It never
chooses ours/theirs, rewrites a proposal, authorizes promotion or writes a
target. Completion decides whether the facts allow completion and creates any
fresh conflict-resolution proposal; workspace applies the resulting exact
plan.

## Validator descriptors, policy and assessments

`ValidatorDescriptor` separates behavior from implementation:

```text
ValidatorDescriptor {
  schemaVersion, validatorId,
  semanticContract: {version, digest, acceptedSchemas, applicability,
                     requiredCapabilities, outputSchema},
  implementation: {providerId, buildId, binaryDigest, provenance},
  dependencies[], orderKey,
  severity: info | warning | error | critical,
  trustTier: trusted | untrusted | heuristic,
  compatibilityClaims[]
}
```

Descriptors are normalized from trusted configuration. The descriptor digest
covers every field, including implementation provenance and compatibility
claims; a separate semantic-contract digest covers behavior-bearing fields.
Until AM-01 is ratified, the wire severity is exactly foundation's
`info|warning|error|critical`. Blocking is a separate policy fact; `critical`
is never silently mapped away and `hint` is not invented as a fifth severity.

Risk, diagnosis and materiality use one tagged assessment shape with
`unassessed`, `insufficient` and `assessed` states. An assessed record includes
evidence refs, uncertainty, author provenance and any distinct approval facts.
Validation establishes schema, digest and policy-selected issuer facts, not the
truth of prose. Authenticity is policy-specific: trusted CI, local review,
signature or attestation may be selected per class. A cryptographic signature
is not required for every local assessment, and a digest alone never proves
who issued it.

Effective policy normalization retains all precedence inputs and provenance:

```text
EffectivePolicy {
  schemaVersion, policyId, definitionDigest,
  organizationInput, projectNarrowingInput, localMetadataInput,
  resolvedRules[], requiredValidators[], requiredCapabilities[],
  approvalAndExceptionFacts[], issuer,
  issuedAt, expiresAt, revocationObservation,
  authenticityBasis, evaluationInstant
}
```

Project/local inputs may narrow but cannot replace a required organization
input. Missing, unverifiable, expired, revoked or stale required inputs produce
a fail-closed fact.

## Deterministic validation stages

`ValidationRequest` contains one immutable snapshot identity, ordered artifact
inputs, policy/profile/descriptor sets, explicit observations and optionally a
separate candidate subject. The stage plan is fixed by profile/policy:

| Order | Stage | Runs when | Failure behavior |
|---:|---|---|---|
| 10 | syntax | raw input exists | fatal malformed input stops stages requiring an AST |
| 20 | schema | syntax produced AST | structural errors stop stages requiring typed values |
| 30 | artifact-graph | typed artifacts exist | indexes stable IDs/derivation relations only |
| 40 | traceability | referenced types exist | reports all safe independent edges |
| 50 | deterministic-policy | policy facts interpretable | fail closed for missing required authority |
| 60 | execution-backed | approved invocation and provider available | unavailable/error is never fail/pass |
| 70 | heuristic | profile permits advice | cannot satisfy a mandatory proof gate |

Every stage emits `pass|fail|error|skipped|unavailable`, a stable reason,
provenance and dependency disposition. `skipped` names the earlier stage that
made it inapplicable; it is not success. Known independent findings are
aggregated, while unsafe provider invocation is short-circuited. Diagnostics
sort by `(stage order, source path bytes, start line, start column, code,
stable subject ID)`.

Time and volatile facts are explicit:

```text
ObservationSet {
  evaluationInstant,
  policyRevocation, issuerTrust, externalReferences,
  candidateObservation, environmentObservation,
  each with schemaVersion, source, observedAt and digest
}
```

The same complete request yields byte-identical ordered result bytes.
Freshness rechecks at authorization/application remain the owning consumer's
responsibility.

## Invocation and provider receipt

The invocation separates artifact subject, implementation candidate and
process argv:

```text
Invocation {
  schemaVersion,
  validatorContract: {id, semanticContractDigest},
  artifactInputs[]: {scope, artifactId, artifactRevision,
                     blobDigest, semanticDigest, derivationDigest?},
  implementationCandidate: CandidateSubject?,
  policy: {id, definitionDigest},
  profile: {name, definitionDigest},
  process: {
    executable: string,
    argv: string[],          // arguments only; argv[0] is not reconstructed text
    cwd: workspace-relative path,
    environment: ordered allowlisted name/value-reference entries
  },
  requestedLimits,
  scratchRequest,
  observationSetDigest
}
```

The provider constructs the real OS argv as exactly `[executable] + argv`.
Arguments remain distinct byte/string values, so `"arg with spaces"` is one
argument. No shell string, interpolation, implicit executable lookup or command
from Markdown is permitted. Authorization binds executable identity, each argv
element, cwd, environment names/value references, limits, artifact inputs and
candidate identity before the provider is called.

The provider owns process groups/job objects, descendant cleanup, pipe closure,
timeouts, cancellation, output caps, resource enforcement, sandboxing and its
isolated scratch area. Scratch is provider-owned temporary state and grants no
permission to alter authoritative workspace/candidate paths. The provider
reports actual enforcement and changed-input observations. An unexpected
authoritative input change invalidates the receipt; validation never blesses
the change or expands authority.

`ProviderReceipt` distinguishes exit, signal, timeout, cancellation,
unavailable and malformed output, and records redaction before any persistence.
It includes contract/implementation/input/candidate/policy/profile/environment
identities, exact authorized argv/cwd, output digests, truncation, enforcement,
scratch cleanup and observation time. It is evidence provenance, not a
completion decision. A fake receipt is marked `test-only` and is rejected by
production trust policy.

## Cache compatibility and freshness

A cache key binds all AV-REQ-020 inputs in a canonical ordered record. Reuse is
deny-by-default:

1. candidate, artifact input, source revision, profile, policy, semantic
   contract, environment, trust and time observations must match exactly;
2. a semantic-contract digest change always invalidates;
3. a binary/build digest change invalidates unless trusted policy contains an
   explicit compatibility edge from old implementation digest to new digest,
   bound to the unchanged semantic contract and current trust/security facts;
4. a security fix, provider/trust change, revocation or uncertain unknown field
   invalidates even if a version label calls the change a patch;
5. cache lookup returns prior provenance and eligibility; it never upgrades an
   untrusted/stale result or decides a completion gate.

Semantic compatibility and binary identity are deliberately separate.
Compatibility is not inferred from semver, filename, self-declaration or equal
stdout.

## Read-only convergence and remediation

`validation.Assess` operates on normalized snapshots and returns bounded
`Finding` and `RemediationProposal` values. Budgets cover artifacts, edges,
depth, input bytes, evidence bytes and output bytes. Missing facts and budget
exhaustion are explicit findings. A proposal contains intended typed changes
and preconditions but is not a `MutationPlan` and cannot be applied. The engine
must start a normal plan/authorize/apply flow to enact one.

Before and after read-only tests compare path set, bytes, modes, workspace/work/
artifact revisions and provider calls. Pure assessment has no provider call;
execution-backed validation may use provider scratch only under the invocation
contract above.

## Engine.Execute flow and public outcomes

The first public walking slice is `sdd check` through the existing command
envelope:

```text
decode CheckCommand
 -> workspace.read bounded RawSnapshot
 -> artifact.Parse / Normalize
 -> validation.Run pure stages
 -> optional provider.Invoke for separately authorized execution stage
 -> CheckResult + foundation Diagnostic[] + NextAction[]
 -> human or JSON renderer
```

Read-only `check` stops before plan/authorize/apply and does not repair stale
state. `CheckCommand`/`CheckResult` and their public schema placement are a
foundation-owned AM-02 extension; artifact-validation supplies only its typed
handler and payload schemas after the extension is allocated. A blocked observation may use exit `0` only when the requested operation
was observation and the outcome clearly says blocked; an attempted validation
condition uses `1`, malformed input `2`, workspace/schema `3`, stale/conflict
`4`, required provider failure `5`, policy/permission `6`, and interruption
`130`.

Public diagnostics use foundation's structured object:

```text
Diagnostic {
  code: VALIDATION_<CATEGORY>_<DESCRIPTION>,
  severity: info | warning | error | critical,
  message, source: {kind, path?, field?, artifactId?},
  range?: {path, start:{line,column,unit}, end?},
  remediation?, retryable, metadata, provenance?
}
```

The source is never a string and a range never invents a competing
`rangeUnit`. Human and JSON renderers consume the same object. Messages are not
machine contracts. Result/diagnostic schema allocation remains provisional
under AM-01 and AM-02.

## State and authority tables

### Artifact observation state

| Input state | Pure result | Allowed next authority | Forbidden |
|---|---|---|---|
| new draft + proposed ID | `ProposedArtifact` only; never an `ArtifactEnvelope` | workspace creation plan may persist allocation and initial revision | parser writes allocation or emits revision zero as allocated |
| allocated + matching binding | normalized current artifact | caller may request authorized edit/move plan | reallocate on rename |
| allocated + stale revision/path | stale/conflict diagnostic | refresh snapshot/replan | silently adopt current bytes |
| tombstoned ID | historical removed fact | explicit new ID or authorized restore policy | ID reuse |
| persisted claim + missing record | integrity error | repair requires separate authorized workflow | silently create record |

### Validation result authority

| Fact | Validation may establish | Validation may not establish |
|---|---|---|
| syntax/schema/trace fact | deterministic result over pinned inputs | artifact acceptance/completion |
| assessment | structure, provenance and uncertainty | truth of diagnosis prose |
| provider receipt | interpreted execution facts and freshness | process isolation it did not report; completion gate |
| policy input | normalized precedence/authenticity observations | organization identity/distribution |
| delta | base/change/conflict facts | conflict resolution or promotion authorization |
| remediation | bounded proposed change | authoritative write |

## Error taxonomy and cancellation

Internal typed errors distinguish malformed input, unsupported version,
resource limit, identity integrity, stale/conflict, policy/authority,
provider/unavailable, cancellation and invariant defect. Expected domain
negative results remain structured stage results; operational inability to
determine a result is an error/unavailable state, never `fail` or a warning
pass. Engine translates once to stable diagnostics and foundation exits.

`context.Context` is passed to expensive parse/validation passes and provider
invocation. Cancellation before provider invocation produces no invocation.
After invocation, the provider reports descendant cleanup and scratch cleanup;
validation does not claim successful interruption until those facts are known.
No unowned goroutine is introduced for parallel validation. If independent
stages are later parallelized, results are buffered and sorted by the fixed
stage/diagnostic order before exposure.

## Requirement-to-design traceability

| Requirement | Design realization |
|---|---|
| AV-REQ-001 | versioned raw/normalized schemas and bounded parse passes 1, 6, 7 |
| AV-REQ-002 | token/anchor dependency scan, saturating pre-expansion estimate and hard maxima |
| AV-REQ-003 | IC-07 envelope, typed profile tree and guided profile seam |
| AV-REQ-004 | scalar revision ownership invariants and observation-state table |
| AV-REQ-005 | proposed versus allocated `IdentityObservation` and transactional first revision |
| AV-REQ-006 | scoped identity indexes, tombstones and exact scoped references |
| AV-REQ-007 | separated provenance/authenticity policy with no universal signature rule |
| AV-REQ-008 | lossless `AV-C14N-1` and field-specific normalization |
| AV-REQ-009 | exact digest framing, ordered closure and real vectors |
| AV-REQ-010 | explicit `ObservationSet` and deterministic sort order |
| AV-REQ-011 | typed `TraceEdge` and `ArtifactIndex`, not workflow compilation |
| AV-REQ-012 | pinned base resolver and pure `DeltaResult.conflictFacts` |
| AV-REQ-013 | split semantic contract/implementation descriptor and foundation severity |
| AV-REQ-014 | fixed seven-stage table and five-state stage result |
| AV-REQ-015 | tagged assessment and distinct author/approval provenance |
| AV-REQ-016 | exact executable/argv invocation, provider-owned supervision/scratch |
| AV-REQ-017 | normalized precedence inputs and fail-closed authority observations |
| AV-REQ-018 | versioned lightweight/ordinary/protocol/guided profile interface |
| AV-REQ-019 | bounded pure assessment and non-applicable remediation proposals |
| AV-REQ-020 | deny-by-default cache identity and explicit compatibility edge |
| AV-REQ-021 | foundation diagnostic object, severity, source units and exit mapping |

## Conformance fixture design

The requirements' 26 fixtures remain authoritative. The following table names
the principal test surface and expected invariant; no fixture is claimed run.

| Fixtures | Surface | Principal assertion |
|---|---|---|
| AV-FIX-001 through AV-FIX-004 | parser package/fuzz corpus | duplicate/bounded/malformed/fenced inputs cannot become a false typed requirement |
| AV-FIX-005 through AV-FIX-006 | canonical codec golden vectors | cosmetic byte changes only affect blob; meaning changes affect semantic digest |
| AV-FIX-007 through AV-FIX-008 | Engine + workspace identity fake | allocation once, move preservation, scoped collisions and no arbitrary basename link |
| AV-FIX-009 through AV-FIX-012 | trace/delta/derivation pure laws | removed/unavailable/location/cycle facts are deterministic and side-effect free |
| AV-FIX-013 through AV-FIX-015 | policy/descriptor/assessment | claims, forged validators and heuristics cannot become trusted proof |
| AV-FIX-016 through AV-FIX-018 | provider contract adapter | inert prose, exact argv, unavailable/timeout/cancel/truncation, redaction and env denial |
| AV-FIX-019 through AV-FIX-020 | receipt/candidate interpretation | malformed or stale-subject receipt cannot become evidence/cache hit |
| AV-FIX-021 through AV-FIX-023 | profile/validator compatibility | renamed version, contract/security change and absent provider invalidate/fail closed |
| AV-FIX-024 | delta/completion consumer contract | exact pinned-base conflict facts; no auto-promotion or write |
| AV-FIX-025 | policy normalization | missing/expired/revoked issuer fails closed |
| AV-FIX-026 | convergence assessment | bounded, redacted, deterministic and byte-for-byte read-only |

Parser fuzzing covers byte mutation, duplicate keys, alias graphs, nesting,
frontmatter boundaries and Markdown fences. Canonical codec property tests cover
round trip, injectivity over the typed model and deterministic map order.
Adapter contract tests use a real platform provider only when that separate
provider exists; deterministic fakes test request/interpretation shape but do
not satisfy production proof. Engine tests compare human/JSON facts, stream
placement, next action, no mutation and foundation exit mapping.

## Shared-contract ledger and amendment proposals

| Contract | Producer -> consumer | This design's concrete obligation |
|---|---|---|
| IC-01 | workspace -> engine/artifact | consume one bounded immutable raw snapshot; no reread |
| IC-04/05/06 | workflow -> engine/validation | consume opaque compiled/routing/composition facts; no second compiler |
| IC-07 | artifact -> graph/validation/completion | normalized envelope, source ownership and three digests |
| IC-08 | artifact -> completion | pinned pure delta/conflict facts only |
| IC-09 | validation/provider -> engine/completion | distinct artifact/candidate subjects, stage state and receipt provenance |
| IC-10 | validation normalizer -> all decision owners | typed policy/authenticity/time facts; owners enforce decisions |
| IC-11 | completion candidate observer -> validation | provisional `CandidateSubject`; AV only binds/interprets it |

Required amendments remain proposals, not silent edits:

1. **AM-01 / foundation FND-001 and FND-003:** retain
   `info|warning|error|critical`; publish blocking separately and add total
   mapping only after central review.
2. **AM-02 / foundation FND-001, FND-002, FND-016:** centrally allocate the
   schema IDs/result discriminator above; do not mutate closed `1.0` or let AV
   independently claim `1.1`.
3. **AM-04 / foundation FND-006:** narrow foundation implementation to native
   record/plan canonicalization; AV owns generic Markdown/profile semantic and
   derivation algorithms. Consumer conformance compares shared results rather
   than keeping two digest engines.
4. **AM-06 / guided GW-004–GW-007 and GW-012:** guided owns typed content
   definitions and journey assertions; AV owns common envelope/parser/digests.
   GW-012 becomes an AV consumer integration slice.
5. **AM-07 / guided GW-008 and GW-013:** consume IC-09 and the future
   completion candidate contract; a fake provider cannot satisfy production
   verification.
6. **AM-10 / guided GW-001 and foundation compatibility text:** store exact
   implementation provenance separately from policy-proven semantic
   compatibility; patch labels alone do not preserve evidence.

## Explicit exclusions

- No workspace file writes, revision allocation, journal, lock, recovery or
  idempotency implementation.
- No workflow topology compiler, predicate evaluator, action router or journey
  definition ownership.
- No completion gate decision, conflict resolution, DONE, receipt acceptance,
  promotion/archive authorization or lifecycle mutation.
- No process tree, sandbox, resource controller, scratch implementation or
  Git mutation.
- No remote organization-policy distribution, issuer provisioning, universal
  signature system, embedded model or automatic semantic truth claim.

## Open decisions and release blockers

1. AM-01 and AM-02 must allocate severity/blocking and protocol extension
   schemas before the wire bundle can be frozen.
2. Parser libraries and their exact grammar behavior must be selected only
   after they pass AV-FIX-001 through AV-FIX-004 and pre-expansion budget tests; this design
   approves no dependency.
3. The completion owner must ratify `CandidateSubject`, bookkeeping exclusions
   and evidence acceptance before IC-09/IC-11 production integration.
4. The execution-provider spec must ratify enforceable limits, platform
   isolation, receipt transport and scratch cleanup before an execution-backed
   pass can be production evidence.
5. Policy owners must supply authenticity, compatibility, expiry/revocation
   and retention rules; signatures are neither universally required nor
   universally sufficient.
6. Workspace must expose the allocation binding and scoped revision facts
   described by WS-REQ-003; AV cannot truthfully test durable allocation/replay
   using only a fake.
7. Cache retention and redaction durations remain company decisions. Cache
   eligibility can be implemented without choosing destructive retention.

## References

- `requirements.md`
- `docs/kiro-input/next-prerequisite-specs.md`, sections 2–4, 8 and 10
- `docs/kiro-input/prerequisite-review.md`, IC-01–IC-12 and AM-01–AM-11
- `.kiro/specs/sdd-engine-foundation/design.md`
- `.kiro/specs/sdd-workspace-transactions/requirements.md`
- `.kiro/specs/sdd-workflow-graph/requirements.md`
- `.kiro/specs/sdd-completion-lifecycle/requirements.md` (provisional consumer contract)
- `.kiro/specs/sdd-guided-workflows/design.md`
- `docs/architecture/go-project-structure.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/cli-ux.md`
- `.kiro/steering/quality-and-safety.md`

## For the chosen ones

The hardest invariant is not “stable IDs” or “deterministic hashes” in
isolation. It is keeping three clocks orthogonal: content meaning changes the
semantic digest, authority-bearing workspace mutations advance scoped integer
revisions, and external trust/time observations change derivation or cache
eligibility. Collapsing any two produces either self-invalidating evidence or
stale proof that appears current. The allocation record and canonical encoder
are valuable precisely because they prevent identity, chronology and meaning
from impersonating one another.
