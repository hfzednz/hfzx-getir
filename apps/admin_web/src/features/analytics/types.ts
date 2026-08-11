export interface AnalyticsKpi {
  id: string;
  label: string;
  value: number;
  unit: "currency" | "percent" | "count" | "days";
  currency?: string;
  deltaPct: number;
  tone: "neutral" | "brand" | "success" | "warning" | "danger";
}

export interface SeriesPoint {
  label: string;
  value: number;
}

export interface FunnelStep {
  id: string;
  label: string;
  count: number;
  conversionPct: number;
}

export interface CohortCell {
  cohort: string;
  week: number;
  retentionPct: number;
}

export interface WarehousePerfRow {
  id: string;
  name: string;
  orders: number;
  pickMins: number;
  slaPct: number;
  stockouts: number;
}

export interface CourierPerfRow {
  id: string;
  name: string;
  deliveries: number;
  onTimePct: number;
  avgMins: number;
  rating: number;
}

export interface CancelReason {
  reason: string;
  count: number;
  pct: number;
}

export interface RefundBucket {
  label: string;
  amountMinor: number;
  count: number;
}

export interface AnalyticsSnapshot {
  cityId: string | null;
  generatedAt: string;
  kpis: AnalyticsKpi[];
  salesSeries: SeriesPoint[];
  revenueSeries: SeriesPoint[];
  retentionSeries: SeriesPoint[];
  funnel: FunnelStep[];
  cohorts: CohortCell[];
  clvSeries: SeriesPoint[];
  warehouses: WarehousePerfRow[];
  couriers: CourierPerfRow[];
  conversionSeries: SeriesPoint[];
  cancelReasons: CancelReason[];
  refunds: RefundBucket[];
  demandForecast: SeriesPoint[];
  demandActual: SeriesPoint[];
}
