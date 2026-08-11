export type CustomerSegment =
  | "new"
  | "loyal"
  | "vip"
  | "churn_risk"
  | "high_aov"
  | "fraud_watch";

export interface CustomerListItem extends Record<string, unknown> {
  id: string;
  name: string;
  email: string;
  phone: string;
  cityId: string;
  segment: CustomerSegment;
  orderCount: number;
  lifetimeValueMinor: number;
  currency: string;
  riskScore: number;
  fraudScore: number;
  loyaltyTier: string;
  walletBalanceMinor: number;
  createdAt: string;
  lastOrderAt: string | null;
}

export interface CustomerAddress {
  id: string;
  label: string;
  line1: string;
  district: string;
  city: string;
  isDefault: boolean;
}

export interface CustomerOrderSummary {
  id: string;
  status: string;
  totalMinor: number;
  currency: string;
  createdAt: string;
}

export interface CustomerWalletTxn {
  id: string;
  type: "credit" | "debit" | "adjustment";
  amountMinor: number;
  currency: string;
  note: string;
  at: string;
}

export interface CustomerLoyalty {
  tier: string;
  points: number;
  pointsToNextTier: number;
}

export interface CustomerCoupon {
  id: string;
  code: string;
  status: "active" | "used" | "expired";
  discountLabel: string;
  expiresAt: string;
}

export interface CustomerSupportTicket {
  id: string;
  subject: string;
  status: "open" | "pending" | "resolved";
  createdAt: string;
}

export interface CustomerNote {
  id: string;
  body: string;
  author: string;
  createdAt: string;
}

export interface CustomerProfile extends CustomerListItem {
  addresses: CustomerAddress[];
  recentOrders: CustomerOrderSummary[];
  walletTxns: CustomerWalletTxn[];
  loyalty: CustomerLoyalty;
  coupons: CustomerCoupon[];
  supportHistory: CustomerSupportTicket[];
  notes: CustomerNote[];
}

export interface CustomerListFilters {
  q?: string;
  segment?: CustomerSegment | "all";
  page?: number;
  pageSize?: number;
  cityId?: string | null;
}

export interface CustomerAdjustmentInput {
  customerId: string;
  type: "wallet" | "loyalty" | "note";
  amountMinor?: number;
  points?: number;
  note: string;
}

export interface CustomerAdjustmentResult {
  customerId: string;
  ok: boolean;
  message: string;
}
