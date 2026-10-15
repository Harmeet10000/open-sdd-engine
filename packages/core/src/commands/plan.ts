import { findWorkspace } from "../core/discovery.js";
import { scaffoldPlan } from "../templates/scaffold.js";

export async function runPlan(cwd: string, slug: string, json: boolean): Promise<void> {
  const workspace = await findWorkspace(cwd);
  const dir = await scaffoldPlan(workspace, slug);
  const out = { command: "plan", slug, path: dir };
  if (json) {
    console.log(JSON.stringify(out, null, 2));
  } else {
    console.log(`Created plan design document: ${dir}`);
  }
}
