# Requirements Document: sdd-workspace-transactions

## Introduction

This specification defines the deep workspace-transaction module behind the
`Engine.Execute` seam. It owns canonical native workspace identity and paths,
bounded raw snapshots, revision allocation, local locking, exact-plan
application, durable idempotency, journaling, and crash recovery. Its public
surface is intentionally small: discover/select a workspace, read one immutable
snapshot, and apply one already-authorized exact mutation plan.

The module does not decide workflow policy, validate typed artifacts, decide
completion, acquire execution claims, supervise processes, import legacy data,
or perform Git commits/merges. It consumes opaque versioned policy and
authorization facts and typed effects. The engine coordinates facts between
modules; this module commits only the effects it receives.

Status: **Accepted for implementation on 2026-10-17** by
`docs/decisions/0006-workspace-transaction-contract.md`. Acceptance fixes the
module invariants; it does not claim implementation, native-platform
certification, production authority-provider approval, destructive-retention
approval, or publication of the allocated protocol extension.

## Glossary and contract boundary

- **Workspace ID** is an immutable ID persisted in the native workspace. The
  selected repository/project path is a mutable coordinate, not identity.
- **Work ID** is immutable; a work slug and path are mutable aliases. Project
  selection and work selection are separate coordinates and both are explicit.
- **Artifact identity** is owned by artifact-validation. Workspace stores and
  compares opaque bytes, revisions, and digests supplied by its caller; it does
  not parse artifact semantics or reimplement semantic digest rules.
- **Snapshot** is one immutable, bounded read set of raw bytes and filesystem
  observations. Evaluation never silently reopens a path after the snapshot.
- **MutationPlan** is an exact, ordered list of foundation-compatible effects,
  preconditions, source observations, authorization binding, limits, and a
  canonical plan digest. A plan is an intention, not permission by itself.
- **Transaction** is a local recoverable application of one exact plan. The
  durable `PREPARED` record is the commit-intent point; it is not a promise of
  cross-file filesystem atomicity.
- **Caller idempotency key** is an opaque request key supplied by the caller.
  It is not a digest and is never replaced by a generated fingerprint.
- **Request fingerprint** is a canonical digest of the complete request and
  relevant authorization scope. It detects changed payloads under one key.
- **Quarantine** means that recovery cannot safely satisfy a recorded
  precondition. It preserves evidence and requires an explicit authorized
  resolution; it is not an invitation to overwrite user files.

### Owned interfaces and consumed contracts

| Interface | Workspace obligation | Owner outside this module |
|---|---|---|
| Native envelope, diagnostics, `exitCode` (`0/1/2/3/4/5/6/130`) | Return foundation-compatible typed outcomes and stable diagnostic codes | `sdd-engine-foundation` |
| Raw `Snapshot` and selection coordinates | Provide immutable bytes, path/file observations, revisions, and bounded read-set identity | Workspace (this spec); engine coordinates |
| Artifact semantic/derivation digests | Store and compare supplied opaque values; never invent semantic meaning | `sdd-artifact-validation` |
| Workflow/completion decisions | Accept opaque policy/authorization facts and effects only | graph and completion specs |
| Claims, leases, process trees, worktrees | Do not acquire, renew, adopt, or expire execution claims | `sdd-bounded-execution` |
| Legacy import journeys | Offer native migration transaction mechanics only | migration/UX spec |

## Requirements

### WS-REQ-001 — Identity, selection, and containment

1. The workspace SHALL persist an immutable `workspaceId`, native format
   version, and canonical native root; a path or project selection SHALL never
   be treated as identity.
2. Work selection SHALL carry an explicit workspace root and either immutable
   `workId` or mutable slug. An explicit invalid root or ID SHALL fail without
   parent fallback; nested workspaces SHALL be discoverable but never selected
   implicitly across a nested boundary.
3. The public workspace interface SHALL represent repository-relative paths in
   normalized POSIX form (`/` separators, no leading `/`), while the filesystem
   adapter MAY translate them to native separators, handles, case rules, and
   Unicode representation. The adapter SHALL retain the native path/file
   identity needed for containment and diagnostics; native spelling SHALL not
   leak into plan identity. `.`/`..`, empty components, separators inside a
   component, platform reserved names, and platform path-length violations
   SHALL be rejected with a source path and a foundation-compatible diagnostic.
4. Before planning and again immediately before every effect, the adapter SHALL
   resolve every existing path component without following a final write target
   into an unowned object, then prove containment under the selected workspace
   root. A symlink or reparse point that resolves outside the root SHALL cause
   rejection; a target swap between planning and apply SHALL be detected and
   cause conflict/quarantine rather than retargeting the effect.
5. Case-fold and Unicode-normalization collisions SHALL be detected using the
   declared filesystem comparison rules. Distinct names that collide SHALL
   fail before application; canonicalization SHALL not rewrite meaningful file
   bytes or silently choose one colliding name.
6. For a managed regular file, an effect that could mutate a hard-linked inode
   SHALL either use an atomic replacement that leaves the old inode untouched
   or reject the effect with an explicit unsupported/unsafe-link diagnostic.
   The module SHALL not assume that a hard link's apparent path proves that all
   aliases are inside the workspace. Hard links are recorded in observations
   when the platform exposes file identity.
7. Containment SHALL be checked at effect time even when a plan was authorized
   earlier. Checks SHALL include destination paths for moves and archive-like
   destinations; the module SHALL never write through a path merely because it
   was contained in a stale plan.

Acceptance criteria: `WS-REQ-001.AC-01` distinguishes workspace ID, project
root, work ID, and slug across rename and nested selection;
`WS-REQ-001.AC-02` rejects symlink/reparse escape and a swapped target at
apply; `WS-REQ-001.AC-03` rejects case, Unicode, reserved-name, separator,
length, and root-escape collisions; `WS-REQ-001.AC-04` covers a hard-linked
managed file without mutating its outside alias.

