package runtime

import (
	"strings"
	"testing"

	"browseforge/internal/config"
	"browseforge/internal/profile"
)

func TestRegistryResolvesExplicitRuntimeID(t *testing.T) {
	reg := NewRegistry(&config.Config{})

	tests := []struct {
		name      string
		runtimeID string
		wantID    ID
	}{
		{name: "camoufox runtime", runtimeID: "camoufox", wantID: Camoufox},
		{name: "cloakbrowser runtime", runtimeID: "cloakbrowser", wantID: CloakBrowser},
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

	got, err := reg.ResolveID("cloakbrowser")
	if err != nil {
		t.Fatalf("ResolveID: %v", err)
	}
	if got != CloakBrowser {
		t.Fatalf("runtime ID = %q, want %q", got, CloakBrowser)
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
		"camoufox":     {BinaryPath: "/opt/camoufox"},
		"cloakbrowser": {BinaryPath: "/opt/cloakbrowser"},
	}})

	got := reg.List()
	if len(got) != 2 {
		t.Fatalf("runtime count = %d, want 2: %#v", len(got), got)
	}
	if got[0].ID != Camoufox || got[1].ID != CloakBrowser {
		t.Fatalf("runtime order = [%q %q], want [%q %q]", got[0].ID, got[1].ID, Camoufox, CloakBrowser)
	}

	camoufox := got[0]
	if camoufox.DisplayName != "Camoufox" || camoufox.Family != FamilyFirefox {
		t.Fatalf("Camoufox metadata = %+v", camoufox)
	}
	if camoufox.BinaryPath != "/opt/camoufox" || !camoufox.Enabled {
		t.Fatalf("Camoufox binary/enabled = %q/%v", camoufox.BinaryPath, camoufox.Enabled)
	}
	if camoufox.FingerprintPoolKey != "firefox" {
		t.Fatalf("Camoufox fingerprint pool = %q, want firefox", camoufox.FingerprintPoolKey)
	}
	if !camoufox.Capabilities.SupportsPersistentContext || !camoufox.Capabilities.SupportsPlaywrightBind || camoufox.Capabilities.SupportsAgentWebSessions {
		t.Fatalf("Camoufox capabilities = %+v", camoufox.Capabilities)
	}

	cloak := got[1]
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
	if camoufox.DisplayName != "Camoufox ESR" || camoufox.BinaryPath != "/custom/camoufox" || camoufox.Enabled {
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
