import type { Id } from "@/shared/types/common";

export type LoyaltyTier = "bronze" | "silver" | "gold" | "platinum" | "vip";

export interface LoyaltyLevel {
  id: Id;
  name: string;
  tier: LoyaltyTier;
  minPoints: number;
  multiplier: number;
  memberCount: number;
}

export interface LoyaltyReward {
  id: Id;
  title: string;
  pointsCost: number;
  stock: number;
  redeemed: number;
  active: boolean;
}

export interface CashbackRule {
  id: Id;
  name: string;
  ratePct: number;
  capMinor: number;
  currency: string;
  active: boolean;
}

export interface ReferralProgram {
  id: Id;
  referrerBonusPoints: number;
  refereeBonusPoints: number;
  conversions: number;
  active: boolean;
}

export interface VipBenefit {
  id: Id;
  title: string;
  description: string;
  tier: LoyaltyTier;
}

export interface LoyaltyAchievement {
  id: Id;
  title: string;
  description: string;
  unlockedCount: number;
  pointsReward: number;
}

export interface LoyaltyChallenge {
  id: Id;
  title: string;
  goalLabel: string;
  progressPct: number;
  endsAt: string;
  active: boolean;
}

export interface LoyaltySnapshot {
  totalMembers: number;
  pointsIssued: number;
  pointsRedeemed: number;
  levels: LoyaltyLevel[];
  rewards: LoyaltyReward[];
  cashback: CashbackRule[];
  referral: ReferralProgram;
  vipBenefits: VipBenefit[];
  achievements: LoyaltyAchievement[];
  challenges: LoyaltyChallenge[];
}