### WS-REQ-002 — Immutable bounded raw snapshots

1. A snapshot SHALL declare its scope, ordered read set, snapshot ID, native
   format/workspace/work revision observations, bounded limits, and per-path
   raw bytes or an explicit absent observation. It SHALL include file type,
   size, mode/permission facts needed by an effect, platform file identity when
   available, and a blob digest of bytes. A timestamp or mtime alone is never a
   snapshot identity.
2. The read-set builder SHALL be explicit and bounded; an operation SHALL not
   silently read the entire repository when its contract names a narrower
   scope. Selection metadata, managed files, journal records, and policy bytes
   SHALL be included only when the command contract requires them.
3. The adapter SHALL use a stable-observation procedure (open/read/stat before
   and after, or a platform equivalent) and SHALL retry within a declared
   bound. If bytes, file identity, size, or relevant metadata change during a
   read, it SHALL return `DIAG_SNAPSHOT_UNSTABLE` with exit code `4` when the
   command cannot safely continue. It SHALL not label a newer reread as the
   original snapshot.
4. A multi-file snapshot SHALL capture a workspace transaction-generation
   marker while holding the read-side snapshot lock, or use an equivalent
   before/after generation validation supplied by the transaction adapter. If a
   writer changes generation during capture, the adapter SHALL retry within its
   bound; if it cannot prove one coherent generation, it SHALL return
   `DIAG_SNAPSHOT_UNSTABLE`/recovery-required and SHALL not expose a mixed set
   as immutable. A per-file stable read alone is insufficient.
5. Evaluation SHALL receive the accepted immutable snapshot value and SHALL not
   reopen files, consult mtime, or refresh policy during evaluation. If the
   caller needs a new observation it must request a new snapshot and new plan.
6. Workspace SHALL record opaque semantic/derivation digest fields only when a
   caller supplies them; parsing, normalization, validator identity, and
   semantic-field selection remain outside this module.
7. A plan SHALL retain the source snapshot ID/read-set digest and expected raw
   observations for every path it may affect. Apply SHALL detect an external
   change, deletion, or newly appearing path and report a conflict rather than
   applying against an unobserved state.
8. When the requested scope includes artifact identities, the same immutable
   snapshot SHALL expose workspace-owned allocation bindings or tombstones with
   `(workspaceId, workId, namespace, artifactId)`, allocation revision and
   provenance, current artifact revision/path, and removal provenance as
   applicable. A claimed persisted ID without an allocation binding SHALL be a
   typed `missing-record` integrity observation. Draft proposals are not
   persisted observations and SHALL NOT be represented by a zero revision.

Acceptance criteria: `WS-REQ-002.AC-01` proves no evaluation reread;
`WS-REQ-002.AC-02` returns a bounded unstable-snapshot diagnostic;
`WS-REQ-002.AC-03` proves narrow read-set selection;
`WS-REQ-002.AC-04` detects an edit after authorization using raw observations
without a second artifact parser; `WS-REQ-002.AC-05` injects a concurrent
transaction-generation change and proves that no mixed-generation snapshot is
accepted; `WS-REQ-002.AC-06` proves allocated/tombstoned/missing artifact
identity observations arise from the same coherent raw snapshot.

### WS-REQ-003 — Scoped revision allocation and mutation table

1. Revisions SHALL be monotonic nonnegative integers (or an explicitly
   versioned compatible representation) allocated under the transaction lock.
   A revision SHALL carry scope, owner identity, format/schema version, and
   allocation provenance; wall-clock time and mtime SHALL never substitute for
   a revision. The public foundation integer revision fields remain intact.
2. The following mutation table is normative. A scope advances only when the
   stated authoritative state changes; a no-op does not advance it.

| Mutation | Workspace revision | Work revision | Artifact revision | Required observation |
|---|---:|---:|---:|---|
| Create workspace/format migration | advance | n/a | n/a | format and root state |
| Create work | advance | allocate initial | n/a | new immutable work ID |
| Edit/submit artifact bytes or metadata | no | no, unless work record changes | advance | old blob/revision plus complete current work read set |
| Rename/move artifact or work slug | no, unless workspace index changes | advance for work slug | advance for artifact path | old/new path and identity |
| Delete artifact/work | no, unless workspace index changes | advance for work deletion | advance tombstone | identity tombstone |
| Approve/accept/reject/stale lifecycle record | no | advance | advance only if artifact bytes/state are authoritative | prior revision |
| Compose/escalate work | no | advance | no unless artifacts are changed | composition digest |
| Graph/completion transition or receipt | no | advance | no unless artifact bytes change | accepted transition effect |
| Journal-only audit/administrative observation | no | no | no | event id and actor |
| Projection/cache rebuild | no | no | no | source journal/snapshot |
| Exact replay of already committed plan | no additional advance | no additional advance | no additional advance | matching idempotency record |
| External edit observed outside control plane | no implicit allocation | no implicit allocation | no implicit allocation | new raw snapshot; mutation must reconcile |

3. A workspace-level revision advances only for an authoritative workspace
   membership/format mutation, not for every file write. A work revision
   advances for authoritative work metadata, lifecycle, composition, or work
   path changes. An artifact revision advances for its authoritative content,
   metadata, path, or lifecycle record. The exact effect SHALL identify which
   scopes it advances.
   An artifact submission that intentionally leaves the work revision unchanged
   SHALL still read and guard the complete current work record, composition,
   lifecycle, and relevant artifact membership at the expected work revision.
   A concurrent work-level change then fails as `DIAG_STALE_REVISION`; the
   no-increment rule is not permission to apply against a partial read set or
   lose a work-level update. If submission changes a work record, its work
   revision advances in the same transaction.
