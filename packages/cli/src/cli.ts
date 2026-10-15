#!/usr/bin/env bun
import { Command } from "commander";
import { runInit } from "@open-sdd/core/commands/init";
import { runNewChange, runNewSpec } from "@open-sdd/core/commands/new";
import { runValidate, runDiagnostics } from "@open-sdd/core/commands/validate";
import { runStatusSelection } from "@open-sdd/core/commands/status";
import { runInstructionsSelection } from "@open-sdd/core/commands/instructions";
import { runArchive } from "@open-sdd/core/commands/archive";
import { runCompileWaves } from "@open-sdd/core/commands/compile-waves";
import { runList } from "@open-sdd/core/commands/list";
import { runRoute } from "@open-sdd/core/commands/route";
import { runResearchCommand } from "@open-sdd/core/commands/research";
import { runSpecify } from "@open-sdd/core/commands/specify";
import { runPlan } from "@open-sdd/core/commands/plan";
import { runAutoResearch } from "@open-sdd/core/commands/autoresearch";
import { discoverSkills, selectSkills } from "@open-sdd/skills";
import { openClarifications, runClarify, runWorkflowGenerate, runWorkflowHandoff, runWorkflowStatus } from "@open-sdd/core/commands/workflow";
import { findWorkspace, loadSpec } from "@open-sdd/core/core/discovery";
import { classifyLoop, parseBudgetHours } from "@open-sdd/core/router/loop-classifier";
import { createInterface } from "node:readline/promises";
import { runConflictAbort, runConflictStatus, runGitBranch, runGitCommit, runGitPr, runGitStatus, runGitWorktree, runProfileInit, runProfileShow } from "@open-sdd/core/commands/git";
import { runMemoryRemember, runMemorySearch } from "@open-sdd/core/commands/memory";

const cwd = () => process.cwd();
const program = new Command()
  .name("sdd")
  .description("Open SDD Engine: OpenSpec change loop plus Kiro/Spec Kit spec loop")
  .version("0.1.0");

program.command("init").option("--tools <tool>", "agent integration", "cursor").option("--okf", "create Google OKF knowledge bundle").option("--no-profile", "skip interactive coding-profile capture").action(async (opts) => {
  await runInit(cwd(), opts.tools, opts.okf);
  if (process.stdin.isTTY && !opts.profile) await runProfileInit(cwd(), false);
});

const create = program.command("new").description("scaffold a Change or Spec loop");
create.action(async () => {
  if (!process.stdin.isTTY) throw new Error("sdd new needs a terminal. Use `sdd new change <id>` or `sdd new spec <slug>` in automation.");
  const prompt = createInterface({ input: process.stdin, output: process.stdout });
  try {
    const description = (await prompt.question("What are you changing? ")).trim();
    const budget = (await prompt.question("Time budget [2h]: ")).trim() || "2h";
    const route = classifyLoop(description, parseBudgetHours(budget));
    const choice = (await prompt.question(`Recommended ${route.recommended} loop. Press Enter to accept or choose change/spec: `)).trim().toLowerCase() || route.recommended;
    if (choice !== "change" && choice !== "spec") throw new Error("Loop must be change or spec.");
    const id = (await prompt.question(`Identifier (${choice}): `)).trim();
    if (!/^[a-z0-9][a-z0-9-]*$/.test(id)) throw new Error("Identifier must be lowercase kebab-case.");
    if (choice === "change") await runNewChange(cwd(), id, false);
    else await runNewSpec(cwd(), id, false);
  } finally {
    prompt.close();
  }
});
create.command("change <id>").option("--json").action((id, opts) => runNewChange(cwd(), id, opts.json));
create.command("spec <slug>").option("--json").action((slug, opts) => runNewSpec(cwd(), slug, opts.json));

program.command("specify <slug>").option("--json").action((slug, opts) => runSpecify(cwd(), slug, opts.json));
program.command("plan <slug>").option("--json").action((slug, opts) => runPlan(cwd(), slug, opts.json));
program.command("clarify <slug>").option("--answer <answer...>", "answers in question order", []).option("--json").action(async (slug, opts) => {
  if (opts.answer.length || !process.stdin.isTTY) return runClarify(cwd(), slug, opts.answer, opts.json);
  const workspace = await findWorkspace(cwd()); const requirements = (await loadSpec(workspace, slug)).requirements;
  if (!requirements) throw new Error("Missing requirements.md");
  const prompt = createInterface({ input: process.stdin, output: process.stdout });
  try {
    const answers: string[] = [];
    for (const question of openClarifications(requirements)) answers.push(await prompt.question(`${question}\n> `));
    return runClarify(cwd(), slug, answers, opts.json);
  } finally { prompt.close(); }
});

const workflow = program.command("workflow").description("generate deterministic planner, implementer, and reviewer contracts");
workflow.command("generate <id>").requiredOption("--type <loop>", "change, spec, or custom").option("--description <text>", "source for custom workflows").option("--phase <name...>", "ordered custom phases").option("--json").action((id, opts) => {
  if (!["change", "spec", "custom"].includes(opts.type)) throw new Error("--type must be change, spec, or custom");
  if (opts.type === "custom" && !opts.description) throw new Error("Custom workflows require --description.");
  return runWorkflowGenerate(cwd(), id, opts.type, opts.json, opts.description, opts.phase);
});
workflow.command("status <id>").option("--json").action((id, opts) => runWorkflowStatus(cwd(), id, opts.json));
workflow.command("handoff <id>").requiredOption("--role <role>", "planner, implementer, or reviewer").option("--json").action((id, opts) => {
  if (!["planner", "implementer", "reviewer"].includes(opts.role)) throw new Error("--role must be planner, implementer, or reviewer");
  return discoverSkills(cwd()).then((skills) => runWorkflowHandoff(cwd(), id, opts.role, opts.json, skills.map((skill) => skill.id)));
});

