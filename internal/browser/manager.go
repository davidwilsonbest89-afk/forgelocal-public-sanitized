package browser

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"browseforge/internal/config"
	"browseforge/internal/fingerprint"
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
	cfg      *config.Config
	pw       *playwright.Playwright
	mu       sync.RWMutex
	sessions map[string]*Session // sessionID → Session
}

func NewManager(cfg *config.Config) (*Manager, error) {
	if err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true}); err != nil {
		return nil, fmt.Errorf("playwright.Install: %w", err)
	}

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

func (m *Manager) launchFirefox(p *profile.Profile) (*Session, error) {
	camoufoxPath := m.cfg.CamoufoxPath
	if _, err := os.Stat(camoufoxPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("camoufox not found: %s", camoufoxPath)
	}

	// Build CAMOU_CONFIG: start from profile fingerprint, then overlay GeoIP
	// WebGL strategy: only pass WebGL fields if we have a COMPLETE profile
	// (with supportedExtensions + parameters). Incomplete WebGL data (only
	// renderer/vendor) causes inconsistency detection. When omitted, Camoufox
	// auto-generates a fully consistent WebGL fingerprint via BrowserForge.
	config := make(map[string]any)
	hasFullWebGL := false
	for k, v := range p.Fingerprint {
		config[k] = v
		if k == "webGl:supportedExtensions" {
			hasFullWebGL = true
		}
	}
	if !hasFullWebGL {
		// Remove partial WebGL data — let Camoufox handle it
		delete(config, "webGl:renderer")
		delete(config, "webGl:vendor")
	}

	// GeoIP: detect timezone/locale from proxy or local IP, inject into CAMOU_CONFIG
	if p.Proxy != nil && p.Proxy.Host != "" {
		tz, locale := fingerprint.DetectProxyGeoResult(p.Proxy.Type, p.Proxy.Host, p.Proxy.Port, p.Proxy.Username, p.Proxy.Password)
		config["timezone"] = tz
		parts := splitLocale(locale)
		config["locale:language"] = parts[0]
		config["locale:region"] = parts[1]
	} else {
		tz, locale := fingerprint.DetectLocalGeoResult()
		config["timezone"] = tz
		parts := splitLocale(locale)
		config["locale:language"] = parts[0]
		config["locale:region"] = parts[1]
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("encode camoufox config: %w", err)
	}

	userDataDir, err := filepath.Abs(filepath.Join(p.ProfileDir, "browser-data"))
	if err != nil {
		return nil, fmt.Errorf("browser data path: %w", err)
	}
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return nil, fmt.Errorf("create browser data dir: %w", err)
	}
	cleanProfileLocks(userDataDir)

	absPath, err := filepath.Abs(camoufoxPath)
	if err != nil {
		return nil, fmt.Errorf("camoufox path: %w", err)
	}

	downloadsDir, err := filepath.Abs(filepath.Join(p.ProfileDir, "downloads"))
	if err != nil {
		return nil, fmt.Errorf("downloads path: %w", err)
	}
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return nil, fmt.Errorf("create downloads dir: %w", err)
	}

	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		ExecutablePath:  playwright.String(absPath),
		Headless:        playwright.Bool(false),
		AcceptDownloads: playwright.Bool(true),
		Env: map[string]string{
			"CAMOU_CONFIG":          string(configJSON),
			"DISPLAY":               os.Getenv("DISPLAY"),
			"HOME":                  os.Getenv("HOME"),
			"LIBGL_ALWAYS_SOFTWARE": os.Getenv("LIBGL_ALWAYS_SOFTWARE"),
		},
		FirefoxUserPrefs: map[string]any{
			"xpinstall.signatures.required":          false,
			"browser.download.folderList":            2,
			"browser.download.dir":                   downloadsDir,
			"browser.download.useDownloadDir":        true,
			"browser.helperApps.neverAsk.saveToDisk": "application/octet-stream,image/jpeg,image/png,application/pdf,application/zip",
			"webgl.disabled":                         false,
			"webgl.force-enabled":                    true,
			"webgl.forbid-software":                  false,
		},
		Viewport: &playwright.Size{Width: 1280, Height: 800},
	}

	// Proxy: SOCKS5 with auth needs relay (Playwright protocol rejects it)
	var relay *SOCKS5Relay

	if p.Proxy != nil {
		needsRelay := p.Proxy.Type == "socks5" && p.Proxy.Username != ""
		if needsRelay {
			upstream := fmt.Sprintf("%s:%d", p.Proxy.Host, p.Proxy.Port)
			var localAddr string
			var err error
			relay, localAddr, err = StartSOCKS5Relay(upstream, p.Proxy.Username, p.Proxy.Password)
			if err != nil {
				return nil, fmt.Errorf("socks5 relay: %w", err)
			}
			opts.Proxy = &playwright.Proxy{Server: "socks5://" + localAddr}
		} else {
			server := fmt.Sprintf("%s://%s:%d", p.Proxy.Type, p.Proxy.Host, p.Proxy.Port)
			opts.Proxy = &playwright.Proxy{
				Server:   server,
				Username: playwright.String(p.Proxy.Username),
				Password: playwright.String(p.Proxy.Password),
			}
		}
	}

	ctx, err := m.pw.Firefox.LaunchPersistentContext(userDataDir, opts)
	if err != nil {
		if relay != nil {
			relay.Close()
		}
		return nil, fmt.Errorf("launch firefox: %w", humanizeError(err))
	}

	dlDir := downloadsDir
	onDl := func(d playwright.Download) { go d.SaveAs(filepath.Join(dlDir, d.SuggestedFilename())) }
	for _, pg := range ctx.Pages() {
		pg.OnDownload(onDl)
	}
	ctx.OnPage(func(pg playwright.Page) { pg.OnDownload(onDl) })

	pages := ctx.Pages()
	var page playwright.Page
	if len(pages) > 0 {
		page = pages[0]
	} else {
		page, err = ctx.NewPage()
		if err != nil {
			if relay != nil {
				relay.Close()
			}
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
		relay:     relay,
	}, nil
}

