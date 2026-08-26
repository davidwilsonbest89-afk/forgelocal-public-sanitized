package browser

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	bfruntime "forgelocal/internal/runtime"
)

// Downloader handles first-run browser binary downloads

// Version constants — bump these to trigger auto-update
const (
	CamoufoxVersion              = "v135.0.1-beta.24"
	CloakBrowserVersion          = "chromium-v146.0.7680.177.4"
	BrowseForgeChromiumVersion   = "v0.1.6-alpha.0"
	BrowseForgeChromiumRelease   = "https://github.com/nczz/browseforge-runtime-chromium/releases/download"
	BrowseForgeChromiumRuntimeID = "browseforge-chromium"
)

var ErrUnsupportedRuntimePlatform = errors.New("unsupported runtime platform")

type UnsupportedRuntimePlatformError struct {
	Runtime string
	Version string
	GOOS    string
	GOARCH  string
}

func (e UnsupportedRuntimePlatformError) Error() string {
	return fmt.Sprintf("%s %s is not available for %s/%s", e.Runtime, e.Version, e.GOOS, e.GOARCH)
}

func (e UnsupportedRuntimePlatformError) Unwrap() error {
	return ErrUnsupportedRuntimePlatform
}

type RuntimeSupport struct {
	RuntimeID         string
	DisplayName       string
	Version           string
	GOOS              string
	GOARCH            string
	PlatformSupported bool
	UnsupportedReason string
}

func unsupportedRuntime(runtimeID, displayName, version, goos, goarch string) RuntimeSupport {
	err := UnsupportedRuntimePlatformError{Runtime: displayName, Version: version, GOOS: goos, GOARCH: goarch}
	return RuntimeSupport{
		RuntimeID:         runtimeID,
		DisplayName:       displayName,
		Version:           version,
		GOOS:              goos,
		GOARCH:            goarch,
		PlatformSupported: false,
		UnsupportedReason: err.Error(),
	}
}

func supportedRuntime(runtimeID, displayName, version, goos, goarch string) RuntimeSupport {
	return RuntimeSupport{
		RuntimeID:         runtimeID,
		DisplayName:       displayName,
		Version:           version,
		GOOS:              goos,
		GOARCH:            goarch,
		PlatformSupported: true,
	}
}

func CurrentRuntimeSupport(runtimeID string) RuntimeSupport {
	return RuntimeSupportFor(runtimeID, runtime.GOOS, runtime.GOARCH)
}

func RuntimeSupportFor(runtimeID, goos, goarch string) RuntimeSupport {
	switch runtimeID {
	case "camoufox":
		return CamoufoxSupportFor(goos, goarch)
	case "cloakbrowser":
		return CloakBrowserSupportFor(goos, goarch)
	case BrowseForgeChromiumRuntimeID:
		return BrowseForgeChromiumSupportFor(goos, goarch)
	default:
		return unsupportedRuntime(runtimeID, runtimeID, "", goos, goarch)
	}
}

// ExpectedCloakBrowserVersion returns the expected version for the current platform.
func ExpectedCloakBrowserVersion() string {
	if runtime.GOOS == "darwin" {
		return "chromium-v145.0.7632.109.2"
	}
	return CloakBrowserVersion
}

type BrowserInfo struct {
	Name    string
	Exists  bool
	Path    string
	Version string
}

func CheckBrowsers(camoufoxPath, cloakPath string) (camoufox, cloak BrowserInfo) {
	camoufox = BrowserInfo{Name: "Camoufox (Firefox)"}
	if camoufoxPath != "" {
		if info, err := os.Stat(camoufoxPath); err == nil && !info.IsDir() {
			camoufox.Exists = true
			camoufox.Path = camoufoxPath
		}
	}

	cloak = BrowserInfo{Name: "CloakBrowser (Chromium)"}
	if cloakPath != "" {
		if info, err := os.Stat(cloakPath); err == nil && !info.IsDir() {
			cloak.Exists = true
			cloak.Path = cloakPath
		}
	}
	return
}

// openVersionMarker opens the runtime marker beneath the selected browser root
// and refuses a symlinked marker to avoid reading outside that root.
func openVersionMarker(baseDir, browserName string) (*os.Root, *os.File, error) {
	runtimeRoot, err := os.OpenRoot(filepath.Join(baseDir, "browsers", browserName))
	if err != nil {
		return nil, nil, err
	}
	info, err := runtimeRoot.Lstat(".version")
	if err != nil {
		_ = runtimeRoot.Close()
		return nil, nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		_ = runtimeRoot.Close()
		return nil, nil, fmt.Errorf("runtime version marker is not a regular file")
	}
	file, err := runtimeRoot.Open(".version")
	if err != nil {
		_ = runtimeRoot.Close()
		return nil, nil, err
	}
	return runtimeRoot, file, nil
}

