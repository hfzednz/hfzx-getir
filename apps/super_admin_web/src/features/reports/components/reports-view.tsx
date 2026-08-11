"use client";

import { useMemo, useState } from "react";
import {
  Button,
  DataGrid,
  FilterBar,
  PageHeader,
  PermissionGate,
  Select,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/platform-permissions";
import { useExportReport, useReports } from "../hooks";
import type {
  ExportFormat,
  ReportDefinition,
  ReportExportResult,
} from "../types";

export function ReportsView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } = useReports();
  const exportMutation = useExportReport();
  const [category, setCategory] = useState("all");
  const [lastExport, setLastExport] = useState<ReportExportResult | null>(null);

  const filtered = useMemo(() => {
    const items = data?.reports ?? [];
    if (category === "all") return items;
    return items.filter((r) => r.category === category);
  }, [data?.reports, category]);

  const cols = useMemo<DataGridColumnDef<ReportDefinition>[]>(
    () => [
      { id: "name", header: "Report", accessorKey: "name" },
      {
        id: "cat",
        header: "Category",
        cell: ({ row }) => (
          <StatusBadge status={row.category} tone="info" />
        ),
        width: 110,
      },
      { id: "desc", header: "Description", accessorKey: "description" },
      {
        id: "sched",
        header: "Schedule",
        cell: ({ row }) => row.schedule ?? "on-demand",
        width: 100,
      },
      { id: "owner", header: "Owner", accessorKey: "owner", width: 140 },
      {
        id: "export",
        header: "Export",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "reports:export")}>
            <div
              className="flex gap-[var(--nx-space-1)]"
              onClick={(e) => e.stopPropagation()}
            >
              {(["csv", "json", "pdf"] as ExportFormat[]).map((format) => (
                <Button
                  key={format}
                  size="sm"
                  variant="secondary"
                  loading={
                    exportMutation.isPending &&
                    exportMutation.variables?.reportId === row.id &&
                    exportMutation.variables?.format === format
                  }
                  onClick={() =>
                    void exportMutation
                      .mutateAsync({ reportId: row.id, format })
                      .then(setLastExport)
                  }
                >
                  {format.toUpperCase()}
                </Button>
              ))}
            </div>
          </PermissionGate>
        ),
        width: 220,
      },
    ],
    [session, exportMutation],
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
          Failed to load reports
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
        title="Reports"
        description={`Platform · infra · financial · compliance · security · AI · tenant${isFetching ? " · refreshing…" : ""}`}
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      />

      <FilterBar>
        <Select
          value={category}
          onChange={(e) => setCategory(e.target.value)}
          aria-label="Filter category"
        >
          <option value="all">All categories</option>
          <option value="platform">Platform</option>
          <option value="infra">Infra</option>
          <option value="financial">Financial</option>
          <option value="compliance">Compliance</option>
          <option value="security">Security</option>
          <option value="ai">AI</option>
          <option value="tenant">Tenant</option>
        </Select>
      </FilterBar>

      <DataGrid columns={cols} data={filtered} getRowId={(r) => r.id} />

      {lastExport ? (
        <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
            Last export preview
          </h3>
          <p className="m-0 mb-[var(--nx-space-2)] text-[12px] text-[var(--nx-text-secondary)]">
            {lastExport.filename} · {lastExport.mimeType} ·{" "}
            {lastExport.byteLength} bytes ·{" "}
            {new Date(lastExport.exportedAt).toLocaleString("en-US")}
          </p>
          <pre className="m-0 p-[var(--nx-space-3)] overflow-auto max-h-48 text-[11px] font-mono bg-[var(--nx-bg-elevated)] rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-subtle)]">
            {lastExport.preview}
          </pre>
        </section>
      ) : null}
    </div>
  );
}
