import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  DeploymentRecord,
  DeploymentsSnapshot,
  SecretRotateProposal,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const mockDeployments: DeploymentRecord[] = [
  {
    id: "dep_1",
    service: "bff-admin",
    environment: "production",
    strategy: "canary",
    version: "2.14.2",
    previousVersion: "2.14.1",
    status: "progressing",
    canaryPct: 15,
    startedAt: new Date(Date.now() - 25 * 60_000).toISOString(),
    finishedAt: null,
  },
  {
    id: "dep_2",
    service: "identity-service",
    environment: "production",
    strategy: "blue_green",
    version: "1.9.1",
    previousVersion: "1.9.0",
    status: "healthy",
    canaryPct: null,
    startedAt: new Date(Date.now() - 6 * 3600_000).toISOString(),
    finishedAt: new Date(Date.now() - 5.5 * 3600_000).toISOString(),
  },
  {
    id: "dep_3",
    service: "config-flags",
    environment: "staging",
    strategy: "rolling",
    version: "3.2.1",
    previousVersion: "3.2.0",
    status: "healthy",
    canaryPct: null,
    startedAt: new Date(Date.now() - 2 * 3600_000).toISOString(),
    finishedAt: new Date(Date.now() - 1.8 * 3600_000).toISOString(),
  },
  {
    id: "dep_4",
    service: "rec-inference",
    environment: "production",
    strategy: "canary",
    version: "0.8.5",
    previousVersion: "0.8.4",
    status: "rolled_back",
    canaryPct: 5,
    startedAt: new Date(Date.now() - 20 * 3600_000).toISOString(),
    finishedAt: new Date(Date.now() - 19 * 3600_000).toISOString(),
  },
];

let mockSecretProposals: SecretRotateProposal[] = [
  {
    id: "prop_sec1",
    action: "secret_rotate",
    secretId: "sec_db_platform",
    secretName: "postgres/platform-primary",
    requesterId: "usr_platform_sre_demo",
    requesterEmail: "sre@nexora.platform",
    status: "pending",
    reason: "90-day rotation policy due",
    createdAt: new Date(Date.now() - 3 * 3600_000).toISOString(),
  },
];

function mockSnapshot(): DeploymentsSnapshot {
  const days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
  return {
    generatedAt: new Date().toISOString(),
    kpis: {
      pipelinesRunning: 2,
      deploys24h: 14,
      failedPipelines: 1,
      canariesActive: 1,
      rollbacks24h: 1,
      secretsDueRotation: 3,
    },
    deployFrequency: days.map((label, i) => ({
      label,
      value: 8 + i * 2 + (i === 3 ? 6 : 0),
    })),
    pipelines: [
      {
        id: "pipe_1",
        name: "bff-admin / main",
        repo: "nexora/bff-admin",
        branch: "main",
        status: "running",
        commitSha: "a1b2c3d",
        triggeredBy: "ci@nexora.platform",
        durationSec: 420,
        updatedAt: new Date().toISOString(),
      },
      {
        id: "pipe_2",
        name: "identity / release",
        repo: "nexora/identity",
        branch: "release/1.9",
        status: "passed",
        commitSha: "9f8e7d6",
        triggeredBy: "owner@nexora.platform",
        durationSec: 610,
        updatedAt: new Date(Date.now() - 2 * 3600_000).toISOString(),
      },
      {
        id: "pipe_3",
        name: "super-admin-web / main",
        repo: "nexora/super_admin_web",
        branch: "main",
        status: "failed",
        commitSha: "c0ffee1",
        triggeredBy: "github-actions",
        durationSec: 180,
        updatedAt: new Date(Date.now() - 40 * 60_000).toISOString(),
      },
      {
        id: "pipe_4",
        name: "config-flags / main",
        repo: "nexora/config",
        branch: "main",
        status: "queued",
        commitSha: "deadbee",
        triggeredBy: "ci@nexora.platform",
        durationSec: 0,
        updatedAt: new Date(Date.now() - 5 * 60_000).toISOString(),
      },
      {
        id: "pipe_5",
        name: "gateway / main",
        repo: "nexora/gateway",
        branch: "main",
        status: "passed",
        commitSha: "1122334",
        triggeredBy: "sre@nexora.platform",
        durationSec: 540,
        updatedAt: new Date(Date.now() - 8 * 3600_000).toISOString(),
      },
    ],
    deployments: [...mockDeployments],
    environments: [
      {
        id: "env_prod",
        name: "production",
        kind: "production",
        region: "eu-west-1",
        cluster: "prod-eu-west",
        activeVersion: "fleet",
        status: "healthy",
        secretCount: 86,
      },
      {
        id: "env_stg",
        name: "staging",
        kind: "staging",
        region: "eu-central-1",
        cluster: "staging-global",
        activeVersion: "fleet",
        status: "healthy",
        secretCount: 64,
      },
      {
        id: "env_dr",
        name: "dr-warm",
        kind: "dr",
        region: "eu-south-1",
        cluster: "dr-warm-eu",
        activeVersion: "fleet",
        status: "healthy",
        secretCount: 52,
      },
      {
        id: "env_sbx",
        name: "sandbox",
        kind: "sandbox",
        region: "eu-west-1",
        cluster: "staging-global",
        activeVersion: "fleet",
        status: "maintenance",
        secretCount: 28,
      },
    ],
    secrets: [
      {
        id: "sec_db_platform",
        name: "postgres/platform-primary",
        environment: "production",
        provider: "vault",
        version: 12,
        rotatedAt: new Date(Date.now() - 88 * 86400_000).toISOString(),
        lastAccessedAt: new Date(Date.now() - 2 * 3600_000).toISOString(),
        owners: "platform_sre",
      },
      {
        id: "sec_redis",
        name: "redis/session-auth",
        environment: "production",
        provider: "aws_sm",
        version: 7,
        rotatedAt: new Date(Date.now() - 40 * 86400_000).toISOString(),
        lastAccessedAt: new Date(Date.now() - 30 * 60_000).toISOString(),
        owners: "platform_sre",
      },
      {
        id: "sec_stripe",
        name: "billing/stripe-api",
        environment: "production",
        provider: "vault",
        version: 4,
        rotatedAt: new Date(Date.now() - 20 * 86400_000).toISOString(),
        lastAccessedAt: new Date(Date.now() - 6 * 3600_000).toISOString(),
        owners: "platform_finops",
      },
      {
        id: "sec_oidc",
        name: "identity/oidc-client",
        environment: "staging",
        provider: "gcp_sm",
        version: 9,
        rotatedAt: new Date(Date.now() - 15 * 86400_000).toISOString(),
        lastAccessedAt: new Date(Date.now() - 12 * 3600_000).toISOString(),
        owners: "platform_security",
      },
      {
        id: "sec_k8s_tls",
        name: "ingress/wildcard-tls",
        environment: "production",
        provider: "k8s",
        version: 3,
        rotatedAt: new Date(Date.now() - 60 * 86400_000).toISOString(),
        lastAccessedAt: new Date(Date.now() - 1 * 3600_000).toISOString(),
        owners: "platform_sre",
      },
    ],
    secretProposals: [...mockSecretProposals],
  };
}

