"use client";

import dynamic from "next/dynamic";
import {
  ChartFrame,
  DataGrid,
  type DataGridColumnDef,
  KpiCard,
  PageHeader,
  Skeleton,
  StatusBadge,
} from "@nexora/ui";
import { useMonitoringSnapshot } from "../hooks";
import type { HealthService, QueueStat, SeriesPoint } from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function lineOption(points: SeriesPoint[], color: string) {
  return {
    grid: { left: 40, right: 12, top: 16, bottom: 28 },
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

const serviceCols: DataGridColumnDef<HealthService>[] = [
  { id: "name", header: "Service", accessorKey: "name" },
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
            s === "healthy" ? "success" : s === "degraded" ? "warning" : "danger"
          }
        />
      );
    },
  },
  {
    id: "lat",
    header: "Latency",
    accessorKey: "latencyMs",
    align: "right",
    cell: ({ value }) => `${value} ms`,
  },
];

const queueCols: DataGridColumnDef<QueueStat>[] = [
  { id: "name", header: "Queue", accessorKey: "name" },
  {
    id: "depth",
    header: "Depth",
    accessorKey: "depth",
    align: "right",
  },
  {
    id: "lag",
    header: "Lag (s)",
    accessorKey: "lagSec",
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
            s === "ok" ? "success" : s === "warn" ? "warning" : "danger"
          }
        />
      );
    },
  },
];

export function MonitoringView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useMonitoringSnapshot();

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
          {Array.from({ length: 4 }).map((_, i) => (
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
          Failed to load monitoring
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
        title="Monitoring"
        description={`Platform health · ${new Date(data.generatedAt).toLocaleTimeString("tr-TR")}${isFetching ? " · refreshing…" : ""}`}
      />

      <section className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
        <KpiCard
          title="DB connections"
          value={String(data.dbConnections)}
          tone="neutral"
        />
        <KpiCard
          title="Slow queries"
          value={String(data.dbSlowQueries)}
          tone={data.dbSlowQueries > 5 ? "warning" : "success"}
        />
        <KpiCard
          title="Storage used"
          value={`${data.storagePct}%`}
          tone={data.storagePct > 80 ? "warning" : "brand"}
        />
        <KpiCard
          title="WebSocket"
          value={`${data.websocket.clients} clients`}
          delta={`${data.websocket.msgPerSec}/s · ${data.websocket.status}`}
          tone={
            data.websocket.status === "connected"
              ? "success"
              : data.websocket.status === "degraded"
                ? "warning"
                : "danger"
          }
        />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="API latency" description="p95 ms">
          <ReactECharts
            option={lineOption(data.apiLatency, "#0B6E6E")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Server load" description="Load average">
          <ReactECharts
            option={lineOption(data.serverLoad, "#0f8585")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="CPU %" description="Cluster average">
          <ReactECharts
            option={lineOption(data.cpu, "#c45c26")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Memory %" description="RSS utilization">
          <ReactECharts
            option={lineOption(data.memory, "#085858")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <div>
          <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
            System health
          </h3>
          <DataGrid
            columns={serviceCols}
            data={data.services}
            getRowId={(r) => r.id}
          />
        </div>
        <div>
          <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
            Queues
          </h3>
          <DataGrid
            columns={queueCols}
            data={data.queues}
            getRowId={(r) => r.id}
          />
        </div>
      </section>
    </div>
  );
}