4. A stale expected revision SHALL fail with `DIAG_STALE_REVISION`, exit code
   `4`, before durable commit intent. Deleted identities SHALL retain a
   tombstone; reusing a path or slug SHALL not reuse the old identity or
   revision.
5. Revision allocation SHALL be transactional with the authoritative effects.
   A failed pre-PREPARED transaction SHALL not publish allocated revisions; a
   committed transaction SHALL publish all scope revisions together with its
   receipt. A replay returns the original revisions and does not increment.
6. Administrative event IDs, plan IDs, idempotency keys, and candidate/content
   digests SHALL remain distinct from revision tokens. Revision equality does
   not prove that implementation bytes are identical.

Acceptance criteria: `WS-REQ-003.AC-01` exercises every mutation-table row;
`WS-REQ-003.AC-02` rejects two writers from one revision;
`WS-REQ-003.AC-03` proves no-op/replay does not advance;
`WS-REQ-003.AC-04` preserves a deleted identity as a tombstone.

### WS-REQ-004 — Exact plan identity and authorization binding

1. A plan SHALL contain a versioned plan descriptor, workspace/work identity,
   source snapshot/read-set digest, expected revisions and raw preconditions,
   ordered foundation effect kinds (`create-file`, `update-file`,
   `delete-file`; versioned extensions may add `create-directory`,
   `remove-empty-directory`, `move-file`, or `metadata-update` only through an
   explicit amendment), the non-circular
   authorization binding fields (principal/scope/policy identity and version,
   but not the final grant/signature), size limits, and a canonical plan digest.
2. Plan identity SHALL be computed from canonical typed plan data, including
   effect order, paths, expected absent/old bytes or digests, content bytes or
   content digests, source snapshot, authorization scope and policy identity/
   version, and extension version. It SHALL exclude the final authorization
   grant/signature and any grant field that points back to the plan digest; the
   grant is issued after the plan digest is known and binds that resulting
   digest. Temporary staging paths, journal offsets, lock nonce, and generated
   transaction IDs SHALL also be excluded from semantic plan identity.
3. Apply SHALL validate the exact plan, authorization principal/scope/plan
   binding/policy digest/expiry/revocation, all expected revisions and raw
   preconditions, path containment, and size limits immediately before durable
   commit intent. It SHALL either apply those effects or reject; it SHALL never
   re-plan or silently widen the read set.
4. A dry-run SHALL be observational and SHALL return the same authorizable
   plan shape without creating staging, journals, locks retained beyond the
   read, revisions, or authoritative files. Unsupported effects SHALL be
   non-authorizable with `DIAG_PLAN_NOT_AUTHORIZABLE` rather than represented as
   executable effects.
5. Plan lifetimes SHALL be explicit. Expired authorization or plan freshness
   SHALL fail closed before `PREPARED`; the caller must create a new snapshot
   and plan. An uncertain result after `PREPARED` returns recovery state and
   the recorded transaction, never a guessed success.

Acceptance criteria: `WS-REQ-004.AC-01` proves exact-plan mismatch is rejected;
`WS-REQ-004.AC-02` proves dry-run has no persistent side effect;
`WS-REQ-004.AC-03` rejects stale policy/expired grant;
`WS-REQ-004.AC-04` proves temporary staging names do not change plan identity.

### WS-REQ-005 — Safe effect application and platform durability

1. Foundation-native create/update/delete effects SHALL enforce expected
   absence or expected old raw digest/revision. A create SHALL fail if the
   target appears; update/delete SHALL fail if it is absent or changed. An
   extension effect SHALL define equivalent preconditions before it is
   supported. Existing permissions and unrelated user files SHALL be preserved.
2. A move/rename extension SHALL be path-bounded, source-identity checked, and
   destination-collision checked. It SHALL not be described as a multi-file
   atomic rename. Archive collision is a conflict unless an exact authorized
   unique destination is part of the plan.
3. Staging writes SHALL be bounded, flushed according to the selected local
   filesystem adapter, and installed using the platform's atomic single-file
   replacement primitive where available. Directory creation/removal and
   journal durability SHALL have documented order and failure results.
4. The module SHALL report whether the configured adapter supports durable
   file data, durable directory entries, atomic single-file replacement, and
   advisory locks. It SHALL not promise generic multi-file atomicity. A held
   open file, permission loss, or disk-full condition SHALL stop before the
   next unsafe effect, preserve evidence, and produce a retryable or manual
   recovery diagnostic.
5. Effect-time checks SHALL repeat containment and expected file identity after
   opening the target and before replacement. A symlink/hard-link target swap
   SHALL be a conflict, not an opportunity to follow a new target.

Acceptance criteria: `WS-REQ-005.AC-01` checks old-content/nonexistence
preconditions; `WS-REQ-005.AC-02` covers permission loss/disk exhaustion
midway; `WS-REQ-005.AC-03` covers held-open Windows behavior;
`WS-REQ-005.AC-04` records actual platform durability and does not claim
cross-file atomicity.

### WS-REQ-006 — Transaction states, durable recovery, and quarantine

1. Legal states SHALL be `PENDING`, `STAGING`, `PREPARED`, `COMMITTED`,
   `FAILED`, and `QUARANTINED`. Durable records SHALL include transaction ID,
   plan digest, source observations, owned staging manifest, intended effects,
   state, journal sequence/checksum, revisions, and idempotency reference.
2. The state transitions SHALL be:

| From | Event | Result |
|---|---|---|
| absent | begin | `PENDING` record |
| `PENDING` | lock/validate/stage begins | `STAGING` |
| `STAGING` | all staged bytes and intent record durably flushed | `PREPARED` |
| `STAGING` | cancellation/fault before intent | `FAILED`, owned cleanup only |
| `PREPARED` | effect preconditions still hold and effects install | `COMMITTED` |
| `PREPARED` | conflicting external change | `QUARANTINED`, preserve intent/staging |
| any nonterminal | corrupt/torn/missing required evidence | `QUARANTINED` unless safe pre-PREPARED cleanup is proven |
| `COMMITTED` | acknowledgment/replay | same `COMMITTED` result |

