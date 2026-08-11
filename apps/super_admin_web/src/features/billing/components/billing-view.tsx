"use client";

import { useMemo, useState } from "react";
import dynamic from "next/dynamic";
import {
  Button,
  ChartFrame,
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
import { can } from "@/shared/permissions/platform-permissions";
import { formatMinorUnits } from "@/shared/lib/money";
import { useBilling, useMarkInvoicePaid } from "../hooks";
import type {
  BillingMeter,
  FinOpsSeriesPoint,
  InvoiceStatus,
  PlatformInvoice,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function invoiceTone(
  status: InvoiceStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "paid":
      return "success";
    case "open":
      return "info";
    case "overdue":
      return "danger";
    case "draft":
      return "neutral";
    case "void":
      return "warning";
    default:
      return "neutral";
  }
}

function dualSeriesOption(
  spend: FinOpsSeriesPoint[],
  forecast: FinOpsSeriesPoint[],
) {
  return {
    grid: { left: 56, right: 16, top: 28, bottom: 28 },
    tooltip: { trigger: "axis" as const },
    legend: {
      data: ["MTD spend", "Forecast"],
      textStyle: { color: "var(--nx-text-secondary)", fontSize: 11 },
    },
    xAxis: {
      type: "category" as const,
      data: spend.map((p) => p.label),
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
      axisLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
    },
    yAxis: {
      type: "value" as const,
      splitLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
    },
    series: [
      {
        name: "MTD spend",
        type: "line" as const,
        smooth: true,
        data: spend.map((p) => Math.round(p.value / 100)),
        lineStyle: { width: 2, color: "#0B6E6E" },
        itemStyle: { color: "#0B6E6E" },
        areaStyle: { opacity: 0.1 },
        showSymbol: false,
      },
      {
        name: "Forecast",
        type: "line" as const,
        smooth: true,
        data: forecast.map((p) => Math.round(p.value / 100)),
        lineStyle: { width: 2, type: "dashed" as const, color: "#085858" },
        itemStyle: { color: "#085858" },
        showSymbol: false,
      },
    ],
  };
}

function breakdownOption(
  breakdown: { category: string; amountMinor: number; pct: number }[],
) {
  return {
    tooltip: { trigger: "item" as const },
    series: [
      {
        type: "pie" as const,
        radius: ["42%", "70%"],
        data: breakdown.map((b) => ({
          name: b.category,
          value: Math.round(b.amountMinor / 100),
        })),
        label: { color: "var(--nx-text-secondary)", fontSize: 11 },
        color: [
          "#0B6E6E",
          "#0f8585",
          "#085858",
          "#147a7a",
          "#3a9a9a",
          "#5cb0b0",
        ],
      },
    ],
  };
}

export function BillingView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } = useBilling();
  const markPaid = useMarkInvoicePaid();
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("all");

  const filtered = useMemo(() => {
    const items = data?.invoices ?? [];
    return items.filter((inv) => {
      if (status !== "all" && inv.status !== status) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        inv.tenantName.toLowerCase().includes(needle) ||
        inv.id.toLowerCase().includes(needle)
      );
    });
  }, [data?.invoices, q, status]);

  const meterCols = useMemo<DataGridColumnDef<BillingMeter>[]>(
    () => [
      {
        id: "cat",
        header: "Category",
        cell: ({ row }) => (
          <StatusBadge status={row.category} tone="info" />
        ),
        width: 120,
      },
      { id: "label", header: "Meter", accessorKey: "label" },
      {
        id: "usage",
        header: "Usage",
        cell: ({ row }) => (
          <span className="tabular-nums">
            {row.usage.toLocaleString("en-US")} {row.unit}
          </span>
        ),
      },
      {
        id: "amount",
        header: "Amount",
        cell: ({ row }) =>
          formatMinorUnits(row.amountMinor, row.currency),
        width: 120,
      },
      {
        id: "delta",
        header: "Δ",
        cell: ({ row }) => {
          const sign = row.deltaPct > 0 ? "+" : "";
          return (
            <span className="tabular-nums text-[12px]">
              {sign}
              {row.deltaPct.toFixed(1)}%
            </span>
          );
        },
        width: 80,
      },
    ],
    [],
  );

  const invoiceCols = useMemo<DataGridColumnDef<PlatformInvoice>[]>(
    () => [
      { id: "id", header: "Invoice", accessorKey: "id", width: 160 },
      { id: "tenant", header: "Tenant", accessorKey: "tenantName" },
      {
        id: "period",
        header: "Period",
        cell: ({ row }) => `${row.periodStart} → ${row.periodEnd}`,
        width: 180,
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={invoiceTone(row.status)} />
        ),
        width: 100,
      },
      {
        id: "total",
        header: "Total",
        cell: ({ row }) =>
          formatMinorUnits(row.totalMinor, row.currency),
        width: 120,
      },
      {
        id: "due",
        header: "Due",
        cell: ({ row }) =>
          new Date(row.dueAt).toLocaleDateString("en-US"),
        width: 110,
      },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "billing:write")}>
            <Button
              size="sm"
              variant="secondary"
              disabled={row.status === "paid" || row.status === "void"}
              loading={markPaid.isPending}
              onClick={() => void markPaid.mutateAsync(row.id)}
            >
              Mark paid
            </Button>
          </PermissionGate>
        ),
        width: 120,
      },
    ],
    [session, markPaid],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} height={96} />
          ))}
        </div>
        <Skeleton height={220} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load billing
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
        title="Billing & FinOps"
        description={`Tenant · API · storage · compute · courier · warehouse · monthly invoices${isFetching ? " · refreshing…" : ""}`}
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      />

      <section
        aria-label="FinOps KPIs"
        className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]"
      >
        <KpiCard
          title="MTD platform spend"
          value={formatMinorUnits(data.mtdSpendMinor, data.currency)}
          tone="brand"
        />
        <KpiCard
          title="Month forecast"
          value={formatMinorUnits(data.forecastMinor, data.currency)}
          tone="neutral"
        />
        <KpiCard
          title="Open invoices"
          value={String(data.invoices.filter((i) => i.status === "open").length)}
          tone="brand"
        />
        <KpiCard
          title="Overdue"
          value={String(
            data.invoices.filter((i) => i.status === "overdue").length,
          )}
          tone="danger"
        />
      </section>

      <section
        aria-label="FinOps charts"
        className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]"
      >
        <ChartFrame title="Spend vs forecast" description="USD major units">
          <ReactECharts
            option={dualSeriesOption(data.spendSeries, data.forecastSeries)}
            style={{ height: 220 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Cost breakdown" description="By billing category">
          <ReactECharts
            option={breakdownOption(data.breakdown)}
            style={{ height: 220 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Usage meters
        </h3>
        <DataGrid columns={meterCols} data={data.meters} getRowId={(r) => r.id} />
      </section>

      <FilterBar>
        <Input
          placeholder="Search invoice or tenant…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
        <Select value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="all">All statuses</option>
          <option value="draft">Draft</option>
          <option value="open">Open</option>
          <option value="paid">Paid</option>
          <option value="overdue">Overdue</option>
          <option value="void">Void</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={invoiceCols}
        data={filtered}
        getRowId={(r) => r.id}
        emptyMessage="No invoices"
      />
    </div>
  );
}
