"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { orderStateLabel } from "@nexora/web-core";
import { customerApi } from "@/shared/api/client";

type OrderRow = {
  id?: string;
  orderId?: string;
  status?: string;
  totalMinor?: number;
  currency?: string;
};

function unwrapOrders(data: unknown): OrderRow[] {
  if (Array.isArray(data)) return data as OrderRow[];
  if (data && typeof data === "object" && "items" in data) {
    const items = (data as { items?: unknown }).items;
    if (Array.isArray(items)) return items as OrderRow[];
  }
  return [];
}

function rowId(row: OrderRow): string {
  return (row.id ?? row.orderId ?? "").toString();
}

export default function OrdersPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["orders"],
    queryFn: async () => {
      const raw = await customerApi().request<unknown>("/v1/customer/orders");
      return unwrapOrders(raw);
    },
  });

  const rows = data ?? [];

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Orders</h1>
      {isLoading ? <p className="text-sm text-neutral-500">Loading…</p> : null}
      {error ? (
        <p className="text-sm text-red-600" role="alert">
          {error instanceof Error ? error.message : "Failed to load orders"}
        </p>
      ) : null}
      {!isLoading && !error && rows.length === 0 ? (
        <p className="text-sm text-neutral-600">No orders yet.</p>
      ) : null}
      {rows.length > 0 ? (
        <ul className="divide-y rounded-xl border">
          {rows.map((row) => {
            const id = rowId(row);
            if (!id) return null;
            return (
              <li key={id}>
                <Link
                  href={`/orders/${id}`}
                  className="flex items-center justify-between gap-3 p-4 hover:bg-neutral-50"
                  style={{ minHeight: 44 }}
                >
                  <span className="min-w-0">
                    <span className="block truncate text-sm font-medium">
                      Order {id.slice(0, 8)}…
                    </span>
                    <span className="block text-xs text-neutral-600">
                      {row.status ? orderStateLabel(row.status) : "Status unavailable"}
                    </span>
                  </span>
                  {typeof row.totalMinor === "number" ? (
                    <span className="text-sm font-medium">
                      {row.currency ?? "TRY"} {(row.totalMinor / 100).toFixed(2)}
                    </span>
                  ) : null}
                </Link>
              </li>
            );
          })}
        </ul>
      ) : null}
    </div>
  );
}
