"use client";

import { useMemo, useState } from "react";
import {
  Button,
  DataGrid,
  type DataGridColumnDef,
  FilterBar,
  PageHeader,
  Select,
  Skeleton,
  StatusBadge,
} from "@nexora/ui";
import { usePermission } from "@/shared/permissions/use-permission";
import { useNotificationsSnapshot } from "../hooks";
import type { AlertCategory, OpsAlert } from "../types";

const categories: Array<AlertCategory | "all"> = [
  "all",
  "operational",
  "stock",
  "security",
  "financial",
  "system",
  "emergency",
];

export function NotificationsView() {
  const { data, isLoading, isError, error, refetch } =
    useNotificationsSnapshot();
  const canWrite = usePermission("notifications:write");
  const [category, setCategory] = useState<AlertCategory | "all">("all");
  const [readFilter, setReadFilter] = useState<"all" | "unread" | "read">(
    "all",
  );
  const [local, setLocal] = useState<OpsAlert[] | null>(null);

  const alerts = useMemo(
    () => local ?? data?.alerts ?? [],
    [local, data?.alerts],
  );

  const filtered = useMemo(() => {
    return alerts.filter((a) => {
      if (category !== "all" && a.category !== category) return false;
      if (readFilter === "unread" && a.read) return false;
      if (readFilter === "read" && !a.read) return false;
      return true;
    });
  }, [alerts, category, readFilter]);

  const columns: DataGridColumnDef<OpsAlert>[] = [
    {
      id: "sev",
      header: "Severity",
      accessorKey: "severity",
      cell: ({ value }) => {
        const s = String(value) as OpsAlert["severity"];
        return (
          <StatusBadge
            status={s}
            tone={
              s === "danger" ? "danger" : s === "warning" ? "warning" : "info"
            }
          />
        );
      },
    },
    {
      id: "cat",
      header: "Category",
      accessorKey: "category",
      cell: ({ value }) => <StatusBadge status={String(value)} tone="neutral" />,
    },
    { id: "title", header: "Title", accessorKey: "title" },
    { id: "body", header: "Detail", accessorKey: "body" },
    {
      id: "when",
      header: "When",
      accessorKey: "createdAt",
      cell: ({ value }) => new Date(String(value)).toLocaleString("tr-TR"),
    },
    {
      id: "read",
      header: "Inbox",
      cell: ({ row }) => (
        <StatusBadge
          status={row.read ? "read" : "unread"}
          tone={row.read ? "neutral" : "info"}
        />
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) =>
        canWrite && !row.read ? (
          <Button
            size="sm"
            variant="ghost"
            onClick={() => {
              setLocal((prev) => {
                const base = prev ?? data?.alerts ?? [];
                return base.map((a) =>
                  a.id === row.id ? { ...a, read: true } : a,
                );
              });
            }}
          >
            Mark read
          </Button>
        ) : null,
    },
  ];

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
          Failed to load notifications
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

  const unread = alerts.filter((a) => !a.read).length;

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Notifications"
        description={`Ops alerts inbox · ${unread} unread`}
      />

      <FilterBar>
        <Select
          value={category}
          onChange={(e) =>
            setCategory(e.target.value as AlertCategory | "all")
          }
          aria-label="Category"
        >
          {categories.map((c) => (
            <option key={c} value={c}>
              {c === "all" ? "All categories" : c}
            </option>
          ))}
        </Select>
        <Select
          value={readFilter}
          onChange={(e) =>
            setReadFilter(e.target.value as "all" | "unread" | "read")
          }
          aria-label="Read state"
        >
          <option value="all">All</option>
          <option value="unread">Unread</option>
          <option value="read">Read</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={filtered}
        getRowId={(r) => r.id}
        emptyMessage="No alerts in this filter"
      />
    </div>
  );
}
