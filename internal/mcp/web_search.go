package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"browseforge/internal/browser"
	"browseforge/internal/profile"

	"github.com/playwright-community/playwright-go"
)

const (
	defaultSessionIdleTTL        = 5 * time.Minute
	defaultSessionSweepEvery     = 1 * time.Minute
	defaultMaxSessionsPerProfile = 10
)

// SessionPool manages agent web_search/web_explore sessions.
//
// Design:
//   - browser.Manager owns one CloakBrowser/Chromium browser per profile.
//   - Each browser.Manager session exposes a Playwright Bind endpoint (ConnectURL).
//   - SessionPool connects to that endpoint and creates one independent Page per agent session.
//   - GC only closes per-agent pages; it never closes the profile browser.
type SessionPool struct {
	mu            sync.RWMutex
	mgr           *browser.Manager
	store         *profile.Store
	pw            *playwright.Playwright
	pools         map[string]*ProfileSessionPool // profileID -> pool
	idleTTL       time.Duration
	maxPerProfile int
}

// ProfileSessionPool holds remote Playwright state and agent pages for one profile browser.
type ProfileSessionPool struct {
	mu         sync.RWMutex
	profileID  string
	browserID  string
	connectURL string
	browser    playwright.Browser
	sessions   map[string]*WebSession // sessionID -> session
}

// WebSession represents one agent session: fixed profile + independent page.
type WebSession struct {
	mu           sync.Mutex
	ID           string
	ProfileID    string
	BrowserID    string
	Page         playwright.Page
	CreatedAt    time.Time
	LastAccessed time.Time
	Closed       bool
}

