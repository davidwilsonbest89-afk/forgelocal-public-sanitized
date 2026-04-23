package browser

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"browseforge/internal/config"
	"browseforge/internal/profile"

	"github.com/playwright-community/playwright-go"
)

// Session represents a running browser profile
type Session struct {
	ID        string
	ProfileID string
	Engine    string
	Browser   playwright.Browser
	Context   playwright.BrowserContext
	Page      playwright.Page
}

// Manager handles browser instances (multi-instance: one process per profile)
type Manager struct {
	cfg      *config.Config
	pw       *playwright.Playwright
	mu       sync.RWMutex
	sessions map[string]*Session // sessionID → Session
}

func NewManager(cfg *config.Config) (*Manager, error) {
	playwright.Install(&playwright.RunOptions{Browsers: []string{}})

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright.Run: %w", err)
	}

	m := &Manager{
		cfg:      cfg,
		pw:       pw,
		sessions: make(map[string]*Session),
	}
	m.recoverOrphanSessions()
	return m, nil
}

// recoverOrphanSessions cleans up stale session state on startup
func (m *Manager) recoverOrphanSessions() {
	// On startup, all previous sessions are dead (processes gone)
	// Just ensure clean state
	slog.Info("session recovery: starting clean")
}

func (m *Manager) LaunchSession(p *profile.Profile) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	for _, s := range m.sessions {
		if s.ProfileID == p.ID {
			return s, nil
		}
	}

	var session *Session
	var err error

	switch p.Engine {
	case "chromium":
		session, err = m.launchChromium(p)
	default:
		session, err = m.launchFirefox(p)
	}

	if err != nil {
		return nil, err
	}

	m.sessions[session.ID] = session
	slog.Info("session launched", "session", session.ID, "profile", p.ID, "engine", p.Engine)
	return session, nil
}

func (m *Manager) launchFirefox(p *profile.Profile) (*Session, error) {
	camoufoxPath := m.cfg.CamoufoxPath
	if _, err := os.Stat(camoufoxPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("camoufox not found: %s", camoufoxPath)
	}

	// Build CAMOU_CONFIG from profile fingerprint
	configJSON, _ := json.Marshal(p.Fingerprint)

	userDataDir, _ := filepath.Abs(filepath.Join(p.ProfileDir, "browser-data"))
	os.MkdirAll(userDataDir, 0755)

	// Camoufox path must also be absolute
	absPath, _ := filepath.Abs(camoufoxPath)

	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		ExecutablePath: playwright.String(absPath),
		Headless:       playwright.Bool(false),
		Env: map[string]string{
			"CAMOU_CONFIG": string(configJSON),
		},
		FirefoxUserPrefs: map[string]any{
			"media.peerconnection.enabled":  false,
			"xpinstall.signatures.required": false,
		},
		Viewport: &playwright.Size{Width: 1280, Height: 800},
	}

	// Per-profile proxy
	if p.Proxy != nil {
		server := fmt.Sprintf("%s://%s:%d", p.Proxy.Type, p.Proxy.Host, p.Proxy.Port)
		opts.Proxy = &playwright.Proxy{
			Server:   server,
			Username: playwright.String(p.Proxy.Username),
			Password: playwright.String(p.Proxy.Password),
		}
	}

	ctx, err := m.pw.Firefox.LaunchPersistentContext(userDataDir, opts)
	if err != nil {
		return nil, fmt.Errorf("launch firefox: %w", err)
	}

	// Get or create first page
	pages := ctx.Pages()
	var page playwright.Page
	if len(pages) > 0 {
		page = pages[0]
	} else {
		page, err = ctx.NewPage()
		if err != nil {
			ctx.Close()
			return nil, fmt.Errorf("new page: %w", err)
		}
	}

	return &Session{
		ID:        fmt.Sprintf("sess_%s", p.ID),
		ProfileID: p.ID,
		Engine:    "firefox",
		Context:   ctx,
		Page:      page,
	}, nil
}

func (m *Manager) launchChromium(p *profile.Profile) (*Session, error) {
	chromiumPath := m.cfg.CloakBrowserPath
	if chromiumPath == "" {
		return nil, fmt.Errorf("cloakbrowser_path not configured")
	}

	userDataDir, _ := filepath.Abs(filepath.Join(p.ProfileDir, "browser-data"))
	os.MkdirAll(userDataDir, 0755)

	absChromiumPath, _ := filepath.Abs(chromiumPath)

	args := []string{
		"--no-first-run",
	}
	if p.FingerprintSeed > 0 {
		args = append(args, fmt.Sprintf("--fingerprint=%d", p.FingerprintSeed))
	}

	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		ExecutablePath: playwright.String(absChromiumPath),
		Headless:       playwright.Bool(false),
		Args:           args,
		Viewport:       &playwright.Size{Width: 1280, Height: 800},
		IgnoreDefaultArgs: []string{"--enable-automation", "--no-sandbox", "--disable-blink-features=AutomationControlled"},
	}

	if p.Proxy != nil {
		server := fmt.Sprintf("%s://%s:%d", p.Proxy.Type, p.Proxy.Host, p.Proxy.Port)
		opts.Proxy = &playwright.Proxy{
			Server:   server,
			Username: playwright.String(p.Proxy.Username),
			Password: playwright.String(p.Proxy.Password),
		}
	}

	ctx, err := m.pw.Chromium.LaunchPersistentContext(userDataDir, opts)
	if err != nil {
		return nil, fmt.Errorf("launch chromium: %w", err)
	}

	pages := ctx.Pages()
	var page playwright.Page
	if len(pages) > 0 {
		page = pages[0]
	} else {
		page, _ = ctx.NewPage()
	}

	return &Session{
		ID:        fmt.Sprintf("sess_%s", p.ID),
		ProfileID: p.ID,
		Engine:    "chromium",
		Context:   ctx,
		Page:      page,
	}, nil
}

func (m *Manager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

func (m *Manager) ListSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Session
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}

func (m *Manager) CloseSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	if s.Context != nil {
		s.Context.Close()
	}
	if s.Browser != nil {
		s.Browser.Close()
	}
	delete(m.sessions, id)
	slog.Info("session closed", "session", id)
	return nil
}

func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, s := range m.sessions {
		if s.Context != nil {
			s.Context.Close()
		}
		delete(m.sessions, id)
	}
	if m.pw != nil {
		m.pw.Stop()
	}
}
