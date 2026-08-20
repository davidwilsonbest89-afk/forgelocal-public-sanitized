/**
 * ForgeLocal — T06 Groupes/Runtimes en lecture seule.
 * Philosophie : le dashboard ne peut recevoir que des projections redacted du
 * Core local. Le test n’envoie ni ne trace de token réel ; les réponses sont
 * simulées en mémoire par le navigateur afin de contrôler le contrat d’UI.
 */
import { expect, test, type Page, type Request } from "@playwright/test";

const dashboardURL = process.env.FORGELOCAL_DASHBOARD_URL ?? "http://127.0.0.1:3000";
const coreBaseURL = "http://127.0.0.1:19280";
const validCode = "a".repeat(64);
const sentinels = ["t06-sentinel", "/t06/private", "T06-RUNTIME-HASH"];

type CatalogOptions = {
  groups?: unknown[];
  runtimes?: unknown[];
  failPath?: "/api/v1/readonly/groups" | "/api/v1/readonly/runtimes";
};

function response(data: unknown) {
  return JSON.stringify({ api_version: "v1", data, page: { limit: 100 } });
}

async function mockReadOnlyCore(page: Page, options: CatalogOptions = {}) {
  const observed: Request[] = [];
  const groups = options.groups ?? [{
    id: "grp_compliance",
    name: "Conformité locale",
    proxy_mode: "direct",
    proxy_configured: false,
    profile_count: 2,
    created_at: "2026-08-15T12:00:00Z",
    updated_at: "2026-08-15T12:00:00Z",
  }];
  const runtimes = options.runtimes ?? [{
    id: "runtime_camoufox_candidate",
    display_name: "Camoufox",
    version: "131.0",
    architecture: "amd64",
    status: "candidate",
    enabled: false,
    platform_supported: true,
    candidate: true,
    launchable: false,
  }];

  const readonlyHandler = async (route: import("@playwright/test").Route) => {
    const request = route.request();
    observed.push(request);
    const path = new URL(request.url()).pathname;

    if (path === "/api/v1/readonly/session/bootstrap") {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          token: "test-token-kept-in-memory-only",
          expires_at: "2099-08-15T12:00:00Z",
          scope: "readonly",
        }),
      });
      return;
    }

    if (path === options.failPath) {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ error: { code: "UNAUTHORIZED" } }),
      });
      return;
    }

    if (path === "/api/v1/readonly/summary") {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ api_version: "v1", data: { profiles: 0, groups: groups.length, runtimes: runtimes.length } }),
      });
      return;
    }
    if (path === "/api/v1/readonly/profiles") {
      await route.fulfill({ status: 200, contentType: "application/json", body: response([]) });
      return;
    }
    if (path === "/api/v1/readonly/groups") {
      await route.fulfill({ status: 200, contentType: "application/json", body: response(groups) });
      return;
    }
    if (path === "/api/v1/readonly/runtimes") {
      await route.fulfill({ status: 200, contentType: "application/json", body: response(runtimes) });
      return;
    }

    await route.fulfill({ status: 404, contentType: "application/json", body: JSON.stringify({ error: { code: "NOT_FOUND" } }) });
  };

  for (const host of ["127.0.0.1", "localhost"]) {
    await page.route(`http://${host}:19280/api/v1/readonly/**`, readonlyHandler);
  }

  return observed;
}

async function connect(page: Page) {
  await page.goto(dashboardURL, { waitUntil: "networkidle" });
  await page.getByLabel("Code local à usage unique").fill(validCode);
  await page.getByRole("button", { name: "Relier au Core local" }).click();
}

function assertNoMutationRequests(observed: Request[]) {
  const allowedPaths = new Set([
    "/api/v1/readonly/session/bootstrap",
    "/api/v1/readonly/summary",
    "/api/v1/readonly/profiles",
    "/api/v1/readonly/groups",
    "/api/v1/readonly/runtimes",
  ]);
  expect(observed.length).toBeGreaterThanOrEqual(5);
  for (const request of observed) {
    const path = new URL(request.url()).pathname;
    expect(allowedPaths.has(path)).toBeTruthy();
    expect(request.method() === "GET" || (request.method() === "POST" && path === "/api/v1/readonly/session/bootstrap")).toBeTruthy();
  }
}

test.describe.configure({ mode: "serial" });

test("T06-3 — affiche les catalogues Core redacted sans mutation ni sentinelle", async ({ page }) => {
  const observed = await mockReadOnlyCore(page);
  await connect(page);

  await expect(page.getByText("Lecture Core sécurisée")).toBeVisible();
  await expect(page.getByTestId("core-groups-list")).toContainText("Conformité locale");
  await expect(page.getByTestId("core-groups-list")).toContainText("2 profils");
  await expect(page.getByTestId("core-runtimes-list")).toContainText("Camoufox");
  await expect(page.getByTestId("core-runtimes-list")).toContainText("candidat non lançable");
  await expect(page.getByTestId("core-groups-count")).toHaveText("1");
  await expect(page.getByTestId("core-runtimes-count")).toHaveText("1");
  await expect(page.getByRole("button", { name: "Lancer" })).toBeDisabled();
  await expect(page.getByRole("button", { name: "Isoler" })).toBeDisabled();

  const bodyText = await page.locator("body").innerText();
  for (const sentinel of sentinels) expect(bodyText).not.toContain(sentinel);
  assertNoMutationRequests(observed);
  process.stdout.write("T06_CATALOGS: PASS groups=1 runtimes=1 sentinels=absent mutations=0\n");
});

test("T06-3 — rend les inventaires vides explicitement", async ({ page }) => {
  await mockReadOnlyCore(page, { groups: [], runtimes: [] });
  await connect(page);

  await expect(page.getByTestId("core-groups-empty")).toBeVisible();
  await expect(page.getByTestId("core-groups-empty")).toContainText("Aucun groupe redacted");
  await expect(page.getByTestId("core-runtimes-empty")).toBeVisible();
  await expect(page.getByTestId("core-runtimes-empty")).toContainText("Aucun runtime redacted");
  await expect(page.getByTestId("core-groups-count")).toHaveText("0");
  await expect(page.getByTestId("core-runtimes-count")).toHaveText("0");
  process.stdout.write("T06_EMPTY_CATALOGS: PASS groups=0 runtimes=0\n");
});

test("T06-3 — invalide la session mémoire après un 401 pendant le chargement", async ({ page }) => {
  await mockReadOnlyCore(page, { failPath: "/api/v1/readonly/groups" });
  await connect(page);

  await expect(page.getByText("Connexion impossible. Vérifiez que ce dashboard est servi par le Core local.")).toBeVisible();
  await expect(page.getByText("Lecture Core sécurisée")).toHaveCount(0);
  await expect(page.getByText("Core non connecté", { exact: true })).toBeVisible();
  await expect(page.getByTestId("core-groups-panel")).toContainText("Inventaire en attente");
  await expect(page.getByTestId("core-runtimes-panel")).toContainText("Inventaire en attente");
  process.stdout.write("T06_FORCED_401: PASS token=cleared ui=Core_non_connecte\n");
});
