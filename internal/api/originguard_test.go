// Style ForgeLocal — G15-B : file d'attente pour originGuard (fail-closed).
// Mutations sans Origin/Referer loopback => 403 ; GET/HEAD/OPTIONS exempts.

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestOriginGuard(t *testing.T) {
	handler := func() http.Handler {
		r := chi.NewRouter()
		r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
		r.Use(originGuard)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
		r.Post("/profiles", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201) })
		r.Put("/profiles/{id}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
		r.Delete("/profiles/{id}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
		r.Patch("/profiles/{id}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
		return r
	}

	cases := []struct {
		method string
		path   string
		header map[string]string
		want   int
	}{
		// GET exempt
		{"GET", "/health", nil, 200},
		{"HEAD", "/health", nil, 200},
		{"OPTIONS", "/profiles", nil, 200},

		// mutations avec Origin loopback — acceptées
		{"POST", "/profiles", map[string]string{"Origin": "http://localhost:3000"}, 201},
		{"PUT", "/profiles/1", map[string]string{"Origin": "https://127.0.0.1:8080"}, 200},
		{"DELETE", "/profiles/1", map[string]string{"Origin": "http://[::1]:3000"}, 200},

		// mutations avec Referer loopback seulement — acceptées
		{"POST", "/profiles", map[string]string{"Referer": "http://localhost:3000/"}, 201},

		// mutations sans header — refusées
		{"POST", "/profiles", nil, 403},
		{"PUT", "/profiles/1", nil, 403},
		{"DELETE", "/profiles/1", nil, 403},
		{"PATCH", "/profiles/1", nil, 403},

		// mutations avec Origin externe — refusées
		{"POST", "/profiles", map[string]string{"Origin": "https://evil.com"}, 403},
		{"POST", "/profiles", map[string]string{"Origin": "http://192.168.1.5"}, 403},
		{"POST", "/profiles", map[string]string{"Origin": "http://127.0.0.1.evil.com"}, 403},

		// mutations avec Referer externe — refusées
		{"POST", "/profiles", map[string]string{"Referer": "https://evil.com/"}, 403},
	}

	for _, c := range cases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(c.method, c.path, nil)
		for k, v := range c.header {
			req.Header.Set(k, v)
		}
		handler().ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Errorf("%s %s headers=%v -> %d, want %d", c.method, c.path, c.header, rec.Code, c.want)
		}
	}
}
