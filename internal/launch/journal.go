package launch

import (
	"context"
	"errors"
	"time"
)

// OperationJournal is the durable contract used by T18. Reserve and
// Transition are fail-closed: callers must not attach a resource when either
// durable step fails. Implementations contain only redacted metadata.
type OperationJournal interface {
	Lookup(ctx context.Context, idempotencyKey string) (op JournalOperation, found bool, err error)
	Reserve(ctx context.Context, in JournalOperation) (op JournalOperation, created bool, err error)
	Transition(ctx context.Context, session sessionID, from, to LaunchState, reason, correlationID string) error
	Reconcile(ctx context.Context) ([]RecoveredSession, error)
}

// JournalOperation is a durable, non-secret projection of one launch intent.
// IdempotencyKey is opaque input used only for deduplication; it is never
// copied to Manager audit events.
type JournalOperation struct {
	SessionID      sessionID
	ProfileID      profileID
	IdempotencyKey string
	State          LaunchState
	CorrelationID  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Reason         string
}

var (
	ErrJournalUnavailable = errors.New("durable operation journal unavailable")
	ErrInvalidTransition  = errors.New("invalid durable operation transition")
)

// WithOperationMetadata attaches a caller-supplied idempotency key and a
// correlation id to one Request. Both fields are redacted/validated before
// persistence. An absent key is replaced by the generated session ID, so it
// cannot accidentally deduplicate independent requests.
func WithOperationMetadata(ctx context.Context, idempotencyKey, correlationID string) context.Context {
	return context.WithValue(ctx, operationMetadataContextKey{}, operationMetadata{
		idempotencyKey: idempotencyKey,
		correlationID:  correlationID,
	})
}

type operationMetadataContextKey struct{}

type operationMetadata struct {
	idempotencyKey string
	correlationID  string
}

func metadataFromContext(ctx context.Context, fallback sessionID) operationMetadata {
	meta, _ := ctx.Value(operationMetadataContextKey{}).(operationMetadata)
	if meta.idempotencyKey == "" {
		meta.idempotencyKey = string(fallback)
	}
	meta.correlationID = redactCorrelationID(meta.correlationID)
	return meta
}

func redactCorrelationID(value string) string {
	if value == "" {
		return ""
	}
	if len(value) > 64 || isSecretish(value) {
		return "[redacted]"
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') &&
			!(r >= '0' && r <= '9') && r != '-' && r != '_' && r != '.' {
			return "[redacted]"
		}
	}
	return value
}

func validTransition(from, to LaunchState) bool {
	switch from {
	case StateQueued:
		return to == StateStarting || to == StateInterrupted
	case StateStarting:
		return to == StateRunning || to == StateError || to == StateInterrupted || to == StateStopped
	case StateRunning, StateStopping:
		return to == StateStopped
	default:
		return false
	}
}
