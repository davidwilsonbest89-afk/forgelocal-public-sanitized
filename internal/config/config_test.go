package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigAdvertisesBrowseForgeChromium(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "config.default.json"))
	if err != nil {
		t.Fatalf("load default config: %v", err)
	}

	rt, ok := cfg.Runtimes["browseforge-chromium"]
	if !ok {
		t.Fatal("default config missing runtimes.browseforge-chromium")
	}
	if rt.DisplayName != "BrowseForge Chromium" || rt.Family != "chromium" {
		t.Fatalf("browseforge-chromium metadata = %#v", rt)
	}
	if rt.BinaryPath != "" {
		t.Fatalf("browseforge-chromium binary_path = %q, want empty until configured", rt.BinaryPath)
	}
	if rt.Enabled == nil || *rt.Enabled {
		t.Fatalf("browseforge-chromium enabled = %v, want explicit false", rt.Enabled)
	}
	if rt.Settings == nil || rt.Settings.PluginsPDF != "enabled" {
		t.Fatalf("browseforge-chromium plugins_pdf = %#v, want enabled", rt.Settings)
	}
}

func TestLoadCloakBrowserConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{
  "port": "19280",
  "profiles_dir": "profiles",
  "data_dir": "data",
  "log_file": "logs/server.log",
  "fingerprint_dir": "data",
  "runtimes": {
    "camoufox": {
      "binary_path": "/opt/browseforge/camoufox"
    },
    "cloakbrowser": {
      "binary_path": "/opt/browseforge/cloakbrowser",
      "settings": {
        "safe_gpu": true,
        "auto_safe_gpu_fallback": true,
        "isolated_runtime_cache": true,
        "repair_transient_cache_on_launch_failure": true,
        "fingerprint_platform": "macos",
        "fonts_dir": "/opt/browseforge/fonts",
        "storage_quota_mb": 2048,
        "target_platform_policy": "warn",
        "plugins_pdf": "enabled",
        "extra_args": ["--disable-features=Translate"]
      }
    }
  }
}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	settings := cfg.CloakBrowserSettings()
	if settings == nil {
		t.Fatal("cloakbrowser settings missing")
	}
	if !settings.SafeGPU ||
		!settings.AutoSafeGPUFallback ||
		!settings.IsolatedRuntimeCache ||
		!settings.RepairTransientCacheOnLaunchFailure {
		t.Fatalf("cloakbrowser settings = %#v", settings)
	}
	if settings.FingerprintPlatform != "macos" {
		t.Fatalf("fingerprint platform = %q, want macos", settings.FingerprintPlatform)
	}
	if settings.FontsDir != "/opt/browseforge/fonts" {
		t.Fatalf("fonts dir = %q, want /opt/browseforge/fonts", settings.FontsDir)
	}
	if settings.StorageQuotaMB != 2048 {
		t.Fatalf("storage quota = %d, want 2048", settings.StorageQuotaMB)
	}
	if settings.TargetPlatformPolicy != "warn" {
		t.Fatalf("target platform policy = %q, want warn", settings.TargetPlatformPolicy)
	}
	if settings.PluginsPDF != "enabled" {
		t.Fatalf("plugins_pdf = %q, want enabled", settings.PluginsPDF)
	}
	if len(settings.ExtraArgs) != 1 || settings.ExtraArgs[0] != "--disable-features=Translate" {
		t.Fatalf("extra args = %#v", settings.ExtraArgs)
	}
}

func TestLoadChromiumRuntimeSettingsByRuntimeID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{
  "port": "19280",
  "profiles_dir": "profiles",
  "data_dir": "data",
  "log_file": "logs/server.log",
  "fingerprint_dir": "data",
  "runtimes": {
    "browseforge-chromium": {
      "binary_path": "/opt/browseforge/chromium",
      "settings": {
        "safe_gpu": true,
        "fingerprint_platform": "linux",
        "storage_quota_mb": 1024
      }
    }
  }
}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	settings := cfg.ChromiumRuntimeSettings("browseforge-chromium")
	if settings == nil {
		t.Fatal("browseforge-chromium settings missing")
	}
	if !settings.SafeGPU || settings.FingerprintPlatform != "linux" || settings.StorageQuotaMB != 1024 {
		t.Fatalf("browseforge-chromium settings = %#v", settings)
	}
}

