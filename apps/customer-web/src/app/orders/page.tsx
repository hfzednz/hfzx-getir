"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { orderStateLabel } from "@nexora/web-core";
import { customerApi } from "@/shared/api/client";

// The customer BFF exposes GET /v1/customer/orders/{id} but no list endpoint, so the
// ids placed from this device are remembered locally and each one is read from the API.
const STORAGE_KEY = "nexora-customer-order-ids";

export default function OrdersPage() {
  const [ids, setIds] = useState<string[] | null>(null);

  useEffect(() => {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      setIds(raw ? (JSON.parse(raw) as string[]) : []);
    } catch {
      setIds([]);
    }
  }, []);

  const { data, isLoading, error } = useQuery({
    queryKey: ["orders", ids],
    enabled: Array.isArray(ids) && ids.length > 0,
    queryFn: async () => {
      const api = customerApi();
      return Promise.all(
        (ids ?? []).map(async (id) => {
          try {
            const order = await api.request<{ status?: string; totalMinor?: number; currency?: string }>(
              `/v1/customer/orders/${id}`,
            );
            return { id, status: order.status, totalMinor: order.totalMinor, currency: order.currency };
          } catch {
            return { id, status: undefined, totalMinor: undefined, currency: undefined };
          }
        }),
      );
    },
  });

  const rows = data ?? (ids ?? []).map((id) => ({ id, status: undefined, totalMinor: undefined, currency: undefined }));

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Orders</h1>
      {ids === null || isLoading ? <p className="text-sm text-neutral-500">Loading…</p> : null}
      {error ? (
        <p className="text-sm text-red-600" role="alert">
          {error instanceof Error ? error.message : "Failed to load orders"}
        </p>
      ) : null}
      {ids !== null && ids.length === 0 ? (
        <p className="text-sm text-neutral-600">No orders yet.</p>
      ) : null}
      {rows.length > 0 ? (
        <ul className="divide-y rounded-xl border">
          {rows.map((row) => (
            <li key={row.id}>
              <Link
                href={`/orders/${row.id}`}
                className="flex items-center justify-between gap-3 p-4 hover:bg-neutral-50"
                style={{ minHeight: 44 }}
              >
                <span className="min-w-0">
                  <span className="block truncate text-sm font-medium">
                    Order {row.id.slice(0, 8)}…
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
          ))}
        </ul>
      ) : null}
    </div>
  );
}
