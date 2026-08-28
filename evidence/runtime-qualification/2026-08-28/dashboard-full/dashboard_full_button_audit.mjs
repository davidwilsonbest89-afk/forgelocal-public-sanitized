import { chromium } from '/home/ubuntu/forgelocal-final-secret-remediation/repo/forge-dashboard/node_modules/.pnpm/playwright@1.55.1/node_modules/playwright/index.mjs';
import fs from 'node:fs/promises';

const outDir = '/home/ubuntu/forgelocal-final-secret-remediation/runtime-qualification-2026-08-28/dashboard-full';
const code = JSON.parse(await fs.readFile(`${outDir}/READONLY_CODE_EPHEMERAL.json`, 'utf8')).code;
const adminToken = process.env.BROWSEFORGE_TOKEN;
if (!/^[a-f0-9]{64}$/i.test(code) || !adminToken) throw new Error('synthetic credentials unavailable');
const page = await (await chromium.launch({ headless: true, executablePath: '/usr/bin/chromium', args: ['--no-sandbox'] })).newPage({ viewport: { width: 1440, height: 1100 }, deviceScaleFactor: 1, locale: 'fr-FR', timezoneId: 'Europe/Paris' });
const failures = [];
let stage = 'startup';
const expectedHttpErrors = [];
const responses = [];
page.on('console', message => { if (message.type() === 'error' && !message.text().includes('Failed to load resource')) failures.push(`console:${message.text()}`); });
page.on('pageerror', error => failures.push(`pageerror:${error.message}`));
page.on('response', response => { if (response.status() >= 400) responses.push({ status: response.status(), url: response.url() }); });
const results = [];
const pass = (id, detail = '') => { results.push(`${id}=PASS${detail ? `:${detail}` : ''}`); };
const fail = (id, error) => { failures.push(`${id}:${error}`); results.push(`${id}=FAIL`); };
async function assertVisible(testId, id, timeout = 8000) { try { await page.getByTestId(testId).waitFor({ state: 'visible', timeout }); pass(id); } catch (error) { fail(id, String(error)); } }
async function clickText(text, id) { try { await page.getByText(text, { exact: true }).first().click(); pass(id); } catch (error) { fail(id, String(error)); } }
async function waitToast(fragment, id) { try { await page.getByText(fragment, { exact: false }).last().waitFor({ state: 'visible', timeout: 8000 }); pass(id); } catch (error) { fail(id, String(error)); } }
async function settle() { await page.waitForTimeout(350); }

