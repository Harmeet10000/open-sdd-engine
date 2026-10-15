# Repository TODO: Kiro specification implementation sequence

This is an implementation-planning checklist. The Kiro specification drafts and their review are not evidence that implementation is complete; every implementation checkbox below intentionally remains unchecked.

## Provenance

- Source conversation (Codex thread ID): `01a1278b-94cb-7c63-a18c-89e44c073dc1`
- Date: 2026-10-13 (Asia/Kolkata)
- Origin: Kiro specification review and prerequisite planning.
- Source brief: [sdd-cli-spec-brief.md](docs/kiro-input/sdd-cli-spec-brief.md)

## Prerequisite ownership

The four missing prerequisite specs must be drafted and accepted before `sdd-guided-workflows` is integrated. Do not infer or invent their implementations in this TODO.

| Spec | Owns | Current planning status |
|---|---|---|
| `sdd-workspace-transactions` | Canonical paths, snapshots, revisions, locking, exact-plan application, idempotency, and crash recovery. | Missing prerequisite; draft and acceptance required. |
| `sdd-workflow-graph` | Graph compilation, valid transitions, conditional gates, reachability, definition pinning, composition, and escalation. | Missing prerequisite; draft and acceptance required. |
| `sdd-artifact-validation` | Artifact schemas, stable IDs, traceability, canonicalization, semantic/derivation digests, and validator execution. | Missing prerequisite; draft and acceptance required. |
| `sdd-completion-lifecycle` | Candidate-bound evidence, approvals, DONE generation, receipt, promotion, archive, and post-verification edits. | Missing prerequisite; draft and acceptance required. |

## Existing specification drafts

These are reviewed/drafted inputs only. Drafting and review do not mean implementation is done.

- [sdd-engine-foundation directory](.kiro/specs/sdd-engine-foundation/)
  - [requirements.md](.kiro/specs/sdd-engine-foundation/requirements.md)
  - [design.md](.kiro/specs/sdd-engine-foundation/design.md)
  - [tasks.md](.kiro/specs/sdd-engine-foundation/tasks.md)
- [sdd-guided-workflows directory](.kiro/specs/sdd-guided-workflows/)
  - [requirements.md](.kiro/specs/sdd-guided-workflows/requirements.md)
  - [design.md](.kiro/specs/sdd-guided-workflows/design.md)
  - [tasks.md](.kiro/specs/sdd-guided-workflows/tasks.md)

## Ordered implementation TODO

Complete in this order. Keep each item unchecked until the corresponding implementation and conformance evidence actually exists.

- [ ] 1. `sdd-engine-foundation` — Implement the native Go executable, public protocol schemas, `Engine.Execute` seam, version/doctor/status slices, and inspectable `start --dry-run`; use the existing draft as design input, not as completion evidence.
- [ ] 2. `sdd-workspace-transactions` — Draft, accept, and implement canonical paths, immutable snapshots, revisions, locking, exact-plan application, idempotency, and crash recovery.
- [ ] 3. `sdd-workflow-graph` — Draft, accept, and implement graph compilation, valid transitions, conditional gates, reachability, definition pinning, composition, and escalation.
- [ ] 4. `sdd-artifact-validation` — Draft, accept, and implement artifact schemas, stable IDs, traceability, typed canonicalization, semantic/derivation digests, and authoritative validator execution.
- [ ] 5. `sdd-completion-lifecycle` — Draft, accept, and implement candidate-bound evidence, approvals, atomic DONE generation, receipts, promotion, archive, and post-verification edit invalidation.
- [ ] 6. `sdd-guided-workflows` — Integrate Bug Fix and Quick Spec only after items 2–5 are drafted/accepted and their contracts conform; use the existing draft as design input, not as completion evidence.
- [ ] 7. `sdd-minimal-ux-migration` — Implement the minimal user-facing CLI journey and migration behavior on the accepted foundation/workflow contracts.
- [ ] 8. `sdd-agent-integrations` — Implement agent-neutral context/action integrations and host adapters while preserving provenance, structured commands, and policy boundaries.
- [ ] 9. `sdd-context-index` — Implement the bounded context index and rebuildable projections; keep authoritative state in the workspace artifacts and journal contracts.
- [ ] 10. `sdd-bounded-execution` — Implement claims/leases, bounded process supervision, execution receipts, worktree/effect controls, and ready-node scheduling.
- [ ] 11. `sdd-release-hardening` — Implement release build provenance, signing/SBOM policy where accepted, native support-matrix validation, migration hardening, and release conformance.

## Reference sources

- [Kiro input brief](docs/kiro-input/sdd-cli-spec-brief.md)
- [Foundation requirements](.kiro/specs/sdd-engine-foundation/requirements.md)
- [Foundation design](.kiro/specs/sdd-engine-foundation/design.md)
- [Foundation tasks](.kiro/specs/sdd-engine-foundation/tasks.md)
- [Guided workflows requirements](.kiro/specs/sdd-guided-workflows/requirements.md)
- [Guided workflows design](.kiro/specs/sdd-guided-workflows/design.md)
- [Guided workflows tasks](.kiro/specs/sdd-guided-workflows/tasks.md)
