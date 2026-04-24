package browser

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Downloader handles first-run browser binary downloads

// Version constants — bump these to trigger auto-update
const (
	CamoufoxVersion = "v135.0.1-beta.24"
)

// ExpectedCloakBrowserVersion returns the expected version for the current platform.
func ExpectedCloakBrowserVersion() string {
	if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" {
		return "chromium-v146.0.7680.177.3"
	}
	return CloakBrowserVersion
}

const CloakBrowserVersion = "chromium-v145.0.7632.109.2"

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

// InstalledVersion reads the .version marker file in a browser directory.
// Returns empty string if not installed.
func InstalledVersion(baseDir, browserName string) string {
	data, err := os.ReadFile(filepath.Join(baseDir, "browsers", browserName, ".version"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeVersionMarker(dir, version string) {
	os.WriteFile(filepath.Join(dir, ".version"), []byte(version), 0644)
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
	case "cloakbrowser":
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
		for _, name := range exeNames {
			p := filepath.Join(dir, entry.Name(), name)
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p
			}
		}
	}
	return ""
}

func DownloadCamoufox(baseDir string) (string, error) {
	arch := runtime.GOARCH
	osName := runtime.GOOS

	var filename, extractedBin string
	switch osName {
	case "darwin":
		suffix := "x86_64"
		if arch == "arm64" {
			suffix = "arm64"
		}
		filename = fmt.Sprintf("camoufox-135.0.1-beta.24-mac.%s.zip", suffix)
		extractedBin = "Camoufox.app/Contents/MacOS/camoufox"
	case "linux":
		suffix := "x86_64"
		if arch == "arm64" {
			suffix = "arm64"
		}
		filename = fmt.Sprintf("camoufox-135.0.1-beta.24-lin.%s.zip", suffix)
		extractedBin = "camoufox"
	case "windows":
		filename = "camoufox-135.0.1-beta.24-win.x86_64.zip"
		extractedBin = "camoufox.exe"
	default:
		return "", fmt.Errorf("unsupported OS: %s", osName)
	}

	url := fmt.Sprintf("https://github.com/daijro/camoufox/releases/download/%s/%s", CamoufoxVersion, filename)
	destDir := filepath.Join(baseDir, "browsers", "camoufox")
	os.MkdirAll(destDir, 0755)

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
	os.Remove(tmpFile)

	binPath := filepath.Join(destDir, extractedBin)
	if osName == "darwin" {
		exec.Command("xattr", "-cr", destDir).Run()
	}
	os.Chmod(binPath, 0755)

	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("binary not found after extract: %s", binPath)
	}

	fmt.Println("✅ Camoufox installed")
	writeVersionMarker(destDir, CamoufoxVersion)
	return binPath, nil
}

func DownloadCloakBrowser(baseDir string) (string, error) {
	arch := runtime.GOARCH
	osName := runtime.GOOS

	var filename, extractedBin string
	switch osName {
	case "darwin":
		suffix := "x64"
		if arch == "arm64" {
			suffix = "arm64"
		}
		filename = fmt.Sprintf("cloakbrowser-darwin-%s.tar.gz", suffix)
		extractedBin = "Chromium.app/Contents/MacOS/Chromium"
	case "linux":
		suffix := "x64"
		if arch == "arm64" {
			suffix = "arm64"
		}
		filename = fmt.Sprintf("cloakbrowser-linux-%s.tar.gz", suffix)
		extractedBin = "chrome"
	case "windows":
		filename = "cloakbrowser-windows-x64.zip"
		extractedBin = "chrome.exe"
	default:
		return "", fmt.Errorf("unsupported OS: %s", osName)
	}

	// Use the version constant
	version := CloakBrowserVersion
	// Linux arm64 only available in v146
	if osName == "linux" && arch == "arm64" {
		version = "chromium-v146.0.7680.177.3"
	}
	url := fmt.Sprintf("https://github.com/CloakHQ/CloakBrowser/releases/download/%s/%s", version, filename)
	destDir := filepath.Join(baseDir, "browsers", "cloakbrowser")
	os.MkdirAll(destDir, 0755)

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
	os.Remove(tmpFile)

	binPath := filepath.Join(destDir, extractedBin)
	if osName == "darwin" {
		exec.Command("xattr", "-cr", destDir).Run()
	}
	os.Chmod(binPath, 0755)

	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("binary not found after extract: %s", binPath)
	}

	fmt.Println("✅ CloakBrowser installed")
	writeVersionMarker(destDir, version)
	return binPath, nil
}

func download(dlURL, dest string) error {
	resp, err := http.Get(dlURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

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
				fmt.Fprintf(os.Stderr, "\r  %d / %d MB (%.0f%%)", written>>20, total>>20, float64(written)/float64(total)*100)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	fmt.Fprintln(os.Stderr)
	return nil
}

func extract(file, destDir string) error {
	if strings.HasSuffix(file, ".zip") {
		return extractZip(file, destDir)
	}
	if strings.HasSuffix(file, ".tar.gz") {
		return extractTarGz(file, destDir)
	}
	return fmt.Errorf("unknown archive format: %s", file)
}

func extractZip(zipFile, destDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue // skip zip slip
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0755)
		mode := f.Mode() | 0644 // ensure writable on Windows (mode 0 = read-only)
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extractTarGz(tarFile, destDir string) error {
	f, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(destDir, hdr.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue // skip path traversal
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			mode := os.FileMode(hdr.Mode) | 0644 // ensure writable on Windows
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			_, err = io.Copy(out, tr)
			out.Close()
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			os.MkdirAll(filepath.Dir(target), 0755)
			os.Remove(target)
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				// Windows: symlink needs admin/dev-mode — copy the target instead
				src := filepath.Join(filepath.Dir(target), hdr.Linkname)
				if data, readErr := os.ReadFile(src); readErr == nil {
					os.WriteFile(target, data, 0755)
				}
			}
		}
	}
	return nil
}
