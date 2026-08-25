package api

import (
	"encoding/json"
	"net/http"
)

// openAPIV1 is deliberately compact: it is the versioned contract index for
// the locally supported API surface. Detailed payload schemas stay owned by
// their handlers until a dedicated schema-authoring lot is opened.
func (h *handler) openAPIV1(w http.ResponseWriter, _ *http.Request) {
	capabilitySchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":  map[string]any{"type": "string"},
			"state": map[string]any{"type": "string", "enum": []string{"SUPPORTED", "UNSUPPORTED"}},
			"note":  map[string]any{"type": "string"},
		},
	}
	environmentDataSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"diagnostic_version": map[string]any{"type": "string", "example": "environment-projection-v4"},
			"observation_mode":   map[string]any{"type": "string", "enum": []string{"PROJECTED_METADATA_ONLY"}},
			"capabilities":       map[string]any{"type": "array", "items": capabilitySchema},
		},
	}
	environmentResponseSchema := map[string]any{
		"type":       "object",
		"properties": map[string]any{"data": environmentDataSchema},
	}
	environmentDiagnosticPath := map[string]any{
		"get": map[string]any{
			"summary": "Read-only redacted environment diagnostic",
			"parameters": []map[string]any{
				{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
			},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Projected metadata only; unsupported browser-runtime controls are never executed",
					"content": map[string]any{
						"application/json": map[string]any{"schema": environmentResponseSchema},
					},
				},
				"401": map[string]any{"description": "Missing or invalid bearer token"},
				"404": map[string]any{"description": "Unknown profile"},
			},
		},
	}

	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":   "ForgeLocal Local API",
			"version": h.cfg.Version,
		},
		"servers":  []map[string]string{{"url": "http://127.0.0.1"}},
		"security": []map[string][]string{{"bearerAuth": {}}},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"bearerAuth": map[string]any{"type": "http", "scheme": "bearer"},
			},
		},
		"paths": map[string]any{
			"/api/v1/profiles":                       map[string]any{"get": map[string]any{"summary": "List profiles"}, "post": map[string]any{"summary": "Create profile"}},
			"/api/v1/profiles/{id}":                  map[string]any{"get": map[string]any{"summary": "Get profile"}, "put": map[string]any{"summary": "Update profile"}, "delete": map[string]any{"summary": "Delete profile"}},
			"/api/v1/environment/profiles/{id}":      environmentDiagnosticPath,
			"/api/v1/groups":                         map[string]any{"get": map[string]any{"summary": "List groups"}},
			"/api/v1/runtimes":                       map[string]any{"get": map[string]any{"summary": "List runtime projections"}},
			"/api/v1/proxies":                        map[string]any{"get": map[string]any{"summary": "List proxies"}, "post": map[string]any{"summary": "Create proxy"}},
			"/api/v1/templates":                      map[string]any{"get": map[string]any{"summary": "List templates"}, "post": map[string]any{"summary": "Create template"}},
			"/api/v1/extensions/import":              map[string]any{"post": map[string]any{"summary": "Import a local ZIP extension package"}},
			"/api/v1/extensions":                     map[string]any{"get": map[string]any{"summary": "List redacted local extension series"}},
			"/api/v1/extensions/{id}":                map[string]any{"get": map[string]any{"summary": "Get a redacted extension series"}, "delete": map[string]any{"summary": "Explicitly purge an unreferenced version"}},
			"/api/v1/extensions/{versionID}/approve": map[string]any{"post": map[string]any{"summary": "Approve a local extension version"}},
			"/api/v1/extensions/{versionID}/assign":  map[string]any{"post": map[string]any{"summary": "Assign an approved version to an existing profile"}},
			"/api/v1/extensions/{seriesID}/update":   map[string]any{"post": map[string]any{"summary": "Import an immutable next version"}},
			"/api/v1/extensions/{seriesID}/rollback": map[string]any{"post": map[string]any{"summary": "Rollback to an approved version"}},
			"/api/v1/extensions/{versionID}/revoke":  map[string]any{"post": map[string]any{"summary": "Revoke or quarantine a version"}},

			"/api/v1/readonly/health":   map[string]any{"get": map[string]any{"summary": "Read-only health"}},
			"/api/v1/readonly/profiles": map[string]any{"get": map[string]any{"summary": "Read-only profile projection"}},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(spec)
}
