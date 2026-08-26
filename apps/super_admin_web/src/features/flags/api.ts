import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  FeatureFlag,
  FlagsSnapshot,
  KillSwitchProposal,
  UpdateRolloutInput,
  UpsertFlagInput,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

let mockFlags: FeatureFlag[] = [
  {
    id: "flg_global_checkout",
    key: "checkout.v2",
    name: "Checkout V2",
    description: "New checkout flow worldwide",
    scope: "global",
    scopeTarget: null,
    kind: "feature",
    status: "enabled",
    enabled: true,
    rolloutPct: 35,
    variants: [],
    scheduledAt: null,
    expiresAt: null,
    updatedAt: new Date(Date.now() - 3 * 3600_000).toISOString(),
    updatedBy: "platform_sre",
  },
  {
    id: "flg_tr_darkstore",
    key: "darkstore.slotting",
    name: "Darkstore slotting",
    description: "Warehouse slotting algorithm — Turkey",
    scope: "country",
    scopeTarget: "TR",
    kind: "feature",
    status: "enabled",
    enabled: true,
    rolloutPct: 100,
    variants: [],
    scheduledAt: null,
    expiresAt: null,
    updatedAt: new Date(Date.now() - 26 * 3600_000).toISOString(),
    updatedBy: "platform_owner",
  },
  {
    id: "flg_acme_ai",
    key: "ai.demand_forecast",
    name: "AI demand forecast",
    description: "Company-scoped forecast model",
    scope: "company",
    scopeTarget: "co_acme",
    kind: "ab_test",
    status: "enabled",
    enabled: true,
    rolloutPct: 50,
    variants: [
      { key: "control", weightPct: 50 },
      { key: "treatment", weightPct: 50 },
    ],
    scheduledAt: null,
    expiresAt: new Date(Date.now() + 14 * 86400_000).toISOString(),
    updatedAt: new Date(Date.now() - 8 * 3600_000).toISOString(),
    updatedBy: "platform_finops",
  },
  {
    id: "flg_user_beta",
    key: "courier.app.beta",
    name: "Courier app beta",
    description: "Per-user beta channel",
    scope: "user",
    scopeTarget: "usr_courier_lead",
    kind: "feature",
    status: "enabled",
    enabled: true,
    rolloutPct: 100,
    variants: [],
    scheduledAt: null,
    expiresAt: null,
    updatedAt: new Date(Date.now() - 48 * 3600_000).toISOString(),
    updatedBy: "platform_sre",
  },
  {
    id: "flg_sched_payments",
    key: "payments.instant_payout",
    name: "Instant payouts",
    description: "Scheduled enable for APAC",
    scope: "global",
    scopeTarget: null,
    kind: "scheduled",
    status: "scheduled",
    enabled: false,
    rolloutPct: 10,
    variants: [],
    scheduledAt: new Date(Date.now() + 2 * 86400_000).toISOString(),
    expiresAt: null,
    updatedAt: new Date(Date.now() - 12 * 3600_000).toISOString(),
    updatedBy: "platform_owner",
  },
  {
    id: "flg_kill_orders",
    key: "kill.orders.intake",
    name: "Kill: order intake",
    description: "Emergency halt of new orders globally",
    scope: "global",
    scopeTarget: null,
    kind: "kill_switch",
    status: "disabled",
    enabled: false,
    rolloutPct: 100,
    variants: [],
    scheduledAt: null,
    expiresAt: null,
    updatedAt: new Date(Date.now() - 72 * 3600_000).toISOString(),
    updatedBy: "platform_security",
  },
  {
    id: "flg_kill_payments",
    key: "kill.payments.capture",
    name: "Kill: payment capture",
    description: "Emergency stop payment capture",
    scope: "global",
    scopeTarget: null,
    kind: "kill_switch",
    status: "disabled",
    enabled: false,
    rolloutPct: 100,
    variants: [],
    scheduledAt: null,
    expiresAt: null,
    updatedAt: new Date(Date.now() - 96 * 3600_000).toISOString(),
    updatedBy: "platform_security",
  },
];

let mockProposals: KillSwitchProposal[] = [
  {
    id: "prop_ks1",
    action: "kill_switch",
    flagId: "flg_kill_orders",
    flagKey: "kill.orders.intake",
    flagName: "Kill: order intake",
    targetEnabled: true,
    requesterId: "usr_platform_sre_demo",
    requesterEmail: "sre@nexora.platform",
    status: "pending",
    reason: "Suspected fraud cascade — pause intake until cleared",
    createdAt: new Date(Date.now() - 45 * 60_000).toISOString(),
  },
];

function snapshot(): FlagsSnapshot {
  return {
    flags: [...mockFlags],
    proposals: [...mockProposals],
    generatedAt: new Date().toISOString(),
  };
}

