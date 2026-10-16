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

Install the pinned development tools once, then run the same contract enforced
by GitHub:

```sh
make install-tools
make ci
make vuln
```

`make ci` checks formatting, module integrity, package loading, vet, shuffled
tests, the race detector, the coverage floor, golangci-lint, gosec, and GitHub
Actions syntax. `make vuln` is separate because it queries the current Go
vulnerability database and therefore requires network access.

Tests that touch files must use a test-owned temporary directory. Concurrency,
locking, cancellation, recovery, parser, and mutation work also requires the
relevant race, fuzz, fault-injection, or platform tests described by its spec.

## Commit contract

Commits and pull-request titles use Conventional Commit subjects:

```text
type(optional-scope): imperative summary
```

Allowed types are `build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`,
`refactor`, `revert`, `style`, and `test`. Use `!` before the colon for a
breaking change. Subjects must be at most 100 characters and must not end with
a period. Validate the current branch with:

```sh
make commitlint
```

Do not hide unrelated work in a passing commit. A mechanically valid subject
does not replace coherent history or a reviewable diff.

## Pull requests

A pull request should contain one coherent vertical slice and explain:

- the requirement and task IDs it implements;
- public protocol or workspace-format impact;
- failure, cancellation, rollback, and recovery behavior;
- dependencies introduced or removed and why;
- commands actually run and evidence produced;
- remaining limitations or follow-up work.

Pull requests are squash-merged. The pull-request title becomes the resulting
commit subject, so it is validated by the same repository-native Go tool as
individual commits. Required checks must pass on the current head before merge.

Generated files and copied fixtures must identify their source and refresh
procedure. Do not commit credentials, local caches, build output, or the
rehydratable repositories under `vendor/upstreams/`.