3. Before `PREPARED`, cleanup MAY remove only files and journal records named
   in the transaction's durable owned-staging manifest and created by that
   transaction. It SHALL not glob all staging, delete unknown journal entries,
   or remove a user file that happens to share a directory. If ownership or
   record completeness cannot be proved, cleanup stops and quarantines.
4. At `PREPARED`, the module SHALL durably record commit intent, exact effects,
   preconditions, staged content references, and revision/idempotency data
   before installing the first authoritative effect. After that point,
   cancellation does not roll back or discard intent. Recovery SHALL roll
   forward idempotently from recorded bytes, or quarantine on a violated
   precondition; it SHALL never guess, partially roll back, or overwrite an
   unrelated changed file.
5. Recovery SHALL be safe to interrupt and rerun. Missing staging, torn
   records, manual journal edits, disk-full cleanup failure, and a committed
   transaction whose response was lost SHALL produce inspectable diagnostics.
   After a transaction is committed, staging bytes MAY be cleaned lazily only
   after a durable committed summary retains the plan digest, final revisions,
   event/result digest, idempotency key/fingerprint, and replay-retention
   expiry. Cleanup SHALL never remove the durable fact needed to distinguish a
   committed replay from a new request. A committed transaction returns its
   stored outcome on replay. A recovery action is explicit and authorized;
   read-only `status` and `doctor` report pending/quarantined state without
   repairing, appending events, or deleting material.
6. Cancellation before durable commit intent returns `130` only after owned
   cleanup completes; cancellation after intent returns an interruption or
   recovery outcome while preserving the prepared transaction. Authorization
   expiry/revocation is checked before `PREPARED`. The default recovery rule is
   to roll forward the already-authorized exact intent without obtaining a new
   permission or re-planning. If a newly observed restrictive policy or
   revocation rule forbids that roll-forward, recovery SHALL preserve the
   prepared evidence and enter an inspectable `QUARANTINED` state; it SHALL not
   silently acquire a replacement grant, widen scope, or substitute a new plan.

Acceptance criteria: `WS-REQ-006.AC-01` covers crash before PREPARED, during
each effect, after commit before reply, and recovery interrupted again;
`WS-REQ-006.AC-02` preserves an unrelated staging file;
`WS-REQ-006.AC-03` quarantines missing staging/external conflict;
`WS-REQ-006.AC-04` proves read-only status is non-mutating;
`WS-REQ-006.AC-05` covers cancellation after durable commit intent.

### WS-REQ-007 — Multi-file visibility and authoritative state

1. Engine readers SHALL observe either the prior committed authoritative state,
   the new committed state, or an explicit recovery-required state. They shall
   not interpret an incomplete transaction directory or projection as a valid
   domain state.
2. Ordinary editors may see intermediate single-file replacements. The module
   SHALL document this window and provide transaction/recovery status; it SHALL
   not promise cross-file filesystem atomicity it cannot implement.
3. Canonical authoritative state SHALL be native workspace files plus durable
   journal events and committed transaction summaries. The summary retained
   after lazy staging cleanup is the canonical replay/idempotency fact for that
   transaction; its retention period is explicit. Caches, SQLite indexes,
   selections, and projections SHALL be disposable and rebuildable. A deleted
   cache or unavailable index SHALL not change correctness or mutate state during
   a read-only command.
4. Rebuilding a projection from the same canonical snapshot/journal SHALL
   produce deterministic queryable facts. Projection writes are not additional
   domain mutations and do not advance workspace/work/artifact revisions.

Acceptance criteria: `WS-REQ-007.AC-01` observes an incomplete multi-file
transaction as recovery-required; `WS-REQ-007.AC-02` deletes cache/index and
rebuilds the same facts; `WS-REQ-007.AC-03` verifies read-only status has no
side effects.

### WS-REQ-008 — Durable idempotency and replay

1. A caller SHALL provide an opaque idempotency key in a declared namespace.
The namespace SHALL include workspace ID, command/effect kind, and caller
principal or equivalent scope. The key is stored separately from:
   (a) a canonical request fingerprint over command, plan, payload, expected
   revisions, authorization identity, and relevant expiry metadata; and
   (b) retention/created/expiry metadata.
2. A retry with the same namespace/key and identical fingerprint SHALL return
   the original committed, failed, quarantined, or in-flight outcome according
   to the state. Before disclosing that stored outcome, the module SHALL check
   that the current caller is still authorized to view the namespace/result;
   this is a disclosure check, not a rerun of the original action or a fresh
   evaluation of its expired grant. A changed payload, plan, scope, principal,
   or authorization binding under the same key SHALL return a hard conflict
   before effects.
3. A concurrent duplicate SHALL join or report the existing request record; it
   SHALL not universally be rejected as a fresh duplicate and SHALL not execute
   effects twice. A lost response after commit is resolved from the durable
   result after the current-caller disclosure check; it does not rerun
   authorization or the action. An in-flight or recovering record exposes a
   stable recovery action subject to the same disclosure check.
4. Retention SHALL preserve enough key, fingerprint, state, plan digest, and
   result data to disambiguate replay for the declared period. After expiry,
   reuse SHALL either be rejected as `DIAG_IDEMPOTENCY_EXPIRED` or require an
   explicitly documented new namespace; it SHALL never silently map to a new
   operation under the old key.
5. Idempotency provides at-most-once application for transactions owned by this
   local module. It does not promise exactly-once behavior for external tools,
   network effects, Git, or process execution.

