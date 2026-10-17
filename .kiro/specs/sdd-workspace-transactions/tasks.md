# Implementation Tasks: sdd-workspace-transactions

## Status and task rules

These tasks are **accepted for implementation on 2026-10-17** by ADR-0006.
Every incomplete checkbox remains unchecked. A fake can prove module behavior
but cannot satisfy production authorization, native filesystem durability,
public protocol publication, or retention-policy acceptance. Proposed commands
remain verification plans until their corresponding tests exist and pass.

`dependsOn` contains only local task IDs and is the normative local DAG.
Cross-spec, interface-ledger, amendment, and company-decision gates are listed
separately as external dependencies. Test drivers belong beside Go packages as
`*_test.go`; `testdata/` contains data only.

## Tasks

### WS-TASK-001 — Deliver the coherent read-only workspace slice

- [ ] WS-TASK-001 — Deliver the coherent read-only workspace slice
- **Requirements:** `WS-REQ-001`, `WS-REQ-002`, `WS-REQ-007`, `WS-REQ-011`, `WS-REQ-012`
- **dependsOn:** []
- **External dependencies:** `sdd-engine-foundation/FND-003`, `FND-005`, `FND-007`, `FND-010`; `IC-01`; `AM-11`
- **Owned files/modules:** `internal/workspace` bounded raw snapshot extension, generation/read-set observation, artifact-identity projection and read-only transaction status; workspace fixtures in `testdata/`. Foundation retains discovery, nearest-root selection, canonical-layout decoding, and the public `status` command.
- Consume the foundation selection result and extend it through `Transactions.Snapshot`: immutable IDs versus slugs, normalized ordered raw observations, durable artifact identity bindings, stable per-file reads, read-side lock/generation coherence, and no evaluator rereads. Newer/partial native formats fail without initialization or parent fallback.
- **Fixtures:** `FX-WS-001`, `FX-WS-002`, `FX-WS-007`–`FX-WS-010`, `FX-WS-029`, `FX-WS-041`, `FX-WS-043`, `FX-WS-046`, `FX-WS-047`.
- **Verification (proposed):** `go test ./internal/workspace ./internal/engine -run 'Test(SnapshotExtension|FoundationSelectionHandoff|SnapshotCoherence|ArtifactIdentity|ReadOnly)'`; compare bytes, modes, mtimes, entries, identity records, journal tail, and transaction records before/after.
- **Done when:** one bounded snapshot is collectively coherent or returns `DIAG_SNAPSHOT_UNSTABLE`/recovery-required; an incomplete transaction is never decoded as domain state; cancellation and every read retry leave authoritative state unchanged.
- **Implementation progress (2026-10-15):** `internal/workspace.Reader` and the raw `Snapshot` bridge now enforce closed scopes, normalized ordered read sets, explicit limits, generation/active-transaction coherence, owned byte/value copies, domain-separated blob/read-set digests, and allocated/tombstoned/missing artifact-identity invariants. Package and race tests pass through a deterministic capture adapter. This task remains unchecked until the native capture adapter, read-side lock/generation record, stable stat/read/stat retry, transaction-status observation, canonical foundation selection handoff, and no-mutation filesystem fixtures exist.
- **Implementation progress (2026-10-17):** Added a strict checksummed
  `sdd.workspace-generation/1` record and a root-confined native capture adapter
  using `os.Root`. The adapter rejects symlink components, bounds bytes,
  captures native identity/link count on Linux, retries the whole read set on a
  changed file or generation, reports an active transaction before reading
  domain files, and has a filesystem before/after fixture proving capture does
  not mutate. This task remains unchecked: interprocess read/write locking,
  durable artifact-identity/transaction-status loading, the foundation
  selection handoff, non-Linux native identity implementations, and native
  certification are still absent.

### WS-TASK-002 — Deliver an authorized create-file commit/replay walking slice

