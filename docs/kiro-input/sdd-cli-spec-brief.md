# Specification brief: production `sdd` CLI

**Status:** Kiro drafting input; proposed, non-normative until accepted  
**Decision source:** `docs/cli-product-plan.md`, revision 6  
**Architecture source:** `docs/architecture/go-project-structure.md`  

## 1. Outcome

Build a company-grade, agent-neutral CLI that turns a software-change intent into a controlled sequence of specification, implementation, validation, and completion actions.

The CLI owns deterministic mechanics and evidence. The coding agent owns codebase exploration, clarification, design judgment, code changes, and review reasoning.

The product succeeds when different agents can complete the same workflow through the same JSON contract without agent-specific behavior leaking into the engine.

## 2. User-facing contract

The primary experience is:

```text
sdd init
sdd start "change intent"
sdd next
sdd status
sdd check
sdd finish
```

`sdd doctor` is a diagnostic/recovery command. Advanced migration, configuration explanation, debugging, integration, and protocol commands may exist outside the primary path.

The CLI must hide artifact graphs, locking, transaction recovery, digests, selection coordinates, context budgets, claims, projections, and adapter manifests during ordinary success.

## 3. Locked product decisions

Treat these as constraints during spec drafting unless the spec is explicitly proposing to supersede one with an ADR:

1. Production language and executable: Go native binary.
2. Public interfaces: CLI behavior, versioned JSON protocol, and `.sdd/` workspace format.
3. Deep application seam: `Engine.Execute(ctx, Command) Outcome`.
4. Canonical workspace: `.sdd/`; `.kiro/`, `openspec/`, and `.specify/` are import formats.
5. Authority: human-readable repository files plus append-only journal.
6. Projections: materialized state and SQLite indexes are disposable/rebuildable.
7. Agent-neutrality: integrations render context/commands but cannot alter workflow semantics.
8. Offline-first: no default network access, telemetry, package installation, push, or PR creation.
9. Mutations: plan → authorize → apply with expected revisions and recoverable multi-file transactions.
10. Completion: policy-driven `done.md` plus snapshot-bound machine receipt; no force-DONE.
11. Extensions: local/version-controlled first; no in-process plugin marketplace in 1.0.
12. Requirements styles: proportional and typed; EARS is optional rather than universal.
13. Knowledge: OKF-compatible Markdown is authoritative; SQLite/FTS is a read model; Obsidian is optional; hosted service deferred.
14. TypeScript: temporary migration oracle only, removed from the shipped product after Go parity.
15. Workflow composition: journey topology, risk-derived rigor, and artifact/validation profiles are orthogonal; a named preset cannot lower the organization or risk floor.

## 4. Primary actors

| Actor | Need |
|---|---|
| Developer | Begin and finish a change without learning internal workflow machinery |
| Coding agent | Receive one precise next action with sufficient, provenance-bearing context |
| Reviewer | Trace intent and requirements to code, tests, evidence, deviations, and residual risk |
| Security/platform owner | Enforce an organization policy floor and inspect every material side effect |
| Maintainer | Recover interrupted work, migrate formats, evolve schemas, and support releases |
| CI consumer | Invoke non-interactive commands with stable JSON and exit codes |

## 5. Core journeys

### 5.1 Initialize

- Discover and validate the intended project root.
- Show an inspectable creation/migration plan.
- Create canonical `.sdd/` files without overwriting unknown content.
- Detect host/agent capabilities without requiring an integration.
- Return equivalent human and JSON outcomes.

### 5.2 Start work

- Create immutable workspace/work identity.
- Classify journey shape, assess observable risk, and resolve a maintained preset into journey + rigor + artifact profiles.
- Record routing evidence and allow an explicit override with rationale.
- Permit overrides to change journey/preset or increase rigor, but never silently lower the applicable policy floor.
- Expose the first action, not the whole internal graph.

### Guided Bug Fix journey

- `sdd start "defect" --workflow bug-fix` uses report → reproduce/observe → investigate → diagnose → regression contract → plan → fix → verify → finish.
- A report records expected/observed behavior and impact without pretending the cause is known.
- Reproduction distinguishes reproduced, intermittent, environment-specific, and not-reproduced states.
- Diagnosis needs evidence that separates the proposed mechanism from competing hypotheses.
- Implementation is blocked until required diagnosis/regression evidence exists unless explicit emergency policy applies; an emergency repair must still reconcile evidence and accepted specs before finish.
- Bug-fix is a journey, not low rigor. Security, migration, public-contract, concurrency, destructive-data, or broad-blast-radius defects escalate independently.

### Quick Spec journey

