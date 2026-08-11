import { apiClient } from "@/shared/api/client";
import type { DeliverySnapshot, DeliveryZoneDetail, DeliveryZoneListItem } from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const MOCK_ZONES: DeliveryZoneListItem[] = [
  {
    id: "zn_kad",
    code: "IST-KAD",
    name: "Kadıköy",
    cityId: "city_ist",
    polygonPoints: 48,
    courierAllocated: 42,
    courierTarget: 50,
    activeOrders: 118,
    slaPct: 91.2,
    avgEtaMinutes: 19,
    status: "active",
  },
  {
    id: "zn_bes",
    code: "IST-BES",
    name: "Beşiktaş",
    cityId: "city_ist",
    polygonPoints: 36,
    courierAllocated: 28,
    courierTarget: 40,
    activeOrders: 86,
    slaPct: 88.4,
    avgEtaMinutes: 22,
    status: "active",
  },
  {
    id: "zn_sis",
    code: "IST-SIS",
    name: "Şişli",
    cityId: "city_ist",
    polygonPoints: 40,
    courierAllocated: 35,
    courierTarget: 35,
    activeOrders: 64,
    slaPct: 94.8,
    avgEtaMinutes: 16,
    status: "active",
  },
  {
    id: "zn_usk",
    code: "IST-USK",
    name: "Üsküdar",
    cityId: "city_ist",
    polygonPoints: 32,
    courierAllocated: 18,
    courierTarget: 28,
    activeOrders: 41,
    slaPct: 92.0,
    avgEtaMinutes: 18,
    status: "paused",
  },
  {
    id: "zn_can",
    code: "ANK-CAN",
    name: "Çankaya",
    cityId: "city_ank",
    polygonPoints: 44,
    courierAllocated: 22,
    courierTarget: 30,
    activeOrders: 55,
    slaPct: 93.5,
    avgEtaMinutes: 17,
    status: "active",
  },
];

function mockSnapshot(cityId: string | null): DeliverySnapshot {
  const zones = cityId
    ? MOCK_ZONES.filter((z) => z.cityId === cityId)
    : MOCK_ZONES;

  return {
    cityId,
    generatedAt: new Date().toISOString(),
    zones,
    queues: [
      {
        id: "dq_1",
        orderId: "ORD-88421",
        zoneName: "Kadıköy",
        mode: "express",
        status: "assigning",
        etaMinutes: 12,
        courierCode: null,
        waitingMinutes: 3,
      },
      {
        id: "dq_2",
        orderId: "ORD-88455",
        zoneName: "Beşiktaş",
        mode: "express",
        status: "delayed",
        etaMinutes: 28,
        courierCode: "CR-1003",
        waitingMinutes: 11,
      },
      {
        id: "dq_3",
        orderId: "ORD-88501",
        zoneName: "Şişli",
        mode: "scheduled",
        status: "queued",
        etaMinutes: 45,
        courierCode: null,
        waitingMinutes: 1,
      },
      {
        id: "dq_4",
        orderId: "ORD-88540",
        zoneName: "Kadıköy",
        mode: "batch",
        status: "dispatched",
        etaMinutes: 20,
        courierCode: "CR-1001",
        waitingMinutes: 0,
      },
      {
        id: "dq_5",
        orderId: "ORD-88566",
        zoneName: "Üsküdar",
        mode: "express",
        status: "queued",
        etaMinutes: 18,
        courierCode: null,
        waitingMinutes: 5,
      },
    ],
    allocations: [
      {
        id: "alloc_1",
        zoneId: "zn_kad",
        zoneName: "Kadıköy",
        courierCode: "CR-1001",
        courierName: "Ahmet Yılmaz",
        liveStatus: "busy",
        capacity: 2,
      },
      {
        id: "alloc_2",
        zoneId: "zn_bes",
        zoneName: "Beşiktaş",
        courierCode: "CR-1002",
        courierName: "Elif Demir",
        liveStatus: "available",
        capacity: 3,
      },
      {
        id: "alloc_3",
        zoneId: "zn_sis",
        zoneName: "Şişli",
        courierCode: "CR-1003",
        courierName: "Can Öztürk",
        liveStatus: "emergency",
        capacity: 1,
      },
    ],
    etaMonitor: zones.map((z) => ({
      zoneName: z.name,
      p50: z.avgEtaMinutes,
      p90: z.avgEtaMinutes + 8,
      breached: Math.max(0, Math.round((100 - z.slaPct) / 2)),
    })),
    modeBreakdown: [
      { mode: "express", count: 186 },
      { mode: "scheduled", count: 42 },
      { mode: "batch", count: 27 },
    ],
  };
}

function mockZoneDetail(id: string): DeliveryZoneDetail {
  const base =
    MOCK_ZONES.find((z) => z.id === id) ??
    ({ ...MOCK_ZONES[0], id } satisfies DeliveryZoneListItem);
  return {
    id: base.id,
    code: base.code,
    name: base.name,
    cityId: base.cityId,
    status: base.status,
    hexCount: Math.round(base.polygonPoints / 2),
    courierAllocated: base.courierAllocated,
    courierTarget: base.courierTarget,
    peakWindows: ["12:00–14:00", "18:00–21:00"],
    notes: "Polygon editable in zones editor; changes audit-logged.",
  };
}

export async function fetchDeliverySnapshot(
  cityId: string | null,
): Promise<DeliverySnapshot> {
  try {
    const q = cityId ? `?cityId=${encodeURIComponent(cityId)}` : "";
    return await apiClient<DeliverySnapshot>(`/admin/delivery${q}`);
  } catch {
    await delay();
    return mockSnapshot(cityId);
  }
}

export async function fetchDeliveryZones(
  cityId: string | null,
): Promise<{ items: DeliveryZoneListItem[]; generatedAt: string }> {
  try {
    const q = cityId ? `?cityId=${encodeURIComponent(cityId)}` : "";
    return await apiClient(`/admin/delivery/zones${q}`);
  } catch {
    await delay();
    const items = cityId
      ? MOCK_ZONES.filter((z) => z.cityId === cityId)
      : MOCK_ZONES;
    return { items, generatedAt: new Date().toISOString() };
  }
}

export async function fetchDeliveryZoneDetail(
  id: string,
): Promise<DeliveryZoneDetail> {
  try {
    return await apiClient<DeliveryZoneDetail>(`/admin/delivery/zones/${id}`);
  } catch {
    await delay();
    return mockZoneDetail(id);
  }
}
