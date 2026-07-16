package browser

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"browseforge/internal/config"
	"browseforge/internal/fingerprint"
	"browseforge/internal/profile"
	bfruntime "browseforge/internal/runtime"

	"github.com/mxschmitt/playwright-go"
)

const chromiumAutomationControlledArg = "--disable-blink-features=AutomationControlled"

const chromiumWebRTCIPHandlingArg = "--force-webrtc-ip-handling-policy=disable_non_proxied_udp"

const chromiumAutomationMitigationInitScript = `(() => {
	const defineGetter = (target, prop, getter) => {
		try {
			Object.defineProperty(target, prop, { get: getter, configurable: true });
		} catch (_) {}
	};
	defineGetter(Navigator.prototype, "webdriver", () => undefined);
	try { delete window.navigator.webdriver; } catch (_) {}
	for (const key of Object.keys(window)) {
		if (/^(cdc_|__webdriver|__driver_evaluate|__webdriver_script_fn|__selenium)/.test(key)) {
			try { delete window[key]; } catch (_) {}
		}
	}
	if (!("chrome" in window)) {
		try {
			Object.defineProperty(window, "chrome", {
				value: { runtime: {} },
				configurable: true,
				enumerable: true,
				writable: false
			});
		} catch (_) {}
	}
})();`

