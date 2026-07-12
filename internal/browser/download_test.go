package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindBinaryLocatesBrowseForgeChromiumArtifactLayouts(t *testing.T) {
	tests := []struct {
		name      string
		binaryRel string
	}{
		{
			name:      "direct app bundle",
			binaryRel: filepath.Join("Chromium.app", "Contents", "MacOS", "Chromium"),
		},
		{
			name:      "direct chrome executable",
			binaryRel: "chrome",
		},
		{
			name:      "direct windows executable",
			binaryRel: "chrome.exe",
		},
		{
			name:      "extracted artifact root with app bundle",
			binaryRel: filepath.Join("browseforge-runtime-chromium-v0.1.0-alpha.0-macos-x64", "Chromium.app", "Contents", "MacOS", "Chromium"),
		},
		{
			name:      "extracted artifact root with linux executable",
			binaryRel: filepath.Join("browseforge-runtime-chromium-v0.1.0-alpha.0-linux-x64", "chrome"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseDir := t.TempDir()
			wantPath := filepath.Join(baseDir, "browsers", BrowseForgeChromiumRuntimeID, tc.binaryRel)
			if err := os.MkdirAll(filepath.Dir(wantPath), 0755); err != nil {
				t.Fatalf("mkdir binary dir: %v", err)
			}
			if err := os.WriteFile(wantPath, []byte("chromium binary"), 0755); err != nil {
				t.Fatalf("write binary: %v", err)
			}

			if got := FindBinary(baseDir, BrowseForgeChromiumRuntimeID); got != wantPath {
				t.Fatalf("FindBinary() = %q, want %q", got, wantPath)
			}
		})
	}
}
