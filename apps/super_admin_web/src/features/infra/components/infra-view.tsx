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
import { useInfraSnapshot } from "../hooks";
import type {
  CdnDistribution,
  Certificate,
  DnsRecord,
  InfraCluster,
  InfraServer,
  K8sDeployment,
  K8sIngress,
  K8sNamespace,
  K8sService,
  StorageVolume,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function StatusCell({ value }: { value: unknown }) {
  const s = String(value);
  return <StatusBadge status={s} tone={healthTone(s)} />;
}

const clusterCols: DataGridColumnDef<InfraCluster>[] = [
  { id: "name", header: "Cluster", accessorKey: "name" },
  { id: "region", header: "Region", accessorKey: "region" },
  { id: "provider", header: "Provider", accessorKey: "provider" },
  { id: "version", header: "K8s", accessorKey: "version" },
  {
    id: "nodes",
    header: "Nodes",
    accessorKey: "nodeCount",
    align: "right",
  },
  {
    id: "cpu",
    header: "CPU %",
    accessorKey: "cpuPct",
    align: "right",
  },
  {
    id: "mem",
    header: "Mem %",
    accessorKey: "memPct",
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const serverCols: DataGridColumnDef<InfraServer>[] = [
  { id: "hostname", header: "Host", accessorKey: "hostname" },
  { id: "role", header: "Role", accessorKey: "role" },
  { id: "type", header: "Instance", accessorKey: "instanceType" },
  { id: "az", header: "AZ", accessorKey: "az" },
  { id: "cpu", header: "CPU %", accessorKey: "cpuPct", align: "right" },
  { id: "mem", header: "Mem %", accessorKey: "memPct", align: "right" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const nsCols: DataGridColumnDef<K8sNamespace>[] = [
  { id: "name", header: "Namespace", accessorKey: "name" },
  { id: "cluster", header: "Cluster", accessorKey: "clusterId" },
  { id: "pods", header: "Pods", accessorKey: "pods", align: "right" },
  { id: "quotas", header: "Quotas", accessorKey: "quotas" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const deployCols: DataGridColumnDef<K8sDeployment>[] = [
  { id: "name", header: "Deployment", accessorKey: "name" },
  { id: "ns", header: "Namespace", accessorKey: "namespace" },
  {
    id: "ready",
    header: "Ready",
    cell: ({ row }) => `${row.ready}/${row.replicas}`,
    align: "right",
  },
  { id: "image", header: "Image", accessorKey: "image" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const svcCols: DataGridColumnDef<K8sService>[] = [
  { id: "name", header: "Service", accessorKey: "name" },
  { id: "ns", header: "Namespace", accessorKey: "namespace" },
  { id: "type", header: "Type", accessorKey: "type" },
  { id: "ip", header: "Cluster IP", accessorKey: "clusterIp" },
  { id: "ports", header: "Ports", accessorKey: "ports" },
];

const ingCols: DataGridColumnDef<K8sIngress>[] = [
  { id: "name", header: "Ingress", accessorKey: "name" },
  { id: "ns", header: "Namespace", accessorKey: "namespace" },
  { id: "hosts", header: "Hosts", accessorKey: "hosts" },
  {
    id: "tls",
    header: "TLS",
    cell: ({ row }) => (row.tls ? "yes" : "no"),
  },
  { id: "class", header: "Class", accessorKey: "className" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const certCols: DataGridColumnDef<Certificate>[] = [
  { id: "domain", header: "Domain", accessorKey: "domain" },
  { id: "issuer", header: "Issuer", accessorKey: "issuer" },
  {
    id: "expires",
    header: "Expires",
    cell: ({ row }) => new Date(row.expiresAt).toLocaleDateString("en-US"),
  },
  {
    id: "renew",
    header: "Auto-renew",
    cell: ({ row }) => (row.autoRenew ? "yes" : "no"),
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const dnsCols: DataGridColumnDef<DnsRecord>[] = [
  { id: "zone", header: "Zone", accessorKey: "zone" },
  { id: "name", header: "Name", accessorKey: "name" },
  { id: "type", header: "Type", accessorKey: "type" },
  { id: "value", header: "Value", accessorKey: "value" },
  { id: "ttl", header: "TTL", accessorKey: "ttl", align: "right" },
  {
    id: "proxied",
    header: "Proxied",
    cell: ({ row }) => (row.proxied ? "yes" : "no"),
  },
];

const storageCols: DataGridColumnDef<StorageVolume>[] = [
  { id: "name", header: "Volume", accessorKey: "name" },
  { id: "type", header: "Type", accessorKey: "type" },
  { id: "region", header: "Region", accessorKey: "region" },
  { id: "size", header: "Size GB", accessorKey: "sizeGb", align: "right" },
  { id: "used", header: "Used %", accessorKey: "usedPct", align: "right" },
  {
    id: "enc",
    header: "Encrypted",
    cell: ({ row }) => (row.encrypted ? "yes" : "no"),
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

const cdnCols: DataGridColumnDef<CdnDistribution>[] = [
  { id: "name", header: "Distribution", accessorKey: "name" },
  { id: "provider", header: "Provider", accessorKey: "provider" },
  { id: "domain", header: "Domain", accessorKey: "domain" },
  {
    id: "hit",
    header: "Cache hit %",
    accessorKey: "cacheHitPct",
    align: "right",
  },
  {
    id: "bw",
    header: "Gbps",
    accessorKey: "bandwidthGbps",
    align: "right",
  },
  { id: "origins", header: "Origins", accessorKey: "origins", align: "right" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => <StatusCell value={value} />,
  },
];

export function InfraView() {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useInfraSnapshot();

  if (isLoading) return <ModuleLoading />;
  if (isError || !data) {
    return (
      <ModuleError
        title="Failed to load infrastructure"
        message={error instanceof Error ? error.message : "Unknown error"}
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Infrastructure"
        description={`Clusters · K8s · DNS · CDN · updated ${new Date(data.generatedAt).toLocaleTimeString("en-US")}${isFetching ? " · refreshing…" : ""}`}
      />

      <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard title="Clusters" value={String(data.kpis.clusters)} tone="brand" />
        <KpiCard title="Nodes" value={String(data.kpis.nodes)} tone="neutral" />
        <KpiCard
          title="Deployments"
          value={String(data.kpis.deployments)}
          tone="neutral"
        />
        <KpiCard
          title="Certs expiring"
          value={String(data.kpis.certsExpiring)}
          tone={data.kpis.certsExpiring > 0 ? "warning" : "success"}
        />
        <KpiCard
          title="CDN hit %"
          value={`${data.kpis.cdnHitPct}%`}
          tone="success"
        />
        <KpiCard
          title="Storage used"
          value={`${data.kpis.storageUsedPct}%`}
          tone={data.kpis.storageUsedPct > 80 ? "warning" : "neutral"}
        />
      </section>

      <ChartFrame title="Cluster CPU" description="Average utilization %">
        <ReactECharts
          option={lineChartOption(data.cpuSeries, "#0B6E6E")}
          style={{ height: 180 }}
          opts={{ renderer: "svg" }}
        />
      </ChartFrame>

      <Tabs defaultValue="clusters">
        <TabsList>
          <TabsTrigger value="clusters">Clusters</TabsTrigger>
          <TabsTrigger value="servers">Servers</TabsTrigger>
          <TabsTrigger value="k8s">Kubernetes</TabsTrigger>
          <TabsTrigger value="certs">Certificates</TabsTrigger>
          <TabsTrigger value="dns">DNS</TabsTrigger>
          <TabsTrigger value="storage">Storage</TabsTrigger>
          <TabsTrigger value="cdn">CDN</TabsTrigger>
        </TabsList>

        <TabsContent value="clusters">
          <DataGrid
            columns={clusterCols}
            data={data.clusters}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="servers">
          <DataGrid
            columns={serverCols}
            data={data.servers}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="k8s">
          <div className="flex flex-col gap-[var(--nx-space-4)]">
            <div>
              <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
                Namespaces
              </h3>
              <DataGrid
                columns={nsCols}
                data={data.namespaces}
                getRowId={(r) => r.id}
              />
            </div>
            <div>
              <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
                Deployments
              </h3>
              <DataGrid
                columns={deployCols}
                data={data.deployments}
                getRowId={(r) => r.id}
              />
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
              <div>
                <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
                  Services
                </h3>
                <DataGrid
                  columns={svcCols}
                  data={data.services}
                  getRowId={(r) => r.id}
                />
              </div>
              <div>
                <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
                  Ingress
                </h3>
                <DataGrid
                  columns={ingCols}
                  data={data.ingresses}
                  getRowId={(r) => r.id}
                />
              </div>
            </div>
          </div>
        </TabsContent>
        <TabsContent value="certs">
          <DataGrid
            columns={certCols}
            data={data.certificates}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="dns">
          <DataGrid
            columns={dnsCols}
            data={data.dnsRecords}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="storage">
          <DataGrid
            columns={storageCols}
            data={data.storage}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="cdn">
          <DataGrid columns={cdnCols} data={data.cdn} getRowId={(r) => r.id} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