func (m *Manager) launchChromium(p *profile.Profile) (*Session, error) {
	desc, err := m.runtimes.ResolveProfile(p)
	if err != nil {
		return nil, err
	}
	chromiumPath := desc.BinaryPath
	if chromiumPath == "" {
		return nil, fmt.Errorf("runtimes.%s.binary_path is not configured", desc.ID)
	}

	userDataDir, err := filepath.Abs(filepath.Join(p.ProfileDir, "browser-data"))
	if err != nil {
		return nil, fmt.Errorf("browser data path: %w", err)
	}
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return nil, fmt.Errorf("create browser data dir: %w", err)
	}
	cleanProfileLocks(userDataDir)

	absChromiumPath, err := filepath.Abs(chromiumPath)
	if err != nil {
		return nil, fmt.Errorf("%s path: %w", desc.ID, err)
	}

	args := []string{
		"--no-first-run",
		"--test-type",
		chromiumAutomationControlledArg,
		chromiumWebRTCIPHandlingArg,
	}
	if p.FingerprintSeed > 0 {
		args = append(args, fmt.Sprintf("--fingerprint=%d", p.FingerprintSeed))
	}

	policy := m.cfg.ChromiumRuntimeSettings(string(desc.ID))
	if quota := cloakStorageQuotaMB(policy); quota < 0 {
		return nil, fmt.Errorf("%s storage_quota_mb must be >= 0", desc.ID)
	}

	var tz, locale, proxyRegion string
	geoResult := fingerprint.GeoDetectionResult{}
	effectiveProxy := m.effectiveProxy(p)
	if effectiveProxy.Proxy != nil {
		proxy := effectiveProxy.Proxy
		if desc.ID == bfruntime.BrowseForgeChromium {
			proxyRegion, err = sanitizeBrowseForgeProxyRegion(proxy.Region)
			if err != nil {
				return nil, err
			}
		} else {
			proxyRegion = strings.TrimSpace(proxy.Region)
		}
		geoResult = fingerprint.DetectProxyGeo(proxy.Type, proxy.Host, proxy.Port, proxy.Username, proxy.Password)
		tz, locale = geoResult.Values()
		if tz == "" {
			tz, locale = fallbackGeoForProxyRegion(proxyRegion)
			geoResult = fingerprint.GeoDetectionResult{Timezone: tz, Locale: locale, Source: "proxy_region_fallback", Status: "geo_provider_unavailable"}
		}
		args = append(args, "--fingerprint-webrtc-ip=auto")
	} else {
		geoResult = fingerprint.DetectLocalGeo()
		tz, locale = geoResult.Values()
	}
	platform, err := resolveCloakFingerprintPlatform(policy, runtime.GOOS)
	if err != nil {
		return nil, err
	}
	if desc.ID == bfruntime.BrowseForgeChromium {
		platform, err = resolveChromiumFingerprintPlatform(policy, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return nil, err
		}
	}
	persona, err := buildChromiumLaunchPersona(p, desc.ID, platform, tz, locale, proxyRegion, runtime.GOARCH, policy)
	if err != nil {
		return nil, err
	}
	persona.Native.Locale.GeoSource = geoResult.Source
	persona.Native.Locale.GeoStatus = geoResult.Status
	if desc.ID == bfruntime.BrowseForgeChromium {
		args = appendChromiumLaunchPersonaArgs(args, persona)
		args = append(args, browseForgeChromiumWindowArgs(persona)...)
		nativeConfigPath, err := writeBrowseForgeNativeConfig(userDataDir, persona)
		if err != nil {
			return nil, err
		}
		args = append(args,
			"--browseforge-stealth-config="+nativeConfigPath,
			"--browseforge-stealth-mode="+browseForgeChromiumNativeMode(policy),
		)
	} else {
		args = append(args,
			"--fingerprint-timezone="+persona.Native.Locale.Timezone,
			"--fingerprint-locale="+persona.Native.Locale.Locale,
			"--fingerprint-platform="+persona.NavigatorPlatform,
		)
		if quota := cloakStorageQuotaMB(policy); quota > 0 {
			args = append(args, fmt.Sprintf("--fingerprint-storage-quota=%d", quota))
		}
		if persona.PluginsPDF != "" {
			args = append(args, "--fingerprint-plugins-pdf="+persona.PluginsPDF)
		}
		args = appendProfileFingerprintArgs(args, p.Fingerprint, persona.Native.Browser.UserAgent, persona.NavigatorPlatform, persona.Native.Locale.AcceptLanguage)
	}

	if m.cfg.NoSandbox {
		args = append(args, "--no-sandbox")
	}

	fontsDir, err := resolveCloakFontsDir(policy)
	if err != nil {
		return nil, err
	}
	if err := validateCloakFingerprintPolicy(policy, platform, runtime.GOOS); err != nil {
		return nil, err
	}
	if fontsDir != "" {
		args = append(args, "--fingerprint-fonts-dir="+fontsDir)
	}

	baseArgs := append([]string(nil), args...)
	args, err = applyCloakBrowserLaunchPolicy(baseArgs, userDataDir, policy, false)
	if err != nil {
		return nil, err
	}

	downloadsDir, err := filepath.Abs(filepath.Join(p.ProfileDir, "downloads"))
	if err != nil {
		return nil, fmt.Errorf("downloads path: %w", err)
	}
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return nil, fmt.Errorf("create downloads dir: %w", err)
	}

	prefsDir := filepath.Join(userDataDir, "Default")
	if err := os.MkdirAll(prefsDir, 0755); err != nil {
		return nil, fmt.Errorf("create chromium prefs dir: %w", err)
	}
	prefsPath := filepath.Join(prefsDir, "Preferences")
	prefs := map[string]any{}
	if data, err := os.ReadFile(prefsPath); err == nil {
		if err := json.Unmarshal(data, &prefs); err != nil {
			return nil, fmt.Errorf("decode chromium preferences: %w", err)
		}
	}
	prefs["savefile"] = map[string]any{"default_directory": downloadsDir}
	prefs["download"] = map[string]any{"default_directory": downloadsDir, "prompt_for_download": false}
	prefs["webrtc"] = map[string]any{
		"ip_handling_policy":      "disable_non_proxied_udp",
		"multiple_routes_enabled": false,
		"nonproxied_udp_enabled":  false,
	}
	out, err := json.Marshal(prefs)
	if err != nil {
		return nil, fmt.Errorf("encode chromium preferences: %w", err)
	}
	if err := os.WriteFile(prefsPath, out, 0644); err != nil {
		return nil, fmt.Errorf("write chromium preferences: %w", err)
	}

	ignoreArgs := []string{
		"--enable-automation",
		"--host-resolver-rules=MAP * ~NOTFOUND , EXCLUDE 127.0.0.1",
	}
	if !m.cfg.NoSandbox {
		ignoreArgs = append(ignoreArgs, "--no-sandbox")
	}

	launch := func(launchArgs []string) (playwright.BrowserContext, *SOCKS5Relay, error) {
		opts := playwright.BrowserTypeLaunchPersistentContextOptions{
			ExecutablePath:    playwright.String(absChromiumPath),
			Headless:          playwright.Bool(false),
			AcceptDownloads:   playwright.Bool(true),
			Args:              launchArgs,
			NoViewport:        playwright.Bool(true),
			Locale:            playwright.String(persona.Native.Locale.Locale),
			TimezoneId:        playwright.String(persona.Native.Locale.Timezone),
			Env:               browseForgeChromiumEnv(persona),
			IgnoreDefaultArgs: ignoreArgs,
		}

		var relay *SOCKS5Relay
		if effectiveProxy.Proxy != nil {
			proxy := effectiveProxy.Proxy
			needsRelay := proxy.Type == "socks5" && proxy.Username != ""
			if needsRelay {
				upstream := fmt.Sprintf("%s:%d", proxy.Host, proxy.Port)
				var localAddr string
				relay, localAddr, err = StartSOCKS5Relay(upstream, proxy.Username, proxy.Password)
				if err != nil {
					return nil, nil, fmt.Errorf("socks5 relay: %w", err)
				}
				opts.Proxy = &playwright.Proxy{Server: "socks5://" + localAddr}
			} else {
				server := fmt.Sprintf("%s://%s:%d", proxy.Type, proxy.Host, proxy.Port)
				opts.Proxy = &playwright.Proxy{
					Server:   server,
					Username: playwright.String(proxy.Username),
					Password: playwright.String(proxy.Password),
				}
			}
		}

		ctx, err := m.pw.Chromium.LaunchPersistentContext(userDataDir, opts)
		if err != nil {
			if relay != nil {
				relay.Close()
			}
			return nil, nil, err
		}
		return ctx, relay, nil
	}

	ctx, relay, err := launch(args)
	fallbackAttempted := false
	if err != nil {
		if policy != nil &&
			(policy.RepairTransientCacheOnLaunchFailure || policy.AutoSafeGPUFallback) &&
			isChromiumGPUOrCacheLaunchFailure(err) {
			slog.Warn("repairing transient chromium cache after launch failure", "profile", p.ID, "userDataDir", userDataDir, "error", err)
			repairTransientChromiumData(userDataDir)
		}
		if shouldAutoFallbackCloakBrowserLaunch(policy, err) {
			fallbackAttempted = true
			slog.Warn("retrying CloakBrowser launch with safe GPU fallback", "profile", p.ID, "userDataDir", userDataDir, "error", err)
			if len(m.sessions) > 0 {
				m.dropSessionsLocked("playwright driver restart before CloakBrowser safe GPU fallback")
			}
			if restartErr := m.restartPlaywright(); restartErr != nil {
				return nil, fmt.Errorf("launch chromium: %w; safe GPU fallback playwright restart failed: %v", humanizeError(err), restartErr)
			}
			fallbackArgs, fallbackArgErr := applyCloakBrowserLaunchPolicy(baseArgs, userDataDir, policy, true)
			if fallbackArgErr != nil {
				return nil, fallbackArgErr
			}
			ctx, relay, err = launch(fallbackArgs)
			if err == nil {
				slog.Info("CloakBrowser launch recovered with safe GPU fallback", "profile", p.ID)
			}
		}
	}
	if err != nil {
		if fallbackAttempted {
			return nil, fmt.Errorf("launch chromium: %w", noManagerRetryError{err: humanizeError(err)})
		}
		return nil, fmt.Errorf("launch chromium: %w", humanizeError(err))
	}
	if err := installChromiumAutomationMitigations(ctx); err != nil {
		if relay != nil {
			relay.Close()
		}
		ctx.Close()
		return nil, fmt.Errorf("install chromium automation mitigations: %w", err)
	}

	dlDir := downloadsDir
	onDl := func(d playwright.Download) { go d.SaveAs(filepath.Join(dlDir, d.SuggestedFilename())) }
	for _, pg := range ctx.Pages() {
		pg.OnDownload(onDl)
	}
	ctx.OnPage(func(pg playwright.Page) { pg.OnDownload(onDl) })

	pages := ctx.Pages()
	var page playwright.Page
	if len(pages) > 0 {
		page = pages[0]
	} else {
		page, err = ctx.NewPage()
		if err != nil {
			if relay != nil {
				relay.Close()
			}
			ctx.Close()
			return nil, fmt.Errorf("new page: %w", err)
		}
	}

	return &Session{
		ID:        fmt.Sprintf("sess_%s", p.ID),
		ProfileID: p.ID,
		RuntimeID: string(desc.ID),
		Context:   ctx,
		Page:      page,
		relay:     relay,
	}, nil
}

