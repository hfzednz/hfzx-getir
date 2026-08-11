import type { Paginated } from "@/shared/types/common";
import { ApiError } from "@/shared/api/client";
import type {
  Campaign,
  CampaignListParams,
  CampaignStatus,
  CampaignType,
  CampaignUpsertInput,
} from "./types";

const delay = (ms = 180) => new Promise((r) => setTimeout(r, ms));

let mockCampaigns: Campaign[] = [
  {
    id: "cmp_flash_01",
    name: "Ramadan flash dairy",
    type: "flash_sale",
    status: "active",
    cityIds: ["city_ist"],
    startsAt: "2026-08-06T10:00:00Z",
    endsAt: "2026-08-06T22:00:00Z",
    budgetMinor: 250_000_00,
    spentMinor: 87_400_00,
    currency: "TRY",
    audience: {
      id: "aud_peak",
      name: "Peak dinner buyers",
      segmentSize: 184200,
      rulesSummary: "Orders >= 3 in 30d · AOV > 200 TRY",
    },
    coupon: null,
    bundle: null,
    flashSale: {
      startsAt: "2026-08-06T10:00:00Z",
      endsAt: "2026-08-06T22:00:00Z",
      stockCap: 5000,
      sold: 2140,
      discountPct: 25,
    },
    personalized: null,
    createdAt: "2026-08-01T09:00:00Z",
    updatedAt: "2026-08-06T08:00:00Z",
  },
  {
    id: "cmp_coupon_02",
    name: "Welcome TRY50",
    type: "coupon",
    status: "scheduled",
    cityIds: ["city_ist", "city_ank"],
    startsAt: "2026-08-10T00:00:00Z",
    endsAt: "2026-09-10T00:00:00Z",
    budgetMinor: 500_000_00,
    spentMinor: 0,
    currency: "TRY",
    audience: {
      id: "aud_new",
      name: "New customers",
      segmentSize: 42000,
      rulesSummary: "First order not completed",
    },
    coupon: {
      code: "WELCOME50",
      discountType: "fixed",
      discountValue: 5000,
      maxRedemptions: 40000,
      redemptions: 0,
      minOrderMinor: 15000,
    },
    bundle: null,
    flashSale: null,
    personalized: null,
    createdAt: "2026-08-04T12:00:00Z",
    updatedAt: "2026-08-04T12:00:00Z",
  },
  {
    id: "cmp_bundle_03",
    name: "Breakfast bundle",
    type: "bundle",
    status: "paused",
    cityIds: ["city_ist"],
    startsAt: "2026-07-01T00:00:00Z",
    endsAt: "2026-08-31T00:00:00Z",
    budgetMinor: 120_000_00,
    spentMinor: 64_200_00,
    currency: "TRY",
    audience: null,
    coupon: null,
    bundle: {
      skuIds: ["sku_egg", "sku_bread", "sku_milk"],
      label: "Morning trio",
      bundlePriceMinor: 8990,
      savingsPct: 18,
    },
    flashSale: null,
    personalized: null,
    createdAt: "2026-06-20T10:00:00Z",
    updatedAt: "2026-08-02T14:00:00Z",
  },
  {
    id: "cmp_pers_04",
    name: "AI snack uplift",
    type: "personalized",
    status: "active",
    cityIds: ["city_ist"],
    startsAt: "2026-08-01T00:00:00Z",
    endsAt: "2026-08-31T00:00:00Z",
    budgetMinor: 80_000_00,
    spentMinor: 22_100_00,
    currency: "TRY",
    audience: {
      id: "aud_snack",
      name: "Snack affinity",
      segmentSize: 96000,
      rulesSummary: "Category snack affinity > 0.6",
    },
    coupon: null,
    bundle: null,
    flashSale: null,
    personalized: {
      model: "promo-ranker-v3",
      upliftPct: 7.4,
      channels: ["push", "in_app"],
    },
    createdAt: "2026-07-28T08:00:00Z",
    updatedAt: "2026-08-05T11:00:00Z",
  },
  {
    id: "cmp_aud_05",
    name: "VIP weekend",
    type: "audience",
    status: "draft",
    cityIds: ["city_ank"],
    startsAt: null,
    endsAt: null,
    budgetMinor: 40_000_00,
    spentMinor: 0,
    currency: "TRY",
    audience: {
      id: "aud_vip",
      name: "VIP gold+",
      segmentSize: 8500,
      rulesSummary: "Loyalty tier >= gold",
    },
    coupon: {
      code: "VIPWEEK",
      discountType: "percent",
      discountValue: 15,
      maxRedemptions: 8500,
      redemptions: 0,
      minOrderMinor: 20000,
    },
    bundle: null,
    flashSale: null,
    personalized: null,
    createdAt: "2026-08-05T16:00:00Z",
    updatedAt: "2026-08-05T16:00:00Z",
  },
];

