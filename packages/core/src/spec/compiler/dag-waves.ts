import type { TaskDependencyGraph } from "../../core/types.js";

export class CycleError extends Error {
  constructor(public involved: string[]) {
    super(`Cycle detected involving: ${involved.join(", ")}`);
  }
}

export function buildAdjacency(
  taskIds: Set<string>,
  explicit: TaskDependencyGraph | null,
  inferred: Map<string, string[]>,
): Map<string, string[]> {
  const adj = new Map<string, string[]>();
  for (const id of taskIds) adj.set(id, []);

  if (explicit?.tasks?.length) {
    for (const t of explicit.tasks) {
      if (!taskIds.has(t.id)) continue;
      adj.set(
        t.id,
        t.depends_on.filter((d) => taskIds.has(d)),
      );
    }
  } else {
    for (const [id, deps] of inferred) {
      if (taskIds.has(id)) adj.set(id, deps.filter((d) => taskIds.has(d)));
    }
  }
  return adj;
}

export function compileWaves(taskIds: string[], adj: Map<string, string[]>): string[][] {
  const remaining = new Set(taskIds);
  const waves: string[][] = [];
  const inDegree = (id: string) => (adj.get(id) ?? []).filter((d) => remaining.has(d)).length;

  while (remaining.size > 0) {
    const wave = [...remaining].filter((id) => inDegree(id) === 0);
    if (wave.length === 0) {
      throw new CycleError([...remaining]);
    }
    waves.push(wave.sort());
    for (const id of wave) remaining.delete(id);
  }
  return waves;
}

export function compileWavesFromGraph(
  taskIds: Set<string>,
  explicit: TaskDependencyGraph | null,
  inferred: Map<string, string[]>,
): string[][] {
  return compileWaves([...taskIds], buildAdjacency(taskIds, explicit, inferred));
}
