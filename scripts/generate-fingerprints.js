#!/usr/bin/env node
// Generate fingerprint pool using fingerprint-suite's Bayesian network
// Usage: node generate-fingerprints.js --browser firefox --os windows --count 500

const { FingerprintGenerator } = require('fingerprint-generator');
const fs = require('fs');
const path = require('path');

const args = process.argv.slice(2);
const getArg = (name, def) => {
  const i = args.indexOf(`--${name}`);
  return i >= 0 && args[i + 1] ? args[i + 1] : def;
};

const browser = getArg('browser', 'firefox');
const os = getArg('os', 'windows');
const count = parseInt(getArg('count', '500'));
const outputDir = getArg('output', 'data');

// Load WebGL profiles for complete fingerprinting
const webglProfilesPath = path.join(__dirname, '..', 'data', 'webgl-profiles', 'webgl-profiles.json');
let webglProfiles = [];
try { webglProfiles = JSON.parse(fs.readFileSync(webglProfilesPath, 'utf8')); } catch {}

function findWebGLProfile(renderer) {
  if (!renderer) return null;
  return webglProfiles.find(p => renderer.includes(p.match)) || null;
}

console.log(`Generating ${count} ${browser}/${os} fingerprints (${webglProfiles.length} WebGL profiles loaded)...`);

const generator = new FingerprintGenerator({ browsers: [browser], operatingSystems: [os] });
const fingerprints = [];

for (let i = 0; i < count; i++) {
  const { fingerprint, headers } = generator.getFingerprint();

  // Convert to Camoufox flat format
  const camoufoxConfig = {
    'navigator.userAgent': fingerprint.navigator.userAgent,
    'navigator.platform': fingerprint.navigator.platform,
    'navigator.hardwareConcurrency': fingerprint.navigator.hardwareConcurrency,
    'navigator.language': fingerprint.navigator.language,
    'navigator.languages': fingerprint.navigator.languages,
    'navigator.oscpu': fingerprint.navigator.oscpu || '',
    'screen.width': fingerprint.screen.width,
    'screen.height': fingerprint.screen.height,
    'screen.availWidth': fingerprint.screen.availWidth,
    'screen.availHeight': fingerprint.screen.availHeight,
    'screen.colorDepth': fingerprint.screen.colorDepth,
    'window.outerWidth': fingerprint.screen.outerWidth,
    'window.outerHeight': fingerprint.screen.outerHeight,
    'window.innerWidth': fingerprint.screen.innerWidth,
    'window.innerHeight': fingerprint.screen.innerHeight,
    'window.devicePixelRatio': fingerprint.screen.devicePixelRatio,
    'webGl:vendor': fingerprint.videoCard?.vendor || '',
    'webGl:renderer': fingerprint.videoCard?.renderer || '',
    'canvas:seed': Math.floor(Math.random() * 4294967295),
    'audio:seed': Math.floor(Math.random() * 4294967295),
    'fonts:spacing_seed': Math.floor(Math.random() * 4294967295),
    'headers.User-Agent': headers['user-agent'] || fingerprint.navigator.userAgent,
    'headers.Accept-Language': headers['accept-language'] || `${fingerprint.navigator.language},en;q=0.9`,
    '_meta': { browser, os, generated: new Date().toISOString() },
  };

  // Attach complete WebGL profile if available for this GPU
  const webglProfile = findWebGLProfile(camoufoxConfig['webGl:renderer']);
  if (webglProfile) {
    const { match, ...webglData } = webglProfile;
    Object.assign(camoufoxConfig, webglData);
  }

  fingerprints.push(camoufoxConfig);
}

fs.mkdirSync(outputDir, { recursive: true });
const outPath = path.join(outputDir, `fingerprints-${browser}-${os}.json`);
fs.writeFileSync(outPath, JSON.stringify(fingerprints, null, 2));
console.log(`Written ${fingerprints.length} fingerprints to ${outPath}`);
