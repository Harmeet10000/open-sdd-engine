# OKF storage, protocol specifications, and evidence-backed completion

**Status:** Research note and architectural recommendation  
**Date:** 2026-10-11  
**Audience:** SDD CLI maintainers, staff engineers, security reviewers, and workflow authors  
**Decision horizon:** local-first 1.0, with a deliberate path to a company knowledge plane  

## Executive conclusion

Three questions look related here, but they have different answers:

1. **The OKF knowledge source should remain Markdown in Git.** That is the reviewable, branchable, portable, offline-capable source of truth for repository-scoped knowledge.
2. **SQLite should be the default local query engine, not the source of truth.** A per-checkout, disposable SQLite database can provide normalized graph traversal, FTS5 search, cached validation facts, and optional vector indexes without turning a binary database into a merge artifact.
3. **Obsidian should be an optional authoring and visualization surface.** It should open the same Markdown tree. It must not become a runtime dependency, synchronization authority, or authorization boundary.
4. **A remote libSQL/Turso service should be optional and downstream.** Introduce it only when the company needs cross-repository discovery, centrally enforced access, usage analytics, or multi-tenant isolation that repository permissions cannot provide. It should ingest committed, digest-addressed facts and remain rebuildable from authoritative repositories.
5. **A completed OpenSpec change needs a gated `done.md` record, not a checked box.** Completion should be the outcome of policy evaluation over fresh artifacts and evidence. The record should explain what changed, why it satisfies the accepted requirements, where the code lives, which verification actually ran, what remains risky, and who or what was authorized to accept it.
6. **The agent protocol should adopt an explicit SDD protocol profile derived from IETF and W3C guidance.** State machines, negative-path behavior, version negotiation, errors, security, privacy, operations, registries, and interoperability vectors need normative treatment. JSON schemas alone cannot specify a protocol.

The core architecture is therefore:

```text
                 human and agent authoring
                           |
                           v
              Markdown + structured metadata
                    in repository Git
                    (authoritative facts)
                      /              \
                     /                \
                    v                  v
       local SQLite projection     optional Obsidian
       FTS + graph + cache          editor and graph view
       ignored by Git                   only
                    |
       committed fact outbox
                    |
                    v
       optional company knowledge service
       libSQL/Turso or equivalent remote store
       auth, tenant boundary, cross-repo queries
       (derived organizational projection)
```

This is not indecision between files and a database. It is an explicit separation between **authority**, **query acceleration**, **human interaction**, and **organizational distribution**.

## Research method and claim discipline

This note uses owning, primary sources: SQLite documentation, Obsidian documentation, Turso/libSQL documentation and source repositories, Git documentation, IETF/RFC Editor publications, W3C specifications, the official Scrum Guide, NIST SSDF, SLSA, in-toto, and the vendored upstream OpenSpec sources.

Statements labeled **Source fact** summarize those sources. Statements labeled **Architectural inference** are recommendations for this project. The distinction matters: a source can establish that SQLite permits one writer or that Obsidian lacks fine-grained shared-vault permissions; it cannot decide which artifact this product should make authoritative.

## 1. What the OKF actually needs to store

Before comparing technologies, define the objects and guarantees. “Knowledge graph” can otherwise mean anything from hyperlinks between notes to a multi-tenant semantic database.

For the SDD CLI, OKF should cover repository and work knowledge such as:

- capabilities, requirements, scenarios, invariants, and constraints;
- architecture decisions and rejected alternatives;
- work items, tasks, dependencies, claims, and completion records;
- code symbols and repository paths;
- validation and operational evidence;
- provenance relationships such as derived-from, verifies, implements, replaces, and supersedes;
- project guidelines and policy sources;
- optional embeddings and retrieval summaries.

The necessary guarantees are not uniform:

| Concern | Necessary guarantee |
|---|---|
| Human-authored product intent | readable without a special service; PR-reviewable; branchable; attributable |
| Stable identity | links survive renames and path moves |
| Graph traversal | typed and directional edges, not filename guessing |
| Search | exact lexical search first; semantic recall optionally |
| Validation | query results identify the source revision and content digest |
| Offline work | normal specification and implementation work continues without the network |
| Concurrency | stale projections never authorize a mutation |
| Access control | the effective principal and tenant boundary are explicit |
| Audit | durable claims point to immutable inputs and evidence |
| Deletion | derived text, indexes, embeddings, replicas, and logs follow retention policy |
| Portability | no editor, hosted database, or model provider owns the only usable representation |

The crucial distinction is between **normative knowledge** and **retrieval infrastructure**. Normative knowledge can authorize work or block completion. Retrieval infrastructure can help locate that knowledge, but must not acquire authority simply because it is faster to query.

## 2. Source facts about each storage option

### 2.1 Markdown and Git

**Source facts**