Acceptance criteria: `WS-REQ-008.AC-01` covers duplicate while first request is
running; `WS-REQ-008.AC-02` rejects changed payload under one key;
`WS-REQ-008.AC-03` replays after lost response; `WS-REQ-008.AC-04` covers
retention expiry and changed retry time.

### WS-REQ-009 — Local lock ownership and ordering

1. Protected resources SHALL include the workspace root, affected work records,
   journal/transaction directory, and effect paths as needed. Locks SHALL be
   acquired in deterministic canonical order and released in reverse order.
2. Acquisition timeout is only the caller's wait limit. Reaching that timeout
   SHALL return a conflict; it SHALL never release, steal, or declare dead a
   lock whose owner may still be active.
3. Lock metadata SHALL contain an owner nonce, process/session identity,
   creation time, resource identity, and heartbeat/last-observed liveness when
   supported. Release SHALL require the nonce/owner proof. PID reuse alone is
   insufficient. Same-process reentrancy SHALL be reference-counted by the
   owner nonce, not by a path-only global.
4. Crash recovery MAY release a lock only after the local adapter proves the
   owner is gone and the lock record/OS primitive permits safe replacement, or
   after an explicit recovery command with the required authority. An active
   owner is never released because a wall-clock deadline elapsed.
5. These locks coordinate local transactions only. They are not distributed
   leases, action claims, process-execution locks, or a scheduling mechanism.
   Unsupported/network filesystems SHALL be rejected or explicitly warned by
   `doctor` until their lock, flush, rename, and identity semantics are
   demonstrated by the supported fault suite.

Acceptance criteria: `WS-REQ-009.AC-01` covers lock timeout with an active
owner; `WS-REQ-009.AC-02` covers crash release and PID reuse;
`WS-REQ-009.AC-03` covers reentrancy and canonical lock ordering;
`WS-REQ-009.AC-04` rejects an unsupported network filesystem.

### WS-REQ-010 — Journal integrity and rebuildability

1. Each journal event SHALL have a unique event ID/namespace, schema version,
   transaction ID, monotonic sequence within the journal, event type, UTC
   timestamp for audit only, affected identities, payload/effect reference,
   previous checksum, and checksum. The checksum detects accidental corruption;
   it is not tamper proof against a user controlling the filesystem.
2. Append SHALL flush according to the selected durability contract. A torn
   record, truncation, sequence gap, checksum mismatch, unknown required
   schema, or manual edit SHALL stop replay and produce an inspectable
   corruption/recovery diagnostic; the module SHALL not invent missing events.
3. Replay SHALL be deterministic and idempotent. Compaction SHALL write a
   checkpoint containing equivalent canonical state and retain enough event and
   transaction references for audit/recovery under the declared retention
   policy. It SHALL not discard a committed result needed for idempotency.
4. Secrets and authorization tokens SHALL be redacted before journal or
   diagnostic persistence. Redaction SHALL not remove hashes, IDs, or facts
   required to prove a plan/recovery relationship.

Acceptance criteria: `WS-REQ-010.AC-01` detects corrupt/torn journal and
manual edits; `WS-REQ-010.AC-02` replays after compaction without changing
facts; `WS-REQ-010.AC-03` verifies secret redaction and retained outcome
references.

### WS-REQ-011 — Authorization seam and initialization

1. An authorization fact SHALL identify trusted principal/source, scope,
   exact plan digest, policy snapshot digest/version, issue and expiry times,
   revocation state, and any required nonce. Workspace validates binding and
   freshness at apply; it does not distribute organization policy or decide
   completion/reviewer authority.
2. Missing, expired, revoked, unverifiable, or plan-mismatched authorization
   SHALL fail closed with the foundation policy/permission category (exit `6`)
   before `PREPARED`. A stale expected revision or changed raw precondition is
   a conflict (exit `4`), not a policy failure.
3. Missing `.sdd`, partial native layout, malformed records, or newer/unknown
   workspace format SHALL produce foundation-compatible workspace/configuration
   diagnostics and exit `3`; they SHALL not trigger implicit initialization or
   parent fallback.
4. Native migration SHALL be planned, authorized, and applied through the same
   recoverable transaction protocol, with a backup/recovery expectation declared
   by the format version. Legacy import UX and mapping policy remain external.

Acceptance criteria: `WS-REQ-011.AC-01` rejects stale/expired/revoked grant;
`WS-REQ-011.AC-02` distinguishes policy denial from stale revision;
`WS-REQ-011.AC-03` covers unsupported workspace version and partial `.sdd`;
`WS-REQ-011.AC-04` recovers a native migration fault.

### WS-REQ-012 — Limits, diagnostics, and filesystem support

1. Snapshot read-set, file, plan, effect, journal, staging, recovery, and
   idempotency-record sizes SHALL have declared bounded limits. Exceeding a
   limit SHALL fail before the next irreversible boundary with a stable code,
   remediation, and retryability.
2. Diagnostics SHALL use foundation's existing taxonomy: usage/malformed input
   `2`, workspace/configuration/schema `3`, conflict/stale `4`, external
   tool/process `5`, policy/permission `6`,
   interruption `130`, and validation/verification `1` only when the operation
   completed and the condition was negative. The module SHALL not create a
   competing exit-code table or envelope.
3. The supported-filesystem policy SHALL be explicit per platform and adapter.
   Local filesystems are supported only when lock, file identity, flush,
   directory durability, rename, case behavior, and fault tests pass. NFS,
   SMB, cloud-sync/userspace mounts, and other unsupported network filesystems
   SHALL be blocked or warned according to the declared policy; the module
   SHALL not silently claim local guarantees.
4. `status`/`doctor`/dry-run are read-only with respect to authoritative files,
   revisions, events, cache repair, and transaction cleanup. They may inspect
   and report a recovery-required state; only an explicit recovery command may
   mutate transaction material.

