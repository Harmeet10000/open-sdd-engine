# Design Document: sdd-workflow-graph

## Draft status and design stance

This is a provisional design derived from `requirements.md`. It is not evidence
of user acceptance, implementation, a released schema, or a company trust
decision. The public result additions below are proposals for the single
foundation-owned minor protocol extension in AM-02; this document deliberately
does **not** assign that extension a version number.

The workflow graph is one deep, pure module. A caller gives it a closed,
versioned definition and one immutable set of revision-bound facts. The module
compiles the definition or deterministically evaluates it. It returns graph
identity, derived state, diagnostics, and typed proposals. It neither obtains
authority nor performs an effect.

```text
trusted definition bytes -> decode -> normalize -> compile -> CompiledGraph

CompiledGraph + immutable EvaluationInput
    -> validate pins/facts -> evaluate predicates/readiness/freshness
    -> derive node/initiative states -> order proposals -> Evaluation
```

The design uses plain values and explicit early returns. There is no generic
`Result[T]`, repository/service layer, object-per-node hierarchy, script
runtime, process-provider ancestor object, or graph-owned filesystem adapter.
The interface is intentionally small; compilation, static proof obligations,
runtime state derivation, freshness propagation, compound progress, and
deterministic recommendation all remain behind it.

## Requirement-to-design traceability

| Requirement | Design realization |
|---|---|
| WG-REQ-001 | closed definition model, typed ports, bounded predicate grammar, exact/semantic identity and self-hash exclusions |
| WG-REQ-002 | ordered compile pipeline, reference/type/cardinality checks, dependency and mixed-readiness cycle analysis, route proof and hard budgets |
| WG-REQ-003 | strong three-valued predicate semantics, applicable-route/cardinality evaluation, mandatory-gate proof and structured runtime blockers |
| WG-REQ-004 | closed node states, separate claim observation, authority-aware transition table and authored-node path without a process provider |
| WG-REQ-005 | immutable compiled topology, workflow-owned action attempts, external execution facts, retry/exhaustion and roll-forward rules |
| WG-REQ-006 | pure evaluation pass, canonical output ordering, one primary proposal, blocked/stale explanations and completion handoff |
| WG-REQ-007 | policy-floor composition, typed assessments and proposal/acceptance/application escalation lifecycle |
| WG-REQ-008 | complete pin set, explicit compatibility/migration, owner-provided digest facts and reverse freshness propagation |
| WG-REQ-009 | generic submission/validation fact contracts; no artifact parser, canonicalizer, provider, or claim trust |
| WG-REQ-010 | stable initiative/leaf identity, deterministic aggregate state, all-leaf resume and partial retry |
| WG-REQ-011 | `Engine.Execute` orchestration and an unallocated, foundation-owned public result proposal |
| WG-REQ-012 | canonical fact inventory, projection-free rebuild and cache-loss behavior |

## Module seams and dependency direction

The in-process workflow module is concrete. The principal implementation
surface is:

```go
package workflow

func Compile(source DefinitionSource, limits CompileLimits) (CompiledGraph, []Diagnostic)

func ResolveComposition(input CompositionInput) (CompositionDecision, []Diagnostic)

func Evaluate(graph CompiledGraph, input EvaluationInput, options EvaluateOptions) Evaluation
```

`Compile` and `ResolveComposition` return diagnostics because invalid input is
an expected domain result. Operational decoder/invariant defects are ordinary
Go errors at the engine boundary; they are not represented by a generic result
monad. `Evaluate` accepts only a successfully compiled graph and a validated
immutable fact set. It always returns an evaluation, including a blocked or
stale one.

No interface is introduced for graph algorithms, predicate evaluation,
canonical encoding, or digest calculation. They are in-process logic. The
engine obtains values from the owners and supplies them to the workflow module:

```text
presentation -> engine
                  |
                  +-> workspace: raw snapshot/revisions/apply/recovery
                  +-> artifact-validation: normalized artifacts, policy,
                  |                       assessments, validation results
                  +-> workflow: compile/compose/evaluate (this module)
                  +-> completion: candidate evidence/gates/DONE proposal
                  +-> bounded execution: execution attempts/results/claims
```

The workflow module does not import workspace transaction implementation,
artifact parsing, validator execution, completion, presentation, or platform
packages. Shared values are passed by the engine. Workspace persists workflow
attempt facts and applies exact plans without importing workflow rules.

### Ownership of attempts and execution facts

`WorkflowAttempt` is a workflow-owned immutable domain fact for every accepted
action submission. Workspace persists it. An authored or structural node can
complete from its accepted submission plus artifact-validation facts and
requires no process provider. Only an `execution` node requires an externally
owned `ExecutionAttemptFact`/`ExecutionResultFact`. A `ClaimObservation` is an
optional bounded-execution fact; it is neither a node state nor authority.

This split realizes corrected WG-C-006 and prevents an unavailable process
provider from disabling ordinary authoring, assessment, or structural
validation routes.

## Closed definition language

The graph language has one normative typed model. `DefinitionSource` contains
exact source bytes, a declared media/schema identity, and a workspace-relative
source coordinate. The first implementation publishes a closed JSON Schema at
`workflow/schema/definition.schema.json`; a different serialization requires a
versioned decoder and byte-identity rule, not an implicit YAML/JSON fallback.
This decoder is specific to graph definitions. It is not a second artifact
Markdown/YAML parser.

