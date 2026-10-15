# Production plan for the SDD CLI

**Status:** Proposed product and implementation plan  
**Date:** 2026-10-11  
**Revision:** 6 — added orthogonal journey/rigor/profile composition plus evidence-backed Bug Fix and bounded Quick Spec presets; retained the Go production decision and contract-first migration  
**Audience:** maintainers, staff engineers, security reviewers, and internal adopters  
**Working name:** `sdd`  

## 1. Executive decision

Build `sdd` as an **agent-neutral, deterministic workflow control plane**, not as an LLM orchestrator.

The coding agent remains responsible for judgment-heavy work: exploring a codebase, asking useful questions, drafting requirements, designing a change, editing code, and reviewing trade-offs. The CLI owns everything that must be predictable and auditable:

- workspace discovery and migration;
- workflow selection and artifact dependencies;
- the next valid action;
- context assembly and progressive disclosure;
- schema, requirements, traceability, and policy validation;
- artifact freshness and content digests;
- task claims, completion receipts, and verification evidence;
- safe process and Git operations;
- append-only workflow history;
- stable human and machine output contracts.

This is the seam that makes the product portable. Claude Code, Codex, Cursor, Copilot, Kiro, Gemini, or a future agent can all use the same engine by invoking a stable JSON interface. Host-specific skills and slash commands should be thin adapters that translate native agent behavior into CLI calls.

The current Bun/TypeScript monorepo is a useful spike, but it is not the production runtime decision. It already proves delta specs, validation, task DAGs, Git operations, skills discovery, and a basic workflow contract. Preserve it temporarily as an executable oracle while a Go implementation replaces behavior slice by slice behind language-neutral JSON contracts. This is a parity migration, not a blank-sheet rewrite: fixtures and observable outcomes survive; prototype-only architecture does not.

## 2. What the research says

At the audit baseline, the repository contained 788 Markdown files, approximately 4.5 MB. All 20 Markdown files under `notes/` and `docs/` that existed at that baseline were read in full; the generated audit report is an output and is not counted as an input. The vendored OpenSpec, Spec Kit, and Ponytail trees were structurally indexed as a complete corpus; their canonical product, architecture, workflow, extension, integration, skill, and benchmark documents were reviewed in detail. Archived upstream changes are useful as implementation evidence, but they should not be mistaken for current product contracts.

### 2.1 Framework lessons worth keeping

| Source | Keep | Avoid copying |
|---|---|---|
| OpenSpec | Brownfield-first delta specs; living domain truth; artifact DAGs; actions instead of phase locks; archive/sync lifecycle; schema-defined workflows | A wide terminal command surface; relying on agent prompts for guarantees; letting every artifact remain silently mutable without downstream freshness checks |
| GitHub Spec Kit | Constitution/project policy; deliberate clarification; spec-plan-task trace; generated agent integrations; presets and extensions; convergence checks | Greenfield bias; mandatory ceremony for small changes; template layering that becomes a second package manager; agent-specific files as the product interface |
| Kiro | Requirements/design/tasks separation; EARS for deterministic behavior; steering context; explicit task dependencies | Treating every problem as EARS-compatible; hard-wiring the product to `.kiro/`; strict waterfall interpretation |
| Ponytail | Inspect before writing; reuse repository code, standard library, native platform, then installed dependencies; smallest correct diff; never trade away safety | Always-on prompt bulk; assuming fewer tokens or fewer lines is universally cheaper or more correct |
| Open Knowledge Format (OKF v0.2) | Portable Markdown bundles; progressive indexes; provenance, verification, trust, lifecycle, and attested-computation metadata | Confusing the exchange format with a serving database; making embeddings or one vendor runtime authoritative |
| AutoResearch | Locked evaluator, isolated writable surface, objective metrics, keep/reject evidence | Reverting a dirty primary worktree; leaving child processes alive; treating a regex match as a sufficient scientific evaluation |

### 2.2 The important synthesis

The strongest common pattern is not a document template. It is a control loop:

```text
intent
  -> resolve uncertainty
  -> establish an observable contract
  -> choose a design
  -> derive executable work
  -> implement in bounded slices
  -> verify against the contract
  -> reconcile discoveries
  -> promote the accepted behavior into durable truth
```

Each framework makes different parts of that loop explicit. `sdd` should model the loop directly and allow workflow schemas to decide which artifacts are necessary for a particular risk level.

