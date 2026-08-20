package proxyprovider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testProvider() Provider {
	return Provider{ID: "acme-test", Name: "Acme QA", SecretRef: "provider.ref.acme-test"}
}

func TestRegisterPersistsReferenceOnlyAndSimulatesTestHost(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := store.Register(testProvider())
	if err != nil {
		t.Fatal(err)
	}
	if provider.Mode != "simulated" {
		t.Fatalf("mode=%q", provider.Mode)
	}
	data, err := os.ReadFile(filepath.Join(dir, "proxy-providers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "api-key") || strings.Contains(string(data), "password") {
		t.Fatalf("secret persisted: %s", data)
	}
	lease, err := store.SimulateResolve(provider.ID, "profile-test", "eu-test")
	if err != nil {
		t.Fatal(err)
	}
	if lease.Host != "eu-test.acme-test.provider.test" || lease.Port != 18080 || lease.Mode != "simulated" {
		t.Fatalf("lease=%#v", lease)
	}
}

func TestRegisterRejectsWrongReferenceOrRegion(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	bad := testProvider()
	bad.SecretRef = "provider.ref.other"
	if _, err := store.Register(bad); err == nil {
		t.Fatal("wrong reference accepted")
	}
	if _, err := store.Register(testProvider()); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SimulateResolve("acme-test", "profile-test", "eu-west"); err == nil {
		t.Fatal("non-test region accepted")
	}
}
