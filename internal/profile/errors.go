// T09 — Profile Writes. Explicit sentinel errors for the profile write
// contract. Handlers map these to machine-readable API error codes; internal
// messages never leak to API responses or audit payloads.
//
// Style contract (project): redaction at the boundary; sentinel errors carry
// short, non-sensitive messages only.

package profile

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a profile id does not identify a stored profile.
var ErrNotFound = errors.New("profile not found")

// ErrLocked is returned when a profile operation could not acquire the
// per-profile lock within the isolation budget (single-mutation contract).
var ErrLocked = errors.New("profile locked by another operation")

// Lifecycle errors are returned when a transition is not authorized by the
// profile state machine (active -> archived, archived -> active only).
var (
	ErrAlreadyArchived = errors.New("profile is already archived")
	ErrNotArchived     = errors.New("profile is not archived")
	ErrQuarantined     = errors.New("quarantined profiles cannot be reopened")
)

// Validation errors are returned for inputs that violate the write contract.
var (
	ErrInvalidName         = errors.New("name is required and must be a short printable string")
	ErrDuplicateName       = errors.New("a profile with this name already exists")
	ErrInvalidGroup        = errors.New("group does not exist")
	ErrInvalidRuntime      = errors.New("runtime is not enabled or not registered")
	ErrInvalidProxy        = errors.New("proxy configuration is inconsistent: type, host and port must agree")
	ErrDuplicateID         = errors.New("profile id already exists")
	ErrTooManyTags         = errors.New("profile exceeds the maximum number of tags")
	ErrInvalidTag          = errors.New("tag is required and must be a short printable string")
	ErrAlreadyTagged       = errors.New("tag is already assigned")
	ErrTagNotAssigned      = errors.New("tag is not assigned to this profile")
	ErrInvalidNote         = errors.New("profile note is invalid")
	ErrInvalidCustomField  = errors.New("profile custom field is invalid")
	ErrTooManyCustomFields = errors.New("profile exceeds the maximum number of custom fields")
)

// IsNotFoundError reports whether the error identifies a missing profile.
func IsNotFoundError(err error) bool {
	return err != nil && errors.Is(err, ErrNotFound)
}

// IsLockedError reports whether the error is a concurrency isolation refusal.
func IsLockedError(err error) bool {
	return errors.Is(err, ErrLocked)
}

// IsLifecycleError reports whether the error is a state machine refusal.
func IsLifecycleError(err error) bool {
	return err != nil && (errors.Is(err, ErrAlreadyArchived) || errors.Is(err, ErrNotArchived) || errors.Is(err, ErrQuarantined))
}

// IsDuplicateError reports whether the error is a name/id uniqueness refusal.
func IsDuplicateError(err error) bool {
	return err != nil && (errors.Is(err, ErrDuplicateName) || errors.Is(err, ErrDuplicateID))
}

// IsValidationError reports whether the error is an input validation refusal.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	for _, target := range []error{ErrInvalidName, ErrInvalidGroup, ErrInvalidRuntime,
		ErrInvalidProxy, ErrTooManyTags, ErrInvalidTag, ErrAlreadyTagged, ErrTagNotAssigned,
		ErrInvalidNote, ErrInvalidCustomField, ErrTooManyCustomFields} {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// LifecycleState is the persisted lifecycle state of a profile.
type LifecycleState string

const (
	// LifecycleActive is the default working state of a profile.
	LifecycleActive LifecycleState = "active"
	// LifecycleArchived marks a profile as dormant; its directory stays on disk
	// and it may be reopened when no session holds it.
	LifecycleArchived LifecycleState = "archived"
	// LifecycleQuarantined marks a profile as blocked by an external authority;
	// it cannot be reopened by clients.
	LifecycleQuarantined LifecycleState = "quarantined"
)

// ValidLifecycleState reports whether the value belongs to the state enum.
func ValidLifecycleState(state LifecycleState) bool {
	switch state {
	case LifecycleActive, LifecycleArchived, LifecycleQuarantined:
		return true
	default:
		return false
	}
}

func wrap(source, detail error) error {
	if detail == nil {
		return source
	}
	return fmt.Errorf("%w: %v", source, detail)
}
