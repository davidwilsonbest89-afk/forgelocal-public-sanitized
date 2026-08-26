package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"forgelocal/internal/backup"
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
			items = append(items, h.redactProfile(item))
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

func (h *handler) readonlyProxies(w http.ResponseWriter, r *http.Request) {
	limit, cursor, ok := readOnlyPageRequest(w, r)
	if !ok {
		return
	}
	type proxyDTO struct {
		ID         string    `json:"id"`
		Name       string    `json:"name"`
		Type       string    `json:"type"`
		Region     string    `json:"region,omitempty"`
		HasSecret  bool      `json:"has_secret"`
		AssignedTo string    `json:"assigned_to,omitempty"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}
	items := make([]proxyDTO, 0)
	if h.proxyStore != nil {
		for _, p := range h.proxyStore.List() {
			items = append(items, proxyDTO{
				ID: p.ID, Name: p.Name, Type: p.Type, Region: p.Region,
				HasSecret: p.HasSecret,
				CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
			})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	start, validCursor := readOnlyStart(items, cursor, func(item proxyDTO) string { return item.ID })
	if !validCursor {
		writeError(w, http.StatusBadRequest, "INVALID_CURSOR", "cursor does not identify a proxy in this result")
		return
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	response := readOnlyPage[proxyDTO]{APIVersion: readOnlyAPIVersion, Data: items[start:end]}
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
		ID              string    `json:"id"`
		Name            string    `json:"name"`
		ProxyMode       string    `json:"proxy_mode"`
		ProxyConfigured bool      `json:"proxy_configured"`
		ProfileCount    int       `json:"profile_count"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
	}
	items := make([]groupDTO, 0)
	if h.readonlyCatalog != nil {
		catalogItems, err := h.readonlyCatalog.ListReadOnlyGroups(context.Background())
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "READONLY_CATALOG_UNAVAILABLE", "readonly catalog unavailable")
			return
		}
		for _, item := range catalogItems {
			created, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)
			updated, _ := time.Parse(time.RFC3339Nano, item.UpdatedAt)
			items = append(items, groupDTO{ID: item.ID, Name: item.Name, ProxyMode: item.ProxyMode, ProxyConfigured: item.ProxyConfigured, ProfileCount: item.ProfileCount, CreatedAt: created, UpdatedAt: updated})
		}
	} else if h.groupStore != nil {
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
	limit, cursor, ok := readOnlyPageRequest(w, r)
	if !ok {
		return
	}
	type runtimeDTO struct {
		ID                string `json:"id"`
		DisplayName       string `json:"display_name"`
		Version           string `json:"version,omitempty"`
		Architecture      string `json:"architecture,omitempty"`
		Status            string `json:"status,omitempty"`
		Enabled           bool   `json:"enabled"`
		PlatformSupported bool   `json:"platform_supported"`
		Candidate         bool   `json:"candidate"`
		Launchable        bool   `json:"launchable"`
	}
	items := make([]runtimeDTO, 0)
	seen := make(map[string]bool)
	if h.readonlyCatalog != nil {
		catalogItems, err := h.readonlyCatalog.ListReadOnlyRuntimeCandidates(context.Background())
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "READONLY_CATALOG_UNAVAILABLE", "readonly catalog unavailable")
			return
		}
		for _, item := range catalogItems {
			items = append(items, runtimeDTO{ID: item.ID, DisplayName: item.Name, Version: item.Version, Architecture: item.Architecture, Status: item.Status, Candidate: true, Launchable: false})
			seen[item.ID] = true
		}
	}
	if h.mgr != nil {
		for _, item := range h.mgr.RuntimeRegistry().List() {
			id := string(item.ID)
			if seen[id] {
				continue
			}
			candidate := item.ID == bfruntime.Camoufox
			launchable := item.Enabled && item.PlatformSupported && !candidate
			version := ""
			status := ""
			if launchable && h.qualifiedRegistry != nil {
				qualified, err := h.qualifiedRegistry.Get(context.Background(), item.ID)
				if err != nil {
					launchable = false
				} else {
					version = qualified.Version
					status = string(qualified.State)
				}
			}
			items = append(items, runtimeDTO{ID: id, DisplayName: item.DisplayName, Version: version, Status: status, Enabled: item.Enabled, PlatformSupported: item.PlatformSupported, Candidate: candidate, Launchable: launchable})
		}
	}
	start, validCursor := readOnlyStart(items, cursor, func(item runtimeDTO) string { return item.ID })
	if !validCursor {
		writeError(w, http.StatusBadRequest, "INVALID_CURSOR", "cursor does not identify a runtime in this result")
		return
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	response := readOnlyPage[runtimeDTO]{APIVersion: readOnlyAPIVersion, Data: items[start:end]}
	response.Page.Limit = limit
	if end < len(items) {
		response.Page.NextCursor = encodeReadOnlyCursor(items[end-1].ID)
	}
	writeJSON(w, http.StatusOK, response)
}

var _ backup.ReadOnlyCatalog = (*backup.SQLiteStore)(nil)

func (h *handler) redactProfile(item *profile.Profile) ReadOnlyProfile {
	if item == nil {
		return ReadOnlyProfile{}
	}
	proxyConfigured := item.Proxy != nil
	if !proxyConfigured && h.proxyStore != nil && item.ID != "" {
		_, proxyConfigured = h.proxyStore.AssignedProxyID(item.ID)
	}
	return ReadOnlyProfile{
		ID: item.ID, Name: item.Name, RuntimeID: item.RuntimeID, Group: item.Group,
		Tags: append([]string(nil), item.Tags...), CreatedAt: item.CreatedAt, LastUsed: item.LastUsed,
		ProxyConfigured: proxyConfigured,
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
