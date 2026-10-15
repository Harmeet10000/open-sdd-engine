# Requirements Document: sdd-engine-foundation

## Introduction

This specification establishes the native, deterministic Go foundation for `sdd`. It freezes one public protocol contract and one deep application seam, `Engine.Execute(ctx, Command) Outcome`, while delivering the read-only/version and inspectable dry-run slices needed by later workflow specifications. It does not own workflow semantics, durable mutation, process supervision, Git mutation, or completion; future specifications may extend reserved seams for those capabilities without being foundation implementation dependencies.

The product plan is authoritative for the public envelope: `schemaVersion`, `command`, `ok`, `workspace`, `work`, discriminated `result`, `diagnostics`, and `nextActions`. This specification makes absent workspace/work context explicit as `null` and adds the typed `exitCode` field. This document does not introduce a competing `protocolVersion` envelope.

## Glossary and authority

- **Workspace**: the discovered project root containing the canonical `.sdd/` directory.
- **Work item**: a stable, immutable work ID plus mutable revision and a display slug. The slug is never identity.
- **Snapshot**: one immutable read of workspace inputs, policy, workflow definitions, artifacts, and implementation facts. Pure passes do not reopen files.
- **Outcome**: the single public result envelope rendered by both human and JSON views.
- **MutationPlan**: ordered effect descriptions with exact preconditions; a dry-run is a plan, not authorization.
- **Expected revision**: the workspace/work revision captured while planning and rechecked immediately before application.
- **Definition digest**: a digest over the exact versioned workflow, policy, schema, and validator definitions used by a decision.
- **Blob, semantic, derivation digest**: respectively bytes, typed meaning, and dependency-derived meaning. They are not interchangeable.
- **Policy authority**: organization policy from a trusted, versioned source; project/local files may narrow it but cannot replace a required organization policy.
- **Agent-neutral**: no model SDK, provider, prompt, or agent claim is trusted as engine authority.
- **DDP and ROP**: DDP means data-driven programming with data-oriented design (not Domain-Driven Design); ROP is an ordinary Go success/failure sequencing model, not a monad framework.

## Requirements

### REQ-001: Native Go executable skeleton

1. The `sdd` executable SHALL compile without CGO for every explicitly declared OS/architecture target.
2. Build metadata SHALL identify semantic version, source revision when available, and a reproducible build-input record. A current wall-clock timestamp SHALL NOT be injected into reproducible artifacts; an explicitly supplied fixed build time may be displayed.
3. `--version` and `--help` SHALL work from an empty, non-workspace directory without Node, Bun, Go, source files, network, or Git.
4. The composition root SHALL only parse/compose; workflow decisions SHALL be behind `Engine.Execute`.
5. Claims about support SHALL distinguish a native runtime smoke job from a cross-build job. A Linux CI runner SHALL not claim runtime support for every foreign binary.

### REQ-002: One public protocol contract

1. Offline schema references SHALL resolve to checked-in schemas; `$ref` targets, discriminators, required fields, and additional-property rules SHALL be validated. Draft-07 schemas use the exact `$schema` URI `http://json-schema.org/draft-07/schema#`, with `draft-07`, not `draft-7`.
2. The public envelope SHALL be:

   ```json
   {
     "schemaVersion": "1.0",
     "command": "status",
     "ok": true,
     "workspace": {"root": "/repo", "formatVersion": 1},
     "work": null,
     "result": {"kind": "status", "workItems": []},
     "diagnostics": [],
     "nextActions": [],
     "exitCode": 0
   }
   ```

   `result` is a discriminated object for successful commands and is `null` only for a failed command whose diagnostics explain the absence. `ok` is true exactly when `exitCode` is zero. The foundation protocol version is the exact major/minor string `"1.0"`. Published schemas are immutable: implementation patch releases preserve that contract; a command/result extension publishes a new minor schema and incompatible meaning requires a new major version. Consumers negotiate/check the supported version instead of assuming the closed `1.0` schema accepts later commands.
