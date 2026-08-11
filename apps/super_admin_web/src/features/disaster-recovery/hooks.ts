"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchDisasterRecovery,
  proposeDrFailover,
  resolveDrFailover,
  runSimulation,
  startRestore,
} from "./api";

export const drKeys = {
  all: ["platform-dr"] as const,
  snapshot: () => [...drKeys.all, "snapshot"] as const,
};

export function useDisasterRecovery() {
  return useQuery({
    queryKey: drKeys.snapshot(),
    queryFn: fetchDisasterRecovery,
  });
}

export function useProposeDrFailover() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: proposeDrFailover,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: drKeys.all });
    },
  });
}

export function useResolveDrFailover() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: resolveDrFailover,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: drKeys.all });
    },
  });
}

export function useStartRestore() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: startRestore,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: drKeys.all });
    },
  });
}

export function useRunSimulation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: runSimulation,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: drKeys.all });
    },
  });
}
