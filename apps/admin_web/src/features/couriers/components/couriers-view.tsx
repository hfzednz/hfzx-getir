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
import { useCouriers } from "../hooks";
import type { CourierListItem, CourierLiveStatus } from "../types";

function statusTone(
  status: CourierLiveStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "available":
      return "success";
    case "busy":
      return "info";
    case "break":
      return "warning";
    case "emergency":
      return "danger";
    default:
      return "neutral";
  }
}

export function CouriersView() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } = useCouriers();
  const [q, setQ] = useState("");
  const [status, setStatus] = useState<string>("all");

  const filtered = useMemo(() => {
    const items = data?.items ?? [];
    return items.filter((c) => {
      if (status !== "all" && c.liveStatus !== status) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        c.fullName.toLowerCase().includes(needle) ||
        c.code.toLowerCase().includes(needle) ||
        c.zoneName.toLowerCase().includes(needle) ||
        c.phone.includes(needle)
      );
    });
  }, [data?.items, q, status]);

  const columns = useMemo<DataGridColumnDef<CourierListItem>[]>(
    () => [
      { id: "code", header: "Code", accessorKey: "code", width: 100 },
      { id: "name", header: "Courier", accessorKey: "fullName" },
      { id: "zone", header: "Zone", accessorKey: "zoneName", width: 120 },
      {
        id: "status",
        header: "Live status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.liveStatus}
            tone={statusTone(row.liveStatus)}
          />
        ),
        width: 120,
      },
      {
        id: "assignments",
        header: "Jobs",
        accessorKey: "activeAssignments",
        align: "right",
        width: 70,
      },
      {
        id: "rating",
        header: "Rating",
        cell: ({ row }) => (
          <span className="tabular-nums">
            {row.rating.toFixed(1)} ({row.ratingCount})
          </span>
        ),
        align: "right",
        width: 110,
      },
      {
        id: "sla",
        header: "On-time %",
        cell: ({ row }) => (
          <span className="tabular-nums">{row.onTimePct.toFixed(1)}%</span>
        ),
        align: "right",
        width: 90,
      },
      {
        id: "vehicle",
        header: "Vehicle",
        accessorKey: "vehicleType",
        width: 110,
      },
      {
        id: "emergency",
        header: "Emergency",
        cell: ({ row }) =>
          row.emergency ? (
            <StatusBadge status="SOS" tone="danger" />
          ) : (
            <span className="text-[var(--nx-text-tertiary)]">—</span>
          ),
        width: 100,
      },
    ],
    [],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={40} />
        <Skeleton height={320} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load couriers
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
        title="Couriers"
        description={`${data.total} couriers · updated ${new Date(data.generatedAt).toLocaleTimeString("tr-TR")}${isFetching ? " · refreshing…" : ""}`}
        actions={
          <PermissionGate allowed={can(session, "couriers:write")}>
            <Button size="sm" variant="secondary">
              Invite courier
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
          placeholder="Search name, code, zone…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search couriers"
        />
        <Select
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          aria-label="Filter by status"
        >
          <option value="all">All statuses</option>
          <option value="available">Available</option>
          <option value="busy">Busy</option>
          <option value="break">Break</option>
          <option value="offline">Offline</option>
          <option value="emergency">Emergency</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={filtered}
        getRowId={(row) => row.id}
        onRowClick={(row) => router.push(`/couriers/${row.id}`)}
        emptyMessage="No couriers match filters"
      />
    </div>
  );
}
