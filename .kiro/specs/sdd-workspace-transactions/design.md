# Design Document: sdd-workspace-transactions

## Status, authority, and review gate

This design is **accepted for core implementation on 2026-10-17** by ADR-0006.
It does not claim implementation, native platform certification, production
authority-provider approval, destructive-retention approval, or a published
public-protocol extension. `requirements.md` remains the normative behavior
source. `DEC-01`, `DEC-02`, `DEC-03`, and `DEC-05` are capability/release gates
as identified below; they no longer block pure/module/fake implementation.

The module is deliberately deep. Callers supply selection coordinates, bounded
read-set intent, an exact already-authorized plan, or an explicit recovery
request. The implementation owns discovery, stable raw observation, canonical
paths, local locks, revision allocation, staging, durable intent, idempotent
installation, journal integrity, recovery, quarantine, and retained replay
facts. A caller does not orchestrate those steps.

## Scope and exclusions

Workspace owns:

- immutable workspace/work identity and mutable path/slug coordinates;
- normalized repository-relative paths and effect-time containment;
- collectively coherent raw snapshots and raw blob observations;
- scoped workspace/work/artifact revisions and tombstones;
- exact-plan identity, application, durable idempotency, and local recovery;
- transaction records, the authoritative append-only journal, and rebuildable
  projections;
- local lock ordering and platform capability reporting.

Workspace consumes but does not define:

- the foundation `Outcome`, `Diagnostic`, `NextAction`, exit codes, source
  positions, and closed public protocol (`sdd-engine-foundation`);
- artifact semantic/derivation digest meaning (`sdd-artifact-validation`);
- workflow, composition, completion, reviewer, or waiver decisions;
- organization policy distribution or identity issuance;
- process supervision, claims/leases, Git mutation, and remote effects.

An authorization fact says that a named trusted provider granted a named
principal a scope for one plan and policy snapshot. Workspace verifies that
fact and its binding. It never decides that a bundle is valid, a transition is
allowed, or completion evidence is sufficient.

## Requirement-to-design traceability

| Requirement | Design sections | Primary verification surface |
|---|---|---|
| `WS-REQ-001` | Identity/path model; containment algorithm; platform rules | path/property tests and native adapter tests |
| `WS-REQ-002` | `Snapshot`; coherent capture algorithm | module interface and race/fault fixtures |
| `WS-REQ-003` | revision records and mutation table | transition/model tests |
| `WS-REQ-004` | `PlanCore`, canonical digest, authorization binding | canonicalization and exact-plan tests |
| `WS-REQ-005` | apply algorithm and durable platform primitives | adapter contract and native fault tests |
| `WS-REQ-006` | transaction records, transition table, crash matrix | restart-at-every-boundary suite |
| `WS-REQ-007` | snapshot generation, authoritative/derived state | incomplete-commit and rebuild tests |
| `WS-REQ-008` | idempotency record/fingerprint/replay | concurrency, lost-response, expiry tests |
| `WS-REQ-009` | lock resource/order/ownership model | interprocess and PID-reuse tests |
| `WS-REQ-010` | journal framing, deduplication, compaction | replay/corruption tests |
| `WS-REQ-011` | authorization verifier seam and migration plan | provider fake and migration faults |
| `WS-REQ-012` | limits, diagnostics, read-only rules, support matrix | boundary and renderer integration tests |

## Architecture and dependency direction

```text
presentation -> engine -> workspace.Transactions
                             |       |
                             |       +--> authorization.Verifier adapter
                             +----------> workspace-owned LocalPlatform port
                                              |
                                   platform/filesystem native adapter
```

`internal/workspace` owns the external module interface and its internal ports.
`internal/platform/filesystem` satisfies the platform port and contains
OS-specific handles, rename/replace behavior, locking, file identity, and flush
details. The engine coordinates artifact/workflow/completion facts into a plan;
workspace imports none of those evaluators. Foundation value types are consumed
through their owning package or generated protocol package; this design does
not create a global interfaces package.

The supported external interface is intentionally small:

```go
// Transactions is the workspace module interface used by Engine.Execute.
// Returned slices and byte buffers are owned by the result and are not aliases
// of adapter buffers.
type Transactions interface {
    Snapshot(context.Context, SnapshotRequest) (Snapshot, error)
    Apply(context.Context, ApplyRequest) (ApplyOutcome, error)
    Recover(context.Context, RecoveryRequest) (RecoveryOutcome, error)
}
```

`Snapshot` incorporates explicit discovery/selection and read-only transaction
inspection. `Apply` owns the entire plan-to-receipt sequence. `Recover` is the
only mutation-capable recovery entry point. There are no public `Stage`,
`AppendEvent`, `AllocateRevision`, `Lock`, or `Cleanup` methods: exposing them
would let callers violate the transaction invariant.

The internal platform seam has two real adapters: the native implementation
and a deterministic fault-injecting implementation. It exposes platform facts
and durable operations, not policy:

```go
type localPlatform interface {
    Capabilities(context.Context, RootHandle) (FilesystemCapabilities, error)
    OpenRoot(context.Context, NativeRoot) (RootHandle, RootIdentity, error)
    Acquire(context.Context, RootHandle, LockRequest) (LockGuard, error)
    Capture(context.Context, RootHandle, CaptureRequest) (CaptureResult, error)
    Persist(context.Context, RootHandle, PersistStep) (PersistResult, error)
}
```

`PersistStep` is a closed internal tagged value for write-new-file, sync-file,
replace-one-file, sync-directory, append-journal-frame, and remove-owned-path.
It is not a public generic filesystem interface. `internal/workspace` chooses
the legal next step and validates every returned native identity; the adapter
performs one requested primitive and reports what the platform actually made
durable.

Authorization proof verification is a second real effect seam because company
providers and deterministic test authorities differ:

```go
type AuthorizationVerifier interface {
    Verify(context.Context, EncodedAuthorizationFact) (ProviderFact, error)
}
```

