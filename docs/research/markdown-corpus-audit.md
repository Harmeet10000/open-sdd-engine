# First-party Markdown corpus audit

**Status:** Audit evidence; non-normative  
**Audit date:** 2026-10-11  
**Scope:** Every Markdown file present under `notes/` and `docs/` at the start of the audit  
**Input files:** 20  
**Input size:** 419,584 bytes; 55,307 words; 8,287 lines  
**Baseline manifest digest:** `sha256:170b0d0f5a9e25744d20fe6700c58a417b7092580253252830215a41d75357a5`  
**Primary output:** `docs/cli-product-plan.md`, now revision 6 after the subsequent runtime/architecture and guided-workflow decisions  

## 1. Verdict

All 20 in-scope Markdown files were read in full and compared against the proposed CLI product model. The material contains strong ideas, but it did not previously distinguish current product contract from historical prototype documentation, research hypothesis, copied conversational output, or domain-specific example. That is the largest company-adoption risk in this corpus: an agent can retrieve a plausible old page and confidently execute the wrong command, write to the wrong directory, or infer authority that the page never had.

The refined direction is:

- keep `docs/cli-product-plan.md` as the proposed decision document until its ADRs are ratified;
- keep the OKF/protocol report and this audit under `docs/research/` as dated, non-normative evidence;
- extract reusable lessons from the remaining notes and drafts;
- mark old CLI, `.kiro`, dual-loop, fixed-agent, and automatic-Git pages as superseded;
- never import universal architectural mandates from the 3D/GIS example drafts;
- generate future command/schema/diagnostic reference from tested contracts;
- make search and context assembly exclude superseded pages by default.

This was a semantic documentation audit, not a claim that every external source cited in the notes was independently revalidated during this pass. Time-sensitive or weakly sourced claims are explicitly identified below.

## 2. Audit method

For each file, the audit evaluated:

1. **Intent:** research, product contract, prototype reference, backlog, example, or generated synthesis.
2. **Authority:** whether the file may currently direct users, agents, validators, or implementation.
3. **Portable insight:** what belongs in the CLI product model independent of one vendor or domain.
4. **Conflict:** where the file contradicts the refined product boundary, safety model, workspace, or command surface.
5. **Evidence quality:** primary versus secondary source, dated versus dynamic claim, and testable versus rhetorical assertion.
6. **Disposition:** retain, revise, extract then supersede, or archive.

The baseline digest is over a sorted `sha256sum` manifest. `docs/cli-product-plan.md` was expected to change as an audit output, and this report did not exist in the input set. Therefore a digest of the current directory should not be expected to equal the baseline digest.

## 3. Authority model

The documentation must use these authority classes:

| Class | Meaning | May direct engine/user behavior? |
|---|---|---:|
| Normative contract | Ratified policy, versioned schema/protocol, generated current reference, or accepted ADR | Yes, within its declared version and scope |
| Proposed decision | Design awaiting ratification, including the current product plan | Not as a silent default; only as an implementation proposal |
| Current explanation | Maintained rationale derived from normative contracts | No new behavior; explanation cannot override its source |
| Research evidence | Dated findings, hypotheses, comparisons, and experiments with provenance | No; it informs a decision |
| Historical/superseded | Prior behavior retained for migration and provenance | No; excluded from normal retrieval |
| Vendored evidence | Pinned upstream material describing another product | No, unless a ratified decision adopts it |

The source-precedence rule belongs in both documentation policy and context-pack assembly. File location alone is not authority, and a recent modification time is not approval.

## 4. Per-file coverage ledger

### 4.1 Product and prototype documents

