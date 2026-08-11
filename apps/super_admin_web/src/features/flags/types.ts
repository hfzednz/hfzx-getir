import type { Id } from "@/shared/types/common";
import type { DualControlAction } from "@/shared/permissions/dual-control";

export type FlagScope = "global" | "country" | "company" | "user";
export type FlagKind = "feature" | "kill_switch" | "ab_test" | "scheduled";
export type FlagStatus = "enabled" | "disabled" | "scheduled" | "rolling_back";

export interface FlagVariant {
  key: string;
  weightPct: number;
}

export interface FeatureFlag {
  id: Id;
  key: string;
  name: string;
  description: string;
  scope: FlagScope;
  scopeTarget: string | null;
  kind: FlagKind;
  status: FlagStatus;
  enabled: boolean;
  rolloutPct: number;
  variants: FlagVariant[];
  scheduledAt: string | null;
  expiresAt: string | null;
  updatedAt: string;
  updatedBy: string;
}

export interface KillSwitchProposal {
  id: Id;
  action: Extract<DualControlAction, "kill_switch">;
  flagId: Id;
  flagKey: string;
  flagName: string;
  targetEnabled: boolean;
  requesterId: Id;
  requesterEmail: string;
  status: "pending" | "approved" | "rejected" | "executed";
  reason: string;
  createdAt: string;
}

export interface FlagsSnapshot {
  flags: FeatureFlag[];
  proposals: KillSwitchProposal[];
  generatedAt: string;
}

export interface UpsertFlagInput {
  key: string;
  name: string;
  description: string;
  scope: FlagScope;
  scopeTarget?: string | null;
  kind: FlagKind;
  rolloutPct: number;
  variants?: FlagVariant[];
  scheduledAt?: string | null;
}

export interface UpdateRolloutInput {
  flagId: Id;
  rolloutPct: number;
  variants?: FlagVariant[];
}