3. `diagnostics` and `nextActions` SHALL be arrays (possibly empty), never `null`. A next action SHALL contain a command name and structured argument/flag fields, not only a shell string.
4. `workspace` and `work` SHALL use the canonical product types and distinguish immutable IDs from mutable slugs and revisions.
5. The `Diagnostic` contract SHALL include `code`, `severity` (`info|warning|error|critical`), `message`, structured `source` (kind/path/field when known), optional range, remediation, and metadata. Category prefixes are catalog namespaces, not a second schema: foundation uses `DIAG_*`; guided workflows use registered `BUG_*`, `QUICK_SPEC_*`, and `WORKFLOW_*` codes.
6. Schemas SHALL have stable `$id` and schema version. Generated Go types or hand-written types SHALL be checked against the same schemas; no example may describe a shape the schemas reject.

### REQ-003: Engine.Execute seam

1. `internal/engine` SHALL expose the single supported seam `Execute(ctx context.Context, cmd Command) Outcome` (a constructed `Engine` may own dependencies; callers still cross this one seam).
2. Unknown commands and malformed payloads SHALL return an outcome with a stable usage diagnostic, never panic.
3. Commands SHALL follow the applicable stages of `decode -> normalize -> discover/snapshot -> decide -> plan -> authorize/apply (if in scope) -> verify -> outcome`. Version/help bypass workspace discovery; read-only commands bypass effects. Cancellation is honored before expensive work and effect boundaries.
4. Internal errors SHALL be translated through `errors.Is`/`errors.As` to stable categories and codes. Raw error text, paths containing secrets, environment values, tokens, and command output SHALL be redacted before rendering.
5. Signal handling SHALL derive a cancellable command context and preserve cancellation causes; handlers SHALL not replace it with `context.Background()`.
6. Human and JSON renderers SHALL consume the same `Outcome`. Renderer write errors SHALL be returned/translated, never ignored.

### REQ-004: Version command

1. `sdd --version` and `sdd version` SHALL produce the same information and exit `0` from any working directory.
2. JSON output SHALL be one schema-valid envelope whose result identifies version, source revision (or `unknown`), build-input/build-time metadata, Go version, and protocol schema version. `buildInputs` and `protocolSchemaVersion` are required result fields; the envelope's exact schema version is `"1.0"`.
3. Version discovery SHALL never require workspace, configuration, network, package installation, or Git at runtime.

### REQ-005: Doctor command

1. `doctor` SHALL report executable, version metadata, current-directory access, workspace discovery, canonical layout, and schema compatibility as structured checks.
2. No workspace is a workspace/configuration error: `doctor` SHALL return exit code `3`, not `0`. Incompatible or malformed workspace likewise returns `3`.
3. Doctor SHALL not mutate, invoke network/external tools, or treat a repository policy file as organization authority.
4. JSON and human output SHALL share the same diagnostics and next actions.

### REQ-006: Status and explicit selection

1. `status` SHALL be read-only. Zero active work items is a successful empty result; one may be selected implicitly; more than one without `--work` returns `DIAG_WORKSPACE_SELECTION_AMBIGUOUS`, candidate IDs and exit `3`. An explicit work ID returns that item.
2. Multiple unrelated active work items SHALL not block starting a new work item. A command that needs an existing work item SHALL require `--work <immutable-id>` or a single unambiguous candidate.
3. Selection SHALL follow explicit root rules: an invalid selected root is an error; the engine SHALL not silently fall back to a valid parent/ancestor root.
4. A missing workspace or invalid selected work ID SHALL return exit `3` with stable diagnostics.
5. Persisted paths SHALL be workspace-relative and normalized; transient display may use native separators.

### REQ-007: Start dry-run and inspectable plan

1. `start --dry-run "intent"` SHALL discover a valid workspace, choose/resolve a work identity and slug, compose the policy floor, and return a meaningful `MutationPlan` describing proposed creation of canonical files, IDs, revisions, and events. An empty operations list is not a useful dry-run.
2. The plan SHALL include `planId`, `workspaceId`, `workId`, `expectedRevision`, input/definition/policy/workflow digests, ordered non-empty operations, non-empty preconditions, an `authorizable` flag, and plan digest.
3. Dry-run SHALL never apply operations. A future/unsupported operation SHALL be marked non-authorizable and produce a diagnostic rather than pretending to be executable.
4. A later mutating start (owned by workspace-transactions) SHALL re-read and compare all preconditions immediately before application.

### REQ-008: Human and JSON rendering

