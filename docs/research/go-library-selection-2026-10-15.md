# Go Library Selection for the SDD CLI

**Research date:** 2026-10-15  
**Status:** Research evidence; non-normative until ratified by an owning spec or ADR  
**Scope:** current first-party documentation and source repositories only  
**Decision context:** the SDD CLI is a machine-first protocol implementation with an optional human terminal adapter. It must preserve deterministic artifacts, exact workspace transactions, bounded execution, stable diagnostics, and a strict no-prompt contract in non-interactive mode.

## Executive decision

Use a small standard-library core and add libraries at the hexagonal boundary where they remove real implementation risk:

1. **Adopt Cobra as the command adapter** (or Kong, but not both). Keep command handlers thin and call application ports; never let the parser own workflow, transaction, or artifact rules.
2. **Adopt `go.yaml.in/yaml/v4` for new YAML input**, with explicit unique-key and known-field checks, a single-document policy, and an application-level byte/node/depth budget. Do not use YAML as the canonical artifact representation.
3. **Adopt `github.com/santhosh-tekuri/jsonschema` v6 for Draft-07 validation**, but compile only bundled/local schemas and place resource limits around parsing and validation. Schema validation complements semantic validators; it does not replace them.
4. **Adopt `github.com/yuin/goldmark/v2` for Markdown semantic parsing**, while retaining the original bytes and source spans for diagnostics and canonicalization. Never round-trip a document by rendering the AST if byte-preserving edits matter.
5. **Adopt `golang.org/x/term` only in the presentation adapter** to detect TTY capability. TTY detection must not decide protocol behavior; explicit `--output`/non-interactive policy wins.
6. **Use `go-cmp` in tests only**, especially for semantic values and rich diffs. Production identity, digest, and equality rules remain explicit domain code.
7. **Defer Bubble Tea v2, Bubbles v2, Lip Gloss v2, and Huh v2** until a real interactive inspect/recovery workflow exists. They are presentation dependencies, not command or domain dependencies.
8. **Defer `modernc.org/sqlite` to the context-index phase**. It is a credible pure-Go SQLite option, but introducing a database before workspace/artifact semantics are stable would make the database schema an accidental domain contract.
9. **Adopt `govulncheck` and Staticcheck in CI/tooling**, and adopt GoReleaser when release artifacts and signing are in scope.

The most important architectural rule is dependency direction: `domain -> application ports -> adapters`. Bubble Tea, Cobra/Kong, YAML, Markdown, SQLite, and terminal detection must all point inward through adapters; no domain package may import them.

## Recommendation matrix

