export interface DashboardKpi {
  id: string;
  label: string;
  value: number;
  unit?: "currency" | "percent" | "count" | "minutes";
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

export interface AiInsight {
  id: string;
  title: string;
  summary: string;
  confidence: number;
}

export interface SystemHealthItem {
  id: string;
  name: string;
  status: "healthy" | "degraded" | "down";
  latencyMs: number;
}

export interface DashboardSnapshot {
  cityId: string | null;
  generatedAt: string;
  kpis: DashboardKpi[];
  revenueSeries: DashboardSeriesPoint[];
  ordersSeries: DashboardSeriesPoint[];
  slaSeries: DashboardSeriesPoint[];
  alerts: DashboardAlert[];
  aiInsights: AiInsight[];
  systemHealth: SystemHealthItem[];
}
