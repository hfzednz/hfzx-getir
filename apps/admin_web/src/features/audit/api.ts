import type { AuditSnapshot } from "./types";
import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient } from "@/shared/api/client";

/** Live audit from bff-admin (derived from orders + tickets). */
export async function fetchAuditSnapshot(): Promise<AuditSnapshot> {
  try {
    return await apiClient<AuditSnapshot>("/admin/audit");
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    return mockAuditSnapshot();
  }
}

function mockAuditSnapshot(): AuditSnapshot {
  const base = Date.now();

  return {
    generatedAt: new Date().toISOString(),
    events: [
      {
        id: "ae1",
        who: "ayse.k@nexora.local",
        when: new Date(base - 3 * 60_000).toISOString(),
        where: "Istanbul HQ",
        device: "Chrome / Windows",
        action: "orders.cancel",
        resource: "ord_9f2a",
        oldValue: "status=en_route",
        newValue: "status=cancelled",
        ip: "185.42.10.22",
        sessionId: "sess_a1b2",
      },
      {
        id: "ae2",
        who: "mehmet.y@nexora.local",
        when: new Date(base - 12 * 60_000).toISOString(),
        where: "Remote VPN",
        device: "Firefox / macOS",
        action: "finance.payout.approve",
        resource: "pay_8821",
        oldValue: "state=pending",
        newValue: "state=approved",
        ip: "10.8.0.14",
        sessionId: "sess_c3d4",
      },
      {
        id: "ae3",
        who: "super@nexora.local",
        when: new Date(base - 28 * 60_000).toISOString(),
        where: "Ankara ops",
        device: "Edge / Windows",
        action: "system.flag.toggle",
        resource: "kill.dispatch",
        oldValue: "enabled=false",
        newValue: "enabled=false (confirm)",
        ip: "176.40.2.9",
        sessionId: "sess_e5f6",
      },
      {
        id: "ae4",
        who: "zeynep.a@nexora.local",
        when: new Date(base - 45 * 60_000).toISOString(),
        where: "Istanbul HQ",
        device: "Chrome / Windows",
        action: "couriers.reassign",
        resource: "ord_3bc1",
        oldValue: "courier=cr_112",
        newValue: "courier=cr_441",
        ip: "185.42.10.41",
        sessionId: "sess_g7h8",
      },
      {
        id: "ae5",
        who: "can.d@nexora.local",
        when: new Date(base - 70 * 60_000).toISOString(),
        where: "Remote",
        device: "Safari / iOS",
        action: "rbac.grant.temp",
        resource: "usr_support_12",
        oldValue: "—",
        newValue: "orders:force_complete until +6h",
        ip: "31.210.5.88",
        sessionId: "sess_i9j0",
      },
      {
        id: "ae6",
        who: "ayse.k@nexora.local",
        when: new Date(base - 95 * 60_000).toISOString(),
        where: "Istanbul HQ",
        device: "Chrome / Windows",
        action: "orders.refund",
        resource: "ord_77de",
        oldValue: "refund=0",
        newValue: "refund=18450 TRY kuruş",
        ip: "185.42.10.22",
        sessionId: "sess_a1b2",
      },
      {
        id: "ae7",
        who: "ops.bot@nexora.local",
        when: new Date(base - 110 * 60_000).toISOString(),
        where: "System",
        device: "service",
        action: "inventory.transfer",
        resource: "xfer_991",
        oldValue: "WH-07:180",
        newValue: "WH-14:180",
        ip: "10.0.0.8",
        sessionId: "sess_svc01",
      },
      {
        id: "ae8",
        who: "mehmet.y@nexora.local",
        when: new Date(base - 140 * 60_000).toISOString(),
        where: "Remote VPN",
        device: "Firefox / macOS",
        action: "customers.write",
        resource: "cus_4421",
        oldValue: "phone=***12",
        newValue: "phone=***88",
        ip: "10.8.0.14",
        sessionId: "sess_c3d4",
      },
    ],
  };
}
