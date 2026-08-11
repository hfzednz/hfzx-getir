import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  LicenseOverrideProposal,
  LicensesSnapshot,
  TenantLicense,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const MOCK_PLANS = [
  {
    id: "plan_starter",
    code: "starter",
    name: "Starter",
    tier: "starter" as const,
    monthlyPriceMinor: 49_900,
    currency: "USD",
    features: ["orders", "couriers", "basic_analytics"],
    limits: {
      warehouses: 5,
      couriers: 50,
      apiCallsPerMonth: 500_000,
      storageGb: 100,
      seats: 10,
    },
  },
  {
    id: "plan_growth",
    code: "growth",
    name: "Growth",
    tier: "growth" as const,
    monthlyPriceMinor: 249_900,
    currency: "USD",
    features: [
      "orders",
      "couriers",
      "analytics",
      "campaigns",
      "ai_assist",
    ],
    limits: {
      warehouses: 50,
      couriers: 500,
      apiCallsPerMonth: 5_000_000,
      storageGb: 1_000,
      seats: 100,
    },
  },
  {
    id: "plan_ent",
    code: "enterprise",
    name: "Enterprise",
    tier: "enterprise" as const,
    monthlyPriceMinor: 0,
    currency: "USD",
    features: [
      "orders",
      "couriers",
      "analytics",
      "campaigns",
      "ai_assist",
      "dedicated_support",
      "custom_sso",
      "isolated_db",
    ],
    limits: {
      warehouses: 500,
      couriers: 10_000,
      apiCallsPerMonth: 50_000_000,
      storageGb: 20_000,
      seats: 5_000,
    },
  },
  {
    id: "plan_custom",
    code: "custom",
    name: "Custom enterprise",
    tier: "custom" as const,
    monthlyPriceMinor: 0,
    currency: "USD",
    features: ["negotiated"],
    limits: {
      warehouses: 0,
      couriers: 0,
      apiCallsPerMonth: 0,
      storageGb: 0,
      seats: 0,
    },
  },
];

const mockLicenses: TenantLicense[] = [
  {
    id: "lic_acme",
    tenantId: "ten_acme",
    tenantName: "ACME Quick Commerce",
    planId: "plan_growth",
    planName: "Growth",
    tier: "growth",
    status: "active",
    seats: 100,
    seatsUsed: 78,
    renewsAt: new Date(Date.now() + 45 * 86400_000).toISOString(),
    overageEnabled: true,
    overageSpendMinor: 12_400_00,
    currency: "USD",
    featureOverrides: [],
    enterpriseContractId: null,
  },
  {
    id: "lic_nova",
    tenantId: "ten_nova",
    tenantName: "Nova Market",
    planId: "plan_ent",
    planName: "Enterprise",
    tier: "enterprise",
    status: "active",
    seats: 2000,
    seatsUsed: 1140,
    renewsAt: new Date(Date.now() + 12 * 86400_000).toISOString(),
    overageEnabled: true,
    overageSpendMinor: 0,
    currency: "USD",
    featureOverrides: ["white_label"],
    enterpriseContractId: "ENT-2024-8841",
  },
  {
    id: "lic_orbit",
    tenantId: "ten_orbit",
    tenantName: "Orbit Enterprise",
    planId: "plan_custom",
    planName: "Custom enterprise",
    tier: "custom",
    status: "active",
    seats: 8000,
    seatsUsed: 6200,
    renewsAt: new Date(Date.now() + 90 * 86400_000).toISOString(),
    overageEnabled: false,
    overageSpendMinor: 0,
    currency: "USD",
    featureOverrides: ["isolated_db", "dedicated_kafka"],
    enterpriseContractId: "ENT-2023-1102",
  },
  {
    id: "lic_delta",
    tenantId: "ten_delta",
    tenantName: "Delta City Ops",
    planId: "plan_starter",
    planName: "Starter",
    tier: "starter",
    status: "past_due",
    seats: 10,
    seatsUsed: 10,
    renewsAt: new Date(Date.now() - 5 * 86400_000).toISOString(),
    overageEnabled: false,
    overageSpendMinor: 3_200_00,
    currency: "USD",
    featureOverrides: [],
    enterpriseContractId: null,
  },
];

let mockProposals: LicenseOverrideProposal[] = [
  {
    id: "prop_lic1",
    action: "license_override",
    licenseId: "lic_delta",
    tenantName: "Delta City Ops",
    requesterId: "usr_platform_finops_demo",
    requesterEmail: "finops@nexora.platform",
    status: "pending",
    reason: "Grace seats during payment remediation",
    payload: { seats: 15, overageEnabled: true },
    createdAt: new Date(Date.now() - 3 * 3600_000).toISOString(),
  },
];

