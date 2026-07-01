package browser

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// splitLocale splits "en-US" into ["en", "US"], fallback to ["en", "US"].
func splitLocale(locale string) [2]string {
	if i := strings.IndexByte(locale, '-'); i > 0 {
		return [2]string{locale[:i], locale[i+1:]}
	}
	return [2]string{"en", "US"}
}

// humanizeError wraps Playwright errors into user-friendly messages.
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
