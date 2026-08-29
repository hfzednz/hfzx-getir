"use client";

import { useEffect, type ReactNode } from "react";
import { useRouter, usePathname } from "next/navigation";
import type { PlatformRole } from "../rbac/roles";
import { hasRole } from "../rbac/roles";

export interface RouteGuardProps {
  session: { roles: string[] } | null;
  allow: PlatformRole | PlatformRole[];
  loginPath?: string;
  onDeny?: () => void;
  children: ReactNode;
}

/**
 * Client route guard. Backend authorization stays authoritative — this only keeps a
 * session that lacks the role from mounting the app shell and its queries.
 */
export function RouteGuard({
  session,
  allow,
  loginPath = "/login",
  onDeny,
  children,
}: RouteGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const allowed = session != null && hasRole(session.roles, allow);

  useEffect(() => {
    if (pathname === loginPath) return;
    if (!session) {
      router.replace(loginPath);
    }
  }, [session, loginPath, pathname, router]);

  if (pathname === loginPath) {
    return <>{children}</>;
  }

  if (!session) {
    return (
      <p className="p-6 text-sm text-neutral-500" role="status">
        Redirecting to sign in…
      </p>
    );
  }

  if (!allowed) {
    return (
      <div className="space-y-3 p-6" role="alert">
        <h1 className="text-lg font-semibold">No access</h1>
        <p className="text-sm text-neutral-600">
          This account does not have a role that can use this application.
        </p>
        <button
          type="button"
          className="rounded-lg border px-4 text-sm"
          style={{ minHeight: 44 }}
          onClick={() => {
            onDeny?.();
            router.replace(loginPath);
          }}
        >
          Sign in with another account
        </button>
      </div>
    );
  }

  return <>{children}</>;
}
