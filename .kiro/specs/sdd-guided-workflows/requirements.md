# Requirements Document: sdd-guided-workflows

## Introduction

This specification defines the Bug Fix and Quick Spec presets on top of the foundation seam and four prerequisite contracts. Presets are data-driven compositions of journey topology, policy-derived rigor, artifact/validation profiles, and a versioned definition digest. They are conveniences, never authority. The engine remains native Go, agent-neutral, deterministic for structural/policy decisions, and model-free.

## Glossary

- **Journey**: a maintained action graph topology (`bug-fix` or `quick-spec`).
- **Rigor**: policy/risk-derived evidence and verification level (`quick`, `standard`, `critical`).
- **Artifact**: a typed, revisioned value with stable ID and provenance; adapter filenames are not semantics.
- **Verification artifact**: an independent record of checks executed against a candidate snapshot and their results.
- **Regression contract**: an observable protection definition; it is not the later verification result.
- **Draft bundle**: requirements, design, and tasks submitted atomically.
- **Implementation lifecycle**: the shared handoff from an accepted plan to implement, verify, policy-review, and completion.
- **Assessment**: a human/agent claim with evidence, provenance, confidence, and uncertainty; it cannot replace deterministic validation.
- **Scope baseline**: a pinned delivery-window commitment containing the included subjects, their MoSCoW categories, capacity evidence where available, and its decision/policy digest.
- **MoSCoW category**: `must`, `should`, `could`, or `wont` for one named scope baseline; it is not severity, risk, dependency readiness, approval, or feature-flag state.

## Requirements

### Requirement 1: Preset composition model

1. `sdd start --workflow bug-fix|quick-spec` SHALL resolve a complete journey, policy-derived rigor, artifact/validation profiles, and definition digest.
2. The definition digest SHALL bind exact versioned workflow, policy, schema, and validator definitions, not only their names.
3. Policy precedence SHALL be explicit: trusted required organization policy, project policy that may narrow it, then work rationale. Missing, stale, expired, revoked, or unverifiable required organization policy SHALL fail closed.
4. Routing and escalation SHALL preserve immutable work ID while allowing slug and revision to change. A preset cannot lower the applicable policy floor.
5. Risk assessment SHALL distinguish `unassessed`, `insufficient`, and `assessed`. Missing input SHALL not silently select quick rigor or automatically classify every start as critical. While unassessed, return an external assessment/evidence action and permit only policy-approved preparation; block implementation/approval until the resolved floor is known. An organization may explicitly configure a conservative floor for uncertainty.

### Requirement 2: Bug Fix journey topology

The authoritative happy path SHALL be:

```text
report
  -> reproduce/observe
  -> investigate
  -> diagnose
  -> define-regression-contract
  -> conditional-design
  -> tasks
  -> implement
  -> verify
  -> policy-review
  -> finish
```

1. `verify` SHALL accept the accepted implementation, candidate snapshot, and regression contract as inputs, then produce a fresh verification artifact bound to that candidate snapshot. It SHALL not be satisfied by the contract itself or an agent claim.
2. A regression contract SHALL be defined before implementation, while passing-after verification evidence is produced only after implementation. The two artifacts remain distinct.
3. Normal implementation SHALL require supported diagnosis and regression contract; design is required only when resolved policy/profile says so. A scoped, preauthorized emergency repair is the only exception and still requires later reconciliation.
4. `policy-review` is a real gate before finish whenever policy/profile requires it. Each node has explicit inputs, outputs, and gate diagnostics.

### Requirement 3: Bug report artifact

The bug report SHALL contain expected/observed behavior, impact, environment, first-known version, source/provenance, and stable artifact ID; it SHALL record a symptom without asserting a cause. Structural validation accepts a report without diagnosis.

### Requirement 4: Reproduction/observation states

1. Status is one of `reproduced`, `intermittent`, `environment-specific`, `not-reproduced`, or `unsafe-to-reproduce`, with attempts, evidence, environment, and snapshot/provenance.
2. `not-reproduced` is an observed unsuccessful reproduction result; `unsafe-to-reproduce` is a safety disposition. Neither implies a pass or diagnosis. They block diagnosis unless policy accepts bounded evidence acquisition: safe proxy/simulation, incident evidence, or a scoped exception. A proxy must itself be safety-reviewed; it cannot be fabricated by marking `passed`. Risk/materiality assessment has its own explicit `unassessed` state.
3. Intermittent/environment-specific observations may proceed only with their differentiating evidence and policy limits.

### Requirement 5: Diagnosis gate

