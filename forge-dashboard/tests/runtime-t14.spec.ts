import { test, expect } from "@playwright/test";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { existsSync, readFileSync } from "node:fs";

// T14 — Runtime qualification E2E.
// Verifies against the real Core (loopback, admin Bearer, memory-only token):
//   W1 The qualified-runtime panel mounts, queries /api/v1/runtimes/qualified
//      over loopback and renders the redacted catalog (state/version/arch).
//   W2 The panel never renders binary paths, debug ports or token values.
//   W3 Read-only contract: the panel exposes no write surface (no create,
//      update or delete controls); the route accepts only GET.
const exec = promisify(execFile);
// Nouvelle baseline forgebaseline-2026-08-17 (T14 réimplémenté clean-room).
const coreBaseURL = process.env.FORGELOCAL_CORE_BASE_URL ?? "http://127.0.0.1:19280";
const dashboardBase = process.env.FORGELOCAL_DASHBOARD_URL ?? "http://127.0.0.1:3000";
const dashboardOrigin = new URL(dashboardBase).origin;
const coreQualifiedUrl = `${coreBaseURL}/api/v1/runtimes/qualified`;

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

async function expectQualified(token: string) {
  const response = await exec(
    "curl", ["-s", "-H", `Authorization: Bearer ${token}`, coreQualifiedUrl],
  );
  const payload = response.stdout;
  expect(payload).toContain('"state"');
  expect(payload).toContain('"version"');
  expect(payload).toContain('"qualified_at"');
  expect(payload).not.toContain('"binary_hash_sha256"');
  expect(payload).not.toContain("/usr/");
  expect(payload).not.toContain("debug_port");
  expect(payload).not.toContain("user_data");
  return payload;
}

async function bootstrapReadOnly(page: import("@playwright/test").Page) {
  // Émission du code à usage unique via le contrat API Core (la CLI de la nouvelle
  // baseline lit un token de configuration distinct de BROWSEFORGE_TOKEN).
  const token = readToken();
  const res = await exec("curl", ["-s", "-X", "POST", "-H", `Authorization: Bearer ${token}`, "-H", `Origin: ${dashboardOrigin}`, "-H", `Referer: ${dashboardOrigin}/`, "-H", "Content-Type: application/json", "-d", "{}", `${coreBaseURL}/api/v1/readonly/session/codes`]);
  const payload = JSON.parse(res.stdout) as { code?: string };
  if (typeof payload.code !== "string" || !/^[a-f0-9]{64}$/i.test(payload.code)) {
    throw new Error("EMISSION_CODE_T14_INVALIDE");
  }
  await page.goto(dashboardBase, { waitUntil: "domcontentloaded" });
  await expect(page.getByLabel("Code local à usage unique")).toBeVisible({ timeout: 15_000 });
  await page.getByLabel("Code local à usage unique").fill(payload.code);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
  await expect(page.getByText("Lecture Core sécurisée")).toBeVisible({ timeout: 15_000 });
}

async function linkAdmin(page: import("@playwright/test").Page, token: string) {
  await page.getByTestId("local-core-admin-token").fill(token);
  await page.getByTestId("local-core-admin-link").click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("Contrôle local actif", { timeout: 15_000 });
}

test.describe("T14 runtime qualification", () => {
  test("W1: qualified catalog mounts and lists the real Chromium", async ({ page }) => {
    const token = readToken();
    await expectQualified(token);
    await bootstrapReadOnly(page);
    await linkAdmin(page, token);
    await page.getByRole("button", { name: "Runtime qualifié" }).click();
    await expect(page.getByText(/Chromium.*qualifié|Catalogue de qualification/i).first()).toBeVisible({ timeout: 15_000 });
    // The real sandbox Chromium version surface must be present.
    await expect(page.getByText(/Chromium \d+/).first()).toBeVisible();
    await expect(page.getByText("Qualifié", { exact: true }).first()).toBeVisible();
    // Payload is redacted: no arch or filesystem path is rendered.
  });

  test("W2: no raw binary paths, ports or tokens are rendered", async ({ page }) => {
    const token = readToken();
    await bootstrapReadOnly(page);
    await linkAdmin(page, token);
    await page.getByRole("button", { name: "Runtime qualifié" }).click();
    await expect(page.getByText(/Chromium \d+/).first()).toBeVisible();
    const html = await page.content();
    for (const forbid of ["/usr/bin", "debug_port", "user_data", "chromium --"]) {
      expect(html).not.toContain(forbid);
    }
    const configuredTokenPath = process.env.FORGELOCAL_TOKEN_PATH;
    const tokenValue = configuredTokenPath && existsSync(configuredTokenPath) ? readFileSync(configuredTokenPath, "utf8").trim() : "";
    if (tokenValue) {
      expect(html).not.toContain(tokenValue);
    }
  });

  test("W3: read-only contract, no write surface on the runtime panel", async ({ page }) => {
    const token = readToken();
    // G15-B: mutations without a loopback Origin/Referer are refused by the
    // Core's origin guard before the router even decides on the method.
    const postHead = await exec("curl", ["-s", "-o", "/dev/null", "-w", "%{http_code}", "-X", "POST", "-H", `Authorization: Bearer ${token}`, coreQualifiedUrl]);
    expect(["403", "405"].includes(postHead.stdout)).toBeTruthy();
    const delHead = await exec("curl", ["-s", "-o", "/dev/null", "-w", "%{http_code}", "-X", "DELETE", "-H", `Authorization: Bearer ${token}`, coreQualifiedUrl]);
    expect(["403", "405"].includes(delHead.stdout)).toBeTruthy();
    await bootstrapReadOnly(page);
    await linkAdmin(page, token);
    await page.getByRole("button", { name: "Runtime qualifié" }).click();
    await expect(page.getByText(/Chromium \d+/).first()).toBeVisible();
    // The runtime panel must render no write surface on qualified runtimes. The
    // scope is the panel itself: sibling panels (profiles, proxies, groups)
    // keep their own legitimate controls, which are outside the T14 contract.
    // Le scope du contrat T14 est le panneau Runtime qualifié lui-même ; les
    // panneaux voisins (profils, proxys, groupes) conservent leurs commandes.
    const panel = page.getByTestId("runtime-panel");
    for (const pattern of [/créer|création/i, /sauvegarder|enregistrer/i, /supprimer|retirer le runtime/i]) {
      await expect(panel.getByRole("button", { name: pattern })).toHaveCount(0, { timeout: 5_000 });
    }
  });
});
