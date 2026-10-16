# Implementation Tasks: sdd-engine-foundation

## Scope and gates

These tasks deliver an independently implementable foundation: native executable, protocol, read-only workspace snapshot/selection, and dry-run plan. They reserve seams for later `sdd-workspace-transactions`, `sdd-workflow-graph`, `sdd-artifact-validation`, and `sdd-completion-lifecycle` work without implementing or depending on those specs. Test drivers are Go `*_test.go` files in the package under test; `testdata/` contains fixtures only.

Each task has a stable ID, requirement links, and a real dependency list. Waves are derived from the DAG; they are not a license to implement all models before observable slices.

## Tasks

### FND-001 — Freeze protocol vocabulary and schema inventory

- [x] FND-001 — Freeze protocol vocabulary and schema inventory
- **Requirements:** REQ-002, REQ-003, REQ-008, REQ-009, REQ-010
- **Depends on:** []
- Record the product-plan envelope (`schemaVersion`, `command`, `ok`, `workspace`, `work`, `result`, `diagnostics`, `nextActions`), discriminated command/result kinds, nullability, severity, source object, structured next actions, and exit taxonomy.
- **Done when:** a checked-in schema inventory and catalog map has no competing `protocolVersion` envelope and every referenced code has severity/exit/remediation metadata.
- **Completed (2026-10-16):** `protocol/extension-inventory.json` centrally preserves published `1.0`, allocates-but-does-not-publish `1.1`, assigns one owner to every prerequisite/guided extension family, and forbids a competing `protocolVersion` envelope. `protocol/diagnostics.catalog.json` now supplies complete metadata for every diagnostic referenced by the prerequisite and guided specifications. Executable tests validate namespaces, contributors, ownership, exit metadata, and scan `.kiro/specs/**/*.md` so a newly referenced code cannot bypass registration.

### FND-002 — Add offline schema resolver and negative fixtures

- [x] FND-002 — Add offline schema resolver and negative fixtures
- **Requirements:** REQ-002
- **Depends on:** [FND-001]
- Add normative schemas with corrected draft-07 dialect URI, stable `$id`, resolved local `$ref`s, closed discriminators, and valid JSON examples. Use a full draft-07 validator, including conditional keywords; do not use jq as schema validation. Convert symbolic documentation examples into complete, digest-correct runtime fixtures.
- **Done when:** valid examples parse as JSON and malformed refs, null arrays, wrong result kinds, missing severity, string sources, and shell-only next actions fail validation.
- **Completed (2026-10-16):** `protocol/schema/foundation-1.0.json` materializes the normative bundle with the exact Draft-07 URI, stable ID, closed command/result variants, local references, conditional operation rules, and canonical SHA-256 wire syntax. An embedded, network-denying resolver compiles the contract and command fragment in tests. Executable positive fixtures cover status, version, unknown command, and start dry-run; negative fixtures cover unregistered refs and every required malformed shape. The dry-run content digest is recomputed from exact proposed bytes. Semantic plan-digest recomputation remains correctly owned by FND-006/FND-011.

### FND-003 — Define immutable protocol types and outcome renderer seam

- [x] FND-003 — Define immutable protocol types and outcome renderer seam
- **Requirements:** REQ-002, REQ-003, REQ-008
- **Depends on:** [FND-002]
- Implement typed command/result/outcome/diagnostic/next-action values and one renderer input. Preserve empty arrays and declared result nullability. Add redaction and renderer write-error paths.
- **Done when:** human and JSON views derive from one outcome and schema fixtures validate both.
- **Completed (2026-10-16):** `internal/engine` materializes the closed immutable protocol model (commands, four result variants, outcome, diagnostics, next-actions, digests) with defensive copies and schema-exact JSON; `internal/presentation/render` provides the single redacting `Input` seam with human/JSON renderers and write-error propagation. Executable tests cover all constructor/validation branches, redaction of every result variant, and Draft-07 conformance of all five rendered outcome kinds (status, version, doctor, unknown, start-dry-run).

### FND-004 — Build deterministic dependency ports

- [x] FND-004 — Build deterministic dependency ports
- **Requirements:** REQ-003, REQ-015, REQ-016
- **Depends on:** [FND-001]
- Define narrow ports for immutable workspace observation, policy snapshot, filesystem fact reads, clock, and ID. Provide fixed-clock/fixed-ID and fault-injecting fakes. Signal entry points derive from caller context.
- **Done when:** tests can control time/identity/cancellation without globals, intent-only IDs, or `context.Background()` replacement.
- **Completed (2026-10-17):** `internal/engine/ports.go` defines the consumer-owned ports (`Clock`, `IDSource`, `Snapshotter`, `PolicySnapshotter`, `FileFacts`) with production adapters (`SystemClock`, `RandomIDSource`, `OSFileFacts`) and `ErrDependencyFault` for engine error mapping; `internal/testkit` provides `FixedClock`, `SequenceIDSource`, `MemFileFacts` (faults, blocking reads, `os.ErrNotExist` parity), `FixedPolicySnapshotter`, and `FixedSnapshotter`, all race-safe. Executable tests prove time/identity control, fault injection, cancellation propagation (blocked reads honor `ctx.Done`), defensive copies, and native round-trips; no port stores or replaces caller context.

