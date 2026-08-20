// Style ForgeLocal — politique de navigation strictement locale (T15-W2).
// Fail-closed : toute URL externe est refusée ; seules file:// (exacte) et
// http(s)://127.0.0.1|localhost sont admises. Aucun préfixage implicite https://.

package api

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

// ErrInvalidURL is returned when a navigation target is not strictly local.
var ErrInvalidURL = errors.New("url not allowed: only local targets (file://, http(s)://127.0.0.1 or localhost) are accepted")

// ValidateLocalURL enforces the fail-closed local-only navigation policy.
//
// Accepted:
//   - file://<path>                 (exact scheme, no further normalization)
//   - http://127.0.0.1[:port]/...   loopback IP
//   - https://127.0.0.1[:port]/...  loopback IP
//   - http://localhost[:port]/...   loopback host
//   - https://localhost[:port]/...  loopback host
//
// Rejected: any other scheme, any external hostname, bare hosts ("example.com"),
// IPv4 private ranges (192.168.x, 10.x, ...), non-loopback IPs and all wildcard
// variants (0.0.0.0, [::], ::1 only loopback via IPv6 loopback).
func ValidateLocalURL(raw string) error {
	if raw == "" {
		return ErrInvalidURL
	}

	u, err := url.Parse(raw)
	if err != nil {
		return ErrInvalidURL
	}

	switch u.Scheme {
	case "file":
		// file:// accepted exactly (local fixture paths).
		return nil
	case "http", "https":
		// fall through to host check
	default:
		return ErrInvalidURL
	}

	host := u.Hostname()
	if host == "" {
		return ErrInvalidURL
	}

	if strings.EqualFold(host, "localhost") {
		return nil
	}

	// IP check: only loopback addresses.
	ip := net.ParseIP(host)
	if ip == nil {
		return ErrInvalidURL
	}
	if !ip.IsLoopback() {
		return ErrInvalidURL
	}
	return nil
}
