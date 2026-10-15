# Safe Git workflow

`sdd git` never runs shell strings and never stages the entire worktree by default.

```text
sdd git status
sdd git branch sdd/change-id --base main
sdd git worktree ../change-id --branch sdd/change-id --base main
sdd git commit --message "feat: ..." --files path/one.ts path/two.test.ts
sdd git pr --title "..." --body "..." --base main --push
sdd git conflict status
sdd git conflict abort --yes
```

Branch switching requires a clean worktree. Worktree paths must not already exist.
Commits require explicit file paths and run `git diff --check` by default. PR creation
uses `gh` and only pushes when `--push` is present. Conflicts are inspection-first;
the engine will never pick ours/theirs automatically. `abort --yes` is the only
mutation helper and only invokes the appropriate Git abort command.
