// BrowseForge Sidebar App
const API = 'http://127.0.0.1:19280/api';
let profiles = [];
let token = '';

async function init() {
  applyI18n();
  const stored = await browser.storage.local.get('apiToken');
  token = stored.apiToken || '';
  await loadProfiles();
  document.getElementById('search').addEventListener('input', e => render(e.target.value));
  document.getElementById('btn-add').addEventListener('click', addProfile);
}

async function loadProfiles() {
  try {
    const res = await fetch(`${API}/profiles`, { headers: { 'Authorization': `Bearer ${token}` } });
    const json = await res.json();
    profiles = json.data || [];
  } catch {
    profiles = [];
  }
  render();
}

function render(filter = '') {
  const list = document.getElementById('profile-list');
  const filtered = profiles.filter(p =>
    !filter || p.name.toLowerCase().includes(filter.toLowerCase()) ||
    (p.group || '').toLowerCase().includes(filter.toLowerCase())
  );

  const groups = {};
  for (const p of filtered) {
    const g = p.group || msg('ungrouped');
    (groups[g] = groups[g] || []).push(p);
  }

  list.innerHTML = '';
  for (const [group, items] of Object.entries(groups)) {
    const header = document.createElement('div');
    header.className = 'group-header';
    header.textContent = `📁 ${group} (${items.length})`;
    list.appendChild(header);

    for (const p of items) {
      const item = document.createElement('div');
      item.className = 'profile-item';
      item.innerHTML = `
        <span class="status ${p._active ? 'active' : ''}"></span>
        <span class="engine">${p.engine === 'chromium' ? '🌐' : '🦊'}</span>
        <span class="name">${esc(p.name)}</span>
      `;
      item.addEventListener('click', () => openProfile(p));
      item.addEventListener('contextmenu', e => { e.preventDefault(); showContextMenu(e, p); });
      list.appendChild(item);
    }
  }
}

async function openProfile(p) {
  await browser.runtime.sendMessage({ type: 'open_profile', cookieStoreId: p.container_id });
}

async function addProfile() {
  const name = prompt(msg('profileNamePrompt'));
  if (!name) return;
  const engine = prompt(msg('enginePrompt'), 'firefox');
  try {
    await fetch(`${API}/profiles`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, engine: engine || 'firefox' }),
    });
    await loadProfiles();
  } catch (e) {
    console.error('Failed to create profile:', e);
  }
}

function showContextMenu(event, profile) {
  // Simple context menu — can be enhanced later
  const action = prompt(msg('contextMenu', profile.name));
  if (action === '3') deleteProfile(profile);
  if (action === '2') duplicateProfile(profile);
}

async function deleteProfile(p) {
  if (!confirm(msg('deleteConfirm', p.name))) return;
  await fetch(`${API}/profiles/${p.id}`, {
    method: 'DELETE',
    headers: { 'Authorization': `Bearer ${token}` },
  });
  await loadProfiles();
}

async function duplicateProfile(p) {
  await fetch(`${API}/profiles/${p.id}/duplicate`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
  });
  await loadProfiles();
}

function esc(s) {
  const d = document.createElement('div');
  d.textContent = s;
  return d.innerHTML;
}

function msg(key, substitutions) {
  const value = browser.i18n.getMessage(key, substitutions);
  return value || key;
}

function applyI18n() {
  document.documentElement.lang = browser.i18n.getUILanguage();
  document.querySelectorAll('[data-i18n]').forEach(el => { el.textContent = msg(el.dataset.i18n); });
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => { el.placeholder = msg(el.dataset.i18nPlaceholder); });
}

init();
