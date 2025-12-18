-- =============================================================================
-- Migration: 000007_create_activity_logs (ROLLBACK)
-- Description: Drop activity logs and notifications tables
-- =============================================================================

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS webhook_logs CASCADE;
DROP TABLE IF EXISTS notification_preferences CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS activity_logs CASCADE;
