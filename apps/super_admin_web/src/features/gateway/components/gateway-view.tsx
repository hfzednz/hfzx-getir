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
import { barChartOption, lineChartOption } from "@/shared/lib/charts";
import { ModuleError, ModuleLoading, healthTone } from "@/shared/ui/module-state";
import { useGatewaySnapshot } from "../hooks";
import type {
  ApiKeyRecord,
  ApiVersion,
  GatewayRoute,
  OAuthClient,
  RateLimitPolicy,
  ServiceDiscoveryEntry,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function StatusCell({ value }: { value: unknown }) {
  const s = String(value);
  return <StatusBadge status={s} tone={healthTone(s)} />;
}

const routeCols: DataGridColumnDef<GatewayRoute>[] = [
  { id: "path", header: "Path", accessorKey: "path" },
  { id: "methods", header: "Methods", accessorKey: "methods" },
  { id: "upstream", header: "Upstream", accessorKey: "upstream" },
  { id: "version", header: "Version", accessorKey: "version" },
  { id: "auth", header: "Auth", accessorKey: "auth" },
  {
    id: "rl",
    header: "RPM",
    accessorKey: "rateLimitRpm",
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const keyCols: DataGridColumnDef<ApiKeyRecord>[] = [
  { id: "name", header: "Name", accessorKey: "name" },
  { id: "prefix", header: "Key", accessorKey: "prefix" },
  {
    id: "tenant",
    header: "Tenant",
    cell: ({ row }) => row.tenantId ?? "platform",
  },
  { id: "scopes", header: "Scopes", accessorKey: "scopes" },
  {
    id: "rpm",
    header: "RPM",
    accessorKey: "rateLimitRpm",
    align: "right",
  },
  {
    id: "last",
    header: "Last used",
    cell: ({ row }) =>
      row.lastUsedAt
        ? new Date(row.lastUsedAt).toLocaleString("en-US")
        : "never",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const oauthCols: DataGridColumnDef<OAuthClient>[] = [
  { id: "clientId", header: "Client ID", accessorKey: "clientId" },
  { id: "name", header: "Name", accessorKey: "name" },
  { id: "grants", header: "Grants", accessorKey: "grantTypes" },
  { id: "scopes", header: "Scopes", accessorKey: "scopes" },
  { id: "redirect", header: "Redirects", accessorKey: "redirectUris" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const rlCols: DataGridColumnDef<RateLimitPolicy>[] = [
  { id: "name", header: "Policy", accessorKey: "name" },
  { id: "scope", header: "Scope", accessorKey: "scope" },
  { id: "target", header: "Target", accessorKey: "target" },
  { id: "limit", header: "RPM", accessorKey: "limitRpm", align: "right" },
  { id: "burst", header: "Burst", accessorKey: "burst", align: "right" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const verCols: DataGridColumnDef<ApiVersion>[] = [
  { id: "version", header: "Version", accessorKey: "version" },
  {
    id: "traffic",
    header: "Traffic %",
    accessorKey: "trafficPct",
    align: "right",
  },
  { id: "routes", header: "Routes", accessorKey: "routes", align: "right" },
  {
    id: "sunset",
    header: "Sunset",
    cell: ({ row }) =>
      row.sunsetAt
        ? new Date(row.sunsetAt).toLocaleDateString("en-US")
        : "—",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const discCols: DataGridColumnDef<ServiceDiscoveryEntry>[] = [
  { id: "service", header: "Service", accessorKey: "service" },
  { id: "registry", header: "Registry", accessorKey: "registry" },
  { id: "endpoint", header: "Endpoint", accessorKey: "endpoint" },
  {
    id: "inst",
    header: "Healthy",
    cell: ({ row }) => `${row.healthy}/${row.instances}`,
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

export function GatewayView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useGatewaySnapshot();

  if (isLoading) return <ModuleLoading />;
  if (isError || !data) {
    return (
      <ModuleError
        title="Failed to load API gateway"
        message={error instanceof Error ? error.message : "Unknown error"}
        onRetry={() => void refetch()}
      />
    );
  }

  const cfg = data.config;

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="API gateway"
        description={`Edge config · keys · OAuth · discovery · updated ${new Date(data.generatedAt).toLocaleTimeString("en-US")}${isFetching ? " · refreshing…" : ""}`}
      />

      <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard
          title="RPS"
          value={data.kpis.rps.toLocaleString("en-US")}
          tone="brand"
        />
        <KpiCard title="p99" value={`${data.kpis.p99Ms} ms`} tone="neutral" />
        <KpiCard
          title="Error rate"
          value={`${data.kpis.errorRatePct}%`}
          tone={data.kpis.errorRatePct > 1 ? "danger" : "success"}
        />
        <KpiCard
          title="API keys"
          value={String(data.kpis.activeKeys)}
          tone="neutral"
        />
        <KpiCard
          title="OAuth clients"
          value={String(data.kpis.oauthClients)}
          tone="neutral"
        />
        <KpiCard
          title="Discovered"
          value={String(data.kpis.discoveredServices)}
          tone="brand"
        />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-3 gap-[var(--nx-space-3)]">
        <ChartFrame title="Traffic" description="Requests / window">
          <ReactECharts
            option={lineChartOption(data.trafficSeries, "#0B6E6E")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Error rate" description="%">
          <ReactECharts
            option={lineChartOption(data.errorSeries, "#c45c26")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Usage by route" description="% of traffic">
          <ReactECharts
            option={barChartOption(data.usageByRoute, "#085858")}
            style={{ height: 180 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <Tabs defaultValue="config">
        <TabsList>
          <TabsTrigger value="config">Config</TabsTrigger>
          <TabsTrigger value="routes">Routes</TabsTrigger>
          <TabsTrigger value="keys">API keys</TabsTrigger>
          <TabsTrigger value="oauth">OAuth</TabsTrigger>
          <TabsTrigger value="limits">Rate limits</TabsTrigger>
          <TabsTrigger value="versions">Versions</TabsTrigger>
          <TabsTrigger value="discovery">Discovery</TabsTrigger>
        </TabsList>

        <TabsContent value="config">
          <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
            <dl className="m-0 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-[var(--nx-space-3)]">
              {(
                [
                  ["Edge region", cfg.edgeRegion],
                  ["TLS min", cfg.tlsMinVersion],
                  ["JWT issuer", cfg.jwtIssuer],
                  ["Request timeout", `${cfg.requestTimeoutMs} ms`],
                  ["Body limit", `${cfg.bodyLimitMb} MB`],
                  ["CORS", cfg.corsMode],
                  ["WAF", cfg.wafEnabled ? "enabled" : "disabled"],
                ] as const
              ).map(([label, value]) => (
                <div key={label}>
                  <dt className="text-[11px] uppercase tracking-wide text-[var(--nx-text-tertiary)]">
                    {label}
                  </dt>
                  <dd className="m-0 mt-[var(--nx-space-1)] text-[13px] font-medium break-all">
                    {value}
                  </dd>
                </div>
              ))}
            </dl>
          </div>
        </TabsContent>
        <TabsContent value="routes">
          <DataGrid
            columns={routeCols}
            data={data.routes}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="keys">
          <DataGrid
            columns={keyCols}
            data={data.apiKeys}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="oauth">
          <DataGrid
            columns={oauthCols}
            data={data.oauthClients}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="limits">
          <DataGrid
            columns={rlCols}
            data={data.rateLimits}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="versions">
          <DataGrid
            columns={verCols}
            data={data.versions}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="discovery">
          <DataGrid
            columns={discCols}
            data={data.discovery}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
