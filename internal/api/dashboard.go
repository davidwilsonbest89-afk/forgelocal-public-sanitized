package api

import (
	_ "embed"
	"log/slog"
	"net/http"
)

//go:embed dashboard.html
var dashboardHTML []byte

func (h *handler) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(dashboardHTML); err != nil {
		slog.Warn("write dashboard response failed", "error", err)
	}
}
