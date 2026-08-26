import { apiClient } from "@/shared/api/client";
import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import type { CourierDetail, CourierListItem, CourierListResponse } from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const MOCK_LIST: CourierListItem[] = [
  {
    id: "cr_1001",
    code: "CR-1001",
    fullName: "Ahmet Yılmaz",
    phone: "+90 532 111 2233",
    cityId: "city_ist",
    zoneName: "Kadıköy",
    liveStatus: "busy",
    vehicleType: "Motorcycle",
    rating: 4.8,
    ratingCount: 1240,
    activeAssignments: 2,
    onTimePct: 96.2,
    emergency: false,
    lastSeenAt: new Date(Date.now() - 45_000).toISOString(),
  },
  {
    id: "cr_1002",
    code: "CR-1002",
    fullName: "Elif Demir",
    phone: "+90 533 222 3344",
    cityId: "city_ist",
    zoneName: "Beşiktaş",
    liveStatus: "available",
    vehicleType: "E-bike",
    rating: 4.9,
    ratingCount: 880,
    activeAssignments: 0,
    onTimePct: 97.5,
    emergency: false,
    lastSeenAt: new Date(Date.now() - 20_000).toISOString(),
  },
  {
    id: "cr_1003",
    code: "CR-1003",
    fullName: "Can Öztürk",
    phone: "+90 534 333 4455",
    cityId: "city_ist",
    zoneName: "Şişli",
    liveStatus: "emergency",
    vehicleType: "Motorcycle",
    rating: 4.4,
    ratingCount: 610,
    activeAssignments: 1,
    onTimePct: 91.0,
    emergency: true,
    lastSeenAt: new Date(Date.now() - 3 * 60_000).toISOString(),
  },
  {
    id: "cr_1004",
    code: "CR-1004",
    fullName: "Merve Kaya",
    phone: "+90 535 444 5566",
    cityId: "city_ist",
    zoneName: "Üsküdar",
    liveStatus: "break",
    vehicleType: "Car",
    rating: 4.7,
    ratingCount: 420,
    activeAssignments: 0,
    onTimePct: 94.8,
    emergency: false,
    lastSeenAt: new Date(Date.now() - 8 * 60_000).toISOString(),
  },
  {
    id: "cr_1005",
    code: "CR-1005",
    fullName: "Burak Şahin",
    phone: "+90 536 555 6677",
    cityId: "city_ank",
    zoneName: "Çankaya",
    liveStatus: "offline",
    vehicleType: "Motorcycle",
    rating: 4.5,
    ratingCount: 300,
    activeAssignments: 0,
    onTimePct: 93.1,
    emergency: false,
    lastSeenAt: new Date(Date.now() - 2 * 60 * 60_000).toISOString(),
  },
];

