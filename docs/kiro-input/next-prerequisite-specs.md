# Generate the next four prerequisite specs in Kiro

Use this handoff to draft the missing prerequisites for the existing foundation and guided-workflow specs. The intended result is four bounded, reviewable specifications with explicit contract ownership, not another whole-product spec.

Status: proposed drafting guidance, not an accepted implementation specification. Existing drafts are inputs, not evidence of acceptance or implementation. The implementation sequence remains in [todo.md](../../todo.md).

Provenance: Codex conversation `01a1278b-94cb-7c63-a18c-89e44c073dc1`; prepared 2026-10-13 (Asia/Kolkata).

## 1. Use this sequence

1. Give Kiro this file and repository access. Run the preflight prompt in section 5 first.
2. Review its ownership map, interface ledger, contradictions, and unresolved decisions.
3. Generate `sdd-workspace-transactions` requirements using section 6. Review requirements, then use section 10 for design and tasks, with a separate review between phases.
4. Repeat for `sdd-workflow-graph`, `sdd-artifact-validation`, and `sdd-completion-lifecycle`, in that order.
5. Run the cross-spec audit in section 11. Reconcile existing consumer specs through explicit amendment proposals before implementation.
6. Only then continue with the remaining specs in `todo.md`. Do not regenerate `sdd-engine-foundation` or `sdd-guided-workflows` as new specs.

This is a drafting order, not permission to invent upstream contracts. A downstream draft may use a clearly labelled provisional interface, but its approval and integration must remain blocked on the named unresolved decision. Keep that decision in the files; do not rely on chat memory.

Use a focused Kiro session for each spec if context becomes crowded. Carry forward the reviewed ownership/interface ledger and the relevant upstream files, rather than the entire research corpus.

## 2. Supply these source files

Paths below are repository-relative. Kiro should read the common inputs and all six existing spec files during preflight. For subsequent drafting, it should retain the common rules and read the relevant accepted/provisional upstream contracts in full.

### Common inputs

- `AGENTS.md`, plus any applicable nested instructions.
- `.kiro/steering/product.md`
- `.kiro/steering/tech.md`
- `.kiro/steering/structure.md`
- `.kiro/steering/cli-ux.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/quality-and-safety.md`
- `.kiro/steering/spec-authoring.md`
- `docs/kiro-input/sdd-cli-spec-brief.md`
- `docs/architecture/go-project-structure.md`
- `todo.md`

### Existing specs: inspect all six files

- `.kiro/specs/sdd-engine-foundation/requirements.md`
- `.kiro/specs/sdd-engine-foundation/design.md`
- `.kiro/specs/sdd-engine-foundation/tasks.md`
- `.kiro/specs/sdd-guided-workflows/requirements.md`
- `.kiro/specs/sdd-guided-workflows/design.md`
- `.kiro/specs/sdd-guided-workflows/tasks.md`

The guided-workflow files are downstream compatibility inputs. Their requested capabilities must be assigned to the right owner; they must not create a reverse dependency from prerequisites to journey-specific implementation.

### Targeted supporting material

Read the indicated sections of `docs/cli-product-plan.md`, including their tables and examples. If numbering changes, locate the named subjects. Follow referenced definitions when needed; do not treat a heading-only scan as reading the contract.

| Target | Product-plan sections and additional input |
|---|---|
| Workspace transactions | Sections 5.1–5.6, 5.8–5.9, 8.5–8.6, 10.5, 11.1, 11.5–11.7, 16.2–16.4; Appendix B filesystem/concurrency cases. |
| Workflow graph | Sections 4.3–4.5, 6.1–6.5, 6.9–6.13, 8.5–8.6, 12.2–12.3, 16.4; `docs/architecture/guided-workflows.md` for consumer expectations, not duplicate journey ownership. |
| Artifact validation | Sections 5.5–5.6, 6.11, 7, 11.6–11.7, 12, 16.4; Appendix B protocol/traceability cases. |
| Completion lifecycle | Sections 5.8, 6.6–6.8, 6.10–6.11, 8.6, 10.5; Appendix B completion/evidence cases; guided-workflow completion conditions. |

Read preceding new specs as they become available. Read their tasks when checking implementation ownership. Do not load every file in `notes/`, vendor research, or old prototypes by default. Research informs decisions; it is not automatically the current product contract.

If sources disagree, cite the exact paths, sections, and requirement/task IDs. Apply an explicitly documented precedence rule only when it resolves the disagreement. Otherwise request a decision and propose the smallest amendment. Do not declare a draft accepted merely because it exists.

## 3. Enforce single-owner contracts

Use this ownership map as the initial allocation to verify in preflight. A consumer may state what it needs and test integration; it must not independently redefine the owner's rules or implement a second copy.