- `sdd start "bounded change" --workflow quick-spec` uses intent → material clarifications → one generated requirements/design/tasks draft bundle → validation → approval → implementation-ready.
- The engine asks only material questions and investigates repository facts instead of asking the user when possible.
- The core does not invoke an LLM. It emits a bounded `generate-spec-bundle` action/context; a host agent or human produces the three artifacts and submits one revision-bound bundle.
- Generation is not acceptance. IDs, traceability, derivation digests, task ordering, verification coverage, and unresolved assumptions are validated.
- Fast means bundled authoring and fewer turns, not weaker verification. Material risk emits `WORKFLOW_ESCALATION_REQUIRED` while preserving prior answers/artifacts.

### 5.3 Get the next action

- Resolve dependencies, freshness, policy gates, clarifications, claims, and capabilities.
- Select exactly one action by default.
- Build the smallest sufficient context pack with authority, provenance, and exclusions.
- Never embed executable authority from repository/research prose.

### 5.4 Check

- Run cheap deterministic validation before expensive verification.
- Emit stable diagnostics with codes, severity, source range, evidence, and remediation.
- Treat implementation convergence as read-only.
- Never report an unrun/interrupted/timed-out/wrong-snapshot/uninterpretable check as passed.

### 5.5 Finish

- Verify every required artifact and task is fresh.
- Map requirements/scenarios and changed paths to implementation and verification evidence.
- Match evidence to the exact result snapshot.
- Reconcile discoveries and perform three-way accepted-spec promotion.
- Render `done.md` and a machine receipt atomically with lifecycle events.
- Reject force-completion; allow only policy-authorized, scoped exception artifacts.

### 5.6 Recover

- Detect interrupted prepared transactions and stale locks.
- Prove safe roll-forward or stop with an inspectable recovery plan.
- Rebuild all projections from authoritative files and journal.
- Resume compound work at the first unresolved leaf rather than retrying the whole operation.

## 6. Essential models

- Workspace identity and canonical root.
- Work identity, lifecycle, and expected revision.
- Journey, rigor, artifact-profile, and named-preset schemas.
- Artifact/action dependency graph.
- Blob, semantic, and derivation digests.
- Stable requirement, scenario, decision, task, finding, evidence, and diagnostic IDs.
- MutationPlan and authorization decision.
- Prepared transaction and recovery state.
- Append-only journal event.
- Claim/lease and idempotent submission.
- Verification receipt and completion receipt.
- Context entry with authority, provenance, trust, freshness, and sensitivity.
- Initiative with shallow child-slice DAG.
- Extension manifest with independent trust facts and requested capabilities.

## 7. Non-negotiable failure scenarios

Specs must cover applicable scenarios from this list:

- explicit invalid root with another valid root above it;
- multiple active work items and no selected work ID;
- stale expected revision or changed input digest;
- two processes attempting the same mutation;
- crash before/after staging, flush, rename, journal append, and projection update;
- disk full, permission failure, antivirus-held file, case-fold collision, Unicode-normalization collision, and symlink swap;
- NFS/SMB/cloud-sync filesystem with weaker locking or durability semantics;
- process cancellation with grandchildren holding pipes, ports, or Git locks;
- Windows and Unix process-tree differences;
- malformed/hostile Markdown, YAML, JSON, schema blocks, paths, and oversized output;
- post-verification edit that makes evidence stale;
- known failing baseline and flaky verification;
- bug report is not reproducible, is intermittent, or conflicts with an accepted spec;
- symptom disappears but root-cause evidence remains ambiguous;
- a `bug-fix` preset encounters critical risk and must escalate without changing journey;
- Quick Spec generation is partial, untraceable, stale, or exceeds the quick-risk ceiling;
- requirements change after a Quick Spec bundle and must stale the derived design/tasks;
- feature implemented but not deployable or not rolled out;
- extension that is schema-valid but unauthenticated, signed but unauthorized, or redirected to another origin;
- obsolete research document outranking a current contract;
- index deleted/corrupt while authoritative files remain healthy;
- migration from a modified legacy workspace with unknown files;
- partial initiative fan-out or changed compound-workflow definition.

## 8. Quality requirements to quantify in individual specs

Do not invent permanent numbers in this brief. Each implementation spec must define measurement method and initial budget for relevant qualities:

- cold `--help`/`--version` startup;
- `status` latency by workspace size;
- validation throughput and memory bounds;
- maximum persisted/output size and truncation behavior;
- lock timeout and cancellation grace period;
- crash/fault recovery coverage;
- binary size and supported OS/architecture matrix;
- deterministic JSON/canonicalization compatibility;
- migration success/rollback evidence;
- diagnostic false-positive and false-negative fixtures;
- agent journey completeness, safety, review comprehension, cost, and latency.

Safety, correctness, and evidence completeness dominate latency, line count, token count, or binary size.

## 9. Architecture constraints

