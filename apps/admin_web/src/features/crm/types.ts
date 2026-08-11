import type { Id } from "@/shared/types/common";

export type CrmChannel = "email" | "sms" | "push" | "whatsapp";

export interface CrmTag {
  id: Id;
  label: string;
  color: string;
}

export interface CrmSegment {
  id: Id;
  name: string;
  size: number;
  rulesSummary: string;
}

export interface CrmNote {
  id: Id;
  author: string;
  body: string;
  createdAt: string;
}

export interface CrmChannelEvent {
  id: Id;
  channel: CrmChannel;
  direction: "inbound" | "outbound";
  subject: string;
  preview: string;
  at: string;
  campaignId: string | null;
}

export interface CrmCustomer {
  id: Id;
  name: string;
  email: string;
  phone: string;
  cityId: string;
  lifetimeValueMinor: number;
  currency: string;
  orderCount: number;
  lastOrderAt: string | null;
  tags: CrmTag[];
  segments: string[];
  riskScore: number;
  notes: CrmNote[];
  channelHistory: CrmChannelEvent[];
  linkedCampaignIds: string[];
}

export interface CrmListParams {
  q?: string;
  tag?: string | "all";
  segment?: string | "all";
  cityId?: string | null;
}

export interface CrmWorkspace {
  customers: CrmCustomer[];
  tags: CrmTag[];
  segments: CrmSegment[];
  total: number;
}