export async function fetchFlags(): Promise<FlagsSnapshot> {
  try {
    return await apiClient<FlagsSnapshot>(platformPath("/flags"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return snapshot();
    }
    throw err;
  }
}

export async function upsertFlag(input: UpsertFlagInput): Promise<FeatureFlag> {
  try {
    return await apiClient<FeatureFlag>(platformPath("/flags"), {
      method: "POST",
      body: input,
      idempotent: true,
    });
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const created: FeatureFlag = {
        id: `flg_${Date.now()}`,
        key: input.key,
        name: input.name,
        description: input.description,
        scope: input.scope,
        scopeTarget: input.scopeTarget ?? null,
        kind: input.kind,
        status: input.scheduledAt ? "scheduled" : "disabled",
        enabled: false,
        rolloutPct: input.rolloutPct,
        variants: input.variants ?? [],
        scheduledAt: input.scheduledAt ?? null,
        expiresAt: null,
        updatedAt: new Date().toISOString(),
        updatedBy: "current_user",
      };
      mockFlags = [created, ...mockFlags];
      return created;
    }
    throw err;
  }
}

export async function updateFlagRollout(
  input: UpdateRolloutInput,
): Promise<FeatureFlag> {
  try {
    return await apiClient<FeatureFlag>(
      platformPath(`/flags/${input.flagId}/rollout`),
      { method: "PATCH", body: input, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockFlags.findIndex((f) => f.id === input.flagId);
      if (idx < 0) throw new Error("Flag not found");
      mockFlags[idx] = {
        ...mockFlags[idx],
        rolloutPct: input.rolloutPct,
        variants: input.variants ?? mockFlags[idx].variants,
        updatedAt: new Date().toISOString(),
      };
      return mockFlags[idx];
    }
    throw err;
  }
}

export async function toggleFeatureFlag(
  flagId: string,
  enabled: boolean,
): Promise<FeatureFlag> {
  try {
    return await apiClient<FeatureFlag>(
      platformPath(`/flags/${flagId}/toggle`),
      { method: "POST", body: { enabled }, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockFlags.findIndex((f) => f.id === flagId);
      if (idx < 0) throw new Error("Flag not found");
      const flag = mockFlags[idx];
      if (flag.kind === "kill_switch") {
        throw new Error("Kill switches require dual-control");
      }
      mockFlags[idx] = {
        ...flag,
        enabled,
        status: enabled ? "enabled" : "disabled",
        updatedAt: new Date().toISOString(),
      };
      return mockFlags[idx];
    }
    throw err;
  }
}

export async function emergencyRollback(flagId: string): Promise<FeatureFlag> {
  try {
    return await apiClient<FeatureFlag>(
      platformPath(`/flags/${flagId}/rollback`),
      { method: "POST", body: {}, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockFlags.findIndex((f) => f.id === flagId);
      if (idx < 0) throw new Error("Flag not found");
      mockFlags[idx] = {
        ...mockFlags[idx],
        enabled: false,
        rolloutPct: 0,
        status: "rolling_back",
        updatedAt: new Date().toISOString(),
      };
      return mockFlags[idx];
    }
    throw err;
  }
}

export async function proposeKillSwitch(input: {
  flagId: string;
  targetEnabled: boolean;
  reason: string;
  requesterId: string;
  requesterEmail: string;
}): Promise<KillSwitchProposal> {
  try {
    return await apiClient<KillSwitchProposal>(
      platformPath("/flags/dual-control"),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const flag = mockFlags.find((f) => f.id === input.flagId);
      if (!flag || flag.kind !== "kill_switch") {
        throw new Error("Not a kill switch");
      }
      const proposal: KillSwitchProposal = {
        id: `prop_${Date.now()}`,
        action: "kill_switch",
        flagId: flag.id,
        flagKey: flag.key,
        flagName: flag.name,
        targetEnabled: input.targetEnabled,
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

export async function resolveKillSwitchProposal(input: {
  proposalId: string;
  decision: "approved" | "rejected";
  approverId: string;
}): Promise<KillSwitchProposal> {
  try {
    return await apiClient<KillSwitchProposal>(
      platformPath(`/flags/dual-control/${input.proposalId}`),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockProposals.findIndex((p) => p.id === input.proposalId);
      if (idx < 0) throw new Error("Proposal not found");
      const current = mockProposals[idx];
      const next: KillSwitchProposal = {
        ...current,
        status: input.decision === "approved" ? "executed" : "rejected",
      };
      mockProposals[idx] = next;
      if (input.decision === "approved") {
        const fIdx = mockFlags.findIndex((f) => f.id === current.flagId);
        if (fIdx >= 0) {
          mockFlags[fIdx] = {
            ...mockFlags[fIdx],
            enabled: current.targetEnabled,
            status: current.targetEnabled ? "enabled" : "disabled",
            updatedAt: new Date().toISOString(),
          };
        }
      }
      return next;
    }
    throw err;
  }
}
