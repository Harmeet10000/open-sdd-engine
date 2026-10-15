---
inclusion: always
---

# Product: `sdd`

`sdd` is an agent-neutral, deterministic workflow control plane for specification-driven software changes.

The user gives `sdd` a change intent. The CLI chooses proportional rigor, exposes the next valid action to any coding agent, validates artifacts and evidence, prevents unsafe or stale mutations, and preserves an auditable path from intent to verified code.

## Primary users

- developers using Codex, Kiro, Claude, Cursor, Copilot, or another coding agent;
- reviewers who need traceability from requirements to code and evidence;
- platform/security teams that need enforceable policy without owning an LLM runtime;
- maintainers importing existing OpenSpec, Spec Kit, or Kiro workspaces.

## Primary experience

```text
sdd init
sdd start "intent"
sdd next
sdd status
sdd check
sdd finish
```

Users should not need to understand artifact graphs, transaction recovery, digests, claims, adapter manifests, context budgets, or knowledge indexes during ordinary work.

Two maintained guided presets are part of the product surface:

- `bug-fix`: investigate and reproduce/observe before evidence-backed diagnosis, regression contract, fix, and verification;
- `quick-spec`: resolve material clarifications, commit an inspectable delivery scope, then produce and validate a linked requirements/design/tasks draft bundle with minimal interaction.

These are journey shapes, not risk levels. The engine independently resolves `quick`, `standard`, or `critical` rigor and may escalate a preset without discarding already accepted answers or artifacts.

## Product invariants

- No lost update.
- No implicit external side effect.
- No false or unverifiable completion.
- No silent artifact drift.
- No hidden policy override.
- No silent scope reduction or priority reclassification.
- No workflow semantics in agent adapters.
- No machine consumer parses human prose to determine success.
- No unsafe recovery guess.
- No authority laundering from agent/research content into policy.
- No fast-path preset may weaken the applicable risk, evidence, verification, or completion floor.

## Non-goals for 1.0

- hosting or selecting LLMs;
- general-purpose agent orchestration;
- remote knowledge or extension marketplaces;
- automatic Git push, PR creation, merge, or conflict resolution;
- mandatory EARS/RFC/BDD for every change;
- arbitrary in-process plugins;
- web research providers in the core.

The current product decision document is `docs/cli-product-plan.md`. It is proposed until its ADRs are ratified. Research under `docs/research/`, historical notes, and vendored upstream material are evidence, not product authority.
