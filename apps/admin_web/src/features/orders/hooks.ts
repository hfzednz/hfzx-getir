"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import {
  bulkOrderAction,
  fetchOrderDetail,
  fetchOrders,
  performOrderAction,
} from "./api";
import type { OrderActionInput, OrderListFilters } from "./types";

export const orderKeys = {
  all: ["orders"] as const,
  list: (filters: OrderListFilters) =>
    [...orderKeys.all, "list", filters] as const,
  detail: (id: string) => [...orderKeys.all, "detail", id] as const,
};

export function useOrdersList(filters: Omit<OrderListFilters, "cityId">) {
  const cityId = useChromeStore((s) => s.cityId);
  const full: OrderListFilters = { ...filters, cityId };

  return useQuery({
    queryKey: orderKeys.list(full),
    queryFn: () => fetchOrders(full),
  });
}

export function useOrderDetail(orderId: string) {
  return useQuery({
    queryKey: orderKeys.detail(orderId),
    queryFn: () => fetchOrderDetail(orderId),
    enabled: Boolean(orderId),
  });
}

export function useOrderAction(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: Omit<OrderActionInput, "orderId">) =>
      performOrderAction({ ...input, orderId }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: orderKeys.detail(orderId) });
      void queryClient.invalidateQueries({ queryKey: orderKeys.all });
    },
  });
}

export function useBulkOrderAction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: {
      orderIds: string[];
      action: "cancel" | "reassign";
      reason?: string;
      courierId?: string;
    }) =>
      bulkOrderAction(
        input.orderIds,
        input.action,
        input.reason,
        input.courierId,
      ),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: orderKeys.all });
    },
  });
}
