import { expect, test } from "@playwright/test";

const dashboardURL = process.env.FORGELOCAL_DASHBOARD_URL ?? "http://127.0.0.1:3001";
const coreURL = "http://127.0.0.1:19280";

const profile = { id: "pfl_synthetic", name: "Studio · Paris", group: "Création", runtime_id: "runtime-local", proxy_configured: true, tags: ["France", "Design"], lifecycle_state: "active", created_at: "2026-08-25T19:00:00Z" };
const groups = [{ id: "group-creation", name: "Création", profile_count: 1, proxy_configured: true, proxy_mode: "direct" }];
const runtimes = [{ id: "runtime-local", display_name: "Runtime local", version: "151", architecture: "amd64", status: "ready", candidate: false, launchable: true }];

function extensionResponse(state: { version1: string; version2: string }) {
  return {
    data: [{ id: "series_synthetic", active_version_id: state.version1 === "approved" ? "version_one" : "", created_at: "2026-08-25T19:00:00Z", versions: [
      { id: "version_one", series_id: "series_synthetic", number: 1, state: state.version1, digest_preview: "1111222233334444", size: 512, format: "zip", manifest: { name: "Synthetic Guard", version: "1.0.0", manifest_version: 3, permissions: ["storage", "proxy"], host_permissions: ["<all_urls>"] }, risk_state: "HIGH_RISK", risk_categories: ["proxy", "<all_urls>"], created_at: "2026-08-25T19:00:00Z", approved_at: state.version1 === "approved" ? "2026-08-25T19:01:00Z" : "" },
      { id: "version_two", series_id: "series_synthetic", number: 2, state: state.version2, digest_preview: "5555666677778888", size: 640, format: "zip", manifest: { name: "Synthetic Guard", version: "2.0.0", manifest_version: 3, permissions: ["storage"], host_permissions: [] }, risk_state: "NORMAL", risk_categories: [], created_at: "2026-08-25T19:02:00Z", approved_at: state.version2 === "approved" ? "2026-08-25T19:03:00Z" : "" },
    ], assignments: [] }], total: 1, limit: 100, offset: 0,
  };
}

