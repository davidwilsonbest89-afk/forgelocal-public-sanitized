#!/usr/bin/env node

const fs = require("fs");

const publicEnglishDocs = [
  "README.md",
  "API.md",
  "docker/README.md",
  "docs/README.md",
  "docs/cli.md",
  "docs/local-quickstart.md",
  "docs/cloud-deployment.md",
  "docs/agent-integration.md",
  "docs/developer-integration.md",
  "docs/linux-server.md",
  "docs/platform-support.md",
  "docs/dual-browser-architecture.md",
  "docs/playwright-patches.md",
  "docs/release.md",
  "docs/i18n.md",
  "docs/documentation-language-audit.md",
];

const allowedCjkTerms = [
  "繁體中文",
];

const cjkPattern = /[\u3400-\u9FFF\uF900-\uFAFF]/u;
let failed = false;

for (const file of publicEnglishDocs) {
  if (!fs.existsSync(file)) {
    console.error(`Missing public documentation file: ${file}`);
    failed = true;
    continue;
  }

  const lines = fs.readFileSync(file, "utf8").split(/\r?\n/);
  lines.forEach((line, index) => {
    const normalizedLine = allowedCjkTerms.reduce(
      (current, term) => current.replaceAll(term, ""),
      line,
    );
    if (!cjkPattern.test(normalizedLine)) {
      return;
    }
    console.error(`${file}:${index + 1}: public English doc contains CJK text: ${line}`);
    failed = true;
  });
}

if (failed) {
  process.exit(1);
}

console.log(`Checked ${publicEnglishDocs.length} public English documentation files.`);
