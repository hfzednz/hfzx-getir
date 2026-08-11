"use client";

import { useQuery } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import { fetchInventorySnapshot } from "./api";

export const inventoryKeys = {
  all: ["inventory"] as const,
  snapshot: (cityId: string | null) =>
    [...inventoryKeys.all, "snapshot", cityId] as const,
};

export function useInventorySnapshot() {
  const cityId = useChromeStore((s) => s.cityId);
  return useQuery({
    queryKey: inventoryKeys.snapshot(cityId),
    queryFn: () => fetchInventorySnapshot(cityId),
    refetchInterval: 30_000,
  });
}
