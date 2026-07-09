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

	"github.com/playwright-community/playwright-go"
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

	var tz, locale string
	effectiveProxy := m.effectiveProxy(p)
	if effectiveProxy.Proxy != nil {
		proxy := effectiveProxy.Proxy
		tz, locale = fingerprint.DetectProxyGeoResult(proxy.Type, proxy.Host, proxy.Port, proxy.Username, proxy.Password)
		args = append(args, "--fingerprint-webrtc-ip=auto")
	} else {
		tz, locale = fingerprint.DetectLocalGeoResult()
	}
	policy := m.cfg.ChromiumRuntimeSettings(string(desc.ID))
	platform, err := resolveCloakFingerprintPlatform(policy, runtime.GOOS)
	if err != nil {
		return nil, err
	}
	if profilePlatform, ok := fingerprintString(p.Fingerprint, "navigator.platform"); ok {
		platform = profilePlatform
	}
	args = append(args,
		"--fingerprint-timezone="+tz,
		"--fingerprint-locale="+locale,
		"--fingerprint-platform="+platform,
	)
	if quota := cloakStorageQuotaMB(policy); quota > 0 {
		args = append(args, fmt.Sprintf("--fingerprint-storage-quota=%d", quota))
	} else if quota < 0 {
		return nil, fmt.Errorf("%s storage_quota_mb must be >= 0", desc.ID)
	}

	if desc.ID == bfruntime.BrowseForgeChromium {
		nativeConfigPath, err := writeBrowseForgeNativeConfig(p, userDataDir, tz, locale, platform, policy)
		if err != nil {
			return nil, err
		}
		args = append(args,
			"--browseforge-stealth-config="+nativeConfigPath,
			"--browseforge-stealth-mode="+browseForgeChromiumNativeMode(policy),
		)
	}
	args = appendProfileFingerprintArgs(args, p.Fingerprint)

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
			Viewport:          &playwright.Size{Width: 1280, Height: 800},
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

func writeBrowseForgeNativeConfig(p *profile.Profile, userDataDir, timezone, locale, platform string, policy *config.CloakBrowserConfig) (string, error) {
	if p == nil {
		return "", fmt.Errorf("profile is nil")
	}
	fp := p.Fingerprint
	userAgent, ok := fingerprintString(fp, "navigator.userAgent")
	if !ok {
		userAgent = defaultChromiumUserAgent(platform)
	}
	fullVersion := chromiumVersionFromUserAgent(userAgent)
	platformOS, platformCH, arch, bitness := nativePersonaPlatform(platform)
	acceptLanguage, ok := fingerprintAcceptLanguage(fp)
	if !ok {
		acceptLanguage = locale
	}
	vendor, ok := fingerprintString(fp, "webGl:vendor")
	if !ok {
		vendor = "Intel Inc."
	}
	renderer, ok := fingerprintString(fp, "webGl:renderer")
	if !ok {
		renderer = "Intel Iris OpenGL Engine"
	}
	storageQuota := int64(8192)
	if quota := cloakStorageQuotaMB(policy); quota > 0 {
		storageQuota = quota
	}
	persona := browseForgeNativePersonaConfig{
		SchemaVersion: "1.0",
		RuntimeID:     "browseforge-chromium",
		Seed:          browseForgePersonaSeed(p),
		Browser: browseForgeNativeBrowserIdentity{
			Family:      "chromium",
			Major:       chromiumMajorVersion(fullVersion),
			FullVersion: fullVersion,
			Brands:      []string{"Chromium", "Google Chrome"},
			UserAgent:   userAgent,
		},
		Platform: browseForgeNativePlatform{
			OS:         platformOS,
			Arch:       arch,
			Platform:   platform,
			PlatformCH: platformCH,
			Mobile:     false,
			Bitness:    bitness,
			Model:      "",
		},
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
			Width:       fingerprintIntDefault(fp, "screen.width", 1920),
			Height:      fingerprintIntDefault(fp, "screen.height", 1080),
			AvailWidth:  fingerprintIntDefault(fp, "screen.availWidth", 1920),
			AvailHeight: fingerprintIntDefault(fp, "screen.availHeight", 1040),
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
			ProxyRegion:       "",
			DirectIPRedaction: true,
		},
		Storage: browseForgeNativeStorage{
			QuotaMB:    int(storageQuota),
			Persistent: true,
		},
	}
	snapshot, err := resolveBrowseForgeNativePersona(persona)
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
	if platform == "MacIntel" {
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36"
	}
	if platform == "Linux x86_64" {
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36"
	}
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36"
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

func nativePersonaPlatform(platform string) (osName, platformCH, arch, bitness string) {
	switch platform {
	case "MacIntel":
		return "macos", "macOS", "x86", "64"
	case "Linux x86_64":
		return "linux", "Linux", "x86", "64"
	default:
		return "windows", "Windows", "x86", "64"
	}
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

func appendProfileFingerprintArgs(args []string, fp map[string]any) []string {
	if len(fp) == 0 {
		return args
	}
	if v, ok := fingerprintString(fp, "navigator.userAgent"); ok {
		args = append(args, "--fingerprint-user-agent="+v)
	}
	if v, ok := fingerprintAcceptLanguage(fp); ok {
		args = append(args, "--fingerprint-accept-language="+v)
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
	case "allow":
		return nil
	case "warn", "strict":
	default:
		return fmt.Errorf("cloakbrowser target_platform_policy must be strict, warn, or allow")
	}
	if platform == "Win32" && goos != "windows" && policy.FontsDir == "" {
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
