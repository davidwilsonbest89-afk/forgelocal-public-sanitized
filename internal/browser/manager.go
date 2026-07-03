package browser

import (
	"fmt"
	"log/slog"
	"sync"

	"browseforge/internal/config"
	"browseforge/internal/groups"
	"browseforge/internal/profile"

	"github.com/playwright-community/playwright-go"
)

// Session represents a running browser profile
type Session struct {
	ID         string
	ProfileID  string
	Engine     string
	Browser    playwright.Browser
	Context    playwright.BrowserContext
	Page       playwright.Page
	relay      *SOCKS5Relay
	ConnectURL string // Bind endpoint for external Playwright clients
}

// Manager handles browser instances (multi-instance: one process per profile)
type Manager struct {
	cfg        *config.Config
	groupStore *groups.Store
	pw         *playwright.Playwright
	mu         sync.RWMutex
	sessions   map[string]*Session // sessionID → Session
}

func NewManager(cfg *config.Config, groupStores ...*groups.Store) (*Manager, error) {
	if err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true}); err != nil {
		return nil, fmt.Errorf("playwright.Install: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright.Run: %w", err)
	}

	var groupStore *groups.Store
	if len(groupStores) > 0 {
		groupStore = groupStores[0]
	}
	m := &Manager{
		cfg:        cfg,
		groupStore: groupStore,
		pw:         pw,
		sessions:   make(map[string]*Session),
	}
	m.recoverOrphanSessions()
	return m, nil
}

// Playwright returns the underlying Playwright driver instance.
// Used by integration components that need to reuse the same Playwright process.
func (m *Manager) Playwright() *playwright.Playwright {
	return m.pw
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

	session, err := m.launchProfile(p)

	if err != nil {
		if shouldRetryLaunch(err) {
			slog.Warn("browser launch failed with recoverable protocol error; restarting Playwright and retrying", "profile", p.ID, "engine", p.Engine, "error", err)
			if len(m.sessions) > 0 {
				m.dropSessionsLocked("playwright driver restart after protocol error")
			}
			if restartErr := m.restartPlaywright(); restartErr != nil {
				return nil, fmt.Errorf("%w; playwright restart failed: %v", err, restartErr)
			}
			session, err = m.launchProfile(p)
			if err == nil {
				slog.Info("browser launch recovered after Playwright restart", "profile", p.ID, "engine", p.Engine)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	// Bind browser for external Playwright clients (WebSocket mode)
	if session.Context != nil {
		if browser := session.Context.Browser(); browser != nil {
			if result, err := browser.Bind("browseforge-"+session.ID, playwright.BrowserBindOptions{
				Host: playwright.String(m.cfg.Host),
				Port: playwright.Int(0),
			}); err == nil {
				session.ConnectURL = result.Endpoint
			} else {
				slog.Warn("browser bind failed", "session", session.ID, "error", err)
			}
		}
	}

	m.sessions[session.ID] = session
	slog.Info("session launched", "session", session.ID, "profile", p.ID, "engine", p.Engine, "connectURL", session.ConnectURL)
	return session, nil
}

func (m *Manager) launchProfile(p *profile.Profile) (*Session, error) {
	switch p.Engine {
	case "chromium":
		return m.launchChromium(p)
	default:
		return m.launchFirefox(p)
	}
}

func (m *Manager) effectiveProxy(p *profile.Profile) groups.EffectiveProxy {
	if m.groupStore != nil {
		return m.groupStore.EffectiveProxy(p)
	}
	if p != nil && p.Proxy != nil && p.Proxy.Host != "" {
		return groups.EffectiveProxy{Proxy: p.Proxy, Source: "profile", Mode: groups.ProxyModeDefault}
	}
	return groups.EffectiveProxy{Source: "none", Mode: groups.ProxyModeDefault}
}

func (m *Manager) restartPlaywright() error {
	if m.pw != nil {
		if err := m.pw.Stop(); err != nil {
			slog.Warn("playwright stop during restart failed", "error", err)
		}
		m.pw = nil
	}
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("playwright.Run: %w", err)
	}
	m.pw = pw
	return nil
}

func (m *Manager) dropSessionsLocked(reason string) {
	for id, s := range m.sessions {
		if s.relay != nil {
			s.relay.Close()
		}
		delete(m.sessions, id)
		slog.Warn("session dropped", "session", id, "reason", reason)
	}
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
	if s.relay != nil {
		s.relay.Close()
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
		if s.relay != nil {
			s.relay.Close()
		}
		delete(m.sessions, id)
	}
	if m.pw != nil {
		m.pw.Stop()
	}
}