The verifier establishes provider authenticity and proof integrity. Workspace
still compares the returned fields to `PlanCore`, the current trusted instant,
and the current revocation observation. It does not ask the verifier whether
domain artifacts are valid.

## Plain Go values and ownership

The following declarations are design-level shapes, not implemented code or a
public Go SDK. Fields represented as `[]byte` or slices are copied at the
module seam. Map iteration never participates in ordering or hashing.

```go
type Selection struct {
    ExplicitRoot *NativeRoot
    WorkspaceID  string
    WorkID       string
    WorkSlug     string
}

type SnapshotRequest struct {
    Selection Selection
    Scope     SnapshotScope
    Paths     []WorkspacePath // already ordered or normalized then sorted
    Limits    Limits
}

type Snapshot struct {
    SnapshotID        string
    Workspace         WorkspaceObservation
    Work              *WorkObservation
    Scope             SnapshotScope
    Generation        uint64
    ReadSetDigest     Digest
    Paths             []PathObservation
    ArtifactIdentities []ArtifactIdentityObservation // only if requested by scope
    Transactions      []TransactionObservation // only if requested by scope
    PolicyBytes       []OpaqueBytesObservation // only if requested
    LimitsApplied     Limits
}

type PathObservation struct {
    Path         WorkspacePath
    Presence     Presence // absent or present
    Kind         FileKind
    Bytes        []byte
    BlobDigest   Digest
    Size         uint64
    Mode         uint32
    NativeID     NativeFileIdentity
    LinkCount    *uint64
    CompareKey   string // collision checking only; never plan identity
}

// ArtifactIdentityObservation is the workspace-owned durable binding used by
// artifact validation. It is derived from the same immutable snapshot as Paths;
// consumers do not open a second identity store or reread workspace files.
type ArtifactIdentityObservation struct {
    WorkspaceID             string
    WorkID                  string
    Namespace               string
    ArtifactID              string
    State                   ArtifactIdentityState // allocated | tombstoned | missing-record
    AllocationRevision      *uint64
    AllocationProvenance    *ProvenanceRef
    CurrentArtifactRevision *uint64
    CurrentPath             *WorkspacePath
    RemovalProvenance       *ProvenanceRef
}

type ApplyRequest struct {
    Plan             MutationPlan
    Content          []ContentBlob
    Authorization    EncodedAuthorizationFact
    Idempotency      IdempotencyInput
    DisclosureCaller CallerFact
}

type RecoveryRequest struct {
    Selection        Selection
    TransactionID    string
    Authorization    RecoveryAuthorization
    DisclosureCaller CallerFact
    Limits           Limits
}
```

`NativeRoot`, `RootHandle`, `NativeFileIdentity`, and `CompareKey` never appear
in semantic plan identity or public machine paths. `WorkspacePath` is normalized
POSIX text with no leading slash. `Presence` is explicit; an absent observation
is not encoded as a zero-valued present file.

`SnapshotScope` is a closed internal enum such as `selection`, `work-current`,
`artifact-identities`, `plan-inputs`, or `recovery-status`. Each scope has an implementation-owned
read-set builder. Callers may narrow a scope with paths but cannot ask an
unbounded “all repository files” question through an omitted limit.

This `Snapshot` is the workspace-owned `RawSnapshot` extension to the narrower
foundation snapshot. The engine first obtains foundation-owned discovery and
selection, then asks workspace to capture this value once. Artifact validation,
workflow evaluation, and completion consume projections of this same value;
they MUST NOT reopen files or maintain a parallel identity database. The
conversion preserves foundation path/source vocabulary and adds raw bytes,
explicit absence, native identity, generation/read-set identity, transaction
status, and durable artifact identity observations. A `NewDraft` proposal is
intentionally absent from `ArtifactIdentities`: it is a consumer value until an
authorized workspace transaction atomically commits the source file,
allocation binding, and initial artifact revision. A claimed persisted ID with
no binding is returned as `missing-record`, never synthesized.

## Identity and canonical path model

The canonical identity tuple is:

```text
workspace = immutable workspaceId
selected project = native root coordinate + proven root identity
work = immutable workId
display/location = mutable slug and workspace-relative path
```

A moved/recloned root may resolve to the same workspace ID, but the newly
opened native root identity is always re-proven. An explicit root is inspected
only at that root. Upward discovery chooses the nearest `.sdd` marker and stops
at a nested marker even if it is malformed; it never crosses the marker to find
a usable parent.

Path normalization is a pure transformation:

1. decode UTF-8 without changing normalization form;
2. require relative POSIX components separated by `/`;
3. reject empty, `.`, `..`, embedded native/alternate separators, NUL, device
   syntax, platform-reserved names, and a component/path over adapter limits;
4. construct a comparison key using declared case and Unicode comparison rules;
5. reject duplicate comparison keys but retain original canonical spelling;
6. never normalize file contents or use the comparison key in plan hashing.

### Containment and TOCTOU algorithm

At snapshot/planning time and again for each effect immediately before install:

1. open the selected root and retain its native identity/handle;
2. walk every existing component relative to that handle using no-follow
   operations; record each component identity and link/reparse facts;
3. for a create, open the parent no-follow and prove the final name is absent;
   for update/delete, open the final object no-follow and prove regular-file
   kind, identity, digest, and relevant mode;
4. prove every opened component descends from the selected root handle rather
   than from string-prefix comparison;
5. compare current component/final identities with the plan observations;
6. immediately before replacement, repeat the final parent/target check using
   handles; a change is `DIAG_PRECONDITION_CHANGED`, never a new target;
7. validate move destinations independently with the same algorithm.

For a managed hard-linked file, the accepted behavior is atomic replacement of
the directory entry, which leaves the old inode and outside aliases unchanged.
If the adapter cannot prove that behavior, or an effect would mutate in place,
it returns `DIAG_UNSAFE_HARD_LINK`.

## Collectively coherent snapshots

Per-file stable reads are necessary but insufficient. Snapshot capture uses a
workspace snapshot lock and a persisted transaction-generation control record:

