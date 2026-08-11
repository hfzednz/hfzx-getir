"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createCompany,
  deleteCompany,
  fetchCompanies,
  fetchCompany,
  updateCompany,
} from "./api";
import type { CreateCompanyInput, UpdateCompanyInput } from "./types";

export const companyKeys = {
  all: ["platform-companies"] as const,
  list: () => [...companyKeys.all, "list"] as const,
  detail: (id: string) => [...companyKeys.all, "detail", id] as const,
};

export function useCompanies() {
  return useQuery({
    queryKey: companyKeys.list(),
    queryFn: fetchCompanies,
  });
}

export function useCompany(id: string) {
  return useQuery({
    queryKey: companyKeys.detail(id),
    queryFn: () => fetchCompany(id),
    enabled: Boolean(id),
  });
}

export function useCreateCompany() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateCompanyInput) => createCompany(input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: companyKeys.all }),
  });
}

export function useUpdateCompany() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { id: string; patch: UpdateCompanyInput }) =>
      updateCompany(input.id, input.patch),
    onSuccess: (_d, vars) => {
      void qc.invalidateQueries({ queryKey: companyKeys.detail(vars.id) });
      void qc.invalidateQueries({ queryKey: companyKeys.list() });
    },
  });
}

export function useDeleteCompany() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteCompany(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: companyKeys.all }),
  });
}
