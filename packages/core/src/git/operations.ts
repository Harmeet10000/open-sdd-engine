import { spawn } from "node:child_process";
import { access, writeFile } from "node:fs/promises";
import { constants } from "node:fs";
import { join, resolve } from "node:path";
import YAML from "yaml";

export interface GitResult { code: number; stdout: string; stderr: string; }
export interface CodingProfile { languages: string[]; guidelines: string; linters: string[]; formatter: string; testCommand: string; branchPrefix: string; }
const defaultProfile: CodingProfile = { languages: ["TypeScript"], guidelines: "Prefer small, readable changes with tests.", linters: [], formatter: "", testCommand: "", branchPrefix: "sdd/" };

export function runGit(cwd: string, args: string[]): Promise<GitResult> {
  return new Promise((resolveResult) => {
    const child = spawn("git", args, { cwd, shell: false }); let stdout = ""; let stderr = "";
    child.stdout.on("data", (chunk: Buffer) => { stdout += chunk.toString(); });
    child.stderr.on("data", (chunk: Buffer) => { stderr += chunk.toString(); });
    child.on("close", (code) => resolveResult({ code: code ?? 1, stdout, stderr }));
    child.on("error", (error) => resolveResult({ code: 127, stdout, stderr: error.message }));
  });
}
async function requireRepository(cwd: string): Promise<void> { const result = await runGit(cwd, ["rev-parse", "--is-inside-work-tree"]); if (result.code !== 0 || result.stdout.trim() !== "true") throw new Error("Not inside a Git worktree."); }
async function requireClean(cwd: string): Promise<void> { const result = await runGit(cwd, ["status", "--porcelain"]); if (result.code !== 0) throw new Error(result.stderr); if (result.stdout.trim()) throw new Error("Worktree is dirty. Commit, stash, or use a separate worktree before changing branches."); }
export async function gitStatus(cwd: string) { await requireRepository(cwd); const [branch, status, conflicts] = await Promise.all([runGit(cwd, ["branch", "--show-current"]), runGit(cwd, ["status", "--porcelain"]), runGit(cwd, ["diff", "--name-only", "--diff-filter=U"])]); return { branch: branch.stdout.trim(), clean: !status.stdout.trim(), entries: status.stdout.trim().split("\n").filter(Boolean), conflicts: conflicts.stdout.trim().split("\n").filter(Boolean) }; }
export async function createBranch(cwd: string, branch: string, base?: string) { await requireRepository(cwd); await requireClean(cwd); if (!/^[A-Za-z0-9._/-]+$/.test(branch)) throw new Error("Invalid branch name."); const result = await runGit(cwd, base ? ["switch", "-c", branch, base] : ["switch", "-c", branch]); if (result.code) throw new Error(result.stderr); return { branch, base: base ?? "HEAD" }; }
export async function addWorktree(cwd: string, path: string, branch: string, base?: string) { await requireRepository(cwd); const target = resolve(cwd, path); try { await access(target, constants.F_OK); throw new Error(`Worktree path already exists: ${target}`); } catch (error) { if (error instanceof Error && error.message.startsWith("Worktree")) throw error; }
  const result = await runGit(cwd, ["worktree", "add", "-b", branch, target, base ?? "HEAD"]); if (result.code) throw new Error(result.stderr); return { path: target, branch, base: base ?? "HEAD" }; }
export async function commitFiles(cwd: string, message: string, files: string[], runChecks: boolean) { await requireRepository(cwd); if (!message.trim()) throw new Error("Commit message is required."); if (!files.length) throw new Error("Pass explicit --files paths; SDD never stages the whole worktree by default."); if (runChecks) { const check = await runGit(cwd, ["diff", "--check"]); if (check.code) throw new Error(`Whitespace check failed: ${check.stdout || check.stderr}`); }
  const staged = (await runGit(cwd, ["diff", "--cached", "--name-only"])).stdout.trim().split("\n").filter(Boolean);
  const selected = new Set(files);
  const unexpected = staged.filter((file) => !selected.has(file));
  if (unexpected.length) throw new Error(`Refusing commit: unrelated staged files exist: ${unexpected.join(", ")}. Unstage or include them explicitly.`);
  const add = await runGit(cwd, ["add", "--", ...files]); if (add.code) throw new Error(add.stderr); const commit = await runGit(cwd, ["commit", "-m", message]); if (commit.code) throw new Error(commit.stderr); return { message, files, commit: (await runGit(cwd, ["rev-parse", "HEAD"])).stdout.trim() }; }
export async function conflictStatus(cwd: string) { await requireRepository(cwd); const status = await runGit(cwd, ["status", "--porcelain"]); const files = (await runGit(cwd, ["diff", "--name-only", "--diff-filter=U"])).stdout.trim().split("\n").filter(Boolean); return { inProgress: /^(UU|AA|DD|AU|UA|DU|UD)/m.test(status.stdout), files, status: status.stdout.trim() }; }
export async function abortConflict(cwd: string, yes: boolean) { if (!yes) throw new Error("Conflict abort requires --yes. No files changed."); await requireRepository(cwd); for (const args of [["merge", "--abort"], ["rebase", "--abort"], ["cherry-pick", "--abort"]]) { const result = await runGit(cwd, args); if (result.code === 0) return { aborted: args[0] }; } throw new Error("No abortable merge, rebase, or cherry-pick is in progress."); }
export async function createPullRequest(cwd: string, title: string, body: string, base: string, push: boolean) {
  await requireRepository(cwd); await requireClean(cwd); const branch = (await runGit(cwd, ["branch", "--show-current"])).stdout.trim(); if (!branch) throw new Error("Detached HEAD cannot open a pull request.");
  if (push) { const pushed = await runGit(cwd, ["push", "-u", "origin", branch]); if (pushed.code) throw new Error(pushed.stderr); }
  const result = await new Promise<GitResult>((resolveResult) => { const child = spawn("gh", ["pr", "create", "--title", title, "--body", body, "--base", base], { cwd, shell: false }); let stdout = ""; let stderr = ""; child.stdout.on("data", (chunk: Buffer) => stdout += chunk.toString()); child.stderr.on("data", (chunk: Buffer) => stderr += chunk.toString()); child.on("close", (code) => resolveResult({ code: code ?? 1, stdout, stderr })); child.on("error", (error) => resolveResult({ code: 127, stdout, stderr: error.message })); });
  if (result.code) throw new Error(`PR creation failed. ${result.stderr || result.stdout} ${push ? "" : "Use --push if the branch has not been pushed."}`.trim()); return { branch, base, url: result.stdout.trim() };
}
export async function saveProfile(workspace: string, profile: CodingProfile) { const path = join(workspace, ".kiro", "coding-profile.yaml"); await writeFile(path, YAML.stringify(profile), "utf-8"); return profile; }
export async function loadProfile(workspace: string): Promise<CodingProfile> { try { return { ...defaultProfile, ...(YAML.parse(await (await import("node:fs/promises")).readFile(join(workspace, ".kiro", "coding-profile.yaml"), "utf-8")) as Partial<CodingProfile>) }; } catch { return defaultProfile; } }
