import { describe, expect, it } from "vitest";
import { validateBdd } from "../../src/spec/validators/bdd-validator.js";
import { openClarifications } from "../../src/commands/workflow.js";
import { createRoleHandoff, hashArtifact, verifyWorkflow, WorkflowSchema } from "../../src/workflow/engine.js";
import { recall, remember } from "../../src/memory/cognee.js";
import { DEFAULT_CONFIG } from "../../src/core/types.js";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { rm } from "node:fs/promises";

describe("BDD and deterministic workflow controls", () => {
  it("rejects incomplete BDD scenarios", () => {
    const result = validateBdd("Feature: Login\nScenario: invalid\n Given a user", true);
    expect(result.errors.map((item) => item.code)).toContain("BDD_INCOMPLETE_SCENARIO");
  });

  it("persists and recalls local workflow memory when Cognee is disabled", async () => {
    const workspace = join(tmpdir(), `sdd-memory-${Date.now()}`);
    await remember(workspace, DEFAULT_CONFIG, { scope: "decision", text: "Auth tokens expire after 24 hours", tags: ["auth"], source: "test" });
    await expect(recall(workspace, DEFAULT_CONFIG, "auth tokens")).resolves.toMatchObject({ provider: "local", records: [expect.objectContaining({ scope: "decision" })] });
    await rm(workspace, { recursive: true, force: true });
  });

  it("finds spec ambiguities before finalization", () => {
    expect(openClarifications("The system SHALL somehow handle login TODO")).toHaveLength(1);
  });

  it("locks a workflow to its source artifact", () => {
    const source = "The system SHALL work";
    const workflow = WorkflowSchema.parse({ version: 1, id: "login", loop: "spec", source, specHash: hashArtifact(source), locked: true, roles: { planner: [], implementer: [], reviewer: ["bdd-validation"] }, phases: ["gather", "clarify", "plan", "implement", "review", "verify"] });
    expect(verifyWorkflow(workflow, source)).toBe(true);
    expect(verifyWorkflow(workflow, `${source}!`)).toBe(false);
    expect(createRoleHandoff(workflow, "reviewer").allowedSkills).toContain("bdd-validation");
  });
});
