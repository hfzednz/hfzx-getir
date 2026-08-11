import type { SeriesPoint } from "@/shared/lib/charts";

export type BrokerKind = "kafka" | "rabbitmq";

export interface MessagingTopic {
  id: string;
  name: string;
  broker: BrokerKind;
  partitions: number;
  replicas: number;
  messagesPerSec: number;
  lag: number;
  retentionHours: number;
  status: "healthy" | "degraded" | "offline";
}

export interface MessagingQueue {
  id: string;
  name: string;
  broker: BrokerKind;
  depth: number;
  consumers: number;
  ackRate: number;
  status: "ok" | "warn" | "critical";
  dlqBound: boolean;
}

export interface MessagingConsumer {
  id: string;
  group: string;
  topicOrQueue: string;
  broker: BrokerKind;
  members: number;
  lag: number;
  status: "stable" | "rebalancing" | "empty" | "dead";
  region: string;
}

export interface DlqEntry {
  id: string;
  source: string;
  broker: BrokerKind;
  depth: number;
  oldestAgeMin: number;
  lastError: string;
  status: "idle" | "draining" | "backed_up";
}

export interface RetryPolicy {
  id: string;
  name: string;
  appliesTo: string;
  maxAttempts: number;
  backoff: "fixed" | "exponential" | "linear";
  initialDelayMs: number;
  maxDelayMs: number;
  dlqOnExhaust: boolean;
}

export interface MessagingSnapshot {
  generatedAt: string;
  kpis: {
    topics: number;
    queues: number;
    consumerGroups: number;
    totalLag: number;
    dlqDepth: number;
    throughputMsgPerSec: number;
  };
  throughputSeries: SeriesPoint[];
  lagSeries: SeriesPoint[];
  topics: MessagingTopic[];
  queues: MessagingQueue[];
  consumers: MessagingConsumer[];
  dlq: DlqEntry[];
  retryPolicies: RetryPolicy[];
}