| File | Baseline SHA-256 | Finding | Disposition |
|---|---|---|---|
| `docs/cli-product-plan.md` | `7743f21aac31471d4ec0d7668e460d2f33af1723c6cdff8f7c10a8a65d07d48e` | The strongest integrated product direction. It already separated deterministic engine behavior from agent judgment and hid advanced mechanisms behind six commands. It needed an evidence ledger, source precedence, fail-closed verification, extension trust decomposition, durable compound-workflow recovery, optional discovery, and documentation lifecycle rules. | **Keep as proposed decision; revised to revision 4.** Split ratified decisions into ADRs and generated reference before 1.0. |
| `docs/cli.md` | `c43433c04932907f3da866577efff6aec90e13f70ac54cd4ce917ba437ee6e5d` | Prototype command inventory exposes internal concepts as top-level commands: `new`, `validate`, `instructions`, `archive`, `compile-waves`, `route`, `skills`, `research`, and `autoresearch`. It contradicts the six-command primary surface and makes provider/runtime mechanics user responsibilities. | **Supersede.** Retain only as a migration map; future CLI reference must be generated from the parser and protocol fixtures. |
| `docs/agent-autonomy.md` | `13f61460eb3cea0dba4ab2b81d0f1a1818e1dc5ed8d48cd4adb24c194e0bf5a1` | Useful boundaries: capability selection must be explainable, must not install/run automatically, minimal-diff policy cannot suppress safety, and experiments are not permission grants. It still depends on obsolete `route`/`skills` commands and frames Ponytail/AutoResearch as product modules. | **Extract then supersede.** The principles now live under authority, complexity, and bounded-execution sections of the plan. |
| `docs/deterministic-agents.md` | `72ff9c708b6926ff647eb0cbae7c778072e1b60ef4666a08a69bad1f5fac2eec` | Stable hashes, clarification records, explicit skill discovery, and reproducible handoffs are valuable. Fixed Planner/Implementer/Reviewer roles, hardcoded provider names, completely frozen plans, and the CLI generating agent teams are too prescriptive. Implementation discovery must be able to revise active design with downstream staleness. | **Extract then supersede.** Keep role packets as optional views, not a built-in agent runtime. |
| `docs/dual-loop.md` | `69dccb61415494758a6e507e5892b8aabc647f851c5e20580cc1ffa13d7d4265` | Brownfield delta versus deeper design is a real distinction. A binary Change/Spec type, two-hour cutoff, description-length routing, `.kiro/domains`, and synchronous waves are fragile proxies that do not model actual risk. | **Supersede.** Replace with schema-defined profiles selected from observable risk signals and overridable with recorded rationale. |
| `docs/git-workflow.md` | `7796ff69e1e05960cf1ae777071fd212035f9736a3eb31261573483ae22ba3df` | Strong safe mechanics: argv rather than shell strings, explicit staged paths, clean-state checks, worktrees, `diff --check`, and no automatic ours/theirs conflict choice. The page wrongly makes branch/commit/PR management look like the default CLI lifecycle and assumes `gh`. | **Revise/extract.** Preserve it later as an opt-in Git how-to generated against current behavior; PR/push remain explicit external side effects. |
| `docs/kiro-operating-model.md` | `e370ed89a7abcd568565896e528e6927efdbd22d6e8b4823975e79202bf6c7be` | Requirements/design/tasks and downstream freshness are useful. Universal EARS, `.kiro/steering`, “recalled memory,” and Kiro-shaped artifacts are not agent-neutral contracts. | **Supersede.** Kiro becomes an import adapter; EARS is one requirement style, not the default for every behavior. |
| `docs/openspec-delta.md` | `ef22eb8d929e7e563885acbc7272786b568a91a49e9c7e2bacb02b9cadc551e2` | Delta operations and deterministic rename → removal → modification → addition ordering are valuable. The `.kiro/changes` location and Change-loop framing are obsolete, and textual ordering alone is not enough without three-way promotion and conflict detection. | **Revise/extract.** Preserve semantics in the typed delta/promotion engine and document from tested fixtures. |

### 4.2 Large architecture and requirements drafts

