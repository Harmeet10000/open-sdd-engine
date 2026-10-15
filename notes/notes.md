Create ~/Projects/open-sdd-engine as Bun/TS monorepo with zod, vitest, commander CLI entry `sdd`, and `.kiro/` init templates for both loops
Implement Change loop: proposal/delta-spec parsers (ADDED/MODIFIED/REMOVED/RENAMED), linear task parser, archive merge into .kiro/domains/
Implement Spec loop: EARS requirements parser, design/Mermaid validator, Kiro tasks parser, DAG wave compiler with JSON graph
Build unified validate/status/instructions commands with --changes|--specs|--all, --strict, --json, diagnostic codes
Add `sdd new` wizard and `sdd route` heuristics (time budget, brownfield vs greenfield) to pick change vs spec loop
Document dual-loop model, OpenSpec delta format, Kiro EARS/DAG format, CLI reference, and when to use each loop
implement Change loop: proposal/delta-spec parsers, linear task parser, archive merge

Build unified validate/status/instructions commands
Implement Spec loop: EARS, design, tasks, DAG wave compiler

Add loop router and research integrations
also need to add google OKF, karpathy's autoresearch, dynamically selecting skills as required.
also need to clone and integrate openspec and spec kit with whats required. make a plan for this. refine it think about it what more can we add such that user isnt telling Agent everything to do. what 
add ponytail too

it should be deterministic where once genrated it will enforce the agent to work in that way with it subagents but at the same time dynamic enough for creating new loops/workflows as needed based on the situation. 
CLI should have validation for the genrated Specs and Behaviour(BDD) 
it should have validations for the schema for Specs(SDD).
it should ask user for clarifications as many times as required to remove doubt before finalising specs
it should auto discover skills and hand it over to subagents for specific task that require skills like context gathering, websearch through firecrawl, code search throght ast-grep, codegraph, graphify
there should be separate planner, implementer, reviewer agents.
should be able to use git, worktrres, commit necessary work, properly use git branches, make PRs resolve merge conflict in a safe way.
ask the user for his coding taste. coding guidelines, prefered languages, linters & formating rules. 