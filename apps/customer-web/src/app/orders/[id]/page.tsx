"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { orderStateLabel } from "@nexora/web-core";
import { customerApi } from "@/shared/api/client";

export default function OrderDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: ["order", id],
    queryFn: () =>
      customerApi().request<Record<string, unknown>>(`/v1/customer/orders/${id}`),
    enabled: Boolean(id) && id !== "latest",
  });

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
      {data ? (
        <>
          <p className="text-sm text-neutral-600">ID: {id}</p>
          <p className="font-medium">
            Status:{" "}
            {orderStateLabel(String(data.status ?? data.state ?? "unknown"))}
          </p>
          <pre className="overflow-auto rounded bg-neutral-50 p-3 text-xs">
            {JSON.stringify(data, null, 2)}
          </pre>
          <Link
            href={`/orders/${id}/track`}
            className="inline-block rounded-lg bg-[var(--nx-brand)] px-4 py-2 text-sm text-white"
          >
            Track delivery
          </Link>
        </>
      ) : null}
    </div>
  );
}
