export type OrderStatus =
  | "created"
  | "confirmed"
  | "picking"
  | "ready"
  | "assigned"
  | "en_route"
  | "delivered"
  | "cancelled"
  | "failed"
  | "refunded";

export type OrderPaymentStatus =
  | "pending"
  | "authorized"
  | "captured"
  | "refunded"
  | "failed";

export interface OrderListItem extends Record<string, unknown> {
  id: string;
  externalRef: string;
  status: OrderStatus;
  paymentStatus: OrderPaymentStatus;
  customerId: string;
  customerName: string;
  customerPhone: string;
  warehouseCode: string;
  courierId: string | null;
  courierName: string | null;
  zone: string;
  cityId: string;
  itemCount: number;
  totalMinor: number;
  currency: string;
  delayMinutes: number;
  createdAt: string;
  updatedAt: string;
}

export interface OrderLineItem {
  id: string;
  sku: string;
  name: string;
  qty: number;
  unitPriceMinor: number;
  totalMinor: number;
}

export interface OrderTimelineEvent {
  id: string;
  type: string;
  title: string;
  detail?: string;
  actor?: string;
  at: string;
}

export interface OrderDetail extends OrderListItem {
  address: string;
  notes: string | null;
  lines: OrderLineItem[];
  timeline: OrderTimelineEvent[];
  refundedMinor: number;
}

export interface OrderListFilters {
  q?: string;
  status?: OrderStatus | "all";
  warehouseCode?: string;
  zone?: string;
  page?: number;
  pageSize?: number;
  cityId?: string | null;
}

export type OrderActionType =
  | "reassign"
  | "cancel"
  | "refund"
  | "replace"
  | "force_complete"
  | "force_cancel";

export interface OrderActionInput {
  orderId: string;
  action: OrderActionType;
  reason?: string;
  courierId?: string;
  refundMinor?: number;
}

export interface OrderActionResult {
  orderId: string;
  action: OrderActionType;
  ok: boolean;
  message: string;
}