func installChromiumAutomationMitigations(ctx playwright.BrowserContext) error {
	script := playwright.Script{Content: playwright.String(chromiumAutomationMitigationInitScript)}
	if err := ctx.AddInitScript(script); err != nil {
		return err
	}
	for _, page := range ctx.Pages() {
		if err := page.AddInitScript(script); err != nil {
			return err
		}
	}
	return nil
}

type browseForgeNativePersonaConfig struct {
	SchemaVersion string                           `json:"schema_version"`
	RuntimeID     string                           `json:"runtime_id"`
	Seed          uint64                           `json:"seed"`
	Browser       browseForgeNativeBrowserIdentity `json:"browser"`
	Platform      browseForgeNativePlatform        `json:"platform"`
	Locale        browseForgeNativeLocale          `json:"locale"`
	Hardware      browseForgeNativeHardware        `json:"hardware"`
	Screen        browseForgeNativeScreen          `json:"screen"`
	GPU           browseForgeNativeGPU             `json:"gpu"`
	WebRTC        browseForgeNativeWebRTC          `json:"webrtc"`
	Storage       browseForgeNativeStorage         `json:"storage"`
}

type browseForgeNativePersonaSnapshot struct {
	SchemaVersion string                           `json:"schema_version"`
	RuntimeID     string                           `json:"runtime_id"`
	Seed          uint64                           `json:"seed"`
	PersonaIDHash string                           `json:"persona_id_hash"`
	OriginSaltKey string                           `json:"origin_salt_key"`
	Browser       browseForgeNativeBrowserIdentity `json:"browser"`
	Platform      browseForgeNativePlatform        `json:"platform"`
	Locale        browseForgeNativeLocale          `json:"locale"`
	Hardware      browseForgeNativeHardware        `json:"hardware"`
	Screen        browseForgeNativeScreen          `json:"screen"`
	GPU           browseForgeNativeGPU             `json:"gpu"`
	WebRTC        browseForgeNativeWebRTC          `json:"webrtc"`
	Storage       browseForgeNativeStorage         `json:"storage"`
}

type browseForgeNativeBrowserIdentity struct {
	Family      string   `json:"family"`
	Major       int      `json:"major"`
	FullVersion string   `json:"full_version"`
	Brands      []string `json:"brands"`
	UserAgent   string   `json:"user_agent"`
}

type browseForgeNativePlatform struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Platform   string `json:"platform"`
	PlatformCH string `json:"platform_ch"`
	Mobile     bool   `json:"mobile"`
	Bitness    string `json:"bitness"`
	Model      string `json:"model"`
}

type browseForgeNativeLocale struct {
	Timezone       string `json:"timezone"`
	Locale         string `json:"locale"`
	AcceptLanguage string `json:"accept_language"`
	GeoSource      string `json:"geo_source,omitempty"`
	GeoStatus      string `json:"geo_status,omitempty"`
}

type browseForgeNativeHardware struct {
	HardwareConcurrency int `json:"hardware_concurrency"`
	DeviceMemoryGB      int `json:"device_memory_gb"`
}

type browseForgeNativeScreen struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	AvailWidth  int     `json:"avail_width"`
	AvailHeight int     `json:"avail_height"`
	DPR         float64 `json:"dpr"`
	ColorDepth  int     `json:"color_depth"`
}

type browseForgeNativeGPU struct {
	Vendor       string            `json:"vendor"`
	Renderer     string            `json:"renderer"`
	ANGLEBackend string            `json:"angle_backend"`
	WebGLParams  map[string]string `json:"webgl_params,omitempty"`
}

type browseForgeNativeWebRTC struct {
	Mode              string `json:"mode"`
	ProxyRegion       string `json:"proxy_region"`
	DirectIPRedaction bool   `json:"direct_ip_redaction"`
}

type browseForgeNativeStorage struct {
	QuotaMB    int  `json:"quota_mb"`
	Persistent bool `json:"persistent"`
}

type chromiumLaunchPersona struct {
	Native            browseForgeNativePersonaConfig
	NavigatorPlatform string
	CanvasNoise       int64
	HasCanvasNoise    bool
	AudioNoise        int64
	HasAudioNoise     bool
	FontsList         string
	HasFontsList      bool
	HasWebGLVendor    bool
	HasWebGLRenderer  bool
	PluginsPDF        string
}

