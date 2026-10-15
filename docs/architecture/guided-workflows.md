# Guided workflows: Bug Fix and Quick Spec

**Status:** Proposed architecture input  
**Decision source:** `docs/cli-product-plan.md`, revision 6  
**Audience:** workflow implementers, Kiro spec authors, agent-adapter authors, and reviewers  

## 1. Product contract

Add two maintained presets without adding another command family:

```text
sdd start "login fails after token refresh" --workflow bug-fix
sdd start "add CSV export" --workflow quick-spec
sdd next
sdd status
sdd check
sdd finish
```

If the user omits `--workflow`, routing may select either preset from repository evidence and the change intent. The outcome explains the selection and permits an explicit alternative, subject to policy.

The presets hide artifact orchestration but do not hide material uncertainty, risk escalation, failed evidence, or pending approval.

## 2. Composition model

Keep these concepts orthogonal:

```text
resolved workflow = journey topology
                  + risk-derived rigor
                  + artifact/validation profiles
                  + organization/project policy floor
```

Examples:

| User-facing preset | Journey | Initial rigor | Possible additions |
|---|---|---|---|
| Bug Fix | `bug-fix` | selected from evidence; often quick/standard | regression, threat model, migration, protocol, performance |
| Quick Spec | `quick-spec` | quick only while within its risk ceiling | protocol or performance only if still bounded; otherwise escalate |

The preset is a convenience, not authority. Explicit selection cannot weaken an organization policy floor. A critical authentication bug stays in the Bug Fix journey while receiving critical rigor and security artifacts. A Quick Spec whose risk ceiling is exceeded keeps its work ID, accepted answers, and valid artifacts while changing to the proposed standard or critical composition.

The resolved composition is persisted with version and reasons:

```yaml
workflow:
  preset: bug-fix
  journey: bug-fix
  rigor: critical
  artifactProfiles: [regression, threat-model]
  schemaVersion: 1
  definitionDigest: sha256:...
  reasons:
    - defect crosses an authentication trust boundary
```

## 3. Bug Fix

### 3.1 User promise

Bug Fix turns an observed defect into an evidence-backed resolution. It separates what was reported, what was observed, what is hypothesized, what is diagnosed, what changed, and what was verified.

The happy-path action graph is:

```text
bug-report
    -> reproduce-or-observe
    -> investigate
    -> diagnose
    -> define-regression-contract
    -> design? -> tasks
    -> implement
    -> verify-regression-and-safety
    -> review?
    -> finish
```

`?` means selected by resolved rigor/profile policy, not agent preference.

### 3.2 Canonical artifacts

| Artifact ID | Purpose | Required content |
|---|---|---|
| `bug-report` | Preserve the claim without assuming a cause | expected behavior, observed behavior, impact, environment, first-known version, source/provenance |
| `reproduction` | Record whether and how the symptom was observed | status, steps/probe, attempts, observed result, environment/snapshot, captured evidence |
| `investigation` | Preserve facts and competing hypotheses | observations, hypothesis IDs, supporting/contradicting evidence, ruled-out explanations, temporary probes |
| `diagnosis` | State the supported failure mechanism | cause/mechanism, affected scope, confidence, discriminating evidence, residual uncertainty |
| `regression-contract` | Define the observable protection | linked requirement, failing-before/passing-after evidence or equivalent oracle, negative/safety cases |
| `design` | Explain a boundary or durable decision change | only when selected by risk or scope |
| `tasks` | Bound implementation and verification work | traceability to diagnosis/regression IDs and rollback/cleanup work |
| `verification` | Bind results to the candidate snapshot | executed commands/checks, exit/result state, environment, output digests, snapshot digest |

The Kiro adapter may combine the bounded report/diagnosis view into Kiro's `bugfix.md`. The canonical engine retains typed artifacts and stable IDs; it does not make an adapter filename part of workflow semantics.

### 3.3 Reproduction states

Use a closed state set:

| State | Meaning | May proceed to diagnosis? |
|---|---|---|
| `reproduced` | Deterministic or adequately controlled observation matches the report | yes |
| `intermittent` | Symptom observed but not deterministically | yes, with frequency/environment evidence and policy limits |
| `environment-specific` | Symptom isolated to a named environment/configuration | yes, with captured differentiators |
| `not-reproduced` | Attempts did not observe the symptom | only if other evidence meets policy or a scoped diagnostic/mitigation path is accepted |
| `unsafe-to-reproduce` | Reproduction would cause unacceptable harm | only through an approved safe proxy, simulation, incident evidence, or exception |

