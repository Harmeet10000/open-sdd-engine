# Foundation conformance review and extension seam record (FND-016)

- Status: Accepted (implementation review; not a product release claim)
- Date: 2026-10-17
- Scope: `sdd-engine-foundation` implementation, FND-001..FND-016

## 1. Package direction audit

Reviewed dependency direction, verified with
`go list -f '{{ join .Imports "\n" }}'` per package:

```text
cmd/sdd            -> internal/cli, internal/engine
internal/cli       -> internal/engine, internal/presentation/*
internal/presentation/render -> internal/engine
internal/presentation/help   -> (stdlib only)
internal/testkit   -> internal/engine, internal/workspace  (tests only)
internal/engine    -> internal/workspace                   (Snapshotter view only)
internal/workspace -> (stdlib only)
```

- Only `internal/cli` depends on Cobra; `internal/engine` has no
  `cobra`/`testkit`/`cli`/`render` dependencies.
- `internal/testkit` is imported exclusively from `*_test.go` files.
- `os.Exit` appears only in `cmd/sdd/main.go` (getwd failure, final code)
  and the pre-existing commitlint tool main.
- `context.Background()` in non-test code appears only at process entry
  (`cmd/sdd`, commitlint tool); every port and engine path propagates the
  caller context.
- Contradiction searches over the implementation: no `protocolVersion`
  envelope, no `"hint"` wire severity, no `"1.1"` schema claim, no second
  outcome renderer, no workspace mutation outside future transaction
  ownership.

## 2. Requirement traceability

Each requirement traces to executable fixtures, or to a labelled future
integration contract. `T` denotes a Go test, `F` a checked-in fixture.

| REQ | Status | Evidence |
|---|---|---|
| REQ-001 native executable | Implemented | `cmd/sdd/main.go`; `make build` with fixed ldflags; reproducible-build CI job; `internal/cli` TTY/redirected fixtures |
| REQ-002 protocol contract | Implemented | `protocol/schema/foundation-1.0.json`; `protocol/schema_test.go`; `protocol/engine_conformance_test.go` (every engine outcome validates); FND-001 inventory test |
| REQ-003 Engine.Execute seam | Partial | Deep seam, cancellation, external identity failure and non-panicking internal fallback are implemented; real conflict/validation producers and the coherent raw-snapshot input remain prerequisite work |
| REQ-004 version | Implemented | Version fixtures in `engine_test.go`, `engine_conformance_test.go`, `cli_test.go`; contact-free proof; FND-013 legacy `version.json` parity |
| REQ-005 doctor | Implemented | Five checks; malformed/incompatible/policy matrices; partial-check failures; workspace/configuration exit 3 and required-policy exit 6 |
| REQ-006 status/selection | Implemented | Zero/one/many/explicit matrices; ID-only selection; `DIAG_WORKSPACE_SELECTION_AMBIGUOUS` with candidates |
| REQ-007 dry-run plan | Implemented | Verified non-empty plan; digest/revision preconditions; non-authorizable warning; determinism proof; no-side-effect proof |
| REQ-008 rendering | Implemented | One-`Input` human/JSON seam; stdout/stderr separation; no ANSI; `--json`/`--output` resolution; TUI honestly rejected |
| REQ-009 exits | Implemented | Taxonomy `0,1,2,3,4,5,6,130` in `engine/value.go`; CLI exit mapping fixtures incl. 130 on cancellation |
| REQ-010 diagnostics | Implemented | Catalog-driven codes/severities/exits; inventory test forbids unregistered codes |
| REQ-011 CLI parsing | Implemented | Cobra command tree; non-interactive proof; concise usage failures; `start` gated on `--dry-run` |
| REQ-012 parity | Implemented | `internal/cli/testdata/legacy-typescript/` (6 frozen fixtures) + `parity_test.go`; doctor/dry-run excluded from parity by corpus allowlist |
| REQ-013 empty directory | Implemented | Version/help succeed; doctor/status/dry-run report `DIAG_WORKSPACE_NOT_FOUND`/3 |
| REQ-014 no side effects | Implemented | Native-tree proofs compare path inventory, bytes, modes, sizes, and modification times for status, doctor, and dry-run |
| REQ-015 cancellation | Implemented | `ctx.Err()` checks before expensive work and effect boundaries; 130 with `DIAG_INTERRUPTED`; blocked-read propagation in fakes |
| REQ-016 build/smoke | Implemented | `make build`; smoke/cross-build/reproducible CI jobs; absolute-path loud-failure smoke; native vs cross reported separately |
| REQ-017 layout/version | Implemented | Strict decode; exact integer versions; duplicate/unknown-field rejection; `.sdd/work/<slug>/work.yaml` recognition |
| REQ-018 selection rules | Implemented | Invalid explicit root never falls back; new-work preview independent of unrelated items |

## 3. Extension seams for future specs

Reserved, implemented seams that later specs consume without modifying
foundation behavior:

- **Workspace transactions** consume `FileFacts` (add write-side effects
  behind exact authorized plans), `Snapshotter` (raw coherent snapshots),
  `VerifyPlanDigest`, and `protocol/testdata/integration/transaction-seam-v1.json`
  (INT-TX-001..007). Foundation never applies.
- **Workflow graph** consumes normalized layout/selection inputs and the
  `WorkflowDefinitionDigest` binding point (currently derived from exact
  embedded preview-definition bytes plus the requested workflow). Foundation never
  compiles or routes.
- **Artifact validation** owns generic parsing, typed normalization, and
  semantic/derivation canonicalization (AM-04); consumes the digest framing
  discipline and `PolicySnapshot` fact shape. Foundation never normalizes
  artifact meaning.
- **Completion lifecycle** consumes validation/execution evidence and the
  dry-run next-action revision convention (`doctor --root` with
  `expectedRevision`) for readiness/staleness handoff. Foundation never decides
  acceptance, generates DONE, or mints receipts.
- **Protocol extension** goes exclusively through the foundation-owned
  `protocol/extension-inventory.json`; no implementation claims `1.1`.

## 4. Known limitations carried forward

- Invariant failures use `DIAG_INTERNAL_ERROR` with exit 3 because the frozen
  taxonomy has no ninth process category; internal construction paths return
  a redacted envelope rather than panicking.
- `Snapshotter` and `workspace.Reader` exist, but no production capturer is
  connected to `Engine`; direct `FileFacts` discovery is not yet a coherent
  multi-file observation under a concurrent writer. FND-005 remains open.
- Local-development policy observes a fixed floor with no authority claims
  (DEC-01); trusted sources arrive with policy integration.
- Safe-slug derivation is identity-as-slug until journey profiles own it;
  separator-bearing identities fail closed.
- Native runtime tests pass on Ubuntu, macOS, and Windows; mutation support is
  still not claimed because foundation only observes and previews.
- `todo.md` distinguishes drafted prerequisite specs from owner acceptance
  and implementation; no directory is treated as accepted merely because it
  contains requirements/design/tasks files.
