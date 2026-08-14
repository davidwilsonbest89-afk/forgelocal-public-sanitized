import { readFileSync } from "node:fs";

const ciWorkflow = readFileSync(".github/workflows/ci.yml", "utf8");
const releaseWorkflow = readFileSync(".github/workflows/release.yml", "utf8");
const provenanceFragments = [
  "actions/setup-node@v6",
  'node-version: "22"',
  "node scripts/check-component-rights.mjs",
  "node scripts/test-component-rights.mjs",
  "node scripts/create-component-rights-evidence.mjs",
  "actions/upload-artifact@v4",
  "name: component-provenance-${{ github.sha }}",
  "path: .ci-artifacts/component-provenance/",
  "if-no-files-found: error",
  "retention-days: 90"
];

const ciMissing = provenanceFragments.filter((fragment) => !ciWorkflow.includes(fragment));
const releaseMissing = provenanceFragments.filter((fragment) => !releaseWorkflow.includes(fragment));
const releaseDependencyFragments = [
  "  verify:\n    needs: provenance",
  "  build:\n    needs: verify",
  "  release:\n    needs: build"
];
const releaseDependencyMissing = releaseDependencyFragments.filter(
  (fragment) => !releaseWorkflow.includes(fragment)
);

if (ciMissing.length > 0 || releaseMissing.length > 0 || releaseDependencyMissing.length > 0) {
  console.error("ci provenance workflow: FAILED");
  for (const fragment of ciMissing) console.error(`- CI: fragment absent: ${fragment}`);
  for (const fragment of releaseMissing) {
    console.error(`- release: fragment absent: ${fragment}`);
  }
  for (const fragment of releaseDependencyMissing) {
    console.error(`- release: dépendance absente: ${JSON.stringify(fragment)}`);
  }
  process.exit(1);
}

console.log(
  `ci provenance workflow: OK (${provenanceFragments.length} invariants CI + release, ${releaseDependencyFragments.length} dépendances de release)`
);
