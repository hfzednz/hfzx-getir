import { apiClient } from "@/shared/api/client";
import type { DashboardSnapshot } from "./types";

/** Live admin dashboard from bff-admin. Throws when BFF unavailable. */
export async function fetchDashboardSnapshot(
  cityId: string | null,
): Promise<DashboardSnapshot> {
  const qs = cityId ? `?cityId=${encodeURIComponent(cityId)}` : "";
  const raw = await apiClient<Record<string, unknown>>(`/admin/dashboard${qs}`);
  const orderCount = Number(raw.orderCount ?? raw.orders ?? 0);
  return {
    cityId,
    generatedAt: new Date().toISOString(),
    kpis: [
      {
        id: "orders",
        label: "Orders",
        value: orderCount,
        unit: "count",
        deltaPct: 0,
        tone: "neutral",
      },
    ],
    revenueSeries: [],
    ordersSeries: [],
    slaSeries: [],
    alerts: [],
    aiInsights: [],
    systemHealth: [
      { id: "bff-admin", name: "bff-admin", status: "healthy", latencyMs: 0 },
    ],
  };
}
