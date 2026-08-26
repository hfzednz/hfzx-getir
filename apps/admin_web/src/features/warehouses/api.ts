import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient } from "@/shared/api/client";
import type {
  WarehouseDetail,
  WarehouseListItem,
  WarehouseListResponse,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const MOCK_LIST: WarehouseListItem[] = [
  {
    id: "wh_07",
    code: "WH-07",
    name: "Kadıköy Dark Store",
    cityId: "city_ist",
    district: "Kadıköy",
    status: "busy",
    capacityPct: 92,
    skuCount: 4820,
    openOrders: 146,
    pickSlaPct: 88.2,
    stockAlerts: 12,
  },
  {
    id: "wh_14",
    code: "WH-14",
    name: "Beşiktaş Hub",
    cityId: "city_ist",
    district: "Beşiktaş",
    status: "open",
    capacityPct: 71,
    skuCount: 5102,
    openOrders: 64,
    pickSlaPct: 95.4,
    stockAlerts: 3,
  },
  {
    id: "wh_21",
    code: "WH-21",
    name: "Şişli Express",
    cityId: "city_ist",
    district: "Şişli",
    status: "open",
    capacityPct: 64,
    skuCount: 3980,
    openOrders: 41,
    pickSlaPct: 96.1,
    stockAlerts: 5,
  },
  {
    id: "wh_03",
    code: "WH-03",
    name: "Çankaya Store",
    cityId: "city_ank",
    district: "Çankaya",
    status: "maintenance",
    capacityPct: 40,
    skuCount: 2200,
    openOrders: 0,
    pickSlaPct: 0,
    stockAlerts: 1,
  },
];

function mockDetail(id: string): WarehouseDetail {
  const base =
    MOCK_LIST.find((w) => w.id === id) ??
    ({ ...MOCK_LIST[0], id, code: id.toUpperCase() } satisfies WarehouseListItem);

  return {
    id: base.id,
    code: base.code,
    name: base.name,
    cityId: base.cityId,
    district: base.district,
    address: `${base.district} Mah. NEXORA Cad. No:12`,
    status: base.status,
    managerName: "Selin Acar",
    openedAt: "2023-06-01T08:00:00.000Z",
    kpis: {
      capacityPct: base.capacityPct,
      skuCount: base.skuCount,
      unitsOnHand: 86_420,
      pickSlaPct: base.pickSlaPct,
      packSlaPct: Math.min(99, base.pickSlaPct + 2.1),
      dispatchSlaPct: Math.min(99, base.pickSlaPct + 1.4),
      avgPickMinutes: 4.2,
      avgPackMinutes: 2.8,
      openTransfers: 3,
      stockAlerts: base.stockAlerts,
    },
    inventorySummary: [
      { category: "Grocery", skuCount: 2100, units: 42_000 },
      { category: "Cold", skuCount: 820, units: 11_200 },
      { category: "Frozen", skuCount: 410, units: 6_400 },
      { category: "Personal care", skuCount: 940, units: 18_000 },
      { category: "Home", skuCount: 550, units: 8_820 },
    ],
    transfers: [
      {
        id: "tr_1",
        direction: "in",
        counterpart: "DC-IST-01",
        skuCount: 120,
        status: "in_transit",
        etaAt: new Date(Date.now() + 90 * 60_000).toISOString(),
      },
      {
        id: "tr_2",
        direction: "out",
        counterpart: "WH-14",
        skuCount: 45,
        status: "picking",
        etaAt: new Date(Date.now() + 40 * 60_000).toISOString(),
      },
      {
        id: "tr_3",
        direction: "in",
        counterpart: "Supplier Nestlé",
        skuCount: 80,
        status: "scheduled",
        etaAt: new Date(Date.now() + 5 * 60 * 60_000).toISOString(),
      },
    ],
    audits: [
      {
        id: "au_1",
        type: "Cycle count · Cold",
        result: "pass",
        auditor: "Ops QA",
        auditedAt: new Date(Date.now() - 2 * 86_400_000).toISOString(),
      },
      {
        id: "au_2",
        type: "Safety inspection",
        result: "minor_findings",
        auditor: "HSE",
        auditedAt: new Date(Date.now() - 10 * 86_400_000).toISOString(),
      },
    ],
    stockAlerts: [
      {
        id: "sa_1",
        sku: "SKU-ICE-044",
        productName: "Magnum Classic",
        severity: "danger",
        onHand: 8,
        safetyStock: 40,
        message: "Predicted stockout before 17:00",
      },
      {
        id: "sa_2",
        sku: "SKU-MLK-012",
        productName: "Full fat milk 1L",
        severity: "warning",
        onHand: 52,
        safetyStock: 80,
        message: "Below safety stock",
      },
    ],
    aiOptimization: [
      {
        id: "ai_1",
        title: "Re-slot high movers",
        summary:
          "Move top 40 SKUs closer to pack stations before dinner peak.",
        impact: "-1.8 min pick time",
        confidence: 0.84,
      },
      {
        id: "ai_2",
        title: "Pre-position cold inventory",
        summary: "Pull ice cream from WH-14 buffer lane by 15:30.",
        impact: "Avoid stockout WH-07",
        confidence: 0.91,
      },
      {
        id: "ai_3",
        title: "Staffing nudge",
        summary: "Add 2 pickers 17:00–20:00 to recover pick SLA.",
        impact: "+4.5 pts pick SLA",
        confidence: 0.77,
      },
    ],
  };
}

export async function fetchWarehouses(
  cityId: string | null,
): Promise<WarehouseListResponse> {
  try {
    const q = cityId ? `?cityId=${encodeURIComponent(cityId)}` : "";
    return await apiClient<WarehouseListResponse>(`/admin/warehouses${q}`);
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await delay();
    const items = cityId
      ? MOCK_LIST.filter((w) => w.cityId === cityId)
      : MOCK_LIST;
    return {
      items,
      total: items.length,
      generatedAt: new Date().toISOString(),
    };
  }
}

export async function fetchWarehouseDetail(id: string): Promise<WarehouseDetail> {
  try {
    return await apiClient<WarehouseDetail>(`/admin/warehouses/${id}`);
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await delay();
    return mockDetail(id);
  }
}
