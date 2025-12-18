-- =============================================================================
-- Migration: 000009_seed_data (ROLLBACK)
-- Description: Remove all seed data
-- =============================================================================

-- Delete seed data in reverse order

-- Feature flags
DELETE FROM feature_flags WHERE slug IN (
    'ocr-processing',
    'auto-categorization',
    'advanced-search',
    'api-access',
    'version-history',
    'webhooks',
    'sso-integration'
);

-- System settings
DELETE FROM system_settings WHERE category IN (
    'feature_flags',
    'limits',
    'storage',
    'ocr',
    'sharing',
    'notifications',
    'security'
);

-- Categories (cascade will delete subcategories)
DELETE FROM categories WHERE is_system = true;

-- Role permissions
DELETE FROM role_permissions WHERE role_id IN (
    SELECT id FROM roles WHERE is_system = true
);

-- System roles
DELETE FROM roles WHERE is_system = true;

-- Permissions
DELETE FROM permissions;

-- Subscription plans
DELETE FROM subscription_plans;

-- Completion message
DO $$
BEGIN
    RAISE NOTICE '✅ Seed data removed successfully!';
END $$;
