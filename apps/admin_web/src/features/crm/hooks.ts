"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import { useAuthStore } from "@/shared/auth/auth-store";
import { addCrmNote, fetchCrmWorkspace } from "./api";
import type { CrmListParams } from "./types";

export const crmKeys = {
  all: ["crm"] as const,
  workspace: (params: CrmListParams) =>
    [...crmKeys.all, "workspace", params] as const,
};

export function useCrmWorkspace(params: Omit<CrmListParams, "cityId"> = {}) {
  const cityId = useChromeStore((s) => s.cityId);
  const full: CrmListParams = { ...params, cityId };
  return useQuery({
    queryKey: crmKeys.workspace(full),
    queryFn: () => fetchCrmWorkspace(full),
  });
}

export function useAddCrmNote() {
  const qc = useQueryClient();
  const session = useAuthStore((s) => s.session);
  return useMutation({
    mutationFn: (input: { customerId: string; body: string }) =>
      addCrmNote(
        input.customerId,
        input.body,
        session?.displayName ?? session?.email ?? "ops",
      ),
    onSuccess: () => void qc.invalidateQueries({ queryKey: crmKeys.all }),
  });
}
