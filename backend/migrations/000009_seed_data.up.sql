-- =============================================================================
-- Migration: 000009_seed_data
-- Description: Insert initial seed data for production
-- =============================================================================

-- =============================================================================
-- SEED SUBSCRIPTION PLANS
-- =============================================================================

INSERT INTO subscription_plans (plan_type, name, description, max_storage_gb, max_ocr_pages_monthly, max_users, price_monthly, price_yearly, features) VALUES
    (
        'free',
        'Free Plan',
        'Perfect for individual users getting started',
        5,
        50,
        1,
        0.00,
        0.00,
        '{
            "ocr": true,
            "categorization": true,
            "basic_sharing": true,
            "search": true,
            "mobile_app": false,
            "api_access": false,
            "priority_support": false,
            "custom_branding": false
        }'::jsonb
    ),
    (
        'pro',
        'Pro Plan',
        'For professionals and small teams',
        100,
        500,
        10,
        9.99,
        99.99,
        '{
            "ocr": true,
            "categorization": true,
            "basic_sharing": true,
            "advanced_sharing": true,
            "search": true,
            "mobile_app": true,
            "api_access": true,
            "priority_support": false,
            "custom_branding": false,
            "version_history": true,
            "advanced_search": true
        }'::jsonb
    ),
    (
        'enterprise',
        'Enterprise Plan',
        'For large organizations with advanced needs',
        1024,
        5000,
        999999,
        49.99,
        499.99,
        '{
            "ocr": true,
            "categorization": true,
            "basic_sharing": true,
            "advanced_sharing": true,
            "search": true,
            "mobile_app": true,
            "api_access": true,
            "priority_support": true,
            "custom_branding": true,
            "version_history": true,
            "advanced_search": true,
            "sso": true,
            "audit_logs": true,
            "dedicated_support": true,
            "sla": true
        }'::jsonb
    );

-- =============================================================================
-- SEED PERMISSIONS
-- =============================================================================

INSERT INTO permissions (name, permission, description, resource, action) VALUES
    -- Document permissions
    ('Create Documents', 'documents.create', 'Create new documents', 'documents', 'create'),
    ('Read Documents', 'documents.read', 'View documents', 'documents', 'read'),
    ('Update Documents', 'documents.update', 'Edit document metadata', 'documents', 'update'),
    ('Delete Documents', 'documents.delete', 'Delete documents', 'documents', 'delete'),
    ('Share Documents', 'documents.share', 'Share documents with others', 'documents', 'share'),

    -- Sharing permissions
    ('Manage Sharing', 'sharing.manage', 'Manage shared documents and links', 'sharing', 'manage'),

    -- Settings permissions
    ('Manage Settings', 'settings.manage', 'Manage account settings', 'settings', 'manage'),

    -- User management permissions
    ('Manage Users', 'users.manage', 'Manage users (Enterprise only)', 'users', 'manage'),

    -- Admin permissions
    ('Admin Access', 'admin.access', 'Full administrative access', 'admin', 'access');

-- =============================================================================
-- SEED SYSTEM ROLES
-- =============================================================================

INSERT INTO roles (name, slug, description, is_system, tenant_id) VALUES
    ('User', 'user', 'Standard user with basic permissions', true, NULL),
    ('Admin', 'admin', 'Administrator with full permissions', true, NULL),
    ('Guest', 'guest', 'Guest user with read-only access', true, NULL);

-- =============================================================================
-- ASSIGN PERMISSIONS TO ROLES
-- =============================================================================

-- User role permissions (basic access)
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'user'
AND p.permission IN (
    'documents.create',
    'documents.read',
    'documents.update',
    'documents.delete',
    'documents.share',
    'sharing.manage',
    'settings.manage'
);

-- Admin role permissions (full access)
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'admin';

-- Guest role permissions (read-only)
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'guest'
AND p.permission IN ('documents.read');

-- =============================================================================
-- SEED CATEGORIES
-- =============================================================================

INSERT INTO categories (name, slug, description, icon, parent_id, is_system, sort_order) VALUES
    -- Top-level categories
    ('Financial', 'financial', 'Financial documents, invoices, receipts', '💰', NULL, true, 1),
    ('Legal', 'legal', 'Contracts, agreements, legal documents', '⚖️', NULL, true, 2),
    ('Personal', 'personal', 'Personal documents and records', '👤', NULL, true, 3),
    ('Work', 'work', 'Work-related documents', '💼', NULL, true, 4),
    ('Medical', 'medical', 'Medical records and health documents', '🏥', NULL, true, 5),
    ('Education', 'education', 'Educational documents, certificates', '🎓', NULL, true, 6),
    ('Travel', 'travel', 'Travel documents, tickets, itineraries', '✈️', NULL, true, 7),
    ('Utilities', 'utilities', 'Utility bills, statements', '🔧', NULL, true, 8),
    ('Insurance', 'insurance', 'Insurance policies and claims', '🛡️', NULL, true, 9),
    ('Other', 'other', 'Uncategorized documents', '📄', NULL, true, 100);

-- Financial subcategories
INSERT INTO categories (name, slug, description, icon, parent_id, is_system, sort_order)
SELECT
    'Invoices', 'invoices', 'Customer and vendor invoices', '🧾',
    id, true, 1
FROM categories WHERE slug = 'financial';

INSERT INTO categories (name, slug, description, icon, parent_id, is_system, sort_order)
SELECT
    'Receipts', 'receipts', 'Purchase receipts', '🧾',
    id, true, 2
FROM categories WHERE slug = 'financial';

