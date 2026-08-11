"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { exportReport, fetchReports } from "./api";
import type { ExportFormat } from "./types";

export const reportKeys = {
  all: ["platform-reports"] as const,
  list: () => [...reportKeys.all, "list"] as const,
};

export function useReports() {
  return useQuery({
    queryKey: reportKeys.list(),
    queryFn: fetchReports,
  });
}

export function useExportReport() {
  return useMutation({
    mutationFn: (input: { reportId: string; format: ExportFormat }) =>
      exportReport(input),
  });
}
