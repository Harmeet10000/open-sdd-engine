# Upstream compatibility policy

Upstreams are shallow-cloned under `vendor/upstreams/` at the commits recorded in
`manifest.json`.
Open SDD Engine does not execute their code or silently copy their agent prompts. Each
upgrade first refreshes fixtures, then passes the adapter compatibility tests.

This keeps the external formats portable while making the engine's public contracts its
own responsibility.
