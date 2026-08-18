import { test } from "@playwright/test";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { readFileSync } from "node:fs";

// Sonde diagnostique T10 : évaluer l'état du formulaire proxy avant le clic create.
const exec = promisify(execFile);
const coreBaseURL = process.env.FORGELOCAL_CORE_BASE_URL ?? "http://127.0.0.1:19280";
const coreBaseDir = process.env.FORGELOCAL_BASE_DIR ?? "/tmp/forge-e2e-base";
const dashboardBase = process.env.FORGELOCAL_DASHBOARD_URL ?? "http://localhost:3000";
const readonlyToken = "e2e-probe-readonly";

async function runCoreCli(...args: string[]): Promise<string> {
  const bin = process.env.FORGELOCAL_BINARY ?? "/tmp/forge-core-e2e";
  const { stdout } = await exec(bin, ["--base-dir", coreBaseDir, ...args], {
    env: { ...process.env, GOTOOLCHAIN: "local" },
  });
  return stdout;
}

async function bootstrapReadOnly(page: import("@playwright/test").Page): Promise<void> {
  await page.goto(dashboardBase, { waitUntil: "networkidle" });
  const code = await runCoreCli("readonly-session", "code", "--base-url", coreBaseURL, "--json");
  const m = /"code"\s*:\s*"([a-f0-9]{64})"/i.exec(code);
  const emissionCode = m?.[1] ?? "";
  if (!/^[a-f0-9]{64}$/i.test(emissionCode)) throw new Error("EMISSION_CODE_INVALIDE");
  await page.getByLabel("Code local à usage unique").fill(emissionCode);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
  await page.getByText("Lecture Core sécurisée").waitFor({ timeout: 15_000 });
}

function resolveAdminToken(): string {
  const tokenPath = process.env.FORGELOCAL_TOKEN_PATH ?? "/tmp/forge-e2e-token.txt";
  let token = String(readFileSync(tokenPath)).trim();
  if (token.startsWith("CORE_TOKEN=")) token = "";
  if (!token) token = (process.env.BROWSEFORGE_TOKEN ?? "").trim();
  if (token.length < 12) throw new Error("CORE_API_TOKEN_ABSENT:probe");
  return token;
}
async function linkAdmin(page: import("@playwright/test").Page): Promise<void> {
  const token = resolveAdminToken();
  await page.getByTestId("local-core-admin-token").fill(token);
  await page.getByTestId("local-core-admin-link").click();
  await page.getByTestId("local-core-admin-message").waitForText?.("Contrôle local actif", { timeout: 15_000 });
  await page.waitForTimeout(2_000);
}

test("T10 probe: form state before create click", async ({ browser }) => {
  const corePage = await browser.newPage();
  const events: string[] = [];
  corePage.on("response", (res) => {
    const u = res.url();
    if (u.includes("/api/") && (u.includes("proxies") || u.includes("profiles"))) {
      events.push(`RESP ${res.status()} ${u.slice(-60)}`);
    }
  });
  await bootstrapReadOnly(corePage);
  await linkAdmin(corePage);
  process.stdout.write("PROBE_LINKED\n");
  // Purge des résidus.
  try {
    const { stdout } = await exec("curl", ["-s", `${coreBaseURL}/api/v1/readonly/proxies?limit=50`, "-H", `Authorization: Bearer ${readonlyToken}`], { env: { ...process.env, GOTOOLCHAIN: "local" } });
    const payload = JSON.parse(stdout) as { data?: Array<{ id: string; name: string }> };
    for (const item of payload.data ?? []) {
      if (item.name.startsWith("E2E · T10") || item.name === "t10-e2e-paris") {
        await exec("curl", ["-s", "-X", "DELETE", `${coreBaseURL}/api/proxies/${item.id}`, "-H", `Authorization: Bearer ${resolveAdminToken()}`, "-H", "Origin: http://localhost:3000"]);
      }
    }
  } catch {
    // ignore
  }
  await corePage.getByTestId("proxy-name").fill("E2E · T10 · Paris");
  const sel = corePage.getByTestId("proxy-type");
  await sel.selectOption("http");
  await corePage.getByTestId("proxy-host").fill("198.51.100.10");
  await corePage.getByTestId("proxy-port").fill("8080");
  await corePage.getByTestId("proxy-region").fill("eu");
  await corePage.waitForTimeout(1_500);
  const state = await corePage.evaluate(() => {
    const name = (document.querySelector('[data-testid="proxy-name"]') as HTMLInputElement)?.value ?? "MISSING";
    const host = (document.querySelector('[data-testid="proxy-host"]') as HTMLInputElement)?.value ?? "MISSING";
    const port = (document.querySelector('[data-testid="proxy-port"]') as HTMLInputElement)?.value ?? "MISSING";
    const region = (document.querySelector('[data-testid="proxy-region"]') as HTMLInputElement)?.value ?? "MISSING";
    const btn = document.querySelector('[data-testid="proxy-create"]') as HTMLButtonElement | null;
    return { name, host, port, region, buttonDisabled: btn?.disabled, buttonText: btn?.textContent?.trim() };
  });
  process.stdout.write(`PROBE_STATE ${JSON.stringify(state)}\n`);
  process.stdout.write(`PROBE_EVENTS\n${events.join("\n")}\n`);
});
