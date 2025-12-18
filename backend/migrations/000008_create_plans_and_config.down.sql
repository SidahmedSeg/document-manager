-- =============================================================================
-- Migration: 000008_create_plans_and_config (ROLLBACK)
-- Description: Drop system configuration and settings tables
-- =============================================================================

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS integrations CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS feature_flags CASCADE;
DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS tenant_settings CASCADE;
DROP TABLE IF EXISTS system_settings CASCADE;
