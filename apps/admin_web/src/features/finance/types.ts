import type { Id } from "@/shared/types/common";

export interface FinanceKpi {
  id: string;
  label: string;
  valueMinor: number;
  currency: string;
  deltaPct: number;
}

export interface RevenueRow {
  id: Id;
  period: string;
  gmvMinor: number;
  netRevenueMinor: number;
  currency: string;
}

export interface RefundRow {
  id: Id;
  orderId: string;
  amountMinor: number;
  currency: string;
  reason: string;
  status: "pending" | "approved" | "paid";
  at: string;
}

export interface TaxRow {
  id: Id;
  jurisdiction: string;
  ratePct: number;
  collectedMinor: number;
  currency: string;
}

export interface InvoiceRow {
  id: Id;
  number: string;
  counterparty: string;
  amountMinor: number;
  currency: string;
  status: "draft" | "sent" | "paid" | "overdue";
  dueAt: string;
}

export interface PaymentRow {
  id: Id;
  method: string;
  amountMinor: number;
  currency: string;
  status: "captured" | "failed" | "refunded";
  at: string;
}

export interface PayoutRow {
  id: Id;
  beneficiary: string;
  amountMinor: number;
  currency: string;
  status: "pending" | "approved" | "paid" | "rejected";
  dualControl: boolean;
  at: string;
}

export interface CourierSettlement {
  id: Id;
  courierId: string;
  courierName: string;
  deliveries: number;
  amountMinor: number;
  currency: string;
  status: "open" | "settled";
  period: string;
}

export interface SupplierPayment {
  id: Id;
  supplier: string;
  amountMinor: number;
  currency: string;
  status: "scheduled" | "paid" | "held";
  dueAt: string;
}

export interface ProfitBreakdown {
  gmvMinor: number;
  cogsMinor: number;
  deliveryCostMinor: number;
  promoCostMinor: number;
  contributionMinor: number;
  currency: string;
}

export interface FinanceReportLink {
  id: Id;
  title: string;
  href: string;
  description: string;
}

export interface FinanceSnapshot {
  kpis: FinanceKpi[];
  revenue: RevenueRow[];
  refunds: RefundRow[];
  taxes: TaxRow[];
  invoices: InvoiceRow[];
  payments: PaymentRow[];
  payouts: PayoutRow[];
  courierSettlements: CourierSettlement[];
  supplierPayments: SupplierPayment[];
  profit: ProfitBreakdown;
  reports: FinanceReportLink[];
}
