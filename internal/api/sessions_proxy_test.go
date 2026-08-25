package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgelocal/internal/profile"
	"forgelocal/internal/proxies"
	"forgelocal/internal/secrets"
)

func newSessionProxyFixture(t *testing.T) (*handler, *profile.Store, *proxies.Store, *secrets.MemoryVault) {
	t.Helper()
	vault := secrets.NewMemoryVault()
	profiles, err := profile.NewStore(t.TempDir(), vault)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := proxies.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return &handler{store: profiles, proxyStore: registry}, profiles, registry, vault
}

func createSessionProxyProfile(t *testing.T, store *profile.Store, id string) *profile.Profile {
	t.Helper()
	p := &profile.Profile{ID: id, Name: id, RuntimeID: "chromium", LifecycleState: profile.LifecycleActive}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	return p
}

func createRegistryProxy(t *testing.T, store *proxies.Store, id, name, host string, port int, secretRef string) *proxies.Proxy {
	t.Helper()
	p := &proxies.Proxy{ID: id, Name: name, Type: "http", Host: host, Port: port, SecretRef: secretRef, HasSecret: secretRef != ""}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSessionProxyAssignmentResolvesDistinctAlphaAndBetaLaunchConfigs(t *testing.T) {
	h, profiles, registry, vault := newSessionProxyFixture(t)
	alphaProfile := createSessionProxyProfile(t, profiles, "profile-alpha")
	betaProfile := createSessionProxyProfile(t, profiles, "profile-beta")
	alpha := createRegistryProxy(t, registry, "proxy-alpha", "alpha", "127.0.0.1", 19282, "proxy.ref.alpha")
	beta := createRegistryProxy(t, registry, "proxy-beta", "beta", "127.0.0.1", 19284, "proxy.ref.beta")
	if err := vault.PutSecret(alpha.SecretRef, []byte(`{"username":"fixture-alpha","password":"fixture-alpha-secret"}`)); err != nil {
		t.Fatal(err)
	}
	if err := vault.PutSecret(beta.SecretRef, []byte(`{"username":"fixture-beta","password":"fixture-beta-secret"}`)); err != nil {
		t.Fatal(err)
	}
	if err := registry.Assign(alphaProfile.ID, alpha.ID); err != nil {
		t.Fatal(err)
	}
	if err := registry.Assign(betaProfile.ID, beta.ID); err != nil {
		t.Fatal(err)
	}

	alphaLaunch, err := h.profileForSessionLaunch(alphaProfile)
	if err != nil {
		t.Fatal(err)
	}
	betaLaunch, err := h.profileForSessionLaunch(betaProfile)
	if err != nil {
		t.Fatal(err)
	}
	if alphaLaunch.LaunchProxy == nil || betaLaunch.LaunchProxy == nil {
		t.Fatal("assigned profiles must carry an ephemeral launch proxy")
	}
	if alphaLaunch.LaunchProxy.Host != "127.0.0.1" || alphaLaunch.LaunchProxy.Port != 19282 || alphaLaunch.LaunchProxy.Username != "fixture-alpha" {
		t.Fatalf("alpha launch proxy = %#v", alphaLaunch.LaunchProxy)
	}
	if betaLaunch.LaunchProxy.Host != "127.0.0.1" || betaLaunch.LaunchProxy.Port != 19284 || betaLaunch.LaunchProxy.Username != "fixture-beta" {
		t.Fatalf("beta launch proxy = %#v", betaLaunch.LaunchProxy)
	}
	if alphaLaunch.LaunchProxy.Port == betaLaunch.LaunchProxy.Port || alphaLaunch.LaunchProxy.Username == betaLaunch.LaunchProxy.Username {
		t.Fatal("alpha and beta launch configurations must remain isolated")
	}

	encoded, err := json.Marshal(alphaLaunch)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "fixture-alpha") || strings.Contains(string(encoded), "fixture-alpha-secret") {
		t.Fatalf("launch credentials leaked into JSON: %s", encoded)
	}
}

func TestSessionProxyAssignmentFailsClosedForUnknownRegistryTarget(t *testing.T) {
	h, profiles, _, _ := newSessionProxyFixture(t)
	p := createSessionProxyProfile(t, profiles, "profile-unknown")
	registryDir := t.TempDir()
	registry, err := proxies.NewStore(registryDir)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"proxies":[],"assigned":{"profile-unknown":"proxy-missing"}}`)
	if err := os.WriteFile(filepath.Join(registryDir, "proxies.json"), payload, 0600); err != nil {
		t.Fatal(err)
	}
	registry, err = proxies.NewStore(registryDir)
	if err != nil {
		t.Fatal(err)
	}
	h.proxyStore = registry
	if _, err := h.profileForSessionLaunch(p); err == nil || !strings.Contains(err.Error(), "PROXY_ASSIGNMENT_UNKNOWN") {
		t.Fatalf("unknown assignment error = %v", err)
	}
}

func TestSessionProxyAssignmentFailsClosedForInvalidRecordAndUnavailableSecret(t *testing.T) {
	h, profiles, registry, _ := newSessionProxyFixture(t)
	invalidProfile := createSessionProxyProfile(t, profiles, "profile-invalid")
	invalid := createRegistryProxy(t, registry, "proxy-invalid", "invalid", "127.0.0.1", 19282, "proxy.ref.invalid")
	invalid.HasSecret = false
	if err := registry.Assign(invalidProfile.ID, invalid.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := h.profileForSessionLaunch(invalidProfile); err == nil || !strings.Contains(err.Error(), "PROXY_ASSIGNMENT_INVALID") {
		t.Fatalf("invalid assignment error = %v", err)
	}

	missingProfile := createSessionProxyProfile(t, profiles, "profile-missing-secret")
	missing := createRegistryProxy(t, registry, "proxy-missing-secret", "missing-secret", "127.0.0.1", 19284, "proxy.ref.missing")
	if err := registry.Assign(missingProfile.ID, missing.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := h.profileForSessionLaunch(missingProfile); err == nil || !strings.Contains(err.Error(), "PROXY_CREDENTIALS_UNAVAILABLE") {
		t.Fatalf("missing secret error = %v", err)
	}
}

func TestSessionWithoutRegistryAssignmentPreservesExplicitNoProxyBehavior(t *testing.T) {
	h, profiles, _, _ := newSessionProxyFixture(t)
	p := createSessionProxyProfile(t, profiles, "profile-direct")
	launch, err := h.profileForSessionLaunch(p)
	if err != nil {
		t.Fatal(err)
	}
	if launch != p {
		t.Fatal("unassigned profile should preserve the existing launch object")
	}
	if launch.LaunchProxy != nil {
		t.Fatal("unassigned profile must not receive a registry proxy override")
	}
}
