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

	"github.com/playwright-community/playwright-go"
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
			Type: "http",
			Host: "127.0.0.1",
			Port: 1,
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
		chromiumAutomationControlledArg,
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
