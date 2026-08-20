// Style ForgeLocal — T15-W2 : file d'attente de cas pour la politique fail-closed.
// Chaque cas est une URL canonique ; le verdict attendu est explicite.

package api

import (
	"errors"
	"testing"
)

func TestValidateLocalURL(t *testing.T) {
	cases := []struct {
		raw  string
		want error
	}{
		// file:// — exact
		{"file:///home/user/page.html", nil},
		{"file://localhost/home/user/page.html", nil},

		// loopback IP — accepté
		{"http://127.0.0.1:3000/", nil},
		{"https://127.0.0.1:8443/path?q=1", nil},
		{"http://[::1]:8080/test", nil},

		// localhost — accepté
		{"http://localhost/", nil},
		{"https://localhost:443/test", nil},
		{"http://localhost:9222/json/version", nil},

		// cas pièges — refus
		{"", ErrInvalidURL},
		{"not-a-url", ErrInvalidURL},
		{"https://example.com/", ErrInvalidURL},
		{"http://127.0.0.1.evil.com/", ErrInvalidURL},
		{"http://localhost.rogue/", ErrInvalidURL},
		{"http://192.168.1.1/", ErrInvalidURL},
		{"http://10.0.0.1/", ErrInvalidURL},
		{"http://0.0.0.0/", ErrInvalidURL},
		{"http://[::]/", ErrInvalidURL},
		{"ftp://127.0.0.1/file", ErrInvalidURL},
		{"javascript:alert(1)", ErrInvalidURL},
		{"data:text/html,<script>", ErrInvalidURL},
		{"http://localhost@evil.com/", ErrInvalidURL},
		{"http://user:pass@127.0.0.1/", nil}, // userinfo local — accepté
		{"http://[::1]/", nil},                // IPv6 loopback
		{"HTTP://LOCALHOST/", nil},            // casse insensible
	}

	for _, c := range cases {
		err := ValidateLocalURL(c.raw)
		switch {
		case c.want == nil && err != nil:
			t.Errorf("ValidateLocalURL(%q) = %v, want nil", c.raw, err)
		case c.want != nil && !errors.Is(err, c.want):
			t.Errorf("ValidateLocalURL(%q) = %v, want %v", c.raw, err, c.want)
		}
	}
}
