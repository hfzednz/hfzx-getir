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
import { useMessagingSnapshot } from "../hooks";
import type {
  DlqEntry,
  MessagingConsumer,
  MessagingQueue,
  MessagingTopic,
  RetryPolicy,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function StatusCell({ value }: { value: unknown }) {
  const s = String(value);
  return <StatusBadge status={s} tone={healthTone(s)} />;
}

const topicCols: DataGridColumnDef<MessagingTopic>[] = [
  { id: "name", header: "Topic", accessorKey: "name" },
  { id: "broker", header: "Broker", accessorKey: "broker" },
  {
    id: "parts",
    header: "Partitions",
    accessorKey: "partitions",
    align: "right",
  },
  {
    id: "replicas",
    header: "Replicas",
    accessorKey: "replicas",
    align: "right",
  },
  {
    id: "mps",
    header: "msg/s",
    accessorKey: "messagesPerSec",
    align: "right",
  },
  { id: "lag", header: "Lag", accessorKey: "lag", align: "right" },
  {
    id: "retention",
    header: "Retention h",
    accessorKey: "retentionHours",
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const queueCols: DataGridColumnDef<MessagingQueue>[] = [
  { id: "name", header: "Queue", accessorKey: "name" },
  { id: "broker", header: "Broker", accessorKey: "broker" },
  { id: "depth", header: "Depth", accessorKey: "depth", align: "right" },
  {
    id: "consumers",
    header: "Consumers",
    accessorKey: "consumers",
    align: "right",
  },
  { id: "ack", header: "Ack/s", accessorKey: "ackRate", align: "right" },
  {
    id: "dlq",
    header: "DLQ",
    cell: ({ row }) => (row.dlqBound ? "bound" : "—"),
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const consumerCols: DataGridColumnDef<MessagingConsumer>[] = [
  { id: "group", header: "Group", accessorKey: "group" },
  { id: "source", header: "Topic / queue", accessorKey: "topicOrQueue" },
  { id: "broker", header: "Broker", accessorKey: "broker" },
  { id: "members", header: "Members", accessorKey: "members", align: "right" },
  { id: "lag", header: "Lag", accessorKey: "lag", align: "right" },
  { id: "region", header: "Region", accessorKey: "region" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const dlqCols: DataGridColumnDef<DlqEntry>[] = [
  { id: "source", header: "Source", accessorKey: "source" },
  { id: "broker", header: "Broker", accessorKey: "broker" },
  { id: "depth", header: "Depth", accessorKey: "depth", align: "right" },
  {
    id: "age",
    header: "Oldest (min)",
    accessorKey: "oldestAgeMin",
    align: "right",
  },
  { id: "err", header: "Last error", accessorKey: "lastError" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const retryCols: DataGridColumnDef<RetryPolicy>[] = [
  { id: "name", header: "Policy", accessorKey: "name" },
  { id: "applies", header: "Applies to", accessorKey: "appliesTo" },
  {
    id: "attempts",
    header: "Max attempts",
    accessorKey: "maxAttempts",
    align: "right",
  },
  { id: "backoff", header: "Backoff", accessorKey: "backoff" },
  {
    id: "initial",
    header: "Initial ms",
    accessorKey: "initialDelayMs",
    align: "right",
  },
  {
    id: "max",
    header: "Max ms",
    accessorKey: "maxDelayMs",
    align: "right",
  },
  {
    id: "dlq",
    header: "DLQ on exhaust",
    cell: ({ row }) => (row.dlqOnExhaust ? "yes" : "no"),
  },
];

export function MessagingView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useMessagingSnapshot();

  if (isLoading) return <ModuleLoading />;
  if (isError || !data) {
    return (
      <ModuleError
        title="Failed to load messaging"
        message={error instanceof Error ? error.message : "Unknown error"}
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Messaging"
        description={`Kafka · RabbitMQ · DLQ · retry · updated ${new Date(data.generatedAt).toLocaleTimeString("en-US")}${isFetching ? " · refreshing…" : ""}`}
      />

      <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard title="Topics" value={String(data.kpis.topics)} tone="brand" />
        <KpiCard title="Queues" value={String(data.kpis.queues)} tone="neutral" />
        <KpiCard
          title="Consumer groups"
          value={String(data.kpis.consumerGroups)}
          tone="neutral"
        />
        <KpiCard
          title="Total lag"
          value={data.kpis.totalLag.toLocaleString("en-US")}
          tone={data.kpis.totalLag > 100000 ? "warning" : "success"}
        />
        <KpiCard
          title="DLQ depth"
          value={data.kpis.dlqDepth.toLocaleString("en-US")}
          tone={data.kpis.dlqDepth > 500 ? "danger" : "warning"}
        />
        <KpiCard
          title="Throughput"
          value={`${(data.kpis.throughputMsgPerSec / 1000).toFixed(1)}k/s`}
          tone="brand"
        />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="Throughput" description="Messages / s">
          <ReactECharts
            option={lineChartOption(data.throughputSeries, "#0B6E6E")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Consumer lag" description="Messages behind">
          <ReactECharts
            option={lineChartOption(data.lagSeries, "#c45c26")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <Tabs defaultValue="topics">
        <TabsList>
          <TabsTrigger value="topics">Topics</TabsTrigger>
          <TabsTrigger value="queues">Queues</TabsTrigger>
          <TabsTrigger value="consumers">Consumers</TabsTrigger>
          <TabsTrigger value="dlq">DLQ</TabsTrigger>
          <TabsTrigger value="retry">Retry policies</TabsTrigger>
        </TabsList>
        <TabsContent value="topics">
          <DataGrid
            columns={topicCols}
            data={data.topics}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="queues">
          <DataGrid
            columns={queueCols}
            data={data.queues}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="consumers">
          <DataGrid
            columns={consumerCols}
            data={data.consumers}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="dlq">
          <DataGrid columns={dlqCols} data={data.dlq} getRowId={(r) => r.id} />
        </TabsContent>
        <TabsContent value="retry">
          <DataGrid
            columns={retryCols}
            data={data.retryPolicies}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
