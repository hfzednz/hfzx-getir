export interface HealthService {
  id: string;
  name: string;
  status: "healthy" | "degraded" | "down";
  latencyMs: number;
}

export interface SeriesPoint {
  label: string;
  value: number;
}

export interface QueueStat {
  id: string;
  name: string;
  depth: number;
  lagSec: number;
  status: "ok" | "warn" | "critical";
}

export interface MonitoringSnapshot {
  generatedAt: string;
  services: HealthService[];
  apiLatency: SeriesPoint[];
  serverLoad: SeriesPoint[];
  cpu: SeriesPoint[];
  memory: SeriesPoint[];
  storagePct: number;
  dbConnections: number;
  dbSlowQueries: number;
  queues: QueueStat[];
  websocket: {
    status: "connected" | "degraded" | "disconnected";
    clients: number;
    msgPerSec: number;
  };
}
