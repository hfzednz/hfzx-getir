export interface SystemSetting {
  id: string;
  key: string;
  value: string;
  category: "app" | "locale" | "currency" | "tax" | "region";
  description: string;
}

export interface FeatureFlag {
  id: string;
  key: string;
  enabled: boolean;
  killSwitch: boolean;
  description: string;
  updatedAt: string;
}

export interface MessageTemplate {
  id: string;
  channel: "email" | "sms" | "push" | "in_app";
  name: string;
  locale: string;
  subject?: string;
  bodyPreview: string;
  updatedAt: string;
}

export interface DeliveryZoneLink {
  id: string;
  name: string;
  city: string;
  href: string;
}

export interface SystemSnapshot {
  generatedAt: string;
  settings: SystemSetting[];
  flags: FeatureFlag[];
  templates: MessageTemplate[];
  zones: DeliveryZoneLink[];
  locales: string[];
  currencies: string[];
}