- [ ] WS-TASK-002 — Deliver an authorized create-file commit/replay walking slice
- **Requirements:** `WS-REQ-003`, `WS-REQ-004`, `WS-REQ-005`, `WS-REQ-006`, `WS-REQ-008`, `WS-REQ-010`, `WS-REQ-011`
- **dependsOn:** [WS-TASK-001]
- **External dependencies:** `sdd-engine-foundation/FND-003`, `FND-006`, `FND-007`, `FND-011`, `FND-014`; `IC-02`, `IC-03`, `IC-10`; `AM-02`, `AM-03`, `DEC-01`, `DEC-05`
- **Owned files/modules:** `internal/workspace` plan core, canonical hash, apply state machine, transaction/manifest/summary serializers, create effect and provider-fact verifier seam; provisional protocol fixtures only in the foundation-owned extension location after allocation.
- Drive one create through snapshot → exact `PlanCore` → provider grant bound to its digest → staging → `PREPARED` → install → journal → committed summary → outcome, then replay the stored outcome. Keep final grant/proof, transaction ID, staging names, and journal offsets out of plan identity; include the authorization-fact digest in the request fingerprint.
- **Fixtures:** `FX-WS-014`–`FX-WS-017`, `FX-WS-023`, `FX-WS-034`.
- **Verification (proposed):** `go test ./internal/workspace ./internal/engine -run 'Test(CreateCommit|PlanDigest|AuthorizationBinding|LostReplyReplay)'` and, after protocol allocation, the foundation schema validator command established by `FND-002`.
- **Done when:** pre/post revision, changed path, event ID, result digest, durability facts, and transaction ID are observable through the module seam; retry creates no second revision/event/effect; faults on either side of every persistence call return old state, resumable intent, or quarantine—never guessed success.
- **Implementation progress (2026-10-17):** Added the immutable workspace-owned
  exact `Plan` core for foundation create/update/delete effects, typed raw
  preconditions, fixed revision intents, bounded staged content, a reusable
  typed canonical-JSON codec, a domain-separated plan digest, non-circular
  authorization/freshness verification, defensive copying, and the accepted
  transaction transition table. Tests prove effect order changes identity,
  grant IDs are excluded from plan identity, stale/revoked/mismatched grants
  fail closed, and `PREPARED` cannot transition to `FAILED`. This task remains
  unchecked: content submission, idempotency records, locking, durable
  `PENDING`/staging/`PREPARED`, install, journal, committed summary, outcome,
  replay, and recovery adapters are not yet implemented.

### WS-TASK-003 — Add update/delete revisions, tombstones, and stale-writer safety

- [ ] WS-TASK-003 — Add update/delete revisions, tombstones, and stale-writer safety
- **Requirements:** `WS-REQ-002`, `WS-REQ-003`, `WS-REQ-004`, `WS-REQ-005`, `WS-REQ-007`
- **dependsOn:** [WS-TASK-002]
- **External dependencies:** `IC-01`, `IC-02`; `AM-03`; `sdd-engine-foundation/FND-014`
- **Owned files/modules:** `internal/workspace` revision allocation, raw preconditions, update/delete effects and tombstone persistence; state-model tests.
- Implement every normative revision-table row, including complete work-record guards when artifact submission intentionally leaves work revision unchanged. Allocate under lock into intent, publish with committed effects, retain deletion identity, and treat external edits as raw observations rather than implicit revisions.
- **Fixtures:** `FX-WS-011`–`FX-WS-013`, `FX-WS-017`, `FX-WS-046`, `FX-WS-047`.
- **Verification (proposed):** `go test ./internal/workspace -run 'Test(RevisionMutationTable|TwoWriters|Tombstone|NoOpReplay)'`.
- **Done when:** two plans from one revision yield one commit and one exit-`4` stale result; no-op/replay allocates nothing; each injected failure before intent publishes no revision and each committed recovery returns the exact recorded revision set.

### WS-TASK-004 — Make prepared recovery safe at every durable boundary

