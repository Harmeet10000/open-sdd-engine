import { readdir, stat, access } from "node:fs/promises";
import { constants } from "node:fs";
import { join } from "node:path";

export interface ChangeBundle {
  id: string;
  dir: string;
  proposal?: string;
  design?: string;
  tasks?: string;
  deltaSpecPaths: string[];
}

export interface SpecBundle {
  slug: string;
  dir: string;
  requirements?: string;
  design?: string;
  tasks?: string;
  behavior?: string;
  meta?: string;
}

async function exists(path: string): Promise<boolean> {
  try {
    await access(path, constants.R_OK);
    return true;
  } catch {
    return false;
  }
}

export async function readFileIfExists(path: string): Promise<string | undefined> {
  try {
    const { readFile } = await import("node:fs/promises");
    return await readFile(path, "utf-8");
  } catch {
    return undefined;
  }
}

export function resolveKiroRoot(workspace: string): string {
  return join(workspace, ".kiro");
}

export async function listChanges(workspace: string): Promise<string[]> {
  const dir = join(workspace, ".kiro", "changes");
  if (!(await exists(dir))) return [];
  const entries = await readdir(dir, { withFileTypes: true });
  return entries
    .filter((e) => e.isDirectory() && e.name !== "archive")
    .map((e) => e.name);
}

export async function listSpecs(workspace: string): Promise<string[]> {
  const dir = join(workspace, ".kiro", "specs");
  if (!(await exists(dir))) return [];
  const entries = await readdir(dir, { withFileTypes: true });
  return entries.filter((e) => e.isDirectory()).map((e) => e.name);
}

async function collectDeltaSpecs(changeDir: string): Promise<string[]> {
  const specsDir = join(changeDir, "specs");
  if (!(await exists(specsDir))) return [];
  const paths: string[] = [];
  async function walk(dir: string) {
    const entries = await readdir(dir, { withFileTypes: true });
    for (const e of entries) {
      const p = join(dir, e.name);
      if (e.isDirectory()) await walk(p);
      else if (e.name.endsWith(".md")) paths.push(p);
    }
  }
  await walk(specsDir);
  return paths;
}

export async function loadChange(workspace: string, id: string): Promise<ChangeBundle> {
  const dir = join(workspace, ".kiro", "changes", id);
  return {
    id,
    dir,
    proposal: await readFileIfExists(join(dir, "proposal.md")),
    design: await readFileIfExists(join(dir, "design.md")),
    tasks: await readFileIfExists(join(dir, "tasks.md")),
    deltaSpecPaths: await collectDeltaSpecs(dir),
  };
}

export async function loadSpec(workspace: string, slug: string): Promise<SpecBundle> {
  const dir = join(workspace, ".kiro", "specs", slug);
  return {
    slug,
    dir,
    requirements: await readFileIfExists(join(dir, "requirements.md")),
    design: await readFileIfExists(join(dir, "design.md")),
    tasks: await readFileIfExists(join(dir, "tasks.md")),
    behavior: await readFileIfExists(join(dir, "behavior.feature")),
    meta: await readFileIfExists(join(dir, "spec.meta.json")),
  };
}

export async function findWorkspace(start: string): Promise<string> {
  let current = start;
  while (true) {
    if (await exists(join(current, ".kiro"))) return current;
    const parent = join(current, "..");
    if (parent === current) return start;
    current = parent;
  }
}

export async function isDirectory(path: string): Promise<boolean> {
  try {
    return (await stat(path)).isDirectory();
  } catch {
    return false;
  }
}
