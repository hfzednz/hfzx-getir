"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchLoyaltySnapshot } from "./api";

export const loyaltyKeys = {
  all: ["loyalty"] as const,
  snapshot: () => [...loyaltyKeys.all, "snapshot"] as const,
};

export function useLoyaltySnapshot() {
  return useQuery({
    queryKey: loyaltyKeys.snapshot(),
    queryFn: () => fetchLoyaltySnapshot(),
  });
}
