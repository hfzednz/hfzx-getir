"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchRolesSnapshot } from "./api";

export const rolesKeys = {
  all: ["platform-roles"] as const,
  snapshot: () => [...rolesKeys.all, "snapshot"] as const,
};

export function useRolesSnapshot() {
  return useQuery({
    queryKey: rolesKeys.snapshot(),
    queryFn: fetchRolesSnapshot,
  });
}
