-- =============================================================================
-- Migration: 000002_create_tenants_and_users (ROLLBACK)
-- Description: Drop tenants and users tables
-- =============================================================================

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS tenant_invitations CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;
DROP TABLE IF EXISTS tenant_users CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;
