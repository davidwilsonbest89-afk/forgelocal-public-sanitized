package browser

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"browseforge/internal/config"
	"browseforge/internal/profile"
	bfruntime "browseforge/internal/runtime"

	"github.com/mxschmitt/playwright-go"
)

func TestShouldRetryLaunchForProtocolEOF(t *testing.T) {
	cases := []string{
		"target closed: could not read protocol padding: EOF",
		"launch firefox: target closed",
		"unexpected EOF",
		"FATAL:content\\browser\\gpu\\gpu_data_manager_impl_private.cc:417] GPU process isn't usable. Goodbye.",
		"ERROR:net\\disk_cache\\disk_cache.cc:284] Unable to create cache",
	}

	for _, msg := range cases {
		if !shouldRetryLaunch(errors.New(msg)) {
			t.Fatalf("expected retryable launch error for %q", msg)
		}
	}
}

func TestShouldRetryLaunchRejectsRegularErrors(t *testing.T) {
	if shouldRetryLaunch(errors.New("profile appears to be in use")) {
		t.Fatal("profile lock errors should not restart Playwright")
	}
}

func TestShouldRetryLaunchRejectsNoManagerRetryError(t *testing.T) {
	err := noManagerRetryError{err: errors.New("target closed: GPU process isn't usable")}
	if shouldRetryLaunch(err) {
		t.Fatal("fallback-exhausted errors should not be retried by manager")
	}
}

func TestCleanProfileLocks(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"SingletonLock", "SingletonCookie", "SingletonSocket"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("lock"), 0644); err != nil {
			t.Fatalf("write lock %s: %v", name, err)
		}
	}

	cleanProfileLocks(dir)

	for _, name := range []string{"SingletonLock", "SingletonCookie", "SingletonSocket"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, got err=%v", name, err)
		}
	}
}

func TestSanitizeExtraChromiumArgs(t *testing.T) {
	got := sanitizeExtraChromiumArgs([]string{
		" --disable-features=Translate ",
		"--user-data-dir=C:\\temp\\profile",
		"--remote-debugging-port=9222",
		"--remote-debugging-pipe",
		"--disk-cache-dir=C:\\temp\\cache",
		"--proxy-server=http://proxy.example",
		"--enable-automation",
		"--disable-blink-features=Other",
		"--force-webrtc-ip-handling-policy=default_public_interface_only",
		"--disable-features=Translate",
		"",
		"--disable-background-networking",
	})

	want := []string{"--disable-features=Translate", "--disable-background-networking"}
	if len(got) != len(want) {
		t.Fatalf("args len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q; all=%#v", i, got[i], want[i], got)
		}
	}
}

func TestSanitizeExtraChromiumArgsRejectsManagedFingerprintFlags(t *testing.T) {
	got := sanitizeExtraChromiumArgs([]string{
		"--fingerprint=123",
		"--fingerprint-platform=linux",
		"--fingerprint-fonts-dir=/tmp/fonts",
		"--fingerprint-storage-quota=1024",
		"--fingerprint-timezone=UTC",
		"--fingerprint-locale=en-US",
		"--fingerprint-webrtc-ip=auto",
		"--fingerprint-screen-width=1280",
		"--fingerprint-screen-height=720",
		"--fingerprint-hardware-concurrency=8",
		"--disable-features=Translate",
	})

	want := []string{"--disable-features=Translate"}
	if len(got) != len(want) {
		t.Fatalf("args len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q; all=%#v", i, got[i], want[i], got)
		}
	}
}

