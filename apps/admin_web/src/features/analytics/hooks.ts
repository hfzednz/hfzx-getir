"use client";

import { useQuery } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import { fetchAnalyticsSnapshot } from "./api";

export const analyticsKeys = {
  all: ["analytics"] as const,
  snapshot: (cityId: string | null) =>
    [...analyticsKeys.all, "snapshot", cityId] as const,
};

export function useAnalyticsSnapshot() {
  const cityId = useChromeStore((s) => s.cityId);

  return useQuery({
    queryKey: analyticsKeys.snapshot(cityId),
    queryFn: () => fetchAnalyticsSnapshot(cityId),
  });
}