function validationOptions(command: Command): Command {
  return command.argument("[name]").option("--changes").option("--specs").option("--okf").option("--all").option("--strict").option("--json");
}
for (const [name, runner] of [["validate", runValidate], ["diagnostics", runDiagnostics]] as const) {
  validationOptions(program.command(name)).action(async (nameArg, opts) => {
    process.exitCode = await runner({ cwd: cwd(), name: nameArg, ...opts });
  });
}

for (const [name, runner] of [["status", runStatusSelection], ["instructions", runInstructionsSelection]] as const) {
  program.command(`${name} [id]`).option("--changes").option("--specs").option("--all").option("--strict").option("--json").action((id, opts) => runner({ cwd: cwd(), id, ...opts }));
}

program.command("archive <change-id>").requiredOption("--yes", "confirm archive").option("--skip-specs").action((id, opts) => runArchive(cwd(), id, opts.yes, opts.skipSpecs));
program.command("compile-waves <slug>").option("--json").action((slug, opts) => runCompileWaves(cwd(), slug, opts.json));
program.command("list").option("--changes").option("--specs").option("--json").action((opts) => runList(cwd(), opts.changes, opts.specs, opts.json));
program.command("route <description>").option("--budget <duration>", "time budget", "2h").option("--json").action((description, opts) => runRoute(description, opts.budget, opts.json));

const skills = program.command("skills").description("discover and select local, trusted skills");
skills.command("select <description>").option("--json").action((description, opts) => {
  const output = { description, selected: selectSkills(description) };
  console.log(opts.json ? JSON.stringify(output, null, 2) : output.selected.map((skill) => `${skill.mode.toUpperCase()} ${skill.id}: ${skill.reason}`).join("\n"));
});
skills.command("discover").option("--json").action(async (opts) => {
  const output = await discoverSkills(cwd());
  console.log(opts.json ? JSON.stringify({ skills: output }, null, 2) : output.map((skill) => `${skill.id}\t${skill.path}`).join("\n"));
});

const research = program.command("research");
for (const kind of ["codebase", "structure", "web", "docs"] as const) {
  research.command(`${kind} <query>`).option("--lang <language>", "source language", "typescript").option("--library <name>").option("--json").action(async (query, opts) => {
    process.exitCode = await runResearchCommand(cwd(), kind, query, { language: opts.lang, library: opts.library, json: opts.json });
  });
}

program.command("autoresearch").command("run").requiredOption("--cmd <command>", "evaluation command").option("--revert-on-fail").option("--metric <regex>").option("--json").action((opts) => runAutoResearch(cwd(), opts));

const git = program.command("git").description("safe Git branches, worktrees, commits, PRs, and conflict inspection");
git.command("status").option("--json").action((opts) => runGitStatus(cwd(), opts.json));
git.command("branch <name>").option("--base <ref>", "base ref").option("--json").action((name, opts) => runGitBranch(cwd(), name, opts.base, opts.json));
git.command("worktree <path>").requiredOption("--branch <name>", "new branch name").option("--base <ref>", "base ref").option("--json").action((path, opts) => runGitWorktree(cwd(), path, opts.branch, opts.base, opts.json));
git.command("commit").requiredOption("--message <message>").requiredOption("--files <paths...>", "explicit files to stage").option("--skip-checks", "skip git diff --check only").option("--json").action((opts) => runGitCommit(cwd(), opts.message, opts.files, opts.skipChecks, opts.json));
git.command("pr").requiredOption("--title <title>").requiredOption("--body <body>").option("--base <branch>", "target branch", "main").option("--push", "push current branch before PR creation").option("--json").action((opts) => runGitPr(cwd(), opts.title, opts.body, opts.base, opts.push, opts.json));
const conflict = git.command("conflict").description("inspect conflicts; never auto-resolve them");
conflict.command("status").option("--json").action((opts) => runConflictStatus(cwd(), opts.json));
conflict.command("abort").requiredOption("--yes", "confirm abort").option("--json").action((opts) => runConflictAbort(cwd(), opts.yes, opts.json));

const profile = program.command("profile").description("capture project coding taste and tooling preferences");
profile.command("init").option("--json").action((opts) => runProfileInit(cwd(), opts.json));
profile.command("show").option("--json").action((opts) => runProfileShow(cwd(), opts.json));

const memory = program.command("memory").description("durable workflow memory, optionally synchronized with Cognee");
memory.command("remember <text>").option("--scope <scope>", "workflow, decision, research, or review", "workflow").option("--tag <tags...>", "tags", []).option("--json").action((text, opts) => runMemoryRemember(cwd(), text, opts.scope, opts.tag, opts.json));
memory.command("search <query>").option("--json").action((query, opts) => runMemorySearch(cwd(), query, opts.json));

program.parseAsync().catch((error) => {
  console.error(error instanceof Error ? error.message : error);
  process.exitCode = 1;
});