| File | Baseline SHA-256 | Finding | Disposition |
|---|---|---|---|
| `docs/architectural_patterns_enriched.md` | `e7238bd69f87e7b4112a6c2f2604e63ecb8b7c5c420d9026fa4a37ed701ec5ae` | Useful glossary of FSD, ROP, hexagonal architecture, configuration-driven systems, arc42, ACLs, DDD, C4, ISO 29148, FMEA, and RFC 9457. It mixes frontend-specific and distributed-system patterns into a universal catalogue, uses secondary sources for several claims, and contains overstatements such as “never mix exceptions,” “circuit breakers everywhere,” and rejecting any design without three alternatives. Its RFC 9457 table incorrectly calls `type`, `title`, and `status` required members; RFC 9457 defines defaults/semantics rather than requiring all three to be emitted. | **Research only.** Extract vocabulary and failure-oriented questions; do not treat pattern choice as compliance. Correct factual errors before republishing. |
| `docs/kiro-ears-dag.md` | `7a3b5d526c542a522b62bd09532c59f1bd4af71c7bff047ed99df3e5af50ad04` | Contains useful criticism of EARS/RFC limits, DAG readiness, property/invariant testing, resource bounds, dry-run/shadow side-effect isolation, threat modeling, traceability, and lifecycle. It is a concatenated conversational draft with broken code fences/markup, 3D/GIS-specific examples, duplicated Vale rules, invented universal requirements, and unsafe claims that regex lint plus single-line formatting yields deterministic semantics. “Auto-correct after lint,” frozen-everything lifecycle, forced diagrams, and automatic code generation are not acceptable defaults. | **Extract then archive/supersede.** Keep domain-specific examples only as labeled examples or validator torture fixtures. |
| `docs/newer_things.md` | `67c856adc829154d296c18cd03576ec14f1616581d3da077cc8d69e7d6f09a97` | A concrete 3D/GIS SDD example plus repeated architecture commentary. It usefully surfaces coordinate systems, plugin ABI boundaries, resource exhaustion, ghost states, reconciliation, schemas, and failure tracks. It also asserts that one stack is a “gold standard,” that patterns have “zero overlap,” and that particular technologies/limits apply universally. The Markdown/YAML/code are frequently flattened and non-executable, and the 40% GC claim is unsupported. | **Research/example only.** Extract portable edge cases; never make its technologies or budgets part of the generic CLI contract. |
| `docs/newer_things 2.md` | `1cce0c230db87c9eaa66581c6d0bf8b162849bda633abbc1e235c05d7ef90b62` | A malformed requirements backlog with duplicated fragments, missing headings and requirement numbers, incomplete findings, contradictory negative-language rules, a fixed four-phase waterfall, mandatory branch-per-feature/commit-per-artifact/PR, a message board as artifact authority, fixed agent roles, forced periodic updates, and impossible properties such as requiring at least one audit finding in every category. | **Archive as negative test material.** It should seed parser/validator fixtures for truncation, duplication, contradiction, orphaned acceptance criteria, and untestable requirements—not shape the product architecture. |

### 4.3 Research and working notes

| File | Baseline SHA-256 | Finding | Disposition |
|---|---|---|---|
| `docs/research/okf-storage-and-protocol-specs.md` | `6b15e0b71bc3c2c65e5a56cfe5129e83dff2d01278e49f21ac25e6bbb805acdb` | Current, focused research supporting Markdown/Git authority, SQLite/FTS projection, optional Obsidian authoring, deferred hosted topology, protocol profile, and corrected RFC usage. It explicitly separates evidence from product decision. | **Keep as dated research.** Revalidate dynamic sources before a later architecture decision changes storage or protocol support. |
| `notes/spec_rfc_methodology.md` | `974516a8ab0aa092f78a3f68ebe8fd668330b1cf8694c7de8bbf5420d336c7ac` | Valuable distinction between normative/informative text, stable requirement IDs, traceability, BCP-style keywords, security considerations, and formal grammar when a protocol actually needs it. It over-applies IETF document ceremony to ordinary software work, references obsolete HTTP/1.1 RFC 7230 as a structural model, omits RFC 8174 from its normative table, treats SemVer and YAML fields as universal, and overclaims direct lossless EARS-to-temporal-logic/formal-verification translation. | **Keep as historical research; supersede as standard.** The plan now uses a dedicated protocol profile only for independently evolving interfaces. |
| `notes/autonomous_agent_patterns.md` | `f468f14bd983676895349495bdc63010ae01d8d51fed0c23a9b7ae9d97eedabe` | Good seeds: locked evaluators, decoupled verifiers, stop rules, metadata-only skill discovery, and durable progress. It uses old OKF v0.1 conclusions, makes `.kiro/state.json` authoritative, and repeats destructive `git checkout --` reversion in the primary worktree. | **Supersede.** Preserve as provenance; use event journal + recoverable projections and isolated experiment worktrees. |
| `notes/detailed_research_report.md` | `7fd9e89021b7cfa60df8e743a101eba5a542b21899cbf1c58ea339f0e75273bc` | Useful early framework comparison. It converts rough heuristics into categorical product claims—OpenSpec ≤2h, Kiro >2h—and recommends old `.kiro`, state-file authority, top-level research commands, and automatic revert. | **Supersede.** Its comparative insights are consolidated into the plan’s source table and risk router. |
| `notes/open_source_specs_research.md` | `30cf0bb13b5945f369cfc8065db4b2b691b299f5515a0bffdeefa826eae05c67` | Concise description of OpenSpec, Spec Kit, and Kiro concepts. It is useful provenance but lacks pinned revisions, treats vendor marketing/shape as stable fact, and does not address failure/recovery/authority. | **Retain as dated research, not current reference.** The pinned vendor corpus and plan supersede it. |
| `notes/hidden.md` | `8a3c4a358d58e3f66b108034a14646b43bf5589e571c3863d88aadabc750f046` | Strong low-level prompts: process-group cleanup, atomic sibling-write/rename, async DAG readiness, Git index contention, and cycle-safe graph traversal. It also contains unsafe or unsupported shortcuts: hidden prompt directives as control, an unmeasured 5× JIT claim, shell-based `flock`, simplistic rename guarantees across platforms/filesystems, and Tarjan SCC as a necessary answer to every Markdown link cycle. | **Extract verified mechanics; quarantine hacks.** Prompt tokens never enforce policy; OS behavior needs platform tests. |
| `notes/notes.md` | `629ac5a559102eb736a349495bdc63010ae01d8ebf0e3a68ee398a6427c6509a` | Raw backlog that records the original ambition. It conflates CLI, agent orchestrator, research provider, Git/PR manager, memory system, and framework clone; it is useful for intent provenance but not executable scope. | **Archive/supersede.** Link to the refined plan and close or migrate remaining ideas into explicit decisions. |
| `notes/workspace_inspection.md` | `57af5c59645967427a155cc92e1eb405a5cbe7c72aa0f4a992261f4ccaf12ce5` | Useful historical inventory. It contains machine-specific absolute paths, a now-stale nested repository description, and claims about other local projects that cannot be portable product facts. | **Historical snapshot only.** Never include absolute local paths in product documentation or test fixtures. |

