export type LoopKind = "change" | "spec";
export type Severity = "error" | "warn" | "info";

export interface DiagnosticIssue {
  code: string;
  severity: Severity;
  message: string;
  file?: string;
  line?: number;
  loop?: LoopKind;
}

export interface ValidationResult {
  id: string;
  loop: LoopKind;
  valid: boolean;
  errors: DiagnosticIssue[];
  warnings: DiagnosticIssue[];
}

export interface DiagnosticsOutput {
  version: "1.0.0";
  status: "PASSED" | "FAILED";
  results: ValidationResult[];
  summary: {
    total: number;
    valid: number;
    invalid: number;
    errorCount: number;
    warningCount: number;
  };
}

export interface TaskDependencyGraph {
  tasks: Array<{ id: string; depends_on: string[] }>;
}

export interface WaveResult {
  specSlug: string;
  waves: string[][];
  parallelism: number[];
}

export interface ResearchToolConfig {
  enabled: boolean;
}

export interface SddConfig {
  version: number;
  loops: {
    change: {
      maxBudgetHours: number;
      requireDesignForBreaking: boolean;
    };
    spec: {
      requireDependencyGraphMinTasks: number;
      requireFilePlanMinTasks: number;
    };
  };
  validation: {
    defaultStrict: boolean;
    concurrency: number;
  };
  research: {
    graphify: ResearchToolConfig & { graphPath: string };
    astGrep: ResearchToolConfig;
    firecrawl: ResearchToolConfig;
    context7: ResearchToolConfig & { mcpServer: string };
  };
  memory: {
    cognee: { enabled: boolean; baseUrl: string; dataset: string; apiKeyEnv: string };
  };
}

export const DEFAULT_CONFIG: SddConfig = {
  version: 1,
  loops: {
    change: { maxBudgetHours: 2, requireDesignForBreaking: true },
    spec: { requireDependencyGraphMinTasks: 6, requireFilePlanMinTasks: 20 },
  },
  validation: { defaultStrict: false, concurrency: 6 },
  research: {
    graphify: { enabled: true, graphPath: "graphify-out/graph.json" },
    astGrep: { enabled: true },
    firecrawl: { enabled: true },
    context7: { enabled: true, mcpServer: "context7" },
  },
  memory: { cognee: { enabled: false, baseUrl: "http://localhost:8000", dataset: "open-sdd", apiKeyEnv: "COGNEE_API_KEY" } },
};