`not-reproduced` is a result, not an error and not a success. The engine must not convert it into “test passed.”

### 3.4 Diagnosis gate

A diagnosis is ready only when it includes:

- a concrete failure mechanism rather than a restatement of the symptom;
- evidence tied to paths, revisions, commands, traces, or captured observations;
- at least one discriminating fact that weakens a plausible competing hypothesis, when alternatives exist;
- affected and explicitly unaffected scope;
- confidence and unresolved uncertainty;
- the accepted-spec relationship: implementation defect, specification defect, or intentional behavior change.

A patch that makes the immediate reproduction pass is resolution evidence, not automatically root-cause evidence.

### 3.5 Investigation changes

Temporary probes, logs, flags, reduced test cases, or instrumentation are tracked as investigation effects. Before finish they must be:

- deliberately retained as supported observability with requirements and review;
- removed with evidence; or
- listed as an unresolved blocker.

The CLI must not allow a temporary probe to disappear into the implementation diff without disposition.

### 3.6 Emergency repair

An organization may authorize a code-first hotfix when delay is more dangerous than incomplete diagnosis. This is an explicit policy route, not `--force`:

1. capture incident constraint and exact starting snapshot;
2. record the scoped exception and owner;
3. plan and authorize the mitigation;
4. verify immediate safety against the candidate snapshot;
5. complete retrospective reproduction/observation, diagnosis, regression contract, and accepted-spec convergence;
6. only then allow normal `finish`.

If policy permits deployment before step 5, deployment and SDD completion remain distinct states.

## 4. Quick Spec

### 4.1 User promise

Quick Spec converts a bounded intent into an implementation-ready requirements/design/tasks set with as few interactions as the evidence permits. It removes ceremony; it does not manufacture certainty.

The action graph is:

```text
intent
  -> discover material unknowns
  -> answer only blocking clarifications
  -> prioritize delivery scope
  -> generate-spec-bundle
       -> requirements
       -> design
       -> tasks
  -> validate bundle and derivation
  -> human/policy approval
  -> implementation-ready
```

### 4.2 Material clarification

A clarification is material when different answers would change one or more of:

- observable behavior or acceptance criteria;
- public compatibility;
- data migration or destructive behavior;
- trust, privacy, authorization, or failure recovery;
- architecture boundary or dependency choice;
- task ordering/parallel safety;
- verification method;
- rollout, rollback, or operational ownership.

Do not ask the user for a fact that can be established safely from the repository, accepted specs, configuration, or an allowed read-only tool. Record that evidence instead. If no material unknown remains, close clarification with zero questions and continue.

### 4.3 Generation without an embedded model

The engine emits a `generate-spec-bundle` action containing:

- accepted intent and clarification records;
- relevant accepted specs and project policy;
- targeted code/repository evidence;
- artifact schemas and stable-ID rules;
- output paths and expected revision;
- pinned delivery-scope baseline and digest;
- byte/token budgets and provenance;
- validators and completion conditions.

A host agent, external tool, or human produces the files. The engine validates and accepts the submission. No model SDK, provider key, prompt template, or probabilistic generation belongs in the core.

Without an integration, `sdd next` returns the packet as JSON and a human-readable instruction. The workflow remains fully usable offline.

Quick Spec uses the [MoSCoW scope-priority contract](moscow-prioritization.md)
to make a delivery commitment before bundle approval. Ordinary human output says
only whether the committed core is on track or needs a scope replan; detailed
categories, estimates, and evidence remain inspectable rather than becoming
ceremony on every command. Priority never substitutes for risk, dependency
order, policy, or technical readiness.

### 4.4 Atomic draft bundle

Requirements, design, and tasks remain separate artifacts but are submitted as one bundle transaction:

- all three artifacts name their stable IDs and source revision;
- design records the requirements derivation digest;
- tasks record requirement/design references and derivation digests;
- the transaction accepts all three or none;
- validation produces diagnostics without partially advancing workflow state;
- an idempotent retry with the same submission key returns the original outcome;
- the bundle is `draft` until required approval occurs.

An edit to requirements marks the design and affected tasks stale. An edit to design marks affected tasks stale. Pure formatting changes use semantic/derivation digest rules rather than blindly invalidating everything.

