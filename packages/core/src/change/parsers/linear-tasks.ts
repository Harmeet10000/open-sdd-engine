export interface LinearTask {
  id: string;
  description: string;
  done: boolean;
  section?: string;
}

export function parseLinearTasks(content: string): LinearTask[] {
  const tasks: LinearTask[] = [];
  let currentSection: string | undefined;

  for (const line of content.split("\n")) {
    const sectionMatch = line.match(/^##\s+(.+)$/);
    if (sectionMatch) {
      currentSection = sectionMatch[1]!.trim();
      continue;
    }
    const taskMatch = line.match(/^-\s+\[([ xX])\]\s+(\d+(?:\.\d+)*)\.?\s+(.+)$/);
    if (taskMatch) {
      tasks.push({
        id: taskMatch[2]!,
        description: taskMatch[3]!.trim(),
        done: taskMatch[1]!.toLowerCase() === "x",
        section: currentSection,
      });
    }
  }
  return tasks;
}

export function countTopLevelSections(content: string): number {
  return (content.match(/^##\s+/gm) ?? []).length;
}
