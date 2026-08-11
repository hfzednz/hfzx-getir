import type { ReactNode } from "react";

export interface PermissionGateProps {
  /** When true, children render; otherwise fallback (or nothing). */
  allowed: boolean;
  children: ReactNode;
  fallback?: ReactNode;
}

export function PermissionGate({
  allowed,
  children,
  fallback = null,
}: PermissionGateProps) {
  if (!allowed) return <>{fallback}</>;
  return <>{children}</>;
}