### FND-005 — Implement snapshot decode/normalize and canonical layout discovery

- [x] FND-005 — Implement snapshot decode/normalize and canonical layout discovery
- **Requirements:** REQ-006, REQ-013, REQ-017, REQ-018
- **Depends on:** [FND-004]
- Read canonical `.sdd/work/<slug>/work.yaml`; preserve immutable work ID, slug, revision, and explicit root rules. Implement no-workspace and invalid-explicit-root diagnostics without ancestor fallback. Do not mutate or reopen during pure evaluation.
- **Done when:** empty, malformed, incompatible, multiple-active, invalid-root, and valid-work fixtures produce one immutable snapshot and stable diagnostics.
- **Completed (2026-10-17):** `internal/engine/discovery.go` implements upward `.sdd` marker search (explicit root inspected only, invalid nearest marker never skipped), strict YAML decode of `workspace.yaml`/`work.yaml` (exact version integers, duplicate-ID/slug-mismatch/unknown-field/malformed rejection), ID-ordered immutable `Layout`, ID-only `SelectWork` with zero/single/ambiguous rules, and `LayoutDiagnosticCode` mapping to the catalog. `docs/decisions/0004-yaml-layout-decoding.md` admits `yaml.v3`. Executable fixtures cover every done-criterion case plus native-directory round-trip and byte-identical no-mutation proof.

### FND-006 — Implement typed canonicalization and digest set

- [x] FND-006 — Implement typed canonicalization and digest set
- **Requirements:** REQ-002, REQ-007, REQ-017
- **Depends on:** [FND-005]
- Compute versioned blob, semantic, and derivation digests; bind workflow/policy/validator definition bytes, not names. Preserve meaningful Markdown whitespace/code/order and invalidate unknown meaning conservatively.
- **Done when:** formatting-only and behavior-changing fixtures demonstrate the intended digest distinctions, including dirty working content not represented by Git HEAD.
- **Completed (2026-10-17):** `internal/engine/digest.go` implements the AM-04-narrowed foundation digest set: domain-separated blob digests, work-record/workspace semantic digests, upstream-binding derivation digests, and `ComputePlanDigest`/`VerifyPlanDigest` over every decision-bearing plan field with ordered effects. Fixtures prove formatting-only stability vs behavior-change invalidation, stale-digest rejection, and git-independent dirty-byte binding. Generic artifact canonicalization stays with artifact-validation.

### FND-007 — Implement Engine.Execute error/cancellation rail

- [x] FND-007 — Implement Engine.Execute error/cancellation rail
- **Requirements:** REQ-003, REQ-009, REQ-010, REQ-015
- **Depends on:** [FND-003, FND-004, FND-005]
- Implement decode → normalize → snapshot → decide → plan → verify → outcome for foundation commands. Map errors with `errors.Is/As` and explicit precedence, redact secrets, preserve cancellation, and propagate renderer writes.
- **Done when:** malformed input, cancellation, workspace/schema error, conflict, policy error, external failure, validation failure, and internal defect each have stable category/code/exit behavior.
- **Completed (2026-10-17):** `internal/engine/engine.go` implements `NewEngine`/`Execute`/`UnknownCommand` over version, doctor (with partial-check failures), status (selection reporting), and start dry-run (verified preview plan, non-authorizable warning). `mapFailure`/`classifyFailure` give every category a stable catalog code/severity/exit with explicit precedence (cancellation, assembly, usage, selection, workspace/schema, policy, observation faults, residual). The category matrix test proves all eight behaviors; conflict/external/validation producers are documented as integration-owned. Residual faults land exit-3 with a flagged DIAG_INTERNAL_ERROR refinement note.

### FND-008 — Implement version and help vertical slice

- [ ] FND-008 — Implement version and help vertical slice
- **Requirements:** REQ-001, REQ-004, REQ-013
- **Depends on:** [FND-003, FND-007]
- Implement version/help outside workspace with fixed/reproducible build metadata and no runtime Git/network/package lookup. Keep composition root shallow.
- **Done when:** version human/JSON output is schema-valid and identical in meaning from empty and read-only working directories; help works without workspace discovery. Missing/inaccessible explicit workspace roots do not affect version/help.

### FND-009 — Implement doctor vertical slice

- [ ] FND-009 — Implement doctor vertical slice
- **Requirements:** REQ-005, REQ-013
- **Depends on:** [FND-005, FND-007]
- Check executable, build metadata, current directory, workspace/layout/schema status, and required policy source. Return exit `3` for no workspace, malformed workspace, incompatible workspace, or missing required policy; never mutate or call external tools.
- **Done when:** no-workspace is not reported as a passing doctor result and all checks share outcome rendering.

### FND-010 — Implement status and explicit-selection vertical slice

