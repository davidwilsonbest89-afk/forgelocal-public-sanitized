package mcp

import "testing"

func TestCR02ToolRunWorkflowIsDisabledBeforePayloadParsing(t *testing.T) {
	s := &Server{}
	raw, mcpErr := s.toolRunWorkflow(map[string]any{"yaml": "not: [valid"})
	if raw != nil {
		t.Fatalf("raw = %#v, want nil", raw)
	}
	if mcpErr == nil || mcpErr.Code != -32601 || mcpErr.Message != "WORKFLOW_EXECUTION_DISABLED" {
		t.Fatalf("mcpErr = %+v, want disabled workflow refusal", mcpErr)
	}
}