func buildChromiumLaunchPersona(p *profile.Profile, runtimeID bfruntime.ID, platform, timezone, locale, proxyRegion, goarch string, policy *config.CloakBrowserConfig) (chromiumLaunchPersona, error) {
	if p == nil {
		return chromiumLaunchPersona{}, fmt.Errorf("profile is nil")
	}
	nativePlatform, err := nativePersonaPlatform(platform, goarch)
	if err != nil {
		return chromiumLaunchPersona{}, err
	}
	if runtimeID == bfruntime.BrowseForgeChromium {
		proxyRegion, err = sanitizeBrowseForgeProxyRegion(proxyRegion)
		if err != nil {
			return chromiumLaunchPersona{}, err
		}
	} else {
		proxyRegion = strings.TrimSpace(proxyRegion)
	}
	fp := p.Fingerprint
	userAgent := effectiveChromiumUserAgent(fp, platform)
	acceptLanguage := effectiveChromiumAcceptLanguage(fp, locale)
	fullVersion := chromiumVersionFromUserAgent(userAgent)
	vendor, hasWebGLVendor := fingerprintString(fp, "webGl:vendor")
	if !hasWebGLVendor {
		vendor = "Intel Inc."
	}
	renderer, hasWebGLRenderer := fingerprintString(fp, "webGl:renderer")
	if !hasWebGLRenderer {
		renderer = "Intel Iris OpenGL Engine"
	}
	if runtimeID == bfruntime.BrowseForgeChromium && browseForgeDockerGPUMode() == "software" && (!hasWebGLVendor || !hasWebGLRenderer) {
		vendor = "Google Inc. (Google)"
		renderer = "ANGLE (Google, Vulkan 1.3.0 (SwiftShader Device (Subzero) (0x0000C0DE)), SwiftShader driver)"
		hasWebGLVendor = true
		hasWebGLRenderer = true
	}
	storageQuota := int64(0)
	if runtimeID == bfruntime.BrowseForgeChromium {
		storageQuota = 8192
	}
	if quota := cloakStorageQuotaMB(policy); quota > 0 {
		storageQuota = quota
	} else if quota < 0 {
		return chromiumLaunchPersona{}, fmt.Errorf("%s storage_quota_mb must be >= 0", runtimeID)
	}
	screenWidth := fingerprintIntDefault(fp, "screen.width", 1920)
	screenHeight := fingerprintIntDefault(fp, "screen.height", 1080)
	availWidth := clampScreenAvail(fingerprintIntDefault(fp, "screen.availWidth", 1920), screenWidth)
	availHeight := clampScreenAvail(fingerprintIntDefault(fp, "screen.availHeight", 1040), screenHeight)
	persona := chromiumLaunchPersona{
		Native: browseForgeNativePersonaConfig{
			SchemaVersion: "1.0",
			RuntimeID:     string(runtimeID),
			Seed:          browseForgePersonaSeed(p),
			Browser: browseForgeNativeBrowserIdentity{
				Family:      "chromium",
				Major:       chromiumMajorVersion(fullVersion),
				FullVersion: fullVersion,
				Brands:      []string{"Chromium", "Google Chrome"},
				UserAgent:   userAgent,
			},
			Platform: nativePlatform,
			Locale: browseForgeNativeLocale{
				Timezone:       timezone,
				Locale:         locale,
				AcceptLanguage: acceptLanguage,
			},
			Hardware: browseForgeNativeHardware{
				HardwareConcurrency: fingerprintIntDefault(fp, "navigator.hardwareConcurrency", 8),
				DeviceMemoryGB:      fingerprintIntDefault(fp, "navigator.deviceMemory", 8),
			},
			Screen: browseForgeNativeScreen{
				Width:       screenWidth,
				Height:      screenHeight,
				AvailWidth:  availWidth,
				AvailHeight: availHeight,
				DPR:         fingerprintFloatDefault(fp, "screen.devicePixelRatio", 1),
				ColorDepth:  fingerprintIntDefault(fp, "screen.colorDepth", 24),
			},
			GPU: browseForgeNativeGPU{
				Vendor:       vendor,
				Renderer:     renderer,
				ANGLEBackend: "",
				WebGLParams:  map[string]string{},
			},
			WebRTC: browseForgeNativeWebRTC{
				Mode:              "disable_non_proxied_udp",
				ProxyRegion:       proxyRegion,
				DirectIPRedaction: true,
			},
			Storage: browseForgeNativeStorage{
				QuotaMB:    int(storageQuota),
				Persistent: true,
			},
		},
		NavigatorPlatform: platform,
		HasWebGLVendor:    hasWebGLVendor,
		HasWebGLRenderer:  hasWebGLRenderer,
		PluginsPDF:        cloakPluginsPDF(policy),
	}
	if v, ok := fingerprintInt(fp, "canvas:seed"); ok {
		persona.CanvasNoise = v
		persona.HasCanvasNoise = true
	}
	if v, ok := fingerprintInt(fp, "audio:seed"); ok {
		persona.AudioNoise = v
		persona.HasAudioNoise = true
	}
	if v, ok := fingerprintStringList(fp, "fonts", "|"); ok {
		persona.FontsList = v
		persona.HasFontsList = true
	}
	return persona, nil
}

func appendChromiumLaunchPersonaArgs(args []string, persona chromiumLaunchPersona) []string {
	native := persona.Native
	args = append(args,
		"--fingerprint-timezone="+native.Locale.Timezone,
		"--fingerprint-locale="+native.Locale.Locale,
		"--fingerprint-platform="+persona.NavigatorPlatform,
	)
	if native.Storage.QuotaMB > 0 {
		args = append(args, fmt.Sprintf("--fingerprint-storage-quota=%d", native.Storage.QuotaMB))
	}
	if persona.PluginsPDF != "" {
		args = append(args, "--fingerprint-plugins-pdf="+persona.PluginsPDF)
	}
	if native.Browser.UserAgent != "" {
		args = append(args, "--user-agent="+native.Browser.UserAgent, "--fingerprint-user-agent="+native.Browser.UserAgent)
		if native.Browser.FullVersion != "" {
			args = append(args, "--fingerprint-ua-full-version="+native.Browser.FullVersion)
		}
	}
	args = append(args,
		"--fingerprint-ua-platform="+native.Platform.PlatformCH,
		"--fingerprint-ua-architecture="+native.Platform.Arch,
		"--fingerprint-ua-bitness="+native.Platform.Bitness,
	)
	if acceptLanguage := chromiumAcceptLanguageSwitchValue(native.Locale.AcceptLanguage); acceptLanguage != "" {
		args = append(args, "--fingerprint-accept-language="+acceptLanguage)
	}
	if native.Hardware.HardwareConcurrency > 0 {
		args = append(args, fmt.Sprintf("--fingerprint-hardware-concurrency=%d", native.Hardware.HardwareConcurrency))
	}
	if native.Hardware.DeviceMemoryGB > 0 {
		args = append(args, fmt.Sprintf("--fingerprint-device-memory=%d", native.Hardware.DeviceMemoryGB))
	}
	if native.Screen.Width > 0 {
		args = append(args, fmt.Sprintf("--fingerprint-screen-width=%d", native.Screen.Width))
	}
	if native.Screen.Height > 0 {
		args = append(args, fmt.Sprintf("--fingerprint-screen-height=%d", native.Screen.Height))
	}
	if native.Screen.AvailWidth > 0 {
		args = append(args, fmt.Sprintf("--fingerprint-screen-avail-width=%d", native.Screen.AvailWidth))
	}
	if native.Screen.AvailHeight > 0 {
		args = append(args, fmt.Sprintf("--fingerprint-screen-avail-height=%d", native.Screen.AvailHeight))
	}
	if persona.HasCanvasNoise {
		args = append(args, fmt.Sprintf("--fingerprint-canvas-noise=%d", persona.CanvasNoise))
	}
	if persona.HasAudioNoise {
		args = append(args, fmt.Sprintf("--fingerprint-audio-noise=%d", persona.AudioNoise))
	}
	if persona.HasFontsList {
		args = append(args, "--fingerprint-fonts-list="+persona.FontsList)
	}
	if persona.HasWebGLVendor {
		args = append(args, "--fingerprint-webgl-vendor="+native.GPU.Vendor)
	}
	if persona.HasWebGLRenderer {
		args = append(args, "--fingerprint-webgl-renderer="+native.GPU.Renderer)
	}
	return args
}