INSERT INTO categories (name, slug, description, icon, parent_id, is_system, sort_order)
SELECT
    'Tax Documents', 'tax-documents', 'Tax returns, W2s, 1099s', '📊',
    id, true, 3
FROM categories WHERE slug = 'financial';

INSERT INTO categories (name, slug, description, icon, parent_id, is_system, sort_order)
SELECT
    'Bank Statements', 'bank-statements', 'Bank and credit card statements', '🏦',
    id, true, 4
FROM categories WHERE slug = 'financial';

-- Legal subcategories
INSERT INTO categories (name, slug, description, icon, parent_id, is_system, sort_order)
SELECT
    'Contracts', 'contracts', 'Business and personal contracts', '📝',
    id, true, 1
FROM categories WHERE slug = 'legal';

INSERT INTO categories (name, slug, description, icon, parent_id, is_system, sort_order)
SELECT
    'Agreements', 'agreements', 'Legal agreements', '🤝',
    id, true, 2
FROM categories WHERE slug = 'legal';

-- =============================================================================
-- SEED SYSTEM SETTINGS
-- =============================================================================

INSERT INTO system_settings (key, value, description, is_public, category) VALUES
    -- Feature flags
    ('features.ocr.enabled', 'true'::jsonb, 'Enable OCR processing', true, 'feature_flags'),
    ('features.categorization.enabled', 'true'::jsonb, 'Enable auto-categorization', true, 'feature_flags'),
    ('features.sharing.enabled', 'true'::jsonb, 'Enable document sharing', true, 'feature_flags'),
    ('features.public_links.enabled', 'true'::jsonb, 'Enable public share links', true, 'feature_flags'),
    ('features.api_keys.enabled', 'false'::jsonb, 'Enable API key management', true, 'feature_flags'),

    -- System limits
    ('limits.max_file_size_mb', '100'::jsonb, 'Maximum file size in MB', true, 'limits'),
    ('limits.max_files_per_upload', '10'::jsonb, 'Maximum files per upload', true, 'limits'),
    ('limits.allowed_file_types', '["pdf", "doc", "docx", "txt", "jpg", "jpeg", "png"]'::jsonb, 'Allowed file types', true, 'limits'),

    -- Storage settings
    ('storage.default_bucket', '"documents"'::jsonb, 'Default MinIO bucket name', false, 'storage'),
    ('storage.retention_days', '30'::jsonb, 'Soft-delete retention period in days', false, 'storage'),

    -- OCR settings
    ('ocr.default_language', '"en"'::jsonb, 'Default OCR language', false, 'ocr'),
    ('ocr.max_concurrent_jobs', '5'::jsonb, 'Maximum concurrent OCR jobs', false, 'ocr'),
    ('ocr.job_timeout_minutes', '10'::jsonb, 'OCR job timeout in minutes', false, 'ocr'),

    -- Sharing settings
    ('sharing.default_link_expiry_days', '7'::jsonb, 'Default share link expiry in days', false, 'sharing'),
    ('sharing.max_link_views', '1000'::jsonb, 'Maximum views per share link', false, 'sharing'),

    -- Notification settings
    ('notifications.digest_enabled', 'true'::jsonb, 'Enable notification digests', false, 'notifications'),
    ('notifications.digest_time', '"09:00"'::jsonb, 'Default digest time (UTC)', false, 'notifications'),

    -- Security settings
    ('security.session_timeout_hours', '24'::jsonb, 'Session timeout in hours', false, 'security'),
    ('security.max_login_attempts', '5'::jsonb, 'Maximum login attempts before lockout', false, 'security'),
    ('security.lockout_duration_minutes', '30'::jsonb, 'Account lockout duration in minutes', false, 'security');

-- =============================================================================
-- SEED FEATURE FLAGS
-- =============================================================================

INSERT INTO feature_flags (name, slug, description, is_enabled, is_beta, rollout_percentage, allowed_plans) VALUES
    (
        'OCR Processing',
        'ocr-processing',
        'Enable OCR text extraction from documents',
        true,
        false,
        100,
        '["free", "pro", "enterprise"]'::jsonb
    ),
    (
        'Auto Categorization',
        'auto-categorization',
        'Automatically categorize documents using ML',
        true,
        false,
        100,
        '["free", "pro", "enterprise"]'::jsonb
    ),
    (
        'Advanced Search',
        'advanced-search',
        'Advanced search with filters and full-text search',
        true,
        false,
        100,
        '["pro", "enterprise"]'::jsonb
    ),
    (
        'API Access',
        'api-access',
        'Programmatic API access with API keys',
        false,
        true,
        0,
        '["pro", "enterprise"]'::jsonb
    ),
    (
        'Version History',
        'version-history',
        'Track and restore document versions',
        false,
        true,
        50,
        '["pro", "enterprise"]'::jsonb
    ),
    (
        'Webhooks',
        'webhooks',
        'Send webhooks for document events',
        false,
        true,
        0,
        '["enterprise"]'::jsonb
    ),
    (
        'SSO Integration',
        'sso-integration',
        'Single Sign-On with SAML/OAuth',
        false,
        true,
        0,
        '["enterprise"]'::jsonb
    );

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE '✅ Seed data inserted successfully!';
    RAISE NOTICE '   - 3 subscription plans';
    RAISE NOTICE '   - 9 permissions';
    RAISE NOTICE '   - 3 system roles';
    RAISE NOTICE '   - 16 categories (10 top-level + 6 subcategories)';
    RAISE NOTICE '   - 17 system settings';
    RAISE NOTICE '   - 7 feature flags';
END $$;
