import type { SpecBundle } from "../../core/discovery.js";
import type { DiagnosticIssue, SddConfig, ValidationResult } from "../../core/types.js";
import { issue } from "../../core/reporter.js";
import { parseRequirements, buildRequirementRegistry } from "../parsers/requirements.js";
import { parseDesign } from "../parsers/design.js";
import { parseTasks, inferDependencies, countTopLevelTasks } from "../parsers/tasks.js";
import { CycleError, compileWavesFromGraph } from "../compiler/dag-waves.js";
import { validateBdd } from "./bdd-validator.js";
import { SpecMetaSchema } from "../../core/schemas.js";

export function validateSpec(bundle: SpecBundle, config: SddConfig, strict: boolean): ValidationResult {
  const errors: DiagnosticIssue[] = [];
  const warnings: DiagnosticIssue[] = [];

  if (!bundle.requirements) {
    errors.push(
      issue("SPEC_MISSING_FILE", "error", "Missing requirements.md", { file: "requirements.md", loop: "spec" }),
    );
  }
  if (!bundle.design) {
    errors.push(issue("SPEC_MISSING_FILE", "error", "Missing design.md", { file: "design.md", loop: "spec" }));
  }
  if (!bundle.tasks) {
    errors.push(issue("SPEC_MISSING_FILE", "error", "Missing tasks.md", { file: "tasks.md", loop: "spec" }));
  }
  if (!bundle.behavior) {
    const target = strict ? errors : warnings;
    target.push(issue("BDD_MISSING_FILE", strict ? "error" : "warn", "Missing behavior.feature", { file: "behavior.feature", loop: "spec" }));
  } else {
    const bdd = validateBdd(bundle.behavior, strict);
    errors.push(...bdd.errors); warnings.push(...bdd.warnings);
  }

  let registry = new Set<string>();
  if (!bundle.meta) {
    (strict ? errors : warnings).push(issue("SDD_SCHEMA_MISSING", strict ? "error" : "warn", "Missing spec.meta.json", { file: "spec.meta.json", loop: "spec" }));
  } else {
    try { SpecMetaSchema.parse(JSON.parse(bundle.meta)); }
    catch (error) { errors.push(issue("SDD_SCHEMA_INVALID", "error", `Invalid spec.meta.json: ${error instanceof Error ? error.message : "unknown error"}`, { file: "spec.meta.json", loop: "spec" })); }
  }
  if (bundle.requirements) {
    const reqs = parseRequirements(bundle.requirements);
    if (!reqs.length) {
      errors.push(issue("EARS_MISSING_REQUIREMENT", "error", "requirements.md requires at least one numbered Requirement section", { file: "requirements.md", loop: "spec" }));
    }
    registry = buildRequirementRegistry(reqs);
    for (const r of reqs) {
      if (!r.criteria.length) {
        errors.push(issue("EARS_MISSING_ACCEPTANCE_CRITERIA", "error", `Requirement ${r.number} has no acceptance criteria`, { file: "requirements.md", loop: "spec" }));
      }
      for (const c of r.criteria) {
        if (c.earsKind === "unknown") {
          const sev = strict ? "error" : "warn";
          (sev === "error" ? errors : warnings).push(
            issue("EARS_NONCONFORMANT", sev, `Line ${c.line} does not conform to EARS: ${c.text.slice(0, 60)}...`, {
              file: "requirements.md",
              line: c.line,
              loop: "spec",
            }),
          );
        }
      }
    }
  }

  if (bundle.design) {
    const design = parseDesign(bundle.design);
    if (!design.hasOverview)
      warnings.push(issue("DESIGN_MISSING_OVERVIEW", "warn", "design.md missing ## Overview", { loop: "spec" }));
    if (!design.hasArchitecture)
      warnings.push(issue("DESIGN_MISSING_ARCHITECTURE", "warn", "design.md missing ## Architecture", { loop: "spec" }));
    for (const block of design.mermaidBlocks) {
      if (!block.valid) {
        warnings.push(
          issue("MERMAID_FENCE_INVALID", "warn", `Invalid mermaid fence at line ${block.startLine}`, {
            file: "design.md",
            line: block.startLine,
            loop: "spec",
          }),
        );
      }
    }
  }

  if (bundle.tasks) {
    const { tasks, dependencyGraph } = parseTasks(bundle.tasks);
    const ids = new Set(tasks.map((t) => t.id));
    const seen = new Set<string>();

    for (const t of tasks) {
      if (seen.has(t.id)) {
        errors.push(
          issue("TASK_DUPLICATE_ID", "error", `Duplicate task id ${t.id}`, {
            file: "tasks.md",
            line: t.line,
            loop: "spec",
          }),
        );
      }
      seen.add(t.id);
      if (t.depth > 2) {
        errors.push(
          issue("TASK_DEPTH_EXCEEDED", "error", `Task ${t.id} exceeds max depth 2`, {
            file: "tasks.md",
            line: t.line,
            loop: "spec",
          }),
        );
      }
      for (const ref of t.requirementRefs) {
        if (!registry.has(ref)) {
          errors.push(
            issue("TRACEABILITY_UNKNOWN_REF", "error", `Unknown requirement ref ${ref} on task ${t.id}`, {
              file: "tasks.md",
              line: t.line,
              loop: "spec",
            }),
          );
        }
      }
    }

    const topLevel = countTopLevelTasks(tasks);
    if (topLevel > config.loops.spec.requireDependencyGraphMinTasks && !dependencyGraph) {
      warnings.push(
        issue(
          "DAG_INCOMPLETE",
          "warn",
          `${topLevel} top-level tasks without Task Dependency Graph JSON`,
          { loop: "spec" },
        ),
      );
    }

    if (bundle.design && topLevel > config.loops.spec.requireFilePlanMinTasks) {
      const design = parseDesign(bundle.design);
      if (!design.hasFileStructure) {
        warnings.push(
          issue("DESIGN_MISSING_FILE_PLAN", "warn", "Large spec should include file structure plan in design.md", {
            loop: "spec",
          }),
        );
      }
    }

    try {
      compileWavesFromGraph(ids, dependencyGraph, inferDependencies(tasks));
    } catch (e) {
      if (e instanceof CycleError) {
        errors.push(
          issue("DAG_CYCLE", "error", e.message, { loop: "spec" }),
        );
      }
    }
  }

  return {
    id: bundle.slug,
    loop: "spec",
    valid: errors.length === 0,
    errors,
    warnings,
  };
}

