import { findWorkspace } from "../core/discovery.js";
import { loadConfig } from "../core/config.js";
import { recall, remember } from "../memory/cognee.js";
export async function runMemoryRemember(cwd: string, text: string, scope: string, tags: string[], json: boolean) { const workspace = await findWorkspace(cwd); const output = await remember(workspace, await loadConfig(workspace), { text, scope, tags, source: "sdd-cli" }); console.log(json ? JSON.stringify(output, null, 2) : `Remembered ${output.id}`); }
export async function runMemorySearch(cwd: string, query: string, json: boolean) { const workspace = await findWorkspace(cwd); const output = await recall(workspace, await loadConfig(workspace), query); console.log(json ? JSON.stringify(output, null, 2) : output.records.map((record) => `- [${record.scope}] ${record.text}`).join("\n")); }
