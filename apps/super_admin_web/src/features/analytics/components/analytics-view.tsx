"use client";

import dynamic from "next/dynamic";
import {
  ChartFrame,
  DataGrid,
  type DataGridColumnDef,
  KpiCard,
  PageHeader,
  Select,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@nexora/ui";
import { barChartOption, lineChartOption } from "@/shared/lib/charts";
import { formatMinorUnits } from "@/shared/lib/money";
import { ModuleError, ModuleLoading } from "@/shared/ui/module-state";
import { useAnalyticsSnapshot } from "../hooks";
import type {
  AnalyticsKpi,
  AnalyticsScope,
  CompanyAggregate,
  CountryAggregate,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function formatKpi(kpi: AnalyticsKpi): string {
  switch (kpi.unit) {
    case "currency":
      return formatMinorUnits(kpi.value, kpi.currency ?? "USD");
    case "percent":
      return `${kpi.value.toFixed(1)}%`;
    default:
      return kpi.value.toLocaleString("en-US");
  }
}

const countryCols: DataGridColumnDef<CountryAggregate>[] = [
  { id: "code", header: "Code", accessorKey: "code" },
  { id: "country", header: "Country", accessorKey: "country" },
  {
    id: "rev",
    header: "GMV",
    cell: ({ row }) => formatMinorUnits(row.revenueUsd, "USD"),
    align: "right",
  },
  {
    id: "orders",
    header: "Orders 24h",
    accessorKey: "orders24h",
    align: "right",
  },
  {
    id: "cos",
    header: "Companies",
    accessorKey: "activeCompanies",
    align: "right",
  },
  {
    id: "growth",
    header: "Growth %",
    accessorKey: "gmvGrowthPct",
    align: "right",
  },
];

const companyCols: DataGridColumnDef<CompanyAggregate>[] = [
  { id: "company", header: "Company", accessorKey: "company" },
  { id: "country", header: "Country", accessorKey: "country" },
  {
    id: "rev",
    header: "GMV",
    cell: ({ row }) => formatMinorUnits(row.revenueUsd, "USD"),
    align: "right",
  },
  {
    id: "orders",
    header: "Orders 24h",
    accessorKey: "orders24h",
    align: "right",
  },
  {
    id: "wh",
    header: "Warehouses",
    accessorKey: "warehouses",
    align: "right",
  },
  {
    id: "health",
    header: "Health",
    accessorKey: "healthScore",
    align: "right",
  },
];

export function AnalyticsView() {
  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
    isFetching,
    scope,
    setScope,
    setScopeId,
  } = useAnalyticsSnapshot();

  if (isLoading) return <ModuleLoading />;
  if (isError || !data) {
    return (
      <ModuleError
        title="Failed to load analytics"
        message={error instanceof Error ? error.message : "Unknown error"}
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Platform analytics"
        description={`Worldwide / country / company aggregates only · updated ${new Date(data.generatedAt).toLocaleTimeString("en-US")}${isFetching ? " · refreshing…" : ""}`}
      />

      <div className="flex flex-wrap items-end gap-[var(--nx-space-3)]">
        <label className="flex flex-col gap-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
          Scope
          <Select
            value={scope}
            onChange={(e) => {
              const next = e.target.value as AnalyticsScope;
              setScope(next);
              if (next === "worldwide") setScopeId(null);
              if (next === "country") setScopeId("ct-tr");
              if (next === "company") setScopeId("co1");
            }}
          >
            <option value="worldwide">Worldwide</option>
            <option value="country">Country</option>
            <option value="company">Company</option>
          </Select>
        </label>
        {scope === "country" && (
          <label className="flex flex-col gap-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
            Country
            <Select
              defaultValue="ct-tr"
              onChange={(e) => setScopeId(e.target.value)}
            >
              {data.countries.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.country}
                </option>
              ))}
            </Select>
          </label>
        )}
        {scope === "company" && (
          <label className="flex flex-col gap-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
            Company
            <Select
              defaultValue="co1"
              onChange={(e) => setScopeId(e.target.value)}
            >
              {data.companies.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.company}
                </option>
              ))}
            </Select>
          </label>
        )}
      </div>

      <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        {data.kpis.map((kpi) => (
          <KpiCard
            key={kpi.id}
            title={kpi.label}
            value={formatKpi(kpi)}
            delta={`${kpi.deltaPct > 0 ? "+" : ""}${kpi.deltaPct.toFixed(1)}%`}
            tone={kpi.tone}
          />
        ))}
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <ChartFrame title="GMV trend" description="Scoped aggregate (USD)">
          <ReactECharts
            option={lineChartOption(
              data.revenueSeries.map((p) => ({
                ...p,
                value: Math.round(p.value / 100),
              })),
              "#0B6E6E",
            )}
            style={{ height: 200 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="Orders trend" description="24h windows">
          <ReactECharts
            option={lineChartOption(data.ordersSeries, "#0f8585")}
            style={{ height: 200 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="GMV by country" description="% share">
          <ReactECharts
            option={barChartOption(data.byCountry, "#085858")}
            style={{ height: 200 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
        <ChartFrame title="GMV by company" description="% share">
          <ReactECharts
            option={barChartOption(data.byCompany, "#0B6E6E")}
            style={{ height: 200 }}
            opts={{ renderer: "svg" }}
          />
        </ChartFrame>
      </section>

      <Tabs defaultValue="countries">
        <TabsList>
          <TabsTrigger value="countries">Countries</TabsTrigger>
          <TabsTrigger value="companies">Companies</TabsTrigger>
        </TabsList>
        <TabsContent value="countries">
          <DataGrid
            columns={countryCols}
            data={data.countries}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="companies">
          <DataGrid
            columns={companyCols}
            data={data.companies}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
