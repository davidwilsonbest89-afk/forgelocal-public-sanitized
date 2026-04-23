package api

import (
	"archive/zip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"camoufoxmulti/internal/profile"

	"github.com/go-chi/chi/v5"
)

func (h *handler) exportProfile(w http.ResponseWriter, r *http.Request) {
	p, err := h.store.Get(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 404, "NOT_FOUND", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+p.ID+".zip")

	zw := zip.NewWriter(w)
	defer zw.Close()

	// Add profile.json
	data, _ := json.MarshalIndent(p, "", "  ")
	f, _ := zw.Create("profile.json")
	f.Write(data)

	// Add cookies backup if exists
	cookiePath := filepath.Join(p.ProfileDir, "cookies-backup.json")
	if raw, err := os.ReadFile(cookiePath); err == nil {
		f, _ := zw.Create("cookies-backup.json")
		f.Write(raw)
	}
}

func (h *handler) importProfile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10MB max
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "NO_FILE", "file field required")
		return
	}
	defer file.Close()

	// Read zip into memory
	body, _ := io.ReadAll(file)
	zr, err := zip.NewReader(strings.NewReader(string(body)), int64(len(body)))
	if err != nil {
		writeError(w, 400, "INVALID_ZIP", err.Error())
		return
	}

	// Find profile.json
	var profileData []byte
	for _, f := range zr.File {
		if f.Name == "profile.json" {
			rc, _ := f.Open()
			profileData, _ = io.ReadAll(rc)
			rc.Close()
		}
	}
	if profileData == nil {
		writeError(w, 400, "NO_PROFILE", "profile.json not found in zip")
		return
	}

	var p profile.Profile
	json.Unmarshal(profileData, &p)
	p.ID = "" // generate new ID
	p.ContainerID = ""

	if err := h.store.Create(&p); err != nil {
		writeError(w, 500, "IMPORT_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"data": p})
}

func (h *handler) backup(w http.ResponseWriter, r *http.Request) {
	profiles := h.store.List("", "")

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=backup-"+time.Now().Format("20060102")+".zip")

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, p := range profiles {
		data, _ := json.MarshalIndent(p, "", "  ")
		f, _ := zw.Create(p.ID + "/profile.json")
		f.Write(data)
	}
}

func (h *handler) restore(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20) // 50MB max
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "NO_FILE", err.Error())
		return
	}
	defer file.Close()

	body, _ := io.ReadAll(file)
	zr, err := zip.NewReader(strings.NewReader(string(body)), int64(len(body)))
	if err != nil {
		writeError(w, 400, "INVALID_ZIP", err.Error())
		return
	}

	imported := 0
	for _, f := range zr.File {
		if filepath.Base(f.Name) != "profile.json" {
			continue
		}
		rc, _ := f.Open()
		data, _ := io.ReadAll(rc)
		rc.Close()

		var p profile.Profile
		if json.Unmarshal(data, &p) != nil {
			continue
		}
		p.ID = ""
		p.ContainerID = ""
		if h.store.Create(&p) == nil {
			imported++
		}
	}
	writeJSON(w, 200, map[string]any{"data": map[string]any{"imported": imported}})
}

func (h *handler) shutdown(w http.ResponseWriter, r *http.Request) {
	h.mgr.Close()
	writeJSON(w, 200, map[string]any{"data": "shutting down"})
	go func() { time.Sleep(500 * time.Millisecond); os.Exit(0) }()
}
