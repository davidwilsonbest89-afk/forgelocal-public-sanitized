package launch

import (
	"strings"
)

// redacted returns a fail-closed, redacted error reason suitable for
// audit logs and session metadata. Any string that looks like a secret
// (bearer token, base64 blob, IP address, absolute file path, UUID-like
// value, or long opaque token) is replaced by a class label instead of
// being passed through.
//
// This is deliberately conservative: anything ambiguous is hidden.
func redacted(reason string) string {
	if reason == "" {
		return ""
	}
	for _, tok := range splitTokens(reason) {
		if isSecretish(tok) {
			return "redacted-error"
		}
	}
	return reason
}

// RecordEvent sanitises an entire audit reason before recording.
func RecordReason(reason string) string { return redacted(reason) }

func splitTokens(s string) []string {
	var out []string
	var cur strings.Builder
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == ',' || r == ':' || r == '=' {
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// isSecretish classifies a token as secret-looking (fail-closed).
func isSecretish(tok string) bool {
	if len(tok) > 16 { // opaque long values hidden (JWT fragments, hashes, tokens)
		return true
	}
	if len(tok) < 4 {
		return false
	}
	lower := strings.ToLower(tok)
	for _, prefix := range []string{"bearer", "basic", "token=", "apikey", "secret", "passwd", "password", "session="} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	// IPv4 / IPv6 fragments
	for _, c := range tok {
		if !(c >= '0' && c <= '9') && c != '.' && c != ':' {
			goto nonIP
		}
	}
	if strings.Count(tok, ".") >= 3 || strings.Count(tok, ":") >= 2 {
		return true
	}
nonIP:
	// absolute paths
	if strings.HasPrefix(tok, "/") || strings.HasPrefix(tok, "\\") || strings.Contains(tok, "://") {
		return true
	}
	return false
}
