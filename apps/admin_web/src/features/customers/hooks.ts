"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import {
  applyCustomerAdjustment,
  fetchCustomerProfile,
  fetchCustomers,
} from "./api";
import type { CustomerAdjustmentInput, CustomerListFilters } from "./types";

export const customerKeys = {
  all: ["customers"] as const,
  list: (filters: CustomerListFilters) =>
    [...customerKeys.all, "list", filters] as const,
  detail: (id: string) => [...customerKeys.all, "detail", id] as const,
};

export function useCustomersList(
  filters: Omit<CustomerListFilters, "cityId">,
) {
  const cityId = useChromeStore((s) => s.cityId);
  const full: CustomerListFilters = { ...filters, cityId };

  return useQuery({
    queryKey: customerKeys.list(full),
    queryFn: () => fetchCustomers(full),
  });
}

export function useCustomerProfile(customerId: string) {
  return useQuery({
    queryKey: customerKeys.detail(customerId),
    queryFn: () => fetchCustomerProfile(customerId),
    enabled: Boolean(customerId),
  });
}

export function useCustomerAdjustment(customerId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: Omit<CustomerAdjustmentInput, "customerId">) =>
      applyCustomerAdjustment({ ...input, customerId }),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: customerKeys.detail(customerId),
      });
      void queryClient.invalidateQueries({ queryKey: customerKeys.all });
    },
  });
}
