import type { FinanceSnapshot } from "./types";
import { ApiError } from "@/shared/api/client";

const delay = (ms = 180) => new Promise((r) => setTimeout(r, ms));

let mockSnapshot: FinanceSnapshot = {
  kpis: [
    {
      id: "gmv",
      label: "GMV (7d)",
      valueMinor: 32_400_000_00,
      currency: "TRY",
      deltaPct: 6.2,
    },
    {
      id: "net",
      label: "Net revenue",
      valueMinor: 4_180_000_00,
      currency: "TRY",
      deltaPct: 4.1,
    },
    {
      id: "refunds",
      label: "Refunds",
      valueMinor: 286_000_00,
      currency: "TRY",
      deltaPct: 1.8,
    },
    {
      id: "payouts",
      label: "Pending payouts",
      valueMinor: 1_120_000_00,
      currency: "TRY",
      deltaPct: -2.4,
    },
  ],
  revenue: [
    {
      id: "rev_1",
      period: "2026-08-01",
      gmvMinor: 4_200_000_00,
      netRevenueMinor: 540_000_00,
      currency: "TRY",
    },
    {
      id: "rev_2",
      period: "2026-08-02",
      gmvMinor: 4_550_000_00,
      netRevenueMinor: 590_000_00,
      currency: "TRY",
    },
    {
      id: "rev_3",
      period: "2026-08-03",
      gmvMinor: 4_810_000_00,
      netRevenueMinor: 610_000_00,
      currency: "TRY",
    },
    {
      id: "rev_4",
      period: "2026-08-04",
      gmvMinor: 5_020_000_00,
      netRevenueMinor: 640_000_00,
      currency: "TRY",
    },
    {
      id: "rev_5",
      period: "2026-08-05",
      gmvMinor: 4_875_000_00,
      netRevenueMinor: 620_000_00,
      currency: "TRY",
    },
  ],
  refunds: [
    {
      id: "rfd_1",
      orderId: "ord_9921",
      amountMinor: 6200,
      currency: "TRY",
      reason: "Missing items",
      status: "pending",
      at: "2026-08-06T09:30:00Z",
    },
    {
      id: "rfd_2",
      orderId: "ord_7755",
      amountMinor: 4500,
      currency: "TRY",
      reason: "Damaged product",
      status: "pending",
      at: "2026-08-06T07:40:00Z",
    },
    {
      id: "rfd_3",
      orderId: "ord_6610",
      amountMinor: 18900,
      currency: "TRY",
      reason: "Cancelled after pay",
      status: "paid",
      at: "2026-08-05T16:00:00Z",
    },
  ],
  taxes: [
    {
      id: "tax_1",
      jurisdiction: "TR-VAT-standard",
      ratePct: 20,
      collectedMinor: 640_000_00,
      currency: "TRY",
    },
    {
      id: "tax_2",
      jurisdiction: "TR-VAT-reduced",
      ratePct: 10,
      collectedMinor: 88_000_00,
      currency: "TRY",
    },
  ],
  invoices: [
    {
      id: "inv_1",
      number: "INV-2026-8841",
      counterparty: "Supplier FreshCo",
      amountMinor: 420_000_00,
      currency: "TRY",
      status: "sent",
      dueAt: "2026-08-15T00:00:00Z",
    },
    {
      id: "inv_2",
      number: "INV-2026-8842",
      counterparty: "City logistics lease",
      amountMinor: 180_000_00,
      currency: "TRY",
      status: "overdue",
      dueAt: "2026-08-01T00:00:00Z",
    },
  ],
  payments: [
    {
      id: "pay_1",
      method: "card",
      amountMinor: 37950,
      currency: "TRY",
      status: "captured",
      at: "2026-08-06T08:12:00Z",
    },
    {
      id: "pay_2",
      method: "wallet",
      amountMinor: 12900,
      currency: "TRY",
      status: "captured",
      at: "2026-08-06T08:18:00Z",
    },
    {
      id: "pay_3",
      method: "card",
      amountMinor: 8900,
      currency: "TRY",
      status: "failed",
      at: "2026-08-06T08:22:00Z",
    },
  ],
  payouts: [
    {
      id: "po_1",
      beneficiary: "Courier batch IST-A",
      amountMinor: 640_000_00,
      currency: "TRY",
      status: "pending",
      dualControl: true,
      at: "2026-08-06T06:00:00Z",
    },
    {
      id: "po_2",
      beneficiary: "Warehouse WH-07 ops",
      amountMinor: 95_000_00,
      currency: "TRY",
      status: "approved",
      dualControl: true,
      at: "2026-08-05T18:00:00Z",
    },
  ],
  courierSettlements: [
    {
      id: "cs_1",
      courierId: "cr_441",
      courierName: "Can Öztürk",
      deliveries: 62,
      amountMinor: 485000,
      currency: "TRY",
      status: "open",
      period: "2026-W31",
    },
    {
      id: "cs_2",
      courierId: "cr_512",
      courierName: "Zeynep Arslan",
      deliveries: 71,
      amountMinor: 552000,
      currency: "TRY",
      status: "settled",
      period: "2026-W31",
    },
  ],
  supplierPayments: [
    {
      id: "sp_1",
      supplier: "FreshCo",
      amountMinor: 420_000_00,
      currency: "TRY",
      status: "scheduled",
      dueAt: "2026-08-15T00:00:00Z",
    },
    {
      id: "sp_2",
      supplier: "SnackWorld",
      amountMinor: 210_000_00,
      currency: "TRY",
      status: "held",
      dueAt: "2026-08-12T00:00:00Z",
    },
  ],
  profit: {
    gmvMinor: 32_400_000_00,
    cogsMinor: 21_800_000_00,
    deliveryCostMinor: 3_900_000_00,
    promoCostMinor: 1_120_000_00,
    contributionMinor: 5_580_000_00,
    currency: "TRY",
  },
  reports: [
    {
      id: "rep_1",
      title: "Daily P&L export",
      href: "/reports?type=pnl",
      description: "City-scoped profit & loss",
    },
    {
      id: "rep_2",
      title: "Tax summary",
      href: "/reports?type=tax",
      description: "VAT collected by jurisdiction",
    },
    {
      id: "rep_3",
      title: "Courier settlement pack",
      href: "/reports?type=courier_settlement",
      description: "Weekly settlement CSV",
    },
  ],
};