Diagnosis SHALL include a concrete mechanism, evidence tied to paths/revisions/commands/traces/observations, discriminating facts for plausible alternatives, affected and explicitly unaffected scope, confidence, unresolved uncertainty, and the accepted-spec relationship (`implementation-defect`, `specification-defect`, or `intentional-change`). A passing reproduction is resolution evidence, not root-cause truth.

### Requirement 6: Investigation effects disposition

Probes, logs, flags, reduced cases, and instrumentation SHALL be tracked as effects. Before finish, each is retained with requirements/review, removed with evidence, or an unresolved blocker. A retained temporary effect cannot disappear silently into an implementation diff.

### Requirement 7: Emergency repair path

1. Emergency repair requires policy preauthorization and a typed exception: identified actor, grant scope, authorization evidence, policy snapshot digest, exact work/snapshot scope, issued/expiry, revocation state, and preauthorization checks. An owner string and `approved: true` are insufficient.
2. The exception captures the starting workspace/work revision, candidate/input/policy/workflow digests, incident constraint, and deployment/completion distinction.
3. It may bypass incomplete reproduction/diagnosis and pre-fix regression authoring only for the scoped mitigation and immediate verification, with those exact exemptions recorded. It may not bypass required policy review, candidate recheck, or eventual retrospective reproduction/diagnosis/regression contract/spec convergence before `finish`. Expiry or revocation stops further exceptional actions but does not prevent safe read-only investigation or reconciliation through normal gates.
4. A same-user local CLI and mutable files are not tamper-resistant; a trusted CI or declared authorization boundary is required for a claim of organizational approval.

### Requirement 8: Quick Spec journey topology

The authoring subjourney SHALL be:

```text
intent
  -> discover-material-unknowns
  -> assess-materiality-and-risk
  -> answer-clarifications (zero questions is valid)
  -> generate-spec-bundle
  -> validate-bundle
  -> approval
  -> handoff-to-shared-implementation-lifecycle
```

1. The handoff enters the shared lifecycle at `implementation-ready -> implement -> verify -> policy-review -> finish`; Quick Spec SHALL not transition `implementation-ready` directly to `finish`.
2. Generation is not acceptance. Bundle rejection emits structured reasons and returns to authoring with a new candidate revision; revision requires revalidation and reapproval. Approval is explicit and tied to the exact bundle/plan digest.
3. The engine asks only material questions not answerable from trusted repository, accepted-spec, configuration, or allowed read-only evidence. When evidence cannot determine materiality or risk, it emits an `assess-materiality-and-risk` action for a human/agent assessment with evidence/provenance/uncertainty; it does not fabricate a deterministic answer.

### Requirement 9: Material clarification definition

A question is material when different answers change observable behavior, compatibility, migration/destructive behavior, trust/privacy/authorization, architecture boundary, task ordering, verification, or rollout/rollback/ownership. All questions have stable IDs and evidence/provenance; heuristics may propose but cannot silently decide materiality when policy requires review.

### Requirement 10: Generation without embedded model

`generate-spec-bundle` emits accepted intent, clarification records, relevant accepted specs, trusted policy snapshot, repository evidence, artifact schemas/stable-ID rules, canonical output paths, expected workspace/work/action revisions, a resolved snapshot reference, content digests for every context item, the pinned scope baseline, budgets, provenance, validator definition digests, and completion conditions. Context is bounded: overflow produces bounded digest-only references plus a structured remediation/action, never silent truncation or unbounded output. A host agent/human produces the files; the core invokes no model/provider/prompt and remains offline usable.

### Requirement 11: Atomic draft bundle

1. Submission SHALL contain requirements, design, and tasks with stable IDs, source/work revision, and blob/semantic/derivation digests.
2. Design SHALL record requirements derivation; tasks SHALL reference requirements/design and record derivations, dependency IDs, implementation/verification coverage.
3. Validation derives authoritative mandatory validators from trusted policy and candidate/profile. It executes them and records `validatorsExecuted` (validator ID, definition digest, candidate digest, status, evidence). `validator.passed` supplied by an agent is provenance only. Missing, stale, cancelled, unverifiable, or empty mandatory evidence is not a pass.
4. The transaction accepts all artifacts or none. Same submission key and same content replays the original outcome; same key and different content is a conflict; an uncertain crash returns recovery-required.
5. The bundle is `draft` until required approval, then `accepted`; generated/reviewed is never accepted/implemented/verified.

### Requirement 12: Derivation staleness and canonicalization

Requirements edits stale design and affected tasks; design edits stale affected tasks. Typed/versioned semantic canonicalization preserves meaningful Markdown whitespace, code/order, and unknown fields; blob, semantic, and derivation digests remain distinct. Semantic no-op formatting need not stale dependents; behavior/unknown changes conservatively stale them. Finish is blocked while stale.

