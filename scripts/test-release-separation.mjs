import assert from "node:assert/strict";
import { evaluateReleaseSeparation } from "./check-release-separation.mjs";

const criticalFiles = [[{ filename: ".github/workflows/ci.yml" }]];
const approved = (login, id = 1) => ({
  id,
  state: "APPROVED",
  submitted_at: `2026-08-15T00:00:0${id}Z`,
  user: { login }
});

function evaluate(author, reviews, files = criticalFiles) {
  return evaluateReleaseSeparation({
    pullRequest: { user: { login: author } },
    reviews,
    files
  });
}

assert.equal(
  evaluate("external-author", [[approved("boucheriechefimane-cmd"), approved("davidwilsonbest89-cloud", 2)]]).status,
  "failed",
  "un reviewer non désigné ne remplace pas Release"
);
assert.deepEqual(
  evaluate("external-author", [[approved("boucheriechefimane-cmd"), approved("davidwilsonbest89-afk", 2)]]).required_roles,
  ["security", "release"]
);
assert.equal(
  evaluate("external-author", [[approved("boucheriechefimane-cmd"), approved("davidwilsonbest89-afk", 2)]]).status,
  "passed"
);
assert.equal(
  evaluate("boucheriechefimane-cmd", [[approved("davidwilsonbest89-afk"), approved("hajarbenmlih91-cloud", 2)]]).status,
  "passed"
);
assert.equal(
  evaluate("davidwilsonbest89-afk", [[approved("boucheriechefimane-cmd"), approved("hajarbenmlih91-cloud", 2)]]).status,
  "passed"
);
assert.equal(
  evaluate("hajarbenmlih91-cloud", [[approved("boucheriechefimane-cmd"), approved("davidwilsonbest89-afk", 2)]]).status,
  "passed"
);
assert.equal(
  evaluate(
    "external-author",
    [[approved("boucheriechefimane-cmd"), { ...approved("davidwilsonbest89-afk", 2), state: "CHANGES_REQUESTED", id: 3 }]]
  ).status,
  "failed",
  "une approbation antérieure est invalidée par une revue plus récente"
);
assert.equal(
  evaluate("external-author", [[]], [[{ filename: "docs/README.md" }]]).status,
  "not_applicable"
);

console.log("release-separation test: 7 scénarios OK");
