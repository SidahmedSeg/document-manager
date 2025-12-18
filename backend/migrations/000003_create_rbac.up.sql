-- =============================================================================
-- Migration: 000003_create_rbac
-- Description: Create Role-Based Access Control (RBAC) tables
-- =============================================================================

-- =============================================================================
-- ROLES TABLE
-- =============================================================================

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Role details
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,

    -- Scope
    is_system BOOLEAN NOT NULL DEFAULT false,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE(tenant_id, slug),
    CHECK (is_system = true OR tenant_id IS NOT NULL)
);

-- Indexes for roles
CREATE INDEX idx_roles_slug ON roles(slug);
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_is_system ON roles(is_system);

-- Comments
COMMENT ON TABLE roles IS 'Roles for RBAC system (system-wide and tenant-specific)';
COMMENT ON COLUMN roles.is_system IS 'System roles cannot be deleted (e.g., admin, user)';
COMMENT ON COLUMN roles.tenant_id IS 'NULL for system roles, set for custom tenant roles';

-- =============================================================================
-- PERMISSIONS TABLE
-- =============================================================================

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Permission details
    name VARCHAR(100) NOT NULL,
    permission permission_type NOT NULL UNIQUE,
    description TEXT,

    -- Grouping
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for permissions
CREATE INDEX idx_permissions_permission ON permissions(permission);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);

-- Comments
COMMENT ON TABLE permissions IS 'Available permissions in the system';
COMMENT ON COLUMN permissions.resource IS 'Resource type (e.g., documents, sharing, settings)';
COMMENT ON COLUMN permissions.action IS 'Action type (e.g., create, read, update, delete)';

-- =============================================================================
-- ROLE PERMISSIONS (Many-to-Many)
-- =============================================================================

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Primary key
    PRIMARY KEY (role_id, permission_id)
);

-- Indexes for role_permissions
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Comments
COMMENT ON TABLE role_permissions IS 'Permissions assigned to roles';

-- =============================================================================
-- USER ROLES (Many-to-Many)
-- =============================================================================

CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- References tenant_users.identity_id or tenants.identity_id
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE(tenant_id, user_id, role_id)
);

-- Indexes for user_roles
CREATE INDEX idx_user_roles_tenant_id ON user_roles(tenant_id);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_tenant_user ON user_roles(tenant_id, user_id);

-- Comments
COMMENT ON TABLE user_roles IS 'Roles assigned to users';
COMMENT ON COLUMN user_roles.user_id IS 'Identity ID from Kratos (can be tenant owner or additional user)';

-- =============================================================================
-- RESOURCE PERMISSIONS (Document-level permissions)
-- =============================================================================

CREATE TABLE resource_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Resource
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    resource_type VARCHAR(100) NOT NULL, -- 'document', 'folder', etc.
    resource_id UUID NOT NULL,

    -- User/Role
    user_id UUID, -- Specific user permission
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE, -- Role-based permission

    -- Permission level
    can_view BOOLEAN NOT NULL DEFAULT false,
    can_edit BOOLEAN NOT NULL DEFAULT false,
    can_delete BOOLEAN NOT NULL DEFAULT false,
    can_share BOOLEAN NOT NULL DEFAULT false,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CHECK (user_id IS NOT NULL OR role_id IS NOT NULL)
);

-- Indexes for resource_permissions
CREATE INDEX idx_resource_permissions_tenant_id ON resource_permissions(tenant_id);
CREATE INDEX idx_resource_permissions_resource ON resource_permissions(resource_type, resource_id);
CREATE INDEX idx_resource_permissions_user_id ON resource_permissions(user_id);
CREATE INDEX idx_resource_permissions_role_id ON resource_permissions(role_id);

-- Comments
COMMENT ON TABLE resource_permissions IS 'Granular permissions for specific resources';
COMMENT ON COLUMN resource_permissions.user_id IS 'Direct user permission (takes precedence over role)';
COMMENT ON COLUMN resource_permissions.role_id IS 'Role-based permission';
