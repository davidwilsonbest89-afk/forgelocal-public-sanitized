#!/usr/bin/env node
'use strict';

const fs = require('fs');

function get(obj, path) {
  return path.split('.').reduce((cur, key) => (cur && Object.prototype.hasOwnProperty.call(cur, key) ? cur[key] : undefined), obj);
}

function normalize(value) {
  if (value === undefined || value === null) return '';
  if (Array.isArray(value) || (typeof value === 'object' && value !== null)) return JSON.stringify(value);
  return String(value);
}

const CONTRACT_FIELDS = [
  ['navigator.userAgent', 'browser.user_agent'],
  ['navigator.platform', 'platform.platform'],
  ['navigator.language', 'locale.navigator_language'],
  ['navigator.languages', 'locale.navigator_languages'],
  ['intl.timeZone', 'locale.timezone'],
  ['date.timezoneOffset', 'locale.timezone_offset_mins'],
  ['hardware.hardwareConcurrency', 'hardware.hardware_concurrency'],
  ['hardware.deviceMemory', 'hardware.device_memory_gb'],
  ['screen.width', 'screen.width'],
  ['screen.height', 'screen.height'],
  ['screen.availWidth', 'screen.avail_width'],
  ['screen.availHeight', 'screen.avail_height'],
  ['screen.devicePixelRatio', 'screen.dpr'],
  ['screen.outerWidth', 'screen.outer_width'],
  ['screen.outerHeight', 'screen.outer_height'],
  ['screen.innerWidth', 'screen.inner_width'],
  ['screen.innerHeight', 'screen.inner_height'],
  ['screen.viewportWidth', 'screen.viewport_width'],
  ['screen.viewportHeight', 'screen.viewport_height'],
  ['screen.colorDepth', 'screen.color_depth'],
  ['screen.touchPoints', 'screen.touch_points'],
  ['screen.orientation', 'screen.orientation'],
  ['webgl.vendor', 'gpu.vendor'],
  ['webgl.renderer', 'gpu.renderer'],
  ['webgl.version', 'gpu.gl_version'],
  ['webgl.shadingLanguageVersion', 'gpu.shading_language_version'],
  ['navigator.userAgentData.platform', 'browser.client_hints.platform'],
  ['navigator.userAgentData.mobile', 'browser.client_hints.mobile'],
  ['navigator.userAgentData.architecture', 'browser.client_hints.architecture'],
  ['navigator.userAgentData.bitness', 'browser.client_hints.bitness'],
  ['navigator.userAgentData.brands', 'browser.brand_versions'],
  ['navigator.userAgentData.fullVersionList', 'browser.full_version_list'],
  ['navigator.userAgentData.platformVersion', 'browser.client_hints.platform_version'],
  ['navigator.userAgentData.model', 'browser.client_hints.model'],
  ['navigator.userAgentData.formFactors', 'browser.client_hints.form_factors'],
];

const OPTIONAL_CONTRACT_FIELDS = [
  ['headers.userAgent', 'browser.user_agent'],
  ['headers.acceptLanguage', 'locale.accept_language'],
  ['headers.secCHUA', 'browser.client_hints.sec_ch_ua'],
  ['headers.secCHUAFullVersionList', 'browser.client_hints.sec_ch_ua_full_version_list'],
  ['headers.secCHUAPlatform', 'browser.client_hints.platform', 'secCHString'],
  ['headers.secCHUAPlatformVersion', 'browser.client_hints.platform_version', 'secCHString'],
  ['headers.secCHUAArch', 'browser.client_hints.architecture', 'secCHString'],
  ['headers.secCHUABitness', 'browser.client_hints.bitness', 'secCHString'],
  ['headers.secCHUAMobile', 'browser.client_hints.mobile', 'secCHBool'],
  ['headers.secCHUAModel', 'browser.client_hints.model', 'secCHString'],
  ['headers.secCHUAFormFactors', 'browser.client_hints.form_factors', 'secCHStringList'],
  ['headers.secCHLang', 'locale.sec_ch_lang'],
];

const DETECTOR_MATRIX = [
  { target: 'https://bot.sannysoft.com/', focus: ['HeadlessChrome', 'webdriver', 'Selenium/WebDriver globals', 'plugins', 'permissions', 'languages'], artifact: 'result table or screenshot' },
  { target: 'https://browserleaks.com/client-hints', focus: ['UA-CH', 'navigator.userAgentData high entropy'], artifact: 'visible fields or JSON' },
  { target: 'https://browserleaks.com/webgl', focus: ['WebGL vendor/renderer', 'extensions', 'limits', 'shader precision'], artifact: 'WebGL report' },
  { target: 'https://browserleaks.com/canvas', focus: ['Canvas', 'text', 'emoji rendering stability'], artifact: 'canvas hash/report' },
  { target: 'https://browserleaks.com/fonts', focus: ['font set', 'font metrics', 'emoji font'], artifact: 'font report' },
  { target: 'https://browserleaks.com/webrtc', focus: ['host/container/public IP leak'], artifact: 'WebRTC report' },
  { target: 'https://browserleaks.com/javascript', focus: ['JavaScript feature shape', 'screen', 'navigator', 'storage', 'geolocation permission', 'client rects'], artifact: 'JavaScript report' },
  { target: 'https://www.browserscan.net/', focus: ['Browser', 'Location', 'IP', 'Hardware', 'Software trust coherence'], artifact: 'report sections' },
  { target: 'https://www.browserscan.net/client-hints', focus: ['UA-CH / JS UAData parity'], artifact: 'Client Hints report' },
  { target: 'https://www.browserscan.net/dns-leak', focus: ['DNS resolver geolocation/ASN leak'], artifact: 'DNS leak report' },
  { target: 'https://www.browserscan.net/webrtc', focus: ['WebRTC host/container/public IP leak'], artifact: 'WebRTC report' },
  { target: 'https://www.browserscan.net/bot-detection', focus: ['automation/headless signals'], artifact: 'bot detection report' },
  { target: 'https://iphey.com/', focus: ['Browser', 'Location', 'IP', 'Hardware', 'Software trust score'], artifact: 'main score sections' },
  { target: 'https://abrahamjuliot.github.io/creepjs/', focus: ['prototype lies', 'realm parity', 'fonts/canvas/audio/math/screen/client-rects'], artifact: 'main result' },
  { target: 'https://abrahamjuliot.github.io/creepjs/tests/workers.html', focus: ['worker parity'], artifact: 'worker test result' },
  { target: 'https://abrahamjuliot.github.io/creepjs/tests/iframes.html', focus: ['iframe parity'], artifact: 'iframe test result' },
  { target: 'https://abrahamjuliot.github.io/creepjs/tests/prototype.html', focus: ['prototype/native descriptor parity'], artifact: 'prototype test result' },
  { target: 'https://pixelscan.net', focus: ['fingerprint consistency', 'IP / timezone / WebRTC / OS mismatch'], artifact: 'fingerprint / bot checker result' },
];

const STABLE_RESULT_CRITERIA = {
  readyState: 'complete',
  stableWindowMs: 3000,
  resourceQuietWindowMs: 3000,
  minTextLength: 100,
  minNodeCount: 50,
};

function stripSecCHQuotes(value) {
  const text = normalize(value).trim();
  if (text.length >= 2 && text.startsWith('"') && text.endsWith('"')) {
    try {
      return JSON.parse(text);
    } catch (_) {
      return text.slice(1, -1);
    }
  }
  return text;
}

function formatExpected(value, format) {
  if (format === 'secCHBool') return value === true ? '?1' : '?0';
  if (format === 'commaList' && Array.isArray(value)) return value.join(',');
  if (format === 'secCHString') return normalize(value);
  if (format === 'secCHStringList' && Array.isArray(value)) return value.map(stripSecCHQuotes).filter(Boolean).join(',');
  return normalize(value);
}

