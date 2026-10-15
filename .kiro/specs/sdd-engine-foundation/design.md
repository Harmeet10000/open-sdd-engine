# Design Document: sdd-engine-foundation

## Overview

The foundation is a native Go executable with one deep application seam. DDP here means data-driven programming with data-oriented design, not Domain-Driven Design; ROP is expressed with idiomatic Go control flow and explicit errors, not a generic monad framework.

```text
CLI parsing -> Engine.Execute(ctx, Command) Outcome -> human/json renderer
```

The engine performs deterministic transformations over immutable snapshots. Platform adapters report facts or apply an already-authorized plan; they do not choose workflow policy. This keeps the core agent-neutral and leaves durable transactions, workflow graphs, typed artifacts, and completion gates to their named prerequisite specs.

## Architecture and package seams

Use capability-oriented internal packages: `engine`, `workspace`, `workflow`, `artifact`, `validation`, `completion`, `contextpack`, and `presentation`. The engine coordinates one use case and is not a service locator. A package exposes the smallest deep interface callers need; interfaces exist only at real effect seams (workspace transaction, process supervision, Git observation, clock/ID).

The dependency direction is:

```text
presentation -> engine -> workspace/workflow/artifact/validation/completion
                               |
                               v
                       effect plans and ports
                               |
                 filesystem/process/Git/clock adapters
```

Artifacts parse and normalize values. Validation consumes normalized immutable values. Engine coordinates snapshots and decisions. No package in an evaluation pass reopens files or relies on mutable global state. The planned artifact/validation package split must not form a cyclic import: artifact owns parsing, typed normalization, identity, and digests; validation imports artifact and emits diagnostics; engine imports both and coordinates.

## Public protocol

The product plan's envelope is normative. A command result is a discriminated type selected by `command`; `diagnostics` and `nextActions` are always arrays. Draft-07 schema documents use `$schema: http://json-schema.org/draft-07/schema#` and resolve all local `$ref`s offline. A conceptual envelope (not a JSON fixture) is:

```text
schemaVersion: 1.0
command: status
ok: true
workspace: optional WorkspaceRef
work: optional WorkRef
result: StatusResult(kind=status, workItems=[])
diagnostics: []
nextActions: []
```

The normative bundle below defines the foundation contract; implementation will materialize it under `protocol/schema/`. No generated schema files are claimed to exist yet. Every schema example must be parsed and validated offline with a full draft-07-capable validator, including conditional keywords and `$ref`; a hand-written jq check is not schema validation. The published `1.0` contract is immutable. Later command/result variants require a new minor schema with explicit supported-version handling; the closed foundation schema does not silently accept them.

### Normative protocol schema bundle

