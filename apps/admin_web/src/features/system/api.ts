import type { SystemSnapshot } from "./types";
import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient } from "@/shared/api/client";

export async function fetchSystemSnapshot(): Promise<SystemSnapshot> {
  try {
    return await apiClient<SystemSnapshot>("/admin/system");
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    return mockSystemSnapshot();
  }
}

async function mockSystemSnapshot(): Promise<SystemSnapshot> {
  await new Promise((r) => setTimeout(r, 220));
  const now = new Date().toISOString();

  return {
    generatedAt: now,
    locales: ["tr-TR", "en-US", "de-DE"],
    currencies: ["TRY", "USD", "EUR"],
    settings: [
      {
        id: "s1",
        key: "app.support_phone",
        value: "+90 850 000 00 00",
        category: "app",
        description: "Customer support hotline",
      },
      {
        id: "s2",
        key: "app.default_eta_buffer_min",
        value: "3",
        category: "app",
        description: "ETA buffer minutes",
      },
      {
        id: "s3",
        key: "locale.default",
        value: "tr-TR",
        category: "locale",
        description: "Default locale",
      },
      {
        id: "s4",
        key: "locale.fallback",
        value: "en-US",
        category: "locale",
        description: "Fallback locale",
      },
      {
        id: "s5",
        key: "currency.default",
        value: "TRY",
        category: "currency",
        description: "Primary currency",
      },
      {
        id: "s6",
        key: "currency.minor_units",
        value: "2",
        category: "currency",
        description: "Minor unit precision",
      },
      {
        id: "s7",
        key: "tax.vat_rate",
        value: "0.20",
        category: "tax",
        description: "Default VAT rate",
      },
      {
        id: "s8",
        key: "tax.include_in_price",
        value: "true",
        category: "tax",
        description: "Prices include tax",
      },
      {
        id: "s9",
        key: "region.active_cities",
        value: "Istanbul,Ankara,Izmir",
        category: "region",
        description: "Active city list",
      },
      {
        id: "s10",
        key: "region.timezone",
        value: "Europe/Istanbul",
        category: "region",
        description: "Ops timezone",
      },
    ],
    flags: [
      {
        id: "f1",
        key: "checkout.apple_pay",
        enabled: true,
        killSwitch: false,
        description: "Apple Pay on checkout",
        updatedAt: now,
      },
      {
        id: "f2",
        key: "live.heatmap",
        enabled: true,
        killSwitch: false,
        description: "Live ops heatmap layer",
        updatedAt: now,
      },
      {
        id: "f3",
        key: "kill.orders_new",
        enabled: false,
        killSwitch: true,
        description: "Kill switch: block new orders",
        updatedAt: now,
      },
      {
        id: "f4",
        key: "kill.payments",
        enabled: false,
        killSwitch: true,
        description: "Kill switch: pause payments",
        updatedAt: now,
      },
      {
        id: "f5",
        key: "ai.surge_suggestions",
        enabled: true,
        killSwitch: false,
        description: "AI surge pricing suggestions",
        updatedAt: now,
      },
      {
        id: "f6",
        key: "kill.dispatch",
        enabled: false,
        killSwitch: true,
        description: "Kill switch: freeze auto-dispatch",
        updatedAt: now,
      },
    ],
    templates: [
      {
        id: "t1",
        channel: "email",
        name: "Order confirmed",
        locale: "tr-TR",
        subject: "Siparişiniz alındı",
        bodyPreview: "Merhaba {{name}}, sipariş {{orderId}}…",
        updatedAt: now,
      },
      {
        id: "t2",
        channel: "sms",
        name: "Courier nearby",
        locale: "tr-TR",
        bodyPreview: "Kuryeniz yaklaşıyor. ETA {{eta}} dk.",
        updatedAt: now,
      },
      {
        id: "t3",
        channel: "push",
        name: "Flash deal",
        locale: "en-US",
        subject: "Flash deal near you",
        bodyPreview: "Save {{pct}}% for the next hour.",
        updatedAt: now,
      },
      {
        id: "t4",
        channel: "email",
        name: "Refund processed",
        locale: "tr-TR",
        subject: "İadeniz tamamlandı",
        bodyPreview: "{{amount}} iadeniz kartınıza…",
        updatedAt: now,
      },
      {
        id: "t5",
        channel: "in_app",
        name: "Stock alert ops",
        locale: "en-US",
        bodyPreview: "WH {{warehouse}} low on {{sku}}",
        updatedAt: now,
      },
      {
        id: "t6",
        channel: "sms",
        name: "OTP login",
        locale: "tr-TR",
        bodyPreview: "NEXORA kodunuz: {{code}}",
        updatedAt: now,
      },
    ],
    zones: [
      {
        id: "z1",
        name: "Kadıköy core",
        city: "Istanbul",
        href: "/delivery/zones",
      },
      {
        id: "z2",
        name: "Beşiktaş evening",
        city: "Istanbul",
        href: "/delivery/zones",
      },
      {
        id: "z3",
        name: "Çankaya",
        city: "Ankara",
        href: "/delivery/zones",
      },
    ],
  };
}
