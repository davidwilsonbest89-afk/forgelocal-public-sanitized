package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/playwright-community/playwright-go"
)

// --- Session endpoints ---

func (h *handler) createSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProfileID string `json:"profile_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "INVALID_BODY", err.Error())
		return
	}
	p, err := h.store.Get(req.ProfileID)
	if err != nil {
		writeError(w, 404, "PROFILE_NOT_FOUND", err.Error())
		return
	}
	sess, err := h.mgr.LaunchSession(p)
	if err != nil {
		writeError(w, 500, "LAUNCH_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"data": map[string]any{
		"session_id": sess.ID,
		"profile_id": sess.ProfileID,
		"engine":     sess.Engine,
	}})
}

func (h *handler) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions := h.mgr.ListSessions()
	var result []map[string]any
	for _, s := range sessions {
		result = append(result, map[string]any{
			"session_id": s.ID,
			"profile_id": s.ProfileID,
			"engine":     s.Engine,
		})
	}
	writeJSON(w, 200, map[string]any{"data": result, "total": len(result)})
}

func (h *handler) deleteSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.mgr.CloseSession(id); err != nil {
		writeError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": "closed"})
}

// --- Browser operation endpoints (via Playwright) ---

func (h *handler) getSessionPage(r *http.Request) (playwright.Page, error) {
	id := chi.URLParam(r, "id")
	sess, ok := h.mgr.GetSession(id)
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return sess.Page, nil
}

func (h *handler) navigate(w http.ResponseWriter, r *http.Request) {
	page, err := h.getSessionPage(r)
	if err != nil {
		writeError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	var req struct {
		URL       string `json:"url"`
		WaitUntil string `json:"wait_until,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	opts := playwright.PageGotoOptions{}
	if req.WaitUntil != "" {
		wu := playwright.WaitUntilState(req.WaitUntil)
		opts.WaitUntil = &wu
	}

	resp, err := page.Goto(req.URL, opts)
	if err != nil {
		writeError(w, 500, "NAVIGATE_FAILED", err.Error())
		return
	}
	status := 0
	if resp != nil {
		status = resp.Status()
	}
	writeJSON(w, 200, map[string]any{"data": map[string]any{"status": status, "url": page.URL()}})
}

func (h *handler) click(w http.ResponseWriter, r *http.Request) {
	page, err := h.getSessionPage(r)
	if err != nil {
		writeError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	var req struct {
		Selector string `json:"selector"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := page.Click(req.Selector); err != nil {
		writeError(w, 500, "CLICK_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": "clicked"})
}

func (h *handler) typeText(w http.ResponseWriter, r *http.Request) {
	page, err := h.getSessionPage(r)
	if err != nil {
		writeError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	var req struct {
		Selector string  `json:"selector"`
		Text     string  `json:"text"`
		Delay    float64 `json:"delay,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	opts := playwright.PageTypeOptions{}
	if req.Delay > 0 {
		opts.Delay = playwright.Float(req.Delay)
	}

	if err := page.Type(req.Selector, req.Text, opts); err != nil {
		writeError(w, 500, "TYPE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": "typed"})
}

func (h *handler) evaluate(w http.ResponseWriter, r *http.Request) {
	page, err := h.getSessionPage(r)
	if err != nil {
		writeError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	var req struct {
		Script string `json:"script"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := page.Evaluate(req.Script)
	if err != nil {
		writeError(w, 500, "EVAL_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": result})
}

func (h *handler) screenshot(w http.ResponseWriter, r *http.Request) {
	page, err := h.getSessionPage(r)
	if err != nil {
		writeError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	fullPage := r.URL.Query().Get("full_page") == "true"
	data, err := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(fullPage),
	})
	if err != nil {
		writeError(w, 500, "SCREENSHOT_FAILED", err.Error())
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(data)
}

func (h *handler) content(w http.ResponseWriter, r *http.Request) {
	page, err := h.getSessionPage(r)
	if err != nil {
		writeError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	selector := r.URL.Query().Get("selector")
	if selector != "" {
		text, err := page.Locator(selector).TextContent()
		if err != nil {
			writeError(w, 500, "CONTENT_FAILED", err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"data": text})
		return
	}
	html, err := page.Content()
	if err != nil {
		writeError(w, 500, "CONTENT_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": html})
}

func (h *handler) waitFor(w http.ResponseWriter, r *http.Request) {
	page, err := h.getSessionPage(r)
	if err != nil {
		writeError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	var req struct {
		Selector string  `json:"selector"`
		Timeout  float64 `json:"timeout,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	opts := playwright.PageWaitForSelectorOptions{}
	if req.Timeout > 0 {
		opts.Timeout = playwright.Float(req.Timeout)
	}

	_, err = page.WaitForSelector(req.Selector, opts)
	if err != nil {
		writeError(w, 408, "TIMEOUT", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": "found"})
}

func (h *handler) getCookies(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, ok := h.mgr.GetSession(id)
	if !ok {
		writeError(w, 404, "NOT_FOUND", "session not found")
		return
	}
	cookies, err := sess.Context.Cookies()
	if err != nil {
		writeError(w, 500, "COOKIES_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": cookies})
}

func (h *handler) setCookies(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, ok := h.mgr.GetSession(id)
	if !ok {
		writeError(w, 404, "NOT_FOUND", "session not found")
		return
	}
	body, _ := io.ReadAll(r.Body)
	var cookies []playwright.OptionalCookie
	if err := json.Unmarshal(body, &cookies); err != nil {
		writeError(w, 400, "INVALID_BODY", err.Error())
		return
	}
	if err := sess.Context.AddCookies(cookies); err != nil {
		writeError(w, 500, "SET_COOKIES_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"data": "cookies set"})
}

func (h *handler) playwrightEndpoint(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"data": map[string]string{
		"note": "Use REST API for browser operations. Direct Playwright connection available in Phase 3.",
	}})
}

// Backup/restore/shutdown are in backup.go
