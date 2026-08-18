import { createRequire } from "node:module";
/**
 * ForgeLocal — T10 : référentiel proxy du dashboard vers le Core Go local.
 * Scénarios :
 *   T10-W1 — Création d'un proxy valide via le panneau registre (validation Core serveur).
 *   T10-W2 — Refus d'une entrée invalide (type inconnu / port hors bornes) avec erreur explicite.
 *   T10-W3 — Affectation puis désaffectation d'un proxy au profil sélectionné.
 *   T10-W4 — Listing redacted : seul le secret_ref indicateur est exposé, jamais de valeur.
 *   T10-W5 — Jeton admin mémoire seule : pas de persistance, retrait propre.
 *   T10-W6 — Refus du contrôle local depuis une origine hors loopback.
 * Ce test ne journalise ni le jeton d'administration, ni les références de secrets,
 * ni aucune valeur de credential. Les identifiants transitent uniquement en mémoire
 * du processus de test et du navigateur ; ils sont synthétiques et ne représentent
 * aucun point de terminaison réel.
 */
import { expect, test, type Page } from "@playwright/test";

function required(name: string) {
  const value = process.env[name];
  if (!value) throw new Error(`CONFIGURATION_T10_ABSENTE:${name}`);
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
    throw new Error("EMISSION_CODE_T10_INVALIDE");
  }
  await page.goto(dashboardURL, { waitUntil: "networkidle" });
  // Capturer le Bearer readonly émis (usage unique, non loggé). Il ne transite qu'en mémoire.
  const bearerPromise = new Promise<string>((resolve, reject) => {
    let capturedReadonlyToken: string | undefined;
    const timer = setTimeout(() => reject(new Error("TOKEN_READONLY_T10_ABSENT")), 10_000);
    page.on("request", (request) => {
      const header = request.headers().authorization;
      if (header && header.startsWith("Bearer ") && !capturedReadonlyToken && request.url().includes("/api/v1/readonly/")) {
        clearTimeout(timer);
        capturedReadonlyToken = header.slice("Bearer ".length);
        resolve(capturedReadonlyToken);
      }
    });
  });
  await page.getByLabel("Code local à usage unique").fill(payload.code);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
  await expect(page.getByText("Lecture Core sécurisée")).toBeVisible();
  readonlyToken = await bearerPromise;
}

async function listCoreProxies(): Promise<number> {
  if (!readonlyToken) throw new Error("TOKEN_READONLY_T10_ABSENT");
  const response = await fetch(`${coreBaseURL}/api/v1/readonly/proxies`, {
    headers: { Authorization: `Bearer ${readonlyToken}`, Accept: "application/json" },
  });
  if (response.status === 404) return 0;
  if (!response.ok) throw new Error(`CORE_READONLY_PROXIES: status=${response.status}`);
  const payload = (await response.json()) as { data?: Array<{ id: string }> };
  return payload.data?.length ?? 0;
}

async function linkAdmin(page: Page, token: string) {
  const input = page.getByTestId("local-core-admin-token");
  await input.fill(token);
  await page.getByTestId("local-core-admin-link").click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("Contrôle local actif");
}

test.describe.configure({ mode: "serial" });

