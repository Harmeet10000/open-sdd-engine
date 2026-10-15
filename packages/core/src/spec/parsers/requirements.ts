export type EarsKind = "ubiquitous" | "event_driven" | "state_driven" | "optional_feature" | "conditional" | "complex" | "unknown";

export interface AcceptanceCriterion {
  id: string;
  text: string;
  line: number;
  earsKind: EarsKind;
}

export interface Requirement {
  number: number;
  title: string;
  criteria: AcceptanceCriterion[];
}

const EARS_PATTERNS: Array<{ kind: EarsKind; pattern: RegExp }> = [
  { kind: "complex", pattern: /^WHILE\s+.+,?\s+WHEN\s+.+,?\s+(THE SYSTEM|[A-Za-z0-9 ]+)\s+SHALL/i },
  { kind: "event_driven", pattern: /^WHEN\s+.+\s+(THE SYSTEM|[A-Za-z0-9 ]+)\s+SHALL/i },
  { kind: "event_driven", pattern: /^WHEN\s+.+\s+THEN\s+.+\s+SHALL/i },
  { kind: "state_driven", pattern: /^WHILE\s+.+\s+(THE SYSTEM|[A-Za-z0-9 ]+)\s+SHALL/i },
  { kind: "optional_feature", pattern: /^WHERE\s+.+\s+(THE SYSTEM|[A-Za-z0-9 ]+)\s+SHALL/i },
  { kind: "conditional", pattern: /^IF\s+.+\s+(THE SYSTEM|[A-Za-z0-9 ]+)\s+SHALL/i },
  { kind: "conditional", pattern: /^IF\s+.+,?\s+THEN\s+(THE SYSTEM|[A-Za-z0-9 ]+)\s+SHALL/i },
  { kind: "ubiquitous", pattern: /^THE SYSTEM SHALL/i },
  { kind: "ubiquitous", pattern: /^[A-Z][A-Za-z0-9 ]+\s+SHALL/i },
];

export function classifyEars(text: string): EarsKind {
  const clean = text.replace(/^\d+\.\s*/, "").trim();
  for (const { kind, pattern } of EARS_PATTERNS) {
    if (pattern.test(clean)) return kind;
  }
  return "unknown";
}

export function parseRequirements(content: string): Requirement[] {
  const requirements: Requirement[] = [];
  const lines = content.split("\n");
  let current: Requirement | null = null;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]!;
    const reqMatch = line.match(/^###\s*Requirement\s+(\d+):\s*(.+)$/i);
    if (reqMatch) {
      if (current) requirements.push(current);
      current = { number: parseInt(reqMatch[1]!, 10), title: reqMatch[2]!.trim(), criteria: [] };
      continue;
    }
    const critMatch = line.match(/^(\d+)\.\s+(.+)$/);
    if (critMatch && current) {
      const text = critMatch[2]!.trim();
      if (/shall/i.test(text) || /^(when|if|while|where)/i.test(text)) {
        current.criteria.push({
          id: `${current.number}.${critMatch[1]}`,
          text,
          line: i + 1,
          earsKind: classifyEars(text),
        });
      }
    }
  }
  if (current) requirements.push(current);
  return requirements;
}

export function buildRequirementRegistry(requirements: Requirement[]): Set<string> {
  const ids = new Set<string>();
  for (const r of requirements) {
    for (const c of r.criteria) ids.add(c.id);
  }
  return ids;
}
