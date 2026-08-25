package browser

import (
	"testing"

	"forgelocal/internal/groups"
	"forgelocal/internal/profile"
)

func TestManagerEffectiveProxyUsesGroupStore(t *testing.T) {
	store, err := groups.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "group.proxy", Port: 1080}, groups.ProxyModeEnforced); err != nil {
		t.Fatal(err)
	}
	m := &Manager{groupStore: store}

	effective := m.effectiveProxy(&profile.Profile{
		Group: "Client A",
		Proxy: &profile.ProxyConfig{Type: "http", Host: "profile.proxy", Port: 8080},
	})
	if effective.Source != "group" || effective.Proxy.Host != "group.proxy" {
		t.Fatalf("effective proxy = %+v", effective)
	}
}

func TestIsLoopbackProxyHost(t *testing.T) {
	for _, host := range []string{"localhost", "127.0.0.1", "127.0.0.9", "[::1]"} {
		if !isLoopbackProxyHost(host) {
			t.Errorf("isLoopbackProxyHost(%q) = false", host)
		}
	}
	for _, host := range []string{"proxy.example", "192.0.2.10"} {
		if isLoopbackProxyHost(host) {
			t.Errorf("isLoopbackProxyHost(%q) = true", host)
		}
	}
}