function formatActual(value, format) {
  if (format === 'secCHString') return stripSecCHQuotes(value);
  if (format === 'secCHStringList') {
    return normalize(value).split(',').map(stripSecCHQuotes).filter(Boolean).join(',');
  }
  return normalize(value);
}

function compareContract(contract, sample) {
  const mismatches = [];
  for (const [samplePath, contractPath] of CONTRACT_FIELDS) {
    const actual = normalize(get(sample, samplePath));
    const expected = normalize(get(contract, contractPath));
    if (expected !== '' && actual !== expected && expected !== 'browser-default') {
      mismatches.push({ field: samplePath, expected, actual });
    }
  }
  for (const [samplePath, contractPath, format] of OPTIONAL_CONTRACT_FIELDS) {
    const rawActual = get(sample, samplePath);
    const actual = formatActual(rawActual, format);
    const expected = formatExpected(get(contract, contractPath), format);
    if (actual !== '' && expected !== '' && actual !== expected && expected !== 'browser-default') {
      mismatches.push({ field: samplePath, expected, actual });
    }
  }
  if (String(get(sample, 'navigator.userAgent') || '').includes('HeadlessChrome')) {
    mismatches.push({ field: 'automation.userAgent', expected: 'no HeadlessChrome token', actual: get(sample, 'navigator.userAgent') });
  }
  if (get(sample, 'navigator.webdriver') === true) {
    mismatches.push({ field: 'automation.webdriver', expected: 'false or undefined', actual: 'true' });
  }
  const automationGlobals = get(sample, 'automation.automationGlobals') || get(sample, 'automation.seleniumGlobals') || [];
  for (const marker of automationGlobals) {
    mismatches.push({ field: 'automation.automationGlobals', expected: `missing automation/CDP global ${marker}`, actual: marker });
  }
  if (Array.isArray(get(sample, 'navigator.languages')) && get(sample, 'navigator.languages').length === 0) {
    mismatches.push({ field: 'navigator.languages', expected: 'non-empty language list', actual: '[]' });
  }
  if (get(contract, 'plugins.pdf_viewer') === true && Array.isArray(get(sample, 'navigator.plugins')) && get(sample, 'navigator.plugins').length === 0) {
    mismatches.push({ field: 'navigator.plugins', expected: 'PDF viewer plugin entries', actual: '[]' });
  }
  if (get(contract, 'plugins.pdf_viewer') === true && Array.isArray(get(sample, 'navigator.mimeTypes')) && !get(sample, 'navigator.mimeTypes').includes('application/pdf')) {
    mismatches.push({ field: 'navigator.mimeTypes', expected: 'application/pdf', actual: get(sample, 'navigator.mimeTypes') });
  }
  if (get(sample, 'navigator.pluginArrayTag') && get(sample, 'navigator.pluginArrayTag') !== '[object PluginArray]') {
    mismatches.push({ field: 'navigator.plugins', expected: '[object PluginArray]', actual: get(sample, 'navigator.pluginArrayTag') });
  }
  if (get(sample, 'navigator.mimeTypeArrayTag') && get(sample, 'navigator.mimeTypeArrayTag') !== '[object MimeTypeArray]') {
    mismatches.push({ field: 'navigator.mimeTypes', expected: '[object MimeTypeArray]', actual: get(sample, 'navigator.mimeTypeArrayTag') });
  }
  if (get(sample, 'automation.chromeRuntimeShape') === false) {
    mismatches.push({ field: 'automation.chromeRuntimeShape', expected: 'present chrome object', actual: 'missing' });
  }
  if ((get(sample, 'automation.webdriverAttributes') || []).length > 0) {
    mismatches.push({ field: 'automation.webdriverAttributes', expected: 'none', actual: get(sample, 'automation.webdriverAttributes') });
  }
  const expectedNotification = get(contract, 'permissions.notification');
  const actualNotification = get(sample, 'automation.notificationPermission');
  if (expectedNotification === 'prompt' && actualNotification && !['prompt', 'default'].includes(actualNotification)) {
    mismatches.push({ field: 'automation.notificationPermission', expected: 'prompt/default', actual: actualNotification });
  }
  const permissionState = get(sample, 'automation.notificationPermissionState');
  if (expectedNotification === 'prompt' && permissionState && !['prompt', 'default'].includes(permissionState)) {
    mismatches.push({ field: 'automation.notificationPermissionState', expected: 'prompt/default', actual: permissionState });
  }
  const geolocationPolicy = get(contract, 'geolocation') || {};
  const expectedGeolocationMode = geolocationPolicy.mode || '';
  if (expectedGeolocationMode) {
    if (get(sample, 'geolocation.available') !== true) {
      mismatches.push({ field: 'geolocation.available', expected: `available for ${expectedGeolocationMode}`, actual: normalize(get(sample, 'geolocation.error')) || 'missing' });
    }
    const geolocationPermissionState = get(sample, 'geolocation.permissionState');
    if (geolocationPermissionState === 'denied') {
      mismatches.push({ field: 'geolocation.permissionState', expected: 'prompt/granted/default', actual: geolocationPermissionState });
    }
  }
  const audioPolicy = get(contract, 'audio') || {};
  const expectedAudioSampleRate = Number(audioPolicy.sample_rate || 0);
  if (expectedAudioSampleRate > 0) {
    const actualAudioSampleRate = Number(get(sample, 'audio.sampleRate') || 0);
    if (actualAudioSampleRate !== expectedAudioSampleRate) {
      mismatches.push({ field: 'audio.sampleRate', expected: String(expectedAudioSampleRate), actual: actualAudioSampleRate > 0 ? String(actualAudioSampleRate) : normalize(get(sample, 'audio.error')) || 'missing' });
    }
  }
  if (audioPolicy.stable === true && !get(sample, 'audio.hash')) {
    mismatches.push({ field: 'audio.hash', expected: 'stable audio fingerprint sample', actual: normalize(get(sample, 'audio.error')) || 'missing' });
  }
  const fontPolicy = get(contract, 'fonts') || {};
  const availableFonts = get(sample, 'fonts.available') || {};
  const expectedFamilies = Array.isArray(fontPolicy.families) ? fontPolicy.families : [];
  if (expectedFamilies.length > 0) {
    const missingFamilies = expectedFamilies.filter((family) => availableFonts[family] !== true);
    if (missingFamilies.length > 0) {
      mismatches.push({ field: 'fonts.families', expected: `available ${missingFamilies.join(',')}`, actual: normalize(get(sample, 'fonts.error')) || 'missing/unavailable' });
    }
  }
  if (fontPolicy.emoji && availableFonts[fontPolicy.emoji] !== true) {
    mismatches.push({ field: 'fonts.emoji', expected: `available ${fontPolicy.emoji}`, actual: normalize(get(sample, 'fonts.error')) || 'missing/unavailable' });
  }
  if (fontPolicy.cjk === true && get(sample, 'fonts.cjkAvailable') !== true) {
    mismatches.push({ field: 'fonts.cjk', expected: 'available CJK font family', actual: normalize(get(sample, 'fonts.cjkAvailable')) || 'missing' });
  }
  if (get(contract, 'canvas.text_metrics_mode') && !get(sample, 'fonts.metricsHash')) {
    mismatches.push({ field: 'fonts.metricsHash', expected: 'stable TextMetrics sample', actual: normalize(get(sample, 'fonts.error')) || 'missing' });
  }
  if (get(contract, 'canvas.stable') === true && !get(sample, 'canvas.dataURL')) {
    mismatches.push({ field: 'canvas.dataURL', expected: 'stable canvas sample', actual: normalize(get(sample, 'canvas.error')) || 'missing' });
  }
  if (get(contract, 'math.stable') === true && !get(sample, 'math.hash')) {
    mismatches.push({ field: 'math.hash', expected: 'stable math fingerprint sample', actual: normalize(get(sample, 'math.error')) || 'missing' });
  }
  if (get(contract, 'geometry.client_rects_stable') === true && !get(sample, 'geometry.clientRectsHash')) {
    mismatches.push({ field: 'geometry.clientRectsHash', expected: 'stable client rect sample', actual: normalize(get(sample, 'geometry.error')) || 'missing' });
  }
  const mediaCodecChecks = [
    ['h264', 'h264Support', 'H264'],
    ['vp9', 'vp9Support', 'VP9'],
    ['av1', 'av1Support', 'AV1'],
  ];
  for (const [contractKey, sampleKey, label] of mediaCodecChecks) {
    if (get(contract, `media.${contractKey}`) === true && get(sample, `media.${sampleKey}`) === '') {
      mismatches.push({ field: `media.${sampleKey}`, expected: `${label} probably or maybe`, actual: '' });
    }
  }
  const expectedMediaDevices = get(contract, 'media.devices') || [];
  if (Array.isArray(expectedMediaDevices) && expectedMediaDevices.length > 0) {
    if (get(sample, 'media.mediaDevices') !== true) {
      mismatches.push({ field: 'media.mediaDevices', expected: 'available mediaDevices', actual: normalize(get(sample, 'media.mediaDevices')) || 'missing' });
    }
    if (get(sample, 'media.enumerateDevices') !== true) {
      mismatches.push({ field: 'media.enumerateDevices', expected: 'available enumerateDevices', actual: normalize(get(sample, 'media.enumerateDevices')) || 'missing' });
    }
  }
  if (get(contract, 'gpu.worker_offscreen_canvas') === true && get(sample, 'workers.offscreenCanvasWebGL') === false) {
    mismatches.push({ field: 'workers.offscreenCanvasWebGL', expected: 'available', actual: 'false' });
  }
  if (get(contract, 'gpu.webgl') === true && get(sample, 'webgl.error')) {
    mismatches.push({ field: 'webgl', expected: 'available WebGL context', actual: get(sample, 'webgl.error') });
  }
  if (get(contract, 'gpu.webgl2') === true && get(sample, 'webgl.webgl2') !== true) {
    mismatches.push({ field: 'webgl.webgl2', expected: 'available WebGL2 context', actual: normalize(get(sample, 'webgl.webgl2')) || 'missing' });
  }
  const webgpuPolicy = get(contract, 'gpu.webgpu');
  if (webgpuPolicy && get(sample, 'webgpu.available') === undefined) {
    mismatches.push({ field: 'webgpu.available', expected: `reported ${webgpuPolicy}`, actual: 'missing' });
  }
  const webrtcPolicy = get(contract, 'webrtc') || {};
  if (webrtcPolicy.mode && get(sample, 'webrtc.available') === undefined) {
    mismatches.push({ field: 'webrtc.available', expected: `reported ${webrtcPolicy.mode}`, actual: 'missing' });
  }
  if (webrtcPolicy.direct_ip_redaction === true) {
    if (get(sample, 'webrtc.hasHostCandidate') === true) {
      mismatches.push({ field: 'webrtc.hostCandidates', expected: 'none with direct IP redaction', actual: 'present' });
    }
    if (get(sample, 'webrtc.hasSrflxCandidate') === true) {
      mismatches.push({ field: 'webrtc.srflxCandidates', expected: 'none with direct IP redaction', actual: 'present' });
    }
    if (get(sample, 'webrtc.hasPrivateAddress') === true) {
      mismatches.push({ field: 'webrtc.privateAddress', expected: 'no host/container private address', actual: 'present' });
    }
  }
  const requiredExtensions = get(contract, 'gpu.extensions') || [];
  const actualExtensions = get(sample, 'webgl.extensions') || [];
  if (Array.isArray(requiredExtensions) && requiredExtensions.length > 0 && Array.isArray(actualExtensions)) {
    const missingExtensions = requiredExtensions.filter((extension) => !actualExtensions.includes(extension));
    if (missingExtensions.length > 0) {
      mismatches.push({ field: 'webgl.extensions', expected: `includes ${missingExtensions.join(',')}`, actual: actualExtensions.join(',') });
    }
  }
  const expectedLimits = get(contract, 'gpu.limits') || {};
  for (const [limitName, expectedLimit] of Object.entries(expectedLimits)) {
    const actualLimit = Number(get(sample, `webgl.limits.${limitName}`) || 0);
    if (Number(expectedLimit) > 0 && actualLimit < Number(expectedLimit)) {
      mismatches.push({ field: `webgl.limits.${limitName}`, expected: `>=${expectedLimit}`, actual: actualLimit > 0 ? String(actualLimit) : 'missing' });
    }
  }
  const expectedFragmentHighFloat = get(contract, 'gpu.shader_precision.fragmentHighFloat');
  const actualFragmentHighFloat = get(sample, 'webgl.shaderPrecision.fragmentHighFloat');
  if (expectedFragmentHighFloat && expectedFragmentHighFloat !== 'browser-default' && actualFragmentHighFloat && actualFragmentHighFloat !== expectedFragmentHighFloat) {
    mismatches.push({ field: 'webgl.shaderPrecision.fragmentHighFloat', expected: expectedFragmentHighFloat, actual: actualFragmentHighFloat });
  }
  const storagePolicy = get(contract, 'storage') || {};
  const storageChecks = [
    ['cookiesEnabled', 'cookies', 'profile-persistent'],
    ['localStorage', 'local_storage', 'profile-persistent'],
    ['sessionStorage', 'session_storage', 'session-scoped'],
    ['indexedDB', 'indexed_db', 'profile-persistent'],
  ];
  for (const [sampleKey, contractKey, availablePolicy] of storageChecks) {
    if (storagePolicy[contractKey] === availablePolicy && get(sample, `storage.${sampleKey}`) !== true) {
      mismatches.push({ field: `storage.${sampleKey}`, expected: 'available', actual: normalize(get(sample, `storage.${sampleKey}`)) || 'missing' });
    }
  }
  if (storagePolicy.persistent === true && get(sample, 'storage.persisted') === false) {
    mismatches.push({ field: 'storage.persisted', expected: 'persistent storage', actual: 'false' });
  }
  const quotaMB = Number(storagePolicy.quota_mb || 0);
  const quotaBytesRaw = get(sample, 'storage.storageEstimate.quota');
  const quotaBytes = Number(quotaBytesRaw || 0);
  if (quotaMB > 0 && quotaBytes <= 0) {
    mismatches.push({ field: 'storage.storageEstimate.quota', expected: `>=${quotaMB}MiB`, actual: normalize(quotaBytesRaw) || 'missing' });
  } else if (quotaMB > 0 && quotaBytes < quotaMB * 1024 * 1024) {
    mismatches.push({ field: 'storage.storageEstimate.quota', expected: `>=${quotaMB}MiB`, actual: String(quotaBytes) });
  }
  const realmComparisons = [
    ['userAgent', 'navigator.userAgent'],
    ['platform', 'navigator.platform'],
    ['language', 'navigator.language'],
    ['languages', 'navigator.languages'],
    ['hardwareConcurrency', 'hardware.hardwareConcurrency'],
    ['deviceMemory', 'hardware.deviceMemory'],
    ['intlTimeZone', 'intl.timeZone'],
    ['intlLocale', 'intl.locale'],
    ['timezoneOffset', 'date.timezoneOffset'],
    ['devicePixelRatio', 'screen.devicePixelRatio'],
    ['webglVendor', 'webgl.vendor'],
    ['webglRenderer', 'webgl.renderer'],
    ['canvasDataURL', 'canvas.dataURL'],
    ['userAgentData.brands', 'navigator.userAgentData.brands'],
    ['userAgentData.mobile', 'navigator.userAgentData.mobile'],
    ['userAgentData.platform', 'navigator.userAgentData.platform'],
    ['userAgentData.architecture', 'navigator.userAgentData.architecture'],
    ['userAgentData.bitness', 'navigator.userAgentData.bitness'],
    ['userAgentData.fullVersionList', 'navigator.userAgentData.fullVersionList'],
    ['userAgentData.model', 'navigator.userAgentData.model'],
    ['userAgentData.platformVersion', 'navigator.userAgentData.platformVersion'],
    ['userAgentData.formFactors', 'navigator.userAgentData.formFactors'],
  ];
  const requiredRealmComparisonKeys = new Set(realmComparisons.map(([realmKey]) => realmKey).filter((realmKey) => realmKey.startsWith('userAgentData.')));
  const realmSamples = Array.isArray(sample.realms) ? sample.realms : [];
  const presentRealms = new Set(realmSamples.filter((realm) => realm && !realm.unsupported && !realm.error).map((realm) => realm.name));
  const requiredRealmTargets = new Set(['same-origin-iframe', 'sandbox-iframe', 'nested-iframe', 'dedicated-worker', 'shared-worker']);
  for (const target of get(contract, 'realms.targets') || []) {
    if (requiredRealmTargets.has(target) && !presentRealms.has(target)) {
      mismatches.push({ field: `realms.${target}`, expected: 'available realm sample', actual: 'missing' });
    }
  }
  if ((get(contract, 'realms.targets') || []).includes('service-worker') && get(sample, 'workers.serviceWorkerAvailable') === false) {
    mismatches.push({ field: 'realms.service-worker', expected: 'service worker API available', actual: 'false' });
  }
  for (const realm of realmSamples) {
    if (realm.unsupported) {
      continue;
    }
    if (realm.error) {
      mismatches.push({ field: `realms.${realm.name}`, expected: 'available realm sample', actual: realm.error });
      continue;
    }
    for (const [realmKey, samplePath] of realmComparisons) {
      const expected = normalize(get(sample, samplePath));
      const actual = normalize(get(realm, realmKey));
      if (expected === '') {
        continue;
      }
      if (actual === '' && requiredRealmComparisonKeys.has(realmKey)) {
        mismatches.push({ field: `realms.${realm.name}.${realmKey}`, expected, actual: 'missing' });
      } else if (actual !== '' && expected !== actual) {
        mismatches.push({ field: `realms.${realm.name}.${realmKey}`, expected, actual });
      }
    }
  }
  return { ok: mismatches.length === 0, mismatches };
}