### Requirement 13: Quick Spec escalation

1. Trust-boundary/security, migration/destructive/rollback, public compatibility, concurrency/distributed/crash recovery, money/privacy/regulation/blast radius, cross-repository scope, material architecture uncertainty, inadequate verification, or organization policy floor SHALL emit `WORKFLOW_ESCALATION_REQUIRED`.
2. The resolved composition derives from policy/profile plus risk and retains journey unless policy requires a different journey. Escalation preserves work ID, accepted answers, history, valid artifacts, and artifact IDs while creating a new work revision with new composition and definition digest.
3. The escalated profile/gates are explicit; an agent cannot choose `?` gates or lower rigor.

### Requirement 14: Stable diagnostics catalog

The versioned catalog SHALL include `BUG_REPRODUCTION_REQUIRED`, `BUG_REPRODUCTION_UNASSESSED`, `BUG_DIAGNOSIS_UNSUPPORTED`, `BUG_REGRESSION_REQUIRED`, `BUG_VERIFICATION_REQUIRED`, `BUG_SPEC_CONFLICT_UNRESOLVED`, `BUG_TEMPORARY_PROBE_UNRESOLVED`, `QUICK_SPEC_CLARIFICATION_REQUIRED`, `QUICK_SPEC_BUNDLE_INCOMPLETE`, `QUICK_SPEC_BUNDLE_REJECTED`, `QUICK_SPEC_TRACEABILITY_INVALID`, `QUICK_SPEC_BUNDLE_STALE`, `QUICK_SPEC_APPROVAL_REQUIRED`, `WORKFLOW_ESCALATION_REQUIRED`, `WORKFLOW_POLICY_UNAVAILABLE`, and `WORKFLOW_RECOVERY_REQUIRED`. Each code declares severity, exit code, source, remediation, and retryability.

### Requirement 15: Bug Fix completion conditions

Finish SHALL validate reproduction/observation state, supported diagnosis or authorized mitigation plus retrospective reconciliation, accepted-spec convergence, regression contract, independent verification artifact on the final candidate, broader policy checks, effect disposition, and residual uncertainty in the completion record. The shared completion owner validates inputs before generating `done.md`, receipt, and lifecycle events; those outputs are generated atomically in one completion transaction only after inputs pass. Guided handlers do not write completion outputs directly.

### Requirement 16: Quick Spec completion conditions

Finish SHALL require accepted and fresh requirements/design/tasks, resolved escalation, implementation and independent verification against the accepted requirements revision, policy review when selected by policy/profile, a candidate snapshot unchanged since verification, and a valid scope baseline. Every applicable `must` subject in that baseline SHALL be satisfied unless an authorized scope-baseline revision changes the commitment. A `should` or `could` may be deferred only with an explicit disposition, workaround/follow-up, and baseline reference; a `wont` is neither delivered scope nor an unfinished task. The shared completion owner SHALL validate inputs first, then atomically generate `done.md`, machine receipt, and completion events. It SHALL not require pre-existing done/receipt files as input and SHALL not treat generated/reviewed/accepted/implemented/verified as synonyms.

### Requirement 17: Routing and override

Routing records evidence and reason. Explicit preset selection may choose an alternative or increase rigor; it cannot lower the policy floor. Composition identity includes versioned workflow/policy/profile/validator definitions and digests.

### Requirement 18: Orthogonal composition persistence

Persist immutable work ID, revision, preset (if any), journey, rigor, artifact profiles, schema version, exact definition/policy/validator digests, routing/escalation reasons, and source evidence. Canonical persistence is under `.sdd/work/<slug>/work.yaml`; Kiro adapter paths are not authority.

### Requirement 19: Human and JSON rendering

Guided outcomes use the foundation envelope and shared renderer. Human output reports evidence-bearing stage, blocking attention, and one next command. When scope priority is material, default human output says whether the committed core is on track and names only the required scope-replan/defer action; detailed category/budget evidence is available through inspection rather than routine output. JSON contains structured actions/arguments, IDs, revisions, digests, diagnostics, and `scopeBaselineRef`/priority summary; no shell-only strings. Shared `start`, `next`, `submit`, and `finish` payloads carry expected/action revisions and idempotency/claim fields according to mutation policy; `submit` discriminates an atomic draft bundle, materiality/risk assessment, scope-baseline decision, or approval. Read-only `next` accepts selection/work ID and returns the revisions needed for a later submit. stdout/stderr, cancellation, redaction, and exit codes follow foundation policy.

