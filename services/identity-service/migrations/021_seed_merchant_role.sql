-- 021_seed_merchant_role.sql
-- Distinct marketplace merchant role; does not replace supplier/partner.

INSERT INTO roles (tenant_id, name, kind, description) VALUES
  (NULL, 'merchant', 'platform', 'Marketplace merchant')
ON CONFLICT ON CONSTRAINT uq_roles_tenant_name DO NOTHING;

INSERT INTO permissions (resource, action, description) VALUES
  ('merchant', 'read', 'Merchant catalog and orders read'),
  ('merchant', 'write', 'Merchant listing and order writes')
ON CONFLICT (resource, action) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON (p.resource, p.action) IN (
  ('identity', 'self_read'), ('identity', 'self_write'),
  ('orders', 'read'), ('merchant', 'read'), ('merchant', 'write')
)
WHERE r.name = 'merchant' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;
