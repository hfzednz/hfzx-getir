import type { Id } from "@/shared/types/common";

export type CampaignStatus =
  | "draft"
  | "scheduled"
  | "active"
  | "paused"
  | "ended";

export type CampaignType =
  | "coupon"
  | "bundle"
  | "flash_sale"
  | "personalized"
  | "audience";

export interface CampaignAudience {
  id: Id;
  name: string;
  segmentSize: number;
  rulesSummary: string;
}

export interface CampaignCoupon {
  code: string;
  discountType: "percent" | "fixed";
  discountValue: number;
  maxRedemptions: number;
  redemptions: number;
  minOrderMinor: number;
}

export interface CampaignBundle {
  skuIds: string[];
  label: string;
  bundlePriceMinor: number;
  savingsPct: number;
}

export interface FlashSaleConfig {
  startsAt: string;
  endsAt: string;
  stockCap: number;
  sold: number;
  discountPct: number;
}

export interface PersonalizedPromo {
  model: string;
  upliftPct: number;
  channels: ("push" | "email" | "in_app")[];
}

export interface Campaign {
  id: Id;
  name: string;
  type: CampaignType;
  status: CampaignStatus;
  cityIds: Id[];
  startsAt: string | null;
  endsAt: string | null;
  budgetMinor: number;
  spentMinor: number;
  currency: string;
  audience: CampaignAudience | null;
  coupon: CampaignCoupon | null;
  bundle: CampaignBundle | null;
  flashSale: FlashSaleConfig | null;
  personalized: PersonalizedPromo | null;
  createdAt: string;
  updatedAt: string;
}

export interface CampaignListParams {
  status?: CampaignStatus | "all";
  type?: CampaignType | "all";
  q?: string;
  cityId?: string | null;
}

export interface CampaignUpsertInput {
  name: string;
  type: CampaignType;
  cityIds: Id[];
  startsAt?: string | null;
  endsAt?: string | null;
  budgetMinor?: number;
  audienceId?: Id | null;
  coupon?: CampaignCoupon | null;
  bundle?: CampaignBundle | null;
  flashSale?: FlashSaleConfig | null;
  personalized?: PersonalizedPromo | null;
}
