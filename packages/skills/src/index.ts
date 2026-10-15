import { z } from "zod";
import { readdir, readFile } from "node:fs/promises";
import { join } from "node:path";

export const SkillDecisionSchema = z.object({
  id: z.string(),
  mode: z.enum(["required", "recommended"]),
  reason: z.string(),
  trust: z.literal("builtin"),
});
export type SkillDecision = z.infer<typeof SkillDecisionSchema>;

const rules: Array<{ id: string; mode: SkillDecision["mode"]; test: RegExp; reason: string }> = [
  { id: "security-review", mode: "required", test: /auth|permission|credential|secret|payment|security/i, reason: "The request changes a trust boundary or protected data." },
  { id: "migration", mode: "required", test: /migrat|schema|database|backfill|rename/i, reason: "The request may make persistent data incompatible." },
  { id: "research", mode: "recommended", test: /unknown|library|sdk|api|benchmark|research/i, reason: "External facts or repository evidence are likely needed before planning." },
  { id: "ponytail", mode: "recommended", test: /refactor|small|simple|cleanup|fix|bug/i, reason: "Prefer the smallest diff that reuses existing code and dependencies." },
  { id: "spec-analysis", mode: "required", test: /new system|architecture|platform|cross.?cutting|distributed/i, reason: "The scope crosses components and needs a traceable spec loop." },
];

export function selectSkills(description: string): SkillDecision[] {
  const selected = rules.filter((rule) => rule.test.test(description)).map(({ id, mode, reason }) => SkillDecisionSchema.parse({ id, mode, reason, trust: "builtin" }));
  if (!selected.some((skill) => skill.id === "ponytail")) {
    selected.push(SkillDecisionSchema.parse({ id: "ponytail", mode: "recommended", reason: "Default to reuse and the smallest working change.", trust: "builtin" }));
  }
  return selected;
}

export interface DiscoveredSkill { id: string; path: string; description?: string; }
export async function discoverSkills(workspace: string): Promise<DiscoveredSkill[]> {
  const roots = [join(workspace, ".kiro", "skills"), join(workspace, ".agents", "skills"), join(workspace, ".cursor", "skills")];
  const found: DiscoveredSkill[] = [];
  for (const root of roots) {
    let entries: Array<{ isDirectory(): boolean; name: string }>;
    try { entries = await readdir(root, { withFileTypes: true, encoding: "utf8" }) as Array<{ isDirectory(): boolean; name: string }>; } catch { continue; }
    for (const entry of entries) {
      if (!entry.isDirectory()) continue;
      const path = join(root, entry.name, "SKILL.md");
      try {
        const content = await readFile(path, "utf-8");
        const description = content.match(/^description:\s*(.+)$/m)?.[1]?.trim();
        found.push({ id: entry.name, path, description });
      } catch { /* a directory is not a skill without SKILL.md */ }
    }
  }
  return found.sort((a, b) => a.id.localeCompare(b.id));
}
