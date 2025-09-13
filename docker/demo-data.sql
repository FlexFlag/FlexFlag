-- FlexFlag Demo Data
-- This creates demo users, projects, and feature flags for the live demo

-- Demo Users
INSERT INTO users (id, email, password, full_name, role, created_at) VALUES 
  (1, 'demo@flexflag.io', '$2a$10$rQZ3J9J5Q5Q5Q5Q5Q5Q5Qu.', 'Demo User', 'admin', NOW()),
  (2, 'guest@flexflag.io', '$2a$10$rQZ3J9J5Q5Q5Q5Q5Q5Q5Qu.', 'Guest User', 'member', NOW())
ON CONFLICT (email) DO NOTHING;

-- Demo Project
INSERT INTO projects (id, name, slug, description, created_by, created_at) VALUES 
  (1, 'FlexFlag Demo', 'demo-project', 'Interactive demo project showcasing FlexFlag capabilities', 1, NOW())
ON CONFLICT (slug) DO NOTHING;

-- Demo Environments
INSERT INTO environments (id, name, slug, project_id, created_at) VALUES 
  (1, 'Production', 'production', 1, NOW()),
  (2, 'Staging', 'staging', 1, NOW()),
  (3, 'Development', 'development', 1, NOW())
ON CONFLICT (slug, project_id) DO NOTHING;

-- Demo API Keys
INSERT INTO api_keys (id, name, key, project_id, environment, permissions, created_by, created_at) VALUES 
  (1, 'Demo API Key', 'ff_demo_abc123xyz789demo', 1, 'production', '["evaluation", "management"]', 1, NOW())
ON CONFLICT (key) DO NOTHING;

-- Demo Feature Flags
INSERT INTO flags (id, key, name, type, description, enabled, default_value, variations, targeting, project_id, environment, created_by, created_at) VALUES 
  (1, 'welcome-banner', 'Welcome Banner', 'boolean', 'Show welcome banner to new users', true, false, '{"enabled": true, "disabled": false}', '{"rules": [], "rollout": {"percentage": 100}}', 1, 'production', 1, NOW()),
  
  (2, 'new-dashboard', 'New Dashboard UI', 'boolean', 'Enable the new dashboard design', true, false, '{"enabled": true, "disabled": false}', '{"rules": [{"condition": "user.plan == \"premium\"", "value": true}], "rollout": {"percentage": 50}}', 1, 'production', 1, NOW()),
  
  (3, 'checkout-version', 'Checkout Flow Version', 'string', 'A/B test for checkout flow', true, 'v1', '{"v1": "Original Checkout", "v2": "Streamlined Checkout", "v3": "One-Click Checkout"}', '{"rules": [], "rollout": {"percentage": 100, "distribution": {"v1": 40, "v2": 40, "v3": 20}}}', 1, 'production', 1, NOW()),
  
  (4, 'max-items', 'Maximum Items', 'number', 'Maximum items per page', true, 10, '{"small": 5, "medium": 10, "large": 20}', '{"rules": [{"condition": "user.plan == \"enterprise\"", "value": 20}], "rollout": {"percentage": 100}}', 1, 'production', 1, NOW()),
  
  (5, 'theme-config', 'Theme Configuration', 'json', 'Dynamic theme settings', true, '{"theme": "light", "accent": "#3b82f6"}', '{"light": "{\"theme\": \"light\", \"accent\": \"#3b82f6\"}", "dark": "{\"theme\": \"dark\", \"accent\": \"#8b5cf6\"}", "auto": "{\"theme\": \"auto\", \"accent\": \"#10b981\"}"}', '{"rules": [], "rollout": {"percentage": 100}}', 1, 'production', 1, NOW())
ON CONFLICT (key, project_id, environment) DO NOTHING;

-- Demo User Segments
INSERT INTO segments (id, key, name, description, conditions, project_id, created_by, created_at) VALUES 
  (1, 'premium-users', 'Premium Users', 'Users with premium subscription', '[{"field": "user.plan", "operator": "equals", "value": "premium"}]', 1, 1, NOW()),
  
  (2, 'beta-testers', 'Beta Testers', 'Users enrolled in beta program', '[{"field": "user.beta", "operator": "equals", "value": true}]', 1, 1, NOW()),
  
  (3, 'enterprise-users', 'Enterprise Users', 'Enterprise plan customers', '[{"field": "user.plan", "operator": "equals", "value": "enterprise"}]', 1, 1, NOW())
ON CONFLICT (key, project_id) DO NOTHING;