| Contract or decision | Owner | Consumer responsibility |
|---|---|---|
| `Engine.Execute`, public envelope, source coordinates, exit-code conventions, protocol compatibility | `sdd-engine-foundation` | Extend through reviewed versioned amendments; preserve old fixtures. |
| Minimal native discovery/read-only selection and foundation dry-run preview | `sdd-engine-foundation` | Workspace adds durable semantics without making foundation depend on a future implementation. |
| Canonical containment, snapshots, revision allocation, lock discipline, durable journal/recovery | `sdd-workspace-transactions` | Consume a snapshot and submit effects; never write authoritative state directly. |
| Exact-plan comparison/application and durable idempotency | `sdd-workspace-transactions` | Supply typed intentions and scoped authorization facts; never re-plan during apply. |
| Generic graph language, compilation, reachability, action transitions, composition and escalation mechanics | `sdd-workflow-graph` | Supply definitions and facts; do not build another router. |
| Generic dependency/freshness propagation and compound leaf progress | `sdd-workflow-graph` | Produce trustworthy artifact/evidence facts and persist transitions through workspace. |
| Markdown/YAML parsing, typed artifact model, IDs, canonicalization, digests and traceability | `sdd-artifact-validation` | Consume normalized values; do not use a second regex/parser/digest implementation. |
| Structural/policy validation, validator descriptors, assessment input schemas and validation results | `sdd-artifact-validation` | Supply trusted inputs and consume diagnostics; do not turn heuristic claims into proof. |
| Candidate-bound evidence acceptance, completion gates, approvals and waivers | `sdd-completion-lifecycle` | Supply receipts/decisions from declared authorities; do not self-certify completion. |
| DONE content, completion receipt, promotion conflict decisions, archive/supersession semantics | `sdd-completion-lifecycle` | Workspace applies the resulting authorized transaction; other modules do not mark work completed. |
| Bug Fix/Quick Spec definitions, specialized artifact profiles, clarification/generation journeys | `sdd-guided-workflows` | Use generic prerequisite contracts; own only journey-specific policy and behavior. |
| Human command journey, interactive presentation and legacy imports | `sdd-minimal-ux-migration` | Use engine outcomes and workspace plans. Basic usable diagnostics still belong in every slice. |
| Host adapters, agent capability negotiation and context/action integrations | `sdd-agent-integrations` | Translate the shared protocol without introducing authority. |
| OKF retrieval, SQLite/FTS projections and index rebuilds | `sdd-context-index` | Index canonical state; do not replace its authority with the cache. |
| Process supervision, execution claims/leases, worktree controls and scheduling | `sdd-bounded-execution` | Implement the execution seam; workspace locks are not action leases. |
| Release provenance, packaging/support matrix and release conformance | `sdd-release-hardening` | Harden existing contracts rather than redefine them. |

### Record an interface ledger before design

For every shared interface, identify one definition owner and its producers/consumers. Record inputs, outputs, stable identifiers, schemas, errors, version rules, trust source, freshness rules, and conformance fixtures. Documentation IDs such as `WS-C-001` are sufficient; do not build a new runtime registry just to document ownership.

At minimum cover:

1. Immutable workspace snapshot, selection coordinates, and revision tokens.
2. Mutation plan, effect vocabulary, preconditions, authorization binding, and apply/recovery outcome.
3. Idempotency scope, request fingerprint, replay record, and retention behavior.
4. Compiled graph and pinned definition identity.
5. Action evaluation and next-action result, including blocked/unknown/stale states.
6. Composition, risk assessment consumption, and escalation proposal.
7. Typed artifact identity, canonical content, source ranges, and digest roles.
8. Proposed artifact delta and the information required for three-way promotion.
9. Validation request/result, validator identity, and execution receipt reference.
10. Effective policy snapshot and scoped authorization/approval facts.
11. Verification subject and candidate-bound evidence/approval references.
12. Completion proposal, DONE record, receipt, promotion and archive effects.

Split policy responsibilities explicitly: artifact-validation owns validating/normalizing policy inputs; graph owns composition from those facts; workspace checks authorization binding at application; completion owns completion-specific gate and reviewer rules. Policy distribution, company identity issuance, and remote attestation are declared external inputs or later decisions, not four new policy engines.

Avoid dependency cycles. Workspace handles opaque versioned effects and does not import graph/completion business rules. Artifact parsing must not import its validators. Engine coordinates values between modules. A graph's need for validation facts must not require a validation implementation during compilation. Keep exported interfaces small while documenting their full invariants and failure semantics.

## 4. Resolve conflicts and reserve later scope

Known questions for preflight, not permission to silently change existing files:

1. Foundation diagnostic severity values are `info|warning|error|critical`; product-plan severity language differs. Choose an explicit mapping or a reviewed protocol amendment, not two enums.
2. Foundation uses an integer workspace format and the product plan discusses current/previous compatibility. Define the actually supported versions; do not invent a historical format or confuse workspace and public schema versions.
3. Foundation's `1.0` schema is closed; guided-workflows proposes `1.1`. Intervening command/effect additions need one version-allocation decision and compatibility fixtures. Four specs must not each independently claim the same next version.
4. Foundation's preview effect vocabulary may not express journal operations, archive moves, and transactional metadata. Propose owned extensions rather than disguising all effects as unrestricted file updates.
5. Distinguish validator semantic-contract identity from exact implementation provenance. Product-plan invalidation rules and existing wording about exact definition digests must be reconciled explicitly. A patch may preserve semantics, but changing trust, declared behavior or required policy cannot be dismissed as cosmetic.
6. Foundation dry-run previews and symbolic example digests are not executable transaction fixtures. Keep legacy examples labelled and require real hashes for new conformance fixtures.
7. Existing guided tasks may describe generic compilation, parsing, bundle acceptance or completion work. Identify exact task IDs to delegate/amend after prerequisite acceptance, preserving journey-specific tests and requirements.

