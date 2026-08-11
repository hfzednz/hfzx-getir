"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchAiPlatformSnapshot } from "./api";

export const aiPlatformKeys = {
  all: ["platform-ai"] as const,
  snapshot: () => [...aiPlatformKeys.all, "snapshot"] as const,
};

export function useAiPlatformSnapshot() {
  return useQuery({
    queryKey: aiPlatformKeys.snapshot(),
    queryFn: fetchAiPlatformSnapshot,
  });
}