Stable IDs use lowercase ASCII `[a-z][a-z0-9._-]{0,127}`. Their bytewise order
is therefore portable. Type references are versioned names, not Go types:
`<namespace>/<name>@<major>`. A profile may narrow a type with a versioned
contract digest but cannot use prose as a type.

```text
Definition {
  schemaVersion, definitionId,
  source: {sourceId, sourceVersion, exactDigest},
  semanticVersion,
  canonicalizer: {id, version, definitionDigest},
  factTypes[], nodes[], edges[], mandatoryGates[],
  compatibilityRules[], declaredExtensions[], sourceCoordinate
}

Node {
  id, version,
  kind: authored | assessment | execution | gate | handoff,
  sourceCoordinate,
  inputs: Port[], outputs: Port[],
  applicabilityPredicate?, skipPredicate?,
  requiredGateIds[], profileBinding,
  action: ActionDefinition
}

Port {
  name, direction: input | output, typeRef,
  cardinality: one | zero-or-one | zero-or-many | at-least-one,
  sourceCoordinate
}

Edge {
  id,
  source: {nodeId, portName}, target: {nodeId, portName},
  relation: dependency | freshness,
  predicate?, sourceCoordinate
}

ActionDefinition {
  id,
  kind: submit | assess | execute | accept-composition | handoff-completion,
  submissionTypeRef?, validatorBinding?, executionBinding?,
  remediationTemplate?, sourceCoordinate
}
```

Only `execution` actions may carry an `executionBinding`, and that binding is a
typed descriptor reference owned by artifact-validation/bounded execution. It
contains no shell string, callback, URL, plugin, template evaluation, or
network instruction. `authored`, `assessment`, and `gate` actions carry typed
submission/assessment references. `handoff` carries completion-input facts,
never DONE contents or file effects.

Unknown top-level or semantic fields are rejected. A declared extension has a
registered name, version, closed value schema, semantic/presentation
classification, and canonicalization rule. An unknown extension is potentially
meaningful and therefore cannot be ignored or retain an old identity.

### Bounded predicate grammar

Predicates are data trees, not expressions in a host language:

```text
Expr = Const(bool)
     | BoolFact(factKey)
     | Present(factKey)
     | Equal(factKey, Scalar)
     | NotEqual(factKey, Scalar)
     | In(factKey, Scalar[])
     | Compare(factKey, lt | le | gt | ge, boundedInteger)
     | Not(Expr)
     | All(Expr[])
     | Any(Expr[])
```

`factKey` resolves against `Definition.factTypes`. Predicate fact types are
`bool`, a finite enum, a bounded integer, or explicit presence. Strings,
digests, IDs, and revisions may be tested only for equality against a typed
literal; there is no regex, glob, arithmetic, user function, property walk,
implicit conversion, or expression-to-expression comparison. Empty `all` and
`any` are rejected rather than acquiring surprising identities.

Runtime evaluation is strong Kleene three-valued logic:

| Operator | Decisive result | Otherwise |
|---|---|---|
| `not` | `not true = false`; `not false = true` | `unknown` |
| `all` | `false` if any operand is false; `true` if all are true | `unknown` with the ordered missing reasons |
| `any` | `true` if any operand is true; `false` if all are false | `unknown` with the ordered missing reasons |
| comparison | normal typed comparison when a current trusted fact exists | `unknown` for missing, stale, unverifiable, wrong-subject, or revoked input |
| `present` | `true`/`false` only when the producing snapshot declares the observation complete | `unknown` when absence itself was not observed |

Every `unknown` retains fact key, owner, reason, expected/current subject or
digest when known, and expression source coordinate. Unknown is never stored as
false. A skip occurs only when its skip predicate is `true` and policy permits
that exact scope; false means applicable and unknown blocks the decision.

## Identity and canonical digests

The module keeps source, semantic, compiled, and runtime identities distinct:

```text
DefinitionRef {
  definitionId, schemaVersion, semanticVersion,
  sourceId, sourceVersion, exactDigest,
  canonicalizerId, canonicalizerVersion, canonicalizerDefinitionDigest,
  semanticDigest
}

CompiledDefinitionRef {
  DefinitionRef,
  compilerContractVersion, compileLimitsDigest, compiledDigest
}
```

- `exactDigest` is SHA-256 over exact definition source bytes. A digest claim
  inside a transport wrapper is excluded from those bytes and compared with
  the computed value; a definition never hashes a field containing its own
  final digest.
- `semanticDigest` is domain-separated SHA-256 over the canonical typed
  semantic projection: schema/canonicalizer identity, IDs/versions, fact types,
  ports/cardinalities, node kinds/actions, edges/relations/predicates, gate and
  profile bindings, compatibility rules, and semantic extensions. It excludes
  its own digest, exact digest claim, source ranges, comments, and fields the
  pinned canonicalizer explicitly marks presentation-only.
- `compiledDigest` is domain-separated SHA-256 over the semantic digest,
  compiler-contract version, enforced limits, and a canonical stable-ID
  projection of normalized topology/proof results. It excludes its own digest
  and runtime-local dense integers.

