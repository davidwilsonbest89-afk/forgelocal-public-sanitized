#!/usr/bin/env node
const fs = require('fs');

function readJSON(path) {
  return JSON.parse(fs.readFileSync(path, 'utf8'));
}

function fail(message) {
  console.error(`i18n check failed: ${message}`);
  process.exit(1);
}

function sameKeys(name, a, b) {
  const ak = Object.keys(a).sort();
  const bk = Object.keys(b).sort();
  const missingInB = ak.filter(k => !bk.includes(k));
  const missingInA = bk.filter(k => !ak.includes(k));
  if (missingInB.length || missingInA.length) {
    fail(`${name} key mismatch. Missing in second: ${missingInB.join(', ') || '-'}; missing in first: ${missingInA.join(', ') || '-'}`);
  }
}

const enMessages = readJSON('extension/_locales/en/messages.json');
const zhMessages = readJSON('extension/_locales/zh_TW/messages.json');
sameKeys('extension locale', enMessages, zhMessages);

for (const [locale, messages] of Object.entries({ en: enMessages, zh_TW: zhMessages })) {
  for (const [key, value] of Object.entries(messages)) {
    if (!value || typeof value.message !== 'string' || value.message.length === 0) {
      fail(`extension ${locale}.${key} must include a non-empty message`);
    }
  }
}

const dashboard = fs.readFileSync('internal/api/dashboard.html', 'utf8');
const match = dashboard.match(/const LOCALES = (\{[\s\S]*?\n\});/);
if (match) {
  const locales = Function(`"use strict"; return (${match[1]});`)();
  if (!locales.en || !locales['zh-TW']) fail('Dashboard must define en and zh-TW locales');
  sameKeys('dashboard locale', locales.en, locales['zh-TW']);

  for (const [locale, messages] of Object.entries(locales)) {
    for (const [key, value] of Object.entries(messages)) {
      if (typeof value !== 'string' || value.length === 0) {
        fail(`dashboard ${locale}.${key} must be a non-empty string`);
      }
    }
  }
  console.log('i18n coverage ok');
} else {
  const historicalDisabled = /Interface historique désactivée/i.test(dashboard) && !/<script\b/i.test(dashboard);
  if (!historicalDisabled) fail('Dashboard has neither a valid LOCALES object nor the approved static disabled form');
  console.log('extension i18n coverage ok; historical dashboard is static and disabled');
}
