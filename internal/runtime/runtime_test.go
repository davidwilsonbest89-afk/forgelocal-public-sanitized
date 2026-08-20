package runtime

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"forgelocal/internal/config"
	"forgelocal/internal/profile"
)

func TestNewRegistryLoadsBrowseForgeChromiumFromDefaultConfig(t *testing.T) {
	cfg, err := config.Load(filepath.Join("..", "..", "config.default.json"))
	if err != nil {
		t.Fatalf("load default config: %v", err)
	}

	reg := NewRegistry(cfg)
	desc, ok := reg.Get(BrowseForgeChromium)
	if !ok {
		t.Fatal("BrowseForge Chromium runtime missing")
	}
	if desc.DisplayName != "BrowseForge Chromium" || desc.Family != FamilyChromium || desc.Capabilities.Family != FamilyChromium {
		t.Fatalf("BrowseForge Chromium metadata = %+v", desc)
	}
	if desc.BinaryPath != "browsers/browseforge-chromium/chrome" || !desc.Enabled {
		t.Fatalf("BrowseForge Chromium binary/enabled = %q/%v, want Docker/GHCR browser path and enabled", desc.BinaryPath, desc.Enabled)
	}
}

func TestRegistryResolvesExplicitRuntimeID(t *testing.T) {
	reg := NewRegistry(&config.Config{})

	tests := []struct {
		name      string
		runtimeID string
		wantID    ID
	}{
		{name: "camoufox runtime", runtimeID: "camoufox", wantID: Camoufox},
		{name: "cloakbrowser runtime", runtimeID: "cloakbrowser", wantID: CloakBrowser},
		{name: "browseforge chromium runtime", runtimeID: "browseforge-chromium", wantID: BrowseForgeChromium},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &profile.Profile{RuntimeID: tc.runtimeID}

			desc, err := reg.ApplyProfileDefaults(p)
			if err != nil {
				t.Fatalf("ApplyProfileDefaults: %v", err)
			}
			if desc.ID != tc.wantID {
				t.Fatalf("runtime ID = %q, want %q", desc.ID, tc.wantID)
			}
			if p.RuntimeID != string(tc.wantID) {
				t.Fatalf("profile runtime_id = %q, want %q", p.RuntimeID, tc.wantID)
			}
		})
	}
}

func TestRegistryRequiresExplicitRuntimeID(t *testing.T) {
	reg := NewRegistry(&config.Config{})
	p := &profile.Profile{}

	if _, err := reg.ApplyProfileDefaults(p); err == nil || !strings.Contains(err.Error(), "runtime_id is required") {
		t.Fatalf("missing runtime_id err = %v", err)
	}
}

func TestRegistryResolvesRuntimeID(t *testing.T) {
	reg := NewRegistry(&config.Config{})

	got, err := reg.ResolveID("browseforge-chromium")
	if err != nil {
		t.Fatalf("ResolveID: %v", err)
	}
	if got != BrowseForgeChromium {
		t.Fatalf("runtime ID = %q, want %q", got, BrowseForgeChromium)
	}
}

func TestRegistryRejectsUnsupportedRuntimeSelectors(t *testing.T) {
	reg := NewRegistry(&config.Config{})

	if _, err := reg.ResolveID("opera"); err == nil || !strings.Contains(err.Error(), `unsupported runtime_id "opera"`) {
		t.Fatalf("unsupported runtime_id err = %v", err)
	}
	if _, err := reg.ResolveID(""); err == nil || !strings.Contains(err.Error(), "runtime_id is required") {
		t.Fatalf("empty runtime_id err = %v", err)
	}
}

