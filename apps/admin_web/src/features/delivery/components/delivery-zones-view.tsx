"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import {
  Button,
  DataGrid,
  FilterBar,
  Input,
  PageHeader,
  PermissionGate,
  Select,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/permissions";
import { useDeliveryZoneDetail, useDeliveryZones } from "../hooks";
import type { DeliveryZoneListItem } from "../types";

export function DeliveryZonesView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useDeliveryZones();
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("all");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const detail = useDeliveryZoneDetail(selectedId ?? "");

  const filtered = useMemo(() => {
    const items = data?.items ?? [];
    return items.filter((z) => {
      if (status !== "all" && z.status !== status) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        z.name.toLowerCase().includes(needle) ||
        z.code.toLowerCase().includes(needle)
      );
    });
  }, [data?.items, q, status]);

  const columns = useMemo<DataGridColumnDef<DeliveryZoneListItem>[]>(
    () => [
      { id: "code", header: "Code", accessorKey: "code", width: 110 },
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
        id: "poly",
        header: "Polygon pts",
        accessorKey: "polygonPoints",
        align: "right",
        width: 100,
      },
      {
        id: "alloc",
        header: "Couriers",
        cell: ({ row }) => `${row.courierAllocated}/${row.courierTarget}`,
        align: "right",
        width: 100,
      },
      {
        id: "orders",
        header: "Active orders",
        accessorKey: "activeOrders",
        align: "right",
        width: 110,
      },
      {
        id: "sla",
        header: "SLA",
        cell: ({ row }) => `${row.slaPct.toFixed(1)}%`,
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
        <Skeleton height={320} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load zones
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

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Delivery zones"
        description={
          <span>
            <Link href="/delivery" className="text-[var(--nx-text-link)]">
              Delivery
            </Link>
            {" · "}
            Polygon editor list · updated{" "}
            {new Date(data.generatedAt).toLocaleTimeString("tr-TR")}
            {isFetching ? " · refreshing…" : ""}
          </span>
        }
        actions={
          <PermissionGate allowed={can(session, "delivery:write")}>
            <Button size="sm">Create zone</Button>
          </PermissionGate>
        }
      />

      <FilterBar
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      >
        <Input
          placeholder="Search zone…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search zones"
        />
        <Select
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          aria-label="Status"
        >
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="paused">Paused</option>
          <option value="draft">Draft</option>
        </Select>
      </FilterBar>

      <div className="grid grid-cols-1 xl:grid-cols-[1fr_320px] gap-[var(--nx-space-3)]">
        <DataGrid
          columns={columns}
          data={filtered}
          getRowId={(row) => row.id}
          onRowClick={(row) => setSelectedId(row.id)}
          emptyMessage="No zones"
        />

        <aside className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            Zone editor
          </h3>
          {!selectedId ? (
            <p className="m-0 text-[13px] text-[var(--nx-text-secondary)]">
              Select a zone to inspect polygon metadata and staffing targets.
            </p>
          ) : detail.isLoading ? (
            <Skeleton height={160} />
          ) : detail.data ? (
            <div className="flex flex-col gap-[var(--nx-space-2)] text-[13px]">
              <p className="m-0 font-semibold">
                {detail.data.name} · {detail.data.code}
              </p>
              <StatusBadge status={detail.data.status} />
              <p className="m-0 text-[var(--nx-text-secondary)]">
                Hexes: {detail.data.hexCount}
              </p>
              <p className="m-0 text-[var(--nx-text-secondary)]">
                Couriers: {detail.data.courierAllocated}/
                {detail.data.courierTarget}
              </p>
              <p className="m-0 text-[var(--nx-text-secondary)]">
                Peaks: {detail.data.peakWindows.join(" · ")}
              </p>
              <p className="m-0 text-[12px] text-[var(--nx-text-tertiary)]">
                {detail.data.notes}
              </p>
              <PermissionGate allowed={can(session, "delivery:write")}>
                <Button size="sm" variant="secondary" className="mt-[var(--nx-space-2)]">
                  Edit polygon
                </Button>
              </PermissionGate>
            </div>
          ) : (
            <p className="m-0 text-[13px] text-[var(--nx-danger)]">
              Failed to load zone detail
            </p>
          )}
        </aside>
      </div>
    </div>
  );
}
