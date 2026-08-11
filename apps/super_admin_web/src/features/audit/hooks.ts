"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchAudit } from "./api";

export const auditKeys = {
  all: ["platform-audit"] as const,
  list: (q: string) => [...auditKeys.all, "list", q] as const,
};

export function useAudit(q = "") {
  return useQuery({
    queryKey: auditKeys.list(q),
    queryFn: () => fetchAudit({ q: q || undefined }),
  });
}
