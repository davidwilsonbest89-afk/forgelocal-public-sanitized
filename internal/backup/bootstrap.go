package backup

import (
	"fmt"
	"os"
	"path/filepath"

	"forgelocal/internal/secrets"
)

// OpenCoreService initializes all BACK-01 state owned by the ForgeLocal Core.
func OpenCoreService(dataDir string) (*Service, *SQLiteStore, error) {
	if dataDir == "" {
		return nil, nil, fmt.Errorf("data directory is required")
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, nil, err
	}
	metadataPath := filepath.Join(dataDir, "metadata.db")
	store, err := OpenSQLite(metadataPath)
	if err != nil {
		return nil, nil, err
	}
	root := filepath.Join(dataDir, "backups")
	if err := os.MkdirAll(root, 0700); err != nil {
		_ = store.Close()
		return nil, nil, err
	}
	service := &Service{Root: root, Vault: secrets.NewSystemVault("ForgeLocal"), Store: store, Locks: NewProfileLocks()}
	if err := service.ReconcileOnce(); err != nil {
		_ = store.Close()
		return nil, nil, fmt.Errorf("backup reconciliation: %w", err)
	}
	return service, store, nil
}
