package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"browseforge/internal/api"
	"browseforge/internal/browser"
	"browseforge/internal/config"
	"browseforge/internal/fingerprint"
	"browseforge/internal/humanize"
	"browseforge/internal/mcp"
	"browseforge/internal/profile"
	"browseforge/internal/workflow"
)

var Version = "dev"

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

// serveFlags holds CLI overrides for serve mode
type serveFlags struct {
	baseDir    string
	configPath string
	host       string
	port       string
	noSandbox  bool
	noOpen     bool
}

func isDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func runMCPStdio(opts mcpStdioOptions) {
	baseDir := opts.baseDir
	if err := os.Chdir(baseDir); err != nil {
		fmt.Fprintf(os.Stderr, "Chdir error: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll("profiles", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Create profiles dir error: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Create data dir error: %v\n", err)
		os.Exit(1)
	}

	// Reuse existing config (don't download browsers in MCP mode)
	if _, err := os.Stat(opts.configPath); os.IsNotExist(err) {
		if err := writeDefaultConfig(opts.configPath, baseDir, false); err != nil {
			fmt.Fprintf(os.Stderr, "Config init error: %v\n", err)
			os.Exit(1)
		}
	}

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}
	cfg.Version = Version
	profileStore, err := profile.NewStore(cfg.ProfilesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Profile store error: %v\n", err)
		os.Exit(1)
	}
	browserMgr, err := browser.NewManager(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Browser manager error: %v\n", err)
		os.Exit(1)
	}
	defer browserMgr.Close()

	sessionPool, err := mcp.NewSessionPool(browserMgr, profileStore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Web session pool warning: %v\n", err)
	}
	if sessionPool != nil {
		stopGC := sessionPool.StartGC(sessionPool.SweepInterval())
		defer stopGC()
		defer sessionPool.CloseAll()
	}

	mcpServer := mcp.NewServer(profileStore, browserMgr, buildHumanizeCfg(cfg), sessionPool, "", cfg.Version)
	mcpServer.RunStdio()
}