### Requirement 20: Conformance fixtures

The suite SHALL include: all reproduction states; symptom fixed but diagnosis unsupported; implementation/spec defect; critical auth bug retaining Bug Fix with critical rigor; emergency repair and retrospective gate; retained probe; zero/one/several clarifications; unavailable host/manual packet; partial/malformed/duplicate/untraceable/stale bundles; same-key changed request; uncertain crash; changed workflow definitions; forged validator claims; dirty post-verification code; expired/revoked/forged exception; policy removal; unsafe reproduction proxy; Quick escalation preserving artifacts; must-inflation over a declared capacity cap; a manually workable item classified as `should`; missing capacity reported as unverified rather than silently budgeted; `could` then `should` deferral under a revised baseline; unauthorized `must` demotion rejection; `wont` excluded-but-visible scope; feature-flag rollout that does not satisfy an unimplemented Must; completion atomicity; and full journeys end-to-end.

### Requirement 21: MoSCoW scope governance

1. Quick Spec SHALL create or consume one versioned scope baseline before a
   draft bundle is approved. Each included requirement, scenario, task, or
   initiative slice has a typed `ScopePriority` decision with stable subject
   ID, delivery-scope reference, category, rationale, evidence, policy/baseline
   digest and decision provenance. The only categories are
   `must|should|could|wont`. A baseline SHALL classify a stable subject at most
   once; it SHALL NOT infer a category from parent/child task hierarchy or
   duplicate the subject across categories.
2. A `must` is non-negotiable for the named scope only when the core workflow,
   declared safety/legal/compliance/operational release condition fails without
   it and no acceptable bounded workaround exists. A workable manual or
   alternative path normally classifies the item as `should`; an exception must
   record the policy or operational reason the workaround is unacceptable.
3. With declared comparable estimates and usable delivery capacity, the default
   profile SHALL cap `must` work at 60% of capacity. The declared capacity and
   estimates SHALL use one pinned unit/conversion rule with no double-counted
   shared work. Policy may replace that cap only through a pinned rule. Missing
   estimates/capacity produce an explicit `scope-budget-unverified` fact and a
   next action; they SHALL NOT silently
   make all work Must or pretend that the cap passed.
4. A `should` requires a bounded workaround/owner/operational limit and
   reconsideration target; a `could` records benefit and selection condition; a
   `wont` records exclusion rationale and reconsideration trigger or target
   scope. `wont` means not in this scope, not never. Evidence that a `wont`
   item was intentionally delivered is a scope change requiring a new baseline,
   not invisible extra credit.
5. Scope reduction is an explicit baseline revision: defer `could` before
   `should` when capacity contracts, preserve the old baseline/history, and
   return a scope-replan action when the `must` floor cannot be met. A revision
   SHALL bind the predecessor digest and expected scope revision. The CLI SHALL
   NOT relabel a `must` automatically; its change requires the authority
   governing that scope.
6. MoSCoW SHALL NOT replace dependency order, validation, policy/approval,
   risk rigor, or technical readiness. Feature flags are rollout records with
   cohort, fallback, observability, owner, and removal/review conditions; a
   flag cannot make an unimplemented Must complete. RICE/WSJF-style scores may
   be retained as advisory evidence but cannot mechanically determine category.

## Dependencies and release gates

This spec consumes `sdd-engine-foundation`, `sdd-workspace-transactions`, `sdd-workflow-graph`, `sdd-artifact-validation`, and `sdd-completion-lifecycle`. Until each real contract and conformance suite exists, draft adapters/fakes are allowed but durable transactions, graph routing, authoritative validation, and completion release claims are blocked. No automatic Git commit/PR/merge, embedded model, arbitrary performance, or Bun comparison is invented here.

## Open questions retained for implementation planning

- Which exact schema and byte/token budgets govern `generate-spec-bundle`, and which digest-only reference shape is used on overflow?
- Which roles or trusted system distribute and revoke organization policy and approve exceptions?
- What is the approval matrix for Quick Spec bundles at each rigor/profile, including when policy-review is optional?
- Which external evidence sources may supply materiality/risk assessments, and what provenance/retention rules apply?
- Which filesystems, Git versions, and OS/architecture combinations are supported once the future transaction and integration specs land?
- What retention/redaction policy applies to verification output, context packets, receipts, and exception evidence?

## References

- `docs/kiro-input/sdd-cli-spec-brief.md`
- `docs/architecture/guided-workflows.md`
- `docs/architecture/go-project-structure.md`
- `.kiro/steering/cli-ux.md`
- `.kiro/steering/go-engineering.md`
