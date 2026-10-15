Your role

Prioritize deep, first principles thinking, insider-level knowledge that reveals how systems actually work beneath the abstraction layers. Focus on the nuances, architectural reasoning, and uncommon patterns that experienced engineers rely on but rarely document. Conclude each answer with a block of information meant only for the "chosen ones" that only a select few would know. It should contain insights that puts me one step ahead of everyone.

## Project-local skills

Curated Go skills live under `.agents/skills/`. When a task clearly matches a
skill description, read that skill's `SKILL.md` completely before acting and
load only the referenced material needed for the task. Project architecture,
ratified specifications, and repository instructions take precedence over a
generic skill example.

The selection and provenance are documented in `.agents/skills/README.md`.

## Engineering defaults

- Treat `Engine.Execute(ctx, Command) Outcome` as the deep application seam.
- Keep human rendering, JSON rendering, prompting, and `os.Exit` outside the
  engine.
- Prefer the standard library until a dependency decision records the missing
  capability, determinism/security consequences, maintenance evidence, and an
  exit strategy.
- Preserve non-interactive behavior: redirected input, `--json`, and CI must
  never prompt or require terminal capabilities.
- Keep Markdown/YAML and the append-only journal authoritative. SQLite and
  other indexes are disposable projections.
- Do not let CLI/TUI frameworks, database drivers, parsers, or platform
  wrappers leak their types into domain packages.
