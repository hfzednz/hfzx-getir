"use client";

import dynamic from "next/dynamic";
import {
  ChartFrame,
  DataGrid,
  type DataGridColumnDef,
  KpiCard,
  PageHeader,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@nexora/ui";
import { lineChartOption } from "@/shared/lib/charts";
import { ModuleError, ModuleLoading, healthTone } from "@/shared/ui/module-state";
import { useAiPlatformSnapshot } from "../hooks";
import type { AiModel, GpuNode, InferenceServer, TrainingJob } from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function StatusCell({ value }: { value: unknown }) {
  const s = String(value);
  return <StatusBadge status={s} tone={healthTone(s)} />;
}

const modelCols: DataGridColumnDef<AiModel>[] = [
  { id: "name", header: "Model", accessorKey: "name" },
  { id: "kind", header: "Kind", accessorKey: "kind" },
  { id: "version", header: "Version", accessorKey: "version" },
  { id: "fw", header: "Framework", accessorKey: "framework" },
  {
    id: "acc",
    header: "Accuracy",
    cell: ({ row }) =>
      row.accuracyPct == null ? "—" : `${row.accuracyPct}%`,
    align: "right",
  },
  {
    id: "p99",
    header: "p99 ms",
    cell: ({ row }) =>
      row.latencyP99Ms == null ? "—" : String(row.latencyP99Ms),
    align: "right",
  },
  {
    id: "rpm",
    header: "Req/min",
    accessorKey: "requestsPerMin",
    align: "right",
  },
  { id: "owner", header: "Owner", accessorKey: "owner" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const trainCols: DataGridColumnDef<TrainingJob>[] = [
  { id: "model", header: "Model", accessorKey: "modelName" },
  { id: "kind", header: "Kind", accessorKey: "kind" },
  { id: "dataset", header: "Dataset", accessorKey: "dataset" },
  { id: "gpus", header: "GPUs", accessorKey: "gpus", align: "right" },
  {
    id: "prog",
    header: "Progress",
    cell: ({ row }) => `${row.progressPct}%`,
    align: "right",
  },
  {
    id: "eta",
    header: "ETA min",
    cell: ({ row }) => (row.etaMin == null ? "—" : String(row.etaMin)),
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const inferCols: DataGridColumnDef<InferenceServer>[] = [
  { id: "name", header: "Server", accessorKey: "name" },
  { id: "models", header: "Models", accessorKey: "models" },
  { id: "region", header: "Region", accessorKey: "region" },
  { id: "runtime", header: "Runtime", accessorKey: "runtime" },
  {
    id: "replicas",
    header: "Replicas",
    accessorKey: "replicas",
    align: "right",
  },
  { id: "qps", header: "QPS", accessorKey: "qps", align: "right" },
  {
    id: "gpu",
    header: "GPU %",
    accessorKey: "gpuUtilPct",
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const gpuCols: DataGridColumnDef<GpuNode>[] = [
  { id: "host", header: "Host", accessorKey: "hostname" },
  { id: "region", header: "Region", accessorKey: "region" },
  { id: "model", header: "GPU", accessorKey: "gpuModel" },
  { id: "count", header: "Count", accessorKey: "gpuCount", align: "right" },
  { id: "util", header: "Util %", accessorKey: "utilPct", align: "right" },
  {
    id: "mem",
    header: "Mem %",
    accessorKey: "memUsedPct",
    align: "right",
  },
  { id: "temp", header: "Temp °C", accessorKey: "tempC", align: "right" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

export function AiPlatformView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useAiPlatformSnapshot();

  if (isLoading) return <ModuleLoading />;
  if (isError || !data) {
    return (
      <ModuleError
        title="Failed to load AI platform"
        message={error instanceof Error ? error.message : "Unknown error"}
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="AI platform"
        description={`Models · training · inference · GPU · updated ${new Date(data.generatedAt).toLocaleTimeString("en-US")}${isFetching ? " · refreshing…" : ""}`}
      />

      <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard
          title="Serving"
          value={String(data.kpis.modelsServing)}
          tone="brand"
        />
        <KpiCard
          title="Training jobs"
          value={String(data.kpis.trainingJobs)}
          tone="neutral"
        />
        <KpiCard
          title="Inference QPS"
          value={data.kpis.inferenceQps.toLocaleString("en-US")}
          tone="brand"
        />
        <KpiCard
          title="GPU util"
          value={`${data.kpis.gpuUtilPct}%`}
          tone={data.kpis.gpuUtilPct > 85 ? "warning" : "success"}
        />
        <KpiCard
          title="Tokens / min"
          value={`${(data.kpis.tokensPerMin / 1_000_000).toFixed(1)}M`}
          tone="neutral"
        />
        <KpiCard
          title="Failed 24h"
          value={String(data.kpis.failedJobs24h)}
          tone={data.kpis.failedJobs24h > 0 ? "danger" : "success"}
        />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="Inference QPS" description="Requests / s">
          <ReactECharts
            option={lineChartOption(data.inferenceSeries, "#0B6E6E")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="GPU utilization" description="%">
          <ReactECharts
            option={lineChartOption(data.gpuUtilSeries, "#085858")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <Tabs defaultValue="models">
        <TabsList>
          <TabsTrigger value="models">Models</TabsTrigger>
          <TabsTrigger value="training">Training</TabsTrigger>
          <TabsTrigger value="inference">Inference</TabsTrigger>
          <TabsTrigger value="gpu">GPU</TabsTrigger>
        </TabsList>
        <TabsContent value="models">
          <DataGrid
            columns={modelCols}
            data={data.models}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="training">
          <DataGrid
            columns={trainCols}
            data={data.trainingJobs}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="inference">
          <DataGrid
            columns={inferCols}
            data={data.inferenceServers}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="gpu">
          <DataGrid
            columns={gpuCols}
            data={data.gpuNodes}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
