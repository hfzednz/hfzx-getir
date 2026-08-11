import type { Paginated } from "@/shared/types/common";
import { ApiError } from "@/shared/api/client";
import type {
  PricingKind,
  PricingListParams,
  PricingRule,
  PricingRuleStatus,
  PricingUpsertInput,
} from "./types";

const delay = (ms = 180) => new Promise((r) => setTimeout(r, ms));

let mockRules: PricingRule[] = [
  {
    id: "prc_base_01",
    name: "SKU milk 1L base",
    kind: "base",
    status: "active",
    skuId: "sku_milk_1l",
    categoryId: null,
    cityId: null,
    warehouseId: null,
    basePriceMinor: 4290,
    overridePriceMinor: null,
    adjustmentPct: null,
    currency: "TRY",
    priority: 100,
    startsAt: null,
    endsAt: null,
    competitorRef: null,
    aiConfidence: null,
    notes: "National list price",
    updatedAt: "2026-08-01T10:00:00Z",
  },
  {
    id: "prc_reg_02",
    name: "Istanbul regional +4%",
    kind: "regional",
    status: "active",
    skuId: null,
    categoryId: "cat_dairy",
    cityId: "city_ist",
    warehouseId: null,
    basePriceMinor: 0,
    overridePriceMinor: null,
    adjustmentPct: 4,
    currency: "TRY",
    priority: 80,
    startsAt: "2026-07-01T00:00:00Z",
    endsAt: null,
    competitorRef: null,
    aiConfidence: null,
    notes: "City cost index",
    updatedAt: "2026-07-15T09:00:00Z",
  },
  {
    id: "prc_wh_03",
    name: "WH-14 warehouse clear",
    kind: "warehouse",
    status: "active",
    skuId: "sku_yogurt",
    categoryId: null,
    cityId: "city_ist",
    warehouseId: "wh_14",
    basePriceMinor: 1890,
    overridePriceMinor: 1490,
    adjustmentPct: null,
    currency: "TRY",
    priority: 60,
    startsAt: "2026-08-05T00:00:00Z",
    endsAt: "2026-08-12T00:00:00Z",
    competitorRef: null,
    aiConfidence: null,
    notes: "Near-expiry clearance",
    updatedAt: "2026-08-05T08:00:00Z",
  },
  {
    id: "prc_dyn_04",
    name: "Dinner surge soft",
    kind: "dynamic",
    status: "scheduled",
    skuId: null,
    categoryId: null,
    cityId: "city_ist",
    warehouseId: null,
    basePriceMinor: 0,
    overridePriceMinor: null,
    adjustmentPct: 6,
    currency: "TRY",
    priority: 40,
    startsAt: "2026-08-06T17:30:00Z",
    endsAt: "2026-08-06T21:00:00Z",
    competitorRef: null,
    aiConfidence: null,
    notes: "SLA protection window",
    updatedAt: "2026-08-06T07:00:00Z",
  },
  {
    id: "prc_comp_05",
    name: "Match competitor water pack",
    kind: "competitor",
    status: "active",
    skuId: "sku_water_6pk",
    categoryId: null,
    cityId: "city_ist",
    warehouseId: null,
    basePriceMinor: 9990,
    overridePriceMinor: 9490,
    adjustmentPct: null,
    currency: "TRY",
    priority: 50,
    startsAt: null,
    endsAt: null,
    competitorRef: "comp_local_qc",
    aiConfidence: null,
    notes: "Parity rule",
    updatedAt: "2026-08-03T12:00:00Z",
  },
  {
    id: "prc_sch_06",
    name: "Weekend beverage promo price",
    kind: "scheduled",
    status: "scheduled",
    skuId: null,
    categoryId: "cat_beverage",
    cityId: "city_ank",
    warehouseId: null,
    basePriceMinor: 0,
    overridePriceMinor: null,
    adjustmentPct: -8,
    currency: "TRY",
    priority: 70,
    startsAt: "2026-08-08T00:00:00Z",
    endsAt: "2026-08-10T23:59:00Z",
    competitorRef: null,
    aiConfidence: null,
    notes: "",
    updatedAt: "2026-08-04T16:00:00Z",
  },
  {
    id: "prc_emg_07",
    name: "Emergency ice cream floor",
    kind: "emergency",
    status: "paused",
    skuId: null,
    categoryId: "cat_ice",
    cityId: "city_ist",
    warehouseId: null,
    basePriceMinor: 0,
    overridePriceMinor: null,
    adjustmentPct: 12,
    currency: "TRY",
    priority: 10,
    startsAt: null,
    endsAt: null,
    competitorRef: null,
    aiConfidence: null,
    notes: "Heatwave stockout guard",
    updatedAt: "2026-08-02T18:00:00Z",
  },
  {
    id: "prc_ai_08",
    name: "AI snack elasticity",
    kind: "ai_assisted",
    status: "draft",
    skuId: null,
    categoryId: "cat_snack",
    cityId: "city_ist",
    warehouseId: null,
    basePriceMinor: 0,
    overridePriceMinor: null,
    adjustmentPct: -3.5,
    currency: "TRY",
    priority: 90,
    startsAt: null,
    endsAt: null,
    competitorRef: null,
    aiConfidence: 0.82,
    notes: "Awaiting catalog_manager publish",
    updatedAt: "2026-08-06T06:30:00Z",
  },
];

