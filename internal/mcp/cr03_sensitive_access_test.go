package mcp

import "testing"

func TestCR03CookieToolsRefuseBeforeContextAccess(t *testing.T) {
	s := &Server{}
	for name, call := range map[string]func(map[string]any) (any, *mcpError){
		"get": s.toolGetCookies,
		"set": s.toolSetCookies,
	} {
		raw, mcpErr := call(map[string]any{"cookies": []any{map[string]any{"value": "CR03_COOKIE_SENTINEL"}}})
		if raw != nil || mcpErr == nil || mcpErr.Code != -32601 || mcpErr.Message != "SENSITIVE_COOKIE_ACCESS_DISABLED" {
			t.Fatalf("%s cookie tool result=%#v error=%+v, want pre-context refusal", name, raw, mcpErr)
		}
	}
}
