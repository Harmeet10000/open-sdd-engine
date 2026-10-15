# Workspace & Project Inspection Notes

This document contains structural details of existing projects in the workspace `/home/harmeet/Desktop/Projects` relevant to the Spec-Driven Development CLI (`sdd`), compiled in June 2026.

---

## 1. Existing Spec Engine (`open-sdd-engine`)

- **Location**: `/home/harmeet/Desktop/Projects/OpenSpec-v2/open-sdd-engine`
- **Scaffold**: Bun + TypeScript workspace
- **Dependencies**:
  - `yaml` (v2.7.0)
  - `zod`
  - `commander` (in typescript cli router wrapper)
  - `vitest` (used for running test suite)

### Directory Structure (Top 2 levels)
```
open-sdd-engine/
├── package.json
├── tsconfig.json
├── README.md
├── schemas/
│   ├── change.meta.schema.json
│   ├── diagnostics-output.schema.json
│   └── task-dependency-graph.schema.json
├── src/
│   ├── cli.ts
│   ├── change/
│   ├── commands/
│   ├── core/
│   ├── research/
│   ├── router/
│   ├── spec/
│   └── templates/
└── tests/
    ├── integration/
    └── unit/
```

### package.json Contents
```json
{
  "name": "open-sdd-engine",
  "version": "0.1.0",
  "description": "Kiro-compatible dual-loop spec-driven development CLI",
  "type": "module",
  "bin": {
    "sdd": "./dist/cli.js"
  },
  "scripts": {
    "build": "bun build src/cli.ts --outdir dist --target node",
    "test": "bun test tests/",
    "prepublishOnly": "bun run build"
  },
  "engines": {
    "node": ">=20"
  },
  "dependencies": {
    "yaml": "^2.7.0"
  }
}
```

---

## 2. Kiro Specs in `sweet` Project

- **Location**: `/home/harmeet/Desktop/Projects/sweet/.kiro/specs/sweet-api-framework`
- **Key Files**:
  - `spec.meta.json` (or `.config.kiro`)
  - `requirements.md` (EARS rules)
  - `design.md` (Mermaid architectural overview)
  - `tasks.md` (checklists with traceability links like `_Requirements: 1.1_`)

---

## 3. Kiro Specs in `ScaleForge` Project

- **Location**: `/home/harmeet/Desktop/Projects/ScaleForge/.kiro/specs/`
- **Spec folders**:
  - `bun-elysia-migration/`
  - `deployment-infrastructure/`
  - `elasticsearch-search/`
  - `fastify-migration/`
  - `novu-notifications/`
  - `saas-payment-system/`

### Example EARS structure (`ScaleForge/bun-elysia-migration/requirements.md`):
Acceptance Criteria follow EARS (Easy Approach to Requirements Syntax) triggers:
- `WHEN the application starts THEN Bun SHALL replace Node.js as the JavaScript runtime`
- `WHEN the application runs THEN Elysia.js SHALL replace Express as the web framework`
- `WHEN errors occur THEN they SHALL be logged with the same format and correlation IDs`

---

## 4. Spec-Kit Searches
- There is no repository named `speckit` or `spec-kit` in `/home/harmeet/Desktop/Projects`.
- However, Spec-Kit workflow commands and stages (`specify` and `plan` CLI scaffolds) have been successfully integrated directly into our new `sdd` engine command router.
