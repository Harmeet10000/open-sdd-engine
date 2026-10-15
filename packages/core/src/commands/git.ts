import { createInterface } from "node:readline/promises";
import { findWorkspace } from "../core/discovery.js";
import { abortConflict, addWorktree, commitFiles, conflictStatus, createBranch, createPullRequest, gitStatus, loadProfile, saveProfile, type CodingProfile } from "../git/operations.js";

function print(output: unknown, json: boolean) { console.log(json ? JSON.stringify(output, null, 2) : typeof output === "string" ? output : JSON.stringify(output, null, 2)); }
export async function runGitStatus(cwd: string, json: boolean) { print(await gitStatus(cwd), json); }
export async function runGitBranch(cwd: string, branch: string, base: string | undefined, json: boolean) { print(await createBranch(cwd, branch, base), json); }
export async function runGitWorktree(cwd: string, path: string, branch: string, base: string | undefined, json: boolean) { print(await addWorktree(cwd, path, branch, base), json); }
export async function runGitCommit(cwd: string, message: string, files: string[], skipChecks: boolean, json: boolean) { print(await commitFiles(cwd, message, files, !skipChecks), json); }
export async function runConflictStatus(cwd: string, json: boolean) { print(await conflictStatus(cwd), json); }
export async function runConflictAbort(cwd: string, yes: boolean, json: boolean) { print(await abortConflict(cwd, yes), json); }
export async function runGitPr(cwd: string, title: string, body: string, base: string, push: boolean, json: boolean) { print(await createPullRequest(cwd, title, body, base, push), json); }
export async function runProfileShow(cwd: string, json: boolean) { print(await loadProfile(await findWorkspace(cwd)), json); }
export async function runProfileInit(cwd: string, json: boolean) {
  const workspace = await findWorkspace(cwd); if (!process.stdin.isTTY) throw new Error("sdd profile init needs a terminal. Use profile configuration file in automation.");
  const prompt = createInterface({ input: process.stdin, output: process.stdout });
  try {
    const ask = async (question: string, fallback: string) => (await prompt.question(`${question} [${fallback}]: `)).trim() || fallback;
    const profile: CodingProfile = { languages: (await ask("Preferred languages (comma-separated)", "TypeScript")).split(",").map((value) => value.trim()).filter(Boolean), guidelines: await ask("Coding guidelines and taste", defaultProfileText()), linters: (await ask("Linters (comma-separated, optional)", "")).split(",").map((value) => value.trim()).filter(Boolean), formatter: await ask("Formatter", ""), testCommand: await ask("Test command", "bun test"), branchPrefix: await ask("Branch prefix", "sdd/") };
    print(await saveProfile(workspace, profile), json);
  } finally { prompt.close(); }
}
function defaultProfileText() { return "Prefer small readable diffs, explicit errors, and tests for behavior changes."; }