```text
SnapshotGeneration {
  schema: "sdd.workspace-generation/1"
  stableGeneration: uint64
  activeTransactionId: string|null
  checksum: digest
}
```

The read-side snapshot lock and writer-exclusive lock are modes of the same
canonical `10/transactions/<workspaceId>` resource. Its lock primitive and
generation record are created during workspace initialization; read-only
capture does not create lock files, emit heartbeats, or update generation.

Capture proceeds as follows:

1. normalize selection and build the bounded ordered read set;
2. acquire the workspace read-side snapshot lock;
3. read and checksum the generation record; if `activeTransactionId` is set or
   a durable `PREPARED` transaction is not finalized, return
   recovery-required rather than domain bytes;
4. for each path, open no-follow; stat-before, bounded-read, hash, stat-after;
   retry the whole capture (not just that file) when identity, size, bytes, or
   relevant metadata changed;
5. reread generation and required transaction summaries while the read lock is
   still held;
6. accept only when both generation observations are identical and stable;
7. compute `readSetDigest` from scope, ordered path observations, workspace/work
   revisions, and generation; then release the lock.

The writer holds the corresponding exclusive lock from final validation until
the stable generation is published. Before the first authoritative effect it
durably sets `activeTransactionId`; after all effects, events, committed
summary, and commit marker are durable, it increments `stableGeneration` and
clears the active ID. A crash leaves an active or prepared transaction, which
forces recovery-required. The generation is coordination metadata, not a
workspace/work/artifact revision and not a substitute for raw observations.

The adapter retries a stable read and whole-snapshot capture only up to limits
in `Limits`. Exhaustion returns `DIAG_SNAPSHOT_UNSTABLE`, exit `4`. The returned
value owns its bytes. Pure evaluators cannot call back through the module to
refresh one path under the same `SnapshotID`. This is coherent against
cooperating workspace transactions. An uncontrolled editor does not take the
lock; stable per-file reads cannot promise universal point-in-time isolation
across several independently edited files. Before `PREPARED`, apply revalidates
the complete guarded read set together and rejects detected drift. Read-only
outcomes expose this observation model rather than claiming a stronger
filesystem snapshot than the adapter can provide.

## Revision records and allocation

Revision coordinates are explicit values:

```go
type RevisionRef struct {
    Scope         RevisionScope // workspace, work, artifact
    OwnerID       string
    Value         uint64
    SchemaVersion uint32
    Provenance    RevisionProvenance
}
```

`RevisionProvenance` records transaction ID, plan digest, and allocation
sequence; wall time is audit data only. A tombstone retains owner ID, final
revision, prior aliases, deletion transaction, and schema version.

Every `RevisionIntent` in the authorized plan contains exact expected and
proposed integer values (normally expected + 1 for an advancing scope). While
the write lock is held, apply validates that the allocator state still permits
those exact values and reserves them; it never substitutes newly chosen outputs
after authorization. Reserved values are recorded in the intent and remain unpublished while the
transaction is `STAGING`. Each affected authoritative record is staged with
its exact allocated value. Publishing the effects and committed receipt makes
the set visible together to engine readers through the generation rule.

The requirements mutation table is implemented literally. In particular:

| Mutation | Allocation in intent | Guard beyond direct target |
|---|---|---|
| artifact byte submission | artifact only unless work record changes | complete current work record, composition, lifecycle, membership |
| work slug change | work; workspace only if its authoritative index changes | old and new path/native identity |
| lifecycle/composition/graph receipt | work; artifact only if its authoritative state changes | prior work revision and referenced facts |
| journal-only observation | none | event ID namespace and prior checksum |
| replay | reuse recorded values | idempotency fingerprint and disclosure |

A no-op plan is either rejected as non-authorizable or returns an explicit
no-op observation before `PENDING`; it does not create a transaction, event, or
revision. An external edit changes a raw observation, not a revision; a later
authorized reconciliation plan must guard those bytes and allocate revisions
according to its actual authoritative effects.

## Exact mutation plan and non-circular authorization

### Plan values

```go
type MutationPlan struct {
    Core           PlanCore
    Digest         Digest
}

type PlanCore struct {
    Descriptor     PlanDescriptor
    WorkspaceID    string
    WorkID         string
    SourceSnapshot SnapshotBinding
    Preconditions  []Precondition
    Effects        []Effect
    Revisions      []RevisionIntent
    AuthScope      AuthorizationScope
    Limits         Limits
}

type AuthorizationScope struct {
    ProviderID     string
    PrincipalID    string
    Scope          string
    PolicyDigest   Digest
    PolicyVersion  string
    GrantKind      string
}
```

`PlanDescriptor` is part of hashed `PlanCore`. It contains the internal plan schema ID, supported effect
vocabulary version, canonicalization version, creation instant, and freshness
deadline. Foundation-native effect kinds are `create-file`, `update-file`, and
`delete-file`. `create-directory`, `remove-empty-directory`, `move-file`, and
`metadata-update` exist only as provisional extension variants and are rejected
unless the centrally allocated foundation extension is negotiated. A file
effect never creates or removes a parent directory implicitly. Internal
staging/journal steps are never domain effects.

Every effect has a stable effect index, kind, path(s), exact before
observation, exact after blob digest/size/mode, revision intents, and optional
opaque owner facts. Effect order is semantic. Content bytes travel in
`ApplyRequest.Content`, keyed by digest; apply checks size and digest before
staging. The digest, not a caller-provided temporary path, selects content.

### Canonical plan digest

Plan identity is:

```text
SHA-256(
  "sdd.workspace-plan-core/1\x00" ||
  FoundationCanonicalJSON(PlanCore)
)
```

The foundation canonical JSON encoder is reused: UTF-8, schema/type domain
separation, sorted object keys, preserved array order, exact string bytes, no
non-finite numbers, and explicit nullability. Workspace adds no second generic
canonicalizer. Before hashing it validates:

- normalized paths and collision-free comparison keys;
- strictly increasing unique effect/precondition indexes;
- all before/after digests and sizes present for the effect kind;
- exact source snapshot/read-set binding;
- revision intents matching the normative mutation table;
- authorization scope and policy identity present;
- supported descriptor/effect versions and declared limits.

Because the descriptor is in `PlanCore`, changing the schema, effect vocabulary,
canonicalization version, creation instant, or freshness deadline changes the
plan digest. The digest excludes `Digest` itself, the final authorization fact/proof,
authorization grant ID/nonce generated after planning, transaction ID, lock
nonce, staging names, journal sequence/offset, retry/evaluation time, and
adapter-native spelling. Thus the provider can issue a grant over the completed
plan digest without a self-hash cycle.

The request fingerprint is distinct:

```text
SHA-256(
  "sdd.workspace-request/1\x00" ||
  canonical(command-kind, plan-digest, content-digest-list,
            idempotency-namespace, authorization-fact-digest,
            expected-revisions, caller-principal)
)
```

`authorization-fact-digest` includes provider/grant identity, issued/expiry
metadata, bound plan/policy/scope/principal, revocation evidence identity, and
proof digest. It excludes the current retry instant. Therefore retrying later
does not change the fingerprint; changing the grant or scope does.

### Authorization check

`ProviderFact` contains provider ID, grant ID, principal, scope, exact plan
digest, policy digest/version, issued/expiry instants, revocation status and
observation instant, nonce, and proof digest. Immediately before `PREPARED`,
workspace requires:

1. verifier success from the configured trusted provider;
2. exact equality with `PlanCore.AuthScope` and `Plan.Digest`;
3. issued/not-before/expiry and revocation observations valid at the injected
   trusted evaluation instant;
4. caller and idempotency namespace consistent with the principal;
5. no policy identity drift from the plan's raw preconditions.

Failure is exit `6` before intent. Workspace does not inspect an artifact,
reviewer role, completion gate, or workflow state to decide permission. After
`PREPARED`, default recovery uses the recorded authorization. A separately
trusted restrictive revocation fact may force `QUARANTINED`; it cannot replace
the plan or silently acquire a new grant.

## Internal durable records and proposed layout

The concrete layout below is the **accepted internal storage baseline**. It is
not a new native-format or public-protocol version:

```text
.sdd/
├── workspace.yaml
├── events.jsonl
└── .runtime/
    └── transactions/
        ├── generation.json
        ├── idempotency/<namespace-digest>/<key-digest>.json
        └── <transaction-id>/
            ├── transaction.json
            ├── manifest.json
            ├── staged/000000.blob
            ├── PREPARED
            ├── COMMITTED
            └── committed-summary.json
```

ADR-0006 fixes the directory at `.sdd/.runtime/transactions/`. Retention
duration still requires `DEC-03`; until then destructive expiry/compaction is
disabled and replay ambiguity facts are retained. User-supplied idempotency
keys are not used as filenames; filenames use a domain-separated digest and
the original key is retained only according to the approved privacy policy.

### `transaction.json`

```text
TransactionRecord {
  schema: "sdd.workspace-transaction/1"
  transactionId
  state: PENDING|STAGING|PREPARED|COMMITTED|FAILED|QUARANTINED
  planDescriptor
  planDigest
  requestFingerprint
  sourceSnapshotId
  readSetDigest
  preconditions[]
  effects[]
  revisionIntents[]
  authorizationReference {providerId, grantId, factDigest}
  idempotencyReference {namespaceDigest, keyDigest}
  manifestDigest
  progress[] {effectIndex, status, observedPostDigest?}
  journal {firstSequence?, lastSequence?, eventIds[], priorChecksum?}
  failure? {code, boundary, retryability, evidenceDigest}
  recordSequence
  previousRecordChecksum
  checksum
}
```

Updates use write-new/sync/atomic-replace/directory-sync. `recordSequence`
prevents an older valid record from replacing a newer one. The checksum detects
accidental corruption, not malicious local tampering.

### `manifest.json`

The owned-staging manifest is durable before cleanup may trust an entry. Each
entry has effect index, relative staging path, blob digest, size, creation
nonce, and persisted state. Paths outside the transaction directory or unknown
entries are never deleted. Manifest replacement itself is a fault boundary.

### `committed-summary.json`

The retained commit fact contains transaction ID, plan digest, request
fingerprint, idempotency namespace/key digests, final revisions, changed paths,
event IDs/result digest, committed journal sequence/checksum, authorization
reference (not secret proof), outcome payload digest/reference, commit instant
for audit, and replay-expiry policy reference. It survives lazy staging and
progress cleanup. Deleting it while its replay period is active is forbidden.

### Structurally valid illustrative record

The following is internal storage JSON, not a public `1.1` envelope and not a
digest-conformance fixture. Runtime fixtures must recompute every digest from
their complete bytes.

```json
{
  "schema": "sdd.workspace-committed-summary/1",
  "transactionId": "tx-01JTEST000000000000000001",
  "planDigest": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
  "requestFingerprint": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
  "finalRevisions": [
    {"scope": "artifact", "ownerId": "artifact-17", "value": 8}
  ],
  "changedPaths": [".sdd/work/example/requirements.md"],
  "eventIds": ["workspace-event:tx-01JTEST000000000000000001:0"],
  "resultDigest": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
  "replayPolicyRef": "company-retention-policy:UNDECIDED"
}
```

## Transaction state machine

### Legal transitions