For each amendment proposal include: issue ID, evidence, current owner, affected consumers, proposed wording/schema change, compatibility impact, tests, required decision, and whether the issue blocks requirements, design, or implementation. Do not leave all contradictions hidden in a generic “open questions” paragraph.

Reserve these for the later five specs:

- Minimal UX/migration: interactive onboarding and full legacy import journeys. Workspace still owns safe native-format initialization, unsupported-format behavior and migration transaction mechanics.
- Agent integrations: MCP/host adapters, role packets and host-specific managed files. Prerequisites expose neutral data contracts only.
- Context index: SQLite/FTS/vector retrieval, Obsidian integration and any hosted Turso topology. Earlier specs only declare authoritative inputs and rebuild/invalidation signals.
- Bounded execution: subprocess trees, resource limits, leases, worktree management and multi-agent schedulers. Validation defines the invocation/result seam, not a competing supervisor.
- Release hardening: installers, signing/SBOM policy, distribution and support certification. Earlier specs still need platform-specific correctness and failure tests.

No embedded LLM, automatic commits/pushes, mandatory hosted backend, generic event bus, speculative plugin runtime, generic repository/service framework, or game-engine ECS machinery is implied. DDP means data-driven/data-oriented programming; ROP means explicit idiomatic Go success/failure control flow. Follow the existing Go steering rather than adding a `Result` monad abstraction.

## 5. First prompt: preflight only

Paste this before asking Kiro to generate any new spec:

```text
Read docs/kiro-input/next-prerequisite-specs.md and run its preflight.
Read the common inputs and all six existing spec files listed in section 2.
Inspect the relevant product-plan sections as you evaluate contract ownership.

Do not implement code, regenerate existing specs, edit todo.md, or draft all four specs now.
Return:
1. An inventory of existing and missing specs; acceptance/implementation status must be
   evidenced or explicitly unknown.
2. A single-owner responsibility matrix, including exclusions and consumer obligations.
3. The shared interface ledger from section 3, with current definitions and missing decisions.
4. Exact contradictions and proposed amendments, referencing file sections and IDs.
5. The draft/implementation dependency DAG, including provisional interfaces and blockers.
6. Company decisions required: trust/identity source, waiver authority, filesystem support,
   evidence retention/privacy, version support and verification execution availability.
7. A list of files actually read and any unavailable input or context-limit omission.

Do not invent acceptance, executed tests, upstream implementations, or authority.
Stop for review. The next generation target after approval is only
sdd-workspace-transactions/requirements.md, using section 6.
```

## 6. Spec 2: sdd-workspace-transactions

### Requirements prompt

```text
Create only .kiro/specs/sdd-workspace-transactions/requirements.md.
Use the reviewed preflight ledger and sections 2–4 and 6 of
docs/kiro-input/next-prerequisite-specs.md. Do not write design, tasks or implementation yet.

Own canonical workspace identity/paths, immutable snapshots, revisions, locks,
exact-plan application, durable idempotency, journaling and crash recovery.
Extend the foundation seam through explicit amendment proposals when needed.
Do not reimplement CLI envelopes, graph policy, typed artifact semantics, completion
decisions, execution leases, legacy imports or the context index.

Use stable WS-REQ-### IDs, observable acceptance criteria and traceable fixture IDs.
Separate behavior from architecture constraints. Apply EARS where it improves clarity.
Include every mandatory topic and adversarial case below, or explain its explicit owner
and exclusion. State non-goals, consumed contracts, unresolved decisions and blockers.
Stop for requirements review.
```

### Mandatory topics

