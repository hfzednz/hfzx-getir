import type { NotificationsSnapshot } from "./types";

/** Mock ops alerts inbox — replaced by GET /admin/notifications when BFF is live. */
export async function fetchNotificationsSnapshot(): Promise<NotificationsSnapshot> {
  await new Promise((r) => setTimeout(r, 180));
  const t = Date.now();

  return {
    generatedAt: new Date().toISOString(),
    alerts: [
      {
        id: "n1",
        category: "emergency",
        title: "Warehouse WH-14 pick queue critical",
        body: "Pick wait > 12 min · Kadıköy dinner peak",
        severity: "danger",
        read: false,
        createdAt: new Date(t - 4 * 60_000).toISOString(),
      },
      {
        id: "n2",
        category: "stock",
        title: "Ice cream SKU stockout risk",
        body: "Predicted cover < 2h at WH-07",
        severity: "warning",
        read: false,
        createdAt: new Date(t - 11 * 60_000).toISOString(),
      },
      {
        id: "n3",
        category: "security",
        title: "Fraud cluster detected",
        body: "14 orders held · new device prepaid pattern",
        severity: "danger",
        read: false,
        createdAt: new Date(t - 18 * 60_000).toISOString(),
      },
      {
        id: "n4",
        category: "financial",
        title: "Payout batch awaiting dual-control",
        body: "3 courier payouts · ₺184,200",
        severity: "warning",
        read: true,
        createdAt: new Date(t - 40 * 60_000).toISOString(),
      },
      {
        id: "n5",
        category: "system",
        title: "courier-gateway latency elevated",
        body: "p95 210ms · degraded status",
        severity: "warning",
        read: true,
        createdAt: new Date(t - 55 * 60_000).toISOString(),
      },
      {
        id: "n6",
        category: "operational",
        title: "Courier shortage Beşiktaş",
        body: "Understaffed hexes for dinner window",
        severity: "warning",
        read: false,
        createdAt: new Date(t - 70 * 60_000).toISOString(),
      },
      {
        id: "n7",
        category: "emergency",
        title: "Weather advisory",
        body: "Heavy rain · ETA buffer +4 min city-wide",
        severity: "info",
        read: true,
        createdAt: new Date(t - 120 * 60_000).toISOString(),
      },
      {
        id: "n8",
        category: "financial",
        title: "Refund spike",
        body: "Missing-item refunds +28% vs yesterday",
        severity: "info",
        read: true,
        createdAt: new Date(t - 150 * 60_000).toISOString(),
      },
    ],
  };
}
