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
import { useObservabilitySnapshot } from "../hooks";
import type {
  HealthProbe,
  ObsAlert,
  ObsIncident,
  ObsLink,
  SlaContract,
  SloPanel,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function StatusCell({ value }: { value: unknown }) {
  const s = String(value);
  return <StatusBadge status={s} tone={healthTone(s)} />;
}

const linkCols: DataGridColumnDef<ObsLink>[] = [
  { id: "kind", header: "Kind", accessorKey: "kind" },
  { id: "name", header: "Name", accessorKey: "name" },
  { id: "provider", header: "Provider", accessorKey: "provider" },
  { id: "desc", header: "Description", accessorKey: "description" },
  {
    id: "url",
    header: "Open",
    cell: ({ row }) => (
      <a
        href={row.url}
        target="_blank"
        rel="noreferrer"
        className="text-[var(--nx-text-link)] underline"
      >
        Link
      </a>
    ),
  },
];

const alertCols: DataGridColumnDef<ObsAlert>[] = [
  {
    id: "sev",
    header: "Severity",
    accessorKey: "severity",
    cell: ({ value }) => <StatusCell value={value} />,
  },
  { id: "title", header: "Alert", accessorKey: "title" },
  { id: "service", header: "Service", accessorKey: "service" },
  {
    id: "fired",
    header: "Fired",
    cell: ({ row }) => new Date(row.firedAt).toLocaleString("en-US"),
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const incidentCols: DataGridColumnDef<ObsIncident>[] = [
  {
    id: "sev",
    header: "Severity",
    accessorKey: "severity",
    cell: ({ value }) => <StatusCell value={value} />,
  },
  { id: "title", header: "Incident", accessorKey: "title" },
  { id: "region", header: "Region", accessorKey: "region" },
  { id: "commander", header: "Commander", accessorKey: "commander" },
  {
    id: "opened",
    header: "Opened",
    cell: ({ row }) => new Date(row.openedAt).toLocaleString("en-US"),
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const healthCols: DataGridColumnDef<HealthProbe>[] = [
  { id: "name", header: "Probe", accessorKey: "name" },
  { id: "kind", header: "Kind", accessorKey: "kind" },
  { id: "region", header: "Region", accessorKey: "region" },
  {
    id: "lat",
    header: "Latency",
    cell: ({ row }) => `${row.latencyMs} ms`,
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const sloCols: DataGridColumnDef<SloPanel>[] = [
  { id: "name", header: "SLO", accessorKey: "name" },
  { id: "sli", header: "SLI", accessorKey: "sli" },
  {
    id: "target",
    header: "Target",
    cell: ({ row }) => `${row.targetPct}%`,
    align: "right",
  },
  {
    id: "current",
    header: "Current",
    cell: ({ row }) => `${row.currentPct}%`,
    align: "right",
  },
  {
    id: "budget",
    header: "Error budget",
    cell: ({ row }) => `${row.errorBudgetRemainingPct}%`,
    align: "right",
  },
  { id: "window", header: "Window", accessorKey: "window" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const slaCols: DataGridColumnDef<SlaContract>[] = [
  { id: "name", header: "SLA", accessorKey: "name" },
  { id: "scope", header: "Scope", accessorKey: "scope" },
  {
    id: "target",
    header: "Target",
    cell: ({ row }) => `${row.uptimeTargetPct}%`,
    align: "right",
  },
  {
    id: "current",
    header: "Current",
    cell: ({ row }) => `${row.currentPct}%`,
    align: "right",
  },
  {
    id: "risk",
    header: "Credits at risk",
    cell: ({ row }) => (
      <StatusBadge
        status={row.creditsAtRisk ? "at risk" : "ok"}
        tone={row.creditsAtRisk ? "danger" : "success"}
      />
    ),
  },
];

export function ObservabilityView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useObservabilitySnapshot();

  if (isLoading) return <ModuleLoading />;
  if (isError || !data) {
    return (
      <ModuleError
        title="Failed to load observability"
        message={error instanceof Error ? error.message : "Unknown error"}
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Observability"
        description={`Logs · metrics · traces · SLO/SLA · updated ${new Date(data.generatedAt).toLocaleTimeString("en-US")}${isFetching ? " · refreshing…" : ""}`}
      />

      <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard
          title="Firing alerts"
          value={String(data.kpis.firingAlerts)}
          tone={data.kpis.firingAlerts > 0 ? "danger" : "success"}
        />
        <KpiCard
          title="Open incidents"
          value={String(data.kpis.openIncidents)}
          tone={data.kpis.openIncidents > 0 ? "warning" : "success"}
        />
        <KpiCard
          title="Probes up"
          value={`${data.kpis.probesUpPct}%`}
          tone="success"
        />
        <KpiCard
          title="SLO breaches"
          value={String(data.kpis.sloBreaches)}
          tone={data.kpis.sloBreaches > 0 ? "danger" : "success"}
        />
        <KpiCard
          title="Error budget"
          value={`${data.kpis.avgErrorBudgetPct}%`}
          tone={data.kpis.avgErrorBudgetPct < 20 ? "warning" : "brand"}
        />
        <KpiCard
          title="Trace ingest"
          value={`${(data.kpis.traceIngestRps / 1000).toFixed(1)}k/s`}
          tone="neutral"
        />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="Error rate" description="%">
          <ReactECharts
            option={lineChartOption(data.errorRateSeries, "#c45c26")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Latency p99" description="ms">
          <ReactECharts
            option={lineChartOption(data.latencySeries, "#0B6E6E")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <Tabs defaultValue="links">
        <TabsList>
          <TabsTrigger value="links">Links</TabsTrigger>
          <TabsTrigger value="alerts">Alerts</TabsTrigger>
          <TabsTrigger value="incidents">Incidents</TabsTrigger>
          <TabsTrigger value="health">Health</TabsTrigger>
          <TabsTrigger value="slo">SLO / SLI</TabsTrigger>
          <TabsTrigger value="sla">SLA</TabsTrigger>
        </TabsList>
        <TabsContent value="links">
          <DataGrid columns={linkCols} data={data.links} getRowId={(r) => r.id} />
        </TabsContent>
        <TabsContent value="alerts">
          <DataGrid
            columns={alertCols}
            data={data.alerts}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="incidents">
          <DataGrid
            columns={incidentCols}
            data={data.incidents}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="health">
          <DataGrid
            columns={healthCols}
            data={data.health}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="slo">
          <DataGrid columns={sloCols} data={data.slos} getRowId={(r) => r.id} />
        </TabsContent>
        <TabsContent value="sla">
          <DataGrid columns={slaCols} data={data.slas} getRowId={(r) => r.id} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
