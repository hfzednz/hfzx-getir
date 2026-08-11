import type { Id } from "@/shared/types/common";
import type { SeriesPoint } from "@/shared/lib/charts";

export type ChannelKind = "email" | "sms" | "push" | "whatsapp" | "webhook";
export type ProviderStatus = "active" | "degraded" | "disabled";
export type DeliveryStatus =
  | "queued"
  | "sent"
  | "delivered"
  | "failed"
  | "bounced";

export interface NotificationProvider {
  id: Id;
  name: string;
  channel: ChannelKind;
  vendor: string;
  status: ProviderStatus;
  region: string;
  dailyCap: number;
  sentToday: number;
  successPct: number;
}

export interface NotificationTemplate {
  id: Id;
  name: string;
  channel: ChannelKind;
  locale: string;
  version: number;
  updatedAt: string;
  owner: string;
}

export interface DeliveryEvent {
  id: Id;
  providerId: Id;
  providerName: string;
  channel: ChannelKind;
  template: string;
  status: DeliveryStatus;
  recipientHash: string;
  latencyMs: number;
  createdAt: string;
  errorCode: string | null;
}

export interface NotificationsSnapshot {
  generatedAt: string;
  kpis: {
    providersActive: number;
    delivered24h: number;
    failed24h: number;
    avgLatencyMs: number;
    templates: number;
    webhookSuccessPct: number;
  };
  volumeSeries: SeriesPoint[];
  providers: NotificationProvider[];
  templates: NotificationTemplate[];
  deliveries: DeliveryEvent[];
}
