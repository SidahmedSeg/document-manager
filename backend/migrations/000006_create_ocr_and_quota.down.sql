-- =============================================================================
-- Migration: 000006_create_ocr_and_quota (ROLLBACK)
-- Description: Drop OCR and quota tables
-- =============================================================================

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS usage_events CASCADE;
DROP TABLE IF EXISTS quota_history CASCADE;
DROP TABLE IF EXISTS quota_usage CASCADE;
DROP TABLE IF EXISTS ocr_jobs CASCADE;
