// T24 — Bounded bulk profile mutations.
//
// This endpoint deliberately coordinates existing Core-owned single-profile
// mutations. It does not introduce a second profile store, global transaction,
// dashboard write path, runtime action or secret-bearing payload. Each target
// is processed in request order and receives an explicit redacted result.

package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"forgelocal/internal/profile"
)

const (
	maxBulkProfileTargets = 50
	maxBulkRequestBytes   = 16 << 10
)

type bulkProfileOperationRequest struct {
	Operation  string   `json:"operation"`
	ProfileIDs []string `json:"profile_ids"`
	Tag        string   `json:"tag,omitempty"`
	Group      string   `json:"group,omitempty"`
}

type bulkProfileOperationResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Code   string `json:"code"`
}

type bulkProfileOperationSummary struct {
	Changed int `json:"changed"`
	Noop    int `json:"noop"`
	Failed  int `json:"failed"`
}

func (h *handler) bulkProfileOperation(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, maxBulkRequestBytes)
	defer r.Body.Close()
	var req bulkProfileOperationRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		profileMutationError(w, http.StatusBadRequest, "INVALID_BULK_REQUEST", "bulk request is invalid", correlationID)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		profileMutationError(w, http.StatusBadRequest, "INVALID_BULK_REQUEST", "bulk request is invalid", correlationID)
		return
	}
	if err := h.validateBulkProfileOperation(&req); err != nil {
		profileMutationError(w, http.StatusBadRequest, "INVALID_BULK_REQUEST", "bulk request is invalid", correlationID)
		return
	}

	results := make([]bulkProfileOperationResult, 0, len(req.ProfileIDs))
	summary := bulkProfileOperationSummary{}
	for index, id := range req.ProfileIDs {
		if err := r.Context().Err(); err != nil {
			for _, remainingID := range req.ProfileIDs[index:] {
				results = append(results, bulkProfileOperationResult{ID: remainingID, Status: "failed", Code: "BULK_CANCELLED"})
				summary.Failed++
			}
			break
		}
		result := h.applyBulkProfileOperation(r, req, id, correlationID, index)
		results = append(results, result)
		switch result.Status {
		case "changed":
			summary.Changed++
		case "noop":
			summary.Noop++
		default:
			summary.Failed++
		}
	}

	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"operation": req.Operation,
		"results":   results,
		"summary":   summary,
	}})
}

func (h *handler) validateBulkProfileOperation(req *bulkProfileOperationRequest) error {
	if req == nil {
		return profile.ErrInvalidName
	}
	req.Operation = strings.TrimSpace(req.Operation)
	if !isBulkOperation(req.Operation) || len(req.ProfileIDs) == 0 || len(req.ProfileIDs) > maxBulkProfileTargets {
		return profile.ErrInvalidName
	}
	seen := make(map[string]struct{}, len(req.ProfileIDs))
	for i, id := range req.ProfileIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			return profile.ErrInvalidName
		}
		if _, duplicate := seen[id]; duplicate {
			return profile.ErrInvalidName
		}
		seen[id] = struct{}{}
		req.ProfileIDs[i] = id
	}
	req.Tag = strings.TrimSpace(req.Tag)
	req.Group = strings.TrimSpace(req.Group)
	switch req.Operation {
	case "add_tag", "remove_tag":
		if !validBulkName(req.Tag) || req.Group != "" {
			return profile.ErrInvalidTag
		}
	case "set_group":
		if !validBulkName(req.Group) || req.Tag != "" {
			return profile.ErrInvalidGroup
		}
		if h.groupStore == nil {
			return profile.ErrInvalidGroup
		}
		if _, ok := h.groupStore.Get(req.Group); !ok {
			return profile.ErrInvalidGroup
		}
	case "archive", "reopen", "clear_group":
		if req.Tag != "" || req.Group != "" {
			return profile.ErrInvalidName
		}
	}
	return nil
}

