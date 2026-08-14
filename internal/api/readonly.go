package api

import (
	"encoding/base64"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"forgelocal/internal/profile"
	bfruntime "forgelocal/internal/runtime"
)

const readOnlyAPIVersion = "v1"

// ReadOnlyProfile is the dashboard-safe projection of a profile. It excludes
// proxy endpoints, secret references, credentials, fingerprints and all paths.
type ReadOnlyProfile struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	RuntimeID       string    `json:"runtime_id"`
	Group           string    `json:"group,omitempty"`
	Tags            []string  `json:"tags"`
	CreatedAt       time.Time `json:"created_at"`
	LastUsed        time.Time `json:"last_used"`
	ProxyConfigured bool      `json:"proxy_configured"`
}

type readOnlyPage[T any] struct {
	APIVersion string `json:"api_version"`
	Data       []T    `json:"data"`
	Page       struct {
		Limit      int    `json:"limit"`
		NextCursor string `json:"next_cursor,omitempty"`
	} `json:"page"`
}

func (h *handler) readonlyHealth(w http.ResponseWriter, _ *http.Request) {
	version := "unknown"
	if h.cfg != nil && h.cfg.Version != "" {
		version = h.cfg.Version
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"api_version":  readOnlyAPIVersion,
		"status":       "ok",
		"core_version": version,
	})
}

func (h *handler) readonlySummary(w http.ResponseWriter, _ *http.Request) {
	profiles := 0
	if h.store != nil {
		profiles = len(h.store.List("", ""))
	}
	groups := 0
	if h.groupStore != nil {
		groups = len(h.groupStore.List())
	}
	runtimes := 0
	if h.mgr != nil {
		runtimes = len(h.mgr.RuntimeRegistry().List())
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"api_version": readOnlyAPIVersion,
		"data": map[string]int{
			"profiles": profiles,
			"groups":   groups,
			"runtimes": runtimes,
		},
	})
}

func (h *handler) readonlyProfiles(w http.ResponseWriter, r *http.Request) {
	limit, cursor, ok := readOnlyPageRequest(w, r)
	if !ok {
		return
	}
	items := make([]ReadOnlyProfile, 0)
	if h.store != nil {
		for _, item := range h.store.List("", "") {
			items = append(items, redactProfile(item))
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	start, validCursor := readOnlyStart(items, cursor, func(item ReadOnlyProfile) string { return item.ID })
	if !validCursor {
		writeError(w, http.StatusBadRequest, "INVALID_CURSOR", "cursor does not identify a profile in this result")
		return
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	response := readOnlyPage[ReadOnlyProfile]{APIVersion: readOnlyAPIVersion, Data: items[start:end]}
	response.Page.Limit = limit
	if end < len(items) {
		response.Page.NextCursor = encodeReadOnlyCursor(items[end-1].ID)
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *handler) readonlyGroups(w http.ResponseWriter, r *http.Request) {
	limit, cursor, ok := readOnlyPageRequest(w, r)
	if !ok {
		return
	}
	type groupDTO struct {
		Name            string    `json:"name"`
		ProxyMode       string    `json:"proxy_mode"`
		ProxyConfigured bool      `json:"proxy_configured"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
	}
	items := make([]groupDTO, 0)
	if h.groupStore != nil {
		for _, item := range h.groupStore.List() {
			items = append(items, groupDTO{Name: item.Name, ProxyMode: item.ProxyMode, ProxyConfigured: item.Proxy != nil, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	start, validCursor := readOnlyStart(items, cursor, func(item groupDTO) string { return item.Name })
	if !validCursor {
		writeError(w, http.StatusBadRequest, "INVALID_CURSOR", "cursor does not identify a group in this result")
		return
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	response := readOnlyPage[groupDTO]{APIVersion: readOnlyAPIVersion, Data: items[start:end]}
	response.Page.Limit = limit
	if end < len(items) {
		response.Page.NextCursor = encodeReadOnlyCursor(items[end-1].Name)
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *handler) readonlyRuntimes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("cursor") != "" || r.URL.Query().Get("limit") != "" {
		writeError(w, http.StatusBadRequest, "PAGINATION_NOT_SUPPORTED", "runtime descriptors are returned as one bounded registry")
		return
	}
	type runtimeDTO struct {
		ID                bfruntime.ID `json:"id"`
		DisplayName       string       `json:"display_name"`
		Enabled           bool         `json:"enabled"`
		PlatformSupported bool         `json:"platform_supported"`
		Candidate         bool         `json:"candidate"`
		Launchable        bool         `json:"launchable"`
	}
	items := make([]runtimeDTO, 0)
	if h.mgr != nil {
		for _, item := range h.mgr.RuntimeRegistry().List() {
			candidate := item.ID == bfruntime.Camoufox
			items = append(items, runtimeDTO{ID: item.ID, DisplayName: item.DisplayName, Enabled: item.Enabled, PlatformSupported: item.PlatformSupported, Candidate: candidate, Launchable: item.Enabled && item.PlatformSupported && !candidate})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"api_version": readOnlyAPIVersion, "data": items})
}

func redactProfile(item *profile.Profile) ReadOnlyProfile {
	if item == nil {
		return ReadOnlyProfile{}
	}
	return ReadOnlyProfile{
		ID: item.ID, Name: item.Name, RuntimeID: item.RuntimeID, Group: item.Group,
		Tags: append([]string(nil), item.Tags...), CreatedAt: item.CreatedAt, LastUsed: item.LastUsed,
		ProxyConfigured: item.Proxy != nil,
	}
}

func readOnlyPageRequest(w http.ResponseWriter, r *http.Request) (int, string, bool) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			writeError(w, http.StatusBadRequest, "INVALID_LIMIT", "limit must be an integer from 1 through 100")
			return 0, "", false
		}
		limit = value
	}
	cursor, err := decodeReadOnlyCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CURSOR", "cursor is malformed")
		return 0, "", false
	}
	return limit, cursor, true
}

func encodeReadOnlyCursor(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func decodeReadOnlyCursor(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) == 0 || len(decoded) > 512 {
		return "", http.ErrNoCookie
	}
	return string(decoded), nil
}

func readOnlyStart[T any](items []T, cursor string, value func(T) string) (int, bool) {
	if cursor == "" {
		return 0, true
	}
	for index, item := range items {
		if strings.EqualFold(value(item), cursor) {
			return index + 1, true
		}
	}
	return 0, false
}
