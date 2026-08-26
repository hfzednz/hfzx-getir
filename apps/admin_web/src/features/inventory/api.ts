import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient } from "@/shared/api/client";
import type { InventorySnapshot, InventoryStockRow } from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const MOCK_STOCK: InventoryStockRow[] = [
  {
    id: "inv_1",
    sku: "SKU-MLK-012",
    productName: "Full fat milk 1L",
    warehouseCode: "WH-07",
    onHand: 52,
    reserved: 14,
    available: 38,
    safetyStock: 80,
    forecast7d: 210,
    kind: "normal",
    lastCountedAt: new Date(Date.now() - 86_400_000).toISOString(),
  },
  {
    id: "inv_2",
    sku: "SKU-ICE-044",
    productName: "Magnum Classic",
    warehouseCode: "WH-07",
    onHand: 8,
    reserved: 2,
    available: 6,
    safetyStock: 40,
    forecast7d: 95,
    kind: "normal",
    lastCountedAt: new Date(Date.now() - 2 * 86_400_000).toISOString(),
  },
  {
    id: "inv_3",
    sku: "SKU-SNK-220",
    productName: "Protein bar cocoa",
    warehouseCode: "WH-14",
    onHand: 0,
    reserved: 0,
    available: 0,
    safetyStock: 24,
    forecast7d: 40,
    kind: "normal",
    lastCountedAt: new Date(Date.now() - 3 * 86_400_000).toISOString(),
  },
  {
    id: "inv_4",
    sku: "SKU-WTR-003",
    productName: "Still water 0.5L",
    warehouseCode: "WH-14",
    onHand: 420,
    reserved: 36,
    available: 384,
    safetyStock: 100,
    forecast7d: 600,
    kind: "reserved",
    lastCountedAt: new Date(Date.now() - 12 * 60 * 60_000).toISOString(),
  },
  {
    id: "inv_5",
    sku: "SKU-YOG-018",
    productName: "Greek yogurt 500g",
    warehouseCode: "WH-07",
    onHand: 12,
    reserved: 0,
    available: 12,
    safetyStock: 30,
    forecast7d: 55,
    kind: "damaged",
    lastCountedAt: new Date(Date.now() - 4 * 60 * 60_000).toISOString(),
  },
  {
    id: "inv_6",
    sku: "SKU-BRD-009",
    productName: "Sourdough loaf",
    warehouseCode: "WH-21",
    onHand: 6,
    reserved: 0,
    available: 6,
    safetyStock: 20,
    forecast7d: 48,
    kind: "expired",
    lastCountedAt: new Date(Date.now() - 6 * 60 * 60_000).toISOString(),
  },
  {
    id: "inv_7",
    sku: "SKU-CHF-055",
    productName: "Cheddar block 200g",
    warehouseCode: "WH-21",
    onHand: 90,
    reserved: 4,
    available: 86,
    safetyStock: 40,
    forecast7d: 70,
    kind: "shrinkage",
    lastCountedAt: new Date(Date.now() - 5 * 86_400_000).toISOString(),
  },
];

function mockSnapshot(cityId: string | null): InventorySnapshot {
  const stock = MOCK_STOCK;
  return {
    cityId,
    generatedAt: new Date().toISOString(),
    stock,
    transfers: [
      {
        id: "it_1",
        fromWarehouse: "DC-IST-01",
        toWarehouse: "WH-07",
        skuCount: 120,
        status: "in_transit",
        requestedAt: new Date(Date.now() - 3 * 60 * 60_000).toISOString(),
      },
      {
        id: "it_2",
        fromWarehouse: "WH-14",
        toWarehouse: "WH-07",
        skuCount: 18,
        status: "pending_approval",
        requestedAt: new Date(Date.now() - 40 * 60_000).toISOString(),
      },
    ],
    cycleCounts: [
      {
        id: "cc_1",
        warehouseCode: "WH-07",
        zone: "Cold A",
        varianceUnits: -4,
        status: "open",
        scheduledAt: new Date(Date.now() + 2 * 60 * 60_000).toISOString(),
      },
      {
        id: "cc_2",
        warehouseCode: "WH-14",
        zone: "Ambient B",
        varianceUnits: 0,
        status: "completed",
        scheduledAt: new Date(Date.now() - 86_400_000).toISOString(),
      },
    ],
    adjustments: [
      {
        id: "adj_1",
        sku: "SKU-YOG-018",
        warehouseCode: "WH-07",
        delta: -12,
        reason: "Damaged cold chain",
        createdAt: new Date(Date.now() - 5 * 60 * 60_000).toISOString(),
      },
      {
        id: "adj_2",
        sku: "SKU-CHF-055",
        warehouseCode: "WH-21",
        delta: -3,
        reason: "Shrinkage write-off",
        createdAt: new Date(Date.now() - 26 * 60 * 60_000).toISOString(),
      },
    ],
    totals: {
      skus: stock.length,
      unitsOnHand: stock.reduce((s, r) => s + r.onHand, 0),
      reserved: stock.reduce((s, r) => s + r.reserved, 0),
      damaged: stock.filter((r) => r.kind === "damaged").length,
      expired: stock.filter((r) => r.kind === "expired").length,
      belowSafety: stock.filter((r) => r.onHand < r.safetyStock).length,
    },
  };
}

export async function fetchInventorySnapshot(
  cityId: string | null,
): Promise<InventorySnapshot> {
  try {
    const q = cityId ? `?cityId=${encodeURIComponent(cityId)}` : "";
    return await apiClient<InventorySnapshot>(`/admin/inventory${q}`);
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await delay();
    return mockSnapshot(cityId);
  }
}
