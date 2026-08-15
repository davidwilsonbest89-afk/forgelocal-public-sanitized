import { createHash } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";

const [packagePath, lockPath, outputPath] = process.argv.slice(2);
if (!packagePath || !lockPath || !outputPath) {
  console.error("usage: node scripts/generate-t07-camoflox-dependency-inventory.mjs <package.json> <package-lock.json> <output.json>");
  process.exit(2);
}

const read = (path) => readFileSync(path);
const sha256 = (value) => createHash("sha256").update(value).digest("hex");
const packageRaw = read(packagePath);
const lockRaw = read(lockPath);
const packageJSON = JSON.parse(packageRaw);
const lock = JSON.parse(lockRaw);

function direct(section, declared) {
  return Object.entries(declared ?? {}).sort(([left], [right]) => left.localeCompare(right)).map(([name, requested]) => {
    const exact = lock.packages?.[`node_modules/${name}`];
    return {
      name,
      scope: section,
      declared_range: requested,
      locked_version: exact?.version ?? "MISSING",
      resolved: exact?.resolved ?? "NOASSERTION",
      integrity: exact?.integrity ?? "NOASSERTION",
      license_declared_in_lock: exact?.license ?? "NOASSERTION",
    };
  });
}

const inventory = {
  schema_version: "1.0",
  purpose: "T07 provenance-only dependency inventory; no package was installed, executed, imported, or integrated.",
  archive_root_package: `${packageJSON.name ?? "unknown"}@${packageJSON.version ?? "unknown"}`,
  package_json_sha256: sha256(packageRaw),
  package_lock_sha256: sha256(lockRaw),
  direct_dependencies: direct("runtime", packageJSON.dependencies),
  direct_dev_dependencies: direct("development", packageJSON.devDependencies),
  lockfile_package_entries: Object.keys(lock.packages ?? {}).filter((path) => path.startsWith("node_modules/")).length,
};

mkdirSync(dirname(resolve(outputPath)), { recursive: true });
writeFileSync(outputPath, `${JSON.stringify(inventory, null, 2)}\n`);
console.log(`t07-dependency-inventory: OK runtime=${inventory.direct_dependencies.length} development=${inventory.direct_dev_dependencies.length} locked=${inventory.lockfile_package_entries}`);
