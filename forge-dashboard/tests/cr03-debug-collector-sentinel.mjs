import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';

const source = fs.readFileSync(new URL('../client/public/__manus__/debug-collector.js', import.meta.url), 'utf8');
const events = [];

class HeadersMock {
  constructor(input = {}) {
    this.values = new Map();
    if (input instanceof HeadersMock) {
      input.values.forEach((value, key) => this.values.set(key, value));
    } else if (Array.isArray(input)) {
      input.forEach(([key, value]) => this.values.set(String(key).toLowerCase(), String(value)));
    } else {
      Object.entries(input).forEach(([key, value]) => this.values.set(key.toLowerCase(), String(value)));
    }
  }
  get(name) { return this.values.get(String(name).toLowerCase()) ?? null; }
  entries() { return this.values.entries(); }
}

class XMLHttpRequestMock {}
XMLHttpRequestMock.prototype.open = function () {};
XMLHttpRequestMock.prototype.send = function () {};

const window = {
  addEventListener() {},
  innerWidth: 1280,
  innerHeight: 720,
  scrollX: 0,
  scrollY: 0,
  fetch: async () => ({ status: 200, statusText: 'OK', headers: new HeadersMock({ 'content-type': 'application/json', 'set-cookie': 'CR03_COOKIE_SENTINEL' }) }),
};
const context = {
  window,
  Headers: HeadersMock,
  XMLHttpRequest: XMLHttpRequestMock,
  Element: class {},
  document: { addEventListener() {}, documentElement: { scrollHeight: 0 } },
  history: { pushState() {}, replaceState() {} },
  location: { href: 'http://127.0.0.1:3100/' },
  navigator: {},
  console: { ...console, debug() {}, warn() {} },
  Error,
  JSON,
  Date,
  Promise,
  setInterval() { return 0; },
};
vm.createContext(context);
vm.runInContext(source, context, { filename: 'debug-collector.js' });

await window.fetch('/api/sentinel', {
  method: 'POST',
  headers: {
    Authorization: 'Bearer CR03_TOKEN_SENTINEL',
    Cookie: 'session=CR03_COOKIE_SENTINEL',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ token: 'CR03_TOKEN_SENTINEL', cookie: 'CR03_COOKIE_SENTINEL' }),
});

const state = window.__MANUS_DEBUG_COLLECTOR__.store.networkRequests;
assert.equal(state.length, 1, 'one fetch event must be recorded');
const entry = state[0];
assert.equal(entry.request.headers['content-type'], 'application/json');
assert.equal(entry.request.headers.authorization, undefined);
assert.equal(entry.request.headers.cookie, undefined);
assert.equal(entry.request.body, '[REDACTED_NOT_CAPTURED]');
assert.equal(entry.response.headers['set-cookie'], undefined);
assert.equal(entry.response.body, '[REDACTED_NOT_CAPTURED]');
assert.equal(JSON.stringify(state).includes('CR03_TOKEN_SENTINEL'), false);
assert.equal(JSON.stringify(state).includes('CR03_COOKIE_SENTINEL'), false);

assert.equal(source.includes('xhr.responseText'), false, 'XHR response bodies must not be captured');
assert.equal(source.includes('Object.fromEntries(response.headers.entries())'), false, 'response headers must use allowlist');
events.push('CR03_DEBUG_COLLECTOR_SENTINEL_PASS');
console.log(events.join('\n'));
