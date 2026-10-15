import { readdir, readFile, access } from "node:fs/promises";
import { constants } from "node:fs";
import { join } from "node:path";
import { validateOkfFile } from "./okf.js";
import type { DiagnosticIssue, ValidationResult } from "./types.js";
import { issue } from "./reporter.js";

async function exists(path: string): Promise<boolean> {
  try {
    await access(path, constants.R_OK);
    return true;
  } catch {
    return false;
  }
}

export async function validateOkfDirectory(workspace: string, strict: boolean): Promise<ValidationResult> {
  const errors: DiagnosticIssue[] = [];
  const warnings: DiagnosticIssue[] = [];
  const knowledgeDir = join(workspace, ".kiro", "knowledge");

  if (!(await exists(knowledgeDir))) {
    return {
      id: "knowledge",
      loop: "change", // categorise under change loop for standard layout
      valid: true,
      errors: [],
      warnings: [],
    };
  }

  try {
    const files = await readdir(knowledgeDir, { withFileTypes: true });
    const mdFiles = files.filter((f) => f.isFile() && f.name.endsWith(".md")).map((f) => f.name);
    
    // Track file links to validate backlinks
    const fileLinkMap = new Map<string, string[]>();

    for (const name of mdFiles) {
      const filePath = join(knowledgeDir, name);
      const content = await readFile(filePath, "utf-8");
      
      const { valid, errors: fileErrors, warnings: fileWarnings, links } = await validateOkfFile(filePath, content);
      
      fileLinkMap.set(name, links);

      for (const e of fileErrors) {
        errors.push(
          issue("OKF_INVALID_FRONTMATTER", "error", `${name}: ${e}`, {
            file: join(".kiro", "knowledge", name),
            loop: "change",
          }),
        );
      }
      
      for (const w of fileWarnings) {
        warnings.push(
          issue("OKF_WARN", "warn", `${name}: ${w}`, {
            file: join(".kiro", "knowledge", name),
            loop: "change",
          }),
        );
      }
    }

    // Check for broken links
    for (const [sourceFile, targets] of fileLinkMap.entries()) {
      for (const target of targets) {
        // Target could be "architecture.md" or "sub/dir.md". We check relative to knowledgeDir.
        const targetPath = join(knowledgeDir, target);
        if (!(await exists(targetPath))) {
          const sev = strict ? "error" : "warn";
          const issueCode = "OKF_BROKEN_LINK";
          const msg = `Broken link in ${sourceFile}: target "${target}" does not exist`;
          
          const iss = issue(issueCode, sev, msg, {
            file: join(".kiro", "knowledge", sourceFile),
            loop: "change",
          });
          
          if (sev === "error") {
            errors.push(iss);
          } else {
            warnings.push(iss);
          }
        }
      }
    }
  } catch (err) {
    errors.push(
      issue("OKF_READ_ERROR", "error", `Failed to read knowledge directory: ${err instanceof Error ? err.message : err}`, {
        loop: "change",
      }),
    );
  }

  return {
    id: "knowledge",
    loop: "change",
    valid: errors.length === 0,
    errors,
    warnings,
  };
}
