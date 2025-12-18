-- =============================================================================
-- Migration: 000010_create_triggers (ROLLBACK)
-- Description: Drop all database triggers and functions
-- =============================================================================

-- Drop triggers first (in reverse order of creation)

-- Share link view count trigger
DROP TRIGGER IF EXISTS trigger_update_share_link_view_count ON share_link_access_log;

-- Quota warning trigger
DROP TRIGGER IF EXISTS trigger_send_quota_warning_notification ON quota_usage;

-- Quota check trigger
DROP TRIGGER IF EXISTS trigger_check_quota_before_document_insert ON documents;

-- OCR quota trigger
DROP TRIGGER IF EXISTS trigger_update_quota_on_ocr_completion ON ocr_jobs;

-- Document quota triggers
DROP TRIGGER IF EXISTS trigger_update_quota_on_document_insert ON documents;
DROP TRIGGER IF EXISTS trigger_update_quota_on_document_update ON documents;
DROP TRIGGER IF EXISTS trigger_update_quota_on_document_delete ON documents;

-- Category document count trigger
DROP TRIGGER IF EXISTS trigger_update_category_document_count ON documents;

-- Tag usage count triggers
DROP TRIGGER IF EXISTS trigger_update_tag_usage_count_insert ON document_tags;
DROP TRIGGER IF EXISTS trigger_update_tag_usage_count_delete ON document_tags;

-- Folder path trigger
DROP TRIGGER IF EXISTS trigger_update_folder_path ON folders;

-- Document search vector trigger
DROP TRIGGER IF EXISTS trigger_update_document_search_vector ON documents;

-- Updated_at triggers (21 tables)
DROP TRIGGER IF EXISTS trigger_update_integrations_updated_at ON integrations;
DROP TRIGGER IF EXISTS trigger_update_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS trigger_update_feature_flags_updated_at ON feature_flags;
DROP TRIGGER IF EXISTS trigger_update_api_keys_updated_at ON api_keys;
DROP TRIGGER IF EXISTS trigger_update_tenant_settings_updated_at ON tenant_settings;
DROP TRIGGER IF EXISTS trigger_update_system_settings_updated_at ON system_settings;
DROP TRIGGER IF EXISTS trigger_update_notification_preferences_updated_at ON notification_preferences;
DROP TRIGGER IF EXISTS trigger_update_notifications_updated_at ON notifications;
DROP TRIGGER IF EXISTS trigger_update_quota_usage_updated_at ON quota_usage;
DROP TRIGGER IF EXISTS trigger_update_ocr_jobs_updated_at ON ocr_jobs;
DROP TRIGGER IF EXISTS trigger_update_shared_folders_updated_at ON shared_folders;
DROP TRIGGER IF EXISTS trigger_update_public_share_links_updated_at ON public_share_links;
DROP TRIGGER IF EXISTS trigger_update_shared_documents_updated_at ON shared_documents;
DROP TRIGGER IF EXISTS trigger_update_tags_updated_at ON tags;
DROP TRIGGER IF EXISTS trigger_update_documents_updated_at ON documents;
DROP TRIGGER IF EXISTS trigger_update_folders_updated_at ON folders;
DROP TRIGGER IF EXISTS trigger_update_resource_permissions_updated_at ON resource_permissions;
DROP TRIGGER IF EXISTS trigger_update_user_roles_updated_at ON user_roles;
DROP TRIGGER IF EXISTS trigger_update_permissions_updated_at ON permissions;
DROP TRIGGER IF EXISTS trigger_update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS trigger_update_subscription_plans_updated_at ON subscription_plans;
DROP TRIGGER IF EXISTS trigger_update_tenant_users_updated_at ON tenant_users;
DROP TRIGGER IF EXISTS trigger_update_tenants_updated_at ON tenants;

-- Drop functions (in reverse order of creation)
DROP FUNCTION IF EXISTS update_share_link_view_count() CASCADE;
DROP FUNCTION IF EXISTS send_quota_warning_notification() CASCADE;
DROP FUNCTION IF EXISTS check_quota_before_document_insert() CASCADE;
DROP FUNCTION IF EXISTS update_quota_on_ocr_completion() CASCADE;
DROP FUNCTION IF EXISTS update_quota_on_document_change() CASCADE;
DROP FUNCTION IF EXISTS update_category_document_count() CASCADE;
DROP FUNCTION IF EXISTS update_tag_usage_count() CASCADE;
DROP FUNCTION IF EXISTS update_folder_path() CASCADE;
DROP FUNCTION IF EXISTS update_document_search_vector() CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- Completion message
DO $$
BEGIN
    RAISE NOTICE '✅ All triggers and functions removed successfully!';
END $$;
