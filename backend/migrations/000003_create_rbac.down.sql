-- =============================================================================
-- Migration: 000003_create_rbac (ROLLBACK)
-- Description: Drop RBAC tables
-- =============================================================================

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS resource_permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
