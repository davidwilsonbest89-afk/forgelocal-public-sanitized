package mcp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"browseforge/internal/browser"
	"browseforge/internal/humanize"
	"browseforge/internal/profile"

	"github.com/playwright-community/playwright-go"
)

// MCP Server — Model Context Protocol (2025-11-25 spec, Streamable HTTP transport)

type Server struct {
	store *profile.Store
	mgr   *browser.Manager
	hcfg  humanize.Config
	reqID atomic.Int64
}

func NewServer(store *profile.Store, mgr *browser.Manager, hcfg humanize.Config) *Server {
	return &Server{store: store, mgr: mgr, hcfg: hcfg}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", 405)
		return
	}

	var req mcpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONRPC(w, nil, newError(-32700, "Parse error"))
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

func (s *Server) handleInitialize(params json.RawMessage) any {
	return map[string]any{
		"protocolVersion": "2025-11-25",
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]any{"name": "BrowseForge", "version": "0.1.0"},
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
	json.Unmarshal(params, &call)

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

func (s *Server) toolCloseBrowser(args map[string]any) (any, *mcpError) {
	id, _ := args["profile_id"].(string)
	sessID := "sess_" + id
	if err := s.mgr.CloseSession(sessID); err != nil {
		return nil, newError(-32000, err.Error())
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

// --- MCP Protocol types ---

var tools = []map[string]any{
	tool("list_profiles", "列出所有瀏覽器 Profile", map[string]any{"group": prop("string", "依分組過濾（選填）"), "tag": prop("string", "依標籤過濾（選填）")}),
	tool("create_profile", "建立新 Profile", map[string]any{
		"name": prop("string", "Profile 名稱"), "engine": prop("string", "firefox 或 chromium"), "group": prop("string", "分組名稱"),
	}),
	tool("delete_profile", "刪除 Profile", map[string]any{"profile_id": prop("string", "Profile ID")}),
	tool("update_profile", "更新 Profile 設定（名稱、分組、Proxy）", map[string]any{"profile_id": prop("string", "Profile ID"), "name": prop("string", "新名稱"), "group": prop("string", "新分組")}),
	tool("open_browser", "開啟瀏覽器", map[string]any{"profile_id": prop("string", "Profile ID")}),
	tool("close_browser", "關閉瀏覽器", map[string]any{"profile_id": prop("string", "Profile ID")}),
	tool("navigate", "導航到 URL", map[string]any{"profile_id": prop("string", "Profile ID"), "url": prop("string", "目標 URL"), "wait_until": prop("string", "等待策略：load/domcontentloaded/networkidle/commit（預設 load）")}),
	tool("click", "點擊元素", map[string]any{"profile_id": prop("string", "Profile ID"), "selector": prop("string", "CSS selector"), "timeout": prop("number", "等待元素出現的毫秒數（選填）")}),
	tool("type_text", "輸入文字", map[string]any{"profile_id": prop("string", "Profile ID"), "selector": prop("string", "CSS selector"), "text": prop("string", "要輸入的文字"), "clear": prop("boolean", "輸入前先清空欄位（預設 false）")}),
	tool("screenshot", "截圖", map[string]any{"profile_id": prop("string", "Profile ID"), "quality": prop("number", "JPEG 品質 1-100（預設 40）"), "full_page": prop("boolean", "是否截全頁（預設 false，僅可視範圍）")}),
	tool("get_content", "取得頁面內容", map[string]any{"profile_id": prop("string", "Profile ID"), "selector": prop("string", "CSS selector（選填）")}),
	tool("evaluate", "執行 JavaScript", map[string]any{"profile_id": prop("string", "Profile ID"), "script": prop("string", "JS 程式碼")}),
}

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newError(code int, msg string) *mcpError { return &mcpError{Code: code, Message: msg} }

func tool(name, desc string, props map[string]any) map[string]any {
	required := []string{}
	for k := range props {
		required = append(required, k)
	}
	return map[string]any{
		"name": name, "description": desc,
		"inputSchema": map[string]any{"type": "object", "properties": props, "required": required},
	}
}

func prop(t, desc string) map[string]string { return map[string]string{"type": t, "description": desc} }

func textResult(text string) map[string]any {
	return map[string]any{"content": []map[string]any{{"type": "text", "text": text}}}
}

func imageResult(data []byte) map[string]any {
	return map[string]any{"content": []map[string]any{{"type": "image", "data": base64.StdEncoding.EncodeToString(data), "mimeType": "image/jpeg"}}}
}

func writeJSONRPC(w http.ResponseWriter, id any, err *mcpError, result ...any) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{"jsonrpc": "2.0", "id": id}
	if err != nil {
		resp["error"] = err
	} else if len(result) > 0 {
		resp["result"] = result[0]
	}
	json.NewEncoder(w).Encode(resp)
}

func mustJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
