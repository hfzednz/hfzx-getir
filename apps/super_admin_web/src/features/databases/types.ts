import type { SeriesPoint } from "@/shared/lib/charts";

export type DbEngine = "postgres" | "redis" | "opensearch" | "clickhouse";
export type DbStatus = "healthy" | "degraded" | "failover" | "down" | "maintenance";

export interface DbCluster {
  id: string;
  name: string;
  engine: DbEngine;
  version: string;
  region: string;
  role: "primary" | "replica" | "cluster";
  status: DbStatus;
  connections: number;
  maxConnections: number;
  storageUsedPct: number;
  lagMs: number | null;
  qps: number;
}

export interface DbBackup {
  id: string;
  clusterId: string;
  clusterName: string;
  engine: DbEngine;
  type: "full" | "incremental" | "snapshot";
  sizeGb: number;
  status: "completed" | "running" | "failed" | "scheduled";
  startedAt: string;
  finishedAt: string | null;
  retentionDays: number;
}

export interface ReplicationLink {
  id: string;
  source: string;
  target: string;
  engine: DbEngine;
  lagMs: number;
  status: "streaming" | "catching_up" | "broken" | "paused";
  mode: "async" | "sync" | "semi-sync";
}

export interface FailoverEvent {
  id: string;
  clusterName: string;
  engine: DbEngine;
  fromNode: string;
  toNode: string;
  reason: string;
  status: "completed" | "in_progress" | "rolled_back";
  at: string;
  dualControl: boolean;
}

export interface MigrationStatus {
  id: string;
  name: string;
  engine: DbEngine;
  fromVersion: string;
  toVersion: string;
  progressPct: number;
  status: "pending" | "running" | "completed" | "failed" | "rolled_back";
  startedAt: string | null;
}

export interface DatabasesSnapshot {
  generatedAt: string;
  kpis: {
    clusters: number;
    unhealthy: number;
    backups24h: number;
    avgLagMs: number;
    migrationsActive: number;
    storagePct: number;
  };
  qpsSeries: SeriesPoint[];
  lagSeries: SeriesPoint[];
  clusters: DbCluster[];
  backups: DbBackup[];
  replication: ReplicationLink[];
  failovers: FailoverEvent[];
  migrations: MigrationStatus[];
}
