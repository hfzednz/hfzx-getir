"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchBilling, markInvoicePaid } from "./api";

export const billingKeys = {
  all: ["platform-billing"] as const,
  snapshot: () => [...billingKeys.all, "snapshot"] as const,
};

export function useBilling() {
  return useQuery({
    queryKey: billingKeys.snapshot(),
    queryFn: fetchBilling,
  });
}

export function useMarkInvoicePaid() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (invoiceId: string) => markInvoicePaid(invoiceId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: billingKeys.all });
    },
  });
}
