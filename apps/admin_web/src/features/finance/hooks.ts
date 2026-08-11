"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  approveFinanceRefund,
  approvePayout,
  fetchFinanceSnapshot,
  settleCourier,
} from "./api";

export const financeKeys = {
  all: ["finance"] as const,
  snapshot: () => [...financeKeys.all, "snapshot"] as const,
};

export function useFinanceSnapshot() {
  return useQuery({
    queryKey: financeKeys.snapshot(),
    queryFn: () => fetchFinanceSnapshot(),
  });
}

export function useApprovePayout() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => approvePayout(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: financeKeys.all }),
  });
}

export function useSettleCourier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => settleCourier(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: financeKeys.all }),
  });
}

export function useApproveFinanceRefund() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => approveFinanceRefund(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: financeKeys.all }),
  });
}
