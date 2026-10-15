# Prompts for generating Kiro specs

Use the manual `spec-authoring` steering context before any prompt below. For the foundation spec, also ensure the focused `cli-ux` and `go-engineering` steering files are active.

## 1. Foundation spec

```text
Create a Requirements-First Feature Spec named `sdd-engine-foundation`.

Use the loaded SDD CLI specification brief and Go architecture document. Draft only the requirements phase first; do not implement code and do not draft tasks yet.

Apply the CLI UX policy to both human and agent consumers. Apply the Go engineering policy's data-driven/data-oriented transformations, compiled workflow constraints, and railway-style error semantics without introducing pointer-rich object graphs, generic service/repository layers, premature ECS/SoA machinery, or a Result monad.

Scope the feature to the native Go skeleton, public command/outcome/diagnostic/mutation-plan JSON Schemas, one deep Engine.Execute seam, version and doctor, read-only status, non-mutating start --dry-run, hermetic TypeScript parity fixtures, and native artifact smoke tests.

Treat the locked product decisions and first-spec exclusions in the brief as constraints. Include stable requirement IDs, observable acceptance criteria, invalid/ambiguous selection, JSON versus human output, exit codes, cancellation, incompatible protocol/workspace versions, empty-directory execution, build provenance, and no-side-effect requirements. Use EARS only when it improves a deterministic trigger/response. Record open questions instead of guessing company policy or support-matrix values.

After drafting requirements.md, stop and ask me to review it. Recommend Analyze Requirements before design.
```

## 2. Design phase review prompt

```text
Draft the design for the accepted `sdd-engine-foundation` requirements.

Use the target project structure but create only packages required by this slice. Center the design on Engine.Execute(ctx, Command) Outcome, a shallow CLI composition root, language-neutral JSON Schemas, and deterministic fake dependencies. Show dependency direction and the end-to-end flows for status and start --dry-run. Explain schema generation, canonical fixture normalization, build provenance, error/exit-code mapping, and how the design avoids making internal Go packages public contracts.

Include alternatives considered: continuing TypeScript/Node, Bun standalone executable, permanent Go+TypeScript hybrid, and Go parity migration. Do not select SQLite, a plugin system, a message bus, or final transaction storage in this design. Stop for review before generating tasks.

Show where errors cross from internal Go values into stable Outcome diagnostics and exit codes. Include a rendering boundary that prevents logs, progress, or ANSI output from contaminating JSON stdout.
```

## 3. Task-generation prompt

```text
Generate implementation tasks for the accepted `sdd-engine-foundation` requirements and design.

Tasks must be vertical, independently verifiable slices through the public engine/CLI behavior. Begin with frozen protocol fixtures and a native version/help artifact, then add doctor, status, and dry-run start one slice at a time. Each task must name the requirements it satisfies, files/modules it is expected to touch, exact verification commands or fixtures, and rollback/deletion expectations.

Do not create horizontal tasks such as “build all models,” “build all adapters,” or “write all tests later.” Do not implement code. Identify task dependencies explicitly and keep the repository green after every top-level task.
```

## 4. Later-spec template

```text
Create a Requirements-First Feature Spec named `<SPEC-NAME>` using the corresponding row in the specification brief’s recommended decomposition table.

Load only the accepted contracts from prerequisite specs plus the relevant product-plan sections. Draft requirements first and stop for review. State prerequisites, in-scope behavior, non-goals, authority boundaries, failure and recovery scenarios, cross-platform implications, observable verification, and migration/compatibility impact. Do not duplicate prerequisite requirements and do not implement code.
```

## 5. Requirement audit prompt

```text
Analyze the current requirements adversarially without editing implementation code.

Look for contradictions, undefined terms, hidden authority expansion, untestable claims, wrong-snapshot verification, failure/cancellation gaps, retry/idempotency gaps, cross-platform assumptions, filesystem/process-tree edge cases, malformed or hostile input, migration ambiguity, and requirements that expose internal package structure as a public contract.

For each finding, cite the requirement IDs involved, classify it as blocking or non-blocking, explain the concrete failure, and propose the smallest correction. Do not invent a finding merely to populate a category.
```

