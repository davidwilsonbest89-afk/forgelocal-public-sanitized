package config

import (
	"encoding/json"
	"log/slog"
	"os"
)

type Config struct {
	Port            string `json:"port"`
	ProfilesDir     string `json:"profiles_dir"`
	DataDir         string `json:"data_dir"`
	LogFile         string `json:"log_file"`
	CamoufoxPath    string `json:"camoufox_path"`
	CloakBrowserPath string `json:"cloakbrowser_path"`
	FingerprintDir  string `json:"fingerprint_dir"`
	APIToken        string `json:"-"` // generated at runtime
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Port:           "19280",
		ProfilesDir:    "profiles",
		DataDir:        "data",
		LogFile:        "logs/server.log",
		CamoufoxPath:   "camoufox/camoufox",
		FingerprintDir: "data",
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Warn("config not found, using defaults", "path", path)
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func SetupLogger(logFile string) *slog.Logger {
	os.MkdirAll("logs", 0755)
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return slog.New(slog.NewJSONHandler(f, nil))
}
