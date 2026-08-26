import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { RolesSnapshot } from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function mockSnapshot(): RolesSnapshot {
  return {
    generatedAt: new Date().toISOString(),
    roles: [
      {
        id: "r1",
        key: "platform_owner",
        label: "Platform Owner",
        scope: "global",
        members: 3,
        inheritsFrom: null,
        description: "Full platform governance",
      },
      {
        id: "r2",
        key: "platform_security",
        label: "Platform Security",
        scope: "global",
        members: 8,
        inheritsFrom: "platform_viewer",
        description: "Security command + dual-control approve",
      },
      {
        id: "r3",
        key: "platform_sre",
        label: "Platform SRE",
        scope: "global",
        members: 12,
        inheritsFrom: "platform_viewer",
        description: "Infra, DR, config write",
      },
      {
        id: "r4",
        key: "tenant_admin",
        label: "Tenant Admin",
        scope: "company",
        members: 42,
        inheritsFrom: null,
        description: "Scoped company admin (delegates to Admin Web)",
      },
      {
        id: "r5",
        key: "dept_finops_lead",
        label: "FinOps Lead",
        scope: "department",
        members: 6,
        inheritsFrom: "platform_finops",
        description: "Department-scoped billing oversight",
      },
    ],
    permissionTemplates: [
      {
        id: "pt1",
        key: "tmpl_read_all",
        label: "Read all platform",
        permissions: ["dashboard:read", "tenants:read", "audit:read"],
      },
      {
        id: "pt2",
        key: "tmpl_tenant_ops",
        label: "Tenant operations",
        permissions: [
          "tenants:read",
          "tenants:write",
          "tenants:suspend",
          "companies:write",
        ],
      },
      {
        id: "pt3",
        key: "tmpl_dual_control",
        label: "Dual-control approver",
        permissions: ["dual_control:approve", "audit:read"],
      },
    ],
    approvalChains: [
      {
        id: "ac1",
        name: "Tenant suspend/delete",
        steps: ["Requester", "platform_security / platform_owner"],
        pending: 1,
        description: "Dual-control for tenant lifecycle",
      },
      {
        id: "ac2",
        name: "Kill switch arm",
        steps: ["Requester", "Second platform_security"],
        pending: 0,
        description: "Global kill dual-control",
      },
      {
        id: "ac3",
        name: "Role elevation",
        steps: ["Manager", "platform_owner"],
        pending: 2,
        description: "Temporary permission grants",
      },
    ],
    inheritance: [
      {
        id: "ih1",
        childRole: "platform_security",
        parentRole: "platform_viewer",
      },
      { id: "ih2", childRole: "platform_sre", parentRole: "platform_viewer" },
      {
        id: "ih3",
        childRole: "dept_finops_lead",
        parentRole: "platform_finops",
      },
    ],
    temporaryPermissions: [
      {
        id: "tp1",
        subject: "sre@nexora.platform",
        permission: "tenants:suspend",
        scope: "global",
        expiresAt: new Date(Date.now() + 8 * 3600_000).toISOString(),
        reason: "Incident bridge",
        status: "active",
      },
      {
        id: "tp2",
        subject: "finops@nexora.platform",
        permission: "licenses:write",
        scope: "company",
        expiresAt: new Date(Date.now() - 3600_000).toISOString(),
        reason: "License override window",
        status: "expired",
      },
      {
        id: "tp3",
        subject: "audit@external.example",
        permission: "audit:read",
        scope: "global",
        expiresAt: new Date(Date.now() + 7 * 86400_000).toISOString(),
        reason: "Quarterly audit",
        status: "active",
      },
    ],
  };
}

export async function fetchRolesSnapshot(): Promise<RolesSnapshot> {
  try {
    return await apiClient<RolesSnapshot>(platformPath("/roles"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockSnapshot();
    }
    throw err;
  }
}