func runServer(flags *serveFlags) {
	if flags == nil {
		baseDir, err := defaultBaseDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Executable path error: %v\n", err)
			os.Exit(1)
		}
		flags = &serveFlags{baseDir: baseDir, configPath: filepath.Join(baseDir, "config.json")}
	}
	// Auto-detect base directory (where the binary lives)
	baseDir := flags.baseDir
	if err := os.Chdir(baseDir); err != nil {
		fmt.Fprintf(os.Stderr, "Chdir error: %v\n", err)
		os.Exit(1)
	}

	// Auto-create directories
	if err := os.MkdirAll("profiles", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Create profiles dir error: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Create data dir error: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll("logs", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Create logs dir error: %v\n", err)
		os.Exit(1)
	}

	// Load config once
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}
	cfg.Version = Version

	// Docker auto-detection
	if isDocker() {
		if cfg.Host == "127.0.0.1" {
			cfg.Host = "0.0.0.0"
		}
		if !cfg.NoSandbox {
			cfg.NoSandbox = true
		}
	}

	// CLI flags override config
	if flags != nil {
		if flags.host != "" {
			cfg.Host = flags.host
		}
		if flags.port != "" {
			cfg.Port = flags.port
		}
		if flags.noSandbox {
			cfg.NoSandbox = true
		}
	}

	// Setup logger (stdout in Docker, file otherwise)
	if isDocker() {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	} else {
		logger := config.SetupLogger(cfg.LogFile)
		slog.SetDefault(logger)
	}

	// Find or download browsers
	camoufoxPath := ""
	if browser.InstalledVersion(baseDir, "camoufox") == browser.CamoufoxVersion {
		camoufoxPath = browser.FindBinary(baseDir, "camoufox")
	}
	if camoufoxPath == "" {
		if browser.InstalledVersion(baseDir, "camoufox") == "" {
			fmt.Println("🦊 Camoufox not found. Downloading...")
		} else {
			fmt.Printf("🦊 Camoufox update available (%s → %s). Downloading...\n",
				browser.InstalledVersion(baseDir, "camoufox"), browser.CamoufoxVersion)
		}
		var err error
		camoufoxPath, err = browser.DownloadCamoufox(baseDir)
		if err != nil {
			slog.Error("download Camoufox", "error", err)
			os.Exit(1)
		}
	}

	chromiumPath := ""
	expectedCloak := browser.ExpectedCloakBrowserVersion()
	if browser.InstalledVersion(baseDir, "cloakbrowser") == expectedCloak {
		chromiumPath = browser.FindBinary(baseDir, "cloakbrowser")
	}
	if chromiumPath == "" {
		if browser.InstalledVersion(baseDir, "cloakbrowser") == "" {
			fmt.Println("🌐 CloakBrowser not found. Downloading...")
		} else {
			fmt.Printf("🌐 CloakBrowser update available (%s → %s). Downloading...\n",
				browser.InstalledVersion(baseDir, "cloakbrowser"), expectedCloak)
		}
		var err error
		chromiumPath, err = browser.DownloadCloakBrowser(baseDir)
		if err != nil {
			slog.Error("download CloakBrowser", "error", err)
			os.Exit(1)
		}
	}

	// Update config with browser paths if changed
	if camoufoxPath != cfg.CamoufoxPath || chromiumPath != cfg.CloakBrowserPath {
		cfg.CamoufoxPath = camoufoxPath
		cfg.CloakBrowserPath = chromiumPath
		cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			slog.Error("config encode", "error", err)
			os.Exit(1)
		}
		if err := os.WriteFile(flags.configPath, cfgJSON, 0644); err != nil {
			slog.Error("config write", "error", err)
			os.Exit(1)
		}
	}

	profileStore, err := profile.NewStore(cfg.ProfilesDir)
	if err != nil {
		slog.Error("profile store", "error", err)
		os.Exit(1)
	}

	fpPool, _ := fingerprint.NewPool(cfg.FingerprintDir)

	browserMgr, err := browser.NewManager(cfg)
	if err != nil {
		slog.Error("browser manager", "error", err)
		os.Exit(1)
	}
	defer browserMgr.Close()

	router, err := api.NewRouter(cfg, profileStore, browserMgr, fpPool)
	if err != nil {
		slog.Error("api router", "error", err)
		os.Exit(1)
	}

	// Web Search & Explore agent sessions (profile browser + Playwright Bind + independent pages)
	sessionPool, err := mcp.NewSessionPool(browserMgr, profileStore)
	if err != nil {
		slog.Warn("web session pool not available", "error", err)
	} else {
		stopGC := sessionPool.StartGC(sessionPool.SweepInterval())
		defer stopGC()
		defer sessionPool.CloseAll()
	}

	// Workflow engine
	wfEngine := workflow.NewEngine("http://127.0.0.1:"+cfg.Port, cfg.APIToken)
	router.Post("/api/workflow/run", api.WorkflowHandler(wfEngine))

	// MCP Server (same HTTP service port, integrated with the main router)
	mcpServer := mcp.NewServer(profileStore, browserMgr, buildHumanizeCfg(cfg), sessionPool, cfg.APIToken, cfg.Version)
	mcpServer.SetWorkflowEngine(wfEngine)
	router.Post("/mcp", mcpServer.ServeHTTP)

	// HTTP Server with error channel (no os.Exit in goroutine)
	srv := &http.Server{Addr: cfg.Host + ":" + cfg.Port, Handler: router}
	serverErr := make(chan error, 1)

	go func() {
		slog.Info("server starting", "host", cfg.Host, "port", cfg.Port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for server ready, then show info
	go func() {
		for i := 0; i < 30; i++ {
			resp, err := http.Get("http://127.0.0.1:" + cfg.Port + "/api/status")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				token := cfg.APIToken
				fmt.Println("╔══════════════════════════════════════════╗")
				fmt.Printf("║        🦊 BrowseForge v%-16s║\n", Version)
				fmt.Println("╠══════════════════════════════════════════╣")
				fmt.Printf("║  Dashboard: http://%s:%-12s║\n", cfg.Host, cfg.Port)
				fmt.Printf("║  MCP:       http://%s:%-12s║\n", cfg.Host, cfg.Port+"/mcp")
				fmt.Printf("║  Token:     %s...  ║\n", tokenPreview(token))
				fmt.Println("╚══════════════════════════════════════════╝")
				if cfg.Host == "127.0.0.1" && !flags.noOpen {
					openBrowser(fmt.Sprintf("http://127.0.0.1:%s#%s", cfg.Port, token))
				}
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Wait for shutdown signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		fmt.Println("\nShutting down...")
	case err := <-serverErr:
		fmt.Fprintf(os.Stderr, "❌ Server failed: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func tokenPreview(token string) string {
	if len(token) <= 16 {
		return token
	}
	return token[:16]
}

// openBrowser opens URL in the default system browser
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func buildHumanizeCfg(cfg *config.Config) humanize.Config {
	if cfg.Humanize == nil {
		return humanize.DefaultConfig()
	}
	return humanize.ConfigFromRaw(cfg.Humanize.Enabled, cfg.Humanize.MouseSpeed, cfg.Humanize.TypingCPM, cfg.Humanize.TypoRate, cfg.Humanize.ScrollStyle)
}