## 5. Contradictions resolved in the refined plan

| Conflict in the corpus | Refined decision |
|---|---|
| `.kiro/`, `openspec/`, and `.specify/` each appear as authority | New work uses one `.sdd/` canonical model; others enter through anti-corruption import adapters. |
| Old docs expose many commands | Users get six primary commands; advanced namespaces exist for maintainers and automation. |
| Fixed Change/Spec loops selected by time/description length | Workflow schemas select progressive rigor from observable risk; users may override with rationale. |
| Plans/specs are frozen before implementation | Accepted capability truth is living; active work may flow back with journaled revisions and downstream staleness; completed history is immutable. |
| Agent roles and message bus are product architecture | The CLI emits role/action packets but does not own an agent runtime or message broker. |
| State lives in `.kiro/state.json` or message-board memory | Authoritative files plus append-only journal; SQLite/state projections are disposable and rebuildable. |
| Tasks execute in bulk-synchronous waves | DAG readiness unlocks each task as its own dependencies complete; “waves” may remain a visualization. |
| CLI automatically commits/reverts/pushes/opens PRs | Git mutation is explicit, path-bounded, inspectable, and opt-in; experiments use isolated writable surfaces. |
| Every requirement must use EARS | EARS is optional for deterministic behavioral clauses; scenarios, schemas, invariants, statistical contracts, examples, and prose are chosen by problem type. |
| Every design is an RFC/arc42 document with diagrams | Protocol profile applies only to durable interfaces between independently evolving participants; ordinary changes pay proportional rigor. |
| Regex/Vale formatting is the semantic contract | Markdown AST + typed frontmatter + schema/graph validation is the contract; prose linters are advisory quality checks. |
| `done` is a checkbox or phase label | Completion is policy plus digest-bound evidence, `done.md`, and a machine receipt; an unverifiable requested check cannot pass. |
| OKF Markdown, SQLite/Turso, and Obsidian compete as one storage decision | Markdown/Git is exchange and authority; SQLite is a local read model; Obsidian is an optional view; hosted libSQL/Turso is a later governed service projection. |
| More architecture is inherently more production-grade | Every dependency, process, interface, config key, and extension point must pay rent; simplifications state ceilings and revisit triggers. |

## 6. Improvements incorporated through revision 6

The audit changed the plan in these material ways:

1. Added a pinned evidence ledger and explicit source-precedence rules.
2. Corrected the baseline corpus count and made the first-party full-read claim auditable.
3. Added the invariant that `not-run`, interrupted, wrong-snapshot, or uninterpretable verification can never be reported as passed.
4. Required active simplifications and their revisit triggers to appear in `status`, `check`, and `finish` rather than disappear into design history.
5. Added durable leaf cursors and partial-fan-out recovery for any future compound workflow.
6. Added optional discovery as intake → evidence → problem → options → decision, explicitly outside the ordinary happy path.
7. Split extension trust into schema validity, curation, integrity, publisher authentication, review, sandbox compatibility, and invocation authorization.
8. Bound credentials to extension identity/origin/invocation and made cross-origin redirects strip authority.
9. Added historical-document tombstones, translation provenance, dynamic-fact expiry, and default exclusion of superseded pages from agent context.
10. Preserved the product boundary: deterministic control plane, not a message bus, model gateway, research-provider bundle, Git autopilot, or universal architecture generator.
11. Selected Go for the production CLI, retained TypeScript only as a temporary parity oracle, and defined the contract-first migration.
12. Defined one deep `Engine.Execute` seam, capability-oriented internal packages, and ports/adapters only at real I/O seams.
13. Distinguished atomic visibility from crash durability and direct-child cancellation from full process-tree supervision.
14. Added Bug Fix and Quick Spec as maintained journey presets without conflating journey shape with the independently risk-derived rigor floor.
15. Defined evidence-backed diagnosis/regression gates, atomic requirements/design/tasks draft generation, no-agent fallback, escalation without lost work, and adversarial fixtures for both guided workflows.

