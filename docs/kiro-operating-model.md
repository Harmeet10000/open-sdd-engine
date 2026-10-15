# Kiro-style operating model

Open SDD Engine uses the same useful operating model without requiring Kiro:
requirements establish observable behavior, design is derived from approved requirements,
and tasks are derived from the design. A change to an upstream artifact requires review of
its downstream artifacts before the workflow is regenerated.

`requirements.md` uses EARS. Required behavior is written as `shall`, including
ubiquitous, event-driven, state-driven, optional-feature (`Where`), unwanted-behavior
(`If ... then`), and combined state/event forms. Every feature spec also records edge
cases, non-goals, invariants, and verification steps. Bugfix invariants use `shall
continue to` to make regression protection explicit.

`sdd init` creates `.kiro/steering/product.md`, `tech.md`, and `structure.md`. These
are persistent project context and should be maintained with the same care as source
code. Agent handoffs should read them with the approved workflow, coding profile, and
recalled memory before planning or implementation.
