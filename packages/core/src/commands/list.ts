import { findWorkspace, listChanges, listSpecs } from "../core/discovery.js";

export async function runList(
  cwd: string,
  changes: boolean,
  specs: boolean,
  json: boolean,
): Promise<void> {
  const workspace = await findWorkspace(cwd);
  const out: { changes: string[]; specs: string[] } = {
    changes: changes || (!changes && !specs) ? await listChanges(workspace) : [],
    specs: specs || (!changes && !specs) ? await listSpecs(workspace) : [],
  };
  if (json) console.log(JSON.stringify(out, null, 2));
  else {
    if (out.changes.length) {
      console.log("Changes:");
      out.changes.forEach((c) => console.log(`  ${c}`));
    }
    if (out.specs.length) {
      console.log("Specs:");
      out.specs.forEach((s) => console.log(`  ${s}`));
    }
  }
}
