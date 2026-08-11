export type Permission =
  | "dashboard:read"
  | "live:read"
  | "live:intervene"
  | "orders:read"
  | "orders:write"
  | "orders:cancel"
  | "orders:force_complete"
  | "orders:reassign"
  | "orders:refund"
  | "customers:read"
  | "customers:write"
  | "couriers:read"
  | "couriers:write"
  | "couriers:intervene"
  | "warehouses:read"
  | "warehouses:write"
  | "catalog:read"
  | "catalog:write"
  | "inventory:read"
  | "inventory:transfer"
  | "delivery:read"
  | "delivery:write"
  | "campaigns:read"
  | "campaigns:write"
  | "pricing:read"
  | "pricing:write"
  | "loyalty:read"
  | "loyalty:write"
  | "crm:read"
  | "crm:write"
  | "support:read"
  | "support:write"
  | "support:escalate"
  | "finance:read"
  | "finance:export"
  | "finance:payout:approve"
  | "finance:settlement"
  | "analytics:read"
  | "ai:read"
  | "ai:write"
  | "system:read"
  | "system:write"
  | "system:flags"
  | "rbac:read"
  | "rbac:write"
  | "audit:read"
  | "notifications:read"
  | "notifications:write"
  | "monitoring:read"
  | "reports:read"
  | "reports:export";

export type Role =
  | "viewer"
  | "city_ops"
  | "support_agent"
  | "support_lead"
  | "catalog_manager"
  | "finance_analyst"
  | "finance_admin"
  | "warehouse_ops"
  | "fraud_analyst"
  | "admin"
  | "super_admin";

const READ_OPS: Permission[] = [
  "dashboard:read",
  "live:read",
  "orders:read",
  "customers:read",
  "couriers:read",
  "warehouses:read",
  "catalog:read",
  "inventory:read",
  "delivery:read",
  "campaigns:read",
  "pricing:read",
  "loyalty:read",
  "crm:read",
  "support:read",
  "finance:read",
  "analytics:read",
  "ai:read",
  "notifications:read",
  "monitoring:read",
  "reports:read",
];

export const ROLE_PERMISSIONS: Record<Role, Permission[]> = {
  viewer: [...READ_OPS],
  city_ops: [
    ...READ_OPS,
    "live:intervene",
    "orders:write",
    "orders:reassign",
    "couriers:intervene",
    "notifications:write",
  ],
  support_agent: [
    ...READ_OPS,
    "support:write",
    "customers:write",
    "orders:refund",
  ],
  support_lead: [
    ...READ_OPS,
    "support:write",
    "support:escalate",
    "customers:write",
    "orders:cancel",
    "orders:refund",
  ],
  catalog_manager: [
    ...READ_OPS,
    "catalog:write",
    "campaigns:write",
    "pricing:write",
  ],
  finance_analyst: [...READ_OPS, "finance:export", "reports:export"],
  finance_admin: [
    ...READ_OPS,
    "finance:export",
    "finance:payout:approve",
    "finance:settlement",
    "orders:refund",
    "reports:export",
  ],
  warehouse_ops: [
    ...READ_OPS,
    "warehouses:write",
    "inventory:transfer",
  ],
  fraud_analyst: [...READ_OPS, "customers:read", "orders:read", "audit:read"],
  admin: [
    ...READ_OPS,
    "live:intervene",
    "orders:write",
    "orders:cancel",
    "orders:force_complete",
    "orders:reassign",
    "orders:refund",
    "customers:write",
    "couriers:write",
    "couriers:intervene",
    "warehouses:write",
    "catalog:write",
    "inventory:transfer",
    "delivery:write",
    "campaigns:write",
    "pricing:write",
    "loyalty:write",
    "crm:write",
    "support:write",
    "support:escalate",
    "finance:export",
    "ai:write",
    "system:read",
    "system:write",
    "rbac:read",
    "audit:read",
    "notifications:write",
    "reports:export",
  ],
  super_admin: [
    ...READ_OPS,
    "live:intervene",
    "orders:write",
    "orders:cancel",
    "orders:force_complete",
    "orders:reassign",
    "orders:refund",
    "customers:write",
    "couriers:write",
    "couriers:intervene",
    "warehouses:write",
    "catalog:write",
    "inventory:transfer",
    "delivery:write",
    "campaigns:write",
    "pricing:write",
    "loyalty:write",
    "crm:write",
    "support:write",
    "support:escalate",
    "finance:export",
    "finance:payout:approve",
    "finance:settlement",
    "ai:write",
    "system:read",
    "system:write",
    "system:flags",
    "rbac:read",
    "rbac:write",
    "audit:read",
    "notifications:write",
    "reports:export",
  ],
};

export interface SessionLike {
  roles: Role[];
  permissions?: string[];
}

export function permissionsForRoles(roles: Role[]): Permission[] {
  const set = new Set<Permission>();
  for (const role of roles) {
    for (const p of ROLE_PERMISSIONS[role] ?? []) {
      set.add(p);
    }
  }
  return [...set];
}

export function can(
  session: SessionLike | null | undefined,
  permission: Permission,
): boolean {
  if (!session) return false;
  if (session.permissions?.includes(permission)) return true;
  return session.roles.some((role) =>
    (ROLE_PERMISSIONS[role] ?? []).includes(permission),
  );
}
