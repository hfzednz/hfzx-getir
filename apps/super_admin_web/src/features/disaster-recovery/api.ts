import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  DisasterRecoverySnapshot,
  DrFailoverProposal,
  DrSimulation,
  RestoreJob,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

let mockProposals: DrFailoverProposal[] = [
  {
    id: "prop_dr1",
    action: "dr_failover",
    fromRegion: "eu-west-1",
    toRegion: "eu-south-1",
    requesterId: "usr_platform_sre_demo",
    requesterEmail: "sre@nexora.platform",
    status: "pending",
    reason: "Primary AZ capacity exhaustion — warm DR promote",
    createdAt: new Date(Date.now() - 45 * 60_000).toISOString(),
  },
];

let mockRestores: RestoreJob[] = [
  {
    id: "rst_1",
    backupId: "bak_pg_daily",
    backupLabel: "Postgres daily full",
    targetEnv: "dr-sandbox",
    status: "completed",
    progressPct: 100,
    requestedBy: "sre@nexora.platform",
    startedAt: new Date(Date.now() - 8 * 3600_000).toISOString(),
    finishedAt: new Date(Date.now() - 7 * 3600_000).toISOString(),
  },
  {
    id: "rst_2",
    backupId: "bak_redis_hourly",
    backupLabel: "Redis RDB hourly",
    targetEnv: "staging",
    status: "running",
    progressPct: 48,
    requestedBy: "owner@nexora.platform",
    startedAt: new Date(Date.now() - 20 * 60_000).toISOString(),
    finishedAt: null,
  },
];

let mockSimulations: DrSimulation[] = [
  {
    id: "sim_1",
    name: "EU primary loss",
    regionFrom: "eu-west-1",
    regionTo: "eu-south-1",
    status: "completed",
    blastRadius: "platform + shared tenants",
    startedAt: new Date(Date.now() - 14 * 86400_000).toISOString(),
    notes: "RTO 18m · RPO 4m observed",
  },
  {
    id: "sim_2",
    name: "APAC Kafka broker loss",
    regionFrom: "ap-southeast-1",
    regionTo: "ap-northeast-1",
    status: "draft",
    blastRadius: "messaging only",
    startedAt: null,
    notes: "Scheduled for next game-day",
  },
];