export async function fetchDeployments(): Promise<DeploymentsSnapshot> {
  try {
    return await apiClient<DeploymentsSnapshot>(platformPath("/deployments"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockSnapshot();
    }
    throw err;
  }
}

export async function promoteCanary(input: {
  deploymentId: string;
  targetPct: number;
}): Promise<DeploymentRecord> {
  try {
    return await apiClient<DeploymentRecord>(
      platformPath(`/deployments/${input.deploymentId}/canary`),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockDeployments.findIndex((d) => d.id === input.deploymentId);
      if (idx < 0) throw new Error("Deployment not found");
      const next: DeploymentRecord = {
        ...mockDeployments[idx],
        canaryPct: input.targetPct,
        status: input.targetPct >= 100 ? "healthy" : "progressing",
        finishedAt:
          input.targetPct >= 100 ? new Date().toISOString() : null,
      };
      mockDeployments[idx] = next;
      return next;
    }
    throw err;
  }
}

export async function rollbackDeployment(input: {
  deploymentId: string;
}): Promise<DeploymentRecord> {
  try {
    return await apiClient<DeploymentRecord>(
      platformPath(`/deployments/${input.deploymentId}/rollback`),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockDeployments.findIndex((d) => d.id === input.deploymentId);
      if (idx < 0) throw new Error("Deployment not found");
      const cur = mockDeployments[idx];
      const next: DeploymentRecord = {
        ...cur,
        status: "rolled_back",
        version: cur.previousVersion,
        canaryPct: null,
        finishedAt: new Date().toISOString(),
      };
      mockDeployments[idx] = next;
      return next;
    }
    throw err;
  }
}

export async function proposeSecretRotate(input: {
  secretId: string;
  secretName: string;
  reason: string;
  requesterId: string;
  requesterEmail: string;
}): Promise<SecretRotateProposal> {
  try {
    return await apiClient<SecretRotateProposal>(
      platformPath("/deployments/secrets/dual-control"),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const proposal: SecretRotateProposal = {
        id: `prop_sec_${Date.now()}`,
        action: "secret_rotate",
        secretId: input.secretId,
        secretName: input.secretName,
        requesterId: input.requesterId,
        requesterEmail: input.requesterEmail,
        status: "pending",
        reason: input.reason,
        createdAt: new Date().toISOString(),
      };
      mockSecretProposals = [proposal, ...mockSecretProposals];
      return proposal;
    }
    throw err;
  }
}

export async function resolveSecretRotate(input: {
  proposalId: string;
  decision: "approved" | "rejected";
  approverId: string;
}): Promise<SecretRotateProposal> {
  try {
    return await apiClient<SecretRotateProposal>(
      platformPath(`/deployments/secrets/dual-control/${input.proposalId}`),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockSecretProposals.findIndex((p) => p.id === input.proposalId);
      if (idx < 0) throw new Error("Proposal not found");
      const next: SecretRotateProposal = {
        ...mockSecretProposals[idx],
        status: input.decision === "approved" ? "executed" : "rejected",
      };
      mockSecretProposals[idx] = next;
      return next;
    }
    throw err;
  }
}
