import type { Id } from "@/shared/types/common";

export type PrivacyRegime = "gdpr" | "kvkk" | "ccpa";
export type ConsentStatus = "granted" | "denied" | "withdrawn" | "pending";
export type PrivacyRequestType =
  | "access"
  | "export"
  | "delete"
  | "rectify"
  | "restrict";
export type PrivacyRequestStatus =
  | "received"
  | "verifying"
  | "in_progress"
  | "completed"
  | "rejected";

export interface RegimePanel {
  regime: PrivacyRegime;
  label: string;
  status: "compliant" | "gaps" | "at_risk";
  openRequests: number;
  retentionAligned: boolean;
  dpaSigned: boolean;
  notes: string;
}

export interface RetentionPolicy {
  id: Id;
  dataClass: string;
  regime: PrivacyRegime | "global";
  retentionDays: number;
  legalHoldExempt: boolean;
  autoDelete: boolean;
}

export interface ConsentRecord {
  id: Id;
  subjectId: Id;
  purpose: string;
  regime: PrivacyRegime;
  status: ConsentStatus;
  channel: string;
  updatedAt: string;
}

export interface PrivacyRequest {
  id: Id;
  type: PrivacyRequestType;
  regime: PrivacyRegime;
  subjectEmail: string;
  tenantName: string;
  status: PrivacyRequestStatus;
  dueAt: string;
  createdAt: string;
  assignee: string | null;
}

export interface ComplianceSnapshot {
  regimes: RegimePanel[];
  retention: RetentionPolicy[];
  consents: ConsentRecord[];
  requests: PrivacyRequest[];
  generatedAt: string;
}