func TestRegistryListReturnsStableRuntimeMetadata(t *testing.T) {
	reg := NewRegistry(&config.Config{Runtimes: map[string]config.RuntimeConfig{
		"camoufox":             {BinaryPath: "/opt/camoufox"},
		"browseforge-chromium": {BinaryPath: "/opt/browseforge-chromium"},
		"cloakbrowser":         {BinaryPath: "/opt/cloakbrowser"},
	}})

	got := reg.List()
	if len(got) != 3 {
		t.Fatalf("runtime count = %d, want 3: %#v", len(got), got)
	}
	if got[0].ID != BrowseForgeChromium || got[1].ID != Camoufox || got[2].ID != CloakBrowser {
		t.Fatalf("runtime order = [%q %q %q], want [%q %q %q]", got[0].ID, got[1].ID, got[2].ID, BrowseForgeChromium, Camoufox, CloakBrowser)
	}

	browseforge := got[0]
	if browseforge.DisplayName != "BrowseForge Chromium" || browseforge.Family != FamilyChromium {
		t.Fatalf("BrowseForge Chromium metadata = %+v", browseforge)
	}
	if browseforge.BinaryPath != "/opt/browseforge-chromium" || !browseforge.Enabled {
		t.Fatalf("BrowseForge Chromium binary/enabled = %q/%v", browseforge.BinaryPath, browseforge.Enabled)
	}
	if !browseforge.Capabilities.SupportsAgentWebSessions || !browseforge.Capabilities.SupportsSeedFingerprint || !browseforge.Capabilities.SupportsPlaywrightBind {
		t.Fatalf("BrowseForge Chromium capabilities = %+v", browseforge.Capabilities)
	}

	camoufox := got[1]
	if camoufox.DisplayName != "Camoufox" || camoufox.Family != FamilyFirefox {
		t.Fatalf("Camoufox metadata = %+v", camoufox)
	}
	if camoufox.BinaryPath != "" || camoufox.Enabled || camoufox.ExecutionAuthorized {
		t.Fatalf("Camoufox policy projection = %+v, want disabled and unauthorized", camoufox)
	}
	if camoufox.FingerprintPoolKey != "firefox" {
		t.Fatalf("Camoufox fingerprint pool = %q, want firefox", camoufox.FingerprintPoolKey)
	}
	if !camoufox.Capabilities.SupportsPersistentContext || !camoufox.Capabilities.SupportsPlaywrightBind || camoufox.Capabilities.SupportsAgentWebSessions {
		t.Fatalf("Camoufox capabilities = %+v", camoufox.Capabilities)
	}

	cloak := got[2]
	if cloak.DisplayName != "CloakBrowser" || cloak.Family != FamilyChromium {
		t.Fatalf("CloakBrowser metadata = %+v", cloak)
	}
	if cloak.BinaryPath != "/opt/cloakbrowser" || !cloak.Enabled {
		t.Fatalf("CloakBrowser binary/enabled = %q/%v", cloak.BinaryPath, cloak.Enabled)
	}
	if !cloak.Capabilities.SupportsAgentWebSessions || !cloak.Capabilities.SupportsSeedFingerprint || !cloak.Capabilities.SupportsPlaywrightBind {
		t.Fatalf("CloakBrowser capabilities = %+v", cloak.Capabilities)
	}
}

func TestPlatformSupportDisablesUnsupportedRuntime(t *testing.T) {
	ok, reason := platformSupported(Camoufox, "windows", "arm64")
	if ok {
		t.Fatal("Camoufox windows/arm64 should be unsupported for the configured release")
	}
	if !strings.Contains(reason, "camoufox") || !strings.Contains(reason, "windows/arm64") {
		t.Fatalf("unsupported reason = %q", reason)
	}

	ok, reason = platformSupported(BrowseForgeChromium, "windows", "amd64")
	if !ok || reason != "" {
		t.Fatalf("BrowseForge Chromium windows/amd64 support = %v %q, want supported", ok, reason)
	}
}

func TestRegistryAppliesRuntimeConfigOverrides(t *testing.T) {
	disabled := false
	enabled := true
	reg := NewRegistry(&config.Config{
		Runtimes: map[string]config.RuntimeConfig{
			"camoufox": {
				DisplayName: "Camoufox ESR",
				BinaryPath:  "/custom/camoufox",
				Enabled:     &disabled,
			},
			"cloakbrowser": {
				DisplayName: "CloakBrowser QA",
				BinaryPath:  "/custom/cloakbrowser",
				Family:      "chromium",
				Enabled:     &enabled,
			},
		},
	})

	camoufox, ok := reg.Get(Camoufox)
	if !ok {
		t.Fatal("Camoufox runtime missing")
	}
	if camoufox.DisplayName != "Camoufox ESR" || camoufox.BinaryPath != "" || camoufox.Enabled || camoufox.ExecutionAuthorized {
		t.Fatalf("Camoufox override = %+v", camoufox)
	}

	cloak, ok := reg.Get(CloakBrowser)
	if !ok {
		t.Fatal("CloakBrowser runtime missing")
	}
	if cloak.DisplayName != "CloakBrowser QA" || cloak.BinaryPath != "/custom/cloakbrowser" || !cloak.Enabled {
		t.Fatalf("CloakBrowser override = %+v", cloak)
	}
	if cloak.Family != FamilyChromium || cloak.Capabilities.Family != FamilyChromium {
		t.Fatalf("CloakBrowser family override = descriptor %q capabilities %q", cloak.Family, cloak.Capabilities.Family)
	}
}

func TestCamoufoxExecutionPolicyCannotBeEnabledByConfig(t *testing.T) {
	enabled := true
	reg := NewRegistry(&config.Config{Runtimes: map[string]config.RuntimeConfig{
		"camoufox": {BinaryPath: "/pretend/camoufox", Enabled: &enabled},
	}})
	desc, ok := reg.Get(Camoufox)
	if !ok {
		t.Fatal("Camoufox runtime missing")
	}
	if desc.Enabled || desc.ExecutionAuthorized || desc.BinaryPath != "" {
		t.Fatalf("configured Camoufox bypassed policy: %+v", desc)
	}
	if !errors.Is(RequireExecution(Camoufox), ErrCamoufoxExecutionNotAuthorized) {
		t.Fatalf("RequireExecution(Camoufox) must fail closed")
	}
}
