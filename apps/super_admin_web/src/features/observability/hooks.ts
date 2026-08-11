"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchObservabilitySnapshot } from "./api";

export const observabilityKeys = {
  all: ["platform-observability"] as const,
  snapshot: () => [...observabilityKeys.all, "snapshot"] as const,
};

export function useObservabilitySnapshot() {
  return useQuery({
    queryKey: observabilityKeys.snapshot(),
    queryFn: fetchObservabilitySnapshot,
  });
}