The canonical graph encoder is versioned and lossless for the definition value.
It sorts unordered maps by raw UTF-8 key bytes and preserves every declared
ordered list. It does not reuse artifact-validation's `AV-C14N-1` role names or
pretend an artifact digest is a graph digest. Conversely, artifact, evidence,
candidate, policy, and validator digests are opaque owner facts; workflow never
recomputes them.

The digest graph is acyclic: source bytes produce exact identity; the typed
semantic projection produces semantic identity; compile identity consumes the
semantic identity; runtime state consumes compile identity and external facts.
No digest includes a value that directly or indirectly includes that digest.

## Compilation and static proof obligations

Compilation is a deterministic sequence. It returns no partial runnable graph:

1. Decode the declared schema; reject unsupported version, unknown critical
   fields, malformed coordinates, and resource overrun.
2. Validate stable IDs, uniqueness, node/action/port kinds, explicit
   cardinalities, type references, profile bindings and predicate types.
3. Sort canonical node, edge, gate, fact-type, and predicate records. Resolve
   every stable reference to a dense runtime index.
4. Check port direction, producer existence, relation legality, type/profile
   compatibility, cardinality feasibility, and mandatory output consumers.
5. Reject a cycle in the possible dependency-edge graph. Constant-false edges
   are absent; every other conditional dependency is possible.
6. Prove mixed dependency/freshness readiness safety using bounded conditional
   valuation analysis described below.
7. Prove every mandatory gate and completion route is statically reachable for
   at least one consistent valuation, and that no required route consumes its
   own output/receipt.
8. Produce stable topological order, adjacency ranges, reverse freshness
   ranges, route obligations, proof summaries, compiled digest, and diagnostic
   order.

### Tractable cycle and impossible-route analysis

Static predicates are normalized into shared atomic tests. For each candidate
strongly connected region or mandatory-route region, compilation collects only
the atoms that can affect that region. The atoms have finite declared domains.
The analyzer enumerates consistent valuations in lexicographic fact/literal
order and debits a deterministic proof-step budget.

For mixed readiness cycles, the analyzer materializes the required-readiness
edges active under one valuation, then performs stable DFS. If any consistent
valuation yields a cycle, compilation rejects the graph and reports one
canonical exact cycle path including edge ID and `dependency`/`freshness`
relation. This finds conditional cycles without rejecting a cycle whose edges
are provably mutually exclusive.

For route proof, the analyzer checks whether each valuation leaves a complete
initial-to-terminal route in which every mandatory gate and required output is
satisfied. If no valuation does, the graph is statically impossible. If a
valuation is structurally complete but its required runtime fact later arrives
as unknown, evaluation is blocked rather than retroactively called a compile
error.

This language is intentionally not general SAT/SMT input. Each region has a
hard atom and proof-step limit. Exhausting it returns
`WORKFLOW_COMPILE_ANALYSIS_LIMIT`; compilation never falls back to an
unbounded solver or punts an unproved safety invariant to runtime.

### Provisional hard limits

These reviewable defaults make the complexity promise concrete. A pinned
profile may lower but never raise them. Changing a hard maximum changes the
compiler-contract identity and requires compatibility fixtures.

| Resource | Hard maximum |
|---|---:|
| exact definition bytes | 2 MiB |
| nodes | 2,048 |
| edges | 8,192 |
| total ports | 16,384 |
| predicate AST nodes | 32,768 |
| predicate depth | 32 |
| distinct atoms in one static proof region | 16 |
| static proof steps across one compilation | 2,000,000 |
| emitted compile diagnostics before terminal limit diagnostic | 4,096 |
| runtime fact records | 65,536 |
| runtime evaluation steps | 1,000,000 |

A proof step is one atom assignment, active-edge decision, graph-visit, or
route-obligation check. Counters use saturating arithmetic. The implementation
checks cancellation between bounded regions and before large allocations; no
goroutine is needed.

### Compiled representation

The default representation is array-of-structs with ordered slices:

```text
CompiledGraph {
  ref: CompiledDefinitionRef,
  nodes[], ports[], edges[], predicates[], routeObligations[],
  nodeIndexByStableId, edgeIndexByStableId,
  outgoingOffsets[], outgoingEdgeIndexes[],
  incomingOffsets[], incomingEdgeIndexes[],
  freshnessReverseOffsets[], freshnessReverseEdgeIndexes[],
  topologicalNodeIndexes[], proofSummary, enforcedLimits
}
```

Nodes and IDs are sorted by stable-ID bytes; adjacency entries sort by source
node ID, target node ID, relation (`dependency` before `freshness`), edge ID,
then source coordinate. Maps are lookup-only. Dense indexes are process-local,
rebuilt on compile/load, bounds-checked before evaluation, and never persisted,
serialized as identity, or exposed in public results. There is no SoA/ECS,
pool, arena, `unsafe`, or custom binary cache without profiling evidence and a
deletion criterion.

## Evaluation input and canonical facts

`EvaluationInput` is a complete immutable value assembled by the engine from
owner-produced facts:

