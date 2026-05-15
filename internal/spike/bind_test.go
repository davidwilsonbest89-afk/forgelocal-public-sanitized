package spike_test

import (
	"net/url"
	"os"
	"testing"

	"github.com/playwright-community/playwright-go"
)

func TestPlaywrightBindEndpointWithCamoufox(t *testing.T) {
	camoufoxPath := os.Getenv("CAMOUFOX_PATH")
	if camoufoxPath == "" {
		camoufoxPath = "browsers/camoufox/Camoufox.app/Contents/MacOS/camoufox"
	}
	if _, err := os.Stat(camoufoxPath); os.IsNotExist(err) {
		t.Skipf("Camoufox not found at %s", camoufoxPath)
	}

	playwright.Install(&playwright.RunOptions{Browsers: []string{}})

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("playwright.Run: %v", err)
	}
	defer pw.Stop()

	userDataDir := t.TempDir()
	ctx, err := pw.Firefox.LaunchPersistentContext(userDataDir, playwright.BrowserTypeLaunchPersistentContextOptions{
		ExecutablePath: playwright.String(camoufoxPath),
		Headless:       playwright.Bool(true),
	})
	if err != nil {
		t.Fatalf("launch persistent context: %v", err)
	}
	defer ctx.Close()

	browser := ctx.Browser()
	if browser == nil {
		t.Fatal("persistent context browser is nil")
	}

	result, err := browser.Bind("browseforge-bind-test", playwright.BrowserBindOptions{
		Host: playwright.String("127.0.0.1"),
		Port: playwright.Int(0),
	})
	if err != nil {
		t.Fatalf("browser.Bind: %v", err)
	}
	defer browser.Unbind()

	endpoint, err := url.Parse(result.Endpoint)
	if err != nil {
		t.Fatalf("parse endpoint %q: %v", result.Endpoint, err)
	}
	if endpoint.Path == "" || endpoint.Path == "/" {
		t.Fatalf("bind endpoint missing websocket path: %s", result.Endpoint)
	}

	pw2, err := playwright.Run()
	if err != nil {
		t.Fatalf("playwright.Run second client: %v", err)
	}
	defer pw2.Stop()

	connected, err := pw2.Firefox.Connect(result.Endpoint)
	if err != nil {
		t.Fatalf("connect to bind endpoint %s: %v", result.Endpoint, err)
	}
	defer connected.Close()

	page, err := connected.NewPage()
	if err != nil {
		t.Fatalf("new page through bind endpoint: %v", err)
	}
	defer page.Close()

	if _, err := page.Goto("about:blank"); err != nil {
		t.Fatalf("goto through bind endpoint: %v", err)
	}
}
