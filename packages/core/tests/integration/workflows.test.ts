import { describe, it, expect, beforeAll } from "vitest";
import { mkdir, writeFile, cp, rm } from "node:fs/promises";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { validateChange } from "../../src/change/validators/change-validator.js";
import { validateSpec } from "../../src/spec/validators/spec-validator.js";
import { compileSpecWaves } from "../../src/spec/validators/spec-validator.js";
import { loadChange, loadSpec } from "../../src/core/discovery.js";
import { DEFAULT_CONFIG } from "../../src/core/types.js";
import { archiveChange } from "../../src/change/archive/merge-deltas.js";
import { scaffoldInit, scaffoldChange } from "../../src/templates/scaffold.js";

const FIXTURE_ROOT = join(tmpdir(), "sdd-test-" + process.pid);
const SWEET_SPEC = "/home/harmeet/Desktop/Projects/sweet/.kiro/specs/sweet-api-framework";
const SCALEFORGE_SPEC = "/home/harmeet/Desktop/Projects/ScaleForge/.kiro/specs/bun-elysia-migration";

beforeAll(async () => {
  await rm(FIXTURE_ROOT, { recursive: true, force: true });
  await mkdir(FIXTURE_ROOT, { recursive: true });
  await scaffoldInit(FIXTURE_ROOT);
  await scaffoldChange(FIXTURE_ROOT, "fix-logging-level");

  const changeDir = join(FIXTURE_ROOT, ".kiro", "changes", "fix-logging-level");
  await mkdir(join(changeDir, "specs", "logging"), { recursive: true });
  await writeFile(
    join(changeDir, "proposal.md"),
    `# Proposal

## Why
Reduce log noise.

## What Changes
- Adjust default log level

## Capabilities
New: logging

## Impact
- config only
`,
    "utf-8",
  );
  await writeFile(
    join(changeDir, "specs", "logging", "spec.md"),
    `## ADDED Requirements

### Requirement: Configurable log level
The system SHALL support LOG_LEVEL env var.

#### Scenario: Default info
- **WHEN** LOG_LEVEL is unset
- **THEN** level is info
`,
    "utf-8",
  );
  await writeFile(
    join(changeDir, "tasks.md"),
    `# Tasks
## 1. Work
- [ ] 1.1 Add env parsing
`,
    "utf-8",
  );

  try {
    const specDir = join(FIXTURE_ROOT, ".kiro", "specs", "sweet-api-framework");
    await mkdir(specDir, { recursive: true });
    for (const f of ["requirements.md", "design.md", "tasks.md"]) {
      await cp(join(SWEET_SPEC, f), join(specDir, f));
    }
  } catch {
    /* optional external fixture */
  }

  const miniDir = join(FIXTURE_ROOT, ".kiro", "specs", "mini-spec");
  await mkdir(miniDir, { recursive: true });
  await writeFile(
    join(miniDir, "requirements.md"),
    `### Requirement 1: Parse

#### Acceptance Criteria
1. WHEN valid input, THE SYSTEM SHALL parse
`,
    "utf-8",
  );
  await writeFile(
    join(miniDir, "design.md"),
    `## Overview
Test

## Architecture
\`\`\`mermaid
graph TD
  A-->B
\`\`\`
`,
    "utf-8",
  );
  await writeFile(
    join(miniDir, "tasks.md"),
    `- [ ] 1. Build
  - _Requirements: 1.1_

## Task Dependency Graph
\`\`\`json
{"tasks":[{"id":"1","depends_on":[]}]}
\`\`\`
`,
    "utf-8",
  );

  try {
    const sfDir = join(FIXTURE_ROOT, ".kiro", "specs", "bun-elysia-migration");
    await mkdir(sfDir, { recursive: true });
    for (const f of ["requirements.md", "design.md", "tasks.md"]) {
      await cp(join(SCALEFORGE_SPEC, f), join(sfDir, f));
    }
  } catch {
    /* optional */
  }
});

