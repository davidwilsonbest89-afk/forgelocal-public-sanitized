import { createRequire } from "node:module";
/**
 * T16 — sonde de diagnostic : captures les valeurs DOM des champs du formulaire
 * proxy APRÈS les fills, pour déterminer si le bouton gelé vient de valeurs
 * DOM vides (fills qui ne tiennent pas) ou de valeurs présentes (autre cause).
 */
import { expect, test, type Page } from "@playwright/test";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
const exec = promisify(execFile);
const baseDir = process.env.FORGELOCAL_BASE_DIR ?? "/tmp/t16-core";
const coreBaseURL = process.env.FORGELOCAL_CORE_BASE_URL ?? "http://127.0.0.1:19280";
const dashboardURL = process.env.FORGELOCAL_DASHBOARD_URL ?? "http://localhost:3000";
const coreBinary = process.env.FORGELOCAL_BINARY ?? "/home/ubuntu/t16-evidence/forgelocal-core-t15-rebuilt";
function payloadOf<T>(raw: string): T {
  try {
    return JSON.parse(raw) as T;
  } catch {
    return {} as T;
  }
}
function required(name: string) {
  const value = process.env[name];
  if (!value) throw new Error(`CONFIGURATION_T16_ABSENTE:${name}`);
  return value;
}
required("FORGELOCAL_BINARY");
required("FORGELOCAL_BASE_DIR");
required("FORGELOCAL_CORE_BASE_URL");
required("FORGELOCAL_DASHBOARD_URL");
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

async function captureBearer(page: Page) {
  return new Promise<string>(resolve => {
    let timer: NodeJS.Timeout;
    page.on("request", request => {
      const header = request.headers()["authorization"];
      if (header && header.startsWith("Bearer ") && !readonlyToken && request.url().includes("/api/v1/readonly/")) {
        clearTimeout(timer);
        readonlyToken = header.slice("Bearer ".length);
        resolve(readonlyToken);
      }
    });
    timer = setTimeout(() => resolve(""), 30_000);
  });
}
async function bootstrapReadOnly(page: Page) {
  const bearerPromise = captureBearer(page);
  const bootstrap = await exec(coreBinary, ["--base-dir", baseDir, "readonly-session", "code", "--base-url", coreBaseURL, "--json"], { env: { ...process.env, GOTOOLCHAIN: "local" } });
  const payload = { code: payloadOf<{ code?: string }>(bootstrap.stdout).code ?? "" };
  await page.goto(dashboardURL);
  await page.getByLabel("Code local à usage unique").fill(payload.code);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
  await expect(page.getByText("Lecture Core sécurisée")).toBeVisible();
  readonlyToken = await bearerPromise;
}
test("T16-VERIFY — fills proxy après purge + capture DOM", async ({ browser }) => {
  const page = await browser.newPage();
  await bootstrapReadOnly(page);
  const adminToken = await readApiToken();
  await page.getByTestId("local-core-admin-token").fill(adminToken);
  await page.getByTestId("local-core-admin-link").click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("Contrôle local actif");
  // purge résidus (identique à la vraie spec T10)
  try {
    const { stdout } = await exec("curl", ["-s", `${coreBaseURL}/api/v1/readonly/proxies?limit=50`, "-H", `Authorization: Bearer ${readonlyToken}`]);
    for (const item of (JSON.parse(stdout) as { data?: Array<{ id: string; name: string }> }).data ?? []) {
      if (item.name.startsWith("E2E") || item.name === "t10-e2e-paris") {
        await exec("curl", ["-s", "-X", "DELETE", `${coreBaseURL}/api/proxies/${item.id}`, "-H", `Authorization: Bearer ${adminToken}`]);
      }
    }
  } catch {
    // Résidu introuvable : on poursuit.
  }
  const captureDom = async () =>
    page.evaluate(() => {
      const inputs = Array.from(document.querySelectorAll<HTMLInputElement>('input[data-testid^="proxy-"]'));
      const btn = document.querySelector('button[data-testid="proxy-create"]');
      return {
        inputs: inputs.map(i => ({ id: i.dataset.testid, value: i.value })),
        button: btn ? { disabled: btn.disabled, text: (btn.textContent ?? "").trim() } : null,
      };
    });
  // fills identiques à la vraie spec, avec capture DOM après chaque étape
  await page.getByTestId("proxy-name").fill("E2E · T16 · Paris");
  console.log("T16_AFTER_NAME:", JSON.stringify(await captureDom()));
  await page.getByTestId("proxy-type").selectOption("http");
  console.log("T16_AFTER_SELECT:", JSON.stringify(await captureDom()));
  await page.getByTestId("proxy-host").fill("198.51.100.10");
  console.log("T16_AFTER_HOST:", JSON.stringify(await captureDom()));
  await page.getByTestId("proxy-port").fill("8080");
  console.log("T16_AFTER_PORT:", JSON.stringify(await captureDom()));
  await page.getByTestId("proxy-region").fill("eu");
  const dom = await captureDom();
  console.log("T16_DOM:", JSON.stringify(dom));
  expect(dom.inputs.find(i => i.id === "proxy-port")?.value).toBe("8080");
  expect(dom.inputs.find(i => i.id === "proxy-name")?.value).toBe("E2E · T16 · Paris");
  expect(dom.button?.disabled).toBe(false);
});
