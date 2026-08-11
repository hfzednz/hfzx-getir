import type { Id } from "@/shared/types/common";
import type { DualControlAction } from "@/shared/permissions/dual-control";
import type { SeriesPoint } from "@/shared/lib/charts";

export type BackupStatus = "ok" | "running" | "stale" | "failed";
export type ReplicationLag = "synced" | "lagging" | "broken";
export type RegionRole = "primary" | "standby" | "warm" | "cold";
export type RestoreStatus =
  | "queued"
  | "running"
  | "completed"
  | "failed"
  | "cancelled";
export type TestStatus = "scheduled" | "running" | "passed" | "failed";
export type SimulationStatus = "draft" | "running" | "completed" | "aborted";

export interface DrBackup {
  id: Id;
  label: string;
  scope: string;
  region: string;
  status: BackupStatus;
  sizeGb: number;
  retentionDays: number;
  takenAt: string;
}

export interface GeoReplicationLink {
  id: Id;
  sourceRegion: string;
  targetRegion: string;
  mode: "async" | "sync" | "semi-sync";
  lagSeconds: number;
  status: ReplicationLag;
  throughputMBps: number;
}

export interface DrRegion {
  id: Id;
  name: string;
  code: string;
  role: RegionRole;
  rpoMinutes: number;
  rtoMinutes: number;
  healthy: boolean;
  lastFailoverAt: string | null;
}

export interface RestoreJob {
  id: Id;
  backupId: Id;
  backupLabel: string;
  targetEnv: string;
  status: RestoreStatus;
  progressPct: number;
  requestedBy: string;
  startedAt: string;
  finishedAt: string | null;
}

export interface RecoveryTest {
  id: Id;
  name: string;
  scenario: string;
  status: TestStatus;
  lastRunAt: string | null;
  nextRunAt: string;
  owner: string;
}

export interface DrSimulation {
  id: Id;
  name: string;
  regionFrom: string;
  regionTo: string;
  status: SimulationStatus;
  blastRadius: string;
  startedAt: string | null;
  notes: string;
}

export interface DrFailoverProposal {
  id: Id;
  action: Extract<DualControlAction, "dr_failover">;
  fromRegion: string;
  toRegion: string;
  requesterId: Id;
  requesterEmail: string;
  status: "pending" | "approved" | "rejected" | "executed";
  reason: string;
  createdAt: string;
}

export interface DisasterRecoverySnapshot {
  generatedAt: string;
  kpis: {
    backupsOk: number;
    backupsFailed: number;
    replicationLinks: number;
    laggingLinks: number;
    pendingFailovers: number;
    lastSuccessfulTestDaysAgo: number;
  };
  rpoSeries: SeriesPoint[];
  backups: DrBackup[];
  replication: GeoReplicationLink[];
  regions: DrRegion[];
  restores: RestoreJob[];
  tests: RecoveryTest[];
  simulations: DrSimulation[];
  proposals: DrFailoverProposal[];
}
