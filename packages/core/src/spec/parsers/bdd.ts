export interface BddScenario { name: string; line: number; steps: Array<{ keyword: string; text: string; line: number }>; }
export interface BddFeature { name?: string; scenarios: BddScenario[]; }

export function parseBdd(content: string): BddFeature {
  const scenarios: BddScenario[] = [];
  let name: string | undefined;
  let current: BddScenario | undefined;
  for (const [index, raw] of content.split(/\r?\n/).entries()) {
    const line = index + 1;
    const feature = raw.match(/^\s*Feature:\s*(.+)$/i);
    if (feature) { name = feature[1]!.trim(); continue; }
    const scenario = raw.match(/^\s*Scenario(?: Outline)?:\s*(.+)$/i);
    if (scenario) { current = { name: scenario[1]!.trim(), line, steps: [] }; scenarios.push(current); continue; }
    const step = raw.match(/^\s*(Given|When|Then|And|But)\s+(.+)$/i);
    if (step && current) current.steps.push({ keyword: step[1]!.toLowerCase(), text: step[2]!.trim(), line });
  }
  return { name, scenarios };
}
