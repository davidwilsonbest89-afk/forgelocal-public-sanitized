package mcp

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"browseforge/internal/browser"
	"browseforge/internal/humanize"
	"browseforge/internal/profile"

	"github.com/playwright-community/playwright-go"
)

// MCP Server — Model Context Protocol (2025-11-25 spec, Streamable HTTP transport)

type Server struct {
	store       *profile.Store
	mgr         *browser.Manager
	hcfg        humanize.Config
	sessionPool *SessionPool
	token       string
	version     string
	reqID       atomic.Int64
}

func NewServer(store *profile.Store, mgr *browser.Manager, hcfg humanize.Config, sessionPool *SessionPool, token, version string) *Server {
	if version == "" {
		version = "dev"
	}
	return &Server{store: store, mgr: mgr, hcfg: hcfg, sessionPool: sessionPool, token: token, version: version}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	// Bearer token auth (MCP spec: MUST return 401 + WWW-Authenticate for HTTP transport)
	if s.token != "" && !validBearerToken(r.Header.Get("Authorization"), s.token) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req mcpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONRPCStatus(w, http.StatusBadRequest, nil, newError(-32700, "Parse error"))
		return
	}

	var result any
	var mcpErr *mcpError

	switch req.Method {
	case "initialize":
		result = s.handleInitialize(req.Params)
	case "tools/list":
		result = s.handleToolsList()
	case "tools/call":
		result, mcpErr = s.handleToolsCall(req.Params)
	default:
		mcpErr = newError(-32601, "Method not found: "+req.Method)
	}

	writeJSONRPC(w, req.ID, mcpErr, result)
}

func validBearerToken(auth, token string) bool {
	const prefix = "Bearer "
	if token == "" || len(auth) < len(prefix) || auth[:len(prefix)] != prefix {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(auth[len(prefix):]), []byte(token)) == 1
}

func (s *Server) handleInitialize(params json.RawMessage) any {
	return map[string]any{
		"protocolVersion": "2025-11-25",
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]any{"name": "BrowseForge", "version": s.version},
	}
}

func (s *Server) handleToolsList() any {
	return map[string]any{"tools": tools}
}

func (s *Server) handleToolsCall(params json.RawMessage) (any, *mcpError) {
	var call struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &call); err != nil {
		return nil, newError(-32700, "Invalid arguments: "+err.Error())
	}

	if call.Name == "" {
		return nil, newError(-32602, "Tool name is required")
	}

	switch call.Name {
	case "list_profiles":
		return s.toolListProfiles(call.Arguments)
	case "create_profile":
		return s.toolCreateProfile(call.Arguments)
	case "delete_profile":
		return s.toolDeleteProfile(call.Arguments)
	case "update_profile":
		return s.toolUpdateProfile(call.Arguments)
	case "open_browser":
		return s.toolOpenBrowser(call.Arguments)
	case "close_browser":
		return s.toolCloseBrowser(call.Arguments)
	case "navigate":
		return s.toolNavigate(call.Arguments)
	case "click":
		return s.toolClick(call.Arguments)
	case "type_text":
		return s.toolTypeText(call.Arguments)
	case "screenshot":
		return s.toolScreenshot(call.Arguments)
	case "get_content":
		return s.toolGetContent(call.Arguments)
	case "evaluate":
		return s.toolEvaluate(call.Arguments)
	case "new_tab":
		return s.toolNewTab(call.Arguments)
	case "list_tabs":
		return s.toolListTabs(call.Arguments)
	case "switch_tab":
		return s.toolSwitchTab(call.Arguments)
	case "close_tab":
		return s.toolCloseTab(call.Arguments)
	case "web_search":
		return s.toolWebSearch(call.Arguments)
	case "web_explore":
		return s.toolWebExplore(call.Arguments)
	case "create_session":
		return s.toolCreateSession(call.Arguments)
	case "destroy_session":
		return s.toolDestroySession(call.Arguments)
	case "list_sessions":
		return s.toolListSessions(call.Arguments)
	case "gc_sessions":
		return s.toolGCSessions(call.Arguments)
	default:
		return nil, newError(-32602, "Unknown tool: "+call.Name)
	}
}

