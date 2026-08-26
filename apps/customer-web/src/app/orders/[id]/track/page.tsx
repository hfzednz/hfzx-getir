"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { orderStateLabel, subscribeOrderSse } from "@nexora/web-core";
import { customerApi } from "@/shared/api/client";

export default function OrderTrackPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const [sseLive, setSseLive] = useState(false);
  const [sseStatus, setSseStatus] = useState<string | null>(null);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["track", id],
    queryFn: () =>
      customerApi().request<{ status?: string; etaMinutes?: number }>(
        `/v1/customer/orders/${id}/track`,
      ),
    enabled: Boolean(id),
    refetchInterval: sseLive ? false : 5000,
  });

  useEffect(() => {
    if (!id) return;
    setSseLive(false);
    const unsub = subscribeOrderSse(
      id,
      (payload) => {
        setSseLive(true);
        if (typeof payload.status === "string") {
          setSseStatus(payload.status);
        }
      },
      () => setSseLive(false),
    );
    return unsub;
  }, [id]);

  const status = sseStatus ?? data?.status;

  return (
    <div className="space-y-4">
      <Link href={`/orders/${id}`} className="text-sm text-[var(--nx-brand)]">
        ← Order detail
      </Link>
      <h1 className="text-xl font-semibold">Live tracking</h1>
      <p className="text-xs text-neutral-500">
        {sseLive ? "Connected via SSE (polling fallback when disconnected)." : "Polling every 5s (SSE reconnecting…)."}
      </p>
      {isLoading ? <p>Loading…</p> : null}
      {error ? (
        <p className="text-red-600" role="alert">
          {error instanceof Error ? error.message : "Tracking unavailable"}
        </p>
      ) : null}
      {status ? (
        <div className="rounded-xl border p-4">
          <p className="font-medium">{orderStateLabel(status ?? "unknown")}</p>
          {data?.etaMinutes != null ? (
            <p className="text-sm text-neutral-600">ETA ~{data.etaMinutes} min</p>
          ) : null}
        </div>
      ) : null}
      <button
        type="button"
        className="rounded-lg border px-4 py-2 text-sm"
        onClick={() => refetch()}
      >
        Refresh
      </button>
    </div>
  );
}
