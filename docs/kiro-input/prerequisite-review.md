# Prerequisite specification review and shared contract ledger

Status: working review record. Requirements, designs and tasks in this batch are drafts, not user-approved contracts or implemented features. Agent review is not company acceptance. No implementation checkbox may be completed by this document.

Source: conversation `01a1278b-94cb-7c63-a18c-89e44c073dc1`, review continued 2026-10-14 (Asia/Kolkata). Scope: repair the three prerequisite requirement drafts, add the missing completion prerequisite, then draft their designs and tasks. The handoff is [next-prerequisite-specs.md](next-prerequisite-specs.md).

## 1. Ownership and integration rules

Foundation owns the public envelope, command/result discriminator extension process, source coordinate units and exit taxonomy. Workspace owns durable effects, raw snapshots, locking, revisions and recovery. Workflow owns compilation and deterministic routing/composition. Artifact-validation owns typed content and deterministic validation, not process supervision. Completion owns acceptance of evidence and completion decisions, not application of effects. Engine coordinates their inputs and outputs.

Guided workflows own maintained Bug Fix/Quick Spec definitions, specialized profiles and journey acceptance fixtures. They consume generic engines. Context indexes, host adapters, action leases/process supervision, migration UX and releases remain with their later specs.

MoSCoW ownership is intentionally narrow: artifact-validation validates the
profile-declared `ScopePriority` record shape; guided workflows decide and pin
the delivery-scope baseline; completion enforces that baseline and reports its
dispositions. Workflow graph does not derive priority from dependencies, and
workspace persists only an already-authorized exact plan rather than assigning
scope meaning.

Module dependency arrows below mean “imports/calls.” These are not chronological drafting arrows:

```text
presentation -> engine
engine -> workspace, workflow, artifact, validation, completion
validation -> artifact
workflow -> normalized values (may reference artifact value contracts; not validation runner)
completion -> artifact/validation result values (not process or transaction adapters)
platform adapters -> consuming module's declared effect interface
```

Workspace must not import graph, artifact parsing, validation policy or completion to commit an exact authorized plan. Engine obtains normalized artifacts from one raw snapshot and passes typed facts to evaluators. Consumer-owned interfaces at actual effect seams avoid a global interfaces/service-locator package. Schemas and type ownership can be referenced without requiring the future owner's runtime implementation.

Drafting order: foundation -> workspace -> graph -> artifact-validation -> completion -> guided integration. This is a review order, not a mandate to build all shared models first or introduce import cycles. Cross-spec task dependencies must identify actual deliverables, not infer readiness from this order.

Runtime fact flow:

```text
raw snapshot -> parsed/normalized artifacts -> validation and routing facts
facts + explicit trusted policy/observations -> proposed action or completion decision
decision -> exact mutation plan -> scoped authorization -> workspace apply/recovery
committed outcome -> rendering and later snapshot
```

## 2. Shared interface ledger

These entries allocate ownership and minimum obligations. Concrete representations belong in the owner's design; unresolved decisions below block publishing an interoperable schema. Ledger IDs are documentation references, not a new runtime registry.