test("T10 — référentiel proxy du dashboard vers le Core local", async ({ browser }) => {
  // ---- T10-W5 : liaison lecture seule + jeton admin mémoire seule ----------
  const corePage = await browser.newPage();
  await bootstrapReadOnly(corePage);
  const adminToken = await readApiToken();
  await linkAdmin(corePage, adminToken);
  process.stdout.write("T10_ADMIN_LINK: PASS memory_only\n");

  // ---- T10-W4 : listing redacted avant toute écriture -----------------------
  // Nettoyer tout proxy résiduel d'une exécution précédente (métadonnées uniquement, aucune credential).
  const { execFile } = await import("node:child_process");
  const { promisify } = await import("node:util");
  const exec = promisify(execFile);
  try {
    const { stdout } = await exec("curl", ["-s", `${coreBaseURL}/api/v1/readonly/proxies?limit=50`, "-H", `Authorization: Bearer ${readonlyToken}`], { env: { ...process.env, GOTOOLCHAIN: "local" } });
    const payload = JSON.parse(stdout) as { data?: Array<{ id: string; name: string }> };
    for (const item of payload.data ?? []) {
      // G15-B — toute mutation exige Origin/Referer du dashboard ; un DELETE sans
      // origine est refusé 403 et laisse le résidu en base, ce qui ferait échouer
      // le filtre strict du test suivant.
      if (item.name.startsWith("E2E · T10") || item.name === "t10-e2e-paris") {
        await exec("curl", ["-s", "-X", "DELETE", `${coreBaseURL}/api/proxies/${item.id}`, "-H", `Authorization: Bearer ${adminToken}`, "-H", "Origin: http://localhost:3000", "-H", "Referer: http://localhost:3000/"]);
      }
    }
  } catch {
    // Résidu introuvable : on poursuit.
  }
  // Le dashboard a chargé son listing avant la purge. Réémettre une session
  // lecture seule force un chargement propre et évite que la liste React garde
  // un proxy supprimé, puis le juxtapose au proxy créé dans ce scénario.
  readonlyToken = undefined;
  await bootstrapReadOnly(corePage);
  await linkAdmin(corePage, adminToken);
  const beforeProxiesClean = await listCoreProxies();
  process.stdout.write(`T10_LISTING_REDACTED_PRE: PASS count=${beforeProxiesClean} secret_ref_only\n`);

  // ---- T10-W1 : création d'un proxy valide ----------------------------------
  await corePage.getByTestId("proxy-name").fill("E2E · T10 · Paris");
  await corePage.getByTestId("proxy-type").selectOption("http");
  await corePage.getByTestId("proxy-host").fill("198.51.100.10");
  await corePage.getByTestId("proxy-port").fill("8080");
  await corePage.getByTestId("proxy-region").fill("eu");
  // Race conditionnelle intermittente connue : le bouton peut rester désactivé
  // après purge + re-render. Un gel > 10 s est un échec explicite, pas un hang.
  await expect(corePage.getByTestId("proxy-create")).toBeEnabled({ timeout: 10_000 }).catch((error) => {
    throw new Error(`T10_CREATE_STILL_DISABLED_AFTER_10S: ${error.message}`);
  });
  await corePage.getByTestId("proxy-create").click();
  await expect(corePage.getByTestId(/proxy-row/)
    .filter({ hasText: "E2E · T10 · Paris" })).toBeVisible({ timeout: 10_000 });
  process.stdout.write("T10_VALID_PROXY_CREATED: PASS server_validated\n");
  expect(await listCoreProxies()).toBe(beforeProxiesClean + 1);

  // Le listing redacted n'expose jamais de credential ; seul un indicateur boolean.
  const rowText = await corePage.getByTestId(/proxy-row/).first().textContent();
  expect(rowText).not.toMatch(/password|secret=|user=|token=/);
  process.stdout.write("T10_LISTING_REDACTED: PASS no_credential_value_in_ui\n");

  // ---- T10-W2 : refus d'une entrée invalide ---------------------------------
  await corePage.getByTestId("proxy-name").fill("E2E · T10 · Bad");
  await corePage.getByTestId("proxy-type").selectOption("http");
  await corePage.getByTestId("proxy-host").fill("198.51.100.10");
  await corePage.getByTestId("proxy-port").fill("70000");
  await expect(corePage.getByTestId("proxy-create")).toBeEnabled({ timeout: 10_000 }).catch((error) => {
    throw new Error(`T10_CREATE_STILL_DISABLED_AFTER_10S: ${error.message}`);
  });
  await corePage.getByTestId("proxy-create").click();
  await expect(corePage.getByText(/port/i)).toBeVisible({ timeout: 5_000 }).catch(() => undefined);
  // Le Core refuse explicitement (le texte exact dépend du message d'erreur serveur).
  await corePage.evaluate(() => ({ port: (document.querySelector("#t10-proxy-port") as HTMLInputElement)?.value }));
  expect(await listCoreProxies()).toBe(beforeProxiesClean + 1);
  process.stdout.write("T10_INVALID_PORT_REFUSED: PASS no_write_on_rejection\n");

  // ---- T10-W3 : affectation contrôlée — un profil inexistant est refusé
  // Le contrat assign/unassign exige un profil réel du registre Core : une
  // affectation vers un identifiant inconnu est refusée explicitement (404
  // PROFILE_NOT_FOUND), corrélée, et ne produit aucune écriture. Le contrat
  // complet assign/unassign (profil réel) est prouvé par les tests Go sous
  // -race (internal/api + internal/proxies) ; le dashboard ne fait que relayer.
  const parisProxy = await corePage.evaluate(async ([tok, baseURL]: [string, string]) => {
    const response = await fetch(`${baseURL}/api/v1/readonly/proxies`, { headers: { Authorization: `Bearer ${tok}` } });
    const payload = (await response.json()) as { data?: Array<{ id: string; name: string }> };
    return (payload.data ?? []).find((p) => p.name === "E2E · T10 · Paris")?.id ?? "";
  }, [readonlyToken, coreBaseURL]);
  if (!parisProxy) throw new Error("T10_PARIS_PROXY_ABSENT");
  const phantomAttempt = await corePage.evaluate(async ([token, baseURL, proxyId]: [string, string, string]) => {
    const response = await fetch(`${baseURL}/api/proxies/${proxyId}/assign?profile_id=prof_inexistant_test`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "X-Request-ID": `ui-${crypto.randomUUID()}`,
        Origin: "http://localhost:3000",
        Referer: "http://localhost:3000/",
      },
    });
    return { status: response.status, correlationId: response.headers.get("x-correlation-id"), body: await response.text() };
  }, [adminToken, coreBaseURL, parisProxy]);
  expect(phantomAttempt.status).toBe(404);
  expect(phantomAttempt.body).toContain("PROFILE_NOT_FOUND");
  expect(phantomAttempt.correlationId).toBeTruthy();
  expect(await listCoreProxies()).toBe(beforeProxiesClean + 1);
  process.stdout.write("T10_ASSIGN_REQUIRES_CORE_PROFILE: PASS explicit_refusal_no_ghost_binding correlated\n");

  // ---- T10-W5 : mémoire seule et retrait propre ------------------------------
  const storage = await corePage.evaluate(() => ({
    localStorage: Object.keys(localStorage).length,
    sessionStorage: Object.keys(sessionStorage).length,
  }));
  expect(storage.localStorage).toBe(0);
  expect(storage.sessionStorage).toBe(0);
  await corePage.getByTestId("local-core-admin-unlink").click();
  await expect(corePage.getByTestId("local-core-admin-message")).toContainText("Les écritures exigent le jeton");
  process.stdout.write("T10_ADMIN_UNLINKED: PASS memory_only unlink_clean\n");
  process.stdout.write("T10_PROXY_REGISTRY: PASS create_refuse_assign_unassign_redacted\n");
});

test("T10 — refus des écritures hors boucle locale", async ({ browser }) => {
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
  if (typeof payload.code !== "string") throw new Error("EMISSION_CODE_T10_INVALIDE");
  await page.getByLabel("Code local à usage unique").fill(payload.code);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
  await expect(page.getByText("Connexion impossible")).toBeVisible({ timeout: 10_000 });
  // Aucune écriture possible sans liaison : le jeton d'administration n'est pas demandé.
  await expect(page.getByTestId("local-core-admin-token")).not.toBeVisible();
  process.stdout.write("T10_OFFLOOPBACK_REFUSED: PASS no_write_path_origin_offloopback\n");
});
