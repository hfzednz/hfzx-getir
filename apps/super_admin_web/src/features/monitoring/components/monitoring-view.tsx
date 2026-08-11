"use client";

import { useMemo, useState } from "react";
import dynamic from "next/dynamic";
import {
  Button,
  ChartFrame,
  DataGrid,
  FilterBar,
  KpiCard,
  PageHeader,
  Select,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { lineChartOption } from "@/shared/lib/charts";
import { useMonitoring } from "../hooks";
import type { MonitorResource, ResourceHealth } from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function healthTone(
  s: ResourceHealth,
): "success" | "warning" | "danger" | "neutral" {
  if (s === "healthy") return "success";
  if (s === "degraded") return "warning";
  if (s === "critical") return "danger";
  return "neutral";
}

export function MonitoringView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useMonitoring();
  const [category, setCategory] = useState("all");

  const filtered = useMemo(() => {
    const items = data?.resources ?? [];
    if (category === "all") return items;
    return items.filter((r) => r.category === category);
  }, [data?.resources, category]);

  const cols = useMemo<DataGridColumnDef<MonitorResource>[]>(
    () => [
      { id: "name", header: "Resource", accessorKey: "name" },
      {
        id: "cat",
        header: "Category",
        cell: ({ row }) => (
          <StatusBadge status={row.category} tone="info" />
        ),
        width: 120,
      },
      { id: "metric", header: "Metric", accessorKey: "metric", width: 120 },
      {
        id: "value",
        header: "Value",
        cell: ({ row }) => `${row.value} ${row.unit}`,
        width: 110,
      },
      {
        id: "thr",
        header: "Threshold",
        cell: ({ row }) => `${row.threshold} ${row.unit}`,
        width: 110,
      },
      {
        id: "status",
        header: "Health",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={healthTone(row.status)} />
        ),
        width: 110,
      },
      { id: "region", header: "Region", accessorKey: "region", width: 120 },
    ],
    [],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <div className="grid grid-cols-2 md:grid-cols-5 gap-[var(--nx-space-3)]">
          {Array.from({ length: 10 }).map((_, i) => (
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

  const k = data.kpis;

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Monitoring"
        description={`Platform-wide CPU · mem · disk · DB · API · queues · WS · Redis · OpenSearch · K8s · cloud${isFetching ? " · refreshing…" : ""}`}
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      />

      <div className="grid grid-cols-2 md:grid-cols-5 gap-[var(--nx-space-3)]">
        <KpiCard title="CPU avg" value={`${k.cpuAvgPct}%`} />
        <KpiCard title="Mem avg" value={`${k.memAvgPct}%`} />
        <KpiCard title="Disk avg" value={`${k.diskAvgPct}%`} />
        <KpiCard title="API p95" value={`${k.apiP95Ms} ms`} />
        <KpiCard title="Queue depth" value={k.queueDepth.toLocaleString("en-US")} />
        <KpiCard title="WS connections" value={k.wsConnections.toLocaleString("en-US")} />
        <KpiCard title="Redis hit" value={`${k.redisHitPct}%`} tone="success" />
        <KpiCard title="OpenSearch lag" value={`${k.opensearchLagMs} ms`} />
        <KpiCard
          title="K8s not ready"
          value={String(k.k8sPodsNotReady)}
          tone={k.k8sPodsNotReady ? "warning" : "success"}
        />
        <KpiCard
          title="Cloud alerts"
          value={String(k.cloudAlerts)}
          tone={k.cloudAlerts ? "danger" : "success"}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-[var(--nx-space-3)]">
        <ChartFrame title="CPU %">
          <ReactECharts
            style={{ height: 200 }}
            option={lineChartOption(data.cpuSeries, "#0B6E6E", "%")}
          />
        </ChartFrame>
        <ChartFrame title="Memory %">
          <ReactECharts
            style={{ height: 200 }}
            option={lineChartOption(data.memSeries, "#147A7A", "%")}
          />
        </ChartFrame>
        <ChartFrame title="API latency (ms)">
          <ReactECharts
            style={{ height: 200 }}
            option={lineChartOption(data.latencySeries, "#C45C26", "ms")}
          />
        </ChartFrame>
      </div>

      <FilterBar>
        <Select
          value={category}
          onChange={(e) => setCategory(e.target.value)}
          aria-label="Filter category"
        >
          <option value="all">All categories</option>
          <option value="compute">CPU / mem</option>
          <option value="disk">Disk</option>
          <option value="database">Database</option>
          <option value="api">API</option>
          <option value="queue">Queues</option>
          <option value="websocket">WebSocket</option>
          <option value="redis">Redis</option>
          <option value="opensearch">OpenSearch</option>
          <option value="kubernetes">Kubernetes</option>
          <option value="cloud">Cloud</option>
        </Select>
      </FilterBar>

      <DataGrid columns={cols} data={filtered} getRowId={(r) => r.id} />
    </div>
  );
}
