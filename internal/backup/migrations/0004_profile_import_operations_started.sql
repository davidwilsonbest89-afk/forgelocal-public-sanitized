-- v4: durable import journal state. SQLite cannot alter a CHECK constraint in place,
-- so the operation ledger is rebuilt atomically while preserving every prior row.
ALTER TABLE profile_import_operations RENAME TO profile_import_operations_v3;

CREATE TABLE profile_import_operations (
  id TEXT PRIMARY KEY,
  source_kind TEXT NOT NULL CHECK (source_kind IN ('profile_json', 'groups_json')),
  source_sha256 TEXT NOT NULL,
  dry_run INTEGER NOT NULL CHECK (dry_run IN (0, 1)),
  state TEXT NOT NULL CHECK (state IN ('planned', 'started', 'validated', 'committed', 'rolled_back', 'failed')),
  summary_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(summary_json)),
  correlation_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  error_code TEXT NOT NULL DEFAULT ''
);

INSERT INTO profile_import_operations (
  id, source_kind, source_sha256, dry_run, state, summary_json,
  correlation_id, created_at, updated_at, error_code
)
SELECT
  id, source_kind, source_sha256, dry_run, state, summary_json,
  correlation_id, created_at, updated_at, error_code
FROM profile_import_operations_v3;

DROP TABLE profile_import_operations_v3;

CREATE INDEX idx_profile_import_operations_state_created
  ON profile_import_operations(state, created_at DESC);
