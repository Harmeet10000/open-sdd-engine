import type { DiagnosticIssue, DiagnosticsOutput, ValidationResult } from "./types.js";
import { DiagnosticsOutputSchema } from "./schemas.js";

export function issue(
  code: string,
  severity: DiagnosticIssue["severity"],
  message: string,
  extra?: Partial<DiagnosticIssue>,
): DiagnosticIssue {
  return { code, severity, message, ...extra };
}

export function mergeResults(results: ValidationResult[]): DiagnosticsOutput {
  const errorCount = results.reduce((n, r) => n + r.errors.length, 0);
  const warningCount = results.reduce((n, r) => n + r.warnings.length, 0);
  const invalid = results.filter((r) => !r.valid).length;
  return DiagnosticsOutputSchema.parse({
    version: "1.0.0",
    status: invalid > 0 || errorCount > 0 ? "FAILED" : "PASSED",
    results,
    summary: {
      total: results.length,
      valid: results.length - invalid,
      invalid,
      errorCount,
      warningCount,
    },
  });
}

export function printDiagnostics(output: DiagnosticsOutput, json: boolean, pretty = false): void {
  if (json) {
    console.log(JSON.stringify(output, null, pretty ? 2 : 0));
    return;
  }
  for (const r of output.results) {
    console.log(`\n${r.loop.toUpperCase()} ${r.id}: ${r.valid ? "PASSED" : "FAILED"}`);
    for (const e of r.errors) {
      const loc = e.file ? ` (${e.file}${e.line ? `:${e.line}` : ""})` : "";
      console.log(`  ✗ [${e.code}] ${e.message}${loc}`);
    }
    for (const w of r.warnings) {
      const loc = w.file ? ` (${w.file}${w.line ? `:${w.line}` : ""})` : "";
      console.log(`  ⚠ [${w.code}] ${w.message}${loc}`);
    }
  }
  console.log(
    `\nSummary: ${output.summary.valid}/${output.summary.total} valid, ${output.summary.errorCount} errors, ${output.summary.warningCount} warnings`,
  );
}

export function exitCodeFromOutput(output: DiagnosticsOutput, strict: boolean): number {
  if (output.summary.errorCount > 0) return 1;
  if (strict && output.summary.warningCount > 0) return 1;
  return 0;
}
