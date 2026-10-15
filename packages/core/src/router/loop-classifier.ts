import type { LoopKind } from "../core/types.js";

const CHANGE_KEYWORDS =
  /\b(refactor|fix|bug|patch|rename|bump|migrate field|typo|cleanup|lint|format|deps?|dependency|hotfix|small change|env var)\b/i;

const SPEC_KEYWORDS =
  /\b(architecture|latency|throughput|protocol|kernel|simd|aot|hft|quant|sub-?ms|million rps|distributed|microservice|greenfield|framework|design system|performance critical|zero.?copy|io_uring|epoll)\b/i;

export interface RouteResult {
  recommended: LoopKind;
  confidence: "high" | "medium" | "low";
  reasons: string[];
}

export function classifyLoop(description: string, budgetHours = 2): RouteResult {
  const reasons: string[] = [];
  let changeScore = 0;
  let specScore = 0;

  if (CHANGE_KEYWORDS.test(description)) {
    changeScore += 2;
    reasons.push("Matches change-loop keywords (refactor/fix/small change)");
  }
  if (SPEC_KEYWORDS.test(description)) {
    specScore += 3;
    reasons.push("Matches spec-loop keywords (architecture/performance/latency)");
  }

  const wordCount = description.split(/\s+/).length;
  if (wordCount < 8) changeScore += 1;
  if (wordCount > 20) specScore += 1;

  if (/\b(from scratch|new system|new framework|entire api)\b/i.test(description)) {
    specScore += 2;
    reasons.push("Greenfield scope suggests spec loop");
  }
  if (/\b(existing|current|brownfield|incremental)\b/i.test(description)) {
    changeScore += 1;
    reasons.push("Brownfield scope suggests change loop");
  }

  if (budgetHours <= 2) {
    changeScore += 1;
    reasons.push(`Budget ${budgetHours}h fits change loop (≤2h default)`);
  } else {
    specScore += 2;
    reasons.push(`Budget ${budgetHours}h exceeds change loop threshold`);
  }

  let recommended: LoopKind = changeScore >= specScore ? "change" : "spec";
  let confidence: RouteResult["confidence"] = "medium";

  if (Math.abs(changeScore - specScore) >= 3) confidence = "high";
  else if (Math.abs(changeScore - specScore) <= 1) confidence = "low";

  if (confidence === "low" && specScore === changeScore) {
    recommended = "change";
    reasons.push("Tie-breaker: default to change loop for ambiguous short tasks");
  }

  return { recommended, confidence, reasons };
}

export function parseBudgetHours(budget: string): number {
  const m = budget.match(/^(\d+(?:\.\d+)?)\s*h?$/i);
  return m ? parseFloat(m[1]!) : 2;
}
