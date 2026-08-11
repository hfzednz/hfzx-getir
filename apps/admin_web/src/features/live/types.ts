export type OpsSeverity = "info" | "warning" | "danger";

export type OrderStreamStatus =
  | "created"
  | "picking"
  | "ready"
  | "assigned"
  | "en_route"
  | "delivered"
  | "failed"
  | "cancelled";

export interface LiveOrderEvent {
  id: string;
  orderId: string;
  status: OrderStreamStatus;
  customerName: string;
  warehouseCode: string;
  zone: string;
  etaMinutes: number | null;
  delayMinutes: number;
  amountMinor: number;
  currency: string;
  updatedAt: string;
}

export interface CourierMarker {
  id: string;
  name: string;
  status: "available" | "busy" | "offline" | "break";
  lat: number;
  lng: number;
  zone: string;
  activeOrderId: string | null;
  batteryPct: number;
  updatedAt: string;
}

export interface WarehouseActivity {
  id: string;
  code: string;
  name: string;
  pickQueueMin: number;
  openOrders: number;
  capacityPct: number;
  status: "healthy" | "busy" | "critical";
}

export interface DelayIncident {
  id: string;
  orderId: string;
  zone: string;
  delayMinutes: number;
  reason: string;
  severity: OpsSeverity;
}

export interface FailedDelivery {
  id: string;
  orderId: string;
  courierName: string;
  reason: string;
  attempts: number;
  failedAt: string;
}

export interface Bottleneck {
  id: string;
  type: "warehouse" | "courier" | "zone" | "payment";
  title: string;
  detail: string;
  severity: OpsSeverity;
  impactScore: number;
}

export interface EmergencyEvent {
  id: string;
  title: string;
  detail: string;
  zone: string;
  raisedAt: string;
  acknowledged: boolean;
}

export interface OpsAlert {
  id: string;
  severity: OpsSeverity;
  title: string;
  detail: string;
  createdAt: string;
}

export interface LiveOpsSnapshot {
  cityId: string | null;
  generatedAt: string;
  connection: "live" | "polling" | "offline";
  orderStream: LiveOrderEvent[];
  couriers: CourierMarker[];
  warehouses: WarehouseActivity[];
  delays: DelayIncident[];
  failedDeliveries: FailedDelivery[];
  bottlenecks: Bottleneck[];
  emergencies: EmergencyEvent[];
  alerts: OpsAlert[];
  counts: {
    activeOrders: number;
    delayedOrders: number;
    availableCouriers: number;
    openEmergencies: number;
  };
}
