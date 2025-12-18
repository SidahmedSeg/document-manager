-- =============================================================================
-- Migration: 000001_create_extensions_and_types (ROLLBACK)
-- Description: Drop custom types and extensions
-- =============================================================================

-- Drop custom types (in reverse order of creation)
DROP TYPE IF EXISTS notification_status CASCADE;
DROP TYPE IF EXISTS notification_type CASCADE;
DROP TYPE IF EXISTS activity_action CASCADE;
DROP TYPE IF EXISTS permission_type CASCADE;
DROP TYPE IF EXISTS subscription_status CASCADE;
DROP TYPE IF EXISTS subscription_plan CASCADE;
DROP TYPE IF EXISTS share_link_type CASCADE;
DROP TYPE IF EXISTS ocr_status CASCADE;
DROP TYPE IF EXISTS document_visibility CASCADE;
DROP TYPE IF EXISTS document_status CASCADE;

-- Note: We don't drop extensions as they might be used by other databases/schemas
-- DROP EXTENSION IF EXISTS "citext";
-- DROP EXTENSION IF EXISTS "pg_trgm";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
