// T09 — Profile Writes. Correlation id propagation.
//
// Every request receives a deterministic correlation id that is attached to
// its responses and to every audit event the request produces. Callers may
// supply one through the X-Correlation-ID header; otherwise the Core generates
// a `cor_<random>` id and echoes it back in the same header of the response.
//
// The correlation id never leaves the boundary as a secret: it is a plain
// request-tracking token, not a credential.

package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const correlationHeader = "X-Correlation-ID"

// generateCorrelationID returns a fresh 48-bit random id with the cor_ prefix.
func generateCorrelationID() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "cor_unreachable"
	}
	return "cor_" + hex.EncodeToString(b)
}

// validCorrelationID rejects header values that are not the expected prefix
// plus hex characters, so a caller cannot smuggle control characters into
// the audit ledger.
func validCorrelationID(value string) bool {
	if len(value) < 4 || value[:4] != "cor_" {
		return false
	}
	for _, c := range value[4:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// correlationMiddleware attaches or validates the request correlation id and
// stores it on the request context for downstream handlers and the audit sink.
func correlationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(correlationHeader)
		if id == "" || !validCorrelationID(id) {
			id = generateCorrelationID()
		}
		w.Header().Set(correlationHeader, id)
		next.ServeHTTP(w, r.WithContext(withCorrelationID(r.Context(), id)))
	})
}

type correlationKey struct{}

func withCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationKey{}, id)
}

// correlationIDFrom returns the id attached by the middleware, defaulting to
// the empty string when no id was attached.
func correlationIDFrom(ctx context.Context) string {
	if id, ok := ctx.Value(correlationKey{}).(string); ok {
		return id
	}
	return ""
}