// SessionInfo is returned by list_sessions and GC diagnostics.
type SessionInfo struct {
	ID           string    `json:"id"`
	ProfileID    string    `json:"profile_id"`
	BrowserID    string    `json:"browser_id"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	IdleSeconds  int64     `json:"idle_seconds"`
}

// NewSessionPool creates a page-level session pool backed by browser.Manager profile browsers.
func NewSessionPool(mgr *browser.Manager, store *profile.Store) (*SessionPool, error) {
	if mgr == nil {
		return nil, fmt.Errorf("browser manager is required")
	}
	if store == nil {
		return nil, fmt.Errorf("profile store is required")
	}
	pw := mgr.Playwright()
	if pw == nil {
		return nil, fmt.Errorf("playwright driver is not available")
	}

	return &SessionPool{
		mgr:           mgr,
		store:         store,
		pw:            pw,
		pools:         make(map[string]*ProfileSessionPool),
		idleTTL:       defaultSessionIdleTTL,
		maxPerProfile: defaultMaxSessionsPerProfile,
	}, nil
}

// SweepInterval returns the recommended GC sweep interval.
func (sp *SessionPool) SweepInterval() time.Duration {
	return defaultSessionSweepEvery
}

// StartGC starts background GC and returns a function that stops it.
func (sp *SessionPool) StartGC(interval time.Duration) func() {
	if interval <= 0 {
		interval = defaultSessionSweepEvery
	}
	stop := make(chan struct{})
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				closed := sp.GC()
				if closed > 0 {
					slog.Info("web session GC completed", "closed", closed)
				}
			case <-stop:
				return
			}
		}
	}()
	return func() { close(stop) }
}

// CreateSession creates a new agent page session for a Chromium profile.
func (sp *SessionPool) CreateSession(profileID string) (*WebSession, error) {
	if profileID == "" {
		return nil, fmt.Errorf("profile_id is required")
	}
	p, pool, err := sp.ensureProfilePool(profileID)
	if err != nil {
		return nil, err
	}
	if p.Engine != "chromium" {
		return nil, fmt.Errorf("profile %s is not a Chromium profile (web search requires Chrome/CloakBrowser)", profileID)
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()
	if len(pool.sessions) >= sp.maxPerProfile {
		return nil, fmt.Errorf("max sessions (%d) reached for profile %s", sp.maxPerProfile, profileID)
	}

	page, err := pool.browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("create agent page: %w", err)
	}

	now := time.Now()
	s := &WebSession{
		ID:           newWebSessionID(),
		ProfileID:    profileID,
		BrowserID:    pool.browserID,
		Page:         page,
		CreatedAt:    now,
		LastAccessed: now,
	}
	pool.sessions[s.ID] = s
	slog.Info("web session created", "session", s.ID, "profile", profileID, "browser", pool.browserID)
	return s, nil
}

// GetOrCreateSession returns an existing session by sessionID, or creates a new one for profileID.
func (sp *SessionPool) GetOrCreateSession(profileID, sessionID string) (*WebSession, bool, error) {
	if sessionID != "" {
		s, err := sp.GetSession(sessionID)
		if err != nil {
			return nil, false, err
		}
		if profileID != "" && profileID != s.ProfileID {
			return nil, false, fmt.Errorf("session %s belongs to profile %s, not %s", sessionID, s.ProfileID, profileID)
		}
		return s, false, nil
	}
	if profileID == "" {
		return nil, false, fmt.Errorf("profile_id is required when session_id is not provided")
	}
	s, err := sp.CreateSession(profileID)
	return s, true, err
}

// GetSession returns an existing agent session by ID.
func (sp *SessionPool) GetSession(sessionID string) (*WebSession, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	sp.mu.RLock()
	pools := make([]*ProfileSessionPool, 0, len(sp.pools))
	for _, pool := range sp.pools {
		pools = append(pools, pool)
	}
	sp.mu.RUnlock()

	for _, pool := range pools {
		pool.mu.RLock()
		s, ok := pool.sessions[sessionID]
		pool.mu.RUnlock()
		if ok {
			return s, nil
		}
	}
	return nil, fmt.Errorf("session not found: %s", sessionID)
}

// DestroySession closes an agent page and removes it from the pool.
func (sp *SessionPool) DestroySession(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	sp.mu.RLock()
	pools := make([]*ProfileSessionPool, 0, len(sp.pools))
	for _, pool := range sp.pools {
		pools = append(pools, pool)
	}
	sp.mu.RUnlock()

	for _, pool := range pools {
		pool.mu.Lock()
		s, ok := pool.sessions[sessionID]
		if ok {
			delete(pool.sessions, sessionID)
		}
		pool.mu.Unlock()
		if ok {
			s.closePage("destroy_session")
			slog.Info("web session destroyed", "session", sessionID, "profile", s.ProfileID, "browser", s.BrowserID, "reason", "destroy_session")
			return nil
		}
	}
	return fmt.Errorf("session not found: %s", sessionID)
}

// DestroyProfileSessions closes all agent pages for a profile and removes its profile pool.
// It intentionally does not close the browser.Manager-owned profile browser.
func (sp *SessionPool) DestroyProfileSessions(profileID string) int {
	if profileID == "" {
		return 0
	}

	sp.mu.Lock()
	pool, ok := sp.pools[profileID]
	if ok {
		delete(sp.pools, profileID)
	}
	sp.mu.Unlock()
	if !ok {
		return 0
	}

	closed := pool.closeAllPages("profile_sessions_destroyed")
	slog.Info("web profile sessions destroyed", "profile", profileID, "closed_sessions", closed)
	return closed
}

// ListSessions lists active agent sessions. If profileID is non-empty, only that profile is listed.
func (sp *SessionPool) ListSessions(profileID string) []SessionInfo {
	sp.mu.RLock()
	pools := make([]*ProfileSessionPool, 0, len(sp.pools))
	for id, pool := range sp.pools {
		if profileID == "" || profileID == id {
			pools = append(pools, pool)
		}
	}
	sp.mu.RUnlock()

	now := time.Now()
	infos := []SessionInfo{}
	for _, pool := range pools {
		pool.mu.RLock()
		for _, s := range pool.sessions {
			s.mu.Lock()
			info := SessionInfo{
				ID:           s.ID,
				ProfileID:    s.ProfileID,
				BrowserID:    s.BrowserID,
				CreatedAt:    s.CreatedAt,
				LastAccessed: s.LastAccessed,
				IdleSeconds:  int64(now.Sub(s.LastAccessed).Seconds()),
			}
			s.mu.Unlock()
			infos = append(infos, info)
		}
		pool.mu.RUnlock()
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].CreatedAt.Before(infos[j].CreatedAt) })
	return infos
}

// GC closes idle pages and enforces the per-profile session cap.
func (sp *SessionPool) GC() int {
	sp.mu.RLock()
	pools := make([]*ProfileSessionPool, 0, len(sp.pools))
	for _, pool := range sp.pools {
		pools = append(pools, pool)
	}
	sp.mu.RUnlock()

	closed := 0
	now := time.Now()
	for _, pool := range pools {
		type closeCandidate struct {
			session *WebSession
			reason  string
			idle    time.Duration
		}
		var toClose []closeCandidate
		pool.mu.Lock()
		for id, s := range pool.sessions {
			s.mu.Lock()
			idle := now.Sub(s.LastAccessed)
			s.mu.Unlock()
			if idle >= sp.idleTTL {
				delete(pool.sessions, id)
				toClose = append(toClose, closeCandidate{session: s, reason: "idle_ttl", idle: idle})
			}
		}
		if len(pool.sessions) > sp.maxPerProfile {
			remaining := make([]*WebSession, 0, len(pool.sessions))
			for _, s := range pool.sessions {
				remaining = append(remaining, s)
			}
			sort.Slice(remaining, func(i, j int) bool { return remaining[i].CreatedAt.Before(remaining[j].CreatedAt) })
			over := len(pool.sessions) - sp.maxPerProfile
			for i := 0; i < over; i++ {
				s := remaining[i]
				delete(pool.sessions, s.ID)
				toClose = append(toClose, closeCandidate{session: s, reason: "max_sessions"})
			}
		}
		pool.mu.Unlock()

		for _, c := range toClose {
			c.session.closePage("gc_" + c.reason)
			slog.Info("web session GC closed session", "session", c.session.ID, "profile", c.session.ProfileID, "browser", c.session.BrowserID, "reason", c.reason, "idle_seconds", int64(c.idle.Seconds()))
			closed++
		}
	}
	return closed
}

// CloseAll closes all agent pages and drops SessionPool metadata.
// It never closes browser.Manager browser sessions or connected browser handles.
// Playwright Go exposes Browser.Close() on Bind-connected handles, but Close may
// close the remote profile browser itself; until a safe non-closing disconnect
// API is available/verified, SessionPool only drops those non-owning handles.
func (sp *SessionPool) CloseAll() {
	sp.mu.Lock()
	pools := make([]*ProfileSessionPool, 0, len(sp.pools))
	for _, pool := range sp.pools {
		pools = append(pools, pool)
	}
	sp.pools = make(map[string]*ProfileSessionPool)
	sp.mu.Unlock()

	for _, pool := range pools {
		pool.closeAllPages("session_pool_close_all")
	}
	slog.Info("web session pool closed", "closed_sessions", len(pools))
}

func (pool *ProfileSessionPool) closeAllPages(reason string) int {
	pool.mu.Lock()
	sessions := make([]*WebSession, 0, len(pool.sessions))
	for _, s := range pool.sessions {
		sessions = append(sessions, s)
	}
	pool.sessions = make(map[string]*WebSession)
	pool.browser = nil
	pool.mu.Unlock()

	for _, s := range sessions {
		s.closePage(reason)
	}
	return len(sessions)
}

func (sp *SessionPool) ensureProfilePool(profileID string) (*profile.Profile, *ProfileSessionPool, error) {
	p, err := sp.store.Get(profileID)
	if err != nil {
		return nil, nil, err
	}
	if p.Engine != "chromium" {
		return nil, nil, fmt.Errorf("profile %s is not a Chromium profile (web search requires Chrome/CloakBrowser)", profileID)
	}

	browserSession, ok := sp.mgr.GetSession("sess_" + profileID)
	if !ok {
		browserSession, err = sp.mgr.LaunchSession(p)
		if err != nil {
			return nil, nil, fmt.Errorf("launch profile browser: %w", err)
		}
	}
	if browserSession.Engine != "chromium" {
		return nil, nil, fmt.Errorf("browser session %s is not chromium", browserSession.ID)
	}
	if browserSession.ConnectURL == "" {
		return nil, nil, fmt.Errorf("browser session %s has no Playwright bind endpoint", browserSession.ID)
	}

	sp.mu.RLock()
	pool, ok := sp.pools[profileID]
	if ok && pool.connectURL == browserSession.ConnectURL && pool.browser != nil {
		sp.mu.RUnlock()
		return p, pool, nil
	}
	sp.mu.RUnlock()

	sp.mu.Lock()
	pool, ok = sp.pools[profileID]
	if ok && pool.connectURL == browserSession.ConnectURL && pool.browser != nil {
		sp.mu.Unlock()
		return p, pool, nil
	}
	oldPool := pool
	// Do not call pool.browser.Close() here: with a Playwright Bind connection,
	// Close may close the remote profile browser owned by browser.Manager.
	// Playwright Go does not currently provide a verified non-closing disconnect
	// path for this handle in BrowseForge, so dropping the handle is safer;
	// GC/CloseAll only close agent pages.

	pw := sp.mgr.Playwright()
	if pw == nil {
		sp.mu.Unlock()
		return nil, nil, fmt.Errorf("playwright driver is not available")
	}
	connected, err := pw.Chromium.Connect(browserSession.ConnectURL)
	if err != nil {
		sp.mu.Unlock()
		return nil, nil, fmt.Errorf("connect to profile browser bind endpoint: %w", err)
	}
	pool = &ProfileSessionPool{
		profileID:  profileID,
		browserID:  browserSession.ID,
		connectURL: browserSession.ConnectURL,
		browser:    connected,
		sessions:   make(map[string]*WebSession),
	}
	sp.pools[profileID] = pool
	sp.mu.Unlock()
	if oldPool != nil {
		closed := oldPool.closeAllPages("profile_pool_replaced")
		if closed > 0 {
			slog.Info("old web profile pool closed", "profile", profileID, "closed_sessions", closed)
		}
	}
	slog.Info("web profile pool connected", "profile", profileID, "browser", browserSession.ID)
	return p, pool, nil
}

// WebSearch performs a provider-backed web search in this session's dedicated page.
func (s *WebSession) WebSearch(query string, engine string, maxResults int) (*SearchResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if maxResults <= 0 || maxResults > 30 {
		maxResults = 10
	}
	provider, err := getSearchProvider(engine)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePageOpenLocked(); err != nil {
		return nil, err
	}

	s.LastAccessed = time.Now()
	searchURL := provider.SearchURL(query)
	_, err = s.Page.Goto(searchURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		return nil, fmt.Errorf("navigate to search: %w", err)
	}

	if err = provider.WaitForResults(s.Page); err != nil {
		slog.Warn("search results selector not found, trying extraction anyway", "session", s.ID, "engine", provider.Name(), "error", err)
	}

	resp, err := provider.Extract(s.Page, maxResults)
	if err != nil {
		logSearchInterstitial(s.ID, s.ProfileID, provider, err)
		return nil, fmt.Errorf("extract search results: %w", err)
	}
	if resp == nil {
		resp = &SearchResponse{Engine: provider.Name(), ExtractionMode: "structured"}
	}
	if resp.Engine == "" {
		resp.Engine = provider.Name()
	}
	if resp.ExtractionMode == "" {
		resp.ExtractionMode = "structured"
	}
	if len(resp.Results) == 0 {
		fallback, err := s.extractSearchRawFallback(provider, maxResults)
		if err != nil {
			return nil, fmt.Errorf("extract raw search fallback: %w", err)
		}
		resp.ExtractionMode = "raw_fallback"
		resp.RawFallback = fallback
		slog.Info("web search raw fallback extracted", "session", s.ID, "profile", s.ProfileID, "engine", provider.Name(), "query", query, "text_chars", len(fallback.Text), "candidate_links", len(fallback.CandidateLinks))
	}

	rawLinks := 0
	if resp.RawFallback != nil {
		rawLinks = len(resp.RawFallback.CandidateLinks)
	}
	slog.Info("web search completed", "session", s.ID, "profile", s.ProfileID, "engine", provider.Name(), "query", query, "results", len(resp.Results), "mode", resp.ExtractionMode, "raw_links", rawLinks)
	return resp, nil
}

// WebExplore navigates this session's dedicated page and extracts structured content.
func (s *WebSession) WebExplore(url string, maxTextLength int, maxLinks int) (*ExploreResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if url == "" {
		return nil, fmt.Errorf("url is required")
	}
	if maxTextLength <= 0 || maxTextLength > 10000 {
		maxTextLength = 3000
	}
	if maxLinks <= 0 || maxLinks > 200 {
		maxLinks = 50
	}
	if err := s.ensurePageOpenLocked(); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	s.LastAccessed = time.Now()
	_, err := s.Page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		return nil, fmt.Errorf("navigate to url: %w", err)
	}

	result, err := s.extractPageContent(maxTextLength, maxLinks)
	if err != nil {
		return nil, fmt.Errorf("extract page content: %w", err)
	}

	result.URL = url
	slog.Info("web explore completed", "session", s.ID, "profile", s.ProfileID, "url", url, "textLen", len(result.Text), "links", len(result.Links))
	return result, nil
}

func (s *WebSession) closePage(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Closed = true
	if s.Page != nil {
		if err := s.Page.Close(); err != nil {
			slog.Warn("web session page close failed", "session", s.ID, "profile", s.ProfileID, "browser", s.BrowserID, "reason", reason, "error", err)
		} else {
			slog.Info("web session page closed", "session", s.ID, "profile", s.ProfileID, "browser", s.BrowserID, "reason", reason)
		}
		s.Page = nil
	}
}

func (s *WebSession) ensurePageOpenLocked() error {
	if s.Closed || s.Page == nil {
		return fmt.Errorf("session is closed: %s", s.ID)
	}
	return nil
}

// SearchResponse represents a web_search extraction result.
type SearchResponse struct {
	Engine         string             `json:"engine"`
	Results        []SearchResult     `json:"results"`
	ExtractionMode string             `json:"extraction_mode"`
	RawFallback    *SearchRawFallback `json:"raw_fallback,omitempty"`
}

// SearchResult represents a single structured search result entry.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// SearchRawFallback is an LLM-friendly SERP snapshot used when structured selectors fail.
type SearchRawFallback struct {
	PageTitle      string    `json:"page_title"`
	Text           string    `json:"text"`
	CandidateLinks []LinkRef `json:"candidate_links"`
}

// ExploreResult represents the extracted content of a single page.
type ExploreResult struct {
	URL         string    `json:"url,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Text        string    `json:"text"`
	Links       []LinkRef `json:"links"`
	Errors      []string  `json:"errors,omitempty"`
}

