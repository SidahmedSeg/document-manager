-- =============================================================================
-- Migration: 000002_create_tenants_and_users
-- Description: Create tenants and users tables with subscription management
-- =============================================================================

-- =============================================================================
-- TENANTS TABLE
-- =============================================================================

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Identity (from Kratos)
    identity_id UUID NOT NULL UNIQUE,
    email CITEXT NOT NULL UNIQUE,
    full_name VARCHAR(255),

    -- Subscription
    subscription_plan subscription_plan NOT NULL DEFAULT 'free',
    subscription_status subscription_status NOT NULL DEFAULT 'active',
    subscription_start_date TIMESTAMPTZ,
    subscription_end_date TIMESTAMPTZ,
    trial_ends_at TIMESTAMPTZ,

    -- Billing
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes for tenants
CREATE INDEX idx_tenants_identity_id ON tenants(identity_id);
CREATE INDEX idx_tenants_email ON tenants(email);
CREATE INDEX idx_tenants_subscription_plan ON tenants(subscription_plan);
CREATE INDEX idx_tenants_subscription_status ON tenants(subscription_status);
CREATE INDEX idx_tenants_stripe_customer_id ON tenants(stripe_customer_id);
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at);
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);

-- Comments
COMMENT ON TABLE tenants IS 'Tenant/user accounts with subscription information';
COMMENT ON COLUMN tenants.identity_id IS 'User ID from Kratos identity system';
COMMENT ON COLUMN tenants.metadata IS 'Additional tenant metadata (preferences, settings, etc.)';

-- =============================================================================
-- TENANT USERS TABLE (for Enterprise plan with multiple users)
-- =============================================================================

CREATE TABLE tenant_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    identity_id UUID NOT NULL,
    email CITEXT NOT NULL,
    full_name VARCHAR(255),

    -- Access control
    is_owner BOOLEAN NOT NULL DEFAULT false,
    is_admin BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Invitation
    invited_by UUID REFERENCES tenant_users(id) ON DELETE SET NULL,
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Constraints
    UNIQUE(tenant_id, identity_id)
);

-- Indexes for tenant_users
CREATE INDEX idx_tenant_users_tenant_id ON tenant_users(tenant_id);
CREATE INDEX idx_tenant_users_identity_id ON tenant_users(identity_id);
CREATE INDEX idx_tenant_users_email ON tenant_users(email);
CREATE INDEX idx_tenant_users_is_owner ON tenant_users(is_owner);
CREATE INDEX idx_tenant_users_is_admin ON tenant_users(is_admin);
CREATE INDEX idx_tenant_users_is_active ON tenant_users(is_active);
CREATE INDEX idx_tenant_users_deleted_at ON tenant_users(deleted_at);

-- Comments
COMMENT ON TABLE tenant_users IS 'Additional users for Enterprise plan tenants';
COMMENT ON COLUMN tenant_users.is_owner IS 'Tenant owner (only one per tenant)';
COMMENT ON COLUMN tenant_users.is_admin IS 'Admin user with full permissions';

-- =============================================================================
-- SUBSCRIPTION PLANS TABLE (Reference data)
-- =============================================================================

CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Plan details
    plan_type subscription_plan NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- Limits
    max_storage_gb INTEGER NOT NULL,
    max_ocr_pages_monthly INTEGER NOT NULL,
    max_users INTEGER NOT NULL,
    max_documents INTEGER,

    -- Pricing
    price_monthly DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    price_yearly DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',

    -- Features (JSONB for flexibility)
    features JSONB DEFAULT '{}',

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for subscription_plans
CREATE INDEX idx_subscription_plans_plan_type ON subscription_plans(plan_type);
CREATE INDEX idx_subscription_plans_is_active ON subscription_plans(is_active);

-- Comments
COMMENT ON TABLE subscription_plans IS 'Available subscription plans and their limits';
COMMENT ON COLUMN subscription_plans.features IS 'JSON object with feature flags (e.g., {"ocr": true, "sharing": true})';

-- =============================================================================
-- TENANT INVITATIONS TABLE
-- =============================================================================

CREATE TABLE tenant_invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    invited_by UUID NOT NULL REFERENCES tenant_users(id) ON DELETE CASCADE,

    -- Invitation details
    email CITEXT NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    is_admin BOOLEAN NOT NULL DEFAULT false,

    -- Status
    accepted_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ NOT NULL,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE(tenant_id, email)
);

-- Indexes for tenant_invitations
CREATE INDEX idx_tenant_invitations_tenant_id ON tenant_invitations(tenant_id);
CREATE INDEX idx_tenant_invitations_email ON tenant_invitations(email);
CREATE INDEX idx_tenant_invitations_token ON tenant_invitations(token);
CREATE INDEX idx_tenant_invitations_expired_at ON tenant_invitations(expired_at);
CREATE INDEX idx_tenant_invitations_accepted_at ON tenant_invitations(accepted_at);

-- Comments
COMMENT ON TABLE tenant_invitations IS 'Pending invitations for Enterprise plan tenants';
COMMENT ON COLUMN tenant_invitations.token IS 'Unique invitation token sent via email';
