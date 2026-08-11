import type { Id } from "@/shared/types/common";

export type DeliveryMode = "express" | "scheduled" | "batch";
export type DispatchQueueStatus = "queued" | "assigning" | "dispatched" | "delayed";

export interface DeliveryZoneListItem extends Record<string, unknown> {
  id: Id;
  code: string;
  name: string;
  cityId: string;
  polygonPoints: number;
  courierAllocated: number;
  courierTarget: number;
  activeOrders: number;
  slaPct: number;
  avgEtaMinutes: number;
  status: "active" | "paused" | "draft";
}

export interface DispatchQueueItem extends Record<string, unknown> {
  id: Id;
  orderId: string;
  zoneName: string;
  mode: DeliveryMode;
  status: DispatchQueueStatus;
  etaMinutes: number;
  courierCode: string | null;
  waitingMinutes: number;
}

export interface CourierAllocation {
  id: Id;
  zoneId: string;
  zoneName: string;
  courierCode: string;
  courierName: string;
  liveStatus: string;
  capacity: number;
}

export interface EtaMonitorPoint {
  zoneName: string;
  p50: number;
  p90: number;
  breached: number;
}

export interface DeliverySnapshot {
  cityId: string | null;
  generatedAt: string;
  zones: DeliveryZoneListItem[];
  queues: DispatchQueueItem[];
  allocations: CourierAllocation[];
  etaMonitor: EtaMonitorPoint[];
  modeBreakdown: { mode: DeliveryMode; count: number }[];
}

export interface DeliveryZoneDetail {
  id: Id;
  code: string;
  name: string;
  cityId: string;
  status: "active" | "paused" | "draft";
  hexCount: number;
  courierAllocated: number;
  courierTarget: number;
  peakWindows: string[];
  notes: string;
}
