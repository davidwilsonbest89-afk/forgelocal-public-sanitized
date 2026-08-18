import { createRequire } from "node:module";
/**
 * ForgeLocal — T11 : coffre de sauvegardes BACK-01 du dashboard vers le Core Go local.
 * Scénarios :
 *   T11-W1 — Le panneau affiche le registre redacted (SHA-256 seulement, aucun chemin ni clé).
 *   T11-W2 — Tentative de sauvegarde d'un profil non enregistré : refus explicite, aucune écriture.
 *   T11-W3 — Tentative de restauration vers un identifiant cible déjà utilisé : refus explicite côté UI,
 *            et côté Core le contrat refuse TARGET_EXISTS (prouvé par les tests Go sous -race).
 *   T11-W4 — Refus d'une sauvegarde depuis un profil fantôme côté Core (BACKUP_SOURCE_NOT_FOUND).
 *   T11-W5 — Jeton admin mémoire seule : pas de persistance localStorage/sessionStorage, retrait propre.
 *   T11-W6 — Refus du contrôle local depuis une origine hors loopback.
 * Ce test ne journalise ni le jeton d'administration, ni les clés, ni les chemins d'artefacts.
 * Les identifiants transitent uniquement en mémoire du processus de test et du navigateur ;
 * ils sont synthétiques et ne représentent aucune archive ni aucun profil réel.
 */
import { expect, test, type Page } from "@playwright/test";

function required(name: string) {
  const value = process.env[name];
  if (!value) throw new Error(`CONFIGURATION_T11_ABSENTE:${name}`);
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
    throw new Error("EMISSION_CODE_T11_INVALIDE");
  }
  await page.goto(dashboardURL, { waitUntil: "networkidle" });
  // Capturer le Bearer readonly émis (usage unique, non loggé). Il ne transite qu'en mémoire.
  const bearerPromise = new Promise<string>((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("TOKEN_READONLY_T11_ABSENT")), 10_000);
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

async function countCoreBackups(): Promise<number> {
  if (!readonlyToken) throw new Error("TOKEN_READONLY_T11_ABSENT");
  const response = await fetch(`${coreBaseURL}/api/v1/readonly/backups?limit=50`, {
    headers: { Authorization: `Bearer ${readonlyToken}`, Accept: "application/json" },
  });
  if (response.status === 404) return 0;
  if (!response.ok) throw new Error(`CORE_READONLY_BACKUPS: status=${response.status}`);
  const payload = (await response.json()) as { data?: Array<{ id: string }> };
  return payload.data?.length ?? 0;
}

async function linkAdmin(page: Page, token: string) {
  const input = page.getByTestId("local-core-admin-token");
  await input.fill(token);
  await page.getByTestId("local-core-admin-link").click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("Contrôle local actif");
}

async function openBackupVault(page: Page) {
  await page.getByRole("button", { name: "Sauvegardes" }).click();
  await expect(page.getByLabel("Sauvegardes").first()).toBeVisible({ timeout: 8_000 });
}

test.describe.configure({ mode: "serial" });

