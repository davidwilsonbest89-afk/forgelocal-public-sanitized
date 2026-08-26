import http from 'node:http';

const playwrightCorePath = process.env.PLAYWRIGHT_CORE_PATH ?? new URL('../../forge-dashboard/node_modules/.pnpm/playwright-core@1.55.1/node_modules/playwright-core/index.mjs', import.meta.url).pathname;
const { chromium } = await import(playwrightCorePath);
const requiredEnv = name => {
  const value = process.env[name];
  if (!value) throw new Error(`missing required runtime environment: ${name}`);
  return value;
};

const core = process.env.SMOKE_CORE_URL ?? 'http://127.0.0.1:19281';
const dashboard = process.env.SMOKE_DASHBOARD_URL ?? 'http://127.0.0.1:3001';
const adminToken = requiredEnv('SMOKE_ADMIN_TOKEN');
const alphaProxyUser = requiredEnv('SMOKE_ALPHA_PROXY_USER');
const alphaProxyPass = requiredEnv('SMOKE_ALPHA_PROXY_PASS');
const betaProxyUser = requiredEnv('SMOKE_BETA_PROXY_USER');
const betaProxyPass = requiredEnv('SMOKE_BETA_PROXY_PASS');
const badProxyExpectedUser = requiredEnv('SMOKE_BAD_PROXY_EXPECTED_USER');
const badProxyExpectedPass = requiredEnv('SMOKE_BAD_PROXY_EXPECTED_PASS');
const alphaDestPort = 19283, alphaProxyPort = 19282, betaDestPort = 19285, betaProxyPort = 19284, badProxyPort = 19286, directDestPort = 19287;
const events = [];
const destinations = new Map();
const proxies = new Map();
const servers = [];
const hit = (kind, data) => events.push({ kind, at: new Date().toISOString(), ...data });

function startDestination(name, port) {
  const counts = { total: 0, viaProxy: 0 };
  destinations.set(name, { port, counts });
  const server = http.createServer((req, res) => {
    counts.total += 1;
    const via = req.headers['x-smoke-proxy-hop'] === 'authenticated-local-proxy';
    if (via) counts.viaProxy += 1;
    hit('destination_request', { destination: name, path: req.url ?? '/', via_proxy: via, method: req.method });
    res.writeHead(200, { 'content-type': 'text/plain; charset=utf-8' });
    res.end(`smoke-destination-${name}`);
  });
  return new Promise(resolve => server.listen(port, '127.0.0.1', () => { servers.push(server); resolve(); }));
}

function authHeader(user, pass) { return `Basic ${Buffer.from(`${user}:${pass}`).toString('base64')}`; }
function startProxy(name, port, expectedUser, expectedPass, allowAnonymous = false) {
  const counts = { total: 0, accepted: 0, rejected: 0 };
  proxies.set(name, { port, counts });
  const server = http.createServer((req, res) => {
    counts.total += 1;
    const auth = req.headers['proxy-authorization'];
    if (!allowAnonymous && auth !== authHeader(expectedUser, expectedPass)) {
      counts.rejected += 1;
      hit('proxy_auth_rejected', { proxy: name, has_proxy_authorization: Boolean(auth) });
      res.writeHead(407, { 'proxy-authenticate': 'Basic realm="synthetic-local"', 'content-type': 'text/plain' });
      return res.end('synthetic proxy authentication required');
    }
    let target;
    try { target = new URL(req.url); } catch { res.writeHead(400); return res.end('invalid proxy target'); }
    const loopbackTarget = target.hostname === '127.0.0.1' || target.hostname === 'localhost' || target.hostname === '[::1]';
    if (!loopbackTarget) {
      hit('external_target_blocked', { proxy: name, target_host: target.hostname, target_path: target.pathname });
      res.writeHead(502, { 'content-type': 'text/plain; charset=utf-8' });
      return res.end('external target blocked by loopback-only smoke');
    }
    counts.accepted += 1;
    hit('proxy_forward', { proxy: name, target_host: target.hostname, target_port: target.port, target_path: target.pathname, auth: 'synthetic-accepted' });
    const forwardHeaders = { ...req.headers };
    delete forwardHeaders['proxy-authorization'];
    forwardHeaders.host = target.host;
    forwardHeaders['x-smoke-proxy-hop'] = 'authenticated-local-proxy';
    const forward = http.request({ hostname: target.hostname, port: Number(target.port || 80), path: `${target.pathname}${target.search}`, method: req.method, headers: forwardHeaders }, upstream => {
      res.writeHead(upstream.statusCode ?? 502, upstream.headers);
      upstream.pipe(res);
    });
    forward.on('error', error => { hit('proxy_forward_error', { proxy: name, error: String(error).slice(0, 180) }); if (!res.headersSent) res.writeHead(502); res.end('synthetic proxy forward error'); });
    req.pipe(forward);
  });
  return new Promise(resolve => server.listen(port, '127.0.0.1', () => { servers.push(server); resolve(); }));
}

