---
inclusion: auto
name: sdd-cli-ux
description: Apply when specifying, designing, implementing, or reviewing commands, flags, prompts, help, terminal output, JSON output, diagnostics, exit behavior, progress, or agent-facing CLI interactions.
---

# CLI experience policy

The CLI is a product interface and a machine protocol, not a debug view over engine internals. Optimize the default path for a developer who knows the intended change but does not know this product's workflow model.

## Experience target

The ordinary journey remains:

```text
sdd init
sdd start "intent"
sdd next
sdd status
sdd check
sdd finish
```

`sdd start` may automatically choose a maintained preset from evidence. Users can explicitly request `--workflow bug-fix` or `--workflow quick-spec`. Human output uses the labels **Bug Fix** and **Quick Spec**; stable machine IDs remain lowercase and hyphenated.

Every command answers three questions, in this order:

1. What happened or what is the current state?
2. What, if anything, needs attention?
3. What is the single most useful next command?

Do not expose artifact graphs, validators, leases, digests, transaction phases, storage backends, adapter names, or workflow taxonomy unless the user asks for detail or the information is required to resolve a failure.

## Command-language rules

- Use a small, consistent verb vocabulary. Prefer `start`, `next`, `status`, `check`, and `finish` over framework-specific jargon.
- A command name describes user intent, not an internal subsystem. Do not add commands named after packages, tables, adapters, or implementation phases.
- The same concept has one canonical name in help, output, schemas, docs, and files. Keep a glossary and reject synonyms that create separate mental models.
- Global flags keep the same spelling and meaning everywhere. Command-local flags do not shadow them.
- Positional arguments are reserved for the one obvious primary object. Use named flags when two strings could be confused or reordered unsafely.
- Do not make abbreviations the only interface. Short flags are optional conveniences for frequent, unambiguous operations.
- Aliases never introduce distinct behavior and do not appear as the canonical form in generated instructions.
- Defaults must be safe, local, deterministic, and explainable. Inference with more than one plausible target fails with choices; it never guesses.
- Workflow selection explains the user-relevant reason and any escalation. Default output says, for example, `Bug Fix · critical safeguards` rather than exposing journey/rigor/profile implementation vocabulary.
- A removed or renamed command gets a bounded deprecation path with a replacement command. Do not silently reinterpret old input.

## Progressive disclosure

Default human output is short. Add detail through `--verbose`, targeted inspection commands, or referenced report files rather than printing the engine's full state.

Use this response shape for ordinary human output:

```text
<outcome in one sentence>
<only the findings that require action>
Next: sdd <command> ...
```

- Do not print a banner on every invocation.
- Do not print success theater. State the material result: created path, checked revision, changed count, or satisfied gate.
- Give one recommended next action first. Secondary actions may follow, clearly labeled.
- Errors lead with the problem in the user's vocabulary, then the cause, remediation, and a copyable command when one exists.
- Bug Fix progress reports evidence-bearing stages—reported, reproduction state, investigating, diagnosis ready, regression defined, fixing, verifying—not a generic percentage.
- Quick Spec reports blocking questions first, then whether the committed core scope is on track or needs a scope replan, and whether the three-file draft bundle is generating, incomplete, ready for review, stale, or approved. It never describes generated drafts as accepted or dumps priority accounting unless the user inspects it.
- Never print a stack trace by default. `--debug` may expose implementation detail on stderr after secrets are redacted.
- Help follows: summary, usage, arguments, options, examples, exit behavior, and related commands. Include at least one realistic example for non-trivial commands.
- Invalid input prints the specific error and concise command help, not the entire root manual.

## Human and machine contracts

Human rendering and JSON rendering are two views of the same `Outcome`; command handlers do not independently reconstruct meaning.

- Primary payload goes to stdout. Diagnostics, progress, and debug logs go to stderr.
- `--json` writes exactly one versioned JSON document to stdout for every completed invocation, including failures that reached command execution.
- JSON mode emits no ANSI sequences, spinner frames, prose prefix, or terminal-width formatting.
- Never require an agent to parse English text. It matches schema fields, stable diagnostic codes, exit codes, IDs, and revisions.
- JSON field meaning is backward-compatible within a protocol major version. Additive fields are permitted; incompatible changes require a new version and negotiation behavior.
- Arrays with no semantic ordering are emitted in an explicitly defined deterministic order. Tests must not rely on JSON object-key order.
- Paths in machine output state whether they are workspace-relative or absolute. Persisted artifacts use normalized workspace-relative paths; transient display may use native separators.
- Timestamps include an offset and are UTC in persisted/machine records unless a schema explicitly says otherwise.
- Secret values, bearer tokens, environment values, and sensitive command output are redacted in both renderers and debug logs.

## Stable exit behavior

Keep the accepted exit-code taxonomy aligned with the product plan:

| Code | Meaning |
|---:|---|
| `0` | command completed and the requested condition holds |
| `1` | validation or verification failed |
| `2` | usage error or malformed input |
| `3` | workspace, configuration, or schema error |
| `4` | conflict, stale revision, or active claim prevents mutation |
| `5` | required external tool or process failed |
| `6` | policy or permission denied the action |
| `130` | interrupted after child-process cleanup |

- Programs branch on the exit code plus stable diagnostic code, never message text.
- A warning does not cause failure unless the selected policy profile elevates it.
- A clean `check` returns `0`; failed requirements return `1`; inability to perform the check returns the operational category that caused it.
- Do not collapse cancellation, validation failure, usage error, and internal defect into exit code `1`.

## Interaction and safety

- Non-interactive execution never prompts. If consent is required, fail with the exact flag or preparatory command needed.
- Detect TTYs only to choose presentation, prompting, and progress behavior; TTY state never changes product semantics.
- Destructive or externally visible operations support an inspectable plan. Confirmation authorizes the exact plan digest, not a newly computed action.
- Prefer `--dry-run` for mutations and require explicit authorization for destructive effects. `--yes` is valid only where the operation and target are already unambiguous.
- Respect cancellation immediately, stop progress rendering, supervise descendant processes, and report whether cleanup completed.
- Never launch an editor, browser, pager, authentication flow, network call, Git commit, or agent process unless the command contract explicitly requests it.
- Read-only commands remain read-only even when state is stale. They may recommend repair but must not perform it.
- When recovery is ambiguous, show the prepared transaction and safe choices; do not guess.

## Terminal quality

- Color reinforces meaning but never carries meaning alone. Honor `NO_COLOR`, non-TTY output, and `TERM=dumb`.
- Use Unicode only when support is known; all information must remain legible in ASCII fallback.
- Keep tables narrow and degradable. For unbounded data, offer stable plain or JSON output rather than truncating identifiers.
- Progress appears only for work with perceptible duration. Use a spinner for unknown duration and measured counts for known work; never invent a percentage.
- Progress redraws stay on stderr and disappear cleanly on completion, failure, or interruption.
- Do not assume a wide terminal, interactive stdin, writable current directory, home directory, Git repository, network, or installed runtime.
- `--version` and `--help` work outside a workspace and from a read-only directory.

## Agent ergonomics

- Every actionable JSON result includes stable `nextActions` with command arguments represented structurally, not only as a shell string.
- Mutation results include the authorized plan ID/digest, before/after revision, changed paths, and receipt location.
- Conflict diagnostics include expected versus observed revisions/digests and a safe refresh/replan action.
- Large context is returned as bounded references with provenance and digests; do not truncate JSON into invalid or ambiguous fragments.
- Idempotency keys and claim IDs are explicit where retries could duplicate effects.
- Commands must be usable without shell interpolation. Examples that invoke subprocesses use executable-plus-argument arrays in the machine protocol.

## UX acceptance checks

Every command spec and implementation includes fixtures for:

1. success in a TTY;
2. success with redirected stdout;
3. schema-valid `--json` success;
4. each expected failure category in human and JSON modes;
5. no workspace, malformed workspace, and incompatible version;
6. non-interactive invocation that would otherwise prompt;
7. narrow terminal, no color, and non-Unicode fallback where rendering is material;
8. interruption and timeout where work can block;
9. secret-bearing inputs and redaction;
10. copy-paste validity of every recommended next command.
11. Bug Fix non-reproduction, intermittent reproduction, critical-risk escalation, and diagnosis-without-evidence behavior.
12. Quick Spec zero-question flow, partial bundle rejection, upstream-edit staleness, no-agent fallback, risk escalation without lost answers, Must-cap replan, unauthorized Must demotion, and deferred/excluded-scope visibility.

Golden tests assert meaning and stream placement. Avoid brittle full-screen snapshots when a focused semantic assertion is sufficient.

## Review rejection criteria

Reject a CLI change when it:

- exposes an internal package or storage mechanism as the user's primary concept;
- changes JSON semantics without protocol/version analysis;
- prints logs or progress to stdout in machine mode;
- asks for input when stdin is non-interactive;
- mutates during `status`, `check --dry-run`, or another declared read-only operation;
- makes English message parsing necessary;
- returns success when requested verification did not run;
- hides ambiguity behind an inferred target;
- adds a flag without defining precedence, default, environment/config interaction, and help text;
- provides an error without a stable code and a feasible remediation path.

## Reference basis

- The community [Command Line Interface Guidelines](https://clig.dev/) inform discoverability, stream separation, TTY-aware presentation, JSON output, prompting, help, and exit behavior.
- The normative product behavior remains `docs/cli-product-plan.md`, especially its stable JSON envelope and exit-code sections.