### 4.5 Escalation

Quick Spec cannot continue under quick rigor when current evidence reveals:

- security or trust-boundary change;
- persistent migration, destructive action, or difficult rollback;
- public API/protocol/compatibility change beyond a bounded additive case;
- concurrency, distributed coordination, or crash-recovery semantics;
- money, privacy, regulated data, or broad production blast radius;
- cross-repository change;
- unresolved architecture choice with material consequences;
- inadequate verification around the affected area;
- organization policy requiring standard or critical rigor.

The result is a structured `WORKFLOW_ESCALATION_REQUIRED` outcome with reasons and a proposed composition. Accepting it creates a new revision of the same work item; it does not discard accepted clarifications or valid artifacts.

## 5. Stable diagnostics

The exact catalog is versioned with the protocol. Initial codes should include:

| Code | Meaning |
|---|---|
| `BUG_REPRODUCTION_REQUIRED` | policy requires reproduction/observation evidence before the next action |
| `BUG_DIAGNOSIS_UNSUPPORTED` | diagnosis lacks required discriminating evidence or scope |
| `BUG_REGRESSION_REQUIRED` | no adequate failing-before/passing-after or equivalent regression oracle exists |
| `BUG_SPEC_CONFLICT_UNRESOLVED` | observed, implemented, and accepted behavior disagree without an explicit decision |
| `BUG_TEMPORARY_PROBE_UNRESOLVED` | investigation instrumentation lacks retain/remove disposition |
| `QUICK_SPEC_CLARIFICATION_REQUIRED` | a material unresolved question blocks generation |
| `QUICK_SPEC_BUNDLE_INCOMPLETE` | one or more required bundle artifacts are absent |
| `QUICK_SPEC_TRACEABILITY_INVALID` | IDs, derivation, task, or verification links are invalid |
| `QUICK_SPEC_BUNDLE_STALE` | an upstream artifact/revision changed after generation |
| `WORKFLOW_ESCALATION_REQUIRED` | resolved risk or policy exceeds the selected preset's allowed rigor |

Machine consumers match codes and structured fields, never message text.

## 6. Completion conditions

Bug Fix adds these conditions to the normal completion gate:

- reproduction/observation state is explicit;
- diagnosis or policy-authorized mitigation path is resolved;
- accepted-spec conflict is reconciled;
- regression contract is satisfied on the candidate snapshot;
- temporary investigation effects are disposed;
- relevant broader safety checks pass;
- residual uncertainty is visible in `done.md` and the receipt.

Quick Spec adds these conditions:

- the draft bundle was accepted according to policy;
- requirements, design, and tasks are fresh and traceable;
- any escalation has been resolved;
- implementation and verification match the accepted requirements revision;
- the pinned delivery-scope baseline is honored: every Must is delivered or an
  authorized revised baseline exists, while deferred Should/Could and excluded
  Won't-have scope remain visible in DONE;
- “generated” is never treated as “reviewed,” “accepted,” “implemented,” or “verified.”

## 7. Required conformance fixtures

At minimum, test:

- reproduced, intermittent, environment-specific, not-reproduced, and unsafe-to-reproduce bugs;
- symptom fixed while diagnosis remains unsupported;
- implementation defect versus accepted-spec defect;
- critical authentication bug retaining Bug Fix journey and receiving critical rigor;
- emergency repair that may deploy but cannot finish before reconciliation;
- temporary instrumentation accidentally retained;
- Quick Spec with zero material questions;
- one and several material clarification questions;
- host agent unavailable/manual packet flow;
- partial bundle, malformed bundle, duplicate IDs, missing traceability, and retry after uncertain submission;
- requirements edited after generation;
- semantic no-op edit versus behavior-changing edit;
- risk discovered before generation and during implementation;
- escalation that preserves work identity, answers, history, and valid artifacts;
- human and JSON rendering, stdout/stderr separation, cancellation, and no side effect from `status`/dry-run.

## 8. Non-goals

- A built-in LLM runtime or provider selection.
- A promise that every bug can be reproduced or assigned one root cause.
- Automatic acceptance of generated specifications.
- Lower verification standards because the word “quick” appears in a preset.
- A separate `sdd bug` or `sdd quick-spec` command tree.
- Kiro-specific filenames as canonical engine semantics.
- Mandatory ECS, event sourcing, or a new service boundary to implement the workflows.