- Obsidian describes a vault as a folder of Markdown-formatted plain-text files on the local filesystem and explicitly notes that other editors and file managers can modify them. Its metadata cache is derived and can become inconsistent with the files, in which case it can be rebuilt. [Obsidian: How Obsidian stores data](https://obsidian.md/help/data-storage)
- Obsidian documents Git as one supported way to synchronize a vault, while keeping a local copy available offline. [Obsidian: Sync your notes across devices](https://obsidian.md/help/Getting%20started/Sync%20your%20notes%20across%20devices)
- Git's built-in diff and three-way merge behavior is designed for text. For binary data, Git commonly reports that binary files differ or requires a custom text conversion/merge driver. [Git attributes documentation](https://git-scm.com/docs/gitattributes)

**Architectural inference**

Markdown in Git is unusually well aligned with spec-driven development because the same unit can be:

- reviewed in a pull request;
- edited by a human or any coding agent;
- branched with the code it describes;
- merged using familiar repository mechanics;
- cited as `commit:path#stable-id`;
- inspected during an incident with no running service;
- retained according to repository history and legal policy.

Its weakness is not that it is unstructured. Frontmatter and validated headings can carry a schema. Its real weaknesses are expensive global queries, weak transactionality across files, ambiguous identities if filenames are used as keys, and limited centrally enforced authorization after a clone exists.

Those weaknesses justify projections and services. They do not justify hiding the authoring truth inside a binary database.

### 2.2 Embedded SQLite

**Source facts**

- SQLite is an in-process, serverless, zero-configuration transactional engine. It stores a complete database in a single cross-platform file and provides ACID transactions, including recovery after power loss. [SQLite: About SQLite](https://www.sqlite.org/about.html)
- SQLite positions itself as local application storage rather than a client/server enterprise repository. It recommends a client/server engine when many clients directly access the same database across a network or many writers must proceed concurrently. SQLite allows multiple readers but only one writer at a time per database file. [SQLite: Appropriate Uses for SQLite](https://www.sqlite.org/whentouse.html)
- WAL mode permits readers and a writer concurrently, but still has only one writer. Its shared-memory WAL index requires all participating processes to be on the same machine, which is why WAL is not suitable on a network filesystem. [SQLite: Write-Ahead Logging](https://www.sqlite.org/wal.html)
- SQLite does not implement `GRANT` and `REVOKE`; access to the database file is governed by operating-system file permissions. [SQLite: SQL Features That SQLite Does Not Implement](https://sqlite.org/omitted.html)
- FTS5 is an official full-text-search virtual table supporting phrases, prefixes, proximity, Boolean composition, column filters, ranking, snippets, and custom tokenization. [SQLite: FTS5](https://www.sqlite.org/fts5.html)
- SQLite's official Vec1 extension currently provides approximate nearest-neighbor search with Euclidean and cosine distance, but its own roadmap still lists more testing, optimization, graph-based indexes, partition keys, and other capabilities. [SQLite: Vec1 Vector Extension](https://www.sqlite.org/vec1.html)
- SQLite itself presents the database file as a richer alternative to a pile of files: schemas, relationships, constraints, and indexes can coexist in one application file. [SQLite as an Application File Format](https://www.sqlite.org/appfileformat.html)

**Architectural inference**

SQLite is excellent for the **local materialized view** of OKF:

- `nodes` and `edges` tables make graph traversal deterministic;
- FTS5 makes lexical search fast and field-aware;
- SQL constraints expose malformed projections early;
- transactions let one indexing pass become visible atomically;
- a schema version and source digest make staleness observable;
- the database can be deleted and rebuilt when corruption or version skew is detected.

SQLite is a poor default **authoritative collaboration format** for this product:

- the database is a binary Git artifact with poor ordinary PR diffs and merges;
- it has OS-level file access, not row-, tenant-, or role-level authorization;
- a shared file on NFS or synchronized cloud storage risks broken locking semantics;
- worktrees and branches can accidentally query an index built from a different checkout;
- writing knowledge through SQL makes ordinary editor and agent workflows dependent on the CLI.

Use a separate database per checkout or worktree, under ignored runtime/cache state. Never commit a live WAL database as the collaborative source of truth.

### 2.3 Obsidian

**Source facts**

- Obsidian stores notes as Markdown files in a local vault and maintains a rebuildable metadata cache for features including graph and outline views. [Obsidian: How Obsidian stores data](https://obsidian.md/help/data-storage)
- Obsidian's search supports content, paths, properties, tags, tasks, sections, regular expressions, and other operators. [Obsidian: Search](https://obsidian.md/help/Plugins/Search)
- Obsidian Sync retains local vaults, queues offline changes, and synchronizes when the device reconnects and Obsidian is open. [Obsidian: Local and remote vaults](https://obsidian.md/help/sync/vault-types)
- Obsidian Sync offers end-to-end encryption for the remote vault, but the local vault itself is not encrypted by Obsidian. [Obsidian: Security and privacy](https://obsidian.md/help/sync/security)
- Shared vaults currently lack fine-grained permissions: collaborators receive the owner's permissions except that only the owner can invite others. Live co-editing of the same file is not supported, and the documented maximum is 20 collaborators. [Obsidian: Collaborate on a shared vault](https://obsidian.md/help/sync/collaborate)
- Obsidian's own setup guidance states that synchronization is not a backup. [Obsidian: Set up Obsidian Sync](https://obsidian.md/help/sync/setup)

**Architectural inference**

Obsidian is a high-quality optional **lens over the files**. It improves backlink navigation, local discovery, authoring, and visual exploration without owning the representation.

It is not the right core runtime or company governance boundary:

- its permissions do not model repository, work-item, field, or classification policy;
- local files remain accessible to the local operating-system principal;
- shared-vault conflict merging is a second synchronization system alongside Git;
- plugins and `.obsidian` configuration can introduce workstation-specific behavior and noisy repository changes;
- requiring a commercial desktop application would violate agent and CI portability.

The supported integration should therefore be small: an optional generated Obsidian configuration, safe ignore recommendations, and `obsidian://` links for humans when available. The CLI must behave identically without Obsidian installed.

### 2.4 libSQL and Turso

The names require care. As of the official project documentation, **libSQL** is the SQLite fork with remote access and embedded replicas; it inherits SQLite's single-writer model. The separately developed **Turso database** is a newer SQLite-compatible Rust implementation with bidirectional offline sync and concurrent-write ambitions, and its own project page currently describes it as beta. [libSQL project README](https://github.com/tursodatabase/libsql)

**Source facts**

- libSQL is embeddable, adds remote access and embedded replicas, preserves SQLite file-format compatibility when incompatible extensions are not used, and retains the single-writer constraint. [libSQL project README](https://github.com/tursodatabase/libsql)
- Turso's official SDK example supports local-only, remote-only, and embedded-replica connections. Embedded replicas synchronize explicitly with `sync()` or by a configured interval. [Turso Rust SDK quickstart](https://docs.turso.tech/sdk/rust)
- Turso Cloud exposes organizations, databases, groups, authentication tokens, replication/sync, branching, team access, and native vector search as managed capabilities. [Turso documentation](https://docs.turso.tech/blog)
- Turso advocates database-per-tenant isolation and documents schemas that can be propagated to child databases. [Turso: Database per tenant](https://turso.tech/multi-tenancy), [Turso: per-user databases](https://turso.tech/blog/working-with-clerk-and-per-user-databases)
- The libSQL server documentation describes a primary accepting writes while replicas serve reads and poll for WAL updates; it also supports authenticated remote service patterns rather than direct multi-host access to one SQLite file. [libSQL server user guide](https://github.com/tursodatabase/libsql/blob/main/docs/USER_GUIDE.md)

**Architectural inference**

A remote SQLite-compatible service becomes valuable when the product has an actual organizational topology:

- users need cross-repository discovery without cloning every repository;
- access must be revoked centrally;
- a service must filter results by tenant, project, sensitivity, or principal;
- teams need organization-wide metrics, policy coverage, or provenance queries;
- many agents need a low-latency shared projection;
- the company needs geographic placement and managed durability.

It should not become the default 1.0 authority merely because it offers SQL and vector search. That would impose network availability, credentials, tenancy design, retention, incident response, and vendor lifecycle on every local `sdd next` call.

The correct seam is a provider-neutral `KnowledgeProjection` interface. A Turso/libSQL adapter can implement it, but the domain must not expose vendor concepts such as groups or sync tokens.

## 3. Decision matrix

The scores below are architectural judgments for this product, not claims made by the vendors. Five is strongest for the criterion.

| Criterion | Markdown + Git | Embedded SQLite | Obsidian as store | Remote libSQL/Turso |
|---|---:|---:|---:|---:|
| Human-readable source | 5 | 2 | 5 | 2 |
| Ordinary PR diff/review | 5 | 1 | 5 | 1 |
| Branch/worktree alignment | 5 | 2 | 5 | 2 |
| Offline operation | 5 | 5 | 5 | 2–4, depending on replica mode |
| Structured graph queries | 2 | 5 | 3 | 5 |
| Full-text performance | 2–3 | 5 | 4 | 5 |
| Vector retrieval | 1 without sidecar | 3–4 with extension | 2 via plugins | 5 as managed feature |
| Multi-file transactionality | 1 | 5 | 1 | 5 |
| Fine-grained centralized access | 2 through repo hosting | 1 through OS only | 1–2 | 4–5 with application policy |
| Multi-writer service concurrency | 2 through Git merge | 2 | 2 | 4 |
| Auditable code correlation | 5 | 3 | 5 | 3 unless commit anchored |
| Zero-service local setup | 5 | 5 | 2–3 | 1 |
| Company-wide cross-repo query | 1–2 | 1–2 | 1 | 5 |
| Vendor/editor independence | 5 | 5 for SQLite | 2 | 2–4 depending on deployment |
| Suitability as normative source | **5** | 2 | 3 only because files remain source | 2–3 |
| Suitability as projection | 2 | **5** | 4 for humans | **5** for organization scope |

No column wins every concern because the concerns belong to different layers.

## 4. Recommended OKF architecture

### 4.1 Layer A: Git-native normative graph

Keep repository-scoped normative knowledge in Markdown with validated metadata. A node should have identity independent of its path:

```yaml
---
okf_version: 1
id: req.auth.recovery.003
kind: requirement
status: accepted
owner: identity-platform
visibility: repository
sensitivity: internal
supersedes: []
tags: [authentication, recovery]
---
```

Edges should be explicit and typed:

```yaml
relations:
  - type: implemented-by
    target: code.auth.recovery.begin
  - type: verified-by
    target: evidence.auth.recovery.integration
  - type: constrained-by
    target: adr.identity.no-security-questions
```

Rules:

- `id` is immutable; title, slug, and path may change.
- IDs are namespaced by kind or registry to avoid accidental collisions.
- A link by path is a convenience; the stable target is the ID.
- The parser rejects duplicate IDs, missing required fields, invalid edge kinds, and dangling mandatory edges.
- Unknown optional metadata is preserved; unknown authority-bearing fields fail closed.
- Every indexed fact retains repository ID, commit ID or working-tree revision, path, source digest, parser version, and classification.
- Secret values are forbidden. A node may contain a secret reference, never secret material.

### 4.2 Layer B: per-checkout SQLite projection

Recommended location:

```text
.sdd/.runtime/<checkout-id>/okf.sqlite
```

It must be ignored by Git. `<checkout-id>` should derive from the canonical repository identity plus worktree path or Git worktree ID, not just the branch name.

Candidate schema:

```sql
CREATE TABLE projection_meta (
  schema_version       INTEGER NOT NULL,
  canonicalizer_version TEXT NOT NULL,
  source_revision      TEXT,
  source_tree_digest   TEXT NOT NULL,
  built_at             TEXT NOT NULL,
  complete             INTEGER NOT NULL CHECK (complete IN (0, 1))
);

CREATE TABLE nodes (
  id              TEXT PRIMARY KEY,
  kind            TEXT NOT NULL,
  title           TEXT,
  status          TEXT,
  repo_id         TEXT NOT NULL,
  source_revision TEXT,
  path            TEXT NOT NULL,
  source_digest   TEXT NOT NULL,
  sensitivity     TEXT NOT NULL,
  body            TEXT NOT NULL,
  metadata_json   TEXT NOT NULL
);

CREATE TABLE edges (
  source_id       TEXT NOT NULL REFERENCES nodes(id),
  relation        TEXT NOT NULL,
  target_id       TEXT NOT NULL,
  target_external INTEGER NOT NULL DEFAULT 0,
  source_digest   TEXT NOT NULL,
  PRIMARY KEY (source_id, relation, target_id)
);

CREATE VIRTUAL TABLE nodes_fts USING fts5(
  id UNINDEXED,
  title,
  body,
  tags,
  content='nodes'
);
```

Projection rules:

1. Read an immutable workspace snapshot or record all observed input digests.
2. Build into a new database or transaction.
3. Mark it complete only after nodes, edges, FTS, and optional embeddings are consistent.
4. Atomically replace the previous projection.
5. Before any authorization or completion decision, compare projection metadata with the current source revision and relevant digests.
6. On mismatch, re-index or fall back to direct source parsing. Never answer “ready” from stale rows.
7. Use short write transactions and a bounded busy timeout. A single indexing coordinator is simpler than allowing every agent process to write directly.
8. Never place WAL mode on NFS or a cloud-synchronized directory.
9. Run SQLite integrity checks as diagnostics, but recover by rebuilding rather than repairing derived facts in place.

### 4.3 Retrieval order

Use deterministic retrieval before probabilistic retrieval:

```text
explicit ID/path lookup
  -> structural filters and graph traversal
  -> FTS5 lexical ranking
  -> optional vector candidate expansion
  -> policy filter
  -> source revalidation
  -> context pack with citations
```

Vector search is a recall enhancer. It must not decide authority, freshness, visibility, or task completion. Every vector hit must resolve back to a current source node and digest.

Store embedding metadata alongside the vector:

```text
source node ID
source semantic digest
chunking contract version
embedding model identifier and revision
vector dimensions
created time
classification
retention class
```

If any of these contracts change, the embedding is stale. Do not compare vectors from incompatible models or dimensions as if they occupied the same semantic space.

### 4.4 Layer C: optional Obsidian surface

An integration may provide:

- an `obsidian://open` link from `sdd show`;
- recommended graph colors and filters by node kind;
- templates that emit valid OKF frontmatter;
- safe `.gitignore` defaults for workspace-layout churn;
- a command to validate the vault without trusting Obsidian's cache.

It must not:

- require Obsidian for parsing or search;
- make `.obsidian` settings authoritative policy;
- enable or install community plugins silently;
- use Obsidian Sync as the official repository merge mechanism;
- assume the Obsidian cache is current;
- expose classified nodes merely because they are in the same visual graph.

### 4.5 Layer D: optional company knowledge plane

Introduce a remote store only through an outbox of committed facts:

```text
repository commit accepted
  -> build projection batch
  -> classify and redact
  -> sign or digest batch
  -> enqueue idempotent outbox record
  -> remote ingest
  -> verify commit and schema
  -> publish tenant-scoped index revision
```

The remote row should retain:

- organization and tenant ID;
- repository identity;
- immutable commit identity;
- source node ID and source digest;
- schema and canonicalizer versions;
- sensitivity and retention class;
- ingestion event ID and idempotency key;
- deletion/tombstone state;
- optional embedding contract.

Do not let a remote service write directly back into repository Markdown in 1.0. Bidirectional synchronization creates two authorities and makes conflict policy a product feature. A proposed remote edit should become an ordinary repository change or pull request.

### 4.6 Multi-tenancy decision

For a company service, authorization belongs in the application boundary, not in the agent's SQL query.

Recommended boundaries:

- database or service namespace per organization for hard organization isolation;
- tenant/repository partitions below it;
- database-per-tenant when regulatory or customer isolation warrants operational overhead;
- a separately authorized aggregate index containing only approved cross-tenant fields;
- no raw SQL MCP tool exposed to agents;
- capability-scoped query operations such as `knowledge.search`, `knowledge.resolve`, and `knowledge.neighbors` that always apply principal and classification filters.

SQLite alone cannot supply this model because its core access boundary is the database file. Turso/libSQL can support a service topology and database-per-tenant deployment, but the SDD service must still define identity, authorization, audit, retention, and exfiltration controls.

### 4.7 When to graduate to a remote service

Do not trigger the migration based on document count alone. Use observed requirements:

| Trigger | Evidence required |
|---|---|
| Cross-repository search | repeated workflows need facts from repositories not present locally |
| Central revocation | access changes must take effect without controlling every clone |
| Multi-tenant product | knowledge queries serve different legal/security principals |
| Organization analytics | policy and trace coverage must be computed centrally |
| Mobile/thin clients | clients cannot clone or index repositories |
| High query concurrency | a service has measured load that local projections cannot serve |
| Compliance | centrally retained, access-logged evidence is an explicit control |

Until one of these is real, an embedded projection is cheaper, safer, and more reliable.

## 5. Storage and retrieval edge cases

The following cases should become architecture tests, not support folklore.

| Edge case | Required behavior |
|---|---|
| User changes branch while an index build runs | The build is labeled with its observed tree/digests and is never installed as current for a different tree. |
| Two worktrees share `.sdd/cache` | Reject shared mutable projection state; allocate per-worktree databases. |
| Repository lives on NFS or synchronized cloud storage | Disable WAL or relocate runtime state to a local filesystem; never pretend unsafe locking is supported. |
| CLI crashes halfway through indexing | Previous complete projection remains readable; incomplete replacement is discarded. |
| Markdown contains unresolved Git conflict markers | Mark the node invalid and exclude it from authority-bearing queries. |
| Case-only rename on a case-insensitive filesystem | Resolve canonical paths and stable IDs; diagnose collisions before indexing. |
| Symlink redirects a source path outside the repository | Enforce workspace containment after canonical path resolution. |
| Node is renamed or moved | Stable ID preserves edges; path-only links are rewritten or diagnosed. |
| Two offline authors create the same ID | Merge validation blocks acceptance and offers deterministic re-keying; never silently merge nodes. |
| Remote events arrive out of order | Compare commit ancestry and immutable event IDs; do not move the visible index backward. |
| Remote ingest succeeds but acknowledgement is lost | Retry with the same idempotency key and return the original logical result. |
| Repository history is shallow | Mark ancestry-dependent facts unavailable; do not infer a complete lineage. |
| A force push removes an indexed commit | Preserve audit records according to policy but stop presenting it as the repository head. |
| A private node links to a public node | Query authorization applies to every returned node and edge; do not leak the hidden node through counts, titles, or embeddings. |
| Access is revoked while a laptop is offline | Central service access ends immediately; local clone risk is handled by endpoint and repository policy, not falsely claimed as remote revocation. |
| A node's classification becomes more restrictive | Purge or quarantine derived excerpts, embeddings, caches, replicas, and support bundles, not just the source row. |
| Embedding model changes | Maintain separate index generations; never mix incompatible spaces. |
| FTS tokenization handles code, CJK, or identifiers poorly | Make tokenizers/version explicit and retain exact ID/path lookup as the reliable path. |
| Obsidian and Git both sync the same vault | Diagnose unsupported dual-sync configuration or require a single documented merge authority. |
| Obsidian creates conflict copies | Treat them as ordinary untrusted files until IDs and provenance validate. |
| Community plugin modifies frontmatter | Validate all authority-bearing fields at the CLI boundary. |
| Markdown contains instructions aimed at the agent | Treat repository content as data; only approved policy and workflow definitions grant authority. |
| SQL/FTS metacharacters appear in user input | Bind SQL parameters and parse query syntax deliberately; never concatenate queries. |
| A huge note or attachment causes resource exhaustion | Apply documented size, depth, edge-count, and parse-time budgets with actionable diagnostics. |
| An embedding or summary service is remote | Require explicit data-export policy, classification checks, redaction, and provider audit. |
| User deletes a source node | Emit a tombstone for remote projections; keep historical evidence only according to retention policy. |
| SQLite schema is newer than the CLI | Open read-only or rebuild with a compatible version; never downgrade in place implicitly. |
| Projection query returns a code line that moved | Cite commit, path, stable symbol, and digest; line number is a convenience, not identity. |

## 6. A protocol-documentation standard for SDD

### 6.1 What the primary standards actually require or recommend

IETF rules are binding for documents entering the IETF/RFC process. They are not automatically binding on this project. The recommendation is to adopt the useful parts as an internal **SDD Protocol Specification Profile**.

**Source facts**

- The RFC Style Guide defines conventional RFC structure, including self-contained abstract and introduction material, references, Security Considerations, and IANA Considerations. It distinguishes normative references required to implement the specification from informative context. [RFC 7322](https://www.rfc-editor.org/rfc/rfc7322.html)
- BCP 14 defines `MUST`, `SHOULD`, `MAY`, and related terms. RFC 8174 clarifies that these words carry the special BCP 14 meaning only when written in uppercase and when the document includes the required boilerplate. [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119.html), [RFC 8174](https://www.rfc-editor.org/rfc/rfc8174.html)
- The guide for Internet Standards writers emphasizes precision, packet/message diagrams, summary tables, state descriptions, errors, management, and implementability. [RFC 2360](https://www.rfc-editor.org/rfc/rfc2360.html)
- IAB guidance on protocol models recommends explaining the problem, overview, participants, messages, trust relationships, and unobvious design features before detailed mechanisms. [RFC 4101](https://www.rfc-editor.org/rfc/rfc4101.html)
- ABNF is the Internet Standard grammar for specifying text protocols. Grammar defines syntax; prose and state models must still define semantics. [RFC 5234](https://www.rfc-editor.org/rfc/rfc5234.html)
- Security Considerations must perform real analysis of the protocol's threats and defenses, including replay, modification, insertion, deletion, impersonation, man-in-the-middle attacks, secrets, and denial of service. [RFC 3552](https://www.rfc-editor.org/rfc/rfc3552.html)
- IAB privacy guidance addresses identifiability, correlation, surveillance, minimization, retention, disclosure, and user participation. [RFC 6973](https://www.rfc-editor.org/rfc/rfc6973.html)
- Operational guidance covers installation, defaults, migration, coexistence, configuration, observability, faults, performance, debugging, and rollback. [RFC 5706](https://www.rfc-editor.org/rfc/rfc5706.html)
- IANA guidance requires registry designers to specify the namespace, entry fields, initial assignments, allocation policy, reserved/private ranges, and change control. [RFC 8126](https://www.rfc-editor.org/rfc/rfc8126.html)
- Extension guidance warns that poorly coordinated extension points create incompatible variants. Long-unused negotiation paths ossify, so extension mechanisms need ongoing exercise with unknown values. [RFC 6709](https://www.rfc-editor.org/rfc/rfc6709.html), [RFC 9170](https://www.rfc-editor.org/rfc/rfc9170.html)
- Modern robustness guidance rejects relying on vague “be liberal in what you accept” behavior; malformed and illegal inputs need precise handling. [RFC 9413](https://www.rfc-editor.org/rfc/rfc9413.html)
- An Internet-Draft may include a time-sensitive implementation-status section describing implementation maturity, feature coverage, versions, and interoperability evidence. [RFC 7942](https://www.rfc-editor.org/rfc/rfc7942.html)
- IETF interoperation reports should describe independent implementation ancestry, tested features, methods, failures, and ambiguities rather than listing product names. [RFC 5657](https://www.rfc-editor.org/rfc/rfc5657.html)
- The W3C QA Specification Guidelines recommend explicit conformance classes, visible normative material, controlled optionality/variability, and a complete conformance clause. [W3C QA Specification Guidelines](https://www.w3.org/TR/qaframe-spec/)
- W3C's test methodology recommends stable identifiers for testable requirements and the identification of actor, preconditions, required behavior, and observable outcome. [W3C Method for Testable Conformance Requirements](https://www.w3.org/TR/test-methodology/)

There is no universal IETF rule that every protocol RFC contain byte-level test vectors or one exact section layout. **Architectural inference:** the SDD protocol needs vectors because canonicalization, digests, replay, idempotency, migrations, and negative-path behavior are central to correctness. RFC 5769 is a useful primary example of exact protocol vectors with encoding, padding, Unicode, and deliberately difficult cases. [RFC 5769](https://www.rfc-editor.org/rfc/rfc5769.html)

### 6.2 Mandatory SDD protocol specification profile

A document claiming `SDD-PROTOCOL-SPEC/1` conformance should contain the following sections.

#### 1. Metadata and lifecycle

- immutable specification ID;
- title and short name;
- document version and protocol-version range;
- status: draft, experimental, candidate, stable, deprecated, or obsolete;
- owners and approval authority;
- last substantive review;
- `updates`, `obsoletes`, and `replaced-by` relationships;
- registry allocations requested by the document.

Document version and wire protocol version are separate. Editorial changes must not falsely imply wire incompatibility.

#### 2. Abstract, scope, and non-goals

The abstract explains the problem and result without normative detail. The body defines:

- applicability;
- assumptions;
- explicit non-goals;
- unsupported environments;
- whether the protocol is local process I/O, files, HTTP, MCP, or another transport.

#### 3. Requirements language and conformance classes

Include the BCP 14 boilerplate when using its keywords. Define actors such as:

```text
CLI producer
agent consumer
workspace reader
workspace mutator
integration adapter
projection service
policy authority
```

Every normative statement identifies the actor carrying the obligation. “The system MUST” is insufficient when several processes can plausibly be “the system.”

#### 4. Terminology and identities

Specify:

- all domain terms;
- identifier syntax and uniqueness scope;
- case sensitivity and Unicode normalization;
- path normalization;
- timestamp syntax and clock assumptions;
- comparison semantics;
- immutable versus mutable identifiers.

#### 5. Protocol model and trust boundaries

Before field definitions, describe:

- participants and their authority;
- authoritative and derived state;
- trust boundaries;
- the ordinary sequence;
- durable effects;
- invariants;
- recovery boundary.

Diagrams are explanatory. The transition model and normative prose decide behavior.

#### 6. Data model and serialization

For each message, request, response, event, and error:

- formal schema;
- required and optional fields;
- ranges and size limits;
- null versus absent behavior;
- Unicode and line-ending rules;
- ordering and duplicate-member behavior;
- canonicalization when bytes are hashed or signed;
- unknown-field behavior;
- precedence if schema, table, and prose disagree.

CI should generate or check schemas and examples from one source. A tie-break rule is still necessary because deployed documents can disagree despite CI.

#### 7. Normative state machine

Each transition needs a stable ID and these fields:

```text
transition ID
current state
input
preconditions and guards
ordered state mutations
durable side effects
output
next state
error behavior
idempotency behavior
```

Specify initial and terminal states, illegal transitions, timeouts, leases, expiration, cancellation, crash/restart, duplicate delivery, and reordered delivery. Generate implementation tests and human tables from the same machine-readable transition definition.

#### 8. Operations and messages

For each operation:

- purpose and authority required;
- request and response;
- preconditions and optimistic revision fields;
- effects and effect ordering;
- idempotency-key scope and retention;
- retry rules;
- cancellation behavior;
- maximum resource use;
- safe diagnostic fields.

#### 9. Errors

Every error defines:

- stable code;
- trigger;
- scope: field, request, action, work item, or session;
- whether durable state changed;
- retryability;
- caller correction;
- exit-status mapping;
- redaction rules.

At minimum, distinguish:

```text
malformed syntax
schema-invalid input
unsupported protocol version
unknown required capability
unknown optional extension
authentication failure
insufficient authority
missing resource
revision conflict
digest/precondition conflict
idempotency-key input mismatch
claim or lease conflict
resource exhaustion
transient implementation failure
internal invariant violation
```

#### 10. Concurrency, ordering, time, and recovery

State:

- total and partial ordering guarantees;
- monotonic sequence or revision source;
- clock-skew behavior;
- duplicate suppression;
- lost-response recovery;
- partial external-effect handling;
- recovery after a prepared transaction;
- behavior when journal and files disagree.

Wall-clock time should not decide correctness when a revision or monotonic lease token can do so.

#### 11. Versioning and extensibility

Define:

- version negotiation and failure when there is no overlap;
- capability negotiation after version selection;
- unknown field/message/event/enum behavior;
- critical versus ignorable extensions;
- reserved and private-use ranges;
- immutable cross-version invariants;
- downgrade protection;
- deprecation and migration periods.

Required examples:

```text
unknown optional observation field  -> ignore or preserve as specified
unknown required capability         -> fail UNSUPPORTED_CAPABILITY
unknown state-changing event        -> fail closed
unsupported major protocol version  -> reject before mutation
security capability stripped        -> reject downgrade
```

#### 12. Security considerations

Include assets, actors, trust boundaries, attacker capabilities, authentication, authorization, confused-deputy risks, replay, tampering, substitution, resource exhaustion, secret handling, diagnostic redaction, downgrade, extension abuse, malicious repository content, residual risks, and explicitly out-of-scope threats.

“No additional security considerations” is not credible for an agent that reads repositories and causes filesystem, process, or Git effects.

#### 13. Privacy considerations

Document:

- data collected and inferred;
- identifiers linkable across repositories or people;
- prompt, source, and tool output export;
- retention, deletion, and replication;
- diagnostic and support-bundle minimization;
- user/admin visibility and control;
- embedding and vector-index deletion.

#### 14. Operational considerations

Cover installation, upgrade, defaults, mixed versions, backup, restore, rollback, health, liveness, logs, metrics, trace correlation, resource budgets, corrupt state, partial migration, stale locks, and verification of correct behavior.

#### 15. Registry considerations

The SDD project needs controlled registries for:

- protocol messages and action kinds;
- event types;
- diagnostic and error codes;
- artifact and edge kinds;
- schema and predicate URIs;
- capabilities and extension names;
- exit statuses.

Each registry defines ownership, entry shape, initial values, review policy, reserved/private ranges, and whether a code can ever be reused. Published codes should be tombstoned, not reassigned to new meanings.

#### 16. Conformance and test vectors

Each testable normative requirement receives a stable ID. A vector contains:

```text
vector ID
requirements covered
specification and protocol versions
exact encoded input bytes
decoded logical value
initial state
expected output bytes or logical response
expected final state and durable effects
expected error for negative vectors
implementation-independent explanation
fixture and generator digest
```

Include happy path, boundaries, malformed input, unknown values, duplicates, retries, stale revisions, clock anomalies, partial effects, and security downgrades. Passing the vector suite proves covered behavior, not total correctness.

#### 17. Implementation and interoperability report

During candidate status, record:

- implementation ancestry and shared libraries;
- protocol versions tested;
- feature and requirement coverage;
- commands and fixture digests;
- failures and ambiguities;
- unsupported optional behavior;
- evidence from at least two independently derived implementations before claiming interoperability.

Two adapters wrapping the same parser are not independent implementations.

#### 18. References

Separate normative sources required to implement the protocol from informative rationale and examples. Unstable draft dependencies should be visible and policy-gated.

### 6.3 Protocol edge cases the suite must force

| Edge case | Normative question the specification must answer |
|---|---|
| Server commits, response is lost, client retries | Does the same idempotency key return the original result without another effect? |
| Same idempotency key, different request | Is this a stable conflict, and which digest is reported safely? |
| Semantically equal JSON uses different member order or number spelling | Is the key bound to raw bytes or canonical meaning? |
| Old client sends known request plus unknown optional field | Ignore, preserve, or reject? |
| Client sends unknown state-changing event | Fail closed before applying any mutation. |
| Version ranges overlap but required capabilities do not | Negotiation fails with a capability-specific error. |
| Unicode normalization differs | Which normalization is normative before comparison or hashing? |
| Null and absent are interchanged | Are they equivalent for this field? |
| Two writers submit against the same revision | Precisely one succeeds; the other receives the current revision without silent overwrite. |
| Lease expires while a live worker runs | Does completion require a fresh claim token, and can work be adopted? |
| Clock moves backwards | Correctness uses monotonic/revision data, not timestamps alone. |
| Message is duplicated, reordered, or truncated | Parser and state machine converge on specified outcomes. |
| External process succeeds before journal commit | Recovery detects and reconciles the effect without blindly replaying it. |
| Response exceeds transport budget | Pagination, references, or a stable size error is specified. |
| Error captures a secret | Redaction happens before persistence and before returning to an untrusted consumer. |
| Deprecated and replacement fields conflict | Precedence or rejection is explicit. |
| Unknown optional values are never exercised | Continuous conformance tests prevent extension-point ossification. |
| Parser accepts what schema rejects | Conformance follows the normative rule; permissiveness is not silently standardized. |
| Two implementations share generated protocol code | Interoperability report discloses common ancestry. |
| Example contradicts transition table | Normative transition rule wins and CI flags the example. |
| Registry code is removed | Code remains reserved/tombstoned forever. |

## 7. A standard for the OpenSpec `DONE` artifact

### 7.1 Lessons from primary completion and evidence guidance

**Source facts**

- The official Scrum Guide defines the Definition of Done as a formal description of the state of the Increment when it meets the product's required quality measures. It creates a shared understanding; work that does not meet it cannot be treated as part of the Increment. If the organization has a standard, it is the minimum for the team. [Official 2020 Scrum Guide](https://scrumguides.org/scrum-guide.html)
- NIST SSDF calls for organization-defined software security checks, review or analysis based on secure coding standards, recorded and triaged findings, and testing of executable code. It explicitly treats review and test results as evidence to inspect, not merely activities to claim. [NIST SP 800-218](https://csrc.nist.gov/pubs/sp/800/218/final)
- SLSA defines provenance as verifiable information describing where, when, and how an artifact was produced. [SLSA provenance](https://slsa.dev/spec/v1.2/provenance)
- The in-toto statement model binds an attestation to immutable subjects by digest and identifies the type-specific predicate. Its predicate guidance supports conditions and references to evidence; its envelope guidance separates authentication from the statement. [in-toto Statement v1](https://github.com/in-toto/attestation/blob/main/spec/v1/statement.md), [in-toto SCAI predicate](https://github.com/in-toto/attestation/blob/main/spec/predicates/scai.md), [in-toto envelope](https://github.com/in-toto/attestation/blob/main/spec/v1/envelope.md)
- W3C PROV models provenance in terms of entities, activities, people/agents, derivation, versioning, and reproducibility so consumers can assess quality, reliability, or trustworthiness. [W3C PROV Overview](https://www.w3.org/TR/prov-overview/)
- Current upstream OpenSpec verification evaluates task completion, requirement implementation, scenarios/tests, and design adherence, and treats critical findings as archive blockers. Its archive workflow currently warns about incomplete artifacts/tasks and performs specification synchronization and archival. [Vendored OpenSpec verify skill](../../vendor/upstreams/openspec/skills/openspec-verify-change/SKILL.md), [vendored OpenSpec archive skill](../../vendor/upstreams/openspec/skills/openspec-archive-change/SKILL.md)

**Architectural inference**

The Definition of Done is a **policy**. `done.md` is an **instance record showing how one change met that policy**. Mixing them causes teams to copy generic checklists into every change and call the duplication evidence.

A checked task means somebody marked a task. It does not prove the requirement is implemented, the test ran against the final code, the evidence is fresh, or the signer had authority. OpenSpec's own upstream history notes that artifact status can be file-existence based; the SDD design should not inherit that weakness.

### 7.2 Lifecycle and gating semantics

Recommended states:

```text
open
  -> implementation-complete
  -> verification-complete
  -> completion-proposed
  -> accepted
  -> archived
```

They are deliberately not one Boolean. A work item may have all implementation tasks complete but still lack verification or acceptance authority.

`done.md` should be drafted from receipts after verification and finalized by `sdd finish`. Its machine metadata is validated; its human explanation remains reviewable Markdown.

The finish gate:

1. Acquire the work and affected-capability locks in canonical order.
2. Re-read accepted requirements, project guidelines, policy, implementation receipts, and evidence from disk.
3. Verify source digests and work revision against the expected snapshot.
4. Reject unresolved merge markers, stale artifacts, unknown mandatory dependencies, and conflicting promotions.
5. Validate that all required artifacts are present and semantically valid; file existence is insufficient.
6. Evaluate every accepted requirement and scenario against evidence coverage.
7. Run required deterministic checks against the exact candidate revision, or validate immutable external receipts for them.
8. Require explicit, authorized waivers for permitted failures; a waiver never becomes a pass.
9. Resolve code references against the candidate commit or tree.
10. Enforce risk-specific independent review and separation-of-duty policy.
11. Reconcile the delta against the current accepted specs with three-way semantic conflict detection.
12. Generate the final `done.md` plus a machine completion receipt.
13. Stage spec promotion, completion record, journal events, and archive move in one recoverable transaction.
14. Re-check revisions/digests immediately before commit.
15. Commit the transition, rebuild projections, and return the completion receipt.

If any precondition changes, completion restarts from the affected gate. It does not reuse a reassuring but stale report.

### 7.3 Required `done.md` structure

```markdown
---
done_schema: sdd.done/v1
work_id: work_01J...
work_revision: 17
status: accepted
candidate_revision: <git-oid-or-tree-digest>
accepted_spec_revision: <digest>
policy_digest: sha256:...
guideline_digests:
  - path: AGENTS.md
    digest: sha256:...
completed_at: 2026-10-11T22:14:09Z
completion_receipt: receipt_01J...
accepted_by:
  type: human
  id: company-id:1234
---

# Done: passkey login and recovery

## Outcome summary

## Delivered and not delivered

## Requirement and scenario evidence

## Implementation ledger

## Code references

## Verification evidence

## Guideline and policy conformance

## Security, privacy, and operations

## Migration and rollback

## Deviations, waivers, and residual risk

## Acceptance

## Reproduction
```

#### Outcome summary

State the user-visible and system-visible result, not an activity log. Include the important design decision and why the result is considered complete.

#### Delivered and not delivered

List included behavior, intentionally deferred scope, compatibility changes, feature flags, migrations, and removals. “Not delivered” prevents readers from inferring that adjacent capabilities shipped.

#### Requirement and scenario evidence

Use a complete matrix:

| Requirement/scenario ID | Result | Evidence IDs | Implementation references | Notes |
|---|---|---|---|---|
| `REQ-AUTH-003` | pass | `ev_integration_41` | `ref_begin_recovery` | recovery tokens are single-use |
| `SCN-AUTH-003-B` | waived | `waiver_09` | none | approved temporary platform gap |

Allowed results should be a closed registry such as `pass`, `fail`, `waived`, `not-applicable`, and `not-run`. Only policy decides which combinations permit acceptance.

#### Implementation ledger: planned and actual order

The user specifically needs the order in which work was done. Record both the declared order and the actual evidence-backed order:

| Seq | Task ID | Depends on | Planned operation | Actual result | Receipt/commit | Status |
|---:|---|---|---|---|---|---|
| 1 | `TASK-04` | — | introduce token table | migration added | `receipt_12` | accepted |
| 2 | `TASK-05` | `TASK-04` | implement issuance | service and API added | `commit:...` | accepted |
| 3 | `TASK-06` | `TASK-05` | add replay tests | 8 vectors added | `ev_test_39` | accepted |

Do not infer actual order from timestamps alone. Use event sequence/revisions and commit ancestry; clocks can skew and parallel tasks can overlap.

The ledger must also disclose:

- tasks reordered and why;
- tasks split, merged, or abandoned;
- unplanned work introduced during implementation;
- discoveries that changed the accepted specification;
- work performed outside the CLI and how it was adopted.

#### Code references

Every material implementation claim uses a durable reference:

```yaml
id: ref_begin_recovery
repository: company/identity-service
revision: 47e9...        # commit or immutable tree identity
path: src/recovery/service.ts
symbol: RecoveryService.begin
lines: 84-137            # informative convenience only
blob_digest: sha256:...
change: modified
implements: [REQ-AUTH-003]
```

Use repository-relative paths. A line range alone is not durable. Prefer commit plus path plus symbol plus blob digest; lines are for navigation. Deleted and generated code should be identified explicitly. Cross-repository references require repository identity and revision.

#### Verification evidence

An evidence receipt should contain:

```yaml
id: ev_test_39
kind: automated-test
subject_revision: 47e9...
command_argv: [bun, test, packages/core/tests/recovery.test.ts]
working_directory: .
environment_contract: ci-linux-x64/v3
started_event_seq: 1842
finished_event_seq: 1843
exit_code: 0
stdout_digest: sha256:...
stderr_digest: sha256:...
tool_versions:
  bun: 1.x.y
result: pass
producer: ci://identity/98421
```

Store redacted excerpts or references rather than unlimited logs. Hashes prove which evidence is referenced; they do not prove the producer was trustworthy. High-risk policy should require authenticated CI or reviewer receipts.

Evidence kinds may include:

- deterministic validation;
- automated test;
- static analysis;
- code review;
- security review;
- migration rehearsal;
- performance measurement;
- manual exploratory check;
- production observation;
- waiver.

Human evidence must identify the observation procedure and authority, not merely state “manually tested.”

#### Guideline and policy conformance

Record every authority-bearing guideline used, its digest, and the checks it introduced:

| Guideline | Digest | Applicable rules | Evidence |
|---|---|---|---|
| `AGENTS.md` | `sha256:...` | deep module boundaries; final chosen-ones note | review receipt |
| `.sdd/project/policy.yaml` | `sha256:...` | independent auth review | security approval |

This protects against a guideline changing after implementation. A path without a digest is not enough.

#### Security, privacy, and operations

For material changes, summarize:

- trust-boundary changes;
- secrets and personal data touched;
- threat mitigations tested;
- logging and redaction behavior;
- monitoring and alert changes;
- rate/resource limits;
- privacy retention/deletion consequences;
- incident and recovery implications.

#### Migration and rollback

Include:

- forward migration command or automation reference;
- compatibility window;
- rollback or roll-forward strategy;
- irreversible steps;
- backup/recovery evidence;
- feature flag or staged-rollout state;
- operator verification.

“Revert the commit” is not a rollback plan when data or external effects changed.

#### Deviations, waivers, and residual risk

Every deviation has:

- violated or unmet rule;
- reason;
- affected scope;
- compensating control;
- approving authority;
- expiry or remediation work item;
- evidence that the waiver itself was authorized.

A waived check remains visibly waived. Aggregated dashboards must never count it as passed.

#### Acceptance and reproduction

Acceptance identifies the actor, authority source, subject digest, and completion receipt. Reproduction gives a bounded set of commands or receipt references that an authorized reviewer can rerun against the same candidate.

### 7.4 Human record plus machine completion envelope

Do not make Markdown parsing the only machine contract. Emit a sibling receipt or fenced JSON object shaped like an attestation:

```json
{
  "_type": "https://sdd.dev/CompletionStatement/v1",
  "subject": [
    {
      "name": "company/identity-service@47e9...",
      "digest": { "gitTree": "..." }
    }
  ],
  "predicateType": "https://sdd.dev/predicate/change-completion/v1",
  "predicate": {
    "workId": "work_01J...",
    "workRevision": 17,
    "policyDigest": "sha256:...",
    "acceptedSpecDigest": "sha256:...",
    "requirements": [],
    "evidence": [],
    "waivers": [],
    "codeReferences": [],
    "guidelines": []
  }
}
```

This borrows the in-toto separation of immutable subject, typed predicate, and optional authenticated envelope. The first SDD release should not claim in-toto conformance unless it uses the actual in-toto statement and parsing rules. It can either define an SDD-native envelope carefully or register/use an appropriate in-toto predicate later.

### 7.5 What `DONE` must not mean

`DONE` must not mean any of the following in isolation:

- every Markdown checkbox is checked;
- the implementation agent says it is complete;
- code compiles on one machine;
- an artifact file exists;
- tests passed before the final changes;
- a reviewer clicked approve without seeing evidence;
- a remote index reports green while local sources changed;
- the branch merged;
- the feature deployed;
- no known bug is open.

Those can be facts contributing to a policy. None is a universal definition of completion.

## 8. Completion edge cases

| Edge case | Required gate behavior |
|---|---|
| Test passed, then code changed | Evidence subject digest is stale; rerun or obtain valid evidence for the new subject. |
| All tasks checked but a requirement has no implementation | Block; task bookkeeping does not satisfy traceability. |
| Requirement is implemented but no task existed | Permit only if the ledger records unplanned work and policy/evidence pass. |
| A flaky test passes on retry | Record all attempts and flakiness policy; do not collapse the history into one clean pass. |
| Verification tool crashes | Result is `not-run` or `error`, never `pass`. |
| Agent fabricates a command output in Markdown | Only engine/CI-generated receipts count as automated evidence. |
| User manually checks a UI | Record procedure, observer, environment, result, and limitations. |
| Reviewer and implementer are the same agent | High-risk separation-of-duty policy can require a different authorized identity. |
| Requirement changes during final verification | Revision precondition fails and affected evidence becomes stale. |
| Main spec changed after work began | Perform structural three-way reconciliation before promotion. |
| Two changes finish against the same base | Independent deltas may commute; overlapping semantics block for explicit reconciliation. |
| Code is generated | Cite generator input, generator version, generated blob, and whether edits are allowed. |
| Code moved after evidence was captured | Resolve by immutable revision and digest, not current line number. |
| Work spans repositories | Each subject has repository/revision identity; acceptance policy covers the whole set or declares partial completion. |
| Migration is irreversible | Require an explicit roll-forward/recovery plan and elevated approval. |
| Required external service is unavailable | Finish remains blocked unless policy permits a scoped, expiring waiver. |
| Waiver expires after archive | Surface the associated remediation work; do not rewrite historical completion as if the waiver never existed. |
| Vulnerability is discovered later | Append a superseding assessment/event; do not mutate the historical claim silently. |
| Completion record is edited after acceptance | Digest/signature breaks; status returns to proposed until reaccepted. |
| Logs contain secrets | Redact before persistence; store digest/reference to protected evidence if required. |
| Dirty worktree contains unrelated changes | Candidate subject is explicit; completion cannot accidentally attest to unscoped files. |
| Git is unavailable or repository is not Git-backed | Use a versioned workspace tree digest and document weaker provenance semantics. |
| Clock is wrong | Event sequence and immutable revision establish order; timestamps remain descriptive. |

## 9. Recommended phased adoption

### Phase A: local and reviewable

1. Define OKF node IDs, edge registry, frontmatter schema, and parser.
2. Keep Markdown authoritative.
3. Implement a disposable per-worktree SQLite projection with FTS5.
4. Make every query return source citations and projection freshness.
5. Add `done.md` schema, evidence receipts, and a finish gate.
6. Define `SDD-PROTOCOL-SPEC/1` and apply it to the CLI-agent JSON contract.

Exit evidence:

- deleting the database and rebuilding produces equivalent query results;
- branch switching cannot expose facts from the previous tree as current;
- `sdd finish` refuses stale evidence and produces a reproducible completion record;
- an independent minimal client passes the protocol vector suite.

### Phase B: optional human knowledge tooling

1. Add Obsidian-friendly templates, URI links, graph groups, and ignore guidance.
2. Validate behavior with Obsidian absent.
3. Diagnose dual-sync and plugin risk rather than taking ownership of them.

Exit evidence:

- opening the repository as a vault changes no normative behavior;
- external edits are detected and indexed safely;
- `.obsidian` churn cannot change policy or completion state.

### Phase C: company projection pilot

1. Define a provider-neutral projection API and committed-fact outbox.
2. Pilot Turso/libSQL or an equivalent service with two repositories and real principal/tenant boundaries.
3. Implement classification filtering, tombstones, retention, replay, idempotency, and audit.
4. Keep local workflows usable during service outage.
5. Measure actual cross-repo query value before committing to company-wide operation.

Exit evidence:

- rebuilding the remote index from repositories reproduces visible facts;
- revocation and tenant isolation tests pass, including inference through counts and vectors;
- out-of-order and duplicate ingestion converge;
- deletion reaches text, FTS, vectors, replicas, logs, and backups according to documented policy;
- remote unavailability does not block repository-local work unless organization policy explicitly requires an online gate.

## 10. Decisions to record as ADRs

1. **Markdown in Git is the normative repository knowledge representation.**
2. **SQLite is a disposable per-checkout projection and cache.**
3. **Obsidian is optional tooling over the normative files.**
4. **Remote knowledge storage is provider-neutral, optional, and downstream of committed facts.**
5. **Vector retrieval cannot grant authority and every result resolves to a fresh source citation.**
6. **Completion is a policy-evaluated state transition with a `done.md` record and machine receipt.**
7. **The CLI-agent contract follows `SDD-PROTOCOL-SPEC/1`.**

## 11. Rejected alternatives

### Put all OKF content only in SQLite

Rejected for 1.0 because it weakens ordinary PR review, branch merge, editor/agent portability, and disaster inspection. SQLite remains the correct projection.

### Treat an Obsidian vault as the company database

Rejected because Obsidian is an editor and synchronization product over local files, not the product's authorization or multi-tenant execution boundary. Its shared-vault permission and live-editing model do not satisfy the required governance.

### Make Turso mandatory for every CLI command

Rejected because local SDD work should not acquire a network, credential, vendor, tenancy, and availability dependency before cross-repository/company requirements exist.

### Keep only Markdown and use recursive grep forever

Rejected because structured graph queries, freshness checks, typed edges, and large-corpus retrieval deserve a real query projection. The projection should be rebuildable, not absent.

### Make `DONE` an agent-written narrative

Rejected because a narrative cannot prove which candidate, policy, guidelines, and evidence were evaluated. The narrative is valuable only when bound to engine-generated receipts and immutable subjects.

### Make `DONE` a fixed universal checklist

Rejected because risk differs across documentation, UI, authentication, migrations, protocols, and experiments. The schema is stable; required evidence is policy- and risk-profile-dependent.

## Final recommendation

Build the simple experience on top of a deliberately layered system:

```text
author once in reviewable files
index locally for speed
view in any editor, optionally Obsidian
publish committed facts to a company service only when needed
finish through evidence, not assertion
specify the agent protocol's failure behavior as carefully as its happy path
```

The first implementation target should be neither Turso nor an Obsidian plugin. It should be a stable OKF node/edge schema, a rebuildable SQLite projection, and an evidence-backed completion contract. Those interfaces preserve the option to add a company service later without rewriting the repository model.

## For the chosen ones

The dangerous failure mode is not “Markdown becomes slow.” It is **a projection quietly becoming authoritative because it is convenient**. Once an agent can complete work, grant access, or answer compliance questions from an index row without revalidating the source revision, a cache has become an ungoverned control plane.

The second non-obvious trap is bidirectional synchronization. Git, Obsidian Sync, an embedded replica, and a hosted database can each be correct alone while the composition is incorrect. Every additional writer multiplies conflict states. Keep one authoring authority; make every other layer a projection or a proposal producer.

Finally, a DONE artifact is strongest when it is boring to write by hand. The engine should already possess task receipts, source digests, project-guideline digests, code references, test results, waivers, and acceptance authority. `done.md` is the human-readable rendering of that state, not a retrospective story the same agent invents after the work. Completion should be a reproducible theorem over recorded facts:

```text
fresh accepted contract
+ exact candidate subject
+ policy-qualified evidence
+ authorized acceptance
+ conflict-free promotion
= completed work
```

If one operand is missing, the honest state is not DONE.
