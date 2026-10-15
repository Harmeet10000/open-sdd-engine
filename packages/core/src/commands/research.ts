import { findWorkspace } from "../core/discovery.js";
import { loadConfig } from "../core/config.js";
import { runResearch, type ResearchKind } from "../research/tools.js";

export async function runResearchCommand(
  cwd: string,
  kind: ResearchKind,
  query: string,
  opts: { language?: string; library?: string; json: boolean },
): Promise<number> {
  const workspace = await findWorkspace(cwd);
  const config = await loadConfig(workspace);
  const result = await runResearch(kind, workspace, config, {
    query,
    language: opts.language,
    library: opts.library,
  });

  if (opts.json) {
    console.log(JSON.stringify(result, null, 2));
  } else {
    console.log(`Tool: ${result.tool} | Success: ${result.success}`);
    if (result.output) console.log(result.output);
    if (result.agentHint) console.log(`\nAgent hint:\n${result.agentHint}`);
    if (result.error) console.error(result.error);
  }
  return result.success ? 0 : 1;
}
