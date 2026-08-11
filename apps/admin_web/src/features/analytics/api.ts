import type { AnalyticsSnapshot, CohortCell } from "./types";

function weeks(): string[] {
  return ["W01", "W02", "W03", "W04", "W05", "W06", "W07", "W08"];
}

function buildCohorts(): CohortCell[] {
  const cohorts = ["Jan", "Feb", "Mar", "Apr", "May", "Jun"];
  const cells: CohortCell[] = [];
  for (const cohort of cohorts) {
    let base = 100;
    for (let week = 0; week < 8; week++) {
      base = Math.max(12, base * (0.72 + (week % 3) * 0.02));
      cells.push({
        cohort,
        week,
        retentionPct: Math.round(base * 10) / 10,
      });
    }
  }
  return cells;
}

/** Mock analytics — replaced by GET /admin/analytics/* when BFF is live. */
export async function fetchAnalyticsSnapshot(
  cityId: string | null,
): Promise<AnalyticsSnapshot> {
  await new Promise((r) => setTimeout(r, 260));
  const days = weeks();

  return {
    cityId,
    generatedAt: new Date().toISOString(),
    kpis: [
      {
        id: "sales",
        label: "Sales (7d)",
        value: 92_450,
        unit: "count",
        deltaPct: 6.2,
        tone: "brand",
      },
      {
        id: "revenue",
        label: "Revenue (7d)",
        value: 34_812_000_00,
        unit: "currency",
        currency: "TRY",
        deltaPct: 4.8,
        tone: "success",
      },
      {
        id: "retention",
        label: "D30 retention",
        value: 38.4,
        unit: "percent",
        deltaPct: 1.1,
        tone: "success",
      },
      {
        id: "clv",
        label: "Avg CLV",
        value: 1_845_00,
        unit: "currency",
        currency: "TRY",
        deltaPct: 2.4,
        tone: "neutral",
      },
      {
        id: "conv",
        label: "Checkout conv.",
        value: 68.2,
        unit: "percent",
        deltaPct: -0.6,
        tone: "warning",
      },
      {
        id: "cancel",
        label: "Cancel rate",
        value: 2.3,
        unit: "percent",
        deltaPct: 0.2,
        tone: "danger",
      },
      {
        id: "refund",
        label: "Refund rate",
        value: 1.1,
        unit: "percent",
        deltaPct: -0.1,
        tone: "success",
      },
      {
        id: "aov",
        label: "AOV",
        value: 376_50,
        unit: "currency",
        currency: "TRY",
        deltaPct: 0.9,
        tone: "neutral",
      },
    ],
    salesSeries: days.map((d, i) => ({
      label: d,
      value: 10_200 + i * 820 + (i % 2) * 400,
    })),
    revenueSeries: days.map((d, i) => ({
      label: d,
      value: 3_800_000_00 + i * 310_000_00 + (i % 3) * 90_000_00,
    })),
    retentionSeries: [
      { label: "D1", value: 62 },
      { label: "D7", value: 48 },
      { label: "D14", value: 42 },
      { label: "D30", value: 38.4 },
      { label: "D60", value: 31 },
      { label: "D90", value: 26 },
    ],
    funnel: [
      { id: "f1", label: "App open", count: 420_000, conversionPct: 100 },
      { id: "f2", label: "Browse", count: 310_000, conversionPct: 73.8 },
      { id: "f3", label: "Add to cart", count: 98_000, conversionPct: 31.6 },
      { id: "f4", label: "Checkout", count: 72_000, conversionPct: 73.5 },
      { id: "f5", label: "Paid", count: 49_100, conversionPct: 68.2 },
    ],
    cohorts: buildCohorts(),
    clvSeries: days.map((d, i) => ({
      label: d,
      value: 1600 + i * 35 + (i % 2) * 20,
    })),
    warehouses: [
      {
        id: "wh1",
        name: "WH-Kadıköy",
        orders: 4200,
        pickMins: 6.2,
        slaPct: 96.1,
        stockouts: 12,
      },
      {
        id: "wh2",
        name: "WH-Beşiktaş",
        orders: 3800,
        pickMins: 7.8,
        slaPct: 93.4,
        stockouts: 28,
      },
      {
        id: "wh3",
        name: "WH-Şişli",
        orders: 5100,
        pickMins: 5.9,
        slaPct: 97.2,
        stockouts: 8,
      },
      {
        id: "wh4",
        name: "WH-Üsküdar",
        orders: 2900,
        pickMins: 8.4,
        slaPct: 91.8,
        stockouts: 34,
      },
    ],
    couriers: [
      {
        id: "c1",
        name: "Top cohort A",
        deliveries: 1840,
        onTimePct: 97.2,
        avgMins: 16,
        rating: 4.9,
      },
      {
        id: "c2",
        name: "Cohort B",
        deliveries: 1620,
        onTimePct: 94.1,
        avgMins: 18,
        rating: 4.7,
      },
      {
        id: "c3",
        name: "Cohort C",
        deliveries: 1410,
        onTimePct: 91.5,
        avgMins: 21,
        rating: 4.5,
      },
      {
        id: "c4",
        name: "New hires",
        deliveries: 980,
        onTimePct: 88.2,
        avgMins: 24,
        rating: 4.3,
      },
    ],
    conversionSeries: days.map((d, i) => ({
      label: d,
      value: 66 + (i % 4) * 1.2 - (i % 3) * 0.5,
    })),
    cancelReasons: [
      { reason: "Customer changed mind", count: 420, pct: 28 },
      { reason: "Out of stock", count: 310, pct: 21 },
      { reason: "Courier delay", count: 245, pct: 16 },
      { reason: "Payment failed", count: 198, pct: 13 },
      { reason: "Address issue", count: 156, pct: 10 },
      { reason: "Other", count: 180, pct: 12 },
    ],
    refunds: [
      { label: "Missing item", amountMinor: 182_000_00, count: 210 },
      { label: "Quality", amountMinor: 96_000_00, count: 88 },
      { label: "Late delivery", amountMinor: 74_000_00, count: 142 },
      { label: "Wrong item", amountMinor: 58_000_00, count: 64 },
      { label: "Promo dispute", amountMinor: 31_000_00, count: 39 },
    ],
    demandActual: days.map((d, i) => ({
      label: d,
      value: 11_000 + i * 700,
    })),
    demandForecast: days.map((d, i) => ({
      label: d,
      value: 10_800 + i * 740 + (i > 5 ? 400 : 0),
    })),
  };
}
