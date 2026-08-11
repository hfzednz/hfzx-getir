"use client";

import type { ReactNode } from "react";
import {
  Badge,
  EmptyState,
  KpiCard,
  PageHeader,
  Skeleton,
  StatusBadge,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { useLiveSnapshot } from "../hooks";
import type {
  CourierMarker,
  LiveOrderEvent,
  OpsSeverity,
  WarehouseActivity,
} from "../types";

function severityTone(
  severity: OpsSeverity,
): "danger" | "warning" | "info" {
  if (severity === "danger") return "danger";
  if (severity === "warning") return "warning";
  return "info";
}

function orderStatusTone(
  status: LiveOrderEvent["status"],
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "delivered":
      return "success";
    case "failed":
    case "cancelled":
      return "danger";
    case "en_route":
    case "assigned":
      return "info";
    case "picking":
    case "ready":
      return "warning";
    default:
      return "neutral";
  }
}

function courierTone(
  status: CourierMarker["status"],
): "success" | "warning" | "danger" | "neutral" {
  switch (status) {
    case "available":
      return "success";
    case "busy":
      return "warning";
    case "offline":
      return "danger";
    default:
      return "neutral";
  }
}

function warehouseTone(
  status: WarehouseActivity["status"],
): "success" | "warning" | "danger" {
  if (status === "healthy") return "success";
  if (status === "busy") return "warning";
  return "danger";
}

function Panel({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)] min-h-[200px]">
      <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
        {title}
      </h3>
      {children}
    </div>
  );
}