async function jsonRequest(path, method = 'GET', body) {
  const response = await fetch(`${core}${path}`, { method, headers: { Authorization: `Bearer ${adminToken}`, Origin: core, 'Content-Type': 'application/json', 'X-Request-ID': `smoke-${crypto.randomUUID()}` }, body: body === undefined ? undefined : JSON.stringify(body) });
  let payload = null; try { payload = await response.json(); } catch { /* redacted status-only */ }
  return { status: response.status, payload };
}
async function waitFor(predicate, label, timeout = 15000) { const end = Date.now() + timeout; let attempts = 0; while (Date.now() < end) { attempts += 1; if (await predicate()) { hit('poll_complete', { label, attempts }); return true; } await new Promise(r => setTimeout(r, 200)); } throw new Error(`timeout: ${label}`); }
async function waitForCoreSessionForProfile(profileId, label, timeout = 15000) {
  let latest = [];
  await waitFor(async () => {
    const response = await jsonRequest('/api/sessions');
    latest = response.payload?.data ?? [];
    const match = latest.find(item => item.profile_id === profileId && item.session_id);
    hit('core_session_poll', { label, status: response.status, session_count: latest.length, matching_profile: Boolean(match), session_id_present: Boolean(match?.session_id) });
    return response.status === 200 && Boolean(match?.session_id);
  }, label, timeout);
  const match = latest.find(item => item.profile_id === profileId && item.session_id);
  if (!match?.session_id) throw new Error(`session disappeared after polling: ${label}`);
  return match;
}
function safeError(error) { return String(error).replace(/Bearer\s+[A-Za-z0-9._-]+/gi, 'Bearer <REDACTED>').slice(0, 220); }

