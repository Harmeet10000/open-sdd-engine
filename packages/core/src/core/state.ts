import { readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";

export interface SddState {
  activeId?: string;
  loopType?: "change" | "spec";
  phase?: string;
  lastEvaluated?: string;
  lastStatus?: "passed" | "failed";
  failuresCount?: number;
  completedTasks?: string[];
  history?: Array<{
    timestamp: string;
    event: string;
    detail?: string;
  }>;
}

export async function loadState(workspace: string): Promise<SddState> {
  const statePath = join(workspace, ".kiro", "state.json");
  try {
    const raw = await readFile(statePath, "utf-8");
    return JSON.parse(raw) as SddState;
  } catch {
    return {
      failuresCount: 0,
      completedTasks: [],
      history: [],
    };
  }
}

export async function saveState(workspace: string, state: SddState): Promise<void> {
  const statePath = join(workspace, ".kiro", "state.json");
  await writeFile(statePath, JSON.stringify(state, null, 2), "utf-8");
}

export async function updateState(workspace: string, update: Partial<SddState>): Promise<SddState> {
  const current = await loadState(workspace);
  const updated: SddState = {
    ...current,
    ...update,
    completedTasks: update.completedTasks ?? current.completedTasks ?? [],
    history: [
      ...(current.history ?? []),
      {
        timestamp: new Date().toISOString(),
        event: "state_update",
        detail: JSON.stringify(update),
      },
    ].slice(-50), // Limit history size
  };
  await saveState(workspace, updated);
  return updated;
}
