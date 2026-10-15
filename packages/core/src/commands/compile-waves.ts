import { findWorkspace, loadSpec } from "../core/discovery.js";
import { compileSpecWaves } from "../spec/validators/spec-validator.js";

export async function runCompileWaves(cwd: string, slug: string, json: boolean): Promise<void> {
  const workspace = await findWorkspace(cwd);
  const bundle = await loadSpec(workspace, slug);
  const waves = compileSpecWaves(bundle);
  const out = {
    specSlug: slug,
    waves,
    parallelism: waves.map((w) => w.length),
  };
  if (json) console.log(JSON.stringify(out, null, 2));
  else {
    waves.forEach((w, i) => console.log(`Wave ${i}: ${w.join(", ")}`));
  }
}
