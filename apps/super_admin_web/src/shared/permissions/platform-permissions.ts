export type Permission =
  | "dashboard:read"
  | "tenants:read"
  | "tenants:write"
  | "tenants:suspend"
  | "tenants:delete"
  | "companies:read"
  | "companies:write"
  | "countries:read"
  | "countries:write"
  | "org:read"
  | "org:write"
  | "roles:read"
  | "roles:write"
  | "config:read"
  | "config:write"
  | "flags:read"
  | "flags:write"
  | "flags:kill"
  | "licenses:read"
  | "licenses:write"
  | "billing:read"
  | "billing:write"
  | "security:read"
  | "security:write"
  | "compliance:read"
  | "compliance:write"
  | "compliance:export"
  | "infra:read"
  | "infra:write"
  | "databases:read"
  | "databases:write"
  | "gateway:read"
  | "gateway:write"
  | "messaging:read"
  | "messaging:write"
  | "observability:read"
  | "ai_platform:read"
  | "ai_platform:write"
  | "analytics:read"
  | "dr:read"
  | "dr:execute"
  | "deployments:read"
  | "deployments:write"
  | "deployments:rollback"
  | "monitoring:read"
  | "notifications:read"
  | "notifications:write"
  | "audit:read"
  | "reports:read"
  | "reports:export"
  | "dual_control:approve";

export type Role =
  | "platform_owner"
  | "platform_security"
  | "platform_sre"
  | "platform_finops"
  | "platform_compliance"
  | "platform_viewer";

export const PLATFORM_ROLES: Role[] = [
  "platform_owner",
  "platform_security",
  "platform_sre",
  "platform_finops",
  "platform_compliance",
  "platform_viewer",
];

const READ_ALL: Permission[] = [
  "dashboard:read",
  "tenants:read",
  "companies:read",
  "countries:read",
  "org:read",
  "roles:read",
  "config:read",
  "flags:read",
  "licenses:read",
  "billing:read",
  "security:read",
  "compliance:read",
  "infra:read",
  "databases:read",
  "gateway:read",
  "messaging:read",
  "observability:read",
  "ai_platform:read",
  "analytics:read",
  "dr:read",
  "deployments:read",
  "monitoring:read",
  "notifications:read",
  "audit:read",
  "reports:read",
];

export const ROLE_PERMISSIONS: Record<Role, Permission[]> = {
  platform_viewer: [...READ_ALL],
  platform_security: [
    ...READ_ALL,
    "security:write",
    "flags:write",
    "flags:kill",
    "audit:read",
    "compliance:read",
    "dual_control:approve",
  ],
  platform_sre: [
    ...READ_ALL,
    "infra:write",
    "databases:write",
    "gateway:write",
    "messaging:write",
    "deployments:write",
    "deployments:rollback",
    "dr:execute",
    "config:write",
    "flags:write",
    "notifications:write",
    "dual_control:approve",
  ],
  platform_finops: [
    ...READ_ALL,
    "billing:write",
    "licenses:write",
    "reports:export",
    "analytics:read",
  ],
  platform_compliance: [
    ...READ_ALL,
    "compliance:write",
    "compliance:export",
    "audit:read",
    "reports:export",
    "dual_control:approve",
  ],
  platform_owner: [
    ...READ_ALL,
    "tenants:write",
    "tenants:suspend",
    "tenants:delete",
    "companies:write",
    "countries:write",
    "org:write",
    "roles:write",
    "config:write",
    "flags:write",
    "flags:kill",
    "licenses:write",
    "billing:write",
    "security:write",
    "compliance:write",
    "compliance:export",
    "infra:write",
    "databases:write",
    "gateway:write",
    "messaging:write",
    "ai_platform:write",
    "dr:execute",
    "deployments:write",
    "deployments:rollback",
    "notifications:write",
    "reports:export",
    "dual_control:approve",
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

export function isPlatformRole(role: string): role is Role {
  return (PLATFORM_ROLES as string[]).includes(role);
}

export function hasPlatformRole(
  session: SessionLike | null | undefined,
): boolean {
  if (!session) return false;
  return session.roles.some((role) => isPlatformRole(role));
}
