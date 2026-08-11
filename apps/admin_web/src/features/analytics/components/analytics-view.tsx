"use client";

import dynamic from "next/dynamic";
import {
  ChartFrame,
  DataGrid,
  type DataGridColumnDef,
  KpiCard,
  PageHeader,
  Skeleton,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { useAnalyticsSnapshot } from "../hooks";
import type {
  AnalyticsKpi,
  CourierPerfRow,
  SeriesPoint,
  WarehousePerfRow,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function formatKpi(kpi: AnalyticsKpi): string {
  switch (kpi.unit) {
    case "currency":
      return formatMinorUnits(kpi.value, kpi.currency ?? "TRY");
    case "percent":
      return `${kpi.value.toFixed(1)}%`;
    case "days":
      return `${kpi.value} d`;
    default:
      return kpi.value.toLocaleString("tr-TR");
  }
}

function lineOption(
  series: { name: string; color: string; points: SeriesPoint[] }[],
  valueDivisor = 1,
) {
  const labels = series[0]?.points.map((p) => p.label) ?? [];
  return {
    grid: { left: 48, right: 16, top: 28, bottom: 28 },
    tooltip: { trigger: "axis" as const },
    legend: {
      data: series.map((s) => s.name),
      textStyle: { color: "var(--nx-text-secondary)", fontSize: 11 },
    },
    xAxis: {
      type: "category" as const,
      data: labels,
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
      axisLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
    },
    yAxis: {
      type: "value" as const,
      splitLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
    },
    series: series.map((s) => ({
      name: s.name,
      type: "line" as const,
      smooth: true,
      showSymbol: false,
      data: s.points.map((p) =>
        valueDivisor > 1 ? Math.round(p.value / valueDivisor) : p.value,
      ),
      lineStyle: { width: 2, color: s.color },
      itemStyle: { color: s.color },
      areaStyle: series.length === 1 ? { opacity: 0.1 } : undefined,
    })),
  };
}

function barOption(labels: string[], values: number[], color: string) {
  return {
    grid: { left: 48, right: 12, top: 16, bottom: 48 },
    tooltip: { trigger: "axis" as const },
    xAxis: {
      type: "category" as const,
      data: labels,
      axisLabel: {
        color: "var(--nx-text-tertiary)",
        fontSize: 10,
        rotate: 20,
      },
    },
    yAxis: {
      type: "value" as const,
      splitLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
    },
    series: [
      {
        type: "bar" as const,
        data: values,
        itemStyle: { color },
        barMaxWidth: 36,
      },
    ],
  };
}

function funnelOption(
  steps: { label: string; count: number }[],
) {
  return {
    tooltip: { trigger: "item" as const },
    series: [
      {
        type: "funnel" as const,
        left: "8%",
        width: "84%",
        label: { color: "var(--nx-text-primary)", fontSize: 11 },
        data: steps.map((s) => ({ name: s.label, value: s.count })),
      },
    ],
  };
}

const warehouseCols: DataGridColumnDef<WarehousePerfRow>[] = [
  { id: "name", header: "Warehouse", accessorKey: "name" },
  {
    id: "orders",
    header: "Orders",
    accessorKey: "orders",
    align: "right",
    cell: ({ value }) => Number(value).toLocaleString("tr-TR"),
  },
  {
    id: "pick",
    header: "Pick (min)",
    accessorKey: "pickMins",
    align: "right",
  },
  {
    id: "sla",
    header: "SLA %",
    accessorKey: "slaPct",
    align: "right",
  },
  {
    id: "stockouts",
    header: "Stockouts",
    accessorKey: "stockouts",
    align: "right",
  },
];

const courierCols: DataGridColumnDef<CourierPerfRow>[] = [
  { id: "name", header: "Cohort", accessorKey: "name" },
  {
    id: "del",
    header: "Deliveries",
    accessorKey: "deliveries",
    align: "right",
    cell: ({ value }) => Number(value).toLocaleString("tr-TR"),
  },
  {
    id: "ontime",
    header: "On-time %",
    accessorKey: "onTimePct",
    align: "right",
  },
  {
    id: "avg",
    header: "Avg min",
    accessorKey: "avgMins",
    align: "right",
  },
  {
    id: "rating",
    header: "Rating",
    accessorKey: "rating",
    align: "right",
  },
];

export function AnalyticsView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useAnalyticsSnapshot();

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} height={88} />
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
          Failed to load analytics
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

  const cohortLabels = [...new Set(data.cohorts.map((c) => c.cohort))];
  const cohortWeeks = [...new Set(data.cohorts.map((c) => c.week))].sort(
    (a, b) => a - b,
  );
  const heatmapData = data.cohorts.map((c) => [
    c.week,
    cohortLabels.indexOf(c.cohort),
    c.retentionPct,
  ]);

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Analytics"
        description={`Sales, retention, funnels & forecasts · ${new Date(data.generatedAt).toLocaleTimeString("tr-TR")}${isFetching ? " · refreshing…" : ""}`}
      />

      <section className="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-8 gap-[var(--nx-space-3)]">
        {data.kpis.map((kpi) => (
          <KpiCard
            key={kpi.id}
            title={kpi.label}
            value={formatKpi(kpi)}
            delta={`${kpi.deltaPct > 0 ? "+" : ""}${kpi.deltaPct.toFixed(1)}%`}
            tone={kpi.tone}
          />
        ))}
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="Sales volume" description="Weekly orders">
          <ReactECharts
            option={lineOption([
              {
                name: "Orders",
                color: "#0B6E6E",
                points: data.salesSeries,
              },
            ])}
            style={{ height: 200 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Revenue" description="TRY (thousands)">
          <ReactECharts
            option={lineOption(
              [
                {
                  name: "Revenue",
                  color: "#0f8585",
                  points: data.revenueSeries,
                },
              ],
              100_00,
            )}
            style={{ height: 200 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-3 gap-[var(--nx-space-3)]">
        <ChartFrame title="Retention curve" description="D1–D90">
          <ReactECharts
            option={lineOption([
              {
                name: "Retention %",
                color: "#085858",
                points: data.retentionSeries,
              },
            ])}
            style={{ height: 200 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Conversion funnel" description="Browse → paid">
          <ReactECharts
            option={funnelOption(data.funnel)}
            style={{ height: 200 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Checkout conversion" description="Daily %">
          <ReactECharts
            option={lineOption([
              {
                name: "Conv %",
                color: "#c45c26",
                points: data.conversionSeries,
              },
            ])}
            style={{ height: 200 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="Cohort retention" description="Week-over-week %">
          <ReactECharts
            option={{
              tooltip: { position: "top" },
              grid: { left: 48, right: 24, top: 16, bottom: 28 },
              xAxis: {
                type: "category" as const,
                data: cohortWeeks.map((w) => `W${w}`),
                splitArea: { show: true },
              },
              yAxis: {
                type: "category" as const,
                data: cohortLabels,
                splitArea: { show: true },
              },
              visualMap: {
                min: 10,
                max: 100,
                calculable: true,
                orient: "horizontal",
                left: "center",
                bottom: 0,
                inRange: { color: ["#e8f4f4", "#0B6E6E"] },
                textStyle: { fontSize: 10 },
              },
              series: [
                {
                  type: "heatmap" as const,
                  data: heatmapData,
                  label: { show: true, fontSize: 9 },
                },
              ],
            }}
            style={{ height: 260 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Customer lifetime value" description="Avg CLV index">
          <ReactECharts
            option={lineOption([
              {
                name: "CLV",
                color: "#0B6E6E",
                points: data.clvSeries,
              },
            ])}
            style={{ height: 260 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <div>
          <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
            Warehouse performance
          </h3>
          <DataGrid
            columns={warehouseCols}
            data={data.warehouses}
            getRowId={(r) => r.id}
          />
        </div>
        <div>
          <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
            Courier performance
          </h3>
          <DataGrid
            columns={courierCols}
            data={data.couriers}
            getRowId={(r) => r.id}
          />
        </div>
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-3 gap-[var(--nx-space-3)]">
        <ChartFrame title="Cancellation reasons" description="Share of cancels">
          <ReactECharts
            option={barOption(
              data.cancelReasons.map((c) => c.reason),
              data.cancelReasons.map((c) => c.count),
              "#b42318",
            )}
            style={{ height: 220 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Refund analysis" description="Amount (TRY thousands)">
          <ReactECharts
            option={barOption(
              data.refunds.map((r) => r.label),
              data.refunds.map((r) => Math.round(r.amountMinor / 100_00)),
              "#c45c26",
            )}
            style={{ height: 220 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Demand forecast" description="Actual vs predicted">
          <ReactECharts
            option={lineOption([
              {
                name: "Actual",
                color: "#0B6E6E",
                points: data.demandActual,
              },
              {
                name: "Forecast",
                color: "#7a9a9a",
                points: data.demandForecast,
              },
            ])}
            style={{ height: 220 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>
    </div>
  );
}
