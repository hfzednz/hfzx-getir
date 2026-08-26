"use client";

import type { ReactNode } from "react";
import type { PlatformRole } from "./roles";
import { hasRole } from "./roles";

interface RoleGateProps {
  roles: string[];
  allow: PlatformRole | PlatformRole[];
  fallback?: ReactNode;
  children: ReactNode;
}

/** UI-only gate; backend authorization remains authoritative. */
export function RoleGate({ roles, allow, fallback = null, children }: RoleGateProps) {
  if (!hasRole(roles, allow)) {
    return <>{fallback}</>;
  }
  return <>{children}</>;
}
