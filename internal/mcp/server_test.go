package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"browseforge/internal/browser"
	"browseforge/internal/config"
	"browseforge/internal/groups"
	"browseforge/internal/humanize"
	"browseforge/internal/profile"
	bfruntime "browseforge/internal/runtime"
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
		"list_runtimes":      {},
		"list_profiles":      {},
		"create_profile":     {"name", "runtime_id", "group"},
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
		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatalf("%s properties type = %T", name, schema["properties"])
		}
		if name == "create_profile" || name == "update_profile" {
			if _, ok := properties["engine"]; ok {
				t.Fatalf("%s schema advertises deprecated profile engine", name)
			}
		}
		if name == "web_search" {
			if _, ok := properties["engine"]; !ok {
				t.Fatalf("web_search schema missing search engine property")
			}
		}
	}
	for name := range expected {
		if !seen[name] {
			t.Fatalf("expected tool missing from registry: %s", name)
		}
	}
}

func TestToolListRuntimesReturnsRuntimeDescriptors(t *testing.T) {
	s := NewServer(nil, testManagerWithRuntimeConfig(t, &config.Config{Runtimes: map[string]config.RuntimeConfig{
		"camoufox":     {BinaryPath: "/opt/camoufox"},
		"cloakbrowser": {BinaryPath: "/opt/cloakbrowser"},
	}}), humanize.Config{}, nil, "", "test")

	raw, mcpErr := s.toolListRuntimes(nil)
	if mcpErr != nil {
		t.Fatalf("toolListRuntimes error = %v", mcpErr)
	}
	var got []bfruntime.Descriptor
	if err := json.Unmarshal([]byte(resultText(t, raw)), &got); err != nil {
		t.Fatalf("decode list_runtimes text: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("runtime count = %d, want 3: %#v", len(got), got)
	}
	if got[0].ID != bfruntime.BrowseForgeChromium || got[0].Enabled {
		t.Fatalf("first runtime = %+v, want disabled BrowseForge Chromium descriptor", got[0])
	}
	if got[1].ID != bfruntime.Camoufox || got[1].BinaryPath != "/opt/camoufox" {
		t.Fatalf("second runtime = %+v, want Camoufox with configured binary path", got[1])
	}
	if got[2].ID != bfruntime.CloakBrowser || got[2].BinaryPath != "/opt/cloakbrowser" {
		t.Fatalf("third runtime = %+v, want CloakBrowser with configured binary path", got[2])
	}
	if got[1].Capabilities.SupportsAgentWebSessions || !got[2].Capabilities.SupportsAgentWebSessions {
		t.Fatalf("agent web session capabilities = Camoufox:%v CloakBrowser:%v", got[1].Capabilities.SupportsAgentWebSessions, got[2].Capabilities.SupportsAgentWebSessions)
	}
}

func TestToolCreateProfileAcceptsRuntimeID(t *testing.T) {
	enabled := true
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	s := NewServer(store, testManagerWithRuntimeConfig(t, &config.Config{Runtimes: map[string]config.RuntimeConfig{
		"cloakbrowser": {Enabled: &enabled},
	}}), humanize.Config{}, nil, "", "test")

	raw, mcpErr := s.toolCreateProfile(map[string]any{
		"name":       "Cloaked",
		"runtime_id": "cloakbrowser",
	})
	if mcpErr != nil {
		t.Fatalf("toolCreateProfile error = %v", mcpErr)
	}
	text := resultText(t, raw)
	if !strings.Contains(text, "runtime: cloakbrowser") {
		t.Fatalf("create_profile text = %s", text)
	}
	profiles := store.List("", "")
	if len(profiles) != 1 {
		t.Fatalf("stored profiles = %d, want 1", len(profiles))
	}
	if profiles[0].RuntimeID != "cloakbrowser" {
		t.Fatalf("stored runtime_id = %q, want cloakbrowser", profiles[0].RuntimeID)
	}
}

func TestToolCreateProfileRejectsDisabledRuntimeID(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	disabled := false
	s := NewServer(store, testManagerWithRuntimeConfig(t, &config.Config{Runtimes: map[string]config.RuntimeConfig{
		"camoufox": {Enabled: &disabled},
	}}), humanize.Config{}, nil, "", "test")

	raw, mcpErr := s.toolCreateProfile(map[string]any{
		"name":       "Disabled",
		"runtime_id": "camoufox",
	})
	if raw != nil {
		t.Fatalf("raw result = %#v, want nil", raw)
	}
	if mcpErr == nil || mcpErr.Code != -32602 || !strings.Contains(mcpErr.Message, `runtime "camoufox" is disabled`) {
		t.Fatalf("mcpErr = %+v, want -32602 disabled runtime", mcpErr)
	}
	if profiles := store.List("", ""); len(profiles) != 0 {
		t.Fatalf("stored profiles = %d, want 0 after disabled runtime rejection", len(profiles))
	}
}

func TestToolUpdateProfileRejectsDisabledRuntimeID(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p := &profile.Profile{Name: "Runtime Profile", RuntimeID: "cloakbrowser"}
	if err := store.Create(p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	enabled := true
	disabled := false
	s := NewServer(store, testManagerWithRuntimeConfig(t, &config.Config{Runtimes: map[string]config.RuntimeConfig{
		"camoufox":     {Enabled: &disabled},
		"cloakbrowser": {Enabled: &enabled},
	}}), humanize.Config{}, nil, "", "test")

	raw, mcpErr := s.toolUpdateProfile(map[string]any{
		"profile_id": p.ID,
		"runtime_id": "camoufox",
	})
	if raw != nil {
		t.Fatalf("raw result = %#v, want nil", raw)
	}
	if mcpErr == nil || mcpErr.Code != -32602 || !strings.Contains(mcpErr.Message, `runtime "camoufox" is disabled`) {
		t.Fatalf("mcpErr = %+v, want -32602 disabled runtime", mcpErr)
	}
	got, err := store.Get(p.ID)
	if err != nil {
		t.Fatalf("stored profile missing: %v", err)
	}
	if got.RuntimeID != "cloakbrowser" {
		t.Fatalf("stored runtime_id = %q, want unchanged cloakbrowser", got.RuntimeID)
	}
}

func TestToolCreateProfileRejectsDeprecatedEngine(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	s := NewServer(store, testManagerWithRuntimeConfig(t, &config.Config{}), humanize.Config{}, nil, "", "test")

	raw, mcpErr := s.toolCreateProfile(map[string]any{
		"name":       "Legacy",
		"runtime_id": "camoufox",
		"engine":     "firefox",
	})
	if raw != nil {
		t.Fatalf("raw result = %#v, want nil", raw)
	}
	if mcpErr == nil || mcpErr.Code != -32602 || !strings.Contains(mcpErr.Message, "engine was removed in v2") {
		t.Fatalf("mcpErr = %+v, want -32602 deprecated field", mcpErr)
	}
}

func TestToolUpdateProfileRejectsDeprecatedEngine(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p := &profile.Profile{Name: "Runtime Profile", RuntimeID: "camoufox"}
	if err := store.Create(p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	s := NewServer(store, testManagerWithRuntimeConfig(t, &config.Config{}), humanize.Config{}, nil, "", "test")

	raw, mcpErr := s.toolUpdateProfile(map[string]any{
		"profile_id": p.ID,
		"engine":     "firefox",
	})
	if raw != nil {
		t.Fatalf("raw result = %#v, want nil", raw)
	}
	if mcpErr == nil || mcpErr.Code != -32602 || !strings.Contains(mcpErr.Message, "engine was removed in v2") {
		t.Fatalf("mcpErr = %+v, want -32602 deprecated field", mcpErr)
	}
}

func TestToolUpdateProfileRejectsNonStringRuntimeID(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p := &profile.Profile{Name: "Runtime Profile", RuntimeID: "camoufox"}
	if err := store.Create(p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	s := NewServer(store, testManagerWithRuntimeConfig(t, &config.Config{}), humanize.Config{}, nil, "", "test")

	raw, mcpErr := s.toolUpdateProfile(map[string]any{
		"profile_id": p.ID,
		"runtime_id": float64(42),
	})
	if raw != nil {
		t.Fatalf("raw result = %#v, want nil", raw)
	}
	if mcpErr == nil || mcpErr.Code != -32602 || !strings.Contains(mcpErr.Message, "runtime_id") {
		t.Fatalf("mcpErr = %+v, want -32602 runtime_id validation", mcpErr)
	}
}

func TestSessionPoolRejectsUnsupportedRuntimeByCapability(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p := &profile.Profile{Name: "Camoufox Agent Session", RuntimeID: "camoufox"}
	if err := store.Create(p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	sp := &SessionPool{
		mgr:   testManagerWithRuntimeConfig(t, &config.Config{}),
		store: store,
		pools: map[string]*ProfileSessionPool{},
	}

	_, err = sp.CreateSession(p.ID)
	if err == nil {
		t.Fatal("expected unsupported runtime to reject agent web session")
	}
	if !strings.Contains(err.Error(), "runtime camoufox does not support agent web sessions") {
		t.Fatalf("error = %q, want capability-based runtime rejection", err.Error())
	}
	if strings.Contains(err.Error(), "Chromium profile") {
		t.Fatalf("error = %q, want runtime capability rejection rather than legacy engine rejection", err.Error())
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
	p := &profile.Profile{Name: "Profile A", RuntimeID: "camoufox", Group: "Client A"}
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

func resultText(t *testing.T, raw any) string {
	t.Helper()

	res, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", raw)
	}
	content, ok := res["content"].([]map[string]any)
	if !ok || len(content) != 1 {
		t.Fatalf("content = %#v", res["content"])
	}
	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatalf("content text type = %T", content[0]["text"])
	}
	return text
}

func testManagerWithRuntimeConfig(t *testing.T, cfg *config.Config) *browser.Manager {
	t.Helper()

	mgr := &browser.Manager{}
	field := reflect.ValueOf(mgr).Elem().FieldByName("runtimes")
	if !field.IsValid() {
		t.Fatal("browser.Manager.runtimes field missing")
	}
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(bfruntime.NewRegistry(cfg)))
	return mgr
}