func browseForgeChromiumWindowArgs(persona chromiumLaunchPersona) []string {
	native := persona.Native
	width := native.Screen.AvailWidth
	height := native.Screen.AvailHeight
	if width <= 0 {
		width = native.Screen.Width
	}
	if height <= 0 {
		height = native.Screen.Height
	}
	if width <= 0 || height <= 0 {
		return nil
	}
	return []string{fmt.Sprintf("--window-size=%d,%d", width, height)}
}

func browseForgeChromiumEnv(persona chromiumLaunchPersona) map[string]string {
	native := persona.Native
	env := map[string]string{
		"DISPLAY":               os.Getenv("DISPLAY"),
		"HOME":                  os.Getenv("HOME"),
		"LIBGL_ALWAYS_SOFTWARE": os.Getenv("LIBGL_ALWAYS_SOFTWARE"),
	}
	if native.Locale.Timezone != "" {
		env["TZ"] = native.Locale.Timezone
	}
	if native.Locale.Locale != "" {
		localeEnv := strings.ReplaceAll(native.Locale.Locale, "-", "_") + ".UTF-8"
		env["LANG"] = localeEnv
		env["LC_ALL"] = localeEnv
		env["BROWSEFORGE_INTL_LOCALE"] = native.Locale.Locale
	}
	if acceptLanguage := native.Locale.AcceptLanguage; acceptLanguage != "" {
		env["BROWSEFORGE_ACCEPT_LANGUAGE"] = acceptLanguage
	}
	return env
}

func browseForgeDockerGPUMode() string {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("BROWSEFORGE_DOCKER_GPU_MODE")))
	switch mode {
	case "software":
		return "software"
	case "native":
		return "native"
	default:
		return ""
	}
}

func writeBrowseForgeNativeConfig(userDataDir string, persona chromiumLaunchPersona) (string, error) {
	snapshot, err := resolveBrowseForgeNativePersona(persona.Native)
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(userDataDir, "BrowseForgeNative")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("create BrowseForge native config dir: %w", err)
	}
	configPath := filepath.Join(configDir, "persona.json")
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode BrowseForge native config: %w", err)
	}
	if err := os.WriteFile(configPath, append(data, '\n'), 0644); err != nil {
		return "", fmt.Errorf("write BrowseForge native config: %w", err)
	}
	return configPath, nil
}

func resolveBrowseForgeNativePersona(cfg browseForgeNativePersonaConfig) (browseForgeNativePersonaSnapshot, error) {
	if err := validateBrowseForgeNativePersonaPlatform(cfg.Platform); err != nil {
		return browseForgeNativePersonaSnapshot{}, err
	}
	canonical, err := json.Marshal(cfg)
	if err != nil {
		return browseForgeNativePersonaSnapshot{}, fmt.Errorf("encode BrowseForge native canonical config: %w", err)
	}
	personaHash := sha256.Sum256(canonical)
	originKey := hmac.New(sha256.New, []byte(fmt.Sprintf("browseforge-origin-salt:%d", cfg.Seed)))
	_, _ = originKey.Write(canonical)
	return browseForgeNativePersonaSnapshot{
		SchemaVersion: cfg.SchemaVersion,
		RuntimeID:     cfg.RuntimeID,
		Seed:          cfg.Seed,
		PersonaIDHash: hex.EncodeToString(personaHash[:16]),
		OriginSaltKey: hex.EncodeToString(originKey.Sum(nil)[:16]),
		Browser:       cfg.Browser,
		Platform:      cfg.Platform,
		Locale:        cfg.Locale,
		Hardware:      cfg.Hardware,
		Screen:        cfg.Screen,
		GPU:           cfg.GPU,
		WebRTC:        cfg.WebRTC,
		Storage:       cfg.Storage,
	}, nil
}

func browseForgeChromiumNativeMode(policy *config.CloakBrowserConfig) string {
	if policy != nil && strings.TrimSpace(policy.NativeMode) != "" {
		return strings.TrimSpace(policy.NativeMode)
	}
	return "enabled"
}

func browseForgePersonaSeed(p *profile.Profile) uint64 {
	if p.FingerprintSeed > 0 {
		return uint64(p.FingerprintSeed)
	}
	if seed, ok := fingerprintInt(p.Fingerprint, "canvas:seed"); ok {
		return uint64(seed)
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(p.ID + ":" + p.RuntimeID))
	seed := h.Sum64()
	if seed == 0 {
		return 1
	}
	return seed
}

func defaultChromiumUserAgent(platform string) string {
	switch platform {
	case "MacIntel":
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36"
	case "Linux aarch64":
		return "Mozilla/5.0 (X11; Linux aarch64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36"
	case "Linux x86_64":
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36"
	}
}

