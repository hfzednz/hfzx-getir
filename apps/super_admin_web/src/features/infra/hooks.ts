"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchInfraSnapshot } from "./api";

export const infraKeys = {
  all: ["platform-infra"] as const,
  snapshot: () => [...infraKeys.all, "snapshot"] as const,
};

export function useInfraSnapshot() {
  return useQuery({
    queryKey: infraKeys.snapshot(),
    queryFn: fetchInfraSnapshot,
  });
}
