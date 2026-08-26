import { apiClient, platformPath } from "@/shared/api/client";
import type { PlatformDashboardSnapshot } from "./types";

/** Live platform stats from platform-ops-service. Throws when unavailable. */
export async function fetchPlatformDashboard(
  tenantContextId: string | null,
): Promise<PlatformDashboardSnapshot> {
  const raw = await apiClient<Record<string, unknown>>(platformPath("/admin/stats"));
  const deployments = Number(raw.deployments ?? raw.activeDeployments ?? 0);
  return {
    tenantContextId,
    generatedAt: new Date().toISOString(),
    kpis: [
      {
        id: "deployments",
        label: "Deployments",
        value: deployments,
        unit: "count",
        deltaPct: 0,
        tone: "neutral",
      },
    ],
    revenueSeries: [],
    trafficSeries: [],
    costSeries: [],
    alerts: [],
    incidents: [],
    systemHealth: [
      { id: "platform-ops", name: "platform-ops-service", status: "healthy", latencyMs: 0 },
    ],
  };
}
