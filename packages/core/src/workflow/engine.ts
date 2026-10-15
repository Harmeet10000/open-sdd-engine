import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { z } from "zod";

export const WorkflowSchema = z.object({
  version: z.literal(1), id: z.string().min(1), loop: z.enum(["change", "spec", "custom"]),
  specHash: z.string().length(64), locked: z.boolean(),
  source: z.string().min(1),
  roles: z.object({ planner: z.array(z.string()), implementer: z.array(z.string()), reviewer: z.array(z.string()) }),
  phases: z.array(z.string().min(1)).min(1),
});
export type Workflow = z.infer<typeof WorkflowSchema>;

const roleSkills = {
  planner: ["codegraph", "graphify", "ast-grep", "context7", "firecrawl"],
  implementer: ["ast-grep", "codegraph", "ponytail"],
  reviewer: ["ast-grep", "codegraph", "bdd-validation", "sdd-schema-validation"],
};
export function hashArtifact(content: string): string { return createHash("sha256").update(content).digest("hex"); }
export async function generateWorkflow(workspace: string, id: string, loop: "change" | "spec" | "custom", source: string, phases?: string[]): Promise<Workflow> {
  const workflow = WorkflowSchema.parse({ version: 1, id, loop, source, specHash: hashArtifact(source), locked: true, roles: roleSkills, phases: phases?.length ? phases : ["gather", "clarify", "plan", "implement", "review", "verify"] });
  const dir = join(workspace, ".kiro", "workflows"); await mkdir(dir, { recursive: true });
  await writeFile(join(dir, `${id}.json`), JSON.stringify(workflow, null, 2), "utf-8");
  return workflow;
}
export async function loadWorkflow(workspace: string, id: string): Promise<Workflow> { return WorkflowSchema.parse(JSON.parse(await readFile(join(workspace, ".kiro", "workflows", `${id}.json`), "utf-8"))); }
export function verifyWorkflow(workflow: Workflow, source: string): boolean { return workflow.locked && workflow.specHash === hashArtifact(source); }
export function createRoleHandoff(workflow: Workflow, role: "planner" | "implementer" | "reviewer") {
  const duties = {
    planner: "Gather evidence, resolve the workflow's clarification record, and produce a plan without changing implementation files.",
    implementer: "Implement only approved tasks. Preserve the workflow contract and attach verification evidence to each completed task.",
    reviewer: "Independently validate SDD schema, EARS, BDD, traceability, tests, and workflow integrity. Do not implement fixes in the review handoff.",
  };
  return { workflowId: workflow.id, role, allowedSkills: workflow.roles[role], phaseOrder: workflow.phases, contract: duties[role], deterministic: workflow.locked };
}
