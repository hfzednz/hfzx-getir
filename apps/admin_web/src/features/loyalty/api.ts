import type { LoyaltySnapshot } from "./types";
import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient } from "@/shared/api/client";

export async function fetchLoyaltySnapshot(): Promise<LoyaltySnapshot> {
  try {
    return await apiClient<LoyaltySnapshot>("/admin/loyalty");
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    return mockLoyaltySnapshot();
  }
}

async function mockLoyaltySnapshot(): Promise<LoyaltySnapshot> {
  return {
    totalMembers: 1_842_300,
    pointsIssued: 94_200_000,
    pointsRedeemed: 41_800_000,
    levels: [
      {
        id: "lvl_bronze",
        name: "Bronze",
        tier: "bronze",
        minPoints: 0,
        multiplier: 1,
        memberCount: 920000,
      },
      {
        id: "lvl_silver",
        name: "Silver",
        tier: "silver",
        minPoints: 500,
        multiplier: 1.1,
        memberCount: 540000,
      },
      {
        id: "lvl_gold",
        name: "Gold",
        tier: "gold",
        minPoints: 2000,
        multiplier: 1.25,
        memberCount: 280000,
      },
      {
        id: "lvl_plat",
        name: "Platinum",
        tier: "platinum",
        minPoints: 8000,
        multiplier: 1.5,
        memberCount: 82000,
      },
      {
        id: "lvl_vip",
        name: "VIP",
        tier: "vip",
        minPoints: 20000,
        multiplier: 2,
        memberCount: 20300,
      },
    ],
    rewards: [
      {
        id: "rw_1",
        title: "Free delivery",
        pointsCost: 400,
        stock: 50000,
        redeemed: 12840,
        active: true,
      },
      {
        id: "rw_2",
        title: "TRY30 voucher",
        pointsCost: 900,
        stock: 20000,
        redeemed: 6100,
        active: true,
      },
      {
        id: "rw_3",
        title: "Snack bundle",
        pointsCost: 1200,
        stock: 3000,
        redeemed: 980,
        active: false,
      },
    ],
    cashback: [
      {
        id: "cb_1",
        name: "Standard cashback",
        ratePct: 1.5,
        capMinor: 5000,
        currency: "TRY",
        active: true,
      },
      {
        id: "cb_2",
        name: "VIP cashback",
        ratePct: 3,
        capMinor: 15000,
        currency: "TRY",
        active: true,
      },
    ],
    referral: {
      id: "ref_1",
      referrerBonusPoints: 300,
      refereeBonusPoints: 200,
      conversions: 18420,
      active: true,
    },
    vipBenefits: [
      {
        id: "vip_1",
        title: "Priority support",
        description: "Skip queue for live chat",
        tier: "vip",
      },
      {
        id: "vip_2",
        title: "Exclusive flash access",
        description: "Early access to flash sales",
        tier: "platinum",
      },
    ],
    achievements: [
      {
        id: "ach_1",
        title: "First 10 orders",
        description: "Complete 10 deliveries",
        unlockedCount: 210000,
        pointsReward: 150,
      },
      {
        id: "ach_2",
        title: "Weekend warrior",
        description: "Order 4 weekends in a row",
        unlockedCount: 42000,
        pointsReward: 250,
      },
    ],
    challenges: [
      {
        id: "ch_1",
        title: "August grocery streak",
        goalLabel: "5 grocery orders this month",
        progressPct: 62,
        endsAt: "2026-08-31T21:00:00Z",
        active: true,
      },
      {
        id: "ch_2",
        title: "Healthy basket",
        goalLabel: "3 produce categories",
        progressPct: 38,
        endsAt: "2026-08-15T21:00:00Z",
        active: true,
      },
    ],
  };
}
