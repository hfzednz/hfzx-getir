import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { AuditSnapshot } from "./types";

function delay(ms = 200): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function mockSnapshot(): AuditSnapshot {
  return {
    generatedAt: new Date().toISOString(),
    total: 6,
    immutable: true,
    items: [
      {
        id: "aud_1",
        actorId: "usr_platform_sre_demo",
        actorEmail: "sre@nexora.platform",
        action: "dr.failover.propose",
        resource: "disaster_recovery",
        resourceId: "prop_dr1",
        when: new Date(Date.now() - 45 * 60_000).toISOString(),
        where: "Istanbul, TR",
        device: "Chrome 128 / Windows",
        ip: "85.108.12.44",
        sessionId: "sess_sre_9f2a",
        oldValue: "primary=eu-west-1",
        newValue: "propose→eu-south-1",
        severity: "critical",
        sealed: true,
      },
      {
        id: "aud_2",
        actorId: "usr_platform_owner_demo",
        actorEmail: "owner@nexora.platform",
        action: "tenant.suspend.approve",
        resource: "tenant",
        resourceId: "ten_nova",
        when: new Date(Date.now() - 3 * 3600_000).toISOString(),
        where: "Dublin, IE",
        device: "Safari 17 / macOS",
        ip: "52.48.10.2",
        sessionId: "sess_own_1a2b",
        oldValue: "status=active",
        newValue: "status=suspended",
        severity: "critical",
        sealed: true,
      },
      {
        id: "aud_3",
        actorId: "usr_platform_security_demo",
        actorEmail: "sec@nexora.platform",
        action: "secret.rotate.propose",
        resource: "secret",
        resourceId: "sec_db_platform",
        when: new Date(Date.now() - 5 * 3600_000).toISOString(),
        where: "Berlin, DE",
        device: "Firefox 129 / Linux",
        ip: "91.64.22.18",
        sessionId: "sess_sec_77cd",
        oldValue: "version=12",
        newValue: "rotation_pending",
        severity: "warning",
        sealed: true,
      },
      {
        id: "aud_4",
        actorId: "usr_platform_sre_demo",
        actorEmail: "sre@nexora.platform",
        action: "deployment.rollback",
        resource: "deployment",
        resourceId: "dep_4",
        when: new Date(Date.now() - 19 * 3600_000).toISOString(),
        where: "Remote (VPN)",
        device: "Chrome 128 / Windows",
        ip: "10.40.1.14",
        sessionId: "sess_sre_9f2a",
        oldValue: "version=0.8.5",
        newValue: "version=0.8.4",
        severity: "warning",
        sealed: true,
      },
      {
        id: "aud_5",
        actorId: "usr_platform_compliance_demo",
        actorEmail: "comp@nexora.platform",
        action: "report.export",
        resource: "report",
        resourceId: "rpt_compliance_dsar",
        when: new Date(Date.now() - 26 * 3600_000).toISOString(),
        where: "Amsterdam, NL",
        device: "Edge 127 / Windows",
        ip: "77.248.12.9",
        sessionId: "sess_comp_44ee",
        oldValue: null,
        newValue: "format=pdf",
        severity: "info",
        sealed: true,
      },
      {
        id: "aud_6",
        actorId: "usr_platform_finops_demo",
        actorEmail: "finops@nexora.platform",
        action: "license.override.propose",
        resource: "license",
        resourceId: "lic_orbit_ent",
        when: new Date(Date.now() - 48 * 3600_000).toISOString(),
        where: "London, GB",
        device: "Chrome 127 / macOS",
        ip: "81.2.69.142",
        sessionId: "sess_fin_55ab",
        oldValue: "seats=500",
        newValue: "seats=2000 (override)",
        severity: "critical",
        sealed: true,
      },
    ],
  };
}

export async function fetchAudit(params?: {
  q?: string;
}): Promise<AuditSnapshot> {
  try {
    const qs = params?.q ? `?q=${encodeURIComponent(params.q)}` : "";
    return await apiClient<AuditSnapshot>(platformPath(`/audit${qs}`));
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const snap = mockSnapshot();
      if (params?.q?.trim()) {
        const n = params.q.trim().toLowerCase();
        const items = snap.items.filter(
          (e) =>
            e.actorEmail.toLowerCase().includes(n) ||
            e.action.toLowerCase().includes(n) ||
            e.resource.toLowerCase().includes(n) ||
            e.ip.includes(n) ||
            e.sessionId.toLowerCase().includes(n),
        );
        return { ...snap, items, total: items.length };
      }
      return snap;
    }
    throw err;
  }
}
