// BrowseForge Fingerprint Injector (content script)
// Runs at document_start in ISOLATED world
// Legacy Firefox-container experiment. The current primary runtime launches one
// isolated browser process per profile and applies native fingerprinting through
// CAMOU_CONFIG for Camoufox or fingerprint_seed/native flags for CloakBrowser.

(async () => {
  try {
    // Get current container's cookieStoreId
    const cookieStoreId = await getCookieStoreId();
    if (!cookieStoreId || cookieStoreId === 'firefox-default') return;

    // Load fingerprint config for this container
    const key = `fp_${cookieStoreId}`;
    const stored = await browser.storage.local.get(key);
    const fp = stored[key];
    if (!fp) return;

    // Call optional Firefox/Camoufox setters when a fork exposes them.
    const w = window.wrappedJSObject;

    if (fp['canvas:seed'] && w.setCanvasSeed)
      w.setCanvasSeed(fp['canvas:seed']);

    if (fp['audio:seed'] && w.setAudioFingerprintSeed)
      w.setAudioFingerprintSeed(fp['audio:seed']);

    if (fp['fonts:spacing_seed'] && w.setFontSpacingSeed)
      w.setFontSpacingSeed(fp['fonts:spacing_seed']);

    if (fp['navigator.platform'] && w.setNavigatorPlatform)
      w.setNavigatorPlatform(fp['navigator.platform']);

    if (fp['navigator.userAgent'] && w.setNavigatorUserAgent)
      w.setNavigatorUserAgent(fp['navigator.userAgent']);

    if (fp['navigator.oscpu'] && w.setNavigatorOscpu)
      w.setNavigatorOscpu(fp['navigator.oscpu']);

    if (fp['webGl:vendor'] && w.setWebGLVendor)
      w.setWebGLVendor(fp['webGl:vendor']);

    if (fp['webGl:renderer'] && w.setWebGLRenderer)
      w.setWebGLRenderer(fp['webGl:renderer']);

    if (fp['screen.width'] && fp['screen.height'] && w.setScreenDimensions)
      w.setScreenDimensions(fp['screen.width'], fp['screen.height']);

    if (fp['screen.colorDepth'] && w.setScreenColorDepth)
      w.setScreenColorDepth(fp['screen.colorDepth']);

    if (fp['timezone'] && w.setTimezone)
      w.setTimezone(fp['timezone']);

    if (fp['webrtc:ipv4'] && w.setWebRTCIPv4)
      w.setWebRTCIPv4(fp['webrtc:ipv4']);

    if (fp['fonts'] && w.setFontList) {
      const fontList = Array.isArray(fp['fonts']) ? fp['fonts'].join(',') : fp['fonts'];
      w.setFontList(fontList);
    }

    if (fp['voices'] && w.setSpeechVoices) {
      const voiceList = Array.isArray(fp['voices']) ? fp['voices'].join(',') : fp['voices'];
      w.setSpeechVoices(voiceList);
    }

  } catch (e) {
    // Silent fail — don't break page loading
  }
})();

async function getCookieStoreId() {
  try {
    const tab = await browser.tabs.getCurrent();
    return tab?.cookieStoreId;
  } catch {
    // content script may not have access, fallback to message
    try {
      return await browser.runtime.sendMessage({ type: 'get_cookie_store_id' });
    } catch {
      return null;
    }
  }
}