function browserCollectorSource() {
  return `(${async function collectBrowseForgeDetectorSample() {
    const hashString = (value) => {
      let checksum = 2166136261;
      const text = String(value);
      for (let i = 0; i < text.length; i += 1) {
        checksum ^= text.charCodeAt(i);
        checksum = Math.imul(checksum, 16777619) >>> 0;
      }
      return checksum.toString(16).padStart(8, '0');
    };
    const canvasFingerprint = (doc = document) => {
      try {
        const canvas = doc.createElement('canvas');
        canvas.width = 240;
        canvas.height = 64;
        const ctx = canvas.getContext('2d');
        if (!ctx) return null;
        ctx.textBaseline = 'alphabetic';
        ctx.font = '16px Arial, "Noto Color Emoji", sans-serif';
        ctx.fillStyle = '#f60';
        ctx.fillRect(4, 4, 120, 28);
        ctx.fillStyle = '#069';
        ctx.fillText('BrowseForge 🧪', 8, 24);
        ctx.strokeStyle = '#0a0';
        ctx.strokeText('Canvas', 8, 48);
        return canvas.toDataURL();
      } catch (error) {
        return { error: String(error && error.message || error) };
      }
    };
    const readAudioFingerprint = async () => {
      try {
        const AudioCtx = window.OfflineAudioContext || window.webkitOfflineAudioContext;
        if (!AudioCtx) return { supported: false, error: 'OfflineAudioContext unavailable' };
        const context = new AudioCtx(1, 2048, 48000);
        const oscillator = context.createOscillator();
        const compressor = context.createDynamicsCompressor();
        oscillator.type = 'triangle';
        oscillator.frequency.value = 1000;
        compressor.threshold.value = -50;
        compressor.knee.value = 40;
        compressor.ratio.value = 12;
        compressor.attack.value = 0;
        compressor.release.value = 0.25;
        oscillator.connect(compressor);
        compressor.connect(context.destination);
        oscillator.start(0);
        oscillator.stop(0.05);
        const buffer = await context.startRendering();
        const samples = buffer.getChannelData(0);
        let checksum = 2166136261;
        for (let i = 0; i < samples.length; i += 16) {
          checksum ^= Math.round((samples[i] || 0) * 1000000);
          checksum = Math.imul(checksum, 16777619) >>> 0;
        }
        return { supported: true, sampleRate: buffer.sampleRate, length: buffer.length, hash: checksum.toString(16).padStart(8, '0') };
      } catch (error) {
        return { supported: false, error: String(error && error.message || error) };
      }
    };
    const readFonts = () => {
      try {
        const candidates = [
          'Noto Sans', 'Noto Serif', 'Noto Sans Mono', 'Liberation Sans', 'Liberation Serif', 'Liberation Mono',
          'DejaVu Sans', 'DejaVu Serif', 'DejaVu Sans Mono', 'Arial', 'Times New Roman', 'Courier New',
          'Noto Color Emoji', 'Apple Color Emoji', 'Segoe UI Emoji',
          'Noto Sans CJK TC', 'Noto Serif CJK TC', 'Noto Sans CJK SC', 'Noto Serif CJK SC',
          'PingFang TC', 'PingFang SC', 'Hiragino Sans', 'Songti TC',
          'Microsoft JhengHei', 'Microsoft YaHei', 'MingLiU', 'SimSun',
        ];
        const available = {};
        for (const family of candidates) {
          available[family] = !!(document.fonts && document.fonts.check && document.fonts.check(`16px "${family}"`));
        }
        const canvas = document.createElement('canvas');
        canvas.width = 360;
        canvas.height = 80;
        const ctx = canvas.getContext('2d');
        const metrics = {};
        if (ctx) {
          for (const [name, font, text] of [
            ['latinSans', '16px "Noto Sans", Arial, sans-serif', 'BrowseForge'],
            ['monoDigits', '16px "Noto Sans Mono", "Courier New", monospace', '0123456789'],
            ['emoji', '16px "Noto Color Emoji", "Apple Color Emoji", "Segoe UI Emoji", sans-serif', '🧪🔒'],
            ['cjk', '16px "Noto Sans CJK TC", "Microsoft JhengHei", sans-serif', '繁體測試'],
          ]) {
            ctx.font = font;
            const value = ctx.measureText(text);
            metrics[name] = Number(value.width.toFixed(3));
          }
        }
        const metricsHash = Object.keys(metrics).sort().map((key) => `${key}:${metrics[key]}`).join('|');
        const cjkFamilies = candidates.filter((family) => /CJK|PingFang|Hiragino|Songti|JhengHei|YaHei|MingLiU|SimSun/.test(family));
        return { available, metrics, metricsHash, cjkAvailable: cjkFamilies.some((family) => available[family] === true) };
      } catch (error) {
        return { error: String(error && error.message || error) };
      }
    };
    const readMathFingerprint = () => {
      try {
        const samples = [
          Math.acos(0.123456789),
          Math.asinh(1.23456789),
          Math.atanh(0.5),
          Math.cbrt(123456.789),
          Math.cos(Math.PI / 7),
          Math.expm1(1),
          Math.hypot(3, 4, 5),
          Math.log1p(10),
          Math.sin(Math.PI / 7),
          Math.tan(0.123456789),
        ].map((value) => Number(value).toPrecision(17));
        return { hash: hashString(samples.join('|')), samples };
      } catch (error) {
        return { error: String(error && error.message || error) };
      }
    };
    const readClientRects = (doc = document) => {
      let node = null;
      try {
        node = doc.createElement('div');
        node.textContent = 'BrowseForge client rect sample 🧪';
        node.style.cssText = 'position:absolute;left:-9999px;top:-9999px;width:180.5px;font:16px Arial, "Noto Color Emoji", sans-serif;line-height:19.25px;letter-spacing:0.125px;white-space:normal;';
        doc.documentElement.appendChild(node);
        const rects = Array.from(node.getClientRects()).map((rect) => ({
          x: Number(rect.x.toFixed(3)),
          y: Number(rect.y.toFixed(3)),
          width: Number(rect.width.toFixed(3)),
          height: Number(rect.height.toFixed(3)),
        }));
        const bounds = node.getBoundingClientRect();
        const summary = { rects, bounds: { width: Number(bounds.width.toFixed(3)), height: Number(bounds.height.toFixed(3)) } };
        return { clientRectsHash: hashString(JSON.stringify(summary)), ...summary };
      } catch (error) {
        return { error: String(error && error.message || error) };
      } finally {
        if (node && node.parentNode) node.parentNode.removeChild(node);
      }
    };
    const readWebGL = (doc = document) => {
      try {
        const canvas = doc.createElement('canvas');
        const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
        if (!gl) return {};
        const dbg = gl.getExtension('WEBGL_debug_renderer_info');
        const attributes = gl.getContextAttributes() || {};
        const extensions = gl.getSupportedExtensions() || [];
        const precision = gl.getShaderPrecisionFormat(gl.FRAGMENT_SHADER, gl.HIGH_FLOAT);
        const precisionValue = precision ? `${precision.precision}/${precision.rangeMin}/${precision.rangeMax}` : '';
        return {
          vendor: dbg ? gl.getParameter(dbg.UNMASKED_VENDOR_WEBGL) : gl.getParameter(gl.VENDOR),
          renderer: dbg ? gl.getParameter(dbg.UNMASKED_RENDERER_WEBGL) : gl.getParameter(gl.RENDERER),
          version: gl.getParameter(gl.VERSION),
          shadingLanguageVersion: gl.getParameter(gl.SHADING_LANGUAGE_VERSION),
          webgl2: !!canvas.getContext('webgl2'),
          contextAttributes: attributes,
          extensions,
          shaderPrecision: precision ? { precision: precision.precision, rangeMin: precision.rangeMin, rangeMax: precision.rangeMax, fragmentHighFloat: precisionValue } : null,
          limits: {
            maxTextureSize: gl.getParameter(gl.MAX_TEXTURE_SIZE),
            maxCubeMapTextureSize: gl.getParameter(gl.MAX_CUBE_MAP_TEXTURE_SIZE),
            maxRenderbufferSize: gl.getParameter(gl.MAX_RENDERBUFFER_SIZE),
            maxVertexAttribs: gl.getParameter(gl.MAX_VERTEX_ATTRIBS),
            maxVaryingVectors: gl.getParameter(gl.MAX_VARYING_VECTORS),
            maxFragmentUniformVectors: gl.getParameter(gl.MAX_FRAGMENT_UNIFORM_VECTORS),
          },
        };
      } catch (error) {
        return { error: String(error && error.message || error) };
      }
    };
    const readWebGPU = async () => {
      try {
        if (!navigator.gpu || !navigator.gpu.requestAdapter) {
          return { available: false, error: 'navigator.gpu unavailable' };
        }
        const adapter = await navigator.gpu.requestAdapter();
        if (!adapter) {
          return { available: false, error: 'WebGPU adapter unavailable' };
        }
        let info = {};
        try {
          if (adapter.requestAdapterInfo) {
            info = await adapter.requestAdapterInfo();
          } else if (adapter.info) {
            info = adapter.info;
          }
        } catch (error) {
          info = { error: String(error && error.message || error) };
        }
        const limits = {};
        for (const key of ['maxTextureDimension1D', 'maxTextureDimension2D', 'maxTextureArrayLayers', 'maxBindGroups', 'maxBufferSize']) {
          if (adapter.limits && adapter.limits[key] !== undefined) limits[key] = adapter.limits[key];
        }
        return { available: true, info, features: Array.from(adapter.features || []).sort(), limits };
      } catch (error) {
        return { available: false, error: String(error && error.message || error) };
      }
    };
    const readUserAgentData = async (nav = navigator) => {
      let userAgentData = null;
      if (nav.userAgentData) {
        userAgentData = {
          brands: nav.userAgentData.brands,
          mobile: nav.userAgentData.mobile,
          platform: nav.userAgentData.platform,
        };
        try {
          Object.assign(userAgentData, await nav.userAgentData.getHighEntropyValues(['architecture', 'bitness', 'fullVersionList', 'model', 'platformVersion', 'wow64']));
        } catch (error) {
          userAgentData.error = String(error && error.message || error);
        }
        try {
          Object.assign(userAgentData, await nav.userAgentData.getHighEntropyValues(['formFactors']));
        } catch (error) {
          if (!userAgentData.error) userAgentData.formFactorsError = String(error && error.message || error);
        }
      }
      return userAgentData;
    };
    const readNavigator = async () => {
      const userAgentData = await readUserAgentData();
      return {
        userAgent: navigator.userAgent,
        platform: navigator.platform,
        language: navigator.language,
        languages: Array.from(navigator.languages || []),
        hardwareConcurrency: navigator.hardwareConcurrency,
        deviceMemory: navigator.deviceMemory,
        webdriver: navigator.webdriver,
        plugins: Array.from(navigator.plugins || []).map((plugin) => plugin.name),
        mimeTypes: Array.from(navigator.mimeTypes || []).map((mime) => mime.type),
        userAgentData,
        pluginArrayTag: Object.prototype.toString.call(navigator.plugins),
        mimeTypeArrayTag: Object.prototype.toString.call(navigator.mimeTypes),
      };
    };
    const readWebRTC = async () => {
      const PeerConnection = window.RTCPeerConnection || window.webkitRTCPeerConnection;
      if (!PeerConnection) return { available: false, error: 'RTCPeerConnection unavailable' };
      const ipPattern = /\b(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\b/g;
      const privatePattern = /\b(?:10\.|127\.|192\.168\.|172\.(?:1[6-9]|2\d|3[0-1])\.)/;
      return new Promise((resolve) => {
        const candidates = [];
        let settled = false;
        let pc;
        const finish = () => {
          if (settled) return;
          settled = true;
          try {
            if (pc) pc.close();
          } catch {}
          const types = new Set();
          for (const candidate of candidates) {
            const match = candidate.match(/\btyp\s+([a-z0-9]+)/i);
            if (match) types.add(match[1].toLowerCase());
          }
          resolve({
            available: true,
            candidateCount: candidates.length,
            candidateTypes: Array.from(types).sort(),
            hasHostCandidate: types.has('host'),
            hasSrflxCandidate: types.has('srflx'),
            hasRelayCandidate: types.has('relay'),
            hasPrivateAddress: candidates.some((candidate) => privatePattern.test(candidate)),
            hasMdnsHostCandidate: candidates.some((candidate) => /\.local\b/i.test(candidate)),
            redactedCandidates: candidates.map((candidate) => candidate.replace(ipPattern, '<ip>')),
          });
        };
        try {
          pc = new PeerConnection({ iceServers: [] });
          pc.createDataChannel('browseforge-detector');
          pc.onicecandidate = (event) => {
            if (event.candidate && event.candidate.candidate) {
              candidates.push(event.candidate.candidate);
            } else {
              finish();
            }
          };
          pc.createOffer().then((offer) => pc.setLocalDescription(offer)).catch((error) => {
            resolve({ available: false, error: String(error && error.message || error) });
          });
          setTimeout(finish, 1500);
        } catch (error) {
          try {
            if (pc) pc.close();
          } catch {}
          resolve({ available: false, error: String(error && error.message || error) });
        }
      });
    };
    const readRealmSnapshot = async (name, win) => {
      const nav = win.navigator;
      const webgl = win.document ? readWebGL(win.document) : {};
      return {
        name,
        userAgent: nav.userAgent,
        platform: nav.platform,
        language: nav.language,
        languages: Array.from(nav.languages || []),
        hardwareConcurrency: nav.hardwareConcurrency,
        deviceMemory: nav.deviceMemory,
        userAgentData: await readUserAgentData(nav),
        intlTimeZone: win.Intl.DateTimeFormat().resolvedOptions().timeZone,
        intlLocale: win.Intl.DateTimeFormat().resolvedOptions().locale,
        timezoneOffset: new win.Date().getTimezoneOffset(),
        devicePixelRatio: win.devicePixelRatio,
        webglVendor: webgl.vendor,
        webglRenderer: webgl.renderer,
        canvasDataURL: win.document ? canvasFingerprint(win.document) : undefined,
      };
    };
    const realmTimeoutMs = 3000;
    const timeoutRealm = (name, cleanup) => new Promise((resolve) => {
      setTimeout(() => {
        try {
          if (cleanup) cleanup();
        } catch {}
        resolve({ name, unsupported: true, error: `realm timed out after ${realmTimeoutMs}ms` });
      }, realmTimeoutMs);
    });
    const iframeRealm = async (name, sandbox, nested, src = 'about:blank') => {
      let iframe;
      const sample = new Promise((resolve) => {
        iframe = document.createElement('iframe');
        if (sandbox) iframe.setAttribute('sandbox', sandbox);
        iframe.src = src;
        iframe.onload = async () => {
          try {
            let win = iframe.contentWindow;
            if (nested) {
              const child = win.document.createElement('iframe');
              child.src = 'about:blank';
              win.document.documentElement.appendChild(child);
              win = child.contentWindow;
            }
            resolve(await readRealmSnapshot(name, win));
          } catch (error) {
            resolve({ name, error: String(error && error.message || error) });
          } finally {
            iframe.remove();
          }
        };
        iframe.onerror = () => {
          iframe.remove();
          resolve({ name, unsupported: true });
        };
        document.documentElement.appendChild(iframe);
      });
      return Promise.race([sample, timeoutRealm(name, () => iframe && iframe.remove())]);
    };
    const detachedIframeRealm = async () => {
      let iframe;
      const sample = new Promise((resolve) => {
        iframe = document.createElement('iframe');
        iframe.src = 'about:blank';
        iframe.onload = () => {
          iframe.remove();
          resolve({ name: 'detached-iframe', unsupported: true, error: 'detached browsing context is not stable enough for contract comparison' });
        };
        iframe.onerror = () => {
          iframe.remove();
          resolve({ name: 'detached-iframe', unsupported: true });
        };
        document.documentElement.appendChild(iframe);
      });
      return Promise.race([sample, timeoutRealm('detached-iframe', () => iframe && iframe.remove())]);
    };
    const dedicatedWorkerRealm = async () => {
      let worker;
      const sample = new Promise((resolve) => {
        try {
          const code = `
            self.onmessage = async () => {
              let userAgentData = null;
              if (navigator.userAgentData) {
                userAgentData = {
                  brands: navigator.userAgentData.brands,
                  mobile: navigator.userAgentData.mobile,
                  platform: navigator.userAgentData.platform,
                };
                try {
                  Object.assign(userAgentData, await navigator.userAgentData.getHighEntropyValues(['architecture', 'bitness', 'fullVersionList', 'model', 'platformVersion', 'wow64']));
                } catch (error) {
                  userAgentData.error = String(error && error.message || error);
                }
                try {
                  Object.assign(userAgentData, await navigator.userAgentData.getHighEntropyValues(['formFactors']));
                } catch (error) {
                  if (!userAgentData.error) userAgentData.formFactorsError = String(error && error.message || error);
                }
              }
              let webglVendor, webglRenderer, offscreenCanvasWebGL = false;
              try {
                if (typeof OffscreenCanvas !== "undefined") {
                  const gl = new OffscreenCanvas(1, 1).getContext("webgl");
                  offscreenCanvasWebGL = !!gl;
                  if (gl) {
                    const dbg = gl.getExtension("WEBGL_debug_renderer_info");
                    webglVendor = dbg ? gl.getParameter(dbg.UNMASKED_VENDOR_WEBGL) : gl.getParameter(gl.VENDOR);
                    webglRenderer = dbg ? gl.getParameter(dbg.UNMASKED_RENDERER_WEBGL) : gl.getParameter(gl.RENDERER);
                  }
                }
              } catch (e) {}
              self.postMessage({
                name: "dedicated-worker",
                userAgent: navigator.userAgent,
                platform: navigator.platform,
                language: navigator.language,
                languages: Array.from(navigator.languages || []),
                hardwareConcurrency: navigator.hardwareConcurrency,
                deviceMemory: navigator.deviceMemory,
                userAgentData,
                intlTimeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
                intlLocale: Intl.DateTimeFormat().resolvedOptions().locale,
                timezoneOffset: new Date().getTimezoneOffset(),
                webglVendor,
                webglRenderer,
                offscreenCanvasWebGL,
              });
            };
          `;
          worker = new Worker(URL.createObjectURL(new Blob([code], { type: 'application/javascript' })));
          worker.onmessage = (event) => { worker.terminate(); resolve(event.data); };
          worker.onerror = (event) => { worker.terminate(); resolve({ name: 'dedicated-worker', error: String(event.message || event) }); };
          worker.postMessage(null);
        } catch (error) {
          resolve({ name: 'dedicated-worker', error: String(error && error.message || error) });
        }
      });
      return Promise.race([sample, timeoutRealm('dedicated-worker', () => worker && worker.terminate())]);
    };
    const sharedWorkerRealm = async () => {
      let worker;
      const sample = new Promise((resolve) => {
        try {
          if (typeof SharedWorker === 'undefined') return resolve({ name: 'shared-worker', unsupported: true });
          const code = `
            self.onconnect = (event) => {
              const port = event.ports[0];
              const send = async () => {
                let userAgentData = null;
                if (navigator.userAgentData) {
                  userAgentData = {
                    brands: navigator.userAgentData.brands,
                    mobile: navigator.userAgentData.mobile,
                    platform: navigator.userAgentData.platform,
                  };
                  try {
                    Object.assign(userAgentData, await navigator.userAgentData.getHighEntropyValues(['architecture', 'bitness', 'fullVersionList', 'model', 'platformVersion', 'wow64']));
                  } catch (error) {
                    userAgentData.error = String(error && error.message || error);
                  }
                  try {
                    Object.assign(userAgentData, await navigator.userAgentData.getHighEntropyValues(['formFactors']));
                  } catch (error) {
                    if (!userAgentData.error) userAgentData.formFactorsError = String(error && error.message || error);
                  }
                }
                port.postMessage({
                  name: "shared-worker",
                  userAgent: navigator.userAgent,
                  platform: navigator.platform,
                  language: navigator.language,
                  languages: Array.from(navigator.languages || []),
                  hardwareConcurrency: navigator.hardwareConcurrency,
                  deviceMemory: navigator.deviceMemory,
                  userAgentData,
                  intlTimeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
                  intlLocale: Intl.DateTimeFormat().resolvedOptions().locale,
                  timezoneOffset: new Date().getTimezoneOffset(),
                });
              };
              send();
            };
          `;
          worker = new SharedWorker(URL.createObjectURL(new Blob([code], { type: 'application/javascript' })));
          worker.port.onmessage = (event) => resolve(event.data);
          worker.port.start();
        } catch (error) {
          resolve({ name: 'shared-worker', error: String(error && error.message || error) });
        }
      });
      return Promise.race([sample, timeoutRealm('shared-worker', () => worker && worker.port && worker.port.close())]);
    };
    const dedicatedWorker = await dedicatedWorkerRealm();
    const sharedWorker = await sharedWorkerRealm();
    const safeFeature = (name, read) => {
      try {
        return read();
      } catch (error) {
        return { unavailable: true, error: String(error && error.message || error), name };
      }
    };
    const readPermissionState = async (descriptor) => {
      try {
        if (!navigator.permissions || !navigator.permissions.query) return '';
        const status = await navigator.permissions.query(descriptor);
        return status && status.state ? status.state : '';
      } catch (error) {
        return '';
      }
    };
    return {
      url: location.href,
      title: document.title,
      navigator: await readNavigator(),
      intl: { timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone, locale: Intl.DateTimeFormat().resolvedOptions().locale },
      date: { timezoneOffset: new Date().getTimezoneOffset() },
      hardware: { hardwareConcurrency: navigator.hardwareConcurrency, deviceMemory: navigator.deviceMemory },
      screen: {
        width: screen.width,
        height: screen.height,
        availWidth: screen.availWidth,
        availHeight: screen.availHeight,
        devicePixelRatio,
        outerWidth: window.outerWidth,
        outerHeight: window.outerHeight,
        innerWidth: window.innerWidth,
        innerHeight: window.innerHeight,
        viewportWidth: document.documentElement ? document.documentElement.clientWidth : window.innerWidth,
        viewportHeight: document.documentElement ? document.documentElement.clientHeight : window.innerHeight,
        colorDepth: screen.colorDepth,
        touchPoints: navigator.maxTouchPoints || 0,
        orientation: screen.orientation && screen.orientation.type ? screen.orientation.type : '',
      },
      webgl: readWebGL(),
      webgpu: await readWebGPU(),
      webrtc: await readWebRTC(),
      canvas: { dataURL: canvasFingerprint() },
      math: readMathFingerprint(),
      geometry: readClientRects(),
      audio: await readAudioFingerprint(),
      fonts: readFonts(),
      storage: {
        cookiesEnabled: navigator.cookieEnabled,
        localStorage: safeFeature('localStorage', () => typeof window.localStorage !== 'undefined'),
        sessionStorage: safeFeature('sessionStorage', () => typeof window.sessionStorage !== 'undefined'),
        indexedDB: safeFeature('indexedDB', () => typeof window.indexedDB !== 'undefined'),
        persisted: navigator.storage && navigator.storage.persisted ? await navigator.storage.persisted().catch((error) => ({ error: String(error && error.message || error) })) : null,
        storageEstimate: navigator.storage && navigator.storage.estimate ? await navigator.storage.estimate().catch((error) => ({ error: String(error && error.message || error) })) : null,
      },
      geolocation: {
        available: typeof navigator.geolocation !== 'undefined',
        permissionState: await readPermissionState({ name: 'geolocation' }),
      },
      media: (() => {
        const video = document.createElement('video');
        return {
          mediaDevices: !!navigator.mediaDevices,
          enumerateDevices: !!(navigator.mediaDevices && navigator.mediaDevices.enumerateDevices),
          h264Support: video.canPlayType('video/mp4; codecs="avc1.42E01E"'),
          vp9Support: video.canPlayType('video/webm; codecs="vp9"'),
          av1Support: video.canPlayType('video/mp4; codecs="av01.0.05M.08"'),
        };
      })(),
      automation: {
        chromeRuntimeShape: !!(window.chrome && typeof window.chrome === 'object'),
        chromeKeys: window.chrome && typeof window.chrome === 'object' ? Object.keys(window.chrome).sort() : [],
        notificationPermission: typeof Notification === 'undefined' ? 'unsupported' : Notification.permission,
        notificationPermissionState: await readPermissionState({ name: 'notifications' }),
        automationGlobals: Object.keys(window).filter((key) => /^(__webdriver|__driver|__selenium|__fxdriver|cdc_|ret_nodes)|selenium|webdriver|chromedriver/i.test(key)),
        webdriverAttributes: Array.from(document.documentElement.attributes || []).map((attr) => attr.name).filter((name) => /webdriver|selenium|driver/i.test(name)),
      },
      workers: {
        dedicated: dedicatedWorker,
        shared: sharedWorker,
        offscreenCanvasWebGL: dedicatedWorker.offscreenCanvasWebGL === true,
        serviceWorkerAvailable: !!navigator.serviceWorker,
        serviceWorkerController: !!(navigator.serviceWorker && navigator.serviceWorker.controller),
        serviceWorkerControllerState: navigator.serviceWorker && navigator.serviceWorker.controller ? navigator.serviceWorker.controller.state : '',
      },
      realms: [
        await iframeRealm('same-origin-iframe', '', false),
        await iframeRealm('sandbox-iframe', 'allow-scripts allow-same-origin', false),
        await iframeRealm('fragment-iframe', '', false, 'about:blank#browseforge-fragment'),
        await iframeRealm('nested-iframe', '', true),
        await detachedIframeRealm(),
        dedicatedWorker,
        sharedWorker,
      ],
    };
  }.toString()})()`;
}

