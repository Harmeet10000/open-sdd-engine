import { readFile, access } from "node:fs/promises";
import { constants } from "node:fs";
import { join } from "node:path";
import YAML from "yaml";
import { DEFAULT_CONFIG, type SddConfig } from "./types.js";

export async function loadConfig(workspace: string): Promise<SddConfig> {
  const configPath = join(workspace, ".kiro", "config.yaml");
  try {
    await access(configPath, constants.R_OK);
    const raw = await readFile(configPath, "utf-8");
    const parsed = YAML.parse(raw) as Partial<SddConfig>;
    return {
      ...DEFAULT_CONFIG,
      ...parsed,
      loops: { ...DEFAULT_CONFIG.loops, ...parsed.loops },
      validation: { ...DEFAULT_CONFIG.validation, ...parsed.validation },
      research: {
        ...DEFAULT_CONFIG.research,
        ...parsed.research,
        graphify: { ...DEFAULT_CONFIG.research.graphify, ...parsed.research?.graphify },
        astGrep: { ...DEFAULT_CONFIG.research.astGrep, ...parsed.research?.astGrep },
        firecrawl: { ...DEFAULT_CONFIG.research.firecrawl, ...parsed.research?.firecrawl },
        context7: { ...DEFAULT_CONFIG.research.context7, ...parsed.research?.context7 },
      },
      memory: { ...DEFAULT_CONFIG.memory, ...parsed.memory, cognee: { ...DEFAULT_CONFIG.memory.cognee, ...parsed.memory?.cognee } },
    };
  } catch {
    return DEFAULT_CONFIG;
  }
}

export function configYamlTemplate(): string {
  return `version: 1
loops:
  change:
    maxBudgetHours: 2
    requireDesignForBreaking: true
  spec:
    requireDependencyGraphMinTasks: 6
    requireFilePlanMinTasks: 20
validation:
  defaultStrict: false
  concurrency: 6
research:
  graphify:
    enabled: true
    graphPath: graphify-out/graph.json
  astGrep:
    enabled: true
  firecrawl:
    enabled: true
  context7:
    enabled: true
    mcpServer: context7
memory:
  cognee:
    enabled: false
    baseUrl: http://localhost:8000
    dataset: open-sdd
    apiKeyEnv: COGNEE_API_KEY
`;
}
