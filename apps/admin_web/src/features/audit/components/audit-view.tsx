"use client";

import { useMemo, useState } from "react";
import {
  Button,
  DataGrid,
  type DataGridColumnDef,
  FilterBar,
  Input,
  PageHeader,
  Select,
  Skeleton,
} from "@nexora/ui";
import { useAuditSnapshot } from "../hooks";
import type { AuditEvent } from "../types";

const columns: DataGridColumnDef<AuditEvent>[] = [
  {
    id: "when",
    header: "When",
    accessorKey: "when",
    cell: ({ value }) => new Date(String(value)).toLocaleString("tr-TR"),
    width: 150,
  },
  { id: "who", header: "Who", accessorKey: "who" },
  { id: "where", header: "Where", accessorKey: "where" },
  { id: "device", header: "Device", accessorKey: "device" },
  { id: "action", header: "Action", accessorKey: "action" },
  { id: "resource", header: "Resource", accessorKey: "resource" },
  { id: "old", header: "Old", accessorKey: "oldValue" },
  { id: "new", header: "New", accessorKey: "newValue" },
  { id: "ip", header: "IP", accessorKey: "ip" },
  { id: "session", header: "Session", accessorKey: "sessionId" },
];

export function AuditView() {
  const { data, isLoading, isError, error, refetch } = useAuditSnapshot();
  const [q, setQ] = useState("");
  const [actionFilter, setActionFilter] = useState("all");

  const filtered = useMemo(() => {
    const events = data?.events ?? [];
    const needle = q.trim().toLowerCase();
    return events.filter((e) => {
      if (actionFilter !== "all" && !e.action.startsWith(actionFilter)) {
        return false;
      }
      if (!needle) return true;
      const hay = [
        e.who,
        e.where,
        e.device,
        e.action,
        e.resource,
        e.oldValue,
        e.newValue,
        e.ip,
        e.sessionId,
      ]
        .join(" ")
        .toLowerCase();
      return hay.includes(needle);
    });
  }, [data?.events, q, actionFilter]);

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
          Failed to load audit log
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
        title="Audit logs"
        description="Who / when / where / device / old / new / IP / session"
      />

      <FilterBar
        actions={
          <Button
            size="sm"
            variant="secondary"
            onClick={() => {
              setQ("");
              setActionFilter("all");
            }}
          >
            Reset
          </Button>
        }
      >
        <Input
          placeholder="Search who, IP, session, resource…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search audit events"
        />
        <Select
          value={actionFilter}
          onChange={(e) => setActionFilter(e.target.value)}
          aria-label="Filter by action prefix"
        >
          <option value="all">All actions</option>
          <option value="orders">orders.*</option>
          <option value="finance">finance.*</option>
          <option value="system">system.*</option>
          <option value="rbac">rbac.*</option>
          <option value="couriers">couriers.*</option>
          <option value="customers">customers.*</option>
          <option value="inventory">inventory.*</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={filtered}
        getRowId={(r) => r.id}
        emptyMessage="No audit events match filters"
      />
    </div>
  );
}
