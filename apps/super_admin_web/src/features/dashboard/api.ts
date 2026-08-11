import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { PlatformDashboardSnapshot } from "./types";

function mockSnapshot(
  tenantContextId: string | null,
): PlatformDashboardSnapshot {
  const hours = ["00", "04", "08", "12", "16", "20"];

  return {
    tenantContextId,
    generatedAt: new Date().toISOString(),
    kpis: [
      {
        id: "revenue",
        label: "Worldwide revenue",
        value: 128_450_000_00,
        unit: "currency",
        currency: "USD",
        deltaPct: 6.2,
        tone: "brand",
      },
      {
        id: "orders",
        label: "Orders (24h)",
        value: 2_184_320,
        unit: "count",
        deltaPct: 4.1,
        tone: "success",
      },
      {
        id: "warehouses",
        label: "Warehouses",
        value: 1842,
        unit: "count",
        deltaPct: 1.0,
        tone: "neutral",
      },
      {
        id: "couriers",
        label: "Active couriers",
        value: 96_420,
        unit: "count",
        deltaPct: 2.4,
        tone: "neutral",
      },
      {
        id: "customers",
        label: "Customers",
        value: 48_210_000,
        unit: "count",
        deltaPct: 0.8,
        tone: "success",
      },
      {
        id: "inventory",
        label: "SKU units",
        value: 312_000_000,
        unit: "count",
        deltaPct: -0.3,
        tone: "neutral",
      },
      {
        id: "server_health",
        label: "Server health",
        value: 99.4,
        unit: "percent",
        deltaPct: 0.1,
        tone: "success",
      },
      {
        id: "api_traffic",
        label: "API req/min",
        value: 1_240_000,
        unit: "count",
        deltaPct: 12.5,
        tone: "brand",
      },
      {
        id: "cloud_cost",
        label: "Cloud cost (MTD)",
        value: 1_842_000_00,
        unit: "currency",
        currency: "USD",
        deltaPct: 3.2,
        tone: "warning",
      },
      {
        id: "ai_usage",
        label: "AI tokens (24h)",
        value: 8_420_000_000,
        unit: "count",
        deltaPct: 18.0,
        tone: "brand",
      },
      {
        id: "fraud",
        label: "Fraud / security",
        value: 37,
        unit: "count",
        deltaPct: -12.0,
        tone: "danger",
      },
      {
        id: "incidents",
        label: "Open incidents",
        value: 5,
        unit: "count",
        deltaPct: -2.0,
        tone: "warning",
      },
    ],
    revenueSeries: hours.map((h, i) => ({
      label: `${h}:00`,
      value: 8_200_000_00 + i * 1_400_000_00 + (i % 2) * 600_000_00,
    })),
    trafficSeries: hours.map((h, i) => ({
      label: `${h}:00`,
      value: 620_000 + i * 95_000 + (i % 3) * 40_000,
    })),
    costSeries: hours.map((h, i) => ({
      label: `${h}:00`,
      value: 48_000_00 + i * 6_500_00,
    })),
    alerts: [
      {
        id: "a1",
        severity: "danger",
        title: "Credential stuffing wave",
        detail: "Identity edge · EU-WEST · 14k failed logins / 5m",
        createdAt: new Date(Date.now() - 6 * 60_000).toISOString(),
      },
      {
        id: "a2",
        severity: "warning",
        title: "Kafka lag rising",
        detail: "order.events · partition 7 lag > 120k",
        createdAt: new Date(Date.now() - 18 * 60_000).toISOString(),
      },
      {
        id: "a3",
        severity: "info",
        title: "License renewals due",
        detail: "3 enterprise tenants renew within 14 days",
        createdAt: new Date(Date.now() - 45 * 60_000).toISOString(),
      },
    ],
    incidents: [
      {
        id: "inc1",
        severity: "sev2",
        title: "Payments latency EU",
        status: "mitigating",
        region: "eu-west-1",
      },
      {
        id: "inc2",
        severity: "sev3",
        title: "CDN cache miss spike APAC",
        status: "investigating",
        region: "ap-southeast-1",
      },
      {
        id: "inc3",
        severity: "sev1",
        title: "Auth MFA provider outage",
        status: "resolved",
        region: "global",
      },
    ],
    systemHealth: [
      { id: "h1", name: "bff-admin /platform", status: "healthy", latencyMs: 38 },
      { id: "h2", name: "identity-service", status: "healthy", latencyMs: 54 },
      { id: "h3", name: "config / flags", status: "healthy", latencyMs: 22 },
      { id: "h4", name: "k8s control plane", status: "degraded", latencyMs: 180 },
      { id: "h5", name: "billing / license", status: "healthy", latencyMs: 71 },
    ],
  };
}

/**
 * Platform dashboard snapshot.
 * Prefers GET /platform/dashboard; falls back to mock when BFF is unavailable.
 */
export async function fetchPlatformDashboard(
  tenantContextId: string | null,
): Promise<PlatformDashboardSnapshot> {
  try {
    const qs =
      tenantContextId != null
        ? `?tenantId=${encodeURIComponent(tenantContextId)}`
        : "";
    return await apiClient<PlatformDashboardSnapshot>(
      `${platformPath("/dashboard")}${qs}`,
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await new Promise((r) => setTimeout(r, 220));
      return mockSnapshot(tenantContextId);
    }
    throw err;
  }
}
