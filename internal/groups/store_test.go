package groups

import (
	"os"
	"path/filepath"
	"testing"

	"browseforge/internal/profile"
)

func TestEffectiveProxyDefaultModeUsesProfileOverride(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "group.proxy", Port: 1080}, ProxyModeDefault); err != nil {
		t.Fatal(err)
	}

	p := &profile.Profile{
		Group: "Client A",
		Proxy: &profile.ProxyConfig{Type: "http", Host: "profile.proxy", Port: 8080},
	}
	effective := store.EffectiveProxy(p)
	if effective.Source != "profile" || effective.Proxy.Host != "profile.proxy" || effective.Mode != ProxyModeDefault {
		t.Fatalf("effective proxy = %+v", effective)
	}
}

func TestEffectiveProxyEnforcedModeUsesGroupOverride(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "group.proxy", Port: 1080}, ProxyModeEnforced); err != nil {
		t.Fatal(err)
	}

	p := &profile.Profile{
		Group: "Client A",
		Proxy: &profile.ProxyConfig{Type: "http", Host: "profile.proxy", Port: 8080},
	}
	effective := store.EffectiveProxy(p)
	if effective.Source != "group" || effective.Proxy.Host != "group.proxy" || effective.Mode != ProxyModeEnforced {
		t.Fatalf("effective proxy = %+v", effective)
	}
}

func TestGroupStorePersistsAndImports(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "group.proxy", Port: 1080, Username: "u", Password: "p"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "groups.json")); err != nil {
		t.Fatalf("groups.json not written: %v", err)
	}

	reloaded, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	g, ok := reloaded.Get("Client A")
	if !ok || g.Proxy == nil || g.Proxy.Host != "group.proxy" || g.ProxyMode != ProxyModeDefault {
		t.Fatalf("reloaded group = %+v ok=%v", g, ok)
	}

	data, err := reloaded.Export()
	if err != nil {
		t.Fatal(err)
	}
	imported, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	n, err := imported.Import(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("imported = %d", n)
	}
	if g, ok := imported.Get("Client A"); !ok || g.Proxy.Host != "group.proxy" {
		t.Fatalf("imported group = %+v ok=%v", g, ok)
	}

	n, err = imported.Import([]byte(`{"groups":[{"name":"Empty","proxy_mode":"default"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("imported empty policies = %d", n)
	}
	if g, ok := imported.Get("Empty"); ok {
		t.Fatalf("empty policy should be skipped: %+v", g)
	}
}

func TestGroupStoreRejectsInvalidProxy(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert("Client A", nil, ProxyModeDefault); err == nil {
		t.Fatal("expected missing proxy error")
	}
	if _, err := store.Upsert("Client A", &profile.ProxyConfig{Type: "ftp", Host: "proxy", Port: 1080}, ProxyModeDefault); err == nil {
		t.Fatal("expected invalid proxy type error")
	}
	if _, err := store.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "proxy", Port: 70000}, ProxyModeDefault); err == nil {
		t.Fatal("expected invalid proxy port error")
	}
	if _, err := store.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "proxy", Port: 1080}, "strict"); err == nil {
		t.Fatal("expected invalid proxy mode error")
	}
}

func TestClearProxyRemovesStoredPolicy(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "proxy", Port: 1080}, ProxyModeDefault); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ClearProxy("Client A"); err != nil {
		t.Fatal(err)
	}
	if g, ok := store.Get("Client A"); ok {
		t.Fatalf("cleared group still exists = %+v", g)
	}
}