Acceptance criteria: `WS-REQ-012.AC-01` exhausts snapshot/plan/journal limits;
`WS-REQ-012.AC-02` verifies foundation exit codes;
`WS-REQ-012.AC-03` covers unsupported filesystem;
`WS-REQ-012.AC-04` proves read-only status during a pending transaction is
non-mutating.

## Required conformance fixtures and coverage matrix

Fixture IDs are stable and are not claims that the fixture has been executed.
Each fixture records preconditions, input/action, expected observable outcome,
authoritative state changes, and forbidden side effects in the implementation
test plan.

| Handoff mandatory topic/case | Requirement/criterion | Fixture IDs |
|---|---|---|
| Identity, nested selection, immutable IDs/slugs | WS-REQ-001.AC-01 | FX-WS-001, FX-WS-002 |
| Symlink/reparse/root escape and effect-time check | WS-REQ-001.AC-02 | FX-WS-003, FX-WS-004 |
| Hard links, case/Unicode/reserved names/separators/length | WS-REQ-001.AC-03, WS-REQ-001.AC-04 | FX-WS-005, FX-WS-006 |
| Bounded immutable snapshot/read set | WS-REQ-002.AC-01, WS-REQ-002.AC-03 | FX-WS-007, FX-WS-008 |
| Unstable read/external edit after authorization and mixed transaction generation | WS-REQ-002.AC-02, WS-REQ-002.AC-04, WS-REQ-002.AC-05 | FX-WS-009, FX-WS-010 |
| Revision scopes, deleted identities, no-op, external edit | WS-REQ-003.AC-01, WS-REQ-003.AC-03, WS-REQ-003.AC-04 | FX-WS-011, FX-WS-012 |
| Two stale writers and changed expected digest | WS-REQ-003.AC-02, WS-REQ-005.AC-01 | FX-WS-013 |
| Exact plan, ordering, identity, dry-run | WS-REQ-004.AC-01, WS-REQ-004.AC-02, WS-REQ-004.AC-04 | FX-WS-014, FX-WS-015 |
| Stale policy/expired grant | WS-REQ-004.AC-03, WS-REQ-011.AC-01 | FX-WS-016 |
| Safe create/update/delete/move and archive collision | WS-REQ-005.AC-01 | FX-WS-017, FX-WS-018 |
| Held-open file, permission loss, disk exhaustion | WS-REQ-005.AC-02, WS-REQ-005.AC-03 | FX-WS-019, FX-WS-020 |
| Native durability/cross-file limits are reported, not overstated | WS-REQ-005.AC-04 | FX-WS-029, FX-WS-044 |
| Crash before PREPARED and during each effect | WS-REQ-006.AC-01 | FX-WS-021, FX-WS-022 |
| Crash after commit before reply and recovery interruption | WS-REQ-006.AC-01 | FX-WS-023, FX-WS-024 |
| Missing staging, torn record, conflicting external change | WS-REQ-006.AC-03 | FX-WS-025, FX-WS-026 |
| Read-only recovery inspection is non-mutating | WS-REQ-006.AC-04, WS-REQ-007.AC-03 | FX-WS-043 |
| Unrelated user file in staging and cleanup failure | WS-REQ-006.AC-02 | FX-WS-027 |
| Cancellation after durable commit intent | WS-REQ-006.AC-05 | FX-WS-028 |
| Multi-file visibility and cache/index deletion | WS-REQ-007.AC-01, WS-REQ-007.AC-02 | FX-WS-029, FX-WS-030 |
| Duplicate while first runs; changed payload; retention expiry | WS-REQ-008.AC-01, WS-REQ-008.AC-02, WS-REQ-008.AC-04 | FX-WS-031, FX-WS-032, FX-WS-033 |
| Lost response / committed-result replay | WS-REQ-008.AC-03 | FX-WS-034 |
| Lock timeout, crash release, PID reuse, reentrancy | WS-REQ-009.AC-01, WS-REQ-009.AC-02, WS-REQ-009.AC-03 | FX-WS-035, FX-WS-036 |
| Unsupported network/cloud-sync coordination semantics | WS-REQ-009.AC-04 | FX-WS-044 |
| Corrupt/torn journal, manual edit, compaction | WS-REQ-010.AC-01, WS-REQ-010.AC-02 | FX-WS-037, FX-WS-038 |
| Redacted journal/outcome disclosure retains non-secret result reference | WS-REQ-010.AC-03 | FX-WS-037, FX-WS-038 |
| Clock moves backward and journal logical sequence | WS-REQ-010.AC-01 | FX-WS-039 |
| Authorization binding/revocation and unsupported version | WS-REQ-011.AC-01, WS-REQ-011.AC-02, WS-REQ-011.AC-03 | FX-WS-040, FX-WS-041 |
| Native migration recovery | WS-REQ-011.AC-04 | FX-WS-042 |
| Read-only status/doctor and unsupported filesystem | WS-REQ-012.AC-03, WS-REQ-012.AC-04 | FX-WS-043, FX-WS-044 |
| Size exhaustion and foundation exit categories | WS-REQ-012.AC-01, WS-REQ-012.AC-02 | FX-WS-045 |
| Durable artifact allocation, move, tombstone, and missing-record observation | WS-REQ-002.AC-06, WS-REQ-003.AC-01, WS-REQ-003.AC-04 | FX-WS-046, FX-WS-047 |
| Create/remove directory preconditions and replay | WS-REQ-004.AC-01, WS-REQ-005.AC-01, WS-REQ-006.AC-01 | FX-WS-048, FX-WS-049 |

### Fixture behavior catalog

The matrix above names coverage; this catalog makes every fixture's contract
observable. These are fixture definitions, not executed-test claims.

