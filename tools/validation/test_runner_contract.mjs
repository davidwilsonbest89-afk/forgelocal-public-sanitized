import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = path.dirname(new URL(import.meta.url).pathname);
const runner = fs.readFileSync(path.join(root, 'integrated_smoke_runner.mjs'), 'utf8');
const config = fs.readFileSync(path.join(root, 'integrated_smoke_runner.env.example'), 'utf8');

assert.match(runner, /waitForCoreSessionForProfile/);
assert.match(runner, /externalForwardObserved|externalForwardAssertion/);
assert.match(runner, /SMOKE_ADMIN_TOKEN/);
assert.doesNotMatch(runner, /waitForTimeout\s*\(/);
for (const forbidden of ['synthetic-user', 'synthetic-pass', 'correct-user', 'correct-pass', 'wrong-user', 'wrong-pass']) {
  assert.doesNotMatch(runner, new RegExp(forbidden));
}
assert.match(config, /<RUNTIME_ONLY_REDACTED>/);
assert.doesNotMatch(config, /Bearer\s+[A-Za-z0-9._-]{12,}/i);
console.log('RUNNER_CONTRACT_PASS');