func chromiumVersionFromUserAgent(userAgent string) string {
	const token = "Chrome/"
	idx := strings.Index(userAgent, token)
	if idx < 0 {
		return "150.0.7871.101"
	}
	version := userAgent[idx+len(token):]
	if end := strings.IndexAny(version, " )"); end >= 0 {
		version = version[:end]
	}
	if version == "" {
		return "150.0.7871.101"
	}
	return version
}

func chromiumMajorVersion(fullVersion string) int {
	major := fullVersion
	if dot := strings.IndexByte(major, '.'); dot >= 0 {
		major = major[:dot]
	}
	parsed, err := strconv.Atoi(major)
	if err != nil || parsed <= 0 {
		return 150
	}
	return parsed
}

func nativePersonaPlatform(platform, goarch string) (browseForgeNativePlatform, error) {
	switch platform {
	case "Win32":
		return browseForgeNativePlatform{OS: "windows", Arch: "x86", Platform: "Win32", PlatformCH: "Windows", Mobile: false, Bitness: "64", Model: ""}, nil
	case "MacIntel":
		arch := "x86"
		if goarch == "arm64" {
			arch = "arm"
		}
		return browseForgeNativePlatform{OS: "macos", Arch: arch, Platform: "MacIntel", PlatformCH: "macOS", Mobile: false, Bitness: "64", Model: ""}, nil
	case "Linux x86_64":
		return browseForgeNativePlatform{OS: "linux", Arch: "x86", Platform: "Linux x86_64", PlatformCH: "Linux", Mobile: false, Bitness: "64", Model: ""}, nil
	case "Linux aarch64":
		return browseForgeNativePlatform{OS: "linux", Arch: "arm", Platform: "Linux aarch64", PlatformCH: "Linux", Mobile: false, Bitness: "64", Model: ""}, nil
	default:
		return browseForgeNativePlatform{}, fmt.Errorf("chromium native persona platform %q is not supported", platform)
	}
}

func validateBrowseForgeNativePersonaPlatform(platform browseForgeNativePlatform) error {
	if platform.Bitness != "64" {
		return fmt.Errorf("chromium native persona platform mismatch: platform=%s os=%s arch=%s bitness=%s platform_ch=%s", platform.Platform, platform.OS, platform.Arch, platform.Bitness, platform.PlatformCH)
	}
	switch platform.OS {
	case "windows":
		if platform.Platform == "Win32" && platform.Arch == "x86" && platform.PlatformCH == "Windows" && !platform.Mobile && platform.Model == "" {
			return nil
		}
	case "macos":
		if platform.Platform == "MacIntel" && (platform.Arch == "x86" || platform.Arch == "arm") && platform.PlatformCH == "macOS" && !platform.Mobile && platform.Model == "" {
			return nil
		}
	case "linux":
		if ((platform.Platform == "Linux x86_64" && platform.Arch == "x86") || (platform.Platform == "Linux aarch64" && platform.Arch == "arm")) && platform.PlatformCH == "Linux" && !platform.Mobile && platform.Model == "" {
			return nil
		}
	}
	return fmt.Errorf("chromium native persona platform mismatch: platform=%s os=%s arch=%s bitness=%s platform_ch=%s", platform.Platform, platform.OS, platform.Arch, platform.Bitness, platform.PlatformCH)
}

func clampScreenAvail(avail, size int) int {
	if size <= 0 {
		return avail
	}
	if avail <= 0 || avail > size {
		return size
	}
	return avail
}
func fallbackGeoForProxyRegion(region string) (timezone, locale string) {
	region = strings.ToLower(strings.TrimSpace(region))
	switch {
	case strings.HasPrefix(region, "us"):
		return "America/New_York", "en-US"
	case strings.HasPrefix(region, "tw"):
		return "Asia/Taipei", "zh-TW"
	case strings.HasPrefix(region, "jp"):
		return "Asia/Tokyo", "ja-JP"
	case strings.HasPrefix(region, "kr"):
		return "Asia/Seoul", "ko-KR"
	case strings.HasPrefix(region, "sg"):
		return "Asia/Singapore", "en-SG"
	case strings.HasPrefix(region, "hk"):
		return "Asia/Hong_Kong", "zh-HK"
	case strings.HasPrefix(region, "de"):
		return "Europe/Berlin", "de-DE"
	case strings.HasPrefix(region, "fr"):
		return "Europe/Paris", "fr-FR"
	case strings.HasPrefix(region, "gb"), strings.HasPrefix(region, "uk"):
		return "Europe/London", "en-GB"
	default:
		return "", ""
	}
}

func sanitizeBrowseForgeProxyRegion(region string) (string, error) {
	region = strings.ToLower(strings.TrimSpace(region))
	if region == "" {
		return "", nil
	}
	if len(region) > 64 {
		return "", fmt.Errorf("browseforge-chromium proxy_region must be a redacted region label of at most 64 characters")
	}
	for _, r := range region {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return "", fmt.Errorf("browseforge-chromium proxy_region must contain only letters, digits, hyphen, or underscore")
	}
	if region[0] == '-' || region[0] == '_' || region[len(region)-1] == '-' || region[len(region)-1] == '_' {
		return "", fmt.Errorf("browseforge-chromium proxy_region must start and end with a letter or digit")
	}
	return region, nil
}

func fingerprintIntDefault(fp map[string]any, key string, fallback int) int {
	if value, ok := fingerprintInt(fp, key); ok {
		return int(value)
	}
	return fallback
}

