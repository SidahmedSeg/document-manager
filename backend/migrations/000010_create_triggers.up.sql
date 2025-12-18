-- =============================================================================
-- Migration: 000010_create_triggers
-- Description: Create database triggers for automation
-- =============================================================================

-- =============================================================================
-- FUNCTION: Update updated_at timestamp
-- =============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to all relevant tables
CREATE TRIGGER trigger_update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_tenant_users_updated_at
    BEFORE UPDATE ON tenant_users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_subscription_plans_updated_at
    BEFORE UPDATE ON subscription_plans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_permissions_updated_at
    BEFORE UPDATE ON permissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_user_roles_updated_at
    BEFORE UPDATE ON user_roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_resource_permissions_updated_at
    BEFORE UPDATE ON resource_permissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_folders_updated_at
    BEFORE UPDATE ON folders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_documents_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_tags_updated_at
    BEFORE UPDATE ON tags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_shared_documents_updated_at
    BEFORE UPDATE ON shared_documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_public_share_links_updated_at
    BEFORE UPDATE ON public_share_links
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_shared_folders_updated_at
    BEFORE UPDATE ON shared_folders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_ocr_jobs_updated_at
    BEFORE UPDATE ON ocr_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_quota_usage_updated_at
    BEFORE UPDATE ON quota_usage
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_system_settings_updated_at
    BEFORE UPDATE ON system_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_tenant_settings_updated_at
    BEFORE UPDATE ON tenant_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_api_keys_updated_at
    BEFORE UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_feature_flags_updated_at
    BEFORE UPDATE ON feature_flags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_integrations_updated_at
    BEFORE UPDATE ON integrations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- FUNCTION: Update document search vector
-- =============================================================================

CREATE OR REPLACE FUNCTION update_document_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.file_name, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_document_search_vector
    BEFORE INSERT OR UPDATE OF title, description, file_name
    ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_document_search_vector();

-- =============================================================================
-- FUNCTION: Update folder materialized path
-- =============================================================================

CREATE OR REPLACE FUNCTION update_folder_path()
RETURNS TRIGGER AS $$
DECLARE
    parent_path TEXT;
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.path := '/' || NEW.name;
    ELSE
        SELECT path INTO parent_path FROM folders WHERE id = NEW.parent_id;
        NEW.path := parent_path || '/' || NEW.name;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_folder_path
    BEFORE INSERT OR UPDATE OF name, parent_id
    ON folders
    FOR EACH ROW
    EXECUTE FUNCTION update_folder_path();

-- =============================================================================
-- FUNCTION: Update tag usage count
-- =============================================================================

CREATE OR REPLACE FUNCTION update_tag_usage_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE tags SET usage_count = usage_count + 1 WHERE id = NEW.tag_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE tags SET usage_count = usage_count - 1 WHERE id = OLD.tag_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_tag_usage_count_insert
    AFTER INSERT ON document_tags
    FOR EACH ROW
    EXECUTE FUNCTION update_tag_usage_count();

CREATE TRIGGER trigger_update_tag_usage_count_delete
    AFTER DELETE ON document_tags
    FOR EACH ROW
    EXECUTE FUNCTION update_tag_usage_count();

-- =============================================================================
-- FUNCTION: Update category document count
-- =============================================================================

CREATE OR REPLACE FUNCTION update_category_document_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        IF NEW.category IS NOT NULL THEN
            UPDATE categories SET document_count = document_count + 1
            WHERE slug = NEW.category;
        END IF;
        IF TG_OP = 'UPDATE' AND OLD.category IS NOT NULL AND OLD.category != NEW.category THEN
            UPDATE categories SET document_count = GREATEST(document_count - 1, 0)
            WHERE slug = OLD.category;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.category IS NOT NULL THEN
            UPDATE categories SET document_count = GREATEST(document_count - 1, 0)
            WHERE slug = OLD.category;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_category_document_count
    AFTER INSERT OR UPDATE OF category OR DELETE
    ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_category_document_count();

-- =============================================================================
-- FUNCTION: Update quota usage on document changes
-- =============================================================================