- [ ] WS-TASK-004 — Make prepared recovery safe at every durable boundary
- **Requirements:** `WS-REQ-005`, `WS-REQ-006`, `WS-REQ-007`, `WS-REQ-010`, `WS-REQ-012`
- **dependsOn:** [WS-TASK-002, WS-TASK-003]
- **External dependencies:** `IC-02`, `IC-03`; `DEC-03`; product-plan §5.3 and §16.4 fault model
- **Owned files/modules:** `internal/workspace` recovery/quarantine transitions and owned cleanup; `internal/testkit` transaction boundary injector; transaction fault fixtures.
- Inject process loss immediately before/after transaction-directory sync, record replacement, every manifest/stage sync, `PREPARED`, active generation, each target replace and parent sync, progress update, each journal sync, summary, `COMMITTED`, stable generation, idempotency result, cleanup, and checkpoint replacement. Recovery recognizes exact intended post-digests and fixed event IDs as already applied.
- **Fixtures:** `FX-WS-019`–`FX-WS-028`.
- **Verification (proposed):** `go test ./internal/workspace -run 'TestCrashBoundaryMatrix|TestRecoveryInterruptedAgain|TestOwnedCleanup'`.
- **Done when:** all pre-intent paths change no authoritative file and delete only manifest-owned material; all post-intent paths roll forward once or quarantine on a violated/missing fact; cancellation after intent preserves the transaction; a second recovery interruption remains resumable.

### WS-TASK-005 — Harden containment and effect-time platform races

- [ ] WS-TASK-005 — Harden containment and effect-time platform races
- **Requirements:** `WS-REQ-001`, `WS-REQ-004`, `WS-REQ-005`, `WS-REQ-009`, `WS-REQ-012`
- **dependsOn:** [WS-TASK-001, WS-TASK-002]
- **External dependencies:** `IC-02`; `DEC-02` for native certification
- **Owned files/modules:** `internal/workspace` path normalization/collision keys and containment checks; `internal/platform/filesystem` no-follow root-handle traversal, file identity, safe replacement and capability facts; path fuzz corpus.
- Reject traversal, separator aliases, reserved/device names, path limits, case/Unicode collisions, symlink/reparse escapes, unsafe hard links, and source/destination swaps. Repeat parent/final identity containment immediately before every effect; preserve permissions and unrelated files.
- **Fixtures:** `FX-WS-003`–`FX-WS-006`, `FX-WS-018`–`FX-WS-020`, `FX-WS-044`.
- **Verification (proposed):** `go test ./internal/workspace ./internal/platform/filesystem -run 'Test(Path|Containment|HardLink|HeldOpen|DiskFull|Permission)'`; `go test ./internal/workspace -fuzz FuzzWorkspacePath`.
- **Done when:** a swap can only conflict/quarantine, no outside alias changes, archive collision never overwrites, disk/permission/held-open failure preserves parseable old or staged recoverable bytes, and unsupported capability facts fail closed before `PREPARED`.

### WS-TASK-006 — Complete durable idempotency, concurrency join, and disclosure

- [ ] WS-TASK-006 — Complete durable idempotency, concurrency join, and disclosure
- **Requirements:** `WS-REQ-006`, `WS-REQ-008`, `WS-REQ-011`, `WS-REQ-012`
- **dependsOn:** [WS-TASK-002, WS-TASK-004]
- **External dependencies:** `IC-03`, `IC-10`; `DEC-01`, `DEC-03`
- **Owned files/modules:** `internal/workspace` idempotency namespace/fingerprint records, join/replay/disclosure/expiry behavior and retained result references.
- Bind namespace to workspace, operation, and principal; preserve an in-flight/recovering state across processes; hard-conflict changed fingerprints; authorize disclosure without rerunning the action grant; retain expiry tombstones or require an explicit new namespace.
- **Fixtures:** `FX-WS-031`–`FX-WS-034`.
- **Verification (proposed):** `go test -race ./internal/workspace -run 'TestIdempotency(ConcurrentJoin|Conflict|Expiry|LostReply|Disclosure)'`.
- **Done when:** N concurrent identical submissions produce one transaction/effect/event/revision and equivalent stored outcomes; changed retry time does not alter the fingerprint; cleanup and compaction cannot remove the active replay fact; denied disclosure leaks no result.

