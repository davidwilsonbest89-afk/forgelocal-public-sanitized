package browser

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Downloader handles first-run browser binary downloads

type BrowserInfo struct {
	Name    string
	Exists  bool
	Path    string
	Version string
}

func CheckBrowsers(camoufoxPath, cloakPath string) (camoufox, cloak BrowserInfo) {
	camoufox = BrowserInfo{Name: "Camoufox (Firefox)"}
	if camoufoxPath != "" {
		if _, err := os.Stat(camoufoxPath); err == nil {
			camoufox.Exists = true
			camoufox.Path = camoufoxPath
		}
	}

	cloak = BrowserInfo{Name: "CloakBrowser (Chromium)"}
	if cloakPath != "" {
		if _, err := os.Stat(cloakPath); err == nil {
			cloak.Exists = true
			cloak.Path = cloakPath
		}
	}
	return
}

func DownloadCamoufox(baseDir string) (string, error) {
	version := "v135.0.1-beta.24"
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
		extractedBin = "camoufox/camoufox"
	default:
		return "", fmt.Errorf("unsupported OS: %s", osName)
	}

	url := fmt.Sprintf("https://github.com/daijro/camoufox/releases/download/%s/%s", version, filename)
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
	default:
		return "", fmt.Errorf("unsupported OS: %s", osName)
	}

	// Use the version that has macOS builds
	version := "chromium-v145.0.7632.109.2"
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
	return binPath, nil
}

func download(url, dest string) error {
	cmd := exec.Command("curl", "-L", "-o", dest, "--progress-bar", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extract(file, destDir string) error {
	if strings.HasSuffix(file, ".zip") {
		return exec.Command("unzip", "-q", "-o", file, "-d", destDir).Run()
	}
	if strings.HasSuffix(file, ".tar.gz") {
		return exec.Command("tar", "xzf", file, "-C", destDir).Run()
	}
	return fmt.Errorf("unknown archive format: %s", file)
}
