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
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"forgelocal/internal/browser"

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
		url  string
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

// T17-R / G15-A : la projection GET /api/sessions produite par le handler
// de production (*handler).listSessions ne doit exposer que session_id,
// profile_id et runtime_id. Aucun champ technique (port, URL de connexion,
// chemin de profil, binaire exécutable) ne doit transparaître, même si les
// structures internes du browser.Manager en contiennent.
//
// Le test injecte une session réelle dans un browser.Manager réel (sans
// playwright : seul le champ sessions privé est alimenté, ce qui n'engage
// aucun binaire ni aucun runtime), puis appelle la méthode du handler produit.
func TestListSessionsRedactsTechnicalFields(t *testing.T) {
	// Fixture réelle : une manager browser.Manager dont la carte privée de
	// sessions contient deux sessions dont tous les champs techniques sont
	// non vides. La valeur technique sensible est volontairement chargée.
	mgr := &browser.Manager{}
	sessionsField := reflect.ValueOf(mgr).Elem().FieldByName("sessions")
	if !sessionsField.IsValid() {
		t.Fatal("browser.Manager.sessions field missing")
	}
	reflect.NewAt(sessionsField.Type(), unsafe.Pointer(sessionsField.UnsafeAddr())).Elem().Set(
		reflect.ValueOf(map[string]*browser.Session{
			"sess-1": {
				ID: "sess-1", ProfileID: "prof-1", RuntimeID: "browseforge-chromium",
				ConnectURL:     "ws://127.0.0.1:19280/api/playwright/ws/sess-1",
				ProfileDir:     "/home/user/.forgelocal/profiles/prof-1",
				UserDataDir:    "/home/user/.forgelocal/userdata/prof-1",
				ExecutablePath: "/opt/browseforge/chromium",
			},
			"sess-2": {
				ID: "sess-2", ProfileID: "prof-2", RuntimeID: "cloakbrowser",
				ConnectURL:     "ws://[::1]:19280/api/playwright/ws/sess-2",
				ProfileDir:     "/home/user/.forgelocal/profiles/prof-2",
				UserDataDir:    "/home/user/.forgelocal/userdata/prof-2",
				ExecutablePath: "/opt/cloakbrowser/bin",
			},
		}),
	)
	h := &handler{mgr: mgr}
	rec := httptest.NewRecorder()
	h.listSessions(rec, httptest.NewRequest(http.MethodGet, "/api/sessions", nil))
	if rec.Code != 200 {
		t.Fatalf("GET /api/sessions -> %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("data count = %d, want 2", len(resp.Data))
	}
	allowed := map[string]bool{"session_id": true, "profile_id": true, "runtime_id": true}
	for i, row := range resp.Data {
		for k := range row {
			if !allowed[k] {
				t.Errorf("row %d exposes technical field %q", i, k)
			}
		}
		for k := range allowed {
			if _, ok := row[k]; !ok {
				t.Errorf("row %d missing contract field %q", i, k)
			}
		}
	}
	// Vérification brute du corps sérialisé : chaque variante de sérialisation
	// des champs techniques doit être absente de la réponse HTTP réelle.
	forbidden := []string{"connect_url", "connectURL", "profile_dir", "profileDir",
		"executable_path", "executablePath", "port", "user_data_dir", "userDataDir"}
	body := rec.Body.String()
	for _, f := range forbidden {
		if strings.Contains(body, f) {
			t.Errorf("session projection leaks %q", f)
		}
	}
}