## 6. CLI UX audit prompt

```text
Review this spec against the active CLI UX steering policy without changing implementation code.

Walk each command through interactive human use, redirected output, CI, and an agent using --json. Check command vocabulary, target ambiguity, defaults, help, stdout/stderr placement, JSON-envelope consistency, stable diagnostics and exit codes, non-interactive prompting, dry-run/authorization semantics, progress cancellation, redaction, narrow terminals, no-color behavior, and the usefulness of nextActions.

For each gap, cite the affected requirement or design section and propose an observable acceptance criterion. Do not request visual ornament or additional flags unless they solve a demonstrated user problem.
```

## 7. Go DDP/ROP architecture audit prompt

```text
Review this design and task list against the active Go engineering steering policy.

Check whether package boundaries follow product capabilities and stable vocabulary; interfaces are owned by consumers at real effect seams; and each use case is an explicit bytes/files → decoded records → normalized records → indexed snapshot → compiled graph → diagnostics/plan/outcome transformation. Verify that workflow variability is data-driven but cannot bypass code-enforced safety. Check canonical ordered slices, stable-ID-to-index maps, snapshot ownership, hot/cold data separation, deterministic graph traversal, and runtime-only dense indexes.

Reject pointer-rich object graphs and per-record interface dispatch when a plain indexed pass is clearer. Also reject ECS, struct-of-arrays, pools, arenas, unsafe code, or custom binary formats unless representative benchmarks and CPU/heap profiles prove the need and the design records a deletion criterion. Verify that errors retain identity with errors.Is/errors.As, translate once into protocol diagnostics, distinguish domain-negative outcomes from operational failures, preserve cancellation, and do not use panic, string matching, log-and-return, Result/Either abstractions, or generic service/repository layers.

Also inspect determinism, goroutine ownership, map ordering, injected time/IDs, process-tree cleanup, transaction fault tests, fuzz targets, dependency admission, and OS-specific claims. Report only concrete violations or missing decisions with the smallest corrective change.
```

## 8. Bug Fix and Quick Spec feature prompt

```text
Create a Requirements-First Feature Spec named `sdd-guided-workflows`.

Load `docs/architecture/guided-workflows.md`. Assume the accepted engine, workspace transaction, workflow graph, artifact validation, and completion contracts as prerequisites. Draft requirements only and stop for review.

Specify journey topology, risk-derived rigor, artifact/validation profiles, and named presets as orthogonal data. Include two maintained presets:

1. Bug Fix: report → reproduce/observe → investigate → diagnose → regression contract → conditional design/tasks → fix → verify → finish.
2. Quick Spec: intent → material clarifications → one revision-bound requirements/design/tasks draft bundle → traceability/derivation validation → approval → implementation-ready.

Require Bug Fix to distinguish reproduced, intermittent, environment-specific, and not-reproduced outcomes; evidence-backed hypotheses and diagnosis; regression checks; accepted-spec conflict handling; temporary-instrumentation cleanup; critical-risk escalation; and explicit emergency/hotfix reconciliation before finish.

Require Quick Spec to ask only material questions, investigate repository facts when possible, keep generation separate from acceptance, atomically reject partial bundles, stale design/tasks after upstream edits, work without an embedded LLM, preserve context during escalation, and never weaken verification or completion policy.

Define human and JSON outcomes, stable action/diagnostic codes, revision and digest behavior, dry-run/read-only behavior, cancellation, idempotent retries, Kiro bugfix.md rendering as an adapter concern, and golden fixtures for every listed failure mode. Do not add new top-level commands: both presets use start, next, status, check, and finish. Do not implement code or draft design/tasks yet. Recommend Analyze Requirements before design.
```
