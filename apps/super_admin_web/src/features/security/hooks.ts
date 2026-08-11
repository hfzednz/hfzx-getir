"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  acknowledgeThreat,
  fetchSecurity,
  revokeSession,
  toggleGeoRule,
  togglePolicy,
  toggleProvider,
} from "./api";

export const securityKeys = {
  all: ["platform-security"] as const,
  snapshot: () => [...securityKeys.all, "snapshot"] as const,
};

export function useSecurity() {
  return useQuery({
    queryKey: securityKeys.snapshot(),
    queryFn: fetchSecurity,
  });
}

export function useAcknowledgeThreat() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (threatId: string) => acknowledgeThreat(threatId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: securityKeys.all });
    },
  });
}

export function useRevokeSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (sessionId: string) => revokeSession(sessionId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: securityKeys.all });
    },
  });
}

export function useToggleGeoRule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (ruleId: string) => toggleGeoRule(ruleId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: securityKeys.all });
    },
  });
}

export function useToggleProvider() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (providerId: string) => toggleProvider(providerId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: securityKeys.all });
    },
  });
}

export function useTogglePolicy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (policyId: string) => togglePolicy(policyId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: securityKeys.all });
    },
  });
}
