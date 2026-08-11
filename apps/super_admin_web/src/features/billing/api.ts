import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { BillingSnapshot, PlatformInvoice } from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const months = ["Mar", "Apr", "May", "Jun", "Jul", "Aug"];

function mockSnapshot(): BillingSnapshot {
  return {
    meters: [
      {
        id: "m_tenant",
        category: "tenant",
        label: "Tenant seats",
        usage: 9_420,
        unit: "seats",
        amountMinor: 420_000_00,
        currency: "USD",
        deltaPct: 4.2,
      },
      {
        id: "m_api",
        category: "api",
        label: "API calls",
        usage: 62_400_000,
        unit: "calls",
        amountMinor: 186_000_00,
        currency: "USD",
        deltaPct: 11.0,
      },
      {
        id: "m_storage",
        category: "storage",
        label: "Object storage",
        usage: 84_200,
        unit: "GB",
        amountMinor: 68_400_00,
        currency: "USD",
        deltaPct: 2.1,
      },
      {
        id: "m_compute",
        category: "compute",
        label: "Compute",
        usage: 1_240_000,
        unit: "vCPU-h",
        amountMinor: 312_000_00,
        currency: "USD",
        deltaPct: 6.8,
      },
      {
        id: "m_courier",
        category: "courier",
        label: "Courier platform fee",
        usage: 96_420,
        unit: "active",
        amountMinor: 145_000_00,
        currency: "USD",
        deltaPct: 1.4,
      },
      {
        id: "m_wh",
        category: "warehouse",
        label: "Warehouse nodes",
        usage: 1_842,
        unit: "sites",
        amountMinor: 98_200_00,
        currency: "USD",
        deltaPct: 0.9,
      },
    ],
    invoices: [
      {
        id: "inv_2026_08_acme",
        tenantId: "ten_acme",
        tenantName: "ACME Quick Commerce",
        periodStart: "2026-08-01",
        periodEnd: "2026-08-31",
        status: "open",
        totalMinor: 28_420_00,
        currency: "USD",
        dueAt: "2026-09-15",
        lines: [
          {
            id: "l1",
            category: "tenant",
            description: "Growth plan seats",
            quantity: 100,
            unitPriceMinor: 249_900,
            amountMinor: 24_990_00,
          },
          {
            id: "l2",
            category: "storage",
            description: "Overage storage",
            quantity: 180,
            unitPriceMinor: 19_00,
            amountMinor: 3_430_00,
          },
        ],
      },
      {
        id: "inv_2026_07_nova",
        tenantId: "ten_nova",
        tenantName: "Nova Market",
        periodStart: "2026-07-01",
        periodEnd: "2026-07-31",
        status: "paid",
        totalMinor: 182_000_00,
        currency: "USD",
        dueAt: "2026-08-15",
        lines: [
          {
            id: "l3",
            category: "tenant",
            description: "Enterprise contract",
            quantity: 1,
            unitPriceMinor: 182_000_00,
            amountMinor: 182_000_00,
          },
        ],
      },
      {
        id: "inv_2026_07_delta",
        tenantId: "ten_delta",
        tenantName: "Delta City Ops",
        periodStart: "2026-07-01",
        periodEnd: "2026-07-31",
        status: "overdue",
        totalMinor: 5_290_00,
        currency: "USD",
        dueAt: "2026-08-05",
        lines: [
          {
            id: "l4",
            category: "tenant",
            description: "Starter plan",
            quantity: 1,
            unitPriceMinor: 499_00,
            amountMinor: 499_00,
          },
          {
            id: "l5",
            category: "api",
            description: "API overage",
            quantity: 1,
            unitPriceMinor: 4_791_00,
            amountMinor: 4_791_00,
          },
        ],
      },
      {
        id: "inv_2026_08_orbit",
        tenantId: "ten_orbit",
        tenantName: "Orbit Enterprise",
        periodStart: "2026-08-01",
        periodEnd: "2026-08-31",
        status: "draft",
        totalMinor: 410_000_00,
        currency: "USD",
        dueAt: "2026-09-30",
        lines: [
          {
            id: "l6",
            category: "compute",
            description: "Dedicated compute",
            quantity: 1,
            unitPriceMinor: 280_000_00,
            amountMinor: 280_000_00,
          },
          {
            id: "l7",
            category: "warehouse",
            description: "Warehouse nodes",
            quantity: 80,
            unitPriceMinor: 1_625_00,
            amountMinor: 130_000_00,
          },
        ],
      },
    ],
    spendSeries: months.map((label, i) => ({
      label,
      value: 780_000_00 + i * 45_000_00 + (i % 2) * 20_000_00,
    })),
    forecastSeries: months.map((label, i) => ({
      label,
      value: 800_000_00 + i * 52_000_00,
    })),
    breakdown: [
      { category: "compute", amountMinor: 312_000_00, pct: 25.4 },
      { category: "tenant", amountMinor: 420_000_00, pct: 34.2 },
      { category: "api", amountMinor: 186_000_00, pct: 15.1 },
      { category: "courier", amountMinor: 145_000_00, pct: 11.8 },
      { category: "warehouse", amountMinor: 98_200_00, pct: 8.0 },
      { category: "storage", amountMinor: 68_400_00, pct: 5.5 },
    ],
    mtdSpendMinor: 1_229_600_00,
    forecastMinor: 1_380_000_00,
    currency: "USD",
    generatedAt: new Date().toISOString(),
  };
}

export async function fetchBilling(): Promise<BillingSnapshot> {
  try {
    return await apiClient<BillingSnapshot>(platformPath("/billing"));
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockSnapshot();
    }
    throw err;
  }
}

export async function markInvoicePaid(
  invoiceId: string,
): Promise<PlatformInvoice> {
  try {
    return await apiClient<PlatformInvoice>(
      platformPath(`/billing/invoices/${invoiceId}/pay`),
      { method: "POST", body: {}, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const snap = mockSnapshot();
      const inv = snap.invoices.find((i) => i.id === invoiceId);
      if (!inv) throw new Error("Invoice not found");
      return { ...inv, status: "paid" };
    }
    throw err;
  }
}
