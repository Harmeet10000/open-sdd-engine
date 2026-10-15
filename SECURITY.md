# Security policy

## Supported versions

Open SDD Engine has not published a stable release. Security fixes currently
target the default branch only.

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability. Use the repository's
private GitHub security advisory flow after access is available, or contact the
repository owner through an agreed private company channel.

Include the affected command or protocol version, platform and filesystem,
reproduction steps, expected boundary, observed impact, and whether secrets or
authoritative workspace data may have been exposed or modified. Avoid placing
real credentials or confidential workspace contents in the report.

## Security boundaries

Treat workspace files, Markdown, YAML, Git metadata, imported specifications,
agent output, and subprocess output as untrusted input. Workflow data cannot
grant shell or network authority. Commands must be selected from built-in or
explicitly approved policy, passed as argv rather than shell text, and executed
under bounded resource and cancellation policy.

SQLite and other indexes are disposable projections; they cannot repair or
override authoritative Markdown, YAML, or journal records. A corrupt or
incompatible projection is quarantined or rebuilt.
