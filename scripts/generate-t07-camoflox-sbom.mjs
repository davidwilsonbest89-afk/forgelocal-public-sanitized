import { createHash } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, relative, resolve } from "node:path";

const [packagePath, lockPath, outputPath] = process.argv.slice(2);
if (!packagePath || !lockPath || !outputPath) {
  console.error("usage: node scripts/generate-t07-camoflox-sbom.mjs <package.json> <package-lock.json> <output.spdx.json>");
  process.exit(2);
}

const packageJSON = JSON.parse(readFileSync(packagePath, "utf8"));
const lock = JSON.parse(readFileSync(lockPath, "utf8"));
const sha256 = (value) => createHash("sha256").update(value).digest("hex");
const packageHash = sha256(readFileSync(packagePath));
const lockHash = sha256(readFileSync(lockPath));
const directDependencies = Object.keys(packageJSON.dependencies ?? {}).sort();

const packages = [{
  SPDXID: "SPDXRef-Camoflox-Archive-Root",
  name: packageJSON.name ?? "camoflox",
  versionInfo: packageJSON.version ?? "NOASSERTION",
  downloadLocation: "NOASSERTION",
  filesAnalyzed: false,
  licenseConcluded: "NOASSERTION",
  licenseDeclared: packageJSON.license ?? "NOASSERTION",
  externalRefs: [{
    referenceCategory: "SECURITY",
    referenceType: "other",
    referenceLocator: `sha256:${packageHash}`,
  }],
}];

for (const [path, entry] of Object.entries(lock.packages ?? {}).sort(([left], [right]) => left.localeCompare(right))) {
  if (!path.startsWith("node_modules/")) continue;
  const name = relative("node_modules", path);
  packages.push({
    SPDXID: `SPDXRef-NPM-${sha256(`${name}@${entry.version ?? "unknown"}`).slice(0, 20)}`,
    name,
    versionInfo: entry.version ?? "NOASSERTION",
    downloadLocation: entry.resolved ?? "NOASSERTION",
    checksums: entry.integrity ? [{ algorithm: "SHA512", checksumValue: entry.integrity.replace(/^sha512-/, "") }] : [],
    filesAnalyzed: false,
    licenseConcluded: "NOASSERTION",
    licenseDeclared: entry.license ?? "NOASSERTION",
    comment: entry.dev ? "development dependency" : "runtime or transitive dependency",
  });
}

const sbom = {
  SPDXID: "SPDXRef-DOCUMENT",
  spdxVersion: "SPDX-2.3",
  name: "ForgeLocal-T07-Camoflox-Provenance-Only-SBOM",
  dataLicense: "CC0-1.0",
  documentNamespace: `urn:sha256:${sha256(`${packageHash}:${lockHash}`)}`,
  creationInfo: {
    creators: ["Tool: ForgeLocal T07 passive SBOM generator"],
    created: "2026-08-15T00:00:00Z",
  },
  documentDescribes: ["SPDXRef-Camoflox-Archive-Root"],
  packages,
  annotations: [{
    annotationType: "OTHER",
    annotator: "Tool: ForgeLocal T07 passive SBOM generator",
    annotationDate: "2026-08-15T00:00:00Z",
    comment: `Provenance-only inventory. direct_dependencies=${directDependencies.join(",")}; package_json_sha256=${packageHash}; package_lock_sha256=${lockHash}; package_count=${packages.length}. No package was installed, executed, imported or integrated.`,
  }],
};

mkdirSync(dirname(resolve(outputPath)), { recursive: true });
writeFileSync(outputPath, `${JSON.stringify(sbom, null, 2)}\n`);
console.log(`t07-sbom: OK packages=${packages.length} direct_dependencies=${directDependencies.length}`);
