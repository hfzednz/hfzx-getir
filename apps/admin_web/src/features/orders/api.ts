import { apiClient } from "@/shared/api/client";
import type { Paginated } from "@/shared/types/common";
import type {
  OrderActionInput,
  OrderActionResult,
  OrderDetail,
  OrderListFilters,
  OrderListItem,
  OrderStatus,
} from "./types";

const STATUSES: OrderStatus[] = [
  "created",
  "confirmed",
  "picking",
  "ready",
  "assigned",
  "en_route",
  "delivered",
  "cancelled",
  "failed",
  "refunded",
];

export async function fetchOrders(
  filters: OrderListFilters,
): Promise<Paginated<OrderListItem>> {
  const params = new URLSearchParams();
  if (filters.q) params.set("q", filters.q);
  if (filters.status && filters.status !== "all")
    params.set("status", filters.status);
  if (filters.warehouseCode)
    params.set("warehouseCode", filters.warehouseCode);
  if (filters.zone) params.set("zone", filters.zone);
  if (filters.page) params.set("page", String(filters.page));
  if (filters.pageSize) params.set("pageSize", String(filters.pageSize));
  if (filters.cityId) params.set("cityId", filters.cityId);
  const qs = params.toString();
  return apiClient<Paginated<OrderListItem>>(
    `/admin/orders${qs ? `?${qs}` : ""}`,
  );
}

export async function fetchOrderDetail(
  orderId: string,
): Promise<OrderDetail> {
  return apiClient<OrderDetail>(`/admin/orders/${orderId}`);
}

export async function performOrderAction(
  input: OrderActionInput,
): Promise<OrderActionResult> {
  return apiClient<OrderActionResult>(
    `/admin/orders/${input.orderId}/actions`,
    {
      method: "POST",
      body: input,
      idempotent: true,
    },
  );
}

export async function bulkOrderAction(
  orderIds: string[],
  action: "cancel" | "reassign",
  reason?: string,
  courierId?: string,
): Promise<OrderActionResult[]> {
  if (action === "reassign" && !courierId?.trim()) {
    throw new Error("courierId is required for reassign");
  }
  const results: OrderActionResult[] = [];
  for (const orderId of orderIds) {
    results.push(
      await performOrderAction({
        orderId,
        action,
        reason,
        courierId: action === "reassign" ? courierId : undefined,
      }),
    );
  }
  return results;
}

export const ORDER_STATUS_OPTIONS: Array<OrderStatus | "all"> = [
  "all",
  ...STATUSES,
];