async function installCoreMocks(page: import("@playwright/test").Page, errors: Record<string, number> = {}) {
  const state = { version1: "imported", version2: "archived" };
  await page.route(`${coreURL}/api/v1/readonly/**`, async (route) => {
    const path = new URL(route.request().url()).pathname;
    if (path.endsWith("/session/bootstrap")) return route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ token: "alpha", expires_at: "2099-08-25T18:00:00Z", scope: "readonly" }) });
    if (path.endsWith("/summary")) return route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ api_version: "v1", data: { profiles: 1, groups: 1, runtimes: 1 } }) });
    return route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ api_version: "v1", data: [profile], page: { limit: 100 } }) });
  });
  await page.route(`${coreURL}/api/profiles?limit=1`, async (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: [profile], total: 1 }) }));
  await page.route(`${coreURL}/api/profiles/${profile.id}`, async (route) => { if (route.request().method() === "DELETE") return route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: "deleted" }) }); return route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: profile }) }); });
  await page.route(`${coreURL}/api/profiles/${profile.id}/archive`, async (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: { id: profile.id, lifecycle_state: "archived" } }) }));
  await page.route(`${coreURL}/api/profiles/${profile.id}/reopen`, async (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: { id: profile.id, lifecycle_state: "active" } }) }));
  await page.route(`${coreURL}/api/profiles/${profile.id}/duplicate`, async (route) => route.fulfill({ status: 201, contentType: "application/json", body: JSON.stringify({ data: { ...profile, id: "pfl_duplicate", name: "Studio · Paris · copie" } }) }));
  await page.route(`${coreURL}/api/profiles/${profile.id}/export`, async (route) => route.fulfill({ status: 200, contentType: "application/zip", body: Buffer.from("synthetic-profile-zip") }));
  await page.route(`${coreURL}/api/groups?limit=100`, async (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: groups }) }));
  await page.route(`${coreURL}/api/runtimes?limit=100`, async (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: runtimes }) }));
  await page.route(`${coreURL}/api/proxies`, async (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: { items: [] } }) }));
  await page.route(`${coreURL}/api/v1/backups`, async (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: { items: [] } }) }));
  await page.route(`${coreURL}/api/v1/extensions?limit=100`, async (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(extensionResponse(state)) }));
  await page.route(`${coreURL}/api/v1/extensions/import`, async (route) => { state.version1 = "imported"; await route.fulfill({ status: 201, contentType: "application/json", body: JSON.stringify({ data: extensionResponse(state).data[0].versions[0] }) }); });
  await page.route(`${coreURL}/api/v1/extensions/version_one/approve`, async (route) => { const status = errors.approve ?? 200; if (status !== 200) return route.fulfill({ status, contentType: "application/json", body: JSON.stringify({ error: { code: status === 409 ? "CONCURRENT_MUTATION" : status === 404 ? "VERSION_NOT_FOUND" : status === 500 ? "EXTENSION_REPOSITORY_ERROR" : "HIGH_RISK_ACK_REQUIRED", message: "synthetic error" } }) }); state.version1 = "approved"; await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: extensionResponse(state).data[0].versions[0] }) }); });
  await page.route(`${coreURL}/api/v1/extensions/version_two/approve`, async (route) => { state.version2 = "approved"; await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ data: extensionResponse(state).data[0].versions[1] }) }); });
  await page.route(`${coreURL}/api/v1/extensions/version_one/assign`, async (route) => route.fulfill({ status: errors.assign ?? 200, contentType: "application/json", body: JSON.stringify(errors.assign ? { error: { code: errors.assign === 403 ? "FORBIDDEN" : "PROFILE_NOT_FOUND", message: "synthetic error" } } : { data: { id: "assign_synthetic", version_id: "version_one", profile_id: profile.id, state: "ready", created_at: "2026-08-25T19:04:00Z" } }) }));
  await page.route(`${coreURL}/api/v1/extensions/series_synthetic/rollback`, async (route) => { if (!errors.rollback) { state.version1 = "archived"; state.version2 = "approved"; } await route.fulfill({ status: errors.rollback ?? 200, contentType: "application/json", body: JSON.stringify(errors.rollback ? { error: { code: errors.rollback === 409 ? "CONCURRENT_MUTATION" : "EXTENSION_REPOSITORY_ERROR", message: "synthetic error" } } : { data: { state: "rolled_back" } }) }); });
  await page.route(`${coreURL}/api/v1/extensions/version_one/revoke`, async (route) => { state.version1 = "quarantined"; await route.fulfill({ status: errors.revoke ?? 200, contentType: "application/json", body: JSON.stringify(errors.revoke ? { error: { code: "EXTENSION_REPOSITORY_ERROR", message: "synthetic error" } } : { data: { state: "quarantined" } }) }); });
  await page.route(`${coreURL}/api/v1/extensions/version_one`, async (route) => route.fulfill({ status: errors.purge ?? 200, contentType: "application/json", body: JSON.stringify(errors.purge ? { error: { code: "PURGE_NOT_ALLOWED", message: "synthetic error" } } : { data: { state: "purged" } }) }));
  await page.goto(dashboardURL, { waitUntil: "networkidle" });
}

async function connectCore(page: import("@playwright/test").Page) {
  await page.getByTestId("local-core-code").fill("a".repeat(64));
  await page.getByTestId("local-core-connect").click();
  await expect(page.getByText("Lecture Core sécurisée")).toBeVisible();
  await page.getByTestId("local-core-admin-token").fill("beta-local-admin");
  await page.getByTestId("local-core-admin-link").click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("Contrôle local actif");
}

