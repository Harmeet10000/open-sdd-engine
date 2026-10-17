# ADR-0003: Ratify prerequisite contracts and company decisions (DEC-01..06)

- Status: **Proposed** — not accepted. Agent authorship is not company acceptance.
- Date: 2026-10-16
- Scope: `sdd-engine-foundation`, `sdd-workspace-transactions`, `sdd-workflow-graph`, `sdd-artifact-validation`, `sdd-completion-lifecycle`, `sdd-guided-workflows`
- Source ledger: [`docs/kiro-input/prerequisite-review.md`](../kiro-input/prerequisite-review.md) §3–§4 (`AM-01`..`AM-11`, `DEC-01`..`DEC-06`)

## Context

The five prerequisite/guided specification drafts are structurally complete
(~8,900 lines) but explicitly unapproved drafts. The review ledger records six
company decisions and eleven cross-spec amendments as blockers. Until they are
resolved, no party may freeze the `1.1` protocol extension, durable transaction
formats, production authorization behavior, completion receipts, retention
behavior, waiver behavior, filesystem support claims, or execution-backed
verification semantics.

This record proposes the company rule for each decision. Acceptance requires
explicit sign-off by the project owner; a checked box in a spec file is not
sign-off. Until accepted, the ledger's drafting rules apply: pure,
provisional-seam implementation may proceed, but no production claim may be
made. This document alone changes no spec file — amendment application is
tracked in the acceptance checklist and happens only after acceptance.

Three readiness levels are kept separate throughout:

```text
ready to prototype pure logic
    !=
ready to freeze an interoperable contract
    !=
ready to claim production authority
```

## Decisions

### DEC-01 — Trusted actors, policy, CI, and reviewer independence

**Proposed rule:**

- Local developers may make ordinary local-development decisions.
- Repository Markdown cannot grant organization authority.
- Critical or production acceptance requires a trusted CI/provider observation.
- Critical completion requires a reviewer distinct from the candidate author.
- Authority is a typed provider fact (per `IC-10`), not a Boolean inside a work artifact.
- Local signatures are optional unless issuer provisioning, key lifecycle,
  revocation, and trust roots are also defined.

Do not require signatures everywhere merely because signatures sound secure. A
signature without a trusted issuer model proves only that some key signed
something.

**Unblocks:** organization-governed production approval; completion reviewer
rules (`CL` evidence acceptance); policy-exception authority.

### DEC-02 — Supported OS and filesystem matrix

**Proposed rule:**

- Initial mutation support: Linux local filesystems ext4 and XFS; macOS APFS;
  Windows NTFS.
- Network filesystems, FUSE/userspace mounts, cloud-synchronized directories,
  SMB, and NFS: read-only inspection permitted with a structured warning;
  authoritative mutation blocked until a native durability/fault suite certifies
  them.
- Cross-compilation proves compilation only.
- Each supported OS/filesystem pair requires native apply, crash, rename, sync,
  held-open-file, permission, disk-full, and process-death tests.
- Earliest walking slice certifies Ubuntu 24.04/ext4 first; macOS/Windows
  mutation support is incomplete until their native fault tests pass.

**Unblocks:** platform durability certification; `WS-TASK-002` production commit
behavior.

### DEC-03 — Retention, privacy, replay, and destructive compaction

**Proposed rule:**

- No automatic destructive deletion in the first implementation.
- Retention classes are typed and policy-controlled.
- Idempotency expiry creates an expiry tombstone rather than silently permitting
  key reuse.
- Completion receipts remain immutable even after evidence is unavailable.
- Evidence references record availability and retention disposition separately.
- Journal compaction is introduced only after checkpoint/replay correctness exists.
- Secrets and raw validator output use separate redaction and retention classes.
- Archive does not mean delete.

This permits implementing bounded retention interfaces without inventing a
company retention duration.

**Unblocks:** destructive retention/compaction deployment; idempotency
retention (`IC-03`); evidence lifecycle (`IC-09`, `IC-12`).

### DEC-04 — Execution provider (`sdd-bounded-execution` first)

**Proposed rule:** production verification evidence requires a
`sdd-bounded-execution` requirements/design/tasks package defining exact argv
execution (never shell strings), working-directory containment, allowlisted
environment, timeouts and cancellation, process-group/job-object termination,
descendant cleanup, output limits and truncation, scratch-space isolation and
cleanup, CPU/memory/process/file limits where enforceable, candidate and
validator-definition binding, execution attempt identity, result and receipt
transport, platform-specific guarantees, crash/restart behavior, trusted
external CI evidence, and the distinction between unavailable, failed,
cancelled, stale, and passed.

Until that spec exists, deterministic fakes may prove interface shape but
cannot produce production verification evidence. Mandatory checks with an
unavailable provider fail closed.

Practical integration order therefore inserts bounded execution between the
completion pure core and completion execution integration:

```text
foundation -> workspace -> workflow graph -> artifact validation
  -> completion pure core -> bounded execution
  -> completion execution integration -> guided workflows
```

**Unblocks:** execution-backed completion in production; `AV` provider-backed
validation; guided Bug Fix verification claims.

### DEC-05 — Shared protocol extension and severity

**Proposed rule:**

- The published foundation `1.0` schema is immutable.
- The centrally allocated extension line is ratified as `1.1`.
- One foundation-owned extension inventory
  (`protocol/extension-inventory.json`); workspace, graph, artifact validation,
  completion, and guided workflows contribute payload schemas to it and none
  independently declares a different `1.1`.
