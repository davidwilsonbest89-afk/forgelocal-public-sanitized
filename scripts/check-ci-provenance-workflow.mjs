import { readFileSync } from "node:fs";

const ciWorkflow = readFileSync(".github/workflows/ci.yml", "utf8");
const releaseWorkflow = readFileSync(".github/workflows/release.yml", "utf8");
const codeowners = readFileSync(".github/CODEOWNERS", "utf8");
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
const releaseEnvironmentFragments = [
  "  release:\n    needs: build\n    runs-on: ubuntu-latest\n    environment:\n      name: production-release"
];
const releaseEnvironmentMissing = releaseEnvironmentFragments.filter(
  (fragment) => !releaseWorkflow.includes(fragment)
);
const primaryReviewOwners = "@boucheriechefimane-cmd @davidwilsonbest89-afk";
const independentReviewer = "@hajarbenmlih91-cloud";
const sensitiveCodeownerPaths = [
  "/.github/CODEOWNERS",
  "/.github/workflows/",
  "/scripts/check-component-rights.mjs",
  "/scripts/test-component-rights.mjs",
  "/scripts/create-component-rights-evidence.mjs",
  "/docs/component-rights-register.json",
  "/package.json",
  "/package-lock.json",
  "/pnpm-lock.yaml",
  "/go.mod",
  "/go.sum",
  "/release/",
  "/dist/"
];
const codeownersMissing = sensitiveCodeownerPaths.filter(
  (path) => !codeowners.includes(`${path} ${primaryReviewOwners}`)
);
const criticalCodeownerPaths = [
  "/.github/CODEOWNERS",
  "/.github/workflows/",
  "/scripts/check-component-rights.mjs",
  "/scripts/test-component-rights.mjs",
  "/scripts/create-component-rights-evidence.mjs",
  "/docs/component-rights-register.json",
  "/release/",
  "/dist/"
];
const independentReviewerMissing = criticalCodeownerPaths.filter(
  (path) =>
    !codeowners.includes(`${path} ${primaryReviewOwners} ${independentReviewer}`)
);

if (
  ciMissing.length > 0 ||
  releaseMissing.length > 0 ||
  releaseDependencyMissing.length > 0 ||
  releaseEnvironmentMissing.length > 0 ||
  codeownersMissing.length > 0 ||
  independentReviewerMissing.length > 0
) {
  console.error("ci provenance workflow: FAILED");
  for (const fragment of ciMissing) console.error(`- CI: fragment absent: ${fragment}`);
  for (const fragment of releaseMissing) {
    console.error(`- release: fragment absent: ${fragment}`);
  }
  for (const fragment of releaseDependencyMissing) {
    console.error(`- release: dépendance absente: ${JSON.stringify(fragment)}`);
  }
  for (const fragment of releaseEnvironmentMissing) {
    console.error(`- release: environnement protégé absent: ${JSON.stringify(fragment)}`);
  }
  for (const path of codeownersMissing) {
    console.error(`- CODEOWNERS: règle sensible absente ou owners incorrects: ${path}`);
  }
  for (const path of independentReviewerMissing) {
    console.error(`- CODEOWNERS: relectrice indépendante absente d’un chemin critique: ${path}`);
  }
  process.exit(1);
}

console.log(
  `ci provenance workflow: OK (${provenanceFragments.length} invariants CI + release, ${releaseDependencyFragments.length} dépendances de release, ${sensitiveCodeownerPaths.length} règles CODEOWNERS, ${criticalCodeownerPaths.length} chemins critiques à trois relecteurs)`
);
