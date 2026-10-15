import { describe, expect, it } from "vitest";
import { discoverSkills, selectSkills } from "../src/index.js";
import { mkdir, writeFile, rm } from "node:fs/promises";
import { join } from "node:path";
import { tmpdir } from "node:os";

describe("skill selection", () => {
  it("requires security review for trust-boundary work", () => {
    expect(selectSkills("add payment credentials to the auth flow")).toEqual(expect.arrayContaining([
      expect.objectContaining({ id: "security-review", mode: "required", trust: "builtin" }),
    ]));
  });

  it("uses Ponytail as a safe default", () => {
    expect(selectSkills("rename an internal helper")).toEqual(expect.arrayContaining([
      expect.objectContaining({ id: "ponytail" }),
    ]));
  });

  it("discovers project-local skills without executing them", async () => {
    const root = join(tmpdir(), `sdd-skill-${Date.now()}`);
    await mkdir(join(root, ".kiro", "skills", "research"), { recursive: true });
    await writeFile(join(root, ".kiro", "skills", "research", "SKILL.md"), "---\ndescription: Gather evidence\n---\n");
    await expect(discoverSkills(root)).resolves.toEqual([expect.objectContaining({ id: "research" })]);
    await rm(root, { recursive: true, force: true });
  });
});
