"use client";

import { useQuery } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import { fetchDashboardSnapshot } from "./api";

export const dashboardKeys = {
  all: ["dashboard"] as const,
  snapshot: (cityId: string | null) =>
    [...dashboardKeys.all, "snapshot", cityId] as const,
};

export function useDashboardSnapshot() {
  const cityId = useChromeStore((s) => s.cityId);

  return useQuery({
    queryKey: dashboardKeys.snapshot(cityId),
    queryFn: () => fetchDashboardSnapshot(cityId),
  });
}
