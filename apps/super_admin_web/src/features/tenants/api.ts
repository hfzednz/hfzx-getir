import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  CreateTenantInput,
  TenantDetail,
  TenantDualControlProposal,
  TenantListItem,
  TenantsListResponse,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const MOCK_TENANTS: TenantListItem[] = [
  {
    id: "ten_acme",
    slug: "acme-qc",
    name: "ACME Quick Commerce",
    companyId: "co_acme",
    companyName: "ACME Holdings A.Ş.",
    isolationMode: "shared",
    status: "active",
    region: "eu-west-1",
    createdAt: "2024-03-12T10:00:00.000Z",
  },
  {
    id: "ten_nova",
    slug: "nova-market",
    name: "Nova Market",
    companyId: "co_nova",
    companyName: "Nova Retail GmbH",
    isolationMode: "hybrid",
    status: "active",
    region: "eu-central-1",
    createdAt: "2024-06-01T08:30:00.000Z",
  },
  {
    id: "ten_orbit",
    slug: "orbit-ent",
    name: "Orbit Enterprise",
    companyId: "co_orbit",
    companyName: "Orbit Logistics Ltd",
    isolationMode: "separate",
    status: "migrating",
    region: "ap-southeast-1",
    createdAt: "2023-11-20T14:15:00.000Z",
  },
  {
    id: "ten_delta",
    slug: "delta-city",
    name: "Delta City Ops",
    companyId: "co_delta",
    companyName: "Delta Commerce Inc",
    isolationMode: "shared",
    status: "suspended",
    region: "us-east-1",
    createdAt: "2025-01-08T16:00:00.000Z",
  },
];

let mockProposals: TenantDualControlProposal[] = [
  {
    id: "prop_t1",
    action: "tenant_suspend",
    tenantId: "ten_nova",
    tenantName: "Nova Market",
    requesterId: "usr_platform_sre_demo",
    requesterEmail: "sre@nexora.platform",
    status: "pending",
    reason: "Billing dispute — freeze until legal clears",
    createdAt: new Date(Date.now() - 2 * 3600_000).toISOString(),
  },
];

function mockDetail(id: string): TenantDetail {
  const base =
    MOCK_TENANTS.find((t) => t.id === id) ??
    ({
      ...MOCK_TENANTS[0],
      id,
      slug: id,
      name: `Tenant ${id}`,
    } satisfies TenantListItem);

  return {
    ...base,
    description: `${base.name} platform tenant under ${base.companyName}.`,
    config: {
      featurePack: base.isolationMode === "separate" ? "enterprise" : "standard",
      maxWarehouses: base.isolationMode === "shared" ? 50 : 500,
      maxUsers: base.isolationMode === "shared" ? 200 : 5000,
      dataResidency: base.region,
      rlsEnabled: base.isolationMode !== "separate",
    },
    customization: {
      primaryColor: "#0B6E6E",
      logoUrl: `/assets/tenants/${base.slug}.svg`,
      customDomain:
        base.isolationMode === "shared" ? null : `app.${base.slug}.example`,
      whiteLabel: base.isolationMode !== "shared",
    },
    migration: {
      status: base.status === "migrating" ? "in_progress" : "idle",
      targetMode: base.status === "migrating" ? "separate" : null,
      progressPct: base.status === "migrating" ? 62 : 0,
      lastMessage:
        base.status === "migrating"
          ? "Copying ledger shards to isolated cluster"
          : "No migration scheduled",
      updatedAt: new Date().toISOString(),
    },
    backups: [
      {
        id: `bak_${base.id}_1`,
        label: "Daily full",
        status: "ok",
        sizeGb: 48.2,
        takenAt: new Date(Date.now() - 6 * 3600_000).toISOString(),
      },
      {
        id: `bak_${base.id}_2`,
        label: "PITR checkpoint",
        status: base.status === "suspended" ? "stale" : "ok",
        sizeGb: 12.4,
        takenAt: new Date(Date.now() - 26 * 3600_000).toISOString(),
      },
    ],
    monitoring: [
      {
        id: "m1",
        label: "API error rate",
        value: "0.12%",
        tone: "success",
      },
      {
        id: "m2",
        label: "DB connections",
        value: base.isolationMode === "shared" ? "340 / 800" : "92 / 200",
        tone: "neutral",
      },
      {
        id: "m3",
        label: "Storage",
        value: "1.8 TB",
        tone: "warning",
      },
      {
        id: "m4",
        label: "SLO burn",
        value: base.status === "suspended" ? "n/a" : "0.4x",
        tone: base.status === "suspended" ? "danger" : "success",
      },
    ],
  };
}