The following is valid draft-07 JSON. Its `definitions.Command`, `definitions.Outcome`, `definitions.Diagnostic`, `definitions.NextAction`, and `definitions.MutationPlan` are the concrete schemas; generated per-type files must preserve these definitions and IDs.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://sdd.dev/schema/foundation-1.0.json",
  "$ref": "#/definitions/Outcome",
  "definitions": {
    "WorkspaceRef": {
      "type": "object",
      "additionalProperties": false,
      "required": ["root", "formatVersion"],
      "properties": {
        "root": {"type": "string", "minLength": 1},
        "formatVersion": {"type": "integer", "minimum": 1}
      }
    },
    "WorkRef": {
      "type": "object",
      "additionalProperties": false,
      "required": ["id", "slug", "revision"],
      "properties": {
        "id": {"type": "string", "minLength": 1},
        "slug": {"type": "string", "minLength": 1},
        "workflow": {"type": "string", "minLength": 1},
        "revision": {"type": "integer", "minimum": 0}
      }
    },
    "SourceRange": {
      "type": "object",
      "additionalProperties": false,
      "required": ["path", "start"],
      "properties": {
        "path": {"type": "string", "minLength": 1},
        "start": {"$ref": "#/definitions/Position"},
        "end": {"$ref": "#/definitions/Position"}
      }
    },
    "Position": {
      "type": "object",
      "additionalProperties": false,
      "required": ["line", "column", "unit"],
      "properties": {
        "line": {"type": "integer", "minimum": 1},
        "column": {"type": "integer", "minimum": 1},
        "unit": {"enum": ["utf8-byte", "unicode-scalar"]}
      }
    },
    "DiagnosticSource": {
      "type": "object",
      "additionalProperties": false,
      "required": ["kind"],
      "properties": {
        "kind": {"type": "string", "minLength": 1},
        "path": {"type": "string"},
        "field": {"type": "string"},
        "artifactId": {"type": "string"},
        "range": {"$ref": "#/definitions/SourceRange"}
      }
    },
    "Diagnostic": {
      "type": "object",
      "additionalProperties": false,
      "required": ["code", "severity", "message", "source", "retryable"],
      "properties": {
        "code": {"type": "string", "pattern": "^[A-Z][A-Z0-9_]*$"},
        "severity": {"enum": ["info", "warning", "error", "critical"]},
        "message": {"type": "string", "minLength": 1},
        "source": {"$ref": "#/definitions/DiagnosticSource"},
        "remediation": {"type": "string"},
        "metadata": {"type": "object", "additionalProperties": true},
        "retryable": {"type": "boolean"}
      }
    },
    "NextAction": {
      "type": "object",
      "additionalProperties": false,
      "required": ["command", "args", "flags", "reason"],
      "properties": {
        "command": {"type": "string", "minLength": 1},
        "args": {"type": "array", "items": {"type": "string"}},
        "flags": {"type": "object", "additionalProperties": {"type": ["string", "boolean", "integer", "number"]}},
        "reason": {"type": "string", "minLength": 1},
        "expectedRevision": {"type": "integer", "minimum": 0}
      }
    },
    "StatusCommand": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type"],
      "properties": {"type": {"const": "status"}, "workId": {"type": "string"}, "root": {"type": "string"}}
    },
    "VersionCommand": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type"],
      "properties": {"type": {"const": "version"}}
    },
    "DoctorCommand": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type"],
      "properties": {"type": {"const": "doctor"}, "root": {"type": "string"}}
    },
    "StartCommand": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type", "dryRun", "intent"],
      "properties": {"type": {"const": "start"}, "dryRun": {"const": true}, "intent": {"type": "string", "minLength": 1}, "workflow": {"type": "string"}, "root": {"type": "string"}}
    },
    "Command": {
      "oneOf": [
        {"$ref": "#/definitions/StatusCommand"},
        {"$ref": "#/definitions/VersionCommand"},
        {"$ref": "#/definitions/DoctorCommand"},
        {"$ref": "#/definitions/StartCommand"}
      ]
    },
    "StatusResult": {
      "type": "object",
      "additionalProperties": false,
      "required": ["kind", "workItems"],
      "properties": {"kind": {"const": "status"}, "workItems": {"type": "array", "items": {"$ref": "#/definitions/WorkRef"}}}
    },
    "VersionResult": {
      "type": "object",
      "additionalProperties": false,
      "required": ["kind", "version", "commit", "buildTime", "goVersion", "buildInputs", "protocolSchemaVersion"],
      "properties": {"kind": {"const": "version"}, "version": {"type": "string"}, "commit": {"type": "string"}, "buildTime": {"type": "string"}, "goVersion": {"type": "string"}, "buildInputs": {"type": "object", "additionalProperties": {"type": "string"}}, "protocolSchemaVersion": {"const": "1.0"}}
    },
    "DoctorResult": {
      "type": "object",
      "additionalProperties": false,
      "required": ["kind", "checks"],
      "properties": {"kind": {"const": "doctor"}, "checks": {"type": "array", "items": {"type": "object", "additionalProperties": true}}}
    },
    "PlanOperation": {
      "type": "object",
      "additionalProperties": false,
      "required": ["kind", "path", "expectedRevision", "authorizable"],
      "properties": {
        "kind": {"enum": ["create-file", "update-file", "delete-file"]},
        "path": {"type": "string", "minLength": 1},
        "content": {"type": "string"},
        "contentDigest": {"type": "string", "minLength": 1},
        "expectedDigest": {"type": "string", "minLength": 1},
        "expectedRevision": {"type": "integer", "minimum": 0},
        "authorizable": {"type": "boolean"}
      },
      "allOf": [
        {"if": {"properties": {"kind": {"enum": ["create-file", "update-file"]}}}, "then": {"required": ["content", "contentDigest"]}},
        {"if": {"properties": {"kind": {"enum": ["update-file", "delete-file"]}}}, "then": {"required": ["expectedDigest"]}}
      ]
    },
    "PlanPrecondition": {
      "type": "object",
      "additionalProperties": false,
      "required": ["kind", "path", "expectedRevision"],
      "properties": {
        "kind": {"enum": ["revision", "digest", "absent"]},
        "path": {"type": "string", "minLength": 1},
        "expectedRevision": {"type": "integer", "minimum": 0},
        "expectedDigest": {"type": "string"}
      },
      "allOf": [
        {"if": {"properties": {"kind": {"const": "digest"}}}, "then": {"required": ["expectedDigest"]}}
      ]
    },
    "MutationPlan": {
      "type": "object",
      "additionalProperties": false,
      "required": ["planId", "workspaceId", "workId", "revisionScope", "expectedRevision", "inputDigest", "policyDigest", "workflowDefinitionDigest", "validatorDefinitionDigest", "operations", "preconditions", "digest", "authorizable"],
      "properties": {
        "planId": {"type": "string", "minLength": 1},
        "workspaceId": {"type": "string", "minLength": 1},
        "workId": {"type": "string", "minLength": 1},
        "revisionScope": {"enum": ["workspace", "work"]},
        "expectedRevision": {"type": "integer", "minimum": 0},
        "inputDigest": {"type": "string", "minLength": 1},
        "policyDigest": {"type": "string", "minLength": 1},
        "workflowDefinitionDigest": {"type": "string", "minLength": 1},
        "validatorDefinitionDigest": {"type": "string", "minLength": 1},
        "operations": {"type": "array", "minItems": 1, "items": {"$ref": "#/definitions/PlanOperation"}},
        "preconditions": {"type": "array", "minItems": 1, "items": {"$ref": "#/definitions/PlanPrecondition"}},
        "digest": {"type": "string", "minLength": 1},
        "idempotencyKey": {"type": "string"},
        "authorizable": {"type": "boolean"}
      }
    },
    "StartDryRunResult": {
      "type": "object",
      "additionalProperties": false,
      "required": ["kind", "plan"],
      "properties": {"kind": {"const": "start-dry-run"}, "plan": {"$ref": "#/definitions/MutationPlan"}}
    },
    "Outcome": {
      "type": "object",
      "additionalProperties": false,
      "required": ["schemaVersion", "command", "ok", "workspace", "work", "result", "diagnostics", "nextActions", "exitCode"],
      "properties": {
        "schemaVersion": {"const": "1.0"},
        "command": {"enum": ["status", "version", "doctor", "start", "unknown"]},
        "requestedCommand": {"type": "string"},
        "ok": {"type": "boolean"},
        "exitCode": {"enum": [0, 1, 2, 3, 4, 5, 6, 130]},
        "workspace": {"anyOf": [{"$ref": "#/definitions/WorkspaceRef"}, {"type": "null"}]},
        "work": {"anyOf": [{"$ref": "#/definitions/WorkRef"}, {"type": "null"}]},
        "result": {"anyOf": [{"$ref": "#/definitions/StatusResult"}, {"$ref": "#/definitions/VersionResult"}, {"$ref": "#/definitions/DoctorResult"}, {"$ref": "#/definitions/StartDryRunResult"}, {"type": "null"}]},
        "diagnostics": {"type": "array", "items": {"$ref": "#/definitions/Diagnostic"}},
        "nextActions": {"type": "array", "items": {"$ref": "#/definitions/NextAction"}}
      },
      "allOf": [
        {"if": {"properties": {"ok": {"const": true}}}, "then": {"properties": {"exitCode": {"const": 0}}, "not": {"properties": {"result": {"type": "null"}}}}, "else": {"properties": {"exitCode": {"enum": [1, 2, 3, 4, 5, 6, 130]}, "diagnostics": {"minItems": 1}}}},
        {"if": {"properties": {"command": {"const": "unknown"}}}, "then": {"required": ["requestedCommand"], "properties": {"ok": {"const": false}, "exitCode": {"const": 2}, "result": {"type": "null"}}}},
        {"if": {"properties": {"ok": {"const": true}, "command": {"enum": ["status", "doctor", "start"]}}}, "then": {"properties": {"workspace": {"$ref": "#/definitions/WorkspaceRef"}}}},
        {"if": {"properties": {"command": {"const": "status"}}}, "then": {"properties": {"result": {"anyOf": [{"$ref": "#/definitions/StatusResult"}, {"type": "null"}]}}}},
        {"if": {"properties": {"command": {"const": "version"}}}, "then": {"properties": {"result": {"anyOf": [{"$ref": "#/definitions/VersionResult"}, {"type": "null"}]}}}},
        {"if": {"properties": {"command": {"const": "doctor"}}}, "then": {"properties": {"result": {"anyOf": [{"$ref": "#/definitions/DoctorResult"}, {"type": "null"}]}}}},
        {"if": {"properties": {"command": {"const": "start"}}}, "then": {"properties": {"result": {"anyOf": [{"$ref": "#/definitions/StartDryRunResult"}, {"type": "null"}]}}}}
      ]
    }
  }
}
```

Examples are checked against the bundle. A successful status uses a non-null `StatusResult`:

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

A malformed/unknown command is representable without trusting raw input as a command:

```json
{
  "schemaVersion": "1.0",
  "command": "unknown",
  "requestedCommand": "bogus",
  "ok": false,
  "workspace": null,
  "work": null,
  "result": null,
  "diagnostics": [{"code": "DIAG_COMMAND_UNKNOWN", "severity": "error", "message": "Unknown command", "source": {"kind": "command"}, "retryable": false}],
  "nextActions": [],
  "exitCode": 2
}
```

A dry-run result uses the public `start` command with `dryRun: true` at input and a discriminated `start-dry-run` result:

```json
{
  "schemaVersion": "1.0",
  "command": "start",
  "ok": true,
  "workspace": {"root": "/repo", "formatVersion": 1},
  "work": null,
  "result": {"kind": "start-dry-run", "plan": {"planId": "plan-1", "workspaceId": "ws-1", "workId": "work-1", "revisionScope": "workspace", "expectedRevision": 4, "inputDigest": "sha256:input", "policyDigest": "sha256:policy", "workflowDefinitionDigest": "sha256:workflow", "validatorDefinitionDigest": "sha256:validators", "operations": [{"kind": "create-file", "path": ".sdd/work/new/work.yaml", "expectedRevision": 4, "content": "workId: work-1\n", "contentDigest": "sha256:content", "authorizable": false}], "preconditions": [{"kind": "absent", "path": ".sdd/work/new/work.yaml", "expectedRevision": 4}], "digest": "sha256:plan", "authorizable": false}},
  "diagnostics": [{"code": "DIAG_PLAN_NOT_AUTHORIZABLE", "severity": "warning", "message": "Preview only; durable application is not implemented in foundation", "source": {"kind": "plan"}, "retryable": false}],
  "nextActions": [],
  "exitCode": 0
}
```

The no-workspace status outcome uses `ok: false`, `workspace: null`, `result: null`, and `DIAG_WORKSPACE_NOT_FOUND` with exit code `3`; a missing workspace cannot appear as a successful status. The dry-run example illustrates wire shape with symbolic digest values and abbreviated work content; it is not an executable fixture. Runtime conformance fixtures must contain complete records and recomputed SHA-256 digests. Foundation plans are always preview-only (`authorizable: false`); a successful preview does not claim application is available.

### Core values

```go
type Command interface{ commandMarker() }
type Outcome struct {
    SchemaVersion string         `json:"schemaVersion"`
    Command       string         `json:"command"`
    OK            bool           `json:"ok"`
    Workspace     *WorkspaceRef  `json:"workspace"`
    Work          *WorkRef       `json:"work"`
    Result        *Result        `json:"result"` // closed tagged value, not a pointer to an interface
    Diagnostics   []Diagnostic   `json:"diagnostics"`
    NextActions   []NextAction   `json:"nextActions"`
    ExitCode      int            `json:"exitCode"`
    RequestedCommand string     `json:"requestedCommand,omitempty"`
}
```

`Diagnostic` has `Code`, `Severity` (`info`, `warning`, `error`, `critical`), `Message`, structured `Source`, optional range, remediation, and metadata. The source is an object (`kind`, workspace-relative path, field, and optional artifact ID), not a string. A `NextAction` contains `command`, `args` (ordered array), `flags` (typed map), `reason`, and optional expected revision; it is not a shell command string.

`result` is a command-specific discriminated object. For example, `StatusResult` has `kind: status` and `workItems`; `StartDryRunResult` has `kind: start-dry-run` and `plan`. Nullability is declared once in the schema and represented identically in Go, JSON, and examples. Empty collections encode no findings; they are never null.

## Workspace and identity

Discovery returns an immutable `Snapshot` containing canonical root, workspace ID/version, ordered work records, normalized policy, workflow/validator definitions, artifact metadata, implementation observation, and input digests. The canonical product layout is `.sdd/work/<slug>/work.yaml`. The record contains an immutable `workId`, mutable display `slug`, monotonic `revision`, journey/composition reference, and lifecycle state. Kiro's `work.md` and `work-items/work-item.yaml` may be import/export adapter files only.

Selection is explicit. A command requiring an existing work item accepts an immutable ID or one unambiguous candidate. An invalid explicitly selected root is an error even if a valid parent exists; there is no ancestor fallback. Multiple unrelated active items do not prevent creating a new work item. Stable IDs are generated by an injected ID source and are not hashes of intent alone; fixed clock/ID adapters make tests reproducible.

Foundation reads `.sdd/workspace.yaml` with required `formatVersion: 1`, immutable `workspaceId`, and non-negative integer `revision`; each `work.yaml` contains `schemaVersion: 1`, `workId`, `slug`, non-negative integer `revision`, `status`, and a composition reference. Unknown format versions, duplicate IDs, malformed records and unknown critical fields fail with exit `3`. Do not compare integer format versions as permissive semver ranges. Without `--root`, discover the nearest `.sdd` marker walking upward; an invalid nearest marker is an error, not permission to skip to a higher project. With `--root`, inspect that root only.

The snapshot owns stable schema references and source coordinates rather than leaving them to renderers:

```text
Snapshot {
  workspace: WorkspaceRef
  schemaRefs: [{name, id, version}]
  workItems: [WorkRef]
  inputs: [{path, blobDigest, semanticDigest, sourceRange, nullability}]
}
```

Every decoded record carries its schema reference, version, and structured source range; range units are explicit (`utf8-byte` or `unicode-scalar`). Optional values have declared nullability in the schema and Go type rather than relying on absent-vs-empty conventions.

## Snapshot and digest pipeline

The only read path is:

```text
bytes/files -> decode -> typed normalize -> immutable Snapshot -> pure passes
```

Normalize once, preserve meaningful Markdown whitespace, code blocks, ordering, and unknown fields, and conservatively invalidate when unknown meaning changes. Compute three explicitly named digests:

- `blobDigest`: exact bytes for audit and conflict detection;
- `semanticDigest`: versioned typed meaning after safe canonicalization;
- `derivationDigest`: semantic digest plus the IDs/digests of upstream definitions and inputs.

`workflowDefinitionDigest`, `policyDigest`, and `validatorDefinitionDigest` bind the exact versioned content used by a decision, not just names. Git commit/tree/status is optional provenance and cannot stand in for work revision or a dirty snapshot digest.

## Engine.Execute flow and errors

Every command follows this visible ROP-style flow:

```text
decode -> normalize -> discover/snapshot -> decide -> plan
       -> authorize/apply (only for an owned mutation) -> verify -> outcome