| ID / interface | Definition owner | Producers -> consumers | Required obligations |
|---|---|---|---|
| IC-01 Snapshot and revision references | Workspace | Raw snapshotter -> engine; normalized facts -> graph/validation/completion | Foundation selection followed by one workspace `RawSnapshot`: workspace/work identity; selected scope; immutable raw bytes and blob identities; stable read-set/generation observation; missing/changed paths; revision coordinates; durable artifact allocation/tombstone/missing-record observations; supported format; no opportunistic rereads or parallel identity store. |
| IC-02 Exact mutation plan and apply/recovery outcome | Workspace; foundation owns public serialization | Engine-coordinated planners -> authorization source -> transaction adapter -> engine | Ordered supported effects; exact old/new digests and nonexistence; snapshot/policy identity; complete read/write preconditions; scope and plan digest; commit intent; replay/recovery state; changed paths and before/after revisions. |
| IC-03 Idempotency | Workspace | Caller key + normalized request -> durable replay record -> engine | Principal/workspace/operation namespace; stable fingerprint independent of retry time; distinct same-key conflict; in-flight/committed/uncertain states; retention tombstone policy; result disclosure authorization. |
| IC-04 Compiled graph and definition identity | Workflow | Trusted definition decoder/compiler -> routing/composition evaluator | Versioned closed language; stable external IDs; deterministic adjacency; typed port references; dependency/freshness distinction; exact source identity distinct from semantic identity; compile diagnostics and limits. |
| IC-05 Action evaluation | Workflow | Graph + snapshot + typed facts -> engine/presentation | State/attempt/action revision; eligibility and blockers; structured next actions; deterministic ordering; candidate/input binding; no effects or invented approvals; terminal node state not whole-work acceptance. |
| IC-06 Composition and escalation | Workflow | Policy/profile facts + typed assessment -> engine and guided journeys | Journey/rigor/profile selection; policy floor; reasons and provenance; explicit unassessed state; pinned definitions; proposal/acceptance/application; preserved history and work identity. |
| IC-07 Typed artifact and digest set | Artifact-validation (artifact module) | Snapshot bytes/profile schema -> normalized artifacts -> validators/graph/completion | Versioned common envelope; profile-defined kinds; persistent IDs; submission state; artifact/source revisions; source ranges with foundation units; blob/semantic/derivation identity and exact dependency ingredients. |
| IC-08 Artifact delta and merge inputs | Artifact-validation (artifact module) | Typed base/proposal/current -> completion | Stable target IDs; pinned base identity and content; add/modify/remove; normalized comparison; invalid references; no durable apply or promotion authorization in parser. |
| IC-09 Validation request/result and execution evidence | Artifact-validation | Approved descriptors + explicit facts/execution provider -> engine/completion | Artifact inputs and separate candidate subject; semantic contract and implementation provenance; status/error distinction; ordered diagnostics; command argv/environment/limits; execution/output references; unavailable provider fails mandatory checks closed. |
| IC-10 Effective policy and authority facts | Artifact-validation normalizes; owner modules enforce their own decisions | Declared trusted source + explicit evaluation instant -> graph/workspace/completion | Version/digest/source; locked floor and value provenance; authenticity basis; scope/principal; expiry/revocation observation; no authority from arbitrary repository prose or a self-asserted role. |
| IC-11 Verification subject and accepted evidence/approval bindings | Completion | Candidate observer + declared verification/reviewer authority -> completion | Implementation read-set including dirty/untracked relevant inputs; environment/build identity; policy-owned exclusions; artifact inputs separate; candidate/policy/gate/plan bindings as required; expiry/revocation; no self-hash cycle. |
| IC-12 Completion proposal, DONE and receipt | Completion | Gate evaluator + merge comparison -> engine -> workspace | Final evidence/approval references; DONE content; immutable receipt; base/current promotion preconditions; lifecycle/archive effects; exact plan; post-verification staleness; preserved historical completion. |

### Common contract obligations

Every owner design must make each applicable IC entry concrete with input/output fields, invariants, failure categories, trust/authority source, freshness rules, format/version handling and positive/negative fixture IDs. An absent field is not a zero/default success. No schema may silently widen the closed foundation `1.0` contract.

Use foundation exit categories: `0` successful observation/action, `1` validation/verification failure, `2` usage/malformed input, `3` workspace/configuration/schema, `4` conflict/stale/recovery, `5` required external process failure, `6` policy/permission, `130` interruption. Successful read-only observation may report blocked work with exit `0`. Follow the foundation's existing internal-error mapping; do not invent a ninth category.

Diagnostics retain structured source and declared `utf8-byte` or `unicode-scalar` units. Until AM-01 is decided, new drafts consume foundation severity vocabulary without silently deleting `critical` or inventing `hint` on its wire format. Whether a diagnostic blocks a transition comes from the applicable code/policy contract, not an unreviewed display-severity conversion.

Fixture ownership follows contract ownership. Consumer fixtures exercise the handoff without reimplementing its producer. Mocks prove integration shape only; actual filesystem, authority and execution guarantees require adapter/conformance evidence. Validation time and external observations are explicit inputs; volatile live state is rechecked at its documented authorization/application point.

## 3. Proposed amendments to existing sources

