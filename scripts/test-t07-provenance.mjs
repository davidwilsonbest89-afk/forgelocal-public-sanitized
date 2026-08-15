import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { join, relative } from "node:path";

const root = process.cwd();
const registry = JSON.parse(readFileSync(join(root, "docs", "component-rights-register.json"), "utf8"));
const sbom = JSON.parse(readFileSync(join(root, "docs", "T07_CAMOFLOX_PROVENANCE_SBOM.spdx.json"), "utf8"));
const dependencyInventory = JSON.parse(readFileSync(join(root, "docs", "T07_CAMOFLOX_DEPENDENCY_INVENTORY.json"), "utf8"));
const candidate = registry.components.find((component) => component.id === "camoflox-audit-source");

function expect(condition, message) {
  if (!condition) throw new Error(message);
}

function walk(path) {
  if (!existsSync(path)) return [];
  const stat = statSync(path);
  if (stat.isFile()) return [path];
  return readdirSync(path, { withFileTypes: true }).flatMap((entry) => walk(join(path, entry.name)));
}

expect(candidate, "entrée camoflox-audit-source absente");
expect(candidate.integration_state === "provenance-qualification-blocked", "Camoflox doit rester bloqué et non intégré");
expect(candidate.rights_status === "authorized", "l’autorisation privée déclarée doit rester explicite");
expect(candidate.source_revision === "sha256:dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2", "empreinte archive inattendue");
expect(candidate.t07_provenance.source_commit === null, "un commit source non vérifié ne doit pas être inventé");
expect(candidate.t07_provenance.source_commit_status === "unknown-blocking", "le commit absent doit rester bloquant");
expect(candidate.t07_provenance.root_license_status === "absent-blocking", "la licence racine absente doit rester bloquante");
expect(candidate.t07_provenance.scan_status === "unknown-blocking", "l’alerte Gitleaks doit rester bloquante");
expect(candidate.t07_provenance.candidate_modules.length === 2, "seuls deux modules conceptuels sont autorisés à l’étude T07");

const selected = new Set(candidate.t07_provenance.candidate_modules.map((module) => module.source_path));
expect(selected.has("lib/concurrency.js") && selected.has("lib/global-action-limiter.js"), "modules conceptuels T07 inattendus");
for (const module of candidate.t07_provenance.candidate_modules) {
  expect(module.t07_decision === "conceptual-review-only", `${module.source_path}: portage interdit dans T07`);
  expect(module.integration_state === "not-integrated", `${module.source_path}: intégration interdite dans T07`);
}

for (const path of ["lib/profile-launch-queue.js", "lib/process-isolation.js", "lib/browser-lifecycle.js"]) {
  expect(candidate.t07_provenance.out_of_scope_modules.includes(path), `${path}: doit rester hors périmètre T07`);
}

expect(Array.isArray(sbom.packages) && sbom.packages.length === 710, "SBOM T07 incomplète ou inattendue");
expect(sbom.packages[0].name === "camoflox", "racine SBOM inattendue");
expect(dependencyInventory.direct_dependencies.length === 8, "inventaire runtime direct incomplet");
expect(dependencyInventory.direct_dev_dependencies.length === 4, "inventaire développement direct incomplet");
expect(dependencyInventory.lockfile_package_entries === 709, "inventaire lockfile incomplet");

const prohibited = /camoflox-v28-etape2|camoflox-FINAL\.zip|lib\/concurrency\.js|lib\/global-action-limiter\.js/iu;
for (const controlledPath of ["cmd", "internal", "extension", "assets", "runtime", "third_party", "vendor"]) {
  for (const absolutePath of walk(join(root, controlledPath))) {
    const contents = readFileSync(absolutePath, "utf8");
    expect(!prohibited.test(contents), `référence de source Camoflox interdite dans ${relative(root, absolutePath)}`);
  }
}

console.log("t07-provenance: BLOCKED_AS_EXPECTED (no Camoflox integration, runtime, queue, lock, port, or launch)");
