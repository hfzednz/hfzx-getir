import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient } from "@/shared/api/client";
import type { Paginated } from "@/shared/types/common";
import type {
  CustomerAdjustmentInput,
  CustomerAdjustmentResult,
  CustomerListFilters,
  CustomerListItem,
  CustomerProfile,
  CustomerSegment,
} from "./types";

const SEGMENTS: CustomerSegment[] = [
  "new",
  "loyal",
  "vip",
  "churn_risk",
  "high_aov",
  "fraud_watch",
];

function daysAgo(d: number): string {
  return new Date(Date.now() - d * 86_400_000).toISOString();
}

function hoursAgo(h: number): string {
  return new Date(Date.now() - h * 3_600_000).toISOString();
}

const MOCK_CUSTOMERS: CustomerListItem[] = [
  {
    id: "cus_1001",
    name: "Ayşe Kaya",
    email: "ayse.kaya@example.com",
    phone: "+90 532 111 2233",
    cityId: "city_ist",
    segment: "loyal",
    orderCount: 48,
    lifetimeValueMinor: 18_450_00,
    currency: "TRY",
    riskScore: 12,
    fraudScore: 8,
    loyaltyTier: "Gold",
    walletBalanceMinor: 125_00,
    createdAt: daysAgo(420),
    lastOrderAt: hoursAgo(0.4),
  },
  {
    id: "cus_1002",
    name: "Mehmet Yılmaz",
    email: "mehmet.y@example.com",
    phone: "+90 533 222 3344",
    cityId: "city_ist",
    segment: "high_aov",
    orderCount: 22,
    lifetimeValueMinor: 21_200_00,
    currency: "TRY",
    riskScore: 28,
    fraudScore: 35,
    loyaltyTier: "Silver",
    walletBalanceMinor: 0,
    createdAt: daysAgo(210),
    lastOrderAt: hoursAgo(0.5),
  },
  {
    id: "cus_1003",
    name: "Deniz Arslan",
    email: "deniz.a@example.com",
    phone: "+90 534 333 4455",
    cityId: "city_ist",
    segment: "vip",
    orderCount: 96,
    lifetimeValueMinor: 62_800_00,
    currency: "TRY",
    riskScore: 5,
    fraudScore: 3,
    loyaltyTier: "Platinum",
    walletBalanceMinor: 890_50,
    createdAt: daysAgo(780),
    lastOrderAt: hoursAgo(0.7),
  },
  {
    id: "cus_1004",
    name: "Caner Tekin",
    email: "caner.t@example.com",
    phone: "+90 535 444 5566",
    cityId: "city_ist",
    segment: "new",
    orderCount: 3,
    lifetimeValueMinor: 540_00,
    currency: "TRY",
    riskScore: 18,
    fraudScore: 12,
    loyaltyTier: "Bronze",
    walletBalanceMinor: 40_00,
    createdAt: daysAgo(14),
    lastOrderAt: hoursAgo(0.9),
  },
  {
    id: "cus_1005",
    name: "Selin Mutlu",
    email: "selin.m@example.com",
    phone: "+90 536 555 6677",
    cityId: "city_ist",
    segment: "fraud_watch",
    orderCount: 11,
    lifetimeValueMinor: 4_120_00,
    currency: "TRY",
    riskScore: 72,
    fraudScore: 81,
    loyaltyTier: "Bronze",
    walletBalanceMinor: 0,
    createdAt: daysAgo(90),
    lastOrderAt: hoursAgo(1.2),
  },
  {
    id: "cus_1006",
    name: "Burak Demir",
    email: "burak.d@example.com",
    phone: "+90 537 666 7788",
    cityId: "city_ist",
    segment: "churn_risk",
    orderCount: 17,
    lifetimeValueMinor: 3_880_00,
    currency: "TRY",
    riskScore: 22,
    fraudScore: 10,
    loyaltyTier: "Silver",
    walletBalanceMinor: 15_00,
    createdAt: daysAgo(300),
    lastOrderAt: daysAgo(45),
  },
];

function filterCustomers(filters: CustomerListFilters): CustomerListItem[] {
  const q = filters.q?.trim().toLowerCase() ?? "";
  const segment =
    filters.segment && filters.segment !== "all" ? filters.segment : null;
  return MOCK_CUSTOMERS.filter((c) => {
    if (filters.cityId && c.cityId !== filters.cityId) return false;
    if (segment && c.segment !== segment) return false;
    if (!q) return true;
    return (
      c.id.includes(q) ||
      c.name.toLowerCase().includes(q) ||
      c.email.toLowerCase().includes(q) ||
      c.phone.includes(q)
    );
  });
}

export function buildMockCustomerList(
  filters: CustomerListFilters,
): Paginated<CustomerListItem> {
  const page = filters.page ?? 1;
  const pageSize = filters.pageSize ?? 25;
  const filtered = filterCustomers(filters);
  const start = (page - 1) * pageSize;
  const items = filtered.slice(start, start + pageSize);
  return {
    items,
    page,
    pageSize,
    total: filtered.length,
    hasMore: start + pageSize < filtered.length,
  };
}

