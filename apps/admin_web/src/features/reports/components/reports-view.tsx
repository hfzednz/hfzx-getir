"use client";

import { useMemo, useState } from "react";
import {
  Button,
  DataGrid,
  type DataGridColumnDef,
  FilterBar,
  PageHeader,
  PermissionGate,
  Select,
  Skeleton,
  StatusBadge,
} from "@nexora/ui";
import { usePermission } from "@/shared/permissions/use-permission";
import {
  buildExportPayload,
  mockDownload,
} from "../api";
import { useReportsCatalog } from "../hooks";
import type { ExportFormat, ReportDomain, ReportTemplate } from "../types";

const domains: Array<ReportDomain | "all"> = [
  "all",
  "orders",
  "customers",
  "products",
  "inventory",
  "couriers",
  "warehouses",
  "finance",
  "crm",
  "campaigns",
  "performance",
  "taxes",
  "operations",
];

export function ReportsView() {
  const { data, isLoading, isError, error, refetch } = useReportsCatalog();
  const canExport = usePermission("reports:export");
  const [domain, setDomain] = useState<ReportDomain | "all">("all");
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const templates = useMemo(() => {
    const list = data?.templates ?? [];
    if (domain === "all") return list;
    return list.filter((t) => t.domain === domain);
  }, [data?.templates, domain]);

  const selected: ReportTemplate | undefined =
    templates.find((t) => t.id === selectedId) ?? templates[0];

  const listCols: DataGridColumnDef<ReportTemplate>[] = [
    { id: "name", header: "Report", accessorKey: "name" },
    {
      id: "domain",
      header: "Domain",
      accessorKey: "domain",
      cell: ({ value }) => <StatusBadge status={String(value)} tone="info" />,
    },
    { id: "desc", header: "Description", accessorKey: "description" },
  ];

  const previewCols: DataGridColumnDef<Record<string, unknown>>[] =
    selected?.columns.map((c) => ({
      id: c,
      header: c,
      accessorKey: c,
    })) ?? [];

  function exportAs(format: ExportFormat) {
    if (!selected) return;
    const payload = buildExportPayload(
      selected.id,
      format,
      selected.sampleRows,
      selected.columns,
    );
    mockDownload(payload.filename, payload.content, payload.mime);
  }

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
        description="Generate domain reports and export CSV / Excel / JSON / PDF (mock client download)"
      />

      <FilterBar>
        <Select
          value={domain}
          onChange={(e) => setDomain(e.target.value as ReportDomain | "all")}
          aria-label="Domain"
        >
          {domains.map((d) => (
            <option key={d} value={d}>
              {d === "all" ? "All domains" : d}
            </option>
          ))}
        </Select>
      </FilterBar>

      <DataGrid
        columns={listCols}
        data={templates}
        getRowId={(r) => r.id}
        onRowClick={(row) => setSelectedId(row.id)}
      />

      {selected ? (
        <div className="flex flex-col gap-[var(--nx-space-3)]">
          <div className="flex flex-wrap items-center justify-between gap-[var(--nx-space-2)]">
            <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
              Preview · {selected.name}
            </h3>
            <PermissionGate
              allowed={canExport}
              fallback={
                <span className="text-[12px] text-[var(--nx-text-tertiary)]">
                  Export requires reports:export
                </span>
              }
            >
              <div className="flex flex-wrap gap-[var(--nx-space-2)]">
                {(["csv", "excel", "json", "pdf"] as ExportFormat[]).map(
                  (fmt) => (
                    <Button
                      key={fmt}
                      size="sm"
                      variant="secondary"
                      onClick={() => exportAs(fmt)}
                    >
                      Export {fmt.toUpperCase()}
                    </Button>
                  ),
                )}
              </div>
            </PermissionGate>
          </div>
          <DataGrid
            columns={previewCols}
            data={selected.sampleRows as Record<string, unknown>[]}
            getRowId={(_, i) => String(i)}
          />
        </div>
      ) : null}
    </div>
  );
}
