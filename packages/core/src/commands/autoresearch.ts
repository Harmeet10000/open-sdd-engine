import { spawn } from "node:child_process";
import { findWorkspace } from "../core/discovery.js";

function parseCommand(command: string): string[] {
  const tokens = command.match(/(?:[^\s"']+|"[^"]*"|'[^']*')+/g) ?? [];
  return tokens.map((token) => token.replace(/^("|')|("|')$/g, ""));
}

function executeCommand(
  command: string,
  cwd: string,
): Promise<{ stdout: string; stderr: string; code: number }> {
  return new Promise((resolve) => {
    const [file, ...args] = parseCommand(command);
    if (!file) {
      resolve({ stdout: "", stderr: "Command is empty", code: 2 });
      return;
    }
    const child = spawn(file, args, { cwd, shell: false });
    let stdout = "";
    let stderr = "";

    child.stdout.on("data", (data) => {
      const chunk = data.toString();
      stdout += chunk;
      process.stdout.write(data);
    });

    child.stderr.on("data", (data) => {
      const chunk = data.toString();
      stderr += chunk;
      process.stderr.write(data);
    });

    child.on("close", (code) => {
      resolve({ stdout, stderr, code: code ?? 1 });
    });
    
    child.on("error", (err) => {
      resolve({ stdout, stderr: err.message, code: 127 });
    });
  });
}

async function getGitModifiedFiles(cwd: string): Promise<{ modified: string[]; untracked: string[] }> {
  const { stdout } = await executeCommand("git status --porcelain", cwd);
  const modified: string[] = [];
  const untracked: string[] = [];

  for (const line of stdout.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    
    // git status format: XY path or XY "path"
    const status = trimmed.slice(0, 2);
    let path = trimmed.slice(3).trim();
    
    // Remove enclosing quotes if path has spaces
    if (path.startsWith('"') && path.endsWith('"')) {
      path = path.slice(1, -1);
    }
    
    if (status === "??" || status === " A") {
      untracked.push(path);
    } else {
      modified.push(path);
    }
  }

  return { modified, untracked };
}

export async function runAutoResearch(
  cwd: string,
  opts: {
    cmd: string;
    revertOnFail: boolean;
    metric?: string;
    json: boolean;
  },
): Promise<void> {
  const workspace = await findWorkspace(cwd);
  
  // 1. Gather initial git status to track what files we can revert if needed
  const beforeStatus = await getGitModifiedFiles(workspace);
  
  console.log(`Starting AutoResearch Loop...`);
  console.log(`Command to execute: "${opts.cmd}"`);
  console.log(`Revert on fail: ${opts.revertOnFail}`);
  if (opts.metric) console.log(`Evaluation metric regex: "${opts.metric}"`);
  
  // 2. Run the evaluation command
  const start = Date.now();
  const runResult = await executeCommand(opts.cmd, workspace);
  const durationMs = Date.now() - start;
  
  let success = runResult.code === 0;
  let metricPassed = true;
  let metricMessage = "";

  if (opts.metric) {
    const combinedOutput = runResult.stdout + "\n" + runResult.stderr;
    const regex = new RegExp(opts.metric);
    metricPassed = regex.test(combinedOutput);
    if (!metricPassed) {
      success = false;
      metricMessage = `Output did not match evaluation metric regex /${opts.metric}/`;
    } else {
      metricMessage = `Output successfully matched evaluation metric regex /${opts.metric}/`;
    }
  }

  const afterStatus = await getGitModifiedFiles(workspace);
  const newlyModified = afterStatus.modified.filter((f) => !beforeStatus.modified.includes(f));
  const newlyUntracked = afterStatus.untracked.filter((f) => !beforeStatus.untracked.includes(f));

  const totalFilesToRevert = [...newlyModified, ...newlyUntracked];

  let reverted = false;
  if (!success && opts.revertOnFail) {
    if (totalFilesToRevert.length > 0) {
      console.log(`\nFailure detected. Preserving ${totalFilesToRevert.length} changed files for inspection.`);
      console.log("Automatic rollback is disabled in the primary worktree. Use an isolated worktree for destructive experiments.");
    } else {
      console.log(`\nFailure detected, but no new modified or untracked files were detected.`);
    }
  }

  const out = {
    command: opts.cmd,
    success,
    exitCode: runResult.code,
    durationMs,
    metricPassed,
    metricMessage,
    reverted,
    revertedFiles: reverted ? totalFilesToRevert : [],
  };

  if (opts.json) {
    console.log(JSON.stringify(out, null, 2));
  } else {
    console.log(`\n--- AutoResearch Loop Result ---`);
    console.log(`Status: ${success ? "PASSED" : "FAILED"}`);
    console.log(`Exit Code: ${runResult.code}`);
    console.log(`Duration: ${(durationMs / 1000).toFixed(2)}s`);
    if (metricMessage) console.log(`Metric: ${metricMessage}`);
    if (reverted) console.log(`Reverted Files: ${totalFilesToRevert.join(", ")}`);
  }
  
  if (!success) {
    process.exit(1);
  }
}