CREATE OR REPLACE FUNCTION update_quota_on_document_change()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Increment storage and document count
        UPDATE quota_usage
        SET
            storage_used_bytes = storage_used_bytes + NEW.file_size,
            document_count = document_count + 1
        WHERE tenant_id = NEW.tenant_id;

        -- Insert usage event
        INSERT INTO usage_events (tenant_id, user_id, event_type, resource_type, resource_id, storage_delta_bytes)
        VALUES (NEW.tenant_id, NEW.created_by, 'storage_added', 'document', NEW.id, NEW.file_size);

    ELSIF TG_OP = 'UPDATE' THEN
        -- Update storage if file size changed
        IF NEW.file_size != OLD.file_size THEN
            UPDATE quota_usage
            SET storage_used_bytes = storage_used_bytes + (NEW.file_size - OLD.file_size)
            WHERE tenant_id = NEW.tenant_id;

            INSERT INTO usage_events (tenant_id, user_id, event_type, resource_type, resource_id, storage_delta_bytes)
            VALUES (NEW.tenant_id, NEW.created_by, 'storage_modified', 'document', NEW.id, NEW.file_size - OLD.file_size);
        END IF;

    ELSIF TG_OP = 'DELETE' THEN
        -- Decrement storage and document count
        UPDATE quota_usage
        SET
            storage_used_bytes = GREATEST(storage_used_bytes - OLD.file_size, 0),
            document_count = GREATEST(document_count - 1, 0)
        WHERE tenant_id = OLD.tenant_id;

        INSERT INTO usage_events (tenant_id, user_id, event_type, resource_type, resource_id, storage_delta_bytes)
        VALUES (OLD.tenant_id, OLD.created_by, 'storage_removed', 'document', OLD.id, -OLD.file_size);
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_quota_on_document_insert
    AFTER INSERT ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_quota_on_document_change();

CREATE TRIGGER trigger_update_quota_on_document_update
    AFTER UPDATE OF file_size ON documents
    FOR EACH ROW
    WHEN (OLD.file_size IS DISTINCT FROM NEW.file_size)
    EXECUTE FUNCTION update_quota_on_document_change();

CREATE TRIGGER trigger_update_quota_on_document_delete
    AFTER DELETE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_quota_on_document_change();

-- =============================================================================
-- FUNCTION: Update quota usage on OCR job completion
-- =============================================================================

CREATE OR REPLACE FUNCTION update_quota_on_ocr_completion()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND (OLD.status IS NULL OR OLD.status != 'completed') THEN
        -- Increment OCR pages used
        UPDATE quota_usage
        SET ocr_pages_used = ocr_pages_used + COALESCE(NEW.page_count, 1)
        WHERE tenant_id = NEW.tenant_id;

        -- Insert usage event
        INSERT INTO usage_events (tenant_id, user_id, event_type, resource_type, resource_id, ocr_pages_delta)
        VALUES (NEW.tenant_id, NEW.requested_by, 'ocr_page_processed', 'ocr_job', NEW.id, COALESCE(NEW.page_count, 1));
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_quota_on_ocr_completion
    AFTER UPDATE OF status ON ocr_jobs
    FOR EACH ROW
    WHEN (NEW.status = 'completed')
    EXECUTE FUNCTION update_quota_on_ocr_completion();

-- =============================================================================
-- FUNCTION: Check quota before insert
-- =============================================================================

CREATE OR REPLACE FUNCTION check_quota_before_document_insert()
RETURNS TRIGGER AS $$
DECLARE
    v_quota_usage quota_usage%ROWTYPE;
BEGIN
    -- Get current quota usage
    SELECT * INTO v_quota_usage FROM quota_usage WHERE tenant_id = NEW.tenant_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Quota not found for tenant %', NEW.tenant_id;
    END IF;

    -- Check storage limit
    IF v_quota_usage.storage_used_bytes + NEW.file_size > v_quota_usage.storage_limit_bytes THEN
        RAISE EXCEPTION 'Storage quota exceeded. Used: % bytes, Limit: % bytes, Attempted: % bytes',
            v_quota_usage.storage_used_bytes,
            v_quota_usage.storage_limit_bytes,
            NEW.file_size;
    END IF;

    -- Check document count limit (if set)
    IF v_quota_usage.document_limit IS NOT NULL AND
       v_quota_usage.document_count >= v_quota_usage.document_limit THEN
        RAISE EXCEPTION 'Document count limit exceeded. Limit: %', v_quota_usage.document_limit;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_quota_before_document_insert
    BEFORE INSERT ON documents
    FOR EACH ROW
    EXECUTE FUNCTION check_quota_before_document_insert();