1. **Identity and containment.** Distinguish immutable workspace/work IDs, mutable slugs, project selection and work selection. Specify canonical paths, nested workspaces, symlinks/reparse points, hard links, case collisions, Unicode names, separators, reserved names, path length and root-escape handling. Check containment at effect time, not only when planning.
2. **Snapshot semantics.** Define the read set, stable observation rules, external-edit detection and what “immutable snapshot” guarantees. No evaluation pass silently rereads newer files. Specify behavior when a stable snapshot cannot be obtained, including diagnostic and next action.
3. **Revision allocation.** Define workspace/work/artifact revision scope, monotonic allocation, comparison, deleted identities, no-op behavior and external edits. A timestamp or mtime alone is not a revision. Specify which administrative writes affect each token.
4. **Exact plans.** Specify canonical plan identity, effect ordering, expected old content/nonexistence, source snapshot, policy/authorization binding, lifetime and size limits. Dry-run has no authoritative side effects. Apply the exact authorized effects or reject; do not silently recompute a different plan.
5. **Safe writes.** Define supported effect types and safe create/update/delete/move behavior. Preserve unrelated user changes and permissions. Handle source/target races and disk-full failures. Explain file/directory durability requirements per supported platform without claiming unsupported guarantees.
6. **Transaction states.** Specify legal states, durable records and recovery at each boundary. Reconcile the product plan's PREPARED point: before it, discard only owned incomplete staging; after it, recover the recorded commit intent, unless conflicting external changes require quarantine. Never destructively “repair” unknown user content.
7. **Multi-file visibility.** Define how engine readers see one coherent transaction state or a recovery-required outcome. Ordinary editors can observe intermediate file renames; do not promise cross-file filesystem atomicity that the implementation cannot supply. State the actual consistency contract.
8. **Recovery.** Specify restart after every write/flush/rename/journal boundary, corrupt/truncated records, missing staging, manual journal edits and committed-but-unacknowledged requests. Make recovery idempotent. Read-only `status`/`doctor` must not quietly mutate authoritative state; expose an explicit recovery action where necessary.
9. **Idempotency.** Define namespace, canonical request fingerprint, same-key/same-request replay, same-key/different-request conflict, concurrent duplicate submission, in-flight/recovering state, result retention and expired-key handling. Persist enough to disambiguate a lost response. Do not promise exactly-once effects beyond owned local transactions.
10. **Locking.** Define protected resources, acquisition order, timeout/cancellation, crash release, same-process reentrancy, ownership and stale-lock policy. Avoid unsafe PID-only lock stealing. Distinguish local interprocess locks from distributed locks and future work-execution leases.
11. **Journal integrity.** Define event identity/order, checksums, schema versions, append behavior, truncation detection, replay and compaction/retention constraints. A local checksum detects accidental corruption; it is not proof against a user who controls the filesystem.
12. **Authoritative versus derived state.** Specify what is canonical and what is rebuildable. Cache loss or SQLite unavailability must not change correctness. Selection records and derived views must not become an undocumented second authority.
13. **Authorization seam.** Define trusted authorization inputs and bind them to principal/scope/plan/policy identity and expiry as applicable. Recheck at apply. The transaction module enforces binding, not organization policy distribution or completion reviewer decisions.
14. **Initialization and formats.** Define missing/partial `.sdd`, supported native versions, newer/unknown versions, safe native-format migration planning and backup/recovery expectations. Leave full legacy import UX to its owner.
15. **Operational limits.** Specify bounded snapshot/plan/journal sizes, cancellation semantics before and after commit intent, error classification, secret redaction, and filesystem assumptions. Explicitly reject or restrict unsupported network filesystems rather than silently claiming local-lock guarantees.

### Required adversarial fixtures

- Two writers from one revision; external edit after authorization; no-op/replayed plan; stale policy or expired grant.
- Crash before PREPARED, during each effect, after commit but before reply, and during recovery itself.
- Duplicate request while the first is running; changed payload under the same key; retention expiry.
- Symlink target swapped during apply; case-insensitive collision; archive destination collision; held-open file on Windows; permission loss/disk exhaustion midway.
- Corrupt/torn journal, missing staged content, unrelated user file in staging, clock moving backwards and PID reuse.
- Read-only status during a pending transaction; cache deleted; unsupported workspace version; cancellation after durable commit intent.

## 7. Spec 3: sdd-workflow-graph

### Requirements prompt

```text
Create only .kiro/specs/sdd-workflow-graph/requirements.md.
Use the reviewed ledger, sections 2–4 and 7 of the handoff, foundation contracts,
and the workspace spec. Treat unresolved upstream contracts as explicit blockers.

Own the generic declarative graph/compiler, legal transitions, gates, reachability,
definition pinning, composition, escalation and deterministic next-action evaluation.
Guided workflows own the actual Bug Fix/Quick Spec definitions and specialized journeys.
Do not duplicate transactions, artifact parsing, validation execution, completion
acceptance, an LLM planner or a process scheduler.

Use WG-REQ-### IDs and observable conformance fixtures. Include the topics and cases
below, scope exclusions, consumed/produced interfaces and amendment proposals.
Stop for requirements review; do not generate design, tasks or code.
```

### Mandatory topics

