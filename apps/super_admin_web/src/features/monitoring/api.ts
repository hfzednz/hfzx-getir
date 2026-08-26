import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { MonitoringSnapshot } from "./types";

function delay(ms = 200): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function mockSnapshot(): MonitoringSnapshot {
  const hours = ["00", "04", "08", "12", "16", "20"];
  return {
    generatedAt: new Date().toISOString(),
    kpis: {
      cpuAvgPct: 58,
      memAvgPct: 71,
      diskAvgPct: 64,
      apiP95Ms: 142,
      queueDepth: 12840,
      wsConnections: 84210,
      redisHitPct: 96.4,
      opensearchLagMs: 38,
      k8sPodsNotReady: 4,
      cloudAlerts: 2,
    },
    cpuSeries: hours.map((label, i) => ({
      label: `${label}:00`,
      value: 48 + i * 4 + (i % 2) * 3,
    })),
    memSeries: hours.map((label, i) => ({
      label: `${label}:00`,
      value: 62 + i * 2,
    })),
    latencySeries: hours.map((label, i) => ({
      label: `${label}:00`,
      value: 110 + i * 8 + (i === 4 ? 40 : 0),
    })),
    resources: [
      {
        id: "r1",
        name: "prod-eu-west nodes",
        category: "compute",
        metric: "CPU",
        value: 58,
        unit: "%",
        threshold: 85,
        status: "healthy",
        region: "eu-west-1",
      },
      {
        id: "r2",
        name: "prod-eu-west nodes",
        category: "compute",
        metric: "Memory",
        value: 71,
        unit: "%",
        threshold: 90,
        status: "healthy",
        region: "eu-west-1",
      },
      {
        id: "r3",
        name: "pg-prod-eu volume",
        category: "disk",
        metric: "Disk used",
        value: 62,
        unit: "%",
        threshold: 80,
        status: "healthy",
        region: "eu-west-1",
      },
      {
        id: "r4",
        name: "clickhouse-apac",
        category: "disk",
        metric: "Disk used",
        value: 84,
        unit: "%",
        threshold: 80,
        status: "degraded",
        region: "ap-southeast-1",
      },
      {
        id: "r5",
        name: "platform-ledger",
        category: "database",
        metric: "Connections",
        value: 340,
        unit: "conn",
        threshold: 800,
        status: "healthy",
        region: "eu-west-1",
      },
      {
        id: "r6",
        name: "bff-admin p95",
        category: "api",
        metric: "Latency",
        value: 142,
        unit: "ms",
        threshold: 250,
        status: "healthy",
        region: "global",
      },
      {
        id: "r7",
        name: "order-events",
        category: "queue",
        metric: "Depth",
        value: 12840,
        unit: "msgs",
        threshold: 50000,
        status: "healthy",
        region: "eu-west-1",
      },
      {
        id: "r8",
        name: "realtime-gateway",
        category: "websocket",
        metric: "Connections",
        value: 84210,
        unit: "conn",
        threshold: 150000,
        status: "healthy",
        region: "global",
      },
      {
        id: "r9",
        name: "session-cache",
        category: "redis",
        metric: "Hit rate",
        value: 96.4,
        unit: "%",
        threshold: 90,
        status: "healthy",
        region: "eu-west-1",
      },
      {
        id: "r10",
        name: "catalog-search",
        category: "opensearch",
        metric: "Indexing lag",
        value: 38,
        unit: "ms",
        threshold: 200,
        status: "healthy",
        region: "us-east-1",
      },
      {
        id: "r11",
        name: "ai-inference ns",
        category: "kubernetes",
        metric: "Pods not ready",
        value: 4,
        unit: "pods",
        threshold: 2,
        status: "degraded",
        region: "ap-southeast-1",
      },
      {
        id: "r12",
        name: "AWS EU account",
        category: "cloud",
        metric: "Service health",
        value: 2,
        unit: "alerts",
        threshold: 0,
        status: "critical",
        region: "eu-west-1",
      },
    ],
  };
}

export async function fetchMonitoring(): Promise<MonitoringSnapshot> {
  try {
    return await apiClient<MonitoringSnapshot>(platformPath("/monitoring"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockSnapshot();
    }
    throw err;
  }
}
