import type { ChangeBundle } from "../../core/discovery.js";
import type { DiagnosticIssue, SddConfig, ValidationResult } from "../../core/types.js";
import { issue } from "../../core/reporter.js";
import { readFileIfExists } from "../../core/discovery.js";
import { parseDeltaSpec } from "../parsers/delta-spec.js";
import { parseProposal } from "../parsers/delta-spec.js";
import { parseLinearTasks } from "../parsers/linear-tasks.js";
import { join } from "node:path";

function scenarioValid(block: { scenarios: string[]; content: string }): boolean {
  if (block.scenarios.length === 0) return false;
  return block.scenarios.every((s) => /^####\s*Scenario:/i.test(s));
}

export async function validateChange(
  bundle: ChangeBundle,
  config: SddConfig,
  strict: boolean,
): Promise<ValidationResult> {
  const errors: DiagnosticIssue[] = [];
  const warnings: DiagnosticIssue[] = [];

  if (!bundle.proposal) {
    errors.push(
      issue("CHANGE_MISSING_PROPOSAL", "error", "Missing proposal.md", {
        file: "proposal.md",
        loop: "change",
      }),
    );
  } else {
    const p = parseProposal(bundle.proposal);
    if (!p.hasWhy)
      warnings.push(issue("CHANGE_MISSING_WHY", "warn", "proposal.md missing Why section", { loop: "change" }));
    if (!p.hasWhat)
      warnings.push(issue("CHANGE_MISSING_WHAT", "warn", "proposal.md missing What Changes section", { loop: "change" }));
    if (p.isBreaking && config.loops.change.requireDesignForBreaking && !bundle.design) {
      const sev = strict ? "error" : "warn";
      (sev === "error" ? errors : warnings).push(
        issue("CHANGE_MISSING_DESIGN", sev, "BREAKING change requires design.md", { loop: "change" }),
      );
    }
  }

  if (bundle.deltaSpecPaths.length === 0) {
    warnings.push(
      issue("DELTA_EMPTY_SECTION", "warn", "No delta spec files under specs/", { loop: "change" }),
    );
  }

  for (const specPath of bundle.deltaSpecPaths) {
    const content = await readFileIfExists(specPath);
    if (!content) continue;
    const plan = parseDeltaSpec(content);
    const hasAny =
      plan.added.length || plan.modified.length || plan.removed.length || plan.renamed.length;
    if (!hasAny) {
      errors.push(
        issue("DELTA_EMPTY_SECTION", "error", `Delta spec has no requirement sections: ${specPath}`, {
          file: specPath,
          loop: "change",
        }),
      );
    }
    for (const block of plan.added) {
      if (!scenarioValid(block)) {
        const sev = strict ? "error" : "warn";
        (sev === "error" ? errors : warnings).push(
          issue(
            "SCENARIO_HEADER_INVALID",
            sev,
            `Requirement "${block.name}" must have at least one #### Scenario:`,
            { file: specPath, loop: "change" },
          ),
        );
      }
    }
    for (const block of plan.modified) {
      if (!scenarioValid(block)) {
        const sev = strict ? "error" : "warn";
        (sev === "error" ? errors : warnings).push(
          issue(
            "SCENARIO_HEADER_INVALID",
            sev,
            `Requirement "${block.name}" must have at least one #### Scenario:`,
            { file: specPath, loop: "change" },
          ),
        );
      }
      if (block.content.length < 50) {
        errors.push(
          issue(
            "MODIFIED_PARTIAL_CONTENT",
            "error",
            `MODIFIED requirement "${block.name}" appears incomplete (full block required)`,
            { file: specPath, loop: "change" },
          ),
        );
      }
    }
    for (const name of plan.removed) {
      const block = plan.added.find((b) => b.name === name);
      void block;
    }
    const removedSection = content.match(/##\s*REMOVED Requirements[\s\S]*/i)?.[0] ?? "";
    if (plan.removed.length && !/\*\*Reason\*\*/i.test(removedSection)) {
      warnings.push(
        issue("REMOVED_MISSING_REASON", "warn", "REMOVED requirements should include **Reason**", {
          file: specPath,
          loop: "change",
        }),
      );
    }
  }

  if (bundle.tasks) {
    const tasks = parseLinearTasks(bundle.tasks);
    const topLevel = tasks.filter((t) => !t.id.includes("."));
    if (topLevel.length > 8) {
      warnings.push(
        issue(
          "TASK_BUDGET_EXCEEDED",
          "warn",
          `${topLevel.length} top-level tasks exceeds change loop budget; consider spec loop`,
          { loop: "change" },
        ),
      );
    }
  } else {
    warnings.push(issue("CHANGE_MISSING_TASKS", "warn", "Missing tasks.md", { loop: "change" }));
  }

  return {
    id: bundle.id,
    loop: "change",
    valid: errors.length === 0,
    errors,
    warnings,
  };
}

export function getChangeStatus(bundle: ChangeBundle): {
  artifacts: Record<string, boolean>;
  phase: string;
} {
  const hasProposal = !!bundle.proposal;
  const hasSpecs = bundle.deltaSpecPaths.length > 0;
  const hasDesign = !!bundle.design;
  const hasTasks = !!bundle.tasks;

  let phase = "proposal";
  if (hasProposal && !hasSpecs) phase = "specs";
  else if (hasProposal && hasSpecs && !hasTasks) phase = "tasks";
  else if (hasProposal && hasSpecs && hasTasks) phase = "apply";
  if (hasDesign && phase === "specs") phase = "design";

  return {
    artifacts: { proposal: hasProposal, specs: hasSpecs, design: hasDesign, tasks: hasTasks },
    phase,
  };
}

export function getChangeInstructions(bundle: ChangeBundle): {
  loop: "change";
  id: string;
  phase: string;
  nextArtifact: string | null;
  pendingTasks: string[];
  contextFiles: string[];
  research: string[];
} {
  const status = getChangeStatus(bundle);
  const pendingTasks = bundle.tasks ? parseLinearTasks(bundle.tasks).filter((t) => !t.done).map((t) => t.id) : [];
  const nextArtifact =
    !status.artifacts.proposal
      ? "proposal.md"
      : !status.artifacts.specs
        ? "specs/<domain>/spec.md"
        : !status.artifacts.tasks
          ? "tasks.md"
          : null;

  const contextFiles = ["proposal.md"];
  if (status.artifacts.specs) contextFiles.push("specs/");
  if (status.artifacts.design) contextFiles.push("design.md");
  if (status.artifacts.tasks) contextFiles.push("tasks.md");

  return {
    loop: "change",
    id: bundle.id,
    phase: status.phase,
    nextArtifact,
    pendingTasks,
    contextFiles,
    research: [
      "ast-grep: structural search for code to modify",
      "graphify query: understand affected modules",
      "context7: library API docs for dependencies",
      "firecrawl search: external references if needed",
    ],
  };
}