1. Human output SHALL answer outcome, attention, and one next command. Machine output SHALL write exactly one envelope to stdout; diagnostics/progress/debug go to stderr.
2. JSON mode SHALL contain no ANSI, spinner, prose prefix, terminal-width formatting, or shell-only next action.
3. Every actionable `nextActions` entry SHALL use structured executable plus argument array/flags and include a reason and expected revision when relevant.
4. TTY, color, Unicode, path separators, and timestamp formatting SHALL affect presentation only; persisted/machine timestamps are UTC with offset.

### REQ-009: Stable exit codes

The CLI SHALL use the product taxonomy: `0` success; `1` validation/verification failure; `2` usage/malformed input; `3` workspace/config/schema error; `4` conflict/stale revision/claim; `5` required external tool failure; `6` policy/permission denial; `130` interrupted after cleanup. A diagnostic category SHALL not be collapsed into `1` merely for convenience.

### REQ-010: Stable diagnostic catalog

1. Core codes SHALL include `DIAG_SUCCESS`, `DIAG_USAGE_ERROR`, `DIAG_COMMAND_UNKNOWN`, `DIAG_WORKSPACE_NOT_FOUND`, `DIAG_WORKSPACE_VERSION_INCOMPATIBLE`, `DIAG_WORKSPACE_SELECTION_AMBIGUOUS`, `DIAG_WORK_ITEM_NOT_FOUND`, `DIAG_STALE_REVISION`, `DIAG_PLAN_NOT_AUTHORIZABLE`, and `DIAG_INTERRUPTED`.
2. Workflow-specific codes are registered by the guided spec and retain their category namespace; consumers match code and structured metadata, never message text.
3. Every code documents severity, exit code, source, remediation, and whether it is retryable.

### REQ-011: CLI parsing and composition

Global flags retain one meaning. Positional input is limited to the one obvious primary object. Non-interactive execution never prompts; consent failure names the required flag or preparatory command. No command may launch an editor, browser, network call, agent, Git commit, PR, merge, or push unless a later explicit contract adds it.

### REQ-012: Migration parity fixtures

Frozen TypeScript fixtures MAY compare only behaviors that actually exist in the legacy CLI (currently version and status, subject to inventory). Fixtures SHALL state command, stdin, stdout, stderr, exit code, schema version, and intentional differences. Doctor and `start --dry-run` are new-contract fixtures unless a legacy implementation is found; they must not be described as parity. The harness is migration evidence only, not a second authority, and is removed when parity is proven.

### REQ-013: Empty-directory execution

`--version` and `--help` succeed outside a workspace. `doctor`, `status`, and `start --dry-run` report `DIAG_WORKSPACE_NOT_FOUND` and exit `3`; no command creates a workspace implicitly.

### REQ-014: No side effects for read-only commands

Version, help, doctor, status, and start dry-run SHALL not create, alter, delete, commit, or lock persistent workspace files. Tests compare file bytes, modes, modification times, and directory entries before/after. Filesystem-maintained read access times are excluded; the CLI does not rewrite timestamps to mask changes.

### REQ-015: Cancellation and interruption

Signal cancellation SHALL stop observation and planning at safe points, propagate the cause, and leave no partial effect. Read-only foundation tests cover this seam; descendant cleanup belongs to the process-supervision prerequisite. `130` is used only when cleanup semantics are known.

### REQ-016: Build and smoke signals

1. CI SHALL run formatting, static analysis, unit tests, schema validation, JSON fixture checks, and native runtime smoke jobs for each declared target. Cross-build jobs are reported separately from runtime smoke.
2. Native smoke tests use an absolute binary path after changing to a temporary empty directory and SHALL fail loudly; no `|| true` masks failures.
3. Reproducible build checks use fixed declared inputs (including a fixed build metadata value where needed), not current wall time.

### REQ-017: Workspace version and canonical layout

1. The foundation SHALL recognize the canonical product layout `.sdd/work/<slug>/work.yaml`; Kiro filenames such as `work.md`, `work-items/`, or `work-item.yaml` are adapters/fixtures, never engine authority.
2. `work.yaml` SHALL hold immutable `workId`, mutable `slug`, `revision`, journey/composition reference, and status. Identity remains stable across slug changes and escalation.
3. Foundation owns minimal read-only discovery, canonical-layout decoding, workspace version checks, and explicit selection. Future `sdd-workspace-transactions` may extend this reader with locking, journals, and durable mutation; foundation does not depend on that later spec to implement its read-only commands.