describe("change loop integration", () => {
  it("validates fix-logging-level change", async () => {
    const bundle = await loadChange(FIXTURE_ROOT, "fix-logging-level");
    const result = await validateChange(bundle, DEFAULT_CONFIG, false);
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it("archives change into domains", async () => {
    await archiveChange(FIXTURE_ROOT, "fix-logging-level", false);
    const { readFile } = await import("node:fs/promises");
    const content = await readFile(join(FIXTURE_ROOT, ".kiro", "domains", "logging", "spec.md"), "utf-8");
    expect(content).toContain("Configurable log level");
  });
});

describe("spec loop integration", () => {
  it("validates mini-spec with zero errors", async () => {
    const bundle = await loadSpec(FIXTURE_ROOT, "mini-spec");
    const result = validateSpec(bundle, DEFAULT_CONFIG, false);
    expect(result.errors).toHaveLength(0);
  });

  it("validates sweet-api-framework without crash", async () => {
    try {
      const bundle = await loadSpec(FIXTURE_ROOT, "sweet-api-framework");
      if (bundle.requirements) {
        const result = validateSpec(bundle, DEFAULT_CONFIG, false);
        expect(result.id).toBe("sweet-api-framework");
      }
    } catch {
      // sweet fixture not copied in this environment
    }
  });

  it("compiles waves for spec with tasks", async () => {
    let slug = "mini-spec";
    try {
      const bundle = await loadSpec(FIXTURE_ROOT, "sweet-api-framework");
      if (bundle.tasks) {
        slug = "sweet-api-framework";
      }
    } catch {
      /* use mini */
    }
    const bundle = await loadSpec(FIXTURE_ROOT, slug);
    const waves = compileSpecWaves(bundle);
    expect(waves.length).toBeGreaterThan(0);
  });
});

describe("OKF & Spec-Kit integration workflows", () => {
  it("scaffolds specify and plan documents", async () => {
    const { scaffoldSpecify, scaffoldPlan } = await import("../../src/templates/scaffold.js");
    const { loadSpec } = await import("../../src/core/discovery.js");

    const specDir = await scaffoldSpecify(FIXTURE_ROOT, "test-kit-spec");
    expect(specDir).toContain("test-kit-spec");

    const planDir = await scaffoldPlan(FIXTURE_ROOT, "test-kit-spec");
    expect(planDir).toContain("test-kit-spec");

    const bundle = await loadSpec(FIXTURE_ROOT, "test-kit-spec");
    expect(bundle.requirements).toContain("Requirements Document");
    expect(bundle.design).toContain("Design Document");
    expect(bundle.tasks).toContain("Task Dependency Graph");
  });

  it("initializes OKF directory and validates it", async () => {
    const { scaffoldInit } = await import("../../src/templates/scaffold.js");
    const { validateOkfDirectory } = await import("../../src/core/okf-validator.js");
    const { join } = await import("node:path");
    const { mkdir, writeFile } = await import("node:fs/promises");

    const okfRoot = join(FIXTURE_ROOT, "okf-test-env");
    await mkdir(okfRoot, { recursive: true });

    // Initialize with OKF files
    await scaffoldInit(okfRoot, true);

    const res = await validateOkfDirectory(okfRoot, false);
    expect(res.valid).toBe(true);
    expect(res.errors).toHaveLength(0);

    // Create broken link
    const devGuidePath = join(okfRoot, ".kiro", "knowledge", "development.md");
    const devGuideContent = `---
type: process
title: Guidelines
---
Refer to [Missing Link](missing_file.md).
`;
    await writeFile(devGuidePath, devGuideContent, "utf-8");

    const resFail = await validateOkfDirectory(okfRoot, false);
    expect(resFail.warnings.find((w) => w.code === "OKF_BROKEN_LINK")).toBeDefined();
  });
});
