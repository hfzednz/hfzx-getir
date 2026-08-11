import type { CrmListParams, CrmWorkspace, CrmCustomer } from "./types";

const delay = (ms = 180) => new Promise((r) => setTimeout(r, ms));

const tags = [
  { id: "tag_vip", label: "VIP", color: "#0B6E6E" },
  { id: "tag_churn", label: "Churn risk", color: "#B45309" },
  { id: "tag_high", label: "High AOV", color: "#1D4ED8" },
];

const segments = [
  {
    id: "seg_new",
    name: "New 30d",
    size: 42000,
    rulesSummary: "First order within 30 days",
  },
  {
    id: "seg_loyal",
    name: "Loyal weekly",
    size: 96000,
    rulesSummary: ">=1 order / week last 8 weeks",
  },
  {
    id: "seg_dormant",
    name: "Dormant 60d",
    size: 71000,
    rulesSummary: "No order in 60 days",
  },
];

const mockCustomers: CrmCustomer[] = [
  {
    id: "cus_1001",
    name: "Ayşe Yılmaz",
    email: "ayse.y@example.com",
    phone: "+90 532 111 2233",
    cityId: "city_ist",
    lifetimeValueMinor: 12_450_00,
    currency: "TRY",
    orderCount: 48,
    lastOrderAt: "2026-08-05T19:20:00Z",
    tags: [tags[0]!, tags[2]!],
    segments: ["Loyal weekly"],
    riskScore: 0.12,
    notes: [
      {
        id: "n1",
        author: "support.lead",
        body: "Prefers evening delivery · allergic to nuts note on profile",
        createdAt: "2026-07-12T10:00:00Z",
      },
    ],
    channelHistory: [
      {
        id: "ch1",
        channel: "push",
        direction: "outbound",
        subject: "Flash dairy",
        preview: "25% off selected dairy until 22:00",
        at: "2026-08-06T10:05:00Z",
        campaignId: "cmp_flash_01",
      },
      {
        id: "ch2",
        channel: "whatsapp",
        direction: "inbound",
        subject: "Delivery ETA",
        preview: "Where is my order ORD-9921?",
        at: "2026-08-04T18:40:00Z",
        campaignId: null,
      },
      {
        id: "ch3",
        channel: "email",
        direction: "outbound",
        subject: "VIP weekend invite",
        preview: "Early access to VIPWEEK",
        at: "2026-08-01T09:00:00Z",
        campaignId: "cmp_aud_05",
      },
    ],
    linkedCampaignIds: ["cmp_flash_01", "cmp_aud_05"],
  },
  {
    id: "cus_1002",
    name: "Mehmet Demir",
    email: "mehmet.d@example.com",
    phone: "+90 533 444 5566",
    cityId: "city_ank",
    lifetimeValueMinor: 3_210_00,
    currency: "TRY",
    orderCount: 9,
    lastOrderAt: "2026-06-01T12:00:00Z",
    tags: [tags[1]!],
    segments: ["Dormant 60d"],
    riskScore: 0.61,
    notes: [
      {
        id: "n2",
        author: "crm.ops",
        body: "Win-back SMS scheduled · do not push flash tonight",
        createdAt: "2026-08-03T14:00:00Z",
      },
    ],
    channelHistory: [
      {
        id: "ch4",
        channel: "sms",
        direction: "outbound",
        subject: "We miss you",
        preview: "TRY40 off your next basket",
        at: "2026-08-03T15:00:00Z",
        campaignId: null,
      },
    ],
    linkedCampaignIds: [],
  },
  {
    id: "cus_1003",
    name: "Elif Kaya",
    email: "elif.k@example.com",
    phone: "+90 535 777 8899",
    cityId: "city_ist",
    lifetimeValueMinor: 8_900_00,
    currency: "TRY",
    orderCount: 31,
    lastOrderAt: "2026-08-06T08:10:00Z",
    tags: [tags[2]!],
    segments: ["Loyal weekly", "New 30d"],
    riskScore: 0.08,
    notes: [],
    channelHistory: [
      {
        id: "ch5",
        channel: "push",
        direction: "outbound",
        subject: "AI snack uplift",
        preview: "Personalized snack picks",
        at: "2026-08-05T11:20:00Z",
        campaignId: "cmp_pers_04",
      },
    ],
    linkedCampaignIds: ["cmp_pers_04"],
  },
];

/** Mock CRM — replaced by GET /admin/crm when BFF is live. */
export async function fetchCrmWorkspace(
  params: CrmListParams = {},
): Promise<CrmWorkspace> {
  await delay();
  let customers = [...mockCustomers];
  if (params.cityId) {
    customers = customers.filter((c) => c.cityId === params.cityId);
  }
  if (params.tag && params.tag !== "all") {
    customers = customers.filter((c) =>
      c.tags.some((t) => t.id === params.tag || t.label === params.tag),
    );
  }
  if (params.segment && params.segment !== "all") {
    customers = customers.filter((c) =>
      c.segments.some((s) => s.toLowerCase().includes(params.segment!.toLowerCase())),
    );
  }
  if (params.q?.trim()) {
    const q = params.q.trim().toLowerCase();
    customers = customers.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        c.email.toLowerCase().includes(q) ||
        c.id.toLowerCase().includes(q) ||
        c.phone.includes(q),
    );
  }
  return {
    customers,
    tags,
    segments,
    total: customers.length,
  };
}

export async function addCrmNote(
  customerId: string,
  body: string,
  author: string,
): Promise<CrmCustomer> {
  await delay(200);
  const cus = mockCustomers.find((c) => c.id === customerId);
  if (!cus) throw new Error("Customer not found");
  cus.notes = [
    {
      id: `n_${Date.now().toString(36)}`,
      author,
      body,
      createdAt: new Date().toISOString(),
    },
    ...cus.notes,
  ];
  return { ...cus, notes: [...cus.notes] };
}
