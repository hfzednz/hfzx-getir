import type { Id } from "@/shared/types/common";
import type { DualControlAction } from "@/shared/permissions/dual-control";

export type PlanTier = "starter" | "growth" | "enterprise" | "custom";
export type LicenseStatus =
  | "active"
  | "trial"
  | "past_due"
  | "cancelled"
  | "expired";

export interface LicensePlan {
  id: Id;
  code: string;
  name: string;
  tier: PlanTier;
  monthlyPriceMinor: number;
  currency: string;
  features: string[];
  limits: {
    warehouses: number;
    couriers: number;
    apiCallsPerMonth: number;
    storageGb: number;
    seats: number;
  };
}

export interface TenantLicense {
  id: Id;
  tenantId: Id;
  tenantName: string;
  planId: Id;
  planName: string;
  tier: PlanTier;
  status: LicenseStatus;
  seats: number;
  seatsUsed: number;
  renewsAt: string;
  overageEnabled: boolean;
  overageSpendMinor: number;
  currency: string;
  featureOverrides: string[];
  enterpriseContractId: string | null;
}

export interface UsageLimitRow {
  id: Id;
  tenantId: Id;
  tenantName: string;
  metric: string;
  used: number;
  limit: number;
  unit: string;
  overagePct: number;
}

export interface LicenseOverrideProposal {
  id: Id;
  action: Extract<DualControlAction, "license_override">;
  licenseId: Id;
  tenantName: string;
  requesterId: Id;
  requesterEmail: string;
  status: "pending" | "approved" | "rejected" | "executed";
  reason: string;
  payload: {
    seats?: number;
    featureOverrides?: string[];
    overageEnabled?: boolean;
  };
  createdAt: string;
}

export interface LicensesSnapshot {
  plans: LicensePlan[];
  licenses: TenantLicense[];
  usage: UsageLimitRow[];
  proposals: LicenseOverrideProposal[];
  renewalsDueDays: number;
  generatedAt: string;
}
