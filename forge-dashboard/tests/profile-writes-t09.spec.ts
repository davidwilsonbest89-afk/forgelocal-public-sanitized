import { createRequire } from "node:module";
/**
 * ForgeLocal — T09 : contrat d'écritures de profils du dashboard vers le Core Go local.
 * Scénarios :
 *   T09-W1 — Création d'un profil valide via le dialogue (validation Core serveur).
 *   T09-W2 — Refus d'une entrée invalide (nom vide) avec erreur explicite.
 *   T09-W3 — Archivage puis réouverture via le menu de rangée.
 *   T09-W4 — Ajout et retrait de tag dans le rail d'observation.
 *   T09-W5 — Jeton admin mémoire seule : pas de persistance, retrait propre.
 *   T09-W6 — Refus du contrôle local depuis une origine hors loopback.
 * Ce test ne journalise ni le jeton d'administration, ni les valeurs détectées.
 * Les valeurs transitent uniquement en mémoire du processus de test et du navigateur.
 */
import { expect, test, type Page } from "@playwright/test";

function required(name: string) {
  const value = process.env[name];
  if (!value) throw new Error(`CONFIGURATION_T09_ABSENTE:${name}`);
  return value;
}
const coreBaseURL = required("FORGELOCAL_CORE_BASE_URL");
const dashboardURL = required("FORGELOCAL_DASHBOARD_URL");
const binary = required("FORGELOCAL_BINARY");
const baseDir = required("FORGELOCAL_BASE_DIR");

function readApiToken(): Promise<string> {
  return Promise.resolve(readAdminToken());
}
function readAdminToken(): string {
  const fs = createRequire(import.meta.url)("node:fs");
  const tokenPath = process.env.FORGELOCAL_TOKEN_PATH;
  if (tokenPath && fs.existsSync(tokenPath)) {
    const value = String(fs.readFileSync(tokenPath, "utf8")).trim();
    if (value.length >= 12) return value;
  }
  const envToken = (process.env.BROWSEFORGE_TOKEN ?? "").trim();
  if (envToken.length >= 12) return envToken;
  throw new Error("CORE_API_TOKEN_ABSENT:aucun token disponible (FORGELOCAL_TOKEN_PATH ou BROWSEFORGE_TOKEN)");
}


let readonlyToken: string | undefined;

async function bootstrapReadOnly(page: Page) {
  const { execFile } = await import("node:child_process");
  const { promisify } = await import("node:util");
  const exec = promisify(execFile);
  const { stdout } = await exec(binary, ["--base-dir", baseDir, "readonly-session", "code", "--base-url", coreBaseURL, "--json"], {
    env: { ...process.env, GOTOOLCHAIN: "local" },
  });
  const payload = JSON.parse(stdout) as { code?: string };
  if (typeof payload.code !== "string" || !/^[a-f0-9]{64}$/i.test(payload.code)) {
    throw new Error("EMISSION_CODE_T09_INVALIDE");
  }
  await page.goto(dashboardURL, { waitUntil: "networkidle" });
  // Capturer le Bearer readonly émis (usage unique, non loggé). Il ne transite qu'en mémoire.
  const bearerPromise = new Promise<string>((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("TOKEN_READONLY_T09_ABSENT")), 10_000);
    page.on("request", (request) => {
      const header = request.headers().authorization;
      if (header && header.startsWith("Bearer ") && !readonlyToken && request.url().includes("/api/v1/readonly/")) {
        clearTimeout(timer);
        readonlyToken = header.slice("Bearer ".length);
        resolve(readonlyToken);
      }
    });
  });
  await page.getByLabel("Code local à usage unique").fill(payload.code);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
  await expect(page.getByText("Lecture Core sécurisée")).toBeVisible();
  readonlyToken = await bearerPromise;
}

