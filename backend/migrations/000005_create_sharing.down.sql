-- =============================================================================
-- Migration: 000005_create_sharing (ROLLBACK)
-- Description: Drop sharing tables
-- =============================================================================

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS shared_folders CASCADE;
DROP TABLE IF EXISTS share_link_access_log CASCADE;
DROP TABLE IF EXISTS public_share_links CASCADE;
DROP TABLE IF EXISTS shared_documents CASCADE;