// splitLocale splits "en-US" into ["en", "US"], fallback to ["en", "US"]
func splitLocale(locale string) [2]string {
	if i := strings.IndexByte(locale, '-'); i > 0 {
		return [2]string{locale[:i], locale[i+1:]}
	}
	return [2]string{"en", "US"}
}

func (m *Manager) launchChromium(p *profile.Profile) (*Session, error) {
	chromiumPath := m.cfg.CloakBrowserPath
	if chromiumPath == "" {
		return nil, fmt.Errorf("cloakbrowser_path not configured")
	}

	userDataDir, err := filepath.Abs(filepath.Join(p.ProfileDir, "browser-data"))
	if err != nil {
		return nil, fmt.Errorf("browser data path: %w", err)
	}
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return nil, fmt.Errorf("create browser data dir: %w", err)
	}
	cleanProfileLocks(userDataDir)

	absChromiumPath, err := filepath.Abs(chromiumPath)
	if err != nil {
		return nil, fmt.Errorf("cloakbrowser path: %w", err)
	}

	// CloakBrowser native flags — fingerprint at C++ level
	args := []string{
		"--no-first-run",
		"--test-type",
	}
	if p.FingerprintSeed > 0 {
		args = append(args, fmt.Sprintf("--fingerprint=%d", p.FingerprintSeed))
	}

	// GeoIP: detect timezone/locale from proxy or local IP, pass as native flags
	var tz, locale string
	if p.Proxy != nil && p.Proxy.Host != "" {
		tz, locale = fingerprint.DetectProxyGeoResult(p.Proxy.Type, p.Proxy.Host, p.Proxy.Port, p.Proxy.Username, p.Proxy.Password)
		args = append(args, "--fingerprint-webrtc-ip=auto")
	} else {
		tz, locale = fingerprint.DetectLocalGeoResult()
	}
	args = append(args,
		"--fingerprint-timezone="+tz,
		"--fingerprint-locale="+locale,
	)

	// Platform spoofing: Linux → appear as Windows (more common fingerprint)
	if runtime.GOOS == "linux" {
		args = append(args, "--fingerprint-platform=windows")
	}

	// Pass fingerprint pool values as native flags (override seed defaults)
	if p.Fingerprint != nil {
		if v, ok := p.Fingerprint["navigator.hardwareConcurrency"]; ok {
			args = append(args, fmt.Sprintf("--fingerprint-hardware-concurrency=%v", v))
		}
		if v, ok := p.Fingerprint["screen.width"]; ok {
			args = append(args, fmt.Sprintf("--fingerprint-screen-width=%v", v))
		}
		if v, ok := p.Fingerprint["screen.height"]; ok {
			args = append(args, fmt.Sprintf("--fingerprint-screen-height=%v", v))
		}
	}

	// Docker/container mode: disable sandbox
	if m.cfg.NoSandbox {
		args = append(args, "--no-sandbox")
	}

	// Docker: use system fonts directory for font fingerprinting
	if _, err := os.Stat("/usr/share/fonts"); err == nil {
		args = append(args, "--fingerprint-fonts-dir=/usr/share/fonts")
	}

	downloadsDir, err := filepath.Abs(filepath.Join(p.ProfileDir, "downloads"))
	if err != nil {
		return nil, fmt.Errorf("downloads path: %w", err)
	}
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return nil, fmt.Errorf("create downloads dir: %w", err)
	}

	// Set Chromium download directory via Preferences (before launch)
	prefsDir := filepath.Join(userDataDir, "Default")
	if err := os.MkdirAll(prefsDir, 0755); err != nil {
		return nil, fmt.Errorf("create chromium prefs dir: %w", err)
	}
	prefsPath := filepath.Join(prefsDir, "Preferences")
	prefs := map[string]any{}
	if data, err := os.ReadFile(prefsPath); err == nil {
		if err := json.Unmarshal(data, &prefs); err != nil {
			return nil, fmt.Errorf("decode chromium preferences: %w", err)
		}
	}
	prefs["savefile"] = map[string]any{"default_directory": downloadsDir}
	prefs["download"] = map[string]any{"default_directory": downloadsDir, "prompt_for_download": false}
	out, err := json.Marshal(prefs)
	if err != nil {
		return nil, fmt.Errorf("encode chromium preferences: %w", err)
	}
	if err := os.WriteFile(prefsPath, out, 0644); err != nil {
		return nil, fmt.Errorf("write chromium preferences: %w", err)
	}

	ignoreArgs := []string{
		"--enable-automation",
		"--disable-blink-features=AutomationControlled",
		"--host-resolver-rules=MAP * ~NOTFOUND , EXCLUDE 127.0.0.1", // CloakBrowser handles DNS internally
	}
	if !m.cfg.NoSandbox {
		ignoreArgs = append(ignoreArgs, "--no-sandbox")
	}

	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		ExecutablePath:    playwright.String(absChromiumPath),
		Headless:          playwright.Bool(false),
		AcceptDownloads:   playwright.Bool(true),
		Args:              args,
		Viewport:          &playwright.Size{Width: 1280, Height: 800},
		IgnoreDefaultArgs: ignoreArgs,
	}

	// Proxy setup — SOCKS5 with auth needs local relay (Playwright protocol rejects it)
	var relay *SOCKS5Relay

	if p.Proxy != nil {
		needsRelay := p.Proxy.Type == "socks5" && p.Proxy.Username != ""
		if needsRelay {
			upstream := fmt.Sprintf("%s:%d", p.Proxy.Host, p.Proxy.Port)
			var localAddr string
			var err error
			relay, localAddr, err = StartSOCKS5Relay(upstream, p.Proxy.Username, p.Proxy.Password)
			if err != nil {
				return nil, fmt.Errorf("socks5 relay: %w", err)
			}
			opts.Proxy = &playwright.Proxy{Server: "socks5://" + localAddr}
		} else {
			server := fmt.Sprintf("%s://%s:%d", p.Proxy.Type, p.Proxy.Host, p.Proxy.Port)
			opts.Proxy = &playwright.Proxy{
				Server:   server,
				Username: playwright.String(p.Proxy.Username),
				Password: playwright.String(p.Proxy.Password),
			}
		}
	}

	ctx, err := m.pw.Chromium.LaunchPersistentContext(userDataDir, opts)
	if err != nil {
		if relay != nil {
			relay.Close()
		}
		return nil, fmt.Errorf("launch chromium: %w", humanizeError(err))
	}

	dlDir := downloadsDir
	onDl := func(d playwright.Download) { go d.SaveAs(filepath.Join(dlDir, d.SuggestedFilename())) }
	for _, pg := range ctx.Pages() {
		pg.OnDownload(onDl)
	}
	ctx.OnPage(func(pg playwright.Page) { pg.OnDownload(onDl) })

	pages := ctx.Pages()
	var page playwright.Page
	if len(pages) > 0 {
		page = pages[0]
	} else {
		page, err = ctx.NewPage()
		if err != nil {
			if relay != nil {
				relay.Close()
			}
			ctx.Close()
			return nil, fmt.Errorf("new page: %w", err)
		}
	}

	return &Session{
		ID:        fmt.Sprintf("sess_%s", p.ID),
		ProfileID: p.ID,
		Engine:    "chromium",
		Context:   ctx,
		Page:      page,
		relay:     relay,
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

// humanizeError wraps Playwright errors into user-friendly messages
func humanizeError(err error) error {
	msg := err.Error()
	switch {
	case shouldRetryLaunch(err):
		return fmt.Errorf("瀏覽器啟動時 Playwright protocol 連線中斷。BrowseForge 會自動重試一次；若仍失敗，請重啟服務或容器。原始錯誤: %w", err)
	case strings.Contains(msg, "sandboxing failed") || strings.Contains(msg, "sandbox"):
		return fmt.Errorf("Chromium sandbox 失敗。Docker 中請使用 --no-sandbox 或 'serve --no-sandbox'。原始錯誤: %w", err)
	case strings.Contains(msg, "XServer") || strings.Contains(msg, "DISPLAY"):
		return fmt.Errorf("找不到 X 顯示器。請設定 DISPLAY 環境變數或使用 xvfb-run。原始錯誤: %w", err)
	case strings.Contains(msg, "profile appears to be in use"):
		return fmt.Errorf("Profile 被鎖定（上次未正常關閉）。請重啟服務或刪除 profiles/*/browser-data/SingletonLock。原始錯誤: %w", err)
	case strings.Contains(msg, "executable doesn't exist") || strings.Contains(msg, "not found"):
		return fmt.Errorf("瀏覽器執行檔不存在。請重新啟動讓 BrowseForge 自動下載。原始錯誤: %w", err)
	default:
		return err
	}
}

func shouldRetryLaunch(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "could not read protocol padding") ||
		strings.Contains(msg, "target closed") ||
		strings.Contains(msg, "EOF")
}

func cleanProfileLocks(userDataDir string) {
	for _, name := range []string{"SingletonLock", "SingletonCookie", "SingletonSocket"} {
		path := filepath.Join(userDataDir, name)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			slog.Warn("remove stale profile lock failed", "path", path, "error", err)
		}
	}
}
