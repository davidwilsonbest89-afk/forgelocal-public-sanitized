package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"forgelocal/internal/profile"

	"github.com/go-chi/chi/v5"
)

// exportProfile is a limited legacy profile export. BACK-01 backup and restore
// are exposed exclusively through authenticated /api/v1 endpoints.
func (h *handler) exportProfile(w http.ResponseWriter, r *http.Request) {
	p, err := h.store.Get(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": p.ID + ".zip"}))
	zw := zip.NewWriter(w)
	defer zw.Close()
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return
	}
	f, err := zw.Create("profile.json")
	if err != nil {
		return
	}
	if _, err := f.Write(data); err != nil {
		return
	}
	// Cookies and browser state are intentionally excluded: portable profile
	// export is not a backup format and must never serialize secrets in cleartext.
}

const (
	maxProfileImportBytes = 10 << 20
	maxProfileImportFiles = 10
	maxProfileImportParts = 2
	maxProfileJSONBytes   = 1 << 20
)

func (h *handler) importProfile(w http.ResponseWriter, r *http.Request) {
	// This import is intentionally limited and never accepts a ForgeLocal
	// backup. The FLBK container is the sole backup/restore format.
	r.Body = http.MaxBytesReader(w, r.Body, maxProfileImportBytes)
	reader, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FORM", err.Error())
		return
	}
	var body []byte
	for parts := 0; ; parts++ {
		part, partErr := reader.NextPart()
		if partErr == io.EOF {
			break
		}
		if partErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_FORM", partErr.Error())
			return
		}
		if parts >= maxProfileImportParts {
			if closeErr := part.Close(); closeErr != nil {
				writeError(w, http.StatusBadRequest, "READ_FAILED", closeErr.Error())
				return
			}
			writeError(w, http.StatusBadRequest, "INVALID_FORM", "too many multipart fields")
			return
		}
		if part.FormName() != "file" {
			if closeErr := part.Close(); closeErr != nil {
				writeError(w, http.StatusBadRequest, "READ_FAILED", closeErr.Error())
				return
			}
			continue
		}
		if body != nil {
			if closeErr := part.Close(); closeErr != nil {
				writeError(w, http.StatusBadRequest, "READ_FAILED", closeErr.Error())
				return
			}
			writeError(w, http.StatusBadRequest, "INVALID_FORM", "multiple file fields")
			return
		}
		body, err = io.ReadAll(io.LimitReader(part, maxProfileImportBytes+1))
		closeErr := part.Close()
		if err != nil {
			writeError(w, http.StatusBadRequest, "READ_FAILED", err.Error())
			return
		}
		if closeErr != nil {
			writeError(w, http.StatusBadRequest, "READ_FAILED", closeErr.Error())
			return
		}
	}
	if body == nil {
		writeError(w, http.StatusBadRequest, "NO_FILE", "file field required")
		return
	}
	if len(body) > maxProfileImportBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "IMPORT_TOO_LARGE", "profile import exceeds maximum size")
		return
	}
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ZIP", err.Error())
		return
	}
	if len(zr.File) == 0 || len(zr.File) > maxProfileImportFiles {
		writeError(w, http.StatusBadRequest, "INVALID_ZIP", "profile import has an invalid file count")
		return
	}
	var profileData []byte
	for _, f := range zr.File {
		if f.Name != "profile.json" || f.FileInfo().IsDir() || f.UncompressedSize64 > maxProfileJSONBytes {
			continue
		}
		if profileData != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ZIP", "profile import contains duplicate profile.json entries")
			return
		}
		rc, err := f.Open()
		if err != nil {
			writeError(w, http.StatusBadRequest, "READ_FAILED", err.Error())
			return
		}
		profileData, err = io.ReadAll(io.LimitReader(rc, maxProfileJSONBytes+1))
		closeErr := rc.Close()
		if err != nil {
			writeError(w, http.StatusBadRequest, "READ_FAILED", err.Error())
			return
		}
		if closeErr != nil {
			writeError(w, http.StatusBadRequest, "READ_FAILED", closeErr.Error())
			return
		}
		if len(profileData) > maxProfileJSONBytes {
			writeError(w, http.StatusBadRequest, "PROFILE_TOO_LARGE", "profile.json exceeds maximum size")
			return
		}
	}
	if profileData == nil {
		writeError(w, http.StatusBadRequest, "NO_PROFILE", "profile.json not found in zip")
		return
	}
	var p profile.Profile
	if err := json.Unmarshal(profileData, &p); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PROFILE", err.Error())
		return
	}
	p.ID = ""
	p.ContainerID = ""
	desc, err := h.mgr.RuntimeRegistry().ApplyProfileDefaults(&p)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_RUNTIME", err.Error())
		return
	}
	if err := requireEnabledRuntime(desc); err != nil {
		writeRuntimeValidationError(w, err)
		return
	}
	if err := h.store.Create(&p); err != nil {
		writeError(w, http.StatusInternalServerError, "IMPORT_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": p})
}

// backup is deliberately kept only for direct legacy callers and cannot create
// an archive. It is not registered as an HTTP route.
func (h *handler) backup(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusGone, "BACKUP_ENDPOINT_RETIRED", "use POST /api/v1/profiles/{id}/backups")
}

// restore is deliberately kept only for direct legacy callers and cannot
// process ZIP input. It is not registered as an HTTP route.
func (h *handler) restore(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusGone, "RESTORE_ENDPOINT_RETIRED", "use POST /api/v1/backups/{id}/restore")
}

func (h *handler) shutdown(w http.ResponseWriter, r *http.Request) {
	h.mgr.Close()
	writeJSON(w, http.StatusOK, map[string]any{"data": "shutting down"})
	go func() { time.Sleep(500 * time.Millisecond); os.Exit(0) }()
}
