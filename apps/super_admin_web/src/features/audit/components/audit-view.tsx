"use client";

import { useMemo, useState } from "react";
import {
  Button,
  DataGrid,
  FilterBar,
  Input,
  PageHeader,
  Select,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useAudit } from "../hooks";
import type { AuditEntry, AuditSeverity } from "../types";

function sevTone(s: AuditSeverity): "info" | "warning" | "danger" {
  if (s === "critical") return "danger";
  if (s === "warning") return "warning";
  return "info";
}

export function AuditView() {
  const [q, setQ] = useState("");
  const [severity, setSeverity] = useState("all");
  const { data, isLoading, isError, error, refetch, isFetching } = useAudit(q);

  const filtered = useMemo(() => {
    const items = data?.items ?? [];
    if (severity === "all") return items;
    return items.filter((e) => e.severity === severity);
  }, [data?.items, severity]);

  const cols = useMemo<DataGridColumnDef<AuditEntry>[]>(
    () => [
      {
        id: "when",
        header: "When",
        cell: ({ row }) => new Date(row.when).toLocaleString("en-US"),
        width: 160,
      },
      { id: "who", header: "Who", accessorKey: "actorEmail", width: 180 },
      { id: "action", header: "Action", accessorKey: "action" },
      {
        id: "resource",
        header: "Resource",
        cell: ({ row }) => `${row.resource}:${row.resourceId}`,
      },
      { id: "where", header: "Where", accessorKey: "where", width: 120 },
      { id: "device", header: "Device", accessorKey: "device", width: 150 },
      { id: "ip", header: "IP", accessorKey: "ip", width: 120 },
      { id: "session", header: "Session", accessorKey: "sessionId", width: 130 },
      {
        id: "old",
        header: "Old",
        cell: ({ row }) => (
          <span className="font-mono text-[11px]">{row.oldValue ?? "—"}</span>
        ),
      },
      {
        id: "new",
        header: "New",
        cell: ({ row }) => (
          <span className="font-mono text-[11px]">{row.newValue ?? "—"}</span>
        ),
      },
      {
        id: "sev",
        header: "Sev",
        cell: ({ row }) => (
          <StatusBadge status={row.severity} tone={sevTone(row.severity)} />
        ),
        width: 90,
      },
      {
        id: "sealed",
        header: "Seal",
        cell: ({ row }) => (
          <StatusBadge
            status={row.sealed ? "immutable" : "open"}
            tone={row.sealed ? "success" : "warning"}
          />
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
        <Skeleton height={280} />
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
        title="Audit"
        description={`Immutable platform audit · ${data.total} entries · sealed=${String(data.immutable)}${isFetching ? " · refreshing…" : ""}`}
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      />

      <FilterBar>
        <Input
          placeholder="Search actor, action, IP, session…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search audit"
        />
        <Select
          value={severity}
          onChange={(e) => setSeverity(e.target.value)}
          aria-label="Filter severity"
        >
          <option value="all">All severities</option>
          <option value="info">Info</option>
          <option value="warning">Warning</option>
          <option value="critical">Critical</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={cols}
        data={filtered}
        getRowId={(r) => r.id}
        emptyMessage="No audit entries"
      />
    </div>
  );
}
