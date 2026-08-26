package api

import "testing"

func TestValidateInternalPlaywrightEndpoint(t *testing.T) {
	valid := []struct {
		name string
		raw  string
		addr string
		path string
	}{
		{"ipv4", "ws://127.0.0.1:19280/api/playwright/ws/sess-1", "127.0.0.1:19280", "/api/playwright/ws/sess-1"},
		{"ipv6-loopback", "ws://[::1]:19280/api/playwright/ws/sess-1", "127.0.0.1:19280", "/api/playwright/ws/sess-1"},
		{"localhost", "ws://localhost:19280/api/playwright/ws/sess-1", "127.0.0.1:19280", "/api/playwright/ws/sess-1"},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			addr, path, err := validateInternalPlaywrightEndpoint(tc.raw, "sess-1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if addr != tc.addr || path != tc.path {
				t.Fatalf("normalized endpoint=(%q,%q), want (%q,%q)", addr, path, tc.addr, tc.path)
			}
		})
	}

	invalid := []struct {
		name string
		raw  string
	}{
		{"external-host", "ws://evil.invalid:19280/api/playwright/ws/sess-1"},
		{"external-ip", "ws://192.0.2.10:19280/api/playwright/ws/sess-1"},
		{"wrong-scheme", "http://127.0.0.1:19280/api/playwright/ws/sess-1"},
		{"missing-port", "ws://127.0.0.1/api/playwright/ws/sess-1"},
		{"invalid-port", "ws://127.0.0.1:70000/api/playwright/ws/sess-1"},
		{"query", "ws://127.0.0.1:19280/api/playwright/ws/sess-1?target=evil"},
		{"wrong-path", "ws://127.0.0.1:19280/api/playwright/ws/other"},
		{"userinfo", "ws://user:pass@127.0.0.1:19280/api/playwright/ws/sess-1"},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if addr, path, err := validateInternalPlaywrightEndpoint(tc.raw, "sess-1"); err == nil {
				t.Fatalf("accepted invalid endpoint: addr=%q path=%q", addr, path)
			}
		})
	}
}