function mockSnapshot(): DisasterRecoverySnapshot {
  const hours = ["00", "04", "08", "12", "16", "20"];
  return {
    generatedAt: new Date().toISOString(),
    kpis: {
      backupsOk: 18,
      backupsFailed: 1,
      replicationLinks: 6,
      laggingLinks: 1,
      pendingFailovers: mockProposals.filter((p) => p.status === "pending")
        .length,
      lastSuccessfulTestDaysAgo: 12,
    },
    rpoSeries: hours.map((label, i) => ({
      label: `${label}:00`,
      value: 3 + (i % 3) + (i === 4 ? 4 : 0),
    })),
    backups: [
      {
        id: "bak_pg_daily",
        label: "Postgres daily full",
        scope: "platform-ledger",
        region: "eu-west-1",
        status: "ok",
        sizeGb: 842,
        retentionDays: 35,
        takenAt: new Date(Date.now() - 5 * 3600_000).toISOString(),
      },
      {
        id: "bak_pg_wal",
        label: "Postgres WAL archive",
        scope: "platform-ledger",
        region: "eu-west-1",
        status: "ok",
        sizeGb: 126,
        retentionDays: 14,
        takenAt: new Date(Date.now() - 20 * 60_000).toISOString(),
      },
      {
        id: "bak_redis_hourly",
        label: "Redis RDB hourly",
        scope: "session-cache",
        region: "eu-west-1",
        status: "running",
        sizeGb: 18.4,
        retentionDays: 7,
        takenAt: new Date().toISOString(),
      },
      {
        id: "bak_os_snap",
        label: "OpenSearch snapshot",
        scope: "search-catalog",
        region: "us-east-1",
        status: "stale",
        sizeGb: 410,
        retentionDays: 21,
        takenAt: new Date(Date.now() - 40 * 3600_000).toISOString(),
      },
      {
        id: "bak_ch_daily",
        label: "ClickHouse daily",
        scope: "analytics",
        region: "eu-central-1",
        status: "failed",
        sizeGb: 0,
        retentionDays: 30,
        takenAt: new Date(Date.now() - 26 * 3600_000).toISOString(),
      },
      {
        id: "bak_obj_geo",
        label: "Object store geo copy",
        scope: "media",
        region: "global",
        status: "ok",
        sizeGb: 120000,
        retentionDays: 90,
        takenAt: new Date(Date.now() - 2 * 3600_000).toISOString(),
      },
    ],
    replication: [
      {
        id: "rep_1",
        sourceRegion: "eu-west-1",
        targetRegion: "eu-south-1",
        mode: "async",
        lagSeconds: 12,
        status: "synced",
        throughputMBps: 84,
      },
      {
        id: "rep_2",
        sourceRegion: "eu-west-1",
        targetRegion: "us-east-1",
        mode: "async",
        lagSeconds: 95,
        status: "lagging",
        throughputMBps: 42,
      },
      {
        id: "rep_3",
        sourceRegion: "us-east-1",
        targetRegion: "us-west-2",
        mode: "semi-sync",
        lagSeconds: 4,
        status: "synced",
        throughputMBps: 110,
      },
      {
        id: "rep_4",
        sourceRegion: "ap-southeast-1",
        targetRegion: "ap-northeast-1",
        mode: "async",
        lagSeconds: 28,
        status: "synced",
        throughputMBps: 56,
      },
      {
        id: "rep_5",
        sourceRegion: "eu-central-1",
        targetRegion: "eu-west-1",
        mode: "sync",
        lagSeconds: 0,
        status: "synced",
        throughputMBps: 22,
      },
      {
        id: "rep_6",
        sourceRegion: "sa-east-1",
        targetRegion: "us-east-1",
        mode: "async",
        lagSeconds: 0,
        status: "broken",
        throughputMBps: 0,
      },
    ],
    regions: [
      {
        id: "reg_euw1",
        name: "EU West (primary)",
        code: "eu-west-1",
        role: "primary",
        rpoMinutes: 5,
        rtoMinutes: 15,
        healthy: true,
        lastFailoverAt: null,
      },
      {
        id: "reg_eus1",
        name: "EU South (warm DR)",
        code: "eu-south-1",
        role: "warm",
        rpoMinutes: 5,
        rtoMinutes: 20,
        healthy: true,
        lastFailoverAt: new Date(Date.now() - 90 * 86400_000).toISOString(),
      },
      {
        id: "reg_use1",
        name: "US East",
        code: "us-east-1",
        role: "standby",
        rpoMinutes: 10,
        rtoMinutes: 30,
        healthy: true,
        lastFailoverAt: null,
      },
      {
        id: "reg_apse1",
        name: "APAC SE",
        code: "ap-southeast-1",
        role: "standby",
        rpoMinutes: 10,
        rtoMinutes: 25,
        healthy: true,
        lastFailoverAt: null,
      },
      {
        id: "reg_sae1",
        name: "SA East (cold)",
        code: "sa-east-1",
        role: "cold",
        rpoMinutes: 60,
        rtoMinutes: 240,
        healthy: false,
        lastFailoverAt: null,
      },
    ],
    restores: [...mockRestores],
    tests: [
      {
        id: "test_1",
        name: "Quarterly full restore",
        scenario: "Restore platform ledger to isolated DR sandbox",
        status: "passed",
        lastRunAt: new Date(Date.now() - 12 * 86400_000).toISOString(),
        nextRunAt: new Date(Date.now() + 78 * 86400_000).toISOString(),
        owner: "platform_sre",
      },
      {
        id: "test_2",
        name: "Failover tabletop",
        scenario: "Dual-control promote warm region without traffic",
        status: "scheduled",
        lastRunAt: new Date(Date.now() - 100 * 86400_000).toISOString(),
        nextRunAt: new Date(Date.now() + 5 * 86400_000).toISOString(),
        owner: "platform_owner",
      },
      {
        id: "test_3",
        name: "Object store integrity",
        scenario: "Checksum sample of geo-replicated media",
        status: "running",
        lastRunAt: new Date().toISOString(),
        nextRunAt: new Date(Date.now() + 30 * 86400_000).toISOString(),
        owner: "platform_sre",
      },
      {
        id: "test_4",
        name: "ClickHouse PITR drill",
        scenario: "Point-in-time recover analytics shard",
        status: "failed",
        lastRunAt: new Date(Date.now() - 3 * 86400_000).toISOString(),
        nextRunAt: new Date(Date.now() + 4 * 86400_000).toISOString(),
        owner: "platform_sre",
      },
    ],
    simulations: [...mockSimulations],
    proposals: [...mockProposals],
  };
}

