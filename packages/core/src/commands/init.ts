import { findWorkspace } from "../core/discovery.js";
import { scaffoldInit, scaffoldCursorSkills } from "../templates/scaffold.js";

export async function runInit(cwd: string, tools?: string, okf?: boolean): Promise<void> {
  const workspace = await findWorkspace(cwd);
  await scaffoldInit(workspace, okf);
  if (tools?.includes("cursor")) {
    await scaffoldCursorSkills(workspace);
  }
  console.log(`Initialized .kiro/ in ${workspace}`);
}
