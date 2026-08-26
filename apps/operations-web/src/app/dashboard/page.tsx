"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { operationsApi, useSession } from "@/shared/api/client";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [dashboard, setDashboard] = useState<Record<string, unknown> | null>(null);
  const [error, setError] = useState("");

  useEffect(() => { if (!session) router.replace("/login"); }, [session, router]);

  useEffect(() => {
    if (!session) return;
    (async () => {
      try {
        const dash = await operationsApi().request<Record<string, unknown>>("/v1/admin/dashboard");
        setDashboard(dash);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Load failed");
      }
    })();
  }, [session]);

  return (
    <div className="space-y-4 p-4">
      <div className="flex justify-between">
        <h1 className="text-xl font-semibold">City operations</h1>
        <button type="button" className="text-sm" onClick={() => { logout(); router.push("/login"); }}>Logout</button>
      </div>
      {error ? <p className="text-sm text-red-600" role="alert">{error}</p> : null}
      <pre className="overflow-auto rounded bg-neutral-100 p-3 text-xs">{JSON.stringify(dashboard, null, 2)}</pre>
    </div>
  );
}