function mockDetail(id: string): CourierDetail {
  const base =
    MOCK_LIST.find((c) => c.id === id) ??
    ({
      ...MOCK_LIST[0],
      id,
      code: id.toUpperCase(),
      fullName: "Unknown Courier",
    } satisfies CourierListItem);

  return {
    id: base.id,
    code: base.code,
    fullName: base.fullName,
    phone: base.phone,
    email: `${base.code.toLowerCase()}@couriers.nexora.local`,
    cityId: base.cityId,
    zoneName: base.zoneName,
    liveStatus: base.liveStatus,
    emergency: base.emergency,
    emergencyReason: base.emergency ? "SOS button · vehicle issue" : null,
    rating: base.rating,
    ratingCount: base.ratingCount,
    joinedAt: "2024-03-12T09:00:00.000Z",
    lastSeenAt: base.lastSeenAt,
    performance: {
      deliveriesToday: 28,
      deliveriesWeek: 142,
      onTimePct: base.onTimePct,
      avgDeliveryMinutes: 16.4,
      acceptanceRatePct: 98.2,
      cancelByCourierPct: 0.4,
    },
    vehicle: {
      plate: "34 NX 1201",
      type: base.vehicleType,
      model: "Honda PCX 125",
      color: "White",
      insuranceExpiresAt: "2026-11-30",
    },
    assignments: [
      {
        id: "asg_1",
        orderId: "ORD-88421",
        status: "picking_up",
        zoneName: base.zoneName,
        etaMinutes: 7,
        assignedAt: new Date(Date.now() - 6 * 60_000).toISOString(),
      },
      {
        id: "asg_2",
        orderId: "ORD-88455",
        status: "queued",
        zoneName: base.zoneName,
        etaMinutes: 22,
        assignedAt: new Date(Date.now() - 2 * 60_000).toISOString(),
      },
    ],
    schedule: [
      { id: "sch_1", day: "Mon", start: "10:00", end: "18:00", zoneName: base.zoneName },
      { id: "sch_2", day: "Tue", start: "10:00", end: "18:00", zoneName: base.zoneName },
      { id: "sch_3", day: "Wed", start: "12:00", end: "20:00", zoneName: "Beşiktaş" },
      { id: "sch_4", day: "Thu", start: "10:00", end: "18:00", zoneName: base.zoneName },
      { id: "sch_5", day: "Fri", start: "14:00", end: "22:00", zoneName: base.zoneName },
    ],
    ratings: [
      {
        id: "rt_1",
        score: 5,
        comment: "Very fast and polite",
        orderId: "ORD-88100",
        createdAt: new Date(Date.now() - 86_400_000).toISOString(),
      },
      {
        id: "rt_2",
        score: 4,
        comment: "Good, slight delay",
        orderId: "ORD-88012",
        createdAt: new Date(Date.now() - 2 * 86_400_000).toISOString(),
      },
    ],
    documents: [
      { id: "doc_1", type: "ID card", status: "valid", expiresAt: null },
      { id: "doc_2", type: "Driver license", status: "valid", expiresAt: "2028-05-01" },
      { id: "doc_3", type: "Criminal record", status: "expiring", expiresAt: "2026-09-15" },
      { id: "doc_4", type: "Health report", status: "valid", expiresAt: "2027-01-01" },
    ],
    payments: [
      {
        id: "pay_1",
        period: "2026-W31",
        baseAmount: 12_500_00,
        bonusAmount: 1_800_00,
        penaltyAmount: 150_00,
        netAmount: 14_150_00,
        currency: "TRY",
        status: "paid",
        paidAt: "2026-08-03T10:00:00.000Z",
      },
      {
        id: "pay_2",
        period: "2026-W32",
        baseAmount: 11_200_00,
        bonusAmount: 900_00,
        penaltyAmount: 0,
        netAmount: 12_100_00,
        currency: "TRY",
        status: "pending",
        paidAt: null,
      },
    ],
    bonuses: [
      {
        id: "bn_1",
        reason: "Peak hour completion streak",
        amount: 500_00,
        currency: "TRY",
        createdAt: new Date(Date.now() - 3 * 86_400_000).toISOString(),
      },
      {
        id: "bn_2",
        reason: "CSAT bonus",
        amount: 300_00,
        currency: "TRY",
        createdAt: new Date(Date.now() - 7 * 86_400_000).toISOString(),
      },
    ],
    penalties: [
      {
        id: "pn_1",
        reason: "Late pickup (>8 min)",
        amount: 150_00,
        currency: "TRY",
        createdAt: new Date(Date.now() - 5 * 86_400_000).toISOString(),
      },
    ],
  };
}

/** GET /admin/couriers — mock fallback when BFF is offline. */
export async function fetchCouriers(
  cityId: string | null,
): Promise<CourierListResponse> {
  try {
    const q = cityId ? `?cityId=${encodeURIComponent(cityId)}` : "";
    return await apiClient<CourierListResponse>(`/admin/couriers${q}`);
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await delay();
    const items = cityId
      ? MOCK_LIST.filter((c) => c.cityId === cityId || cityId === "city_ist")
      : MOCK_LIST;
    return {
      items: cityId === "city_ank" ? MOCK_LIST.filter((c) => c.cityId === "city_ank") : items,
      total: cityId === "city_ank" ? 1 : items.length,
      generatedAt: new Date().toISOString(),
    };
  }
}

/** GET /admin/couriers/:id */
export async function fetchCourierDetail(id: string): Promise<CourierDetail> {
  try {
    return await apiClient<CourierDetail>(`/admin/couriers/${id}`);
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await delay();
    return mockDetail(id);
  }
}
