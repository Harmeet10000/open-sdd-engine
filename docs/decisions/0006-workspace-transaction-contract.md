# ADR-0006: Accept the workspace transaction contract

- Status: Accepted
- Date: 2026-10-17
- Scope: coherent raw snapshots, exact local mutation plans, and crash recovery
- Source conversation: `01a1278b-94cb-7c63-a18c-89e44c073dc1`

## Context

The workflow graph, artifact-validation, completion, and guided-workflow
modules all need one trustworthy answer to two questions: which exact bytes
were evaluated, and which exact authorized effects were committed. Leaving
those answers to individual commands would create mixed-generation reads,
incompatible revision rules, duplicated idempotency code, and recovery paths
that cannot distinguish an interrupted commit from a new request.

The requirements, design, and task decomposition in
`.kiro/specs/sdd-workspace-transactions/` have been reviewed as one contract.
This ADR accepts that contract for implementation while keeping platform,
authority-provider, retention, and public-schema certification as explicit
release gates.

## Decision

Accept `WS-REQ-001` through `WS-REQ-012`, the design's module boundary, and
the `WS-TASK-001` through `WS-TASK-014` implementation DAG with these binding
implementation baselines:

1. Authoritative transaction material lives below
   `.sdd/.runtime/transactions/`. User-provided keys never become path names.
2. The currently published public protocol remains `1.0`. Workspace consumes
   the centrally allocated `1.1` ownership in
   `protocol/extension-inventory.json`, but no implementation may advertise
   `1.1` until the foundation conformance schemas are ratified.
3. Mutating operations fail closed on an unknown, network, userspace, or
   sync-managed filesystem. A warning-only mode may be added only by an
   explicit support-matrix decision and may not claim untested durability.
4. Managed hard-linked regular files may be changed only by a proven atomic
   directory-entry replacement that leaves the old inode and every alias
   untouched. In-place mutation and an adapter that cannot prove the property
   are rejected.
5. Action authorization is checked immediately before durable `PREPARED`
   intent. Expiry after `PREPARED` does not revoke the recorded intent.
   Recovery normally rolls that exact intent forward; a separately trusted,
   explicitly restrictive revocation fact may quarantine it but may not
   rewrite or replace the plan.
6. Until company retention/privacy policy is configured, committed summaries,
   idempotency ambiguity tombstones, and journal facts are retained. No
   destructive expiry or compaction may silently choose a duration.
7. Production provider trust, native platform certification, numeric limit
   profiles, and destructive retention remain deployment decisions. They do
   not block pure values, deterministic fakes, coherent capture, exact-plan
   validation, or recovery state-machine implementation.

## Consequences

- `workspace.Transactions` is the deep module seam. Callers can request a
  snapshot, apply an already-authorized exact plan, or explicitly recover;
  they cannot individually stage, lock, allocate revisions, append events, or
  clean transaction material.
- Markdown/YAML, the append-only journal, and retained transaction summaries
  remain authoritative. SQLite and other indexes are disposable projections.
- Downstream modules consume the same accepted snapshot and can propose
  effects, but they cannot reopen files during evaluation or implement another
  commit protocol.
- Initial implementation may use deterministic platform and authorization
  adapters. Such evidence proves module behavior, not native durability or
  company-grade authority.

## Release gates that remain open

- `DEC-01`: production identity/authorization provider and restrictive
  revocation authority.
- `DEC-02`: certified OS/filesystem matrix and native fault evidence.
- `DEC-03`: retention, privacy, compaction, and deletion durations.
- `DEC-05`: publication of the allocated shared protocol extension.
- Company numeric limit and performance profiles.

These are capability and release gates, not permission to change the accepted
transaction invariants.

## Rejected alternatives

- Treating Git as the transaction manager: the CLI must also work before a
  commit and must recover local state without rewriting user Git history.
- Treating SQLite as the authority: it would create two sources of truth and
  make recovery depend on a disposable projection.
- Best-effort multi-file writes: a lost response could not safely replay, and
  a crash could expose files without durable intent or provenance.
- Rechecking policy by creating a new plan during recovery: that would replace
  recorded authority after effects may already have been installed.

## For the chosen ones

The irreversible boundary is not the first target-file rename. It is the
durable publication of enough intent to prove that every later rename is the
continuation of one already-authorized operation. That is why revision outputs,
event IDs, staged digests, and the idempotency fingerprint must be frozen
before `PREPARED`: recovery is replay of a fact, not another planning pass.
