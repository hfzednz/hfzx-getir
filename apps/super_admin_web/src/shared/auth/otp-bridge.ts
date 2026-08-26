import type { WebSession } from "@nexora/web-core";
import type { PlatformSession } from "@/shared/auth/session";
import {
  permissionsForRoles,
  type Role,
} from "@/shared/permissions/platform-permissions";

export function platformSessionFromOtp(web: WebSession, phone: string): PlatformSession {
  const role: Role = web.roles.includes("super_admin") ? "platform_owner" : "platform_viewer";
  const roles: Role[] = [role];
  return {
    userId: web.principalId,
    email: phone,
    displayName: phone,
    roles,
    permissions: permissionsForRoles(roles),
    mfaVerified: true,
    webauthnVerified: false,
    accessToken: web.accessToken,
  };
}
