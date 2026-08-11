"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchAuditSnapshot } from "./api";

export const auditKeys = {
  all: ["audit"] as const,
  snapshot: () => [...auditKeys.all, "snapshot"] as const,
};

export function useAuditSnapshot() {
  return useQuery({
    queryKey: auditKeys.snapshot(),
    queryFn: () => fetchAuditSnapshot(),
  });
}