// --- Tool implementations ---

func (s *Server) toolListProfiles(args map[string]any) (any, *mcpError) {
	group, _ := args["group"].(string)
	tag, _ := args["tag"].(string)
	profiles := s.store.List(group, tag)
	var items []map[string]string
	for _, p := range profiles {
		items = append(items, map[string]string{"id": p.ID, "name": p.Name, "engine": p.Engine, "group": p.Group})
	}
	return textResult(fmt.Sprintf("Found %d profiles:\n%s", len(items), mustJSON(items))), nil
}

func (s *Server) toolCreateProfile(args map[string]any) (any, *mcpError) {
	name, _ := args["name"].(string)
	if name == "" {
		return nil, newError(-32602, "name is required")
	}
	engine, _ := args["engine"].(string)
	if engine == "" {
		engine = "firefox"
	}
	group, _ := args["group"].(string)

	p := &profile.Profile{Name: name, Engine: engine, Group: group}
	if err := s.store.Create(p); err != nil {
		return nil, newError(-32000, err.Error())
	}
	return textResult(fmt.Sprintf("Created profile %s (%s, %s)", p.ID, p.Name, p.Engine)), nil
}

func (s *Server) toolDeleteProfile(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	if s.sessionPool != nil {
		s.sessionPool.DestroyProfileSessions(id)
	}
	if err := s.closeProfileBrowser(id, true); err != nil {
		return nil, err
	}
	if err := s.store.Delete(id); err != nil {
		return nil, newError(-32000, err.Error())
	}
	return textResult("Deleted profile " + id), nil
}

func (s *Server) toolUpdateProfile(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	updates := map[string]any{}
	if v, ok := args["name"]; ok {
		updates["name"] = v
	}
	if v, ok := args["group"]; ok {
		updates["group"] = v
	}
	if v, ok := args["proxy"]; ok {
		updates["proxy"] = v
	}
	p, err := s.store.Update(id, updates)
	if err != nil {
		return nil, newError(-32000, err.Error())
	}
	return textResult(fmt.Sprintf("Updated profile %s (%s)", p.ID, p.Name)), nil
}

func (s *Server) toolOpenBrowser(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	p, err := s.store.Get(id)
	if err != nil {
		return nil, newError(-32000, err.Error())
	}
	sess, err := s.mgr.LaunchSession(p)
	if err != nil {
		return nil, newError(-32000, err.Error())
	}
	return textResult(fmt.Sprintf("Opened browser for %s (session: %s, engine: %s)", p.Name, sess.ID, sess.Engine)), nil
}

func (s *Server) closeProfileBrowser(profileID string, ignoreNotFound bool) *mcpError {
	if s.mgr == nil {
		return nil
	}
	sessID := "sess_" + profileID
	if err := s.mgr.CloseSession(sessID); err != nil {
		if ignoreNotFound && strings.HasPrefix(err.Error(), "session not found: ") {
			return nil
		}
		return newError(-32000, err.Error())
	}
	return nil
}

func (s *Server) toolCloseBrowser(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	if s.sessionPool != nil {
		s.sessionPool.DestroyProfileSessions(id)
	}
	if err := s.closeProfileBrowser(id, false); err != nil {
		return nil, err
	}
	return textResult("Closed browser for " + id), nil
}

func (s *Server) toolNavigate(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	url, _ := args["url"].(string)
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session for "+id)
	}
	opts := playwright.PageGotoOptions{}
	if wu, ok := args["wait_until"].(string); ok && wu != "" {
		wus := playwright.WaitUntilState(wu)
		opts.WaitUntil = &wus
	}
	if _, err := sess.Page.Goto(url, opts); err != nil {
		return nil, newError(-32000, err.Error())
	}
	return textResult(fmt.Sprintf("Navigated to %s", url)), nil
}

func (s *Server) toolClick(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	selector, _ := args["selector"].(string)
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session")
	}
	if t, ok := args["timeout"].(float64); ok && t > 0 {
		if err := sess.Page.Locator(selector).WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(t)}); err != nil {
			return nil, newError(-32000, err.Error())
		}
	}
	if err := humanize.Click(sess.Page, selector, s.hcfg); err != nil {
		return nil, newError(-32000, err.Error())
	}
	return textResult("Clicked " + selector), nil
}