-- =============================================================================
-- FUNCTION: Send quota warning notifications
-- =============================================================================

CREATE OR REPLACE FUNCTION send_quota_warning_notification()
RETURNS TRIGGER AS $$
DECLARE
    v_storage_percent DECIMAL;
    v_ocr_percent DECIMAL;
BEGIN
    -- Calculate storage usage percentage
    v_storage_percent := (NEW.storage_used_bytes::DECIMAL / NEW.storage_limit_bytes::DECIMAL) * 100;

    -- Send storage warning at 80%
    IF v_storage_percent >= 80 AND NOT NEW.storage_warning_sent THEN
        INSERT INTO notifications (tenant_id, user_id, type, title, message, status)
        SELECT
            NEW.tenant_id,
            identity_id,
            'quota_warning',
            'Storage Quota Warning',
            format('You have used %s%% of your storage quota (%s GB / %s GB). Consider upgrading your plan or deleting unused documents.',
                ROUND(v_storage_percent, 0),
                ROUND(NEW.storage_used_bytes::DECIMAL / 1073741824, 2),
                ROUND(NEW.storage_limit_bytes::DECIMAL / 1073741824, 2)
            ),
            'pending'
        FROM tenants
        WHERE id = NEW.tenant_id;

        NEW.storage_warning_sent := true;
    END IF;

    -- Calculate OCR usage percentage
    v_ocr_percent := (NEW.ocr_pages_used::DECIMAL / NEW.ocr_pages_limit::DECIMAL) * 100;

    -- Send OCR warning at 80%
    IF v_ocr_percent >= 80 AND NOT NEW.ocr_warning_sent THEN
        INSERT INTO notifications (tenant_id, user_id, type, title, message, status)
        SELECT
            NEW.tenant_id,
            identity_id,
            'quota_warning',
            'OCR Quota Warning',
            format('You have used %s%% of your monthly OCR quota (%s / %s pages). Your quota will reset on %s.',
                ROUND(v_ocr_percent, 0),
                NEW.ocr_pages_used,
                NEW.ocr_pages_limit,
                TO_CHAR(NEW.ocr_reset_at, 'YYYY-MM-DD')
            ),
            'pending'
        FROM tenants
        WHERE id = NEW.tenant_id;

        NEW.ocr_warning_sent := true;
    END IF;

    -- Reset warning flags if usage drops below 80%
    IF v_storage_percent < 80 THEN
        NEW.storage_warning_sent := false;
    END IF;

    IF v_ocr_percent < 80 THEN
        NEW.ocr_warning_sent := false;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_send_quota_warning_notification
    BEFORE UPDATE ON quota_usage
    FOR EACH ROW
    EXECUTE FUNCTION send_quota_warning_notification();

-- =============================================================================
-- FUNCTION: Update share link view count
-- =============================================================================

CREATE OR REPLACE FUNCTION update_share_link_view_count()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.action = 'view' AND NEW.success = true THEN
        UPDATE public_share_links
        SET
            current_views = current_views + 1,
            last_accessed_at = NEW.accessed_at
        WHERE id = NEW.share_link_id;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_share_link_view_count
    AFTER INSERT ON share_link_access_log
    FOR EACH ROW
    WHEN (NEW.action = 'view' AND NEW.success = true)
    EXECUTE FUNCTION update_share_link_view_count();

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE '✅ Database triggers created successfully!';
    RAISE NOTICE '   - updated_at triggers on 21 tables';
    RAISE NOTICE '   - Document search vector trigger';
    RAISE NOTICE '   - Folder path trigger';
    RAISE NOTICE '   - Tag usage count triggers';
    RAISE NOTICE '   - Category document count trigger';
    RAISE NOTICE '   - Quota management triggers';
    RAISE NOTICE '   - Quota warning notification trigger';
    RAISE NOTICE '   - Share link view count trigger';
END $$;