export async function fetchDisasterRecovery(): Promise<DisasterRecoverySnapshot> {
  try {
    return await apiClient<DisasterRecoverySnapshot>(
      platformPath("/disaster-recovery"),
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockSnapshot();
    }
    throw err;
  }
}

export async function proposeDrFailover(input: {
  fromRegion: string;
  toRegion: string;
  reason: string;
  requesterId: string;
  requesterEmail: string;
}): Promise<DrFailoverProposal> {
  try {
    return await apiClient<DrFailoverProposal>(
      platformPath("/disaster-recovery/dual-control"),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const proposal: DrFailoverProposal = {
        id: `prop_dr_${Date.now()}`,
        action: "dr_failover",
        fromRegion: input.fromRegion,
        toRegion: input.toRegion,
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

export async function resolveDrFailover(input: {
  proposalId: string;
  decision: "approved" | "rejected";
  approverId: string;
}): Promise<DrFailoverProposal> {
  try {
    return await apiClient<DrFailoverProposal>(
      platformPath(`/disaster-recovery/dual-control/${input.proposalId}`),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockProposals.findIndex((p) => p.id === input.proposalId);
      if (idx < 0) throw new Error("Proposal not found");
      const next: DrFailoverProposal = {
        ...mockProposals[idx],
        status: input.decision === "approved" ? "executed" : "rejected",
      };
      mockProposals[idx] = next;
      return next;
    }
    throw err;
  }
}

export async function startRestore(input: {
  backupId: string;
  backupLabel: string;
  targetEnv: string;
  requestedBy: string;
}): Promise<RestoreJob> {
  try {
    return await apiClient<RestoreJob>(
      platformPath("/disaster-recovery/restores"),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const job: RestoreJob = {
        id: `rst_${Date.now()}`,
        backupId: input.backupId,
        backupLabel: input.backupLabel,
        targetEnv: input.targetEnv,
        status: "queued",
        progressPct: 0,
        requestedBy: input.requestedBy,
        startedAt: new Date().toISOString(),
        finishedAt: null,
      };
      mockRestores = [job, ...mockRestores];
      return job;
    }
    throw err;
  }
}

export async function runSimulation(input: {
  name: string;
  regionFrom: string;
  regionTo: string;
  blastRadius: string;
}): Promise<DrSimulation> {
  try {
    return await apiClient<DrSimulation>(
      platformPath("/disaster-recovery/simulations"),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const sim: DrSimulation = {
        id: `sim_${Date.now()}`,
        name: input.name,
        regionFrom: input.regionFrom,
        regionTo: input.regionTo,
        status: "running",
        blastRadius: input.blastRadius,
        startedAt: new Date().toISOString(),
        notes: "Simulation started (mock)",
      };
      mockSimulations = [sim, ...mockSimulations];
      return sim;
    }
    throw err;
  }
}
