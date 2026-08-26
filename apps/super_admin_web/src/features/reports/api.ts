import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  ExportFormat,
  ReportExportResult,
  ReportsSnapshot,
} from "./types";

function delay(ms = 200): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function mockSnapshot(): ReportsSnapshot {
  return {
    generatedAt: new Date().toISOString(),
    reports: [
      {
        id: "rpt_platform_health",
        name: "Platform health summary",
        category: "platform",
        description: "Global KPIs, incidents, and dual-control activity",
        schedule: "daily",
        lastRunAt: new Date(Date.now() - 6 * 3600_000).toISOString(),
        owner: "platform_owner",
      },
      {
        id: "rpt_infra_capacity",
        name: "Infrastructure capacity",
        category: "infra",
        description: "Cluster, storage, CDN, and DR readiness",
        schedule: "weekly",
        lastRunAt: new Date(Date.now() - 2 * 86400_000).toISOString(),
        owner: "platform_sre",
      },
      {
        id: "rpt_finops_monthly",
        name: "FinOps monthly rollup",
        category: "financial",
        description: "License, compute, storage, and API billing",
        schedule: "monthly",
        lastRunAt: new Date(Date.now() - 10 * 86400_000).toISOString(),
        owner: "platform_finops",
      },
      {
        id: "rpt_compliance_dsar",
        name: "Compliance DSAR ledger",
        category: "compliance",
        description: "GDPR/KVKK/CCPA request throughput",
        schedule: "weekly",
        lastRunAt: new Date(Date.now() - 1 * 86400_000).toISOString(),
        owner: "platform_compliance",
      },
      {
        id: "rpt_security_threats",
        name: "Security threat overview",
        category: "security",
        description: "Login anomalies, sessions, policy changes",
        schedule: "daily",
        lastRunAt: new Date(Date.now() - 4 * 3600_000).toISOString(),
        owner: "platform_security",
      },
      {
        id: "rpt_ai_usage",
        name: "AI platform usage",
        category: "ai",
        description: "Inference spend, model versions, guardrail hits",
        schedule: "weekly",
        lastRunAt: new Date(Date.now() - 3 * 86400_000).toISOString(),
        owner: "platform_sre",
      },
      {
        id: "rpt_tenant_isolation",
        name: "Tenant isolation inventory",
        category: "tenant",
        description: "Shared/hybrid/separate modes and backup posture",
        schedule: null,
        lastRunAt: new Date(Date.now() - 12 * 3600_000).toISOString(),
        owner: "platform_owner",
      },
    ],
  };
}

function buildPreview(
  reportId: string,
  format: ExportFormat,
): { preview: string; mimeType: string; filename: string; byteLength: number } {
  const stamp = new Date().toISOString().slice(0, 10);
  if (format === "csv") {
    const preview = `report_id,generated_at,metric,value\n${reportId},${stamp},rows,128\n`;
    return {
      preview,
      mimeType: "text/csv",
      filename: `${reportId}-${stamp}.csv`,
      byteLength: preview.length,
    };
  }
  if (format === "json") {
    const preview = JSON.stringify(
      { reportId, generatedAt: stamp, metrics: { rows: 128 }, mock: true },
      null,
      2,
    );
    return {
      preview,
      mimeType: "application/json",
      filename: `${reportId}-${stamp}.json`,
      byteLength: preview.length,
    };
  }
  const preview = `%PDF-1.4 mock\nReport: ${reportId}\nGenerated: ${stamp}\n(NEXORA platform mock PDF)\n%%EOF`;
  return {
    preview,
    mimeType: "application/pdf",
    filename: `${reportId}-${stamp}.pdf`,
    byteLength: preview.length * 12,
  };
}

export async function fetchReports(): Promise<ReportsSnapshot> {
  try {
    return await apiClient<ReportsSnapshot>(platformPath("/reports"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockSnapshot();
    }
    throw err;
  }
}

export async function exportReport(input: {
  reportId: string;
  format: ExportFormat;
}): Promise<ReportExportResult> {
  try {
    return await apiClient<ReportExportResult>(
      platformPath(`/reports/${input.reportId}/export`),
      { method: "POST", body: { format: input.format }, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay(320);
      const built = buildPreview(input.reportId, input.format);
      return {
        reportId: input.reportId,
        format: input.format,
        ...built,
        exportedAt: new Date().toISOString(),
      };
    }
    throw err;
  }
}