// InstalledVersion reads the .version marker file in a browser directory.
// Returns empty string if not installed or if the marker is unsafe.
func InstalledVersion(baseDir, browserName string) string {
	root, file, err := openVersionMarker(baseDir, browserName)
	if err != nil {
		return ""
	}
	defer root.Close()
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeVersionMarker(dir, version string) error {
	return os.WriteFile(filepath.Join(dir, ".version"), []byte(version), 0600)
}

// FindBinary searches a browser directory for a known executable.
// Checks exact candidate paths first, then scans one level of subdirectories.
func FindBinary(baseDir, browserName string) string {
	dir := filepath.Join(baseDir, "browsers", browserName)

	var candidates []string
	var exeNames []string
	switch browserName {
	case "camoufox":
		candidates = []string{
			filepath.Join("Camoufox.app", "Contents", "MacOS", "camoufox"),
			"camoufox.exe",
			"camoufox",
		}
		exeNames = []string{"camoufox.exe", "camoufox"}
	case "cloakbrowser", "browseforge-chromium":
		candidates = []string{
			filepath.Join("Chromium.app", "Contents", "MacOS", "Chromium"),
			"chrome.exe",
			"chrome",
			"chromium",
		}
		exeNames = []string{"chrome.exe", "chrome", "chromium"}
	}

	// Check known candidate paths
	for _, c := range candidates {
		p := filepath.Join(dir, c)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}

	// Fallback: scan one level of subdirectories for known exe names
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		for _, c := range candidates {
			p := filepath.Join(dir, entry.Name(), c)
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p
			}
		}
		for _, name := range exeNames {
			p := filepath.Join(dir, entry.Name(), name)
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p
			}
		}
	}
	return ""
}

func ExpectedBrowseForgeChromiumVersion() string {
	return BrowseForgeChromiumVersion
}

func BrowseForgeChromiumSupportFor(goos, goarch string) RuntimeSupport {
	if _, err := browseForgeChromiumPlatformFor(goos, goarch); err != nil {
		return unsupportedRuntime(BrowseForgeChromiumRuntimeID, "BrowseForge Chromium", BrowseForgeChromiumVersion, goos, goarch)
	}
	return supportedRuntime(BrowseForgeChromiumRuntimeID, "BrowseForge Chromium", BrowseForgeChromiumVersion, goos, goarch)
}

func browseForgeChromiumPlatform() (string, error) {
	return browseForgeChromiumPlatformFor(runtime.GOOS, runtime.GOARCH)
}

func browseForgeChromiumPlatformFor(goos, goarch string) (string, error) {
	switch goos {
	case "darwin":
		switch goarch {
		case "arm64":
			return "macos-arm64", nil
		case "amd64":
			return "macos-x64", nil
		}
	case "linux":
		switch goarch {
		case "amd64":
			return "linux-x64", nil
		case "arm64":
			return "linux-arm64", nil
		}
	case "windows":
		if goarch == "amd64" {
			return "windows-x64", nil
		}
	}
	return "", fmt.Errorf("unsupported BrowseForge Chromium platform: %s/%s", goos, goarch)
}

func CamoufoxSupportFor(goos, goarch string) RuntimeSupport {
	if _, err := camoufoxDownloadFilenameFor(CamoufoxVersion, goos, goarch); err != nil {
		return unsupportedRuntime("camoufox", "Camoufox", CamoufoxVersion, goos, goarch)
	}
	return supportedRuntime("camoufox", "Camoufox", CamoufoxVersion, goos, goarch)
}

func camoufoxDownloadFilenameFor(version, goos, goarch string) (string, error) {
	switch goos {
	case "darwin":
		switch goarch {
		case "amd64":
			return "camoufox-135.0.1-beta.24-mac.x86_64.zip", nil
		case "arm64":
			return "camoufox-135.0.1-beta.24-mac.arm64.zip", nil
		}
	case "linux":
		switch goarch {
		case "amd64":
			return "camoufox-135.0.1-beta.24-lin.x86_64.zip", nil
		case "arm64":
			return "camoufox-135.0.1-beta.24-lin.arm64.zip", nil
		}
	case "windows":
		if goarch == "amd64" {
			return "camoufox-135.0.1-beta.24-win.x86_64.zip", nil
		}
	}
	return "", UnsupportedRuntimePlatformError{Runtime: "Camoufox", Version: version, GOOS: goos, GOARCH: goarch}
}

func camoufoxDownloadURLFor(version, goos, goarch string) (string, string, error) {
	filename, err := camoufoxDownloadFilenameFor(version, goos, goarch)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("https://github.com/daijro/camoufox/releases/download/%s/%s", version, filename), filename, nil
}