```text
EvaluationInput {
  snapshot: {snapshotId, workspaceId, workspaceRevision,
             workId, workRevision, observedReadSetDigest},
  pins: PinSet,
  composition: CompositionRevision,
  policy: EffectivePolicy,
  evaluationInstant,
  assessments[], artifactFacts[], validationFacts[],
  workflowAttempts[], executionAttempts[], executionResults[], claimObservations[],
  completionGateFacts[], initiativeFacts[], workspaceOutcomes[]
}
```

Every slice has a declared canonical key and arrives sorted or is copied and
sorted before acceptance. The workflow module does not retain mutable caller
aliases. Duplicate or contradictory facts for the same owner/subject/revision
produce a diagnostic; “last element wins” is forbidden. `evaluationInstant`
is an explicit observation used for expiry checks, not a call to the system
clock.

Canonical workflow facts are the pin/composition records, workflow attempts,
owner-produced submission/validation facts, initiative/leaf relationships,
and workspace outcome references. Derived node states, dense indexes,
selection views and action lists are projections. Deleting them cannot change
correctness.

## Runtime state derivation

Evaluation performs these stable passes:

1. Verify the complete pins, snapshot coordinates, policy authenticity/expiry,
   fact schemas, subjects, revisions and provenance. Invalid mandatory owner
   facts are blockers, not absent success.
2. Compare current external semantic/derivation/evidence validity with pinned
   inputs and propagate staleness through reverse freshness adjacency.
3. Evaluate node/edge applicability and skip predicates in topological order.
4. Resolve each applicable port's values and cardinality. A dependency is
   satisfied only by a completed source with its required output and current
   validity; terminal failed/cancelled/stale/skipped is not success.
5. Reconcile immutable workflow attempts with external execution/result and
   workspace replay/recovery facts; derive one node state and ordered blockers.
6. Aggregate initiatives from all leaves, not from a display cursor.
7. Build eligible proposals, or a blocked/completed result; sort and select the
   primary proposal.
8. Canonically encode state facts, diagnostics and actions to compute a state
   digest. No timestamps, random IDs, Go map order, cache state or wall clock
   enter the output unless explicitly present in the input.

The closed node states are `unstarted`, `ready`, `running`, `blocked`, `stale`,
`failed`, `cancelled`, `skipped`, and `completed`. `failed` carries
`retry-eligible` or `exhausted`; the latter is terminal for that node under the
current policy. `claimed` is reported only as a separate observation on an
execution-backed action.

### Runtime transition table

The graph derives or proposes transitions; a workspace commit establishes any
authoritative persisted transition.

| State/input | Required authority and facts | Derived result/proposal | Forbidden behavior |
|---|---|---|---|
| unstarted evaluation | compiled graph + current pins/facts | ready, blocked, stale, or policy-authorized skipped | mutate or treat unknown as false |
| ready action proposal | expected snapshot/work/action revisions | typed proposal and required submission schema | claim, authorize, or execute |
| ready + accepted authored/structural validation | committed submission, trusted validation result, matching candidate/action | completed, failed on a negative result, or blocked when result unavailable | require a process provider merely because this is a transition |
| ready + accepted assessment | committed typed assessment with status/provenance | completed only for sufficient trusted assessment; otherwise blocked | infer truth from prose/null |
| ready + execution start | committed workflow attempt + matching provider execution-attempt fact | running | synthesize provider attempt/claim |
| running + execution result | matching workflow/execution attempt and candidate/input subject | completed, failed, cancelled, stale, or blocked | accept another candidate's result |
| running + claim observation | provider-owned lease fact | state remains running; expose separate claim status | acquire, renew, expire, release, or adopt claim |
| any nonterminal + changed required digest | declared freshness edge and owner fact | stale with edge and expected/current digest | rewrite topology or whole-work invalidate without an edge |
| applicable + true skip | policy permits scope and remaining route proof holds | skipped with reason/scope | waive mandatory output/gate |
| failed + retry authorization | non-exhausted or explicit policy intervention, new key/revisions | proposal for next workflow attempt number | overwrite prior attempt or add retry edge |
| running + cancellation outcome | workspace/provider durable outcome | cancelled before intent, or committed/recovery-required after intent | invent rollback or delete history |
| completed reevaluation | output/validity still current | completed or stale | mutate completed attempt in place |
| no graph action | completion prerequisites summarized | completion handoff proposal | decide completion, create DONE/receipt/checklist |

An attempted transition with stale expected coordinates is a conflict. A
read-only evaluation can successfully report the conflict/blocker with exit
`0`; an attempted mutation uses the foundation nonzero category.

## Workflow attempts, retries, and roll-forward

```text
WorkflowAttempt {
  attemptId,
  workId, workRevision,
  nodeId, nodeVersion,
  actionId, actionRevision,
  candidateDigest?, orderedInputDigests[],
  attemptNumber,
  submissionRef?, validationResultRef?, executionAttemptRef?,
  status: proposed | committed | running | accepted | rejected |
          failed | cancelled | stale | recovery-required,
  workspaceOutcomeRef, source, provenance
}
```

`attemptId` is an opaque allocation persisted by workspace; identity validity
also requires the complete tuple above. Attempt numbers are monotonic within
`(workId,nodeId,nodeVersion,actionId,candidate/input subject)`. Missing numbers,
duplicate numbers with different content, or a result that changes the tuple
are integrity diagnostics. Retry allocates the next number; it never edits the
old record or recompiles topology.

