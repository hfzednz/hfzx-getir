import type { Id } from "@/shared/types/common";

export type BillingCategory =
  | "tenant"
  | "api"
  | "storage"
  | "compute"
  | "courier"
  | "warehouse";

export type InvoiceStatus = "draft" | "open" | "paid" | "overdue" | "void";

export interface BillingMeter {
  id: Id;
  category: BillingCategory;
  label: string;
  usage: number;
  unit: string;
  amountMinor: number;
  currency: string;
  deltaPct: number;
}

export interface InvoiceLine {
  id: Id;
  category: BillingCategory;
  description: string;
  quantity: number;
  unitPriceMinor: number;
  amountMinor: number;
}

export interface PlatformInvoice {
  id: Id;
  tenantId: Id;
  tenantName: string;
  periodStart: string;
  periodEnd: string;
  status: InvoiceStatus;
  totalMinor: number;
  currency: string;
  dueAt: string;
  lines: InvoiceLine[];
}

export interface FinOpsSeriesPoint {
  label: string;
  value: number;
}

export interface FinOpsBreakdown {
  category: BillingCategory;
  amountMinor: number;
  pct: number;
}

export interface BillingSnapshot {
  meters: BillingMeter[];
  invoices: PlatformInvoice[];
  spendSeries: FinOpsSeriesPoint[];
  forecastSeries: FinOpsSeriesPoint[];
  breakdown: FinOpsBreakdown[];
  mtdSpendMinor: number;
  forecastMinor: number;
  currency: string;
  generatedAt: string;
}
