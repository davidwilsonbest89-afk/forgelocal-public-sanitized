package backup

import (
	"database/sql"
	_ "embed"
	"fmt"
)

const (
	backupSchemaVersion       = 1
	productSchemaVersion      = 2
	proxyIndexesSchemaVersion = 3
)

//go:embed migrations/0001_back01.sql
var backupSchemaSQL string

//go:embed migrations/0002_product.sql
var productSchemaSQL string

//go:embed migrations/0003_proxy_reference_indexes.sql
var proxyIndexesSchemaSQL string

type schemaMigration struct {
	version int
	sql     string
}

var schemaMigrations = []schemaMigration{
	{version: backupSchemaVersion, sql: backupSchemaSQL},
	{version: productSchemaVersion, sql: productSchemaSQL},
	{version: proxyIndexesSchemaVersion, sql: proxyIndexesSchemaSQL},
}

// Migrate applies the Core SQLite schema in order. Each migration and its ledger
// row commit atomically; a failed migration leaves its own version unapplied.
func Migrate(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA journal_mode = WAL`); err != nil {
		return fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		return fmt.Errorf("set busy timeout: %w", err)
	}

	for _, migration := range schemaMigrations {
		applied, err := migrationApplied(db, migration.version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigration(db, migration); err != nil {
			return err
		}
	}
	return nil
}

func migrationApplied(db *sql.DB, version int) (bool, error) {
	var ledgerExists int
	if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations')`).Scan(&ledgerExists); err != nil {
		return false, fmt.Errorf("check schema migration ledger: %w", err)
	}
	if ledgerExists == 0 {
		return false, nil
	}

	var found int
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)`, version).Scan(&found)
	if err != nil {
		return false, fmt.Errorf("check schema migration %d: %w", version, err)
	}
	return found == 1, nil
}

func applyMigration(db *sql.DB, migration schemaMigration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin schema migration %d: %w", migration.version, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(migration.sql); err != nil {
		return fmt.Errorf("apply schema migration %d: %w", migration.version, err)
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (?, strftime('%Y-%m-%dT%H:%M:%fZ','now'))`, migration.version); err != nil {
		return fmt.Errorf("record schema migration %d: %w", migration.version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit schema migration %d: %w", migration.version, err)
	}
	return nil
}
