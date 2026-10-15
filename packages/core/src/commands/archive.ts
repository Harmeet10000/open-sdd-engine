import { findWorkspace } from "../core/discovery.js";
import { archiveChange } from "../change/archive/merge-deltas.js";

export async function runArchive(
  cwd: string,
  changeId: string,
  yes: boolean,
  skipSpecs: boolean,
): Promise<void> {
  if (!yes) {
    console.log(`Archive change "${changeId}"? Use --yes to confirm.`);
    process.exit(2);
  }
  const workspace = await findWorkspace(cwd);
  const { archivedTo } = await archiveChange(workspace, changeId, skipSpecs);
  console.log(`Archived to ${archivedTo}`);
}