export function buildMockCustomerProfile(
  customerId: string,
): CustomerProfile | null {
  const base = MOCK_CUSTOMERS.find((c) => c.id === customerId);
  if (!base) return null;

  return {
    ...base,
    addresses: [
      {
        id: `${customerId}_addr1`,
        label: "Home",
        line1: "NEXORA Cad. No:12 D:4",
        district: "Kadıköy",
        city: "İstanbul",
        isDefault: true,
      },
      {
        id: `${customerId}_addr2`,
        label: "Work",
        line1: "Levent Plaza Kat:8",
        district: "Beşiktaş",
        city: "İstanbul",
        isDefault: false,
      },
    ],
    recentOrders: [
      {
        id: "ord_88421",
        status: "en_route",
        totalMinor: 312_50,
        currency: "TRY",
        createdAt: hoursAgo(0.4),
      },
      {
        id: "ord_88201",
        status: "delivered",
        totalMinor: 210_00,
        currency: "TRY",
        createdAt: hoursAgo(3),
      },
      {
        id: "ord_87011",
        status: "delivered",
        totalMinor: 156_80,
        currency: "TRY",
        createdAt: daysAgo(2),
      },
    ],
    walletTxns: [
      {
        id: `${customerId}_w1`,
        type: "credit",
        amountMinor: 50_00,
        currency: "TRY",
        note: "Campaign cashback",
        at: daysAgo(3),
      },
      {
        id: `${customerId}_w2`,
        type: "debit",
        amountMinor: 25_00,
        currency: "TRY",
        note: "Applied at checkout",
        at: daysAgo(1),
      },
    ],
    loyalty: {
      tier: base.loyaltyTier,
      points: base.segment === "vip" ? 12400 : 2400,
      pointsToNextTier: base.segment === "vip" ? 0 : 600,
    },
    coupons: [
      {
        id: `${customerId}_cp1`,
        code: "WELCOME40",
        status: base.segment === "new" ? "active" : "used",
        discountLabel: "40 TL off",
        expiresAt: daysAgo(-20),
      },
      {
        id: `${customerId}_cp2`,
        code: "FLASH15",
        status: "active",
        discountLabel: "15% up to 100 TL",
        expiresAt: daysAgo(-5),
      },
    ],
    supportHistory: [
      {
        id: "tkt_4412",
        subject: "Missing item in order",
        status: "resolved",
        createdAt: daysAgo(12),
      },
      {
        id: "tkt_4588",
        subject: "Late delivery complaint",
        status: base.segment === "fraud_watch" ? "open" : "pending",
        createdAt: daysAgo(2),
      },
    ],
    notes: [
      {
        id: `${customerId}_n1`,
        body: "Prefers contact via SMS. Door code 4488.",
        author: "support_agent",
        createdAt: daysAgo(30),
      },
    ],
  };
}

export async function fetchCustomers(
  filters: CustomerListFilters,
): Promise<Paginated<CustomerListItem>> {
  try {
    const params = new URLSearchParams();
    if (filters.q) params.set("q", filters.q);
    if (filters.segment && filters.segment !== "all")
      params.set("segment", filters.segment);
    if (filters.page) params.set("page", String(filters.page));
    if (filters.pageSize) params.set("pageSize", String(filters.pageSize));
    if (filters.cityId) params.set("cityId", filters.cityId);
    const qs = params.toString();
    return await apiClient<Paginated<CustomerListItem>>(
      `/admin/customers${qs ? `?${qs}` : ""}`,
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await new Promise((r) => setTimeout(r, 200));
    return buildMockCustomerList(filters);
  }
}

export async function fetchCustomerProfile(
  customerId: string,
): Promise<CustomerProfile> {
  try {
    return await apiClient<CustomerProfile>(`/admin/customers/${customerId}`);
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await new Promise((r) => setTimeout(r, 180));
    const mock = buildMockCustomerProfile(customerId);
    if (!mock) throw new Error(`Customer ${customerId} not found`);
    return mock;
  }
}

export async function applyCustomerAdjustment(
  input: CustomerAdjustmentInput,
): Promise<CustomerAdjustmentResult> {
  try {
    return await apiClient<CustomerAdjustmentResult>(
      `/admin/customers/${input.customerId}/adjustments`,
      {
        method: "POST",
        body: input,
        idempotent: true,
      },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await new Promise((r) => setTimeout(r, 220));
    return {
      customerId: input.customerId,
      ok: true,
      message: `Mock ${input.type} adjustment applied (BFF unavailable)`,
    };
  }
}

export const CUSTOMER_SEGMENT_OPTIONS: Array<CustomerSegment | "all"> = [
  "all",
  ...SEGMENTS,
];
