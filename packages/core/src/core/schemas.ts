import { z } from "zod";

export const LoopKindSchema = z.enum(["change", "spec"]);
export const SeveritySchema = z.enum(["error", "warn", "info"]);
export const DiagnosticIssueSchema = z.object({
  code: z.string().regex(/^[A-Z][A-Z0-9_]+$/),
  severity: SeveritySchema,
  message: z.string().min(1),
  file: z.string().optional(),
  line: z.number().int().positive().optional(),
  loop: LoopKindSchema.optional(),
});
export const ValidationResultSchema = z.object({
  id: z.string().min(1),
  loop: LoopKindSchema,
  valid: z.boolean(),
  errors: z.array(DiagnosticIssueSchema),
  warnings: z.array(DiagnosticIssueSchema),
});
export const DiagnosticsOutputSchema = z.object({
  version: z.literal("1.0.0"),
  status: z.enum(["PASSED", "FAILED"]),
  results: z.array(ValidationResultSchema),
  summary: z.object({ total: z.number().int().nonnegative(), valid: z.number().int().nonnegative(), invalid: z.number().int().nonnegative(), errorCount: z.number().int().nonnegative(), warningCount: z.number().int().nonnegative() }),
});
export const TaskDependencyGraphSchema = z.object({
  tasks: z.array(z.object({ id: z.string().min(1), depends_on: z.array(z.string()) })),
});
export const SpecMetaSchema = z.object({
  loop: z.literal("spec"),
  estimatedDays: z.number().positive(),
  complexity: z.enum(["low", "medium", "high"]),
}).strict();

export type DiagnosticsOutputInput = z.input<typeof DiagnosticsOutputSchema>;