| Fixture IDs | Preconditions | Action/input | Expected result | Forbidden mutation |
|---|---|---|---|---|
| FX-WS-001, FX-WS-002 | Two nested native roots; one work whose slug is renamed | Select by explicit root/work ID, then by old/new slug | ID remains stable; invalid explicit root/ID fails without parent fallback | No parent fallback, identity replacement, or unrelated work change |
| FX-WS-003, FX-WS-004 | Target symlink/reparse component points outside; target is swapped after plan | Plan then replace/swap link before apply | Containment conflict/recovery diagnostic; no outside write | No traversal or write through the swapped target |
| FX-WS-005, FX-WS-006 | Hard-linked managed file; colliding case/Unicode/reserved/long names | Plan mutation or create each invalid name | Safe replacement or explicit unsafe-link rejection; invalid names fail | No outside inode mutation or name rewriting |
| FX-WS-007, FX-WS-008 | Workspace has files outside requested scope | Request bounded snapshot with declared read set | Snapshot contains only required ordered observations and limits | No silent full-repository read or authoritative write |
| FX-WS-009, FX-WS-010 | File/transaction generation changes during read; authorized plan has changed file | Inject read race, generation change, or edit after authorization | Unstable/mixed-generation snapshot or stale conflict with retry guidance | No newer reread presented as original; no mixed snapshot or effect applied |
| FX-WS-011, FX-WS-012 | Workspace/work/artifact records at known revisions, including tombstone | Apply create/edit/delete/no-op/external-observation sequence | Mutation table scopes advance exactly; tombstone remains; no-op does not | No timestamp-only revision or identity reuse |
| FX-WS-013 | Two plans share one expected work/artifact revision | Apply writer A, then writer B | A commits; B returns `DIAG_STALE_REVISION`, exit `4` | No partial B effect or revision increment |
| FX-WS-014, FX-WS-015 | Authorized plan has ordered effects and exact preconditions | Change order/staging name or invoke dry-run | Mismatch rejects; staging name does not alter digest; dry-run is side-effect free | No silent re-plan, staging residue, lock retention, or revision change |
| FX-WS-016 | Grant is expired/revoked or policy digest differs before intent | Apply exact plan | Closed policy/permission result, exit `6` | No staging commit, effect, or replacement grant |
| FX-WS-017, FX-WS-018 | Create/update/delete/move targets with expected old/absent state; archive destination exists | Apply effects and collision case | Exact preconditions enforce safe effect; destination collision is conflict | No overwrite of changed source/destination or unrelated file |
| FX-WS-019, FX-WS-020 | Held-open/permission-denied target; fault injector exhausts disk midway | Apply bounded retry plan | Retryable/manual diagnostic with evidence; no unsafe next effect | No truncated authoritative file or unowned orphan cleanup |
| FX-WS-021, FX-WS-022 | Fault boundary before PREPARED and at each effect boundary | Crash/restart transaction | Owned pre-PREPARED material only is cleaned; prepared recovery resumes/quarantines | No unknown staging/journal deletion or guessed rollback |
| FX-WS-023, FX-WS-024 | Effects committed but reply lost; recovery itself is interrupted | Restart/retry recovery | Stored committed result or resumable recovery state; rerun is idempotent | No duplicate effects or fabricated success |
| FX-WS-025, FX-WS-026 | Prepared record lacks staged bytes or target changed externally | Run explicit recovery | Quarantine with exact evidence/manual action | No destructive repair, re-plan, or overwrite |
| FX-WS-027 | Transaction staging contains an unrelated user file | Fail before PREPARED and recover | Only manifest-owned material is considered; unknown file preserved | No glob deletion or user-file removal |
| FX-WS-028 | Durable PREPARED intent exists and caller cancels | Cancel/restart recovery | Prepared record remains and returns interruption/recovery state | No rollback, intent deletion, or new grant acquisition |
| FX-WS-029, FX-WS-030 | Multi-file commit is incomplete; disposable cache/index is deleted | Read status and rebuild projection | Recovery-required state or deterministic rebuilt facts | No read-only repair, event append, or revision advance |
| FX-WS-031, FX-WS-032, FX-WS-033 | First request is in flight; same key changed payload; record expired | Submit duplicate/changed/expired key | Join/replay, hard conflict, or explicit expiry result respectively | No second effect or silent key remapping |
| FX-WS-034 | Commit completed and response channel is lost | Retry same key after restart | Durable outcome returned after disclosure check | No authorization rerun or action rerun |
| FX-WS-035, FX-WS-036 | Active owner holds lock; crashed owner/PID reuse; nested same-process call | Timeout, recover, and re-enter lock | Active owner remains; crash release requires proof; reentrancy counted | No timeout-based steal, PID-only release, or deadlock |
| FX-WS-037, FX-WS-038 | Journal has torn/checksum-invalid record; valid checkpoint exists | Replay/compact/replay again | Corruption is surfaced; equivalent checkpoint replay is deterministic | No invented event or discarded committed result |
| FX-WS-039 | Wall clock moves backward between journal appends | Append events | Logical sequence remains ordered; timestamp is audit-only | No revision rollback or event reordering |
| FX-WS-040, FX-WS-041 | Authorization binding is revoked/mismatched; `.sdd` is newer/partial | Apply or inspect workspace | Policy denial or workspace/schema error with correct exit category | No implicit initialization, parent fallback, or migration write |
| FX-WS-042 | Native migration plan and backup/recovery metadata exist | Crash migration at a persistence boundary | Explicit recovery resumes or quarantines without loss | No legacy import or destructive source deletion |
| FX-WS-043, FX-WS-044 | Pending transaction or unsupported network/cloud-sync filesystem | Run status/doctor | Non-mutating recovery report or declared filesystem block/warning | No cleanup, cache repair, lock theft, or transaction creation when blocked |
| FX-WS-045 | Snapshot/plan/journal input is one byte beyond its declared limit | Submit oversized operation and inspect envelope | Stable bounded-limit diagnostic with foundation exit category | No partial staging, revision allocation, or alternate envelope |
| FX-WS-046 | An allocated artifact is edited, moved, then removed under exact transactions | Capture `artifact-identities` after each commit | One binding preserves workspace/work/namespace/ID, advances only committed artifact revisions, updates path on move, then returns a tombstone with allocation/removal provenance | No ID reallocation, second identity database, or consumer reread |
| FX-WS-047 | Persisted artifact bytes claim an ID whose allocation record is absent/corrupt | Capture the same bounded raw snapshot | `missing-record` identity observation and integrity diagnostic with claimed scope; no synthesized allocation | No draft downgrade, silent allocation, repair, or revision advance |
| FX-WS-048 | Parent directory identity is observed and destination directory is absent | Apply authorized `create-directory`, replay, then race by swapping the parent | Exact create succeeds once; replay is a no-op; changed parent conflicts before effect | No implicit parent creation, traversal, or duplicate revision/event |
| FX-WS-049 | Target directory identity and empty child set are observed | Apply authorized `remove-empty-directory`, replay, then repeat with an unexpected child | Empty directory is removed once; replay is a no-op; non-empty/changed identity conflicts | No recursive deletion, unknown-child removal, or guessed recovery |

