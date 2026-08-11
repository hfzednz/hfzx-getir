import type { DashboardSnapshot } from "./types";

/** Mock dashboard payload — replaced by GET /admin/dashboard when BFF is live. */
export async function fetchDashboardSnapshot(
  cityId: string | null,
): Promise<DashboardSnapshot> {
  // Simulate network
  await new Promise((r) => setTimeout(r, 280));

  const hours = ["06", "08", "10", "12", "14", "16", "18", "20", "22"];

  return {
    cityId,
    generatedAt: new Date().toISOString(),
    kpis: [
      {
        id: "gmv",
        label: "GMV (today)",
        value: 4_875_400_00,
        unit: "currency",
        currency: "TRY",
        deltaPct: 8.4,
        tone: "brand",
      },
      {
        id: "orders",
        label: "Orders",
        value: 12840,
        unit: "count",
        deltaPct: 5.1,
        tone: "success",
      },
      {
        id: "aov",
        label: "AOV",
        value: 379_50,
        unit: "currency",
        currency: "TRY",
        deltaPct: 1.2,
        tone: "neutral",
      },
      {
        id: "sla",
        label: "SLA on-time",
        value: 94.2,
        unit: "percent",
        deltaPct: -0.8,
        tone: "warning",
      },
      {
        id: "eta",
        label: "Median ETA",
        value: 18,
        unit: "minutes",
        deltaPct: -3.4,
        tone: "success",
      },
      {
        id: "couriers",
        label: "Active couriers",
        value: 1842,
        unit: "count",
        deltaPct: 2.0,
        tone: "neutral",
      },
      {
        id: "cancel",
        label: "Cancel rate",
        value: 2.1,
        unit: "percent",
        deltaPct: 0.3,
        tone: "danger",
      },
      {
        id: "nps",
        label: "CSAT",
        value: 4.6,
        unit: "count",
        deltaPct: 0.1,
        tone: "success",
      },
    ],
    revenueSeries: hours.map((h, i) => ({
      label: `${h}:00`,
      value: 180_000_00 + i * 42_000_00 + (i % 3) * 15_000_00,
    })),
    ordersSeries: hours.map((h, i) => ({
      label: `${h}:00`,
      value: 420 + i * 95 + (i % 2) * 40,
    })),
    slaSeries: hours.map((h, i) => ({
      label: `${h}:00`,
      value: Math.min(98, 88 + i * 0.7 + (i % 4 === 0 ? -2 : 1)),
    })),
    alerts: [
      {
        id: "a1",
        severity: "danger",
        title: "Warehouse WH-14 capacity",
        detail: "Pick queue > 12 min · zone Kadıköy",
        createdAt: new Date(Date.now() - 4 * 60_000).toISOString(),
      },
      {
        id: "a2",
        severity: "warning",
        title: "Courier shortage",
        detail: "Beşiktaş hex understaffed for dinner peak",
        createdAt: new Date(Date.now() - 12 * 60_000).toISOString(),
      },
      {
        id: "a3",
        severity: "info",
        title: "Campaign spike",
        detail: "Flash deal driving +22% orders vs forecast",
        createdAt: new Date(Date.now() - 28 * 60_000).toISOString(),
      },
    ],
    aiInsights: [
      {
        id: "i1",
        title: "Pre-position inventory",
        summary:
          "Move cold SKUs toward WH-07 before 17:00 — predicted stockout on ice cream category.",
        confidence: 0.86,
      },
      {
        id: "i2",
        title: "Surge pricing window",
        summary:
          "Open soft surge in Şişli 18:30–20:00 to protect SLA without killing conversion.",
        confidence: 0.79,
      },
      {
        id: "i3",
        title: "Fraud hold pattern",
        summary:
          "New device clusters on prepaid high-AOV carts — review 14 orders in fraud queue.",
        confidence: 0.91,
      },
    ],
    systemHealth: [
      { id: "h1", name: "bff-admin", status: "healthy", latencyMs: 42 },
      { id: "h2", name: "order-service", status: "healthy", latencyMs: 68 },
      { id: "h3", name: "courier-gateway", status: "degraded", latencyMs: 210 },
      { id: "h4", name: "payments", status: "healthy", latencyMs: 95 },
    ],
  };
}