func BrowseForgeChromiumDownloadURL(version string) (string, string, error) {
	return browseForgeChromiumDownloadURLFor(version, runtime.GOOS, runtime.GOARCH)
}

func browseForgeChromiumDownloadURLFor(version, goos, goarch string) (string, string, error) {
	platform, err := browseForgeChromiumPlatformFor(goos, goarch)
	if err != nil {
		return "", "", err
	}
	filename := fmt.Sprintf("browseforge-runtime-chromium-%s-%s.zip", version, platform)
	baseURL := strings.TrimRight(os.Getenv("BROWSEFORGE_CHROMIUM_RELEASE_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = BrowseForgeChromiumRelease
	}
	return fmt.Sprintf("%s/%s/%s", baseURL, version, filename), filename, nil
}

func DownloadBrowseForgeChromium(baseDir string) (string, error) {
	version := BrowseForgeChromiumVersion
	url, filename, err := BrowseForgeChromiumDownloadURL(version)
	if err != nil {
		return "", err
	}
	destDir := filepath.Join(baseDir, "browsers", BrowseForgeChromiumRuntimeID)
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return "", err
	}

	tmpFile := filepath.Join(os.TempDir(), filename)
	slog.Info("downloading BrowseForge Chromium", "url", url)
	fmt.Printf("📥 Downloading BrowseForge Chromium (%s)...\n", filename)

	if err := download(url, tmpFile); err != nil {
		return "", err
	}

	fmt.Println("📦 Extracting...")
	if err := extract(tmpFile, destDir); err != nil {
		return "", err
	}
	if err := os.Remove(tmpFile); err != nil {
		return "", fmt.Errorf("remove downloaded archive: %w", err)
	}

	// Flatten: if the archive extracted into a single subdirectory, hoist its
	// contents up so the binary lives directly under destDir (e.g. chrome is at
	// browsers/browseforge-chromium/chrome instead of .../subdir/chrome).
	if err := flattenSingleSubdir(destDir); err != nil {
		slog.Warn("flatten after extract failed", "err", err)
	}

	if runtime.GOOS == "darwin" {
		if err := clearMacOSQuarantine(destDir); err != nil {
			slog.Warn("clearing macOS quarantine failed", "err", err)
		}
	}

	binPath := FindBinary(baseDir, BrowseForgeChromiumRuntimeID)
	if binPath == "" {
		return "", fmt.Errorf("binary not found after extract in %s", destDir)
	}
	if err := os.Chmod(binPath, 0755); err != nil {
		return "", err
	}

	fmt.Println("✅ BrowseForge Chromium installed")
	if err := writeVersionMarker(destDir, version); err != nil {
		return "", err
	}
	return binPath, nil
}