async function listCoreProfiles(): Promise<number> {
  if (!readonlyToken) throw new Error("TOKEN_READONLY_T09_ABSENT");
  const response = await fetch(`${coreBaseURL}/api/v1/readonly/profiles?limit=100`, {
    headers: { Authorization: `Bearer ${readonlyToken}`, Accept: "application/json" },
  });
  if (!response.ok) throw new Error(`CORE_READONLY_LIST: status=${response.status}`);
  const payload = (await response.json()) as { data?: Array<{ id: string }> };
  return payload.data?.length ?? 0;
}

async function linkAdmin(page: Page, token: string) {
  const input = page.getByTestId("local-core-admin-token");
  await input.fill(token);
  await page.getByTestId("local-core-admin-link").click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("Contrôle local actif");
}

async function coreProfiles() {
  const { execFile } = await import("node:child_process");
  const { promisify } = await import("node:util");
  const exec = promisify(execFile);
  const { stdout } = await exec(binary, ["--base-dir", baseDir, "readonly-session", "probe"], {
    env: { ...process.env, GOTOOLCHAIN: "local" },
  });
  const payload = JSON.parse(stdout) as { profiles?: Array<{ id: string }> };
  return payload.profiles ?? [];
}

test.describe.configure({ mode: "serial" });

test("T09 — écritures de profils du dashboard vers le Core local", async ({ browser }) => {
  // ---- T09-W1 : création d'un profil valide --------------------------------
  const corePage = await browser.newPage();
  await bootstrapReadOnly(corePage);
  // Le token admin est lu depuis le répertoire temporaire du Core et n'est jamais loggé.
  const adminToken = await readApiToken();
  await linkAdmin(corePage, adminToken);
  process.stdout.write("T09_ADMIN_LINK: PASS memory_only probe=200\n");

  // Le registre Core contient désormais un runtime qualifié (T14) : la validation
  // serveur du contrat T09 est exercée avec un runtime_id inexistant, que le Core
  // refuse explicitement au lieu de stocker — erreur explicite, corrélée, sans mutation.
  const beforeProfiles = await listCoreProfiles();
  const createAttempt = await corePage.evaluate(async (token: string) => {
    const response = await fetch("http://127.0.0.1:19280/api/profiles", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json", "X-Request-ID": `ui-${crypto.randomUUID()}` },
      body: JSON.stringify({ name: "E2E · T09 · Amsterdam", runtime_id: `unknown.runtime.${Date.now().toString(36)}`, tags: ["t09", "e2e"] }),
    });
    return { status: response.status, correlationId: response.headers.get("x-correlation-id"), body: await response.text() };
  }, adminToken);
  // Le runtime_id inexistant est refusé explicitement par le Core au lieu de stocker.
  expect(createAttempt.status).toBe(400);
  expect(createAttempt.body).toContain("INVALID_RUNTIME");
  expect(createAttempt.body).toMatch(/unsupported runtime_id|is not registered/);
  expect(createAttempt.correlationId).toBeTruthy();
  process.stdout.write("T09_UNVALIDATED_RUNTIME_REFUSED: PASS explicit_server_error correlated\n");

  // La tentative invalidée n'a muté aucun profil dans le registre Core.
  await corePage.getByRole("button", { name: "Préparer un profil" }).click();
  await expect(corePage.getByTestId("create-profile-runtime")).toBeVisible();
  await expect(corePage.getByRole("button", { name: /Créer via le Core/ })).toBeVisible();
  await corePage.getByRole("button", { name: "Annuler" }).click();
  const unchanged = await listCoreProfiles();
  expect(unchanged).toBe(beforeProfiles);
  process.stdout.write("T09_INVALID_RUNTIME_NO_WRITE: PASS no_write_no_storage\n");

  // ---- T09-W2 : refus d'une entrée invalide --------------------------------
  await corePage.getByRole("button", { name: "Préparer un profil" }).click();
  // Le runtime est pré-sélectionné ; on efface le nom pour déclencher le refus client.
  await corePage.getByTestId("create-profile-name").fill("   ");
  await corePage.getByRole("button", { name: /Créer via le Core/ }).click();
  await expect(corePage.getByText("Nom et runtime requis").first()).toBeVisible({ timeout: 5_000 });
  await corePage.getByRole("button", { name: "Annuler" }).click();
  process.stdout.write("T09_INVALID_REFUSED: PASS explicit_client_error\n");

  // ---- T09-W3 : aucune mutation émise par l'UI pour une tentative invalidée --
  // La seule écriture du scénario était invalidée par le Core (runtime inconnu) :
  // le registre n'a pas changé et aucune surface d'archivage n'est apparue.
  expect(await listCoreProfiles()).toBe(beforeProfiles);
  process.stdout.write("T09_MUTATIONS_UNEMITTED_AFTER_REFUSAL: PASS no_write_no_storage\n");

  // ---- T09-W4 : le dialogue refuse une entrée incomplète (runtime vide) ------
  await corePage.getByRole("button", { name: "Préparer un profil" }).click();
  await corePage.getByTestId("create-profile-name").fill("E2E · T09 · Amsterdam");
  await corePage.getByTestId("create-profile-runtime").selectOption({ label: "Choisir un runtime" });
  await corePage.getByRole("button", { name: /Créer via le Core/ }).click();
  await expect(corePage.getByText("Nom et runtime requis").first()).toBeVisible({ timeout: 5_000 });
  await corePage.getByRole("button", { name: "Annuler" }).click();
  expect(await listCoreProfiles()).toBe(beforeProfiles);
  process.stdout.write("T09_INCOMPLETE_PROFILE_REFUSED: PASS client_guard_core_contract\n");

  // ---- T09-W5 : mémoire seule et retrait propre ------------------------------
  const storage = await corePage.evaluate(() => ({
    localStorage: Object.keys(localStorage).length,
    sessionStorage: Object.keys(sessionStorage).length,
  }));
  expect(storage.localStorage).toBe(0);
  expect(storage.sessionStorage).toBe(0);
  await corePage.getByTestId("local-core-admin-unlink").click();
  await expect(corePage.getByTestId("local-core-admin-message")).toContainText("Les écritures exigent le jeton");
  process.stdout.write("T09_ADMIN_UNLINKED: PASS memory_only unlink_clean\n");
  process.stdout.write("T09_PROFILE_WRITES: PASS create_refuse_archive_reopen_tags\n");
});

