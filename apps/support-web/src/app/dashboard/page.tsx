"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { RouteGuard } from "@nexora/web-core";
import { supportApi, useSession } from "@/shared/api/client";

type OrderRow = {
  id?: string;
  status?: string;
  currency?: string;
  totalMinor?: number;
};

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [lookupId, setLookupId] = useState("");
  const [lookup, setLookup] = useState<OrderRow | null>(null);
  const [lookupError, setLookupError] = useState("");
  const [lookupBusy, setLookupBusy] = useState(false);

  const { data, isLoading, error } = useQuery({
    queryKey: ["support-orders"],
    enabled: Boolean(session),
    queryFn: async () => {
      const api = supportApi();
      const [dashboard, orders] = await Promise.all([
        api.request<{ ordersLive?: number }>("/v1/admin/dashboard"),
        api.request<{ items?: OrderRow[] } | OrderRow[]>("/v1/admin/orders?pageSize=25"),
      ]);
      const items = Array.isArray(orders) ? orders : orders.items ?? [];
      return { ordersLive: dashboard.ordersLive, items };
    },
  });

  async function findOrder() {
    const id = lookupId.trim();
    if (!id) return;
    setLookupBusy(true);
    setLookupError("");
    setLookup(null);
    try {
      setLookup(await supportApi().request<OrderRow>(`/v1/admin/orders/${id}`));
    } catch (e) {
      setLookupError(e instanceof Error ? e.message : "Order lookup failed");
    } finally {
      setLookupBusy(false);
    }
  }

  const items = data?.items ?? [];

  return (
    <RouteGuard session={session} allow={["support_agent", "admin", "super_admin"]} onDeny={logout}>
      <div className="space-y-4 p-4">
        <div className="flex items-center justify-between gap-3">
          <h1 className="text-xl font-semibold">Support</h1>
          <button
            type="button"
            className="rounded-lg px-3 text-sm"
            style={{ minHeight: 44 }}
            onClick={() => {
              logout();
              router.push("/login");
            }}
          >
            Logout
          </button>
        </div>

        <section className="space-y-2 rounded-xl border p-4">
          <h2 className="font-medium">Order lookup</h2>
          <label className="block text-sm">
            Order id
            <input
              className="mt-1 w-full rounded-lg border px-3 py-3"
              style={{ minHeight: 44 }}
              value={lookupId}
              onChange={(e) => setLookupId(e.target.value)}
              placeholder="uuid"
              aria-invalid={lookupError ? true : undefined}
              aria-describedby={lookupError ? "support-lookup-error" : undefined}
            />
          </label>
          <button
            type="button"
            className="w-full rounded-lg border py-3 text-sm disabled:opacity-60"
            style={{ minHeight: 44 }}
            onClick={findOrder}
            disabled={lookupBusy || !lookupId.trim()}
          >
            {lookupBusy ? "Searching…" : "Find order"}
          </button>
          {lookupError ? (
            <p id="support-lookup-error" className="text-sm text-red-600" role="alert">
              {lookupError}
            </p>
          ) : null}
          {lookup ? (
            <dl className="text-sm">
              <div className="flex justify-between gap-3">
                <dt className="text-neutral-600">Status</dt>
                <dd className="font-medium">{lookup.status ?? "unknown"}</dd>
              </div>
              {typeof lookup.totalMinor === "number" ? (
                <div className="flex justify-between gap-3">
                  <dt className="text-neutral-600">Total</dt>
                  <dd className="font-medium">
                    {lookup.currency ?? "TRY"} {(lookup.totalMinor / 100).toFixed(2)}
                  </dd>
                </div>
              ) : null}
            </dl>
          ) : null}
        </section>

        <section className="rounded-xl border p-4">
          <h2 className="font-medium">
            Recent orders{data?.ordersLive != null ? ` (${data.ordersLive} live)` : ""}
          </h2>
          {isLoading ? <p className="text-sm text-neutral-500">Loading…</p> : null}
          {error ? (
            <p className="text-sm text-red-600" role="alert">
              {error instanceof Error ? error.message : "Load failed"}
            </p>
          ) : null}
          {!isLoading && !error && items.length === 0 ? (
            <p className="text-sm text-neutral-500">No orders for this tenant yet.</p>
          ) : null}
          {items.length > 0 ? (
            <ul className="mt-2 divide-y text-sm">
              {items.map((order, index) => (
                <li key={order.id ?? index} className="flex items-center justify-between gap-3 py-2">
                  <span className="min-w-0 truncate">{order.id ?? "—"}</span>
                  <span className="shrink-0 text-neutral-600">{order.status ?? "unknown"}</span>
                </li>
              ))}
            </ul>
          ) : null}
        </section>
      </div>
    </RouteGuard>
  );
}
