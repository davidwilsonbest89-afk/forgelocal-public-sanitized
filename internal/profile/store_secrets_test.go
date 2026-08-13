package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgelocal/internal/secrets"
)

func TestProxyCredentialsRemainInVaultAndOutOfProfileJSON(t *testing.T) {
	vault := secrets.NewMemoryVault()
	dir := t.TempDir()
	store, err := NewStore(dir, vault)
	if err != nil {
		t.Fatal(err)
	}
	p := &Profile{
		ID:        "proxy-secret-profile",
		Name:      "Proxy secret",
		RuntimeID: "chromium",
		Proxy: &ProxyConfig{
			Type:     "http",
			Host:     "proxy.local",
			Port:     8080,
			Username: "alice",
			Password: "never-on-disk",
		},
	}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	if p.Proxy.SecretRef == "" {
		t.Fatal("missing proxy secret reference")
	}
	secret, err := vault.GetSecret(p.Proxy.SecretRef)
	if err != nil || !strings.Contains(string(secret), "never-on-disk") {
		t.Fatalf("proxy credential was not stored in vault: %q, %v", secret, err)
	}
	profileJSON, err := os.ReadFile(filepath.Join(p.ProfileDir, "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(profileJSON), "alice") || strings.Contains(string(profileJSON), "never-on-disk") {
		t.Fatalf("proxy credential leaked into profile.json: %s", profileJSON)
	}
	responseJSON, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(responseJSON), "alice") || strings.Contains(string(responseJSON), "never-on-disk") {
		t.Fatalf("proxy credential leaked into serialized profile: %s", responseJSON)
	}

	// A fresh Core-owned store process can resolve the credentials only via the
	// same vault; no credential recovery is possible from profile.json alone.
	reopened, err := NewStore(dir, vault)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := reopened.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Proxy == nil || loaded.Proxy.Username != "alice" || loaded.Proxy.Password != "never-on-disk" {
		t.Fatalf("vault-backed credentials not restored in Core memory: %#v", loaded.Proxy)
	}

	// An imported profile cannot nominate the first profile's secret reference.
	attacker := &Profile{ID: "proxy-secret-attacker", Name: "No secret", RuntimeID: "chromium", Proxy: &ProxyConfig{Host: "proxy.local", Port: 8080, SecretRef: p.Proxy.SecretRef}}
	if err := reopened.Create(attacker); err != nil {
		t.Fatal(err)
	}
	if attacker.Proxy.SecretRef != "" || attacker.Proxy.Username != "" || attacker.Proxy.Password != "" {
		t.Fatalf("foreign proxy secret reference was retained: %#v", attacker.Proxy)
	}
}
