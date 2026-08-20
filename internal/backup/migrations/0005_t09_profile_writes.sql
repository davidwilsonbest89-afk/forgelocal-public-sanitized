-- T09 — Profile Writes. Indexes for lifecycle state lookups and write audit
-- throughput. The profile business state remains owned by the Core profile
-- store (files, single-process mutation per profile per T08 session locks);
-- this migration only improves SQLite read/audit performance and does not
-- change any schema semantics.
CREATE INDEX IF NOT EXISTS idx_profiles_lifecycle_state ON profiles (lifecycle_state);
CREATE INDEX IF NOT EXISTS idx_profiles_updated_at ON profiles (updated_at);
CREATE INDEX IF NOT EXISTS idx_audit_events_created_at ON audit_events (created_at);
CREATE INDEX IF NOT EXISTS idx_audit_events_correlation_id ON audit_events (correlation_id);