| Current | Trigger and durable evidence | Next | Retry behavior |
|---|---|---|---|
| absent | accepted unique idempotency request; `PENDING` record durable | `PENDING` | duplicate joins this record |
| `PENDING` | lock acquired, all inputs available, manifest creation starts | `STAGING` | restart may continue or fail with owned cleanup |
| `STAGING` | all stage blobs, manifest, revisions, exact intent and preconditions synced; `PREPARED` marker directory-synced | `PREPARED` | only roll forward/quarantine |
| `PENDING`/`STAGING` | cancellation/fault and owned cleanup proven complete | `FAILED` | stored failed outcome replays |
| `PENDING`/`STAGING` | corrupt/incomplete ownership proof | `QUARANTINED` | manual authorized recovery only |
| `PREPARED` | all effects/events/summary/commit marker durable | `COMMITTED` | stored outcome replays |
| `PREPARED` | violated precondition, missing stage, restrictive revocation, or corrupt evidence | `QUARANTINED` | preserve all evidence |
| `COMMITTED` | replay/acknowledgment | `COMMITTED` | no writes/revisions/events |
| any nonterminal | torn/unknown required record where safe cleanup cannot be proved | `QUARANTINED` | no guessing |

Forbidden transitions include `PREPARED -> FAILED`, `QUARANTINED -> STAGING`,
`COMMITTED -> PREPARED`, and any transition that changes plan digest,
authorization binding, revision intents, stage digests, or event IDs. Resolution
of quarantine is a separately authorized recovery operation that either proves
the original roll-forward or records a terminal manual disposition; it never
edits the old plan into a new one.

### Apply phases

1. **Pure acceptance:** decode plan, recompute digest/fingerprint, check limits,
   effect vocabulary, content digests, path syntax, and idempotency lookup.
2. **Locked validation:** acquire canonical locks; recapture generation,
   revisions, raw observations, collision keys, containment, platform support,
   and authorization facts. No revision is yet public.
3. **Owned staging:** create `PENDING`, then `STAGING`; persist manifest entries
   before treating files as owned; write and sync complete content blobs.
4. **Durable intent:** allocate exact revision values, persist complete intent,
   sync it, create/sync `PREPARED`, and sync the transaction directory.
5. **Roll-forward:** publish active generation; for each effect repeat
   containment/precondition checks, install one file, sync its parent as
   supported, and persist progress. Append deduplicated domain events.
6. **Commit fact:** persist committed summary, `COMMITTED`, journal linkage, and
   stable generation in that order; then return the stored outcome.
7. **Lazy cleanup:** remove only manifest-owned staged bytes. Cleanup failure is
   reported but cannot uncommit or remove retained replay facts.

Cancellation before `PREPARED` stops new staging and returns `130` only after
owned cleanup succeeds. If cleanup cannot complete or ownership is uncertain,
the outcome is recovery/quarantine (exit `4`), not a falsely clean
interruption. Cancellation at or after `PREPARED` preserves intent and returns
an interruption/recovery reference; recovery continues from the recorded plan.

## Crash and durable-boundary matrix

Every `Persist` action below has injected failures immediately before and after
the primitive. “After” includes loss of the process response before the module
records the next step.

| Durable boundary | State visible after restart | Required restart action |
|---|---|---|
| transaction directory creation / directory sync | absent or unowned directory | ignore absent; quarantine unknown unmanifested material |
| initial `PENDING` record replace / sync | absent or `PENDING` | no authoritative effect; join or owned cleanup |
| manifest entry replace / sync | prior or new complete manifest | delete only entries proven by the accepted manifest |
| staged blob write / file sync | incomplete/unlisted or complete listed blob | remove proven incomplete pre-intent bytes or restage; never glob |
| `STAGING` record replace / sync | `PENDING` or `STAGING` | validate manifest and continue/clean safely |
| exact intent write / file and directory sync | `STAGING` | no roll-forward until marker is durable |
| `PREPARED` marker create / file and directory sync | `STAGING` or `PREPARED` | only a proven marker enters roll-forward; uncertainty quarantines |
| active-generation marker replace / sync | stable old or active transaction | stable old with prepared record resumes; active forces recovery-required |
| each target replace / target-parent sync | expected-old or exact intended-after | apply once or recognize exact already-applied postcondition |
| progress record replace / sync | old progress or effect recorded | target observation, not progress bit alone, decides resume |
| each journal frame append / journal sync | missing frame or one valid complete frame | append missing event ID once; matching event is already applied |
| committed summary replace / sync | absent or complete summary | reconstruct only from complete intent/events; otherwise quarantine |
| `COMMITTED` marker / directory sync | prepared or committed | finalize commit only after summary/effects/events prove complete |
| stable generation replace / sync | active or stable new generation | committed recovery clears active and publishes same generation once |
| idempotency result replace / sync | in-flight or committed result | link to retained summary; never execute effects again |
| staged cleanup remove / directory sync | some or no owned blobs | retry lazy cleanup; committed fact remains |
| checkpoint/compaction replace / sync | old journal/checkpoint or new complete checkpoint | choose latest valid linked record; never invent events |

### Idempotent effect recognition

For a prepared effect, recovery compares the current target with both recorded
pre- and postconditions:

- exact precondition match: the effect has not applied; install it;
- exact postcondition digest, kind, size, mode policy, and destination identity
  match: treat it as already applied and record progress without replacing it;
- neither match, or a path component identity changed: quarantine.

This rule applies only under a durable `PREPARED` intent. It cannot convert an
unprepared coincidence into a commit. Revision outputs were fixed in the
intent, so recognizing an installed effect never allocates again.

## Effect semantics

| Kind | Required precondition | Installation and postcondition |
|---|---|---|
| `create-directory` (provisional extension) | parent identity exact; final path absent | create exactly one declared directory with declared mode; sync parent as supported |
| `remove-empty-directory` (provisional extension) | directory identity exact, owned, and empty | remove only the declared empty directory; sync parent as supported |
| `create-file` | final path absent; parent identities/containment exact | same-directory exclusive stage, file sync, no-overwrite install, parent sync |
| `update-file` | regular file; old blob/revision/native identity exact | atomic single-file replacement preserving declared permission policy; final digest exact |
| `delete-file` | regular file; old blob/revision/native identity exact | path-bounded removal; parent sync; tombstone effect in same intent |
| `move-file` (provisional extension) | source exact; destination absent; both parents contained | platform rename/replace only under extension contract; no multi-file atomicity claim |
| `metadata-update` (provisional extension) | exact supported metadata and file identity | closed metadata fields only; unsupported metadata fails before intent |

