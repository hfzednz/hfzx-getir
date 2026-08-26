"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { supportApi, useSession } from "@/shared/api/client";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [dashboard, setDashboard] = useState<Record<string, unknown> | null>(null);
  const [orders, setOrders] = useState<unknown[]>([]);
  const [error, setError] = useState("");

  useEffect(() => { if (!session) router.replace("/login"); }, [session, router]);

  useEffect(() => {
    if (!session) return;
    (async () => {
      try {
        const dash = await supportApi().request<Record<string, unknown>>("/v1/admin/dashboard");
        const ord = await supportApi().request<{ items?: unknown[] } | unknown[]>("/v1/admin/orders");
        setDashboard(dash);
        setOrders(Array.isArray(ord) ? ord : (ord.items ?? []));
      } catch (e) {
        setError(e instanceof Error ? e.message : "Load failed");
      }
    })();
  }, [session]);

  return (
    <div className="space-y-4 p-4">
      <div className="flex justify-between">
        <h1 className="text-xl font-semibold">Support</h1>
        <button type="button" className="text-sm" onClick={() => { logout(); router.push("/login"); }}>Logout</button>
      </div>
      {error ? <p className="text-sm text-red-600" role="alert">{error}</p> : null}
      <pre className="overflow-auto rounded bg-neutral-100 p-3 text-xs">{JSON.stringify(dashboard, null, 2)}</pre>
      <p className="text-sm">Orders loaded: {orders.length}</p>
    </div>
  );
}
