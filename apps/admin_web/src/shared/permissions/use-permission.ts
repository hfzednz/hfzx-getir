"use client";

import { useAuthStore } from "@/shared/auth/auth-store";
import { can, type Permission } from "@/shared/permissions/permissions";

/** Soft-gate helper bound to the current admin session. */
export function usePermission(permission: Permission): boolean {
  const session = useAuthStore((s) => s.session);
  return can(session, permission);
}

export function useSessionPermissions(): {
  can: (permission: Permission) => boolean;
  session: ReturnType<typeof useAuthStore.getState>["session"];
} {
  const session = useAuthStore((s) => s.session);
  return {
    session,
    can: (permission) => can(session, permission),
  };
}
