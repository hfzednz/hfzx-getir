"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
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
import { useWarehouses } from "../hooks";
import type { WarehouseListItem, WarehouseStatus } from "../types";

function statusTone(
  status: WarehouseStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "open":
      return "success";
    case "busy":
      return "warning";
    case "maintenance":
      return "info";
    default:
      return "neutral";
  }
}

export function WarehousesView() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useWarehouses();
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("all");

  const filtered = useMemo(() => {
    const items = data?.items ?? [];
    return items.filter((w) => {
      if (status !== "all" && w.status !== status) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        w.name.toLowerCase().includes(needle) ||
        w.code.toLowerCase().includes(needle) ||
        w.district.toLowerCase().includes(needle)
      );
    });
  }, [data?.items, q, status]);

  const columns = useMemo<DataGridColumnDef<WarehouseListItem>[]>(
    () => [
      { id: "code", header: "Code", accessorKey: "code", width: 90 },
      { id: "name", header: "Warehouse", accessorKey: "name" },
      { id: "district", header: "District", accessorKey: "district", width: 110 },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={statusTone(row.status)} />
        ),
        width: 120,
      },
      {
        id: "capacity",
        header: "Capacity",
        cell: ({ row }) => (
          <span
            className="tabular-nums"
            style={{
              color:
                row.capacityPct >= 90
                  ? "var(--nx-danger)"
                  : "var(--nx-text-primary)",
            }}
          >
            {row.capacityPct}%
          </span>
        ),
        align: "right",
        width: 90,
      },
      {
        id: "skus",
        header: "SKUs",
        accessorKey: "skuCount",
        align: "right",
        width: 80,
      },
      {
        id: "orders",
        header: "Open orders",
        accessorKey: "openOrders",
        align: "right",
        width: 100,
      },
      {
        id: "pick",
        header: "Pick SLA",
        cell: ({ row }) => (
          <span className="tabular-nums">{row.pickSlaPct.toFixed(1)}%</span>
        ),
        align: "right",
        width: 90,
      },
      {
        id: "alerts",
        header: "Alerts",
        accessorKey: "stockAlerts",
        align: "right",
        width: 70,
      },
    ],
    [],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={40} />
        <Skeleton height={280} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load warehouses
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
        title="Warehouses"
        description={`${data.total} dark stores · updated ${new Date(data.generatedAt).toLocaleTimeString("tr-TR")}${isFetching ? " · refreshing…" : ""}`}
        actions={
          <PermissionGate allowed={can(session, "warehouses:write")}>
            <Button size="sm" variant="secondary">
              Add warehouse
            </Button>
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
          placeholder="Search code, name, district…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search warehouses"
        />
        <Select
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          aria-label="Filter by status"
        >
          <option value="all">All statuses</option>
          <option value="open">Open</option>
          <option value="busy">Busy</option>
          <option value="closed">Closed</option>
          <option value="maintenance">Maintenance</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={filtered}
        getRowId={(row) => row.id}
        onRowClick={(row) => router.push(`/warehouses/${row.id}`)}
        emptyMessage="No warehouses match filters"
      />
    </div>
  );
}
