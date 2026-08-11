"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createTenant,
  fetchTenant,
  fetchTenants,
  proposeTenantAction,
  resolveTenantProposal,
  updateTenantIsolation,
} from "./api";
import type { CreateTenantInput, TenantListItem } from "./types";

export const tenantKeys = {
  all: ["platform-tenants"] as const,
  list: () => [...tenantKeys.all, "list"] as const,
  detail: (id: string) => [...tenantKeys.all, "detail", id] as const,
};

export function useTenants() {
  return useQuery({
    queryKey: tenantKeys.list(),
    queryFn: fetchTenants,
  });
}

export function useTenant(id: string) {
  return useQuery({
    queryKey: tenantKeys.detail(id),
    queryFn: () => fetchTenant(id),
    enabled: Boolean(id),
  });
}

export function useCreateTenant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateTenantInput) => createTenant(input),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: tenantKeys.all });
    },
  });
}

export function useUpdateTenantIsolation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: {
      id: string;
      isolationMode: TenantListItem["isolationMode"];
    }) => updateTenantIsolation(input.id, input.isolationMode),
    onSuccess: (_data, vars) => {
      void qc.invalidateQueries({ queryKey: tenantKeys.detail(vars.id) });
      void qc.invalidateQueries({ queryKey: tenantKeys.list() });
    },
  });
}

export function useProposeTenantAction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: proposeTenantAction,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: tenantKeys.all });
    },
  });
}

export function useResolveTenantProposal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: resolveTenantProposal,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: tenantKeys.all });
    },
  });
}