func TestRepairTransientChromiumDataPreservesProfileState(t *testing.T) {
	dir := t.TempDir()
	removePaths := []string{
		filepath.Join("Default", "Cache", "data.bin"),
		filepath.Join("Default", "Code Cache", "code.bin"),
		filepath.Join("Default", "GPUCache", "gpu.bin"),
		filepath.Join("BrowseForgeRuntimeCache", "cache-1", "data.bin"),
		"ShaderCache",
		"GrShaderCache",
		"component_crx_cache",
	}
	for _, rel := range removePaths {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte("cache"), 0644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	preservePaths := []string{
		filepath.Join("Default", "Cookies"),
		filepath.Join("Default", "Local Storage", "leveldb", "state.log"),
		filepath.Join("Default", "Preferences"),
	}
	for _, rel := range preservePaths {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte("state"), 0644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	repairTransientChromiumData(dir)

	for _, rel := range removePaths {
		if _, err := os.Stat(filepath.Join(dir, rel)); !os.IsNotExist(err) {
			t.Fatalf("expected transient path %s removed, err=%v", rel, err)
		}
	}
	for _, rel := range preservePaths {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected profile state %s preserved: %v", rel, err)
		}
	}
}

func TestShouldAutoFallbackCloakBrowserLaunch(t *testing.T) {
	cases := []struct {
		name   string
		policy *config.CloakBrowserConfig
		err    error
		want   bool
	}{
		{
			name:   "nil policy never falls back",
			policy: nil,
			err:    errors.New("GPU process isn't usable. Goodbye."),
		},
		{
			name:   "fallback must be explicitly enabled",
			policy: &config.CloakBrowserConfig{},
			err:    errors.New("GPU process isn't usable. Goodbye."),
		},
		{
			name:   "GPU launch failure falls back when enabled",
			policy: &config.CloakBrowserConfig{AutoSafeGPUFallback: true},
			err:    errors.New("GPU process isn't usable. Goodbye."),
			want:   true,
		},
		{
			name:   "cache launch failure falls back when enabled",
			policy: &config.CloakBrowserConfig{AutoSafeGPUFallback: true},
			err:    errors.New("ERROR:net\\disk_cache\\disk_cache.cc:284] Unable to create cache"),
			want:   true,
		},
		{
			name:   "non GPU or cache errors do not fallback",
			policy: &config.CloakBrowserConfig{AutoSafeGPUFallback: true},
			err:    errors.New("profile appears to be in use"),
		},
		{
			name: "policy already using fallback-equivalent settings does not fallback again",
			policy: &config.CloakBrowserConfig{
				AutoSafeGPUFallback:  true,
				SafeGPU:              true,
				IsolatedRuntimeCache: true,
			},
			err: errors.New("GPU process isn't usable. Goodbye."),
		},
		{
			name: "safe GPU without isolated cache still falls back for cache failures",
			policy: &config.CloakBrowserConfig{
				AutoSafeGPUFallback: true,
				SafeGPU:             true,
			},
			err:  errors.New("Unable to create cache"),
			want: true,
		},
		{
			name: "isolated cache without safe GPU still falls back for GPU failures",
			policy: &config.CloakBrowserConfig{
				AutoSafeGPUFallback:  true,
				IsolatedRuntimeCache: true,
			},
			err:  errors.New("GPU process launch failed"),
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldAutoFallbackCloakBrowserLaunch(tc.policy, tc.err); got != tc.want {
				t.Fatalf("fallback = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestApplyCloakBrowserLaunchPolicyFallbackArgs(t *testing.T) {
	dir := t.TempDir()
	args, err := applyCloakBrowserLaunchPolicy(
		[]string{"--no-first-run"},
		dir,
		&config.CloakBrowserConfig{
			AutoSafeGPUFallback: true,
			ExtraArgs:           []string{"--disable-features=Translate", "--user-data-dir=C:\\temp"},
		},
		true,
	)
	if err != nil {
		t.Fatalf("apply fallback policy: %v", err)
	}

	joined := strings.Join(args, "\n")
	for _, want := range []string{
		"--disable-gpu",
		"--disable-gpu-compositing",
		"--disable-gpu-sandbox",
		"--disable-gpu-shader-disk-cache",
		"--in-process-gpu",
		"--disable-features=Translate",
		"--disk-cache-dir=",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("fallback args missing %q: %#v", want, args)
		}
	}
	if strings.Contains(joined, "--user-data-dir") {
		t.Fatalf("unsafe extra arg was not filtered: %#v", args)
	}
}

func TestLaunchChromiumRejectsNegativeStorageQuotaBeforeBrowserLaunch(t *testing.T) {
	enabled := true
	cfg := &config.Config{
		Runtimes: map[string]config.RuntimeConfig{
			"browseforge-chromium": {
				BinaryPath: filepath.Join(t.TempDir(), "browseforge-chromium"),
				Enabled:    &enabled,
				Settings: &config.CloakBrowserConfig{
					StorageQuotaMB:       -1,
					TargetPlatformPolicy: "allow",
				},
			},
		},
	}
	manager := &Manager{
		cfg:      cfg,
		runtimes: bfruntime.NewRegistry(cfg),
		sessions: make(map[string]*Session),
	}

	_, err := manager.launchChromium(&profile.Profile{
		ID:        "storage-quota-negative",
		RuntimeID: "browseforge-chromium",
		Proxy: &profile.ProxyConfig{
			Type: "http",
			Host: "127.0.0.1",
			Port: 1,
		},
		ProfileDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected negative storage quota to fail before launching Chromium")
	}
	if !strings.Contains(err.Error(), "browseforge-chromium storage_quota_mb must be >= 0") {
		t.Fatalf("error = %q, want storage quota validation", err.Error())
	}
}

func TestLaunchProfileDispatchesBrowseForgeChromiumToChromiumLauncher(t *testing.T) {
	enabled := true
	launchErr := errors.New("captured browseforge-chromium launch")
	browserType := &capturingBrowserType{t: t, launchErr: launchErr}
	cfg := &config.Config{
		Runtimes: map[string]config.RuntimeConfig{
			"browseforge-chromium": {
				BinaryPath: filepath.Join(t.TempDir(), "browseforge-chromium"),
				Enabled:    &enabled,
				Settings: &config.CloakBrowserConfig{
					TargetPlatformPolicy: "allow",
				},
			},
		},
	}
	manager := &Manager{
		cfg:      cfg,
		runtimes: bfruntime.NewRegistry(cfg),
		pw:       &playwright.Playwright{Chromium: browserType},
		sessions: make(map[string]*Session),
	}

	_, err := manager.launchProfile(&profile.Profile{
		ID:         "browseforge-chromium-dispatch",
		RuntimeID:  "browseforge-chromium",
		ProfileDir: t.TempDir(),
	})
	if !errors.Is(err, launchErr) {
		t.Fatalf("launch error = %v, want captured launch error", err)
	}
	if browserType.calls != 1 {
		t.Fatalf("launch calls = %d, want 1", browserType.calls)
	}
}

func TestLaunchChromiumAssemblesProxyFingerprintArgsWithoutLaunchingBrowser(t *testing.T) {
	enabled := true
	fontsDir := t.TempDir()
	launchErr := errors.New("captured launch")
	browserType := &capturingBrowserType{t: t, launchErr: launchErr}
	cfg := &config.Config{
		Runtimes: map[string]config.RuntimeConfig{
			"browseforge-chromium": {
				BinaryPath: filepath.Join(t.TempDir(), "browseforge-chromium"),
				Enabled:    &enabled,
				Settings: &config.CloakBrowserConfig{
					FingerprintPlatform:  "windows",
					FontsDir:             fontsDir,
					StorageQuotaMB:       2048,
					TargetPlatformPolicy: "allow",
					PluginsPDF:           "enabled",
				},
			},
		},
	}
	manager := &Manager{
		cfg:      cfg,
		runtimes: bfruntime.NewRegistry(cfg),
		pw:       &playwright.Playwright{Chromium: browserType},
		sessions: make(map[string]*Session),
	}

	profileDir := t.TempDir()
	_, err := manager.launchChromium(&profile.Profile{
		ID:        "proxy-fingerprint-args",
		RuntimeID: "browseforge-chromium",
		Proxy: &profile.ProxyConfig{
			Type:   "http",
			Host:   "127.0.0.1",
			Port:   1,
			Region: "  us-ny  ",
		},
		Fingerprint: map[string]any{
			"navigator.userAgent":           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/150.0.0.0 Safari/537.36",
			"navigator.platform":            "Win32",
			"navigator.languages":           []any{"en-US", "en"},
			"navigator.hardwareConcurrency": float64(8),
			"navigator.deviceMemory":        float64(8),
			"screen.width":                  float64(1920),
			"screen.height":                 float64(1080),
			"screen.availWidth":             float64(1920),
			"screen.availHeight":            float64(1032),
			"canvas:seed":                   float64(12345),
			"audio:seed":                    float64(67890),
			"fonts":                         []any{"Segoe UI", "Calibri", "Consolas"},
			"webGl:vendor":                  "Google Inc. (NVIDIA)",
			"webGl:renderer":                "ANGLE (NVIDIA, NVIDIA GeForce RTX 4050 Laptop GPU Direct3D11)",
		},
		ProfileDir: profileDir,
	})
	if !errors.Is(err, launchErr) {
		t.Fatalf("launch error = %v, want captured launch error", err)
	}
	if browserType.calls != 1 {
		t.Fatalf("launch calls = %d, want 1", browserType.calls)
	}
	fontsDirAbs, err := filepath.Abs(fontsDir)
	if err != nil {
		t.Fatalf("abs fonts dir: %v", err)
	}
	for _, want := range []string{
		chromiumWebRTCIPHandlingArg,
		"--fingerprint-webrtc-ip=auto",
		"--fingerprint-timezone=America/New_York",
		"--fingerprint-locale=en-US",
		"--fingerprint-platform=Win32",
		"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/150.0.0.0 Safari/537.36",
		"--fingerprint-user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/150.0.0.0 Safari/537.36",
		"--fingerprint-ua-full-version=150.0.0.0",
		"--fingerprint-ua-platform=Windows",
		"--fingerprint-ua-architecture=x86",
		"--fingerprint-ua-bitness=64",
		"--fingerprint-accept-language=en-US,en",
		"--fingerprint-hardware-concurrency=8",
		"--fingerprint-device-memory=8",
		"--fingerprint-screen-width=1920",
		"--fingerprint-screen-height=1080",
		"--fingerprint-screen-avail-width=1920",
		"--fingerprint-screen-avail-height=1032",
		"--force-device-scale-factor=1",
		"--fingerprint-screen-device-scale-factor=1",
		"--window-position=0,0",
		"--window-size=1920,1032",
		"--fingerprint-canvas-noise=12345",
		"--fingerprint-audio-noise=67890",
		"--fingerprint-fonts-list=Segoe UI|Calibri|Consolas",
		"--fingerprint-webgl-vendor=Google Inc. (NVIDIA)",
		"--fingerprint-webgl-renderer=ANGLE (NVIDIA, NVIDIA GeForce RTX 4050 Laptop GPU Direct3D11)",
		"--fingerprint-storage-quota=2048",
		"--fingerprint-plugins-pdf=enabled",
		"--fingerprint-fonts-dir=" + fontsDirAbs,
		"--browseforge-stealth-config=" + filepath.Join(profileDir, "browser-data", "BrowseForgeNative", "persona.json"),
		"--browseforge-stealth-mode=enabled",
	} {
		if !containsArg(browserType.options.Args, want) {
			t.Fatalf("launch args missing %q: %#v", want, browserType.options.Args)
		}
	}
	if browserType.options.NoViewport == nil || !*browserType.options.NoViewport {
		t.Fatalf("NoViewport = %#v, want true", browserType.options.NoViewport)
	}
	if browserType.options.Locale != nil {
		t.Fatalf("Locale = %#v, want nil so native/env locale owns Intl behavior", browserType.options.Locale)
	}
	if browserType.options.TimezoneId != nil {
		t.Fatalf("TimezoneId = %#v, want nil so native/env timezone owns Intl behavior", browserType.options.TimezoneId)
	}
	if browserType.options.Env["TZ"] != "America/New_York" || browserType.options.Env["BROWSEFORGE_INTL_LOCALE"] != "en-US" {
		t.Fatalf("launch env = %#v, want TZ and BROWSEFORGE_INTL_LOCALE", browserType.options.Env)
	}
	nativeConfigPath := filepath.Join(profileDir, "browser-data", "BrowseForgeNative", "persona.json")
	nativeConfigData, err := os.ReadFile(nativeConfigPath)
	if err != nil {
		t.Fatalf("read native config: %v", err)
	}
	var nativeConfig map[string]any
	if err := json.Unmarshal(nativeConfigData, &nativeConfig); err != nil {
		t.Fatalf("decode native config: %v", err)
	}
	if got := nativeConfig["runtime_id"]; got != "browseforge-chromium" {
		t.Fatalf("native runtime_id = %#v, want browseforge-chromium", got)
	}
	if got := nativeConfig["seed"]; got != float64(12345) {
		t.Fatalf("native seed = %#v, want 12345", got)
	}
	nativeLocale, ok := nativeConfig["locale"].(map[string]any)
	if !ok {
		t.Fatalf("native locale missing: %#v", nativeConfig)
	}
	if got := nativeLocale["geo_source"]; got != "proxy_region_fallback" {
		t.Fatalf("native locale geo_source = %#v, want proxy_region_fallback", got)
	}
	if got := nativeLocale["geo_status"]; got != "geo_provider_unavailable" {
		t.Fatalf("native locale geo_status = %#v, want geo_provider_unavailable", got)
	}
	for _, key := range []string{"persona_id_hash", "origin_salt_key"} {
		value, ok := nativeConfig[key].(string)
		if !ok || len(value) != 32 {
			t.Fatalf("native %s = %#v, want 32-char hex string", key, nativeConfig[key])
		}
	}
	nativeGPU, ok := nativeConfig["gpu"].(map[string]any)
	if !ok {
		t.Fatalf("native gpu missing: %#v", nativeConfig)
	}
	if got := nativeGPU["vendor"]; got != "Google Inc. (NVIDIA)" {
		t.Fatalf("native gpu vendor = %#v", got)
	}
	if got := nativeGPU["renderer"]; got != "ANGLE (NVIDIA, NVIDIA GeForce RTX 4050 Laptop GPU Direct3D11)" {
		t.Fatalf("native gpu renderer = %#v", got)
	}
	nativeWebRTC, ok := nativeConfig["webrtc"].(map[string]any)
	if !ok {
		t.Fatalf("native webrtc missing: %#v", nativeConfig)
	}
	if got := nativeWebRTC["direct_ip_redaction"]; got != true {
		t.Fatalf("native webrtc direct_ip_redaction = %#v, want true", got)
	}
	if got := nativeWebRTC["proxy_region"]; got != "us-ny" {
		t.Fatalf("native webrtc proxy_region = %#v, want us-ny", got)
	}
	nativeStorage, ok := nativeConfig["storage"].(map[string]any)
	if !ok {
		t.Fatalf("native storage missing: %#v", nativeConfig)
	}
	if got := nativeStorage["quota_mb"]; got != float64(2048) {
		t.Fatalf("native storage quota_mb = %#v, want 2048", got)
	}
	if got := nativeStorage["persistent"]; got != true {
		t.Fatalf("native storage persistent = %#v, want true", got)
	}
	if browserType.options.Proxy == nil {
		t.Fatalf("launch proxy was not configured")
	}
	if got := browserType.options.Proxy.Server; got != "http://127.0.0.1:1" {
		t.Fatalf("proxy server = %q, want http://127.0.0.1:1", got)
	}
	prefsPath := filepath.Join(profileDir, "browser-data", "Default", "Preferences")
	prefsData, err := os.ReadFile(prefsPath)
	if err != nil {
		t.Fatalf("read prefs: %v", err)
	}
	var prefs map[string]any
	if err := json.Unmarshal(prefsData, &prefs); err != nil {
		t.Fatalf("decode prefs: %v", err)
	}
	webrtcPrefs, ok := prefs["webrtc"].(map[string]any)
	if !ok {
		t.Fatalf("webrtc prefs missing: %#v", prefs)
	}
	if got := webrtcPrefs["ip_handling_policy"]; got != "disable_non_proxied_udp" {
		t.Fatalf("webrtc ip_handling_policy = %#v, want disable_non_proxied_udp", got)
	}
	if got := webrtcPrefs["multiple_routes_enabled"]; got != false {
		t.Fatalf("webrtc multiple_routes_enabled = %#v, want false", got)
	}
	if got := webrtcPrefs["nonproxied_udp_enabled"]; got != false {
		t.Fatalf("webrtc nonproxied_udp_enabled = %#v, want false", got)
	}
}

func TestResolveCloakFontsDirExplicitDirectory(t *testing.T) {
	fontsDir := filepath.Join(t.TempDir(), "fonts")
	if err := os.MkdirAll(fontsDir, 0755); err != nil {
		t.Fatalf("mkdir fonts dir: %v", err)
	}

	got, err := resolveCloakFontsDir(&config.CloakBrowserConfig{FontsDir: fontsDir})
	if err != nil {
		t.Fatalf("resolve fonts dir: %v", err)
	}
	want, err := filepath.Abs(fontsDir)
	if err != nil {
		t.Fatalf("abs fonts dir: %v", err)
	}
	if got != want {
		t.Fatalf("fonts dir = %q, want %q", got, want)
	}
}

func TestResolveCloakFontsDirExplicitMissingPathFails(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-fonts")

	_, err := resolveCloakFontsDir(&config.CloakBrowserConfig{FontsDir: missing})
	if err == nil {
		t.Fatal("expected missing explicit fonts dir to fail")
	}
	if !strings.Contains(err.Error(), "fonts_dir unavailable") {
		t.Fatalf("error = %q, want unavailable fonts dir", err.Error())
	}
}

func TestResolveCloakFontsDirUsesSystemFontsWhenPresent(t *testing.T) {
	info, err := os.Stat("/usr/share/fonts")
	if err != nil || !info.IsDir() {
		t.Skip("/usr/share/fonts is not a directory on this host")
	}

	got, err := resolveCloakFontsDir(nil)
	if err != nil {
		t.Fatalf("resolve default fonts dir: %v", err)
	}
	if got != "/usr/share/fonts" {
		t.Fatalf("fonts dir = %q, want /usr/share/fonts", got)
	}
}

func TestResolveCloakFingerprintPlatform(t *testing.T) {
	cases := []struct {
		name   string
		policy *config.CloakBrowserConfig
		goos   string
		want   string
	}{
		{
			name:   "auto uses macos on darwin",
			policy: &config.CloakBrowserConfig{FingerprintPlatform: "auto"},
			goos:   "darwin",
			want:   "MacIntel",
		},
		{
			name:   "empty policy uses windows-compatible profile off darwin",
			policy: &config.CloakBrowserConfig{},
			goos:   "linux",
			want:   "Win32",
		},
		{
			name:   "explicit linux is preserved on darwin",
			policy: &config.CloakBrowserConfig{FingerprintPlatform: "linux"},
			goos:   "darwin",
			want:   "Linux x86_64",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveCloakFingerprintPlatform(tc.policy, tc.goos)
			if err != nil {
				t.Fatalf("resolve platform: %v", err)
			}
			if got != tc.want {
				t.Fatalf("platform = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestResolveChromiumFingerprintPlatformUsesNativeLinuxArch(t *testing.T) {
	cases := []struct {
		name   string
		policy *config.CloakBrowserConfig
		goos   string
		goarch string
		want   string
	}{
		{name: "auto linux amd64", policy: &config.CloakBrowserConfig{FingerprintPlatform: "auto"}, goos: "linux", goarch: "amd64", want: "Linux x86_64"},
		{name: "auto linux arm64", policy: &config.CloakBrowserConfig{FingerprintPlatform: "auto"}, goos: "linux", goarch: "arm64", want: "Linux aarch64"},
		{name: "explicit linux arm64", policy: &config.CloakBrowserConfig{FingerprintPlatform: "linux"}, goos: "linux", goarch: "arm64", want: "Linux aarch64"},
		{name: "explicit windows on arm64", policy: &config.CloakBrowserConfig{FingerprintPlatform: "windows"}, goos: "linux", goarch: "arm64", want: "Win32"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveChromiumFingerprintPlatform(tc.policy, tc.goos, tc.goarch)
			if err != nil {
				t.Fatalf("resolve platform: %v", err)
			}
			if got != tc.want {
				t.Fatalf("platform = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestChromiumLaunchPersonaPlatformMappings(t *testing.T) {
	cases := []struct {
		name     string
		platform string
		goarch   string
		want     browseForgeNativePlatform
	}{
		{
			name:     "windows",
			platform: "Win32",
			want:     browseForgeNativePlatform{OS: "windows", Arch: "x86", Platform: "Win32", PlatformCH: "Windows", Bitness: "64"},
		},
		{
			name:     "macos x64",
			platform: "MacIntel",
			goarch:   "amd64",
			want:     browseForgeNativePlatform{OS: "macos", Arch: "x86", Platform: "MacIntel", PlatformCH: "macOS", Bitness: "64"},
		},
		{
			name:     "macos arm64",
			platform: "MacIntel",
			goarch:   "arm64",
			want:     browseForgeNativePlatform{OS: "macos", Arch: "arm", Platform: "MacIntel", PlatformCH: "macOS", Bitness: "64"},
		},
		{
			name:     "linux x64",
			platform: "Linux x86_64",
			want:     browseForgeNativePlatform{OS: "linux", Arch: "x86", Platform: "Linux x86_64", PlatformCH: "Linux", Bitness: "64"},
		},
		{
			name:     "linux arm64",
			platform: "Linux aarch64",
			want:     browseForgeNativePlatform{OS: "linux", Arch: "arm", Platform: "Linux aarch64", PlatformCH: "Linux", Bitness: "64"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			persona, err := buildChromiumLaunchPersona(
				&profile.Profile{ID: "platform-" + tc.name, RuntimeID: "browseforge-chromium"},
				bfruntime.BrowseForgeChromium,
				tc.platform,
				"UTC",
				"en-US",
				"",
				tc.goarch,
				nil,
			)
			if err != nil {
				t.Fatalf("build persona: %v", err)
			}
			if persona.Native.Platform != tc.want {
				t.Fatalf("platform = %#v, want %#v", persona.Native.Platform, tc.want)
			}
			args := appendChromiumLaunchPersonaArgs(nil, persona)
			for _, wantArg := range []string{
				"--fingerprint-platform=" + tc.want.Platform,
				"--fingerprint-ua-platform=" + tc.want.PlatformCH,
				"--fingerprint-ua-architecture=" + tc.want.Arch,
				"--fingerprint-ua-bitness=" + tc.want.Bitness,
			} {
				if !containsArg(args, wantArg) {
					t.Fatalf("args missing %q: %#v", wantArg, args)
				}
			}
		})
	}
}

func TestChromiumLaunchPersonaRepairsIncompatiblePoolValues(t *testing.T) {
	cases := []struct {
		name      string
		userAgent string
	}{
		{
			name:      "windows pool on linux arm64",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/150.0.0.0 Safari/537.36",
		},
		{
			name:      "linux x86 pool on linux arm64",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/150.0.0.0 Safari/537.36",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			persona, err := buildChromiumLaunchPersona(
				&profile.Profile{
					ID:        "repair-" + tc.name,
					RuntimeID: "browseforge-chromium",
					Fingerprint: map[string]any{
						"navigator.userAgent": tc.userAgent,
						"navigator.platform":  "Win32",
						"navigator.languages": []any{"en-US", "en"},
					},
				},
				bfruntime.BrowseForgeChromium,
				"Linux aarch64",
				"Asia/Taipei",
				"zh-TW",
				"",
				"arm64",
				nil,
			)
			if err != nil {
				t.Fatalf("build persona: %v", err)
			}
			if !strings.Contains(persona.Native.Browser.UserAgent, "Linux aarch64") {
				t.Fatalf("user agent = %q, want Linux arm64 repair", persona.Native.Browser.UserAgent)
			}
			if persona.Native.Platform.Platform != "Linux aarch64" || persona.Native.Platform.Arch != "arm" {
				t.Fatalf("platform = %#v, want Linux aarch64/arm", persona.Native.Platform)
			}
			if got := persona.Native.Locale.AcceptLanguage; got != "zh-TW,zh;q=0.9" {
				t.Fatalf("accept language = %q, want zh-TW repair", got)
			}
			args := appendChromiumLaunchPersonaArgs(nil, persona)
			for _, forbidden := range []string{"--fingerprint-accept-language=en-US,en", "--fingerprint-platform=Win32"} {
				if containsArg(args, forbidden) {
					t.Fatalf("args preserved incompatible value %q: %#v", forbidden, args)
				}
			}
			if !containsArg(args, "--fingerprint-accept-language=zh-TW,zh") {
				t.Fatalf("args missing q-stripped Chromium accept-language switch: %#v", args)
			}
		})
	}
}

func TestChromiumLaunchPersonaPreservesCompatiblePoolValues(t *testing.T) {
	const linuxArmUA = "Mozilla/5.0 (X11; Linux aarch64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36"
	persona, err := buildChromiumLaunchPersona(
		&profile.Profile{
			ID:        "compatible-pool",
			RuntimeID: "browseforge-chromium",
			Fingerprint: map[string]any{
				"navigator.userAgent":           linuxArmUA,
				"navigator.platform":            "Linux aarch64",
				"navigator.languages":           []any{"zh-TW", "zh"},
				"navigator.hardwareConcurrency": float64(12),
				"navigator.deviceMemory":        float64(16),
				"screen.width":                  float64(1440),
				"screen.height":                 float64(900),
				"screen.availWidth":             float64(1440),
				"screen.availHeight":            float64(860),
				"canvas:seed":                   float64(111),
				"audio:seed":                    float64(222),
				"fonts":                         []any{"Noto Sans", "Arial"},
				"webGl:vendor":                  "Google Inc. (AMD)",
				"webGl:renderer":                "ANGLE (AMD, Radeon, Vulkan)",
			},
		},
		bfruntime.BrowseForgeChromium,
		"Linux aarch64",
		"Asia/Taipei",
		"zh-TW",
		"tw-taipei",
		"arm64",
		&config.CloakBrowserConfig{StorageQuotaMB: 4096},
	)
	if err != nil {
		t.Fatalf("build persona: %v", err)
	}
	if got := persona.Native.Browser.UserAgent; got != linuxArmUA {
		t.Fatalf("user agent = %q, want profile UA", got)
	}
	if got := persona.Native.Locale.AcceptLanguage; got != "zh-TW,zh" {
		t.Fatalf("accept language = %q, want profile language chain", got)
	}
	if got := persona.Native.Storage.QuotaMB; got != 4096 {
		t.Fatalf("storage quota = %d, want 4096", got)
	}
	args := appendChromiumLaunchPersonaArgs(nil, persona)
	for _, want := range []string{
		"--fingerprint-hardware-concurrency=12",
		"--fingerprint-device-memory=16",
		"--fingerprint-screen-width=1440",
		"--fingerprint-screen-height=900",
		"--fingerprint-screen-avail-width=1440",
		"--fingerprint-screen-avail-height=860",
		"--fingerprint-canvas-noise=111",
		"--fingerprint-audio-noise=222",
		"--fingerprint-fonts-list=Noto Sans|Arial",
		"--fingerprint-webgl-vendor=Google Inc. (AMD)",
		"--fingerprint-webgl-renderer=ANGLE (AMD, Radeon, Vulkan)",
		"--fingerprint-storage-quota=4096",
	} {
		if !containsArg(args, want) {
			t.Fatalf("args missing %q: %#v", want, args)
		}
	}
}

func TestChromiumLaunchPersonaClampsScreenAvailToScreenSize(t *testing.T) {
	persona, err := buildChromiumLaunchPersona(
		&profile.Profile{
			ID:        "screen-clamp",
			RuntimeID: "browseforge-chromium",
			Fingerprint: map[string]any{
				"screen.width":       float64(1366),
				"screen.height":      float64(768),
				"screen.availWidth":  float64(1920),
				"screen.availHeight": float64(900),
			},
		},
		bfruntime.BrowseForgeChromium,
		"Linux x86_64",
		"UTC",
		"en-US",
		"",
		"amd64",
		nil,
	)
	if err != nil {
		t.Fatalf("build persona: %v", err)
	}
	if got := persona.Native.Screen.AvailWidth; got != 1366 {
		t.Fatalf("avail width = %d, want clamped 1366", got)
	}
	if got := persona.Native.Screen.AvailHeight; got != 768 {
		t.Fatalf("avail height = %d, want clamped 768", got)
	}
	args := appendChromiumLaunchPersonaArgs(nil, persona)
	for _, want := range []string{"--fingerprint-screen-avail-width=1366", "--fingerprint-screen-avail-height=768"} {
		if !containsArg(args, want) {
			t.Fatalf("args missing clamped value %q: %#v", want, args)
		}
	}
}

func TestLaunchChromiumRejectsInvalidBrowseForgeProxyRegionBeforeLaunch(t *testing.T) {
	enabled := true
	browserType := &capturingBrowserType{t: t, launchErr: errors.New("should not launch")}
	cfg := &config.Config{
		Runtimes: map[string]config.RuntimeConfig{
			"browseforge-chromium": {
				BinaryPath: filepath.Join(t.TempDir(), "browseforge-chromium"),
				Enabled:    &enabled,
				Settings: &config.CloakBrowserConfig{
					TargetPlatformPolicy: "allow",
				},
			},
		},
	}
	manager := &Manager{
		cfg:      cfg,
		runtimes: bfruntime.NewRegistry(cfg),
		pw:       &playwright.Playwright{Chromium: browserType},
		sessions: make(map[string]*Session),
	}

	_, err := manager.launchChromium(&profile.Profile{
		ID:        "invalid-proxy-region",
		RuntimeID: "browseforge-chromium",
		Proxy: &profile.ProxyConfig{
			Type:   "http",
			Host:   "127.0.0.1",
			Port:   1,
			Region: "192.0.2.1",
		},
		ProfileDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected invalid proxy region to fail")
	}
	if !strings.Contains(err.Error(), "proxy_region") {
		t.Fatalf("error = %q, want proxy_region validation", err.Error())
	}
	if browserType.calls != 0 {
		t.Fatalf("launch calls = %d, want 0", browserType.calls)
	}
}

func TestEffectiveChromiumIdentityKeepsPlatformAndLocaleConsistent(t *testing.T) {
	fp := map[string]any{
		"navigator.userAgent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36",
		"navigator.languages": []any{"en-US", "en"},
	}
	userAgent := effectiveChromiumUserAgent(fp, "Linux aarch64")
	if !strings.Contains(userAgent, "Linux aarch64") {
		t.Fatalf("user agent = %q, want Linux arm64 default", userAgent)
	}
	if got := effectiveChromiumAcceptLanguage(fp, "zh-TW"); got != "zh-TW,zh;q=0.9" {
		t.Fatalf("accept language = %q, want zh-TW,zh;q=0.9", got)
	}
	linux64 := "Mozilla/5.0 (X11; Linux aarch64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36"
	fp["navigator.userAgent"] = linux64
	fp["navigator.languages"] = []any{"zh-TW", "zh"}
	if got := effectiveChromiumUserAgent(fp, "Linux aarch64"); got != linux64 {
		t.Fatalf("user agent = %q, want profile arm64 UA", got)
	}
	if got := effectiveChromiumAcceptLanguage(fp, "zh-TW"); got != "zh-TW,zh" {
		t.Fatalf("accept language = %q, want profile zh-TW chain", got)
	}
}

func TestBrowseForgeLaunchOptionsExposeNativeLocaleAndRealViewport(t *testing.T) {
	t.Setenv("DISPLAY", ":1")
	t.Setenv("HOME", "/tmp/browseforge-home")
	t.Setenv("LIBGL_ALWAYS_SOFTWARE", "1")
	persona := chromiumLaunchPersona{
		Native: browseForgeNativePersonaConfig{
			Locale: browseForgeNativeLocale{
				Timezone:       "Asia/Taipei",
				Locale:         "zh-TW",
				AcceptLanguage: "zh-TW,zh;q=0.9",
			},
			Screen: browseForgeNativeScreen{
				Width:       1920,
				Height:      1080,
				AvailWidth:  1920,
				AvailHeight: 1040,
			},
		},
	}

	env := browseForgeChromiumEnv(persona)
	if env["TZ"] != "Asia/Taipei" {
		t.Fatalf("TZ env = %q, want Asia/Taipei", env["TZ"])
	}
	if env["LANG"] != "zh_TW.UTF-8" || env["LC_ALL"] != "zh_TW.UTF-8" {
		t.Fatalf("locale env = LANG %q LC_ALL %q, want zh_TW.UTF-8", env["LANG"], env["LC_ALL"])
	}
	if env["BROWSEFORGE_INTL_LOCALE"] != "zh-TW" {
		t.Fatalf("BROWSEFORGE_INTL_LOCALE = %q, want zh-TW", env["BROWSEFORGE_INTL_LOCALE"])
	}
	if env["DISPLAY"] != ":1" || env["HOME"] != "/tmp/browseforge-home" || env["LIBGL_ALWAYS_SOFTWARE"] != "1" {
		t.Fatalf("display env = DISPLAY %q HOME %q LIBGL_ALWAYS_SOFTWARE %q", env["DISPLAY"], env["HOME"], env["LIBGL_ALWAYS_SOFTWARE"])
	}
	windowArgs := browseForgeChromiumWindowArgs(persona)
	for _, want := range []string{"--window-position=0,0", "--window-size=1920,1040"} {
		if !containsArg(windowArgs, want) {
			t.Fatalf("window args = %#v, want %q", windowArgs, want)
		}
	}
}

func TestBrowseForgeDockerSoftwareModeUsesSwiftShaderPersona(t *testing.T) {
	t.Setenv("BROWSEFORGE_DOCKER_GPU_MODE", "software")
	p := &profile.Profile{Fingerprint: map[string]any{
		"navigator.userAgent": "Mozilla/5.0 (X11; Linux aarch64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.7871.101 Safari/537.36",
	}}

	persona, err := buildChromiumLaunchPersona(p, bfruntime.BrowseForgeChromium, "Linux aarch64", "Asia/Taipei", "zh-TW", "", "arm64", nil)
	if err != nil {
		t.Fatalf("build persona: %v", err)
	}
	if !persona.HasWebGLVendor || !persona.HasWebGLRenderer {
		t.Fatalf("software Docker persona must explicitly own WebGL strings: %#v", persona)
	}
	if persona.Native.GPU.Vendor != "Google Inc. (Google)" {
		t.Fatalf("GPU vendor = %q, want SwiftShader vendor", persona.Native.GPU.Vendor)
	}
	if !strings.Contains(persona.Native.GPU.Renderer, "SwiftShader") {
		t.Fatalf("GPU renderer = %q, want SwiftShader renderer", persona.Native.GPU.Renderer)
	}
}

func TestBrowseForgeDockerSoftwareModeEnablesSwiftShaderWebGLFallback(t *testing.T) {
	t.Setenv("BROWSEFORGE_DOCKER_GPU_MODE", "software")
	args := appendBrowseForgeDockerSoftwareGPUArgs([]string{"--no-first-run"})

	for _, want := range []string{
		"--no-first-run",
		"--use-gl=angle",
		"--use-angle=swiftshader-webgl",
		"--enable-unsafe-swiftshader",
	} {
		if !containsArg(args, want) {
			t.Fatalf("software GPU args missing %q: %#v", want, args)
		}
	}
}

func TestBrowseForgeDockerNativeModeDoesNotForceSwiftShaderWebGLFallback(t *testing.T) {
	t.Setenv("BROWSEFORGE_DOCKER_GPU_MODE", "native")
	args := appendBrowseForgeDockerSoftwareGPUArgs([]string{"--no-first-run"})

	for _, forbidden := range []string{
		"--use-gl=angle",
		"--use-angle=swiftshader-webgl",
		"--enable-unsafe-swiftshader",
	} {
		if containsArg(args, forbidden) {
			t.Fatalf("native GPU args unexpectedly contain %q: %#v", forbidden, args)
		}
	}
}

func TestResolveCloakFingerprintPlatformRejectsInvalidValues(t *testing.T) {
	_, err := resolveCloakFingerprintPlatform(&config.CloakBrowserConfig{FingerprintPlatform: "ios"}, "darwin")
	if err == nil {
		t.Fatal("expected invalid fingerprint platform to fail")
	}
	if !strings.Contains(err.Error(), "auto, macos, windows, or linux") {
		t.Fatalf("error = %q, want allowed platform list", err.Error())
	}
}

func TestValidateCloakFingerprintPolicyTargetPlatformPolicy(t *testing.T) {
	cases := []struct {
		name      string
		policy    config.CloakBrowserConfig
		platform  string
		goos      string
		wantError string
	}{
		{
			name:      "invalid mode is rejected",
			policy:    config.CloakBrowserConfig{TargetPlatformPolicy: "audit"},
			platform:  "Win32",
			goos:      "linux",
			wantError: "target_platform_policy must be strict, warn, or allow",
		},
		{
			name:      "invalid plugins PDF mode is rejected",
			policy:    config.CloakBrowserConfig{PluginsPDF: "maybe"},
			platform:  "MacIntel",
			goos:      "darwin",
			wantError: "plugins_pdf must be enabled/true/1 or disabled/false/0",
		},
		{
			name:      "strict rejects windows fingerprint on non-windows host without fonts",
			policy:    config.CloakBrowserConfig{TargetPlatformPolicy: "strict"},
			platform:  "Win32",
			goos:      "linux",
			wantError: "Windows CloakBrowser fingerprint on non-Windows host should configure runtimes.cloakbrowser.settings.fonts_dir",
		},
		{
			name:     "warn allows windows fingerprint on non-windows host without fonts",
			policy:   config.CloakBrowserConfig{TargetPlatformPolicy: "warn"},
			platform: "Win32",
			goos:     "linux",
		},
		{
			name:     "allow bypasses windows font-pack validation",
			policy:   config.CloakBrowserConfig{TargetPlatformPolicy: "allow"},
			platform: "Win32",
			goos:     "linux",
		},
		{
			name:     "strict allows windows fingerprint on non-windows host with fonts",
			policy:   config.CloakBrowserConfig{TargetPlatformPolicy: "strict", FontsDir: "C:\\Windows\\Fonts"},
			platform: "Win32",
			goos:     "linux",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCloakFingerprintPolicy(&tc.policy, tc.platform, tc.goos)
			if tc.wantError == "" {
				if err != nil {
					t.Fatalf("validate policy: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q", tc.wantError)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tc.wantError)
			}
		})
	}
}

func TestApplyCloakBrowserLaunchPolicyKeepsManagedFingerprintArgsOwnedByManager(t *testing.T) {
	baseArgs := []string{
		"--fingerprint=4242",
		"--fingerprint-platform=macos",
		"--fingerprint-fonts-dir=/managed/fonts",
		"--fingerprint-storage-quota=4096",
		"--fingerprint-timezone=Asia/Taipei",
		"--fingerprint-locale=zh-TW",
		"--fingerprint-webrtc-ip=auto",
		"--fingerprint-screen-width=1440",
		"--fingerprint-screen-height=900",
		"--fingerprint-hardware-concurrency=10",
		"--fingerprint-device-memory=8",
		"--fingerprint-screen-avail-width=1440",
		"--fingerprint-screen-avail-height=860",
		"--fingerprint-accept-language=zh-TW,zh",
		"--fingerprint-user-agent=Mozilla/5.0",
		"--fingerprint-audio-noise=222",
		"--fingerprint-canvas-noise=333",
		"--fingerprint-webgl-vendor=Google Inc. (NVIDIA)",
		"--fingerprint-fonts-list=Segoe UI|Calibri",
		"--fingerprint-webgl-renderer=ANGLE (NVIDIA)",
	}
	extraArgs := []string{
		"--fingerprint=1",
		"--fingerprint-platform=linux",
		"--fingerprint-fonts-dir=/tmp/fonts",
		"--fingerprint-storage-quota=1",
		"--fingerprint-timezone=UTC",
		"--fingerprint-locale=en-US",
		"--fingerprint-webrtc-ip=public",
		"--fingerprint-screen-width=1",
		"--fingerprint-screen-height=1",
		"--fingerprint-hardware-concurrency=1",
		"--fingerprint-device-memory=1",
		"--fingerprint-screen-avail-width=1",
		"--fingerprint-screen-avail-height=1",
		"--fingerprint-accept-language=en-US,en",
		"--fingerprint-user-agent=evil",
		"--fingerprint-audio-noise=1",
		"--fingerprint-canvas-noise=1",
		"--fingerprint-webgl-vendor=evil",
		"--fingerprint-fonts-list=evil",
		"--fingerprint-webgl-renderer=evil",
		"--disable-background-networking",
	}

	args, err := applyCloakBrowserLaunchPolicy(
		baseArgs,
		t.TempDir(),
		&config.CloakBrowserConfig{ExtraArgs: extraArgs},
		false,
	)
	if err != nil {
		t.Fatalf("apply policy: %v", err)
	}

	for _, want := range append(baseArgs, "--disable-background-networking") {
		if !containsArg(args, want) {
			t.Fatalf("args missing %q: %#v", want, args)
		}
	}
	for _, blocked := range extraArgs[:len(extraArgs)-1] {
		if containsArg(args, blocked) {
			t.Fatalf("managed fingerprint extra arg was not filtered: %q in %#v", blocked, args)
		}
	}
}

func TestCamoufoxEnvChunksLargeNonASCIIConfigWithoutCorruption(t *testing.T) {
	const jsonPrefix = `{"label":"`
	paddingLen := camouConfigChunkSize - len(jsonPrefix) - 1
	if paddingLen < 0 {
		t.Fatalf("test setup cannot place a multibyte rune across the chunk boundary")
	}
	configText := jsonPrefix + strings.Repeat("x", paddingLen) + "世" + strings.Repeat("y", camouConfigChunkSize) + `"}`
	configJSON := []byte(configText)
	if !utf8.Valid(configJSON) {
		t.Fatal("test setup produced invalid UTF-8 JSON")
	}

	env := camoufoxEnv(configJSON, map[string]string{"HOME": "/tmp/home"})

	if _, ok := env["CAMOU_CONFIG"]; ok {
		t.Fatalf("CAMOU_CONFIG should not be set when config is chunked: %#v", env)
	}
	if env["HOME"] != "/tmp/home" {
		t.Fatalf("base env was not preserved: %#v", env)
	}
	chunks := []string{
		env["CAMOU_CONFIG_1"],
		env["CAMOU_CONFIG_2"],
		env["CAMOU_CONFIG_3"],
	}
	if _, ok := env["CAMOU_CONFIG_4"]; ok {
		t.Fatalf("unexpected fourth CAMOU_CONFIG chunk: %#v", env)
	}
	var rebuilt strings.Builder
	for i, chunk := range chunks {
		if chunk == "" {
			t.Fatalf("CAMOU_CONFIG_%d was not set: %#v", i+1, env)
		}
		if !utf8.ValidString(chunk) {
			t.Fatalf("CAMOU_CONFIG_%d is not valid UTF-8; len=%d", i+1, len(chunk))
		}
		rebuilt.WriteString(chunk)
	}
	got := rebuilt.String()
	if got != configText {
		t.Fatalf("chunked config did not round trip: got len %d, want %d", len(got), len(configText))
	}
	if strings.Contains(got, "\uFFFD") {
		t.Fatalf("chunked config contains replacement characters after reconstruction")
	}
}

func TestNormalizeCamouWebGLProfileDropsOnlyPartialWebGL2(t *testing.T) {
	config := map[string]any{
		"webGl:renderer":               "webgl-renderer",
		"webGl:vendor":                 "webgl-vendor",
		"webGl:supportedExtensions":    "webgl-extensions",
		"webGl:parameters":             "webgl-parameters",
		"webGl:shaderPrecisionFormats": "webgl-precision",
		"webGl:contextAttributes":      "webgl-attributes",
		"webGl2:renderer":              "partial-webgl2-renderer",
		"webGl2:parameters":            "partial-webgl2-parameters",
		"navigator:hardwareMemory":     8,
		"screen:width":                 1920,
	}

	normalizeCamouWebGLProfile(config)

	wantWebGL := map[string]any{
		"webGl:renderer":               "webgl-renderer",
		"webGl:vendor":                 "webgl-vendor",
		"webGl:supportedExtensions":    "webgl-extensions",
		"webGl:parameters":             "webgl-parameters",
		"webGl:shaderPrecisionFormats": "webgl-precision",
		"webGl:contextAttributes":      "webgl-attributes",
	}
	for key, want := range wantWebGL {
		if got, ok := config[key]; !ok || got != want {
			t.Fatalf("complete WebGL1 key %q = %#v, present=%v; want %#v in %#v", key, got, ok, want, config)
		}
	}
	for key := range config {
		if strings.HasPrefix(key, "webGl2:") {
			t.Fatalf("partial WebGL2 key %q was not removed: %#v", key, config)
		}
	}
	if got := config["navigator:hardwareMemory"]; got != 8 {
		t.Fatalf("non-WebGL hardware memory was not preserved: got %#v in %#v", got, config)
	}
	if got := config["screen:width"]; got != 1920 {
		t.Fatalf("non-WebGL screen width was not preserved: got %#v in %#v", got, config)
	}
}

func TestNormalizeCamouWebGLProfilePreservesCompleteWebGL2Profile(t *testing.T) {
	config := map[string]any{
		"webGl2:renderer":               "webgl2-renderer",
		"webGl2:supportedExtensions":    "webgl2-extensions",
		"webGl2:parameters":             "webgl2-parameters",
		"webGl2:shaderPrecisionFormats": "webgl2-precision",
		"webGl2:contextAttributes":      "webgl2-attributes",
		"navigator:platform":            "Linux x86_64",
	}

	normalizeCamouWebGLProfile(config)

	wantWebGL2 := map[string]any{
		"webGl2:renderer":               "webgl2-renderer",
		"webGl2:supportedExtensions":    "webgl2-extensions",
		"webGl2:parameters":             "webgl2-parameters",
		"webGl2:shaderPrecisionFormats": "webgl2-precision",
		"webGl2:contextAttributes":      "webgl2-attributes",
	}
	for key, want := range wantWebGL2 {
		if got, ok := config[key]; !ok || got != want {
			t.Fatalf("complete WebGL2 key %q = %#v, present=%v; want %#v in %#v", key, got, ok, want, config)
		}
	}
	if got := config["navigator:platform"]; got != "Linux x86_64" {
		t.Fatalf("non-WebGL platform was not preserved: got %#v in %#v", got, config)
	}
}

func TestNormalizeCamouWebGLProfileDropsOnlyPartialWebGL1(t *testing.T) {
	config := map[string]any{
		"webGl:renderer":                "partial-webgl-renderer",
		"webGl:parameters":              "partial-webgl-parameters",
		"webGl2:renderer":               "webgl2-renderer",
		"webGl2:supportedExtensions":    "webgl2-extensions",
		"webGl2:parameters":             "webgl2-parameters",
		"webGl2:shaderPrecisionFormats": "webgl2-precision",
		"webGl2:contextAttributes":      "webgl2-attributes",
		"navigator:platform":            "Linux x86_64",
	}

	normalizeCamouWebGLProfile(config)

	for key := range config {
		if strings.HasPrefix(key, "webGl:") {
			t.Fatalf("partial WebGL1 key %q was not removed: %#v", key, config)
		}
	}
	wantWebGL2 := map[string]any{
		"webGl2:renderer":               "webgl2-renderer",
		"webGl2:supportedExtensions":    "webgl2-extensions",
		"webGl2:parameters":             "webgl2-parameters",
		"webGl2:shaderPrecisionFormats": "webgl2-precision",
		"webGl2:contextAttributes":      "webgl2-attributes",
	}
	for key, want := range wantWebGL2 {
		if got, ok := config[key]; !ok || got != want {
			t.Fatalf("complete WebGL2 key %q = %#v, present=%v; want %#v in %#v", key, got, ok, want, config)
		}
	}
	if got := config["navigator:platform"]; got != "Linux x86_64" {
		t.Fatalf("non-WebGL platform was not preserved: got %#v in %#v", got, config)
	}
}

type capturingBrowserType struct {
	t           *testing.T
	launchErr   error
	calls       int
	userDataDir string
	options     playwright.BrowserTypeLaunchPersistentContextOptions
}

func (b *capturingBrowserType) Connect(string, ...playwright.BrowserTypeConnectOptions) (playwright.Browser, error) {
	panic("unexpected BrowserType.Connect call")
}

func (b *capturingBrowserType) ConnectOverCDP(string, ...playwright.BrowserTypeConnectOverCDPOptions) (playwright.Browser, error) {
	panic("unexpected BrowserType.ConnectOverCDP call")
}

func (b *capturingBrowserType) ExecutablePath() string {
	panic("unexpected BrowserType.ExecutablePath call")
}

func (b *capturingBrowserType) Launch(...playwright.BrowserTypeLaunchOptions) (playwright.Browser, error) {
	panic("unexpected BrowserType.Launch call")
}

func (b *capturingBrowserType) LaunchPersistentContext(userDataDir string, options ...playwright.BrowserTypeLaunchPersistentContextOptions) (playwright.BrowserContext, error) {
	b.calls++
	b.userDataDir = userDataDir
	if len(options) != 1 {
		b.t.Fatalf("launch options len = %d, want 1", len(options))
	}
	b.options = options[0]
	return nil, b.launchErr
}

func (b *capturingBrowserType) Name() string {
	return "chromium"
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}