## Legacy reference mapping

The former unnumbered headings are retained by this mapping so review comments
remain traceable: “Workspace identity and path containment” → WS-REQ-001;
“Immutable snapshot semantics” → WS-REQ-002; “Revision allocation and
comparison” → WS-REQ-003; “Exact plans” → WS-REQ-004; “Safe write effects” →
WS-REQ-005; “Transaction states and recovery” → WS-REQ-006; “Multi-file
visibility and consistency” → WS-REQ-007; “Idempotency” → WS-REQ-008;
“Locking” → WS-REQ-009; “Journal integrity” → WS-REQ-010;
“Authorization seam” and “Initialization and format support” → WS-REQ-011;
“Operational limits” → WS-REQ-012; “Deterministic state and projection
rebuildability” → WS-REQ-007. Former unnumbered adversarial bullets map to
FX-WS-011 through FX-WS-049 in the table above.

## Dependencies, non-goals, and explicit amendments

This specification depends on `sdd-engine-foundation` for the command/result
envelope, native layout/discovery, `MutationPlan` vocabulary, source
coordinates, diagnostic catalog, exit codes, and `Engine.Execute` seam. It
proposes the following foundation amendment before design acceptance:

1. Extend the foundation `PlanOperation.kind` enum from
   `create-file/update-file/delete-file` only through a versioned, additive
   extension for `create-directory`, `remove-empty-directory`, `move-file`, and
   `metadata-update`, with explicit parent/source/destination identity,
   emptiness, old-content, and absence preconditions. Foundation 1.0
   remains valid for its existing three kinds; workspace MUST reject an
   extension when the negotiated plan version does not support it.
2. Reserve workspace transaction fields (snapshot/read-set identity,
   idempotency namespace/key/fingerprint, authorization binding, transaction
   state, recovery reference) in the mutation seam rather than changing the
   foundation envelope or inventing a second public envelope.
3. Foundation's read-only snapshot contract must permit opaque raw-byte/file
   observations and bounded read-set identity. Workspace adds durable checks;
   it does not make foundation import artifact-validation.

Module non-goals are distributed locking/consensus, remote/network effects,
automatic Git operations, process supervision, execution claims/leases,
artifact semantic parsing, workflow graph compilation, completion/DONE
decisions, legacy import UX, and automatic destructive repair of unknown files.

## Ratified implementation baselines and remaining release gates

ADR-0006 resolves the choices needed to begin implementation:

1. Transaction material uses `.sdd/.runtime/transactions/`; destructive
   retention and compaction remain disabled until `DEC-03` supplies policy.
2. Unknown, network, userspace, and sync-managed filesystems fail closed for
   mutation until `DEC-02` certifies a native adapter. Read-only inspection may
   report the unsupported capability without mutating.
3. Hard-linked regular files are supported only through a proven atomic
   directory-entry replacement that leaves the old inode untouched; otherwise
   they fail as unsafe.
4. `protocol/extension-inventory.json` centrally allocates workspace's effect
   variants to the unpublished `1.1` extension. Implementations may build the
   internal typed variants but may not claim public `1.1` support before the
   foundation conformance schema is ratified.
5. Authorization expiry is checked before `PREPARED`. Expiry after that point
   does not undo intent; exact recovery rolls forward unless a separately
   trusted restrictive revocation fact requires quarantine.

`DEC-01`, `DEC-02`, `DEC-03`, `DEC-05`, native certification, and company
numeric limits remain production/release gates. They do not permit an adapter
to weaken these accepted invariants.

## References

- `docs/kiro-input/next-prerequisite-specs.md`, sections 2–4 and 6
- `docs/cli-product-plan.md`, sections 5.1–5.6, 5.8–5.9, 8.5–8.6, 10.5,
  11.1, 11.5–11.7, 16.2–16.4, and Appendix B
- `docs/architecture/go-project-structure.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/quality-and-safety.md`
- `.kiro/specs/sdd-engine-foundation/requirements.md`
- `.kiro/specs/sdd-engine-foundation/design.md`

## For the chosen ones

The subtle transaction invariant is not “all files changed together.” It is
that every authoritative effect is justified by one immutable observation and
one durable intent record. A filesystem rename can expose an intermediate
state, but a correct reader can distinguish that state from a committed state
and recovery can prove which bytes it is allowed to install. Once that seam is
explicit, idempotency, stale revisions, lock ownership, and quarantine become
different projections of the same fact: *which exact authority may change which
exact bytes, and at what durable point?*
