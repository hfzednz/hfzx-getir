"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchSystemSnapshot } from "./api";

export const systemKeys = {
  all: ["system"] as const,
  snapshot: () => [...systemKeys.all, "snapshot"] as const,
};

export function useSystemSnapshot() {
  return useQuery({
    queryKey: systemKeys.snapshot(),
    queryFn: () => fetchSystemSnapshot(),
  });
}
