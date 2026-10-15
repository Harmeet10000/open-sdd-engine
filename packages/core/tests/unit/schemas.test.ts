import { describe, expect, it } from "vitest";
import { DiagnosticsOutputSchema, SpecMetaSchema, TaskDependencyGraphSchema } from "../../src/core/schemas.js";

describe("public schemas", () => {
  it("rejects malformed diagnostic codes", () => {
    expect(() => DiagnosticsOutputSchema.parse({ version: "1.0.0", status: "PASSED", results: [{ id: "x", loop: "change", valid: true, errors: [{ code: "bad-code", severity: "error", message: "bad" }], warnings: [] }], summary: { total: 1, valid: 1, invalid: 0, errorCount: 1, warningCount: 0 } })).toThrow();
  });

  it("requires task IDs and dependency arrays", () => {
    expect(TaskDependencyGraphSchema.safeParse({ tasks: [{ id: "1", depends_on: [] }] }).success).toBe(true);
    expect(TaskDependencyGraphSchema.safeParse({ tasks: [{ id: "", depends_on: "1" }] }).success).toBe(false);
  });

  it("validates the SDD spec schema", () => {
    expect(SpecMetaSchema.safeParse({ loop: "spec", estimatedDays: 3, complexity: "high" }).success).toBe(true);
    expect(SpecMetaSchema.safeParse({ loop: "change", estimatedDays: 0, complexity: "wild" }).success).toBe(false);
  });
});