```

Read-only foundation commands stop before authorization/application. A constructor wires concrete adapters and fakes; callers cross only `Engine.Execute`. The method checks `ctx.Err()` before expensive passes and before any irreversible boundary. Signal handling derives the context from the process root and preserves cancellation causes; no handler uses `context.Background()`.

Error translation uses `errors.Is`/`errors.As` with explicit precedence: cancellation/interruption, usage, workspace/schema, conflict, policy, external tool, validation, then internal defect. Internal errors are logged only through redacted debug paths and mapped to stable diagnostics. Renderer write errors are propagated into an operational outcome; they are not discarded. Human and JSON renderers receive the same outcome and differ only in presentation.

If stdout fails after a partial JSON write, do not append a second envelope or claim complete JSON. Report a redacted transport diagnostic on stderr where possible and exit nonzero. Broken output transport is an explicit exception to the complete-envelope guarantee, never successful verification.

## Commands

### Version and doctor

`version` reads build metadata only. `doctor` checks executable metadata, current-directory access, workspace discovery, canonical layout, schema compatibility, and policy source status without network, external tools, or mutation. No workspace is a configuration/workspace error with exit `3`; it is not a passing health check.

### Status

Status reads one snapshot and renders zero, one, or selected work records. It does not lock, repair, update projections, or infer a different root after a failed explicit selection. A stale workspace can be reported as stale with a read-only remediation.

### Start dry-run

Dry-run resolves a work ID (from the injected ID source), safe slug, journey/policy floor, and artifact profile, then produces an inspectable plan. The plan contains meaningful proposed operations such as creating `.sdd/work/<slug>/work.yaml`, initial journal event, and initial composition record; each operation has path, precondition, and expected revision. It does not write. A future stub is explicitly `non-authorizable` and cannot be passed to an apply adapter.

Plan fields are:

```text
planId
workspaceId
workId
revisionScope (workspace for creation, work for existing-work changes)
expectedRevision
inputDigest
policyDigest
workflowDefinitionDigest
validatorDefinitionDigest
operations[]
preconditions[]
digest
authorizable
idempotencyKey (when a mutation command supplies one)
```

The eventual workspace transaction re-reads and rechecks every field immediately before applying. Same key plus identical content returns the recorded result; same key plus changed content is a conflict. A crash or timeout after prepare but before an observable commit returns `recovery-required` with the prepared transaction and never guesses success.

The plan digest binds every decision-bearing field and ordered effect, including targets, revision scope, preconditions, input/definition digests and proposed content, excluding the digest field itself. The versioned canonical encoder sorts object keys, preserves array order and Unicode content, rejects non-finite numbers, and uses UTF-8 with type/schema domain separation. Any changed target, policy, content or precondition requires a new authorization. Reject absolute paths, traversal, symlink escapes and platform-equivalent collisions before evaluation and recheck at the effect seam. Creates require absence; updates/deletes require prior blob digests; operation revisions use the plan's declared revision scope. An `authorizable` Boolean describes capability availability, not an approval grant; approval is separate evidence bound to the digest.

## Policy and authorization

Policy precedence is explicit: required organization policy from a trusted, versioned distribution; project policy may narrow it; work-local metadata may record rationale but cannot lower it. Missing, stale, expired, revoked, or unverifiable required policy fails closed. A same-user CLI and mutable local files are not a tamper-resistant trust boundary. Claims of approval or exception authority require trusted CI or another declared boundary.

An exception is a typed record, not owner text and a Boolean. It includes identified actor, grant scope, authorization evidence, policy snapshot digest, work/snapshot scope, issued/expiry timestamps, revocation state, and preauthorization checks. The authorization decision binds the exact plan digest; a newly computed plan cannot inherit an earlier approval.

## Validation architecture

Authoritative validators are selected from trusted policy and the versioned validator definition set. For a candidate, the engine derives the mandatory validator IDs, executes those validators, records `validatorsExecuted` with definition digests and results, and rejects missing, stale, cancelled, or unverifiable evidence. An agent-provided `validator.passed` is provenance only; it cannot turn an empty validator list into a pass. Field validators and execution records use one canonical name (`validatorsExecuted`).

Semantic classification is not a deterministic root-cause oracle. Human/agent assessments carry evidence references, provenance, confidence, and uncertainty. Deterministic checks cover structure, policy, IDs, digests, and traceability; heuristics are advisory and cannot satisfy a required gate alone.

## Build, smoke, and test design

Use native matrix jobs to build and run binaries on supported OS/architectures; cross-builds validate compilation only. Smoke tests pass an absolute binary path after entering a temporary directory and fail on errors. Build metadata is fixed for reproducibility. Test drivers live in the package under test (`*_test.go`, external `package_test` where useful); `testdata/` contains only fixtures, never executable test drivers.

Schema conformance uses a full draft-07 validator with `$ref` resolution and invalid fixtures. Foundation interface tests use deterministic filesystem, clock, ID, and policy fakes. Required foundation fixtures cover invalid explicit root, multiple active unrelated items, non-empty dry-run plans, cancellation, renderer failure, redaction, no-side-effect reads, and native-versus-cross-build reporting. Stale writers, uncertain commits, dirty post-verification candidates, revoked exceptions, changed workflow definitions, unsafe proxies, and forged workflow validator claims are recorded as future prerequisite acceptance scenarios rather than executed by this independent foundation slice.

## Dependencies and release gates

The foundation is independently implementable. Future `sdd-workspace-transactions`, `sdd-workflow-graph`, `sdd-artifact-validation`, and `sdd-completion-lifecycle` extend reserved seams for durable mutation, graph execution, rich artifacts, and completion. This design does not invent those internals or claim their integration; their conformance suites become later release gates.

## Deliberate non-goals

No automatic Git commit, branch, push, PR, merge, Bun comparison, arbitrary latency claim, network authority, embedded LLM, process supervisor, SQLite projection, or second protocol envelope is designed here. Git observation may later appear as optional provenance, never as the work revision.

## References

- `docs/kiro-input/sdd-cli-spec-brief.md`
- `docs/architecture/go-project-structure.md`
- `.kiro/steering/go-engineering.md`
- `.kiro/steering/cli-ux.md`
