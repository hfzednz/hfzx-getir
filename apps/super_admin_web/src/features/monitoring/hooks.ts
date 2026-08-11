"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchMonitoring } from "./api";

export const monitoringKeys = {
  all: ["platform-monitoring"] as const,
  snapshot: () => [...monitoringKeys.all, "snapshot"] as const,
};

export function useMonitoring() {
  return useQuery({
    queryKey: monitoringKeys.snapshot(),
    queryFn: fetchMonitoring,
    refetchInterval: 30_000,
  });
}
