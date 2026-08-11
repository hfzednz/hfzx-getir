import type { Id } from "@/shared/types/common";

export type CourierLiveStatus =
  | "available"
  | "busy"
  | "offline"
  | "break"
  | "emergency";

export type CourierDocumentStatus = "valid" | "expiring" | "expired" | "missing";

export interface CourierListItem extends Record<string, unknown> {
  id: Id;
  code: string;
  fullName: string;
  phone: string;
  cityId: string;
  zoneName: string;
  liveStatus: CourierLiveStatus;
  vehicleType: string;
  rating: number;
  ratingCount: number;
  activeAssignments: number;
  onTimePct: number;
  emergency: boolean;
  lastSeenAt: string;
}

export interface CourierAssignment extends Record<string, unknown> {
  id: Id;
  orderId: string;
  status: string;
  zoneName: string;
  etaMinutes: number;
  assignedAt: string;
}

export interface CourierScheduleSlot extends Record<string, unknown> {
  id: Id;
  day: string;
  start: string;
  end: string;
  zoneName: string;
}

export interface CourierPerformance {
  deliveriesToday: number;
  deliveriesWeek: number;
  onTimePct: number;
  avgDeliveryMinutes: number;
  acceptanceRatePct: number;
  cancelByCourierPct: number;
}

export interface CourierRating extends Record<string, unknown> {
  id: Id;
  score: number;
  comment: string;
  orderId: string;
  createdAt: string;
}

export interface CourierDocument extends Record<string, unknown> {
  id: Id;
  type: string;
  status: CourierDocumentStatus;
  expiresAt: string | null;
}

export interface CourierVehicle {
  plate: string;
  type: string;
  model: string;
  color: string;
  insuranceExpiresAt: string;
}

export interface CourierPayment extends Record<string, unknown> {
  id: Id;
  period: string;
  baseAmount: number;
  bonusAmount: number;
  penaltyAmount: number;
  netAmount: number;
  currency: string;
  status: string;
  paidAt: string | null;
}

export interface CourierBonus extends Record<string, unknown> {
  id: Id;
  reason: string;
  amount: number;
  currency: string;
  createdAt: string;
}

export interface CourierPenalty extends Record<string, unknown> {
  id: Id;
  reason: string;
  amount: number;
  currency: string;
  createdAt: string;
}

export interface CourierDetail {
  id: Id;
  code: string;
  fullName: string;
  phone: string;
  email: string;
  cityId: string;
  zoneName: string;
  liveStatus: CourierLiveStatus;
  emergency: boolean;
  emergencyReason: string | null;
  rating: number;
  ratingCount: number;
  joinedAt: string;
  lastSeenAt: string;
  performance: CourierPerformance;
  vehicle: CourierVehicle;
  assignments: CourierAssignment[];
  schedule: CourierScheduleSlot[];
  ratings: CourierRating[];
  documents: CourierDocument[];
  payments: CourierPayment[];
  bonuses: CourierBonus[];
  penalties: CourierPenalty[];
}

export interface CourierListResponse {
  items: CourierListItem[];
  total: number;
  generatedAt: string;
}
