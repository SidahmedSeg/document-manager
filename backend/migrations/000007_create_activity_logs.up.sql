-- =============================================================================
-- Migration: 000007_create_activity_logs
-- Description: Create activity logs and notifications tables
-- =============================================================================

-- =============================================================================
-- ACTIVITY LOGS TABLE
-- =============================================================================

CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- Identity ID from Kratos

    -- Activity details
    action activity_action NOT NULL,
    resource_type VARCHAR(100) NOT NULL, -- 'document', 'folder', 'share', 'user', 'settings'
    resource_id UUID,

    -- Context
    description TEXT NOT NULL,
    ip_address INET,
    user_agent TEXT,

    -- Changes (for update actions)
    changes JSONB, -- {"field": {"old": "value1", "new": "value2"}}

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for activity_logs
CREATE INDEX idx_activity_logs_tenant_id ON activity_logs(tenant_id);
CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_action ON activity_logs(action);
CREATE INDEX idx_activity_logs_resource ON activity_logs(resource_type, resource_id);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX idx_activity_logs_tenant_created ON activity_logs(tenant_id, created_at DESC);

-- JSONB index for changes
CREATE INDEX idx_activity_logs_changes ON activity_logs USING GIN(changes);

-- Partitioning hint
COMMENT ON TABLE activity_logs IS 'Audit trail of all user activities (consider partitioning by created_at)';
COMMENT ON COLUMN activity_logs.changes IS 'JSON object with field changes for update actions';
COMMENT ON COLUMN activity_logs.description IS 'Human-readable description of the action';

-- =============================================================================
-- NOTIFICATIONS TABLE
-- =============================================================================

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- Identity ID from Kratos (recipient)

    -- Notification details
    type notification_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,

    -- Related resource
    resource_type VARCHAR(100),
    resource_id UUID,

    -- Delivery
    status notification_status NOT NULL DEFAULT 'pending',
    email_sent_at TIMESTAMPTZ,
    push_sent_at TIMESTAMPTZ,

    -- User interaction
    read_at TIMESTAMPTZ,
    clicked_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,

    -- Action button (optional)
    action_url TEXT,
    action_label VARCHAR(100),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for notifications
CREATE INDEX idx_notifications_tenant_id ON notifications(tenant_id);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_read_at ON notifications(read_at);
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, read_at) WHERE read_at IS NULL;
CREATE INDEX idx_notifications_user_created ON notifications(user_id, created_at DESC);

-- Comments
COMMENT ON TABLE notifications IS 'In-app and email notifications';
COMMENT ON COLUMN notifications.status IS 'Delivery status (pending, sent, failed)';
COMMENT ON COLUMN notifications.action_url IS 'Optional URL for notification action button';

-- =============================================================================
-- NOTIFICATION PREFERENCES TABLE
-- =============================================================================

CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- Identity ID from Kratos

    -- Email preferences
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    email_document_shared BOOLEAN NOT NULL DEFAULT true,
    email_share_accessed BOOLEAN NOT NULL DEFAULT false,
    email_quota_warning BOOLEAN NOT NULL DEFAULT true,
    email_quota_exceeded BOOLEAN NOT NULL DEFAULT true,
    email_plan_expired BOOLEAN NOT NULL DEFAULT true,
    email_system_announcements BOOLEAN NOT NULL DEFAULT true,

    -- In-app preferences
    inapp_enabled BOOLEAN NOT NULL DEFAULT true,
    inapp_document_shared BOOLEAN NOT NULL DEFAULT true,
    inapp_share_accessed BOOLEAN NOT NULL DEFAULT true,
    inapp_quota_warning BOOLEAN NOT NULL DEFAULT true,
    inapp_quota_exceeded BOOLEAN NOT NULL DEFAULT true,

    -- Digest settings
    digest_enabled BOOLEAN NOT NULL DEFAULT false,
    digest_frequency VARCHAR(20) NOT NULL DEFAULT 'daily', -- 'daily', 'weekly'
    digest_time TIME NOT NULL DEFAULT '09:00:00',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE(tenant_id, user_id)
);

-- Indexes for notification_preferences
CREATE INDEX idx_notification_preferences_tenant_id ON notification_preferences(tenant_id);
CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX idx_notification_preferences_digest ON notification_preferences(digest_enabled, digest_frequency);

-- Comments
COMMENT ON TABLE notification_preferences IS 'User notification preferences';
COMMENT ON COLUMN notification_preferences.digest_frequency IS 'Frequency of digest emails (daily, weekly)';
COMMENT ON COLUMN notification_preferences.digest_time IS 'Time of day to send digest (UTC)';

-- =============================================================================
-- WEBHOOK LOGS TABLE (for future webhook support)
-- =============================================================================

CREATE TABLE webhook_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Webhook details
    event_type VARCHAR(100) NOT NULL, -- 'document.created', 'document.deleted', etc.
    webhook_url TEXT NOT NULL,

    -- Request
    request_headers JSONB,
    request_body JSONB NOT NULL,

    -- Response
    response_status INTEGER,
    response_headers JSONB,
    response_body TEXT,
    response_time_ms INTEGER, -- Duration in milliseconds

    -- Status
    success BOOLEAN NOT NULL DEFAULT false,
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CHECK (retry_count >= 0),
    CHECK (response_status IS NULL OR (response_status >= 100 AND response_status < 600))
);

-- Indexes for webhook_logs
CREATE INDEX idx_webhook_logs_tenant_id ON webhook_logs(tenant_id);
CREATE INDEX idx_webhook_logs_event_type ON webhook_logs(event_type);
CREATE INDEX idx_webhook_logs_success ON webhook_logs(success);
CREATE INDEX idx_webhook_logs_created_at ON webhook_logs(created_at DESC);
CREATE INDEX idx_webhook_logs_webhook_url ON webhook_logs(webhook_url);

-- Comments
COMMENT ON TABLE webhook_logs IS 'Webhook delivery logs (for future webhook feature)';
COMMENT ON COLUMN webhook_logs.response_time_ms IS 'Time taken for webhook request in milliseconds';