1. **Definition language.** Specify versioned node/edge/port identities, typed input/output references, dependency versus freshness relations, conditional predicates, cardinality and supported extension points. Keep the language closed and inspectable; untrusted definitions cannot inject shell execution or arbitrary code.
2. **Compilation.** Detect duplicate/missing references, cycles, incompatible contracts, invalid predicates, unreachable required nodes and impossible completion routes. Bound node/edge counts, recursion and evaluation complexity. Distinguish compile errors from runtime missing information.
3. **Reachability and gates.** A condition cannot silently remove a mandatory policy gate. Define unknown/insufficient input explicitly rather than coercing it to false or passed. Check required outputs on all applicable routes, including conditional skip routes.
4. **Runtime states.** Specify a transition table with inputs, authority, preconditions, outputs, revision changes and diagnostics. Distinguish ready, running/claimed facts if supported, blocked, stale, failed, skipped and completed semantics. Do not conflate graph-node completion with whole-work completion.
5. **Static topology versus attempts.** Keep the static DAG separate from retries, rejection/revision events and emergency reconciliation. Runtime attempts can revisit work without inventing cycles or replacing immutable history. Define attempt identity and exhaustion outcomes.
6. **Deterministic next action.** Identical snapshots, definitions, policy and assessment yield identical ordered recommendations. Specify tie-breaking, one useful human next action versus machine-readable alternatives, blocked explanations and stale expected revisions. Recommending an action does not authorize its effects.
7. **Composition.** Resolve journey topology, rigor and artifact/validation profiles with explicit policy precedence. Preserve a policy floor; missing or revoked required policy fails closed. Consume typed external assessments with provenance, never infer authoritative risk from prose or a null default.
8. **Definition pinning.** Specify which versions/digests are pinned, how they are resolved offline and what changes make work stale. Updating a named preset must not silently rewrite active work. Define explicit migration/recomposition with revision-bound acceptance.
9. **Escalation.** Define a proposal/acceptance/application lifecycle and preserve evidence and history. Retain journey topology when allowed; change rigor/profile only under policy. Lowering rigor, exceptions and emergency paths require explicit authority and cannot bypass later reconciliation.
10. **Freshness propagation.** Consume artifact/derivation/evidence validity facts from their owners. Specify invalidation propagation and minimal recomputation. A node cannot require its own output to become ready: verification receipts follow verification, and bundle acceptance cannot require the bundle already be accepted.
11. **Submissions.** Define the generic transition interface for draft/submitted/rejected/accepted/stale artifact facts and assessments. The graph requests validation and consumes acceptance facts; it neither parses artifacts nor trusts an agent's submitted `accepted` status.
12. **Initiatives and compound progress.** Define parent/child identity, leaf progress, dependencies, aggregate status, failure/cancellation and resume cursor behavior. Persist only through workspace transactions. Do not create a second specification hierarchy or retry an entire completed compound operation after one leaf fails.
13. **Engine integration.** Define additions to start/next/submit/evaluation results using the foundation schema process. Keep specialized guided payloads with their owner. Include human/JSON parity, deterministic diagnostics and actionable recovery without introducing unnecessary commands.

### Required adversarial fixtures

- Cycle, disconnected mandatory gate, impossible predicate combination, missing producer, duplicate ID and incompatible input/output profile.
- Unknown risk, stale policy, missing pinned definition, policy floor overridden by an explicit low-risk flag.
- Conditional skip bypassing review, emergency mitigation with unreconciled ordinary gates, invalid escalation or changed candidate during acceptance.
- Definition edited under the same display name; one upstream semantic edit invalidating several consumers; cosmetic edit preserving freshness where allowed.
- Verification requires its own receipt; a draft bundle requires prior acceptance; reject both dependency deadlocks.
- Rejected artifact retry preserves attempt history; two simultaneous submissions; stale next action; deterministic tie-breaking.
- Compound operation with three successful leaves and one failed leaf; resume without duplicating successful effects; dependency cancellation and parent status consistency.

## 8. Spec 4: sdd-artifact-validation

### Requirements prompt

```text
Create only .kiro/specs/sdd-artifact-validation/requirements.md.
Use sections 2–4 and 8 of the handoff and reviewed upstream interfaces.
Own typed artifact parsing/schemas, stable IDs, traceability, canonicalization,
semantic/derivation digests, validators and structured assessment/policy inputs.

Own the validator invocation/result contract but not the later process supervisor.
Completion owns whether evidence satisfies completion; guided workflows own their
specialized content profiles. Do not duplicate either decision.

Use AV-REQ-### IDs and observable fixtures, with normative examples for hashes,
malformed input and traceability. Identify every external trust dependency.
Stop for requirements review; no design, tasks or implementation yet.
```

### Mandatory topics

