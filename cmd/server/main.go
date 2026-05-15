package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
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
	if len(os.Args) < 2 {
		runServer(nil)
		return
	}

	switch os.Args[1] {
	case "--mcp":
		runMCPStdio()
	case "serve":
		flags := parseServeFlags(os.Args[2:])
		runServer(flags)
	case "token":
		runToken()
	case "doctor":
		runDoctor()
	default:
		runServer(nil)
	}
}

// serveFlags holds CLI overrides for serve mode
type serveFlags struct {
	host      string
	port      string
	noSandbox bool
}

func parseServeFlags(args []string) *serveFlags {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	f := &serveFlags{}
	fs.StringVar(&f.host, "host", "", "Listen address (default: 127.0.0.1, Docker auto-detects 0.0.0.0)")
	fs.StringVar(&f.port, "port", "", "API port (default: 19280)")
	fs.BoolVar(&f.noSandbox, "no-sandbox", false, "Disable Chromium sandbox (required in Docker)")
	fs.Parse(args)
	return f
}

func runToken() {
	exe, _ := os.Executable()
	baseDir := filepath.Dir(exe)
	tokenPath := filepath.Join(baseDir, "data", ".api-token")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Token not found. Start the server first.\n")
		os.Exit(1)
	}
	fmt.Println(strings.TrimSpace(string(data)))
}

func runDoctor() {
	exe, _ := os.Executable()
	baseDir := filepath.Dir(exe)

	fmt.Printf("BrowseForge v%s — Environment Check\n\n", Version)

	// Docker detection
	if isDocker() {
		fmt.Println("🐳 Docker: detected")
	} else {
		fmt.Println("💻 Docker: no (native)")
	}

	// Display
	display := os.Getenv("DISPLAY")
	if display != "" {
		fmt.Printf("🖥️  Display: %s\n", display)
	} else {
		fmt.Println("⚠️  Display: not set (browsers need DISPLAY or --headless)")
	}

	// Camoufox
	ver := browser.InstalledVersion(baseDir, "camoufox")
	if ver != "" {
		path := browser.FindBinary(baseDir, "camoufox")
		fmt.Printf("✅ Camoufox: %s (%s)\n", ver, path)
	} else {
		fmt.Println("❌ Camoufox: not installed (will download on first run)")
	}

	// CloakBrowser
	ver = browser.InstalledVersion(baseDir, "cloakbrowser")
	if ver != "" {
		path := browser.FindBinary(baseDir, "cloakbrowser")
		fmt.Printf("✅ CloakBrowser: %s (%s)\n", ver, path)
	} else {
		fmt.Println("❌ CloakBrowser: not installed (will download on first run)")
	}

	// Sandbox
	if isDocker() {
		fmt.Println("⚠️  Sandbox: Docker detected — use 'serve --no-sandbox'")
	} else {
		fmt.Println("✅ Sandbox: native environment (should work)")
	}

	// Token
	tokenPath := filepath.Join(baseDir, "data", ".api-token")
	if _, err := os.Stat(tokenPath); err == nil {
		fmt.Println("✅ API Token: configured")
	} else {
		fmt.Println("ℹ️  API Token: will be generated on first start")
	}
}

func isDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func runMCPStdio() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Executable path error: %v\n", err)
		os.Exit(1)
	}
	baseDir := filepath.Dir(exe)
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
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		camoufoxPath := browser.FindBinary(baseDir, "camoufox")
		chromiumPath := browser.FindBinary(baseDir, "cloakbrowser")
		cfg := &config.Config{
			Host: "127.0.0.1", Port: "19280", ProfilesDir: "profiles", DataDir: "data",
			LogFile: "logs/server.log", FingerprintDir: "data",
			CamoufoxPath: camoufoxPath, CloakBrowserPath: chromiumPath,
		}
		cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config encode error: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile("config.json", cfgJSON, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Config write error: %v\n", err)
			os.Exit(1)
		}
	}

	cfg, err := config.Load("config.json")
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

	mcpServer := mcp.NewServer(profileStore, browserMgr, buildHumanizeCfg(cfg), "", cfg.Version)
	mcpServer.RunStdio()
}

func runServer(flags *serveFlags) {
	// Auto-detect base directory (where the binary lives)
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Executable path error: %v\n", err)
		os.Exit(1)
	}
	baseDir := filepath.Dir(exe)
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
	cfg, err := config.Load("config.json")
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
		if err := os.WriteFile("config.json", cfgJSON, 0644); err != nil {
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

	// Workflow engine
	wfEngine := workflow.NewEngine("http://127.0.0.1:"+cfg.Port, cfg.APIToken)
	router.Post("/api/workflow/run", api.WorkflowHandler(wfEngine))

	// MCP Server
	mcpServer := mcp.NewServer(profileStore, browserMgr, buildHumanizeCfg(cfg), cfg.APIToken, cfg.Version)
	go func() {
		slog.Info("MCP server starting", "port", "19281")
		if err := http.ListenAndServe(cfg.Host+":19281", mcpServer); err != nil && err != http.ErrServerClosed {
			slog.Error("MCP server failed", "error", err)
		}
	}()

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
				fmt.Printf("║  MCP:       http://%s:19281       ║\n", cfg.Host)
				fmt.Printf("║  Token:     %s...  ║\n", token[:16])
				fmt.Println("╚══════════════════════════════════════════╝")
				if cfg.Host == "127.0.0.1" {
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
