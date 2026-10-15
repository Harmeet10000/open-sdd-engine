import { findWorkspace } from "../core/discovery.js";
import { scaffoldChange, scaffoldSpec } from "../templates/scaffold.js";

export async function runNewChange(cwd: string, id: string, json: boolean): Promise<void> {
  const workspace = await findWorkspace(cwd);
  const dir = await scaffoldChange(workspace, id);
  const out = { loop: "change", id, path: dir };
  if (json) console.log(JSON.stringify(out, null, 2));
  else console.log(`Created change loop: ${dir}`);
}

export async function runNewSpec(cwd: string, slug: string, json: boolean): Promise<void> {
  const workspace = await findWorkspace(cwd);
  const dir = await scaffoldSpec(workspace, slug);
  const out = { loop: "spec", id: slug, path: dir };
  if (json) console.log(JSON.stringify(out, null, 2));
  else console.log(`Created spec loop: ${dir}`);
}
