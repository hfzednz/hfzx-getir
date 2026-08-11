import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { ObservabilitySnapshot } from "./types";

function mockSnapshot(): ObservabilitySnapshot {
  const h = ["00", "04", "08", "12", "16", "20"];
  return {
    generatedAt: new Date().toISOString(),
    kpis: {
      firingAlerts: 7,
      openIncidents: 3,
      probesUpPct: 98.6,
      sloBreaches: 1,
      avgErrorBudgetPct: 72,
      traceIngestRps: 18400,
    },
    errorRateSeries: h.map((label, i) => ({
      label: `${label}:00`,
      value: 0.15 + i * 0.05 + (i === 3 ? 0.35 : 0),
    })),
    latencySeries: h.map((label, i) => ({
      label: `${label}:00`,
      value: 95 + i * 12 + (i === 4 ? 40 : 0),
    })),
    links: [
      {
        id: "l1",
        kind: "logs",
        name: "Loki — platform",
        provider: "Grafana Loki",
        url: "https://obs.nexora.global/explore?orgId=1&left=loki",
        description: "Structured logs for /platform namespace",
      },
      {
        id: "l2",
        kind: "metrics",
        name: "Prometheus — control plane",
        provider: "Prometheus",
        url: "https://obs.nexora.global/prometheus",
        description: "Cluster & service golden signals",
      },
      {
        id: "l3",
        kind: "traces",
        name: "Tempo — distributed traces",
        provider: "Grafana Tempo",
        url: "https://obs.nexora.global/explore?left=tempo",
        description: "End-to-end traces across BFF and services",
      },
      {
        id: "l4",
        kind: "dashboards",
        name: "Grafana — SRE overview",
        provider: "Grafana",
        url: "https://obs.nexora.global/d/sre-overview",
        description: "Platform SRE landing dashboard",
      },
    ],
    alerts: [
      {
        id: "a1",
        severity: "critical",
        title: "High Kafka lag order.events",
        service: "messaging",
        firedAt: new Date(Date.now() - 12 * 60000).toISOString(),
        status: "firing",
      },
      {
        id: "a2",
        severity: "warning",
        title: "APAC cluster CPU > 80%",
        service: "infra",
        firedAt: new Date(Date.now() - 28 * 60000).toISOString(),
        status: "acked",
      },
      {
        id: "a3",
        severity: "critical",
        title: "OpenSearch yellow cluster",
        service: "databases",
        firedAt: new Date(Date.now() - 45 * 60000).toISOString(),
        status: "firing",
      },
      {
        id: "a4",
        severity: "info",
        title: "Cert expiring < 21d",
        service: "infra",
        firedAt: new Date(Date.now() - 2 * 3600000).toISOString(),
        status: "firing",
      },
      {
        id: "a5",
        severity: "warning",
        title: "Gateway error rate shadow breach",
        service: "gateway",
        firedAt: new Date(Date.now() - 8 * 60000).toISOString(),
        status: "resolved",
      },
    ],
    incidents: [
      {
        id: "inc1",
        severity: "sev2",
        title: "Payments latency EU",
        status: "mitigating",
        region: "eu-west-1",
        openedAt: new Date(Date.now() - 90 * 60000).toISOString(),
        commander: "sre.oncall",
      },
      {
        id: "inc2",
        severity: "sev3",
        title: "CDN cache miss spike APAC",
        status: "investigating",
        region: "ap-southeast-1",
        openedAt: new Date(Date.now() - 40 * 60000).toISOString(),
        commander: "edge.oncall",
      },
      {
        id: "inc3",
        severity: "sev1",
        title: "Auth MFA provider outage",
        status: "resolved",
        region: "global",
        openedAt: new Date(Date.now() - 6 * 3600000).toISOString(),
        commander: "sec.oncall",
      },
    ],
    health: [
      {
        id: "h1",
        name: "bff-admin /platform",
        kind: "http",
        status: "up",
        latencyMs: 38,
        region: "eu-west-1",
      },
      {
        id: "h2",
        name: "identity-service",
        kind: "grpc",
        status: "up",
        latencyMs: 54,
        region: "eu-west-1",
      },
      {
        id: "h3",
        name: "gateway edge synth",
        kind: "synthetic",
        status: "degraded",
        latencyMs: 210,
        region: "ap-southeast-1",
      },
      {
        id: "h4",
        name: "kafka brokers",
        kind: "tcp",
        status: "up",
        latencyMs: 12,
        region: "eu-west-1",
      },
      {
        id: "h5",
        name: "postgres primary",
        kind: "tcp",
        status: "up",
        latencyMs: 4,
        region: "eu-west-1",
      },
    ],
    slos: [
      {
        id: "slo1",
        name: "API availability",
        sli: "successful requests / total",
        targetPct: 99.9,
        currentPct: 99.94,
        errorBudgetRemainingPct: 68,
        window: "30d",
        status: "healthy",
      },
      {
        id: "slo2",
        name: "Checkout latency p99",
        sli: "p99 < 400ms",
        targetPct: 99.0,
        currentPct: 98.2,
        errorBudgetRemainingPct: 12,
        window: "7d",
        status: "burn",
      },
      {
        id: "slo3",
        name: "Auth success rate",
        sli: "login success / attempts",
        targetPct: 99.5,
        currentPct: 97.8,
        errorBudgetRemainingPct: 0,
        window: "7d",
        status: "breached",
      },
      {
        id: "slo4",
        name: "Trace ingest freshness",
        sli: "ingest lag < 30s",
        targetPct: 99.0,
        currentPct: 99.6,
        errorBudgetRemainingPct: 82,
        window: "30d",
        status: "healthy",
      },
    ],
    slas: [
      {
        id: "sla1",
        name: "Enterprise platform uptime",
        scope: "All enterprise tenants",
        uptimeTargetPct: 99.95,
        currentPct: 99.97,
        creditsAtRisk: false,
      },
      {
        id: "sla2",
        name: "API gateway latency",
        scope: "Partner API keys",
        uptimeTargetPct: 99.5,
        currentPct: 99.1,
        creditsAtRisk: true,
      },
      {
        id: "sla3",
        name: "Data residency EU",
        scope: "EU regulated tenants",
        uptimeTargetPct: 100,
        currentPct: 100,
        creditsAtRisk: false,
      },
    ],
  };
}

export async function fetchObservabilitySnapshot(): Promise<ObservabilitySnapshot> {
  try {
    return await apiClient<ObservabilitySnapshot>(
      platformPath("/observability"),
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await new Promise((r) => setTimeout(r, 200));
      return mockSnapshot();
    }
    throw err;
  }
}
