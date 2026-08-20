// Style ForgeLocal — T15-W2 : test E2E minimal du contrat complet.
// POST /api/sessions/{id}/navigate avec une URL externe => 400 URL_REJECTED_LOCAL_ONLY,
// même si l'authentification est valide. Test sans Playwright réel : on injecte un
// manager vide et on vérifie le code produit par la politique avant tout appel navigateur.

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// chiRouterProbe builds the real navigate route without starting Playwright.
func chiRouterProbe() http.Handler {
	r := chi.NewRouter()
	r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(405) }))
	r.Use(corsLocal)
	r.Use(originGuard)
	var h handler
	r.Post("/api/sessions/{id}/navigate", h.navigate)
	return r
}

// navigatePolicyProbe reproduit le contrat exact du handler navigate pour la partie politique.
func navigatePolicyProbe(t *testing.T, url string) (int, string) {
	t.Helper()
	r := chiRouterProbe()
	rec := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"url": url, "wait_until": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/sessions/s1/navigate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:3000") // mutation — loopback requis par originGuard
	r.ServeHTTP(rec, req)
	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	return rec.Code, resp.Error.Code
}

func TestNavigateRejectsExternalURL(t *testing.T) {
	cases := []struct {
		url  string
		code int
		err  string
	}{
		// externes => rejetés AVANT tout accès session (fail-closed W2)
		{"https://example.com/", 400, "URL_REJECTED_LOCAL_ONLY"},
		{"http://192.168.1.1/admin", 400, "URL_REJECTED_LOCAL_ONLY"},
		{"javascript:alert(1)", 400, "URL_REJECTED_LOCAL_ONLY"},
		{"data:text/html,x", 400, "URL_REJECTED_LOCAL_ONLY"},
		{"http://127.0.0.1.evil.com/", 400, "URL_REJECTED_LOCAL_ONLY"},
		// URL valides : le handler n'atteint la session que si l'URL passe ;
		// avec un manager nil, on vérifie la politique côté URL uniquement via
		// ValidateLocalURL (couvert par TestValidateLocalURL) — ici on confirme
		// que l'ordre de validation refuse les externes avant getSessionPage.
	}
	for _, c := range cases {
		code, err := navigatePolicyProbe(t, c.url)
		if code != c.code || (c.err != "" && err != c.err) {
			t.Errorf("navigate url=%q -> %d/%s, want %d/%s", c.url, code, err, c.code, c.err)
		}
	}
}
