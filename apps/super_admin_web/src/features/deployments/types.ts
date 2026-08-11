import type { Id } from "@/shared/types/common";
import type { DualControlAction } from "@/shared/permissions/dual-control";
import type { SeriesPoint } from "@/shared/lib/charts";

export type DeployStrategy = "blue_green" | "canary" | "rolling";
export type PipelineStatus =
  | "queued"
  | "running"
  | "passed"
  | "failed"
  | "cancelled";
export type DeployStatus =
  | "idle"
  | "progressing"
  | "healthy"
  | "degraded"
  | "rolled_back";
export type EnvKind = "production" | "staging" | "dr" | "sandbox" | "preview";

export interface CicdPipeline {
  id: Id;
  name: string;
  repo: string;
  branch: string;
  status: PipelineStatus;
  commitSha: string;
  triggeredBy: string;
  durationSec: number;
  updatedAt: string;
}

export interface DeploymentRecord {
  id: Id;
  service: string;
  environment: string;
  strategy: DeployStrategy;
  version: string;
  previousVersion: string;
  status: DeployStatus;
  canaryPct: number | null;
  startedAt: string;
  finishedAt: string | null;
}

export interface DeployEnvironment {
  id: Id;
  name: string;
  kind: EnvKind;
  region: string;
  cluster: string;
  activeVersion: string;
  status: "healthy" | "degraded" | "maintenance";
  secretCount: number;
}

/** Metadata only — never includes secret values. */
export interface SecretMeta {
  id: Id;
  name: string;
  environment: string;
  provider: "vault" | "aws_sm" | "gcp_sm" | "k8s";
  version: number;
  rotatedAt: string;
  lastAccessedAt: string;
  owners: string;
}

export interface SecretRotateProposal {
  id: Id;
  action: Extract<DualControlAction, "secret_rotate">;
  secretId: Id;
  secretName: string;
  requesterId: Id;
  requesterEmail: string;
  status: "pending" | "approved" | "rejected" | "executed";
  reason: string;
  createdAt: string;
}

export interface DeploymentsSnapshot {
  generatedAt: string;
  kpis: {
    pipelinesRunning: number;
    deploys24h: number;
    failedPipelines: number;
    canariesActive: number;
    rollbacks24h: number;
    secretsDueRotation: number;
  };
  deployFrequency: SeriesPoint[];
  pipelines: CicdPipeline[];
  deployments: DeploymentRecord[];
  environments: DeployEnvironment[];
  secrets: SecretMeta[];
  secretProposals: SecretRotateProposal[];
}
