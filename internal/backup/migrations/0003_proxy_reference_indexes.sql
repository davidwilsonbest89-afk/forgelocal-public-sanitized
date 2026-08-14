-- ForgeLocal Product Schema v0.1: additive proxy reference indexes.
--
-- This migration preserves the frozen BACK-01 release and the v2 table contract.
-- It adds lookup indexes only after the v2 columns, foreign keys, and CHECK
-- constraints have been introduced. Secret references remain opaque vault IDs,
-- never secret values.

CREATE INDEX IF NOT EXISTS idx_profiles_proxy_provider_id_not_null
  ON profiles(proxy_provider_id) WHERE proxy_provider_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_profiles_proxy_secret_ref_not_empty
  ON profiles(proxy_secret_ref) WHERE proxy_secret_ref <> '';
CREATE INDEX IF NOT EXISTS idx_groups_proxy_provider_id_not_null
  ON groups(proxy_provider_id) WHERE proxy_provider_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_groups_proxy_secret_ref_not_empty
  ON groups(proxy_secret_ref) WHERE proxy_secret_ref <> '';