1. **Parsing.** Parse structured Markdown with an AST and frontmatter with a constrained YAML parser. Specify grammar/profile versions, duplicate keys, aliases, nesting/size limits, UTF-8/BOM/line endings and treatment of fenced examples. A requirement-like heading inside a code fence is not a real requirement.
2. **Typed model.** Define common artifact envelope, kind/status, immutable identity, work association, revisions, source coordinates and typed content extension rules. Preserve source provenance. Separate parsing, structural validity, submission and trusted acceptance.
3. **Stable IDs.** Define namespaces, uniqueness scope, creation, rename/move stability, tombstones, cross-work references and collisions. Do not derive identity solely from headings, filenames or line numbers. Fail on ambiguity rather than linking to an arbitrary match.
4. **Canonicalization.** Specify exact normalized representation and hash algorithm/version, ordering, null/absent distinctions, Unicode treatment, whitespace and unknown fields. Preserve semantically meaningful order, code blocks, normative wording and examples where profile rules require them.
5. **Digest roles.** Distinguish raw byte/blob identity, semantic identity and derivation identity. Define the dependency closure for derivation, including relevant source IDs, profiles, policy and validator semantic contracts. Avoid self-referential digests and incidental timestamps/formatting in semantic identity.
6. **Provenance and implementation identity.** Record exact producer/validator implementation provenance separately from declared semantic compatibility. Define when a patch may preserve freshness and when trust/security or behavioral changes invalidate it. Reconcile upstream wording before design acceptance.
7. **Traceability.** Specify requirement-to-scenario/design/task/code/check/evidence edges, reverse coverage, invalid targets, optionality/waivers and removal impact. A URL or code line is a locator, not proof of a valid relationship. Keep structural completeness separate from semantic correctness.
8. **Artifact deltas.** Define add/modify/remove operations, base references, IDs and normalized comparison inputs required by completion's three-way promotion. Parsing/delta validity belong here; conflict resolution and promotion authorization belong to completion.
9. **Validator descriptors.** Specify stable ID, version, semantic contract, accepted inputs, deterministic ordering, applicability, severity, trust tier, required capabilities and outputs. Resolve approved definitions from trusted configuration, not artifact prose.
10. **Validation pipeline.** Separate syntax, schema, graph, traceability, deterministic policy, execution-backed checks and heuristic assessment. Define pass/fail/error/skipped/unavailable distinctions; mandatory unavailable checks fail closed. Aggregate diagnostics without turning execution failure into a warning pass.
11. **Authority and assessments.** Typed external risk/diagnosis/materiality assessments carry evidence, uncertainty, author/approval provenance and explicit insufficient/unassessed states. Validate their structure and trust, not the truth of root-cause prose. Heuristics may suggest; they cannot satisfy mandatory proof gates alone.
12. **Execution seam.** Specify argv/cwd/environment allowlist, input/candidate identity, requested limits, stdout/stderr/output digest, exit/signal/timeout/cancellation and receipt provenance. Do not implement an unrestricted shell runner here. Before bounded-execution exists, declare approved external CI/manual evidence policy or an unavailable result. Test fakes are not production verification.
13. **Policy data.** Define normalized effective-policy inputs, schema validation, precedence inputs, pinned identity and trust/expiry status. Approval grants and evidence are typed facts with explicit issuers. Do not build a remote organization policy distribution product in this spec.
14. **Profiles.** Define extension contracts for lightweight, ordinary, protocol/RFC-style and guided artifact profiles. Protocol validation covers normative keyword usage, syntax, state/failure semantics, security/privacy/operations and conformance structure. Do not claim IETF publication or standards certification from local linting.
15. **Read-only assessment and limits.** Validation/convergence assessment returns findings and remediation proposals without modifying artifacts. Applying a fix requires a new normal plan. Specify bounded traversal, redaction, deterministic output/source locations and behavior when repository facts are missing.

### Required adversarial fixtures

- Duplicate YAML keys, alias expansion, malformed UTF-8, enormous nesting, hidden control characters, requirements inside fenced examples and ambiguous duplicate IDs.
- CRLF/LF and permitted cosmetic changes preserve semantic identity; normative wording, executable examples, ordering where meaningful and dependency/policy changes invalidate it.
- Removed requirement still referenced by a task; code moved; duplicate basename; external reference unavailable; cyclic derivation dependency.
- Agent-provided `passed` or `accepted`; forged validator name; heuristic confidence mistaken for evidence; executable commands embedded in Markdown.
- Validator missing, timeout, truncated output, stale candidate, secret in stderr, malformed receipt and unauthorized environment request.
- Profile upgraded under the same name; validator patch compatibility versus semantic change; no process adapter present; validation must not mutate files.

## 9. Spec 5: sdd-completion-lifecycle

### Requirements prompt

```text
Create only .kiro/specs/sdd-completion-lifecycle/requirements.md.
Use sections 2–4 and 9 of the handoff, all reviewed prerequisite interfaces and
the existing guided-workflow completion expectations.

Own candidate-bound evidence acceptance, approvals/waivers, completion gates,
DONE content, receipts, three-way promotion, archive and post-verification edits.
Workspace owns committing the effects; graph owns generic routing; artifact-validation
owns parsing and validation; execution belongs to its declared provider.

Use CL-REQ-### IDs, explicit lifecycle/authority tables and adversarial fixtures.
Do not redefine shared envelopes or implement process supervision.
Stop for requirements review; no design, tasks or implementation yet.
```

### Mandatory topics

