"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  emergencyRollback,
  fetchFlags,
  proposeKillSwitch,
  resolveKillSwitchProposal,
  toggleFeatureFlag,
  updateFlagRollout,
  upsertFlag,
} from "./api";
import type { UpdateRolloutInput, UpsertFlagInput } from "./types";

export const flagKeys = {
  all: ["platform-flags"] as const,
  snapshot: () => [...flagKeys.all, "snapshot"] as const,
};

export function useFlags() {
  return useQuery({
    queryKey: flagKeys.snapshot(),
    queryFn: fetchFlags,
  });
}

export function useUpsertFlag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpsertFlagInput) => upsertFlag(input),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: flagKeys.all });
    },
  });
}

export function useUpdateFlagRollout() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateRolloutInput) => updateFlagRollout(input),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: flagKeys.all });
    },
  });
}

export function useToggleFeatureFlag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { flagId: string; enabled: boolean }) =>
      toggleFeatureFlag(input.flagId, input.enabled),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: flagKeys.all });
    },
  });
}

export function useEmergencyRollback() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (flagId: string) => emergencyRollback(flagId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: flagKeys.all });
    },
  });
}

export function useProposeKillSwitch() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: proposeKillSwitch,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: flagKeys.all });
    },
  });
}

export function useResolveKillSwitchProposal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: resolveKillSwitchProposal,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: flagKeys.all });
    },
  });
}
