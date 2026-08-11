"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  advancePrivacyRequest,
  createPrivacyRequest,
  fetchCompliance,
  updateRetention,
} from "./api";
import type { PrivacyRequest, RetentionPolicy } from "./types";

export const complianceKeys = {
  all: ["platform-compliance"] as const,
  snapshot: () => [...complianceKeys.all, "snapshot"] as const,
};

export function useCompliance() {
  return useQuery({
    queryKey: complianceKeys.snapshot(),
    queryFn: fetchCompliance,
  });
}

export function useAdvancePrivacyRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: advancePrivacyRequest,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: complianceKeys.all });
    },
  });
}

export function useCreatePrivacyRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createPrivacyRequest,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: complianceKeys.all });
    },
  });
}

export function useUpdateRetention() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: {
      policyId: string;
      patch: Partial<Pick<RetentionPolicy, "retentionDays" | "autoDelete">>;
    }) => updateRetention(input.policyId, input.patch),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: complianceKeys.all });
    },
  });
}

export type { PrivacyRequest };
