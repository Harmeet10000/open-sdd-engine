# OpenSpec delta format

Each Change loop contains a proposal, linear task list, and one or more capability
deltas under `.kiro/changes/<id>/specs/<capability>/spec.md`.

```md
## ADDED Requirements
### Requirement: Session timeout
The system SHALL expire inactive sessions.

#### Scenario: Default timeout
- **WHEN** 24 hours pass
- **THEN** invalidate the session
```

Supported operations are `ADDED`, `MODIFIED`, `REMOVED`, and `RENAMED Requirements`.
Archive validates every delta before applying it. Rename occurs first, followed by
removal, modification, and addition, so a change cannot silently recreate a removed
requirement under an old name.