/** Mock pricing — replaced by GET /admin/pricing when BFF is live. */
export async function fetchPricingRules(
  params: PricingListParams = {},
): Promise<Paginated<PricingRule>> {
  await delay();
  let items = [...mockRules];
  if (params.kind && params.kind !== "all") {
    items = items.filter((r) => r.kind === params.kind);
  }
  if (params.status && params.status !== "all") {
    items = items.filter((r) => r.status === params.status);
  }
  if (params.cityId) {
    items = items.filter(
      (r) => r.cityId === params.cityId || r.cityId === null,
    );
  }
  if (params.q?.trim()) {
    const q = params.q.trim().toLowerCase();
    items = items.filter(
      (r) =>
        r.name.toLowerCase().includes(q) ||
        r.id.toLowerCase().includes(q) ||
        (r.skuId?.toLowerCase().includes(q) ?? false),
    );
  }
  return { items, page: 1, pageSize: 100, total: items.length, hasMore: false };
}

export async function createPricingRule(
  input: PricingUpsertInput,
): Promise<PricingRule> {
  await delay(240);
  const created: PricingRule = {
    id: `prc_${Date.now().toString(36)}`,
    name: input.name,
    kind: input.kind,
    status: "draft",
    skuId: input.skuId ?? null,
    categoryId: input.categoryId ?? null,
    cityId: input.cityId ?? null,
    warehouseId: input.warehouseId ?? null,
    basePriceMinor: input.basePriceMinor,
    overridePriceMinor: input.overridePriceMinor ?? null,
    adjustmentPct: input.adjustmentPct ?? null,
    currency: "TRY",
    priority: input.priority ?? 100,
    startsAt: input.startsAt ?? null,
    endsAt: input.endsAt ?? null,
    competitorRef: input.competitorRef ?? null,
    aiConfidence: input.kind === "ai_assisted" ? 0.7 : null,
    notes: input.notes ?? "",
    updatedAt: new Date().toISOString(),
  };
  mockRules = [created, ...mockRules];
  return created;
}

export async function updatePricingRuleStatus(
  id: string,
  status: PricingRuleStatus,
): Promise<PricingRule> {
  await delay(180);
  const idx = mockRules.findIndex((r) => r.id === id);
  if (idx < 0) {
    throw new ApiError(404, {
      code: "not_found",
      message: "Pricing rule not found",
      traceId: "mock",
    });
  }
  const next = {
    ...mockRules[idx]!,
    status,
    updatedAt: new Date().toISOString(),
  };
  mockRules = mockRules.map((r, i) => (i === idx ? next : r));
  return next;
}

export const PRICING_KINDS: PricingKind[] = [
  "base",
  "regional",
  "warehouse",
  "dynamic",
  "competitor",
  "scheduled",
  "emergency",
  "ai_assisted",
];
