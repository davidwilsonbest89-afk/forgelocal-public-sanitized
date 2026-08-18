import { createRequire } from "node:module";
/**
 * T16 — reproduction isolée du hang T10 (bouton proxy-create gelé au 2e submit).
 * Reprend fidèlement le setup et le 1er describe de proxies-t10.spec.ts.
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

let readonlyToken: string | undefined;

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
  const bearerPromise = new Promise<string>((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("TOKEN_READONLY_T10_ABSENT")), 10_000);
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

async function listCoreProxies(): Promise<number> {
  if (!readonlyToken) throw new Error("TOKEN_READONLY_T10_ABSENT");
  const response = await fetch(`${coreBaseURL}/api/v1/readonly/proxies`, {
    headers: { Authorization: `Bearer ${readonlyToken}`, Accept: "application/json" },
  });
  if (response.status === 404) return 0;
  if (!response.ok) throw new Error(`CORE_READONLY_PROXIES: status=${response.status}`);
  const payload = (await response.json()) as { data?: Array<{ id: string }> };
  return Array.isArray(payload.data) ? payload.data.length : (payload as { data?: { items?: unknown[] } }).data?.items?.length ?? 0;
}

async function linkAdmin(page: Page, token: string) {
  await page.getByTestId("local-core-admin-token").fill(token);
  await page.getByTestId("local-core-admin-link").click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("Contrôle local actif");
}

async function purgeE2EProxiesReal(): Promise<void> {
  // Purge identique à la spec réelle : listing readonly, suppression admin par curl.
  const { execFile } = await import("node:child_process");
  const { promisify } = await import("node:util");
  const exec = promisify(execFile);
  const adminToken = await readApiToken();
  try {
    const { stdout } = await exec("curl", ["-s", `${coreBaseURL}/api/v1/readonly/proxies?limit=50`, "-H", `Authorization: Bearer ${readonlyToken}`], { env: { ...process.env, GOTOOLCHAIN: "local" } });
    const payload = JSON.parse(stdout) as { data?: Array<{ id: string; name: string }> };
    for (const item of payload.data ?? []) {
      if (item.name.startsWith("E2E · T10") || item.name === "t10-e2e-paris") {
        await exec("curl", ["-s", "-X", "DELETE", `${coreBaseURL}/api/proxies/${item.id}`, "-H", `Authorization: Bearer ${adminToken}`]);
      }
    }
  } catch {
    // Résidu introuvable : on poursuit.
  }
}

test.describe("T10 — référentiel proxy standalone (reproduction hang)", () => {
  test("W1 valide puis W2 port invalide", async ({ page: corePage }) => {
    await bootstrapReadOnly(corePage);
    await linkAdmin(corePage, await readApiToken());
    await purgeE2EProxiesReal();
    const beforeProxiesClean = await listCoreProxies();

    // W1 — proxy valide
    const nameW1 = `E2E · T10 · Paris · ${Date.now()}`;
    await corePage.getByTestId("proxy-name").fill(nameW1);
    await corePage.getByTestId("proxy-type").selectOption("http");
    await corePage.getByTestId("proxy-host").fill("198.51.100.10");
    await corePage.getByTestId("proxy-port").fill("8080");
    await corePage.getByTestId("proxy-region").fill("eu");
    await corePage.getByTestId("proxy-create").click();
    await expect(corePage.getByTestId(/proxy-row/).filter({ hasText: nameW1 })).toBeVisible({ timeout: 10_000 });
    expect(await listCoreProxies()).toBe(beforeProxiesClean + 1);

    // Reproduction du rowText check de la spec réelle (entre W1 et W2).
    const rowText = await corePage.getByTestId(/proxy-row/).first().textContent();
    expect(rowText).not.toMatch(/password|secret=|user=|token=/);

    // W2 — port invalide (le serveur doit refuser explicitement)
    await corePage.getByTestId("proxy-name").fill(`E2E · T10 · Refus · ${Date.now()}`);
    await corePage.getByTestId("proxy-type").selectOption("http");
    await corePage.getByTestId("proxy-host").fill("198.51.100.11");
    await corePage.getByTestId("proxy-port").fill("70000");
    const disabledBefore = await corePage.evaluate(() => (document.querySelector('[data-testid="proxy-create"]') as HTMLButtonElement)?.disabled);
    process.stdout.write(`T10_STANDALONE_DISABLED_BEFORE_CLICK ${JSON.stringify(disabledBefore)}\n`);
    await corePage.getByTestId("proxy-create").click({ timeout: 30_000 });
    const portError = await corePage.evaluate(() => document.body.innerText.includes("port"));
    process.stdout.write(`T10_STANDALONE_W2_AFTERCLICK ${portError ? "PORT_ERROR_VISIBLE" : "NO_PORT_ERROR"}\n`);
    await expect(portError).toBe(true);
    expect(await listCoreProxies()).toBe(beforeProxiesClean + 1);
  });
});
