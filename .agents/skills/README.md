# Curated Go skills

These project-local skills were copied from:

```text
/home/harmeet/Desktop/prompts/skills/backend/Golang
```

Source repository revision:

```text
https://github.com/Harmeet10000/skills.git
a77a0d07c5cedb7e27aac6b3f55906bdb592d867
```

The copied files retain their per-skill author, version, and license metadata.
The source repository's license is preserved in `SOURCE-LICENSE`.

They are guidance for coding agents, not executable dependencies and not
product authority. Ratified `.kiro/specs/`, accepted ADRs, protocol schemas,
and repository instructions win when a generic skill conflicts with this
project.

## Included

| Skill | Why it belongs here |
| --- | --- |
| `golang-cli` | Command parsing, I/O separation, stable exit codes, completion, and test seams |
| `golang-code-style` | Idiomatic and reviewable Go implementation |
| `golang-context` | Cancellation and deadline propagation through the engine |
| `golang-concurrency` | Bounded execution, leak prevention, and deterministic shutdown |
| `golang-continuous-integration` | GitHub Actions, security gates, and release pipelines |
| `golang-dependency-management` | Minimal Version Selection, vulnerability review, and upgrade policy |
| `golang-design-patterns` | Hexagonal boundaries, lifecycle, and manual composition |
| `golang-documentation` | CLI/reference documentation and contribution guidance |
| `golang-error-handling` | Error identity, wrapping, diagnostics, and single-point rendering |
| `golang-project-layout` | Capability-oriented package and command layout |
| `golang-safety` | Defensive copying, nil safety, and corruption-resistant code |
| `golang-security` | Filesystem, command, secret, and dependency threat controls |
| `golang-testing` | Tables, fuzzing, race tests, fault injection, and integration tests |
| `golang-troubleshooting` | Root-cause workflow for Bug Fix and production investigation |
| `golang-benchmark` | Evidence before data-layout or performance complexity |
| `golang-database` | Later SQLite projection and transaction discipline |
| `golang-popular-libraries` | Structured review when proposing a third-party module |

## Deliberately excluded

- Vendor-specific `golang-samber-*` skills: the engine is standard-library
  first and no corresponding dependency has been admitted.
- `golang-stretchr-testify`: standard `testing`, fuzzing, and focused helpers
  are sufficient until repeated assertion or mocking pain is demonstrated.
- Dependency-injection frameworks: the composition root uses explicit manual
  wiring.
- gRPC and web-service observability bundles: the v1 deliverable is a local
  CLI, not a network service.
- Duplicate broad skills such as `golang-pro` and `go-concurrency-patterns`:
  the selected focused skills already own those concerns.

## Maintenance

When refreshing these files, copy a complete skill directory so relative
references and assets remain valid. Update the source revision above, inspect
the diff, and run the repository checks. Do not automatically import new
skills or accept a skill's suggested dependency into `go.mod`.
