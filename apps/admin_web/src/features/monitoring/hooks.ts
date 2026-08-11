"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchMonitoringSnapshot } from "./api";

export const monitoringKeys = {
  all: ["monitoring"] as const,
  snapshot: () => [...monitoringKeys.all, "snapshot"] as const,
};

export function useMonitoringSnapshot() {
  return useQuery({
    queryKey: monitoringKeys.snapshot(),
    queryFn: () => fetchMonitoringSnapshot(),
    refetchInterval: 15_000,
  });
}
