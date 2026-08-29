import type { WebSession } from "@nexora/web-core";
import type { PlatformSession } from "@/shared/auth/session";
import {
  permissionsForRoles,
  type Role,
} from "@/shared/permissions/platform-permissions";

export function platformSessionFromOtp(web: WebSession, phone: string): PlatformSession {
  if (!web.roles.includes("super_admin")) {
    throw new Error("Super admin role required");
  }
  const roles: Role[] = ["platform_owner"];
  return {
    userId: web.principalId,
    email: phone,
    displayName: phone,
    roles,
    permissions: permissionsForRoles(roles),
    mfaVerified: false,
    webauthnVerified: false,
    accessToken: web.accessToken,
  };
}
