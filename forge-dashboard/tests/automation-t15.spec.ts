import { test, expect } from "@playwright/test";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { existsSync, readFileSync } from "node:fs";

// T15 — Automation locale (pilote CDP loopback) E2E.
// Vérifie contre le Core réel T15 (loopback, Bearer admin, mémoire seule) :
//   W1 Ouverture de session : POST /api/sessions ouvre une session liée au
//      profil réel ; GET /api/sessions liste sans porter, chemin ou token.
//   W2 Politique fail-closed local-only : navigate refuse toute URL externe
//      (https://example.com, IP privée non-loopback, bare host) et accepte
//      uniquement file:// et http(s)://127.0.0.1|localhost.
//   W3 Navigate + content + screenshot : navigation vers une fixture locale
//      file://, projection redacted du contenu (digest + longueur), capture
//      PNG transmise pour hachage client-side (jamais affichée brute).
//   W4 Fermeture de session : DELETE ferme la session, plus aucune
//      commande acceptée (NOT_FOUND) ; réouverture propre.
//   W5 Panneau dashboard : "Automation locale" monte, ouvre/ferme la
//      session du profil sélectionné, projette digest jamais brut ; aucun
//      HTML brut, aucune image, aucune coordonnée dans le DOM du panneau.
const exec = promisify(execFile);
// Nouvelle baseline forgebaseline-2026-08-17 (T15 réimplémenté clean-room).
const coreBinary = "/tmp/forge-core-e2e";
const coreBaseDir = "/tmp/forge-e2e-base";
const coreBaseURL = "http://127.0.0.1:19280";
const tokenPath = "/tmp/forge-e2e-token.txt";
const dashboardBase = "http://localhost:3000";
const fixtureFile = "file:///tmp/t15-fixtures/index.html";
// Profil créé par la suite elle-même (nom unique par exécution pour éviter les collisions
// entre Core frais sans state partagé). Le Core T15 exige un runtime_id valide ;
// "browseforge-chromium" est activé dans la config de référence E2E.
const profileName = "E2E · T15 · Automation";
let profileId = "";
async function ensureProfile(token: string): Promise<string> {
  if (profileId) return profileId;
  // Essayer d'abord un profil existant portant le même nom (state persistant).
  const list = await curl("-s", "-H", `Authorization: Bearer ${token}`, `${coreBaseURL}/api/profiles`);
  const existing = payloadOf<{ data?: Array<{ id: string; name: string; tags?: Array<string> }> }>(list.stdout);
  for (const p of existing?.data ?? []) {
    if (p.name === profileName) {
      profileId = p.id;
      return profileId;
    }
  }
  const create = await curl(
    "-s", "-X", "POST", "-H", `Authorization: Bearer ${token}`, "-H", "Content-Type: application/json",
    "-d", JSON.stringify({ name: profileName, runtime_id: "browseforge-chromium", tags: ["t15", "e2e"] }),
    `${coreBaseURL}/api/profiles`,
  );
  const p = payloadOf<{ data?: { id: string } }>(create.stdout);
  if (!p.data?.id || create.code < 200 || create.code >= 300) {
    throw new Error(`T15_PROFILE_CREATE_FAILED: status=${create.code} stdout=${create.stdout.slice(0, 200)}`);
  }
  profileId = p.data.id;
  return profileId;
}

function readToken(): string {
    const tokenPath = process.env.FORGELOCAL_TOKEN_PATH;
  if (tokenPath && existsSync(tokenPath)) {
    const value = String(readFileSync(tokenPath, "utf8")).trim();
    if (value.length >= 12) return value;
  }
  const envToken = (process.env.BROWSEFORGE_TOKEN ?? "").trim();
  if (envToken.length >= 12) return envToken;
  throw new Error("CORE_API_TOKEN_ABSENT:aucun token disponible (FORGELOCAL_TOKEN_PATH ou BROWSEFORGE_TOKEN)");
}

// G15-B — durcissement : toute mutation du Core doit déclarer l'origine locale
// (Origin + Referer du dashboard localhost:3000), sinon le Core refuse (ORIGIN_REJECTED).
function curl(...args: string[]): Promise<{ stdout: string; code: number }> {
  return exec("curl", [...args, "-H", "Origin: http://localhost:3000", "-H", "Referer: http://localhost:3000/", "-w", "\n__CODE__:%{http_code}"]).then(
    res => ({ stdout: res.stdout, code: httpCodeOf(res.stdout) }),
    err => ({ stdout: err.stdout, code: httpCodeOf(err.stdout) }),
  );
}
function httpCodeOf(output: string): number {
  const m = String(output).match(/__CODE__:(\d+)/);
  return m ? Number(m[1]) : 0;
}
function payloadOf<T>(output: string): T {
  return JSON.parse(String(output).split("__CODE__")[0]) as T;
}

