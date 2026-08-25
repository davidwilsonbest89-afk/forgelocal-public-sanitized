/**
 * ForgeLocal — T05 BOOTSTRAP-RO-01.
 * Ce test ne journalise ni le code à usage unique, ni le Bearer court. Les
 * valeurs transitent uniquement en mémoire du processus de test et du navigateur.
 */
import AxeBuilder from "@axe-core/playwright";
import { expect, test, type Page } from "@playwright/test";
import { execFile as execFileCallback } from "node:child_process";
import { writeFile } from "node:fs/promises";
import { promisify } from "node:util";

const execFile = promisify(execFileCallback);
const coreBaseURL = required("FORGELOCAL_CORE_BASE_URL");
const dashboardURL = required("FORGELOCAL_DASHBOARD_URL");
const binary = required("FORGELOCAL_BINARY");
const baseDir = required("FORGELOCAL_BASE_DIR");
const hostedDashboardURL = process.env.FORGELOCAL_HOSTED_DASHBOARD_URL;

function required(name: string) {
  const value = process.env[name];
  if (!value) throw new Error(`CONFIGURATION_T05_ABSENTE:${name}`);
  return value;
}

async function issueCode() {
  const { stdout } = await execFile(binary, [
    "--base-dir", baseDir, "readonly-session", "code", "--base-url", coreBaseURL, "--json",
  ], { env: { ...process.env, GOTOOLCHAIN: "local" } });
  const payload = JSON.parse(stdout) as { code?: unknown; expires_at?: unknown };
  if (typeof payload.code !== "string" || !/^[a-f0-9]{64}$/i.test(payload.code) || typeof payload.expires_at !== "string") {
    throw new Error("EMISSION_CODE_T05_INVALIDE");
  }
  return payload.code;
}

async function connect(page: Page, code: string) {
  await page.goto(dashboardURL, { waitUntil: "networkidle" });
  await page.getByLabel("Code local à usage unique").fill(code);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
}

test.describe.configure({ mode: "serial" });

test("T05 — bootstrap loopback, rejeu, expiration, 401 et non-persistance", async ({ browser }) => {
  const observed: Array<{ url: string; method: string; authorization: boolean }> = [];
  const context = await browser.newContext();
  const page = await context.newPage();
  // Le dashboard résout le Core depuis window.location.hostname (localhost), tandis que
  // coreBaseURL pointe vers 127.0.0.1 ; les deux hôtes loopback atteignent le même Core.
  const coreHosts = [coreBaseURL, coreBaseURL.replace("127.0.0.1", "localhost")];
  page.on("request", (request) => {
    if (coreHosts.some((host) => request.url().startsWith(host))) {
      observed.push({
        url: request.url(),
        method: request.method(),
        authorization: Boolean(request.headers().authorization),
      });
    }
  });

  const firstCode = await issueCode();
  await connect(page, firstCode);
  await expect(page.getByText("Lecture Core sécurisée")).toBeVisible();
  await expect(page.getByText("Lecture Core active — session limitée et non persistée.")).toBeVisible();
  await expect(page.getByText(/affiché.*depuis le Core, en lecture seule/)).toBeVisible();
  const axeResults = await new AxeBuilder({ page }).analyze();
  const axeResultsPath = process.env.FORGELOCAL_AXE_RESULTS_PATH;
  if (axeResultsPath) {
    await writeFile(axeResultsPath, JSON.stringify(axeResults, null, 2));
  }
  const blockingAxeViolations = axeResults.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical");
  expect(blockingAxeViolations, "Axe serious/critical violations").toEqual([]);
  await expect.poll(() => observed.length).toBeGreaterThanOrEqual(3);
  expect(observed.some((request) => request.method === "POST" && request.url.endsWith("/api/v1/readonly/session/bootstrap") && !request.authorization)).toBeTruthy();
  expect(observed.filter((request) => request.method === "GET").every((request) => request.authorization)).toBeTruthy();
  expect(observed.every((request) => !request.url.includes(firstCode) && !new URL(request.url).searchParams.has("token"))).toBeTruthy();

  const browserStorage = await page.evaluate(async () => ({
    href: location.href,
    localStorage: Object.entries(localStorage),
    sessionStorage: Object.entries(sessionStorage),
    indexedDB: typeof indexedDB.databases === "function" ? await indexedDB.databases() : [],
    cacheNames: "caches" in globalThis ? await caches.keys() : [],
  }));
  expect(browserStorage.href).not.toContain(firstCode);
  expect(browserStorage.localStorage).toEqual([]);
  expect(browserStorage.sessionStorage).toEqual([]);
  expect(browserStorage.indexedDB).toEqual([]);
  expect(browserStorage.cacheNames).toEqual([]);
  process.stdout.write("T05_BROWSER_STORAGE: PASS url=clean localStorage=0 sessionStorage=0 indexedDB=0 caches=0\n");

  const replayPage = await context.newPage();
  await connect(replayPage, firstCode);
  await expect(replayPage.getByText("Code refusé, expiré ou déjà utilisé. Générez-en un nouveau localement.")).toBeVisible();
  await expect(replayPage.getByText("Lecture Core sécurisée")).toHaveCount(0);
  process.stdout.write("T05_REPLAY: PASS status=401 ui=disconnected\n");

  const expirationCode = await issueCode();
  await page.waitForTimeout(605_000);
  const expirationPage = await context.newPage();
  await connect(expirationPage, expirationCode);
  await expect(expirationPage.getByText("Code refusé, expiré ou déjà utilisé. Générez-en un nouveau localement.")).toBeVisible();
  process.stdout.write("T05_EXPIRY: PASS status=401 ui=disconnected\n");

  const unauthorizedPage = await context.newPage();
  await unauthorizedPage.route("**/api/v1/readonly/summary", (route) => route.fulfill({
    status: 401,
    contentType: "application/json",
    body: '{"error":{"code":"UNAUTHORIZED"}}',
  }));
  const unauthorizedCode = await issueCode();
  await connect(unauthorizedPage, unauthorizedCode);
  await expect(unauthorizedPage.getByText("Connexion impossible. Vérifiez que ce dashboard est servi par le Core local.")).toBeVisible();
  await expect(unauthorizedPage.getByText("Lecture Core sécurisée")).toHaveCount(0);
  await expect(unauthorizedPage.getByText("Core non connecté", { exact: true })).toBeVisible();
  process.stdout.write("T05_FORCED_401: PASS token=cleared ui=Core_non_connecte\n");

  if (hostedDashboardURL) {
    const hostedPage = await context.newPage();
    await hostedPage.goto(hostedDashboardURL, { waitUntil: "domcontentloaded" });
    await expect(hostedPage.getByText("Cette prévisualisation hébergée ne peut pas recevoir un code local.")).toBeVisible();
    await expect(hostedPage.getByLabel("Code local à usage unique")).toHaveCount(0);
    process.stdout.write("T05_HOSTED_ORIGIN: PASS local_code_input=absent\n");
  }
  process.stdout.write("T05_BOOTSTRAP_LOOPBACK: PASS exchange=accepted reads=redacted\n");
  await context.close();
});
