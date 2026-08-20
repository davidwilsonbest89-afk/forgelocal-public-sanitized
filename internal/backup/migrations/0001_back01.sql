PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS backup_operations (
  id TEXT PRIMARY KEY,
  profile_id TEXT NOT NULL,
  state TEXT NOT NULL CHECK(state IN ('staging','published_unregistered','committed','quarantined')),
  artifact_path TEXT NOT NULL,
  key_id TEXT NOT NULL,
  correlation_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  error_code TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS backups (
  id TEXT PRIMARY KEY,
  profile_id TEXT NOT NULL,
  artifact_path TEXT NOT NULL UNIQUE,
  key_id TEXT NOT NULL,
  sha256 TEXT NOT NULL,
  created_at TEXT NOT NULL,
  restored_from_backup_id TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS restore_operations (
  id TEXT PRIMARY KEY,
  backup_id TEXT NOT NULL,
  source_profile_id TEXT NOT NULL,
  target_profile_id TEXT NOT NULL,
  target_path TEXT NOT NULL,
  state TEXT NOT NULL CHECK(state IN ('staging','committed','failed')),
  correlation_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  error_code TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS audit_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  correlation_id TEXT NOT NULL,
  details_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);
