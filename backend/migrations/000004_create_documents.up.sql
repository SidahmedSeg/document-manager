-- =============================================================================
-- Migration: 000004_create_documents
-- Description: Create documents, folders, and tags tables
-- =============================================================================

-- =============================================================================
-- FOLDERS TABLE
-- =============================================================================

CREATE TABLE folders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Ownership
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_by UUID NOT NULL, -- Identity ID from Kratos

    -- Folder details
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7), -- Hex color code

    -- Hierarchy
    parent_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    path TEXT, -- Materialized path for quick hierarchy queries

    -- Visibility
    visibility document_visibility NOT NULL DEFAULT 'private',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Constraints
    UNIQUE(tenant_id, parent_id, name)
);

-- Indexes for folders
CREATE INDEX idx_folders_tenant_id ON folders(tenant_id);
CREATE INDEX idx_folders_created_by ON folders(created_by);
CREATE INDEX idx_folders_parent_id ON folders(parent_id);
CREATE INDEX idx_folders_path ON folders(path);
CREATE INDEX idx_folders_deleted_at ON folders(deleted_at);
CREATE INDEX idx_folders_created_at ON folders(created_at DESC);

-- Comments
COMMENT ON TABLE folders IS 'Folder hierarchy for organizing documents';
COMMENT ON COLUMN folders.path IS 'Materialized path (e.g., /parent/child/grandchild)';

-- =============================================================================
-- DOCUMENTS TABLE
-- =============================================================================

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Ownership
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_by UUID NOT NULL, -- Identity ID from Kratos

    -- Document metadata
    title VARCHAR(255) NOT NULL,
    description TEXT,
    folder_id UUID REFERENCES folders(id) ON DELETE SET NULL,

    -- File information
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL, -- Size in bytes
    file_type VARCHAR(100) NOT NULL, -- MIME type
    file_extension VARCHAR(50),

    -- Storage
    storage_key VARCHAR(500) NOT NULL UNIQUE, -- MinIO object key
    storage_bucket VARCHAR(255) NOT NULL DEFAULT 'documents',

    -- Processing status
    status document_status NOT NULL DEFAULT 'pending',
    processing_error TEXT,

    -- Visibility and sharing
    visibility document_visibility NOT NULL DEFAULT 'private',

    -- Categorization (from ML service)
    category VARCHAR(100),
    confidence_score DECIMAL(5,4), -- 0.0000 to 1.0000
    auto_categorized_at TIMESTAMPTZ,

    -- Full-text search
    search_vector tsvector,

    -- Metadata (flexible JSON)
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Constraints
    CHECK (file_size >= 0),
    CHECK (confidence_score IS NULL OR (confidence_score >= 0 AND confidence_score <= 1))
);

-- Indexes for documents
CREATE INDEX idx_documents_tenant_id ON documents(tenant_id);
CREATE INDEX idx_documents_created_by ON documents(created_by);
CREATE INDEX idx_documents_folder_id ON documents(folder_id);
CREATE INDEX idx_documents_storage_key ON documents(storage_key);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_visibility ON documents(visibility);
CREATE INDEX idx_documents_category ON documents(category);
CREATE INDEX idx_documents_file_type ON documents(file_type);
CREATE INDEX idx_documents_deleted_at ON documents(deleted_at);
CREATE INDEX idx_documents_created_at ON documents(created_at DESC);
CREATE INDEX idx_documents_updated_at ON documents(updated_at DESC);

-- Full-text search index
CREATE INDEX idx_documents_search_vector ON documents USING GIN(search_vector);

-- JSONB index for metadata queries
CREATE INDEX idx_documents_metadata ON documents USING GIN(metadata);

-- Comments
COMMENT ON TABLE documents IS 'User documents with metadata and processing status';
COMMENT ON COLUMN documents.storage_key IS 'MinIO object key (tenant_id/document_id/filename)';
COMMENT ON COLUMN documents.search_vector IS 'Full-text search vector (auto-updated by trigger)';
COMMENT ON COLUMN documents.metadata IS 'Additional metadata (custom fields, tags, etc.)';

-- =============================================================================
-- TAGS TABLE
-- =============================================================================

CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Ownership
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_by UUID NOT NULL,

    -- Tag details
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7), -- Hex color code
    description TEXT,

    -- Usage count (denormalized for performance)
    usage_count INTEGER NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE(tenant_id, name)
);

-- Indexes for tags
CREATE INDEX idx_tags_tenant_id ON tags(tenant_id);
CREATE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_tags_usage_count ON tags(usage_count DESC);

-- Comments
COMMENT ON TABLE tags IS 'User-defined tags for documents';
COMMENT ON COLUMN tags.usage_count IS 'Number of documents using this tag (updated by trigger)';

-- =============================================================================
-- DOCUMENT TAGS (Many-to-Many)
-- =============================================================================

CREATE TABLE document_tags (
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Primary key
    PRIMARY KEY (document_id, tag_id)
);

-- Indexes for document_tags
CREATE INDEX idx_document_tags_document_id ON document_tags(document_id);
CREATE INDEX idx_document_tags_tag_id ON document_tags(tag_id);

-- Comments
COMMENT ON TABLE document_tags IS 'Tags assigned to documents';

-- =============================================================================
-- DOCUMENT VERSIONS TABLE (for future versioning support)
-- =============================================================================

CREATE TABLE document_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Relationships
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_by UUID NOT NULL,

    -- Version details
    version_number INTEGER NOT NULL,
    comment TEXT,

    -- File information (at this version)
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR(100) NOT NULL,
    storage_key VARCHAR(500) NOT NULL,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE(document_id, version_number),
    CHECK (version_number > 0),
    CHECK (file_size >= 0)
);

-- Indexes for document_versions
CREATE INDEX idx_document_versions_document_id ON document_versions(document_id);
CREATE INDEX idx_document_versions_tenant_id ON document_versions(tenant_id);
CREATE INDEX idx_document_versions_version_number ON document_versions(document_id, version_number DESC);
CREATE INDEX idx_document_versions_created_at ON document_versions(created_at DESC);

-- Comments
COMMENT ON TABLE document_versions IS 'Document version history';
COMMENT ON COLUMN document_versions.version_number IS 'Sequential version number (1, 2, 3, ...)';
