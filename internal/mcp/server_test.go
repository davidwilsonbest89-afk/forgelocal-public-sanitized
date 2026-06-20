package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"browseforge/internal/humanize"
)

func TestBuildWebSearchMCPResultRawFallback(t *testing.T) {
	resp := &SearchResponse{
		Engine:         "duckduckgo",
		ExtractionMode: "raw_fallback",
		Results:        nil,
		RawFallback: &SearchRawFallback{
			PageTitle: "Synthetic SERP",
			Text:      "visible SERP text for LLM interpretation",
			CandidateLinks: []LinkRef{
				{Text: "Candidate A", URL: "https://example.com/a"},
				{Text: "Candidate B", URL: "https://example.com/b"},
			},
		},
	}

	got := buildWebSearchMCPResult("synthetic query", resp, "sess_test", "prof_test", true)

	if got["session_id"] != "sess_test" {
		t.Fatalf("session_id = %v", got["session_id"])
	}
	if got["profile_id"] != "prof_test" {
		t.Fatalf("profile_id = %v", got["profile_id"])
	}
	if got["session_created"] != true {
		t.Fatalf("session_created = %v", got["session_created"])
	}
	if got["extraction_mode"] != "raw_fallback" {
		t.Fatalf("extraction_mode = %v", got["extraction_mode"])
	}
	if got["engine"] != "duckduckgo" {
		t.Fatalf("engine = %v", got["engine"])
	}
	results, ok := got["results"].([]map[string]string)
	if !ok {
		t.Fatalf("results type = %T", got["results"])
	}
	if len(results) != 0 {
		t.Fatalf("results len = %d", len(results))
	}
	fallback, ok := got["raw_fallback"].(*SearchRawFallback)
	if !ok {
		t.Fatalf("raw_fallback type = %T", got["raw_fallback"])
	}
	if fallback.PageTitle != "Synthetic SERP" || fallback.Text == "" || len(fallback.CandidateLinks) != 2 {
		t.Fatalf("unexpected fallback = %+v", fallback)
	}

	content, ok := got["content"].([]map[string]any)
	if !ok || len(content) != 1 {
		t.Fatalf("content = %#v", got["content"])
	}
	text, _ := content[0]["text"].(string)
	for _, want := range []string{"duckduckgo", "mode: raw_fallback", "raw_fallback", "candidate_links", "visible SERP text"} {
		if !strings.Contains(text, want) {
			t.Fatalf("content text missing %q: %s", want, text)
		}
	}
}

func TestServeHTTPMalformedJSONReturnsBadRequest(t *testing.T) {
	srv := NewServer(nil, nil, humanizeNoopConfig(), nil, "", "test")
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("{"))
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":-32700`) {
		t.Fatalf("body missing parse error: %s", rec.Body.String())
	}
}

func TestBuildWebSearchMCPResultDefaultsEngine(t *testing.T) {
	got := buildWebSearchMCPResult("synthetic query", &SearchResponse{}, "sess_test", "prof_test", false)

	if got["engine"] != "google" {
		t.Fatalf("engine = %v", got["engine"])
	}
	if got["extraction_mode"] != "structured" {
		t.Fatalf("extraction_mode = %v", got["extraction_mode"])
	}
	content := got["content"].([]map[string]any)
	if !strings.Contains(content[0]["text"].(string), "Found 0 google results") {
		t.Fatalf("content text = %s", content[0]["text"])
	}
}

func TestWebSessionClosedReturnsExplicitError(t *testing.T) {
	sess := &WebSession{ID: "sess_test", Closed: true}

	_, err := sess.WebSearch("query", "", 1)
	if err == nil || !strings.Contains(err.Error(), "session is closed: sess_test") {
		t.Fatalf("WebSearch err = %v", err)
	}

	_, err = sess.WebExplore("https://example.com", 100, 1)
	if err == nil || !strings.Contains(err.Error(), "session is closed: sess_test") {
		t.Fatalf("WebExplore err = %v", err)
	}
}

func humanizeNoopConfig() humanize.Config {
	return humanize.Config{}
}

func TestSearchProviderRegistry(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"", "google"},
		{"google", "google"},
		{"BING", "bing"},
		{"duckduckgo", "duckduckgo"},
		{"ddg", "duckduckgo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := getSearchProvider(tt.name)
			if err != nil {
				t.Fatalf("getSearchProvider(%q): %v", tt.name, err)
			}
			if provider.Name() != tt.want {
				t.Fatalf("provider.Name() = %q, want %q", provider.Name(), tt.want)
			}
		})
	}

	_, err := getSearchProvider("unknown")
	if err == nil || !strings.Contains(err.Error(), "unsupported search engine") {
		t.Fatalf("unknown provider err = %v", err)
	}
}

func TestSearchProviderURLs(t *testing.T) {
	for name, wantHost := range map[string]string{
		"google":     "https://www.google.com/search?",
		"bing":       "https://www.bing.com/search?",
		"duckduckgo": "https://duckduckgo.com/html/?",
	} {
		provider, err := getSearchProvider(name)
		if err != nil {
			t.Fatalf("getSearchProvider(%q): %v", name, err)
		}
		got := provider.SearchURL("BrowseForge MCP")
		if !strings.HasPrefix(got, wantHost) {
			t.Fatalf("%s SearchURL = %q, want prefix %q", name, got, wantHost)
		}
		if !strings.Contains(got, "BrowseForge") || strings.Contains(got, " ") {
			t.Fatalf("%s SearchURL query not encoded as expected: %q", name, got)
		}
	}
}
