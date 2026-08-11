export type ReportDomain =
  | "orders"
  | "customers"
  | "products"
  | "inventory"
  | "couriers"
  | "warehouses"
  | "finance"
  | "crm"
  | "campaigns"
  | "performance"
  | "taxes"
  | "operations";

export type ExportFormat = "csv" | "excel" | "json" | "pdf";

export interface ReportTemplate {
  id: string;
  domain: ReportDomain;
  name: string;
  description: string;
  columns: string[];
  sampleRows: Record<string, string | number>[];
}

export interface ReportsCatalog {
  generatedAt: string;
  templates: ReportTemplate[];
}