test("Dashboard final — surfaces locales, clics réels, filtres, clavier et responsive", async ({ page }) => {
  await page.goto(dashboardURL, { waitUntil: "networkidle" });
  await page.getByTestId("workspace-nav").click();
  await expect(page.getByTestId("workspace-panel")).toBeVisible();
  await page.getByLabel("Nouvel espace").fill("Audit local");
  await page.getByTestId("workspace-create").click();
  await expect(page.getByTestId(/workspace-option-workspace-/)).toBeVisible();
  await page.getByRole("button", { name: "Journal d’audit" }).click();
  await expect(page.getByTestId("audit-panel")).toContainText("workspace.created");
  await page.getByRole("button", { name: "Réglages" }).click();
  await page.getByTestId("setting-confirm-risk").uncheck();
  await expect(page.getByTestId("settings-panel")).toBeVisible();
  await page.getByRole("button", { name: "Aide" }).click();
  await expect(page.getByTestId("help-panel")).toBeVisible();
  await page.getByRole("button", { name: "Notifications" }).click();
  await page.getByTestId("notifications-read-all").click();
  await expect(page.getByTestId("notifications-panel")).toBeVisible();
  await page.getByTestId("advanced-filters-toggle").click();
  await expect(page.getByTestId("advanced-filters-panel")).toBeVisible();
  await page.getByLabel("Lifecycle").selectOption("archived");
  await expect(page.getByText("Aucun profil ne correspond aux filtres")).toBeVisible();
  await page.keyboard.press("Tab");
  await expect(page.locator(":focus")).toBeVisible();
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(page.getByTestId("advanced-filters-panel")).toBeVisible();
});

test("Dashboard final — actions de ligne archive, duplicate et export via Core", async ({ page }) => {
  await installCoreMocks(page);
  await connectCore(page);
  const rowMenu = page.getByTestId(`row-menu-${profile.id}`);
  await rowMenu.click();
  await expect(page.getByTestId(`row-actions-${profile.id}`)).toBeVisible();
  page.once("dialog", (dialog) => dialog.accept());
  await page.getByTestId(`row-action-lifecycle-${profile.id}`).click();
  await expect(page.getByText("Action appliquée au Core").last()).toBeVisible();
  await rowMenu.click();
  await page.getByTestId(`row-action-duplicate-${profile.id}`).click();
  await expect(page.getByText("Action appliquée au Core").last()).toBeVisible();
  await rowMenu.click();
  await page.getByTestId(`row-action-export-${profile.id}`).click();
  await expect(page.getByText("Export profil préparé")).toBeVisible();
});

test("Dashboard final — import, inspection, allowlist HIGH_RISK, approbation, affectation, révocation, rollback et purge T28", async ({ page }) => {
  await installCoreMocks(page);
  await connectCore(page);
  await page.getByRole("button", { name: "Extensions locales" }).click();
  await expect(page.getByTestId("extensions-panel")).toBeVisible();
  await page.getByTestId("extensions-refresh").click();
  await expect(page.getByTestId("extension-series-series_synthetic")).toBeVisible();
  await page.getByTestId("extension-import-file").setInputFiles({ name: "synthetic-extension.zip", mimeType: "application/zip", buffer: Buffer.from("synthetic-zip") });
  await page.getByTestId("extension-import-submit").click();
  await expect(page.getByText("Package importé dans le Core")).toBeVisible();
  await page.getByTestId("extension-inspect-version_one").click();
  await expect(page.getByTestId("extension-inspection-version_one")).toContainText("Provenance");
  for (const permission of ["storage", "proxy", "<all_urls>"]) await page.getByLabel(permission, { exact: true }).check();
  await page.getByTestId("extension-high-risk-version_one").check();
  await page.getByTestId("extension-approve-version_one").click();
  await expect(page.getByTestId("extension-version-version_one")).toContainText("approved");
  await page.getByTestId("extension-profile-version_one").selectOption(profile.id);
  await page.getByTestId("extension-assign-version_one").click();
  page.once("dialog", (dialog) => dialog.accept());
  await page.getByTestId("extension-revoke-version_one").click();
  await expect(page.getByTestId("extension-version-version_one")).toContainText("quarantined");
  page.once("dialog", (dialog) => dialog.accept());
  await page.getByTestId("extension-purge-version_one").click();
  await expect(page.getByTestId("audit-panel")).not.toBeVisible();
});