Unknown files, directories, modes, ACLs, alternate data streams, and xattrs are
preserved unless the exact effect and adapter contract name them. A directory
is removed only when the plan explicitly owns it, proves it empty, and its
identity is unchanged. Archive destination collision is a conflict; no suffix
is generated during apply.

Disk-full, permission loss, antivirus/held-open replacement, and directory-sync
failure stop before the next effect. Before `PREPARED` they may clean owned
material; after it they retain the recovery evidence. Bounded retries are
adapter-policy inputs, never an infinite loop.

## Durable idempotency

```text
IdempotencyRecord {
  schema: "sdd.workspace-idempotency/1"
  namespace {workspaceId, operationKind, principalId}
  keyDigest
  encryptedOrRedactedOriginalKey? // company policy
  requestFingerprint
  transactionId
  state: in-flight|recovering|committed|failed|quarantined|expired
  resultDigest
  resultReference
  createdAt
  expiryPolicyRef
  checksum
}
```

Lookup occurs before staging and again under the idempotency/transaction lock.
Same namespace/key plus a different fingerprint is a hard conflict. An exact
match returns or joins the existing state. Join does not require in-process
goroutine sharing; another process can poll/read the durable record under the
caller wait limit. A committed replay verifies current disclosure authority,
then returns the stored result without rechecking the expired action grant or
rerunning effects. Disclosure denial reveals no stored payload.

Expiry never deletes ambiguity. The record becomes an expiry tombstone or the
namespace is explicitly rolled; reuse returns `DIAG_IDEMPOTENCY_EXPIRED`.
Exact retention duration and key privacy are blocked by `DEC-03`.

## Locks

Lock resources are canonical strings derived after selection:

```text
00/workspace/<workspaceId>               format or membership changes
10/transactions/<workspaceId>            journal/generation/idempotency
20/work/<workId>                          work record/lifecycle/composition
30/path/<normalized-workspace-path>       effect source/destination
```

Resources are de-duplicated, sorted by bytewise canonical order, acquired in
that order, and released in reverse. Reentrancy is keyed by `{sessionId,
ownerNonce, resource}` and reference-counted. Metadata records owner nonce,
session/process identity, native lock identity, creation time, optional
heartbeat/liveness evidence, and resources. Release requires the guard plus
nonce; a path or PID is insufficient.

Timeout only bounds waiting and returns exit `4`; it never steals. After a
crash, the native primitive and liveness adapter must prove the owner is gone,
or an explicit authorized recovery is required. A reused PID without matching
session/nonce/native-lock evidence is treated as a different process. These are
local transaction locks, not workflow claims or distributed leases.

## Journal, event deduplication, and projections

`events.jsonl` contains one canonical JSON event per newline. Each event has
schema/event payload version, namespace/event ID, transaction ID, monotonic
journal sequence, event type, audit UTC timestamp, affected identities,
payload/effect reference, previous checksum, and checksum. The checksum domain
includes the exact prior checksum and canonical event bytes excluding the
checksum field.

Prepared intent fixes all event IDs and payload digests. Recovery appends an
event only when replay to the current valid tail finds no matching event ID. A
matching ID with the same transaction/payload digest is already applied; a
matching ID with different content is corruption and quarantines. Thus a crash
after journal sync but before progress update cannot duplicate a logical event.
Wall-clock movement never changes sequence or event identity.

Replay stops at the first torn line, sequence gap, checksum break, unknown
required state-changing event, or manual edit. It does not truncate or repair
during status. Compaction writes a checkpoint to a new file, syncs it, links it
to the source tail checksum/sequence, replaces the checkpoint pointer, and only
then may retire covered events under approved retention. It retains every
commit/idempotency reference required for replay.

Canonical authority is native workspace records, valid journal events, active
transaction intents, and retained committed summaries. Projections and SQLite
indexes are derived. Rebuild consumes one coherent snapshot and valid journal
prefix, emits deterministic facts sorted by stable identity, and advances no
domain revision. Read-only commands do not rebuild automatically.

## Diagnostics and failure precedence

Foundation owns the envelope, severity vocabulary, and exit taxonomy. Workspace
returns typed errors/facts; engine translates once. Provisional stable codes:

| Code | Exit | Meaning / next action |
|---|---:|---|
| `DIAG_WORKSPACE_NOT_FOUND` / `DIAG_WORKSPACE_VERSION_INCOMPATIBLE` | 3 | inspect/select/migrate explicitly |
| `DIAG_PATH_UNSAFE` / `DIAG_PATH_COLLISION` | 3 | correct exact path or use supported platform name |
| `DIAG_SNAPSHOT_UNSTABLE` | 4 | retry a new snapshot; do not reuse the old ID |
| `DIAG_STALE_REVISION` / `DIAG_PRECONDITION_CHANGED` | 4 | refresh, reevaluate, and obtain a new authorization |
| `DIAG_LOCK_TIMEOUT` | 4 | retry after active owner completes; no lock theft |
| `DIAG_RECOVERY_REQUIRED` / `DIAG_TRANSACTION_QUARANTINED` | 4 | inspect then run explicit authorized recovery |
| `DIAG_IDEMPOTENCY_CONFLICT` / `DIAG_IDEMPOTENCY_EXPIRED` | 4 | use original request or a documented new namespace |
| `DIAG_PLAN_NOT_AUTHORIZABLE` | 6 | unsupported effect/version or missing capability |
| `DIAG_AUTHORIZATION_INVALID` / `DIAG_AUTHORIZATION_REVOKED` | 6 | obtain a valid grant for a newly evaluated plan when pre-intent |
| `DIAG_FILESYSTEM_UNSUPPORTED` | 3 | move to/certify a supported local filesystem |
| `DIAG_JOURNAL_CORRUPT` | 4 | preserve evidence and perform explicit recovery |
| `DIAG_LIMIT_EXCEEDED` | 2, 3, or 4 | category is fixed by source: request, stored workspace, or active transaction |
| `DIAG_INTERRUPTED` | 130 | only before intent with completed cleanup; otherwise recovery reference |