| Component | Decision | Where it belongs | Why | Guardrail |
|---|---|---|---|---|
| Go standard library (`encoding/json`, `crypto/sha256`, `io/fs`, `os`, `os/exec`, `log/slog`, `testing`, fuzzing) | **Adopt now** | Domain/application and small adapters | Deterministic hashing, JSON envelopes, filesystem ports, process control, structured logs, and tests need no third-party policy surface | Inject clocks, IDs, environment, filesystem, and process runners; never hash map iteration order or wall-clock output |
| Cobra | **Adopt one CLI parser** | `internal/presentation/cli` and `cmd/sdd` | Subcommands, nested commands, POSIX flags, help, suggestions, aliases, completion, and man-page generation match `start`, `next`, `status`, `check`, `finish`, `inspect`, and `recover` | `RunE` only translates arguments to an engine command and renders an outcome; no business logic in `cmd` |
| Kong | **Viable alternative; do not add with Cobra** | Same CLI adapter boundary | Struct/tag-driven command definitions can reduce repetitive flag wiring and supports nested commands, custom decoders, validation, and plugins | Keep parser structs separate from domain requests; choose Kong only if its declarative model is preferred over Cobra’s ecosystem/docs conventions |
| Bubble Tea v2 | **Defer; adopt for an explicit TUI** | `internal/adapters/tui` | Production-grade stateful terminal UI based on The Elm Architecture, suitable for full-window, inline, or mixed UIs | Never run it for JSON output, piped stdout, CI, or `--non-interactive`; TUI receives already-computed view models |
| Bubbles v2 | **Defer with Bubble Tea** | TUI adapter | Reusable inputs, lists, tables, viewport, progress, and other widgets | Pin Bubble Tea, Bubbles, and Lip Gloss to compatible v2 lines; they are a coordinated set |
| Lip Gloss v2 | **Defer with Bubble Tea** | Presentation adapter | Deterministic terminal styles/layout and controlled I/O in v2 | Render to an injected writer; do not put ANSI bytes into machine protocol responses |
| Huh v2 | **Defer / optional** | Interactive wizard adapter | Forms/prompts with an accessible mode and Bubble Tea integration | The application must return a typed `interactive_required` diagnostic instead of prompting when non-interactive |
| Goldmark v2 | **Adopt when Markdown artifacts land** | Markdown parser adapter | CommonMark-compliant AST/CST pipeline, extensibility, source-oriented parsing, and good performance | Preserve source bytes and offsets; impose input limits; reject or explicitly sanitize unsafe URL/HTML behavior for untrusted content |
| `santhosh-tekuri/jsonschema` v6 | **Adopt for schema validation** | Artifact validation adapter | Supports Draft-04/06/07/2019-09/2020-12, schema cycles, output/error inspection, formats, and multiple instances | Bundle/allowlist schema resolvers; prohibit network fetches; set size/depth/ref budgets; map errors to stable diagnostics |
| `go.yaml.in/yaml/v4` | **Adopt for new YAML** | Input parser adapter | Official YAML-org maintained fork; v4 is the active line and adds explicit `WithUniqueKeys`, `WithKnownFields`, and single-document options | Parse into a node/typed intermediate with limits; reject duplicate keys and aliases beyond a budget; do not use emitted YAML as digest input |
| `goccy/go-yaml` | **Avoid as the default** | — | Rich tokenizer/parser and reference features, but more capability than the SDD contract needs | Consider only if a measured source-preserving requirement cannot be met with the official YAML library; document the security review |
| `golang.org/x/term` | **Adopt at TTY edge** | Terminal capability adapter | Official terminal support; `IsTerminal`, `GetSize`, raw mode, and password support | Capability detection is advisory; explicit output mode and `NO_COLOR`/CI policy are authoritative |
| `github.com/google/go-cmp` | **Adopt test-only** | `_test.go` and test helpers | Safer semantic comparison/diffs than `reflect.DeepEqual`, with controlled comparers and transformers | Never use `cmp.Equal` as the implementation of artifact identity, canonical equality, or security checks |
| `modernc.org/sqlite` | **Defer** | Future context-index adapter | Pure-Go, no cgo, `database/sql` driver, virtual-table extension surface, and broad platform support | Benchmark against workload and native SQLite; review transitive/generated code licenses; keep SQLite behind a repository port |
| `govulncheck` | **Adopt in CI/release gates** | Tooling, not application import | Go’s low-noise vulnerability analysis follows reachable calls rather than flagging every transitive package | Pin tool versions in CI and record database timestamp/results in the build receipt |
| Staticcheck | **Adopt in CI** | Tooling | Fast, focused checks designed for low false-positive CI use, complementary to `go vet` | Treat findings as a merge gate after establishing a baseline; avoid suppressions without a reason and owner |
| GoReleaser | **Defer until release hardening** | `.goreleaser.yaml`, release workflow | Automates cross-platform build/archive/publish/announce flow and supports verifiable Go builds | Require clean tree, SemVer tag, reproducible inputs, provenance/SBOM/signing policy, and a dry-run before publishing |

## Bubble Tea and the Charm ecosystem

Bubble Tea v2’s import path is `charm.land/bubbletea/v2`; the project describes itself as a stateful terminal framework based on The Elm Architecture and supports inline, full-window, and mixed terminal applications. The v2 migration is intentionally coordinated with Bubbles v2 and Lip Gloss v2. Sources:

