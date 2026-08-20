package browser

import (
	"errors"
	"os"
	"testing"

	"forgelocal/internal/config"
	"forgelocal/internal/profile"
	bfruntime "forgelocal/internal/runtime"
)

func TestCR01DownloadCamoufoxRefusesBeforeFilesystemOrNetwork(t *testing.T) {
	base := t.TempDir()
	_, err := DownloadCamoufox(base)
	if !errors.Is(err, bfruntime.ErrCamoufoxExecutionNotAuthorized) {
		t.Fatalf("DownloadCamoufox error = %v, want authorization refusal", err)
	}
	if _, statErr := os.Stat(base + "/browsers/camoufox"); !os.IsNotExist(statErr) {
		t.Fatalf("Camoufox directory was created before refusal: %v", statErr)
	}
}

func TestCR01LaunchFirefoxRefusesWithoutBinaryOrPlaywright(t *testing.T) {
	manager := &Manager{runtimes: bfruntime.NewRegistry(&config.Config{})}
	_, err := manager.launchFirefox(&profile.Profile{ID: "cr01", RuntimeID: string(bfruntime.Camoufox), ProfileDir: t.TempDir()})
	if !errors.Is(err, bfruntime.ErrCamoufoxExecutionNotAuthorized) {
		t.Fatalf("launchFirefox error = %v, want authorization refusal", err)
	}
	if code := ErrorCode(err); code != bfruntime.CamoufoxExecutionNotAuthorizedCode {
		t.Fatalf("launchFirefox code = %q", code)
	}
}
