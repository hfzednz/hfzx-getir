import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { MessagingSnapshot } from "./types";

function mockSnapshot(): MessagingSnapshot {
  const h = ["00", "04", "08", "12", "16", "20"];
  return {
    generatedAt: new Date().toISOString(),
    kpis: {
      topics: 86,
      queues: 42,
      consumerGroups: 118,
      totalLag: 184200,
      dlqDepth: 1240,
      throughputMsgPerSec: 92000,
    },
    throughputSeries: h.map((label, i) => ({
      label: `${label}:00`,
      value: 42000 + i * 9000 + (i % 2) * 4000,
    })),
    lagSeries: h.map((label, i) => ({
      label: `${label}:00`,
      value: 42000 + i * 22000 + (i === 4 ? 80000 : 0),
    })),
    topics: [
      {
        id: "t1",
        name: "order.events",
        broker: "kafka",
        partitions: 48,
        replicas: 3,
        messagesPerSec: 28000,
        lag: 120400,
        retentionHours: 168,
        status: "degraded",
      },
      {
        id: "t2",
        name: "inventory.mutations",
        broker: "kafka",
        partitions: 24,
        replicas: 3,
        messagesPerSec: 12000,
        lag: 8400,
        retentionHours: 72,
        status: "healthy",
      },
      {
        id: "t3",
        name: "platform.audit",
        broker: "kafka",
        partitions: 12,
        replicas: 3,
        messagesPerSec: 2200,
        lag: 120,
        retentionHours: 720,
        status: "healthy",
      },
      {
        id: "t4",
        name: "ai.inference.requests",
        broker: "kafka",
        partitions: 16,
        replicas: 3,
        messagesPerSec: 6400,
        lag: 2100,
        retentionHours: 24,
        status: "healthy",
      },
    ],
    queues: [
      {
        id: "q1",
        name: "notifications.push",
        broker: "rabbitmq",
        depth: 4200,
        consumers: 12,
        ackRate: 980,
        status: "ok",
        dlqBound: true,
      },
      {
        id: "q2",
        name: "billing.settlements",
        broker: "rabbitmq",
        depth: 180,
        consumers: 4,
        ackRate: 42,
        status: "ok",
        dlqBound: true,
      },
      {
        id: "q3",
        name: "search.reindex",
        broker: "rabbitmq",
        depth: 28400,
        consumers: 2,
        ackRate: 18,
        status: "critical",
        dlqBound: true,
      },
      {
        id: "q4",
        name: "webhooks.outbound",
        broker: "rabbitmq",
        depth: 3200,
        consumers: 8,
        ackRate: 210,
        status: "warn",
        dlqBound: true,
      },
    ],
    consumers: [
      {
        id: "cg1",
        group: "order-projector",
        topicOrQueue: "order.events",
        broker: "kafka",
        members: 12,
        lag: 98400,
        status: "stable",
        region: "eu-west-1",
      },
      {
        id: "cg2",
        group: "fraud-scorer",
        topicOrQueue: "order.events",
        broker: "kafka",
        members: 6,
        lag: 22000,
        status: "rebalancing",
        region: "eu-west-1",
      },
      {
        id: "cg3",
        group: "notif-workers",
        topicOrQueue: "notifications.push",
        broker: "rabbitmq",
        members: 12,
        lag: 4200,
        status: "stable",
        region: "multi",
      },
      {
        id: "cg4",
        group: "search-reindex",
        topicOrQueue: "search.reindex",
        broker: "rabbitmq",
        members: 2,
        lag: 28400,
        status: "stable",
        region: "us-east-1",
      },
    ],
    dlq: [
      {
        id: "dlq1",
        source: "order.events → fraud-scorer",
        broker: "kafka",
        depth: 840,
        oldestAgeMin: 42,
        lastError: "downstream timeout 504",
        status: "backed_up",
      },
      {
        id: "dlq2",
        source: "webhooks.outbound",
        broker: "rabbitmq",
        depth: 320,
        oldestAgeMin: 18,
        lastError: "partner 429 Too Many Requests",
        status: "draining",
      },
      {
        id: "dlq3",
        source: "billing.settlements",
        broker: "rabbitmq",
        depth: 80,
        oldestAgeMin: 6,
        lastError: "ledger lock contention",
        status: "idle",
      },
    ],
    retryPolicies: [
      {
        id: "rp1",
        name: "default-exponential",
        appliesTo: "*",
        maxAttempts: 5,
        backoff: "exponential",
        initialDelayMs: 500,
        maxDelayMs: 60000,
        dlqOnExhaust: true,
      },
      {
        id: "rp2",
        name: "webhooks-aggressive",
        appliesTo: "webhooks.outbound",
        maxAttempts: 8,
        backoff: "exponential",
        initialDelayMs: 1000,
        maxDelayMs: 300000,
        dlqOnExhaust: true,
      },
      {
        id: "rp3",
        name: "billing-fixed",
        appliesTo: "billing.settlements",
        maxAttempts: 3,
        backoff: "fixed",
        initialDelayMs: 5000,
        maxDelayMs: 5000,
        dlqOnExhaust: true,
      },
      {
        id: "rp4",
        name: "search-linear",
        appliesTo: "search.reindex",
        maxAttempts: 10,
        backoff: "linear",
        initialDelayMs: 2000,
        maxDelayMs: 30000,
        dlqOnExhaust: false,
      },
    ],
  };
}

export async function fetchMessagingSnapshot(): Promise<MessagingSnapshot> {
  try {
    return await apiClient<MessagingSnapshot>(platformPath("/messaging"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await new Promise((r) => setTimeout(r, 200));
      return mockSnapshot();
    }
    throw err;
  }
}
