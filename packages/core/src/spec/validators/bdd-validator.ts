import type { DiagnosticIssue } from "../../core/types.js";
import { issue } from "../../core/reporter.js";
import { parseBdd } from "../parsers/bdd.js";

export function validateBdd(content: string, strict: boolean): { errors: DiagnosticIssue[]; warnings: DiagnosticIssue[] } {
  const errors: DiagnosticIssue[] = [];
  const warnings: DiagnosticIssue[] = [];
  const feature = parseBdd(content);
  if (!feature.name) errors.push(issue("BDD_MISSING_FEATURE", "error", "behavior.feature requires Feature:", { file: "behavior.feature", loop: "spec" }));
  if (!feature.scenarios.length) errors.push(issue("BDD_MISSING_SCENARIO", "error", "behavior.feature requires at least one Scenario:", { file: "behavior.feature", loop: "spec" }));
  for (const scenario of feature.scenarios) {
    const keywords = new Set(scenario.steps.map((step) => step.keyword));
    if (!keywords.has("when") || !keywords.has("then")) errors.push(issue("BDD_INCOMPLETE_SCENARIO", "error", `Scenario "${scenario.name}" requires When and Then steps`, { file: "behavior.feature", line: scenario.line, loop: "spec" }));
    if (!keywords.has("given")) (strict ? errors : warnings).push(issue("BDD_MISSING_GIVEN", strict ? "error" : "warn", `Scenario "${scenario.name}" has no Given context`, { file: "behavior.feature", line: scenario.line, loop: "spec" }));
  }
  return { errors, warnings };
}
