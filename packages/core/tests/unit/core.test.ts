import { describe, it, expect } from "vitest";
import { parseDeltaSpec } from "../../src/change/parsers/delta-spec.js";
import { classifyLoop } from "../../src/router/loop-classifier.js";
import { classifyEars, parseRequirements } from "../../src/spec/parsers/requirements.js";
import { compileWaves } from "../../src/spec/compiler/dag-waves.js";
import { parseTasks } from "../../src/spec/parsers/tasks.js";
import { parseOkf, validateOkfFile } from "../../src/core/okf.js";
import { loadState, updateState } from "../../src/core/state.js";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { mkdir, rm } from "node:fs/promises";

describe("delta-spec parser", () => {
  it("parses ADDED requirements with scenarios", () => {
    const content = `## ADDED Requirements
 
### Requirement: Token refresh
The system SHALL refresh tokens.
 
#### Scenario: Expired token
- **WHEN** token expires
- **THEN** refresh succeeds
`;
    const plan = parseDeltaSpec(content);
    expect(plan.added).toHaveLength(1);
    expect(plan.added[0]!.scenarios).toHaveLength(1);
  });
});

describe("loop classifier", () => {
  it("recommends change for refactors", () => {
    const r = classifyLoop("refactor middleware error handling");
    expect(r.recommended).toBe("change");
  });

  it("recommends spec for HFT API", () => {
    const r = classifyLoop("design sub-millisecond HTTP server for HFT quant devs");
    expect(r.recommended).toBe("spec");
  });
});

describe("EARS classifier", () => {
  it("classifies THE SYSTEM SHALL", () => {
    expect(classifyEars("THE SYSTEM SHALL parse requests")).toBe("ubiquitous");
  });

  it("classifies WHEN THEN SHALL", () => {
    expect(classifyEars("WHEN the app starts THEN Bun SHALL replace Node")).toBe("event_driven");
  });

  it("classifies optional-feature and canonical complex EARS", () => {
    expect(classifyEars("WHERE OAuth is enabled, the auth service SHALL validate callbacks")).toBe("optional_feature");
    expect(classifyEars("WHILE a websocket is active, WHEN idle timeout expires, the websocket layer SHALL close it")).toBe("complex");
  });
});

describe("requirements parser", () => {
  it("builds requirement IDs", () => {
    const content = `### Requirement 1: Auth
 
#### Acceptance Criteria
 
1. WHEN user logs in, THE SYSTEM SHALL issue a token
2. THE SYSTEM SHALL expire sessions
`;
    const reqs = parseRequirements(content);
    expect(reqs[0]!.criteria.map((c) => c.id)).toEqual(["1.1", "1.2"]);
  });
});

describe("DAG compiler", () => {
  it("compiles waves from dependencies", () => {
    const adj = new Map([
      ["1", []],
      ["2", []],
      ["3", ["1"]],
      ["4", ["1", "2"]],
    ]);
    const waves = compileWaves(["1", "2", "3", "4"], adj);
    expect(waves[0]).toContain("1");
    expect(waves[0]).toContain("2");
    expect(waves.length).toBeGreaterThan(1);
  });
});

describe("task parser", () => {
  it("parses Kiro checkbox format", () => {
    const content = `- [ ] 1. Setup
  - _Requirements: 1.1_
- [ ] 2. Implement
  - [ ] 2.1 Subtask
`;
    const { tasks } = parseTasks(content);
    expect(tasks.find((t) => t.id === "1")).toBeTruthy();
    expect(tasks.find((t) => t.id === "2.1")?.parentId).toBe("2");
  });
});

describe("Google OKF parser", () => {
  it("parses YAML frontmatter and body and extracts markdown links", () => {
    const content = `---
type: guide
title: Testing Guide
tags: [unit, integration]
---
# Testing

Refer to [Development guidelines](development.md) and [System Architecture](architecture.md#flow).
Do not match external links like [Google](https://google.com).
`;
    const { meta, body, links } = parseOkf(content);
    expect(meta).not.toBeNull();
    expect(meta!.type).toBe("guide");
    expect(meta!.title).toBe("Testing Guide");
    expect(body).toContain("# Testing");
    expect(links).toContain("development.md");
    expect(links).toContain("architecture.md");
    expect(links).not.toContain("https://google.com");
  });

  it("fails validation if type or title are missing", async () => {
    const invalidContent = `---
tags: [tag1]
---
# No Type or Title
`;
    const res = await validateOkfFile("test.md", invalidContent);
    expect(res.valid).toBe(false);
    expect(res.errors.length).toBeGreaterThanOrEqual(2);
  });
});

describe("Loop state management", () => {
  it("loads default state and handles updates", async () => {
    const testDir = join(tmpdir(), "sdd-state-test-" + Date.now());
    await mkdir(testDir, { recursive: true });
    await mkdir(join(testDir, ".kiro"), { recursive: true });

    const state = await loadState(testDir);
    expect(state.failuresCount).toBe(0);

    const updated = await updateState(testDir, {
      activeId: "my-change",
      failuresCount: 3,
    });
    expect(updated.activeId).toBe("my-change");
    expect(updated.failuresCount).toBe(3);

    await rm(testDir, { recursive: true, force: true });
  });
});
