-- 019_seed_roles_permissions.sql
-- Default platform roles and permissions for NEXORA IAM.
-- Matches 011_roles_permissions.sql (resource+action permissions, name-keyed roles).

INSERT INTO permissions (resource, action, description) VALUES
  ('identity', 'self_read', 'Read own identity'),
  ('identity', 'self_write', 'Update own identity'),
  ('orders', 'read', 'Read orders'),
  ('orders', 'write', 'Write orders'),
  ('warehouse', 'pick', 'Pick tasks'),
  ('warehouse', 'pack', 'Pack stations'),
  ('warehouse', 'dispatch', 'Dispatch handoff'),
  ('admin', 'read', 'Admin read'),
  ('admin', 'write', 'Admin write'),
  ('platform', 'read', 'Platform read'),
  ('platform', 'write', 'Platform write'),
  ('platform', 'flags_kill', 'Kill switches'),
  ('finance', 'read', 'Finance read'),
  ('support', 'read', 'Support read'),
  ('support', 'write', 'Support write')
ON CONFLICT (resource, action) DO NOTHING;

INSERT INTO roles (tenant_id, name, kind, description) VALUES
  (NULL, 'customer', 'platform', 'End customer'),
  (NULL, 'courier', 'platform', 'Delivery courier'),
  (NULL, 'picker', 'platform', 'Warehouse picker'),
  (NULL, 'packer', 'platform', 'Warehouse packer'),
  (NULL, 'dispatcher', 'platform', 'Warehouse dispatcher'),
  (NULL, 'city_ops', 'platform', 'City operations'),
  (NULL, 'support_agent', 'platform', 'Support agent'),
  (NULL, 'finance_analyst', 'platform', 'Finance analyst'),
  (NULL, 'admin', 'platform', 'Company admin'),
  (NULL, 'super_admin', 'platform', 'Platform owner'),
  (NULL, 'service_account', 'platform', 'Machine identity'),
  (NULL, 'partner', 'platform', 'External partner'),
  (NULL, 'supplier', 'platform', 'Supplier portal')
ON CONFLICT ON CONSTRAINT uq_roles_tenant_name DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON (p.resource, p.action) IN (
  ('identity', 'self_read'), ('identity', 'self_write'), ('orders', 'read'), ('orders', 'write')
)
WHERE r.name = 'customer' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON (p.resource, p.action) IN (
  ('identity', 'self_read'), ('identity', 'self_write')
)
WHERE r.name = 'courier' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.resource = 'warehouse' AND p.action = 'pick'
WHERE r.name = 'picker' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.resource = 'warehouse' AND p.action = 'pack'
WHERE r.name = 'packer' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.resource = 'warehouse' AND p.action = 'dispatch'
WHERE r.name = 'dispatcher' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.resource = 'admin'
WHERE r.name = 'admin' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.resource = 'platform'
WHERE r.name = 'super_admin' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;
