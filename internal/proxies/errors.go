// T10 — Proxies. Explicit sentinel errors for the proxy registry contract.
//
// Handlers map these to machine-readable API error codes; internal messages
// never leak to API responses or audit payloads.
//
// Style contract (project): redaction at the boundary; sentinel errors carry
// short, non-sensitive messages only. No proxy network, browser, runtime,
// Camoufox, provider integration or release activity lives in this package.
package proxies

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a proxy id does not identify a stored proxy.
var ErrNotFound = errors.New("proxy not found")

// ErrDuplicateName is returned when a proxy with the same name already exists
// (case-insensitive comparison).
var ErrDuplicateName = errors.New("a proxy with this name already exists")

// ErrProxyInUse is returned when a mutation targets a proxy currently
// assigned to one or more profiles.
var ErrProxyInUse = errors.New("the proxy is assigned to profiles")

// ErrInvalidProxy is returned when a proxy payload violates the contract
// (type, host, port or region rules).
var ErrInvalidProxy = errors.New("proxy data is invalid")

// ErrInvalidName is returned when the proxy name is missing or too long.
var ErrInvalidName = errors.New("name is required and must be a short printable string")

// ErrInvalidHost is returned when the proxy host is missing or malformed.
var ErrInvalidHost = errors.New("host is required and must be a short printable string")

// ErrInvalidPort is returned when the proxy port is outside the 1..65535 range.
var ErrInvalidPort = errors.New("port must be between 1 and 65535")

// ErrInvalidType is returned when the proxy type is neither http nor socks5.
var ErrInvalidType = errors.New("type must be http or socks5")

// ErrInvalidRegion is returned when the region value is not a short printable string.
var ErrInvalidRegion = errors.New("region must be a short printable string")

// ErrDuplicateID is returned when an explicit proxy id is already registered.
var ErrDuplicateID = errors.New("proxy id already exists")

// ErrProxyLocked is returned when a proxy operation could not acquire the
// per-proxy lock within the isolation budget (single-mutation contract).
var ErrProxyLocked = errors.New("proxy locked by another operation")

// IsNotFoundError reports whether the error identifies a missing proxy.
func IsNotFoundError(err error) bool {
	return err != nil && errors.Is(err, ErrNotFound)
}

// IsDuplicateError reports whether the error identifies a name conflict.
func IsDuplicateError(err error) bool {
	return err != nil && (errors.Is(err, ErrDuplicateName) || errors.Is(err, ErrProxyInUse) || errors.Is(err, ErrDuplicateID))
}

// IsLockedError reports whether the error is a concurrency isolation refusal.
func IsLockedError(err error) bool {
	return err != nil && errors.Is(err, ErrProxyLocked)
}

// IsValidationError reports whether the error is an input validation refusal.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	for _, target := range []error{ErrInvalidProxy, ErrInvalidName, ErrInvalidHost,
		ErrInvalidPort, ErrInvalidType, ErrInvalidRegion} {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func wrap(source, detail error) error {
	if detail == nil {
		return source
	}
	return fmt.Errorf("%w: %v", source, detail)
}