function snapshot(): LicensesSnapshot {
  return {
    plans: MOCK_PLANS,
    licenses: [...mockLicenses],
    usage: [
      {
        id: "u1",
        tenantId: "ten_acme",
        tenantName: "ACME Quick Commerce",
        metric: "API calls",
        used: 4_820_000,
        limit: 5_000_000,
        unit: "calls",
        overagePct: 0,
      },
      {
        id: "u2",
        tenantId: "ten_acme",
        tenantName: "ACME Quick Commerce",
        metric: "Storage",
        used: 1180,
        limit: 1000,
        unit: "GB",
        overagePct: 18,
      },
      {
        id: "u3",
        tenantId: "ten_nova",
        tenantName: "Nova Market",
        metric: "Warehouses",
        used: 120,
        limit: 500,
        unit: "sites",
        overagePct: 0,
      },
      {
        id: "u4",
        tenantId: "ten_delta",
        tenantName: "Delta City Ops",
        metric: "Seats",
        used: 10,
        limit: 10,
        unit: "users",
        overagePct: 0,
      },
      {
        id: "u5",
        tenantId: "ten_orbit",
        tenantName: "Orbit Enterprise",
        metric: "Couriers",
        used: 8400,
        limit: 10_000,
        unit: "active",
        overagePct: 0,
      },
    ],
    proposals: [...mockProposals],
    renewalsDueDays: 14,
    generatedAt: new Date().toISOString(),
  };
}

export async function fetchLicenses(): Promise<LicensesSnapshot> {
  try {
    return await apiClient<LicensesSnapshot>(platformPath("/licenses"));
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return snapshot();
    }
    throw err;
  }
}

export async function renewLicense(licenseId: string): Promise<TenantLicense> {
  try {
    return await apiClient<TenantLicense>(
      platformPath(`/licenses/${licenseId}/renew`),
      { method: "POST", body: {}, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockLicenses.findIndex((l) => l.id === licenseId);
      if (idx < 0) throw new Error("License not found");
      mockLicenses[idx] = {
        ...mockLicenses[idx],
        status: "active",
        renewsAt: new Date(Date.now() + 365 * 86400_000).toISOString(),
      };
      return mockLicenses[idx];
    }
    throw err;
  }
}

export async function proposeLicenseOverride(input: {
  licenseId: string;
  reason: string;
  requesterId: string;
  requesterEmail: string;
  payload: LicenseOverrideProposal["payload"];
}): Promise<LicenseOverrideProposal> {
  try {
    return await apiClient<LicenseOverrideProposal>(
      platformPath("/licenses/dual-control"),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const lic = mockLicenses.find((l) => l.id === input.licenseId);
      const proposal: LicenseOverrideProposal = {
        id: `prop_${Date.now()}`,
        action: "license_override",
        licenseId: input.licenseId,
        tenantName: lic?.tenantName ?? input.licenseId,
        requesterId: input.requesterId,
        requesterEmail: input.requesterEmail,
        status: "pending",
        reason: input.reason,
        payload: input.payload,
        createdAt: new Date().toISOString(),
      };
      mockProposals = [proposal, ...mockProposals];
      return proposal;
    }
    throw err;
  }
}

export async function resolveLicenseProposal(input: {
  proposalId: string;
  decision: "approved" | "rejected";
  approverId: string;
}): Promise<LicenseOverrideProposal> {
  try {
    return await apiClient<LicenseOverrideProposal>(
      platformPath(`/licenses/dual-control/${input.proposalId}`),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockProposals.findIndex((p) => p.id === input.proposalId);
      if (idx < 0) throw new Error("Proposal not found");
      const current = mockProposals[idx];
      const next: LicenseOverrideProposal = {
        ...current,
        status: input.decision === "approved" ? "executed" : "rejected",
      };
      mockProposals[idx] = next;
      if (input.decision === "approved") {
        const lIdx = mockLicenses.findIndex((l) => l.id === current.licenseId);
        if (lIdx >= 0) {
          mockLicenses[lIdx] = {
            ...mockLicenses[lIdx],
            seats: current.payload.seats ?? mockLicenses[lIdx].seats,
            overageEnabled:
              current.payload.overageEnabled ??
              mockLicenses[lIdx].overageEnabled,
            featureOverrides:
              current.payload.featureOverrides ??
              mockLicenses[lIdx].featureOverrides,
          };
        }
      }
      return next;
    }
    throw err;
  }
}