1. **Lifecycle.** Define legal work states and transitions for verification, review, finish, completion, archive, abandonment, supersession and reopening where supported. Distinguish failed attempt, cancelled request and terminal work state. No direct `ready -> completed` shortcut.
2. **Verification subject.** Define candidate identity from the actual implementation/read set, including relevant tracked/untracked files, configuration, generated inputs, dependencies and build/test environment. Handle no-Git repositories, dirty trees, submodules and ignored files explicitly; Git HEAD alone cannot identify a dirty candidate.
3. **Avoid self-invalidating proof.** Distinguish candidate content identity from administrative work revision. Predeclare policy-owned exclusions for DONE, receipts, journals and other bookkeeping so recording evidence does not invalidate itself. The agent must not choose exclusions that hide changed implementation files.
4. **Verification state.** Distinguish planned/not-run, running, passed, failed, inconclusive, unavailable, stale and explicitly waived. A checkbox, expected result or test command string is not evidence that execution occurred.
5. **Baseline and regression evidence.** Define failing-before/passing-after relationships and safe alternatives for intermittent, environment-specific or unsafe reproduction. Approved proxy evidence must state limits. An inability to reproduce is not a passing regression check.
6. **Evidence authority.** Bind each check to candidate, command/input/environment, executed validator identity, result/output digests and trusted producer. Specify replay, expiry, revoked trust, partial/truncated evidence and independently verifiable external receipts. Locally editable files are not tamper-resistant attestations by themselves.
7. **Approvals and waivers.** Bind reviewer decisions to candidate, policy and gate scope. Define independence and author/reviewer identity rules, expiry, revocation and emergency limits. An agent-supplied display name or role string is not proof of an authorized reviewer.
8. **Gate evaluation.** Define deterministic required-gate evaluation and actionable blocked results. Decide whether all gates are evaluated for diagnostics or short-circuited for safety. Missing/unknown evidence never passes by default. Completion rechecks required facts against the application snapshot.
9. **DONE content.** Require delivered scope, summary of how it was done, requirement/scenario coverage, design decisions, code references, project-guideline references, planned and actual task order, verification, deviations, waivers, residual risk, follow-ups, security/privacy/migration/operations, acceptance and reproduction instructions. Tailor required sections by trusted profile; do not demand meaningless filler.
10. **Durable references.** Code references include repository-relative path, symbol when available, immutable revision or candidate digest and precise optional line range. Guideline references identify version/content and applicable clauses. Detect stale references; do not pretend line numbers alone survive edits.
11. **Completion receipt.** Define versioned immutable content, work/candidate/policy/definition identity, evidence/approval references, timestamps as metadata, completion attempt and transaction linkage. Explain the acyclic digest graph: a receipt cannot include its own final hash or require itself as a prerequisite.
12. **Atomic completion proposal.** Evaluate gates and produce DONE, receipt, promotion, lifecycle and archive effects as one exact plan. Workspace owns recoverable application. Handle a candidate/policy/baseline edit between evaluation and apply; no partially completed authoritative state or silent re-plan.
13. **Three-way promotion.** Compare pinned base, proposed change and current canonical target by stable identity and semantic delta. Specify clean/disjoint merges, delete-versus-edit, rename/ID collision, conflicting modifications and reviewer resolution. Do not overwrite another work item's accepted change. Conflict resolution is a fresh authorized plan.
14. **Archive.** Preserve IDs, provenance and evidence references; define collision handling and discovery of completed work after relocation. Keep completion, archive and eventual deletion/retention separate. Archiving is not permission to delete evidence or Git branches.
15. **Retry and partial scope.** Integrate workspace idempotency and crash recovery. A duplicate finish returns the committed receipt under the documented replay contract. Define documentation-only work, no-change outcomes, partially delivered scope and deferred requirements without quietly claiming unimplemented scope completed.
16. **Post-verification changes.** Specify which edits stale evidence/approvals, invalidate pending completion or require a new work revision. Preserve historical receipts; do not rewrite history to match changed code. Define reopen/supersede behavior and its relation to immutable completed records.
17. **Privacy and observability.** Minimize evidence storage, redact secrets and define reference availability/retention requirements. Distinguish auditability from storing entire logs. Report blocked gate, affected subject and safe next action consistently in human/JSON modes.

### Required adversarial fixtures

- All tasks checked but no execution evidence; fake pass, forged reviewer, missing mandatory gate or self-approval where forbidden.
- Dirty/untracked implementation file after verification; policy revoked before apply; excluded-file rules modified by the agent; administrative receipt write must not stale its own candidate.
- Verification passed against another environment/candidate; CI receipt cannot be authenticated; output unavailable or redacted; no execution provider.
- No Git, docs-only change, flaky test, unsafe reproduction, approved proxy and expired emergency waiver.
- Simultaneous promotions from the same base; disjoint changes versus same-requirement conflict; delete/modify collision; archived destination already exists.
- Crash during each completion effect; response lost after commit; concurrent duplicate finish; retry after evidence expiry where prior commit must be distinguished from a new attempt.
- Post-completion code edit; stale code/guideline references; reopen preserving prior receipt; unfinished follow-up cannot be concealed as delivered scope.

## 10. Follow-up prompts for each spec

Replace `<TARGET-SPEC>` with one of the four names. Do not send these until the preceding phase has been reviewed. If that review is unavailable, describe the phase as provisional rather than accepted.

### Design prompt

```text
Draft only .kiro/specs/<TARGET-SPEC>/design.md from the reviewed requirements.
Follow docs/kiro-input/next-prerequisite-specs.md and its ownership/interface ledger.
Do not implement code, draft tasks, or silently edit other specs.

Include:
1. Requirement-to-design traceability and explicit exclusions.
2. Small deep module interfaces, plain immutable data transformations and effect seams;
   idiomatic Go errors/early returns, with no generic Result/service/repository framework.
3. Concrete types, versioned schemas, invariants, source coordinates and error taxonomy.
4. State/transition tables including authority, freshness, retries and forbidden transitions.
5. End-to-end flows through Engine.Execute and exact workspace plans; identify pure and
   effectful phases and every authoritative write owner.
6. Dependency direction and shared-contract ledger, including downstream guided consumers.
7. Trust model, failure/cancellation/crash paths, redaction, bounded resources and platforms.
8. Conformance fixtures with explicit inputs and expected outputs; valid and invalid examples.
9. Public-schema/workspace-format compatibility and exact upstream amendment proposals.
10. Open decisions, blocked dependencies and a test strategy that distinguishes pure unit
    tests, adapter contracts, fault injection and real integration evidence.

Do not hide the difficult contract inside undefined helpers named validate(), authorize(),
recover() or complete(). Specify their observable guarantees without writing implementation.
Validate any embedded schema examples where tooling exists; report only checks actually run.
If a contract conflict prevents design acceptance, state exactly what needs a decision.
Stop for design review.
```

