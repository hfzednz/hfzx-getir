export interface DashboardKpi {
  id: string;
  label: string;
  value: number;
  unit?: "currency" | "percent" | "count" | "usd";
  currency?: string;
  deltaPct: number;
  tone: "neutral" | "brand" | "success" | "warning" | "danger";
}

export interface DashboardSeriesPoint {
  label: string;
  value: number;
}

export interface DashboardAlert {
  id: string;
  severity: "info" | "warning" | "danger";
  title: string;
  detail: string;
  createdAt: string;
}

export interface IncidentItem {
  id: string;
  severity: "sev1" | "sev2" | "sev3";
  title: string;
  status: "investigating" | "mitigating" | "resolved";
  region: string;
}

export interface SystemHealthItem {
  id: string;
  name: string;
  status: "healthy" | "degraded" | "down";
  latencyMs: number;
}

export interface PlatformDashboardSnapshot {
  tenantContextId: string | null;
  generatedAt: string;
  kpis: DashboardKpi[];
  revenueSeries: DashboardSeriesPoint[];
  trafficSeries: DashboardSeriesPoint[];
  costSeries: DashboardSeriesPoint[];
  alerts: DashboardAlert[];
  incidents: IncidentItem[];
  systemHealth: SystemHealthItem[];
}
