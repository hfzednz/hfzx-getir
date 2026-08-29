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
  const [sseOpen, setSseOpen] = useState(false);
  const [sseEvent, setSseEvent] = useState(false);
  const [sseStatus, setSseStatus] = useState<string | null>(null);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["track", id],
    queryFn: () =>
      customerApi().request<{ status?: string; etaMinutes?: number }>(
        `/v1/customer/orders/${id}/track`,
      ),
    enabled: Boolean(id),
    refetchInterval: sseOpen ? false : 5000,
  });

  useEffect(() => {
    if (!id) return;
    setSseOpen(false);
    setSseEvent(false);
    const unsub = subscribeOrderSse(
      id,
      (payload) => {
        setSseEvent(true);
        if (typeof payload.status === "string") {
          setSseStatus(payload.status);
        }
      },
      () => setSseOpen(false),
      () => setSseOpen(true),
      async () => {
        const res = await customerApi().request<{ ticket: string }>(
          `/v1/customer/orders/${id}/realtime-ticket`,
          { method: "POST", body: {} },
        );
        if (!res.ticket) throw new Error("sse ticket missing");
        return res.ticket;
      },
    );
    return unsub;
  }, [id]);

  const status = sseStatus ?? data?.status;

  return (
    <div className="space-y-4">
      <Link
        href={`/orders/${id}`}
        className="inline-flex items-center text-sm text-[var(--nx-brand)]"
        style={{ minHeight: 44 }}
      >
        ← Order detail
      </Link>
      <h1 className="text-xl font-semibold">Live tracking</h1>
      <p
        className="text-xs text-neutral-500"
        data-testid="sse-connection"
        aria-live="polite"
      >
        {sseOpen
          ? sseEvent
            ? "SSE connected; live event received."
            : "SSE connected (waiting for event)."
          : "Polling every 5s (SSE reconnecting…)."}
      </p>
      <p
        data-testid="sse-event-status"
        className="text-xs text-neutral-500"
        aria-live="polite"
      >
        Event: {sseEvent ? sseStatus ?? "received" : "none"}
      </p>
      {isLoading ? <p>Loading…</p> : null}
      {error ? (
        <p className="text-red-600" role="alert">
          {error instanceof Error ? error.message : "Tracking unavailable"}
        </p>
      ) : null}
      {status ? (
        <div className="rounded-xl border p-4">
          <p className="font-medium" data-testid="track-status">
            {orderStateLabel(status ?? "unknown")}
          </p>
          {data?.etaMinutes != null ? (
            <p className="text-sm text-neutral-600">ETA ~{data.etaMinutes} min</p>
          ) : null}
        </div>
      ) : null}
      <button
        type="button"
        className="rounded-lg border px-4 text-sm"
        style={{ minHeight: 44 }}
        onClick={() => refetch()}
      >
        Refresh
      </button>
    </div>
  );
}