- Use the target structure in `docs/architecture/go-project-structure.md`.
- Organize internal code by product capability, not global technical layers.
- Use ports/adapters only for real effects or hard external dependencies.
- Keep the CLI composition root shallow.
- Keep platform behavior out of workflow policy.
- Use language-neutral JSON Schema as the public machine contract.
- Embed built-in workflow resources but keep them inspectable/exportable.
- Do not expose internal package names in workspace files or protocol payloads.
- Do not choose a SQLite driver until the Phase 5 projection spike compares build portability, FTS support, backup/rebuild behavior, licensing, maintenance, and performance.

## 10. Recommended Kiro spec decomposition

Create these as separate specs, in order. Later specs may be drafted early but should not assume an unimplemented dependency is complete.

| Order | Spec name | Scope | Depends on |
|---:|---|---|---|
| 1 | `sdd-engine-foundation` | Go skeleton, public schemas, deep Execute seam, version/doctor/status/dry-run start, TypeScript parity fixtures | none |
| 2 | `sdd-workspace-transactions` | selection, snapshots, locks, journal, recoverable transaction, projection rebuild, platform fault tests | foundation |
| 3 | `sdd-workflow-graph` | workflow schemas, action graph, risk routing, freshness, initiatives, compound cursor | foundation + workspace |
| 4 | `sdd-artifact-validation` | Markdown AST, typed artifacts, IDs, diagnostics, traceability, protocol profile | workflow graph |
| 5 | `sdd-completion-lifecycle` | evidence, receipts, DONE gate, convergence, three-way promotion, archive | workspace + validation |
| 6 | `sdd-guided-workflows` | orthogonal journey/rigor/profile composition; Bug Fix investigation/diagnosis/regression flow; Quick Spec clarification and atomic draft-bundle flow; escalation fixtures | workflow + validation + completion |
| 7 | `sdd-minimal-ux-migration` | six-command UX, legacy imports, human rendering, documentation journeys | foundation through guided workflows |
| 8 | `sdd-agent-integrations` | capability handshake, context packs, managed files, generic/Codex/Claude/Kiro adapters | minimal UX |
| 9 | `sdd-context-index` | OKF ingestion, authority filtering, SQLite/FTS rebuildable projection, Obsidian guidance | artifact validation |
| 10 | `sdd-bounded-execution` | process supervision, claims/leases, receipts, worktrees, ready-node scheduling | workspace + workflow graph |
| 11 | `sdd-release-hardening` | signing, provenance, SBOM, support matrix, security audit, pilot and compatibility freeze | all required 1.0 capabilities |

Do not create one spec containing all eleven areas. That would produce an unreviewable design and tasks too broad for independent verification.

## 11. First spec boundary: `sdd-engine-foundation`

### In scope

- initialize a Go module and native `sdd` executable;
- embed version/build provenance;
- define normative `Command`, `Outcome`, `Diagnostic`, and `MutationPlan` JSON Schemas;
- generate or validate Go protocol types from the schemas;
- implement one deep `Engine.Execute` interface;
- implement `version`, `doctor`, `status`, and non-mutating `start --dry-run` slices;
- separate stdout JSON, stderr operational logs, human rendering, and exit codes;
- capture hermetic TypeScript prototype fixtures and compare normalized read-only outcomes;
- native build/smoke jobs for the proposed support matrix;
- no filesystem mutation except explicitly scoped fixture/test directories.

### Out of scope

- final workspace transaction implementation;
- completion/promotion/archive;
- SQLite;
- remote integrations;
- arbitrary plugin execution;
- process runner for project verification;
- automatic Git changes;
- full legacy workspace migration.

### Required demonstrations

1. A native artifact runs `--version` and `--help` in an empty directory without Node, Bun, Go, or source files.
2. `status --json` returns a schema-valid outcome for no workspace, valid workspace fixture, incompatible workspace version, and ambiguous work selection.
3. `start --dry-run --json` returns an inspectable, non-mutating plan.
4. The CLI contains no workflow decision beyond parsing and rendering.
5. Engine interface tests use deterministic fake dependencies.
6. Every error path has stable code, exit code, and human/JSON rendering fixture.
7. Go output matches normalized legacy fixtures where old behavior is intentionally preserved; intentional differences are recorded decisions.

## 12. Explicit questions Kiro should raise rather than guess

- Which exact OS/architecture combinations will the company support first?
- Is the repository internal-only or intended for public open-source release?
- Which organization policy authority signs or distributes policy bundles?
- What evidence identifies independent reviewers inside the company?
- Which existing repositories form the first golden migration set?
- Which native artifact distribution channels are approved?
- What minimum Git version is supported when Git features are enabled?
- Which filesystems are supported versus warned/blocked?
- Are FIPS or other regulated cryptographic requirements applicable?
- What retention/redaction policy applies to verification output and receipts?

Unknown answers should become explicit open questions, assumptions with expiry, or deferred requirements—not invented facts.