All proposals below remain unratified. Existing foundation and guided files are preserved during this batch so a reviewer can see exactly what must change. “Block” names the point that cannot truthfully be claimed complete; provisional design can still describe alternatives and chosen recommendations.

| ID | Exact source | Proposal | Compatibility / verification | Block |
|---|---|---|---|---|
| AM-01 | Product plan §12.2; foundation design `Diagnostic.severity`; foundation FND-001 | Reconcile ERROR/WARNING/INFO/HINT with foundation info/warning/error/critical. Publish a total explicit mapping or versioned change; retain hint as structured advisory metadata only if approved. Do not map away critical and call it merely documentation. | Fixtures for every source severity, resulting wire value and independent blocking rule. | Diagnostic schema publication. |
| AM-02 | Foundation design public protocol; guided design “Shared command payloads” text proposing 1.1; GW-018 | One foundation-owned extension schema inventory includes prerequisite effect/command/results and guided payloads. Guided contributes profile payload definitions, not another envelope. Final version allocation needs one decision; do not assign four incompatible 1.1 schemas. | Closed 1.0 positive/negative fixtures unchanged; extension negotiation and unknown-variant fixtures. | Public extension/schema freeze and real integration. |
| AM-03 | Foundation `MutationPlan` and `Operation` definitions; FND-014 | Keep existing create-file/update-file/delete-file spellings; add only necessary typed create-directory/remove-empty-directory/move-file/metadata effects under the shared extension. Directory creation/removal is inspectable plan data, not an implicit side effect of file writes. Separate domain effects from internal journal housekeeping. Rich revision references must translate to existing integer wire coordinates or receive an explicit extension. | Exact-plan digest, unsupported effects, parent/source/destination/emptiness preconditions and replay fixtures. | Durable mutating command release. |
| AM-04 | Foundation FND-006, FND-011, FND-014 | Narrow foundation digest work to native foundation records/preview inputs. Generic typed Markdown AST, profiles, semantic/derivation canonicalization belong to artifact-validation. Foundation may consume versioned raw pinned facts without a reverse dependency on the future parser. | Foundation remains independently testable; generic parser fixtures exist once under AV; consumer tests compare shared results. | Implementation task allocation. |
| AM-05 | Guided GW-001, GW-002, GW-003 | Retain journey definitions and journey-specific composition assertions; call the generic graph compiler/resolver/escalation interface instead of building a second one. | Bug Fix/Quick Spec conformance runs against the prerequisite compiler and pinned definitions. | Guided generic graph implementation. |
| AM-06 | Guided design common artifact envelope; GW-004..GW-007, GW-012 | AV owns common envelope/parser/digests; guided owns specialized content profiles and their fixtures. Preserve submitted/rejected/stale states and profile-defined kinds. GW-012 becomes bundle integration and traceability consumption, not a parser implementation task. | Cross-profile common-envelope fixtures; draft-bundle positive submission; malformed/ambiguous references fail. | Guided typed-artifact integration. |
| AM-07 | Guided GW-008, GW-013 | Consume AV execution evidence and CL candidate/evidence acceptance. Keep journey-specific regression assertions. Bounded-execution owns process controls; approved external-provider evidence must declare trust. No fake process adapter satisfies production verification. | Candidate mismatch, missing provider, forged result, timeout and independent reviewer tests. | Production execution/evidence claim. |
| AM-08 | Guided GW-014 | Keep atomic bundle integration; delegate commit/recovery/idempotency to workspace and acceptance facts to their owner. | Partial bundle rejection, exact same-key replay, changed request conflict and crash recovery. | Guided durable submission integration. |
| AM-09 | Guided GW-015, GW-016 | Retain journey-specific gate inputs and lifecycle handoff; completion owns gate acceptance, DONE/receipt generation and promotion/archive proposal. Workspace owns commit. | No ready-to-finish shortcut; post-verification edit blocked; output artifact cannot satisfy its own prerequisite. | Guided completion integration. |
| AM-10 | Product plan §5.5; guided composition definition digest wording; GW-001 | Separate exact source/binary provenance from semantic compatibility/freshness identity. A patch label alone proves neither compatibility nor continuing trust; policy/contract revocation can invalidate despite equal semantics. | Cosmetic vs behavioral vs security/trust changes; old receipts immutable; cache eligibility tested. | Definition and evidence compatibility acceptance. |
| AM-11 | Product plan §8.5; foundation native-format reader | Current native format is the documented integer 1; no historical previous format is invented. Define supported migration pairs when an actual prior format contract is selected. | Unknown/newer/partial format fails safely; init/migration only explicit mutations. | Native-format migration support claims. |

