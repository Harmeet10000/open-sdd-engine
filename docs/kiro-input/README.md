# Kiro specification input pack

This directory contains bounded context for drafting `sdd` specs in Kiro. It is input, not generated Kiro output and not a second product authority.

## Files Kiro should receive

Kiro automatically receives the concise foundation steering files:

- `.kiro/steering/product.md`
- `.kiro/steering/tech.md`
- `.kiro/steering/structure.md`

Kiro should auto-load these focused steering files when the spec or task touches their described surface:

- `.kiro/steering/cli-ux.md` for commands, flags, output, diagnostics, help, prompting, exit codes, or agent interactions;
- `.kiro/steering/go-engineering.md` for Go packages, data-driven/data-oriented representations and passes, ROP-style error flow, concurrency, dependencies, performance, or testing.

When drafting or reviewing a spec, manually include:

- `.kiro/steering/spec-authoring.md`

That steering file loads:

- `docs/kiro-input/sdd-cli-spec-brief.md`
- `docs/architecture/go-project-structure.md`

Use `docs/kiro-input/spec-prompts.md` to create the specs in the recommended order.

Load `docs/architecture/guided-workflows.md` only when drafting or reviewing `sdd-guided-workflows`, Bug Fix, Quick Spec, routing composition, or their conformance fixtures.

Kiro should generate its normal outputs under `.kiro/specs/<spec-name>/`:

- `requirements.md` for feature requirements;
- `design.md` for architecture and implementation design;
- `tasks.md` for vertical, testable work;
- or `bugfix.md` instead of `requirements.md` when using a bugfix spec.

Do not preload every file under `notes/`, `docs/research/`, or `vendor/`. Those sources contain historical and conflicting proposals. Consult a research source only to resolve a named question and preserve its provenance.

## Recommended workflow

1. Start Kiro in the repository root so workspace steering is discovered.
2. Invoke the manual `spec-authoring` steering context.
3. Confirm that `cli-ux` and `go-engineering` were included for the foundation spec; add them manually if auto inclusion did not trigger.
4. Paste the first prompt from `spec-prompts.md`.
5. Choose Requirements-First for the foundation and safety-sensitive specs.
6. Review `requirements.md`, then run Kiro’s Analyze Requirements step.
7. Resolve contradictions and unknowns before approving design.
8. Review `design.md` against the architecture document and product non-goals.
9. Require tasks to deliver end-to-end observable slices through `Engine.Execute`.
10. Do not run all implementation tasks until the spec’s dependencies and repository baseline are actually ready.

The full product plan is intentionally not always-on steering. It is too large and contains proposed future work that would crowd out the current spec.
