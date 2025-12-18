-- =============================================================================
-- Migration: 000005_create_sharing
-- Description: Create document sharing and public link tables
-- =============================================================================

-- =============================================================================
-- SHARED DOCUMENTS TABLE
-- =============================================================================

CREATE TABLE shared_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    shared_by UUID NOT NULL, -- Identity ID from Kratos
    shared_with UUID NOT NULL, -- Identity ID from Kratos (recipient)

    -- Permissions
    can_view BOOLEAN NOT NULL DEFAULT true,
    can_download BOOLEAN NOT NULL DEFAULT false,
    can_edit BOOLEAN NOT NULL DEFAULT false,
    can_reshare BOOLEAN NOT NULL DEFAULT false,

    -- Expiration
    expires_at TIMESTAMPTZ,

    -- Metadata
    message TEXT,
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,

    -- Constraints
    UNIQUE(document_id, shared_with)
);

-- Indexes for shared_documents
CREATE INDEX idx_shared_documents_document_id ON shared_documents(document_id);
CREATE INDEX idx_shared_documents_tenant_id ON shared_documents(tenant_id);
CREATE INDEX idx_shared_documents_shared_by ON shared_documents(shared_by);
CREATE INDEX idx_shared_documents_shared_with ON shared_documents(shared_with);
CREATE INDEX idx_shared_documents_expires_at ON shared_documents(expires_at);
CREATE INDEX idx_shared_documents_revoked_at ON shared_documents(revoked_at);
CREATE INDEX idx_shared_documents_created_at ON shared_documents(created_at DESC);

-- Comments
COMMENT ON TABLE shared_documents IS 'Direct document sharing between users';
COMMENT ON COLUMN shared_documents.expires_at IS 'Share expires after this date (NULL = never expires)';
COMMENT ON COLUMN shared_documents.revoked_at IS 'When the share was revoked (NULL = active)';

-- =============================================================================
-- PUBLIC SHARE LINKS TABLE
-- =============================================================================

CREATE TABLE public_share_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_by UUID NOT NULL, -- Identity ID from Kratos

    -- Link details
    token VARCHAR(255) NOT NULL UNIQUE,
    link_type share_link_type NOT NULL DEFAULT 'view',

    -- Access control
    password_hash VARCHAR(255), -- Bcrypt hash if password-protected
    max_views INTEGER, -- NULL = unlimited
    current_views INTEGER NOT NULL DEFAULT 0,

    -- Expiration
    expires_at TIMESTAMPTZ,

    -- Metadata
    title VARCHAR(255),
    description TEXT,
    metadata JSONB DEFAULT '{}',

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ,

    -- Constraints
    CHECK (max_views IS NULL OR max_views > 0),
    CHECK (current_views >= 0)
);

-- Indexes for public_share_links
CREATE INDEX idx_public_share_links_document_id ON public_share_links(document_id);
CREATE INDEX idx_public_share_links_tenant_id ON public_share_links(tenant_id);
CREATE INDEX idx_public_share_links_created_by ON public_share_links(created_by);
CREATE INDEX idx_public_share_links_token ON public_share_links(token);
CREATE INDEX idx_public_share_links_is_active ON public_share_links(is_active);
CREATE INDEX idx_public_share_links_expires_at ON public_share_links(expires_at);
CREATE INDEX idx_public_share_links_created_at ON public_share_links(created_at DESC);
CREATE INDEX idx_public_share_links_last_accessed_at ON public_share_links(last_accessed_at DESC);

-- Comments
COMMENT ON TABLE public_share_links IS 'Public shareable links for documents';
COMMENT ON COLUMN public_share_links.token IS 'Unique URL-safe token for public access';
COMMENT ON COLUMN public_share_links.password_hash IS 'Optional password protection (bcrypt hash)';
COMMENT ON COLUMN public_share_links.max_views IS 'Maximum number of views allowed (NULL = unlimited)';

-- =============================================================================
-- SHARE LINK ACCESS LOG TABLE
-- =============================================================================

CREATE TABLE share_link_access_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    share_link_id UUID NOT NULL REFERENCES public_share_links(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Access details
    ip_address INET,
    user_agent TEXT,
    referer TEXT,

    -- Action
    action VARCHAR(50) NOT NULL, -- 'view', 'download', 'password_attempt'
    success BOOLEAN NOT NULL DEFAULT true,

    -- Geolocation (optional)
    country VARCHAR(2),
    city VARCHAR(100),

    -- Timestamp
    accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for share_link_access_log
CREATE INDEX idx_share_link_access_log_share_link_id ON share_link_access_log(share_link_id);
CREATE INDEX idx_share_link_access_log_tenant_id ON share_link_access_log(tenant_id);
CREATE INDEX idx_share_link_access_log_ip_address ON share_link_access_log(ip_address);
CREATE INDEX idx_share_link_access_log_accessed_at ON share_link_access_log(accessed_at DESC);
CREATE INDEX idx_share_link_access_log_action ON share_link_access_log(action);

-- Comments
COMMENT ON TABLE share_link_access_log IS 'Access log for public share links';
COMMENT ON COLUMN share_link_access_log.action IS 'Type of access (view, download, password_attempt)';

-- =============================================================================
-- SHARED FOLDERS TABLE (for future folder sharing support)
-- =============================================================================

CREATE TABLE shared_folders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    folder_id UUID NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    shared_by UUID NOT NULL,
    shared_with UUID NOT NULL,

    -- Permissions
    can_view BOOLEAN NOT NULL DEFAULT true,
    can_upload BOOLEAN NOT NULL DEFAULT false,
    can_edit BOOLEAN NOT NULL DEFAULT false,
    can_delete BOOLEAN NOT NULL DEFAULT false,

    -- Expiration
    expires_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,

    -- Constraints
    UNIQUE(folder_id, shared_with)
);

-- Indexes for shared_folders
CREATE INDEX idx_shared_folders_folder_id ON shared_folders(folder_id);
CREATE INDEX idx_shared_folders_tenant_id ON shared_folders(tenant_id);
CREATE INDEX idx_shared_folders_shared_by ON shared_folders(shared_by);
CREATE INDEX idx_shared_folders_shared_with ON shared_folders(shared_with);
CREATE INDEX idx_shared_folders_expires_at ON shared_folders(expires_at);
CREATE INDEX idx_shared_folders_revoked_at ON shared_folders(revoked_at);

-- Comments
COMMENT ON TABLE shared_folders IS 'Folder sharing between users';
COMMENT ON COLUMN shared_folders.can_upload IS 'Allow uploading documents to shared folder';