### REQ-018: Invalid and ambiguous selection

An explicitly requested invalid root/work ID is an error with no parent fallback. Selection of a new work item is independent from unrelated active work items. Any policy, workflow, validator, or completion prerequisite not accepted by its named spec is a release gate, not an invented local implementation.

## Cross-cutting contract: mutation, policy, identity, and digests

- A plan captures `workspaceId`, `workId`, `revisionScope`, `expectedRevision`, `inputDigest`, `policyDigest`, `workflowDefinitionDigest`, `validatorDefinitionDigest`, and `digest` (the plan digest). `revisionScope` is `workspace` for new-work creation and `work` for existing-work changes. Git commit/tree/status is optional provenance and never substitutes for work revision or dirty-snapshot identity.
- Recheck the complete precondition set immediately before applying. A transaction with the same idempotency key and identical content returns the original result; the same key with changed content returns a conflict. Crash/timeout with an uncertain commit returns an explicit recovery state and the prepared transaction, never a guessed success.
- Policy evaluation loads required organization policy from a trusted versioned source with explicit precedence. Missing, expired, revoked, unverifiable, or stale required policy fails closed. Local project policy may narrow requirements. A same-user local CLI/filesystem is not tamper-resistant; trusted CI or another declared trust boundary is required for claims of origin, approval, or exception authority.
- A policy exception names an identified actor, grant scope, authorization evidence, policy snapshot digest, issued/expiry times, revocation state, exact work/snapshot scope, and preauthorization checks. `owner: "alice", approved: true` is not authorization.
- `Clock` and `ID` are injected dependencies. Deterministic tests supply fixed values; no permanent identity is derived solely from intent text.
- Canonicalization is typed and versioned. Preserve meaningful Markdown whitespace, code blocks, ordering, and unknown fields conservatively. Distinguish blob, semantic, and derivation digests and invalidate derived artifacts when unknown meaning may have changed.
- Error-category precedence is deterministic: cancellation/interruption, usage, workspace/schema, conflict, policy, external-tool, validation, then internal defect. The typed `exitCode` in the envelope is authoritative; human messages never select a category.

## Dependencies and non-goals

Foundation is independently implementable for its native executable, protocol, read-only workspace snapshot, selection, and dry-run plan. Future `sdd-workspace-transactions`, `sdd-workflow-graph`, `sdd-artifact-validation`, and `sdd-completion-lifecycle` extend reserved seams; durable multi-file mutation, workflow execution, full artifact validation, and completion release claims are future integration gates, not foundation dependencies. Automatic Git commits/PRs/merges, Bun comparisons, arbitrary performance claims, process supervision, SQLite, and embedded model execution are not part of this spec.

## Required foundation fixtures

Foundation fixtures cover invalid explicit root with valid parent; multiple unrelated active work items; empty-op dry-run rejection; schema `$ref` and negative fixtures; cancellation and renderer-write failure; redaction; no-side-effect reads; and native-vs-cross-build smoke behavior. Concurrent stale writers, uncertain commits, dirty post-verification code, exception revocation, policy removal, changed workflow definitions, unsafe reproduction proxies, and forged workflow validator claims are named compatibility scenarios for the future prerequisite specs, not foundation tests requiring those runtimes.

## Open questions retained for implementation planning

- Which exact OS/architecture pairs receive native runtime support first, as opposed to cross-build support?
- How should environment variables interact with command flags while preserving deterministic precedence?
- Which filesystems are supported, warned, or blocked for future durable transactions?
- What minimum Git version, if any, is required when optional Git observation is introduced?
- Are FIPS or other regulated cryptographic requirements applicable to future policy/receipt signing?
- Is distribution internal-only or public, which channels are approved, and which native runners are available? Until decided, CI reports tested targets without claiming an approved support matrix.
- Which trusted policy/actor sources and retention/redaction rules will company integration use? Foundation fixtures must not invent these authorities.

## References

- `docs/kiro-input/sdd-cli-spec-brief.md`
- `docs/architecture/go-project-structure.md`
- `docs/architecture/guided-workflows.md`
- `.kiro/steering/cli-ux.md`
- `.kiro/steering/go-engineering.md`