/** Mock finance — replaced by GET /admin/finance/* when BFF is live. */
export async function fetchFinanceSnapshot(): Promise<FinanceSnapshot> {
  await delay();
  return structuredClone(mockSnapshot);
}

export async function approvePayout(id: string): Promise<FinanceSnapshot> {
  await delay(240);
  const po = mockSnapshot.payouts.find((p) => p.id === id);
  if (!po) {
    throw new ApiError(404, {
      code: "not_found",
      message: "Payout not found",
      traceId: "mock",
    });
  }
  mockSnapshot = {
    ...mockSnapshot,
    payouts: mockSnapshot.payouts.map((p) =>
      p.id === id ? { ...p, status: "approved" as const } : p,
    ),
  };
  return structuredClone(mockSnapshot);
}

export async function settleCourier(id: string): Promise<FinanceSnapshot> {
  await delay(240);
  mockSnapshot = {
    ...mockSnapshot,
    courierSettlements: mockSnapshot.courierSettlements.map((c) =>
      c.id === id ? { ...c, status: "settled" as const } : c,
    ),
  };
  return structuredClone(mockSnapshot);
}

export async function approveFinanceRefund(
  id: string,
): Promise<FinanceSnapshot> {
  await delay(240);
  mockSnapshot = {
    ...mockSnapshot,
    refunds: mockSnapshot.refunds.map((r) =>
      r.id === id ? { ...r, status: "approved" as const } : r,
    ),
  };
  return structuredClone(mockSnapshot);
}
