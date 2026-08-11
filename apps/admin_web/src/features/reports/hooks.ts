"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchReportsCatalog } from "./api";

export const reportsKeys = {
  all: ["reports"] as const,
  catalog: () => [...reportsKeys.all, "catalog"] as const,
};

export function useReportsCatalog() {
  return useQuery({
    queryKey: reportsKeys.catalog(),
    queryFn: () => fetchReportsCatalog(),
  });
}
