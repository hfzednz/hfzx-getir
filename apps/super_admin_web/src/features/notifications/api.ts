import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { NotificationsSnapshot, NotificationProvider } from "./types";

function delay(ms = 200): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const mockProviders: NotificationProvider[] = [
  {
    id: "np_ses",
    name: "AWS SES primary",
    channel: "email",
    vendor: "aws-ses",
    status: "active",
    region: "eu-west-1",
    dailyCap: 2_000_000,
    sentToday: 842_110,
    successPct: 99.2,
  },
  {
    id: "np_twilio",
    name: "Twilio SMS",
    channel: "sms",
    vendor: "twilio",
    status: "active",
    region: "global",
    dailyCap: 500_000,
    sentToday: 128_440,
    successPct: 98.1,
  },
  {
    id: "np_fcm",
    name: "FCM push",
    channel: "push",
    vendor: "firebase",
    status: "active",
    region: "global",
    dailyCap: 5_000_000,
    sentToday: 1_920_000,
    successPct: 97.4,
  },
  {
    id: "np_wa",
    name: "WhatsApp Business",
    channel: "whatsapp",
    vendor: "meta",
    status: "degraded",
    region: "eu",
    dailyCap: 200_000,
    sentToday: 41_200,
    successPct: 91.0,
  },
  {
    id: "np_hook",
    name: "Partner webhooks",
    channel: "webhook",
    vendor: "nexora-relay",
    status: "active",
    region: "global",
    dailyCap: 10_000_000,
    sentToday: 3_410_000,
    successPct: 99.6,
  },
  {
    id: "np_sendgrid",
    name: "SendGrid failover",
    channel: "email",
    vendor: "sendgrid",
    status: "disabled",
    region: "us-east-1",
    dailyCap: 1_000_000,
    sentToday: 0,
    successPct: 0,
  },
];

function mockSnapshot(): NotificationsSnapshot {
  const hours = ["00", "04", "08", "12", "16", "20"];
  return {
    generatedAt: new Date().toISOString(),
    kpis: {
      providersActive: mockProviders.filter((p) => p.status === "active").length,
      delivered24h: 4_210_000,
      failed24h: 18_420,
      avgLatencyMs: 210,
      templates: 64,
      webhookSuccessPct: 99.6,
    },
    volumeSeries: hours.map((label, i) => ({
      label: `${label}:00`,
      value: 120_000 + i * 80_000,
    })),
    providers: [...mockProviders],
    templates: [
      {
        id: "tpl_1",
        name: "platform.security.mfa_challenge",
        channel: "email",
        locale: "en",
        version: 4,
        updatedAt: new Date(Date.now() - 10 * 86400_000).toISOString(),
        owner: "platform_security",
      },
      {
        id: "tpl_2",
        name: "platform.billing.invoice_ready",
        channel: "email",
        locale: "en",
        version: 2,
        updatedAt: new Date(Date.now() - 30 * 86400_000).toISOString(),
        owner: "platform_finops",
      },
      {
        id: "tpl_3",
        name: "platform.dr.failover_notice",
        channel: "sms",
        locale: "en",
        version: 1,
        updatedAt: new Date(Date.now() - 60 * 86400_000).toISOString(),
        owner: "platform_sre",
      },
      {
        id: "tpl_4",
        name: "platform.ops.incident_push",
        channel: "push",
        locale: "en",
        version: 6,
        updatedAt: new Date(Date.now() - 5 * 86400_000).toISOString(),
        owner: "platform_sre",
      },
      {
        id: "tpl_5",
        name: "platform.tenant.suspend_notice",
        channel: "whatsapp",
        locale: "tr",
        version: 3,
        updatedAt: new Date(Date.now() - 14 * 86400_000).toISOString(),
        owner: "platform_compliance",
      },
      {
        id: "tpl_6",
        name: "platform.webhook.license_event",
        channel: "webhook",
        locale: "en",
        version: 8,
        updatedAt: new Date(Date.now() - 2 * 86400_000).toISOString(),
        owner: "platform_owner",
      },
    ],
    deliveries: [
      {
        id: "del_1",
        providerId: "np_ses",
        providerName: "AWS SES primary",
        channel: "email",
        template: "platform.billing.invoice_ready",
        status: "delivered",
        recipientHash: "sha256:a1…f9",
        latencyMs: 180,
        createdAt: new Date(Date.now() - 12 * 60_000).toISOString(),
        errorCode: null,
      },
      {
        id: "del_2",
        providerId: "np_twilio",
        providerName: "Twilio SMS",
        channel: "sms",
        template: "platform.dr.failover_notice",
        status: "sent",
        recipientHash: "sha256:b2…c0",
        latencyMs: 420,
        createdAt: new Date(Date.now() - 8 * 60_000).toISOString(),
        errorCode: null,
      },
      {
        id: "del_3",
        providerId: "np_wa",
        providerName: "WhatsApp Business",
        channel: "whatsapp",
        template: "platform.tenant.suspend_notice",
        status: "failed",
        recipientHash: "sha256:c3…d1",
        latencyMs: 2100,
        createdAt: new Date(Date.now() - 5 * 60_000).toISOString(),
        errorCode: "WA_RATE_LIMIT",
      },
      {
        id: "del_4",
        providerId: "np_hook",
        providerName: "Partner webhooks",
        channel: "webhook",
        template: "platform.webhook.license_event",
        status: "delivered",
        recipientHash: "sha256:d4…e2",
        latencyMs: 95,
        createdAt: new Date(Date.now() - 2 * 60_000).toISOString(),
        errorCode: null,
      },
      {
        id: "del_5",
        providerId: "np_fcm",
        providerName: "FCM push",
        channel: "push",
        template: "platform.ops.incident_push",
        status: "bounced",
        recipientHash: "sha256:e5…f3",
        latencyMs: 310,
        createdAt: new Date(Date.now() - 1 * 60_000).toISOString(),
        errorCode: "DEVICE_UNREGISTERED",
      },
      {
        id: "del_6",
        providerId: "np_ses",
        providerName: "AWS SES primary",
        channel: "email",
        template: "platform.security.mfa_challenge",
        status: "queued",
        recipientHash: "sha256:f6…a4",
        latencyMs: 0,
        createdAt: new Date().toISOString(),
        errorCode: null,
      },
    ],
  };
}

export async function fetchNotifications(): Promise<NotificationsSnapshot> {
  try {
    return await apiClient<NotificationsSnapshot>(
      platformPath("/notifications"),
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockSnapshot();
    }
    throw err;
  }
}

export async function setProviderStatus(input: {
  providerId: string;
  status: NotificationProvider["status"];
}): Promise<NotificationProvider> {
  try {
    return await apiClient<NotificationProvider>(
      platformPath(`/notifications/providers/${input.providerId}`),
      { method: "PATCH", body: { status: input.status }, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockProviders.findIndex((p) => p.id === input.providerId);
      if (idx < 0) throw new Error("Provider not found");
      mockProviders[idx] = { ...mockProviders[idx], status: input.status };
      return mockProviders[idx];
    }
    throw err;
  }
}
