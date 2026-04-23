// BrowseForge Background Script
// Responsibilities: proxy routing, container management, server communication

const API_BASE = 'http://127.0.0.1:19280/api';
let apiToken = '';
let profileMap = new Map();       // profileId → profile
let containerMap = new Map();     // cookieStoreId → profileId
let ws = null;

// --- Initialization ---

async function init() {
  console.log('[BrowseForge] Background script starting...');

  // Load token from storage
  const stored = await browser.storage.local.get('apiToken');
  apiToken = stored.apiToken || '';

  // Load profile-container mappings
  const maps = await browser.storage.local.get('containerMap');
  if (maps.containerMap) {
    containerMap = new Map(Object.entries(maps.containerMap));
  }

  // Register proxy listener
  browser.proxy.onRequest.addListener(handleProxyRequest, { urls: ['<all_urls>'] });

  // Register auth listener (suppress 407 popups)
  browser.webRequest.onAuthRequired.addListener(
    handleProxyAuth,
    { urls: ['<all_urls>'] },
    ['blocking']
  );

  // Register tab listeners
  browser.tabs.onUpdated.addListener(handleTabUpdated);
  browser.tabs.onRemoved.addListener(handleTabRemoved);

  // Connect to Control Server
  connectWebSocket();

  console.log('[BrowseForge] Background script ready');
}

// --- Proxy Routing (synchronous, per-container) ---

function handleProxyRequest(requestInfo) {
  const profileId = containerMap.get(requestInfo.cookieStoreId);
  if (!profileId) return { type: 'direct' };

  const profile = profileMap.get(profileId);
  if (!profile?.proxy) return { type: 'direct' };

  return {
    type: profile.proxy.type || 'direct',
    host: profile.proxy.host,
    port: profile.proxy.port,
    username: profile.proxy.username || '',
    password: profile.proxy.password || '',
    proxyDNS: true,
  };
}

function handleProxyAuth(details) {
  const profileId = containerMap.get(details.cookieStoreId);
  if (!profileId) return {};

  const profile = profileMap.get(profileId);
  if (!profile?.proxy?.username) return {};

  return {
    authCredentials: {
      username: profile.proxy.username,
      password: profile.proxy.password,
    },
  };
}

// --- Container Management ---

async function createContainer(profileId, name, color) {
  const container = await browser.contextualIdentities.create({
    name: name,
    color: color || 'blue',
    icon: 'fingerprint',
  });
  containerMap.set(container.cookieStoreId, profileId);
  await saveContainerMap();
  return container;
}

async function removeContainer(cookieStoreId) {
  // Close all tabs in this container
  const tabs = await browser.tabs.query({ cookieStoreId });
  for (const tab of tabs) {
    await browser.tabs.remove(tab.id);
  }
  // Clear container data
  await browser.browsingData.remove(
    { cookieStoreId },
    { cookies: true, localStorage: true, indexedDB: true, cache: true }
  );
  // Remove container
  await browser.contextualIdentities.remove(cookieStoreId);
  containerMap.delete(cookieStoreId);
  await saveContainerMap();
}

async function openContainerTab(cookieStoreId, url) {
  return browser.tabs.create({
    cookieStoreId,
    url: url || 'about:blank',
  });
}

// --- Tab Monitoring ---

function handleTabUpdated(tabId, changeInfo, tab) {
  if (!changeInfo.title) return;
  const keywords = ['驗證', 'verify', 'checkpoint', 'captcha', 'confirm your identity'];
  const title = (changeInfo.title || '').toLowerCase();
  if (keywords.some(k => title.includes(k))) {
    const profileId = containerMap.get(tab.cookieStoreId);
    if (profileId) {
      notifyServer('need_attention', { profile_id: profileId, tab_id: tabId, reason: 'captcha', title: changeInfo.title });
    }
  }
}

function handleTabRemoved(tabId, removeInfo) {
  notifyServer('tab_closed', { tab_id: tabId });
}

// --- WebSocket Communication ---

function connectWebSocket() {
  const port = 19280;
  ws = new WebSocket(`ws://127.0.0.1:${port}/ws`);

  ws.onopen = () => {
    console.log('[BrowseForge] WebSocket connected');
    ws.send(JSON.stringify({ type: 'extension_ready', payload: { version: '0.1.0' } }));
  };

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    handleServerMessage(msg);
  };

  ws.onclose = () => {
    console.log('[BrowseForge] WebSocket disconnected, reconnecting...');
    setTimeout(connectWebSocket, 2000);
  };
}

async function handleServerMessage(msg) {
  switch (msg.type) {
    case 'open_tab':
      await openContainerTab(msg.payload.cookie_store_id, msg.payload.url);
      break;
    case 'close_tab':
      if (msg.payload.tab_id) await browser.tabs.remove(msg.payload.tab_id);
      break;
    case 'update_profile':
      profileMap.set(msg.payload.profile_id, msg.payload.profile);
      // Update fingerprint in storage for content script
      if (msg.payload.profile.fingerprint) {
        const key = `fp_${msg.payload.cookie_store_id}`;
        await browser.storage.local.set({ [key]: msg.payload.profile.fingerprint });
      }
      break;
    case 'delete_profile':
      if (msg.payload.cookie_store_id) await removeContainer(msg.payload.cookie_store_id);
      profileMap.delete(msg.payload.profile_id);
      break;
  }
}

function notifyServer(type, payload) {
  if (ws?.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type, payload }));
  }
}

// --- Storage ---

async function saveContainerMap() {
  const obj = Object.fromEntries(containerMap);
  await browser.storage.local.set({ containerMap: obj });
}

// --- Message handling from sidebar ---

browser.runtime.onMessage.addListener((msg, sender) => {
  switch (msg.type) {
    case 'get_profiles':
      return Promise.resolve([...profileMap.values()]);
    case 'get_container_map':
      return Promise.resolve(Object.fromEntries(containerMap));
    case 'open_profile':
      return openContainerTab(msg.cookieStoreId, msg.url);
  }
});

// Start
init();
