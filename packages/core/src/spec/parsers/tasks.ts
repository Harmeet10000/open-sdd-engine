import type { TaskDependencyGraph } from "../../core/types.js";

export interface TaskNode {
  id: string;
  description: string;
  done: boolean;
  optional: boolean;
  depth: number;
  parentId?: string;
  requirementRefs: string[];
  line: number;
}

const TASK_LINE = /^(\s*)-\s+\[([ xX\*])\]\s+(\d+(?:\.\d+)*)\.?\s*(.*)$/;
const REQ_REF = /_Requirements:\s*([\d.,\s]+)_/;

export function parseTasks(content: string): {
  tasks: TaskNode[];
  dependencyGraph: TaskDependencyGraph | null;
} {
  const tasks: TaskNode[] = [];
  const lines = content.split("\n");
  let pendingReqRefs: string[] = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]!;
    const reqMatch = line.match(REQ_REF);
    if (reqMatch) {
      pendingReqRefs = reqMatch[1]!.split(",").map((s) => s.trim()).filter(Boolean);
      continue;
    }

    const m = line.match(TASK_LINE);
    if (m) {
      const indent = m[1]!.length;
      const depth = indent >= 4 ? 2 : indent >= 2 ? 1 : 0;
      const check = m[2]!;
      const id = m[3]!;
      const description = m[4]?.trim() || id;
      const parts = id.split(".");
      const parentId = parts.length > 1 ? parts.slice(0, -1).join(".") : undefined;

      tasks.push({
        id,
        description,
        done: check.toLowerCase() === "x",
        optional: check === "*",
        depth,
        parentId,
        requirementRefs: [...pendingReqRefs],
        line: i + 1,
      });
      pendingReqRefs = [];
    }
  }

  const graphSection = content.match(/##\s*Task Dependency Graph\s*\n```json\s*\n([\s\S]*?)\n```/);
  let dependencyGraph: TaskDependencyGraph | null = null;
  if (graphSection) {
    try {
      dependencyGraph = JSON.parse(graphSection[1]!) as TaskDependencyGraph;
    } catch {
      dependencyGraph = null;
    }
  }

  return { tasks, dependencyGraph };
}

export function inferDependencies(tasks: TaskNode[]): Map<string, string[]> {
  const deps = new Map<string, string[]>();
  for (const t of tasks) {
    const list: string[] = [];
    if (t.parentId) list.push(t.parentId);
    deps.set(t.id, list);
  }
  return deps;
}

export function countTopLevelTasks(tasks: TaskNode[]): number {
  return tasks.filter((t) => !t.id.includes(".")).length;
}
