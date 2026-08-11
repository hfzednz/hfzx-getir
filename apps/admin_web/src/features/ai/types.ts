export interface AiKpi {
  id: string;
  label: string;
  value: number;
  unit: "count" | "percent" | "currency";
  currency?: string;
  tone: "neutral" | "brand" | "success" | "warning" | "danger";
}

export interface SeriesPoint {
  label: string;
  value: number;
}

export interface FraudAlert {
  id: string;
  orderId: string;
  score: number;
  reason: string;
  status: "open" | "reviewing" | "cleared" | "blocked";
  createdAt: string;
}

export interface PricingRec {
  id: string;
  zone: string;
  skuCategory: string;
  currentMultiplier: number;
  suggestedMultiplier: number;
  liftPct: number;
  confidence: number;
}

export interface CampaignOpt {
  id: string;
  campaign: string;
  recommendation: string;
  expectedRoiLift: number;
  status: "pending" | "applied" | "dismissed";
}

export interface SegmentRow {
  id: string;
  name: string;
  size: number;
  avgAovMinor: number;
  churnRisk: number;
}

export interface RiskItem {
  id: string;
  area: string;
  severity: "low" | "medium" | "high";
  summary: string;
}

export interface OpsInsight {
  id: string;
  title: string;
  detail: string;
  confidence: number;
}

export interface AiCommandSnapshot {
  cityId: string | null;
  generatedAt: string;
  kpis: AiKpi[];
  demandForecast: SeriesPoint[];
  inventoryForecast: SeriesPoint[];
  fraudAlerts: FraudAlert[];
  deliveryOptSeries: SeriesPoint[];
  recommendationCtr: SeriesPoint[];
  pricingRecs: PricingRec[];
  campaignOpts: CampaignOpt[];
  segments: SegmentRow[];
  risks: RiskItem[];
  opsInsights: OpsInsight[];
}
