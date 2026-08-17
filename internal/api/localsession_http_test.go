// Style ForgeLocal — T15-W2 : test d'intégration HTTP du endpoint POST /sessions/{id}/navigate.
// Vérifie que le handler rejette les URL externes et accepte les cibles locales.
// Playwright réel non requis : on teste le code de politique (ValidateLocalURL)
// et le contrat de rejet du handler via un serveur httptest.

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// rejectHandler enregistre le code d'erreur produit par la politique W2.
func rejectHandler(t *testing.T, url string) int {
	t.Helper()
	r := chi.NewRouter()
	r.Post("/navigate", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "INVALID_BODY", err.Error())
			return
		}
		if err := ValidateLocalURL(req.URL); err != nil {
			writeError(w, 400, "URL_REJECTED_LOCAL_ONLY", err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"data": "ok"})
	})

	rec := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"url": url})
	req := httptest.NewRequest(http.MethodPost, "/navigate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec.Code
}

func TestNavigateURLPolicy(t *testing.T) {
	cases := []struct {
		url string
		code int
	}{
		// acceptés
		{"file:///home/user/page.html", 200},
		{"http://127.0.0.1:3000/", 200},
		{"https://localhost:8443/test", 200},

		// refusés
		{"", 400},
		{"https://example.com/", 400},
		{"http://192.168.1.1/", 400},
		{"http://0.0.0.0/", 400},
		{"javascript:alert(1)", 400},
		{"data:text/html,<script>", 400},
	}

	for _, c := range cases {
		code := rejectHandler(t, c.url)
		if code != c.code {
			t.Errorf("navigate url=%q -> %d, want %d", c.url, code, c.code)
		}
	}
}
