import type { AiCommandSnapshot } from "./types";
import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient } from "@/shared/api/client";

export async function fetchAiCommandSnapshot(
  cityId: string | null,
): Promise<AiCommandSnapshot> {
  try {
    const q = cityId ? `?cityId=${encodeURIComponent(cityId)}` : "";
    return await apiClient<AiCommandSnapshot>(`/admin/ai${q}`);
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    return mockAiCommandSnapshot(cityId);
  }
}

async function mockAiCommandSnapshot(
  cityId: string | null,
): Promise<AiCommandSnapshot> {
  await new Promise((r) => setTimeout(r, 240));
  const hours = ["10", "12", "14", "16", "18", "20", "22"];

  return {
    cityId,
    generatedAt: new Date().toISOString(),
    kpis: [
      {
        id: "models",
        label: "Models online",
        value: 12,
        unit: "count",
        tone: "success",
      },
      {
        id: "fraud",
        label: "Fraud queue",
        value: 14,
        unit: "count",
        tone: "warning",
      },
      {
        id: "prec",
        label: "Demand MAPE",
        value: 6.4,
        unit: "percent",
        tone: "brand",
      },
      {
        id: "rec",
        label: "Rec CTR",
        value: 4.8,
        unit: "percent",
        tone: "success",
      },
      {
        id: "price",
        label: "Pricing lifts",
        value: 7,
        unit: "count",
        tone: "neutral",
      },
      {
        id: "risk",
        label: "High risks",
        value: 3,
        unit: "count",
        tone: "danger",
      },
    ],
    demandForecast: hours.map((h, i) => ({
      label: `${h}:00`,
      value: 900 + i * 180 + (i === 4 ? 220 : 0),
    })),
    inventoryForecast: hours.map((h, i) => ({
      label: `${h}:00`,
      value: Math.max(40, 220 - i * 18 - (i > 4 ? 30 : 0)),
    })),
    fraudAlerts: [
      {
        id: "fa1",
        orderId: "ord_9f2a",
        score: 0.94,
        reason: "New device + prepaid + high AOV",
        status: "open",
        createdAt: new Date(Date.now() - 8 * 60_000).toISOString(),
      },
      {
        id: "fa2",
        orderId: "ord_3bc1",
        score: 0.88,
        reason: "Velocity: 5 carts / 12 min",
        status: "reviewing",
        createdAt: new Date(Date.now() - 22 * 60_000).toISOString(),
      },
      {
        id: "fa3",
        orderId: "ord_77de",
        score: 0.81,
        reason: "Address mismatch cluster",
        status: "open",
        createdAt: new Date(Date.now() - 41 * 60_000).toISOString(),
      },
      {
        id: "fa4",
        orderId: "ord_11aa",
        score: 0.72,
        reason: "Promo abuse pattern",
        status: "cleared",
        createdAt: new Date(Date.now() - 90 * 60_000).toISOString(),
      },
    ],
    deliveryOptSeries: hours.map((h, i) => ({
      label: `${h}:00`,
      value: 18 - i * 0.4 + (i % 2),
    })),
    recommendationCtr: hours.map((h, i) => ({
      label: `${h}:00`,
      value: 3.8 + (i % 4) * 0.35,
    })),
    pricingRecs: [
      {
        id: "pr1",
        zone: "Kadıköy",
        skuCategory: "Ice cream",
        currentMultiplier: 1.0,
        suggestedMultiplier: 1.08,
        liftPct: 3.2,
        confidence: 0.84,
      },
      {
        id: "pr2",
        zone: "Beşiktaş",
        skuCategory: "Beverages",
        currentMultiplier: 1.05,
        suggestedMultiplier: 1.12,
        liftPct: 2.1,
        confidence: 0.79,
      },
      {
        id: "pr3",
        zone: "Şişli",
        skuCategory: "Snacks",
        currentMultiplier: 1.0,
        suggestedMultiplier: 0.97,
        liftPct: 1.4,
        confidence: 0.71,
      },
    ],
    campaignOpts: [
      {
        id: "co1",
        campaign: "Flash Friday",
        recommendation: "Narrow geo to underutilized hexes 18–20",
        expectedRoiLift: 12,
        status: "pending",
      },
      {
        id: "co2",
        campaign: "Free delivery >200₺",
        recommendation: "Raise threshold to 250₺ in peak zones",
        expectedRoiLift: 8,
        status: "applied",
      },
      {
        id: "co3",
        campaign: "New user 20%",
        recommendation: "Cap to first 2 orders to cut abuse",
        expectedRoiLift: 15,
        status: "pending",
      },
    ],
    segments: [
      {
        id: "s1",
        name: "Power weekly",
        size: 42_100,
        avgAovMinor: 485_00,
        churnRisk: 0.08,
      },
      {
        id: "s2",
        name: "Promo hunters",
        size: 68_400,
        avgAovMinor: 210_00,
        churnRisk: 0.22,
      },
      {
        id: "s3",
        name: "At-risk dormant",
        size: 31_200,
        avgAovMinor: 340_00,
        churnRisk: 0.61,
      },
      {
        id: "s4",
        name: "High-value nights",
        size: 18_900,
        avgAovMinor: 620_00,
        churnRisk: 0.11,
      },
    ],
    risks: [
      {
        id: "r1",
        area: "Inventory",
        severity: "high",
        summary: "Cold chain SKUs WH-14 predicted stockout < 3h",
      },
      {
        id: "r2",
        area: "Fraud",
        severity: "medium",
        summary: "Device cluster growth +18% vs 7d baseline",
      },
      {
        id: "r3",
        area: "SLA",
        severity: "high",
        summary: "Beşiktaş dinner peak understaffed vs demand model",
      },
      {
        id: "r4",
        area: "Pricing",
        severity: "low",
        summary: "Competitor undercut on beverages in 2 hexes",
      },
    ],
    opsInsights: [
      {
        id: "oi1",
        title: "Pre-position ice cream",
        detail: "Shift 180 units WH-07 → WH-14 before 17:00.",
        confidence: 0.86,
      },
      {
        id: "oi2",
        title: "Batch courier staging",
        detail: "Stage 24 couriers near Beşiktaş hexes 18:20.",
        confidence: 0.81,
      },
      {
        id: "oi3",
        title: "Suppress low-ROI push",
        detail: "Pause segment Promo hunters until CTR recovers.",
        confidence: 0.77,
      },
    ],
  };
}
