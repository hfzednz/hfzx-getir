"use client";

import { useQuery } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import { fetchPlatformDashboard } from "./api";

export const dashboardKeys = {
  all: ["platform-dashboard"] as const,
  snapshot: (tenantContextId: string | null) =>
    [...dashboardKeys.all, "snapshot", tenantContextId] as const,
};

export function usePlatformDashboard() {
  const tenantContextId = useChromeStore((s) => s.tenantContextId);

  return useQuery({
    queryKey: dashboardKeys.snapshot(tenantContextId),
    queryFn: () => fetchPlatformDashboard(tenantContextId),
  });
}
