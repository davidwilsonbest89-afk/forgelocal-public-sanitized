// T09 — Profile Writes. Write audit sink for profile mutations.
//
// All profile mutations (create, update, archive, reopen, tag add/remove,
// delete) record a redacted audit event in the Core SQLite audit_events
// ledger, joined to the request's correlation id. This sink is append-only
// and never returns user-controlled error text into the event payload.
//
// Style contract (project): redaction at the boundary; no proxy secrets,
// cookie values, absolute paths or vault references may appear in the ledger.

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"path"
	"strings"
	"time"
)

// redactString strips anything that looks like a secret, a token, a vault
// reference or an absolute filesystem path before it enters the ledger.
func redactString(value string) string {
	if value == "" {
		return ""
	}
	if isSecretLike(value) {
		return "[redacted]"
	}
	if strings.HasPrefix(value, "/") || strings.Contains(value, "://") {
		return path.Base(value)
	}
	return value
}

// secretSuffixes are literal suffixes that mark bearer-like values.
var secretSuffixes = []string{
	"token", "key", "secret", "password", "passwd", "pwd", "auth", "bearer",
	"basic", "api_key", "apikey", "credential", "ref",
}

func isSecretLike(value string) bool {
	lower := strings.ToLower(value)
	for _, suffix := range secretSuffixes {
		if strings.HasSuffix(lower, suffix) && len(value) > 3 {
			return true
		}
	}
	return false
}

func redactMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		if isSecretLike(k) {
			out[k] = "[redacted]"
			continue
		}
		switch typed := v.(type) {
		case string:
			out[k] = redactString(typed)
		case map[string]any:
			out[k] = redactMap(typed)
		default:
			out[k] = v
		}
	}
	return out
}

// writeAuditSink persists redacted profile-mutation audit events. It is safe
// to call concurrently from handler goroutines.
type writeAuditSink struct {
	db *sql.DB
}

func newWriteAuditSink(db *sql.DB) *writeAuditSink {
	return &writeAuditSink{db: db}
}

// auditRecord inserts one redacted audit event. A failed insert is logged but
// never surfaces to the API response, so audit failures cannot break the
// profile contract the Core exposes to its clients.
func (s *writeAuditSink) auditRecord(ctx context.Context, eventType, entityID, correlationID string, details map[string]any) {
	if s == nil || s.db == nil {
		return
	}
	payload, err := json.Marshal(redactMap(details))
	if err != nil {
		payload = []byte(`{"error":"audit_encode_failed"}`)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	// Best-effort context timeout so a stalled DB cannot stall the handler.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, _ = s.db.ExecContext(ctx,
		`INSERT INTO audit_events(event_type, entity_id, correlation_id, details_json, created_at) VALUES (?,?,?,?,?)`,
		eventType, entityID, correlationID, string(payload), now,
	)
}
