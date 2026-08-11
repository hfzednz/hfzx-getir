"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchRbacSnapshot } from "./api";

export const rbacKeys = {
  all: ["rbac"] as const,
  snapshot: () => [...rbacKeys.all, "snapshot"] as const,
};

export function useRbacSnapshot() {
  return useQuery({
    queryKey: rbacKeys.snapshot(),
    queryFn: () => fetchRbacSnapshot(),
  });
}
