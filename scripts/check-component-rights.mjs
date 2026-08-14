import { createHash } from "node:crypto";
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { join, relative } from "node:path";

const root = process.cwd();
const registryPath = join(root, "docs", "component-rights-register.json");
const registry = JSON.parse(readFileSync(registryPath, "utf8"));
const errors = [];
const componentByID = new Map(registry.components.map((component) => [component.id, component]));
const integratedStatuses = new Set(registry.policy.integrated_statuses);

function fail(message) {
  errors.push(message);
}

function sha256(path) {
  return createHash("sha256").update(readFileSync(path)).digest("hex");
}

function walk(path) {
  if (!existsSync(path)) return [];
  const stat = statSync(path);
  if (stat.isFile()) return [path];
  return readdirSync(path, { withFileTypes: true }).flatMap((entry) => walk(join(path, entry.name)));
}

for (const component of registry.components) {
  const integrated = component.integration_state === "integrated";
  if (integrated && !integratedStatuses.has(component.rights_status)) {
    fail(`${component.id}: statut ${component.rights_status} interdit pour un composant intégré`);
  }
  if (integrated && !component.source_revision) {
    fail(`${component.id}: révision exacte manquante`);
  }
  if (component.rights_status === "authorized" && integrated) {
    if (!component.evidence_ref || !component.owner || !Array.isArray(component.scope) || component.scope.length === 0) {
      fail(`${component.id}: preuve, responsable ou portée d’autorisation manquant`);
    }
    if (!Array.isArray(component.dependency_inventory_ref) || component.dependency_inventory_ref.length === 0) {
      fail(`${component.id}: inventaire de dépendances manquant`);
    }
  }
  if (component.rights_status === "not_required" && (!component.first_party || !component.owner || !component.source_revision)) {
    fail(`${component.id}: not_required est réservé à une source first-party avec responsable et révision`);
  }
  if (["denied", "unknown"].includes(component.rights_status) && integrated) {
    fail(`${component.id}: source ${component.rights_status} ne peut pas être intégrée`);
  }
}

for (const input of registry.build_inputs) {
  const absolutePath = join(root, input.path);
  if (!existsSync(absolutePath)) {
    fail(`input build introuvable: ${input.path}`);
    continue;
  }
  const actual = sha256(absolutePath);
  if (actual !== input.sha256) fail(`empreinte inattendue: ${input.path}`);
  for (const owner of input.owners) {
    const component = componentByID.get(owner);
    if (!component) fail(`${input.path}: propriétaire absent du registre: ${owner}`);
    else if (component.integration_state !== "integrated" || !integratedStatuses.has(component.rights_status)) {
      fail(`${input.path}: propriétaire non autorisé pour un input build: ${owner}`);
    }
  }
}

const markers = registry.forbidden_markers.map((marker) => marker.toLowerCase());
for (const controlledPath of registry.controlled_paths) {
  for (const absolutePath of walk(join(root, controlledPath))) {
    const relativePath = relative(root, absolutePath);
    const text = readFileSync(absolutePath, "utf8").toLowerCase();
    for (const marker of markers) {
      if (relativePath.toLowerCase().includes(marker) || text.includes(marker)) {
        fail(`marqueur interdit ${marker} trouvé dans ${relativePath}`);
      }
    }
  }
}

if (errors.length > 0) {
  console.error("component-rights: FAILED");
  for (const error of errors) console.error(`- ${error}`);
  process.exit(1);
}

console.log(`component-rights: OK (${registry.components.length} composants, ${registry.build_inputs.length} inputs vérifiés)`);