/** Mock campaigns — replaced by GET /admin/campaigns when BFF is live. */
export async function fetchCampaigns(
  params: CampaignListParams = {},
): Promise<Paginated<Campaign>> {
  await delay();
  let items = [...mockCampaigns];
  if (params.status && params.status !== "all") {
    items = items.filter((c) => c.status === params.status);
  }
  if (params.type && params.type !== "all") {
    items = items.filter((c) => c.type === params.type);
  }
  if (params.cityId) {
    items = items.filter((c) => c.cityIds.includes(params.cityId!));
  }
  if (params.q?.trim()) {
    const q = params.q.trim().toLowerCase();
    items = items.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        c.id.toLowerCase().includes(q) ||
        c.coupon?.code.toLowerCase().includes(q),
    );
  }
  return { items, page: 1, pageSize: 50, total: items.length, hasMore: false };
}

export async function fetchCampaign(id: string): Promise<Campaign> {
  await delay();
  const found = mockCampaigns.find((c) => c.id === id);
  if (!found) {
    throw new ApiError(404, {
      code: "not_found",
      message: "Campaign not found",
      traceId: "mock",
    });
  }
  return { ...found };
}

export async function createCampaign(
  input: CampaignUpsertInput,
): Promise<Campaign> {
  await delay(260);
  const now = new Date().toISOString();
  const created: Campaign = {
    id: `cmp_${Date.now().toString(36)}`,
    name: input.name,
    type: input.type,
    status: "draft",
    cityIds: input.cityIds,
    startsAt: input.startsAt ?? null,
    endsAt: input.endsAt ?? null,
    budgetMinor: input.budgetMinor ?? 0,
    spentMinor: 0,
    currency: "TRY",
    audience: input.audienceId
      ? {
          id: input.audienceId,
          name: "Selected audience",
          segmentSize: 0,
          rulesSummary: "—",
        }
      : null,
    coupon: input.coupon ?? null,
    bundle: input.bundle ?? null,
    flashSale: input.flashSale ?? null,
    personalized: input.personalized ?? null,
    createdAt: now,
    updatedAt: now,
  };
  mockCampaigns = [created, ...mockCampaigns];
  return created;
}

export async function updateCampaign(
  id: string,
  input: Partial<CampaignUpsertInput>,
): Promise<Campaign> {
  await delay(240);
  const idx = mockCampaigns.findIndex((c) => c.id === id);
  if (idx < 0) {
    throw new ApiError(404, {
      code: "not_found",
      message: "Campaign not found",
      traceId: "mock",
    });
  }
  const prev = mockCampaigns[idx]!;
  const next: Campaign = {
    ...prev,
    name: input.name ?? prev.name,
    type: input.type ?? prev.type,
    cityIds: input.cityIds ?? prev.cityIds,
    startsAt: input.startsAt !== undefined ? input.startsAt : prev.startsAt,
    endsAt: input.endsAt !== undefined ? input.endsAt : prev.endsAt,
    budgetMinor: input.budgetMinor ?? prev.budgetMinor,
    coupon: input.coupon !== undefined ? input.coupon : prev.coupon,
    bundle: input.bundle !== undefined ? input.bundle : prev.bundle,
    flashSale:
      input.flashSale !== undefined ? input.flashSale : prev.flashSale,
    personalized:
      input.personalized !== undefined
        ? input.personalized
        : prev.personalized,
    updatedAt: new Date().toISOString(),
  };
  mockCampaigns = mockCampaigns.map((c, i) => (i === idx ? next : c));
  return next;
}

export async function duplicateCampaign(id: string): Promise<Campaign> {
  await delay(220);
  const src = mockCampaigns.find((c) => c.id === id);
  if (!src) {
    throw new ApiError(404, {
      code: "not_found",
      message: "Campaign not found",
      traceId: "mock",
    });
  }
  const now = new Date().toISOString();
  const copy: Campaign = {
    ...structuredClone(src),
    id: `cmp_${Date.now().toString(36)}`,
    name: `${src.name} (copy)`,
    status: "draft",
    spentMinor: 0,
    createdAt: now,
    updatedAt: now,
  };
  mockCampaigns = [copy, ...mockCampaigns];
  return copy;
}

function patchStatus(
  id: string,
  status: CampaignStatus,
  extra: Partial<Campaign> = {},
): Campaign {
  const idx = mockCampaigns.findIndex((c) => c.id === id);
  if (idx < 0) {
    throw new ApiError(404, {
      code: "not_found",
      message: "Campaign not found",
      traceId: "mock",
    });
  }
  const next = {
    ...mockCampaigns[idx]!,
    ...extra,
    status,
    updatedAt: new Date().toISOString(),
  };
  mockCampaigns = mockCampaigns.map((c, i) => (i === idx ? next : c));
  return next;
}

export async function scheduleCampaign(
  id: string,
  startsAt: string,
  endsAt: string | null,
): Promise<Campaign> {
  await delay(200);
  return patchStatus(id, "scheduled", { startsAt, endsAt });
}

export async function pauseCampaign(id: string): Promise<Campaign> {
  await delay(160);
  return patchStatus(id, "paused");
}

export async function resumeCampaign(id: string): Promise<Campaign> {
  await delay(160);
  return patchStatus(id, "active");
}

export const CAMPAIGN_TYPES: CampaignType[] = [
  "coupon",
  "bundle",
  "flash_sale",
  "personalized",
  "audience",
];

export const CAMPAIGN_STATUSES: CampaignStatus[] = [
  "draft",
  "scheduled",
  "active",
  "paused",
  "ended",
];
