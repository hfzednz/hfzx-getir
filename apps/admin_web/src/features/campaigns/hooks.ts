"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import {
  createCampaign,
  duplicateCampaign,
  fetchCampaign,
  fetchCampaigns,
  pauseCampaign,
  resumeCampaign,
  scheduleCampaign,
  updateCampaign,
} from "./api";
import type { CampaignListParams, CampaignUpsertInput } from "./types";

export const campaignKeys = {
  all: ["campaigns"] as const,
  list: (params: CampaignListParams) =>
    [...campaignKeys.all, "list", params] as const,
  detail: (id: string) => [...campaignKeys.all, "detail", id] as const,
};

export function useCampaigns(params: Omit<CampaignListParams, "cityId"> = {}) {
  const cityId = useChromeStore((s) => s.cityId);
  const full: CampaignListParams = { ...params, cityId };
  return useQuery({
    queryKey: campaignKeys.list(full),
    queryFn: () => fetchCampaigns(full),
  });
}

export function useCampaign(id: string) {
  return useQuery({
    queryKey: campaignKeys.detail(id),
    queryFn: () => fetchCampaign(id),
    enabled: Boolean(id),
  });
}

function invalidateCampaigns(
  qc: ReturnType<typeof useQueryClient>,
  id?: string,
) {
  void qc.invalidateQueries({ queryKey: campaignKeys.all });
  if (id) void qc.invalidateQueries({ queryKey: campaignKeys.detail(id) });
}

export function useCreateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CampaignUpsertInput) => createCampaign(input),
    onSuccess: () => invalidateCampaigns(qc),
  });
}

export function useUpdateCampaign(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: Partial<CampaignUpsertInput>) =>
      updateCampaign(id, input),
    onSuccess: () => invalidateCampaigns(qc, id),
  });
}

export function useDuplicateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => duplicateCampaign(id),
    onSuccess: () => invalidateCampaigns(qc),
  });
}

export function useScheduleCampaign(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { startsAt: string; endsAt: string | null }) =>
      scheduleCampaign(id, input.startsAt, input.endsAt),
    onSuccess: () => invalidateCampaigns(qc, id),
  });
}

export function usePauseCampaign(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => pauseCampaign(id),
    onSuccess: () => invalidateCampaigns(qc, id),
  });
}

export function useResumeCampaign(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => resumeCampaign(id),
    onSuccess: () => invalidateCampaigns(qc, id),
  });
}
