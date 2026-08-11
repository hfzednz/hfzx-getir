"use client";

import { useEffect, useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { AdminShell } from "@/components/shell/admin-shell";
import { useAuthStore } from "@/shared/auth/auth-store";

export default function AppLayout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    setHydrated(useAuthStore.persist.hasHydrated());
    return useAuthStore.persist.onFinishHydration(() => setHydrated(true));
  }, []);

  useEffect(() => {
    if (hydrated && !session) {
      router.replace("/login");
    }
  }, [hydrated, session, router]);

  if (!hydrated) {
    return (
      <div className="min-h-screen flex items-center justify-center text-[var(--nx-text-secondary)] text-[13px]">
        Loading session…
      </div>
    );
  }

  if (!session) {
    return null;
  }

  return <AdminShell>{children}</AdminShell>;
}
