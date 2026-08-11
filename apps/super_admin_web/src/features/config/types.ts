import type { Id } from "@/shared/types/common";

export interface ConfigSetting {
  id: Id;
  key: string;
  value: string;
  category:
    | "platform"
    | "brand"
    | "api"
    | "locale"
    | "currency"
    | "region"
    | "tax"
    | "notification";
  description: string;
}

export interface NotificationProviderConfig {
  id: Id;
  name: string;
  channel: "email" | "sms" | "push" | "webhook";
  status: "configured" | "disabled" | "error";
  endpoint: string;
}

export interface ConfigSnapshot {
  generatedAt: string;
  settings: ConfigSetting[];
  locales: string[];
  currencies: string[];
  regions: string[];
  taxEngines: string[];
  notificationProviders: NotificationProviderConfig[];
}