- [ ] FND-010 — Implement status and explicit-selection vertical slice
- **Requirements:** REQ-006, REQ-018
- **Depends on:** [FND-005, FND-007]
- Report zero/one/selected work item; permit new-work selection with unrelated active items; reject an invalid explicit root without parent fallback. Keep status read-only.
- **Done when:** selection fixtures cover zero, one, many, explicit valid, explicit invalid, and valid-parent/invalid-root cases.

### FND-011 — Implement dry-run plan vertical slice

- [ ] FND-011 — Implement dry-run plan vertical slice
- **Requirements:** REQ-007, REQ-014, REQ-016
- **Depends on:** [FND-006, FND-007, FND-010]
- Build a meaningful, ordered, inspectable plan for canonical work creation with all digest/revision preconditions. Mark unsupported future operations non-authorizable; never apply.
- **Done when:** `--dry-run` has non-empty operations, stable plan digest, structured next action, and byte-for-byte no-side-effect verification.

### FND-012 — Implement CLI parsing, stream separation, and exits

- [ ] FND-012 — Implement CLI parsing, stream separation, and exits
- **Requirements:** REQ-008, REQ-009, REQ-011
- **Depends on:** [FND-007, FND-008, FND-009, FND-010, FND-011]
- Implement global flag semantics, non-interactive no-prompt behavior, concise usage failures, stdout/stderr separation, structured next actions, and stable exit mapping.
- **Done when:** TTY, redirected, JSON, no-color, narrow-terminal, secret-redaction, and malformed-input fixtures pass without requiring prose parsing.

### FND-013 — Add migration parity harness

- [ ] FND-013 — Add migration parity harness
- **Requirements:** REQ-012
- **Depends on:** [FND-008, FND-009, FND-010, FND-011, FND-012]
- Inventory the actual legacy CLI first (currently version and status); place only those frozen input/output fixtures in `testdata/legacy-typescript/`. Place executable harnesses in the relevant Go package as `*_test.go`. Doctor and `start --dry-run` receive new-contract fixtures, not fabricated legacy parity. Record intentional differences and remove the migration harness when parity is proven.
- **Done when:** fixture metadata includes command/stdin/stdout/stderr/exit/schema and no test driver is hidden under `testdata`.

### FND-014 — Record future transaction integration contract

- [ ] FND-014 — Record future transaction integration contract
- **Requirements:** REQ-007, REQ-009, REQ-017, REQ-018
- **Depends on:** [FND-006, FND-011]
- Record the future transaction seam: expected revision, workspace/work IDs, input/policy/workflow/validator digests, idempotency replay/conflict, uncertain recovery, trusted policy precedence, fail-closed removal/expiry/revocation, typed exception scope, and optional Git provenance. Do not execute these scenarios in foundation; they require the later transaction/workflow/completion runtimes.
- **Done when:** the contract and fixture inventory are versioned as integration inputs, without claiming implementation or passing runtime tests here.

### FND-015 — Add schema, static, and build verification

- [ ] FND-015 — Add schema, static, and build verification
- **Requirements:** REQ-001, REQ-002, REQ-016
- **Depends on:** [FND-002, FND-012, FND-013]
- Run format/static checks, full schema/ref/example validation, unit tests, reproducible builds with fixed inputs, native runtime smoke jobs for declared targets, and separate cross-build jobs. Use absolute binary paths after temporary-directory changes; never mask failures.
- **Done when:** CI reports native runtime support separately from cross-build success and demonstrates deterministic build metadata.

### FND-016 — Foundation conformance review and extension seam record

- [ ] FND-016 — Foundation conformance review and extension seam record
- **Requirements:** REQ-003, REQ-017, REQ-018
- **Depends on:** [FND-014, FND-015]
- Review package direction and deep `Engine.Execute` seam, inspect contradiction searches, and record extension seams for future transaction, workflow, artifact-validation, and completion specs. Do not claim their integration in foundation.
- **Done when:** a reviewer can trace each REQ to an executable foundation fixture or a clearly labelled future integration contract.

## DAG and derived scheduling

The dependency graph above is the source of truth. A topological scheduler derives executable waves at runtime; no hand-written wave list is normative. It must reject unknown task IDs, duplicate IDs, self-dependencies, and cycles. A task is complete only when its observable slice and linked fixtures are complete.

## Verification checklist

- All JSON-labelled examples parse and validate against the same checked-in schemas.
- `$ref` resolution works offline; negative fixtures reject missing refs and invalid discriminators.
- No `*_test.go` lives under `testdata`.
- No command mutates during status, doctor, or dry-run.
- No runtime claim is inferred from a foreign cross-build.
- No unsupported future Git/PR/merge, Bun, or performance claim appears in implementation acceptance.
- Required foundation fixtures cover stale-free read-only selection, schema refs, empty-op dry-run rejection, redaction, renderer/cancellation failures, and native/cross-build distinction. Future transaction/workflow/completion scenarios (stale writers, idempotency, crash uncertainty, dirty candidate, policy exception revocation, and validator provenance) remain recorded as integration inputs and are not claimed as foundation runtime tests.

## References

- `requirements.md`
- `docs/kiro-input/sdd-cli-spec-brief.md`
- `docs/architecture/go-project-structure.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/cli-ux.md`
