package workflow

import "testing"

func TestValidateWorkflowURLQueryAllowlist(t *testing.T) {
	accepted := []string{
		"http://127.0.0.1:19280/api/sessions/sess-1/screenshot?full_page=true",
		"http://localhost:19280/api/sessions/sess-1/screenshot?full_page=false",
	}
	for _, raw := range accepted {
		if _, err := validateWorkflowURL(raw); err != nil {
			t.Fatalf("validateWorkflowURL(%q) rejected contract query: %v", raw, err)
		}
	}
	rejected := []string{
		"http://127.0.0.1:19280/api/status?target=external",
		"http://127.0.0.1:19280/api/status?full_page=true&target=external",
		"http://127.0.0.1:19280/api/status?full_page=1",
		"http://example.invalid/api/status?full_page=true",
		"http://user:pass@127.0.0.1:19280/api/status?full_page=true",
	}
	for _, raw := range rejected {
		if _, err := validateWorkflowURL(raw); err == nil {
			t.Fatalf("validateWorkflowURL(%q) accepted unsafe or non-contract query", raw)
		}
	}
}
