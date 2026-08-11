import { ApiError } from "@/shared/api/client";
import type {
  SupportTicket,
  SupportWorkspace,
  TicketListParams,
  TicketStatus,
} from "./types";

const delay = (ms = 180) => new Promise((r) => setTimeout(r, ms));

let mockTickets: SupportTicket[] = [
  {
    id: "tkt_901",
    subject: "Missing dairy items",
    status: "open",
    priority: "high",
    category: "order",
    customerId: "cus_1001",
    customerName: "Ayşe Yılmaz",
    orderId: "ord_9921",
    assignee: "agent.selin",
    createdAt: "2026-08-06T09:10:00Z",
    updatedAt: "2026-08-06T09:40:00Z",
    messages: [
      {
        id: "m1",
        author: "Ayşe Yılmaz",
        role: "customer",
        body: "Milk and yogurt missing from bag.",
        at: "2026-08-06T09:10:00Z",
      },
      {
        id: "m2",
        author: "NEXORA AI",
        role: "ai",
        body: "I can offer a partial refund or re-delivery within 20 min.",
        at: "2026-08-06T09:11:00Z",
      },
      {
        id: "m3",
        author: "agent.selin",
        role: "agent",
        body: "Sorry about that — checking warehouse pick log.",
        at: "2026-08-06T09:25:00Z",
      },
    ],
    refund: {
      id: "rf_1",
      amountMinor: 6200,
      currency: "TRY",
      reason: "Missing items",
      status: "pending",
      orderId: "ord_9921",
    },
    qc: [
      {
        id: "qc1",
        skuId: "sku_milk_1l",
        issue: "Not packed",
        severity: "high",
      },
    ],
    escalated: false,
    aiSuggestedReply:
      "We can refund TRY 62.00 for the missing dairy or schedule a free re-delivery in ~18 minutes. Which do you prefer?",
  },
  {
    id: "tkt_902",
    subject: "Late delivery complaint",
    status: "escalated",
    priority: "urgent",
    category: "complaint",
    customerId: "cus_1002",
    customerName: "Mehmet Demir",
    orderId: "ord_8802",
    assignee: "lead.burak",
    createdAt: "2026-08-05T21:00:00Z",
    updatedAt: "2026-08-06T08:00:00Z",
    messages: [
      {
        id: "m4",
        author: "Mehmet Demir",
        role: "customer",
        body: "Order arrived 40 min late. Unacceptable.",
        at: "2026-08-05T21:00:00Z",
      },
    ],
    refund: null,
    qc: [],
    escalated: true,
    aiSuggestedReply:
      "Offer goodwill voucher TRY 30 and explain courier shortage in Beşiktaş.",
  },
  {
    id: "tkt_903",
    subject: "Damaged packaging QC",
    status: "pending",
    priority: "medium",
    category: "product_qc",
    customerId: "cus_1003",
    customerName: "Elif Kaya",
    orderId: "ord_7755",
    assignee: null,
    createdAt: "2026-08-06T07:30:00Z",
    updatedAt: "2026-08-06T07:45:00Z",
    messages: [
      {
        id: "m5",
        author: "Elif Kaya",
        role: "customer",
        body: "Juice bottle leaked in the bag.",
        at: "2026-08-06T07:30:00Z",
      },
    ],
    refund: {
      id: "rf_2",
      amountMinor: 4500,
      currency: "TRY",
      reason: "Damaged product",
      status: "pending",
      orderId: "ord_7755",
    },
    qc: [
      {
        id: "qc2",
        skuId: "sku_juice",
        issue: "Leak / seal failure",
        severity: "medium",
      },
    ],
    escalated: false,
    aiSuggestedReply: null,
  },
];

/** Mock support — replaced by GET /admin/support/tickets when BFF is live. */
export async function fetchSupportWorkspace(
  params: TicketListParams = {},
): Promise<SupportWorkspace> {
  await delay();
  let tickets = [...mockTickets];
  if (params.status && params.status !== "all") {
    tickets = tickets.filter((t) => t.status === params.status);
  }
  if (params.category && params.category !== "all") {
    tickets = tickets.filter((t) => t.category === params.category);
  }
  if (params.q?.trim()) {
    const q = params.q.trim().toLowerCase();
    tickets = tickets.filter(
      (t) =>
        t.subject.toLowerCase().includes(q) ||
        t.id.toLowerCase().includes(q) ||
        t.customerName.toLowerCase().includes(q) ||
        (t.orderId?.toLowerCase().includes(q) ?? false),
    );
  }
  return {
    tickets,
    liveChat: {
      activeSessions: 14,
      queued: 3,
      avgWaitSec: 48,
      agentsOnline: 11,
    },
    aiChatbot: {
      enabled: true,
      containmentRatePct: 64.2,
      handoffRatePct: 21.5,
      topIntents: [
        { intent: "where_is_order", count: 420 },
        { intent: "refund_request", count: 188 },
        { intent: "coupon_help", count: 96 },
      ],
    },
    complaintCount: tickets.filter((t) => t.category === "complaint").length,
    openRefunds: tickets.filter((t) => t.refund?.status === "pending").length,
  };
}

export async function fetchTicket(id: string): Promise<SupportTicket> {
  await delay();
  const found = mockTickets.find((t) => t.id === id);
  if (!found) {
    throw new ApiError(404, {
      code: "not_found",
      message: "Ticket not found",
      traceId: "mock",
    });
  }
  return structuredClone(found);
}

export async function escalateTicket(id: string): Promise<SupportTicket> {
  await delay(200);
  return patchTicket(id, { status: "escalated", escalated: true });
}

export async function resolveTicket(id: string): Promise<SupportTicket> {
  await delay(200);
  return patchTicket(id, { status: "resolved" });
}

export async function approveTicketRefund(
  id: string,
): Promise<SupportTicket> {
  await delay(240);
  const t = await fetchTicket(id);
  if (!t.refund) {
    throw new ApiError(400, {
      code: "no_refund",
      message: "No refund on ticket",
      traceId: "mock",
    });
  }
  return patchTicket(id, {
    refund: { ...t.refund, status: "approved" },
    status: "pending" as TicketStatus,
  });
}

function patchTicket(
  id: string,
  patch: Partial<SupportTicket>,
): SupportTicket {
  const idx = mockTickets.findIndex((t) => t.id === id);
  if (idx < 0) {
    throw new ApiError(404, {
      code: "not_found",
      message: "Ticket not found",
      traceId: "mock",
    });
  }
  const next = {
    ...mockTickets[idx]!,
    ...patch,
    updatedAt: new Date().toISOString(),
  };
  mockTickets = mockTickets.map((t, i) => (i === idx ? next : t));
  return structuredClone(next);
}