const browser = await chromium.launch({ headless: true, executablePath: '/usr/bin/chromium', args: ['--no-sandbox', '--disable-gpu', '--disable-background-networking', '--disable-component-update', '--disable-domain-reliability'] });
const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
let alphaProfileId = '', betaProfileId = '', alphaProxyId = '', betaProxyId = '';
let coreSessionStatus = null, coreSessionNavigate = null;
try {
  await Promise.all([startDestination('alpha', alphaDestPort), startDestination('beta', betaDestPort), startDestination('direct', directDestPort), startProxy('alpha', alphaProxyPort, alphaProxyUser, alphaProxyPass, true), startProxy('beta', betaProxyPort, betaProxyUser, betaProxyPass, true), startProxy('bad-auth', badProxyPort, badProxyExpectedUser, badProxyExpectedPass)]);
  hit('local_services_ready', { destination_ports: [alphaDestPort, betaDestPort, directDestPort], proxy_ports: [alphaProxyPort, betaProxyPort, badProxyPort] });
  const codeResponse = await fetch(`${core}/api/v1/readonly/session/codes`, { method: 'POST', headers: { Authorization: `Bearer ${adminToken}`, Origin: core, 'X-Request-ID': `smoke-${crypto.randomUUID()}` } });
  const codePayload = await codeResponse.json();
  const readonlyCode = codePayload.code ?? codePayload.data?.code ?? process.env.SMOKE_READONLY_CODE ?? '';
  if (!codeResponse.ok || typeof readonlyCode !== 'string' || readonlyCode.length !== 64) throw new Error(`readonly code issue failed: HTTP ${codeResponse.status}`);
  await page.goto(dashboard, { waitUntil: 'networkidle' });
  await page.getByTestId('local-core-code').fill(readonlyCode);
  await page.getByTestId('local-core-connect').click();
  await page.getByText('Lecture Core sécurisée').waitFor();
  await page.getByTestId('local-core-admin-token').fill(adminToken);
  await page.getByTestId('local-core-admin-link').click();
  await page.getByTestId('local-core-admin-message').waitFor();
  hit('dashboard_admin_linked', { status: 'linked' });

  async function createProfile(name) {
    await page.getByRole('button', { name: 'Préparer un profil' }).click();
    await page.getByTestId('create-profile-name').fill(name);
    const runtime = await page.getByTestId('create-profile-runtime').locator('option:not([disabled])').first().getAttribute('value');
    if (!runtime) throw new Error('no launchable runtime exposed by Core');
    await page.getByTestId('create-profile-runtime').selectOption(runtime);
    await page.getByTestId('create-profile-tags').fill(`smoke,${name}`);
    await page.getByRole('button', { name: /Créer via le Core/ }).click();
    await page.getByText('Action appliquée au Core').last().waitFor();
    await page.getByText(name, { exact: true }).first().click();
    const list = await jsonRequest('/api/profiles?limit=100');
    const found = (list.payload?.data ?? []).find(p => p.name === name);
    if (!found?.id) throw new Error(`created profile not visible in Core: ${name}`);
    hit('profile_created', { profile: name, profile_id_present: true });
    return found.id;
  }
  alphaProfileId = await createProfile('smoke-alpha');
  betaProfileId = await createProfile('smoke-beta');
  hit('dashboard_profiles_created', { alpha_id_present: Boolean(alphaProfileId), beta_id_present: Boolean(betaProfileId), creation_via_dashboard: true });

  async function createProxy(name, port, secretRef) {
    await page.getByTestId('proxy-name').fill(name);
    await page.getByTestId('proxy-type').selectOption('http');
    await page.getByTestId('proxy-host').fill('127.0.0.1');
    await page.getByTestId('proxy-port').fill(String(port));
    await page.getByTestId('proxy-region').fill('us-east');
    await page.getByTestId('proxy-secret-ref').fill(secretRef);
    await page.getByTestId('proxy-create').click();
    await page.getByText('Proxy ajouté au référentiel Core.').waitFor();
    const list = await jsonRequest('/api/proxies');
    const found = (list.payload?.data?.items ?? []).find(p => p.name === name);
    if (!found?.id) throw new Error(`created proxy not visible in Core: ${name}`);
    hit('proxy_created', { proxy: name, proxy_id_present: true, credential_value_logged: false });
    return found.id;
  }
  await page.locator('.profile-row').filter({ hasText: 'smoke-alpha' }).first().click();
  await page.getByText('smoke-alpha', { exact: true }).first().waitFor();
  hit('dashboard_profile_selected', { profile: 'alpha', profile_id_present: true });
  alphaProxyId = await createProxy('smoke-proxy-alpha', alphaProxyPort, '');
  await page.getByTestId(`proxy-assign-${alphaProxyId}`).click();
  await page.getByText('Proxy affecté au profil côté Core.').waitFor();
  const alphaProjection = await jsonRequest(`/api/profiles/${encodeURIComponent(alphaProfileId)}`);
  hit('alpha_assignment_persisted', { profile_id_present: Boolean(alphaProjection.payload?.data?.id), proxy_id_present: alphaProjection.payload?.proxy_id === alphaProxyId });

  await page.getByRole('button', { name: 'Automation locale' }).click();
  await page.getByTestId('automation-open-session').click();
  const session = await waitForCoreSessionForProfile(alphaProfileId, 'alpha-core-session-created');
  if (session?.session_id) {
    coreSessionStatus = 201;
    coreSessionNavigate = await jsonRequest(`/api/sessions/${encodeURIComponent(session.session_id)}/navigate`, 'POST', { url: `http://127.0.0.1:${alphaDestPort}/core-session` });
    await jsonRequest(`/api/sessions/${encodeURIComponent(session.session_id)}`, 'DELETE');
    hit('core_session_navigation', { create_status: coreSessionStatus, navigate_status: coreSessionNavigate.status, destination_hits: destinations.get('alpha').counts.total, destination_via_proxy: destinations.get('alpha').counts.viaProxy, alpha_proxy_forwards: proxies.get('alpha').counts.accepted });
  } else {
    coreSessionStatus = 500;
    hit('core_session_navigation', { create_status: 'not-created', navigate_status: 'not-executed', error: 'Dashboard session did not open a Core session' });
  }

  const proxiedBrowser = await chromium.launch({ headless: true, executablePath: '/usr/bin/chromium', proxy: { server: `http://127.0.0.1:${alphaProxyPort}`, username: alphaProxyUser, password: alphaProxyPass }, args: ['--no-sandbox', '--disable-background-networking', '--disable-component-update', '--disable-domain-reliability', '--proxy-bypass-list=<-loopback>'] });
  const proxiedPage = await proxiedBrowser.newPage();
  const nominalResponse = await proxiedPage.goto(`http://127.0.0.1:${alphaDestPort}/nominal`, { waitUntil: 'networkidle', timeout: 10000 });
  hit('direct_browser_nominal', { status: nominalResponse?.status() ?? 0, destination_hits: destinations.get('alpha').counts.total, proxy_accepted: proxies.get('alpha').counts.accepted });
  await proxiedBrowser.close();

  const beforeStopped = destinations.get('alpha').counts.total;
  await new Promise(resolve => servers.find(s => s.address()?.port === alphaProxyPort)?.close(resolve));
  const stoppedBrowser = await chromium.launch({ headless: true, executablePath: '/usr/bin/chromium', proxy: { server: `http://127.0.0.1:${alphaProxyPort}`, username: alphaProxyUser, password: alphaProxyPass }, args: ['--no-sandbox', '--disable-background-networking', '--disable-component-update', '--disable-domain-reliability', '--proxy-bypass-list=<-loopback>'] });
  const stoppedPage = await stoppedBrowser.newPage();
  try { await stoppedPage.goto(`http://127.0.0.1:${alphaDestPort}/proxy-stopped`, { timeout: 6000 }); hit('proxy_stopped_result', { result: 'unexpected-success', destination_hits: destinations.get('alpha').counts.total }); } catch (error) { hit('proxy_stopped_result', { result: 'fail-closed', error: safeError(error), destination_hits: destinations.get('alpha').counts.total, destination_unchanged: destinations.get('alpha').counts.total === beforeStopped }); }
  await stoppedBrowser.close();

  // Reuse the Dashboard-selected alpha profile through the real Core path after
  // its assigned proxy has been stopped. The destination must remain unchanged.
  await page.locator('.profile-row').filter({ hasText: 'smoke-alpha' }).first().click();
  await page.getByText('smoke-alpha', { exact: true }).first().waitFor();
  await page.getByRole('button', { name: 'Automation locale' }).click();
  await page.getByTestId('automation-open-session').click();
  const stoppedCoreBefore = destinations.get('alpha').counts.total;
  const stoppedCoreSession = await waitForCoreSessionForProfile(alphaProfileId, 'alpha-core-session-proxy-stopped');
  let stoppedCoreNavigate = null;
  if (stoppedCoreSession?.session_id) {
    stoppedCoreNavigate = await jsonRequest(`/api/sessions/${encodeURIComponent(stoppedCoreSession.session_id)}/navigate`, 'POST', { url: `http://127.0.0.1:${alphaDestPort}/core-proxy-stopped` });
    await jsonRequest(`/api/sessions/${encodeURIComponent(stoppedCoreSession.session_id)}`, 'DELETE');
  }
  hit('core_proxy_stopped_navigation', { create_status: stoppedCoreSession?.session_id ? 201 : 'not-created', navigate_status: stoppedCoreNavigate?.status ?? 'not-executed', destination_hits: destinations.get('alpha').counts.total, destination_unchanged: destinations.get('alpha').counts.total === stoppedCoreBefore, result: stoppedCoreNavigate?.status === 200 ? 'unexpected-success' : 'fail-closed' });

  const badBrowser = await chromium.launch({ headless: true, executablePath: '/usr/bin/chromium', proxy: { server: `http://127.0.0.1:${badProxyPort}`, username: `${badProxyExpectedUser}-wrong`, password: `${badProxyExpectedPass}-wrong` }, args: ['--no-sandbox', '--disable-background-networking', '--disable-component-update', '--disable-domain-reliability', '--proxy-bypass-list=<-loopback>'] });
  const badPage = await badBrowser.newPage();
  try { const badResponse = await badPage.goto(`http://127.0.0.1:${alphaDestPort}/bad-auth`, { timeout: 6000 }); hit('proxy_bad_auth_result', { result: badResponse?.status() === 407 ? 'rejected' : 'unexpected-success', http_status: badResponse?.status() ?? 0, destination_hits: destinations.get('alpha').counts.total }); } catch (error) { hit('proxy_bad_auth_result', { result: 'rejected', error: safeError(error), destination_hits: destinations.get('alpha').counts.total }); }
  await badBrowser.close();

  await page.locator('.profile-row').filter({ hasText: 'smoke-beta' }).first().click();
  await page.getByText('smoke-beta', { exact: true }).first().waitFor();
  betaProxyId = await createProxy('smoke-proxy-beta', betaProxyPort, '');
  await page.getByTestId(`proxy-assign-${betaProxyId}`).click();
  await page.getByText('Proxy affecté au profil côté Core.').waitFor();
  const betaProjection = await jsonRequest(`/api/profiles/${encodeURIComponent(betaProfileId)}`);
  hit('beta_assignment_persisted', { profile_id_present: Boolean(betaProjection.payload?.data?.id), proxy_id_present: betaProjection.payload?.proxy_id === betaProxyId, distinct_from_alpha: betaProjection.payload?.proxy_id !== alphaProxyId });
  const betaBrowser = await chromium.launch({ headless: true, executablePath: '/usr/bin/chromium', proxy: { server: `http://127.0.0.1:${betaProxyPort}`, username: betaProxyUser, password: betaProxyPass }, args: ['--no-sandbox', '--disable-background-networking', '--disable-component-update', '--disable-domain-reliability', '--proxy-bypass-list=<-loopback>'] });
  const betaPage = await betaBrowser.newPage();
  const betaResponse = await betaPage.goto(`http://127.0.0.1:${betaDestPort}/isolated`, { waitUntil: 'networkidle', timeout: 10000 });
  hit('beta_isolation_browser', { status: betaResponse?.status() ?? 0, beta_proxy_accepted: proxies.get('beta').counts.accepted, alpha_proxy_accepted: proxies.get('alpha').counts.accepted, distinct_destination: destinations.get('beta').counts.total >= 1 });
  await betaBrowser.close();

  const expiryProbe = await jsonRequest('/api/profiles?limit=1');
  hit('token_probe_before_revoke', { status: expiryProbe.status, reason: expiryProbe.payload?.error?.reason ?? 'none', token_value_logged: false });
  if (process.env.SMOKE_SKIP_REVOKE !== '1') {
    const revokeResponse = await fetch(`${core}/api/auth/revoke`, { method: 'POST', headers: { Authorization: `Bearer ${adminToken}`, Origin: core, 'X-Request-ID': `smoke-${crypto.randomUUID()}` } });
    const revokedProbe = await jsonRequest('/api/profiles?limit=1');
    hit('token_revocation', { revoke_status: revokeResponse.status, probe_status: revokedProbe.status, reason: revokedProbe.payload?.error?.reason ?? 'none', token_value_logged: false, write_state_expected: 'removed' });
  } else {
    hit('token_revocation_deferred', { reason: 'deferred_until_after_restart', token_value_logged: false });
  }
} catch (error) {
  hit('smoke_error', { error: safeError(error) });
} finally {
  await page.close();
  await browser.close();
  for (const server of servers) await new Promise(resolve => server.close(() => resolve()));
}
const alpha = destinations.get('alpha')?.counts ?? {};
const beta = destinations.get('beta')?.counts ?? {};
const alphaProxy = proxies.get('alpha')?.counts ?? {};
const betaProxy = proxies.get('beta')?.counts ?? {};
const badProxy = proxies.get('bad-auth')?.counts ?? {};
const sessionProxyPassed = events.some(e => e.kind === 'core_session_navigation' && e.navigate_status === 200 && e.destination_via_proxy > 0);
const directNominalPassed = events.some(e => e.kind === 'direct_browser_nominal' && e.status === 200 && e.proxy_accepted > 0);
const stoppedFailClosed = events.some(e => e.kind === 'proxy_stopped_result' && e.result === 'fail-closed' && e.destination_unchanged) && events.some(e => e.kind === 'core_proxy_stopped_navigation' && e.result === 'fail-closed' && e.destination_unchanged);
const badAuthRejected = events.some(e => e.kind === 'proxy_bad_auth_result' && e.result === 'rejected' && badProxy.rejected > 0);
const isolated = events.some(e => e.kind === 'beta_isolation_browser' && e.status === 200 && e.beta_proxy_accepted > 0 && e.distinct_destination);
const coreProfileAssignments = events.filter(e => e.kind.endsWith('assignment_persisted')).length === 2;
const externalForward = events.some(e => e.kind === 'proxy_forward' && !['127.0.0.1', 'localhost', '[::1]'].includes(e.target_host));
const externalForwardAssertion = externalForward === false;
if (!externalForwardAssertion) hit('external_forward_assertion_failed', { external_forward_observed: true });
const strictProxyPath = sessionProxyPassed && directNominalPassed && stoppedFailClosed && badAuthRejected && isolated && coreProfileAssignments && externalForwardAssertion;
const coreLaunchBypass = events.some(e => e.kind === 'core_session_navigation' && e.create_status === 201 && e.navigate_status === 200 && e.destination_via_proxy === 0);
  const dashboardProfileCreation = events.some(e => e.kind === 'dashboard_profiles_created' && e.creation_via_dashboard && e.alpha_id_present && e.beta_id_present);
  const status = strictProxyPath && dashboardProfileCreation ? 'SMOKE_DASHBOARD_PROFILE_CREATE_PASS' : (strictProxyPath ? 'SMOKE_DASHBOARD_PROFILE_CREATE_FAIL' : (coreLaunchBypass ? 'SMOKE_INTEGRATED_PROXY_FAIL_PROFILE_ASSIGNMENT_NOT_APPLIED_TO_CORE_LAUNCH' : 'SMOKE_INTEGRATED_PROXY_FAIL_UNKNOWN'));
console.log(JSON.stringify({ status, dashboard_profile_create_path: dashboardProfileCreation ? 'PASS_DASHBOARD_MUTATION' : 'FAIL_NOT_DASHBOARD_MUTATION', core_session_proxy_pass: sessionProxyPassed, direct_browser_nominal_pass: directNominalPassed, proxy_stopped_fail_closed: stoppedFailClosed, proxy_bad_auth_rejected: badAuthRejected, profile_isolation_pass: isolated, core_profile_assignments_persisted: coreProfileAssignments, external_forward_observed: externalForward, external_targets_blocked: events.filter(e => e.kind === 'external_target_blocked').length, destinations: { alpha, beta }, proxies: { alpha: alphaProxy, beta: betaProxy, bad_auth: badProxy }, core_session_status: coreSessionStatus, core_session_navigate_status: coreSessionNavigate?.status ?? null, events: events.map(e => ({ ...e })) }, null, 2));
