import type { SeriesPoint } from "@/shared/lib/charts";

export type AnalyticsScope = "worldwide" | "country" | "company";

export interface AnalyticsKpi {
  id: string;
  label: string;
  value: number;
  unit: "currency" | "percent" | "count";
  currency?: string;
  deltaPct: number;
  tone: "neutral" | "brand" | "success" | "warning" | "danger";
}

export interface CountryAggregate {
  id: string;
  country: string;
  code: string;
  revenueUsd: number;
  orders24h: number;
  activeCompanies: number;
  gmvGrowthPct: number;
}

export interface CompanyAggregate {
  id: string;
  company: string;
  country: string;
  revenueUsd: number;
  orders24h: number;
  warehouses: number;
  healthScore: number;
}

export interface AnalyticsSnapshot {
  generatedAt: string;
  scope: AnalyticsScope;
  scopeId: string | null;
  kpis: AnalyticsKpi[];
  revenueSeries: SeriesPoint[];
  ordersSeries: SeriesPoint[];
  byCountry: SeriesPoint[];
  byCompany: SeriesPoint[];
  countries: CountryAggregate[];
  companies: CompanyAggregate[];
}
