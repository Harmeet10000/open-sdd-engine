# Agent autonomy policy

`sdd route` chooses a loop. `sdd skills select` chooses only built-in, explainable
capabilities. A selection reports why it was made and whether it is required. It never
installs a plugin, runs a remote skill, or writes code by itself.

The default Ponytail policy is conservative: reuse repository code, platform features,
and installed dependencies before adding anything new. It never suppresses validation,
security review, or acceptance criteria.

AutoResearch is an experiment runner, not a permission grant. Experiments need an
explicit command and preserve the primary worktree on failure; a future isolated
worktree mode may make an accepted experiment mergeable.
