import { findWorkspace } from "../core/discovery.js";
import { scaffoldSpecify } from "../templates/scaffold.js";

export async function runSpecify(cwd: string, slug: string, json: boolean): Promise<void> {
  const workspace = await findWorkspace(cwd);
  const dir = await scaffoldSpecify(workspace, slug);
  const out = { command: "specify", slug, path: dir };
  if (json) {
    console.log(JSON.stringify(out, null, 2));
  } else {
    console.log(`Created specify requirements document: ${dir}`);
  }
}