export function LiveOpsView() {
  const { data, isLoading, isError, error, refetch, connection } =
    useLiveSnapshot();

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} height={88} />
          ))}
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
          <Skeleton height={280} />
          <Skeleton height={280} />
        </div>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <EmptyState
        title="Failed to load live ops"
        description={
          error instanceof Error ? error.message : "Unknown error"
        }
        action={
          <button
            type="button"
            onClick={() => void refetch()}
            className="text-[var(--nx-text-link)] underline cursor-pointer bg-transparent border-0"
          >
            Retry
          </button>
        }
      />
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Live operations"
        description={`City scope · ${new Date(data.generatedAt).toLocaleTimeString("tr-TR")}`}
        actions={
          <Badge
            tone={
              connection === "live"
                ? "success"
                : connection === "polling"
                  ? "warning"
                  : "danger"
            }
          >
            {connection === "live"
              ? "WebSocket live"
              : connection === "polling"
                ? "Polling fallback"
                : "Offline"}
          </Badge>
        }
      />

      <section
        aria-label="Live KPIs"
        className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]"
      >
        <KpiCard
          title="Active orders"
          value={data.counts.activeOrders.toLocaleString("tr-TR")}
          tone="brand"
        />
        <KpiCard
          title="Delayed"
          value={String(data.counts.delayedOrders)}
          tone="warning"
        />
        <KpiCard
          title="Available couriers"
          value={String(data.counts.availableCouriers)}
          tone="success"
        />
        <KpiCard
          title="Emergencies"
          value={String(data.counts.openEmergencies)}
          tone="danger"
        />
      </section>

      <section
        aria-label="Streams"
        className="grid grid-cols-1 xl:grid-cols-2 gap-[var(--nx-space-3)]"
      >
        <Panel title="Order stream">
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)] max-h-[320px] overflow-auto">
            {data.orderStream.map((evt) => (
              <li
                key={evt.id}
                className="flex items-start justify-between gap-[var(--nx-space-3)] py-[var(--nx-space-2)] border-b border-[var(--nx-border-subtle)] last:border-0"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-[var(--nx-space-2)] flex-wrap">
                    <span className="text-[13px] font-semibold tabular-nums">
                      {evt.orderId}
                    </span>
                    <StatusBadge
                      status={evt.status.replace("_", " ")}
                      tone={orderStatusTone(evt.status)}
                    />
                    {evt.delayMinutes > 0 ? (
                      <Badge tone="warning">+{evt.delayMinutes}m</Badge>
                    ) : null}
                  </div>
                  <p className="m-0 mt-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
                    {evt.customerName} · {evt.warehouseCode} · {evt.zone}
                    {evt.etaMinutes != null ? ` · ETA ${evt.etaMinutes}m` : ""}
                  </p>
                </div>
                <span className="text-[12px] tabular-nums text-[var(--nx-text-tertiary)] shrink-0">
                  {formatMinorUnits(evt.amountMinor, evt.currency)}
                </span>
              </li>
            ))}
          </ul>
        </Panel>

        <Panel title="Courier markers">
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)] max-h-[320px] overflow-auto">
            {data.couriers.map((c) => (
              <li
                key={c.id}
                className="flex items-center justify-between gap-[var(--nx-space-3)] py-[var(--nx-space-2)] border-b border-[var(--nx-border-subtle)] last:border-0"
              >
                <div>
                  <div className="flex items-center gap-[var(--nx-space-2)]">
                    <span className="text-[13px] font-semibold">{c.name}</span>
                    <StatusBadge status={c.status} tone={courierTone(c.status)} />
                  </div>
                  <p className="m-0 mt-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)] tabular-nums">
                    {c.zone} · {c.lat.toFixed(4)}, {c.lng.toFixed(4)}
                    {c.activeOrderId ? ` · ${c.activeOrderId}` : ""}
                  </p>
                </div>
                <span className="text-[11px] tabular-nums text-[var(--nx-text-tertiary)]">
                  {c.batteryPct}%
                </span>
              </li>
            ))}
          </ul>
        </Panel>
      </section>

      <section
        aria-label="Ops detail"
        className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-[var(--nx-space-3)]"
      >
        <Panel title="Warehouse activity">
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
            {data.warehouses.map((w) => (
              <li
                key={w.id}
                className="flex items-center justify-between gap-[var(--nx-space-2)] py-[var(--nx-space-1)] border-b border-[var(--nx-border-subtle)] last:border-0"
              >
                <div>
                  <p className="m-0 text-[13px] font-semibold">
                    {w.code} · {w.name}
                  </p>
                  <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                    Queue {w.pickQueueMin}m · {w.openOrders} open ·{" "}
                    {w.capacityPct}%
                  </p>
                </div>
                <StatusBadge status={w.status} tone={warehouseTone(w.status)} />
              </li>
            ))}
          </ul>
        </Panel>

        <Panel title="Delays">
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
            {data.delays.map((d) => (
              <li key={d.id} className="flex flex-col gap-[var(--nx-space-1)]">
                <div className="flex items-center gap-[var(--nx-space-2)]">
                  <StatusBadge
                    status={`+${d.delayMinutes}m`}
                    tone={severityTone(d.severity)}
                  />
                  <span className="text-[13px] font-semibold tabular-nums">
                    {d.orderId}
                  </span>
                </div>
                <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                  {d.zone} · {d.reason}
                </p>
              </li>
            ))}
          </ul>
        </Panel>

        <Panel title="Failed deliveries">
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
            {data.failedDeliveries.map((f) => (
              <li key={f.id}>
                <p className="m-0 text-[13px] font-semibold tabular-nums">
                  {f.orderId}
                </p>
                <p className="m-0 mt-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
                  {f.courierName} · attempt {f.attempts} · {f.reason}
                </p>
              </li>
            ))}
          </ul>
        </Panel>

        <Panel title="Bottlenecks">
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
            {data.bottlenecks.map((b) => (
              <li key={b.id}>
                <div className="flex items-center gap-[var(--nx-space-2)]">
                  <StatusBadge
                    status={b.type}
                    tone={severityTone(b.severity)}
                  />
                  <span className="text-[13px] font-semibold">{b.title}</span>
                  <span className="ml-auto text-[11px] tabular-nums text-[var(--nx-text-tertiary)]">
                    impact {b.impactScore}
                  </span>
                </div>
                <p className="m-0 mt-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
                  {b.detail}
                </p>
              </li>
            ))}
          </ul>
        </Panel>

        <Panel title="Emergencies">
          {data.emergencies.length === 0 ? (
            <p className="m-0 text-[13px] text-[var(--nx-text-secondary)]">
              No open emergencies
            </p>
          ) : (
            <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
              {data.emergencies.map((e) => (
                <li key={e.id}>
                  <div className="flex items-center gap-[var(--nx-space-2)]">
                    <StatusBadge
                      status={e.acknowledged ? "acked" : "open"}
                      tone={e.acknowledged ? "neutral" : "danger"}
                    />
                    <span className="text-[13px] font-semibold">{e.title}</span>
                  </div>
                  <p className="m-0 mt-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
                    {e.zone} · {e.detail}
                  </p>
                </li>
              ))}
            </ul>
          )}
        </Panel>

        <Panel title="Alerts">
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
            {data.alerts.map((a) => (
              <li key={a.id} className="flex flex-col gap-[var(--nx-space-1)]">
                <div className="flex items-center gap-[var(--nx-space-2)]">
                  <StatusBadge
                    status={a.severity}
                    tone={severityTone(a.severity)}
                  />
                  <span className="text-[13px] font-semibold">{a.title}</span>
                </div>
                <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                  {a.detail}
                </p>
              </li>
            ))}
          </ul>
        </Panel>
      </section>
    </div>
  );
}