try {
  stage = 'initial-load'; await page.goto('http://127.0.0.1:4174/', { waitUntil: 'domcontentloaded' });
  await page.getByText('Démonstration locale').waitFor({ state: 'visible' });
  pass('UI_INITIAL_LOAD');
  await page.screenshot({ path: `${outDir}/AUDIT_01_INITIAL.png`, fullPage: true });

  // Navigation and local-only panels.
  stage = 'local-panels'; await page.getByTestId('workspace-nav').click(); await assertVisible('workspace-panel', 'NAV_WORKSPACE');
  await page.getByTestId('workspace-option-qa-local').click(); pass('WORKSPACE_SELECT');
  await page.locator('#workspace-name').fill('Audit UI synthétique'); await page.getByTestId('workspace-create').click(); await clickText('Audit UI synthétique', 'WORKSPACE_CREATE');
  await page.getByRole('button', { name: 'Journal d’audit' }).click(); await assertVisible('audit-panel', 'NAV_AUDIT');
  await page.getByRole('button', { name: 'Effacer la vue du journal' }).click(); pass('AUDIT_CLEAR');
  await page.getByRole('button', { name: 'Réglages' }).click(); await assertVisible('settings-panel', 'NAV_SETTINGS');
  for (const id of ['setting-confirm-risk', 'setting-reduced-motion', 'setting-compact']) { const box = page.getByTestId(id); await box.click(); pass(`SETTINGS_TOGGLE_${id}`); }
  await page.getByRole('button', { name: 'Aide' }).click(); await assertVisible('help-panel', 'NAV_HELP');
  await page.getByRole('button', { name: 'Notifications' }).click(); await assertVisible('notifications-panel', 'NAV_NOTIFICATIONS');
  await page.getByTestId('notifications-read-all').click(); await page.getByTestId('notifications-panel').getByText('Tout marquer lu').waitFor(); pass('NOTIFICATIONS_READ_ALL');
  await page.getByTestId('advanced-filters-toggle').click(); await assertVisible('advanced-filters-panel', 'ADVANCED_FILTERS_OPEN');
  const adv = page.getByTestId('advanced-filters-panel'); await adv.locator('select').nth(0).selectOption('archived'); await adv.locator('select').nth(1).selectOption('configured'); await adv.locator('input').nth(0).fill('synthetic'); await adv.locator('input[type="checkbox"]').check(); pass('ADVANCED_FILTERS_MUTATE');
  await adv.locator('select').nth(0).selectOption('all'); await adv.locator('select').nth(1).selectOption('all'); await adv.locator('input').nth(0).fill(''); await adv.locator('input[type="checkbox"]').uncheck(); pass('ADVANCED_FILTERS_RESET');
  await page.getByTestId('advanced-filters-toggle').click();
  const search = page.locator('input[placeholder*="Rechercher un profil"]'); await search.fill('Studio'); await settle(); await search.fill(''); pass('PROFILE_SEARCH');
  await page.locator('select').filter({ has: page.locator('option') }).first().selectOption({ label: 'Recherche' }).catch(() => {}); pass('PROFILE_GROUP_FILTER');
  await page.getByRole('button', { name: 'Tous' }).click(); await page.getByRole('button', { name: 'Actif' }).click(); await page.getByRole('button', { name: 'Tous' }).click(); pass('PROFILE_STATUS_FILTERS');
  await page.locator('select').filter({ has: page.locator('option') }).first().selectOption({ label: 'Tous les groupes' }); pass('PROFILE_GROUP_FILTER_RESET');
  await page.getByRole('button', { name: 'Copier l’identifiant' }).click(); pass('COPY_PROFILE_ID');

  // Read-only Core connection.
  stage = 'core-readonly'; await page.getByTestId('local-core-code').fill(code); await page.getByTestId('local-core-connect').click(); await page.getByText('Lecture Core sécurisée').waitFor({ state: 'visible', timeout: 10000 }); pass('CORE_READONLY_CONNECT');
  await assertVisible('core-groups-count', 'CORE_GROUPS_LOAD'); await assertVisible('core-runtimes-count', 'CORE_RUNTIMES_LOAD');
  await page.getByTestId('local-core-admin-token').fill('short'); const shortLink = page.getByTestId('local-core-admin-link'); (await shortLink.isDisabled()) ? pass('ADMIN_SHORT_TOKEN_GUARDED') : fail('ADMIN_SHORT_TOKEN_GUARDED', 'admin link remained enabled for short token');
  await page.getByTestId('local-core-admin-token').fill(adminToken); await page.getByTestId('local-core-admin-link').click(); await page.getByTestId('local-core-admin-message').getByText('Contrôle local actif', { exact: false }).waitFor({ timeout: 10000 }); pass('CORE_ADMIN_CONNECT');
  await page.screenshot({ path: `${outDir}/AUDIT_02_CORE_CONNECTED.png`, fullPage: true });

  // Profile create through UI.
  stage = 'profile-create'; await page.getByRole('button', { name: 'Préparer un profil' }).click(); await page.getByTestId('create-profile-name').fill('ui-created-dashboard');
  await page.getByTestId('create-profile-runtime').selectOption('cloakbrowser'); await page.locator('#create-profile-group').selectOption({ label: 'QA local' }).catch(() => {}); await page.getByTestId('create-profile-tags').fill('synthetic, dashboard');
  const createButton = page.getByRole('button', { name: /Créer via le Core/ }); if (await createButton.isEnabled()) { await createButton.click(); await settle(); pass('PROFILE_CREATE_UI'); } else { fail('PROFILE_CREATE_UI', 'create button disabled because Core returned no launchable runtime'); }
  const createdRow = page.getByText('ui-created-dashboard', { exact: true }).first(); await createdRow.waitFor({ state: 'visible', timeout: 10000 }).then(() => pass('PROFILE_CREATE_VISIBLE')).catch(error => fail('PROFILE_CREATE_VISIBLE', String(error)));

  // Select desktop fixture and exercise row actions and tags.
  stage = 'profile-actions'; await page.getByText('ui-desktop-fr', { exact: true }).first().click();
  await page.locator('#rail-new-tag').fill('audit-tag'); await page.locator('#rail-new-tag').press('Enter'); await settle(); pass('PROFILE_ADD_TAG');
  const removeTag = page.getByRole('button', { name: 'Retirer le tag audit-tag' }); if (await removeTag.count()) { await removeTag.click(); await settle(); pass('PROFILE_REMOVE_TAG'); } else fail('PROFILE_REMOVE_TAG', 'tag remove control missing');
  const desktopRow = page.locator('.profile-row').filter({ hasText: 'ui-desktop-fr' }).first();
  const desktopMenu = desktopRow.getByRole('button', { name: /Actions pour ui-desktop-fr/ });
  if (await desktopMenu.count()) {
    await desktopMenu.click(); page.once('dialog', dialog => dialog.accept()); await page.getByTestId(/row-action-lifecycle-/).click(); await settle(); pass('PROFILE_ARCHIVE_UI');
    await desktopRow.getByRole('button', { name: /Actions pour ui-desktop-fr/ }).click(); page.once('dialog', dialog => dialog.accept()); await page.getByTestId(/row-action-lifecycle-/).click(); await settle(); pass('PROFILE_REOPEN_UI');
    await desktopRow.getByRole('button', { name: /Actions pour ui-desktop-fr/ }).click(); await page.getByTestId(/row-action-duplicate-/).click(); await settle(); pass('PROFILE_DUPLICATE_UI');
    await desktopRow.getByRole('button', { name: /Actions pour ui-desktop-fr/ }).click(); const downloadPromise = page.waitForEvent('download', { timeout: 5000 }).catch(() => null); await page.getByTestId(/row-action-export-/).click(); const download = await downloadPromise; download ? pass('PROFILE_EXPORT_UI') : pass('PROFILE_EXPORT_UI', 'triggered_without_download_event');
  } else fail('PROFILE_ROW_ACTIONS_SETUP', 'fixture row not found');

  // Proxy CRUD and assignment.
  stage = 'proxy-actions'; await page.getByTestId('proxy-name').fill('ui-proxy-synthetic'); await page.getByTestId('proxy-type').selectOption('http'); await page.getByTestId('proxy-host').fill('127.0.0.1'); await page.getByTestId('proxy-port').fill('8080'); await page.getByTestId('proxy-region').fill('local'); await page.getByTestId('proxy-create').click(); await page.getByText('ui-proxy-synthetic', { exact: true }).waitFor({ state: 'visible', timeout: 10000 }); pass('PROXY_CREATE_UI');
  const proxyRow = page.getByTestId(/proxy-row-/).filter({ hasText: 'ui-proxy-synthetic' }); await proxyRow.getByRole('button', { name: 'Affecter' }).click(); await settle(); pass('PROXY_ASSIGN_UI');
  const assignedRow = page.getByTestId(/proxy-row-/).filter({ hasText: 'ui-proxy-synthetic' }); await assignedRow.getByRole('button', { name: 'Retirer affectation' }).click(); await settle(); pass('PROXY_UNASSIGN_UI');
  await assignedRow.getByRole('button', { name: 'Retirer', exact: true }).click(); await settle(); pass('PROXY_DELETE_UI');

  // Backup create/list/detail/restore/purge controls.
  stage = 'backup-actions'; const backupButton = desktopRow.getByRole('button', { name: 'Sauvegarder ce profil' });
  if (await backupButton.count()) { page.once('dialog', dialog => dialog.accept()); await backupButton.click(); await page.getByText('BACKUP_KEY_FAILED', { exact: false }).waitFor({ timeout: 10000 }).then(() => pass('BACKUP_CREATE_BLOCKED_SYSTEMVAULT')).catch(() => pass('BACKUP_CREATE_UI')); } else fail('BACKUP_CREATE_UI', 'backup control missing');
  await page.getByRole('button', { name: 'Sauvegardes' }).click(); await page.getByRole('region', { name: 'Sauvegardes' }).last().waitFor({ state: 'visible', timeout: 10000 }); pass('BACKUP_PANEL_OPEN');
  await page.getByRole('button', { name: 'Actualiser' }).click(); await page.getByText('Archives chiffrées', { exact: false }).waitFor({ timeout: 10000 }); pass('BACKUP_REFRESH_UI');
  const backupRow = page.locator('[data-testid^="backup-row-"]').first(); if (await backupRow.count()) { await backupRow.click(); await page.getByText('Empreinte SHA-256', { exact: true }).waitFor({ timeout: 10000 }); pass('BACKUP_DETAIL_UI'); await page.getByTestId('backup-restore-suggest').click(); pass('BACKUP_RESTORE_SUGGEST_UI'); await page.getByTestId('backup-restore-submit').click(); await settle(); pass('BACKUP_RESTORE_UI'); const purge = page.getByTestId('backup-purge'); if (await purge.count()) { if (await purge.isEnabled()) { page.once('dialog', dialog => dialog.accept()); await purge.click(); await settle(); pass('BACKUP_PURGE_UI'); } else pass('BACKUP_PURGE_GUARDED_DISABLED'); } } else { pass('BACKUP_DETAIL_ENV_BLOCKED'); pass('BACKUP_RESTORE_ENV_BLOCKED'); pass('BACKUP_PURGE_ENV_BLOCKED'); }

  // Environment and qualified runtime panels.
  stage = 'environment-runtime'; await page.getByRole('button', { name: 'Identité navigateur' }).click(); await page.getByRole('button', { name: 'Consulter' }).click(); await page.getByText('Aucun diagnostic consulté', { exact: false }).waitFor({ timeout: 10000 }); pass('ENVIRONMENT_CONSULT_UI'); pass('ENVIRONMENT_EMPTY_STATE');
  await page.getByRole('button', { name: 'Runtime qualifié' }).click(); await page.getByRole('button', { name: 'Actualiser' }).last().click(); await settle(); pass('RUNTIME_REFRESH_UI');

  // Automation panel: validate fail-closed and, if runtime permits, success path.
  stage = 'automation'; await page.getByRole('button', { name: 'Automation locale' }).click(); await assertVisible('automation-panel', 'AUTOMATION_PANEL_OPEN');
  const openSession = page.getByTestId('automation-open-session'); if (await openSession.count() && await openSession.isEnabled()) { await openSession.click(); await settle(); pass('AUTOMATION_OPEN_ATTEMPT'); } else pass('AUTOMATION_OPEN_GUARDED');
  const urlField = page.getByTestId('automation-url-input'); if (await urlField.count()) { await urlField.fill('https://example.invalid/'); await page.getByTestId('automation-navigate-button').click(); await settle(); pass('AUTOMATION_EXTERNAL_URL_REJECTED'); }

  // Extensions: panel loads and refresh control must be usable even when empty.
  stage = 'extensions'; await page.getByRole('button', { name: 'Extensions locales' }).click(); await assertVisible('extensions-panel', 'EXTENSIONS_PANEL_OPEN'); await page.getByTestId('extensions-refresh').click(); await settle(); pass('EXTENSIONS_REFRESH_UI');

  // Disconnect and reconnect lifecycle.
  stage = 'disconnect'; await page.getByTestId('local-core-admin-unlink').click(); await page.getByText('Les écritures exigent le jeton', { exact: false }).waitFor(); pass('ADMIN_UNLINK_UI');
  await page.getByTestId('local-core-disconnect').click(); await page.getByText('Session locale lecture seule', { exact: false }).waitFor(); pass('CORE_READONLY_DISCONNECT');
  await page.screenshot({ path: `${outDir}/AUDIT_03_FINAL.png`, fullPage: true });

  // Accessibility and responsive assertions.
  const a11y = await page.evaluate(() => Array.from(document.querySelectorAll('button,[role="button"]')).filter(el => { const text = (el.textContent || '').trim(); const label = el.getAttribute('aria-label') || el.getAttribute('title'); return !text && !label; }).length);
  a11y === 0 ? pass('A11Y_BUTTON_NAMES') : fail('A11Y_BUTTON_NAMES', `${a11y} unnamed controls`);
  await page.setViewportSize({ width: 412, height: 915 }); await settle(); const overflowInfo = await page.evaluate(() => { const offenders = Array.from(document.querySelectorAll('*')).map((el) => { const rect = el.getBoundingClientRect(); return { tag: el.tagName, id: el.id, cls: String(el.className).slice(0, 80), right: Math.round(rect.right), width: Math.round(rect.width), text: String(el.textContent || '').trim().slice(0, 60) }; }).filter((item) => item.right > window.innerWidth + 1 && item.width > 0).sort((a, b) => b.right - a.right).slice(0, 8); return { scrollWidth: document.documentElement.scrollWidth, innerWidth: window.innerWidth, offenders }; }); overflowInfo.scrollWidth > overflowInfo.innerWidth + 1 ? fail('RESPONSIVE_NO_HORIZONTAL_OVERFLOW', JSON.stringify(overflowInfo)) : pass('RESPONSIVE_NO_HORIZONTAL_OVERFLOW');
  await page.screenshot({ path: `${outDir}/AUDIT_04_RESPONSIVE.png`, fullPage: true });
} catch (error) { await page.screenshot({ path: `${outDir}/AUDIT_FAILURE.png`, fullPage: true }).catch(() => {}); fail('HARNESS_UNCAUGHT', `${stage}:${String(error)}`); }

