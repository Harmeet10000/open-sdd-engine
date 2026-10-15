import { readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { findWorkspace, loadChange, loadSpec } from "../core/discovery.js";
import { createRoleHandoff, generateWorkflow, loadWorkflow, verifyWorkflow } from "../workflow/engine.js";
import { loadProfile } from "../git/operations.js";
import { loadConfig } from "../core/config.js";
import { recall } from "../memory/cognee.js";

export async function runWorkflowGenerate(cwd: string, id: string, loop: "change" | "spec" | "custom", json: boolean, description?: string, phases?: string[]): Promise<void> {
  const workspace = await findWorkspace(cwd);
  const source = loop === "custom" ? description : loop === "spec" ? (await loadSpec(workspace, id)).requirements : (await loadChange(workspace, id)).proposal;
  if (!source) throw new Error(`Cannot generate workflow: missing ${loop === "spec" ? "requirements.md" : "proposal.md"}.`);
  if (loop === "spec") {
    const questions = openClarifications(source);
    if (questions.length) {
      try {
        const record = JSON.parse(await readFile(join(workspace, ".kiro", "specs", id, "clarifications.json"), "utf-8")) as { finalized?: boolean };
        if (!record.finalized) throw new Error(`Clarification gate is open. Run: sdd clarify ${id}`);
      } catch (error) {
        if (error instanceof Error && error.message.startsWith("Clarification gate")) throw error;
        throw new Error(`Clarification gate is open. Run: sdd clarify ${id}`);
      }
    }
  }
  const output = await generateWorkflow(workspace, id, loop, source, phases);
  console.log(json ? JSON.stringify(output, null, 2) : `Generated locked ${loop} workflow: .kiro/workflows/${id}.json`);
}

export async function runWorkflowStatus(cwd: string, id: string, json: boolean): Promise<void> {
  const workspace = await findWorkspace(cwd); const workflow = await loadWorkflow(workspace, id);
  const source = workflow.loop === "custom" ? workflow.source : workflow.loop === "spec" ? (await loadSpec(workspace, id)).requirements : (await loadChange(workspace, id)).proposal;
  const output = { ...workflow, deterministic: !!source && verifyWorkflow(workflow, source) };
  console.log(json ? JSON.stringify(output, null, 2) : `${id}: ${output.deterministic ? "locked and deterministic" : "spec changed, regenerate workflow"}`);
}

export async function runWorkflowHandoff(cwd: string, id: string, role: "planner" | "implementer" | "reviewer", json: boolean, discoveredSkills: string[] = []): Promise<void> {
  const workspace = await findWorkspace(cwd); const workflow = await loadWorkflow(workspace, id);
  const source = workflow.loop === "custom" ? workflow.source : workflow.loop === "spec" ? (await loadSpec(workspace, id)).requirements : (await loadChange(workspace, id)).proposal;
  if (!source || !verifyWorkflow(workflow, source)) throw new Error("Workflow source changed. Regenerate before handing work to a subagent.");
  const memory = await recall(workspace, await loadConfig(workspace), `${workflow.id} ${role}`);
  const handoff = { ...createRoleHandoff(workflow, role), discoveredSkills, codingProfile: await loadProfile(workspace), memory };
  console.log(json ? JSON.stringify(handoff, null, 2) : `${role}: ${handoff.contract}\nAllowed skills: ${handoff.allowedSkills.join(", ")}`);
}

export function openClarifications(source: string): string[] {
  const questions: string[] = [];
  for (const [index, line] of source.split(/\r?\n/).entries()) if (/\b(TODO|TBD|TBA|somehow|appropriate|etc\.?|as needed)\b/i.test(line)) questions.push(`Line ${index + 1}: What precise behavior replaces "${line.trim()}"?`);
  if (!/\b(SHALL|MUST)\b/i.test(source)) questions.push("What observable behavior must the system guarantee?");
  return questions;
}

export async function runClarify(cwd: string, id: string, answers: string[], json: boolean): Promise<void> {
  const workspace = await findWorkspace(cwd); const spec = await loadSpec(workspace, id);
  if (!spec.requirements) throw new Error("Missing requirements.md");
  const questions = openClarifications(spec.requirements); const unresolved = questions.slice(answers.length);
  const path = join(workspace, ".kiro", "specs", id, "clarifications.json");
  const output = { id, questions, answers, unresolved, finalized: unresolved.length === 0 };
  await writeFile(path, JSON.stringify(output, null, 2), "utf-8");
  console.log(json ? JSON.stringify(output, null, 2) : unresolved.length ? unresolved.join("\n") : "All clarification questions answered. Run `sdd workflow generate <id> --type spec`.");
}