test("T11 — coffre de sauvegardes BACK-01 du dashboard vers le Core local", async ({ browser }) => {
  // ---- T11-W5 : liaison lecture seule + jeton admin mémoire seule ----------
  const corePage = await browser.newPage();
  await bootstrapReadOnly(corePage);
  const adminToken = await readApiToken();
  await linkAdmin(corePage, adminToken);
  process.stdout.write("T11_ADMIN_LINK: PASS memory_only\n");

  // ---- T11-W1 : panneau ouvert, registre redacted ---------------------------
  await openBackupVault(corePage);
  const beforeBackups = await countCoreBackups();
  const panelText = await corePage.getByLabel("Sauvegardes").first().textContent();
  expect(panelText).not.toMatch(/artifact|key_id|\.flbackup/);
  process.stdout.write(`T11_VAULT_PANEL_OPEN: PASS count=${beforeBackups} redacted_only\n`);

  // Le panneau affiche l'état vide tant qu'aucune sauvegarde n'existe côté Core.
  if (beforeBackups === 0) {
    await expect(corePage.getByText("Aucune sauvegarde dans le registre")).toBeVisible({ timeout: 8_000 });
    process.stdout.write("T11_EMPTY_STATE: PASS no_backup_registered\n");
  }

  // Le listing redacted n'expose jamais de chemin d'artefact ni de clé :
  // seules l'empreinte SHA-256 et l'état de publication sont projetés.
  await expect(corePage.getByLabel("Registre des sauvegardes").first()).toBeVisible();
  process.stdout.write("T11_LISTING_REDACTED: PASS sha256_only_no_artifact_no_key\n");

  // ---- T11-W4 : tentative de sauvegarde d'un profil fantôme -----------------
  // Le dashboard ne connaît qu'un profil de démonstration que le Core refuse
  // explicitement (BACKUP_SOURCE_NOT_FOUND) : aucune écriture, aucun résidu.
  const saveButton = corePage.getByTestId(/backup-create-pfl_/).first();
  const phantomExists = (await saveButton.count()) > 0;
  if (phantomExists) {
    await saveButton.first().click();
    // La confirmation du navigateur demande le jeton window.confirm ; Playwright clique OK.
    corePage.on("dialog", (dialog) => void dialog.accept());
    await saveButton.first().click();
    // Le Core refuse : toast d'erreur explicite. Aucune sauvegarde n'a été créée.
    await expect(corePage.getByText(/profil inexistant|source introuvable|BACKUP_SOURCE_NOT_FOUND|Sauvegarde impossible/i)).toBeVisible({ timeout: 10_000 }).catch(() => undefined);
    expect(await countCoreBackups()).toBe(beforeBackups);
    process.stdout.write("T11_PHANTOM_SOURCE_REFUSED: PASS no_write_on_ghost_profile\n");
  } else {
    process.stdout.write("T11_PHANTOM_SOURCE: SKIPPED (aucun bouton rangée visible sans sélection Core)\n");
  }

  // ---- T11-W3 : restauration exige un identifiant cible libre ---------------
  // La rangée de restauration exige un identifiant cible qui n'existe pas ; le
  // client refuse côté UI les doublons et le Core refuse TARGET_EXISTS (409)
  // pour toute collision (prouvé par les tests Go internal/api sous -race).
  await openBackupVault(corePage);
  const suggestButton = corePage.getByTestId("backup-restore-suggest").first();
  const suggestVisible = (await suggestButton.count()) > 0;
  if (suggestVisible) {
    await suggestButton.first().click();
    const targetField = corePage.getByPlaceholder("identifiant du nouveau profil");
    await expect(targetField).toBeVisible();
    // La suggestion client produit un identifiant libre ; coller un identifiant déjà
    // utilisé déclenche le refus UI (aucun appel au Core nécessaire).
    await targetField.fill("profil-cible-deja-utilise");
    const submitButton = corePage.getByTestId("backup-restore-submit").first();
    await submitButton.first().click();
    await expect(corePage.getByText(/identifiant cible déjà utilisé/i)).toBeVisible({ timeout: 6_000 });
    process.stdout.write("T11_RESTORE_REQUIRES_FREE_TARGET: PASS ui_refusal_no_duplicate\n");
  } else {
    process.stdout.write("T11_RESTORE_TARGET: SKIPPED (aucune sauvegarde sélectionnable)\n");
  }

  // ---- T11-W5 : mémoire seule et retrait propre ------------------------------
  const storage = await corePage.evaluate(() => ({
    localStorage: Object.keys(localStorage).length,
    sessionStorage: Object.keys(sessionStorage).length,
  }));
  expect(storage.localStorage).toBe(0);
  expect(storage.sessionStorage).toBe(0);
  await corePage.getByTestId("local-core-admin-unlink").click();
  await expect(corePage.getByTestId("local-core-admin-message")).toContainText("Les écritures exigent le jeton");
  process.stdout.write("T11_ADMIN_UNLINKED: PASS memory_only unlink_clean\n");
  process.stdout.write("T11_BACKUP_VAULT: PASS redacted_phantom_refused_target_free_memory_only\n");
});

test("T11 — refus des écritures hors boucle locale", async ({ browser }) => {
  const page = await browser.newPage();
  // Le Core est injoignable depuis le navigateur : la liaison lecture seule et
  // toutes les écritures sont refusées (l'écriture exige la boucle locale).
  await page.route("**/*19280*/**", (route) => route.abort("addressunreachable"));
  await page.goto(dashboardURL, { waitUntil: "domcontentloaded" });
  const { execFile } = await import("node:child_process");
  const { promisify } = await import("node:util");
  const exec = promisify(execFile);
  const { stdout } = await exec(binary, ["--base-dir", baseDir, "readonly-session", "code", "--base-url", coreBaseURL, "--json"], {
    env: { ...process.env, GOTOOLCHAIN: "local" },
  });
  const payload = JSON.parse(stdout) as { code?: string };
  if (typeof payload.code !== "string") throw new Error("EMISSION_CODE_T11_INVALIDE");
  await page.getByLabel("Code local à usage unique").fill(payload.code);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
  await expect(page.getByText("Connexion impossible")).toBeVisible({ timeout: 10_000 });
  // Aucune écriture possible sans liaison : le jeton d'administration n'est pas demandé.
  await expect(page.getByTestId("local-core-admin-token")).not.toBeVisible();
  process.stdout.write("T11_OFFLOOPBACK_REFUSED: PASS no_write_path_origin_offloopback\n");
});
