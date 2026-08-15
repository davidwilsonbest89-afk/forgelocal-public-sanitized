#!/usr/bin/env node
// T07-R validator regression tests. Every document here is synthetic test data,
// never an external attestation and never an input to the evidence collector.
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

const root = new URL("..", import.meta.url).pathname;
const validator = join(root, "scripts", "validate-t07-r-attestation.mjs");
const expectedArchive = "dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2";
const sha = (character) => character.repeat(64);

function baseAttestation() {
  return {
    candidate: { archive_sha256: expectedArchive },
    source_revision: {
      private_repository_or_attestation_id: "synthetic-test-attestation",
      revision_kind: "immutable_snapshot",
      revision_identifier: "synthetic-snapshot-r1",
      snapshot_sha256: sha("a"),
      attested_candidate_archive_sha256: expectedArchive,
      independent_review: {
        reference: "synthetic-review-reference",
        coverage: {
          revision_or_snapshot: true,
          candidate_archive_sha256: true,
          rights_scope: true,
          license_or_agreement: true,
          notices: true,
          security_triage: true,
        },
      },
    },
    rights_and_license: {
      rights_holder_or_authorized_grantor: "synthetic-rights-holder",
      rights_scope: {
        internal_use: "yes",
        modification: "yes",
        redistribution: "not_granted",
        third_party_obligations_reference: "synthetic-obligations-reference",
      },
      license_or_agreement_reference: "synthetic-license-reference",
      notices_reference: "synthetic-notices-reference",
    },
    security_triage: {
      secret_value_recorded_here: false,
      maintainer_decision: "FALSE_POSITIVE",
      independent_reviewer_decision: "FALSE_POSITIVE",
      new_redacted_snapshot_reference: "synthetic-redacted-snapshot",
      new_redacted_snapshot_sha256: sha("b"),
      new_snapshot_rescan_reference: "synthetic-rescan-reference",
    },
    constraints: {
      no_private_source_in_git: true,
      no_credentials_in_git: true,
      t08_authorized: false,
    },
  };
}

async function runCase(directory, name, document, expectedExit, expectedFragment) {
  const file = join(directory, `${name}.json`);
  await writeFile(file, `${JSON.stringify(document)}\n`, { mode: 0o600 });
  const run = spawnSync(process.execPath, [validator, file], { encoding: "utf8" });
  if (run.status !== expectedExit || !run.stdout.includes(expectedFragment)) {
    throw new Error(`${name}: expected exit ${expectedExit} and fragment ${expectedFragment}; got ${run.status}: ${run.stdout}${run.stderr}`);
  }
}

const directory = await mkdtemp(join(tmpdir(), "forgelocal-t07r-validator-"));
try {
  const accepted = baseAttestation();
  await runCase(directory, "accepted-passive-audit", accepted, 0, '"future_distribution":"blocked"');

  const deniedModification = baseAttestation();
  deniedModification.rights_and_license.rights_scope.modification = "no";
  await runCase(directory, "modification-no", deniedModification, 1, "must_be_yes_for_review_eligibility");

  const incompleteReviewScope = baseAttestation();
  incompleteReviewScope.source_revision.independent_review.coverage.notices = false;
  await runCase(directory, "independent-review-incomplete", incompleteReviewScope, 1, "independent_review_coverage_must_be_true");

  const divergentTriage = baseAttestation();
  divergentTriage.security_triage.independent_reviewer_decision = "REAL_SECRET";
  await runCase(directory, "divergent-triage", divergentTriage, 1, '"triage_operational_status":"UNKNOWN"');

  const incompleteSecretRemediation = baseAttestation();
  incompleteSecretRemediation.security_triage.maintainer_decision = "REAL_SECRET";
  incompleteSecretRemediation.security_triage.independent_reviewer_decision = "REAL_SECRET";
  delete incompleteSecretRemediation.security_triage.new_redacted_snapshot_reference;
  delete incompleteSecretRemediation.security_triage.new_redacted_snapshot_sha256;
  delete incompleteSecretRemediation.security_triage.new_snapshot_rescan_reference;
  incompleteSecretRemediation.security_triage.revocation_or_rotation_reference = "synthetic-rotation-reference";
  await runCase(directory, "real-secret-needs-new-snapshot", incompleteSecretRemediation, 1, "new_candidate_snapshot_reference");

  console.log("t07r_validator_regressions=PASS");
} finally {
  await rm(directory, { recursive: true, force: true });
}
