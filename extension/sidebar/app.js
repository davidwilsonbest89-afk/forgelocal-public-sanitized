// BrowseForge Sidebar App
const API = 'http://127.0.0.1:19280/api';
let profiles = [];
let token = '';

async function init() {
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
    const g = p.group || '未分組';
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
  const name = prompt('Profile 名稱：');
  if (!name) return;
  const engine = prompt('引擎 (firefox / chromium)：', 'firefox');
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
  const action = prompt(`${profile.name}\n\n1. 編輯\n2. 複製\n3. 刪除\n\n選擇 (1-3)：`);
  if (action === '3') deleteProfile(profile);
  if (action === '2') duplicateProfile(profile);
}

async function deleteProfile(p) {
  if (!confirm(`確定刪除 ${p.name}？`)) return;
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

init();
