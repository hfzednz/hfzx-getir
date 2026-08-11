"use client";

import {
  DataGrid,
  type DataGridColumnDef,
  KpiCard,
  PageHeader,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@nexora/ui";
import { useLoyaltySnapshot } from "../hooks";
import type {
  LoyaltyAchievement,
  LoyaltyChallenge,
  LoyaltyLevel,
  LoyaltyReward,
  VipBenefit,
} from "../types";

export function LoyaltyView() {
  const { data, isLoading, isError, error, refetch } = useLoyaltySnapshot();

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={320} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load loyalty
        </p>
        <p className="m-0 mt-1 text-[var(--nx-text-secondary)]">
          {error instanceof Error ? error.message : "Unknown error"}
        </p>
        <button
          type="button"
          onClick={() => void refetch()}
          className="mt-3 text-[var(--nx-text-link)] underline cursor-pointer bg-transparent border-0"
        >
          Retry
        </button>
      </div>
    );
  }

  const levelCols: DataGridColumnDef<LoyaltyLevel & Record<string, unknown>>[] =
    [
      { id: "name", header: "Level", accessorKey: "name" },
      { id: "tier", header: "Tier", accessorKey: "tier" },
      {
        id: "min",
        header: "Min points",
        align: "right",
        cell: ({ row }) => row.minPoints.toLocaleString("tr-TR"),
      },
      {
        id: "mult",
        header: "Multiplier",
        align: "right",
        cell: ({ row }) => `${row.multiplier}×`,
      },
      {
        id: "members",
        header: "Members",
        align: "right",
        cell: ({ row }) => row.memberCount.toLocaleString("tr-TR"),
      },
    ];

  const rewardCols: DataGridColumnDef<LoyaltyReward & Record<string, unknown>>[] =
    [
      { id: "title", header: "Reward", accessorKey: "title" },
      {
        id: "cost",
        header: "Points",
        align: "right",
        cell: ({ row }) => row.pointsCost.toLocaleString("tr-TR"),
      },
      {
        id: "stock",
        header: "Stock",
        align: "right",
        cell: ({ row }) =>
          `${row.redeemed.toLocaleString("tr-TR")} / ${row.stock.toLocaleString("tr-TR")}`,
      },
      {
        id: "active",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.active ? "active" : "inactive"} />
        ),
      },
    ];

  const achCols: DataGridColumnDef<
    LoyaltyAchievement & Record<string, unknown>
  >[] = [
    { id: "title", header: "Achievement", accessorKey: "title" },
    { id: "desc", header: "Description", accessorKey: "description" },
    {
      id: "unlocked",
      header: "Unlocked",
      align: "right",
      cell: ({ row }) => row.unlockedCount.toLocaleString("tr-TR"),
    },
    {
      id: "pts",
      header: "Reward pts",
      align: "right",
      cell: ({ row }) => row.pointsReward,
    },
  ];

  const challengeCols: DataGridColumnDef<
    LoyaltyChallenge & Record<string, unknown>
  >[] = [
    { id: "title", header: "Challenge", accessorKey: "title" },
    { id: "goal", header: "Goal", accessorKey: "goalLabel" },
    {
      id: "progress",
      header: "Progress",
      align: "right",
      cell: ({ row }) => `${row.progressPct}%`,
    },
    {
      id: "ends",
      header: "Ends",
      cell: ({ row }) => new Date(row.endsAt).toLocaleDateString("tr-TR"),
    },
  ];

  const vipCols: DataGridColumnDef<VipBenefit & Record<string, unknown>>[] = [
    { id: "title", header: "Benefit", accessorKey: "title" },
    { id: "desc", header: "Description", accessorKey: "description" },
    { id: "tier", header: "From tier", accessorKey: "tier" },
  ];

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Loyalty"
        description="Points, levels, rewards, cashback, referral, VIP, achievements, challenges"
      />

      <section className="grid grid-cols-1 md:grid-cols-3 gap-[var(--nx-space-3)]">
        <KpiCard
          title="Members"
          value={data.totalMembers.toLocaleString("tr-TR")}
          tone="brand"
        />
        <KpiCard
          title="Points issued"
          value={data.pointsIssued.toLocaleString("tr-TR")}
          tone="success"
        />
        <KpiCard
          title="Points redeemed"
          value={data.pointsRedeemed.toLocaleString("tr-TR")}
          tone="neutral"
        />
      </section>

      <Tabs defaultValue="levels">
        <TabsList>
          <TabsTrigger value="levels">Levels</TabsTrigger>
          <TabsTrigger value="rewards">Rewards</TabsTrigger>
          <TabsTrigger value="cashback">Cashback</TabsTrigger>
          <TabsTrigger value="referral">Referral</TabsTrigger>
          <TabsTrigger value="vip">VIP</TabsTrigger>
          <TabsTrigger value="achievements">Achievements</TabsTrigger>
          <TabsTrigger value="challenges">Challenges</TabsTrigger>
        </TabsList>

        <TabsContent value="levels">
          <DataGrid
            columns={levelCols}
            data={data.levels as (LoyaltyLevel & Record<string, unknown>)[]}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="rewards">
          <DataGrid
            columns={rewardCols}
            data={data.rewards as (LoyaltyReward & Record<string, unknown>)[]}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="cashback">
          <ul className="m-0 p-0 list-none flex flex-col gap-3">
            {data.cashback.map((c) => (
              <li
                key={c.id}
                className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-4 flex justify-between gap-3"
              >
                <div>
                  <p className="m-0 font-semibold text-[13px]">{c.name}</p>
                  <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                    {c.ratePct}% · cap {(c.capMinor / 100).toFixed(2)} {c.currency}
                  </p>
                </div>
                <StatusBadge status={c.active ? "active" : "inactive"} />
              </li>
            ))}
          </ul>
        </TabsContent>
        <TabsContent value="referral">
          <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-4 flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <p className="m-0 font-semibold">Referral program</p>
              <StatusBadge
                status={data.referral.active ? "active" : "inactive"}
              />
            </div>
            <p className="m-0 text-[13px]">
              Referrer +{data.referral.referrerBonusPoints} pts · Referee +
              {data.referral.refereeBonusPoints} pts
            </p>
            <p className="m-0 text-[12px] tabular-nums text-[var(--nx-text-secondary)]">
              Conversions: {data.referral.conversions.toLocaleString("tr-TR")}
            </p>
          </div>
        </TabsContent>
        <TabsContent value="vip">
          <DataGrid
            columns={vipCols}
            data={data.vipBenefits as (VipBenefit & Record<string, unknown>)[]}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="achievements">
          <DataGrid
            columns={achCols}
            data={
              data.achievements as (LoyaltyAchievement & Record<string, unknown>)[]
            }
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="challenges">
          <DataGrid
            columns={challengeCols}
            data={
              data.challenges as (LoyaltyChallenge & Record<string, unknown>)[]
            }
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