func (s *Server) toolTypeText(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	selector, _ := args["selector"].(string)
	text, _ := args["text"].(string)
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session")
	}
	if clear, _ := args["clear"].(bool); clear {
		sess.Page.Locator(selector).Clear()
	}
	if err := humanize.Type(sess.Page, selector, text, s.hcfg); err != nil {
		return nil, newError(-32000, err.Error())
	}
	return textResult("Typed text into " + selector), nil
}

func (s *Server) toolScreenshot(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session")
	}

	quality := 40
	if q, ok := args["quality"].(float64); ok && q > 0 {
		quality = int(q)
	}
	fullPage := false
	if fp, ok := args["full_page"].(bool); ok {
		fullPage = fp
	}

	opts := playwright.PageScreenshotOptions{
		Type:     playwright.ScreenshotTypeJpeg,
		Quality:  playwright.Int(quality),
		FullPage: playwright.Bool(fullPage),
	}
	data, err := sess.Page.Screenshot(opts)
	if err != nil {
		return nil, newError(-32000, err.Error())
	}
	return imageResult(data), nil
}

func (s *Server) toolGetContent(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	selector, _ := args["selector"].(string)
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session")
	}
	var text string
	var err error
	if selector != "" {
		text, err = sess.Page.Locator(selector).TextContent()
	} else {
		text, err = sess.Page.Content()
	}
	if err != nil {
		return nil, newError(-32000, err.Error())
	}
	return textResult(text), nil
}

func (s *Server) toolEvaluate(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	script, _ := args["script"].(string)
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session")
	}
	result, err := sess.Page.Evaluate(script)
	if err != nil {
		return nil, newError(-32000, err.Error())
	}
	return textResult(fmt.Sprintf("%v", result)), nil
}

func (s *Server) toolNewTab(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session")
	}
	page, err := sess.Context.NewPage()
	if err != nil {
		return nil, newError(-32000, err.Error())
	}
	sess.Page = page
	url, _ := args["url"].(string)
	if url != "" {
		if _, err := page.Goto(url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateLoad,
			Timeout:   playwright.Float(30000),
		}); err != nil {
			return nil, newError(-32000, "navigate to new tab: "+err.Error())
		}
	}
	return textResult(fmt.Sprintf("New tab opened (total: %d)", len(sess.Context.Pages()))), nil
}

func (s *Server) toolListTabs(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session")
	}
	pages := sess.Context.Pages()
	var tabs []map[string]any
	for i, p := range pages {
		active := p == sess.Page
		tabs = append(tabs, map[string]any{"index": i, "url": p.URL(), "active": active})
	}
	return textResult(mustJSON(tabs)), nil
}

func (s *Server) toolSwitchTab(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	index := int(args["index"].(float64))
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session")
	}
	pages := sess.Context.Pages()
	if index < 0 || index >= len(pages) {
		return nil, newError(-32000, fmt.Sprintf("tab index %d out of range (0-%d)", index, len(pages)-1))
	}
	sess.Page = pages[index]
	return textResult(fmt.Sprintf("Switched to tab %d: %s", index, sess.Page.URL())), nil
}

func (s *Server) toolCloseTab(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	index := int(args["index"].(float64))
	sess, ok := s.mgr.GetSession("sess_" + id)
	if !ok {
		return nil, newError(-32000, "no active session")
	}
	pages := sess.Context.Pages()
	if index < 0 || index >= len(pages) {
		return nil, newError(-32000, fmt.Sprintf("tab index %d out of range (0-%d)", index, len(pages)-1))
	}
	pages[index].Close()
	if pages[index] == sess.Page {
		remaining := sess.Context.Pages()
		if len(remaining) > 0 {
			sess.Page = remaining[len(remaining)-1]
		}
	}
	return textResult(fmt.Sprintf("Closed tab %d (remaining: %d)", index, len(sess.Context.Pages()))), nil
}
