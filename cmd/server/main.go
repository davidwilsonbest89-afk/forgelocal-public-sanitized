package main

import (
	"context"
	"encoding/json"
	"fmt"
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

const Version = "1.3.0"

func main() {
	// MCP stdio mode: BrowseForge --mcp
	if len(os.Args) > 1 && os.Args[1] == "--mcp" {
		runMCPStdio()
		return
	}

	// Normal server mode
	runServer()
}

func runMCPStdio() {
	exe, _ := os.Executable()
	baseDir := filepath.Dir(exe)
	os.Chdir(baseDir)
	os.MkdirAll("profiles", 0755)
	os.MkdirAll("data", 0755)

	// Reuse existing config (don't download browsers in MCP mode)
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		camoufoxPath := browser.FindBinary(baseDir, "camoufox")
		chromiumPath := browser.FindBinary(baseDir, "cloakbrowser")
		cfg := &config.Config{
			Port: "19280", ProfilesDir: "profiles", DataDir: "data",
			LogFile: "logs/server.log", FingerprintDir: "data",
			CamoufoxPath: camoufoxPath, CloakBrowserPath: chromiumPath,
		}
		cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile("config.json", cfgJSON, 0644)
	}

	cfg, _ := config.Load("config.json")
	profileStore, _ := profile.NewStore(cfg.ProfilesDir)
	browserMgr, _ := browser.NewManager(cfg)

	mcpServer := mcp.NewServer(profileStore, browserMgr, buildHumanizeCfg(cfg))
	mcpServer.RunStdio()
}

func runServer() {
	// Auto-detect base directory (where the binary lives)
	exe, _ := os.Executable()
	baseDir := filepath.Dir(exe)
	os.Chdir(baseDir)

	// Auto-create directories
	os.MkdirAll("profiles", 0755)
	os.MkdirAll("data", 0755)
	os.MkdirAll("logs", 0755)

	// Find or download browsers — version-based detection
	cfg, _ := config.Load("config.json")

	// Camoufox: check installed version vs expected
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
			fmt.Printf("⚠️  Camoufox download failed: %v\n", err)
		}
	}

	// CloakBrowser: check installed version vs expected
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
			fmt.Printf("⚠️  CloakBrowser download failed: %v (Chromium engine disabled)\n", err)
		}
	}

	// Save config if browser paths changed
	if camoufoxPath != cfg.CamoufoxPath || chromiumPath != cfg.CloakBrowserPath {
		cfg.CamoufoxPath = camoufoxPath
		cfg.CloakBrowserPath = chromiumPath
		cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile("config.json", cfgJSON, 0644)
	}

	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}
	cfg.Version = Version

	logger := config.SetupLogger(cfg.LogFile)
	slog.SetDefault(logger)

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

	router := api.NewRouter(cfg, profileStore, browserMgr, fpPool)

	// Workflow engine
	wfEngine := workflow.NewEngine("http://127.0.0.1:"+cfg.Port, cfg.APIToken)
	router.Post("/api/workflow/run", api.WorkflowHandler(wfEngine))

	// MCP Server
	mcpServer := mcp.NewServer(profileStore, browserMgr, buildHumanizeCfg(cfg))
	go func() {
		slog.Info("MCP server starting", "port", "19281")
		http.ListenAndServe("127.0.0.1:19281", mcpServer)
	}()

	srv := &http.Server{Addr: "127.0.0.1:" + cfg.Port, Handler: router}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for server ready, then open dashboard
	go func() {
		for i := 0; i < 30; i++ {
			resp, err := http.Get("http://127.0.0.1:" + cfg.Port + "/api/status")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				token := cfg.APIToken
				url := fmt.Sprintf("http://127.0.0.1:%s#%s", cfg.Port, token)
				fmt.Println("╔══════════════════════════════════════════╗")
				fmt.Printf("║        🦊 BrowseForge v%-16s║\n", Version)
				fmt.Println("╠══════════════════════════════════════════╣")
				fmt.Printf("║  Dashboard: http://127.0.0.1:%-12s║\n", cfg.Port)
				fmt.Printf("║  MCP:       http://127.0.0.1:19281       ║\n")
				fmt.Printf("║  Token:     %s...  ║\n", token[:16])
				fmt.Println("╚══════════════════════════════════════════╝")
				openBrowser(url)
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	browserMgr.Close()
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
