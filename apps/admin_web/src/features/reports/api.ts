import type { ReportsCatalog } from "./types";

/** Mock report catalog — replaced by GET /admin/reports when BFF is live. */
export async function fetchReportsCatalog(): Promise<ReportsCatalog> {
  await new Promise((r) => setTimeout(r, 200));

  return {
    generatedAt: new Date().toISOString(),
    templates: [
      {
        id: "rep_orders",
        domain: "orders",
        name: "Orders summary",
        description: "Order volume, GMV, cancel & refund rates by day",
        columns: ["date", "orders", "gmv_try", "cancel_pct", "refund_pct"],
        sampleRows: [
          {
            date: "2026-08-05",
            orders: 12840,
            gmv_try: 4875400,
            cancel_pct: 2.1,
            refund_pct: 1.0,
          },
          {
            date: "2026-08-04",
            orders: 12110,
            gmv_try: 4622100,
            cancel_pct: 2.0,
            refund_pct: 1.1,
          },
        ],
      },
      {
        id: "rep_customers",
        domain: "customers",
        name: "Customer cohorts",
        description: "New vs returning, retention, CLV bands",
        columns: ["cohort", "new", "returning", "retention_d30", "avg_clv"],
        sampleRows: [
          {
            cohort: "2026-07",
            new: 42000,
            returning: 88000,
            retention_d30: 38.4,
            avg_clv: 1845,
          },
        ],
      },
      {
        id: "rep_products",
        domain: "products",
        name: "Product performance",
        description: "Top SKUs by units and revenue",
        columns: ["sku", "name", "units", "revenue_try"],
        sampleRows: [
          { sku: "SKU-1001", name: "Ayran 1L", units: 18400, revenue_try: 368000 },
        ],
      },
      {
        id: "rep_inventory",
        domain: "inventory",
        name: "Inventory snapshot",
        description: "On-hand, reserved, stockouts by warehouse",
        columns: ["warehouse", "on_hand", "reserved", "stockouts"],
        sampleRows: [
          { warehouse: "WH-Kadıköy", on_hand: 42000, reserved: 3100, stockouts: 12 },
        ],
      },
      {
        id: "rep_couriers",
        domain: "couriers",
        name: "Courier performance",
        description: "Deliveries, on-time %, ratings",
        columns: ["courier_id", "deliveries", "on_time_pct", "rating"],
        sampleRows: [
          { courier_id: "cr_441", deliveries: 62, on_time_pct: 97.2, rating: 4.9 },
        ],
      },
      {
        id: "rep_warehouses",
        domain: "warehouses",
        name: "Warehouse ops",
        description: "Pick time, SLA, capacity utilization",
        columns: ["warehouse", "pick_min", "sla_pct", "util_pct"],
        sampleRows: [
          { warehouse: "WH-Şişli", pick_min: 5.9, sla_pct: 97.2, util_pct: 81 },
        ],
      },
      {
        id: "rep_finance",
        domain: "finance",
        name: "Finance settlements",
        description: "GMV, fees, payouts, refunds",
        columns: ["date", "gmv", "fees", "payouts", "refunds"],
        sampleRows: [
          {
            date: "2026-08-05",
            gmv: 4875400,
            fees: 243770,
            payouts: 410000,
            refunds: 52000,
          },
        ],
      },
      {
        id: "rep_crm",
        domain: "crm",
        name: "CRM engagement",
        description: "Segments reached, open/click rates",
        columns: ["segment", "reached", "open_pct", "click_pct"],
        sampleRows: [
          { segment: "Power weekly", reached: 42100, open_pct: 42, click_pct: 8.1 },
        ],
      },
      {
        id: "rep_campaigns",
        domain: "campaigns",
        name: "Campaign ROI",
        description: "Spend, incremental orders, ROI",
        columns: ["campaign", "spend", "orders", "roi"],
        sampleRows: [
          { campaign: "Flash Friday", spend: 120000, orders: 8400, roi: 3.2 },
        ],
      },
      {
        id: "rep_perf",
        domain: "performance",
        name: "City SLA performance",
        description: "On-time, ETA, cancel by city",
        columns: ["city", "sla_pct", "eta_min", "cancel_pct"],
        sampleRows: [
          { city: "Istanbul", sla_pct: 94.2, eta_min: 18, cancel_pct: 2.1 },
        ],
      },
      {
        id: "rep_taxes",
        domain: "taxes",
        name: "Tax summary",
        description: "VAT collected by rate band",
        columns: ["rate", "taxable", "vat"],
        sampleRows: [{ rate: 0.2, taxable: 4062833, vat: 812567 }],
      },
      {
        id: "rep_ops",
        domain: "operations",
        name: "Operations incidents",
        description: "Incident counts by severity and zone",
        columns: ["zone", "danger", "warning", "info"],
        sampleRows: [{ zone: "Kadıköy", danger: 2, warning: 8, info: 14 }],
      },
    ],
  };
}

export function mockDownload(
  filename: string,
  content: string,
  mime: string,
): void {
  if (typeof window === "undefined") return;
  const blob = new Blob([content], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export function buildExportPayload(
  templateId: string,
  format: "csv" | "excel" | "json" | "pdf",
  rows: Record<string, string | number>[],
  columns: string[],
): { filename: string; content: string; mime: string } {
  const stamp = new Date().toISOString().slice(0, 10);
  if (format === "json") {
    return {
      filename: `${templateId}_${stamp}.json`,
      content: JSON.stringify({ templateId, rows }, null, 2),
      mime: "application/json",
    };
  }
  if (format === "csv" || format === "excel") {
    const header = columns.join(",");
    const body = rows
      .map((r) => columns.map((c) => JSON.stringify(r[c] ?? "")).join(","))
      .join("\n");
    const ext = format === "excel" ? "xls" : "csv";
    return {
      filename: `${templateId}_${stamp}.${ext}`,
      content: `${header}\n${body}`,
      mime:
        format === "excel"
          ? "application/vnd.ms-excel"
          : "text/csv;charset=utf-8",
    };
  }
  // pdf mock: plain text payload
  return {
    filename: `${templateId}_${stamp}.pdf.txt`,
    content: `NEXORA Report ${templateId}\n${columns.join(" | ")}\n${rows
      .map((r) => columns.map((c) => String(r[c] ?? "")).join(" | "))
      .join("\n")}`,
    mime: "text/plain",
  };
}
