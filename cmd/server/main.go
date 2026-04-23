package main

import (
	"context"
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
	"browseforge/internal/mcp"
	"browseforge/internal/profile"
	"browseforge/internal/workflow"
)

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
		camoufoxPath := findBrowser(baseDir, "camoufox")
		chromiumPath := findBrowser(baseDir, "cloakbrowser")
		cfgJSON := fmt.Sprintf(`{
  "port": "19280",
  "profiles_dir": "profiles",
  "data_dir": "data",
  "log_file": "logs/server.log",
  "camoufox_path": %q,
  "cloakbrowser_path": %q,
  "fingerprint_dir": "data"
}`, camoufoxPath, chromiumPath)
		os.WriteFile("config.json", []byte(cfgJSON), 0644)
	}

	cfg, _ := config.Load("config.json")
	profileStore, _ := profile.NewStore(cfg.ProfilesDir)
	browserMgr, _ := browser.NewManager(cfg)

	mcpServer := mcp.NewServer(profileStore, browserMgr)
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

	// Find or download browsers
	camoufoxPath := findBrowser(baseDir, "camoufox")
	chromiumPath := findBrowser(baseDir, "cloakbrowser")

	if camoufoxPath == "" {
		fmt.Println("🦊 Camoufox not found. Downloading...")
		var err error
		camoufoxPath, err = browser.DownloadCamoufox(baseDir)
		if err != nil {
			fmt.Printf("⚠️  Camoufox download failed: %v\n", err)
		}
	}

	if chromiumPath == "" {
		fmt.Println("🌐 CloakBrowser not found. Downloading...")
		var err error
		chromiumPath, err = browser.DownloadCloakBrowser(baseDir)
		if err != nil {
			fmt.Printf("⚠️  CloakBrowser download failed: %v (Chromium engine disabled)\n", err)
		}
	}

	// Generate config
	cfgJSON := fmt.Sprintf(`{
  "port": "19280",
  "profiles_dir": "profiles",
  "data_dir": "data",
  "log_file": "logs/server.log",
  "camoufox_path": %q,
  "cloakbrowser_path": %q,
  "fingerprint_dir": "data"
}`, camoufoxPath, chromiumPath)
	os.WriteFile("config.json", []byte(cfgJSON), 0644)

	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

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
	mcpServer := mcp.NewServer(profileStore, browserMgr)
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
				fmt.Println("║        🦊 BrowseForge v0.2.0          ║")
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

// findBrowser locates browser binary relative to base directory
func findBrowser(baseDir, name string) string {
	candidates := []string{
		// Camoufox (Firefox)
		filepath.Join("browsers", name, "Camoufox.app", "Contents", "MacOS", "camoufox"),
		filepath.Join("camoufox", "Camoufox.app", "Contents", "MacOS", "camoufox"),
		filepath.Join("browsers", name, "camoufox"),
		// CloakBrowser (Chromium)
		filepath.Join("browsers", name, "Chromium.app", "Contents", "MacOS", "Chromium"),
		filepath.Join("browsers", name, "chrome"),
		filepath.Join("browsers", name, "chromium"),
	}
	for _, c := range candidates {
		abs := filepath.Join(baseDir, c)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return ""
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
