"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchDeployments,
  promoteCanary,
  proposeSecretRotate,
  resolveSecretRotate,
  rollbackDeployment,
} from "./api";

export const deploymentKeys = {
  all: ["platform-deployments"] as const,
  snapshot: () => [...deploymentKeys.all, "snapshot"] as const,
};

export function useDeployments() {
  return useQuery({
    queryKey: deploymentKeys.snapshot(),
    queryFn: fetchDeployments,
  });
}

export function usePromoteCanary() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: promoteCanary,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: deploymentKeys.all });
    },
  });
}

export function useRollbackDeployment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: rollbackDeployment,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: deploymentKeys.all });
    },
  });
}

export function useProposeSecretRotate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: proposeSecretRotate,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: deploymentKeys.all });
    },
  });
}

export function useResolveSecretRotate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: resolveSecretRotate,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: deploymentKeys.all });
    },
  });
}