Cancellation has highest precedence. Then malformed request, workspace/schema,
conflict/recovery, policy, external adapter failure, and validation/domain
negative follow the foundation precedence. Workspace does not emit exit `1`
for an operational inability to commit.

Source coordinates use the foundation `DiagnosticSource` and explicit
`utf8-byte` or `unicode-scalar` ranges. Paths are normalized workspace-relative
in machine output. Secrets, proof bytes, raw idempotency keys, environment
values, and file content are redacted; digests, IDs, effect indexes, and
bounded observed/expected metadata remain.

## Engine.Execute and user experience

### Mutating flow

```text
Engine.Execute
  -> workspace.Snapshot(plan-input scope)
  -> artifact/workflow/completion owners evaluate immutable facts
  -> engine assembles exact MutationPlan
  -> authorization provider grants PlanDigest
  -> workspace.Apply(plan, content, grant, idempotency)
  -> engine maps ApplyOutcome to the foundation-owned Outcome
```

Workspace never calls an evaluator during apply and never substitutes a fresh
snapshot. `ApplyOutcome` is plain internal data: transaction ID/state,
plan/result digests, before/after revision refs, changed paths, journal event
IDs, replay status, recovery reference, durability facts, and warnings.

### Human recovery example

```text
This change is prepared but not fully committed.
Transaction tx-01JTEST... stopped after effect 2; its recorded files are preserved.
Next: sdd recover --transaction tx-01JTEST...
```

Default success output omits transaction phases. Detailed/JSON output may show
the transaction/recovery fields needed by agents.

### Public JSON compatibility

Foundation `1.0` remains closed. This design does **not** label workspace apply
or recovery payloads `1.1`. Until `AM-02`, `AM-03`, and `DEC-05` allocate one
shared extension, only existing foundation result variants may be emitted.
The provisional extension payload that foundation must eventually embed as a
discriminated result is:

```json
{
  "kind": "workspace-apply",
  "transactionId": "tx-01JTEST000000000000000001",
  "transactionState": "committed",
  "planDigest": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
  "resultDigest": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
  "replay": false,
  "changedPaths": [".sdd/work/example/requirements.md"],
  "beforeRevisions": [{"scope": "artifact", "ownerId": "artifact-17", "value": 7}],
  "afterRevisions": [{"scope": "artifact", "ownerId": "artifact-17", "value": 8}],
  "recovery": null
}
```

This is a candidate result object only, not a standalone envelope or accepted
schema. The eventual public schema must also cover prepared/quarantined/replay
results, structured next actions, and unknown-variant rejection while
preserving all foundation `1.0` fixtures.

## Platform and filesystem support

```go
type FilesystemCapabilities struct {
    Platform                   string
    FilesystemKind             string
    StableFileIdentity         CapabilityState
    NoFollowTraversal          CapabilityState
    AdvisoryInterprocessLocks  CapabilityState
    DurableFileData            CapabilityState
    DurableDirectoryEntries    CapabilityState
    AtomicSingleFileReplace    CapabilityState
    CaseComparison             CaseRule
    UnicodeComparison          UnicodeRule
    MaxComponentBytes          uint32
    MaxPathBytes               uint32
    NetworkOrSyncManaged       bool
}
```

`supported`, `unsupported`, and `unknown` are distinct. Mutation requires every
capability used by the selected effect/durability profile. Unknown is not
success. `doctor` reports facts read-only. Native certification requires the
same adapter contract and kill/fault suite on the claimed OS/filesystem;
cross-builds do not certify runtime semantics. NFS, SMB, FUSE/userspace, and
cloud-sync mounts are blocked or warned only according to the approved
`DEC-02` matrix; warning-only mode cannot claim guarantees it did not test.

Unix and Windows adapters may differ internally in root handles, no-follow
opening, advisory lock primitives, replacement behavior, sharing violations,
and directory sync. Shared domain logic never branches on `runtime.GOOS`.

## Limits

`Limits` names maxima for paths/components, read-set entries and bytes,
per-file bytes, effects, total staged bytes, transaction records, journal event
and replay bytes, capture retries, replacement retries, recovery steps, and
idempotency records. Limits and the selected limits-policy digest participate
in the plan. Exceeding a request limit fails before `PENDING`; a disk or stored
record exceeding a limit during an active prepared transaction preserves
evidence and returns recovery/quarantine.

No permanent numeric defaults are accepted here. Company/release policy must
publish the initial values and profiles. Tests exercise exactly-at and
one-byte/one-entry-over boundaries regardless of those values.

## Format migration

The only currently documented native workspace format is integer `1`.
Therefore no historical migration pair is invented. An unknown/newer format or
partial `.sdd` fails read-only with exit `3`. When a real pair is approved, the
migration owner supplies a complete exact plan including source format/root
observations, backup manifest and digests, destination version, staged outputs,
revision intents, recovery expectations, and authorization. Workspace applies
that plan through the same protocol. Legacy Kiro/OpenSpec/Spec Kit mapping and
interactive UX remain outside this module.

## Conformance fixture design

Every fixture records setup, command/module request, injected boundary,
expected typed outcome, exact authoritative mutations, and forbidden side
effects. IDs map as follows:

| Fixture IDs | Test layer | Observable assertion |
|---|---|---|
| `FX-WS-001`–`006` | module + native path adapters | stable identities, explicit selection, collision/escape/link safety |
| `FX-WS-007`–`010` | snapshot interface + capture fake | bounded read set, no reread, unstable/mixed generation rejected |
| `FX-WS-011`–`013` | apply/model tests | mutation table, tombstones, replay/no-op, stale writer |
| `FX-WS-014`–`016` | canonical plan/provider fake | order/content binding, dry-run purity, staging-name exclusion, stale grant |
| `FX-WS-017`–`020` | adapter contract + native OS | exact create/update/delete/provisional move, collision, held-open/disk/permission faults |
| `FX-WS-021`–`028` | restart fault harness | every pre/post-intent crash, recovery interruption, quarantine, owned cleanup, cancellation |
| `FX-WS-029`–`030` | engine snapshot/rebuild | incomplete commit is recovery-required; cache rebuild deterministic |
| `FX-WS-031`–`034` | interprocess idempotency tests | join, conflict, expiry, lost-response replay/disclosure |
| `FX-WS-035`–`036` | interprocess lock tests | timeout/no steal, owner proof, PID reuse, reentrancy/order |
| `FX-WS-037`–`039` | journal replay/compaction | torn/manual edit detected, checkpoint equivalence, clock-independent order |
| `FX-WS-040`–`042` | provider/format/migration harness | binding vs stale category, partial/new format, recoverable migration |
| `FX-WS-043`–`045` | engine/adapter/limit boundary | read-only inspection, unsupported FS, exact limit categories |
| `FX-WS-046`–`047` | snapshot identity projection | allocated/moved/tombstoned binding and missing-record integrity |
| `FX-WS-048`–`049` | directory-effect adapter/model tests | exact parent/absence/identity/emptiness preconditions and replay |

Pure tests cover path/canonical hash/revision/state-machine laws. Module tests
cross `Transactions` with deterministic adapters. Adapter contract tests run
against both fault fake and native implementation. Kill-process, Windows
held-open, directory durability, and network-filesystem claims require real
native integration evidence; fakes cannot certify them. State-machine model
tests generate apply/crash/recover/replay sequences and compare revisions,
events, generation, and visible outcomes after every step.

## Shared-contract ledger

| Contract | Concrete workspace output/input | Trust/freshness rule | Consumers |
|---|---|---|---|
| `IC-01` snapshot/revisions | workspace `RawSnapshot` extension (`Snapshot`), ordered raw/path/artifact-identity observations, generation/read-set digest, `RevisionRef` | foundation selects; workspace captures once; no reread under ID and no parallel identity store | engine, graph, artifact, validation, completion |
| `IC-02` plan/outcome | `PlanCore`, digest, effects, `ApplyOutcome`/`RecoveryOutcome` | exact plan only; raw and revision checks at apply | engine and all mutation-producing owners |
| `IC-03` idempotency | namespaced key, request fingerprint, durable record, retained result | disclosure checked on replay; action not rerun | engine/automation/guided journeys |
| `IC-10` policy/authority facts | opaque policy identity in plan plus verified provider fact | workspace checks binding/freshness only | engine/provider; graph/completion remain decision owners |
| `IC-12` completion plan | opaque completion-owned effects passed through `IC-02` | workspace cannot infer DONE from files | completion and guided workflows |

Downstream guided `GW-014` must consume `IC-02`/`IC-03` for atomic draft
submission; `GW-015` consumes committed work revisions but graph/completion own
transition eligibility; `GW-016` obtains completion effects from the completion
owner and submits them through `IC-02`. These guided tasks must not implement a
second transaction or idempotency algorithm.

## Upstream amendments and compatibility gates

| Amendment | Exact need | Compatibility test | Gate |
|---|---|---|---|
| `AM-02` / `DEC-05` | foundation-owned result/command extension inventory for apply/recover and guided commands | all closed `1.0` fixtures unchanged; negotiated unknown variants fail | public serialization publication |
| `AM-03` / foundation `FND-014` | add only approved move/metadata effect variants and richer preconditions/revision refs | native create/update/delete remain unchanged; unsupported extension non-authorizable | extension effects and public schema |
| `AM-01` | retain foundation severity including `critical`; do not introduce workspace severity enum | code/severity/blocking mapping fixture | diagnostic catalog publication |
| `AM-11` | document only format `1` until real migration pairs exist | unknown/newer/partial format fail, no invented migration | migration support claim |

Internal record schema IDs above are workspace-format details, not claims that
the public protocol version is `1.1` or that native format `2` exists.

## Open production/release gates

1. `DEC-01`: trusted provider/actor model, proof verification, and restrictive
   post-`PREPARED` revocation source. Fakes can test shape; production authority
   cannot be claimed.
2. `DEC-02`: supported OS/filesystem matrix, directory durability guarantees,
   and block-versus-warning policy for sync/network filesystems. Cross-builds
   are insufficient.
3. `DEC-03`: journal, committed summary, idempotency, audit, secret, and expiry
   retention/privacy. Compaction/destructive expiry is blocked.
4. `DEC-05` plus `AM-02`/`AM-03`: one public extension allocation and effect
   vocabulary. No workspace-specific `1.1` may be published independently.
5. Company-selected numeric limits and performance budgets. The design defines
   measurement and boundaries, not permanent values.

ADR-0006 resolves the transaction path, conservative hard-link behavior, and
post-intent expiry rule. Pure/module/fake and internal extension implementation
may proceed. Until the gates above are resolved, no task may claim protocol
publication, destructive retention, native durability certification,
production authorization, or public move/metadata support.

## References

- `requirements.md`
- `docs/kiro-input/prerequisite-review.md`
- `docs/kiro-input/next-prerequisite-specs.md`, sections 2–4, 6, and 10
- `.kiro/specs/sdd-engine-foundation/design.md`
- `.kiro/specs/sdd-engine-foundation/tasks.md`
- `docs/architecture/go-project-structure.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/cli-ux.md`
- `.kiro/steering/quality-and-safety.md`

## For the chosen ones

The durable commit fact is more fundamental than the transaction directory.
Staging is expendable after commit; the proof that a caller key already spent
its authority is not. That is why the committed summary, event IDs, planned
revision values, and request fingerprint survive cleanup as one retention
unit. Systems that retain the new files but discard that unit are crash-safe
only superficially: after a lost response they can no longer distinguish
replay from a second authorized mutation.
