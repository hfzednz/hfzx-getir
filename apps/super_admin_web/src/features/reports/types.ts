import type { Id } from "@/shared/types/common";

export type ReportCategory =
  | "platform"
  | "infra"
  | "financial"
  | "compliance"
  | "security"
  | "ai"
  | "tenant";

export type ExportFormat = "csv" | "json" | "pdf";

export interface ReportDefinition {
  id: Id;
  name: string;
  category: ReportCategory;
  description: string;
  schedule: string | null;
  lastRunAt: string | null;
  owner: string;
}

export interface ReportExportResult {
  reportId: Id;
  format: ExportFormat;
  filename: string;
  mimeType: string;
  /** Mock payload preview (truncated). */
  preview: string;
  byteLength: number;
  exportedAt: string;
}

export interface ReportsSnapshot {
  generatedAt: string;
  reports: ReportDefinition[];
}
