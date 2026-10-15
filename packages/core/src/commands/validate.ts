import { findWorkspace, listChanges, listSpecs, loadChange, loadSpec } from "../core/discovery.js";
import { loadConfig } from "../core/config.js";
import { mergeResults, printDiagnostics, exitCodeFromOutput } from "../core/reporter.js";
import { validateChange } from "../change/validators/change-validator.js";
import { validateSpec } from "../spec/validators/spec-validator.js";
import { validateOkfDirectory } from "../core/okf-validator.js";

export async function runValidate(opts: {
  cwd: string;
  name?: string;
  changes: boolean;
  specs: boolean;
  okf?: boolean;
  all: boolean;
  strict: boolean;
  json: boolean;
}): Promise<number> {
  const workspace = await findWorkspace(opts.cwd);
  const config = await loadConfig(workspace);
  const strict = opts.strict || config.validation.defaultStrict;
  const results = [];

  const runChanges = opts.all || opts.changes || (!opts.changes && !opts.specs && !opts.okf);
  const runSpecs = opts.all || opts.specs || (!opts.changes && !opts.specs && !opts.okf);
  const runOkf = opts.all || opts.okf || (!opts.changes && !opts.specs && !opts.okf);

  if (runChanges) {
    const ids = opts.name ? [opts.name] : await listChanges(workspace);
    for (const id of ids) {
      results.push(await validateChange(await loadChange(workspace, id), config, strict));
    }
  }
  if (runSpecs) {
    const slugs = opts.name ? [opts.name] : await listSpecs(workspace);
    for (const slug of slugs) {
      results.push(validateSpec(await loadSpec(workspace, slug), config, strict));
    }
  }
  if (runOkf) {
    results.push(await validateOkfDirectory(workspace, strict));
  }

  const output = mergeResults(results);
  printDiagnostics(output, opts.json, opts.json);
  return exitCodeFromOutput(output, strict);
}

export async function runDiagnostics(opts: Parameters<typeof runValidate>[0]): Promise<number> {
  return runValidate(opts);
}