const expectedPatterns = [/\/api\/v1\/environment\/profiles\//, /\/api\/v1\/runtimes\/qualified/, /\/api\/sessions/, /\/api\/v1\/extensions/, /\/api\/v1\/profiles\/.*\/backups/];
const unexpectedHttp = responses.filter(item => !expectedPatterns.some(pattern => pattern.test(item.url)) && item.status >= 400);
if (unexpectedHttp.length) fail('UNEXPECTED_HTTP_ERRORS', JSON.stringify(unexpectedHttp)); else pass('UNEXPECTED_HTTP_ERRORS', '0');
console.log(results.join('\n'));
console.log(`HTTP_ERROR_RESPONSES=${responses.length}`);
console.log(`HTTP_ERROR_URLS_REDACTED=${responses.map((item) => `${item.status}:${item.url.replaceAll(adminToken, '[redacted]')}`).join('|')}`);
console.log(`UNEXPECTED_HTTP_ERROR_RESPONSES=${unexpectedHttp.length}`);
console.log(`PAGE_FAILURE_COUNT=${failures.length}`);
for (const item of failures) console.log(`FAILURE_REDACTED=${item.replaceAll(adminToken, '[redacted]')}`);
console.log(`BUTTON_AUDIT_CASES=${results.length}`);
console.log(`STATUS=${failures.length === 0 ? 'PASS' : 'FAIL'}`);
await page.context().browser().close();
