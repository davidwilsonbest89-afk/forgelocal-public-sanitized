import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { pathToFileURL } from "node:url";

const SECURITY_REVIEWER = "boucheriechefimane-cmd";
const RELEASE_REVIEWER = "davidwilsonbest89-afk";
const INDEPENDENT_REVIEWER = "hajarbenmlih91-cloud";

const criticalPaths = [
  ".github/CODEOWNERS",
  ".github/workflows/",
  "scripts/check-component-rights.mjs",
  "scripts/test-component-rights.mjs",
  "scripts/create-component-rights-evidence.mjs",
  "docs/component-rights-register.json",
  "release/",
  "dist/"
];

function fail(message) {
  throw new Error(`release-separation: ${message}`);
}

function flattenPages(payload, label) {
  if (!Array.isArray(payload)) fail(`${label} doit être un tableau JSON.`);
  return payload.flatMap((page) => (Array.isArray(page) ? page : [page]));
}

function roleFor(login) {
  if (login === SECURITY_REVIEWER) return "security";
  if (login === RELEASE_REVIEWER) return "release";
  if (login === INDEPENDENT_REVIEWER) return "independent";
  return "other";
}

function isCriticalFile(filename) {
  return criticalPaths.some((path) => {
    if (path.endsWith("/")) return filename.startsWith(path);
    return filename === path;
  });
}

function latestReviewByUser(reviews) {
  const latest = new Map();
  for (const review of reviews) {
    const login = review?.user?.login;
    if (typeof login !== "string" || login.length === 0) continue;
    const timestamp = Date.parse(review.submitted_at ?? "") || 0;
    const id = Number.isInteger(review.id) ? review.id : 0;
    const previous = latest.get(login);
    if (!previous || timestamp > previous.timestamp || (timestamp === previous.timestamp && id > previous.id)) {
      latest.set(login, { state: review.state, timestamp, id });
    }
  }
  return latest;
}

function requirementsForAuthor(author) {
  switch (roleFor(author)) {
    case "security":
      return ["release", "independent"];
    case "release":
      return ["security", "independent"];
    case "independent":
      return ["security", "release"];
    default:
      return ["security", "release"];
  }
}

export function evaluateReleaseSeparation({ pullRequest, reviews, files }) {
  const author = pullRequest?.user?.login;
  if (typeof author !== "string" || author.length === 0) {
    fail("l’auteur de la pull request est absent.");
  }

  const normalizedFiles = flattenPages(files, "files");
  for (const file of normalizedFiles) {
    if (typeof file?.filename !== "string" || file.filename.length === 0) {
      fail("une entrée files ne contient pas filename.");
    }
  }

  const criticalFiles = normalizedFiles
    .map((file) => file.filename)
    .filter(isCriticalFile)
    .sort();
  const authorRole = roleFor(author);
  const requiredRoles = requirementsForAuthor(author);

  if (criticalFiles.length === 0) {
    return {
      schema_version: 1,
      status: "not_applicable",
      critical_change: false,
      author_role: authorRole,
      required_roles: requiredRoles,
      approved_roles: [],
      missing_roles: [],
      critical_file_count: 0
    };
  }

  const normalizedReviews = flattenPages(reviews, "reviews");
  const latest = latestReviewByUser(normalizedReviews);
  const approvedRoles = new Set();
  for (const [login, review] of latest.entries()) {
    if (login === author || review.state !== "APPROVED") continue;
    approvedRoles.add(roleFor(login));
  }

  const missingRoles = requiredRoles.filter((role) => !approvedRoles.has(role));
  return {
    schema_version: 1,
    status: missingRoles.length === 0 ? "passed" : "failed",
    critical_change: true,
    author_role: authorRole,
    required_roles: requiredRoles,
    approved_roles: [...approvedRoles].filter((role) => role !== "other").sort(),
    missing_roles: missingRoles,
    critical_file_count: criticalFiles.length
  };
}

function parseArguments(argv) {
  const values = new Map();
  for (let index = 0; index < argv.length; index += 2) {
    const flag = argv[index];
    const value = argv[index + 1];
    if (!flag?.startsWith("--") || !value || values.has(flag)) {
      fail("usage: --pr <fichier> --reviews <fichier> --files <fichier> [--output <fichier>]");
    }
    values.set(flag, value);
  }
  for (const required of ["--pr", "--reviews", "--files"]) {
    if (!values.has(required)) fail(`argument requis absent: ${required}`);
  }
  return values;
}

function readJSON(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch (error) {
    fail(`JSON invalide ou illisible (${file}): ${error.message}`);
  }
}

function main() {
  const args = parseArguments(process.argv.slice(2));
  const evidence = evaluateReleaseSeparation({
    pullRequest: readJSON(args.get("--pr")),
    reviews: readJSON(args.get("--reviews")),
    files: readJSON(args.get("--files"))
  });

  if (args.has("--output")) {
    const output = args.get("--output");
    mkdirSync(dirname(output), { recursive: true });
    writeFileSync(output, `${JSON.stringify(evidence, null, 2)}\n`, { mode: 0o600 });
  }

  console.log(`release-separation: ${evidence.status}`);
  if (evidence.status === "failed") {
    console.error(`release-separation: rôles requis absents: ${evidence.missing_roles.join(", ")}`);
    process.exitCode = 1;
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  main();
}
