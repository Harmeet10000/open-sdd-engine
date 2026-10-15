import { readFile, writeFile, mkdir, rename } from "node:fs/promises";
import { join, dirname } from "node:path";
import { parseDeltaSpec } from "../parsers/delta-spec.js";
import { readFileIfExists } from "../../core/discovery.js";

const REQUIREMENT_HEADER = /^###\s*Requirement:\s*(.+)$/im;

function normalizeName(name: string): string {
  return name.trim().toLowerCase();
}

function mergeRequirementBlocks(existing: string, added: string[], modified: string[]): string {
  let result = existing.trim();
  const blocks = splitRequirements(result);
  const blockMap = new Map(blocks.map((b) => [normalizeName(b.name), b.raw]));

  for (const mod of modified) {
    const nameMatch = mod.match(REQUIREMENT_HEADER);
    if (nameMatch) blockMap.set(normalizeName(nameMatch[1]!), mod.trim());
  }
  for (const add of added) {
    const nameMatch = add.match(REQUIREMENT_HEADER);
    if (nameMatch && !blockMap.has(normalizeName(nameMatch[1]!))) {
      blockMap.set(normalizeName(nameMatch[1]!), add.trim());
    }
  }

  return `# Specification\n\n${[...blockMap.values()].join("\n\n")}\n`;
}

function splitRequirements(content: string): Array<{ name: string; raw: string }> {
  const blocks: Array<{ name: string; raw: string }> = [];
  const parts = content.split(/(?=^###\s*Requirement:)/m);
  for (const part of parts) {
    const m = part.match(REQUIREMENT_HEADER);
    if (m && part.trim()) blocks.push({ name: m[1]!.trim(), raw: part.trim() });
  }
  return blocks;
}

function removeRequirements(existing: string, removedNames: string[]): string {
  const removeSet = new Set(removedNames.map(normalizeName));
  const blocks = splitRequirements(existing).filter((b) => !removeSet.has(normalizeName(b.name)));
  if (blocks.length === 0) return "# Specification\n\n";
  return `# Specification\n\n${blocks.map((b) => b.raw).join("\n\n")}\n`;
}

function applyRenames(existing: string, renamed: Array<{ from: string; to: string }>): string {
  let result = existing;
  for (const { from, to } of renamed) {
    result = result.replace(
      new RegExp(`(###\\s*Requirement:\\s*)${escapeRegex(from)}`, "gi"),
      `$1${to}`,
    );
  }
  return result;
}

function escapeRegex(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

export async function archiveChange(
  workspace: string,
  changeId: string,
  skipSpecs: boolean,
): Promise<{ archivedTo: string }> {
  const changeDir = join(workspace, ".kiro", "changes", changeId);
  const date = new Date().toISOString().slice(0, 10);
  const archiveDir = join(workspace, ".kiro", "changes", "archive", `${date}-${changeId}`);

  if (!skipSpecs) {
    const specsDir = join(changeDir, "specs");
    const { readdir } = await import("node:fs/promises");
    try {
      const domains = await readdir(specsDir, { withFileTypes: true });
      for (const d of domains) {
        if (!d.isDirectory()) continue;
        const deltaPath = join(specsDir, d.name, "spec.md");
        const deltaContent = await readFileIfExists(deltaPath);
        if (!deltaContent) continue;

        const plan = parseDeltaSpec(deltaContent);
        const domainPath = join(workspace, ".kiro", "domains", d.name, "spec.md");
        await mkdir(dirname(domainPath), { recursive: true });

        let existing = (await readFileIfExists(domainPath)) ?? "# Specification\n\n";
        existing = applyRenames(existing, plan.renamed);
        existing = removeRequirements(existing, plan.removed);

        const addedBlocks = plan.added.map((b) => `### Requirement: ${b.name}\n${b.content}`);
        const modifiedBlocks = plan.modified.map((b) => `### Requirement: ${b.name}\n${b.content}`);
        existing = mergeRequirementBlocks(existing, addedBlocks, modifiedBlocks);

        await writeFile(domainPath, existing, "utf-8");
      }
    } catch {
      // no specs dir
    }
  }

  await mkdir(dirname(archiveDir), { recursive: true });
  await rename(changeDir, archiveDir);
  return { archivedTo: archiveDir };
}
