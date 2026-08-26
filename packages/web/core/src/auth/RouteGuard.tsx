"use client";

import { useEffect, type ReactNode } from "react";
import { useRouter, usePathname } from "next/navigation";
import type { PlatformRole } from "../rbac/roles";
import { hasRole } from "../rbac/roles";

export interface RouteGuardProps {
  session: { roles: string[] } | null;
  allow: PlatformRole | PlatformRole[];
  loginPath?: string;
  children: ReactNode;
}

/** Client route guard — backend authorization remains authoritative. */
export function RouteGuard({
  session,
  allow,
  loginPath = "/login",
  children,
}: RouteGuardProps) {
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (pathname === loginPath) return;
    if (!session) {
      router.replace(loginPath);
      return;
    }
    if (!hasRole(session.roles, allow)) {
      router.replace("/unauthorized");
    }
  }, [session, allow, loginPath, pathname, router]);

  if (pathname === loginPath) {
    return <>{children}</>;
  }
  if (!session || !hasRole(session.roles, allow)) {
    return null;
  }
  return <>{children}</>;
}