### WS-TASK-007 — Enforce local lock ownership and deterministic ordering

- [ ] WS-TASK-007 — Enforce local lock ownership and deterministic ordering
- **Requirements:** `WS-REQ-009`, `WS-REQ-012`
- **dependsOn:** [WS-TASK-001, WS-TASK-002]
- **External dependencies:** `DEC-02`; `sdd-engine-foundation/REQ-015` cancellation contract
- **Owned files/modules:** `internal/workspace` lock resource/order/reentrancy policy; `internal/platform/filesystem` Unix/Windows lock adapters; interprocess lock test helper.
- Acquire workspace/transaction/work/path resources in canonical order and release in reverse. Require nonce/session/native-lock proof for release; timeout never steals; crash cleanup requires proven dead owner or explicit authorized recovery; reentrancy is nonce-scoped and counted.
- **Fixtures:** `FX-WS-035`, `FX-WS-036`.
- **Verification (proposed):** `go test -race ./internal/workspace ./internal/platform/filesystem -run 'TestLock(Timeout|Order|Reentrant|Crash|PIDReuse)'` plus native multi-process jobs.
- **Done when:** active owners survive timeout/cancellation, PID reuse cannot release another guard, lock-order property tests cannot form a cycle, and unsupported network semantics cannot claim coordination support.

### WS-TASK-008 — Deliver journal deduplication, corruption stop, and projection rebuild

- [ ] WS-TASK-008 — Deliver journal deduplication, corruption stop, and projection rebuild
- **Requirements:** `WS-REQ-006`, `WS-REQ-007`, `WS-REQ-010`, `WS-REQ-012`
- **dependsOn:** [WS-TASK-002, WS-TASK-004]
- **External dependencies:** `IC-02`, `IC-03`; `DEC-03`; `AM-01`
- **Owned files/modules:** `internal/workspace` event framing/checksum/replay/deduplication, checkpoint compaction and pure projection builder.
- Fix event IDs/payload digests in prepared intent; append a missing event once, recognize an exact matching event after a lost progress write, and quarantine a duplicate ID with different payload. Replay stops at torn/gapped/unknown-required/manual edits. Checkpoints link the source tail and retain commit/idempotency facts.
- **Fixtures:** `FX-WS-030`, `FX-WS-037`–`FX-WS-039`.
- **Verification (proposed):** `go test ./internal/workspace -run 'Test(JournalReplay|EventDedup|Compaction|ProjectionRebuild|ClockRollback)'`.
- **Done when:** enumeration order and backward wall clock do not change facts; rebuild from the same canonical input is byte-deterministic and advances no revision; corruption is inspectable and never auto-truncated; compaction survives faults with either old or new valid state.

### WS-TASK-009 — Apply provider-bound authorization and native migration safely

- [ ] WS-TASK-009 — Apply provider-bound authorization and native migration safely
- **Requirements:** `WS-REQ-004`, `WS-REQ-006`, `WS-REQ-011`
- **dependsOn:** [WS-TASK-001, WS-TASK-002, WS-TASK-004]
- **External dependencies:** `IC-10`; `DEC-01`; `AM-11`; a project-owner-approved native migration pair and backup contract
- **Owned files/modules:** `internal/workspace` provider-fact comparison, recovery revocation branch, and migration plan adapter under `migrations/workspace` only after a real pair is approved.
- Check provider authenticity, principal/scope/plan/policy binding, issue/expiry, nonce, and revocation immediately before intent without evaluating domain validity. Route pre-intent denial to exit `6` and stale raw/revision facts to exit `4`. Apply a future approved format pair as an exact backed-up transaction; do not invent a format predecessor.
- **Fixtures:** `FX-WS-016`, `FX-WS-040`–`FX-WS-042`.
- **Verification (proposed):** `go test ./internal/workspace ./migrations/workspace/... -run 'Test(AuthorizationBinding|PolicyVersusConflict|FormatRead|MigrationRecovery)'` only after the external migration pair exists.
- **Done when:** stale/expired/revoked/mismatched grants cause no `PREPARED`; restrictive post-intent policy preserves evidence in quarantine; unknown/partial format never initializes; every migration boundary resumes or quarantines without source loss.

