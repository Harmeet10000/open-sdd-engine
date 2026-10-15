# Deterministic agent workflows

`sdd workflow generate <id> --type spec` creates a locked workflow manifest whose
SHA-256 hash binds it to finalized requirements. `sdd workflow status <id>` reports
whether the requirements still match. A changed spec must generate a new workflow.

- **Planner** gathers context through CodeGraph, Graphify, ast-grep, Context7, and Firecrawl.
- **Implementer** follows the frozen plan with repository evidence and minimal-diff policy.
- **Reviewer** independently checks SDD schema, EARS, BDD, traceability, and evidence.

`sdd clarify` records every discovered ambiguity and blocks workflow generation until
all questions are answered. `sdd skills discover` finds local `SKILL.md` files without
executing them, keeping role handoffs explicit and auditable.

`sdd workflow handoff <id> --role planner|implementer|reviewer` emits the isolated
agent contract, including the role's permitted research skills and newly discovered
project-local skills. The engine does not execute a discovered skill implicitly. The
handoff also includes the project's saved coding profile, so an implementer receives
the same language, lint, formatting, test, and branch conventions as the reviewer.

For a situation that does not fit the Change or Spec loops, create a locked custom
workflow with `sdd workflow generate <id> --type custom --description "..." --phase ...`.