export function getSpecStatus(bundle: SpecBundle): {
  artifacts: Record<string, boolean>;
  phase: string;
} {
  return {
    artifacts: {
      requirements: !!bundle.requirements,
      design: !!bundle.design,
      tasks: !!bundle.tasks,
    },
    phase:
      !bundle.requirements
        ? "requirements"
        : !bundle.design
          ? "design"
          : !bundle.tasks
            ? "tasks"
            : "implement",
  };
}

export function getSpecInstructions(bundle: SpecBundle): {
  loop: "spec";
  id: string;
  phase: string;
  nextArtifact: string | null;
  pendingTasks: string[];
  contextFiles: string[];
  research: string[];
} {
  const status = getSpecStatus(bundle);
  const pendingTasks =
    bundle.tasks
      ? parseTasks(bundle.tasks)
          .tasks.filter((t) => !t.done)
          .map((t) => t.id)
      : [];

  const nextArtifact =
    !status.artifacts.requirements
      ? "requirements.md"
      : !status.artifacts.design
        ? "design.md"
        : !status.artifacts.tasks
          ? "tasks.md"
          : null;

  return {
    loop: "spec",
    id: bundle.slug,
    phase: status.phase,
    nextArtifact,
    pendingTasks,
    contextFiles: ["requirements.md", "design.md", "tasks.md"].filter(
      (_, i) => Object.values(status.artifacts)[i],
    ),
    research: [
      "graphify query: architecture exploration and dependency paths",
      "ast-grep: structural patterns across codebase",
      "firecrawl search: benchmarks, papers, external system designs",
      "context7: framework and library documentation",
    ],
  };
}

export function compileSpecWaves(bundle: SpecBundle): string[][] {
  if (!bundle.tasks) return [];
  const { tasks, dependencyGraph } = parseTasks(bundle.tasks);
  const ids = new Set(tasks.map((t) => t.id));
  return compileWavesFromGraph(ids, dependencyGraph, inferDependencies(tasks));
}
