package browser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"browseforge/internal/fingerprint"
	"browseforge/internal/profile"

	"github.com/playwright-community/playwright-go"
)

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

	args := []string{
		"--no-first-run",
		"--test-type",
	}
	if p.FingerprintSeed > 0 {
		args = append(args, fmt.Sprintf("--fingerprint=%d", p.FingerprintSeed))
	}

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

	if runtime.GOOS == "linux" {
		args = append(args, "--fingerprint-platform=windows")
	}

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

	if m.cfg.NoSandbox {
		args = append(args, "--no-sandbox")
	}

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
		"--host-resolver-rules=MAP * ~NOTFOUND , EXCLUDE 127.0.0.1",
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

	var relay *SOCKS5Relay
	if p.Proxy != nil {
		needsRelay := p.Proxy.Type == "socks5" && p.Proxy.Username != ""
		if needsRelay {
			upstream := fmt.Sprintf("%s:%d", p.Proxy.Host, p.Proxy.Port)
			var localAddr string
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
