# The dual-loop model

Use the **Change loop** for bounded work in an existing codebase: fixes, focused
refactors, dependency updates, and behavioral edits that can be reviewed as a delta.
Its durable output is an archived merge into `.kiro/domains/`.

Use the **Spec loop** when the work creates a new subsystem, spans multiple components,
or carries architectural, performance, or compatibility risk. Its durable output is a
requirements-to-design-to-task trace, then a dependency graph that schedules execution
in safe parallel waves.

`sdd route` considers stated time budget, brownfield/greenfield signals, complexity
keywords, and description length. The recommendation is evidence, not a lock: `sdd new`
always lets the developer choose the other loop.
