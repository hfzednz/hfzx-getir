"use client";

import { useEffect, useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { PlatformShell } from "@/components/shell/platform-shell";
import { useAuthStore } from "@/shared/auth/auth-store";
import { hasPlatformRole } from "@/shared/permissions/platform-permissions";

export default function AppLayout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const logout = useAuthStore((s) => s.logout);
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    setHydrated(useAuthStore.persist.hasHydrated());
    return useAuthStore.persist.onFinishHydration(() => setHydrated(true));
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    if (!session || !hasPlatformRole(session)) {
      if (session && !hasPlatformRole(session)) {
        logout();
      }
      router.replace("/login");
    }
  }, [hydrated, session, router, logout]);

  if (!hydrated) {
    return (
      <div className="min-h-screen flex items-center justify-center text-[var(--nx-text-secondary)] text-[13px]">
        Loading session…
      </div>
    );
  }

  if (!session || !hasPlatformRole(session)) {
    return null;
  }

  return <PlatformShell>{children}</PlatformShell>;
}