func isBulkOperation(operation string) bool {
	switch operation {
	case "archive", "reopen", "add_tag", "remove_tag", "set_group", "clear_group":
		return true
	default:
		return false
	}
}

func validBulkName(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, c := range value {
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	return true
}

func (h *handler) applyBulkProfileOperation(r *http.Request, req bulkProfileOperationRequest, id, correlationID string, index int) bulkProfileOperationResult {
	unlock, err := h.store.WithHistorySequence(id)
	if err != nil {
		return h.bulkProfileFailure(r, req.Operation, id, correlationID, index, err)
	}
	defer unlock()

	changed, err := h.applyBulkProfileOperationLocked(r, req, id, correlationID)
	if err != nil {
		if isBulkNoop(req.Operation, err) {
			h.auditSink.auditRecord(r.Context(), "profile.bulk_noop", id, correlationID, map[string]any{"operation": req.Operation, "index": index})
			return bulkProfileOperationResult{ID: id, Status: "noop", Code: "NO_CHANGE"}
		}
		return h.bulkProfileFailure(r, req.Operation, id, correlationID, index, err)
	}
	if !changed {
		h.auditSink.auditRecord(r.Context(), "profile.bulk_noop", id, correlationID, map[string]any{"operation": req.Operation, "index": index})
		return bulkProfileOperationResult{ID: id, Status: "noop", Code: "NO_CHANGE"}
	}
	h.auditSink.auditRecord(r.Context(), "profile.bulk_changed", id, correlationID, map[string]any{"operation": req.Operation, "index": index})
	return bulkProfileOperationResult{ID: id, Status: "changed", Code: "OK"}
}

func (h *handler) applyBulkProfileOperationLocked(r *http.Request, req bulkProfileOperationRequest, id, correlationID string) (bool, error) {
	var p *profile.Profile
	var changed bool
	var err error
	switch req.Operation {
	case "archive":
		p, changed, err = h.store.ArchiveProfileResult(id)
	case "reopen":
		p, err = h.store.ReopenProfileResult(id)
		changed = err == nil
	case "add_tag":
		err = h.store.AddProfileTag(id, req.Tag)
		changed = err == nil
		if err == nil {
			p, err = h.store.Get(id)
		}
	case "remove_tag":
		err = h.store.RemoveProfileTag(id, req.Tag)
		changed = err == nil
		if err == nil {
			p, err = h.store.Get(id)
		}
	case "set_group":
		p, changed, err = h.store.SetProfileGroup(id, req.Group)
	case "clear_group":
		p, changed, err = h.store.SetProfileGroup(id, "")
	}
	if err != nil || !changed {
		return changed, err
	}
	if err := h.captureProfileHistory(r.Context(), p, historyActionForBulk(req.Operation), correlationID); err != nil {
		return false, err
	}
	return true, nil
}

func historyActionForBulk(operation string) string {
	switch operation {
	case "archive":
		return "archive"
	case "reopen":
		return "reopen"
	case "add_tag":
		return "tag_add"
	case "remove_tag":
		return "tag_remove"
	case "set_group":
		return "group_set"
	case "clear_group":
		return "group_clear"
	default:
		return ""
	}
}

func (h *handler) bulkProfileFailure(r *http.Request, operation, id, correlationID string, index int, err error) bulkProfileOperationResult {
	code := mapErrorCode(err)
	h.auditSink.auditRecord(r.Context(), "profile.bulk_failed", id, correlationID, map[string]any{"operation": operation, "index": index, "error": code})
	return bulkProfileOperationResult{ID: id, Status: "failed", Code: code}
}

func isBulkNoop(operation string, err error) bool {
	switch operation {
	case "add_tag":
		return errors.Is(err, profile.ErrAlreadyTagged)
	case "remove_tag":
		return errors.Is(err, profile.ErrTagNotAssigned)
	default:
		return false
	}
}
