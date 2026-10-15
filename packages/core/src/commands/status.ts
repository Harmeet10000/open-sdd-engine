import { findWorkspace, listChanges, listSpecs, loadChange, loadSpec } from "../core/discovery.js";
import { getChangeStatus } from "../change/validators/change-validator.js";
import { getSpecStatus } from "../spec/validators/spec-validator.js";

export async function runStatus(
  cwd: string,
  id: string,
  type: "change" | "spec" | "auto",
  json: boolean,
): Promise<void> {
  const workspace = await findWorkspace(cwd);
  let out: unknown;

  if (type === "spec" || (type === "auto" && id.includes("-") === false)) {
    try {
      const bundle = await loadSpec(workspace, id);
      if (bundle.requirements !== undefined || type === "spec") {
        out = { loop: "spec", id, ...getSpecStatus(bundle) };
        if (json) console.log(JSON.stringify(out, null, 2));
        else printStatus(out as { loop: string; id: string; artifacts: Record<string, boolean>; phase: string });
        return;
      }
    } catch {
      /* fall through */
    }
  }

  const bundle = await loadChange(workspace, id);
  out = { loop: "change", id, ...getChangeStatus(bundle) };
  if (json) console.log(JSON.stringify(out, null, 2));
  else printStatus(out as { loop: string; id: string; artifacts: Record<string, boolean>; phase: string });
}

export async function runStatusSelection(opts: {
  cwd: string; id?: string; changes: boolean; specs: boolean; all: boolean; strict: boolean; json: boolean;
}): Promise<void> {
  if (opts.id) return runStatus(opts.cwd, opts.id, opts.specs && !opts.changes ? "spec" : opts.changes && !opts.specs ? "change" : "auto", opts.json);
  const workspace = await findWorkspace(opts.cwd);
  const includeChanges = opts.all || opts.changes || (!opts.changes && !opts.specs);
  const includeSpecs = opts.all || opts.specs || (!opts.changes && !opts.specs);
  const output: Array<Record<string, unknown>> = [];
  if (includeChanges) for (const id of await listChanges(workspace)) {
    const bundle = await loadChange(workspace, id);
    output.push({ loop: "change", id, ...getChangeStatus(bundle) });
  }
  if (includeSpecs) for (const id of await listSpecs(workspace)) {
    const bundle = await loadSpec(workspace, id);
    output.push({ loop: "spec", id, ...getSpecStatus(bundle) });
  }
  if (opts.json) console.log(JSON.stringify({ items: output, strict: opts.strict }, null, 2));
  else for (const item of output) printStatus(item as { loop: string; id: string; artifacts: Record<string, boolean>; phase: string });
}

function printStatus(s: { loop: string; id: string; artifacts: Record<string, boolean>; phase: string }): void {
  console.log(`${s.loop.toUpperCase()} ${s.id} — phase: ${s.phase}`);
  for (const [k, v] of Object.entries(s.artifacts)) {
    console.log(`  ${v ? "✓" : "○"} ${k}`);
  }
}
