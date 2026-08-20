import { cpSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

const root = process.cwd();
const scenarios = [
  {
    name: "authorized",
    mutate: () => {},
    expectedStatus: 0,
  },
  {
    name: "denied-integrated",
    mutate: (registry) => {
      registry.components[0].rights_status = "denied";
    },
    expectedStatus: 1,
  },
  {
    name: "unknown-integrated",
    mutate: (registry) => {
      registry.components[0].rights_status = "unknown";
    },
    expectedStatus: 1,
  },
  {
    name: "missing-revision",
    mutate: (registry) => {
      registry.components[0].source_revision = null;
    },
    expectedStatus: 1,
  },
  {
    name: "not-first-party",
    mutate: (registry) => {
      registry.components[1].first_party = false;
    },
    expectedStatus: 1,
  },
  {
    name: "absent-owner",
    mutate: (registry) => {
      registry.build_inputs[0].owners = ["not-in-register"];
    },
    expectedStatus: 1,
  },
];

for (const scenario of scenarios) {
  const fixture = mkdtempSync(join(tmpdir(), "forgelocal-rights-"));
  try {
    cpSync(root, fixture, {
      recursive: true,
      filter: (path) => !path.includes("/.git") && !path.includes("/.manus-logs"),
    });
    const path = join(fixture, "docs", "component-rights-register.json");
    const registry = JSON.parse(readFileSync(path, "utf8"));
    scenario.mutate(registry);
    writeFileSync(path, `${JSON.stringify(registry, null, 2)}\n`);
    const result = spawnSync(process.execPath, ["scripts/check-component-rights.mjs"], {
      cwd: fixture,
      encoding: "utf8",
    });
    if (result.status !== scenario.expectedStatus) {
      throw new Error(`${scenario.name}: expected ${scenario.expectedStatus}, got ${result.status}\n${result.stdout}\n${result.stderr}`);
    }
    console.log(`component-rights test: ${scenario.name} OK`);
  } finally {
    rmSync(fixture, { recursive: true, force: true });
  }
}
