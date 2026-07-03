package mcp

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"browseforge/internal/groups"
	"browseforge/internal/humanize"
	"browseforge/internal/profile"
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

func TestToolSchemasRequiredFields(t *testing.T) {
	expected := map[string][]string{
		"list_profiles":      {},
		"create_profile":     {"name", "engine", "group"},
		"delete_profile":     {"profile_id"},
		"update_profile":     {"profile_id"},
		"list_groups":        {},
		"get_group":          {"group"},
		"update_group_proxy": {"group", "proxy"},
		"clear_group_proxy":  {"group"},
		"delete_group":       {"group"},
		"open_browser":       {"profile_id"},
		"close_browser":      {"profile_id"},
		"navigate":           {"profile_id", "url"},
		"click":              {"profile_id", "selector"},
		"type_text":          {"profile_id", "selector", "text"},
		"screenshot":         {"profile_id"},
		"get_content":        {"profile_id"},
		"evaluate":           {"profile_id", "script"},
		"new_tab":            {"profile_id"},
		"list_tabs":          {"profile_id"},
		"switch_tab":         {"profile_id", "index"},
		"close_tab":          {"profile_id", "index"},
		"web_search":         {"query"},
		"web_explore":        {"url"},
		"create_session":     {"profile_id"},
		"destroy_session":    {"session_id"},
		"list_sessions":      {},
		"gc_sessions":        {},
		"wait_for":           {"selector"},
		"get_page_state":     {},
		"get_cookies":        {},
		"set_cookies":        {"cookies"},
		"run_workflow":       {},
		"form_fill":          {"fields"},
		"select_option":      {"selector"},
		"check":              {"selector"},
		"press_key":          {"key"},
		"list_downloads":     {},
		"delete_download":    {"name"},
		"read_download":      {"name"},
		"web_extract":        {"schema"},
		"doctor_profile":     {"profile_id"},
	}

	seen := map[string]bool{}
	for _, toolDef := range tools {
		name, _ := toolDef["name"].(string)
		seen[name] = true
		want, ok := expected[name]
		if !ok {
			t.Fatalf("unexpected tool in registry: %s", name)
		}
		schema, ok := toolDef["inputSchema"].(map[string]any)
		if !ok {
			t.Fatalf("%s inputSchema type = %T", name, toolDef["inputSchema"])
		}
		rawRequired, ok := schema["required"].([]string)
		if !ok {
			t.Fatalf("%s required type = %T", name, schema["required"])
		}
		got := append([]string(nil), rawRequired...)
		slices.Sort(got)
		want = append([]string(nil), want...)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatalf("%s required = %v, want %v", name, got, want)
		}
	}
	for name := range expected {
		if !seen[name] {
			t.Fatalf("expected tool missing from registry: %s", name)
		}
	}
}

func TestToolUpdateGroupProxyReportsRestartOnlyWhenActive(t *testing.T) {
	groupStore, err := groups.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(nil, nil, humanize.Config{}, nil, "", "test", groupStore)

	raw, mcpErr := s.toolUpdateGroupProxy(map[string]any{
		"group":      "Client A",
		"proxy_mode": groups.ProxyModeEnforced,
		"proxy": map[string]any{
			"type": "socks5",
			"host": "proxy.example.com",
			"port": float64(1080),
		},
	})
	if mcpErr != nil {
		t.Fatalf("toolUpdateGroupProxy error = %v", mcpErr)
	}
	res, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", raw)
	}
	if res["restart_required"] != false {
		t.Fatalf("restart_required = %v", res["restart_required"])
	}
	if res["active_sessions"] != 0 {
		t.Fatalf("active_sessions = %v", res["active_sessions"])
	}
	g, ok := res["group"].(*groups.Group)
	if !ok || g.Proxy == nil || g.Proxy.Host != "proxy.example.com" {
		t.Fatalf("group result = %#v", res["group"])
	}
}

func TestToolDeleteGroupUngroupsProfilesAndClearsProxy(t *testing.T) {
	profileStore, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := &profile.Profile{Name: "Profile A", Engine: "firefox", Group: "Client A"}
	if err := profileStore.Create(p); err != nil {
		t.Fatal(err)
	}
	groupStore, err := groups.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := groupStore.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "proxy.example.com", Port: 1080}, groups.ProxyModeDefault); err != nil {
		t.Fatal(err)
	}
	s := NewServer(profileStore, nil, humanize.Config{}, nil, "", "test", groupStore)

	raw, mcpErr := s.toolDeleteGroup(map[string]any{"group": "Client A"})
	if mcpErr != nil {
		t.Fatalf("toolDeleteGroup error = %v", mcpErr)
	}
	res, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", raw)
	}
	if res["profiles_ungrouped"] != 1 {
		t.Fatalf("profiles_ungrouped = %v", res["profiles_ungrouped"])
	}
	updated, err := profileStore.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Group != "" {
		t.Fatalf("profile group = %q, want empty", updated.Group)
	}
	if g, ok := groupStore.Get("Client A"); ok {
		t.Fatalf("group proxy still exists = %+v", g)
	}
}

func TestActiveGroupDeleteErrorIsStructured(t *testing.T) {
	err := activeGroupDeleteError("Client A", 2)
	if err.Code != -32000 {
		t.Fatalf("code = %d", err.Code)
	}
	data, ok := err.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T", err.Data)
	}
	if data["code"] != "GROUP_HAS_ACTIVE_SESSIONS" || data["group"] != "Client A" || data["active_sessions"] != 2 || data["restart_required"] != true {
		t.Fatalf("data = %#v", data)
	}
}

func TestParseWorkflowArgsRequiresWorkflow(t *testing.T) {
	_, err := parseWorkflowArgs(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "workflow or yaml is required") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseWorkflowArgsFromObject(t *testing.T) {
	wf, err := parseWorkflowArgs(map[string]any{
		"workflow": map[string]any{
			"name": "test workflow",
			"steps": []any{
				map[string]any{"name": "sleep", "action": "sleep", "params": map[string]any{"seconds": 1}},
			},
		},
	})
	if err != nil {
		t.Fatalf("parseWorkflowArgs: %v", err)
	}
	if wf.Name != "test workflow" || len(wf.Steps) != 1 || wf.Steps[0].Action != "sleep" {
		t.Fatalf("workflow = %+v", wf)
	}
}

func TestResolveArtifactPathStaysInProfileArtifacts(t *testing.T) {
	got, err := resolveArtifactPath("/tmp/profile", "shots/home", ".jpg")
	if err != nil {
		t.Fatalf("resolveArtifactPath: %v", err)
	}
	if got != "/tmp/profile/artifacts/shots/home.jpg" {
		t.Fatalf("path = %q", got)
	}

	if _, err := resolveArtifactPath("/tmp/profile", "../escape.jpg", ".jpg"); err == nil {
		t.Fatal("expected traversal error")
	}
	if _, err := resolveArtifactPath("/tmp/profile", "/tmp/escape.jpg", ".jpg"); err == nil {
		t.Fatal("expected absolute path error")
	}
}

func TestResolveDownloadPathRequiresFileName(t *testing.T) {
	got, name, err := resolveDownloadPath("/tmp/profile", map[string]any{"name": "report.csv"})
	if err != nil {
		t.Fatalf("resolveDownloadPath: %v", err)
	}
	if name != "report.csv" || got != "/tmp/profile/downloads/report.csv" {
		t.Fatalf("path=%q name=%q", got, name)
	}
	if _, _, err := resolveDownloadPath("/tmp/profile", map[string]any{"name": "../report.csv"}); err == nil {
		t.Fatal("expected traversal error")
	}
}
