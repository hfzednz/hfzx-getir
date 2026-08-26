/** Canonical order-service states (order.go). */
export const ORDER_STATES = [
  "draft",
  "pending_payment",
  "payment_processing",
  "payment_failed",
  "inventory_reservation",
  "inventory_failed",
  "warehouse_assigned",
  "picking",
  "packing",
  "ready_for_dispatch",
  "courier_assigned",
  "out_for_delivery",
  "delivered",
  "completed",
  "cancelled",
  "refund_pending",
  "refunded",
  "failed",
  "archived",
] as const;

export type OrderState = (typeof ORDER_STATES)[number];

export function orderStateLabel(state: string): string {
  return state.replace(/_/g, " ");
}

export function isTerminalOrderState(state: string): boolean {
  return [
    "cancelled",
    "refunded",
    "failed",
    "archived",
    "completed",
    "payment_failed",
    "inventory_failed",
  ].includes(state);
}
