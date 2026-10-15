---
inclusion: manual
---

# Authoring `sdd` specifications in Kiro

Use this steering file only while creating or reviewing Kiro specs for `sdd`.

Load these source documents:

#[[file:docs/kiro-input/sdd-cli-spec-brief.md]]

#[[file:docs/architecture/go-project-structure.md]]

When the spec touches CLI behavior, load `.kiro/steering/cli-ux.md`. When it touches Go design or implementation, load `.kiro/steering/go-engineering.md`. Most foundation specs require both. These focused files are engineering policy, not substitutes for feature requirements.

When drafting `sdd-guided-workflows` or reviewing Bug Fix/Quick Spec behavior, also load `docs/architecture/guided-workflows.md`. Do not load that specialized document for unrelated specs.

When drafting or reviewing Quick Spec delivery scope, MoSCoW records, scope-baseline changes, or completion treatment of deferred work, also load `docs/architecture/moscow-prioritization.md`. MoSCoW is a scoped commitment contract; do not use it to replace dependency, risk, readiness, or authority rules.

The complete decision source is `docs/cli-product-plan.md`; consult only the relevant section when the brief explicitly points to it. Do not load all research notes or vendored documents into the drafting context.

## Drafting rules

1. Create several bounded specs, not one repository-wide mega-spec.
2. Start with the engine/protocol foundation before persistence, workflow graph, validation, completion, adapters, indexing, or parallel execution.
3. Draft requirements before implementation tasks unless a narrow technical feasibility spike is explicitly selected.
4. Use EARS only where trigger/state/response syntax improves a deterministic behavior. Use tables, schemas, invariants, examples, and prose where they fit better.
5. Separate user-visible requirements from architecture constraints.
6. Every requirement needs a stable ID and observable verification method.
7. State non-goals and authority boundaries explicitly.
8. Include failure, cancellation, retry, stale-revision, crash-recovery, cross-platform, and malicious-input scenarios when applicable.
9. Design against `Engine.Execute`, public JSON schemas, and the `.sdd/` format. Do not make internal package names public contracts.
10. Tasks must be vertical, independently verifiable slices; avoid “build all models,” then “build all repositories,” then “build all handlers.”
11. Do not implement code while drafting the spec unless the user explicitly starts implementation.
12. CLI acceptance criteria must cover human and JSON rendering, stdout/stderr separation, stable diagnostic/exit behavior, non-interactive use, and a useful next action.
13. Go designs use explicit data transformations and success/failure flow without pointer-rich object graphs, generic repository/service layers, premature ECS/SoA machinery, or a `Result` monad abstraction.