function listResponse(): TenantsListResponse {
  return {
    items: [...MOCK_TENANTS],
    total: MOCK_TENANTS.length,
    generatedAt: new Date().toISOString(),
    proposals: [...mockProposals],
  };
}

export async function fetchTenants(): Promise<TenantsListResponse> {
  try {
    return await apiClient<TenantsListResponse>(platformPath("/tenants"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return listResponse();
    }
    throw err;
  }
}

export async function fetchTenant(id: string): Promise<TenantDetail> {
  try {
    return await apiClient<TenantDetail>(platformPath(`/tenants/${id}`));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockDetail(id);
    }
    throw err;
  }
}

export async function createTenant(
  input: CreateTenantInput,
): Promise<TenantListItem> {
  try {
    return await apiClient<TenantListItem>(platformPath("/tenants"), {
      method: "POST",
      body: input,
      idempotent: true,
    });
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const created: TenantListItem = {
        id: `ten_${input.slug.replace(/[^a-z0-9]/gi, "_")}`,
        slug: input.slug,
        name: input.name,
        companyId: input.companyId,
        companyName: "New company",
        isolationMode: input.isolationMode,
        status: "pending",
        region: input.region,
        createdAt: new Date().toISOString(),
      };
      MOCK_TENANTS.unshift(created);
      return created;
    }
    throw err;
  }
}

export async function updateTenantIsolation(
  id: string,
  isolationMode: TenantListItem["isolationMode"],
): Promise<TenantDetail> {
  try {
    return await apiClient<TenantDetail>(
      platformPath(`/tenants/${id}/isolation`),
      { method: "PATCH", body: { isolationMode }, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = MOCK_TENANTS.findIndex((t) => t.id === id);
      if (idx >= 0) {
        MOCK_TENANTS[idx] = { ...MOCK_TENANTS[idx], isolationMode };
      }
      return mockDetail(id);
    }
    throw err;
  }
}

export async function proposeTenantAction(input: {
  tenantId: string;
  action: "tenant_suspend" | "tenant_delete";
  reason: string;
  requesterId: string;
  requesterEmail: string;
}): Promise<TenantDualControlProposal> {
  try {
    return await apiClient<TenantDualControlProposal>(
      platformPath("/tenants/dual-control"),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const tenant = MOCK_TENANTS.find((t) => t.id === input.tenantId);
      const proposal: TenantDualControlProposal = {
        id: `prop_${Date.now()}`,
        action: input.action,
        tenantId: input.tenantId,
        tenantName: tenant?.name ?? input.tenantId,
        requesterId: input.requesterId,
        requesterEmail: input.requesterEmail,
        status: "pending",
        reason: input.reason,
        createdAt: new Date().toISOString(),
      };
      mockProposals = [proposal, ...mockProposals];
      return proposal;
    }
    throw err;
  }
}

export async function resolveTenantProposal(input: {
  proposalId: string;
  decision: "approved" | "rejected";
  approverId: string;
}): Promise<TenantDualControlProposal> {
  try {
    return await apiClient<TenantDualControlProposal>(
      platformPath(`/tenants/dual-control/${input.proposalId}`),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockProposals.findIndex((p) => p.id === input.proposalId);
      if (idx < 0) throw new Error("Proposal not found");
      const current = mockProposals[idx];
      const next: TenantDualControlProposal = {
        ...current,
        status: input.decision === "approved" ? "executed" : "rejected",
      };
      mockProposals[idx] = next;
      if (input.decision === "approved") {
        const tIdx = MOCK_TENANTS.findIndex((t) => t.id === current.tenantId);
        if (tIdx >= 0) {
          if (current.action === "tenant_suspend") {
            MOCK_TENANTS[tIdx] = {
              ...MOCK_TENANTS[tIdx],
              status: "suspended",
            };
          } else {
            MOCK_TENANTS.splice(tIdx, 1);
          }
        }
      }
      return next;
    }
    throw err;
  }
}