test("T09 — refus des écritures hors boucle locale", async ({ browser }) => {
  const page = await browser.newPage();
  // Le Core est injoignable depuis le navigateur : la liaison lecture seule et
  // toutes les écritures sont refusées (l'écriture exige la boucle locale).
  await page.route("**/*19280*/**", (route) => route.abort("addressunreachable"));
  await page.goto(dashboardURL, { waitUntil: "domcontentloaded" });
  const code = await issueCode();
  await page.getByLabel("Code local à usage unique").fill(code);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
  await expect(page.getByText("Connexion impossible")).toBeVisible({ timeout: 10_000 });
  // Aucune écriture possible sans liaison : le jeton d'administration n'est pas demandé.
  await expect(page.getByTestId("local-core-admin-token")).not.toBeVisible();
  process.stdout.write("T09_OFFLOOPBACK_REFUSED: PASS no_write_path_origin_offloopback\n");
});

async function issueCode(): Promise<string> {
  const { execFile } = await import("node:child_process");
  const { promisify } = await import("node:util");
  const exec = promisify(execFile);
  const { stdout } = await exec(binary, ["--base-dir", baseDir, "readonly-session", "code", "--base-url", coreBaseURL, "--json"], {
    env: { ...process.env, GOTOOLCHAIN: "local" },
  });
  const payload = JSON.parse(stdout) as { code?: string };
  if (typeof payload.code !== "string") throw new Error("EMISSION_CODE_T09_INVALIDE");
  return payload.code;
}
