import { apiClient } from "@/shared/api/client";
import type { LiveOpsSnapshot } from "./types";

function minutesAgo(m: number): string {
  return new Date(Date.now() - m * 60_000).toISOString();
}

/** Mock live ops payload — used when BFF is unavailable. */
export function buildMockLiveSnapshot(
  cityId: string | null,
  connection: LiveOpsSnapshot["connection"] = "polling",
): LiveOpsSnapshot {
  const tick = Math.floor(Date.now() / 15_000) % 5;

  return {
    cityId,
    generatedAt: new Date().toISOString(),
    connection,
    orderStream: [
      {
        id: `evt_${tick}_1`,
        orderId: "ord_88421",
        status: "en_route",
        customerName: "Ayşe K.",
        warehouseCode: "WH-07",
        zone: "Kadıköy",
        etaMinutes: 9 + tick,
        delayMinutes: tick > 2 ? 4 : 0,
        amountMinor: 312_50,
        currency: "TRY",
        updatedAt: minutesAgo(1),
      },
      {
        id: `evt_${tick}_2`,
        orderId: "ord_88418",
        status: "picking",
        customerName: "Mehmet Y.",
        warehouseCode: "WH-14",
        zone: "Beşiktaş",
        etaMinutes: 22,
        delayMinutes: 6,
        amountMinor: 189_00,
        currency: "TRY",
        updatedAt: minutesAgo(2),
      },
      {
        id: `evt_${tick}_3`,
        orderId: "ord_88405",
        status: "ready",
        customerName: "Deniz A.",
        warehouseCode: "WH-03",
        zone: "Şişli",
        etaMinutes: 14,
        delayMinutes: 0,
        amountMinor: 455_90,
        currency: "TRY",
        updatedAt: minutesAgo(3),
      },
      {
        id: `evt_${tick}_4`,
        orderId: "ord_88391",
        status: "assigned",
        customerName: "Caner T.",
        warehouseCode: "WH-07",
        zone: "Üsküdar",
        etaMinutes: 18,
        delayMinutes: 0,
        amountMinor: 98_40,
        currency: "TRY",
        updatedAt: minutesAgo(4),
      },
      {
        id: `evt_${tick}_5`,
        orderId: "ord_88370",
        status: "failed",
        customerName: "Selin M.",
        warehouseCode: "WH-11",
        zone: "Bakırköy",
        etaMinutes: null,
        delayMinutes: 25,
        amountMinor: 267_20,
        currency: "TRY",
        updatedAt: minutesAgo(8),
      },
    ],
    couriers: [
      {
        id: "cr_101",
        name: "Emre B.",
        status: "busy",
        lat: 41.0082,
        lng: 28.9784,
        zone: "Kadıköy",
        activeOrderId: "ord_88421",
        batteryPct: 72,
        updatedAt: minutesAgo(0),
      },
      {
        id: "cr_214",
        name: "Zeynep S.",
        status: "available",
        lat: 41.0422,
        lng: 29.0067,
        zone: "Beşiktaş",
        activeOrderId: null,
        batteryPct: 91,
        updatedAt: minutesAgo(1),
      },
      {
        id: "cr_308",
        name: "Burak D.",
        status: "busy",
        lat: 41.0602,
        lng: 28.9877,
        zone: "Şişli",
        activeOrderId: "ord_88405",
        batteryPct: 44,
        updatedAt: minutesAgo(1),
      },
      {
        id: "cr_412",
        name: "İrem K.",
        status: "break",
        lat: 41.0225,
        lng: 29.015,
        zone: "Üsküdar",
        activeOrderId: null,
        batteryPct: 58,
        updatedAt: minutesAgo(5),
      },
      {
        id: "cr_519",
        name: "Onur L.",
        status: "offline",
        lat: 40.981,
        lng: 28.874,
        zone: "Bakırköy",
        activeOrderId: null,
        batteryPct: 12,
        updatedAt: minutesAgo(40),
      },
    ],
    warehouses: [
      {
        id: "wh_07",
        code: "WH-07",
        name: "Kadıköy Dark",
        pickQueueMin: 4 + tick,
        openOrders: 38,
        capacityPct: 71,
        status: "healthy",
      },
      {
        id: "wh_14",
        code: "WH-14",
        name: "Beşiktaş Dark",
        pickQueueMin: 13 + tick,
        openOrders: 62,
        capacityPct: 94,
        status: "critical",
      },
      {
        id: "wh_03",
        code: "WH-03",
        name: "Şişli Dark",
        pickQueueMin: 8,
        openOrders: 45,
        capacityPct: 82,
        status: "busy",
      },
      {
        id: "wh_11",
        code: "WH-11",
        name: "Bakırköy Dark",
        pickQueueMin: 6,
        openOrders: 29,
        capacityPct: 64,
        status: "healthy",
      },
    ],
    delays: [
      {
        id: "dl_1",
        orderId: "ord_88418",
        zone: "Beşiktaş",
        delayMinutes: 6 + tick,
        reason: "Pick queue backlog",
        severity: "warning",
      },
      {
        id: "dl_2",
        orderId: "ord_88370",
        zone: "Bakırköy",
        delayMinutes: 25,
        reason: "Customer unreachable",
        severity: "danger",
      },
      {
        id: "dl_3",
        orderId: "ord_88355",
        zone: "Kadıköy",
        delayMinutes: 11,
        reason: "Traffic near bridge",
        severity: "warning",
      },
    ],
    failedDeliveries: [
      {
        id: "fd_1",
        orderId: "ord_88370",
        courierName: "Onur L.",
        reason: "Address not found / no answer",
        attempts: 2,
        failedAt: minutesAgo(8),
      },
      {
        id: "fd_2",
        orderId: "ord_88201",
        courierName: "Emre B.",
        reason: "Payment dispute at door",
        attempts: 1,
        failedAt: minutesAgo(55),
      },
    ],
    bottlenecks: [
      {
        id: "bn_1",
        type: "warehouse",
        title: "WH-14 pick queue",
        detail: "Queue > 12 min · dinner peak understaffed",
        severity: "danger",
        impactScore: 92,
      },
      {
        id: "bn_2",
        type: "courier",
        title: "Beşiktaş courier shortage",
        detail: "Available / busy ratio below SLA threshold",
        severity: "warning",
        impactScore: 74,
      },
      {
        id: "bn_3",
        type: "zone",
        title: "Bridge traffic hex",
        detail: "ETA inflation +4–7 min on Kadıköy ↔ Beşiktaş",
        severity: "info",
        impactScore: 41,
      },
    ],
    emergencies: [
      {
        id: "em_1",
        title: "Courier SOS",
        detail: "cr_308 reported incident near Şişli metro",
        zone: "Şişli",
        raisedAt: minutesAgo(3),
        acknowledged: false,
      },
    ],
    alerts: [
      {
        id: "oa_1",
        severity: "danger",
        title: "Emergency open",
        detail: "Unacknowledged SOS in Şişli",
        createdAt: minutesAgo(3),
      },
      {
        id: "oa_2",
        severity: "warning",
        title: "SLA risk cluster",
        detail: "14 orders projected late in Beşiktaş",
        createdAt: minutesAgo(7),
      },
      {
        id: "oa_3",
        severity: "info",
        title: "Flash campaign load",
        detail: "+18% order ingress vs forecast (last 20 min)",
        createdAt: minutesAgo(15),
      },
    ],
    counts: {
      activeOrders: 1284 + tick,
      delayedOrders: 47 + tick,
      availableCouriers: 312 - tick,
      openEmergencies: 1,
    },
  };
}

export async function fetchLiveSnapshot(
  cityId: string | null,
): Promise<LiveOpsSnapshot> {
  try {
    const qs = cityId ? `?cityId=${encodeURIComponent(cityId)}` : "";
    return await apiClient<LiveOpsSnapshot>(`/admin/live/snapshot${qs}`);
  } catch {
    await new Promise((r) => setTimeout(r, 180));
    return buildMockLiveSnapshot(cityId, "polling");
  }
}