func clearMacOSQuarantine(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := exec.CommandContext(ctx, "xattr", "-cr", path).Run()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

func DownloadCamoufox(baseDir string) (string, error) {
	if err := bfruntime.RequireExecution(bfruntime.Camoufox); err != nil {
		return "", err
	}
	url, filename, err := camoufoxDownloadURLFor(CamoufoxVersion, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	destDir := filepath.Join(baseDir, "browsers", "camoufox")
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return "", err
	}

	tmpFile := filepath.Join(os.TempDir(), filename)
	slog.Info("downloading Camoufox", "url", url)
	fmt.Printf("📥 Downloading Camoufox (%s)...\n", filename)

	if err := download(url, tmpFile); err != nil {
		return "", err
	}

	fmt.Println("📦 Extracting...")
	if err := extract(tmpFile, destDir); err != nil {
		return "", err
	}
	if err := os.Remove(tmpFile); err != nil {
		return "", fmt.Errorf("remove downloaded archive: %w", err)
	}

	if runtime.GOOS == "darwin" {
		if err := clearMacOSQuarantine(destDir); err != nil {
			slog.Warn("clearing macOS quarantine failed", "err", err)
		}
	}

	binPath := FindBinary(baseDir, "camoufox")
	if binPath == "" {
		return "", fmt.Errorf("binary not found after extract in %s", destDir)
	}
	if err := os.Chmod(binPath, 0755); err != nil {
		return "", err
	}

	fmt.Println("✅ Camoufox installed")
	if err := writeVersionMarker(destDir, CamoufoxVersion); err != nil {
		return "", err
	}
	return binPath, nil
}

func CloakBrowserSupportFor(goos, goarch string) RuntimeSupport {
	supported := false
	switch goos {
	case "darwin", "linux":
		supported = goarch == "amd64" || goarch == "arm64"
	case "windows":
		supported = goarch == "amd64"
	}
	version := CloakBrowserVersion
	if goos == "darwin" {
		version = "chromium-v145.0.7632.109.2"
	}
	if !supported {
		return unsupportedRuntime("cloakbrowser", "CloakBrowser", version, goos, goarch)
	}
	return supportedRuntime("cloakbrowser", "CloakBrowser", version, goos, goarch)
}

func DownloadCloakBrowser(baseDir string) (string, error) {
	arch := runtime.GOARCH
	osName := runtime.GOOS

	var filename string
	switch osName {
	case "darwin":
		suffix := "x64"
		if arch == "arm64" {
			suffix = "arm64"
		}
		filename = fmt.Sprintf("cloakbrowser-darwin-%s.tar.gz", suffix)
	case "linux":
		suffix := "x64"
		if arch == "arm64" {
			suffix = "arm64"
		}
		filename = fmt.Sprintf("cloakbrowser-linux-%s.tar.gz", suffix)
	case "windows":
		filename = "cloakbrowser-windows-x64.zip"
	default:
		return "", fmt.Errorf("unsupported OS: %s", osName)
	}

	version := CloakBrowserVersion
	if runtime.GOOS == "darwin" {
		version = "chromium-v145.0.7632.109.2"
	}
	url := fmt.Sprintf("https://github.com/CloakHQ/CloakBrowser/releases/download/%s/%s", version, filename)
	destDir := filepath.Join(baseDir, "browsers", "cloakbrowser")
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return "", err
	}

	tmpFile := filepath.Join(os.TempDir(), filename)
	slog.Info("downloading CloakBrowser", "url", url)
	fmt.Printf("📥 Downloading CloakBrowser (%s)...\n", filename)

	if err := download(url, tmpFile); err != nil {
		return "", err
	}

	fmt.Println("📦 Extracting...")
	if err := extract(tmpFile, destDir); err != nil {
		return "", err
	}
	if err := os.Remove(tmpFile); err != nil {
		return "", fmt.Errorf("remove downloaded archive: %w", err)
	}

	if osName == "darwin" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := exec.CommandContext(ctx, "xattr", "-cr", destDir).Run()
		cancel()
		if err != nil {
			return "", fmt.Errorf("clear macOS extended attributes: %w", err)
		}
	}

	binPath := FindBinary(baseDir, "cloakbrowser")
	if binPath == "" {
		return "", fmt.Errorf("binary not found after extract in %s", destDir)
	}
	if err := os.Chmod(binPath, 0755); err != nil {
		return "", err
	}

	fmt.Println("✅ CloakBrowser installed")
	if err := writeVersionMarker(destDir, version); err != nil {
		return "", err
	}
	return binPath, nil
}

// flattenSingleSubdir hoists the contents of a lone subdirectory up into dir.
// This handles archives that wrap everything in a version-named folder, e.g.
// browseforge-runtime-chromium-v0.1.6-alpha.0-linux-x64/chrome → chrome.
// Files already at the top level (like .version) are preserved.
func flattenSingleSubdir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Count subdirectories and top-level files (ignoring .version marker).
	var dirs []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		}
	}

	// Only flatten when there is exactly one subdirectory.
	if len(dirs) != 1 {
		return nil
	}

	subdir := filepath.Join(dir, dirs[0].Name())
	subEntries, err := os.ReadDir(subdir)
	if err != nil {
		return err
	}

	// Move each entry from the subdirectory up to dir.
	for _, e := range subEntries {
		src := filepath.Join(subdir, e.Name())
		dst := filepath.Join(dir, e.Name())
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("flatten: rename %s → %s: %w", src, dst, err)
		}
	}

	// Remove the now-empty subdirectory.
	return os.Remove(subdir)
}

func download(dlURL, dest string) (result error) {
	req, err := http.NewRequest(http.MethodGet, dlURL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return fmt.Errorf("download redirects are not allowed")
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); result == nil {
			result = closeErr
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	outRoot, out, err := openBrowserRegularFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); result == nil {
			result = closeErr
		}
		if rootErr := outRoot.Close(); result == nil {
			result = rootErr
		}
	}()

	total := resp.ContentLength
	var written int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := out.Write(buf[:n]); err != nil {
				return err
			}
			written += int64(n)
			if total > 0 {
				if _, err := fmt.Fprintf(os.Stderr, "\r  %d / %d MB (%.0f%%)", written>>20, total>>20, float64(written)/float64(total)*100); err != nil {
					return err
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	_, result = fmt.Fprintln(os.Stderr)
	return result
}

func extract(file, destDir string) error {
	if strings.HasSuffix(file, ".zip") {
		return secureExtractArchive(file, destDir)
	}
	if strings.HasSuffix(file, ".tar.gz") {
		return secureExtractArchive(file, destDir)
	}
	return fmt.Errorf("unknown archive format: %s", file)
}

func extractZip(zipFile, destDir string) error {
	return secureExtractArchive(zipFile, destDir)
}

func extractTarGz(tarFile, destDir string) error {
	return secureExtractArchive(tarFile, destDir)
}
