# ADR-0004: Use yaml.v3 for canonical layout decoding

- Status: Accepted
- Date: 2026-10-17
- Scope: Foundation read-only discovery (`workspace.yaml`, `work.yaml` decoding in FND-005)

## Context

Foundation discovery must decode `.sdd/workspace.yaml` and
`.sdd/work/<slug>/work.yaml` with three hard requirements from the foundation
design: malformed records fail, unknown fields fail conservatively (never
silently accepted), and format/schema versions compare as exact integers, not
semver ranges. The Go standard library has no YAML implementation.

## Decision

Use `gopkg.in/yaml.v3` pinned to `v3.0.1`, always through a strict decoder
(`KnownFields(true)`) constructed at the single decode boundary in the
foundation discovery code. Duplicate mapping keys, unknown fields, negative
values into unsigned integers, and empty documents are decode errors, not
defaults.

This decision covers canonical layout records only. Artifact Markdown/YAML
parsing, profiles, and semantic canonicalization belong to
`artifact-validation`, whose parser selection (per ADR-0003 ancillary notes)
remains a separate, spike-confirmed decision. Nothing here pre-selects it.

## Why this dependency

- Strict mode enforces the unknown-field rejection the contract requires; a
  hand-rolled subset parser would reimplement YAML grammar edge cases
  (anchors, duplicate keys, number forms) worse.
- Pure Go: no CGO, no impact on the no-CGO release target.
- Zero transitive dependencies; MIT license.
- Small integration surface: decode bytes into two closed structs at one
  boundary. The library type never crosses into engine values — decoded
  structs convert immediately into validated immutable layout values.

## Rejected alternatives

- Standard library only: no YAML implementation exists.
- Hand-rolled minimal parser: cannot match YAML's grammar cheaply, and every
  missed edge case becomes a workspace-acceptance divergence.
- `sigs.k8s.io/yaml` (JSON bridge): adds transitive dependencies and
  JSON-round-trip type quirks for no benefit at this narrow boundary.

## Consequences and controls

- `go.mod`/`go.sum` pin the dependency; CI uses `-mod=readonly`.
- Discovery tests cover empty, malformed, duplicate-key, unknown-field,
  negative-integer, and wrong-version fixtures through the strict decoder.
- `govulncheck` remains a release signal for reachable vulnerabilities.

## Exit strategy

Replacement requires only a decoder that rejects unknown fields and duplicate
keys, reports malformed input as errors, and passes the same layout fixtures.
No layout value, port, or diagnostic code changes to swap the library.
