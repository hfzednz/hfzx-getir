"use client";

import { useQuery } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import { fetchAiCommandSnapshot } from "./api";

export const aiKeys = {
  all: ["ai"] as const,
  snapshot: (cityId: string | null) =>
    [...aiKeys.all, "snapshot", cityId] as const,
};

export function useAiCommandSnapshot() {
  const cityId = useChromeStore((s) => s.cityId);

  return useQuery({
    queryKey: aiKeys.snapshot(cityId),
    queryFn: () => fetchAiCommandSnapshot(cityId),
  });
}
