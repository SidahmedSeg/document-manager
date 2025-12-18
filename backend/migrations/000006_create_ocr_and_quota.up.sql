-- =============================================================================
-- Migration: 000006_create_ocr_and_quota
-- Description: Create OCR jobs and quota tracking tables
-- =============================================================================

-- =============================================================================
-- OCR JOBS TABLE
-- =============================================================================

CREATE TABLE ocr_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    requested_by UUID NOT NULL, -- Identity ID from Kratos

    -- Job details
    status ocr_status NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 5, -- 1 (highest) to 10 (lowest)

    -- OCR results
    extracted_text TEXT,
    confidence_score DECIMAL(5,4), -- 0.0000 to 1.0000
    language VARCHAR(10), -- ISO 639-1 code (e.g., 'en', 'fr')
    page_count INTEGER,

    -- Processing details
    processing_started_at TIMESTAMPTZ,
    processing_completed_at TIMESTAMPTZ,
    processing_duration_ms INTEGER, -- Duration in milliseconds
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CHECK (priority >= 1 AND priority <= 10),
    CHECK (confidence_score IS NULL OR (confidence_score >= 0 AND confidence_score <= 1)),
    CHECK (retry_count >= 0),
    CHECK (page_count IS NULL OR page_count > 0)
);

-- Indexes for ocr_jobs
CREATE INDEX idx_ocr_jobs_document_id ON ocr_jobs(document_id);
CREATE INDEX idx_ocr_jobs_tenant_id ON ocr_jobs(tenant_id);
CREATE INDEX idx_ocr_jobs_status ON ocr_jobs(status);
CREATE INDEX idx_ocr_jobs_priority ON ocr_jobs(priority);
CREATE INDEX idx_ocr_jobs_created_at ON ocr_jobs(created_at DESC);
CREATE INDEX idx_ocr_jobs_processing_started_at ON ocr_jobs(processing_started_at);
CREATE INDEX idx_ocr_jobs_status_priority ON ocr_jobs(status, priority); -- Composite for job queue

-- Full-text search index for extracted text
CREATE INDEX idx_ocr_jobs_extracted_text ON ocr_jobs USING GIN(to_tsvector('english', extracted_text));

-- Comments
COMMENT ON TABLE ocr_jobs IS 'OCR processing jobs and results';
COMMENT ON COLUMN ocr_jobs.priority IS 'Job priority (1=highest, 10=lowest)';
COMMENT ON COLUMN ocr_jobs.processing_duration_ms IS 'Time taken to process OCR in milliseconds';
COMMENT ON COLUMN ocr_jobs.retry_count IS 'Number of times job has been retried';

-- =============================================================================
-- QUOTA USAGE TABLE
-- =============================================================================

CREATE TABLE quota_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Storage quota
    storage_used_bytes BIGINT NOT NULL DEFAULT 0,
    storage_limit_bytes BIGINT NOT NULL,

    -- OCR quota (monthly)
    ocr_pages_used INTEGER NOT NULL DEFAULT 0,
    ocr_pages_limit INTEGER NOT NULL,
    ocr_reset_at TIMESTAMPTZ NOT NULL, -- When OCR quota resets (monthly)

    -- Document count
    document_count INTEGER NOT NULL DEFAULT 0,
    document_limit INTEGER, -- NULL = unlimited

    -- User count (for Enterprise)
    user_count INTEGER NOT NULL DEFAULT 1,
    user_limit INTEGER NOT NULL,

    -- Warning flags
    storage_warning_sent BOOLEAN NOT NULL DEFAULT false,
    ocr_warning_sent BOOLEAN NOT NULL DEFAULT false,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CHECK (storage_used_bytes >= 0),
    CHECK (storage_limit_bytes > 0),
    CHECK (ocr_pages_used >= 0),
    CHECK (ocr_pages_limit > 0),
    CHECK (document_count >= 0),
    CHECK (user_count >= 0),
    CHECK (user_limit > 0),
    UNIQUE(tenant_id)
);

-- Indexes for quota_usage
CREATE INDEX idx_quota_usage_tenant_id ON quota_usage(tenant_id);
CREATE INDEX idx_quota_usage_storage_warning ON quota_usage(storage_warning_sent) WHERE storage_warning_sent = false;
CREATE INDEX idx_quota_usage_ocr_warning ON quota_usage(ocr_warning_sent) WHERE ocr_warning_sent = false;
CREATE INDEX idx_quota_usage_ocr_reset_at ON quota_usage(ocr_reset_at);

-- Comments
COMMENT ON TABLE quota_usage IS 'Real-time quota tracking per tenant';
COMMENT ON COLUMN quota_usage.ocr_reset_at IS 'When monthly OCR quota resets (usually first day of month)';
COMMENT ON COLUMN quota_usage.storage_warning_sent IS 'Whether 80% storage warning has been sent';
COMMENT ON COLUMN quota_usage.ocr_warning_sent IS 'Whether 80% OCR warning has been sent';

-- =============================================================================
-- QUOTA HISTORY TABLE (for analytics)
-- =============================================================================

CREATE TABLE quota_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Snapshot date
    snapshot_date DATE NOT NULL,

    -- Storage snapshot
    storage_used_bytes BIGINT NOT NULL,
    storage_limit_bytes BIGINT NOT NULL,

    -- OCR snapshot
    ocr_pages_used INTEGER NOT NULL,
    ocr_pages_limit INTEGER NOT NULL,

    -- Document snapshot
    document_count INTEGER NOT NULL,

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE(tenant_id, snapshot_date)
);

-- Indexes for quota_history
CREATE INDEX idx_quota_history_tenant_id ON quota_history(tenant_id);
CREATE INDEX idx_quota_history_snapshot_date ON quota_history(snapshot_date DESC);
CREATE INDEX idx_quota_history_tenant_date ON quota_history(tenant_id, snapshot_date DESC);

-- Comments
COMMENT ON TABLE quota_history IS 'Historical quota usage snapshots (for analytics and trends)';
COMMENT ON COLUMN quota_history.snapshot_date IS 'Date of snapshot (one per day)';

-- =============================================================================
-- USAGE EVENTS TABLE (for detailed tracking)
-- =============================================================================

CREATE TABLE usage_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- Identity ID from Kratos

    -- Event details
    event_type VARCHAR(50) NOT NULL, -- 'storage_added', 'storage_removed', 'ocr_page_processed'
    resource_type VARCHAR(50) NOT NULL, -- 'document', 'ocr_job'
    resource_id UUID,

    -- Metrics
    storage_delta_bytes BIGINT, -- Change in storage (positive or negative)
    ocr_pages_delta INTEGER, -- Change in OCR pages (positive only)

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for usage_events
CREATE INDEX idx_usage_events_tenant_id ON usage_events(tenant_id);
CREATE INDEX idx_usage_events_user_id ON usage_events(user_id);
CREATE INDEX idx_usage_events_event_type ON usage_events(event_type);
CREATE INDEX idx_usage_events_created_at ON usage_events(created_at DESC);
CREATE INDEX idx_usage_events_resource ON usage_events(resource_type, resource_id);

-- Partitioning hint for large datasets
COMMENT ON TABLE usage_events IS 'Detailed usage event log (consider partitioning by created_at for large datasets)';
COMMENT ON COLUMN usage_events.event_type IS 'Type of usage event';
COMMENT ON COLUMN usage_events.storage_delta_bytes IS 'Change in storage bytes (can be negative for deletions)';