### WS-TASK-010 — Expose bounded, non-mutating status/doctor/recovery UX

- [ ] WS-TASK-010 — Expose bounded, non-mutating status/doctor/recovery UX
- **Requirements:** `WS-REQ-002`, `WS-REQ-004`, `WS-REQ-006`, `WS-REQ-007`, `WS-REQ-012`
- **dependsOn:** [WS-TASK-001, WS-TASK-004, WS-TASK-006, WS-TASK-008, WS-TASK-009]
- **External dependencies:** `sdd-engine-foundation/FND-003`, `FND-007`, `FND-009`, `FND-010`, `FND-012`; `AM-01`, `AM-02`, `DEC-05`
- **Owned files/modules:** `internal/engine` workspace outcome mapping, foundation-owned diagnostics/catalog extension, human/JSON fixtures after shared protocol allocation; no independent envelope.
- Render current/recovery/quarantine facts, actual platform durability, structured expected/observed values, and one safe next action. Dry-run/status/doctor create no transaction, cleanup, event, revision, projection, or retained lock. Enforce exact-at/one-over limits and secret/proof/key/content redaction.
- **Fixtures:** `FX-WS-014`, `FX-WS-028`, `FX-WS-029`, `FX-WS-043`–`FX-WS-045`.
- **Verification (proposed):** `go test ./internal/engine ./internal/presentation/... -run 'Test(WorkspaceOutcome|ReadOnlyRecoveryStatus|Limits|Redaction)'` and the foundation schema/example validator once allocated.
- **Done when:** human/JSON meanings match, public `1.0` remains unchanged, no draft labels itself `1.1`, every failure has foundation severity/exit/remediation, and filesystem snapshots prove read-only behavior even during pending/quarantined transactions.

### WS-TASK-011 — Add directory/move/metadata effects only through the shared extension

- [ ] WS-TASK-011 — Add directory/move/metadata effects only through the shared extension
- **Requirements:** `WS-REQ-001`, `WS-REQ-003`, `WS-REQ-004`, `WS-REQ-005`
- **dependsOn:** [WS-TASK-003, WS-TASK-005, WS-TASK-009]
- **External dependencies:** ratified `AM-02`, `AM-03`, `DEC-05`; foundation-owned extension task amending `sdd-engine-foundation/FND-001`, `FND-002`, `FND-014`
- **Owned files/modules:** version-gated `internal/workspace` create-directory/remove-empty-directory/move/metadata variants and consumer conformance fixtures; foundation owns public schema files.
- Implement only allocated closed variants with exact parent/source/destination/native-identity/absence/emptiness/metadata preconditions. File effects never create or remove parents implicitly; journal housekeeping is not a domain effect; no multi-file atomic rename is promised. Without negotiation, return non-authorizable before staging.
- **Fixtures:** `FX-WS-006`, `FX-WS-014`, `FX-WS-018`, `FX-WS-048`, `FX-WS-049`.
- **Verification (proposed):** `go test ./internal/workspace -run 'Test(EffectVersionNegotiation|DirectoryPrecondition|MoveCollision|MetadataPrecondition)'` plus shared positive/negative schema fixtures.
- **Done when:** foundation-native create/update/delete stay wire-compatible, unknown variants fail closed, archive collisions preserve both sides, and each fault leaves the source/destination in an old, exact-new, or quarantined state.

