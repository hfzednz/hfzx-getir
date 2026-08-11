"use client";

import dynamic from "next/dynamic";
import {
  ChartFrame,
  DataGrid,
  type DataGridColumnDef,
  KpiCard,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { usePermission } from "@/shared/permissions/use-permission";
import { useAiCommandSnapshot } from "../hooks";
import type {
  CampaignOpt,
  FraudAlert,
  PricingRec,
  SegmentRow,
  SeriesPoint,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function lineOption(points: SeriesPoint[], color: string, name: string) {
  return {
    grid: { left: 40, right: 12, top: 24, bottom: 28 },
    tooltip: { trigger: "axis" as const },
    xAxis: {
      type: "category" as const,
      data: points.map((p) => p.label),
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
    },
    yAxis: {
      type: "value" as const,
      splitLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
    },
    series: [
      {
        name,
        type: "line" as const,
        smooth: true,
        showSymbol: false,
        data: points.map((p) => p.value),
        lineStyle: { width: 2, color },
        itemStyle: { color },
        areaStyle: { opacity: 0.1 },
      },
    ],
  };
}

const fraudCols: DataGridColumnDef<FraudAlert>[] = [
  { id: "order", header: "Order", accessorKey: "orderId" },
  {
    id: "score",
    header: "Score",
    accessorKey: "score",
    align: "right",
    cell: ({ value }) => `${(Number(value) * 100).toFixed(0)}%`,
  },
  { id: "reason", header: "Reason", accessorKey: "reason" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => {
      const s = String(value);
      const tone =
        s === "blocked" || s === "open"
          ? "danger"
          : s === "reviewing"
            ? "warning"
            : "success";
      return <StatusBadge status={s} tone={tone} />;
    },
  },
];

const pricingCols: DataGridColumnDef<PricingRec>[] = [
  { id: "zone", header: "Zone", accessorKey: "zone" },
  { id: "cat", header: "Category", accessorKey: "skuCategory" },
  {
    id: "cur",
    header: "Current",
    accessorKey: "currentMultiplier",
    align: "right",
  },
  {
    id: "sug",
    header: "Suggested",
    accessorKey: "suggestedMultiplier",
    align: "right",
  },
  {
    id: "lift",
    header: "Lift %",
    accessorKey: "liftPct",
    align: "right",
  },
  {
    id: "conf",
    header: "Conf.",
    accessorKey: "confidence",
    align: "right",
    cell: ({ value }) => `${(Number(value) * 100).toFixed(0)}%`,
  },
];

const campaignCols: DataGridColumnDef<CampaignOpt>[] = [
  { id: "campaign", header: "Campaign", accessorKey: "campaign" },
  { id: "rec", header: "Recommendation", accessorKey: "recommendation" },
  {
    id: "roi",
    header: "ROI lift %",
    accessorKey: "expectedRoiLift",
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => {
      const s = String(value);
      return (
        <StatusBadge
          status={s}
          tone={
            s === "applied" ? "success" : s === "pending" ? "warning" : "neutral"
          }
        />
      );
    },
  },
];

const segmentCols: DataGridColumnDef<SegmentRow>[] = [
  { id: "name", header: "Segment", accessorKey: "name" },
  {
    id: "size",
    header: "Size",
    accessorKey: "size",
    align: "right",
    cell: ({ value }) => Number(value).toLocaleString("tr-TR"),
  },
  {
    id: "aov",
    header: "Avg AOV",
    accessorKey: "avgAovMinor",
    align: "right",
    cell: ({ value }) => formatMinorUnits(Number(value), "TRY"),
  },
  {
    id: "churn",
    header: "Churn risk",
    accessorKey: "churnRisk",
    align: "right",
    cell: ({ value }) => `${(Number(value) * 100).toFixed(0)}%`,
  },
];

export function AiCommandView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useAiCommandSnapshot();
  const canWrite = usePermission("ai:write");

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} height={88} />
          ))}
        </div>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load AI command center
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
        title="AI Command Center"
        description={`Demand, fraud, pricing & ops insights · ${new Date(data.generatedAt).toLocaleTimeString("tr-TR")}${isFetching ? " · refreshing…" : ""}`}
        actions={
          <PermissionGate allowed={canWrite}>
            <span className="text-[12px] text-[var(--nx-text-tertiary)]">
              Write actions enabled
            </span>
          </PermissionGate>
        }
      />

      <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        {data.kpis.map((kpi) => (
          <KpiCard
            key={kpi.id}
            title={kpi.label}
            value={
              kpi.unit === "percent"
                ? `${kpi.value}%`
                : kpi.unit === "currency"
                  ? formatMinorUnits(kpi.value, kpi.currency ?? "TRY")
                  : kpi.value.toLocaleString("tr-TR")
            }
            tone={kpi.tone}
          />
        ))}
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="Demand prediction" description="Orders / hour">
          <ReactECharts
            option={lineOption(data.demandForecast, "#0B6E6E", "Demand")}
            style={{ height: 190 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Inventory prediction" description="Critical SKU cover (units)">
          <ReactECharts
            option={lineOption(data.inventoryForecast, "#c45c26", "Cover")}
            style={{ height: 190 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="Delivery optimization" description="Predicted ETA (min)">
          <ReactECharts
            option={lineOption(data.deliveryOptSeries, "#0f8585", "ETA")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Recommendation CTR" description="Model monitoring">
          <ReactECharts
            option={lineOption(data.recommendationCtr, "#085858", "CTR %")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <section>
        <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
          Fraud detection
        </h3>
        <DataGrid
          columns={fraudCols}
          data={data.fraudAlerts}
          getRowId={(r) => r.id}
        />
      </section>

      <section>
        <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
          Pricing recommendations
        </h3>
        <DataGrid
          columns={pricingCols}
          data={data.pricingRecs}
          getRowId={(r) => r.id}
        />
      </section>

      <section>
        <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
          Campaign optimization
        </h3>
        <DataGrid
          columns={campaignCols}
          data={data.campaignOpts}
          getRowId={(r) => r.id}
        />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <div>
          <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
            Segmentation
          </h3>
          <DataGrid
            columns={segmentCols}
            data={data.segments}
            getRowId={(r) => r.id}
          />
        </div>
        <div className="grid grid-cols-1 gap-[var(--nx-space-3)]">
          <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
            <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
              Risk
            </h3>
            <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
              {data.risks.map((r) => (
                <li key={r.id} className="flex flex-col gap-[var(--nx-space-1)]">
                  <div className="flex items-center gap-[var(--nx-space-2)]">
                    <StatusBadge
                      status={r.severity}
                      tone={
                        r.severity === "high"
                          ? "danger"
                          : r.severity === "medium"
                            ? "warning"
                            : "info"
                      }
                    />
                    <span className="text-[13px] font-semibold">{r.area}</span>
                  </div>
                  <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                    {r.summary}
                  </p>
                </li>
              ))}
            </ul>
          </div>
          <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
            <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
              Operational insights
            </h3>
            <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
              {data.opsInsights.map((i) => (
                <li key={i.id}>
                  <p className="m-0 text-[13px] font-semibold">{i.title}</p>
                  <p className="m-0 mt-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
                    {i.detail}
                  </p>
                  <p className="m-0 mt-[var(--nx-space-1)] text-[11px] text-[var(--nx-text-tertiary)]">
                    Confidence {(i.confidence * 100).toFixed(0)}%
                  </p>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>
    </div>
  );
}
