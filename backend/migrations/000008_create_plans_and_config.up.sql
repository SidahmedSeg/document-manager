-- =============================================================================
-- Migration: 000008_create_plans_and_config
-- Description: Create system configuration and settings tables
-- =============================================================================

-- =============================================================================
-- SYSTEM SETTINGS TABLE
-- =============================================================================

CREATE TABLE system_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Setting details
    key VARCHAR(255) NOT NULL UNIQUE,
    value JSONB NOT NULL,
    description TEXT,

    -- Metadata
    is_public BOOLEAN NOT NULL DEFAULT false,
    is_encrypted BOOLEAN NOT NULL DEFAULT false,
    category VARCHAR(100), -- 'feature_flags', 'limits', 'integrations', etc.

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for system_settings
CREATE INDEX idx_system_settings_key ON system_settings(key);
CREATE INDEX idx_system_settings_category ON system_settings(category);
CREATE INDEX idx_system_settings_is_public ON system_settings(is_public);

-- Comments
COMMENT ON TABLE system_settings IS 'Global system configuration and feature flags';
COMMENT ON COLUMN system_settings.is_public IS 'Whether setting can be exposed to frontend';
COMMENT ON COLUMN system_settings.is_encrypted IS 'Whether value is encrypted at rest';

-- =============================================================================
-- TENANT SETTINGS TABLE
-- =============================================================================

CREATE TABLE tenant_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Settings
    key VARCHAR(255) NOT NULL,
    value JSONB NOT NULL,
    description TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE(tenant_id, key)
);

-- Indexes for tenant_settings
CREATE INDEX idx_tenant_settings_tenant_id ON tenant_settings(tenant_id);
CREATE INDEX idx_tenant_settings_key ON tenant_settings(key);

-- Comments
COMMENT ON TABLE tenant_settings IS 'Per-tenant configuration settings';

-- =============================================================================
-- API KEYS TABLE (for programmatic access)
-- =============================================================================

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_by UUID NOT NULL, -- Identity ID from Kratos

    -- Key details
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE, -- Bcrypt hash of API key
    key_prefix VARCHAR(20) NOT NULL, -- First few chars for identification (e.g., "sk_live_abc123")

    -- Permissions
    scopes JSONB NOT NULL DEFAULT '[]', -- Array of allowed scopes

    -- Rate limiting
    rate_limit_per_hour INTEGER, -- NULL = no limit
    rate_limit_per_day INTEGER, -- NULL = no limit

    -- Usage tracking
    last_used_at TIMESTAMPTZ,
    usage_count BIGINT NOT NULL DEFAULT 0,

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- Indexes for api_keys
CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX idx_api_keys_created_by ON api_keys(created_by);
CREATE INDEX idx_api_keys_last_used_at ON api_keys(last_used_at);

-- Comments
COMMENT ON TABLE api_keys IS 'API keys for programmatic access';
COMMENT ON COLUMN api_keys.key_hash IS 'Bcrypt hash of the API key';
COMMENT ON COLUMN api_keys.key_prefix IS 'First few characters of key for identification';
COMMENT ON COLUMN api_keys.scopes IS 'Array of allowed scopes (e.g., ["documents.read", "documents.write"])';

-- =============================================================================
-- FEATURE FLAGS TABLE
-- =============================================================================

CREATE TABLE feature_flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Feature details
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,

    -- Status
    is_enabled BOOLEAN NOT NULL DEFAULT false,
    is_beta BOOLEAN NOT NULL DEFAULT false,

    -- Rollout control
    rollout_percentage INTEGER NOT NULL DEFAULT 0, -- 0-100
    allowed_plans JSONB DEFAULT '[]', -- Array of subscription_plan values
    allowed_tenants JSONB DEFAULT '[]', -- Array of tenant IDs (for targeted rollout)

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100)
);

-- Indexes for feature_flags
CREATE INDEX idx_feature_flags_slug ON feature_flags(slug);
CREATE INDEX idx_feature_flags_is_enabled ON feature_flags(is_enabled);
CREATE INDEX idx_feature_flags_is_beta ON feature_flags(is_beta);

-- JSONB indexes
CREATE INDEX idx_feature_flags_allowed_plans ON feature_flags USING GIN(allowed_plans);
CREATE INDEX idx_feature_flags_allowed_tenants ON feature_flags USING GIN(allowed_tenants);

-- Comments
COMMENT ON TABLE feature_flags IS 'Feature flags for gradual rollout and A/B testing';
COMMENT ON COLUMN feature_flags.rollout_percentage IS 'Percentage of users to enable feature for (0-100)';
COMMENT ON COLUMN feature_flags.allowed_plans IS 'Array of plans that can access this feature';
COMMENT ON COLUMN feature_flags.allowed_tenants IS 'Array of specific tenant IDs for targeted rollout';

-- =============================================================================
-- CATEGORIES TABLE (for document categorization)
-- =============================================================================

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Category details
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(50), -- Icon name or emoji

    -- Hierarchy
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,

    -- Visibility
    is_system BOOLEAN NOT NULL DEFAULT false, -- System categories cannot be deleted
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Usage count (denormalized)
    document_count INTEGER NOT NULL DEFAULT 0,

    -- Display order
    sort_order INTEGER NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for categories
CREATE INDEX idx_categories_slug ON categories(slug);
CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_is_system ON categories(is_system);
CREATE INDEX idx_categories_is_active ON categories(is_active);
CREATE INDEX idx_categories_sort_order ON categories(sort_order);
CREATE INDEX idx_categories_document_count ON categories(document_count DESC);

-- Comments
COMMENT ON TABLE categories IS 'Document categories for auto-categorization';
COMMENT ON COLUMN categories.is_system IS 'System categories cannot be deleted or modified';
COMMENT ON COLUMN categories.document_count IS 'Number of documents in this category (updated by trigger)';

-- =============================================================================
-- INTEGRATIONS TABLE (for third-party integrations)
-- =============================================================================

CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_by UUID NOT NULL,

    -- Integration details
    provider VARCHAR(100) NOT NULL, -- 'google_drive', 'dropbox', 'onedrive', etc.
    name VARCHAR(255) NOT NULL,

    -- Authentication
    access_token_encrypted TEXT, -- Encrypted access token
    refresh_token_encrypted TEXT, -- Encrypted refresh token
    expires_at TIMESTAMPTZ,

    -- Configuration
    config JSONB DEFAULT '{}', -- Provider-specific configuration

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_sync_at TIMESTAMPTZ,
    last_sync_error TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE(tenant_id, provider)
);

-- Indexes for integrations
CREATE INDEX idx_integrations_tenant_id ON integrations(tenant_id);
CREATE INDEX idx_integrations_provider ON integrations(provider);
CREATE INDEX idx_integrations_is_active ON integrations(is_active);
CREATE INDEX idx_integrations_last_sync_at ON integrations(last_sync_at);

-- Comments
COMMENT ON TABLE integrations IS 'Third-party service integrations (future feature)';
COMMENT ON COLUMN integrations.access_token_encrypted IS 'Encrypted OAuth access token';
COMMENT ON COLUMN integrations.config IS 'Provider-specific settings (folders to sync, etc.)';
