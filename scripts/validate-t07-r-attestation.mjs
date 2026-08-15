#!/usr/bin/env node
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const EXPECTED_ARCHIVE_SHA256 = "dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2";
const ALLOWED_DECISIONS = new Set(["REAL_SECRET", "FALSE_POSITIVE", "UNKNOWN"]);
const PLACEHOLDER = /<required|^pending$|private_commit\|immutable_snapshot/i;

function get(object, path) {
  return path.split(".").reduce((value, key) => (value && typeof value === "object" ? value[key] : undefined), object);
}

function isRequiredValue(value) {
  return typeof value === "string" && value.trim().length > 0 && !PLACEHOLDER.test(value.trim());
}

function isSha256(value) {
  return typeof value === "string" && /^[a-f0-9]{64}$/i.test(value);
}

function report(errors, path, message) {
  errors.push({ path, message });
}

function requireValue(document, errors, path) {
  if (!isRequiredValue(get(document, path))) {
    report(errors, path, "missing_or_placeholder");
  }
}

function requireSha(document, errors, path) {
  if (!isSha256(get(document, path))) {
    report(errors, path, "missing_or_invalid_sha256");
  }
}

async function main() {
  const input = process.argv[2];
  if (!input) {
    console.error("usage: node scripts/validate-t07-r-attestation.mjs <redacted-attestation.json>");
    process.exit(2);
  }

  let document;
  try {
    document = JSON.parse(await readFile(resolve(input), "utf8"));
  } catch (error) {
    console.error(JSON.stringify({ status: "invalid", errors: [{ path: "$", message: "unreadable_or_invalid_json" }] }));
    process.exit(1);
  }

  const errors = [];
  const warnings = [];
  requireValue(document, errors, "source_revision.private_repository_or_attestation_id");
  requireValue(document, errors, "source_revision.revision_identifier");
  requireValue(document, errors, "source_revision.independent_review.reference");
  requireSha(document, errors, "source_revision.snapshot_sha256");
  for (const area of ["revision_or_snapshot", "candidate_archive_sha256", "rights_scope", "license_or_agreement", "notices", "security_triage"]) {
    const path = `source_revision.independent_review.coverage.${area}`;
    if (get(document, path) !== true) {
      report(errors, path, "independent_review_coverage_must_be_true");
    }
  }

  const revisionKind = get(document, "source_revision.revision_kind");
  if (!["private_commit", "immutable_snapshot"].includes(revisionKind)) {
    report(errors, "source_revision.revision_kind", "must_be_private_commit_or_immutable_snapshot");
  }

  if (get(document, "candidate.archive_sha256") !== EXPECTED_ARCHIVE_SHA256) {
    report(errors, "candidate.archive_sha256", "unexpected_candidate_archive_sha256");
  }
  if (get(document, "source_revision.attested_candidate_archive_sha256") !== EXPECTED_ARCHIVE_SHA256) {
    report(errors, "source_revision.attested_candidate_archive_sha256", "candidate_link_mismatch");
  }

  requireValue(document, errors, "rights_and_license.rights_holder_or_authorized_grantor");
  requireValue(document, errors, "rights_and_license.license_or_agreement_reference");
  requireValue(document, errors, "rights_and_license.notices_reference");
  requireValue(document, errors, "rights_and_license.rights_scope.third_party_obligations_reference");

  for (const path of ["rights_and_license.rights_scope.internal_use", "rights_and_license.rights_scope.modification"]) {
    if (get(document, path) !== "yes") {
      report(errors, path, "must_be_yes_for_review_eligibility");
    }
  }
  if (!["granted", "not_granted"].includes(get(document, "rights_and_license.rights_scope.redistribution"))) {
    report(errors, "rights_and_license.rights_scope.redistribution", "must_be_granted_or_not_granted");
  }

  if (get(document, "security_triage.secret_value_recorded_here") !== false) {
    report(errors, "security_triage.secret_value_recorded_here", "must_be_false");
  }
  const maintainerDecision = get(document, "security_triage.maintainer_decision");
  const independentDecision = get(document, "security_triage.independent_reviewer_decision");
  if (!ALLOWED_DECISIONS.has(maintainerDecision)) {
    report(errors, "security_triage.maintainer_decision", "invalid_or_missing_decision");
  }
  if (!ALLOWED_DECISIONS.has(independentDecision)) {
    report(errors, "security_triage.independent_reviewer_decision", "invalid_or_missing_decision");
  }
  const decisionsConcordant = ALLOWED_DECISIONS.has(maintainerDecision) && ALLOWED_DECISIONS.has(independentDecision) && maintainerDecision === independentDecision;
  const triageOperationalStatus = decisionsConcordant ? maintainerDecision : "UNKNOWN";
  if (!decisionsConcordant) {
    report(errors, "security_triage", "missing_invalid_or_divergent_decisions_force_unknown");
  }

  if (triageOperationalStatus === "REAL_SECRET") {
    requireValue(document, errors, "security_triage.revocation_or_rotation_reference");
    requireValue(document, errors, "security_triage.new_candidate_snapshot_reference");
    requireSha(document, errors, "security_triage.new_candidate_snapshot_sha256");
    requireValue(document, errors, "security_triage.new_candidate_snapshot_rescan_reference");
  }
  if (triageOperationalStatus === "FALSE_POSITIVE") {
    requireValue(document, errors, "security_triage.new_redacted_snapshot_reference");
    requireSha(document, errors, "security_triage.new_redacted_snapshot_sha256");
    requireValue(document, errors, "security_triage.new_snapshot_rescan_reference");
  }
  if (triageOperationalStatus === "UNKNOWN") {
    warnings.push("triage_unknown_keeps_t07_blocked");
  }

  const redistribution = get(document, "rights_and_license.rights_scope.redistribution");
  const futureDistribution = redistribution === "not_granted" ? "blocked" : redistribution === "granted" ? "requires_independent_rights_review" : "blocked";
  if (redistribution === "not_granted") {
    warnings.push("redistribution_not_granted_keeps_audit_passive_only");
  }

  if (get(document, "constraints.no_private_source_in_git") !== true || get(document, "constraints.no_credentials_in_git") !== true || get(document, "constraints.t08_authorized") !== false) {
    report(errors, "constraints", "required_t07_safety_constraints_not_preserved");
  }

  const output = {
    status: errors.length === 0 ? "complete_for_independent_review_only" : "incomplete_or_inconsistent",
    t08_authorized: false,
    triage_operational_status: triageOperationalStatus,
    future_distribution: futureDistribution,
    errors,
    warnings,
  };
  process.stdout.write(`${JSON.stringify(output)}\n`);
  process.exit(errors.length === 0 ? 0 : 1);
}

main();
