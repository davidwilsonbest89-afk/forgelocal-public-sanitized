import { createHash } from "node:crypto";
import { copyFileSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { join } from "node:path";

const root = process.cwd();
const registryRelativePath = "docs/component-rights-register.json";
const registryPath = join(root, registryRelativePath);
const artifactDirectory = join(root, ".ci-artifacts", "component-provenance");
const registryBytes = readFileSync(registryPath);
const registry = JSON.parse(registryBytes.toString("utf8"));
const commit = process.env.GITHUB_SHA || execFileSync("git", ["rev-parse", "HEAD"], { cwd: root, encoding: "utf8" }).trim();

mkdirSync(artifactDirectory, { recursive: true });
copyFileSync(registryPath, join(artifactDirectory, "component-rights-register.json"));

const evidence = {
  schema_version: "1.0",
  evidence_kind: "component-provenance-ci",
  registry_id: registry.registry_id,
  registry_path: registryRelativePath,
  registry_sha256: createHash("sha256").update(registryBytes).digest("hex"),
  source_commit: commit,
  checks: [
    "node scripts/check-component-rights.mjs",
    "node scripts/test-component-rights.mjs"
  ],
  release_authorization: "none",
  release_status: "PUBLIC_RELEASE_BLOCKED"
};

writeFileSync(join(artifactDirectory, "component-provenance-evidence.json"), `${JSON.stringify(evidence, null, 2)}\n`);
console.log(`component-provenance evidence: ${evidence.registry_sha256} @ ${commit}`);