The repository’s older OKF research points at the frozen v0.1 copy in `knowledge-catalog`. The canonical project has moved to `GoogleCloudPlatform/open-knowledge-format`, and the current specification is v0.2. The v0.2 specification still makes a deliberate distinction between a portable Markdown bundle and whatever database, search index, UI, or agent serves it. See the [canonical OKF repository](https://github.com/GoogleCloudPlatform/open-knowledge-format) and [OKF v0.2 specification](https://github.com/GoogleCloudPlatform/open-knowledge-format/blob/main/SPEC.md).

### 2.3 Evidence ledger and source precedence

Research is reproducible only when the evidence set is pinned. The audit baseline is:

| Corpus | Snapshot | Files | Manifest digest | Authority |
|---|---:|---:|---|---|
| first-party `notes/` and `docs/` inputs | 2026-10-11 | 20 | `sha256:170b0d0f5a9e25744d20fe6700c58a417b7092580253252830215a41d75357a5` | hypotheses, proposed contracts, and historical design evidence; authority decided per file |
| vendored OpenSpec | `27b22ab4cbf530fa00e17f0f6b75a44d56777542` | 599 | `sha256:483f856541c570d940c3bff3127c51e59b94499e9318a86b86c009140aac3da1` | pinned upstream evidence, never an `sdd` product contract |
| vendored Spec Kit | `a5560fcf137eb48b4cb0b3a230e202362a011142` | 125 | `sha256:0c1bd22789c6fd3d1b55dfe0b7262c1f378c16f586cefeed41219d43894ad634` | pinned upstream evidence, never an `sdd` product contract |
| vendored Ponytail | `16f29800fd2681bdf24f3eb4ccffe38be3baec6b` | 38 | `sha256:34e866ad9080a184b64363d241f0f9a56738b8e5fc79db0528b28b59ae1b6cb6` | pinned upstream evidence, never an `sdd` product contract |

The digest is over a sorted `sha256sum` manifest, not a claim that every sentence is correct. The detailed per-file findings and dispositions live in `docs/research/markdown-corpus-audit.md`.

When sources disagree, use this order:

1. ratified organization policy and versioned engine/schema contracts;
2. accepted project policy, ADRs, and current generated reference;
3. this plan while its decisions remain proposed;
4. dated first-party research with explicit provenance;
5. pinned current upstream documentation;
6. archived upstream changes, copied notes, and conversational drafts.

Recency alone does not grant authority. A copied draft cannot override a tested contract, and a vendor document describes that vendor unless an `sdd` decision explicitly adopts it. Dynamic claims—current versions, support matrices, benchmark numbers, prices, provider behavior, and security guidance—must record a verification date and source or be labeled unverified.

## 3. Product promise

The user-facing promise should fit in one sentence:

> Tell `sdd` what you want to change; it gives any coding agent the right context and next action, checks every result, and preserves an auditable path from intent to verified code.

### 3.1 The six-command happy path

```bash
sdd init
sdd start "add passkey login with account recovery"
sdd next
sdd status
sdd check
sdd finish
```

The commands mean:

| Command | User meaning | Engine behavior hidden behind it |
|---|---|---|
| `sdd init` | Prepare this repository | Detect repo conventions, create `.sdd/`, capture policy, install selected thin agent adapters, and run health checks |
| `sdd start <intent>` | Begin one unit of work | Create a work ID, assess risk, select a workflow, record the decision, and expose the first action |
| `sdd next [id]` | Tell me what to do now | Resolve the artifact graph, freshness, open questions, claims, and policy gates; return exactly one actionable packet by default |
| `sdd status [id]` | Show where this work stands | Render the materialized workflow state, stale artifacts, blockers, active claims, and evidence |
| `sdd check [id]` | Prove the current state is coherent | Run deterministic validators and configured verification commands; emit diagnostics and remediation actions |
| `sdd finish [id]` | Reconcile and close the work | Verify completion, update durable specs, write the final record, and archive; Git/PR actions remain explicit options |

`sdd doctor` is the one additional troubleshooting command exposed in ordinary help. Everything else belongs under advanced namespaces and should not be necessary for a normal feature.

### 3.2 Example interaction

```text
$ sdd start "add passkey login with account recovery"

Created work auth-passkeys.
Risk: high — authentication, credential recovery, and persistent identity data.
Workflow: security-change.

Next: resolve 3 decisions before requirements can be approved.
Run `sdd next` or ask your coding agent to continue the SDD workflow.
```

The user should never need to choose “OpenSpec loop versus Kiro loop,” understand a DAG, select a validator, name a role, or remember an artifact path. Those are engine decisions with visible reasoning and an override when needed.

### 3.3 Advanced namespaces

Advanced capabilities remain available without polluting the happy path:

```text
sdd agent ...         stable machine protocol for coding agents
sdd workflow ...      inspect, explain, or author workflow schemas
sdd schema ...        inspect and validate artifact schemas
sdd integration ...   install, update, diagnose, and remove agent adapters
sdd config ...        project and user policy
sdd git ...           opt-in safe Git helpers
sdd experiment ...    isolated evaluator-driven experiments
sdd migrate ...       import or upgrade prior workspace formats
sdd debug ...         graph, journal, and context diagnostics
```

Compatibility aliases can preserve the current `new`, `validate`, `instructions`, `archive`, `compile-waves`, `route`, and `profile` commands during the pre-1.0 migration. They should call the same application interface and print a deprecation hint; they must not retain separate implementations.

### 3.4 Three operating modes, one meaning

The same command must preserve its semantics in three environments:

| Mode | Primary caller | Interaction rule |
|---|---|---|
| Interactive | Engineer at a terminal | Concise explanation, optional prompts, color only when supported |
| Agent | Coding agent or tool adapter | Versioned JSON, no prompts, deterministic ordering, bounded context |
| CI | Policy gate or verification job | Non-interactive, reproducible, cache-aware, precise exit code and report artifact |

TTY detection may change presentation, never workflow meaning. Any command that can prompt must accept all answers through flags or structured input and must fail with a usage diagnostic under `--no-input` rather than hanging. `--json` implies `--no-input`, disables ANSI and progress animation, and emits exactly one protocol document to stdout.

The primary users are deliberately ordered:

1. an engineer trying to move one change forward;
2. a coding agent asking for the next bounded action;
3. a reviewer reconstructing intent, evidence, and scope;
4. a platform team defining non-bypassable organizational policy;
5. a workflow author extending the system without forking the engine.

If a design improves workflow-author convenience while making the first two callers learn more concepts, it is pointed in the wrong direction.

## 4. Product principles

### 4.1 Determinism belongs in the engine

An LLM prompt cannot guarantee ordering, safety, traceability, or completion. The CLI must compute these from files, schemas, hashes, policies, and process results. Agent prose may recommend; engine diagnostics decide whether a contract is satisfied.

### 4.2 Files are the collaboration format

Human-readable Markdown and YAML remain the durable interface. The engine builds typed internal models from them and emits JSON for automation. A database may later index the files, but it must never be required to recover the authoritative state.

### 4.3 Dependencies enable work; freshness protects work

The workflow is not a waterfall. Users can revise an earlier artifact at any time. However, changing an upstream artifact makes affected downstream artifacts stale by digest. The engine explains what must be reviewed or regenerated before implementation or finalization continues.

This resolves the false choice between OpenSpec’s fluidity and deterministic enforcement:

```text
editable at any time != valid for every transition
```

### 4.4 Rigor scales with risk

A CSS copy change and an authentication migration must not pay the same planning cost. The router chooses a built-in workflow based on observable risk signals, not a fixed two-hour estimate.

### 4.5 No silent authority expansion

Planning does not authorize commits, pushes, PR creation, package installation, network research, or external issue creation. Such operations require an explicit command or project policy. Read-only discovery may be automatic; material external side effects may not be.

### 4.6 One deep engine, thin adapters

The CLI parser, generated skills, and IDE integrations should contain no workflow logic. They all invoke one application interface. If deleting the engine caused its complexity to reappear in every adapter, the engine is earning its place.

### 4.7 Progressive disclosure everywhere

Humans see the next decision. Agents receive the smallest sufficient context pack. Maintainers can inspect the complete graph and event journal. The same system serves all three without dumping everything into every prompt.

### 4.8 Non-negotiable product invariants

These are release contracts, not aspirations:

1. **Rebuildability:** deleting projections and caches never destroys authoritative work.
2. **No lost update:** a mutation with a stale revision or changed input digest is rejected.
3. **No implicit side effect:** planning and validation never commit, push, install, access the network, or mutate an external system.
4. **No false completion:** a file existing is not proof that its action completed; completion requires a valid receipt against the current inputs.
5. **No silent drift:** changing a normative input makes precisely its affected descendants stale.
6. **No hidden override:** the effective workflow and policy are explainable, including the source of every value.
7. **No adapter semantics:** removing every agent integration leaves the workflow behavior unchanged.
8. **No machine-output ambiguity:** a protocol consumer never needs to parse human prose to decide what happened.
9. **No unsafe recovery guess:** when recovery cannot prove that roll-forward is safe, it stops with an inspectable recovery plan.
10. **No authority laundering:** an agent recommendation cannot elevate itself into organizational approval.
11. **No unverifiable pass:** requested reproduction or verification that was not executed, was interrupted, used the wrong snapshot, or lacks interpretable evidence is `not-run`, `partial`, or `unverifiable`—never `passed`.

Every invariant must have at least one black-box acceptance test. If an implementation choice makes one impossible to test through the engine interface, the seam is wrong.

## 5. The canonical workspace model

Use an agent-neutral `.sdd/` directory for new projects. Treat `.kiro/`, `openspec/`, and `.specify/` as import formats through anti-corruption adapters, not as three simultaneous sources of truth.

```text
.sdd/
├── config.yaml                    # project policy and selected workflow defaults
├── project/
│   ├── product.md                 # users, outcomes, product constraints
│   ├── engineering.md             # runtime, style, testing, deployment constraints
│   ├── structure.md               # module map, ownership, dependency rules
│   └── glossary.md                # stable project vocabulary
├── specs/                         # accepted current behavior by capability
│   └── auth/
│       └── spec.md
├── initiatives/                   # optional shallow DAGs of independently finishable work
│   └── identity-modernization/
│       └── initiative.yaml        # outcome, slice boundaries, dependencies, derived status
├── work/                          # active units of work
│   └── auth-passkeys/
│       ├── work.yaml              # selected workflow, risk decision, artifact manifest
│       ├── proposal.md
│       ├── requirements.md        # optional by workflow
│       ├── specs/                 # delta specs when current behavior changes
│       ├── design.md              # optional by workflow/risk
│       ├── tasks.md
│       ├── review.md              # findings and dispositions
│       ├── done.md                # proof-bearing completion record, created by finish
│       ├── completion.receipt.json # machine-verifiable acceptance envelope
│       ├── evidence.jsonl         # task/test/review receipts
│       └── events.jsonl           # append-only workflow journal
├── archive/                       # closed work, immutable except migration
├── schemas/                       # project-local workflow and artifact schemas
├── integrations/                  # managed adapter manifests, not copied business logic
├── cache/                         # disposable projections and indexes; ignored by Git
└── .runtime/                      # locks and recoverable transactions; ignored by Git
```

### 5.1 Source-of-truth rules

- Markdown/YAML artifacts and `events.jsonl` are authoritative.
- `state.json`, if retained for performance, is a disposable materialized view and must be rebuildable from the journal and filesystem.
- Every completed artifact has a canonical content digest in `work.yaml`.
- Every downstream artifact records the upstream digests it was derived from.
- Generated integration files are tracked in a managed-file manifest with original SHA-256 hashes.
- Cache content is safe to delete.
- Archive is append-only from the user’s perspective; corrections use an erratum or superseding work item.

There are three different kinds of truth and they must not be conflated:

- artifact bytes are authoritative for what a person wrote;
- journal events are authoritative for accepted lifecycle transitions;
- projections are authoritative for nothing and exist only to make reads fast.

An out-of-band edit is therefore neither ignored nor automatically accepted as complete. A read-only command reports the observed digest mismatch. The next mutating command records an external-edit event, invalidates the affected receipt, and requires revalidation.

### 5.2 Safe persistence for one file

All mutable structured files must use the same write protocol:

1. acquire a workspace-scoped lock;
2. write a sibling temporary file;
3. flush the file and, where supported, its directory entry;
4. atomically rename the temporary file over the projection;
5. release the lock.

Events require an event ID and idempotency key. Replaying the same completion receipt must be harmless. Interrupted writes must leave either the old valid state or the new valid state, never truncated JSON.

### 5.3 Recoverable transactions for many files

Atomic rename protects one file; it does not make `finish`, migration, integration installation, or spec promotion atomic. Every multi-file mutation must use a recoverable transaction stored on the same filesystem:

```text
.sdd/.runtime/transactions/<transaction-id>/
├── plan.json          # ordered operations, preconditions, target and staged digests
├── staged/            # complete new bytes for every target
├── PREPARED           # durable point after which recovery rolls forward
└── COMMITTED          # all effects and domain events are durable
```

The transaction protocol is:

1. compute a `MutationPlan` without changing authoritative state;
2. stage complete outputs and flush them;
3. acquire locks in canonical path order and re-check every expected revision and digest;
4. durably mark the transaction `PREPARED`;
5. apply each operation idempotently and append its domain events;
6. write `COMMITTED`, update disposable projections, and release locks;
7. clean committed transaction material lazily.

Before `PREPARED`, an interrupted transaction can be discarded. After `PREPARED`, recovery rolls forward from staged bytes. If external changes violate a precondition during recovery, the transaction is quarantined and the CLI emits an exact manual recovery plan; it never guesses or partially rolls back user content.

Locks carry a nonce, owner process, creation time, and heartbeat where supported. Breaking a stale lock requires proving that its owner is gone or an explicit recovery command. Cross-work locks are acquired in lexical order to prevent deadlock. Workspace-format migration alone takes the global workspace lock.

### 5.4 Identity and path rules

- A work item has an immutable machine ID and a mutable human slug. The slug is never the database key.
- Requirement, scenario, decision, task, finding, action, event, receipt, and transaction IDs occupy separate namespaces.
- IDs are case-stable and compared exactly; paths are checked for case-fold collisions on case-insensitive filesystems.
- Stored paths use normalized repository-relative POSIX form. Native paths exist only at the filesystem adapter.
- Renaming an artifact changes its path, not its identity. The journal records the rename.
- A path is validated once while planning and again immediately before mutation to reduce symlink and time-of-check/time-of-use risk.

### 5.5 Two digests, not one misleading hash

Use separate digests for integrity and meaning:

```text
blobDigest       = SHA-256(exact file bytes)
semanticDigest   = SHA-256(schema ID + canonical parsed representation)
derivationDigest = SHA-256(
  workflow digest + effective policy digest +
  ordered input semantic digests + validator/generator contract versions
)
```

The blob digest detects any edit and preserves audit evidence. The semantic digest prevents a formatting-only change from invalidating an entire graph. Canonicalization is artifact-type-specific, versioned, and lossless with respect to normative meaning; when the parser cannot prove that a change is non-semantic, it must conservatively change the semantic digest.

Receipts record all three relevant values. Validator implementation patch versions do not make all old work stale, but a validator contract-version change does. The exact digest ingredients are visible through `sdd debug digest <artifact>`.

### 5.6 Configuration and policy precedence

Ordinary preferences use explicit precedence:

```text
built-in defaults
  < user preferences
  < repository config
  < workflow defaults
  < work-item override
  < command flags
```

Security and compliance policy do not use ordinary last-writer-wins merging. An optional organization policy bundle establishes a floor and may lock fields. Lower layers may tighten a locked rule but cannot weaken it. Every effective value retains `{value, source, locked, reason}` provenance, exposed by `sdd config explain`.

Environment variables are limited to documented scalar overrides and paths to policy bundles; the CLI does not accept an opaque environment-provided configuration blob. Unknown keys are errors in CI and warnings only during an explicit forward-compatibility mode.

### 5.7 OKF storage: format, authoring view, index, and service are different layers

Do not ask whether OKF should be “Markdown or SQLite.” OKF defines the portable knowledge bundle; SQLite, Obsidian, and Turso solve different consumption and collaboration problems.

```text
                         authoritative, reviewable exchange
                    +----------------------------------------+
                    | OKF v0.2 Markdown bundle in Git/files |
                    +----------------------------------------+
                       |                 |                 |
             author/view adapter   local read model   published export
                       |                 |                 |
                  Obsidian          SQLite + FTS5     hosted knowledge service
                                          ^                 |
                                          |          optionally libSQL/Turso
                                          +------ never dual-writer ------+
```

The recommended 1.0 architecture is:

- **Canonical knowledge:** OKF v0.2 Markdown, frontmatter, standard Markdown links, and Git history.
- **Local retrieval:** a disposable SQLite index under `.sdd/cache/knowledge.sqlite`.
- **Human browsing:** any editor; Obsidian is an optional view over the same files, not a new source of truth.
- **Remote/multi-tenant serving:** deferred to a separately deployed knowledge service. Turso/libSQL is one possible storage adapter for that service, not a dependency of the local CLI.
- **Agent session memory:** a separate subsystem with separate retention, privacy, and user-identity rules. Conversation “golden nuggets” never enter curated organizational knowledge without provenance and an explicit promotion/review action.

This follows the format’s own scope. OKF v0.2 describes a Markdown bundle, recommends Git as a distribution form, and intentionally does not prescribe storage or serving infrastructure. It explicitly allows consumers such as Obsidian, search indexes, static servers, and graph viewers. See the [OKF v0.2 specification](https://github.com/GoogleCloudPlatform/open-knowledge-format/blob/main/SPEC.md) and the [canonical project explanation](https://github.com/GoogleCloudPlatform/open-knowledge-format).

#### SQLite is the default read model, not the source

SQLite is well suited to a local application index: it is embedded, transactional, portable, and supports full-text search through FTS5. SQLite itself distinguishes this local-application role from a centralized enterprise-data server; it serializes writes, while WAL permits concurrent readers and a writer on the same host. WAL is not a network-filesystem coordination mechanism. See SQLite’s [appropriate-use guidance](https://www.sqlite.org/whentouse.html), [isolation model](https://www.sqlite.org/isolation.html), [WAL constraints](https://www.sqlite.org/wal.html), and [FTS5 documentation](https://www.sqlite.org/fts5.html).

The initial index should contain normalized, rebuildable facts:

```text
concepts        id, path, type, title, description, resource,
                blob_digest, semantic_digest, lifecycle, trust_tier
links           source_id, target_ref, resolved_target_id, status
sources         concept_id, source_id, resource, actor, last_modified
tags            concept_id, tag
verification    concept_id, actor, method, verified_at, stale_after
fts_concepts    title, description, body, tags        # FTS5 virtual table
index_manifest  path, blob_digest, parser_version, indexed_at
```

The database MUST remain disposable:

- deleting it and rebuilding produces the same queryable facts for the same bundle snapshot;
- every result returns concept IDs, source paths, digests, and trust/freshness metadata;
- an index schema or parser-version mismatch triggers rebuild, not an in-place best-effort guess;
- the database file, WAL, and shared-memory files are ignored by Git;
- a branch switch or external file edit invalidates affected rows before they can be returned as current;
- a query never treats an embedding or rank as proof of authority;
- write operations always modify Markdown through the artifact transaction path and then update the index.

Vector search is optional recall assistance. It may nominate candidates, but policy, exact IDs, explicit links, source provenance, and deterministic filters decide what enters an agent context pack. Embeddings are derived, model-versioned cache entries and may be dropped without data loss.

#### Obsidian is an optional authoring and graph-view adapter

Obsidian stores notes as Markdown files in a local vault and maintains its own metadata cache; its documentation warns that the cache can become out of sync with the underlying files. Shared vaults do not currently provide fine-grained permissions, and simultaneous editing is synchronized rather than live-collaborative. See [how Obsidian stores data](https://obsidian.md/help/data-storage) and [shared-vault collaboration limitations](https://obsidian.md/help/sync/collaborate).

Therefore an Obsidian adapter should:

- open the OKF bundle directly without converting it;
- prefer standard Markdown links over Obsidian-only wikilinks;
- ignore `.obsidian/workspace*.json` and other per-user UI state;
- require no community plugin for bundle correctness;
- treat Obsidian’s graph and metadata cache as disposable views;
- rely on repository or filesystem access control, not the Obsidian graph;
- avoid nested vaults and dual sync systems that can race Git or another file synchronizer.

Obsidian is excellent for one person or a small group exploring and curating a bundle. It is not the company authorization layer and should not be required by agents.

#### Turso/libSQL is a later service topology, not a bigger cache

Turso embedded replicas can serve reads locally while synchronizing with a remote primary; by default writes go to the remote primary, other replicas observe them after synchronization, and offline writes change the conflict and freshness problem. The platform also supports local encryption keys. See the [Turso embedded-replica behavior](https://docs.turso.tech/features/embedded-replicas/introduction).

Adopt a remote database only when one or more of these needs is observed in real workflows, not merely predicted from document count:

- repeated cross-repository searches cannot be satisfied from checked-out bundles;
- access revocation must take effect centrally without controlling every clone;
- different legal or security principals require tenant-isolated queries;
- organization-wide traceability, policy coverage, or audit analytics must be computed centrally;
- mobile or thin clients cannot clone and index the relevant repositories;
- measured concurrency or corpus size exceeds the local-index service objective;
- centrally retained and access-logged evidence is an explicit compliance control.

Those triggers are necessary but not sufficient. A continuously available service, backup/restore policy, regional and retention policy, schema-migration discipline, credential lifecycle, incident owner, and tested authorization boundary must exist before the hosted mode can become supported.

At that point, choose exactly one authority model:

1. **Repository-authoritative:** the service ingests signed/committed OKF snapshots and is a read-only serving projection.
2. **Service-authoritative:** the service owns mutations and emits deterministic OKF exports for portability and review.

Never permit both Git files and a remote database to accept independent authoritative writes. “Sync later” is not a conflict policy.

The recommended hosted evolution is repository-authoritative and one-way:

```text
accepted repository commit
  -> build projection batch
  -> classify and redact
  -> bind repository, commit, concept, and semantic digests
  -> enqueue idempotent outbox record
  -> remote service verifies schema and commit identity
  -> publish a tenant-scoped index revision
```

The outbox key is stable for the same repository, commit, concept, and schema version, so retry cannot create a second logical fact. The remote row retains tenant/repository identity, commit and source digests, schema/canonicalizer version, sensitivity and retention class, ingestion event ID, and tombstone state. The service remains rebuildable from accepted commits and does not write directly into repository Markdown. A remote edit becomes an ordinary reviewed repository proposal. This keeps the network side effect outside the atomic local `finish` transaction and prevents a successful local completion from depending on remote availability.

A database also does not create access control by itself. Authorization belongs in the service interface and must be tested against tenant, concept, source, and export operations. For the local 1.0 product, sensitive knowledge should be separated into bundles/repositories with existing repository ACLs; frontmatter labels and search filters are not security controls.

#### Selection rule

| Need | Use |
|---|---|
| Portable, diffable, reviewed organizational knowledge | OKF Markdown in Git |
| Fast local filtering, graph traversal, and lexical retrieval | Rebuildable SQLite/FTS5 index |
| Personal or small-team visual authoring | Optional Obsidian vault over the Markdown |
| Session/personal agent memory | Separate memory store with retention and consent policy |
| Multi-tenant remote queries and centralized authorization | Hosted knowledge service; evaluate Turso/libSQL as one adapter |
| Low-latency semantic candidate retrieval | Optional derived vector index, never authority |

The evidence and trade-off analysis behind this decision is preserved in [OKF storage, protocol specifications, and evidence-backed completion](./research/okf-storage-and-protocol-specs.md).

### 5.8 Use one deliberate temporal model

The corpus describes “flow-back,” “flow-forward,” “living spec,” and “spec-as-source” as alternatives. `sdd` should not make every team rediscover the choice or mix them accidentally. Its model is:

- **Accepted capability specs are living current truth.** `.sdd/specs/` describes behavior the product currently promises, not the history of how it arrived there.
- **Active work is allowed to flow back.** An implementation discovery may revise the active requirements, delta, design, or task graph, but the revision is journaled and its downstream evidence becomes stale.
- **Completed work flows forward.** A finished/archive record is immutable evidence about one decision and candidate snapshot. Corrections use an erratum or superseding work item.
- **Plans and task lists are derivations, not co-equal product truth.** Durable rationale is promoted into accepted specs, ADRs, project guidance, or explicit decisions before the active work is archived.
- **Code is not assumed to be generated from specs.** `sdd` enforces alignment and evidence; it does not pretend arbitrary production code can be safely regenerated from prose.

Emergency code-first repair is still possible. The CLI records a drift/hotfix work item from the observed code snapshot, captures the incident constraint, and requires the accepted spec and evidence to converge before the work can finish. It never rewrites an old completion record to make history appear cleaner.

This hybrid gives maintainers one answer to “what is true now?” and a different, immutable answer to “why and how did it change?”

### 5.9 Project selection and work selection are separate coordinates

A monorepo can contain several `.sdd/` projects, while a project can contain several active work items. Root resolution and work selection therefore must not be collapsed into the current directory, Git branch, or one shared `current-feature.json` file.

Project selection rules:

1. an explicit `--root` or structured protocol field wins and must name an existing `.sdd/` project;
2. otherwise, walk upward and select the nearest `.sdd/` root;
3. an explicit but invalid root fails without falling back to another directory;
4. every project has an immutable workspace ID so moving or recloning it does not create a new logical project accidentally.

Work selection rules:

- machine and agent mutations carry an explicit work ID plus expected revision;
- an adapter may retain a session-local binding returned by `sdd start`, but it passes that binding back on every command;
- an interactive command may offer the only active work item as a convenience, but ambiguity produces `SELECTION_REQUIRED` rather than guessing;
- a Git branch name may be displayed as context but never selects a work item;
- local selection state lives under ignored `.sdd/.runtime/`, is scoped by worktree/session, and is never synchronized as project truth.

The effective target is a tuple `{workspaceId, workId, expectedRevision}` and is echoed in every mutation plan and receipt. In a monorepo, project state may be directory-scoped while Git branches remain repository-scoped; `doctor` must explain that shared branch namespace. Cross-project atomic work is outside the local 1.0 contract and must be decomposed into independently finishable work items.

## 6. Workflow model

### 6.1 Schema-defined artifact graphs

Replace hardcoded `change` and `spec` branches with a versioned workflow schema interpreted by the core engine.

```yaml
schemaVersion: 1
name: standard-change
description: Brownfield behavior change with explicit verification

artifacts:
  - id: proposal
    path: proposal.md
    needs: []
    validatesWith: [proposal-structure]

  - id: requirements
    path: requirements.md
    needs: [proposal]
    validatesWith: [requirement-ids, requirement-quality]

  - id: delta-specs
    path: specs/**/*.md
    needs: [proposal, requirements]
    validatesWith: [delta-semantics, scenarios]

  - id: design
    path: design.md
    needs: [proposal, requirements]
    when: risk.designRequired
    validatesWith: [design-decisions, failure-modes]

  - id: tasks
    path: tasks.md
    needs: [requirements, delta-specs, design?]
    validatesWith: [task-ids, traceability, task-graph]

actions:
  implement:
    needs: [tasks]
    policy: implementation-ready
  finish:
    needs: [implement, verify, review]
    policy: completion-ready
```

The schema compiler validates unique IDs, known dependencies, acyclicity, output-path containment, validator availability, and condition expressions before the workflow can be used.

### 6.2 Compose journey, rigor, and artifact profiles

Do not put intent shape and risk level into one overloaded profile enum. Resolve a work item from three orthogonal dimensions:

- **Journey:** the artifact/action topology—general change, bug investigation, or fast specification.
- **Rigor:** the policy floor selected from `quick`, `standard`, or `critical` using observable risk.
- **Artifact profiles:** optional maintained additions such as protocol, migration, threat-model, performance, or regulatory evidence.

Ship named presets so ordinary users never have to compose those dimensions manually:

| Preset | Default composition | Typical work | Default artifacts and gates |
|---|---|---|---|
| `quick-change` | general-change + quick rigor | Copy/UI adjustment, narrow refactor, low-risk local behavior | intent, acceptance/delta note, tasks, focused checks |
| `standard-change` | general-change + standard rigor | Product feature or cross-file behavior change | proposal, requirements, delta specs, conditional design, tasks, verify, review |
| `critical-change` | general-change + critical rigor | Auth, payments, destructive data, migrations, concurrency, public interfaces, compliance | standard artifacts plus applicable threat model, rollback/migration plan, decisions, independent review, explicit human approval |
| `bug-fix` | bug-fix journey + risk-selected rigor | Reproducible or reported defect whose cause is not yet proven | bug report, reproduction evidence, investigation, diagnosis, regression contract, conditional design, tasks, focused verification |
| `quick-spec` | quick-spec journey + quick rigor, with mandatory escalation | Bounded low-risk behavior that needs a compact implementation-ready spec | intent, blocking clarifications, requirements, design, tasks, approval, verification contract |

`sdd start` chooses a preset, resolves its dimensions, and prints the evidence behind the choice. `--workflow <name>` selects a preset explicitly, but it cannot lower organization policy or a risk-derived rigor floor. A critical auth defect may use the bug-fix journey with critical rigor; it must never inherit quick controls merely because it is a bug. If `quick-spec` discovers material security, migration, public-contract, cross-cutting, or unresolved architectural risk, the engine emits `WORKFLOW_ESCALATION_REQUIRED` and proposes the resolved standard/critical composition.

The routing record becomes part of `work.yaml`:

```yaml
workflow:
  preset: bug-fix
  journey: bug-fix
  rigor: standard
  artifactProfiles: [regression]
  version: 1
  reasons:
    - existing behavior contradicts an accepted requirement
    - fix crosses the authentication adapter boundary
```

#### Bug Fix journey

The default action sequence is:

```text
report -> reproduce/observe -> investigate -> diagnose
       -> define regression contract -> plan -> fix -> verify -> finish
```

- The report records expected versus observed behavior, impact, environment, first-known version, and available reproduction material without asserting a cause.
- Reproduction may end as `reproduced`, `intermittent`, `environment-specific`, or `not-reproduced`; the last state is not silently rewritten as success.
- Investigation stores observations, ruled-out hypotheses, probes, and evidence provenance. Temporary instrumentation is isolated and cannot become an unnoticed production change.
- Diagnosis names the failure mechanism, evidence that distinguishes it from competing hypotheses, affected scope, and confidence. “Changing this line makes the test pass” is not by itself root-cause evidence.
- The regression contract links a failing-before/passing-after test or other observable check to the affected requirement. A conditional design artifact is added when the fix crosses a boundary or changes a durable decision.
- Implementation remains blocked until the required diagnosis/regression evidence exists, except under an explicit emergency/hotfix policy. Emergency code-first repair must reconcile diagnosis, accepted specs, and exact-snapshot evidence before `finish`.
- The Kiro adapter may render the bounded problem/diagnosis contract as `bugfix.md`; the canonical engine retains typed artifacts rather than making Kiro's filename authoritative.

#### Quick Spec journey

The default action sequence is:

```text
intent -> blocking clarifications -> generate draft requirements/design/tasks bundle
       -> validate links and derivation -> human/policy approval -> implementation-ready
```

- Clarification asks only questions whose answers materially change requirements, design, task ordering, safety, or verification. Resolvable repository facts are investigated instead of asking the user.
- “Auto-generate” does not put an LLM inside the core. The engine emits one bounded `generate-spec-bundle` action and context pack; the host agent produces `requirements.md`, `design.md`, and `tasks.md`, then submits them as one revision-bound bundle.
- Requirements, design, and tasks remain separately addressable artifacts. The design records the requirements digest, tasks record their design/requirement derivation, and one stale upstream artifact invalidates the appropriate descendants.
- Generated content is a draft, not acceptance. The engine validates IDs, traceability, task ordering, verification coverage, unresolved clarifications, and policy before declaring the work implementation-ready.
- Fast means fewer turns and a bundled authoring action—not weaker verification, hidden assumptions, skipped review, or automatic acceptance.

Both journeys use the existing six-command surface. A user may write:

```text
sdd start "login fails after token refresh" --workflow bug-fix
sdd start "add CSV export to the report" --workflow quick-spec
sdd next
sdd status
sdd check
sdd finish
```

Agent integrations may make the clarify/generate cycle feel automatic, but standalone CLI behavior remains deterministic and useful without a hosted model.

The bounded artifact, diagnostic, escalation, and conformance contract for both presets is maintained in [Guided workflows: Bug Fix and Quick Spec](./architecture/guided-workflows.md).

### 6.3 Risk routing signals

Use evidence, not prompt length or a two-hour threshold:

- security or trust-boundary changes;
- persistent data or irreversible migration;
- public contract, protocol, or compatibility changes;
- concurrency, distributed state, or failure-recovery behavior;
- external services, money, privacy, or regulated data;
- blast radius across modules or repositories;
- novelty relative to repository patterns;
- weak or missing tests around the affected area;
- user-declared criticality or deadline;
- confidence of the repository evidence used for classification.

Time budget may influence task sizing, but never lowers a safety requirement.

### 6.4 Action state

Each artifact or action has one computed state:

```text
blocked -> ready -> claimed -> produced -> valid
                           \-> invalid
valid + changed dependency -> stale
claimed + expired lease -> ready
```

The filesystem alone is insufficient for multi-agent state: a partially written file is not a completed artifact. Completion requires a receipt containing the artifact digest, producer, source digests, validation outcome, and event ID.

### 6.5 Clarification as a convergence algorithm

Do not detect uncertainty only through words such as `TODO`, `somehow`, or `appropriate`. Store clarification items as data:

```yaml
- id: Q-AUTH-003
  question: What is the recovery path when every registered passkey is unavailable?
  affects: [REQ-AUTH-007, threat-model, rollout]
  impact: security
  status: open
  answer: null
  source: user
```

The clarification loop stops when:

- no open question can materially change scope, observable behavior, safety, compatibility, or verification;
- every accepted assumption is explicit and has an owner;
- deferred questions have a reason and do not block the selected workflow policy;
- validators find no unresolved ambiguity markers in normative requirements.

The agent proposes questions; the engine records, deduplicates, and gates them.

### 6.6 Finishing is an optimistic three-way promotion

Two active work items may start from the same accepted spec and propose different deltas. `sdd finish` must never let the later finisher silently erase the earlier one.

For every affected capability, retain:

- the **base** semantic digest observed when work began;
- the **current** accepted spec at finish time;
- the **proposed** result obtained by applying this work item’s delta to its base.

The promotion algorithm is:

1. validate all completion policy, traceability, review, and evidence requirements;
2. lock the affected capability specs in canonical path order;
3. compare their current digests with the recorded bases;
4. if unchanged, stage the proposed specs directly;
5. if changed, compute a structural three-way comparison and report conflicting requirement/scenario IDs;
6. require an explicit reconciled delta when conflicts are semantic;
7. commit updated specs, archive material, receipts, and events through one recoverable transaction.

Textual non-overlap is not enough to prove semantic compatibility. Automatic reconciliation is permitted only for structurally independent IDs with schema-valid results; otherwise the engine blocks and explains the collision. Git commits and pushes remain outside this transaction and require separate authority.

### 6.7 Work lifecycle and terminal states

The lifecycle is small and explicit:

```text
open <-> paused
open -> abandoned
open -> finished -> archived
```

`finished` means the selected completion policy passed and accepted specs were reconciled. `archived` means the immutable historical record was durably moved. `abandoned` preserves rationale and partial evidence; it is not deletion. Reopening finished work creates a superseding work item rather than rewriting history.

### 6.8 A proof-bearing `DONE` artifact

A checked checkbox, a confident agent message, a green test command from an older revision, and a file named `done.md` are all claims—not proof. `sdd finish` must create a completion record only after the engine can bind those claims to the current workspace snapshot.

Keep three layers distinct:

1. the **organizational Definition of Done** is versioned policy that declares which gates and authorities apply to a class of work;
2. `done.md` is the human-reviewable instance record explaining how this exact candidate satisfied that policy;
3. `completion.receipt.json` is the machine contract binding the decision, evidence, and acceptance authority to immutable digests.

Copying the organization checklist into every change is not evidence. Conversely, a machine receipt without a comprehensible account of delivered behavior is not a usable engineering handoff.

Each finished work item contains `done.md` with this stable shape:

```markdown
---
type: Completion Record
work_id: 019f2c5e-87c4-7b31-9d9e-4be7379a4f21
work_revision: 42
status: accepted
result_snapshot: sha256:...
policy_digest: sha256:...
completed_at: 2026-10-11T21:30:00+05:30
completion_receipt: completion.receipt.json
---

# Completion Record: Passkey login and recovery

## Summary
<What behavior now exists, who benefits, and how it was achieved at a high level; no task-list transcription.>

## Delivered Scope
<Planned behavior, delivered behavior, and explicit non-deliveries.>

## Planned and Actual Execution Order
| Planned order | Actual order/group | Task | Dependencies | Implementation summary | Receipt | Result |

## Requirements and Scenarios
| Requirement | Scenario | Implementation | Verification | Status |

## Code and Project-Guideline References
| Task/Decision | Code reference | Guideline reference | Disposition |

## Verification Evidence
| Check | Command/Method | Snapshot | Result | Evidence |

## Deviations, Waivers, Residual Risk, and Follow-ups
<Design changes, authorized exceptions, remaining risk, deferred work IDs, and known limitations.>

## Security, Privacy, Migration, and Operations
<Threat/privacy impact, data lifecycle, migration, rollout, rollback or irreversibility, observability, and support notes.>

## Acceptance and Reproduction
<Acceptance authority, exact subject snapshot, bounded reproduction commands, and immutable receipt references.>

## DONE
| Gate | Status | Evidence |
|---|---|---|
| Required tasks | PASS | ... |
| Requirement coverage | PASS | ... |
| Current verification | PASS | ... |
| Guideline compliance | PASS | ... |
| Review and promotion | PASS | ... |

Accepted by policy `<profile>@<version>` at work revision 42.
Completion digest: `sha256:...`
```

The document has two audiences: the first screen is a concise handoff for a reviewer; the tables below it are a reproducible evidence index. The engine renders the final `## DONE` section from receipts and policy results. An agent may draft the narrative sections, but it cannot write its own acceptance stamp. Editing any accepted completion input or `done.md` afterward invalidates the completion digest and returns the work item to `open` or `stale` according to policy.

The sibling receipt is deliberately not inferred by reparsing prose:

```json
{
  "schemaVersion": "sdd.completion/v1",
  "receiptId": "receipt_019f...",
  "subject": {
    "workId": "019f2c5e-87c4-7b31-9d9e-4be7379a4f21",
    "workRevision": 42,
    "resultSnapshot": { "gitTree": "47e9..." }
  },
  "policy": { "profile": "critical-change", "digest": "sha256:..." },
  "acceptedSpecDigest": "sha256:...",
  "doneDocumentDigest": "sha256:...",
  "gateResults": [],
  "evidenceDigests": [],
  "waivers": [],
  "acceptedBy": { "type": "human", "id": "company-id:1234" },
  "acceptedAt": "2026-10-11T16:00:00Z",
  "eventId": "event_019f..."
}
```

This is an SDD-native completion envelope, not a claim of in-toto conformance. Its structure follows the useful attestation distinction between immutable subject and typed predicate; a future ADR may adopt a registered in-toto predicate and signing envelope if supply-chain consumers need it. `done.md`, its receipt, the spec promotion, the archive move, and the `completion.accepted` journal event are staged in one recoverable local transaction. The receipt cannot authorize itself: the engine obtains acceptance from the authority named by the active policy.

#### Code references that survive review

A code reference records:

- repository-relative path;
- symbol, heading, test name, or schema pointer when available;
- line range as a convenience snapshot, not the identity;
- blob digest or result commit/tree digest;
- owning task, requirement, and receipt IDs.

Project-guideline references use the same discipline: path, section anchor, and semantic digest from `.sdd/project/engineering.md`, `.sdd/project/structure.md`, `AGENTS.md`, or another policy-declared source. The completion record marks each applicable guideline `followed`, `not-applicable`, or `exception-approved`; “reviewed guidelines” without references is not evidence.

#### Planned order versus actual order

`tasks.md` defines the intended dependency order before implementation. The event journal and task receipts define the actual execution order. `done.md` shows both when they differ and explains why. Parallel tasks share an execution group rather than being forced into a fictional total order. A task completed before one of its declared prerequisites is a completion error unless the plan was revised first.

#### Mandatory completion gates

The built-in completion policy checks:

1. every required artifact is valid and fresh;
2. every required task is complete and has a receipt against its current input digest;
3. every requirement and required scenario has implementation and verification evidence, or an approved applicability disposition;
4. every changed path is mapped to a task, generated-output rule, or explicit incidental-change record;
5. verification evidence was produced from the same result snapshot being finished;
6. required test, lint, type, build, migration, performance, and security checks passed according to the resolved rigor and artifact profiles;
7. applicable project guidelines are referenced and deviations carry authorized exceptions;
8. design discoveries were reconciled back into requirements, design, tasks, or accepted specs;
9. migrations include rollout, rollback or irreversibility, compatibility window, and operator evidence when required;
10. required review identities are independent where policy demands it and every blocking finding is disposed;
11. there are no blocking diagnostics, unresolved material clarifications, expired claims, or uncommitted prepared transactions;
12. accepted-spec promotion has no unresolved three-way conflict;
13. newly introduced dependencies, interfaces, configuration, processes, and extension points have traceable necessity decisions, while deliberate simplifications name their ceiling and revisit trigger.
14. every requested verification step has a snapshot-bound result; `not-run`, `partial`, `cancelled`, `timed-out`, and `unverifiable` results block any gate that requires a pass.

There is no `--force-done`. A policy-authorized exception is a separate signed/identified artifact naming the exact diagnostic, scope, owner, expiry, and rationale. An incomplete change can be abandoned and archived as abandoned; it cannot be relabeled done.

#### Completion edge cases

- **Docs-only change:** code references may point to documentation, schemas, or examples; code-test gates become not applicable only when the workflow says so.
- **Research or spike:** finish requires a hypothesis, reproducible method, results, decision, and follow-up—not production-code evidence.
- **Manual verification:** record the actor, environment, steps, observed result, and artifact/screenshots; policy decides whether manual evidence is sufficient.
- **Known failing baseline:** record the pre-existing failure fingerprint and prove the result snapshot adds no new matching failure; do not convert the entire test command to PASS.
- **Flaky test:** retries and pass rate are evidence, not a clean pass. A policy threshold decides whether it blocks.
- **Generated code or lockfile:** map the generator command, source inputs, generated paths, and reproducibility result.
- **No Git repository:** use a deterministic workspace snapshot manifest rather than fabricating a commit reference.
- **Post-verification edit:** even a one-line change stales affected evidence until the digest graph proves it irrelevant.
- **Deferred scope:** create a traceable follow-up work ID and remove or supersede the original required task; do not merely leave an unchecked box under a DONE stamp.
- **Feature flag rollout:** distinguish code complete, deployable, enabled for cohort, and fully rolled out. The workflow chooses which state constitutes done.

### 6.9 Initiatives decompose context; they do not create another specification layer

Large work fails when one agent invocation must retain an epic, every artifact, every task, and the changing codebase at once. Before inventing a larger context window or an autonomous super-loop, decompose the outcome into independently demonstrable work items.

An optional `initiative.yaml` is deliberately shallow:

```yaml
schemaVersion: sdd.initiative/v1
id: init_identity_modernization
outcome: Users can authenticate and recover access without shared secrets.
successMeasures: [SM-IDENTITY-01, SM-IDENTITY-02]
slices:
  - id: SLICE-01
    intent: Introduce passkey registration and authentication.
    boundary: No account recovery or password removal.
    dependsOn: []
    workId: work_passkeys
  - id: SLICE-02
    intent: Add recovery with approved factors.
    boundary: No migration of existing passwords.
    dependsOn: [SLICE-01]
    workId: work_recovery
```

The initiative owns outcome, slice boundaries, stable slice IDs, dependency edges, and cross-slice acceptance. It does not duplicate each child’s requirements, design, or tasks. Slice status is projected from child work receipts rather than maintained as another checkbox. A parent cannot be accepted merely because every child says `done`; cross-slice invariants, integration evidence, and the initiative success measures must also pass.

Use the least expensive context control first: scope one action, then one task wave, then parallel independent work, and only then create an initiative. `sdd start` can return a `decompose` action when estimated context, ownership, risk, or dependency breadth exceeds policy, so users do not need another primary command. Recursive initiatives are allowed only when a slice still exceeds the same measured limit and policy caps depth.

### 6.10 Complexity must pay rent

Correctness review asks whether a solution works. Necessity review asks whether all of that solution needed to exist. They are separate review dimensions and neither can substitute for the other.

Before introducing a new dependency, deployable process, public interface, abstraction, configuration key, extension point, or cluster of generated files, the work records:

- what existing repository facility was inspected;
- whether the language/runtime, platform, database, or an installed dependency already provides the behavior;
- the concrete requirement that the new mechanism satisfies;
- its operational, upgrade, security, and deletion cost;
- the observed threshold that would justify a more complex version later.

This is not another document for ordinary work. The fields appear only in the proposal/design decision that introduces the complexity, and `sdd check` reports an unexplained addition. The engine may detect likely additions from dependency manifests, public exports, configuration schemas, processes, and file topology, but a human or agent supplies the rationale.

Deliberate simplifications are recorded with a ceiling and a revisit trigger:

```yaml
decision: use one workspace-scoped lock
ceiling: independent work items cannot mutate the same workspace concurrently
revisitWhen: measured lock wait exceeds 500 ms p95 during the pilot
owner: architecture-steward
```

They are not buried after approval. `sdd status` shows active simplifications whose scope intersects the selected work, `sdd finish` requires an explicit `retained`, `revised`, `removed`, or `triggered` disposition, and `sdd check` emits a stable diagnostic when available measurements cross a revisit trigger. If a trigger cannot be measured automatically, the decision names the manual review event that can close it.

Never optimize a size metric by deleting trust-boundary validation, data-loss protection, security controls, accessibility, observability required for recovery, or verification. A smaller unsafe implementation loses regardless of line count. Conversely, a familiar abstraction with one implementation and no demonstrated variation does not become justified merely because it looks architecturally conventional.

### 6.11 Convergence assessment is read-only; remediation is an ordinary mutation

Post-implementation comparison is valuable, but a checker must not silently append tasks or edit requirements while it is deciding whether the implementation matches them. `sdd check --implementation` emits stable findings classified as:

- `missing` — required behavior is absent;
- `partial` — behavior exists but does not satisfy the full contract;
- `contradicts` — implementation conflicts with accepted intent or locked policy;
- `unrequested` — material behavior or complexity has no traceable requirement/decision;
- `unverifiable` — the behavior may exist, but current evidence cannot establish it.

Each finding names its source requirement/decision, observed code or runtime evidence, severity, affected snapshot, and proposed remediation. The report changes no authoritative file. An agent may turn findings into a revisioned task/spec patch, but that patch passes through the normal `MutationPlan`, expected-revision, deduplication, and authorization path. Re-running assessment against unchanged inputs returns the same finding IDs instead of growing an append-only pile of duplicate “convergence tasks.”

`sdd finish` reruns or validates the assessment against the exact result snapshot. “The checker generated no new tasks” is not completion evidence; a digest-bound coverage result is.

### 6.12 Compound workflows need durable leaf progress, not a giant retry

The six-command experience may eventually coordinate migration batches, initiative fan-out, or organization-defined compound workflows. The engine still executes one idempotent leaf mutation at a time. A compound run records a durable cursor containing the workflow/version digest, ordered or DAG-resolved leaf IDs, completed receipt IDs, pending leaves, fan-out parent, and expected workspace revisions.

After interruption, replay verifies completed receipts and resumes at the first unresolved leaf. It does not rerun the entire compound operation. Changing the workflow definition invalidates the cursor and requires an inspectable replan. Partial fan-out is a first-class recoverable state: successfully created child work remains identified, missing children can be created idempotently, and rollback never deletes a child that has acquired independent work.

This is a future-proofing contract, not a new primary command. Ordinary users continue to see one current action and one recovery message.

### 6.13 Discovery is an optional funnel, not mandatory ceremony

Some work begins with a solution request before the problem is understood. After the core workflow is stable, an optional discovery profile may model:

```text
intake -> evidence -> problem framing -> options -> decision -> ordinary sdd start
```

Discovery outputs are proposals with provenance, confidence, open questions, and expiry; they are not accepted requirements. The decision records which evidence and option trade-offs justified starting work. A user with a bounded, already-understood change bypasses discovery entirely. No product workflow should require ideation workshops, web research, or a multi-agent panel merely to edit known behavior.

## 7. Requirements and specification model

### 7.1 Use structured Markdown, not line-oriented regex as the contract

Parse Markdown into an AST, preserve source locations, and recognize typed blocks through headings and frontmatter. Regex may assist lexical checks, but must not be the primary parser. Multi-line requirements, lists, tables, code fences, and generated formatting must survive round trips.

Every normative requirement has a stable ID that is never silently renumbered:

```markdown
### REQ-AUTH-007 — Recovery without a registered passkey

When an account has no usable registered passkey, the authentication service
MUST require an approved recovery factor before issuing a new credential.

#### SCN-AUTH-007-A — Recovery factor is valid

- **GIVEN** an account with no usable passkey
- **WHEN** the user presents an approved recovery factor
- **THEN** the service permits registration of a replacement passkey
```

### 7.2 Progressive requirements styles

EARS is one supported authoring profile, not the universal representation.

| Behavior type | Recommended contract |
|---|---|
| Deterministic state/event behavior | EARS or equivalent trigger/state/response structure |
| API/protocol behavior | Request/response schemas, examples, error taxonomy, RFC 2119/8174 terms where interoperability requires them |
| Probabilistic AI behavior | Evaluation dataset, scorer, threshold, confidence interval, sample size, and unacceptable failure cases |
| UX/qualitative behavior | User outcome, interaction examples, accessibility criteria, and human evaluation rubric |
| Performance/reliability | Workload model, percentile/SLO, measurement window, tolerance, and test environment |
| Safety invariant | Explicit invariant, failure response, and property/model-based verification where justified |

This avoids forcing probabilistic systems, concurrent behavior, or subjective interaction quality into fake Boolean certainty.

### 7.3 Requirement quality checks

Validators should check facts that can be checked deterministically:

- stable and unique IDs;
- recognizable actor and observable outcome;
- explicit trigger or applicability when needed;
- no unresolved placeholders or escape clauses in normative text;
- measurable units for claimed thresholds;
- scenarios linked to requirements;
- schemas/examples parse and validate;
- requirement terms exist in the glossary when strict terminology is enabled;
- no impossible or conflicting duplicate identifiers;
- every MUST/SHALL has a planned verification method before implementation is finalized.

Do not enforce weak universal rules such as “SHALL NOT is always invalid,” “every requirement must be one physical line,” or “every event requirement must contain a number.” Those produce false confidence and train users to work around the linter.

### 7.4 Bidirectional traceability

Maintain a graph, not a table assembled at release time:

```text
intent
  -> requirement
  -> scenario
  -> design decision
  -> task
  -> changed path/commit
  -> verification evidence
  -> review finding/disposition
```

Explicit IDs are the primary links. Code annotations are optional adapters, not a universal mandate. A task completion receipt must name its requirement IDs and evidence. `sdd check --traceability` reports orphan requirements, ungrounded tasks, missing verification, and code changes with no declared task.

### 7.5 Protocol specifications are a distinct artifact profile

Do not use “RFC” to mean “long design document.” A protocol specification defines behavior between independently evolving participants. It is required when work creates or materially changes a wire protocol, cross-process command protocol, event contract, public API semantics, durable file format, compatibility contract, or extension registry.

The profile borrows engineering discipline, not IETF pageantry. [RFC 7322](https://www.rfc-editor.org/rfc/rfc7322.html) is useful for precise structure and for separating normative from informative references, while [RFC 2360](https://www.rfc-editor.org/rfc/rfc2360.html) emphasizes implementability, message/state descriptions, errors, and management. Internal specifications do not copy IETF boilerplate, publication categories, or an IANA section unless those mechanisms actually apply.

Use the smallest correct profile:

| Change | Primary artifact |
|---|---|
| Internal implementation choice with no interoperability surface | design document plus ADR |
| HTTP API | OpenAPI/schema artifacts plus HTTP semantic profile and examples |
| Asynchronous events/messages | AsyncAPI or message schemas plus delivery/order semantics |
| CLI/agent machine protocol | protocol RFC plus JSON Schemas and conformance fixtures |
| Durable file/workspace format | format RFC plus grammar/schema, canonicalization, and migration rules |
| New binary/text wire protocol | protocol RFC plus ABNF/binary grammar, state machine, and test vectors |

The Markdown document explains semantics; it does not duplicate an authoritative OpenAPI, AsyncAPI, JSON Schema, Protobuf, or ABNF artifact line by line.

#### Required protocol-spec structure

```text
1.  Status, version, authors, and change relation
2.  Abstract
3.  Scope, goals, non-goals, and deployment assumptions
4.  Terminology, roles, trust boundaries, and BCP 14 convention
5.  Architecture and transport
6.  Syntax, data model, framing, and canonical representation
7.  State machine and lifecycle
8.  Message exchanges and normative participant behavior
9.  Errors, cancellation, retries, idempotency, and backpressure
10. Ordering, concurrency, delivery, timeout, and partial-failure semantics
11. Version negotiation, compatibility, extensibility, and unknown fields
12. Security considerations
13. Privacy and data-lifecycle considerations
14. Operational and observability considerations
15. Registry / identifier-allocation considerations
16. Conformance requirements and machine-readable test vectors
17. Migration, rollout, downgrade, and rollback
18. Alternatives and rejected designs
19. Open issues
20. Normative and informative references
```

Sections may say “not applicable” only with a concrete reason. They are prompts for analysis, not boilerplate targets.

#### Normative language

When the document opts into BCP 14, include the RFC 2119/RFC 8174 convention and treat only uppercase keywords as carrying the defined special meaning. RFC 2119 also warns that these imperatives should be used sparingly for interoperability or harm prevention, not to force one implementation technique. See [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119) and its capitalization clarification in [RFC 8174](https://www.rfc-editor.org/rfc/rfc8174).

Rules for the SDD profile:

- `MUST`/`MUST NOT` statements need a stable requirement ID and a conformance test or justified non-testable analysis.
- `SHOULD`/`SHOULD NOT` statements define the valid reasons and observable consequences of deviation; they are not weak `MUST`s.
- `MAY` features state how implementations that do and do not implement the option interoperate.
- Informative rationale, examples, and implementation notes are marked as such and contain no hidden requirements.
- Normative references are only documents required to implement this specification; everything else is informative.

#### Syntax, state, and failure semantics

- Use ABNF for text syntax when appropriate, following [RFC 5234](https://www.rfc-editor.org/rfc/rfc5234); use the native schema language for JSON, Protobuf, OpenAPI, or AsyncAPI payloads.
- Define every participant role and externally visible state. A state diagram without transition preconditions, emitted outputs, timers, and invalid-event behavior is illustrative, not normative. [RFC 4101](https://www.rfc-editor.org/rfc/rfc4101) is useful guidance for writing protocol models.
- Specify truncation, duplication, reordering, replay, timeout, retry, cancellation, unknown messages, unsupported versions, resource exhaustion, and peer disappearance.
- For HTTP APIs, follow current HTTP semantics and use [RFC 9457 Problem Details](https://www.rfc-editor.org/rfc/rfc9457) only when its problem model fits; it is not a universal error envelope for non-HTTP protocols. [RFC 9205](https://www.rfc-editor.org/rfc/rfc9205) provides guidance for protocols built on HTTP.
- Extension points define collision avoidance, ownership, allocation policy, unknown-value behavior, and exhaustion. Public Internet registries follow [RFC 8126](https://www.rfc-editor.org/rfc/rfc8126); internal protocols use an equivalent repository-owned registry without pretending it is IANA.
- Extension points are exercised continuously with unknown optional values and retired/reserved values. [RFC 6709](https://www.rfc-editor.org/rfc/rfc6709.html) explains why casual extensibility creates incompatible variants, and [RFC 9170](https://www.rfc-editor.org/rfc/rfc9170.html) documents protocol ossification when negotiated paths stop being exercised.
- Parsers do not rely on a vague “accept anything reasonable” rule. Illegal and malformed inputs get bounded, specified handling so different implementations do not normalize the same attack differently; this follows the modern robustness guidance in [RFC 9413](https://www.rfc-editor.org/rfc/rfc9413.html).

#### Security, privacy, and operations

A protocol security section identifies assets, attackers, trust assumptions, authentication, authorization, confidentiality, integrity, replay/downgrade risks, denial of service, resource bounds, and failure of dependencies. “Uses TLS” is not a threat model. Apply [RFC 3552](https://www.rfc-editor.org/rfc/rfc3552) as the analysis guide.

Privacy analysis covers identifiers, correlation, retained data, secondary use, disclosure, user participation, and stored-data compromise where relevant, following [RFC 6973](https://www.rfc-editor.org/rfc/rfc6973). Operational analysis covers configuration, monitoring, counters, logs, fault isolation, safe defaults, upgrades, and rollback; [RFC 5706](https://www.rfc-editor.org/rfc/rfc5706) provides a protocol operations checklist.

#### Conformance and running code

Each protocol profile produces:

```text
protocol.md
schemas/
conformance/
├── valid/
├── invalid/
├── sequences/
└── expected-results.json
```

Examples in the prose are generated from or checked against these fixtures. A stable internal protocol requires one real implementation plus an independently executed conformance harness. A protocol whose purpose is third-party interoperability should demonstrate two independent implementations or an equivalent cross-implementation test before stable status.

Record implementation maturity, coverage, supported draft version, and interoperability evidence while the protocol is evolving. [RFC 7942](https://www.rfc-editor.org/rfc/rfc7942) treats implementation status as time-dependent evidence rather than normative protocol text; in `sdd`, that evidence belongs in work artifacts and `done.md`, not in the accepted protocol spec.

#### Corrections to the existing local RFC research

Treat `notes/spec_rfc_methodology.md` as a research draft until it is revised:

- cite BCP 14 as RFC 2119 plus RFC 8174, not RFC 2119 alone;
- do not use obsolete HTTP/1.1 RFC 7230 as the generic model for current HTTP semantics;
- do not copy IETF publication categories, boilerplate, or an IANA section into every internal design;
- do not force every requirement into EARS when a schema, grammar, state machine, temporal property, or evaluation contract is the correct representation;
- do not claim that EARS text translates directly into formal verification without a separately defined formal semantics;
- require substantive security, privacy, operations, compatibility, and conformance analysis instead of section-presence checks.

## 8. Agent contract

### 8.1 Stable JSON envelope

Every machine command prints one JSON document to stdout, sends logs to stderr, emits no spinner or ANSI codes, and follows the same envelope:

```json
{
  "schemaVersion": "1.0",
  "command": "agent.next",
  "ok": true,
  "workspace": {
    "root": "/repo",
    "formatVersion": 1
  },
  "work": {
    "id": "019f2c5e-87c4-7b31-9d9e-4be7379a4f21",
    "slug": "auth-passkeys",
    "workflow": "critical-change",
    "revision": 12
  },
  "result": {
    "actionId": "ACT-0008",
    "kind": "create-artifact",
    "artifact": "requirements",
    "outputPath": ".sdd/work/auth-passkeys/requirements.md",
    "instruction": "Resolve the accepted clarification record into observable requirements.",
    "context": [
      {
        "path": ".sdd/work/auth-passkeys/proposal.md",
        "blobDigest": "sha256:...",
        "semanticDigest": "sha256:...",
        "trust": "project-authored",
        "required": true
      }
    ],
    "validators": ["requirement-ids", "requirement-quality"]
  },
  "diagnostics": [],
  "nextActions": []
}
```

Required machine commands:

```text
sdd agent next [--work <id>] --json
sdd agent context <action-id> --json
sdd agent claim <action-id> --agent <identity> --json
sdd agent submit <action-id> --receipt <path-or-json> --json
sdd agent fail <action-id> --diagnostic <path-or-json> --json
sdd agent heartbeat <claim-id> --json
```

Claims use expiring leases. Submission is rejected when the claim is stale, the action revision changed, an input digest changed, or the output path escapes the work root.

### 8.2 Exit codes

| Code | Meaning |
|---:|---|
| 0 | command completed and requested condition holds |
| 1 | validation or verification failed |
| 2 | usage error or malformed input |
| 3 | workspace/configuration/schema error |
| 4 | conflict, stale revision, or active claim prevents mutation |
| 5 | required external tool or process failed |
| 6 | policy or permission denied the requested action |
| 130 | interrupted; child process cleanup completed |

Diagnostics use stable uppercase codes, severity, message, source range, related IDs, remediation, and optional documentation URL. Programs match on codes, never English messages.

### 8.3 Context packs

The engine constructs context in layers:

1. action contract and output schema;
2. direct dependency artifacts;
3. relevant project policy and glossary entries;
4. targeted code evidence with path and digest;
5. optional external research references;
6. previous failures or decisions relevant to the action.

The default packet has a configurable byte/token budget. Overflow becomes linked, addressable resources rather than truncated prose. Context selection is recorded so a review can tell what evidence the agent actually received.

### 8.4 Role packets without a built-in agent runtime

Planner, implementer, verifier, and reviewer are useful capability profiles, but the CLI should not own model selection, temperatures, or vendor calls. It emits role packets and permissions; the host agent decides how to execute them.

- **Planner:** read-only repository evidence; may write planning artifacts only.
- **Implementer:** may claim specific tasks and write only within the declared worktree/repository scope.
- **Verifier:** runs allowlisted verification commands; does not edit implementation.
- **Reviewer:** reads artifacts, evidence, and diff; emits findings and dispositions; does not silently fix its own findings.

The same physical agent may perform multiple roles in a small project, but each receipt records the role. Critical workflows may require reviewer identity independence by policy.

### 8.5 Protocol and format versioning

Do not use one version number for unrelated compatibility surfaces:

| Surface | Versioned as | Compatibility rule |
|---|---|---|
| CLI distribution | semantic version | normal release/deprecation policy |
| machine protocol | `major.minor` | same major is additive; consumer declares supported range |
| workspace format | integer | current CLI reads current and previous stable format; writes current only after explicit migration |
| workflow schema | integer plus schema `$id` | compiler rejects unknown required semantics |
| artifact schema | stable schema ID and version | artifact declares its contract; migration is explicit |
| event vocabulary | event type plus payload version | replay never silently ignores an unknown state-changing event |
| agent adapter | adapter and host capability version | adapter may evolve without changing workflow meaning |

Every machine request may declare `acceptProtocol`. The engine either returns a compatible envelope or fails with `PROTOCOL_VERSION_UNSUPPORTED`; it never emits a best-effort hybrid. Additive fields are allowed within a protocol major and consumers must ignore unknown fields, but unknown event types are safe to ignore only when explicitly marked observational.

Downgrades are read-only by default. A newer workspace is never rewritten by an older binary. `sdd migrate` creates a complete operation plan, backup manifest, and compatibility report before touching authoritative files.

### 8.6 Idempotency, revisions, and retries

Every mutating machine command accepts an idempotency key and expected work revision. The key is scoped to `{workspace ID, command kind, caller identity}` and maps to the original outcome digest. Retrying the same request returns the original logical outcome; reusing a key with different input is an error.

Human commands generate keys internally for a single invocation. Automation must supply them. Timeouts are therefore ambiguous only until the caller retries; the caller never needs to guess whether a mutation landed.

## 9. Agent integrations

### 9.1 Canonical behavior and thin adapters

Maintain canonical agent instructions once. Render host-specific files from adapters:

```text
canonical action protocol
  -> Codex skill
  -> Claude skill/command
  -> Cursor rule/command
  -> GitHub Copilot instruction/skill
  -> Gemini command/skill
  -> Kiro steering file
  -> generic AGENTS.md + SKILL.md
```

An adapter may define paths, invocation syntax, supported capabilities, and rendering format. It must not reimplement workflow selection or validation.

### 9.2 Managed-file safety

For every generated file, record:

- adapter and adapter version;
- destination path;
- original SHA-256 digest;
- current observed digest;
- ownership (`managed`, `shared`, or `user`);
- generation source and parameters.

Update or uninstall removes only unmodified managed files. Modified files are preserved and reported unless the user passes an explicit force option. Installation is staged in a temporary directory, validated, then committed atomically; partial failure rolls back.

### 9.3 Capability handshake

`sdd integration doctor --json` reports whether an agent host supports:

- skills;
- slash commands;
- project instructions;
- hooks;
- MCP/tool discovery;
- structured tool output;
- subagents;
- approvals and filesystem restrictions.

Workflow behavior must degrade honestly. If a host cannot enforce permissions, the packet says `advisory` rather than claiming isolation.

### 9.4 Skills and tools

The core may discover skill metadata and recommend capabilities. It must never install or execute a discovered third-party skill implicitly. A skill decision records the evidence, trust source, required permissions, and whether the choice is required or recommended.

Research tools belong to the host or an extension adapter. The core emits a research need such as `library-docs` or `code-call-path`; it should not hardcode Firecrawl, Context7, Graphify, or CodeGraph as permanent domain concepts.

## 10. Core architecture

### 10.1 External interface

Expose one deep application module:

```go
type Engine struct {
	// private implementation dependencies and internal seams
}

func (e *Engine) Execute(ctx context.Context, command Command) Outcome {
	// parse facts, decide, plan effects, and return one stable outcome
}
```

`Command` is a closed tagged union represented by generated Go types and normative JSON Schema. `Outcome` always contains the stable result/diagnostic envelope, including cancellation and unexpected internal failure. The CLI, tests, agent integrations, and any future MCP/daemon adapter all cross this seam. CLI parsing, prompting, rendering, and `os.Exit` remain outside the engine.

This is intentionally one interface method, not one exported method per command and not a forest of public ports. Deleting this module should force workspace selection, policy, validation, mutation planning, journaling, and receipt behavior back into every caller; that is the deletion test showing that the module is deep.

### 10.2 Internal modules

```text
CLI / agent adapters
        |
        v
Command application module
        |
        +-- Workspace module          root, format, migration, locking
        +-- Workflow module           schema compile, graph, action state
        +-- Artifact module           parse, render, digest, freshness
        +-- Clarification module      questions, assumptions, convergence
        +-- Validation module         diagnostics, policies, traceability
        +-- Context module            progressive evidence packs
        +-- Execution module          claims, receipts, process evidence
        +-- Lifecycle module          sync, reconcile, finish, archive
        +-- Integration module        managed agent adapters
        +-- Journal module            events and materialized state
        |
        v
Filesystem / Git / process / clock adapters
```

These are modules with small interfaces, not a collection of one-method classes. Keep internal seams private unless two real adapters or a hard external dependency justify one.

### 10.3 Real seams

Create adapters where variation exists in production or tests:

| Seam | Production adapter | Test/alternate adapter |
|---|---|---|
| Artifact store | local atomic filesystem | in-memory filesystem |
| Git | argv-based Git CLI | in-memory/fake repository |
| Process runner | local process group runner | deterministic fake runner |
| Clock/ID | system clock and UUID | fixed clock and deterministic IDs |
| Output renderer | human terminal | JSON protocol |
| Agent integration | Codex/Claude/etc. renderers | fixture adapter |

Do not create ports for every parser or policy. Pure in-process logic is directly testable.

### 10.4 Why there is no message bus in v1

The proposed in-memory message board duplicates direct in-process coordination while adding delivery semantics, registration, queues, and timing claims that do not improve a local CLI. It is a shallow module with one meaningful caller and one adapter.

Use the append-only event journal as the internal coordination record. If a future daemon or distributed runner creates a second deployment topology, project journal events into a queue adapter then. The domain event vocabulary can remain stable without forcing a broker into v1.

### 10.5 Every mutation is plan, authorize, then apply

The engine should not scatter writes throughout command handlers. A mutating command has one internal execution shape:

```text
parse command
  -> load immutable workspace snapshot
  -> decide and validate
  -> produce MutationPlan + diagnostics
  -> authorize declared effects
  -> apply through transaction module
  -> return receipt + new revision
```

`--dry-run` returns the real `MutationPlan` before the authorize/apply steps; it is not a parallel simulation implementation. The plan lists every file create/update/move/delete, every process invocation, every Git effect, every required permission, and every precondition digest. This makes destructive behavior reviewable and gives tests a stable surface below presentation but above I/O.

Read-only commands execute against one immutable snapshot. They do not repair projections, append “observed” events, refresh integrations, or otherwise mutate as a side effect of inspection.

### 10.6 Package boundaries follow public contracts

Do not turn every internal module in section 10.2 into a public Go package. That would convert private seams into permanent compatibility promises. Use `internal/` aggressively and expose only what distribution or independent versioning requires:

```text
cmd/sdd/                    argument parsing, prompting, rendering, process exit
internal/engine/            deep application module and Execute seam
internal/workspace/         discovery, snapshots, locks, transactions, journal
internal/workflow/          graph compilation, action state, built-in profiles
internal/artifact/          Markdown models, digests, freshness, traceability
internal/validation/        diagnostic and completion policy
internal/execution/         process supervision, claims, receipts
internal/integration/       thin agent renderers and managed-file behavior
internal/testkit/           golden workspaces, fake effects, fault injection
protocol/schema/            normative public JSON Schemas
workflow/builtin/           embedded schemas, templates, validator profiles
```

The supported 1.0 interface is the CLI, workspace format, and machine protocol—not importable Go packages. Promise a public in-process library only after at least two independent consumers prove that seam. Tests inside the repository may use internal seams without exporting them.

The expanded target tree, dependency rules, slice example, and test placement are maintained in `docs/architecture/go-project-structure.md`; that document must remain derived from this plan and its ratified ADRs rather than becoming an independent authority.

Release one signed native executable per supported OS/architecture plus checksums, provenance, and an SBOM. Provide package-manager recipes only as installers for those artifacts; do not make Node, Bun, npm, or a Go toolchain a user prerequisite. The CLI never self-updates in the background.

### 10.7 Language and runtime decision: Go for the product, TypeScript for parity evidence

Use Go for the production engine and CLI. Keep the TypeScript implementation only until the Go engine passes the same black-box fixtures, then remove it from the shipped product.

| Criterion | TypeScript on Node/Bun | Go | Decision impact |
|---|---|---|---|
| Existing investment | roughly 3,000 source lines and useful parser/behavior examples | port required | small enough for controlled parity migration; do not preserve the wrong architecture to save a small prototype |
| User installation | npm/runtime path is easy for JS teams; Bun can also emit standalone cross-target executables | normal output is a native executable | both can ship binaries, so this is not the deciding myth |
| Core workload | excellent authoring ecosystem for Markdown/schema manipulation | strong standard library for files, hashing, embedding, concurrency, and argv-based processes | Go fits the control-plane workload without needing runtime dynamism |
| Recovery and concurrency | possible, but runtime/package/bundler combinations enlarge the support matrix | explicit contexts, goroutines, race detector, native fuzzing, and simple deployment artifact | Go reduces the number of moving parts around the hardest invariants |
| Dynamic extensions | npm modules are convenient but would create in-process authority and compatibility pressure | favors schemas, files, and subprocess adapters | Go reinforces the chosen extension seam instead of inviting arbitrary in-process plugins |
| Agent integrations | agent SDKs often appear first in TypeScript | integrations here render files and speak JSON; the host owns the SDK | no reason to put the deterministic core in the SDK ecosystem |
| SQLite projection | Node has a built-in SQLite module, but its stability and runtime version become part of support | requires a selected driver and an explicit CGO/pure-Go decision | defer the driver decision to the Phase 5 projection spike; SQLite is not core authority |

Verified 2026-10-11: Bun documents standalone cross-compilation, so the decision is not based on a false claim that TypeScript cannot ship a binary. Node’s single-executable facility remains marked active development in its official documentation. Go provides native executables, embedded resources, build metadata, native fuzzing, a race detector, and vulnerability analysis in the official toolchain ecosystem. Relevant primary sources: [Bun standalone executables](https://bun.sh/docs/bundler/executables), [Node single-executable applications](https://nodejs.org/api/single-executable-applications.html), [Go build/install](https://go.dev/doc/tutorial/compile-install), [`embed`](https://pkg.go.dev/embed), [race detector](https://go.dev/doc/articles/race_detector), and [Go vulnerability management](https://go.dev/doc/security/vuln/).

The migration rule is **contract first, leaf by leaf**:

1. stop adding product features to the TypeScript command tree;
2. make the prototype runnable enough to capture golden input/output/exit-code/workspace fixtures;
3. define JSON Schemas and canonical fixture normalization outside either language;
4. create `cmd/sdd` and `internal/engine` with `version`, `doctor`, `status`, and dry-run `start`;
5. run TypeScript and Go against identical read-only fixtures and compare normalized outcomes;
6. port mutation behavior only after plan/receipt schemas exist, using fault injection rather than duplicating unsafe writes;
7. switch each command when the Go result meets its parity and new-invariant tests;
8. delete the TypeScript implementation, Bun build path, and compatibility harness after the six-command golden journey and legacy import fixtures pass in Go.

Do not maintain a permanent Go core plus TypeScript sidecar. That hybrid would double schemas, release tooling, dependency scanning, and incident ownership while adding no product capability.

## 11. Safety and security

### 11.1 Filesystem

- Resolve and canonicalize the workspace root once.
- Reject `..`, symlink escapes, device paths, and outputs outside declared roots.
- Never destructively target a home directory, workspace root, or unresolved glob.
- Distinguish visibility atomicity from crash durability. A structured-file replacement uses a same-directory exclusive temporary file, bounded write, file flush where supported, replace/rename, and parent-directory flush where supported; the platform adapter reports weaker guarantees rather than pretending every filesystem behaves like local POSIX.
- Use bounded workspace locks with owner metadata and recovery rules. Do not promise correctness on network or userspace filesystems until their semantics pass the same fault suite.
- Preserve unknown user files during migrations and integration updates.
- Offer `--dry-run` for every multi-file or destructive operation.

### 11.2 Processes

- Accept executable plus argument array, never a shell string.
- Define timeout, output limit, environment allowlist, and working directory.
- Start an isolated process tree and terminate all descendants on cancel/timeout: Unix process groups and Windows Job Objects are separate adapters. A language runtime’s direct-child `Kill` is not evidence that grandchildren stopped.
- Attempt policy-defined graceful cancellation, wait a bounded interval, then force termination and close inherited pipes. Treat descendant cleanup as failed until ports/locks and process enumeration prove otherwise.
- Record command, exit code, duration, and output digests; redact configured secret patterns.
- Run experiments only in an isolated worktree or disposable directory.

### 11.3 Git

- Git features are opt-in; the core workflow does not auto-commit or auto-push.
- Never stage the whole worktree by default.
- Require explicit paths and reject unrelated staged files.
- Inspect conflicts; never choose ours/theirs automatically.
- PR creation and push require explicit flags and a clean, non-detached branch.
- Multi-agent implementation defaults to separate worktrees with claim ownership.

### 11.4 Extensions

Do not ship a remote extension marketplace in the first stable release. Start with local, version-controlled extensions and built-in adapters.

An extension manifest must declare files, commands, network access, environment variables, external tools, compatibility, integrity digest, hooks, credential classes, and redirect policy. Install plans are inspectable before execution. Unknown or unsigned remote packages remain unsupported until sandboxing and supply-chain verification exist.

Do not collapse extension trust into one `trusted: true` bit. Report these independent states:

- **schema-valid:** the manifest parses and satisfies its declared schema;
- **curated:** it appears in a reviewed catalog at a pinned revision;
- **integrity-pinned:** retrieved bytes match the expected digest;
- **publisher-authenticated:** an accepted identity signed the artifact and the signature verifies;
- **reviewed:** the exact version received code/security review under named policy;
- **sandbox-compatible:** declared capabilities fit an available enforcement boundary;
- **authorized:** this project/user granted the requested capabilities for this invocation.

None implies all the others. Authentication failures, digest mismatches, undeclared capabilities, or unsupported critical permissions fail closed. Credentials are capability-scoped and bound to the extension identity, publisher, allowed origins, and invocation; they are never placed in generated prompts, logs, or general child-process environments. Redirects crossing an allowed origin strip credentials and require a policy decision rather than inheriting authority.

### 11.5 Privacy and telemetry

The CLI works fully offline by default. Any telemetry is opt-in for company deployments and must exclude prompts, paths, artifact content, Git remotes, usernames, and environment values. Local diagnostic bundles require explicit creation and show their manifest before export.

### 11.6 Trust and prompt-injection model

The repository, generated agent output, tool output, and external research are all potentially adversarial data. Context entries carry provenance and one trust label:

| Label | Meaning | May define policy or commands? |
|---|---|---|
| `organization-policy` | authenticated administrative policy bundle | yes, within declared schema |
| `project-policy` | repository-owned SDD configuration and approved instructions | yes, subject to organization floor |
| `project-authored` | reviewed artifacts created by the team | no implicit executable authority |
| `repository-content` | source, tests, docs, comments, fixtures | no |
| `agent-generated` | unaccepted model output | no |
| `external-content` | web pages, issues, copied documents, tool output | no |

The CLI never promotes prose found in a source file, dependency, issue, or research page into an executable command. Process commands come only from built-in workflow definitions or explicitly approved project policy. Agent adapters must preserve provenance in context packets and warn that lower-trust content can contain instructions intended to manipulate the agent.

This does not “solve” prompt injection; it prevents the control plane from turning an injected instruction into higher authority.

### 11.7 Secret and environment handling

- Child processes receive a minimal inherited environment plus an explicit allowlist.
- Secrets are referenced by name, never copied into workspace artifacts or event payloads.
- Captured output passes through configurable redaction before persistence; the unredacted stream is not cached.
- Diagnostic bundles inventory filenames and metadata before export and require explicit confirmation.
- Context packs exclude ignored files, common secret paths, and files exceeding policy limits unless explicitly requested.
- Crash reports contain codes and structural metadata, not artifact bodies, by default.

## 12. Validation architecture

### 12.1 Layers

Run cheap deterministic checks first and expensive checks only when earlier layers pass:

1. file and frontmatter syntax;
2. artifact schema and stable IDs;
3. workflow graph and freshness;
4. cross-artifact traceability;
5. domain-specific requirement and design rules;
6. repository policy and architecture boundaries;
7. configured lint/type/test commands;
8. semantic agent review;
9. human approval where policy requires it.

The CLI must distinguish deterministic validation from agent opinion. A semantic finding cannot masquerade as a parser error.

### 12.2 Severity and policy

Use four severities:

- `ERROR`: artifact or command is invalid; transition blocked.
- `WARNING`: risk is material; transition allowed only if policy permits.
- `INFO`: useful state or observation.
- `HINT`: optional improvement.

Blocking is a policy decision derived from diagnostic code, workflow, risk, and environment. It is not embedded ad hoc in each validator.

### 12.3 Validation profiles

Ship composable profiles rather than one universal strict mode:

```text
base
requirements-ears
api-contract
security-critical
data-migration
distributed-systems
ai-evaluation
accessibility
company-policy
```

The workflow selects profiles. Projects may extend severity mappings without replacing core diagnostic semantics.

## 13. Current code disposition

### 13.1 Keep and harden

| Current area | Direction |
|---|---|
| Delta parsing and archive merge | Preserve behavior; move behind artifact models; add conflict, idempotency, and golden tests |
| Task graph and cycle detection | Preserve the pure graph kernel; validate unknown references instead of silently filtering them; expose scheduler-ready dependencies, not only fixed waves |
| Diagnostic codes and JSON intent | Standardize into the global envelope and source-range model |
| Argv-based Git spawning | Keep; add path containment, cancellation, timeouts, process groups, and full integration tests |
| Skills discovery | Keep as metadata-only discovery; add trust and capability records |
| BDD/scenario validation | Keep as an optional profile linked to requirement IDs |
| TypeScript prototype | Preserve only as a temporary fixture oracle; migrate behavior behind JSON contracts and remove it from the shipped product after Go parity |

### 13.2 Replace or deepen

| Current area | Problem | Replacement |
|---|---|---|
| `.kiro/state.json` | Non-atomic, last-writer-wins, history capped and not replayable | append-only event journal plus atomic materialized projection |
| Hardcoded change/spec branching | Product complexity leaks into every command | schema-compiled artifact graph and data-defined journey/rigor/profile presets |
| Two-hour loop classifier | Time is a weak proxy for risk | evidence-based risk routing with recorded reasons |
| Regex requirements parser | Breaks on formatting variation and cannot model typed requirements | Markdown AST plus stable typed blocks and source locations |
| One-file workflow hash | Requirements, design, policy, and tasks can drift independently | manifest of dependency digests and stale propagation |
| Commander actions printing directly | Core behavior is coupled to terminal I/O | Go `(*Engine).Execute` returning typed outcomes |
| Hardcoded research providers | Vendor names leak into the domain model | capability requests resolved by host/extension adapters |
| Fixed role skill arrays | Skills change by host and task | capability requirements plus explicit host resolution |
| Static templates in TypeScript strings | Hard to version, override, or inspect | versioned embedded workflow resources that remain inspectable/exportable |

### 13.3 Quarantine or defer

- Cognee memory integration: the referenced implementation is currently missing and breaks imports. Remove it from the default build or complete it only as an optional extension after the journal/context model is stable.
- AutoResearch: keep the evaluator concept, but do not market it as a loop until isolated worktrees, process cleanup, experiment manifests, and metric comparison are implemented.
- Remote web research commands: make these agent/extension capabilities, not core commands.
- Distributed message board: defer until a daemon or remote runner proves the need.
- Automatic PR generation, commit-per-artifact, and conflict resolution: opt-in integrations, never default workflow behavior.
- Remote extension marketplace: post-1.0 after permissions, integrity, rollback, and trust UX exist.

### 13.4 Retire from the primary interface

- top-level `route`, `compile-waves`, `instructions`, `diagnostics`, and provider-specific `research` commands;
- the fixed choice between “change” and “spec” in `sdd new`;
- duplicate `validate`/`diagnostics` behavior;
- mandatory user knowledge of planner/implementer/reviewer handoff commands;
- claims that generated workflows are deterministic merely because one file was hashed.

The underlying capabilities remain available through `start`, `next`, `check`, status explanation, or advanced debug commands.

## 14. Implementation roadmap

Every phase ends in a shippable, demonstrable state. Do not start a later phase while the current phase’s exit gate is red.

### Delivery rules

- Establish one complete golden journey early and keep it green while implementations behind the engine seam are replaced.
- No more than two consecutive pull requests may be infrastructure-only; the next PR must improve an observable journey, diagnostic, recovery case, or measured quality property.
- New architecture ships behind the same commands, not beside them as `v2` command trees.
- Compatibility fixtures land before the behavior they protect is migrated.
- Every phase is dogfooded to build at least part of the following phase.
- A red safety or recovery invariant stops feature work; a red cosmetic target does not.

The permanent golden journey is:

```text
clean repository
  -> sdd init
  -> sdd start
  -> sdd next until implementation-ready
  -> submit implementation and verification receipts
  -> sdd check
  -> sdd finish
  -> delete projections
  -> rebuild and obtain the same finished state
```

### Phase 0 — Stabilize the repository

**Goal:** establish a trustworthy baseline before adding architecture.

Work:

1. freeze feature work in the TypeScript prototype and make its dependency installation reproducible from a clean clone;
2. scope its root test command to first-party packages and give vendored upstreams separate opt-in commands;
3. remove the broken Cognee commands/imports rather than recreating an unneeded core memory subsystem;
4. make the prototype’s typecheck, unit tests, integration tests, build, and CLI smoke tests pass only far enough to capture parity evidence;
5. replace absolute external fixture paths with checked-in fixtures;
6. capture current CLI JSON, exit-code, and workspace outputs as language-neutral migration fixtures;
7. initialize the Go module, `cmd/sdd`, embedded build version, and Linux/macOS/Windows build matrix;
8. add a release smoke check that runs the native artifact in an empty directory and invokes `sdd --version` and `sdd --help`;
9. classify every first-party document by authority and add visible/machine-readable tombstones to prototype-era pages before they can enter normal agent context.

Exit gate:

- clean clone -> TypeScript parity-fixture capture is green and the Go skeleton builds/tests without a user runtime;
- no test discovers `vendor/` unless explicitly requested;
- no source import points at a missing module;
- native artifacts run on every claimed platform and expose embedded version/build provenance.

### Phase 1 — Define the contracts and engine seam

**Goal:** freeze the vocabulary and interfaces before moving behavior.

Work:

1. write ADRs for the deterministic-control-plane thesis, `.sdd/` canonical layout, event journal, workflow schemas, and Git authority model;
2. define language-neutral `Command`, `Outcome`, diagnostic, source range, error, and exit-code JSON Schemas, then generate Go protocol types;
3. define `MutationPlan`, receipt, revision, idempotency, workspace format, and migration transaction contracts;
4. implement human and JSON output renderers over the same outcome;
5. expose `(*Engine).Execute` and route `status`, `doctor`, and a dry-run `start` through it end to end;
6. add protocol compatibility fixtures before migrating the remaining commands.

Exit gate:

- JSON Schema fixtures validate every outcome variant;
- stdout/stderr and exit-code behavior are contract-tested;
- the CLI parser contains no business decision for migrated commands;
- `sdd start --dry-run --json` shows a complete, non-mutating operation plan.
- the TypeScript and Go implementations produce equivalent normalized outcomes for every migrated read-only fixture.

### Phase 2 — Build crash-safe state and the walking skeleton

**Goal:** make state recovery and concurrent access trustworthy.

Work:

1. implement canonical root discovery and path containment;
2. implement atomic structured-file writes and workspace locks;
3. implement append-only events with idempotency keys;
4. implement recoverable multi-file transactions and canonical lock ordering;
5. build a materialized state projection and `sdd debug rebuild-state`;
6. add failure injection around every write boundary;
7. run native Go `init`, `start`, `next`, `status`, `check`, and `finish` commands over one built-in standard workflow through the engine;
8. complete one real repository change using this path and preserve it as the golden journey.

Exit gate:

- forced termination at every persistence step cannot corrupt recoverable state;
- replay produces byte-equivalent canonical projections;
- concurrent mutation returns a conflict, never silent lost updates;
- the six-command golden journey completes and survives projection deletion/rebuild;
- every multi-file mutation exposes the same operation plan used to apply it.

### Phase 3 — Replace loops with the artifact graph

**Goal:** make workflow shape data-driven and extensible.

Work:

1. define and compile workflow schema v1;
2. implement artifact/action states, conditional nodes, and freshness propagation;
3. create versioned `quick-change`, `standard-change`, `critical-change`, `bug-fix`, and `quick-spec` preset resources from composable journey, rigor, and artifact-profile data;
4. move templates from TypeScript constants into versioned workflow resources;
5. implement `sdd start`, `sdd next`, and graph-aware `sdd status` on the engine;
6. replace the walking skeleton’s static graph adapter without changing its command fixtures;
7. map legacy change/spec workspaces to equivalent graph state.

Exit gate:

- the same engine executes all five built-in presets without command-specific workflow branching;
- journey, rigor, and artifact-profile composition is inspectable; a bug-fix may receive critical rigor and quick-spec escalation preserves the current work revision and accepted answers;
- editing an upstream artifact deterministically marks the correct descendants stale;
- cycles, unknown validators, and path escapes fail schema compilation;
- no command branches on `loop === change/spec` outside migration code.

### Phase 4 — Make specs executable contracts

**Goal:** replace brittle document checks with typed artifact validation.

Work:

1. introduce Markdown AST parsing and typed frontmatter;
2. implement stable requirements, scenarios, decisions, tasks, findings, and evidence IDs;
3. port delta semantics, EARS, BDD, design, schema-block, and task graph validators;
4. implement clarification records and convergence policy;
5. implement cross-artifact traceability and stale evidence detection;
6. implement layered `sdd check` with stable diagnostics and remediation;
7. add probabilistic/AI evaluation contract support without pretending it is EARS;
8. add the protocol-specification profile, schema sidecars, and conformance fixtures;
9. implement proof-bearing `done.md`, the machine completion receipt, and completion-gate validation;
10. complete the Bug Fix evidence/diagnosis/regression gates and the Quick Spec revision-bound requirements/design/tasks bundle, including partial-generation, staleness, no-agent, and risk-escalation behavior.

Exit gate:

- multiline and reformatted golden documents parse identically;
- every diagnostic has a stable code and accurate source range;
- a requirement-to-evidence report is generated for golden projects;
- validator false-positive and false-negative fixtures cover known research counterexamples;
- changing code after verification makes the completion candidate stale;
- protocol examples and schemas pass their conformance fixtures.
- Bug Fix cannot confuse non-reproduction or symptom removal with a proven diagnosis, and Quick Spec cannot confuse generated drafts with accepted implementation-ready contracts.

### Phase 5 — Productize the minimal UX and migration path

**Goal:** let internal developers use the product without learning the engine.

Work:

1. polish `init`, `start`, `next`, `status`, `check`, `finish`, and `doctor` without widening the primary command surface;
2. add interactive defaults with non-interactive equivalents;
3. implement legacy aliases and warning policy;
4. implement `sdd migrate --from kiro|openspec|spec-kit --dry-run`;
5. write the first-run tutorial and common how-to guides;
6. implement a rebuildable SQLite/FTS5 OKF index behind the context module;
7. add optional Obsidian-safe ignore/link guidance without making Obsidian a dependency;
8. dogfood the workflow on the remaining CLI work.

Exit gate:

- a new user can initialize and start work in under two minutes;
- all ordinary flows use six commands or fewer;
- no migration overwrites an unknown or modified user file;
- at least three real brownfield changes complete through `finish`;
- deleting the knowledge index and rebuilding it preserves query results and provenance;
- each finished journey produces a current, policy-valid `done.md` and digest-bound completion receipt.

### Phase 6 — Ship agent portability

**Goal:** prove that the control plane works with different agents.

Work:

1. finalize the agent JSON protocol and claim/receipt lifecycle;
2. build the generic `AGENTS.md`/`SKILL.md` adapter first;
3. add Codex, Claude Code, Cursor, and GitHub Copilot adapters;
4. implement managed-file manifests, atomic install, safe update, and safe uninstall;
5. implement integration capability detection and honest degraded-mode reporting;
6. add adapter fixture tests and one end-to-end golden task per supported host where automation is available.

Exit gate:

- each supported agent consumes the same action fixture and produces a valid receipt;
- adapter upgrades preserve modified files by default;
- removing one integration cannot delete another integration’s files;
- unsupported enforcement is labeled advisory.

### Phase 7 — Add bounded execution and multi-agent safety

**Goal:** support parallel work without turning the CLI into a model gateway.

Work:

1. add task/action claims with leases, heartbeats, revisions, and idempotent submission;
2. add worktree creation and cleanup plans with explicit user authority;
3. add safe process runner, cancellation, process-tree cleanup, and output caps;
4. record implementation and verification receipts;
5. schedule ready DAG nodes asynchronously rather than forcing bulk-synchronous wave barriers;
6. enforce reviewer independence only when selected policy requires it.

Exit gate:

- two agents cannot claim the same exclusive action;
- expired claims recover automatically without accepting stale output;
- cancellation leaves no child process or Git index lock behind;
- parallel golden tasks merge only after declared checks pass.

### Phase 8 — Production hardening and 1.0

**Goal:** make the CLI suitable for company-wide adoption.

Work:

1. publish threat model, support policy, versioning policy, and migration guarantees;
2. add performance budgets for startup, status, graph resolution, and validation;
3. create signed/reproducible release artifacts and SBOM;
4. add shell completions and accessible terminal output;
5. complete Diataxis documentation;
6. run a security review and destructive-operation audit;
7. run an internal pilot across different repositories, languages, and agent hosts;
8. freeze machine protocol v1 and workspace format v1.

Exit gate:

- all 1.0 acceptance criteria in section 18 pass;
- upgrade from the last pre-1.0 format is tested;
- no default command requires network access;
- the incident/recovery runbook has been exercised.

### Maturity gates

Version numbers alone do not communicate operational readiness:

| Gate | Audience | Required evidence |
|---|---|---|
| `R0 — engineering preview` | CLI maintainers | Phase 2 exit gate; golden journey; no support promise |
| `R1 — dogfood` | teams building `sdd` | Phases 3–4; weekly issue triage; migrations may still break with tooling |
| `R2 — design-partner pilot` | 2–4 volunteer company teams | Phase 5; rollback tested; named support owner; adoption metrics captured |
| `R3 — company supported` | approved internal repositories | Phase 7; security review; support matrix; incident and recovery runbooks |
| `R4 — stable 1.0` | default internal adoption and external users if desired | Phase 8; compatibility guarantees frozen; signed release artifacts |

Promotion requires evidence; deadlines do not waive a gate. A team may stay at a lower gate while gathering evidence rather than relabeling preview software as production-ready.

## 15. Suggested pull-request sequence

Keep each change reviewable and independently green:

1. freeze the TypeScript prototype, remove broken memory/provider paths, check in hermetic fixtures, and capture its usable behavior;
2. initialize the Go module, native build matrix, embedded version, and empty-directory smoke test;
3. add normative JSON Schemas for command/outcome, diagnostics, exit codes, idempotency, and `MutationPlan`, plus generated Go types;
4. introduce `(*Engine).Execute` with `status`, `doctor`, and dry-run `start`, compared against normalized prototype fixtures;
5. add atomic filesystem adapter, recoverable transactions, and fault-injection tests;
6. add event journal and rebuildable state projection;
7. land the native six-command standard-workflow walking skeleton and its golden journey, then remove the shipped TypeScript CLI path;
8. add workspace v1 plus `.kiro` import dry-run;
9. add workflow schema compiler and graph kernel behind existing journey fixtures;
10. move templates into the first built-in workflow package;
11. add blob/semantic/derivation digests and stale propagation;
12. add Markdown AST artifact models;
13. port delta, task, BDD, and EARS validators;
14. add clarification records, traceability graph, and three-way spec promotion;
15. add protocol-specification profile and machine-readable conformance fixtures;
16. add proof-bearing completion records, machine receipts, and the `DONE` gate;
17. add orthogonal journey/rigor/profile composition plus Bug Fix and Quick Spec presets with escalation, diagnosis, and atomic-bundle fixtures;
18. add rebuildable OKF v0.2 SQLite/FTS5 index and provenance-aware context queries;
19. add legacy aliases and OpenSpec/Spec Kit importers;
20. add generic agent adapter and managed-file manifest;
21. add Codex and Claude adapters;
22. add claims, receipts, and isolated process runner;
23. add organization policy floor, decomposed extension trust/credential controls, documentation-authority enforcement, CI mode, pilot fixtures, signed packaging, and 1.0 release gates.

Do not combine a storage migration, interface redesign, and validator rewrite in one PR. Contract fixtures should land before implementation changes that rely on them.

## 16. Test strategy

### 16.1 Test pyramid by interface

- **Pure property tests:** graph acyclicity, topological readiness, digest propagation, delta merge laws, path containment, idempotent replay.
- **Golden parser tests:** real Markdown variations, source ranges, round-trip stability, malformed documents, Unicode, CRLF, large files.
- **Module interface tests:** run the engine against in-memory filesystem/Git/process adapters and assert outcomes, not internal calls.
- **CLI contract tests:** stdout is valid JSON in machine mode, stderr contains logs, exit codes are stable, no ANSI/spinner pollution.
- **Filesystem fault tests:** crash before/after temp write, flush, rename, journal append, and projection update.
- **Concurrency tests:** competing claims, lock timeout, stale revision, duplicate receipt, interrupted process.
- **Migration tests:** fixtures from current `.kiro`, OpenSpec, and Spec Kit layouts; dry-run and rollback.
- **Adapter tests:** generated paths, invocation syntax, file hashes, update/uninstall preservation, multi-install collisions.
- **End-to-end journeys:** evidence-backed Bug Fix, zero-question and escalated Quick Spec, standard feature, critical auth migration, resumed crash, parallel tasks, archive conflict.
- **Artifact tests:** signed native release executed in an empty environment on every supported OS/architecture.

### 16.2 Golden repository matrix

Maintain small checked-in repositories for:

- TypeScript monorepo;
- Python service;
- Rust CLI;
- frontend-only application;
- Git repository with dirty worktree;
- non-Git directory;
- monorepo with nested project roots;
- symlink/path traversal cases;
- conflicting concurrent changes to the same capability;
- AI feature with probabilistic evaluation criteria.

Tests must never depend on absolute paths to unrelated projects on a maintainer’s machine.

### 16.3 Quality budgets

Initial targets to validate during the pilot:

- `sdd status` warm p95 under 150 ms for 1,000 artifacts;
- `sdd next` warm p95 under 250 ms without external tools;
- no full-repository rescan when tracked inputs are unchanged;
- deterministic output ordering on every platform;
- bounded diagnostic output with a machine-accessible full report;
- zero data loss in fault-injection persistence tests;
- zero unapproved network, Git push, package install, or remote mutation in default flows.

### 16.4 State-machine and conformance testing

Example tests are insufficient for the control-plane kernel. Build a small executable reference model of work/action state and generate command sequences against both the model and real engine:

```text
start -> edit -> submit -> invalidate -> retry -> claim -> expire
      -> crash -> recover -> duplicate receipt -> finish conflict
```

After every step, compare visible actions, revision, diagnostics, events, and freshness. Inject termination at every transaction boundary and then replay. Useful metamorphic properties include:

- formatting-only edits preserve semantic digests but change blob digests;
- changing a normative requirement invalidates every and only its derived descendants;
- event replay in different filesystem enumeration orders produces identical projections;
- retrying an idempotent command never adds a second logical event;
- plan followed by apply has the same declared effects unless a precondition changes;
- any unsupported protocol or event version fails closed rather than partially loading;
- finishing two non-overlapping deltas commutes, while overlapping semantic edits conflict.

Publish a protocol conformance kit containing JSON fixtures, a fake workspace, and expected outcomes. Agent integrations should be tested against that kit rather than against private engine internals.

### 16.5 Agent-behavior evaluation without benchmark theater

Engine tests prove deterministic mechanics; they do not prove that an agent using `sdd` produces more complete, safer, or easier-to-review work. Maintain a separate evaluation harness with these controls:

- compare the same agent/model with and without `sdd`, plus a short-prompt ablation where useful;
- use pinned real repositories and seeded, independently isolated worktrees;
- start every cell with a fresh agent context and an explicit host configuration;
- disable global plugins, user instructions, credentials, and hooks that could contaminate the baseline;
- run deterministic scorer self-tests against known-good and known-bad implementations before spending on agent runs;
- measure completeness, requirement coverage, safety/adversarial behavior, review comprehension, and recovery before secondary metrics such as lines, tokens, cost, turns, or latency;
- exclude tests and required safety controls from “bloat” metrics;
- preserve workspaces, prompts, tool transcripts allowed by privacy policy, scorer versions, and raw receipts so metrics can be recomputed offline;
- report distributions and per-task results with model/host versions, not one universal percentage;
- disclose failed/time-out cells, evaluator limitations, shared-library ancestry, and any manual exclusions.

Add contamination probes: a sentinel instruction and hook visible only to one arm, a deliberately unsafe small implementation that the safety scorer must reject, and a complete-but-larger implementation that a completeness scorer must rank above a stub. If the harness cannot distinguish those controls, its product claims are blocked.

The benchmark must be capable of disproving the product thesis. Lower ceremony or fewer lines is a loss when requirement coverage, safety, or reviewer understanding falls. Cost claims from single-shot chat completions must never be generalized to multi-turn coding-agent sessions.

## 17. Documentation and adoption

Keep documentation types separate.

### Tutorials

- Build and finish your first change.
- Use `sdd` with your coding agent.
- Migrate an existing OpenSpec/Kiro/Spec Kit project.

### How-to guides

- Add company engineering policy.
- Create a custom workflow.
- Resolve stale artifacts after a requirements edit.
- Resume interrupted work.
- Run parallel agents in worktrees.
- Add a new agent adapter.
- Diagnose a failed validator or process.

### Reference

- command and option reference;
- JSON protocol and schemas;
- workspace format;
- workflow schema;
- diagnostic code catalog;
- event and receipt catalog;
- exit codes and environment variables;
- integration capability matrix.

### Explanations

- why the CLI is a control plane, not an agent runtime;
- actions, dependencies, freshness, and policy gates;
- living specs versus historical work records;
- progressive rigor and risk routing;
- agent trust, permissions, and side effects;
- how traceability and evidence prevent specification drift.

The current research repository contains overlapping notes, proposed requirements, enriched pattern catalogs, copied framework material, and product-facing guides at the same apparent authority level. Before company adoption, publish an explicit documentation source map:

```text
docs/tutorials/       learning journeys tested from a clean project
docs/how-to/          one operational task per page
docs/reference/       generated or contract-tested facts
docs/explanation/     rationale and trade-offs
docs/architecture/    ratified ADRs and stable system explanations
docs/research/        dated evidence and hypotheses; never product contract
archive/research/     superseded first-party drafts retained for provenance
vendor/upstreams/     pinned external evidence; never normative for sdd
```

`docs/cli-product-plan.md` is a decision/RFC artifact. Once its ADRs are ratified, split stable decisions into the architecture pages above and stop extending the plan as a second manual reference. Generate command, schema, diagnostic, event, and integration-capability reference pages from the same contracts used by tests. A small documentation manifest records each page’s type, owner, normative status, source contract, and last verified version. CI checks links, examples, code fences, generated-reference drift, duplicate stable IDs, and pages that claim current behavior from a superseded research source.

Kiro specification drafting uses a bounded authoring pack rather than loading this entire research corpus. The always-on steering files define product, runtime, and package direction; focused steering defines CLI UX and Go engineering policy; manual spec-authoring context loads the short product brief and target project structure. The pack lives at `docs/kiro-input/` and `.kiro/steering/`. Generated `requirements.md`, `design.md`, and `tasks.md` are proposals until reviewed and accepted; Kiro output never outranks product ADRs merely because it is newer or more detailed.

Historical first-party pages stay readable for provenance but begin with machine-readable frontmatter such as:

```yaml
status: superseded
normative: false
supersededBy: docs/architecture/workflow-engine.md
verifiedThrough: 2026-10-11
```

The rendered page displays a tombstone banner before any old command or path. Search and context assembly exclude superseded pages by default and include them only as explicitly labeled historical evidence. Translation is a separate, non-normative derivative with `translationOf`, source revision, locale, translator, and verification status; the canonical contract remains unambiguous. Time-sensitive facts carry `verifiedAt` and either an expiry/recheck policy or a build-generated source. Documentation CI downgrades an expired dynamic claim to `stale`; it does not silently keep presenting the claim as current.

Adoption should begin with a small internal pilot. Capture where users abandon the workflow, override routing, ignore diagnostics, or edit generated integration files. Those are product signals, not training failures.

### Pilot protocol

Use evidence-bearing cohorts rather than a broad announcement:

1. record a lightweight baseline from recent changes in each participating repository;
2. choose 2–4 volunteer teams across different languages and repository shapes;
3. have each team complete at least one quick and two standard brownfield changes;
4. pair the first change, observe the second, and leave the third unassisted;
5. review every workflow override, abandoned work item, blocking false positive, manual recovery, and out-of-band edit;
6. compare review comprehension, rework, escaped requirement gaps, and lead time with the baseline;
7. promote maturity only after the unassisted journeys meet their gate.

Initial pilot hypotheses—not permanent targets—are:

- at least 90% of work items recover from interruption without maintainer intervention;
- fewer than 5% of blocking diagnostics are overridden as false positives;
- at least 80% of standard changes finish without users invoking an advanced namespace;
- quick changes do not regress median lead time by more than 20%;
- reviewers can identify intent, changed requirements, and evidence without reconstructing them from chat history;
- zero data-loss, authority-escalation, or unapproved external-side-effect incidents.

If teams repeatedly bypass the same gate, investigate the product model before adding training. If experts need `sdd debug` during ordinary work, the primary interface is leaking.

## 18. Release acceptance criteria

Version 1.0 is ready only when all of the following are true.

### User experience

- A fresh project can initialize, start, validate, and finish work through the six-command path.
- The CLI chooses a sensible workflow for the golden quick, standard, and critical cases and explains why.
- A user can revise any artifact and receive precise stale-descendant guidance.
- Help output keeps advanced machinery out of the primary path.
- Ambiguous project/work selection produces one actionable choice response and never mutates the wrong root or work item.
- Oversized work can be decomposed into initiative slices without introducing another ordinary command or duplicating child specifications.

### Agent portability

- At least four materially different agent hosts pass the same end-to-end protocol fixtures.
- Machine output is versioned, schema-valid, deterministic, and free of presentation noise.
- Agent adapters contain no workflow business logic.
- Capability limitations are reported honestly.

### Correctness and recovery

- Workflow state is recoverable from authoritative files and journal after forced termination.
- Duplicate commands and receipts are idempotent.
- Concurrent writers cannot silently overwrite each other.
- Artifact freshness and traceability reports are correct for all golden journeys.
- A requested verification that is not run, interrupted, timed out, bound to another snapshot, or uninterpretable can never satisfy a passing gate.
- A work item cannot reach `finished` without a current completion record whose evidence matches the result snapshot.
- Deleting and rebuilding the SQLite knowledge index preserves deterministic filters, graph links, provenance, and lexical query fixtures.
- Accepted specs remain living current truth while archived work remains immutable history; flow-back revisions stale exactly the affected descendants.
- Re-running implementation convergence against unchanged inputs returns the same findings and never appends duplicate tasks as a read side effect.

### Safety

- No default flow pushes, opens a PR, installs a package, executes a discovered skill, or accesses the network.
- Destructive operations support dry-run, identify exact targets, and require explicit confirmation.
- Process cancellation terminates descendants and releases locks.
- Path traversal and symlink escape tests pass on every supported platform.

### Engineering quality

- Go format/vet/test/race/fuzz smoke, native builds, empty-directory artifact smoke, and vulnerability checks pass in CI; TypeScript checks remain only while the migration oracle exists.
- First-party tests are hermetic and do not execute vendored upstream test suites.
- Public JSON/workspace contracts are documented and versioned; generated Go types are checked against their schemas but are not an additional public contract.
- Migration and rollback from the current prototype are tested.
- Protocol profiles have schema, sequence, version-skew, malformed-input, and interoperability conformance fixtures.
- The agent-behavior harness detects its contaminated-baseline, unsafe-minimal, and incomplete-stub controls before any product claim is accepted.

### Operations and governance

- Every supported release has an owner, support window, compatibility matrix, and security contact.
- Runbooks cover stale locks, interrupted prepared transactions, corrupt projections, failed migration, adapter drift, runaway child processes, and incompatible workspace versions.
- A support bundle can be generated without artifact bodies and lists exactly what it includes.
- Security-sensitive changes require review from the declared code owner; protocol and workspace changes require an ADR and migration analysis.
- Deprecations emit a machine-readable diagnostic for at least one supported release before removal.
- The project can revoke a compromised adapter or workflow package without requiring a workspace rewrite.
- Extension status never conflates schema validity, curation, integrity, publisher authentication, review, sandbox compatibility, and invocation authorization; credential and redirect tests fail closed.
- The documentation source map distinguishes normative contracts, generated reference, current explanations, research, archives, and vendored evidence; CI detects drift between contract-generated references and implementation fixtures.

## 19. Success metrics

Measure whether the product improves engineering outcomes, not how many Markdown files it creates.

### Leading indicators

- median time from intent to approved implementation-ready plan;
- number of clarification rounds and percentage that materially changed scope;
- percentage of work using the recommended workflow without override;
- stale artifact detection before implementation resumes;
- requirements with linked verification evidence;
- agent context bytes per action and cache hit rate;
- validation false-positive override rate;
- work resumed successfully after interruption;
- percentage of completion attempts blocked by stale evidence, unmapped changed paths, or unresolved deviations;
- OKF query results carrying current provenance/trust/freshness metadata;
- percentage of introduced dependencies, public interfaces, processes, and configuration keys with an explicit necessity decision.

### Outcome indicators

- rework caused by misunderstood requirements;
- escaped defects tied to missing scenarios or failure modes;
- review time to understand intent and verify scope;
- implementation drift found before merge;
- median diff size relative to comparable work before adoption;
- incidents involving unsafe Git/process behavior;
- percentage of archived specs that still match observed product behavior after 30/90 days.

### Guardrail indicators

- quick-change lead-time regression;
- percentage of commands requiring advanced namespaces;
- blocking-diagnostic false-positive rate;
- stale-lock and manual-recovery frequency;
- average context pack size and excluded-secret events;
- migration rollback frequency;
- adapter-generated file conflict rate;
- number of policy exceptions and how long they remain open.

Do not optimize for artifact count, agent token count in isolation, or strict workflow compliance. Those can improve while the product becomes slower and less trustworthy.

## 20. Risk register

| Risk | Failure mode | Mitigation |
|---|---|---|
| Scope explosion | The CLI becomes agent runtime, memory system, web research platform, Git client, and package manager | Enforce non-goals and move providers into adapters/extensions |
| False determinism | A hash or prompt is presented as proof of correct execution | Define determinism as reproducible state transitions over declared inputs; preserve semantic uncertainty explicitly |
| Validator theater | Regex rules generate noise and users ignore all diagnostics | AST parsing, evidence-based rules, false-positive fixtures, per-profile severity policy |
| Documentation drift | Specs freeze while hotfixes change behavior | freshness digests, finish-time convergence, optional code/test trace scanning, periodic drift audits |
| Adapter churn | Agent vendors change file paths and command syntax | canonical protocol, thin versioned adapters, capability detection, managed-file manifests |
| Extension supply chain | A workflow package gains agent-level privileges | local-only v1, inspectable manifest, independent integrity/authentication/review/authorization states, origin-bound credentials, no implicit install |
| Multi-agent collisions | Two agents edit the same task or shared file | leases, revisions, worktrees, explicit merge/review actions |
| Context bloat | Every action receives the entire research/spec corpus | progressive packs, budgets, linked resources, selection receipts |
| Hidden destructive behavior | “Convenient” finish operations commit/push/archive unexpectedly | explicit authority boundaries, dry-run, exact operation plans, safe defaults |
| Methodology dogma | EARS/RFC rules are applied to domains they cannot model | typed requirement profiles and workflow-selected validation |
| Partial multi-file commit | A crash updates specs but not archive/events, or vice versa | prepared transaction directory, idempotent roll-forward, fault injection at every operation |
| Digest dishonesty | Formatting churn invalidates everything or canonicalization hides a normative edit | distinct blob and semantic digests, versioned conservative canonicalizers, metamorphic tests |
| Concurrent spec promotion | Two valid work items overwrite each other’s accepted behavior | base/current/proposed three-way promotion and capability locks |
| Policy confusion | A repository or command flag silently weakens company rules | monotonic locked policy floor and `config explain` provenance |
| Platform abstraction leak | Path, process, or rename behavior differs across operating systems | native adapters, platform golden tests, honest support matrix |
| Infrastructure tunnel vision | The team builds a framework for months without a useful user journey | Phase 2 walking skeleton and maximum two infrastructure-only PRs in sequence |
| Knowledge dual authority | Markdown, SQLite, Obsidian Sync, and a remote database all accept edits | one declared authority per topology; indexes/views are rebuildable and writes return through the artifact transaction path |
| Knowledge leakage | Search results or embeddings reveal concepts a caller cannot access | bundle separation in local mode; service-side authorization before retrieval in hosted mode; filters are not ACLs |
| Completion theater | An agent checks boxes and writes a persuasive summary without current evidence | engine-rendered DONE gate, snapshot-bound receipts, changed-path mapping, independent review policy |
| Remote-sync ambiguity | Offline and remote writes reconcile differently or overwrite newer knowledge | no Turso core dependency; explicit conflict/version policy before hosted mode; never dual-write Git and service authority |
| Protocol prose/schema drift | Markdown examples and generated clients disagree with schemas | machine-readable normative schemas, checked examples, conformance vectors, compatibility tests |
| Ambient-target confusion | Current directory, branch, or a shared pointer selects the wrong project/work item | separate root/work coordinates, explicit machine IDs and revisions, no fallback from an invalid explicit root |
| Initiative recursion | “Spec of specs” becomes an unbounded hierarchy and duplicated truth | shallow outcome/slice DAG, measured decomposition trigger, depth cap, child artifacts remain authoritative for child behavior |
| Complexity metric gaming | Fewer files/lines is rewarded despite incomplete or unsafe work | necessity review separate from correctness; safety/completeness gates dominate size/cost metrics |
| Research authority confusion | Old notes or vendored proposals are quoted as current product contracts | documentation source map, normative-status metadata, pinned corpus snapshots, explicit supersession |
| Authority aliasing in retrieval | Semantically similar obsolete prose outranks current contracts | filter by authority/status/version before lexical or vector ranking; show provenance on every context entry |

## 21. Explicit non-goals for 1.0

- hosting or selecting LLMs;
- a general-purpose multi-agent scheduler;
- a Redis/Postgres message bus;
- semantic memory, vector search, or Cognee as a core dependency;
- Turso, Redis, or another remote knowledge service as a core dependency;
- Obsidian as a required editor, sync system, or authorization layer;
- automatic promotion of conversational memory into organizational OKF knowledge;
- a web research proxy;
- automatic code generation by the CLI itself;
- automatic conflict resolution or merging;
- Jira/Linear/GitHub issue synchronization in core;
- a remote extension marketplace;
- formal verification generation for every project;
- forcing all teams into EARS, RFC-style design, BDD, or commit-per-artifact.

These may become integrations or later products. Keeping them out is what lets the v1 engine be small, portable, and reliable.

## 22. Decisions to ratify

The plan makes strong defaults, but company adoption still needs explicit ownership of these decisions:

| Decision | Recommended default | Must be settled before |
|---|---|---|
| Product and executable name | `sdd`; treat `@open-sdd/cli` only as a legacy prototype identifier | publishing the first preview artifact |
| Canonical workspace | `.sdd/`; import other formats, never dual-write them | PR 7 |
| Supported interface | CLI, JSON protocol, and workspace format; Go packages remain internal | protocol preview |
| Runtime distribution | signed native Go binaries; package managers install pinned artifacts rather than executing source | PR 2 / R2 pilot |
| Operating-system support | claim only platforms with path, process-tree, lock, and packed-install CI | R2 pilot |
| Initial agent integrations | generic adapter first, then Codex and Claude; select two more from pilot demand | R3 |
| Organization policy delivery | local authenticated/read-only bundle; no policy server in 1.0 | PR 22 |
| Telemetry | off by default; pilot metrics exported explicitly from local summaries | R2 pilot |
| Release audience and license | internal-first until security and compatibility gates pass | external preview |
| Archive retention and sensitive data policy | repository-owned retention with no prompts or secrets in events | R2 pilot |
| OKF authority and retrieval | OKF v0.2 Markdown is canonical; SQLite/FTS5 is derived; Obsidian is optional; hosted service deferred | PR 17 |
| Completion standard | organization policy plus engine-rendered `done.md` and machine receipt; no force-DONE; exceptions are explicit artifacts | PR 16 |
| Protocol documentation | profile selected by interface type; BCP 14 plus schemas/state/conformance, not generic RFC ceremony | PR 15 |
| Temporal artifact model | living accepted specs; flow-back active work; immutable flow-forward completion history | PR 9 |
| Target selection | nearest validated project root plus explicit work identity/revision; Git branch is never authority | PR 5 |
| Initiative model | optional shallow slice DAG; child work remains independently specifiable and finishable | PR 9 |
| Complexity governance | necessity and correctness are independent checks; safety/completeness dominate size | PR 14 |
| Documentation authority | typed authority/status metadata; superseded sources excluded before relevance ranking | Phase 0 / PR 22 |
| Extension trust | independent trust facts plus per-invocation authorization; no aggregate `trusted` flag | PR 22 |
| Guided workflows | `bug-fix` and `quick-spec` are journey presets composed with an independently risk-derived rigor floor | workflow-graph/guided-workflows specs |

Assign four accountable roles even if one person fills several: product owner for user surface and workflow outcomes; architecture steward for protocol/workspace compatibility; security owner for authority and trust policy; release owner for packaging, support matrix, and incident handling. “The team” is not an owner.

## 23. Immediate next move

Begin with Phase 0 and these decision records:

1. **ADR-001: `sdd` is a deterministic control plane, not an agent runtime.**
2. **ADR-002: `.sdd/` is the canonical v1 workspace; prior formats enter through import adapters.**
3. **ADR-003: all multi-file mutations use recoverable plan/prepare/apply transactions.**
4. **ADR-004: blob, semantic, and derivation digests have separate meanings.**
5. **ADR-005: the supported automation interface is the versioned JSON protocol.**
6. **ADR-006: OKF Markdown is authoritative and SQLite is a disposable local query projection.**
7. **ADR-007: Definition of Done is policy; completion is a snapshot-bound evidence gate recorded as `done.md` plus a machine receipt.**
8. **ADR-008: protocol contracts use a dedicated schema/state/conformance profile.**
9. **ADR-009: accepted specs are living truth, active work may flow back, and completed work is immutable flow-forward history.**
10. **ADR-010: project root and work identity are separate coordinates; Git branches and shared current-feature files are non-authoritative.**
11. **ADR-011: initiatives are shallow outcome/slice DAGs whose status is derived from independently finished child work.**
12. **ADR-012: necessity review is distinct from correctness, and complexity requires traceable evidence.**
13. **ADR-013: documentation authority and temporal validity are evaluated before relevance ranking.**
14. **ADR-014: extension trust is a set of independent facts plus explicit per-invocation authorization.**
15. **ADR-015: the production CLI is Go; the TypeScript prototype is a temporary parity oracle, not a permanent sidecar.**
16. **ADR-016: journey topology, risk rigor, and artifact profiles are orthogonal; presets may add rigor but never lower the policy floor.**

Then land the first three PRs from section 15. Do not build another user-visible capability until the clean-clone pipeline is green and the machine output contract exists.

The first implementation milestone is not “the architecture is in place.” It is demonstrable evidence that a native CLI can execute the Phase 2 golden journey, survive forced interruption, rebuild its projection, and return the same finished state.

## Appendix A — Domain language

Use these terms consistently in code, schemas, documentation, and diagnostics:

| Term | Exact meaning |
|---|---|
| Work item | One bounded proposed change with immutable identity and a lifecycle |
| Initiative | Optional shallow DAG of bounded work items tied to one larger outcome; not a container for duplicate child specifications |
| Workflow | A versioned artifact/action graph interpreted by the engine |
| Journey | Artifact/action topology selected from the shape of the work, such as general change, bug fix, or quick spec |
| Rigor | Risk-derived policy floor (`quick`, `standard`, or `critical`) that a preset or user override cannot silently lower |
| Profile | A composable maintained artifact/validation policy, such as protocol, migration, threat-model, or regression |
| Preset | Named user-facing composition of a journey, default rigor, and artifact profiles; routing may safely escalate it |
| Artifact | A human- or tool-authored durable input/output governed by a schema |
| Action | One currently possible transition or production task computed from facts and policy |
| Accepted spec | Current durable truth for one product capability |
| Delta | A proposed structural change from a recorded accepted-spec base |
| Policy | Rules deciding permission, severity, approval, and transition eligibility |
| Diagnostic | Stable coded observation with severity, source, and remediation |
| Evidence | Recorded observation used to support a requirement, task, or review conclusion |
| Receipt | A structured claim that an action ran against named inputs and produced named outputs/evidence |
| Event | Immutable record of an accepted lifecycle transition or observation |
| Projection | Disposable state derived from artifacts and events for efficient reads |
| Claim | Expiring coordination lease for an exclusive action; not proof of completion |
| Integration | Installed host-specific presentation or invocation support |
| Adapter | Implementation satisfying a real seam while preserving canonical behavior |

Avoid using `phase` for runtime workflow state, `agent` for a workflow role, `state` for both authoritative history and cache, or `spec` for every Markdown file. Ambiguous language becomes ambiguous schemas.

## Appendix B — Real-world edge-case catalog

These cases belong in golden fixtures, fault injection, or conformance suites. They are not “later hardening” if their failure can corrupt truth or falsely claim completion.

### Bug Fix and Quick Spec journeys

| Edge case | Required behavior / acceptance test |
|---|---|
| A one-line authentication bug is submitted as `bug-fix` | Journey remains investigation-first, but rigor escalates to critical; the preset never lowers the risk floor |
| Report cannot be reproduced locally | Record `not-reproduced` plus environment/evidence gaps; do not invent a cause or label verification passed |
| A change removes the symptom but competing root-cause hypotheses remain | Diagnosis gate remains unresolved unless policy explicitly accepts a scoped mitigation with residual-risk evidence |
| The only reproduction is flaky | Store attempt counts/environment and distinguish intermittent reproduction from deterministic regression evidence |
| Observed behavior conflicts with an accepted spec | Require an explicit decision: code defect, spec defect, or intentional behavior change; do not silently “fix” one side |
| Suspected defect is in generated or vendored code | Trace to the generator/upstream boundary and prevent a local patch from masquerading as a durable fix |
| Emergency hotfix is implemented before diagnosis | Explicit hotfix policy records the exception; `finish` still requires retrospective diagnosis, convergence, and exact-snapshot verification |
| Quick Spec reveals a migration, public contract, security boundary, or broad blast radius | Emit `WORKFLOW_ESCALATION_REQUIRED`; preserve answers and upgrade the composition rather than restarting or continuing with quick rigor |
| No material clarification is needed | Clarification stage closes with recorded evidence and proceeds without asking ceremonial questions |
| Host agent produces only two of the three Quick Spec files | Reject the bundle atomically; preserve the prior revision and report missing artifact IDs |
| Requirements change after bundled generation | Design and affected tasks become stale by derivation digest; implementation-ready status is revoked |
| Generated requirements/design/tasks are polished but untraceable | Validation rejects missing stable IDs, derivation links, verification coverage, or unresolved assumptions |
| Quick Spec grows while implementation proceeds | Re-route using current evidence; preserve artifacts and revisions while adding required rigor/gates |
| No agent/model integration is available | CLI emits a bounded generation packet and schemas for manual/external completion; core behavior remains useful offline |

### Workspace, filesystem, and packaging

| Edge case | Required behavior / acceptance test |
|---|---|
| Process dies before `PREPARED` | Next run discards the unprepared transaction; authoritative files are unchanged |
| Process dies after `PREPARED` but before all renames | Recovery idempotently rolls forward and emits exactly one logical event |
| Disk becomes full during staging or journal append | Command fails without a completion event; existing files remain parseable |
| Target is on NFS, SMB, OneDrive, Dropbox, or another sync filesystem | `doctor` detects or warns about unsupported atomic/WAL assumptions; SQLite WAL is not used as cross-host coordination |
| Antivirus/indexer holds a target on Windows | Bounded retry with diagnostic; no infinite spin and no deletion of the staged recovery material |
| A symlink target changes between planning and apply | Final containment/precondition check rejects the mutation |
| `Foo.md` and `foo.md` coexist on Linux then move to macOS/Windows | Case-fold collision is reported before migration or archive |
| NFC/NFD Unicode filenames compare differently across platforms | Canonical collision check blocks ambiguous paths while preserving original display names |
| User switches Git branches while a command is waiting | Worktree/head precondition changes; mutation conflicts instead of applying to the new branch |
| Repository is moved or cloned to a new absolute path | Stored relative identities still resolve; caches containing old roots rebuild |
| Workspace is read-only | Read commands work; mutation returns permission diagnostics before partial staging |
| Nested repository, submodule, or nested `.sdd/` exists | Root selection is deterministic and explainable; command never crosses the selected workspace |
| Native CLI runs without Bun, Node, a Go toolchain, or source files | Empty-directory artifact smoke invokes version/help and one golden read command |
| System clock moves backward | IDs remain unique; leases use monotonic elapsed time where possible; timestamps never decide mutation ordering |

### Concurrency, retries, and lifecycle

| Edge case | Required behavior / acceptance test |
|---|---|
| Two users start work with the same slug | Immutable IDs differ; slug collision gets deterministic disambiguation or a clear conflict |
| Agent submits after requirements changed | Expected revision/input digest rejects the stale receipt and returns the replacement action |
| Agent times out after a successful mutation | Retry with the same idempotency key returns the original outcome without a second event |
| Idempotency key is reused with different input | Hard conflict; original mapping is preserved |
| Lease expires while the original worker is still running | Another worker may claim, but the old worker’s eventual submission is rejected by claim/revision preconditions |
| Two work items finish against the same accepted spec | Structural three-way promotion combines independent IDs and blocks semantic overlap |
| Current spec was renamed during active work | Stable capability identity resolves the rename or requires explicit remapping; no new duplicate spec is created |
| CI and a developer run `finish` concurrently | Canonical locks plus revision checks allow one winner and one inspectable conflict |
| Prepared transaction is discovered after a branch checkout | Recovery verifies original branch/tree preconditions and quarantines rather than applying into a different history |
| Same model performs implementer and reviewer roles | Allowed only for profiles without independence policy; receipt identities make the lack of independence visible |
| Work is no longer wanted halfway through | `abandoned` preserves evidence and reason; it cannot produce a DONE stamp or promote specs |
| A finished work item needs correction | Create a superseding work item or erratum; never rewrite archived history silently |

### OKF, retrieval, Obsidian, and hosted knowledge

| Edge case | Required behavior / acceptance test |
|---|---|
| OKF contains a broken link | Preserve the document; report according to project policy. OKF conformance itself tolerates broken links |
| Knowledge graph contains cycles | Traversal uses visited sets/budgets; cycles are reported, not recursively expanded forever |
| Markdown changes outside the CLI while SQLite is open | Query validates the manifest/digest and reindexes or marks the result stale before returning it as current |
| SQLite index is corrupt or from a newer schema | Quarantine/delete the cache and rebuild from Markdown; never repair authoritative content from the index |
| Branch switch leaves old indexed concepts | Snapshot identity changes and the index is rebuilt or partitioned; results never cross branch snapshots |
| Obsidian rewrites links as `[[wikilinks]]` | Adapter warns or converts through an explicit plan; canonical OKF continues to use standard Markdown links |
| Obsidian metadata cache disagrees with files | Files win; SDD never reads Obsidian’s cache as truth |
| Obsidian Sync and Git edit the same file | `doctor` warns against dual synchronization; conflict files are treated as unresolved artifacts |
| A concept contains credentials or personal data | Secret/privacy scanner blocks index/export according to policy; embeddings and diagnostic bundles exclude it |
| Sensitive and public concepts share one local bundle | Local mode cannot claim per-concept ACL; split bundles/repositories or move behind an authorized service |
| FTS returns a high-ranking but unverified/stale concept | Result includes provenance/trust/freshness and ranking cannot elevate authority |
| Embedding model changes | Vector index version changes and rebuilds; deterministic IDs/links/filters remain stable |
| Corpus contains huge generated files or binary attachments | Policy limits indexing and reports exclusions; no unbounded context or memory allocation |
| Turso replica is offline and accepts local writes | Hosted topology defines conflict/version behavior before enabling offline mutation; local v1 never relies on it |
| Remote database and Git both receive edits | Configuration is invalid because two authorities are declared; sync is refused |
| Conversation memory proposes a “golden nugget” | It remains untrusted memory until explicit promotion creates a sourced, reviewed OKF concept |
| Remote ingestion events arrive twice or out of order | Idempotency keys collapse duplicates; commit ancestry and index revision prevent an older batch from becoming current |
| Git history is rebased after a batch was published | Prior facts remain attributable but become superseded/tombstoned; the service never guesses that equal paths imply equal facts |
| A fork preserves upstream concept IDs | Repository and tenant identity are part of the global key; provenance still records the upstream relationship |
| User loses access after a local clone or index already exists | Central revocation blocks service queries but cannot erase uncontrolled clones; sensitive deployments require managed-device/retention controls and state this limitation |
| A user can retrieve a document but not one of its linked sources | Query filters every node and edge before ranking/context assembly; snippets, counts, and embeddings cannot reveal the hidden neighbor |
| A deleted secret remains in SQLite, vectors, diagnostics, or remote backup | Deletion propagates as a tombstone and cache purge; retention/backups follow an explicit erasure policy and report what cannot be immediately destroyed |
| Same concept changes during an outbox retry | The idempotent record remains bound to the original commit/digest; a new snapshot produces a new record rather than mutating the old fact |
| Malicious knowledge contains agent instructions | Knowledge content is untrusted data, never system policy; context packaging preserves provenance and separates quoted content from executable agent instructions |

### Completion and evidence

| Edge case | Required behavior / acceptance test |
|---|---|
| Every checkbox is `[x]` but no code exists | Missing receipts, changed-path mapping, and requirement evidence block DONE |
| Tests passed and code changed afterward | Result snapshot mismatch stales the evidence |
| Repository had failing tests before the change | Baseline and result fingerprints distinguish pre-existing failures from regressions; report is not a blanket PASS |
| Test is skipped, quarantined, or expected-failing | Status is preserved as skipped/xfail; policy decides whether requirement evidence is sufficient |
| Flaky test passes on retry | Receipt records attempts and failure rate; one eventual pass cannot masquerade as deterministic success |
| Verification is manual | Actor, steps, environment, observation, timestamp, and attachments are required; critical policy may still reject it |
| Change is documentation-only | Workflow requires link/render/example checks rather than irrelevant code coverage |
| Work is a research spike | DONE summarizes hypothesis, method, data, result, decision, and follow-up; it does not claim production behavior |
| Feature is merged but disabled behind a flag | Record separate code-complete, deployable, cohort-enabled, and fully-rolled-out states |
| Migration cannot roll back | Irreversibility, backup/restore, compatibility window, dry run, and explicit approval become completion gates |
| External CI result arrives after local finish attempt | Work remains waiting; receipt is accepted only if it names the current snapshot and trusted CI identity |
| Generated source is committed | Completion links generator, source inputs, tool version, command, and reproducibility check |
| Code moved after line references were recorded | Symbol/path/digest identity detects movement; line numbers are refreshed as convenience metadata |
| Verification output contains a secret | Redaction happens before persistence; policy may invalidate and require credential rotation |
| Required task was intentionally deferred | Original requirement/task is revised or superseded and a follow-up ID is linked; unchecked work cannot sit beneath DONE |
| `done.md` is edited after acceptance | Its digest no longer matches the receipt; status becomes stale and archive/export is blocked until completion is re-evaluated |
| Completion receipt is copied to another branch or work item | Subject work ID, revision, repository identity, and tree/snapshot digest fail validation |
| Organization policy changes after completion | Historical receipt remains an honest decision under the recorded policy digest; current compliance reporting may classify it as legacy or require re-attestation without rewriting history |
| Reviewer authority is revoked after signing | New completions reject that authority immediately; historical validity follows company revocation policy and is reported, not silently rewritten |
| Human approves while files change | Final lock/revision recheck detects the time-of-check/time-of-use race and requires fresh evidence/acceptance |
| External CI says success but cannot identify the exact source | Evidence is unbound and cannot satisfy a gate; branch name or pull-request number alone is insufficient |
| Submodule or Git LFS content changes without the pointer changing locally | Snapshot calculation and evidence policy explicitly include required submodule commits and resolved/LFS object identities |
| `done.md` says PASS while one gate has an approved waiver | The gate remains `WAIVED`, never `PASS`; receipt names approver, reason, scope, expiry, and residual risk |

### Protocol specifications and interoperability

| Edge case | Required behavior / acceptance test |
|---|---|
| Frame/message is truncated mid-field | Parser returns a specified error, consumes bounded resources, and never treats partial input as complete |
| Message is duplicated, reordered, or replayed | Idempotency/order/replay rules produce deterministic state transitions |
| Older peer sees an unknown field or enum value | Specification says ignore, preserve, reject, or negotiate; conformance fixtures cover the behavior |
| Client and server support disjoint versions | Negotiation fails explicitly without downgrade to an undefined hybrid |
| Attacker forces downgrade | Version transcript is authenticated or policy rejects insecure versions; threat appears in security analysis |
| Two extensions allocate the same identifier | Registry/allocation policy detects collision and defines ownership |
| Peer sends oversized nesting, collection, or decompressed payload | Limits are normative; parser fails before unbounded CPU/memory/disk use |
| Text contains invalid UTF-8 or visually confusable identifiers | Encoding and normalization behavior are normative and tested |
| Signed JSON has multiple equivalent serializations | Canonical representation is specified or signatures bind exact bytes; ordinary JSON objects are not assumed canonical |
| Error response is itself malformed | Receiver has a fallback failure path and does not recursively retry forever |
| Retry occurs without an idempotency key | Protocol declares whether operation is unsafe, rejected, or at-most-once by another mechanism |
| Clock-based expiry meets skewed peers | Allowed skew, monotonic timers, and expiry authority are defined |
| Prose example drifts from JSON Schema/ABNF | CI validates or generates examples from conformance fixtures and blocks publication |
| Implementation supports only part of a draft | Implementation-status evidence names exact version and coverage; it is not promoted into normative text |
| Optional extension becomes de facto mandatory | Baseline conformance runs without it and negotiation fixtures prove both enabled and absent behavior |
| Unknown critical extension is received | Peer fails with the specified unsupported-capability error rather than ignoring behavior required for safety |
| Two parsers accept different representations of the same signed message | Differential/canonicalization vectors force one byte or semantic representation before signature/digest verification |
| Batch partially succeeds before the connection drops | Atomicity boundary and replay token make the retry converge without duplicating already-applied operations |
| Cancellation races with commit | Protocol defines the linearization point and returns committed, cancelled, or indeterminate with a resolution operation—never a reassuring guess |
| Diagnostic text echoes hostile payloads | Logs and errors bound, escape, and redact attacker-controlled content while retaining a correlation ID for authorized inspection |

---

## For the chosen ones

The durable advantage is not a better prompt or a more elaborate template. It is making **the transition function** the product:

```text
(workspace facts, artifact digests, policy, event history, command)
  -> (new events, diagnostics, next valid actions)
```

If that function is deterministic, replayable, and versioned, agent models become replaceable execution adapters. If it is not, every “workflow” is merely a prompt convention and every audit trail is retrospective fiction.

The second-order insight is that fluid workflows need stronger—not weaker—state semantics. Let people edit anything, but make invalidation precise. A requirements edit should not lock the team into a phase; it should propagate a digest mismatch to exactly the design decisions, tasks, and evidence derived from the old requirement. That is how the CLI hides complexity without hiding risk.

Finally, resist building a universal plugin system early. The highest-leverage extension point is the workflow schema plus the stable action protocol. Once those are deep enough, most “plugins” collapse into data, templates, validators, or thin host adapters. Every extension mechanism added before that seam stabilizes becomes permanent compatibility debt.
