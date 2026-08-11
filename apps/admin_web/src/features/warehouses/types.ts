import type { Id } from "@/shared/types/common";

export type WarehouseStatus = "open" | "busy" | "closed" | "maintenance";

export interface WarehouseListItem extends Record<string, unknown> {
  id: Id;
  code: string;
  name: string;
  cityId: string;
  district: string;
  status: WarehouseStatus;
  capacityPct: number;
  skuCount: number;
  openOrders: number;
  pickSlaPct: number;
  stockAlerts: number;
}

export interface WarehouseKpis {
  capacityPct: number;
  skuCount: number;
  unitsOnHand: number;
  pickSlaPct: number;
  packSlaPct: number;
  dispatchSlaPct: number;
  avgPickMinutes: number;
  avgPackMinutes: number;
  openTransfers: number;
  stockAlerts: number;
}

export interface WarehouseTransfer {
  id: Id;
  direction: "in" | "out";
  counterpart: string;
  skuCount: number;
  status: string;
  etaAt: string;
}

export interface WarehouseAudit {
  id: Id;
  type: string;
  result: string;
  auditor: string;
  auditedAt: string;
}

export interface WarehouseStockAlert {
  id: Id;
  sku: string;
  productName: string;
  severity: "warning" | "danger";
  onHand: number;
  safetyStock: number;
  message: string;
}

export interface AiOptimizationItem {
  id: Id;
  title: string;
  summary: string;
  impact: string;
  confidence: number;
}

export interface WarehouseDetail {
  id: Id;
  code: string;
  name: string;
  cityId: string;
  district: string;
  address: string;
  status: WarehouseStatus;
  managerName: string;
  openedAt: string;
  kpis: WarehouseKpis;
  inventorySummary: { category: string; skuCount: number; units: number }[];
  transfers: WarehouseTransfer[];
  audits: WarehouseAudit[];
  stockAlerts: WarehouseStockAlert[];
  aiOptimization: AiOptimizationItem[];
}

export interface WarehouseListResponse {
  items: WarehouseListItem[];
  total: number;
  generatedAt: string;
}
