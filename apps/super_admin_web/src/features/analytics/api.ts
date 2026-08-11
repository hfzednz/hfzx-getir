import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { AnalyticsScope, AnalyticsSnapshot } from "./types";

function mockSnapshot(
  scope: AnalyticsScope,
  scopeId: string | null,
): AnalyticsSnapshot {
  const h = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
  const scale =
    scope === "worldwide" ? 1 : scope === "country" ? 0.18 : 0.04;

  return {
    generatedAt: new Date().toISOString(),
    scope,
    scopeId,
    kpis: [
      {
        id: "revenue",
        label:
          scope === "worldwide"
            ? "Worldwide GMV"
            : scope === "country"
              ? "Country GMV"
              : "Company GMV",
        value: Math.round(128_450_000_00 * scale),
        unit: "currency",
        currency: "USD",
        deltaPct: 6.2,
        tone: "brand",
      },
      {
        id: "orders",
        label: "Orders (24h)",
        value: Math.round(2_184_320 * scale),
        unit: "count",
        deltaPct: 4.1,
        tone: "success",
      },
      {
        id: "companies",
        label: "Active companies",
        value: scope === "company" ? 1 : scope === "country" ? 42 : 318,
        unit: "count",
        deltaPct: 1.2,
        tone: "neutral",
      },
      {
        id: "countries",
        label: "Countries live",
        value: scope === "worldwide" ? 28 : 1,
        unit: "count",
        deltaPct: 0,
        tone: "neutral",
      },
      {
        id: "aov",
        label: "Avg order value",
        value: 1840,
        unit: "currency",
        currency: "USD",
        deltaPct: 0.8,
        tone: "success",
      },
      {
        id: "fill",
        label: "Fill rate",
        value: 97.4,
        unit: "percent",
        deltaPct: 0.3,
        tone: "success",
      },
    ],
    revenueSeries: h.map((label, i) => ({
      label,
      value: Math.round((18_000_000_00 + i * 1_200_000_00) * scale),
    })),
    ordersSeries: h.map((label, i) => ({
      label,
      value: Math.round((280_000 + i * 18_000) * scale),
    })),
    byCountry: [
      { label: "TR", value: 28 },
      { label: "DE", value: 16 },
      { label: "UK", value: 14 },
      { label: "NL", value: 9 },
      { label: "US", value: 8 },
      { label: "Other", value: 25 },
    ],
    byCompany: [
      { label: "Getir TR", value: 22 },
      { label: "Nexora DE", value: 12 },
      { label: "Quick NL", value: 9 },
      { label: "Flash UK", value: 8 },
      { label: "Other", value: 49 },
    ],
    countries: [
      {
        id: "ct-tr",
        country: "Türkiye",
        code: "TR",
        revenueUsd: 36_000_000_00,
        orders24h: 620_000,
        activeCompanies: 48,
        gmvGrowthPct: 7.4,
      },
      {
        id: "ct-de",
        country: "Germany",
        code: "DE",
        revenueUsd: 20_500_000_00,
        orders24h: 310_000,
        activeCompanies: 36,
        gmvGrowthPct: 5.1,
      },
      {
        id: "ct-uk",
        country: "United Kingdom",
        code: "UK",
        revenueUsd: 18_200_000_00,
        orders24h: 280_000,
        activeCompanies: 28,
        gmvGrowthPct: 3.8,
      },
      {
        id: "ct-nl",
        country: "Netherlands",
        code: "NL",
        revenueUsd: 11_400_000_00,
        orders24h: 190_000,
        activeCompanies: 22,
        gmvGrowthPct: 4.6,
      },
      {
        id: "ct-us",
        country: "United States",
        code: "US",
        revenueUsd: 10_200_000_00,
        orders24h: 150_000,
        activeCompanies: 18,
        gmvGrowthPct: 2.1,
      },
    ],
    companies: [
      {
        id: "co1",
        company: "Getir TR Holdings",
        country: "TR",
        revenueUsd: 28_400_000_00,
        orders24h: 480_000,
        warehouses: 420,
        healthScore: 94,
      },
      {
        id: "co2",
        company: "Nexora Germany GmbH",
        country: "DE",
        revenueUsd: 12_100_000_00,
        orders24h: 180_000,
        warehouses: 96,
        healthScore: 91,
      },
      {
        id: "co3",
        company: "Flash Commerce UK Ltd",
        country: "UK",
        revenueUsd: 9_800_000_00,
        orders24h: 140_000,
        warehouses: 72,
        healthScore: 88,
      },
      {
        id: "co4",
        company: "Quick NL B.V.",
        country: "NL",
        revenueUsd: 7_200_000_00,
        orders24h: 110_000,
        warehouses: 54,
        healthScore: 92,
      },
      {
        id: "co5",
        company: "Rapid US Inc",
        country: "US",
        revenueUsd: 6_400_000_00,
        orders24h: 95_000,
        warehouses: 48,
        healthScore: 85,
      },
    ],
  };
}

export async function fetchAnalyticsSnapshot(
  scope: AnalyticsScope,
  scopeId: string | null,
): Promise<AnalyticsSnapshot> {
  try {
    const qs = new URLSearchParams({ scope });
    if (scopeId) qs.set("scopeId", scopeId);
    return await apiClient<AnalyticsSnapshot>(
      `${platformPath("/analytics")}?${qs.toString()}`,
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await new Promise((r) => setTimeout(r, 200));
      return mockSnapshot(scope, scopeId);
    }
    throw err;
  }
}