### WS-TASK-012 — Certify each claimed native platform/filesystem adapter

- [ ] WS-TASK-012 — Certify each claimed native platform/filesystem adapter
- **Requirements:** `WS-REQ-001`, `WS-REQ-002`, `WS-REQ-005`, `WS-REQ-006`, `WS-REQ-009`, `WS-REQ-012`
- **dependsOn:** [WS-TASK-004, WS-TASK-005, WS-TASK-007, WS-TASK-008, WS-TASK-010, WS-TASK-011]
- **External dependencies:** approved `DEC-02` support matrix and native CI runners; `sdd-engine-foundation/FND-015`
- **Owned files/modules:** OS-specific `internal/platform/filesystem` adapters and native integration jobs; support-matrix evidence.
- Run the same no-follow identity, shared/exclusive lock, replacement, file/directory sync, held-open, kill-process, case/Unicode, path-limit, permission, disk-full, and recovery suite on every claimed OS/filesystem. Cross-builds and skipped capability tests are recorded separately and confer no support.
- **Fixtures:** adapter-native executions of `FX-WS-003`–`FX-WS-006`, `FX-WS-009`, `FX-WS-019`–`FX-WS-024`, `FX-WS-035`, `FX-WS-036`, `FX-WS-044`.
- **Verification (proposed):** native CI `go test -race ./internal/platform/filesystem ./internal/workspace -run 'TestNative'`; support report must name OS, filesystem, mount facts, capabilities, and skipped tests.
- **Done when:** every claimed capability has native runtime evidence and every unsupported/unknown combination blocks or warns exactly per approved policy; kill/fault runs show zero unsafe loss and retained recovery evidence.

### WS-TASK-013 — Run state-model, fuzz, and full fixture conformance

- [ ] WS-TASK-013 — Run state-model, fuzz, and full fixture conformance
- **Requirements:** `WS-REQ-001`, `WS-REQ-002`, `WS-REQ-003`, `WS-REQ-004`, `WS-REQ-005`, `WS-REQ-006`, `WS-REQ-007`, `WS-REQ-008`, `WS-REQ-009`, `WS-REQ-010`, `WS-REQ-011`, `WS-REQ-012`
- **dependsOn:** [WS-TASK-003, WS-TASK-004, WS-TASK-005, WS-TASK-006, WS-TASK-007, WS-TASK-008, WS-TASK-009, WS-TASK-010, WS-TASK-011, WS-TASK-012]
- **External dependencies:** `IC-01`, `IC-02`, `IC-03`, `IC-10`; approved portions of `DEC-01`–`DEC-05`; foundation conformance kit from `FND-015`
- **Owned files/modules:** `internal/testkit` model/fault runner, package fuzz tests, all `FX-WS-001`–`FX-WS-049` golden repositories and expected outcomes.
- Generate apply/crash/recover/replay/cancel/duplicate sequences and compare transaction state, generation, revisions, files, events, idempotency, and outcomes after every step. Fuzz paths, manifests, records, journal frames, effect ordering, canonicalization and size boundaries; drive public behavior primarily through `Engine.Execute`.
- **Verification (proposed):** `go test -race ./internal/...`; targeted `go test ./internal/workspace -fuzz 'Fuzz(Path|Manifest|Journal|Plan)'`; native matrix from WS-TASK-012; schema conformance command from foundation.
- **Done when:** all 49 fixtures name exact expected and forbidden mutations, local and public outputs are deterministic, every persistence edge has before/after coverage, and no fake/platform skip is counted as native durability or production authority evidence.

### WS-TASK-014 — Reconcile consumers and close the implementation review

