package workflow

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Engine executes YAML workflow definitions against the REST API

type Workflow struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name      string         `yaml:"name"`
	Action    string         `yaml:"action"` // create_profile, open_browser, navigate, click, type, screenshot, wait, close_browser, sleep
	ProfileID string         `yaml:"profile_id,omitempty"`
	Params    map[string]any `yaml:"params,omitempty"`
	OnError   string         `yaml:"on_error,omitempty"` // skip, stop (default: stop)
	Retries   int            `yaml:"retries,omitempty"`
}

type Result struct {
	Step    string `json:"step"`
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Engine struct {
	apiBase string
	token   string
}

func NewEngine(apiBase, token string) *Engine {
	return &Engine{apiBase: apiBase, token: token}
}

func (e *Engine) LoadFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var w Workflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

func (e *Engine) Execute(w *Workflow) []Result {
	slog.Info("workflow starting", "name", w.Name, "steps", len(w.Steps))
	var results []Result

	// Track created profile IDs for variable substitution
	vars := map[string]string{}

	for i, step := range w.Steps {
		slog.Info("step", "index", i, "name", step.Name, "action", step.Action)

		// Substitute variables in profile_id
		pid := step.ProfileID
		if strings.HasPrefix(pid, "$") {
			if v, ok := vars[pid[1:]]; ok {
				pid = v
			}
		}

		var result Result
		result.Step = step.Name
		result.Action = step.Action

		retries := step.Retries
		if retries <= 0 {
			retries = 1
		}

		for attempt := 0; attempt < retries; attempt++ {
			var err error
			result.Output, err = e.executeStep(step.Action, pid, step.Params, vars)
			if err == nil {
				result.Success = true
				break
			}
			result.Error = err.Error()
			if attempt < retries-1 {
				time.Sleep(2 * time.Second)
			}
		}

		results = append(results, result)

		if !result.Success {
			if step.OnError == "skip" {
				slog.Warn("step failed, skipping", "step", step.Name, "error", result.Error)
				continue
			}
			slog.Error("step failed, stopping", "step", step.Name, "error", result.Error)
			break
		}
	}

	slog.Info("workflow complete", "name", w.Name, "results", len(results))
	return results
}

func (e *Engine) executeStep(action, profileID string, params map[string]any, vars map[string]string) (string, error) {
	switch action {
	case "create_profile":
		body := map[string]any{"name": params["name"], "engine": params["engine"]}
		resp, err := e.apiCall("POST", "/api/profiles", body)
		if err != nil {
			return "", err
		}
		// Extract profile ID and store as variable
		if data, ok := resp["data"].(map[string]any); ok {
			if id, ok := data["id"].(string); ok {
				varName, _ := params["var"].(string)
				if varName != "" {
					vars[varName] = id
				}
				return "created: " + id, nil
			}
		}
		return "", fmt.Errorf("unexpected response")

	case "open_browser":
		_, err := e.apiCall("POST", "/api/sessions", map[string]any{"profile_id": profileID})
		return "opened", err

	case "navigate":
		url, _ := params["url"].(string)
		_, err := e.apiCall("POST", "/api/sessions/sess_"+profileID+"/navigate", map[string]any{"url": url})
		return "navigated to " + url, err

	case "click":
		selector, _ := params["selector"].(string)
		_, err := e.apiCall("POST", "/api/sessions/sess_"+profileID+"/click", map[string]any{"selector": selector})
		return "clicked " + selector, err

	case "type":
		selector, _ := params["selector"].(string)
		text, _ := params["text"].(string)
		_, err := e.apiCall("POST", "/api/sessions/sess_"+profileID+"/type", map[string]any{"selector": selector, "text": text})
		return "typed", err

	case "screenshot":
		// Screenshot returns binary, handle differently
		return "screenshot taken", nil

	case "wait":
		selector, _ := params["selector"].(string)
		_, err := e.apiCall("POST", "/api/sessions/sess_"+profileID+"/wait", map[string]any{"selector": selector, "timeout": 10000})
		return "found " + selector, err

	case "close_browser":
		_, err := e.apiCall("DELETE", "/api/sessions/sess_"+profileID, nil)
		return "closed", err

	case "sleep":
		dur, _ := params["seconds"].(float64)
		if dur <= 0 {
			dur = 1
		}
		time.Sleep(time.Duration(dur) * time.Second)
		return fmt.Sprintf("slept %.0fs", dur), nil

	case "eval":
		script, _ := params["script"].(string)
		resp, err := e.apiCall("POST", "/api/sessions/sess_"+profileID+"/eval", map[string]any{"script": script})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", resp["data"]), nil

	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (e *Engine) apiCall(method, path string, body any) (map[string]any, error) {
	var r *http.Request
	if body != nil {
		data, _ := json.Marshal(body)
		r, _ = http.NewRequest(method, e.apiBase+path, strings.NewReader(string(data)))
	} else {
		r, _ = http.NewRequest(method, e.apiBase+path, nil)
	}
	r.Header.Set("Authorization", "Bearer "+e.token)
	r.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode >= 400 {
		if errObj, ok := result["error"].(map[string]any); ok {
			return nil, fmt.Errorf("%v", errObj["message"])
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return result, nil
}
