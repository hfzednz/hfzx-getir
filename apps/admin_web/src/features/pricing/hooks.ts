"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import {
  createPricingRule,
  fetchPricingRules,
  updatePricingRuleStatus,
} from "./api";
import type {
  PricingListParams,
  PricingRuleStatus,
  PricingUpsertInput,
} from "./types";

export const pricingKeys = {
  all: ["pricing"] as const,
  list: (params: PricingListParams) =>
    [...pricingKeys.all, "list", params] as const,
};

export function usePricingRules(
  params: Omit<PricingListParams, "cityId"> = {},
) {
  const cityId = useChromeStore((s) => s.cityId);
  const full: PricingListParams = { ...params, cityId };
  return useQuery({
    queryKey: pricingKeys.list(full),
    queryFn: () => fetchPricingRules(full),
  });
}

export function useCreatePricingRule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: PricingUpsertInput) => createPricingRule(input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: pricingKeys.all }),
  });
}

export function useUpdatePricingStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { id: string; status: PricingRuleStatus }) =>
      updatePricingRuleStatus(input.id, input.status),
    onSuccess: () => void qc.invalidateQueries({ queryKey: pricingKeys.all }),
  });
}