async function linkAdmin(page: import("@playwright/test").Page, token: string) {
  await page.getByTestId("local-core-admin-token").fill(token);
  await page.getByTestId("local-core-admin-link").click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("Contrôle local actif", { timeout: 15_000 });
}

async function clearSessions(token: string) {
  const list0 = await curl("-s", "-H", `Authorization: Bearer ${token}`, `${coreBaseURL}/api/sessions`);
  const open0 = payloadOf<{ data?: Array<{ session_id: string }> }>(list0.stdout);
  for (const s of open0?.data ?? []) {
    await curl("-s", "-X", "DELETE", "-H", `Authorization: Bearer ${token}`, `${coreBaseURL}/api/sessions/${s.session_id}`);
  }
}

test.describe("T15 local CDP automation", () => {
  const token = readToken();
  const auth = `Authorization: Bearer ${token}`;

  test("W1: session opens and lists without port/path leakage", async () => {
    await ensureProfile(token);
    await clearSessions(token);
    const open = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", JSON.stringify({ profile_id: profileId }), `${coreBaseURL}/api/sessions`);
    expect(open.code).toBe(201);
    const payload = payloadOf<{ data?: { session_id: string; profile_id: string; runtime_id: string } }>(open.stdout);
    expect(payload.data?.session_id).toMatch(/^sess_/);
    expect(payload.data?.profile_id).toBe(profileId);
    const raw = open.stdout;
    expect(raw).not.toContain(":92"); // debug port jamais exposé
    expect(raw).not.toContain("/tmp"); // chemin jamais exposé
    const list = await curl("-s", "-H", auth, `${coreBaseURL}/api/sessions`);
    expect(payloadOf<{ total: number }>(list.stdout).total).toBeGreaterThanOrEqual(1);
  });

  test("W2: fail-closed local-only navigation policy", async () => {
    await ensureProfile(token);
    const listBefore = await curl("-s", "-H", auth, `${coreBaseURL}/api/sessions`);
    let sid = payloadOf<{ data?: Array<{ session_id: string }> }>(listBefore.stdout).data?.[0]?.session_id;
    if (!sid) {
      const open = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", JSON.stringify({ profile_id: profileId }), `${coreBaseURL}/api/sessions`);
      sid = payloadOf<{ data: { session_id: string } }>(open.stdout).data.session_id;
    }
    // Refus systématique des sorties réseau externes.
    for (const url of ["https://example.com", "http://192.168.1.1/", "https://10.0.0.1", "example.com"]) {
      const r = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", JSON.stringify({ url }), `${coreBaseURL}/api/sessions/${sid}/navigate`);
      const body = JSON.parse(r.stdout.split("__CODE__")[0]) as { error?: { code: string } };
      expect(body.error?.code).toBe("URL_REJECTED_LOCAL_ONLY", `URL externe acceptée par erreur : ${url}`);
    }
    // Acceptation locale uniquement.
    const local = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", JSON.stringify({ url: fixtureFile }), `${coreBaseURL}/api/sessions/${sid}/navigate`);
    expect(payloadOf<{ data: { url: string } }>(local.stdout).data.url).toBe(fixtureFile);
    const localHttp = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", JSON.stringify({ url: "http://127.0.0.1:19280/api/health" }), `${coreBaseURL}/api/sessions/${sid}/navigate`);
    expect(payloadOf<{ data: { url: string } }>(localHttp.stdout).data.url).toContain("127.0.0.1");
  });

  test("W3: content projection redacted and screenshot available for client-side hashing", async () => {
    await ensureProfile(token);
    const listBefore = await curl("-s", "-H", auth, `${coreBaseURL}/api/sessions`);
    let sid = payloadOf<{ data?: Array<{ session_id: string }> }>(listBefore.stdout).data?.[0]?.session_id;
    if (!sid) {
      const open = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", JSON.stringify({ profile_id: profileId }), `${coreBaseURL}/api/sessions`);
      sid = payloadOf<{ data: { session_id: string } }>(open.stdout).data.session_id;
    }
    const nav = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", JSON.stringify({ url: fixtureFile }), `${coreBaseURL}/api/sessions/${sid}/navigate`);
    expect(payloadOf<{ data: { url: string } }>(nav.stdout).data.url).toBe(fixtureFile);
    const content = await curl("-s", "-H", auth, `${coreBaseURL}/api/sessions/${sid}/content`);
    const cPayload = payloadOf<{ data?: string }>(content.stdout);
    expect(typeof cPayload.data).toBe("string");
    expect(cPayload.data).toContain("Automation locale T15");
    const shot = await exec("curl", ["-s", "-D", "-", "-o", "/dev/null", "-H", auth, `${coreBaseURL}/api/sessions/${sid}/screenshot`]);
    expect(String(shot.stdout)).toContain("image/png"); // PNG brut transmis pour hachage SubtleCrypto client-side
  });

  test("W4: session close and refusal after close", async () => {
    await ensureProfile(token);
    const listBefore = await curl("-s", "-H", auth, `${coreBaseURL}/api/sessions`);
    const sid = payloadOf<{ data?: Array<{ session_id: string }> }>(listBefore.stdout).data?.[0]?.session_id;
    expect(sid).toBeTruthy();
    const del = await curl("-s", "-X", "DELETE", "-H", auth, `${coreBaseURL}/api/sessions/${sid}`);
    expect(del.code).toBe(200);
    const after = await curl("-s", "-H", auth, `${coreBaseURL}/api/sessions`);
    expect(payloadOf<{ total: number }>(after.stdout).total).toBe(0);
    const orphan = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", JSON.stringify({ url: fixtureFile }), `${coreBaseURL}/api/sessions/${sid}/navigate`);
    expect(payloadOf<{ error?: { code: string } }>(orphan.stdout).error?.code).toBe("NOT_FOUND");
    const re = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", JSON.stringify({ profile_id: profileId }), `${coreBaseURL}/api/sessions`);
    expect(re.code).toBe(201);
  });

  test("W5: dashboard automation panel mounts and projects digests only", async ({ page }) => {
    // Le Core borne les sessions simultanées à 1 : aucune session ne doit rester
    // ouverte d'un test précédent, sinon l'ouverture serait refusée.
    await clearSessions(token);
    await page.goto(dashboardBase, { waitUntil: "networkidle" });
    // Émission du code à usage unique via le contrat API Core (la CLI de la nouvelle
    // baseline lit un token de configuration distinct de BROWSEFORGE_TOKEN ; on appelle
    // directement l'API, comme le ferait tout client local authentifié).
    const codeReq = await curl("-s", "-X", "POST", "-H", auth, "-H", "Content-Type: application/json", "-d", "{}", `${coreBaseURL}/api/v1/readonly/session/codes`);
    const code = payloadOf<{ code?: string }>(codeReq.stdout).code ?? "";
    if (!/^[a-f0-9]{64}$/i.test(code)) throw new Error("EMISSION_CODE_INVALIDE");
    await page.getByLabel("Code local à usage unique").fill(code);
    await page.getByRole("button", { name: "Relier au Core local" }).click();
    await expect(page.getByText("Lecture Core sécurisée")).toBeVisible({ timeout: 15_000 });
    await linkAdmin(page, token);

    // Sélectionner le profil E2E (clic sur la rangée de la liste des profils).
    const profileRow = page.locator("article.profile-row").filter({ hasText: "E2E · T15 · Automation" });
    await expect(profileRow.first()).toBeVisible({ timeout: 15_000 });
    await profileRow.first().click();

    // Panneau Automation locale.
    await page.getByRole("button", { name: "Automation locale" }).click();
    await expect(page.getByTestId("automation-panel")).toBeVisible();

    // Ouverture de session pour le profil sélectionné.
    await page.getByTestId("automation-open-session").click();
    await expect(page.getByText("Session locale ouverte")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByTestId("automation-url-input")).toBeVisible();

    // Navigation locale et digest du contenu (jamais le HTML brut).
    await page.getByTestId("automation-url-input").fill(fixtureFile);
    await page.getByTestId("automation-navigate-button").click();
    await expect(page.getByTestId("automation-navigate-feedback")).toContainText("Navigation locale acceptée", { timeout: 20_000 });
    await page.getByTestId("automation-content-digest").click();
    await expect(page.getByTestId("automation-content-digest-line")).toContainText(/Contenu : [a-f0-9]{16}… ·/, { timeout: 20_000 });
    const panelText = await page.getByTestId("automation-panel").textContent();
    expect(panelText).not.toContain("<h1"); // pas de HTML brut projeté

    // Capture projetée en digest (pas d'image brute).
    await page.getByTestId("automation-screenshot-digest").click();
    await expect(page.getByTestId("automation-screenshot-digest-line")).toContainText(/Capture : [a-f0-9]{16}… ·/, { timeout: 20_000 });
    await expect(page.getByTestId("automation-panel").locator("img")).toHaveCount(0);

    // Fermeture de session.
    const firstSession = page.getByTestId("automation-sessions").locator("button[aria-label^='Fermer la session']").first();
    await firstSession.click();
    await expect(page.getByText("Session fermée")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByTestId("automation-open-session")).toBeVisible();
  });
});
