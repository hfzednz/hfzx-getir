"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchConfigSnapshot } from "./api";

export const configKeys = {
  all: ["platform-config"] as const,
  snapshot: () => [...configKeys.all, "snapshot"] as const,
};

export function useConfigSnapshot() {
  return useQuery({
    queryKey: configKeys.snapshot(),
    queryFn: fetchConfigSnapshot,
  });
}