// LinkRef represents a hyperlink found on a page.
type LinkRef struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

func (s *WebSession) extractSearchRawFallback(provider SearchProvider, maxLinks int) (*SearchRawFallback, error) {
	data, err := s.Page.Evaluate(`() => {
		const normalizeSearchURL = (rawHref) => {
			if (!rawHref) return '';
			const u = new URL(rawHref, window.location.href);
			if (u.pathname === '/url' && u.searchParams.has('q')) {
				return u.searchParams.get('q') || '';
			}
			if (u.hostname.endsWith('bing.com') && u.searchParams.has('u')) {
				const encoded = u.searchParams.get('u') || '';
				try {
					const payload = encoded.startsWith('a1') ? encoded.slice(2) : encoded;
					const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
					if (decoded.startsWith('http://') || decoded.startsWith('https://')) return decoded;
				} catch (_) {}
			}
			if (u.pathname.includes('/l/') && u.searchParams.has('uddg')) {
				return decodeURIComponent(u.searchParams.get('uddg') || '');
			}
			return u.href;
		};
		const isUsefulHref = (href) => {
			if (!href || href.startsWith('javascript:') || href.startsWith('#') || href.startsWith('mailto:')) return false;
			const u = new URL(href, window.location.href);
			if (u.hostname.endsWith('google.com') && ['/search', '/preferences', '/advanced_search', '/maps', '/imghp', '/webhp'].includes(u.pathname)) return false;
			return true;
		};
		const body = document.body;
		let text = body && body.innerText ? body.innerText.replace(/[\s]+/g, ' ').trim() : '';
		const links = [];
		const seen = new Set();
		for (const a of document.querySelectorAll('a[href]')) {
			const href = normalizeSearchURL(a.getAttribute('href') || a.href);
			if (!isUsefulHref(href) || seen.has(href)) continue;
			const linkText = (a.innerText || a.textContent || '').replace(/[\s]+/g, ' ').trim();
			if (!linkText && !href) continue;
			seen.add(href);
			links.push({ text: linkText, url: href });
			if (links.length >= 30) break;
		}
		return { page_title: document.title || '', text, candidate_links: links };
	}`)
	if err != nil {
		return nil, err
	}
	raw, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected raw fallback type")
	}
	fallback := &SearchRawFallback{
		PageTitle:      asString(raw["page_title"]),
		Text:           truncateString(asString(raw["text"]), 4000),
		CandidateLinks: []LinkRef{},
	}
	if rawLinks, ok := raw["candidate_links"].([]any); ok {
		for i, item := range rawLinks {
			if i >= maxLinks {
				break
			}
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			linkURL := asString(m["url"])
			if !provider.UsefulFallbackLink(linkURL) {
				continue
			}
			fallback.CandidateLinks = append(fallback.CandidateLinks, LinkRef{Text: asString(m["text"]), URL: linkURL})
		}
	}
	return fallback, nil
}