- [ ] WS-TASK-014 — Reconcile consumers and close the implementation review
- **Requirements:** `WS-REQ-001`, `WS-REQ-002`, `WS-REQ-003`, `WS-REQ-004`, `WS-REQ-005`, `WS-REQ-006`, `WS-REQ-007`, `WS-REQ-008`, `WS-REQ-009`, `WS-REQ-010`, `WS-REQ-011`, `WS-REQ-012`
- **dependsOn:** [WS-TASK-013]
- **External dependencies:** accepted results for `AM-01`–`AM-03`, `AM-08`, `AM-09`, `AM-11`; decisions `DEC-01`–`DEC-05`; `sdd-guided-workflows/GW-014`, `GW-015`, `GW-016`, `GW-018`, `GW-019`
- **Owned files/modules:** workspace conformance report and interface-ledger evidence; amendments to other specs/tasks only in their owner-approved follow-up, not silently in this task.
- Propose exact consumer edits: `GW-014` keeps bundle-specific plan construction/acceptance tests but delegates commit/recovery/idempotency to `IC-02`/`IC-03`; `GW-015` consumes committed revisions while graph/completion own eligibility; `GW-016` consumes a completion-owned exact plan and workspace outcome; `GW-018` uses the centrally allocated protocol extension; `GW-019` requires real workspace adapter evidence.
- **Verification (proposed):** run the repository's spec traceability/DAG checker once it exists; mechanically check requirement IDs, fixture IDs, undefined dependencies, cycles, unchecked tasks, and forbidden competing envelopes; review `git diff --check` only after files are tracked/staged or pass paths explicitly.
- **Done when:** every requirement and fixture maps to evidence, the local DAG is acyclic, all consumer duplication is assigned to an owner amendment, unresolved decisions remain visibly blocked, and reviewers—not this task—record acceptance. No implementation checkbox is auto-completed.

## Local DAG and coverage audit

The local graph is acyclic by construction:

```text
001 -> 002 -> 003 -> 004
 |      |       |      +-> 006 -> 010
 |      |       +-> 011 ----+
 |      +-> 005 ----------> 012
 |      +-> 007 ----------> 012
 |      +-> 008 ----------> 010/012
 +----------> 009 --------> 010/011
003..012 -> 013 -> 014
```

Requirement coverage: `WS-REQ-001` → 001/005/011/012/013/014;
`002` → 001/003/010/012/013/014; `003` → 002/003/011/013/014;
`004` → 002/003/009/010/011/013/014; `005` → 002/003/004/005/011/012/013/014;
`006` → 002/004/006/008/009/010/012/013/014; `007` → 001/003/004/008/010/013/014;
`008` → 002/006/013/014; `009` → 005/007/012/013/014;
`010` → 002/004/008/013/014; `011` → 001/002/006/009/013/014;
`012` → 001/004/005/006/007/008/010/012/013/014.

Fixture coverage is explicit in task bodies: `FX-WS-001`–`010` (001/005/012),
`011`–`020` (002/003/005/009/011/012), `021`–`030` (002/004/008/010/012),
`031`–`039` (006/007/008), `040`–`045` (001/005/009/010/012), and
`046`–`049` (001/003/011). WS-TASK-013
executes the complete set and verifies there are no gaps.

## Remaining production/release gates

- No production authorization claim until `DEC-01` identifies trusted providers.
- No native durability/support claim until `DEC-02` and WS-TASK-012 pass on native runners.
- No destructive compaction/expiry until `DEC-03` sets retention/privacy policy.
- No public apply/recover schema or move/metadata effect until `DEC-05` and `AM-02`/`AM-03` are ratified.
- No native migration implementation until an actual source/target pair and backup contract are approved under `AM-11`.
- Hard-linked files are accepted only through proven atomic replacement; all
  in-place or unproven behavior fails closed under ADR-0006.

## For the chosen ones

The fault suite must kill the process after the filesystem primitive succeeds
but before progress metadata advances. That tiny interval is where shallow
transaction implementations duplicate events or revisions. Testing only
failures returned by an adapter exercises error handling; testing successful
syscalls followed by process death exercises recovery truth.
