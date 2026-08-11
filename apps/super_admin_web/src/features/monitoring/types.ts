import type { SeriesPoint } from "@/shared/lib/charts";

export type ResourceHealth = "healthy" | "degraded" | "critical" | "unknown";

export interface MonitorResource {
  id: string;
  name: string;
  category:
    | "compute"
    | "disk"
    | "database"
    | "api"
    | "queue"
    | "websocket"
    | "redis"
    | "opensearch"
    | "kubernetes"
    | "cloud";
  metric: string;
  value: number;
  unit: string;
  threshold: number;
  status: ResourceHealth;
  region: string;
}

export interface MonitoringSnapshot {
  generatedAt: string;
  kpis: {
    cpuAvgPct: number;
    memAvgPct: number;
    diskAvgPct: number;
    apiP95Ms: number;
    queueDepth: number;
    wsConnections: number;
    redisHitPct: number;
    opensearchLagMs: number;
    k8sPodsNotReady: number;
    cloudAlerts: number;
  };
  cpuSeries: SeriesPoint[];
  memSeries: SeriesPoint[];
  latencySeries: SeriesPoint[];
  resources: MonitorResource[];
}
