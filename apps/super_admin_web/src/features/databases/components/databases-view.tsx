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
import { useDatabasesSnapshot } from "../hooks";
import type {
  DbBackup,
  DbCluster,
  FailoverEvent,
  MigrationStatus,
  ReplicationLink,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function StatusCell({ value }: { value: unknown }) {
  const s = String(value);
  return <StatusBadge status={s} tone={healthTone(s)} />;
}

const clusterCols: DataGridColumnDef<DbCluster>[] = [
  { id: "name", header: "Cluster", accessorKey: "name" },
  { id: "engine", header: "Engine", accessorKey: "engine" },
  { id: "version", header: "Version", accessorKey: "version" },
  { id: "region", header: "Region", accessorKey: "region" },
  { id: "role", header: "Role", accessorKey: "role" },
  {
    id: "conn",
    header: "Connections",
    cell: ({ row }) => `${row.connections}/${row.maxConnections}`,
    align: "right",
  },
  { id: "qps", header: "QPS", accessorKey: "qps", align: "right" },
  {
    id: "lag",
    header: "Lag ms",
    cell: ({ row }) => (row.lagMs == null ? "—" : String(row.lagMs)),
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const backupCols: DataGridColumnDef<DbBackup>[] = [
  { id: "cluster", header: "Cluster", accessorKey: "clusterName" },
  { id: "engine", header: "Engine", accessorKey: "engine" },
  { id: "type", header: "Type", accessorKey: "type" },
  { id: "size", header: "Size GB", accessorKey: "sizeGb", align: "right" },
  {
    id: "started",
    header: "Started",
    cell: ({ row }) => new Date(row.startedAt).toLocaleString("en-US"),
  },
  {
    id: "retention",
    header: "Retention d",
    accessorKey: "retentionDays",
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const repCols: DataGridColumnDef<ReplicationLink>[] = [
  { id: "source", header: "Source", accessorKey: "source" },
  { id: "target", header: "Target", accessorKey: "target" },
  { id: "engine", header: "Engine", accessorKey: "engine" },
  { id: "mode", header: "Mode", accessorKey: "mode" },
  { id: "lag", header: "Lag ms", accessorKey: "lagMs", align: "right" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const foCols: DataGridColumnDef<FailoverEvent>[] = [
  { id: "cluster", header: "Cluster", accessorKey: "clusterName" },
  { id: "engine", header: "Engine", accessorKey: "engine" },
  {
    id: "path",
    header: "Failover",
    cell: ({ row }) => `${row.fromNode} → ${row.toNode}`,
  },
  { id: "reason", header: "Reason", accessorKey: "reason" },
  {
    id: "dc",
    header: "Dual-control",
    cell: ({ row }) => (row.dualControl ? "required" : "n/a"),
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const migCols: DataGridColumnDef<MigrationStatus>[] = [
  { id: "name", header: "Migration", accessorKey: "name" },
  { id: "engine", header: "Engine", accessorKey: "engine" },
  {
    id: "ver",
    header: "Versions",
    cell: ({ row }) => `${row.fromVersion} → ${row.toVersion}`,
  },
  {
    id: "prog",
    header: "Progress",
    cell: ({ row }) => `${row.progressPct}%`,
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

export function DatabasesView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useDatabasesSnapshot();

  if (isLoading) return <ModuleLoading />;
  if (isError || !data) {
    return (
      <ModuleError
        title="Failed to load databases"
        message={error instanceof Error ? error.message : "Unknown error"}
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Databases"
        description={`Postgres · Redis · OpenSearch · ClickHouse · updated ${new Date(data.generatedAt).toLocaleTimeString("en-US")}${isFetching ? " · refreshing…" : ""}`}
      />

      <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard title="Clusters" value={String(data.kpis.clusters)} tone="brand" />
        <KpiCard
          title="Unhealthy"
          value={String(data.kpis.unhealthy)}
          tone={data.kpis.unhealthy > 0 ? "danger" : "success"}
        />
        <KpiCard
          title="Backups 24h"
          value={String(data.kpis.backups24h)}
          tone="neutral"
        />
        <KpiCard
          title="Avg lag"
          value={`${data.kpis.avgLagMs} ms`}
          tone={data.kpis.avgLagMs > 100 ? "warning" : "success"}
        />
        <KpiCard
          title="Migrations"
          value={String(data.kpis.migrationsActive)}
          tone="brand"
        />
        <KpiCard
          title="Storage"
          value={`${data.kpis.storagePct}%`}
          tone={data.kpis.storagePct > 80 ? "warning" : "neutral"}
        />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="Query throughput" description="QPS">
          <ReactECharts
            option={lineChartOption(data.qpsSeries, "#0B6E6E")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Replication lag" description="ms">
          <ReactECharts
            option={lineChartOption(data.lagSeries, "#c45c26")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <Tabs defaultValue="clusters">
        <TabsList>
          <TabsTrigger value="clusters">Clusters</TabsTrigger>
          <TabsTrigger value="backups">Backups</TabsTrigger>
          <TabsTrigger value="replication">Replication</TabsTrigger>
          <TabsTrigger value="failover">Failover</TabsTrigger>
          <TabsTrigger value="migrations">Migrations</TabsTrigger>
        </TabsList>
        <TabsContent value="clusters">
          <DataGrid
            columns={clusterCols}
            data={data.clusters}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="backups">
          <DataGrid
            columns={backupCols}
            data={data.backups}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="replication">
          <DataGrid
            columns={repCols}
            data={data.replication}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="failover">
          <DataGrid
            columns={foCols}
            data={data.failovers}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="migrations">
          <DataGrid
            columns={migCols}
            data={data.migrations}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