func TestLoadMigratesLegacyRootRuntimeFieldsToV2Runtimes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{
  "camoufox_path": "/legacy/camoufox",
  "cloakbrowser_path": "/legacy/cloakbrowser",
  "cloakbrowser": {
    "safe_gpu": true,
    "auto_safe_gpu_fallback": true,
    "isolated_runtime_cache": true,
    "repair_transient_cache_on_launch_failure": true,
    "fingerprint_platform": "windows",
    "fonts_dir": "/legacy/fonts",
    "storage_quota_mb": 4096,
    "target_platform_policy": "strict",
    "extra_args": ["--legacy-flag"]
  }
}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := cfg.Runtime("camoufox").BinaryPath; got != "/legacy/camoufox" {
		t.Fatalf("camoufox binary_path = %q, want /legacy/camoufox", got)
	}
	if got := cfg.Runtime("cloakbrowser").BinaryPath; got != "/legacy/cloakbrowser" {
		t.Fatalf("cloakbrowser binary_path = %q, want /legacy/cloakbrowser", got)
	}
	settings := cfg.CloakBrowserSettings()
	if settings == nil {
		t.Fatal("cloakbrowser settings missing")
	}
	if !settings.SafeGPU ||
		!settings.AutoSafeGPUFallback ||
		!settings.IsolatedRuntimeCache ||
		!settings.RepairTransientCacheOnLaunchFailure {
		t.Fatalf("cloakbrowser settings = %#v", settings)
	}
	if settings.FingerprintPlatform != "windows" {
		t.Fatalf("fingerprint platform = %q, want windows", settings.FingerprintPlatform)
	}
	if settings.FontsDir != "/legacy/fonts" {
		t.Fatalf("fonts dir = %q, want /legacy/fonts", settings.FontsDir)
	}
	if settings.StorageQuotaMB != 4096 {
		t.Fatalf("storage quota = %d, want 4096", settings.StorageQuotaMB)
	}
	if settings.TargetPlatformPolicy != "strict" {
		t.Fatalf("target platform policy = %q, want strict", settings.TargetPlatformPolicy)
	}
	if len(settings.ExtraArgs) != 1 || settings.ExtraArgs[0] != "--legacy-flag" {
		t.Fatalf("extra args = %#v", settings.ExtraArgs)
	}

	encoded, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal migrated config: %v", err)
	}
	var emitted map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &emitted); err != nil {
		t.Fatalf("decode migrated config json: %v", err)
	}
	for _, legacyKey := range []string{"camoufox_path", "cloakbrowser_path", "cloakbrowser"} {
		if _, ok := emitted[legacyKey]; ok {
			t.Fatalf("migrated config re-emits legacy root key %q: %s", legacyKey, encoded)
		}
	}
}

func TestLoadKeepsExplicitRuntimesWhenLegacyRootFieldsAlsoExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{
  "camoufox_path": "/legacy/camoufox",
  "cloakbrowser_path": "/legacy/cloakbrowser",
  "cloakbrowser": {
    "fingerprint_platform": "windows",
    "fonts_dir": "/legacy/fonts",
    "storage_quota_mb": 4096,
    "target_platform_policy": "strict",
    "extra_args": ["--legacy-flag"]
  },
  "runtimes": {
    "camoufox": {
      "binary_path": "/v2/camoufox",
      "family": "firefox",
      "display_name": "V2 Camoufox"
    },
    "cloakbrowser": {
      "binary_path": "/v2/cloakbrowser",
      "family": "chromium",
      "display_name": "V2 CloakBrowser",
      "settings": {
        "fingerprint_platform": "linux",
        "fonts_dir": "/v2/fonts",
        "storage_quota_mb": 2048,
        "target_platform_policy": "warn",
        "extra_args": ["--v2-flag"]
      }
    }
  }
}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	camoufox := cfg.Runtime("camoufox")
	if camoufox.BinaryPath != "/v2/camoufox" {
		t.Fatalf("camoufox binary_path = %q, want /v2/camoufox", camoufox.BinaryPath)
	}
	if camoufox.Family != "firefox" || camoufox.DisplayName != "V2 Camoufox" {
		t.Fatalf("camoufox runtime metadata overwritten: %#v", camoufox)
	}
	cloakbrowser := cfg.Runtime("cloakbrowser")
	if cloakbrowser.BinaryPath != "/v2/cloakbrowser" {
		t.Fatalf("cloakbrowser binary_path = %q, want /v2/cloakbrowser", cloakbrowser.BinaryPath)
	}
	if cloakbrowser.Family != "chromium" || cloakbrowser.DisplayName != "V2 CloakBrowser" {
		t.Fatalf("cloakbrowser runtime metadata overwritten: %#v", cloakbrowser)
	}
	settings := cloakbrowser.Settings
	if settings == nil {
		t.Fatal("cloakbrowser settings missing")
	}
	if settings.FingerprintPlatform != "linux" {
		t.Fatalf("fingerprint platform = %q, want linux", settings.FingerprintPlatform)
	}
	if settings.FontsDir != "/v2/fonts" {
		t.Fatalf("fonts dir = %q, want /v2/fonts", settings.FontsDir)
	}
	if settings.StorageQuotaMB != 2048 {
		t.Fatalf("storage quota = %d, want 2048", settings.StorageQuotaMB)
	}
	if settings.TargetPlatformPolicy != "warn" {
		t.Fatalf("target platform policy = %q, want warn", settings.TargetPlatformPolicy)
	}
	if len(settings.ExtraArgs) != 1 || settings.ExtraArgs[0] != "--v2-flag" {
		t.Fatalf("extra args = %#v", settings.ExtraArgs)
	}
}