function runSelfTest() {
  const contract = {
    browser: {
      user_agent: 'Mozilla/5.0 Chrome/150',
      brand_versions: [{ brand: 'Chromium', version: '150' }],
      full_version_list: [{ brand: 'Chromium', version: '150.0.0.0' }],
      client_hints: {
        sec_ch_ua: '"Chromium";v="150"',
        sec_ch_ua_full_version_list: '"Chromium";v="150.0.0.0"',
        platform: 'Linux',
        platform_version: '',
        architecture: 'arm',
        bitness: '64',
        model: '',
        form_factors: ['Desktop'],
      },
    },
    locale: { accept_language: 'en-US,en', sec_ch_lang: 'en-US,en' },
    screen: {
      width: 1920,
      height: 1080,
      avail_width: 1920,
      avail_height: 1040,
      dpr: 1,
      outer_width: 1920,
      outer_height: 1040,
      inner_width: 1920,
      inner_height: 948,
      viewport_width: 1920,
      viewport_height: 948,
      color_depth: 24,
      touch_points: 0,
      orientation: 'landscape-primary',
    },
    permissions: { notification: 'prompt' },
    plugins: { pdf_viewer: true },
    media: { h264: true, vp9: true, av1: true, devices: ['default-camera'] },
    audio: { sample_rate: 48000, stable: true },
    fonts: { families: ['Noto Sans', 'Noto Color Emoji', 'Noto Sans CJK TC'], emoji: 'Noto Color Emoji', cjk: true },
    canvas: { stable: true, text_metrics_mode: 'stable-profile' },
    math: { stable: true },
    geometry: { client_rects_stable: true },
    gpu: {
      worker_offscreen_canvas: true,
      webgl: true,
      webgl2: true,
      webgpu: 'browser-default',
      extensions: ['WEBGL_debug_renderer_info', 'OES_texture_float'],
      limits: { maxTextureSize: 16384 },
      shader_precision: { fragmentHighFloat: '23/127/127' },
    },
    storage: {
      quota_mb: 4096,
      persistent: true,
      cookies: 'profile-persistent',
      local_storage: 'profile-persistent',
      session_storage: 'session-scoped',
      indexed_db: 'profile-persistent',
      quota_behavior: 'chromium-profile-quota',
    },
    webrtc: { mode: 'disable_non_proxied_udp', direct_ip_redaction: true },
    geolocation: { mode: 'proxy-aligned', country_code: 'US', region_code: 'NY' },
    realms: { targets: ['same-origin-iframe', 'nested-iframe', 'dedicated-worker', 'shared-worker', 'service-worker', 'offscreen-canvas-worker'] },
  };
  const sample = {
    navigator: {
      userAgent: 'Mozilla/5.0 HeadlessChrome/120',
      webdriver: true,
      languages: [],
      plugins: [],
      mimeTypes: [],
      pluginArrayTag: '[object Array]',
      mimeTypeArrayTag: '[object Array]',
      userAgentData: {
        brands: [{ brand: 'Chromium', version: '150' }],
        mobile: false,
        platform: 'Linux',
        architecture: 'x86',
        bitness: '64',
        fullVersionList: [{ brand: 'Chromium', version: '150.0.0.0' }],
        platformVersion: '',
        model: '',
        formFactors: ['Mobile'],
      },
    },
    headers: {
      userAgent: 'Mozilla/5.0 HeadlessChrome/150',
      acceptLanguage: 'zh-TW,zh',
      secCHUA: '"Not Chromium";v="150"',
      secCHUAFullVersionList: '"Not Chromium";v="150.0.0.0"',
      secCHUAPlatform: 'Windows',
      secCHUAArch: 'x86',
      secCHUAMobile: '?1',
      secCHUAFormFactors: 'Mobile',
      secCHLang: 'zh-TW,zh',
    },
    automation: {
      chromeRuntimeShape: false,
      notificationPermission: 'denied',
      notificationPermissionState: 'denied',
      automationGlobals: ['cdc_adoQpoasnfa76pfcZLmcfl_Array'],
      webdriverAttributes: ['webdriver'],
    },
    screen: {
      width: 1366,
      height: 768,
      availWidth: 1366,
      availHeight: 728,
      devicePixelRatio: 2,
      outerWidth: 1200,
      outerHeight: 700,
      innerWidth: 1180,
      innerHeight: 640,
      viewportWidth: 1170,
      viewportHeight: 620,
      colorDepth: 30,
      touchPoints: 5,
      orientation: 'portrait-primary',
    },
    media: { mediaDevices: false, enumerateDevices: false, h264Support: '', vp9Support: '', av1Support: '' },
    audio: { sampleRate: 44100, hash: '' },
    math: { hash: '' },
    geometry: { clientRectsHash: '' },
    fonts: {
      available: { 'Noto Sans': false, 'Noto Color Emoji': false, 'Noto Sans CJK TC': false },
      cjkAvailable: false,
      metricsHash: '',
    },
    webgl: {
      webgl2: false,
      extensions: ['OES_texture_float'],
      limits: { maxTextureSize: 4096 },
      shaderPrecision: { fragmentHighFloat: '10/15/15' },
    },
    workers: { offscreenCanvasWebGL: false, serviceWorkerAvailable: false, serviceWorkerController: false },
    geolocation: { available: false, permissionState: 'denied' },
    webrtc: { available: true, hasHostCandidate: true, hasSrflxCandidate: true, hasPrivateAddress: true },
    storage: {
      cookiesEnabled: false,
      localStorage: false,
      sessionStorage: false,
      indexedDB: false,
      persisted: false,
      storageEstimate: { quota: 1024 * 1024 },
    },
    realms: [
      { name: 'optional-detached-iframe', unsupported: true, error: 'detached access blocked' },
      {
        name: 'shared-worker',
        userAgent: 'Mozilla/5.0 SharedWorkerMismatch',
        userAgentData: {
          brands: [{ brand: 'Chromium', version: '150' }],
          mobile: false,
          platform: 'Linux',
          architecture: 'arm',
          fullVersionList: [{ brand: 'Chromium', version: '150.0.0.0' }],
          platformVersion: '',
          model: '',
          formFactors: ['Mobile'],
        },
      },
    ],
  };
  const result = compareContract(contract, sample);
  const fields = new Set(result.mismatches.map((mismatch) => mismatch.field));
  for (const field of [
    'automation.userAgent',
    'automation.webdriver',
    'automation.automationGlobals',
    'navigator.languages',
    'navigator.plugins',
    'navigator.mimeTypes',
    'automation.chromeRuntimeShape',
    'automation.webdriverAttributes',
    'automation.notificationPermission',
    'automation.notificationPermissionState',
    'geolocation.available',
    'geolocation.permissionState',
    'navigator.userAgentData.architecture',
    'navigator.userAgentData.formFactors',
    'headers.userAgent',
    'headers.acceptLanguage',
    'headers.secCHUA',
    'headers.secCHUAFullVersionList',
    'headers.secCHUAPlatform',
    'headers.secCHUAArch',
    'headers.secCHUAMobile',
    'headers.secCHUAFormFactors',
    'headers.secCHLang',
    'screen.width',
    'screen.height',
    'screen.availWidth',
    'screen.availHeight',
    'screen.devicePixelRatio',
    'screen.outerWidth',
    'screen.outerHeight',
    'screen.innerWidth',
    'screen.innerHeight',
    'screen.viewportWidth',
    'screen.viewportHeight',
    'screen.colorDepth',
    'screen.touchPoints',
    'screen.orientation',
    'audio.sampleRate',
    'audio.hash',
    'canvas.dataURL',
    'math.hash',
    'geometry.clientRectsHash',
    'fonts.families',
    'fonts.emoji',
    'fonts.cjk',
    'fonts.metricsHash',
    'media.h264Support',
    'media.vp9Support',
    'media.av1Support',
    'media.mediaDevices',
    'media.enumerateDevices',
    'workers.offscreenCanvasWebGL',
    'webgl.webgl2',
    'webgpu.available',
    'webrtc.hostCandidates',
    'webrtc.srflxCandidates',
    'webrtc.privateAddress',
    'webgl.extensions',
    'webgl.limits.maxTextureSize',
    'webgl.shaderPrecision.fragmentHighFloat',
    'storage.cookiesEnabled',
    'storage.localStorage',
    'storage.sessionStorage',
    'storage.indexedDB',
    'storage.persisted',
    'storage.storageEstimate.quota',
    'realms.same-origin-iframe',
    'realms.nested-iframe',
    'realms.dedicated-worker',
    'realms.service-worker',
    'realms.shared-worker.userAgent',
    'realms.shared-worker.userAgentData.architecture',
    'realms.shared-worker.userAgentData.bitness',
  ]) {
    if (!fields.has(field)) {
      throw new Error(`selftest missing mismatch for ${field}`);
    }
  }
  if (fields.has('realms.optional-detached-iframe')) {
    throw new Error('selftest should skip unsupported optional realms');
  }
  return result;
}

