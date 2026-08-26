package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// RunStdio runs the MCP server over stdin/stdout (for Kiro, Claude, etc.)
func (s *Server) RunStdio() error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req mcpRequest
		if err := json.Unmarshal(line, &req); err != nil {
			if err := writeStdout(map[string]any{"jsonrpc": "2.0", "id": nil, "error": map[string]any{"code": -32700, "message": "Parse error"}}); err != nil {
				return fmt.Errorf("write parse error response: %w", err)
			}
			continue
		}

		var result any
		var mcpErr *mcpError

		switch req.Method {
		case "initialize":
			result = s.handleInitialize(req.Params)
		case "tools/list":
			result = s.handleToolsList()
		case "tools/call":
			result, mcpErr = s.handleToolsCall(req.Params, nil)
		case "notifications/initialized":
			continue // client ack, ignore
		default:
			mcpErr = newError(-32601, "Method not found: "+req.Method)
		}

		resp := map[string]any{"jsonrpc": "2.0", "id": req.ID}
		if mcpErr != nil {
			resp["error"] = mcpErr
		} else if result != nil {
			resp["result"] = result
		}
		if err := writeStdout(resp); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read MCP stdin: %w", err)
	}
	return nil
}

func writeStdout(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s\n", data); err != nil {
		return err
	}
	return os.Stdout.Sync()
}