`ExecutionAttemptFact` and `ExecutionResultFact` are externally owned and
reference `workflowAttemptId`. They distinguish unavailable, not-run, running,
passed, failed, timed-out, cancelled, truncated, malformed, and stale. Only the
appropriate trusted result can satisfy an execution node. A test fake proves
the value handoff but is marked test-only and cannot satisfy production policy.

After an interruption, graph evaluation consumes the workspace/provider's
committed, replayed, in-flight, failed, quarantined, or recovery-required
outcome. It does not reread mutable files to reconstruct the result, retry an
effect, delete staging, or choose rollback.

## Deterministic next-action selection

An `ActionProposal` is mutation-neutral plain data:

```text
ActionProposal {
  actionId, actionRevision, nodeId, nodeVersion, kind,
  workId, expectedWorkspaceRevision, expectedWorkRevision,
  definitionRef, compositionRef, policyDigest,
  orderedInputRefs[], candidateDigest?, submissionSchemaRef?,
  executionBindingRef?, completionInputRefs[],
  reasonCode, sourceCoordinate
}
```

It contains no approval Boolean, grant, claim, shell command, staging path, or
DONE content. The engine may ask workspace to build an exact mutation plan from
an accepted proposal and submission; workflow never translates it into file
operations.

Eligible proposals sort by stable node ID, action ID, then source coordinate
`(path bytes, start line, start column, unit)`. The default evaluation returns
the first proposal as `primaryAction`. `includeAlternatives` adds the remaining
ordered eligible proposals without changing the primary choice. When none is
eligible, blockers sort by affected node ID, fact/edge/gate ID, diagnostic
code, and source coordinate; one safe owner-provided remediation becomes the
primary next action. When all graph work is complete, no graph action is
invented: the result contains a completion handoff.

Diagnostics and actions never rely on English for semantics. Remediation is a
structured command/argument proposal owned by foundation presentation plus a
human explanation.

## Composition, policy floor, and escalation

`ResolveComposition` receives typed values, not repository prose:

```text
CompositionInput {
  workId, expectedWorkRevision,
  presetRef?, journeyRef,
  requestedRigor?,
  organizationPolicy: RequiredPolicyFact,
  projectNarrowingPolicy?, workRationale?,
  assessment: AssessmentFact,
  graphDefinitionRef, artifactProfileRefs[], validatorContractRefs[],
  evaluationInstant
}

CompositionDecision {
  status: resolved | blocked | escalation-proposed,
  currentCompositionRef?, proposedComposition?,
  effectiveRigor, requiredProfiles[], validatorBindings[],
  reasons[], evidenceRefs[], diagnostics[]
}
```

Rigor is ordered `quick < standard < critical`. The effective floor is the
maximum of trusted organization minimum, project narrowing, trusted assessed
risk, and an explicit requested increase. Required profile/validator sets are
unions; lower-precedence inputs may add but never remove higher-precedence
requirements. Work rationale records a reason but cannot lower a floor.
Missing, expired, revoked, stale, or unverifiable required organization policy
returns blocked. `unassessed` and `insufficient` remain distinct typed states
and yield an evidence-gathering/assessment action when policy allows one.

Escalation has three distinct facts:

1. `CompositionProposal`: current and proposed pins, exact reasons/evidence,
   expected revisions, authority required, scope, expiry requirements.
2. `CompositionAcceptanceFact`: external trusted actor/authority, scope,
   proposal digest, policy digest, issue/expiry/revocation observations.
3. `CompositionApplicationFact`: workspace committed/replayed outcome and new
   work/composition revision.

Evaluation cannot skip from proposal to application. A proposal does not
change the active graph. Acceptance does not persist it. Application creates a
new revision only after workspace verifies the exact accepted proposal. The
journey and compatible history remain when policy allows; changing journey or
lowering rigor uses explicit migration/exception authority. Emergency
reconciliation never removes ordinary later gates.

The issuer, policy distribution, reviewer independence, waiver authority,
retention, and revocation mechanisms remain DEC-01/03/06 company decisions.
The design preserves those facts but does not select a vendor or invent a
signature requirement.

## Pinning, compatibility, and freshness

```text
PinSet {
  graph: CompiledDefinitionRef,
  composition: {id, revision, digest},
  policy: {id, version, definitionDigest},
  artifactProfiles[]: {name, version, definitionDigest},
  validatorContracts[]: {id, version, semanticContractDigest},
  compatibilityContract: {id, version, digest}
}
```

Every pin is exact and resolved by the engine from the immutable snapshot or
offline built-in bundle before evaluation. The graph never falls back to a
display name, newest preset, network lookup, or current embedded definition.
A same-name different digest produces `WORKFLOW_UPDATE_AVAILABLE`/stale facts
and migration data; it does not change active work.

Compatibility is a trusted, versioned rule evaluated over old/new exact and
semantic identities plus an explicit history mapping. A patch label,
self-asserted compatibility flag, or equal display name proves nothing. Node,
predicate, port, gate, profile, policy, trust, or validator-contract changes
conservatively stale the affected facts unless the applicable compatibility
rule proves preservation. Adding a node is compatible only when existing
accepted paths and mandatory-gate/output obligations remain valid.

Freshness consumes owner facts of this form:

```text
ValidityFact {
  subjectRef, subjectRevision,
  semanticDigest, derivationDigest?, evidenceDigest?,
  validity: current | stale | unknown | revoked | unavailable,
  producer, schemaVersion, provenance, observedSnapshotRef
}
```

On change, the evaluator starts at matching freshness edges and traverses the
precompiled reverse adjacency in stable order. It marks only applicable
declared descendants. Propagation stops when an edge is inapplicable or the
target's expected owner digest still matches. Dependency edges are reevaluated
for success but are not silently treated as freshness edges. The stale fact
records edge ID, owner subject, expected/current digest, derivation inputs, and
safe regeneration action.

No node may require its own output, acceptance, verification receipt, or
completion receipt to become ready. The compiler catches direct and indirect
self-output readiness cycles before runtime.

## Generic submissions and validation facts

The graph consumes the IC-07/IC-09 shapes defined by artifact-validation. It
does not parse source, allocate artifact IDs/revisions, canonicalize content,
compute blob/semantic/derivation digests, choose validators, run a provider, or
establish trusted acceptance.

```text
SubmissionFact {
  schemaVersion,
  status: draft | submitted | rejected | accepted | stale,
  artifactId, workId, artifactRevision, sourceRevision,
  actionId, actionRevision,
  blobDigest, semanticDigest, derivationDigest?,
  idempotencyRef?, source, provenance
}

ValidationFact {
  schemaVersion,
  validatorId, validatorSemanticContractDigest,
  validatorImplementationProvenance,
  artifactInputs[], candidateDigest?, inputDigest,
  status: pass | fail | error | skipped | unavailable | stale,
  evidenceRefs[], diagnostics[], producer, provenance
}
```

Graph evaluation accepts a validation fact only when the action, artifact,
candidate/input, validator contract, policy and snapshot subjects match the
proposal. `accepted`/`passed` fields supplied inside agent content are ignored
as authority. Missing, stale, cancelled, truncated, empty, malformed,
unavailable or unverifiable mandatory evidence blocks and never becomes pass.

The graph returns a `TransitionIntent` containing the accepted proposal,
normalized owner-fact references, expected revisions, next workflow-attempt
number, and transition reason. The engine gives that intent plus the submitted
payload to workspace's exact-plan interface. Workspace alone decides
committed/replayed/conflict/recovery state. Same-key/same-fingerprint replay
and same-key/different-fingerprint conflict are consumed as workspace outcomes;
the graph has no second idempotency store.

## Initiatives and compound progress

An initiative is an aggregation view over ordinary graph leaves:

```text
InitiativeFact {
  initiativeId, workId, initiativeRevision,
  leaves[] {leafId, nodeId, stableOrdinal},
  parentSource, provenance
}
```

Parent and leaf IDs never derive from current status or a resume cursor. A new
initiative allocates a new parent ID; retries retain both IDs. Leaf dependency
and freshness behavior uses the same compiled edge semantics as all nodes.

Aggregate status is deterministic:

| Condition in precedence order | Parent status |
|---|---|
| every applicable leaf completed or policy-authorized skipped, and required parent outputs hold | completed |
| any required leaf is stale | stale |
| any leaf is running | active |
| any leaf is ready | actionable |
| a required leaf is failed(exhausted) and blocks every completion route | failed |
| all unresolved required leaves are cancelled with no eligible retry | cancelled |
| no action is eligible and an unresolved required leaf remains | blocked |
| otherwise | unstarted |

The result also returns counts by leaf state and ordered leaf blockers, so an
aggregate label never hides partial success. Failure/cancellation of one leaf
does not cancel an unrelated leaf. A dependent leaf blocks or conditionally
skips according to its compiled edge. Resume scans every leaf and preserves all
completed attempts. `resumeCursor` is a display optimization only and cannot
exclude an eligible leaf. Atomic bundle acceptance is one workspace outcome;
the graph never decomposes or retries a partial bundle effect.

## Engine.Execute and public result proposal

The engine orchestration is:

```text
decode command
 -> workspace obtains one immutable snapshot
 -> artifact-validation supplies normalized policy/artifact/validation facts
 -> workflow verifies pins, compiles/loads, resolves composition, evaluates
 -> command-specific result + foundation diagnostics/nextActions
 -> for submit only: workspace plans/authorizes/applies exact TransitionIntent
 -> re-evaluate from committed owner outcome
 -> foundation human/JSON renderer
```

`next` stops before plan/authorize/apply. `start` records routing/composition
evidence through workspace without mutating a pinned definition. `submit`
checks workspace/work/action revisions and delegates persistence. `finish`
passes a completion handoff to the completion owner.

The closed foundation `1.0` result union cannot carry these values. The
following is a **typed payload proposal** for the single centrally allocated
minor extension; no protocol version is assigned here:

```text
GraphEvaluationResult {
  kind: next-action,
  snapshotRef,
  workId, workRevision,
  compiledDefinitionRef, compositionRef, policyDigest,
  evaluationStatus: actionable | blocked | graph-complete,
  stateDigest,
  nodeStates[], initiativeStates[],
  primaryAction?, alternatives[], blockers[],
  completionHandoff?
}

WorkflowSubmissionResult {
  kind: workflow-submission,
  workId, beforeRevision, afterRevision?,
  actionId, actionRevision, workflowAttemptRef,
  workspaceOutcome: committed | replayed | conflict | recovery-required,
  acceptedFactRefs[], stateDigest?, nextAction?
}

CompositionResult {
  kind: workflow-composition,
  workId, workRevision,
  status: resolved | blocked | escalation-proposed,
  activeCompositionRef?, proposal?, reasons[], evidenceRefs[]
}

CompletionHandoffResult {
  kind: completion-handoff,
  workId, workRevision, candidateDigest?,
  compositionRef, policyDigest, gateFactRefs[], blockers[]
}
```

Foundation must decide the final command/result discriminator names,
nullability, schema IDs, version negotiation, and how these variants join its
closed envelope. Graph-owned nested values are closed and required arrays are
never null. The proposal reuses foundation `WorkspaceRef`, `WorkRef`,
`Diagnostic`, `NextAction`, source units, severities, and exit codes. It does
not reproduce the envelope or claim the guided design's provisional `1.1`.

Human and JSON renderers receive the same result. A blocked read-only `next`
is a successful observation (`0`). A failed attempted validation is `1`,
malformed definition/submission `2` or schema/workspace `3` as appropriate,
stale/conflict/recovery/claim is `4`, required provider failure is `5`, policy
denial is `6`, and interruption after owner cleanup is `130`.

## Diagnostics and failure taxonomy

The stable catalog begins with:

| Code family/example | Meaning | Attempted-operation exit |
|---|---|---:|
| `WORKFLOW_DEFINITION_INVALID`, `WORKFLOW_REFERENCE_MISSING`, `WORKFLOW_TYPE_MISMATCH` | closed schema/reference/type error | 3 |
| `WORKFLOW_DEPENDENCY_CYCLE`, `WORKFLOW_READINESS_CYCLE` | exact static cycle path | 3 |
| `WORKFLOW_ROUTE_IMPOSSIBLE`, `WORKFLOW_GATE_UNREACHABLE` | static proof failure | 3 |
| `WORKFLOW_COMPILE_LIMIT`, `WORKFLOW_COMPILE_ANALYSIS_LIMIT` | deterministic resource exhaustion | 3 |
| `WORKFLOW_FACT_UNKNOWN`, `WORKFLOW_GATE_BLOCKED` | runtime fact/gate unavailable | 0 for observation; owner category for attempt |
| `WORKFLOW_POLICY_UNAVAILABLE`, `WORKFLOW_POLICY_FLOOR` | authority absent/revoked or requested lowering | 6 |
| `WORKFLOW_DEFINITION_MISSING`, `WORKFLOW_UPDATE_AVAILABLE` | missing pin or explicit migration needed | 3 / 4 |
| `WORKFLOW_ACTION_STALE`, `WORKFLOW_SUBMISSION_CONFLICT` | revision/candidate mismatch | 4 |
| `WORKFLOW_VALIDATION_FAILED` | trusted check ran and was negative | 1 |
| `WORKFLOW_EXECUTION_UNAVAILABLE` | required execution provider absent | 5 |
| `WORKFLOW_ATTEMPT_EXHAUSTED` | retry policy reached | 1 or 6 when intervention is denied |
| `WORKFLOW_RECOVERY_REQUIRED` | workspace/provider durable outcome uncertain | 4 |
| `WORKFLOW_CACHE_REBUILD_REQUIRED` | derived projection absent/corrupt | 0 for read-only rebuild/report |

Every diagnostic uses foundation severity and structured source coordinates,
ordered metadata, remediation and retryability. Blocking is a state/policy
fact, not inferred from severity text. AM-01 remains the only severity
reconciliation path.

Cancellation cannot create a state that an owner did not report. During pure
compile/evaluate it stops between bounded passes with no effects. During
submit, workspace/provider outcome determines cancelled versus committed or
recovery-required. Workflow never cleans staging, releases locks, kills a
process, or adopts a lease.

## Verification strategy

The interface is the test surface. Tests are divided by evidence strength:

- package tests exercise definition decoding, canonical identity, compiler
  laws, tri-state truth tables, state transitions, ordering, composition,
  freshness, initiative aggregation and projection rebuild through
  `Compile`/`ResolveComposition`/`Evaluate`;
- fuzz tests mutate definition bytes, IDs/references, predicate trees,
  conditional readiness regions and fact sets under hard limits;
- engine tests use deterministic workspace/artifact/policy/completion/provider
  fakes to prove orchestration shape, human/JSON parity, stdout/stderr, exits,
  redaction and no side effects;
- producer/consumer contract tests run against real workspace,
  artifact-validation, completion and bounded-execution adapters when those
  specs land. A fake passing cannot satisfy durability, process supervision,
  authority, or evidence release gates;
- rebuild tests delete every derived graph cache/projection and compare
  byte-identical evaluation results from canonical facts.

Each WG-F fixture retains the preconditions, action, expected result and
forbidden side effect stated in `requirements.md`. Tasks map every WG-F-001
through WG-F-063 individually to one primary implementation slice; no
consumer fixture is reassigned to the producer or used to justify a duplicate
parser/provider/transaction implementation.

## Shared contracts, amendments, and blockers