function usage() {
  console.error('Usage: node scripts/detector-harness.js compare <persona.json> <sample.json>');
  console.error('       node scripts/detector-harness.js collector');
  console.error('       node scripts/detector-harness.js matrix');
  console.error('       node scripts/detector-harness.js selftest');
}

if (require.main === module) {
  const [cmd, ...args] = process.argv.slice(2);
  if (cmd === 'collector') {
    process.stdout.write(browserCollectorSource() + '\n');
  } else if (cmd === 'compare' && args.length === 2) {
    const contract = JSON.parse(fs.readFileSync(args[0], 'utf8'));
    const sample = JSON.parse(fs.readFileSync(args[1], 'utf8'));
    const result = compareContract(contract, sample);
    process.stdout.write(JSON.stringify(result, null, 2) + '\n');
    process.exitCode = result.ok ? 0 : 1;
  } else if (cmd === 'matrix') {
    process.stdout.write(JSON.stringify({ criteria: STABLE_RESULT_CRITERIA, targets: DETECTOR_MATRIX }, null, 2) + '\n');
  } else if (cmd === 'selftest') {
    const result = runSelfTest();
    process.stdout.write(JSON.stringify({ ok: true, mismatchCount: result.mismatches.length }, null, 2) + '\n');
  } else {
    usage();
    process.exitCode = 2;
  }
}

module.exports = { CONTRACT_FIELDS, OPTIONAL_CONTRACT_FIELDS, DETECTOR_MATRIX, STABLE_RESULT_CRITERIA, browserCollectorSource, compareContract, runSelfTest };
