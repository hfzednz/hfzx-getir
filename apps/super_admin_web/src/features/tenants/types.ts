import type { Id } from "@/shared/types/common";
import type { DualControlAction } from "@/shared/permissions/dual-control";

export type TenantIsolationMode = "shared" | "separate" | "hybrid";
export type TenantStatus = "active" | "suspended" | "pending" | "migrating";
export type MigrationStatus =
  | "idle"
  | "planned"
  | "in_progress"
  | "completed"
  | "failed";
export type BackupStatus = "ok" | "stale" | "failed" | "running";

export interface TenantListItem {
  id: Id;
  slug: string;
  name: string;
  companyId: Id;
  companyName: string;
  isolationMode: TenantIsolationMode;
  status: TenantStatus;
  region: string;
  createdAt: string;
}

export interface TenantConfig {
  featurePack: string;
  maxWarehouses: number;
  maxUsers: number;
  dataResidency: string;
  rlsEnabled: boolean;
}

export interface TenantCustomization {
  primaryColor: string;
  logoUrl: string;
  customDomain: string | null;
  whiteLabel: boolean;
}

export interface TenantMigration {
  status: MigrationStatus;
  targetMode: TenantIsolationMode | null;
  progressPct: number;
  lastMessage: string;
  updatedAt: string;
}

export interface TenantBackup {
  id: Id;
  label: string;
  status: BackupStatus;
  sizeGb: number;
  takenAt: string;
}

export interface TenantMonitorMetric {
  id: string;
  label: string;
  value: string;
  tone: "success" | "warning" | "danger" | "neutral";
}

export interface TenantDetail extends TenantListItem {
  description: string;
  config: TenantConfig;
  customization: TenantCustomization;
  migration: TenantMigration;
  backups: TenantBackup[];
  monitoring: TenantMonitorMetric[];
}

export interface TenantDualControlProposal {
  id: Id;
  action: Extract<DualControlAction, "tenant_suspend" | "tenant_delete">;
  tenantId: Id;
  tenantName: string;
  requesterId: Id;
  requesterEmail: string;
  status: "pending" | "approved" | "rejected" | "executed";
  reason: string;
  createdAt: string;
}

export interface CreateTenantInput {
  name: string;
  slug: string;
  companyId: Id;
  isolationMode: TenantIsolationMode;
  region: string;
}

export interface TenantsListResponse {
  items: TenantListItem[];
  total: number;
  generatedAt: string;
  proposals: TenantDualControlProposal[];
}