func fingerprintFloatDefault(fp map[string]any, key string, fallback float64) float64 {
	if fp == nil {
		return fallback
	}
	switch typed := fp[key].(type) {
	case int:
		if typed > 0 {
			return float64(typed)
		}
	case int64:
		if typed > 0 {
			return float64(typed)
		}
	case float64:
		if typed > 0 {
			return typed
		}
	case json.Number:
		if parsed, err := strconv.ParseFloat(string(typed), 64); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func applyCloakBrowserLaunchPolicy(args []string, userDataDir string, policy *config.CloakBrowserConfig, fallback bool) ([]string, error) {
	out := append([]string(nil), args...)
	if policy == nil {
		return out, nil
	}

	if policy.SafeGPU || fallback {
		out = appendUniqueChromiumArgs(out,
			"--disable-gpu",
			"--disable-gpu-compositing",
			"--disable-gpu-sandbox",
			"--disable-gpu-shader-disk-cache",
			"--in-process-gpu",
		)
	}
	if policy.IsolatedRuntimeCache || fallback {
		cacheDir := filepath.Join(userDataDir, "BrowseForgeRuntimeCache", fmt.Sprintf("cache-%d-%d", os.Getpid(), time.Now().UnixNano()))
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return nil, fmt.Errorf("create chromium runtime cache dir: %w", err)
		}
		out = append(out, "--disk-cache-dir="+cacheDir)
	}
	out = appendUniqueChromiumArgs(out, sanitizeExtraChromiumArgs(policy.ExtraArgs)...)
	return out, nil
}

func resolveCloakFingerprintPlatform(policy *config.CloakBrowserConfig, goos string) (string, error) {
	platform := "Win32"
	if goos == "darwin" {
		platform = "MacIntel"
	}
	if policy == nil || policy.FingerprintPlatform == "" || policy.FingerprintPlatform == "auto" {
		return platform, nil
	}
	switch policy.FingerprintPlatform {
	case "windows":
		return "Win32", nil
	case "macos":
		return "MacIntel", nil
	case "linux":
		return "Linux x86_64", nil
	default:
		return "", fmt.Errorf("cloakbrowser fingerprint_platform must be auto, macos, windows, or linux")
	}
}

func resolveChromiumFingerprintPlatform(policy *config.CloakBrowserConfig, goos, goarch string) (string, error) {
	if policy == nil || policy.FingerprintPlatform == "" || policy.FingerprintPlatform == "auto" {
		switch goos {
		case "darwin":
			return "MacIntel", nil
		case "linux":
			if goarch == "arm64" {
				return "Linux aarch64", nil
			}
			return "Linux x86_64", nil
		default:
			return "Win32", nil
		}
	}
	switch policy.FingerprintPlatform {
	case "windows":
		return "Win32", nil
	case "macos":
		return "MacIntel", nil
	case "linux":
		if goarch == "arm64" {
			return "Linux aarch64", nil
		}
		return "Linux x86_64", nil
	default:
		return "", fmt.Errorf("chromium fingerprint_platform must be auto, macos, windows, or linux")
	}
}

func effectiveChromiumUserAgent(fp map[string]any, platform string) string {
	if ua, ok := fingerprintString(fp, "navigator.userAgent"); ok && userAgentMatchesPlatform(ua, platform) {
		return ua
	}
	return defaultChromiumUserAgent(platform)
}

func userAgentMatchesPlatform(userAgent, platform string) bool {
	switch platform {
	case "MacIntel":
		return strings.Contains(userAgent, "Macintosh")
	case "Linux aarch64":
		return strings.Contains(userAgent, "Linux aarch64")
	case "Linux x86_64":
		return strings.Contains(userAgent, "Linux x86_64")
	default:
		return strings.Contains(userAgent, "Windows NT")
	}
}

func effectiveChromiumAcceptLanguage(fp map[string]any, locale string) string {
	profileAcceptLanguage, ok := fingerprintAcceptLanguage(fp)
	if ok && acceptLanguageMatchesLocale(profileAcceptLanguage, locale) {
		return profileAcceptLanguage
	}
	return acceptLanguageForLocale(locale)
}

func acceptLanguageMatchesLocale(acceptLanguage, locale string) bool {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return strings.TrimSpace(acceptLanguage) != ""
	}
	first := strings.TrimSpace(strings.Split(acceptLanguage, ",")[0])
	return strings.EqualFold(first, locale)
}

func acceptLanguageForLocale(locale string) string {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return "en-US,en;q=0.9"
	}
	if dash := strings.IndexByte(locale, '-'); dash > 0 {
		primary := locale[:dash]
		return locale + "," + primary + ";q=0.9"
	}
	return locale
}

func chromiumAcceptLanguageSwitchValue(acceptLanguage string) string {
	parts := strings.Split(acceptLanguage, ",")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		lang := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if lang != "" {
			cleaned = append(cleaned, lang)
		}
	}
	return strings.Join(cleaned, ",")
}

func appendProfileFingerprintArgs(args []string, fp map[string]any, userAgent, platform, acceptLanguage string) []string {
	if userAgent != "" {
		args = append(args, "--user-agent="+userAgent, "--fingerprint-user-agent="+userAgent)
		if fullVersion := chromiumVersionFromUserAgent(userAgent); fullVersion != "" {
			args = append(args, "--fingerprint-ua-full-version="+fullVersion)
		}
	}
	if platform != "" {
		nativePlatform, err := nativePersonaPlatform(platform, runtime.GOARCH)
		if err == nil {
			args = append(args,
				"--fingerprint-ua-platform="+nativePlatform.PlatformCH,
				"--fingerprint-ua-architecture="+nativePlatform.Arch,
				"--fingerprint-ua-bitness="+nativePlatform.Bitness,
			)
		}
	}
	if acceptLanguage := chromiumAcceptLanguageSwitchValue(acceptLanguage); acceptLanguage != "" {
		args = append(args, "--fingerprint-accept-language="+acceptLanguage)
	}
	if v, ok := fingerprintInt(fp, "navigator.hardwareConcurrency"); ok {
		args = append(args, fmt.Sprintf("--fingerprint-hardware-concurrency=%d", v))
	}
	if v, ok := fingerprintInt(fp, "navigator.deviceMemory"); ok {
		args = append(args, fmt.Sprintf("--fingerprint-device-memory=%d", v))
	}
	if v, ok := fingerprintInt(fp, "screen.width"); ok {
		args = append(args, fmt.Sprintf("--fingerprint-screen-width=%d", v))
	}
	if v, ok := fingerprintInt(fp, "screen.height"); ok {
		args = append(args, fmt.Sprintf("--fingerprint-screen-height=%d", v))
	}
	if v, ok := fingerprintInt(fp, "screen.availWidth"); ok {
		args = append(args, fmt.Sprintf("--fingerprint-screen-avail-width=%d", v))
	}
	if v, ok := fingerprintInt(fp, "screen.availHeight"); ok {
		args = append(args, fmt.Sprintf("--fingerprint-screen-avail-height=%d", v))
	}
	if v, ok := fingerprintInt(fp, "canvas:seed"); ok {
		args = append(args, fmt.Sprintf("--fingerprint-canvas-noise=%d", v))
	}
	if v, ok := fingerprintInt(fp, "audio:seed"); ok {
		args = append(args, fmt.Sprintf("--fingerprint-audio-noise=%d", v))
	}
	if v, ok := fingerprintStringList(fp, "fonts", "|"); ok {
		args = append(args, "--fingerprint-fonts-list="+v)
	}
	if v, ok := fingerprintString(fp, "webGl:vendor"); ok {
		args = append(args, "--fingerprint-webgl-vendor="+v)
	}
	if v, ok := fingerprintString(fp, "webGl:renderer"); ok {
		args = append(args, "--fingerprint-webgl-renderer="+v)
	}
	return args
}