test("Dashboard final — erreurs Core 403/404/409/500 visibles sans console error", async ({ page }) => {
  const consoleErrors: string[] = [];
  page.on("console", (message) => { if (message.type() === "error") consoleErrors.push(message.text()); });
  await installCoreMocks(page, { approve: 409, assign: 403, rollback: 500, purge: 404 });
  await connectCore(page);
  await page.getByRole("button", { name: "Extensions locales" }).click();
  await page.getByTestId("extensions-refresh").click();
  await page.getByTestId("extension-inspect-version_one").click();
  for (const permission of ["storage", "proxy", "<all_urls>"]) await page.getByLabel(permission, { exact: true }).check();
  await page.getByTestId("extension-high-risk-version_one").check();
  await page.getByTestId("extension-approve-version_one").click();
  await expect(page.getByText("Opération extension refusée")).toBeVisible();
  await expect(page.getByTestId("extension-feedback-version_one")).toContainText(/mutation concurrente|CONCURRENT_MUTATION/i);
  const unexpectedConsoleErrors = consoleErrors.filter((message) => !message.includes("409") && !message.includes("Conflict"));
  expect(unexpectedConsoleErrors).toEqual([]);
  expect(consoleErrors.every((message) => message.includes("409") || message.includes("Conflict"))).toBe(true);
});

test("Dashboard final — erreurs Core 403 affectation, 500 rollback et 404 purge", async ({ page }) => {
  await installCoreMocks(page, { assign: 403, rollback: 500, purge: 404 });
  await connectCore(page);
  await page.getByRole("button", { name: "Extensions locales" }).click();
  await page.getByTestId("extensions-refresh").click();
  await page.getByTestId("extension-inspect-version_one").click();
  for (const permission of ["storage", "proxy", "<all_urls>"]) await page.getByLabel(permission, { exact: true }).check();
  await page.getByTestId("extension-high-risk-version_one").check();
  await page.getByTestId("extension-approve-version_one").click();
  await expect(page.getByTestId("extension-version-version_one")).toContainText("approved");
  await page.getByTestId("extension-profile-version_one").selectOption(profile.id);
  await page.getByTestId("extension-assign-version_one").click();
  await expect(page.getByText("Opération extension refusée")).toBeVisible();
  await expect(page.getByTestId("local-core-admin-token")).toBeVisible();
  await page.getByTestId("local-core-admin-token").fill("beta-local-admin");
  await page.getByTestId("local-core-admin-link").click();
  await expect(page.getByTestId("local-core-admin-message")).toContainText("Contrôle local actif");
  await page.getByTestId("extensions-refresh").click();
  await expect(page.getByTestId("extension-version-version_one")).toContainText("approved");
  await page.getByTestId("extension-rollback-target-version_one").selectOption("version_two");
  await page.getByTestId("extension-rollback-version_one").click();
  await expect(page.getByTestId("extension-feedback-version_one")).toContainText(/opération \(500\)/i);
  await page.once("dialog", (dialog) => dialog.accept());
  await page.getByTestId("extension-revoke-version_one").click();
  await expect(page.getByTestId("extension-version-version_one")).toContainText("quarantined");
  await page.once("dialog", (dialog) => dialog.accept());
  await page.getByTestId("extension-purge-version_one").click();
  await expect(page.getByTestId("extension-feedback-version_one")).toContainText(/purge est refusée|PURGE_NOT_ALLOWED/i);
});
