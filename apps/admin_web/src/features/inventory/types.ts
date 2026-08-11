import type { Id } from "@/shared/types/common";

export type InventoryRowKind =
  | "normal"
  | "reserved"
  | "damaged"
  | "expired"
  | "shrinkage";

export interface InventoryStockRow extends Record<string, unknown> {
  id: Id;
  sku: string;
  productName: string;
  warehouseCode: string;
  onHand: number;
  reserved: number;
  available: number;
  safetyStock: number;
  forecast7d: number;
  kind: InventoryRowKind;
  lastCountedAt: string;
}

export interface InventoryTransfer {
  id: Id;
  fromWarehouse: string;
  toWarehouse: string;
  skuCount: number;
  status: string;
  requestedAt: string;
}

export interface InventoryCycleCount {
  id: Id;
  warehouseCode: string;
  zone: string;
  varianceUnits: number;
  status: string;
  scheduledAt: string;
}

export interface InventoryAdjustment {
  id: Id;
  sku: string;
  warehouseCode: string;
  delta: number;
  reason: string;
  createdAt: string;
}

export interface InventorySnapshot {
  cityId: string | null;
  generatedAt: string;
  stock: InventoryStockRow[];
  transfers: InventoryTransfer[];
  cycleCounts: InventoryCycleCount[];
  adjustments: InventoryAdjustment[];
  totals: {
    skus: number;
    unitsOnHand: number;
    reserved: number;
    damaged: number;
    expired: number;
    belowSafety: number;
  };
}
