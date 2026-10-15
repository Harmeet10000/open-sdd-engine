import { spawn } from "node:child_process";
import { access } from "node:fs/promises";
import { constants } from "node:fs";
import { join } from "node:path";
import type { LoopKind, SddConfig } from "../core/types.js";

export interface ResearchResult {
  tool: "graphify" | "ast-grep" | "firecrawl" | "context7";
  success: boolean;
  output: string;
  error?: string;
  agentHint?: string;
}

function runCommand(cmd: string, args: string[], cwd: string): Promise<{ stdout: string; stderr: string; code: number }> {
  return new Promise((resolve) => {
    const child = spawn(cmd, args, { cwd, shell: false });
    let stdout = "";
    let stderr = "";
    child.stdout.on("data", (d) => (stdout += d.toString()));
    child.stderr.on("data", (d) => (stderr += d.toString()));
    child.on("close", (code) => resolve({ stdout, stderr, code: code ?? 1 }));
    child.on("error", () => resolve({ stdout, stderr: `${cmd} not found`, code: 127 }));
  });
}

async function exists(path: string): Promise<boolean> {
  try {
    await access(path, constants.R_OK);
    return true;
  } catch {
    return false;
  }
}

export function researchHintsForLoop(loop: LoopKind): string[] {
  if (loop === "change") {
    return [
      "ast-grep: find code structures to modify",
      "graphify query: trace module relationships",
      "context7: verify library APIs",
      "firecrawl search: quick external lookup",
    ];
  }
  return [
    "graphify query: map architecture and data flows",
    "ast-grep: inventory patterns and hotspots",
    "firecrawl search: research latency/throughput techniques",
    "context7: deep framework documentation",
  ];
}

export async function runGraphifyQuery(
  workspace: string,
  query: string,
  config: SddConfig,
): Promise<ResearchResult> {
  if (!config.research.graphify.enabled) {
    return { tool: "graphify", success: false, output: "", error: "graphify disabled in config" };
  }
  const graphPath = join(workspace, config.research.graphify.graphPath);
  if (!(await exists(graphPath))) {
    return {
      tool: "graphify",
      success: false,
      output: "",
      error: `No graph at ${config.research.graphify.graphPath}. Run: graphify ${workspace}`,
      agentHint: "Build graph first with /graphify or graphify CLI",
    };
  }
  const { stdout, stderr, code } = await runCommand("graphify", ["query", query], workspace);
  return {
    tool: "graphify",
    success: code === 0,
    output: stdout || stderr,
    error: code !== 0 ? stderr : undefined,
  };
}

export async function runAstGrep(
  workspace: string,
  pattern: string,
  language: string,
  config: SddConfig,
): Promise<ResearchResult> {
  if (!config.research.astGrep.enabled) {
    return { tool: "ast-grep", success: false, output: "", error: "ast-grep disabled in config" };
  }
  const { stdout, stderr, code } = await runCommand(
    "ast-grep",
    ["run", "--pattern", pattern, "--lang", language, workspace],
    workspace,
  );
  return {
    tool: "ast-grep",
    success: code === 0 || stdout.length > 0,
    output: stdout || stderr,
    error: code !== 0 && !stdout ? stderr : undefined,
  };
}

export async function runFirecrawlSearch(
  workspace: string,
  query: string,
  config: SddConfig,
): Promise<ResearchResult> {
  if (!config.research.firecrawl.enabled) {
    return { tool: "firecrawl", success: false, output: "", error: "firecrawl disabled in config" };
  }
  const { stdout, stderr, code } = await runCommand(
    "firecrawl",
    ["search", query, "--limit", "5"],
    workspace,
  );
  return {
    tool: "firecrawl",
    success: code === 0,
    output: stdout || stderr,
    error: code !== 0 ? stderr : undefined,
    agentHint: code === 127 ? "Install firecrawl CLI and set FIRECRAWL_API_KEY" : undefined,
  };
}

export async function runContext7Hint(
  libraryName: string,
  query: string,
  config: SddConfig,
): Promise<ResearchResult> {
  if (!config.research.context7.enabled) {
    return { tool: "context7", success: false, output: "", error: "context7 disabled in config" };
  }
  return {
    tool: "context7",
    success: true,
    output: "",
    agentHint: [
      `Use Context7 MCP server "${config.research.context7.mcpServer}":`,
      `1. resolve-library-id(libraryName="${libraryName}", query="${query}")`,
      `2. query-docs(libraryId=<selected>, query="${query}")`,
    ].join("\n"),
  };
}

export type ResearchKind = "codebase" | "structure" | "web" | "docs";

export async function runResearch(
  kind: ResearchKind,
  workspace: string,
  config: SddConfig,
  opts: { query: string; language?: string; library?: string },
): Promise<ResearchResult> {
  switch (kind) {
    case "codebase":
      return runGraphifyQuery(workspace, opts.query, config);
    case "structure":
      return runAstGrep(workspace, opts.query, opts.language ?? "typescript", config);
    case "web":
      return runFirecrawlSearch(workspace, opts.query, config);
    case "docs":
      return runContext7Hint(opts.library ?? opts.query.split(" ")[0] ?? "unknown", opts.query, config);
  }
}