- Preserve `info`, `warning`, `error`, and `critical`. “Blocks transition” is
  represented independently from display severity. Do not map `critical` down
  to `error`. `hint` is optional advisory metadata, never a fifth wire severity.

This is primarily ratification and consumer reconciliation, not new
architecture — the central inventory already exists (FND-001).

**Unblocks:** interoperable schema freeze; guided payload contribution;
diagnostic publication.

### DEC-06 — Emergency waiver policy

**Proposed rule:**

- No `--force-done`. No Boolean waiver.
- Waivers are scoped to explicit gates and an exact candidate or plan, carrying
  issuer, authority, rationale, evidence, issue time, expiry, revocation
  observation, and required retrospective actions.
- Waivers cannot silently eliminate mandatory evidence.
- Emergency deployment and SDD completion remain separate.
- The ordinary completion gate stays blocked until retrospective requirements
  are reconciled.
- Critical waivers require a distinct reviewer or trusted incident authority.

**Unblocks:** emergency production journey; completion exception scope.

### Ancillary ratifications (review §4 follow-ons)

- **Hard-link behavior:** out of scope for the first slice. Workspace
  observation records link-count > 1 on a plan path as a blocking diagnostic
  for mutation; read-only inspection proceeds. A permissive rule, if ever
  wanted, needs its own digest-identity argument.
- **Native transaction directory:** the exact path literal is owned and approved
  in `sdd-workspace-transactions` (per its REQ on owner approval); no other
  spec hardcodes it, and foundation/engine never references it.
- **Parser selections:** Goldmark v2 and YAML v4 are the selected candidates
  for artifact parsing. Selection is confirmed only by the parser conformance
  spike recording actual grammar behavior (alias/anchor limits, pre-expansion
  budgets, unknown-field and source-coordinate behavior); the libraries are not
  adopted by reference alone.
- **Candidate read-set policy:** `CandidateReadRequest` / `CandidateSubject` as
  designed in `sdd-completion-lifecycle`; policy-owned exclusions are explicit
  inputs, never silent omissions.
- **Graph serialization identity:** exact source/binary provenance is distinct
  from semantic compatibility/freshness identity (AM-10); a patch label proves
  neither compatibility nor continuing trust.
- **Graph-limit benchmarking:** limits are set by a recorded benchmarking
  procedure in the workflow spec, not by assertion.
- **Retention classes:** typed classes per DEC-03, shared by workspace,
  validation, and completion; no spec invents its own duration.

## Amendment dispositions (AM-01..AM-11)

All proposals remain unapplied until this record is accepted. On acceptance,
each is applied to its owner file exactly as below — no amendment authorizes a
second implementation of any mechanism:

| ID | Disposition | Owner change |
|---|---|---|
| AM-01 | Accept | Reconcile severity centrally; guided consumes foundation vocabulary, `critical` retained, `hint` advisory-only |
| AM-02 | Accept | One foundation-owned extension inventory; guided contributes payloads, never publishes `1.1` (reword guided design §“publishes 1.1”) |
| AM-03 | Accept | Additional mutation operations defined centrally; directory/move effects are inspectable plan data |
| AM-04 | Accept | FND-006 narrowed to foundation records/preview inputs; generic Markdown/YAML canonicalization moves to artifact validation |
| AM-05 | Accept | GW-002 authors journey definitions, submits to `workflow.Compile`; no guided graph compiler |
| AM-06 | Accept | GW-012 owns bundle rules/traceability, consumes AV parsing/digests; no guided parser |
| AM-07 | Accept | GW-013 selects validator profiles; AV invokes execution, completion judges evidence; no guided validator execution |
| AM-08 | Accept | GW-014 keeps atomic-bundle invariant, delegates idempotency/recovery/replay to workspace |
| AM-09 | Accept | GW-016 contributes journey completion inputs; completion owns gates/DONE/receipt/promotion |
| AM-10 | Accept | Exact provenance separated from semantic compatibility in composition/definition wording |
| AM-11 | Accept | Only real native-format versions documented; no invented migrations |

## Rejected alternatives

- Accepting specs by Markdown existence: the drafts are rigorous in isolation
  but contradict on ownership; acceptance without this ledger would freeze
  duplicate owners for irreversible facts.
- Per-module `1.1` schemas: produces four incompatible wire contracts with one
  version number.
- Signatures-everywhere without an issuer model: theater that proves key
  possession, not authority.
- Choosing retention durations or filesystem certifications in this record:
  company data the authors do not have; the record fixes invariants and
  procedures instead.

## Consequences and controls

- On acceptance: apply the amendment table to owner files; fix `todo.md` to
  distinguish draft-exists / reviewed / decisions-accepted / amendments-applied
  / ready / complete; repair Kiro metadata through Kiro's own flow (backup,
  preserve foundation/workspace identities, issue guided/graph/completion
  identities); add the mechanical spec-audit tool as a CI check.
- FND-003 (immutable protocol model) is unaffected and already implemented
  against the `1.0` schema.
- Unblocked foundation tasks (FND-004 pure ports, FND-005 discovery) may
  proceed behind provisional seams; durable/integrated tasks stay gated.

## Acceptance checklist

- [ ] Project owner signs each of DEC-01..DEC-06 (counter-proposals recorded here, not in five places).
- [ ] Ancillary ratifications confirmed or counter-proposed.
- [ ] AM-01..AM-11 applied to owner files and guided consumer tasks reworded.
- [ ] `todo.md` states refined beyond present/missing.
- [ ] Kiro metadata repaired via Kiro flow with backup.
- [ ] This record's status flipped to Accepted with date and signatory.