func (s *WebSession) extractPageContent(maxTextLength, maxLinks int) (*ExploreResult, error) {
	data, err := s.Page.Evaluate(`() => {
		const result = { title: document.title || '', description: '', text: '', links: [] };
		const metaDesc = document.querySelector('meta[name="description"]');
		if (metaDesc) result.description = metaDesc.getAttribute('content') || '';
		const body = document.body;
		if (body) {
			const scripts = body.querySelectorAll('script, style, noscript, iframe');
			scripts.forEach(s => s.remove());
			let text = body.innerText || '';
			text = text.replace(/[\s]+/g, ' ').trim();
			result.text = text;
		}
		const links = document.querySelectorAll('a[href]');
		for (const a of links) {
			const href = a.getAttribute('href');
			if (!href || href.startsWith('javascript:') || href.startsWith('#') || href.startsWith('mailto:')) continue;
			const text = a.innerText ? a.innerText.trim() : '';
			if (!text && !href) continue;
			result.links.push({ text, url: href });
		}
		return result;
	}`)
	if err != nil {
		return nil, err
	}

	raw, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected page content type")
	}

	er := &ExploreResult{
		Title:       asString(raw["title"]),
		Description: asString(raw["description"]),
		Text:        asString(raw["text"]),
	}
	er.Text = truncateString(er.Text, maxTextLength)

	if rawLinks, ok := raw["links"].([]any); ok {
		for i, item := range rawLinks {
			if i >= maxLinks {
				break
			}
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			er.Links = append(er.Links, LinkRef{Text: asString(m["text"]), URL: asString(m["url"])})
		}
	}
	return er, nil
}

func newWebSessionID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("sess_search_%d", time.Now().UnixNano())
	}
	return "sess_search_" + hex.EncodeToString(b)
}

func truncateString(s string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	r := []rune(s)
	return string(r[:maxRunes]) + "\n… (truncated)"
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return fmt.Sprintf("%g", t)
	case nil:
		return ""
	default:
		s := fmt.Sprintf("%v", t)
		if utf8.RuneCountInString(s) > 5000 {
			return s[:5000] + "…"
		}
		return s
	}
}
