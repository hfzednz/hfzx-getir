import type { WebSession } from "@nexora/web-core";
import type { AdminSession } from "@/shared/auth/session";
import {
  permissionsForRoles,
  type Role,
} from "@/shared/permissions/permissions";

const IDENTITY_TO_ADMIN: Record<string, Role> = {
  admin: "admin",
  super_admin: "super_admin",
  city_ops: "city_ops",
  support_agent: "support_agent",
  finance_analyst: "finance_analyst",
  dispatcher: "warehouse_ops",
  picker: "warehouse_ops",
  packer: "warehouse_ops",
};

export function adminSessionFromOtp(web: WebSession, phone: string): AdminSession {
  const roles = Array.from(
    new Set(web.roles.map((r) => IDENTITY_TO_ADMIN[r]).filter(Boolean) as Role[]),
  );
  if (roles.length === 0) {
    // An identity principal with no admin-side role must not be granted a fallback
    // read-only console session.
    throw new Error("This account is not allowed to use the admin console.");
  }
  return {
    userId: web.principalId,
    email: phone,
    displayName: phone,
    roles,
    permissions: permissionsForRoles(roles),
    cityIds: [],
    mfaVerified: false,
    accessToken: web.accessToken,
  };
}
