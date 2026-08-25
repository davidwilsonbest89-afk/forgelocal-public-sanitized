import { expect, test } from "@playwright/test";

const dashboardURL = process.env.FORGELOCAL_DASHBOARD_URL ?? "http://127.0.0.1:3001";
const coreURL = "http://127.0.0.1:19280";

async function installReadonlyMocks(page: import("@playwright/test").Page) {
  await page.route(`${coreURL}/api/v1/readonly/**`, async (route) => {
    const path = new URL(route.request().url()).pathname;
    if (path.endsWith("/session/bootstrap")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ token: "readonly-r2-synthetic", expires_at: "2099-08-25T18:00:00Z", scope: "readonly" }) });
      return;
    }
    if (path.endsWith("/summary")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ api_version: "v1", data: { profiles: 0, groups: 0, runtimes: 0 } }) });
      return;
    }
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ api_version: "v1", data: [], page: { limit: 100 } }) });
  });
}

test("R2 — affiche distinctement token admin expiré et révoqué", async ({ page }) => {
  await installReadonlyMocks(page);
  let reason: "expired" | "revoked" = "expired";
  await page.route(`${coreURL}/api/profiles?limit=1`, async (route) => {
    await route.fulfill({ status: 401, contentType: "application/json", body: JSON.stringify({ error: { code: "UNAUTHORIZED", reason } }) });
  });
  await page.goto(dashboardURL, { waitUntil: "networkidle" });
  await page.getByLabel("Code local à usage unique").fill("a".repeat(64));
  await page.getByRole("button", { name: "Relier au Core local" }).click();
  await expect(page.getByText("Lecture Core sécurisée")).toBeVisible();
  const tokenInput = page.getByTestId("local-core-admin-token");
  const linkButton = page.getByTestId("local-core-admin-link");
  await tokenInput.fill("admin-r2-synthetic-token");
  await linkButton.click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("expiré");

  reason = "revoked";
  await tokenInput.fill("admin-r2-synthetic-token");
  await linkButton.click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("révoqué");
  const body = await page.locator("body").innerText();
  expect(body).not.toContain("readonly-r2-synthetic");
});
