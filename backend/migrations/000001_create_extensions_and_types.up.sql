-- =============================================================================
-- Migration: 000001_create_extensions_and_types
-- Description: Enable PostgreSQL extensions and create custom types
-- =============================================================================

-- Enable UUID extension for generating unique identifiers
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable full-text search
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Enable case-insensitive text matching
CREATE EXTENSION IF NOT EXISTS "citext";

-- =============================================================================
-- CUSTOM TYPES
-- =============================================================================

-- Document status enum
CREATE TYPE document_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed'
);

-- Document visibility enum
CREATE TYPE document_visibility AS ENUM (
    'private',
    'shared',
    'public'
);

-- OCR status enum
CREATE TYPE ocr_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed'
);

-- Share link type
CREATE TYPE share_link_type AS ENUM (
    'view',
    'download'
);

-- Subscription plan enum
CREATE TYPE subscription_plan AS ENUM (
    'free',
    'pro',
    'enterprise'
);

-- Subscription status enum
CREATE TYPE subscription_status AS ENUM (
    'active',
    'past_due',
    'canceled',
    'trialing'
);

-- Permission enum for RBAC
CREATE TYPE permission_type AS ENUM (
    'documents.create',
    'documents.read',
    'documents.update',
    'documents.delete',
    'documents.share',
    'sharing.manage',
    'settings.manage',
    'users.manage',
    'admin.access'
);

-- Activity log action types
CREATE TYPE activity_action AS ENUM (
    'document.created',
    'document.updated',
    'document.deleted',
    'document.viewed',
    'document.downloaded',
    'document.shared',
    'share.created',
    'share.revoked',
    'share.accessed',
    'user.login',
    'user.logout',
    'settings.updated',
    'plan.upgraded',
    'plan.downgraded'
);

-- Notification type
CREATE TYPE notification_type AS ENUM (
    'document_shared',
    'share_access',
    'quota_warning',
    'quota_exceeded',
    'plan_expired',
    'system_announcement'
);

-- Notification status
CREATE TYPE notification_status AS ENUM (
    'pending',
    'sent',
    'failed'
);

-- Create index on enum types for faster lookups
COMMENT ON TYPE document_status IS 'Document processing status';
COMMENT ON TYPE document_visibility IS 'Document visibility level';
COMMENT ON TYPE ocr_status IS 'OCR processing status';
COMMENT ON TYPE share_link_type IS 'Type of share link access';
COMMENT ON TYPE subscription_plan IS 'Available subscription plans';
COMMENT ON TYPE subscription_status IS 'Subscription status';
COMMENT ON TYPE permission_type IS 'Available permissions in the system';
COMMENT ON TYPE activity_action IS 'Types of activities logged in the system';
COMMENT ON TYPE notification_type IS 'Types of notifications';
COMMENT ON TYPE notification_status IS 'Notification delivery status';
