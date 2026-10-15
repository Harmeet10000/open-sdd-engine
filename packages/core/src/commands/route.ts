import { classifyLoop, parseBudgetHours } from "../router/loop-classifier.js";

export function runRoute(description: string, budget: string, json: boolean): void {
  const hours = parseBudgetHours(budget);
  const result = classifyLoop(description, hours);
  const out = { description, budgetHours: hours, ...result };
  if (json) console.log(JSON.stringify(out, null, 2));
  else {
    console.log(`Recommended loop: ${result.recommended} (${result.confidence} confidence)`);
    for (const r of result.reasons) console.log(`  - ${r}`);
    console.log(`\nRun: sdd new ${result.recommended} <id>`);
  }
}
