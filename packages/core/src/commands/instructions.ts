import { findWorkspace, listChanges, listSpecs, loadChange, loadSpec } from "../core/discovery.js";
import { getChangeInstructions } from "../change/validators/change-validator.js";
import { getSpecInstructions } from "../spec/validators/spec-validator.js";
import { loadState } from "../core/state.js";

export interface InstructionsOutput {
  loop: string;
  id: string;
  phase: string;
  nextArtifact: string | null;
  pendingTasks: string[];
  contextFiles: string[];
  research: string[];
  failuresCount?: number;
  lastStatus?: string;
  agentTips?: string[];
}

export async function runInstructionsSelection(opts: {
  cwd: string; id?: string; changes: boolean; specs: boolean; all: boolean; strict: boolean; json: boolean;
}): Promise<void> {
  if (opts.id) return runInstructions(opts.cwd, opts.id, opts.specs && !opts.changes ? "spec" : opts.changes && !opts.specs ? "change" : "auto", opts.json);
  const workspace = await findWorkspace(opts.cwd);
  const includeChanges = opts.all || opts.changes || (!opts.changes && !opts.specs);
  const includeSpecs = opts.all || opts.specs || (!opts.changes && !opts.specs);
  const items: InstructionsOutput[] = [];
  const collect = async (id: string, type: "change" | "spec") => {
    const original = console.log;
    let captured = "";
    console.log = (value?: unknown) => { captured = String(value ?? ""); };
    try { await runInstructions(workspace, id, type, true); }
    finally { console.log = original; }
    items.push(JSON.parse(captured) as InstructionsOutput);
  };
  if (includeChanges) for (const id of await listChanges(workspace)) await collect(id, "change");
  if (includeSpecs) for (const id of await listSpecs(workspace)) await collect(id, "spec");
  if (opts.json) console.log(JSON.stringify({ items, strict: opts.strict }, null, 2));
  else for (const item of items) printInstructions(item);
}

export async function runInstructions(
  cwd: string,
  id: string,
  type: "change" | "spec" | "auto",
  json: boolean,
): Promise<void> {
  const workspace = await findWorkspace(cwd);
  const state = await loadState(workspace);

  let rawOut: any;

  if (type === "spec") {
    const bundle = await loadSpec(workspace, id);
    rawOut = getSpecInstructions(bundle);
  } else if (type === "change") {
    const bundle = await loadChange(workspace, id);
    rawOut = getChangeInstructions(bundle);
  } else {
    // auto: try spec first if requirements exist
    const specBundle = await loadSpec(workspace, id);
    if (specBundle.requirements) {
      rawOut = getSpecInstructions(specBundle);
    } else {
      const changeBundle = await loadChange(workspace, id);
      rawOut = getChangeInstructions(changeBundle);
    }
  }

  // Enrich with state details
  const agentTips: string[] = [];
  if (state.failuresCount && state.failuresCount > 0) {
    agentTips.push(`You have encountered ${state.failuresCount} failures in the autoresearch loop.`);
    if (state.failuresCount >= 3) {
      agentTips.push(`WARNING: 3+ failures detected. Consider using 'sdd research web' or 'sdd research codebase' to verify your assumptions.`);
      agentTips.push(`Tip: Check if your changes need to be reverted using 'git checkout' to prevent compile error loops.`);
    }
  } else {
    agentTips.push(`Clean execution path. Proceed with the planned task sequence.`);
  }

  const out: InstructionsOutput = {
    ...rawOut,
    failuresCount: state.failuresCount ?? 0,
    lastStatus: state.lastStatus,
    agentTips,
  };

  if (json) {
    console.log(JSON.stringify(out, null, 2));
  } else {
    printInstructions(out);
  }
}

function printInstructions(out: InstructionsOutput): void {
  console.log(`Loop: ${out.loop} | ID: ${out.id} | Phase: ${out.phase}`);
  console.log(`Next artifact: ${out.nextArtifact ?? "(apply)"}`);
  console.log(`Context: ${out.contextFiles.join(", ")}`);
  if (out.pendingTasks.length) console.log(`Pending tasks: ${out.pendingTasks.join(", ")}`);
  
  if (out.failuresCount && out.failuresCount > 0) {
    console.log(`Failures Count: ${out.failuresCount} | Last Status: ${out.lastStatus ?? "unknown"}`);
  }

  console.log("Research tools:");
  for (const r of out.research) console.log(`  - ${r}`);

  if (out.agentTips && out.agentTips.length) {
    console.log("Agent Tips:");
    for (const t of out.agentTips) console.log(`  - ${t}`);
  }
}
