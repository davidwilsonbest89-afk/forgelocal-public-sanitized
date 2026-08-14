-- ForgeLocal Product Schema v0.1 (planned Core migration 0002)
--
-- Scope: local-first product metadata. This migration is designed to run after
-- BACK-01 migration 0001 in the same SQLite database and uses the existing
-- schema_migrations ledger. It must not alter the frozen BACK-01 release
-- artifact, runtime, SBOM, checksum, gates, or pilot status.
--
-- Security invariant: no password, username, token, API key, or other secret
-- value is stored in SQLite. A *_secret_ref field is an opaque reference to the
-- operating-system vault; it is not an access credential.

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS proxy_providers (
  id TEXT PRIMARY KEY,
  adapter_id TEXT NOT NULL,
  display_name TEXT NOT NULL COLLATE NOCASE UNIQUE,
  api_base_url TEXT NOT NULL DEFAULT '',
  credential_secret_ref TEXT NOT NULL DEFAULT '' CHECK (
    credential_secret_ref = '' OR credential_secret_ref GLOB 'proxy.provider.[0-9A-Za-z._-]*'
  ),
  public_config_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(public_config_json)),
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS runtime_candidates (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  architecture TEXT NOT NULL,
  binary_path TEXT NOT NULL DEFAULT '',
  binary_sha256 TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL CHECK (status IN ('candidate', 'validated', 'quarantined', 'retired')),
  created_at TEXT NOT NULL,
  CHECK ((binary_path = '' AND binary_sha256 = '') OR (binary_path <> '' AND binary_sha256 <> '')),
  UNIQUE (name, version, architecture, binary_path)
);

CREATE TABLE IF NOT EXISTS groups (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL COLLATE NOCASE UNIQUE,
  proxy_mode TEXT NOT NULL DEFAULT 'default' CHECK (proxy_mode IN ('default', 'enforced')),
  proxy_provider_id TEXT REFERENCES proxy_providers(id) ON DELETE RESTRICT,
  proxy_secret_ref TEXT NOT NULL DEFAULT '' CHECK (
    proxy_secret_ref = '' OR proxy_secret_ref GLOB 'proxy.group.[0-9A-Za-z._-]*'
  ),
  proxy_type TEXT NOT NULL DEFAULT '' CHECK (proxy_type IN ('', 'http', 'socks5')),
  proxy_host TEXT NOT NULL DEFAULT '',
  proxy_port INTEGER NOT NULL DEFAULT 0 CHECK (proxy_port BETWEEN 0 AND 65535),
  proxy_region TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  CHECK (
    (proxy_type = '' AND proxy_host = '' AND proxy_port = 0 AND proxy_secret_ref = '') OR
    (proxy_type IN ('http', 'socks5') AND proxy_host <> '' AND proxy_port BETWEEN 1 AND 65535)
  ),
  CHECK (proxy_mode = 'default' OR proxy_type <> '')
);

CREATE TABLE IF NOT EXISTS profiles (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL COLLATE NOCASE,
  runtime_id TEXT NOT NULL REFERENCES runtime_candidates(id) ON DELETE RESTRICT,
  group_id TEXT REFERENCES groups(id) ON DELETE RESTRICT,
  lifecycle_state TEXT NOT NULL DEFAULT 'active' CHECK (lifecycle_state IN ('active', 'archived', 'quarantined')),
  profile_dir TEXT NOT NULL UNIQUE,
  container_id TEXT NOT NULL DEFAULT '',
  fingerprint_seed INTEGER,
  identity_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(identity_json)),
  fingerprint_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(fingerprint_json)),
  metadata_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(metadata_json)),
  proxy_provider_id TEXT REFERENCES proxy_providers(id) ON DELETE RESTRICT,
  proxy_secret_ref TEXT NOT NULL DEFAULT '' CHECK (
    proxy_secret_ref = '' OR proxy_secret_ref GLOB 'proxy.[0-9A-Za-z._-]*'
  ),
  proxy_type TEXT NOT NULL DEFAULT '' CHECK (proxy_type IN ('', 'http', 'socks5')),
  proxy_host TEXT NOT NULL DEFAULT '',
  proxy_port INTEGER NOT NULL DEFAULT 0 CHECK (proxy_port BETWEEN 0 AND 65535),
  proxy_region TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  last_used_at TEXT NOT NULL DEFAULT '',
  CHECK (
    (proxy_type = '' AND proxy_host = '' AND proxy_port = 0 AND proxy_secret_ref = '') OR
    (proxy_type IN ('http', 'socks5') AND proxy_host <> '' AND proxy_port BETWEEN 1 AND 65535)
  )
);