- [Bubble Tea repository](https://github.com/charmbracelet/bubbletea)
- [Bubble Tea v2 announcement and import-path change](https://github.com/charmbracelet/bubbletea/discussions/1374)
- [Bubbles v2 upgrade guide](https://github.com/charmbracelet/bubbles/blob/main/UPGRADE_GUIDE_V2.md)
- [Lip Gloss v2 repository](https://github.com/charmbracelet/lipgloss)
- [Huh? forms and prompts](https://github.com/charmbracelet/huh)

Bubble Tea is a good fit for a future `sdd inspect --tui`, `sdd recover --interactive`, or a live bounded-execution monitor. It is not a good fit for the protocol path itself because terminal rendering, alternate-screen state, key handling, and terminal ownership are inherently presentation concerns. The CLI should first compute a pure view model:

```text
workspace snapshot -> workflow evaluation -> application result -> view model
                                                        |-> JSON renderer
                                                        |-> text renderer
                                                        `-> Bubble Tea renderer
```

The TUI must not perform writes directly. A confirmation action should produce an application command with a candidate ID and expected snapshot/plan digest; the transaction port then performs the same guarded commit used by non-TUI callers. This prevents “interactive mode” from becoming a second, weaker authorization path.

Two subtle v2 concerns matter here:

- Lip Gloss v2 makes style and I/O behavior more explicit. Inject the writer and terminal profile rather than allowing global stdin/stdout discovery to leak into protocol decisions.
- Bubble Tea’s normal rendering occupies the terminal output. Logs and diagnostics must go to an injected file/stderr sink, never stdout, when a TUI owns the terminal.

## Cobra versus Kong

### Cobra: recommended default

Cobra documents nested subcommands, POSIX-compatible flags, cascading flags, suggestions, help generation, shell completion, aliases, and man pages. Those map directly to a company CLI that must be discoverable by humans while preserving a stable machine contract. Source: [Cobra README](https://github.com/spf13/cobra/blob/main/README.md) and [Cobra user guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md).

Use Cobra as a deliberately thin adapter:

```text
cmd/sdd/main.go
  -> cmd/sdd/root.go                      (composition and dispatch)
  -> internal/presentation/cli/*          (parse, classify output mode)
  -> internal/engine                      (typed command and outcome)
  -> internal/presentation/{jsonout,human,tui} (render only)
```

Avoid Cobra’s optional Viper integration unless a configuration contract is explicitly designed. Ambient config precedence (`flags > environment > files > defaults`) is an easy way to make a workflow non-reproducible. If configuration is eventually needed, parse it into a versioned input object and include the effective-config digest in the request/receipt.

### Kong: credible alternative

Kong v1.0.0 expresses commands and flags as Go types/tags, supports nested commands, custom decoders, validation, plugins, and generated help. Source: [Kong repository](https://github.com/alecthomas/kong).

Kong can reduce repetitive parser code, but the type-driven structure can tempt developers to reuse command structs as domain DTOs. Do not do that: parser tags are an adapter concern, and domain requests must not acquire CLI defaults, environment interpolation, or help metadata. Choose Kong instead of Cobra only after a small spike proves that its generated usage, completion, error classification, and JSON-mode behavior satisfy the protocol contract.

### Do not use both

Using Cobra for the root and Kong for subcommands would create two flag semantics, two help models, and two places to test precedence. Pick one. The parser is replaceable; the application ports are the durable boundary.

## Markdown and artifact parsing

[Goldmark](https://github.com/yuin/goldmark) describes v2 as CommonMark-compliant, AST/CST-capable, extensible, and CJK-friendly. It is a strong semantic parser for requirements, design, tasks, and DONE sections. It should not be treated as a canonical serializer.

Implement the adapter in two layers:

1. Keep the original bounded byte slice and a line/column index.
2. Parse a semantic AST with Goldmark and attach source offsets/section IDs to domain nodes.

Canonicalization must be a domain operation over normalized records, not “parse then render Markdown.” Rendering can change whitespace, list markers, escaping, or extension output, which would change digests and invalidate candidate-bound evidence. Use a dedicated canonical JSON representation for semantic and derivation digests; retain source digests separately.

Goldmark’s own security guidance says raw HTML and potentially dangerous URLs are not rendered by default and recommends an HTML sanitizer such as bluemonday for more permissive rendering. For this CLI, prefer no HTML rendering at all in the machine path and sanitize only in an explicitly opt-in human renderer.

## JSON Schema Draft-07 validation

[Santhosh Tekuri’s JSON Schema implementation](https://github.com/santhosh-tekuri/jsonschema) currently documents v6.0.2 and support for Draft-04, Draft-06, Draft-07, Draft 2019-09, and Draft 2020-12. It also documents schema/validation cycle detection, introspectable error output, custom schema URLs, format assertions, and content assertions.

Use it for structural contracts:

- protocol envelope shape;
- artifact file shape after parsing/canonical conversion;
- workflow definition shape;
- receipt and diagnostic shape.

Keep semantic and policy checks outside JSON Schema:

- duplicate stable IDs across files;
- graph reachability and valid transition edges;
- evidence candidate binding and snapshot digest matching;
- transaction preconditions and lock ownership;
- MoSCoW rules such as a release’s Must budget;
- bounded execution effects.

Security and determinism requirements:

- Compile schemas from `go:embed` or a repository-controlled schema registry. Do not resolve arbitrary `http(s)` references during a normal CLI invocation.
- Allow only known `$schema` dialects and known local `$ref` roots.
- Bound input bytes, nesting depth, object/array counts, and reference expansion before invoking validation.
- Preserve validator error paths and map them into stable diagnostic codes; do not expose library-specific wording as a compatibility contract.
- Pin the validator module and test every schema against invalid, boundary, duplicate, unknown-field, and large-input fixtures.

## YAML safety

The active official YAML project is now the YAML organization’s maintained fork, importing as `go.yaml.in/yaml/v4`. Its README states that v1–v3 are frozen legacy lines and that ongoing features and routine fixes happen in v4. The v4 migration guide documents `WithKnownFields`, `WithUniqueKeys`, and `WithSingleDocument`. Sources: [official YAML repository](https://github.com/yaml/go-yaml) and [v3-to-v4 migration guide](https://github.com/yaml/go-yaml/blob/main/docs/v3-to-v4-migration.md).

That makes v4 the best starting point for a new CLI, but its options are not a complete resource policy. The adapter must still enforce:

- maximum bytes read (use `io.LimitReader` and reject truncation);
- one document only for specs/configuration;
- `WithUniqueKeys()` and `WithKnownFields()` where decoding into structs;
- maximum aliases/anchors, node count, scalar length, and nesting depth;
- an explicit policy for tags and merge keys;
- no remote reference directory or file inclusion;
- cancellation for any streaming parse path.

Parse to a node/intermediate representation first when source ranges, duplicate-key diagnostics, or canonical ordering matter. Decode into a typed struct only after the structural policy has passed. YAML values must be normalized into the same canonical JSON-like domain model used by JSON input, otherwise semantically equivalent YAML and JSON can produce different identities.

`goccy/go-yaml` exposes a tokenizer/parser and additional reference features, but that larger surface is not automatically safer. Its own documentation shows file-based reference directories and options that allow duplicate map keys. Treat it as an exception requiring a measured source-preservation need and a separate security review. Source: [goccy/go-yaml](https://github.com/goccy/go-yaml).

## SQLite and the future context index

[modernc.org/sqlite](https://modernc.org/sqlite) documents a pure-Go, no-cgo SQLite driver, a `database/sql` interface, virtual-table support, broad platform combinations, and generated/transpiled SQLite sources. It also explicitly notes that the transpiled core is slower than native C SQLite and calls out its third-party/public-domain licensing and SBOM considerations.

It is a reasonable future adapter for `sdd-context-index`, especially for a single-binary distribution where cgo is undesirable. It is not a reason to move the OKF or workspace truth into a database now:

- the canonical truth is currently the workspace/artifact snapshot and its byte/identity digests;
- a database introduces migration, lock, recovery, and backup semantics before those are settled;
- SQLite query order and schema evolution can accidentally become public protocol behavior;
- vector/full-text indexing is an optimization and must be rebuildable from canonical artifacts.

When the index spec begins, define a repository port such as `ContextIndex` and persist only projections with source artifact ID, revision, semantic digest, and derivation version. The index must be disposable and rebuildable. Benchmark `modernc.org/sqlite` against the target workload and compare with a native/cgo driver before committing to one. If using modernc, follow the project’s guidance to align the `modernc.org/libc` version and review its generated-code and transitive license inventory.

## Terminal detection and output policy

[`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term) is the official Go terminal support package; it exposes `IsTerminal`, `GetSize`, raw-mode, password, and restore APIs. Use it at the edge to answer “can this presentation render here?” Do not use it to silently switch the protocol mode.

Recommended precedence:

```text
explicit --output=json|text|tui
  > explicit --non-interactive / --interactive
  > protocol caller (stdin/stdout are pipes, CI, MCP/agent invocation)
  > TTY capability heuristic
```

If TUI is requested but unsupported, return a stable diagnostic and exit code. If no output mode is specified, default to text only for an interactive human invocation and JSON for machine/pipe invocation, with the chosen mode recorded in debug metadata but not in semantic digests.

## Testing and static analysis

Use [go-cmp](https://github.com/google/go-cmp) in tests for semantic equality and useful diffs. Its documented purpose is a safer, more configurable alternative to `reflect.DeepEqual`; custom comparers should explicitly ignore nondeterministic fields such as timestamps or receipt IDs. Never use it to define production identity.

Run `go vet`, native fuzz tests, and Staticcheck. [Staticcheck’s documentation](https://staticcheck.dev/docs/) describes 150+ focused checks intended to be fast, precise, and suitable for CI; its CLI is designed to behave like `go vet`/`go build`. Source: [Staticcheck getting started](https://staticcheck.dev/docs/getting-started/). Staticcheck consumes Go compiler export data, so run the current pinned release under its supported current Go toolchain in a separate CI job while `make verify` continues to prove the module's older minimum Go version.

Add [govulncheck](https://go.dev/doc/security/vuln/) to release/CI gates. Go’s documentation describes it as low-noise analysis that reports vulnerabilities reachable through functions your code actually calls, and documents the canonical Go vulnerability database. This is particularly useful once optional UI, YAML, Markdown, and SQLite dependencies arrive. Record tool version and database timestamp in CI logs; do not turn a transient database outage into an unexplained protocol failure for ordinary local development.

## Release tooling

[GoReleaser’s workflow](https://www.goreleaser.com/getting-started/how-it-works/) covers defaulting, building, packaging, publishing, and announcing. Its quick start documents `healthcheck`, build-only mode, and `--skip=publish`; its [verifiable builds documentation](https://www.goreleaser.com/customization/builds/verifiable_builds/) describes retaining module information and verifying modules through Go’s proxy/checksum infrastructure.

Adopt it in `sdd-release-hardening`, not in the foundation. The release gate should require:

- clean working tree and a signed/verified SemVer tag;
- `make verify`, Staticcheck, govulncheck, and cross-platform build tests;
- embedded protocol/schema/diagnostic inventory hashes;
- SBOM and license inventory;
- provenance/signature policy;
- a dry-run artifact comparison before publishing.

## Suggested dependency boundary

```text
cmd/sdd
  -> internal/presentation/cli         (Cobra; or Kong, exclusively)
  -> internal/engine
  -> internal/presentation/jsonout     (stdlib)
  -> internal/presentation/human       (stdlib; optional Lip Gloss)
  -> internal/presentation/tui         (Bubble Tea + Bubbles + Lip Gloss)

internal/artifact                      (Goldmark v2 and YAML v4 stay private here)
internal/validation                    (jsonschema v6 stays private here)
internal/presentation/terminal         (x/term)
internal/platform/sqlite               (modernc.org/sqlite, future)

engine-facing domain values and protocols (stdlib only, ideally)
```

The adapter packages may depend on application interfaces, but application/domain packages must not import these third-party libraries. This keeps the protocol testable with deterministic fakes and makes a future daemon/MCP adapter possible without dragging a terminal stack into it.

## Implementation sequence

1. Keep the current foundation and workspace snapshot work dependency-free.
2. Add Cobra (or Kong after a parser spike) only when the first executable command is wired.
3. Add YAML v4 and JSON Schema validation with resource limits as part of artifact/workflow validation.
4. Add Goldmark when Markdown artifact parsing is implemented; create source-offset and canonicalization tests before accepting real files.
5. Add x/term alongside explicit output-mode policy.
6. Add Staticcheck and govulncheck to CI before dependency count grows.
7. Add a `tui` package only when a user story needs interactive inspection/recovery; add Bubble Tea/Bubbles/Lip Gloss/Huh together and test all commands under pipes/CI.
8. Add SQLite only behind the context-index port after projections are proven rebuildable from canonical artifacts.
9. Add GoReleaser in release hardening after the binary, schema inventory, and protocol compatibility policy are stable.

## Sources

- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Bubble Tea v2 announcement](https://github.com/charmbracelet/bubbletea/discussions/1374)
- [Bubbles v2 upgrade guide](https://github.com/charmbracelet/bubbles/blob/main/UPGRADE_GUIDE_V2.md)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [Huh?](https://github.com/charmbracelet/huh)
- [Cobra README](https://github.com/spf13/cobra/blob/main/README.md)
- [Cobra user guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md)
- [Kong](https://github.com/alecthomas/kong)
- [Goldmark](https://github.com/yuin/goldmark)
- [Santhosh Tekuri JSON Schema](https://github.com/santhosh-tekuri/jsonschema)
- [JSON Schema Draft-07](https://json-schema.org/draft-07)
- [Official YAML Go repository](https://github.com/yaml/go-yaml)
- [Official YAML v3-to-v4 migration](https://github.com/yaml/go-yaml/blob/main/docs/v3-to-v4-migration.md)
- [goccy/go-yaml](https://github.com/goccy/go-yaml)
- [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term)
- [go-cmp](https://github.com/google/go-cmp)
- [Go vulnerability management](https://go.dev/doc/security/vuln/)
- [Staticcheck](https://staticcheck.dev/docs/)
- [GoReleaser how it works](https://www.goreleaser.com/getting-started/how-it-works/)
- [GoReleaser verifiable builds](https://www.goreleaser.com/customization/builds/verifiable_builds/)
- [modernc.org/sqlite](https://modernc.org/sqlite)
