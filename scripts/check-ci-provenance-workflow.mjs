import { readFileSync } from "node:fs";

const workflow = readFileSync(".github/workflows/ci.yml", "utf8");
const requiredFragments = [
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

const missing = requiredFragments.filter((fragment) => !workflow.includes(fragment));
if (missing.length > 0) {
  console.error("ci provenance workflow: FAILED");
  for (const fragment of missing) console.error(`- fragment absent: ${fragment}`);
  process.exit(1);
}

console.log(`ci provenance workflow: OK (${requiredFragments.length} invariants)`);