Earlier refinements informed by the same corpus already introduced:

- independent project-root and work-item selection coordinates;
- living accepted specs versus immutable completion history;
- shallow initiatives made of independently finishable slices;
- read-only convergence assessment;
- proof-bearing `DONE` records with planned versus actual order and code/policy references;
- a protocol-specification profile using current HTTP/problem-details guidance only where applicable;
- benchmark controls that can disprove product claims;
- OKF Markdown/Git authority with SQLite/FTS5 as a disposable projection.

## 7. Claims that must not become product contracts without new evidence

The following claims occur in the source corpus but are unverified, too broad, or context-dependent:

- exact percentage improvements in GC pressure, token consumption, or agent quality;
- exact JIT warm-up multipliers and fixed warm-up counts;
- universal duration boundaries such as “brownfield under two hours” versus “greenfield over two hours”;
- the claim that EARS maps losslessly to temporal logic or directly enables formal verification;
- the claim that hidden HTML comments reliably force agent behavior;
- the claim that a particular architecture stack is a universal “gold standard” with zero overlap;
- fixed message delivery intervals, agent counts, phase counts, and mandatory audit finding counts;
- fixed resource, latency, rollout, or geometry thresholds copied from a 3D/GIS example;
- the assumption that `rename`, file locking, signals, symlinks, and process groups behave identically on every supported OS/filesystem;
- any current vendor feature, version, support matrix, pricing, or security posture without a pinned primary source and verification date.

These can become hypotheses, test fixtures, example policy, or measured project-specific requirements. They cannot become silent defaults.

## 8. Documentation cleanup required before an internal pilot

The safest cleanup sequence is:

1. Add frontmatter and visible tombstone banners to every superseded first-party page.
2. Create `docs/tutorials/`, `docs/how-to/`, `docs/reference/`, `docs/explanation/`, and `docs/architecture/` only as content is ratified; do not copy old files into new authoritative locations.
3. Move raw conversational drafts and malformed requirement dumps to `archive/research/` or a parser-fixture directory with `normative: false`.
4. Generate CLI, JSON protocol, schema, diagnostics, event, and integration reference from tested source contracts.
5. Add a documentation manifest with owner, authority, source contract, version, and `verifiedAt`/expiry for dynamic claims.
6. Make context assembly/search exclude `superseded`, `archived`, fixtures, and vendored content unless the action explicitly requests historical/upstream evidence.
7. Add CI checks for broken links, code fences, duplicate stable IDs, stale generated reference, invalid frontmatter, missing tombstones, and expired dynamic claims.
8. Ratify the plan’s ADRs and then stop growing the plan as a second reference manual.

No bulk move or tombstoning was performed during this audit because the user asked to improve the plan, and moving research files would be a separate repository-organization change with link and history consequences.

## 9. Definition of audit completion

This documentation audit is complete when:

- every baseline file appears exactly once in the coverage ledger;
- the baseline count, byte/word/line totals, hashes, and manifest digest are recorded;
- contradictory current-looking pages have an explicit disposition;
- extracted insights are traceable to plan changes;
- unsupported claims are not promoted into product requirements;
- the audit output is clearly non-normative;
- future cleanup work is bounded and ordered.

Those conditions are satisfied for the 20-file `notes/` and `docs/` baseline.

## For the chosen ones

The dangerous failure mode in an agent-facing documentation repository is not merely “stale docs.” It is **authority aliasing**: two files describe different eras or scopes using the same confident grammatical voice, and retrieval erases the provenance that would let the agent distinguish them. A vector search system can make this worse because semantic similarity promotes the obsolete page precisely when it looks most relevant.

The mitigation is not a better embedding model. It is a typed evidence plane: authority, status, source revision, temporal validity, and supersession are filter predicates applied before ranking. Lexical or vector relevance should order only the candidates that remain after policy filtering. This is why Markdown frontmatter, tombstones, provenance, and a rebuildable index are part of the safety architecture rather than documentation housekeeping.
