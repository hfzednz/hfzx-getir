import type { SeriesPoint } from "@/shared/lib/charts";

export type AiModelKind =
  | "recommendation"
  | "forecast"
  | "pricing"
  | "search"
  | "ocr"
  | "chatbot";

export interface AiModel {
  id: string;
  name: string;
  kind: AiModelKind;
  version: string;
  framework: string;
  status: "serving" | "staging" | "training" | "deprecated" | "failed";
  accuracyPct: number | null;
  latencyP99Ms: number | null;
  requestsPerMin: number;
  owner: string;
}

export interface TrainingJob {
  id: string;
  modelName: string;
  kind: AiModelKind;
  status: "queued" | "running" | "succeeded" | "failed" | "cancelled";
  progressPct: number;
  gpus: number;
  startedAt: string | null;
  etaMin: number | null;
  dataset: string;
}

export interface InferenceServer {
  id: string;
  name: string;
  models: string;
  region: string;
  replicas: number;
  status: "healthy" | "degraded" | "scaling" | "down";
  qps: number;
  gpuUtilPct: number;
  runtime: "triton" | "torchserve" | "vllm" | "custom";
}

export interface GpuNode {
  id: string;
  hostname: string;
  region: string;
  gpuModel: string;
  gpuCount: number;
  utilPct: number;
  memUsedPct: number;
  tempC: number;
  status: "healthy" | "throttled" | "down" | "maintenance";
}

export interface AiPlatformSnapshot {
  generatedAt: string;
  kpis: {
    modelsServing: number;
    trainingJobs: number;
    inferenceQps: number;
    gpuUtilPct: number;
    tokensPerMin: number;
    failedJobs24h: number;
  };
  inferenceSeries: SeriesPoint[];
  gpuUtilSeries: SeriesPoint[];
  models: AiModel[];
  trainingJobs: TrainingJob[];
  inferenceServers: InferenceServer[];
  gpuNodes: GpuNode[];
}
