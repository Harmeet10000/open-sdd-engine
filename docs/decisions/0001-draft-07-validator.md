# ADR-0001: Use jsonschema/v6 for protocol conformance

- Status: Accepted
- Date: 2026-10-16
- Scope: Build-time and test-time validation of the published SDD protocol

## Context

The foundation contract is JSON Schema Draft-07 and relies on local `$ref`,
`oneOf`, `const`, `additionalProperties`, and `if`/`then`/`else`. Go's standard
library parses JSON but does not implement JSON Schema. A hand-written subset
would create a second, weaker contract and would be especially likely to miss
conditional or reference behavior.

Validation must be deterministic and offline. A schema reference that is not
present in the checked-in registry must fail compilation; it must never trigger
an HTTP fetch during a build, test, or CLI invocation.

## Decision

Use `github.com/santhosh-tekuri/jsonschema/v6` pinned to `v6.0.2` behind the
protocol conformance tests. Configure the compiler explicitly for Draft-07,
register checked-in resources by stable `$id`, and install a loader that rejects
every unregistered URL.

The library is not a domain type and must not cross into engine, workflow,
workspace, artifact, or completion APIs. Published JSON schemas remain the
authority; the validator is replaceable machinery.

## Why this dependency

- It enforces the complete feature set used by the contract instead of a local
  approximation.
- It is pure Go, so it does not weaken the no-CGO release target.
- The API permits a closed resource registry and a network-denying loader.
- It has a small integration surface: compile a schema, then validate decoded
  JSON values.
- Its MIT license and maintenance evidence were reviewed in
  [`../research/go-library-selection-2026-10-15.md`](../research/go-library-selection-2026-10-15.md).

## Rejected alternatives

- Standard library only: no JSON Schema implementation.
- `jq` or hand-written structural checks: do not implement Draft-07 semantics
  and can incorrectly accept broken references or conditional payloads.
- Runtime network resolution: non-reproducible and a supply-chain boundary the
  CLI does not need.
- Generating Go types as the sole validator: loses schema-level compatibility
  checks and makes Go's decoder behavior the public contract.

## Consequences and controls

- `go.mod` and `go.sum` pin the dependency; CI uses `-mod=readonly`.
- The schema suite includes both accepted and rejected payloads, a closed
  command discriminator, conditional-keyword coverage, and a deliberately
  unregistered reference.
- `govulncheck` remains a release signal for reachable vulnerabilities.
- The top-level `vendor/` directory is research evidence, not Go vendoring;
  module commands continue to select module mode explicitly.

## Exit strategy

Replacement requires only a validator capable of compiling the same checked-in
Draft-07 schemas with a deny-by-default offline resolver and passing the same
fixtures. No schema, engine type, or public envelope should change merely to
swap the library.
