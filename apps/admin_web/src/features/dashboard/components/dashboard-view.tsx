"use client";

import dynamic from "next/dynamic";
import {
  ChartFrame,
  KpiCard,
  PageHeader,
  Skeleton,
  StatusBadge,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { useDashboardSnapshot } from "../hooks";
import type { DashboardKpi, DashboardSeriesPoint } from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function formatKpiValue(kpi: DashboardKpi): string {
  switch (kpi.unit) {
    case "currency":
      return formatMinorUnits(kpi.value, kpi.currency ?? "TRY");
    case "percent":
      return `${kpi.value.toFixed(1)}%`;
    case "minutes":
      return `${kpi.value} min`;
    default:
      return kpi.value.toLocaleString("tr-TR");
  }
}

function formatDelta(deltaPct: number): string {
  const sign = deltaPct > 0 ? "+" : "";
  return `${sign}${deltaPct.toFixed(1)}% vs yesterday`;
}

function seriesOption(
  points: DashboardSeriesPoint[],
  color: string,
  yName?: string,
) {
  return {
    grid: { left: 40, right: 12, top: 24, bottom: 28 },
    tooltip: { trigger: "axis" as const },
    xAxis: {
      type: "category" as const,
      data: points.map((p) => p.label),
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
      axisLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
    },
    yAxis: {
      type: "value" as const,
      name: yName,
      splitLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
    },
    series: [
      {
        type: "line" as const,
        smooth: true,
        data: points.map((p) =>
          p.value > 10_000 ? Math.round(p.value / 100) : p.value,
        ),
        areaStyle: { opacity: 0.12 },
        lineStyle: { width: 2, color },
        itemStyle: { color },
        showSymbol: false,
      },
    ],
  };
}

export function DashboardView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useDashboardSnapshot();

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} height={96} />
          ))}
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-[var(--nx-space-3)]">
          <Skeleton height={220} />
          <Skeleton height={220} />
          <Skeleton height={220} />
        </div>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load dashboard
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
        title="Operations dashboard"
        description={`City scope · updated ${new Date(data.generatedAt).toLocaleTimeString("tr-TR")}${isFetching ? " · refreshing…" : ""}`}
      />

      <section
        aria-label="KPIs"
        className="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-8 gap-[var(--nx-space-3)]"
      >
        {data.kpis.map((kpi) => (
          <KpiCard
            key={kpi.id}
            title={kpi.label}
            value={formatKpiValue(kpi)}
            delta={formatDelta(kpi.deltaPct)}
            tone={kpi.tone}
          />
        ))}
      </section>

      <section
        aria-label="Charts"
        className="grid grid-cols-1 lg:grid-cols-3 gap-[var(--nx-space-3)]"
      >
        <ChartFrame title="Revenue" description="Today by hour (TRY)">
          <ReactECharts
            option={seriesOption(data.revenueSeries, "#0B6E6E")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Orders" description="Completed volume">
          <ReactECharts
            option={seriesOption(data.ordersSeries, "#0f8585")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="SLA" description="On-time %">
          <ReactECharts
            option={seriesOption(data.slaSeries, "#085858", "%")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <section
        aria-label="Ops panels"
        className="grid grid-cols-1 lg:grid-cols-3 gap-[var(--nx-space-3)]"
      >
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            Live alerts
          </h3>
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
            {data.alerts.map((a) => (
              <li key={a.id} className="flex flex-col gap-[var(--nx-space-1)]">
                <div className="flex items-center gap-[var(--nx-space-2)]">
                  <StatusBadge
                    status={a.severity}
                    tone={
                      a.severity === "danger"
                        ? "danger"
                        : a.severity === "warning"
                          ? "warning"
                          : "info"
                    }
                  />
                  <span className="text-[13px] font-semibold text-[var(--nx-text-primary)]">
                    {a.title}
                  </span>
                </div>
                <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                  {a.detail}
                </p>
              </li>
            ))}
          </ul>
        </div>

        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            AI insights
          </h3>
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
            {data.aiInsights.map((insight) => (
              <li key={insight.id}>
                <p className="m-0 text-[13px] font-semibold text-[var(--nx-text-primary)]">
                  {insight.title}
                </p>
                <p className="m-0 mt-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
                  {insight.summary}
                </p>
                <p className="m-0 mt-[var(--nx-space-1)] text-[11px] text-[var(--nx-text-tertiary)] tabular-nums">
                  Confidence {(insight.confidence * 100).toFixed(0)}%
                </p>
              </li>
            ))}
          </ul>
        </div>

        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            System health
          </h3>
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
            {data.systemHealth.map((h) => (
              <li
                key={h.id}
                className="flex items-center justify-between gap-[var(--nx-space-2)] py-[var(--nx-space-1)] border-b border-[var(--nx-border-subtle)] last:border-0"
              >
                <span className="text-[13px] font-medium">{h.name}</span>
                <div className="flex items-center gap-[var(--nx-space-2)]">
                  <span className="text-[11px] tabular-nums text-[var(--nx-text-tertiary)]">
                    {h.latencyMs} ms
                  </span>
                  <StatusBadge
                    status={h.status}
                    tone={
                      h.status === "healthy"
                        ? "success"
                        : h.status === "degraded"
                          ? "warning"
                          : "danger"
                    }
                  />
                </div>
              </li>
            ))}
          </ul>
        </div>
      </section>
    </div>
  );
}
