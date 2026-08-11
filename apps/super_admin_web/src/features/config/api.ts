import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { ConfigSnapshot, ConfigSetting } from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function mockSnapshot(): ConfigSnapshot {
  const settings: ConfigSetting[] = [
    {
      id: "s1",
      key: "platform.name",
      value: "NEXORA",
      category: "platform",
      description: "Platform display name",
    },
    {
      id: "s2",
      key: "platform.support_email",
      value: "platform@nexora.example",
      category: "platform",
      description: "Platform ops contact",
    },
    {
      id: "s3",
      key: "brand.primary_color",
      value: "#0B6E6E",
      category: "brand",
      description: "Default brand primary",
    },
    {
      id: "s4",
      key: "brand.logo_url",
      value: "/assets/nexora-mark.svg",
      category: "brand",
      description: "Default logo asset",
    },
    {
      id: "s5",
      key: "api.rate_limit_rpm",
      value: "120000",
      category: "api",
      description: "Global API requests per minute",
    },
    {
      id: "s6",
      key: "api.burst_multiplier",
      value: "1.5",
      category: "api",
      description: "Burst allowance over rate limit",
    },
    {
      id: "s7",
      key: "locale.default",
      value: "en-US",
      category: "locale",
      description: "Fallback locale",
    },
    {
      id: "s8",
      key: "currency.settlement",
      value: "USD",
      category: "currency",
      description: "Platform settlement currency",
    },
    {
      id: "s9",
      key: "region.default",
      value: "eu-west-1",
      category: "region",
      description: "Default control-plane region",
    },
    {
      id: "s10",
      key: "tax.default_engine",
      value: "avalara",
      category: "tax",
      description: "Default tax calculation engine",
    },
    {
      id: "s11",
      key: "notification.retry_max",
      value: "5",
      category: "notification",
      description: "Provider retry attempts (config only)",
    },
  ];

  return {
    generatedAt: new Date().toISOString(),
    settings,
    locales: ["en-US", "tr-TR", "de-DE", "en-SG"],
    currencies: ["USD", "EUR", "TRY", "SGD"],
    regions: ["eu-west-1", "eu-central-1", "us-east-1", "ap-southeast-1"],
    taxEngines: ["avalara", "gib", "internal"],
    notificationProviders: [
      {
        id: "np1",
        name: "SendGrid",
        channel: "email",
        status: "configured",
        endpoint: "https://api.sendgrid.com/v3",
      },
      {
        id: "np2",
        name: "Twilio SMS",
        channel: "sms",
        status: "configured",
        endpoint: "https://api.twilio.com",
      },
      {
        id: "np3",
        name: "FCM Push",
        channel: "push",
        status: "configured",
        endpoint: "https://fcm.googleapis.com",
      },
      {
        id: "np4",
        name: "Ops webhook",
        channel: "webhook",
        status: "disabled",
        endpoint: "https://hooks.nexora.example/platform",
      },
    ],
  };
}

export async function fetchConfigSnapshot(): Promise<ConfigSnapshot> {
  try {
    return await apiClient<ConfigSnapshot>(platformPath("/config"));
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockSnapshot();
    }
    throw err;
  }
}
