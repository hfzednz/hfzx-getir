import type { MonitoringSnapshot } from "./types";

/** Mock monitoring — replaced by GET /admin/monitoring/health when BFF is live. */
export async function fetchMonitoringSnapshot(): Promise<MonitoringSnapshot> {
  await new Promise((r) => setTimeout(r, 210));
  const mins = ["-25m", "-20m", "-15m", "-10m", "-5m", "now"];

  return {
    generatedAt: new Date().toISOString(),
    services: [
      { id: "h1", name: "bff-admin", status: "healthy", latencyMs: 42 },
      { id: "h2", name: "order-service", status: "healthy", latencyMs: 68 },
      { id: "h3", name: "courier-gateway", status: "degraded", latencyMs: 210 },
      { id: "h4", name: "payments", status: "healthy", latencyMs: 95 },
      { id: "h5", name: "inventory", status: "healthy", latencyMs: 74 },
      { id: "h6", name: "realtime-gw", status: "healthy", latencyMs: 38 },
    ],
    apiLatency: mins.map((m, i) => ({
      label: m,
      value: 55 + i * 8 + (i === 4 ? 40 : 0),
    })),
    serverLoad: mins.map((m, i) => ({
      label: m,
      value: 0.35 + i * 0.08,
    })),
    cpu: mins.map((m, i) => ({
      label: m,
      value: 32 + i * 6 + (i % 2) * 4,
    })),
    memory: mins.map((m, i) => ({
      label: m,
      value: 58 + i * 2.5,
    })),
    storagePct: 67,
    dbConnections: 184,
    dbSlowQueries: 3,
    queues: [
      {
        id: "q1",
        name: "dispatch.assign",
        depth: 42,
        lagSec: 1.2,
        status: "ok",
      },
      {
        id: "q2",
        name: "notifications.push",
        depth: 310,
        lagSec: 8.4,
        status: "warn",
      },
      {
        id: "q3",
        name: "finance.settlement",
        depth: 12,
        lagSec: 0.4,
        status: "ok",
      },
      {
        id: "q4",
        name: "fraud.score",
        depth: 88,
        lagSec: 14,
        status: "critical",
      },
    ],
    websocket: {
      status: "connected",
      clients: 1260,
      msgPerSec: 420,
    },
  };
}
