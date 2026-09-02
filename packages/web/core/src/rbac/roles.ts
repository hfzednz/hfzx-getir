/** Identity-service seeded platform roles (019_seed_roles_permissions.sql). */
export const PLATFORM_ROLES = [
  "customer",
  "courier",
  "picker",
  "packer",
  "dispatcher",
  "city_ops",
  "support_agent",
  "finance_analyst",
  "admin",
  "super_admin",
  "service_account",
  "partner",
  "supplier",
  "merchant",
] as const;

export type PlatformRole = (typeof PLATFORM_ROLES)[number];

export function hasRole(roles: string[], required: PlatformRole | PlatformRole[]): boolean {
  const needed = Array.isArray(required) ? required : [required];
  return needed.some((r) => roles.includes(r));
}