### Tasks prompt

```text
Draft only .kiro/specs/<TARGET-SPEC>/tasks.md from the reviewed requirements and design.
Follow the handoff and check existing foundation/guided tasks for duplicate ownership.
Do not implement tasks or check them off.

Use numbered checkboxes with stable task IDs and explicit dependsOn IDs.
Each task must specify owned files/modules, requirement IDs, consumed contract IDs,
observable outcome, verification command/fixture, and relevant rollback/recovery checks.
Use independently verifiable vertical slices; avoid all-models/all-repositories/all-handlers
horizontal task batches. State cross-spec dependencies separately from local task IDs.

Start with an end-to-end walking slice through the public interface, then add failure,
concurrency, authority, version and adversarial cases. Include schema fixture validation,
deterministic ordering, cancellation, real adapter contract tests and integration tests.
Check the local and cross-spec dependency DAG for cycles and undefined task IDs.

Do not create tasks implementing contracts owned elsewhere. Name the required upstream
deliverable and consumer conformance test instead. If the adapter will arrive later,
label fake-based tests explicitly and keep production integration blocked; a fake passing
must not satisfy real execution or durability acceptance criteria.

Include a final acceptance checklist, unresolved blockers and an exact amendment list for
existing consumer tasks. Stop for review; leave all implementation checkboxes unchecked.
```

## 11. Final cross-spec audit prompt

```text
Audit the foundation, four new prerequisite specs and guided workflows together.
Read their requirements/design/tasks and the reviewed interface/amendment ledger.
Do not edit specs or implementation during this audit.

Check:
1. Every decision and shared contract has exactly one owner.
2. Every consumed contract has a producer, concrete schema/type and conformance fixture.
3. No generic prerequisite task is also implemented independently in guided workflows.
4. Requirements, designs and task DAGs are mutually consistent and traceable.
5. Foundation schema versions, severity values, source coordinates and effect vocabularies
   are extended through one compatible proposal, not conflicting parallel definitions.
6. All authoritative mutations use the exact-plan transaction interface.
7. Candidate identity, administrative revision and evidence/receipt digests form no cycles;
   recording a receipt cannot invalidate itself or hide an implementation edit.
8. Unknown assessments, missing validators, stale policy and untrusted claims fail closed
   for mandatory gates, while allowing explicitly permitted evidence-gathering actions.
9. Graph conditions cannot bypass policy; retries are not static-DAG cycles; actions do
   not require their own output or prior acceptance to begin.
10. Test fakes, manual evidence, external CI and future process execution have explicit
    trust/availability rules; unavailable capabilities are not claimed as implemented.
11. CLI human/JSON behavior remains understandable, deterministic and non-interactive-safe.
12. Later UX, host integration, index, execution and release scopes remain reserved.

Walk these end-to-end scenarios and cite exact requirements/tasks at each handoff:
- Start a work item, inspect a dry-run, authorize and apply its exact plan.
- Reject a stale plan and recover a crash after commit intent without losing user edits.
- Retry a committed request whose response was lost.
- Submit malformed then corrected artifacts without bypassing authority or losing history.
- Escalate Quick Spec risk while preserving required gates and accepted evidence.
- Complete Bug Fix with real candidate-bound regression evidence and independent review.
- Reject a changed candidate between verification and finish.
- Resolve conflicting promotion while preserving another work item's accepted changes.
- Resume a compound workflow after partial success without replaying successful effects.

Report P1/P2 findings with path/section/requirement/task IDs, failure scenario, single owner,
proposed amendment, consumer impact and test. List unresolved decisions and missing specs.
Do not claim the set is non-overlapping solely because folder names differ.
```

## 12. Acceptance checklist for this drafting batch

- [ ] Four separate prerequisite specs; existing foundation/guided specs remain recognizable consumers/owners.
- [ ] Requirements, design and tasks reviewed as separate phases; no invented acceptance.
- [ ] Every normative requirement has a stable ID and observable verification.
- [ ] Shared-interface ledger has one owner, producers/consumers, trust/freshness rules and fixtures.
- [ ] Known source contradictions have decisions or explicit phase-blocking amendment proposals.
- [ ] Generic implementation tasks removed/delegated through reviewed guided-spec amendments, not duplicated.
- [ ] Local and cross-spec task dependencies are defined and acyclic.
- [ ] Snapshot, graph, artifact, evidence and transaction identities fit together without self-reference.
- [ ] Failure/cancellation/retry/concurrency/crash/malicious-input cases have testable outcomes.
- [ ] Schema examples and contract tests distinguish actually executed checks from future test plans.
- [ ] Missing execution/trust/platform capabilities are honest blockers, not fake-backed production claims.
- [ ] The later five specs remain reserved, and no implementation checkbox is marked done from documentation alone.

## For the chosen ones

The strongest non-overlap test is to ask which module may make a decision and which module may commit its consequences. Completion decides whether the evidence permits finishing; workspace commits the authorized effects; graph exposes what can happen next. If two modules can independently decide that work is done, they overlap even when their types and folders differ. If nobody can make that decision without calling back into its own consumer, the decomposition contains a hidden cycle. Review authority and dependency together, not just package names.
