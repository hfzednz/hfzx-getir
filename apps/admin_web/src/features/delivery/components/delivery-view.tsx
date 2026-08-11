"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import {
  Button,
  DataGrid,
  FilterBar,
  Input,
  KpiCard,
  PageHeader,
  PermissionGate,
  Select,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/permissions";
import { useDeliverySnapshot } from "../hooks";
import type {
  CourierAllocation,
  DeliveryZoneListItem,
  DispatchQueueItem,
  DispatchQueueStatus,
} from "../types";

function queueTone(
  status: DispatchQueueStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "dispatched":
      return "success";
    case "assigning":
      return "info";
    case "delayed":
      return "danger";
    default:
      return "warning";
  }
}

function Panel({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
      <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
        {title}
      </h3>
      {children}
    </section>
  );
}

export function DeliveryView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useDeliverySnapshot();
  const [mode, setMode] = useState("all");
  const [q, setQ] = useState("");

  const queues = useMemo(() => {
    const items = data?.queues ?? [];
    return items.filter((row) => {
      if (mode !== "all" && row.mode !== mode) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        row.orderId.toLowerCase().includes(needle) ||
        row.zoneName.toLowerCase().includes(needle) ||
        (row.courierCode?.toLowerCase().includes(needle) ?? false)
      );
    });
  }, [data?.queues, mode, q]);

  const zoneCols = useMemo<DataGridColumnDef<DeliveryZoneListItem>[]>(
    () => [
      { id: "code", header: "Code", accessorKey: "code", width: 100 },
      { id: "name", header: "Zone", accessorKey: "name" },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "active"
                ? "success"
                : row.status === "paused"
                  ? "warning"
                  : "neutral"
            }
          />
        ),
        width: 100,
      },
      {
        id: "couriers",
        header: "Couriers",
        cell: ({ row }) => `${row.courierAllocated}/${row.courierTarget}`,
        align: "right",
        width: 90,
      },
      {
        id: "orders",
        header: "Orders",
        accessorKey: "activeOrders",
        align: "right",
        width: 80,
      },
      {
        id: "sla",
        header: "SLA",
        cell: ({ row }) => (
          <span className="tabular-nums">{row.slaPct.toFixed(1)}%</span>
        ),
        align: "right",
        width: 80,
      },
      {
        id: "eta",
        header: "Avg ETA",
        cell: ({ row }) => `${row.avgEtaMinutes} min`,
        align: "right",
        width: 90,
      },
    ],
    [],
  );

  const queueCols = useMemo<DataGridColumnDef<DispatchQueueItem>[]>(
    () => [
      { id: "order", header: "Order", accessorKey: "orderId", width: 120 },
      { id: "zone", header: "Zone", accessorKey: "zoneName" },
      { id: "mode", header: "Mode", accessorKey: "mode", width: 100 },
      {
        id: "status",
        header: "Queue",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={queueTone(row.status)} />
        ),
        width: 110,
      },
      {
        id: "eta",
        header: "ETA",
        cell: ({ row }) => `${row.etaMinutes} min`,
        align: "right",
        width: 80,
      },
      {
        id: "wait",
        header: "Wait",
        cell: ({ row }) => `${row.waitingMinutes} min`,
        align: "right",
        width: 70,
      },
      {
        id: "courier",
        header: "Courier",
        cell: ({ row }) => row.courierCode ?? "—",
        width: 100,
      },
    ],
    [],
  );

  const allocCols = useMemo<DataGridColumnDef<CourierAllocation>[]>(
    () => [
      { id: "zone", header: "Zone", accessorKey: "zoneName" },
      { id: "code", header: "Courier", accessorKey: "courierCode", width: 100 },
      { id: "name", header: "Name", accessorKey: "courierName" },
      { id: "live", header: "Live", accessorKey: "liveStatus", width: 100 },
      {
        id: "cap",
        header: "Capacity",
        accessorKey: "capacity",
        align: "right",
        width: 80,
      },
    ],
    [],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={96} />
        <Skeleton height={280} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load delivery ops
        </p>
        <p className="m-0 mt-[var(--nx-space-1)] text-[var(--nx-text-secondary)]">
          {error instanceof Error ? error.message : "Unknown error"}
        </p>
        <button
          type="button"
          onClick={() => void refetch()}
          className="mt-[var(--nx-space-3)] text-[var(--nx-text-link)] underline cursor-pointer bg-transparent border-0"
        >
          Retry
        </button>
      </div>
    );
  }

  const express = data.modeBreakdown.find((m) => m.mode === "express")?.count ?? 0;
  const scheduled =
    data.modeBreakdown.find((m) => m.mode === "scheduled")?.count ?? 0;
  const batch = data.modeBreakdown.find((m) => m.mode === "batch")?.count ?? 0;
  const delayed = data.queues.filter((q) => q.status === "delayed").length;

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Delivery"
        description={`Dispatch & ETA · updated ${new Date(data.generatedAt).toLocaleTimeString("tr-TR")}${isFetching ? " · refreshing…" : ""}`}
        actions={
          <div className="flex gap-[var(--nx-space-2)]">
            <Link href="/delivery/zones" className="no-underline">
              <Button size="sm" variant="secondary">
                Zones editor
              </Button>
            </Link>
            <PermissionGate allowed={can(session, "delivery:write")}>
              <Button size="sm">Rebalance couriers</Button>
            </PermissionGate>
          </div>
        }
      />

      <section className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
        <KpiCard title="Express" value={express} tone="brand" />
        <KpiCard title="Scheduled" value={scheduled} tone="neutral" />
        <KpiCard title="Batch" value={batch} tone="neutral" />
        <KpiCard title="Delayed queue" value={delayed} tone="danger" />
      </section>

      <FilterBar
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      >
        <Input
          placeholder="Search order, zone, courier…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search dispatch queue"
        />
        <Select
          value={mode}
          onChange={(e) => setMode(e.target.value)}
          aria-label="Delivery mode"
        >
          <option value="all">All modes</option>
          <option value="express">Express</option>
          <option value="scheduled">Scheduled</option>
          <option value="batch">Batch</option>
        </Select>
      </FilterBar>

      <Panel title="Dispatch queues">
        <DataGrid
          columns={queueCols}
          data={queues}
          getRowId={(r) => r.id}
          emptyMessage="Queue empty"
        />
      </Panel>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <Panel title="Zones overview">
          <DataGrid
            columns={zoneCols}
            data={data.zones}
            getRowId={(r) => r.id}
          />
        </Panel>
        <Panel title="Courier allocation">
          <DataGrid
            columns={allocCols}
            data={data.allocations}
            getRowId={(r) => r.id}
          />
        </Panel>
        <Panel title="ETA monitoring">
          <DataGrid
            columns={[
              { id: "zone", header: "Zone", accessorKey: "zoneName" },
              {
                id: "p50",
                header: "P50",
                cell: ({ row }) => `${row.p50} min`,
                align: "right",
              },
              {
                id: "p90",
                header: "P90",
                cell: ({ row }) => `${row.p90} min`,
                align: "right",
              },
              {
                id: "br",
                header: "Breached",
                accessorKey: "breached",
                align: "right",
              },
            ]}
            data={data.etaMonitor}
            getRowId={(r) => r.zoneName}
          />
        </Panel>
      </div>
    </div>
  );
}
