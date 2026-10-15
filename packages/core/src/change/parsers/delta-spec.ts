export interface RequirementBlock {
  name: string;
  content: string;
  scenarios: string[];
}

export interface DeltaPlan {
  added: RequirementBlock[];
  modified: RequirementBlock[];
  removed: string[];
  renamed: Array<{ from: string; to: string }>;
  sectionPresence: {
    added: boolean;
    modified: boolean;
    removed: boolean;
    renamed: boolean;
  };
}

const REQUIREMENT_HEADER = /^###\s*Requirement:\s*(.+)$/i;

function normalizeLineEndings(s: string): string {
  return s.replace(/\r\n/g, "\n");
}

function splitTopLevelSections(content: string): Map<string, string> {
  const sections = new Map<string, string>();
  const lines = normalizeLineEndings(content).split("\n");
  let currentKey = "";
  let body: string[] = [];

  for (const line of lines) {
    const m = line.match(/^##\s+(.+)$/);
    if (m) {
      if (currentKey) sections.set(currentKey, body.join("\n").trim());
      currentKey = m[1]!.trim();
      body = [];
    } else if (currentKey) {
      body.push(line);
    }
  }
  if (currentKey) sections.set(currentKey, body.join("\n").trim());
  return sections;
}

function getSection(sections: Map<string, string>, name: string): { found: boolean; body: string } {
  for (const [key, body] of sections) {
    if (key.toLowerCase() === name.toLowerCase()) return { found: true, body };
  }
  return { found: false, body: "" };
}

function parseRequirementBlocks(sectionBody: string): RequirementBlock[] {
  if (!sectionBody.trim()) return [];
  const blocks: RequirementBlock[] = [];
  const lines = normalizeLineEndings(sectionBody).split("\n");
  let current: RequirementBlock | null = null;
  let contentLines: string[] = [];

  for (const line of lines) {
    const m = line.match(REQUIREMENT_HEADER);
    if (m) {
      if (current) {
        current.content = contentLines.join("\n").trim();
        blocks.push(current);
      }
      current = { name: m[1]!.trim(), content: "", scenarios: [] };
      contentLines = [];
      continue;
    }
    if (current) {
      if (/^####\s*Scenario:/i.test(line)) {
        current.scenarios.push(line);
      }
      contentLines.push(line);
    }
  }
  if (current) {
    current.content = contentLines.join("\n").trim();
    blocks.push(current);
  }
  return blocks;
}

function parseRemovedNames(sectionBody: string): string[] {
  const names: string[] = [];
  for (const line of normalizeLineEndings(sectionBody).split("\n")) {
    const m = line.match(REQUIREMENT_HEADER);
    if (m) names.push(m[1]!.trim());
  }
  return names;
}

function parseRenamedPairs(sectionBody: string): Array<{ from: string; to: string }> {
  const pairs: Array<{ from: string; to: string }> = [];
  for (const line of normalizeLineEndings(sectionBody).split("\n")) {
    const from = line.match(/FROM:\s*`?###\s*Requirement:\s*(.+?)`?\s*$/i);
    const to = line.match(/TO:\s*`?###\s*Requirement:\s*(.+?)`?\s*$/i);
    if (from) pairs.push({ from: from[1]!.trim(), to: "" });
    else if (to && pairs.length) pairs[pairs.length - 1]!.to = to[1]!.trim();
  }
  return pairs.filter((p) => p.from && p.to);
}

export function parseDeltaSpec(content: string): DeltaPlan {
  const sections = splitTopLevelSections(content);
  const added = getSection(sections, "ADDED Requirements");
  const modified = getSection(sections, "MODIFIED Requirements");
  const removed = getSection(sections, "REMOVED Requirements");
  const renamed = getSection(sections, "RENAMED Requirements");

  return {
    added: parseRequirementBlocks(added.body),
    modified: parseRequirementBlocks(modified.body),
    removed: parseRemovedNames(removed.body),
    renamed: parseRenamedPairs(renamed.body),
    sectionPresence: {
      added: added.found,
      modified: modified.found,
      removed: removed.found,
      renamed: renamed.found,
    },
  };
}

export function parseProposal(content: string): {
  hasWhy: boolean;
  hasWhat: boolean;
  hasCapabilities: boolean;
  isBreaking: boolean;
} {
  const lower = content.toLowerCase();
  return {
    hasWhy: /##\s*why/i.test(content) || lower.includes("why"),
    hasWhat: /##\s*what/i.test(content) || /what changes/i.test(content),
    hasCapabilities: /capabilities/i.test(content),
    isBreaking: /\*\*breaking\*\*/i.test(content) || /breaking change/i.test(content),
  };
}
