"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchOrgSnapshot } from "./api";

export const orgKeys = {
  all: ["platform-org"] as const,
  snapshot: () => [...orgKeys.all, "snapshot"] as const,
};

export function useOrgSnapshot() {
  return useQuery({
    queryKey: orgKeys.snapshot(),
    queryFn: fetchOrgSnapshot,
  });
}