| Contract | Owner -> graph use | Design consequence / release blocker |
|---|---|---|
| IC-01 / WG-C-001 | workspace -> workflow | immutable snapshot/revision/read-set facts; no reread; blocked until workspace design/schema/task contract is published |
| IC-04 / WG-C-002 | workflow -> guided/engine | definition/compiler/identity in this document |
| IC-05 / WG-C-003 | workflow -> engine/presentation | evaluation/proposal payload; public publication blocked on AM-02 |
| IC-06 / WG-C-005 | policy/AV facts -> workflow -> engine | composition/escalation; production authority blocked on DEC-01/06 |
| IC-07/09 / WG-C-004 | artifact-validation -> workflow | consume AV envelope/result exactly; no parser/digest/provider duplicate |
| IC-02/03 / WG-C-007 | workspace -> workflow/engine | exact-plan/replay/cancellation outcomes; graph emits only `TransitionIntent` |
| IC-11 / WG-C-006/008 | completion/provider -> workflow | candidate/evidence linkage only where the node requires it; no reverse runtime import |
| IC-12 / WG-C-008 | completion -> workflow/engine | graph emits handoff and consumes gate facts; never DONE/receipt authority |
| WG-C-009 | foundation -> all public consumers | central result union/version allocation required; no locally invented `1.1` |

The following reviewed amendments remain necessary:

- **AM-02 / WG-A-001:** foundation FND-001/FND-002/FND-003/FND-016 owns one
  extension inventory and final version. Existing guided GW-018 must consume
  it rather than publish an independent version.
- **AM-04:** foundation FND-006 must not implement artifact Markdown semantic
  digests; graph identity is graph-specific, while IC-07 digests come from AV.
- **AM-05:** guided GW-001..GW-003 retain journey definitions/fixtures but call
  this compiler/composer rather than implement another graph engine.
- **AM-06/07/08/09:** guided artifact parsing, validator execution,
  transaction, and completion tasks become consumer integration against their
  named owners.
- **AM-10:** exact definition/provenance identity stays distinct from
  policy-proven compatibility.

The concurrent artifact-validation design is consistent with this seam: it
exports IC-07/09 values and explicitly declines workflow compilation. At this
draft point, workspace and completion have requirements but no accepted design
or task contract, and bounded-execution has no spec. Their production
integrations remain blocked; they do not block pure compilation or structural
authored-node evaluation.

## Explicit exclusions

- No `.sdd` path choice, snapshot read, revision allocation, lock, journal,
  idempotency store, staging, apply, recovery, cleanup, or platform durability.
- No artifact Markdown/YAML parsing, artifact identity allocation,
  canonicalization, digest calculation, traceability graph, validator
  selection/execution, provider scratch, or result attestation.
- No process tree, shell, subprocess, resource limit enforcement, scheduler,
  action lease, claim adoption, or ancestor process-supervisor object.
- No completion gate acceptance, candidate observer, DONE/checklist/receipt,
  promotion, archive, or lifecycle effect.
- No Bug Fix/Quick Spec topology or specialized content policy; guided
  workflows own those definitions.
- No second public envelope, hidden network resolution, embedded model,
  arbitrary plugin runtime, Git mutation, auto-fix, or automatic migration.

## Open decisions

1. DEC-01/03/06 must select trusted policy/assessment/waiver issuers,
   revocation and retention. Injectable explicit facts are used until then.
2. DEC-05/AM-01/AM-02 must allocate the shared protocol extension and severity
   treatment before public schema fixtures can freeze.
3. The provisional hard limits above require corpus/benchmark review. Lowering
   them is compatible only when affected definitions already fit; raising them
   changes compiler-contract identity.
4. Workspace must publish IC-01/02/03 concrete types and real adapter tests.
5. Completion must publish IC-11/12 candidate/gate facts. Bounded execution
   must publish execution-attempt/result/claim facts. Their absence blocks only
   their dependent routes and production release claims.
6. The first graph-definition serialization and schema ID must be approved
   with built-in definition migration fixtures; the typed language and digest
   projection above remain the compatibility contract regardless of encoding.

## References

- `requirements.md`
- `docs/kiro-input/next-prerequisite-specs.md`, sections 2–4, 7 and 10
- `docs/kiro-input/prerequisite-review.md`, IC-01–IC-12 and AM-01–AM-11
- `.kiro/specs/sdd-engine-foundation/design.md`
- `.kiro/specs/sdd-workspace-transactions/requirements.md`
- `.kiro/specs/sdd-artifact-validation/design.md`
- `.kiro/specs/sdd-completion-lifecycle/requirements.md`
- `.kiro/specs/sdd-guided-workflows/design.md`
- `docs/architecture/go-project-structure.md`
- `docs/architecture/guided-workflows.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/cli-ux.md`
- `.kiro/steering/quality-and-safety.md`

## For the chosen ones

The real safety proof is not “the graph has no cycle.” It is that readiness is
a proof over two different relations and three different identities.
Dependency proves a successful value exists; freshness proves the consumer is
still bound to the current valid value; revisions prove the proposal was made
against the same administrative state. Keeping those proofs orthogonal lets an
authored node complete without a process provider, lets one semantic edit stale
only declared descendants, and prevents a receipt or retry from quietly
rewriting topology. Most workflow engines collapse those questions into one
terminal-state bit; that is exactly where false progress enters.
