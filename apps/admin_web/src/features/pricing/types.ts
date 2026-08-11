import type { Id } from "@/shared/types/common";

export type PricingKind =
  | "base"
  | "regional"
  | "warehouse"
  | "dynamic"
  | "competitor"
  | "scheduled"
  | "emergency"
  | "ai_assisted";

export type PricingRuleStatus = "draft" | "active" | "scheduled" | "expired" | "paused";

export interface PricingRule {
  id: Id;
  name: string;
  kind: PricingKind;
  status: PricingRuleStatus;
  skuId: string | null;
  categoryId: string | null;
  cityId: string | null;
  warehouseId: string | null;
  basePriceMinor: number;
  overridePriceMinor: number | null;
  adjustmentPct: number | null;
  currency: string;
  priority: number;
  startsAt: string | null;
  endsAt: string | null;
  competitorRef: string | null;
  aiConfidence: number | null;
  notes: string;
  updatedAt: string;
}

export interface PricingListParams {
  kind?: PricingKind | "all";
  status?: PricingRuleStatus | "all";
  q?: string;
  cityId?: string | null;
}

export interface PricingUpsertInput {
  name: string;
  kind: PricingKind;
  skuId?: string | null;
  categoryId?: string | null;
  cityId?: string | null;
  warehouseId?: string | null;
  basePriceMinor: number;
  overridePriceMinor?: number | null;
  adjustmentPct?: number | null;
  priority?: number;
  startsAt?: string | null;
  endsAt?: string | null;
  competitorRef?: string | null;
  notes?: string;
}
