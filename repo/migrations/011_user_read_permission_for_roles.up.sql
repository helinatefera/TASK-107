-- Grant user.read to user and merchant roles so they can access their own profiles
INSERT INTO role_permissions (role, permission_id)
SELECT 'user', id FROM permissions WHERE name = 'user.read'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'merchant', id FROM permissions WHERE name = 'user.read'
ON CONFLICT DO NOTHING;
