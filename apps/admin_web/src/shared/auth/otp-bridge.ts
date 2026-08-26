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
  const mapped = web.roles
    .map((r) => IDENTITY_TO_ADMIN[r])
    .filter(Boolean) as Role[];
  const roles: Role[] = mapped.length ? mapped : ["viewer"];
  return {
    userId: web.principalId,
    email: phone,
    displayName: phone,
    roles,
    permissions: permissionsForRoles(roles),
    cityIds: ["city_ist"],
    mfaVerified: true,
    accessToken: web.accessToken,
  };
}