func fingerprintString(fp map[string]any, key string) (string, bool) {
	if fp == nil {
		return "", false
	}
	value, ok := fp[key]
	if !ok {
		return "", false
	}
	s, ok := value.(string)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	return s, s != ""
}

func fingerprintAcceptLanguage(fp map[string]any) (string, bool) {
	if value, ok := fp["navigator.languages"]; ok {
		switch typed := value.(type) {
		case []string:
			if len(typed) > 0 {
				return strings.Join(typed, ","), true
			}
		case []any:
			langs := make([]string, 0, len(typed))
			for _, item := range typed {
				if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
					langs = append(langs, strings.TrimSpace(s))
				}
			}
			if len(langs) > 0 {
				return strings.Join(langs, ","), true
			}
		}
	}
	return fingerprintString(fp, "navigator.language")
}

func fingerprintStringList(fp map[string]any, key string, sep string) (string, bool) {
	if fp == nil {
		return "", false
	}
	value, ok := fp[key]
	if !ok {
		return "", false
	}
	var items []string
	switch typed := value.(type) {
	case []string:
		items = typed
	case []any:
		items = make([]string, 0, len(typed))
		for _, item := range typed {
			if s, ok := item.(string); ok {
				items = append(items, s)
			}
		}
	default:
		return "", false
	}
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" && !strings.Contains(item, sep) {
			cleaned = append(cleaned, item)
		}
	}
	if len(cleaned) == 0 {
		return "", false
	}
	return strings.Join(cleaned, sep), true
}

func fingerprintInt(fp map[string]any, key string) (int64, bool) {
	if fp == nil {
		return 0, false
	}
	value, ok := fp[key]
	if !ok {
		return 0, false
	}
	var n int64
	switch typed := value.(type) {
	case int:
		n = int64(typed)
	case int64:
		n = typed
	case float64:
		if typed != float64(int64(typed)) {
			return 0, false
		}
		n = int64(typed)
	case json.Number:
		parsed, err := strconv.ParseInt(string(typed), 10, 64)
		if err != nil {
			return 0, false
		}
		n = parsed
	default:
		return 0, false
	}
	return n, n > 0
}

func cloakStorageQuotaMB(policy *config.CloakBrowserConfig) int64 {
	if policy == nil {
		return 0
	}
	return policy.StorageQuotaMB
}

func cloakPluginsPDF(policy *config.CloakBrowserConfig) string {
	if policy == nil {
		return ""
	}
	switch policy.PluginsPDF {
	case "", "enabled", "true", "1", "disabled", "false", "0":
		return policy.PluginsPDF
	default:
		return ""
	}
}

func resolveCloakFontsDir(policy *config.CloakBrowserConfig) (string, error) {
	if policy != nil && policy.FontsDir != "" {
		fontsDir, err := filepath.Abs(policy.FontsDir)
		if err != nil {
			return "", fmt.Errorf("cloakbrowser fonts_dir: %w", err)
		}
		info, err := os.Stat(fontsDir)
		if err != nil {
			return "", fmt.Errorf("cloakbrowser fonts_dir unavailable: %w", err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("cloakbrowser fonts_dir is not a directory: %s", fontsDir)
		}
		return fontsDir, nil
	}
	if info, err := os.Stat("/usr/share/fonts"); err == nil && info.IsDir() {
		return "/usr/share/fonts", nil
	}
	return "", nil
}

func validateCloakFingerprintPolicy(policy *config.CloakBrowserConfig, platform string, goos string) error {
	if policy == nil {
		return nil
	}
	mode := policy.TargetPlatformPolicy
	if mode == "" {
		mode = "warn"
	}
	switch mode {
	case "allow", "warn", "strict":
	default:
		return fmt.Errorf("cloakbrowser target_platform_policy must be strict, warn, or allow")
	}
	switch policy.PluginsPDF {
	case "", "enabled", "true", "1", "disabled", "false", "0":
	default:
		return fmt.Errorf("cloakbrowser plugins_pdf must be enabled/true/1 or disabled/false/0")
	}
	if mode != "allow" && platform == "Win32" && goos != "windows" && policy.FontsDir == "" {
		msg := "Windows CloakBrowser fingerprint on non-Windows host should configure runtimes.cloakbrowser.settings.fonts_dir with a Windows-compatible font pack"
		if mode == "strict" {
			return fmt.Errorf("%s", msg)
		}
		slog.Warn(msg, "goos", goos, "platform", platform)
	}
	return nil
}

func shouldAutoFallbackCloakBrowserLaunch(policy *config.CloakBrowserConfig, err error) bool {
	return policy != nil &&
		policy.AutoSafeGPUFallback &&
		isChromiumGPUOrCacheLaunchFailure(err) &&
		(!policy.SafeGPU || !policy.IsolatedRuntimeCache)
}

func appendUniqueChromiumArgs(args []string, extra ...string) []string {
	seen := make(map[string]bool, len(args)+len(extra))
	for _, arg := range args {
		seen[arg] = true
	}
	for _, arg := range extra {
		if arg == "" || seen[arg] {
			continue
		}
		seen[arg] = true
		args = append(args, arg)
	}
	return args
}