## 4. Company decisions and honest implementation gates

| ID | Decision | Drafting rule until decided | Blocks |
|---|---|---|---|
| DEC-01 | Trusted actor/policy/CI provider and reviewer independence | Use explicit provider facts and injectable test authorities; local editable files are not tamper-resistant attestation. Do not mandate signatures for every local assessment without an issuer/key model. | Organization-governed production approval. |
| DEC-02 | Native OS/filesystem support and durability matrix | Document necessary primitives; unsupported mutation fails closed. Cross-build is not native runtime proof; sync-managed directories are not automatically equivalent to a network filesystem. | Platform durability certification. |
| DEC-03 | Evidence, replay and audit retention/privacy | Specify required invariants and bounded policies, with safe expiry outcomes. Do not choose a company retention duration implicitly. | Destructive retention/compaction deployment. |
| DEC-04 | Approved execution provider before bounded-execution | Mandatory unavailable checks cannot pass. Pure validators and test fakes are usable; external CI/manual evidence requires declared scope/authenticity and cannot masquerade as executed tests. | Execution-backed completion in production. |
| DEC-05 | Protocol extension allocation and severity reconciliation | Preserve 1.0, mark extension drafts provisional, resolve AM-01..03 centrally. | Interoperable schema freeze. |
| DEC-06 | Emergency exception/waiver policy | Explicit scoped grants, authority, expiry and reconciliation requirements; no Boolean bypass. | Emergency production journey. |

## 5. Review deliverables and verification record

Each repaired requirements file must contain stable requirement/acceptance IDs, old-to-new mapping, handoff topic coverage and adversarial fixtures with expected outcomes. Each design must map requirements to concrete interfaces, state transitions, exact-plan effects and test surfaces. Each task must remain unchecked, have a stable ID, traceable requirements, explicit dependencies, owned files, verification and completion criteria.

Actual document-level checks completed on 2026-10-15 (Asia/Kolkata): all four
prerequisite requirement files have unique requirement IDs and balanced code
fences; the requirement counts are workspace 12, workflow graph 12,
artifact-validation 21 and completion 17. The local task DAGs are acyclic with
no unknown local dependency: workspace 14 tasks/40 edges, workflow graph 14/31,
artifact-validation 14/33 and completion 15/35. Every requirement ID is named
by at least one local task; all task checkboxes remain unchecked. Fixture
coverage is expressed through canonical inclusive ranges and task matrices:
workspace 49, graph 63, artifact-validation 26 and completion 26 fixture IDs.

These are structural reviews only. No Go implementation, schema suite, native
filesystem test, execution-provider test, company authority verification, or
production acceptance was run or implied. The remaining company decisions and
amendments above remain blockers.

### Kiro metadata caution

This review intentionally did not rewrite `.config.kiro` metadata. The current
workspace has duplicate `specId` values for `sdd-engine-foundation` and
`sdd-guided-workflows`, and separately for `sdd-workspace-transactions` and
`sdd-workflow-graph`; `sdd-completion-lifecycle` has no `.config.kiro` file.
If Kiro uses `specId` as a persistent primary key, blindly assigning or
changing those values could disconnect conversation/history state. Before using
these drafts in Kiro automation, create/reconcile the completion metadata and
repair duplicate IDs through Kiro's own migration/creation flow, preserving a
backup of its state. This is a tooling-integrity issue, not a defect in the
Markdown contracts.

## For the chosen ones

Two authorities are deliberately separated: domain modules decide what is permitted, and workspace decides whether those exact permitted effects can still be committed against observed state. A receipt records the resulting decision; it does not confer authority on itself. Reviewing this separation catches more integration errors than counting interfaces or checking directory names.
