import type { Id } from "@/shared/types/common";

export type TicketStatus =
  | "open"
  | "pending"
  | "escalated"
  | "resolved"
  | "closed";

export type TicketPriority = "low" | "medium" | "high" | "urgent";

export type TicketCategory =
  | "order"
  | "refund"
  | "complaint"
  | "delivery"
  | "product_qc"
  | "other";

export interface TicketMessage {
  id: Id;
  author: string;
  role: "customer" | "agent" | "system" | "ai";
  body: string;
  at: string;
}

export interface RefundRequest {
  id: Id;
  amountMinor: number;
  currency: string;
  reason: string;
  status: "pending" | "approved" | "rejected";
  orderId: string;
}

export interface QcFinding {
  id: Id;
  skuId: string;
  issue: string;
  severity: "low" | "medium" | "high";
}

export interface SupportTicket {
  id: Id;
  subject: string;
  status: TicketStatus;
  priority: TicketPriority;
  category: TicketCategory;
  customerId: string;
  customerName: string;
  orderId: string | null;
  assignee: string | null;
  createdAt: string;
  updatedAt: string;
  messages: TicketMessage[];
  refund: RefundRequest | null;
  qc: QcFinding[];
  escalated: boolean;
  aiSuggestedReply: string | null;
}

export interface TicketListParams {
  status?: TicketStatus | "all";
  category?: TicketCategory | "all";
  q?: string;
}

export interface LiveChatStub {
  activeSessions: number;
  queued: number;
  avgWaitSec: number;
  agentsOnline: number;
}

export interface AiChatbotPanel {
  enabled: boolean;
  containmentRatePct: number;
  handoffRatePct: number;
  topIntents: { intent: string; count: number }[];
}

export interface SupportWorkspace {
  tickets: SupportTicket[];
  liveChat: LiveChatStub;
  aiChatbot: AiChatbotPanel;
  complaintCount: number;
  openRefunds: number;
}
