DELETE FROM role_permissions WHERE role IN ('user', 'merchant') AND permission_id IN (SELECT id FROM permissions WHERE name = 'user.read');
