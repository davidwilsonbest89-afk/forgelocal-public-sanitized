package mcp

var tools = []map[string]any{
	toolWithRequired("list_profiles", "列出瀏覽器 profiles，可依 group 或 tag 過濾", map[string]any{
		"group": prop("string", "Optional. 依分組過濾"),
		"tag":   prop("string", "Optional. 依標籤過濾"),
	}, []string{}),
	tool("create_profile", "建立新 Profile", map[string]any{
		"name": prop("string", "Profile 名稱"), "engine": prop("string", "firefox 或 chromium"), "group": prop("string", "分組名稱"),
	}),
	tool("delete_profile", "刪除 Profile", map[string]any{"profile_id": prop("string", "Profile ID")}),
	toolWithRequired("update_profile", "更新 Profile 設定", map[string]any{
		"profile_id": prop("string", "Profile ID"),
		"name":       prop("string", "Optional. 新名稱"),
		"group":      prop("string", "Optional. 新分組"),
		"proxy":      prop("object", "Optional. Proxy 設定"),
	}, []string{"profile_id"}),
	tool("open_browser", "開啟瀏覽器", map[string]any{"profile_id": prop("string", "Profile ID")}),
	tool("close_browser", "關閉瀏覽器", map[string]any{"profile_id": prop("string", "Profile ID")}),
	toolWithRequired("navigate", "導航目前分頁到指定 URL", map[string]any{
		"profile_id": prop("string", "Profile ID"),
		"url":        prop("string", "目標 URL"),
		"wait_until": prop("string", "Optional. 等待策略：load/domcontentloaded/networkidle/commit；預設 Playwright 行為"),
	}, []string{"profile_id", "url"}),
	toolWithRequired("click", "點擊頁面元素", map[string]any{
		"profile_id": prop("string", "Profile ID"),
		"selector":   prop("string", "CSS selector"),
		"timeout":    prop("number", "Optional. 點擊前等待元素出現的毫秒數"),
	}, []string{"profile_id", "selector"}),
	toolWithRequired("type_text", "在頁面元素中輸入文字", map[string]any{
		"profile_id": prop("string", "Profile ID"),
		"selector":   prop("string", "CSS selector"),
		"text":       prop("string", "要輸入的文字"),
		"clear":      prop("boolean", "Optional. 輸入前先清空欄位；預設 false"),
	}, []string{"profile_id", "selector", "text"}),
	toolWithRequired("screenshot", "擷取頁面截圖", map[string]any{
		"profile_id": prop("string", "Profile ID"),
		"quality":    prop("number", "Optional. JPEG 品質 1-100；預設 40"),
		"full_page":  prop("boolean", "Optional. 是否截全頁；預設 false"),
	}, []string{"profile_id"}),
	toolWithRequired("get_content", "取得頁面 HTML 或指定元素文字", map[string]any{
		"profile_id": prop("string", "Profile ID"),
		"selector":   prop("string", "Optional. CSS selector；未提供時回傳頁面 HTML"),
	}, []string{"profile_id"}),
	tool("evaluate", "執行 JavaScript", map[string]any{"profile_id": prop("string", "Profile ID"), "script": prop("string", "JS 程式碼")}),
	toolWithRequired("new_tab", "開啟新分頁", map[string]any{
		"profile_id": prop("string", "Profile ID"),
		"url":        prop("string", "Optional. 新分頁要導航的 URL"),
	}, []string{"profile_id"}),
	tool("list_tabs", "列出所有分頁", map[string]any{"profile_id": prop("string", "Profile ID")}),
	tool("switch_tab", "切換到指定分頁", map[string]any{"profile_id": prop("string", "Profile ID"), "index": prop("number", "分頁索引（從 0 開始）")}),
	tool("close_tab", "關閉指定分頁", map[string]any{"profile_id": prop("string", "Profile ID"), "index": prop("number", "要關閉的分頁索引")}),
	toolWithRequired("web_search", "網頁搜尋，使用指定 profile 的 CloakBrowser 並為 agent session 開獨立分頁", map[string]any{
		"query":       prop("string", "搜尋查詢文字"),
		"engine":      prop("string", "搜尋引擎：google、bing、duckduckgo（預設 google）"),
		"profile_id":  prop("string", "Chromium profile ID；session_id 未提供時必填"),
		"session_id":  prop("string", "agent session ID；提供時重用既有分頁，未提供則自動建立"),
		"max_results": prop("number", "最大結果數量（預設 10，最大 30）"),
	}, []string{"query"}),
	toolWithRequired("web_explore", "探索指定網頁，使用指定 profile 的 CloakBrowser 並為 agent session 開獨立分頁", map[string]any{
		"url":             prop("string", "要探索的 URL（可省略 http/https 前綴）"),
		"profile_id":      prop("string", "Chromium profile ID；session_id 未提供時必填"),
		"session_id":      prop("string", "agent session ID；提供時重用既有分頁，未提供則自動建立"),
		"max_text_length": prop("number", "最大文字長度（預設 3000）"),
		"max_links":       prop("number", "最大連結數量（預設 50）"),
	}, []string{"url"}),
	toolWithRequired("create_session", "為指定 Chromium profile 建立 agent web session（獨立分頁）", map[string]any{
		"profile_id": prop("string", "Chromium profile ID"),
	}, []string{"profile_id"}),
	toolWithRequired("destroy_session", "銷毀指定 agent web session（關閉分頁）", map[string]any{
		"session_id": prop("string", "agent session ID"),
	}, []string{"session_id"}),
	toolWithRequired("list_sessions", "列出活躍 agent web sessions", map[string]any{
		"profile_id": prop("string", "依 profile ID 過濾（選填）"),
	}, []string{}),
	toolWithRequired("gc_sessions", "立即執行 agent web session GC", map[string]any{}, []string{}),
}

func tool(name, desc string, props map[string]any) map[string]any {
	required := []string{}
	for k := range props {
		required = append(required, k)
	}
	return toolWithRequired(name, desc, props, required)
}

func toolWithRequired(name, desc string, props map[string]any, required []string) map[string]any {
	return map[string]any{
		"name": name, "description": desc,
		"inputSchema": map[string]any{"type": "object", "properties": props, "required": required},
	}
}

func prop(t, desc string) map[string]string { return map[string]string{"type": t, "description": desc} }
