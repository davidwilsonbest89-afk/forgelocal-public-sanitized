package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/templates"
)

const maxTemplateBodyBytes = 64 << 10

type createTemplateRequest struct {
	Name    string            `json:"name"`
	Content templates.Content `json:"content"`
}

type createVersionRequest struct {
	ExpectedActiveVersion int               `json:"expected_active_version"`
	Content               templates.Content `json:"content"`
}

type draftTemplateRequest struct {
	BaseDraft templates.Content `json:"base_draft"`
}

func (h *handler) decodeTemplateBody(w http.ResponseWriter, r *http.Request, value any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxTemplateBodyBytes)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New("multiple JSON values")
	}
	return nil
}

func (h *handler) createTemplate(w http.ResponseWriter, r *http.Request) {
	var req createTemplateRequest
	correlation := correlationIDFrom(r.Context())
	if err := h.decodeTemplateBody(w, r, &req); err != nil {
		writeTemplateError(w, errors.New("invalid template request"), correlation)
		return
	}
	item, err := h.templateStore.Create(r.Context(), req.Name, req.Content, correlation)
	if err != nil {
		writeTemplateError(w, err, correlation)
		return
	}
	w.Header().Set(correlationHeader, correlation)
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *handler) listTemplates(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := templatePagination(w, r)
	if !ok {
		return
	}
	result, err := h.templateStore.List(r.Context(), limit, offset)
	if err != nil {
		writeTemplateError(w, err, correlationIDFrom(r.Context()))
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) getTemplateVersion(w http.ResponseWriter, r *http.Request) {
	version, ok := templateVersionParam(w, r)
	if !ok {
		return
	}
	item, err := h.templateStore.GetVersion(r.Context(), chi.URLParam(r, "id"), version)
	if err != nil {
		writeTemplateError(w, err, correlationIDFrom(r.Context()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *handler) createTemplateVersion(w http.ResponseWriter, r *http.Request) {
	var req createVersionRequest
	correlation := correlationIDFrom(r.Context())
	if err := h.decodeTemplateBody(w, r, &req); err != nil {
		writeTemplateError(w, errors.New("invalid template request"), correlation)
		return
	}
	item, err := h.templateStore.NewVersion(r.Context(), chi.URLParam(r, "id"), req.ExpectedActiveVersion, req.Content, correlation)
	if err != nil {
		writeTemplateError(w, err, correlation)
		return
	}
	w.Header().Set(correlationHeader, correlation)
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *handler) archiveTemplateVersion(w http.ResponseWriter, r *http.Request) {
	version, ok := templateVersionParam(w, r)
	if !ok {
		return
	}
	correlation := correlationIDFrom(r.Context())
	if err := h.templateStore.Archive(r.Context(), chi.URLParam(r, "id"), version, correlation); err != nil {
		writeTemplateError(w, err, correlation)
		return
	}
	w.Header().Set(correlationHeader, correlation)
	writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

func (h *handler) calculateTemplateDraft(w http.ResponseWriter, r *http.Request) {
	version, ok := templateVersionParam(w, r)
	if !ok {
		return
	}
	var req draftTemplateRequest
	correlation := correlationIDFrom(r.Context())
	if err := h.decodeTemplateBody(w, r, &req); err != nil {
		writeTemplateError(w, errors.New("invalid template request"), correlation)
		return
	}
	result, err := h.templateStore.Draft(r.Context(), chi.URLParam(r, "id"), version, req.BaseDraft, correlation)
	if err != nil && !errors.Is(err, templates.ErrConflict) {
		writeTemplateError(w, err, correlation)
		return
	}
	w.Header().Set(correlationHeader, correlation)
	if errors.Is(err, templates.ErrConflict) {
		writeJSON(w, http.StatusConflict, map[string]any{"data": result})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func templateVersionParam(w http.ResponseWriter, r *http.Request) (int, bool) {
	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil || version < 1 {
		writeTemplateError(w, templates.ErrInvalidVersion, correlationIDFrom(r.Context()))
		return 0, false
	}
	return version, true
}

func templatePagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit, offset := 50, 0
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			writeTemplateError(w, templates.ErrInvalidTemplate, correlationIDFrom(r.Context()))
			return 0, 0, false
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			writeTemplateError(w, templates.ErrInvalidTemplate, correlationIDFrom(r.Context()))
			return 0, 0, false
		}
	}
	return limit, offset, true
}

func writeTemplateError(w http.ResponseWriter, err error, correlation string) {
	w.Header().Set(correlationHeader, correlation)
	status, code := http.StatusBadRequest, "INVALID_TEMPLATE"
	switch {
	case errors.Is(err, templates.ErrNotFound):
		status, code = http.StatusNotFound, "TEMPLATE_NOT_FOUND"
	case errors.Is(err, templates.ErrNameActive):
		status, code = http.StatusConflict, "TEMPLATE_NAME_ACTIVE"
	case errors.Is(err, templates.ErrStaleVersion):
		status, code = http.StatusConflict, "TEMPLATE_VERSION_STALE"
	case errors.Is(err, templates.ErrVersionNotActive):
		status, code = http.StatusConflict, "TEMPLATE_VERSION_NOT_ACTIVE"
	case errors.Is(err, templates.ErrConflict):
		status, code = http.StatusConflict, "TEMPLATE_DRAFT_CONFLICT"
	case errors.Is(err, templates.ErrInvalidVersion):
		status, code = http.StatusBadRequest, "INVALID_TEMPLATE_VERSION"
	}
	writeError(w, status, code, "template request rejected")
}
