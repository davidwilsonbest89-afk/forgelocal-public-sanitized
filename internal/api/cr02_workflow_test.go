package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type cr02PanicBody struct{}

func (cr02PanicBody) Read([]byte) (int, error) { panic("workflow body must not be read") }
func (cr02PanicBody) Close() error             { return nil }

func TestCR02WorkflowHandlerRefusesBeforeBodyRead(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/workflow/run", nil)
	req.Body = io.NopCloser(cr02PanicBody{})
	rec := httptest.NewRecorder()

	WorkflowHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Body.String(); got == "" || !cr02ContainsString(got, WorkflowExecutionDisabledCode) {
		t.Fatalf("body = %q, want %s", got, WorkflowExecutionDisabledCode)
	}
}

func cr02ContainsString(s, want string) bool {
	return len(want) > 0 && len(s) >= len(want) && (s == want || cr02ContainsAt(s, want))
}

func cr02ContainsAt(s, want string) bool {
	for i := 0; i+len(want) <= len(s); i++ {
		if s[i:i+len(want)] == want {
			return true
		}
	}
	return false
}
