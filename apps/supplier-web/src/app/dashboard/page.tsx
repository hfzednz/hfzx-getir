"use client";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { RouteGuard } from "@nexora/web-core";
import { supplierApi, useSession } from "@/shared/api/client";

type Row = { id?: string; name?: string; status?: string; code?: string };

function rows(payload: { items?: Row[] } | Row[]): Row[] {
  return Array.isArray(payload) ? payload : payload.items ?? [];
}

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);

  const { data, isLoading, error } = useQuery({
    queryKey: ["supplier-dashboard"],
    enabled: Boolean(session),
    queryFn: async () => {
      const api = supplierApi();
      const [suppliers, purchaseOrders, sellers, merchant] = await Promise.all([
        api.request<{ items?: Row[] } | Row[]>("/v1/supplier/suppliers"),
        api.request<{ items?: Row[] } | Row[]>("/v1/supplier/purchase-orders"),
        api
          .request<{ items?: Row[] } | Row[]>("/v1/supplier/sellers")
          .then((res) => ({ ok: true as const, rows: rows(res) }))
          .catch((err: unknown) => ({
            ok: false as const,
            message: err instanceof Error ? err.message : "Sellers unavailable",
          })),
        api
          .request<{
            listings?: Row[];
            summary?: { activeOrders?: number; products?: number; inventory?: number };
          }>("/v1/supplier/merchant/dashboard")
          .then((res) => ({ ok: true as const, ...res }))
          .catch((err: unknown) => ({
            ok: false as const,
            message: err instanceof Error ? err.message : "Merchant dashboard unavailable",
          })),
      ]);
      return {
        suppliers: rows(suppliers),
        purchaseOrders: rows(purchaseOrders),
        sellers,
        merchant,
      };
    },
  });

  const sections = [
    { title: "Suppliers", items: data?.suppliers ?? [] },
    { title: "Purchase orders", items: data?.purchaseOrders ?? [] },
  ];

  return (
    <RouteGuard session={session} allow={["supplier", "partner", "merchant"]} onDeny={logout}>
      <div className="space-y-4 p-4">
        <div className="flex items-center justify-between gap-3">
          <h1 className="text-xl font-semibold">Supplier / merchant portal</h1>
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
        {isLoading ? <p className="text-sm text-neutral-500">Loading…</p> : null}
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error instanceof Error ? error.message : "Load failed"}
          </p>
        ) : null}

        {sections.map((section) => (
          <section key={section.title} className="rounded-xl border p-4">
            <h2 className="font-medium">
              {section.title} ({section.items.length})
            </h2>
            {!isLoading && section.items.length === 0 ? (
              <p className="text-sm text-neutral-500">Nothing to show yet.</p>
            ) : null}
            {section.items.length > 0 ? (
              <ul className="mt-2 divide-y text-sm">
                {section.items.map((row, index) => (
                  <li key={row.id ?? index} className="flex items-center justify-between gap-3 py-2">
                    <span className="min-w-0 truncate">
                      {row.name ?? row.code ?? row.id ?? "—"}
                    </span>
                    {row.status ? (
                      <span className="shrink-0 text-neutral-600">{row.status}</span>
                    ) : null}
                  </li>
                ))}
              </ul>
            ) : null}
          </section>
        ))}

        <section className="rounded-xl border p-4">
          <h2 className="font-medium">
            Marketplace sellers
            {data?.sellers.ok ? ` (${data.sellers.rows.length})` : ""}
          </h2>
          {data && !data.sellers.ok ? (
            <p className="text-sm text-red-600" role="alert">
              {data.sellers.message}
            </p>
          ) : null}
          {data?.sellers.ok && data.sellers.rows.length === 0 ? (
            <p className="text-sm text-neutral-500">No sellers registered yet.</p>
          ) : null}
          {data?.sellers.ok && data.sellers.rows.length > 0 ? (
            <ul className="mt-2 divide-y text-sm">
              {data.sellers.rows.map((row, index) => (
                <li key={row.id ?? index} className="truncate py-2">
                  {row.name ?? row.id ?? "—"}
                </li>
              ))}
            </ul>
          ) : null}
        </section>

        <section className="rounded-xl border p-4">
          <h2 className="font-medium">
            Catalog listings
            {data?.merchant.ok && data.merchant.listings ? ` (${data.merchant.listings.length})` : ""}
          </h2>
          {data && !data.merchant.ok ? (
            <p className="text-sm text-red-600" role="alert">
              {data.merchant.message}
            </p>
          ) : null}
          {data?.merchant.ok && data.merchant.summary ? (
            <p className="text-sm text-neutral-600">
              Active orders {data.merchant.summary.activeOrders ?? 0} · products{" "}
              {data.merchant.summary.products ?? 0} · inventory units{" "}
              {data.merchant.summary.inventory ?? 0}
            </p>
          ) : null}
          {data?.merchant.ok && (data.merchant.listings ?? []).length === 0 ? (
            <p className="text-sm text-neutral-500">No listings yet.</p>
          ) : null}
          {data?.merchant.ok && (data.merchant.listings ?? []).length > 0 ? (
            <ul className="mt-2 divide-y text-sm">
              {(data.merchant.listings ?? []).map((row, index) => (
                <li key={row.id ?? index} className="flex items-center justify-between gap-3 py-2">
                  <span className="min-w-0 truncate">{row.name ?? row.code ?? row.id ?? "—"}</span>
                  {row.status ? (
                    <span className="shrink-0 text-neutral-600">{row.status}</span>
                  ) : null}
                </li>
              ))}
            </ul>
          ) : null}
        </section>
      </div>
    </RouteGuard>
  );
}
