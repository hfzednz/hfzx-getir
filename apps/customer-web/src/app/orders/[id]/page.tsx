"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { orderStateLabel } from "@nexora/web-core";
import { customerApi } from "@/shared/api/client";

type OrderLine = {
  sku?: string;
  name?: string;
  qty?: number;
  lineTotalMinor?: number;
};

type Order = {
  id?: string;
  status?: string;
  state?: string;
  currency?: string;
  subtotalMinor?: number;
  discountMinor?: number;
  totalMinor?: number;
  placedAt?: string;
  completedAt?: string | null;
  addressSnapshot?: Record<string, unknown> | null;
  lines?: OrderLine[];
};

function money(minor: number | undefined, currency: string) {
  return typeof minor === "number" ? `${currency} ${(minor / 100).toFixed(2)}` : "—";
}

function addressText(snapshot: Record<string, unknown> | null | undefined) {
  if (!snapshot) return "";
  const parts = ["line1", "line2", "district", "city", "country"]
    .map((key) => snapshot[key])
    .filter((v): v is string => typeof v === "string" && v.trim().length > 0);
  return Array.from(new Set(parts)).join(", ");
}

export default function OrderDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: ["order", id],
    queryFn: () => customerApi().request<Order>(`/v1/customer/orders/${id}`),
    enabled: Boolean(id) && id !== "latest",
  });

  const currency = data?.currency ?? "TRY";
  const status = data?.status ?? data?.state;
  const address = addressText(data?.addressSnapshot);
  const lines = data?.lines ?? [];

  return (
    <div className="space-y-4">
      <Link href="/orders" className="text-sm text-[var(--nx-brand)]">
        ← Back to orders
      </Link>
      <h1 className="text-xl font-semibold">Order</h1>
      {isLoading ? <p>Loading…</p> : null}
      {error ? (
        <p className="text-red-600" role="alert">
          {error instanceof Error ? error.message : "Failed to load order"}
        </p>
      ) : null}
      {!isLoading && !error && !data ? (
        <p className="text-sm text-neutral-500">This order could not be found.</p>
      ) : null}
      {data ? (
        <>
          <p className="text-sm text-neutral-600">ID: {id}</p>
          <p className="font-medium" data-testid="order-status">
            Status: {orderStateLabel(String(status ?? "unknown"))}
          </p>
          {data.placedAt ? (
            <p className="text-sm text-neutral-600">
              Placed: {new Date(data.placedAt).toLocaleString()}
            </p>
          ) : null}
          {address ? (
            <section className="rounded-xl border p-4 text-sm">
              <h2 className="mb-1 font-medium">Delivery address</h2>
              <p className="text-neutral-600">{address}</p>
            </section>
          ) : null}
          {lines.length > 0 ? (
            <section className="rounded-xl border p-4 text-sm">
              <h2 className="mb-2 font-medium">Items</h2>
              <ul className="space-y-1">
                {lines.map((line, index) => (
                  <li key={`${line.sku ?? "line"}-${index}`} className="flex justify-between gap-3">
                    <span className="min-w-0 truncate text-neutral-700">
                      {line.qty ?? 1} × {line.name ?? line.sku ?? "Item"}
                    </span>
                    <span className="font-medium">{money(line.lineTotalMinor, currency)}</span>
                  </li>
                ))}
              </ul>
            </section>
          ) : null}
          <section className="rounded-xl border p-4 text-sm">
            <h2 className="mb-2 font-medium">Payment</h2>
            <dl className="space-y-1">
              <div className="flex justify-between gap-3">
                <dt className="text-neutral-600">Subtotal</dt>
                <dd>{money(data.subtotalMinor, currency)}</dd>
              </div>
              <div className="flex justify-between gap-3">
                <dt className="text-neutral-600">Discount</dt>
                <dd>{money(data.discountMinor, currency)}</dd>
              </div>
              <div className="flex justify-between gap-3">
                <dt className="text-neutral-600">Total</dt>
                <dd className="font-semibold">{money(data.totalMinor, currency)}</dd>
              </div>
            </dl>
          </section>
          <Link
            href={`/orders/${id}/track`}
            className="inline-flex items-center rounded-lg bg-[var(--nx-brand)] px-4 text-sm text-white"
            style={{ minHeight: 44 }}
          >
            Track delivery
          </Link>
        </>
      ) : null}
    </div>
  );
}
