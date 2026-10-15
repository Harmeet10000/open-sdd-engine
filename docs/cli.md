# CLI reference

```text
sdd init [--tools cursor] [--okf]
sdd new                         # interactive route-backed wizard
sdd new change <id> [--json]
sdd new spec <slug> [--json]
sdd validate [name] [--changes|--specs|--all] [--okf] [--strict] [--json]
sdd status [id] [--changes|--specs|--all] [--strict] [--json]
sdd instructions [id] [--changes|--specs|--all] [--strict] [--json]
sdd archive <change-id> --yes [--skip-specs]
sdd compile-waves <slug> [--json]
sdd route <description> [--budget 2h] [--json]
sdd skills select <description> [--json]
sdd research codebase|structure|web|docs <query> [--json]
sdd autoresearch run --cmd <command> [--metric <regex>] [--revert-on-fail] [--json]
```

The diagnostics JSON contract has stable uppercase codes, source location fields, and
an output schema at `packages/core/schemas/diagnostics-output.schema.json`.
