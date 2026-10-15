# Contributing to Open SDD Engine

Open SDD Engine is being built contract-first. A change is complete only when
its observable behavior, protocol surface, recovery behavior, and verification
evidence agree.

## Before changing code

1. Read `README.md`, `docs/cli-product-plan.md`, and the owning Kiro spec.
2. Identify the capability package that owns the behavior. Do not add generic
   `utils`, `helpers`, `services`, or `repositories` packages.
3. Record a dependency or public-contract decision before adding one. Prefer
   standard-library behavior and explicit data transformations.
4. Keep CLI/TUI and platform concerns behind the existing application and
   effect seams.

## Local verification

Run:

```sh
make verify
```

Tests that touch files must use a test-owned temporary directory. Concurrency,
locking, cancellation, recovery, parser, and mutation work also requires the
relevant race, fuzz, fault-injection, or platform tests described by its spec.

## Pull requests

A pull request should contain one coherent vertical slice and explain:

- the requirement and task IDs it implements;
- public protocol or workspace-format impact;
- failure, cancellation, rollback, and recovery behavior;
- dependencies introduced or removed and why;
- commands actually run and evidence produced;
- remaining limitations or follow-up work.

Generated files and copied fixtures must identify their source and refresh
procedure. Do not commit credentials, local caches, build output, or the
rehydratable repositories under `vendor/upstreams/`.