CREATE TABLE IF NOT EXISTS profile_tags (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL COLLATE NOCASE UNIQUE,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS profile_tag_assignments (
  profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  tag_id TEXT NOT NULL REFERENCES profile_tags(id) ON DELETE RESTRICT,
  created_at TEXT NOT NULL,
  PRIMARY KEY (profile_id, tag_id)
);

CREATE TABLE IF NOT EXISTS proxy_test_runs (
  id TEXT PRIMARY KEY,
  profile_id TEXT REFERENCES profiles(id) ON DELETE CASCADE,
  group_id TEXT REFERENCES groups(id) ON DELETE CASCADE,
  outcome TEXT NOT NULL CHECK (outcome IN ('success', 'failure', 'timeout', 'blocked')),
  latency_ms INTEGER CHECK (latency_ms IS NULL OR latency_ms >= 0),
  error_code TEXT NOT NULL DEFAULT '',
  evidence_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(evidence_json)),
  created_at TEXT NOT NULL,
  CHECK (
    (profile_id IS NOT NULL AND group_id IS NULL) OR
    (profile_id IS NULL AND group_id IS NOT NULL)
  )
);

-- JSON-to-SQLite migration is dry-run first. Source digests and summaries are
-- retained for parity and rollback evidence; source credentials are never stored.
CREATE TABLE IF NOT EXISTS profile_import_operations (
  id TEXT PRIMARY KEY,
  source_kind TEXT NOT NULL CHECK (source_kind IN ('profile_json', 'groups_json')),
  source_sha256 TEXT NOT NULL,
  dry_run INTEGER NOT NULL CHECK (dry_run IN (0, 1)),
  state TEXT NOT NULL CHECK (state IN ('planned', 'validated', 'committed', 'rolled_back', 'failed')),
  summary_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(summary_json)),
  correlation_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  error_code TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS profile_json_parity_checks (
  id TEXT PRIMARY KEY,
  profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  source_sha256 TEXT NOT NULL,
  sqlite_record_sha256 TEXT NOT NULL,
  result TEXT NOT NULL CHECK (result IN ('match', 'mismatch', 'not_applicable')),
  checked_at TEXT NOT NULL,
  correlation_id TEXT NOT NULL
);

-- details_json is restricted by application contract to redacted metadata only.
CREATE TABLE IF NOT EXISTS product_audit_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_type TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  correlation_id TEXT NOT NULL,
  details_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(details_json)),
  created_at TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_runtime_candidates_sha256_not_empty
  ON runtime_candidates(binary_sha256) WHERE binary_sha256 <> '';
CREATE INDEX IF NOT EXISTS idx_profiles_group_last_used
  ON profiles(group_id, lifecycle_state, last_used_at DESC);
CREATE INDEX IF NOT EXISTS idx_profiles_runtime_state
  ON profiles(runtime_id, lifecycle_state);
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_container_id_not_empty
  ON profiles(container_id) WHERE container_id <> '';
CREATE INDEX IF NOT EXISTS idx_profile_tag_assignments_tag
  ON profile_tag_assignments(tag_id, profile_id);
CREATE INDEX IF NOT EXISTS idx_proxy_test_runs_profile_created
  ON proxy_test_runs(profile_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_proxy_test_runs_group_created
  ON proxy_test_runs(group_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_profile_import_operations_state_created
  ON profile_import_operations(state, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_product_audit_events_entity_created
  ON product_audit_events(entity_type, entity_id, created_at DESC);
