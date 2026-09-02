import type { RbacSnapshot } from "./types";
import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient } from "@/shared/api/client";

/** Live RBAC from bff-admin → identity role catalog. */
export async function fetchRbacSnapshot(): Promise<RbacSnapshot> {
  try {
    return await apiClient<RbacSnapshot>("/admin/rbac");
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    return mockRbacSnapshot();
  }
}

function mockRbacSnapshot(): RbacSnapshot {
  const matrixRoles = ["viewer", "city_ops", "support_lead", "admin", "super_admin"];
  const matrixPermissions = [
    "orders:read",
    "orders:cancel",
    "orders:force_complete",
    "finance:payout:approve",
    "system:flags",
    "rbac:write",
    "audit:read",
  ];

  const grants: Record<string, string[]> = {
    viewer: ["orders:read", "audit:read"],
    city_ops: ["orders:read"],
    support_lead: ["orders:read", "orders:cancel"],
    admin: [
      "orders:read",
      "orders:cancel",
      "orders:force_complete",
      "audit:read",
      "rbac:write",
    ],
    super_admin: matrixPermissions,
  };

  const matrix = matrixRoles.flatMap((role) =>
    matrixPermissions.map((permission) => ({
      role,
      permission,
      granted: (grants[role] ?? []).includes(permission),
    })),
  );

  return {
    generatedAt: new Date().toISOString(),
    departments: [
      {
        id: "d1",
        name: "City Operations",
        headcount: 48,
        roles: ["city_ops", "admin"],
      },
      {
        id: "d2",
        name: "Customer Support",
        headcount: 120,
        roles: ["support_agent", "support_lead"],
      },
      {
        id: "d3",
        name: "Finance",
        headcount: 22,
        roles: ["finance_analyst", "finance_admin"],
      },
      {
        id: "d4",
        name: "Catalog & Growth",
        headcount: 35,
        roles: ["catalog_manager"],
      },
      {
        id: "d5",
        name: "Risk & Fraud",
        headcount: 14,
        roles: ["fraud_analyst"],
      },
      {
        id: "d6",
        name: "Warehouse Ops",
        headcount: 60,
        roles: ["warehouse_ops"],
      },
    ],
    roles: [
      {
        id: "r1",
        key: "viewer",
        label: "Viewer",
        members: 40,
        description: "Read-only dashboards",
      },
      {
        id: "r2",
        key: "city_ops",
        label: "City Ops",
        members: 48,
        description: "Live ops & reassign",
      },
      {
        id: "r3",
        key: "support_lead",
        label: "Support Lead",
        members: 18,
        description: "Escalations & cancel",
      },
      {
        id: "r4",
        key: "admin",
        label: "Admin",
        members: 12,
        description: "City admin (non-kill)",
      },
      {
        id: "r5",
        key: "super_admin",
        label: "Super Admin",
        members: 4,
        description: "Flags & kill switches",
      },
    ],
    matrix,
    matrixPermissions,
    matrixRoles,
    customPermissions: [
      {
        id: "cp1",
        key: "orders:bulk_export",
        description: "Export order batches beyond default caps",
        createdBy: "ops_platform",
      },
      {
        id: "cp2",
        key: "warehouses:override_capacity",
        description: "Temporarily raise pick capacity",
        createdBy: "city_ist_admin",
      },
    ],
    temporaryGrants: [
      {
        id: "tg1",
        user: "ayse.k@nexora.local",
        permission: "orders:force_complete",
        expiresAt: new Date(Date.now() + 6 * 3600_000).toISOString(),
        reason: "Incident bridge coverage",
        status: "active",
      },
      {
        id: "tg2",
        user: "mehmet.y@nexora.local",
        permission: "finance:export",
        expiresAt: new Date(Date.now() - 2 * 3600_000).toISOString(),
        reason: "Month-end close",
        status: "expired",
      },
      {
        id: "tg3",
        user: "zeynep.a@nexora.local",
        permission: "system:write",
        expiresAt: new Date(Date.now() + 24 * 3600_000).toISOString(),
        reason: "Locale rollout",
        status: "active",
      },
    ],
    approvals: [
      {
        id: "aw1",
        name: "Payout dual-control",
        steps: 2,
        pending: 3,
        description: "Requester + finance_admin approver",
      },
      {
        id: "aw2",
        name: "Kill switch arm",
        steps: 2,
        pending: 0,
        description: "super_admin propose + second super_admin",
      },
      {
        id: "aw3",
        name: "Role elevation",
        steps: 2,
        pending: 1,
        description: "RBAC write + audit notification",
      },
    ],
  };
}
