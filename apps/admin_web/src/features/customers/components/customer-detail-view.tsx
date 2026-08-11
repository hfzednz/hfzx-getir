"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Button,
  ConfirmDialog,
  EmptyState,
  Input,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  Textarea,
} from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { formatMinorUnits } from "@/shared/lib/money";
import { can } from "@/shared/permissions/permissions";
import { useCustomerAdjustment, useCustomerProfile } from "../hooks";

function riskTone(score: number): "success" | "warning" | "danger" {
  if (score >= 60) return "danger";
  if (score >= 30) return "warning";
  return "success";
}

export function CustomerDetailView({ customerId }: { customerId: string }) {
  const session = useAuthStore((s) => s.session);
  const canWrite = can(session, "customers:write");
  const { data, isLoading, isError, error, refetch } =
    useCustomerProfile(customerId);
  const adjust = useCustomerAdjustment(customerId);

  const [walletAmount, setWalletAmount] = useState("50");
  const [loyaltyPoints, setLoyaltyPoints] = useState("100");
  const [noteBody, setNoteBody] = useState("");
  const [pending, setPending] = useState<
    null | "wallet" | "loyalty" | "note"
  >(null);
  const [flash, setFlash] = useState<string | null>(null);

  const confirmAdjust = async () => {
    if (!pending) return;
    const result = await adjust.mutateAsync(
      pending === "wallet"
        ? {
            type: "wallet",
            amountMinor: Math.round(Number(walletAmount) * 100),
            note: noteBody || "Manual wallet adjustment",
          }
        : pending === "loyalty"
          ? {
              type: "loyalty",
              points: Number(loyaltyPoints),
              note: noteBody || "Manual loyalty adjustment",
            }
          : {
              type: "note",
              note: noteBody || "Operator note",
            },
    );
    setFlash(result.message);
    setPending(null);
    setNoteBody("");
  };

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={160} />
        <Skeleton height={320} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <EmptyState
        title="Customer not found"
        description={
          error instanceof Error ? error.message : `No customer ${customerId}`
        }
        action={
          <div className="flex gap-[var(--nx-space-2)]">
            <Button variant="secondary" size="sm" onClick={() => void refetch()}>
              Retry
            </Button>
            <Link href="/customers">
              <Button variant="primary" size="sm">
                Back to list
              </Button>
            </Link>
          </div>
        }
      />
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title={data.name}
        description={
          <span className="flex items-center gap-[var(--nx-space-2)] flex-wrap">
            <span className="tabular-nums">{data.id}</span>
            <StatusBadge
              status={data.segment.replace("_", " ")}
              tone={
                data.segment === "fraud_watch"
                  ? "danger"
                  : data.segment === "vip"
                    ? "info"
                    : "neutral"
              }
            />
            <span>{data.email}</span>
            <span>{data.phone}</span>
          </span>
        }
        actions={
          <Link href="/customers">
            <Button variant="secondary" size="sm">
              Back
            </Button>
          </Link>
        }
      />

      {flash ? (
        <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-subtle)] bg-[var(--nx-bg-surface)] p-[var(--nx-space-3)] text-[13px]">
          {flash}
        </div>
      ) : null}

      <section
        aria-label="Profile KPIs"
        className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]"
      >
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-3)]">
          <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)]">LTV</p>
          <p className="m-0 text-[16px] font-semibold tabular-nums">
            {formatMinorUnits(data.lifetimeValueMinor, data.currency)}
          </p>
        </div>
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-3)]">
          <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)]">Orders</p>
          <p className="m-0 text-[16px] font-semibold tabular-nums">
            {data.orderCount}
          </p>
        </div>
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-3)]">
          <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)]">Wallet</p>
          <p className="m-0 text-[16px] font-semibold tabular-nums">
            {formatMinorUnits(data.walletBalanceMinor, data.currency)}
          </p>
        </div>
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-3)]">
          <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)]">
            Risk / Fraud
          </p>
          <div className="flex gap-[var(--nx-space-1)] mt-[var(--nx-space-1)]">
            <StatusBadge
              status={`R ${data.riskScore}`}
              tone={riskTone(data.riskScore)}
            />
            <StatusBadge
              status={`F ${data.fraudScore}`}
              tone={riskTone(data.fraudScore)}
            />
          </div>
        </div>
      </section>

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="orders">Orders</TabsTrigger>
          <TabsTrigger value="wallet">Wallet & loyalty</TabsTrigger>
          <TabsTrigger value="support">Support & notes</TabsTrigger>
          <TabsTrigger value="adjust">Adjustments</TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)] mt-[var(--nx-space-3)]">
            <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
              <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
                Addresses
              </h3>
              <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
                {data.addresses.map((a) => (
                  <li key={a.id}>
                    <div className="flex items-center gap-[var(--nx-space-2)]">
                      <span className="text-[13px] font-semibold">{a.label}</span>
                      {a.isDefault ? (
                        <StatusBadge status="default" tone="info" />
                      ) : null}
                    </div>
                    <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                      {a.line1} · {a.district}, {a.city}
                    </p>
                  </li>
                ))}
              </ul>
            </div>

            <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
              <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
                Coupons
              </h3>
              <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
                {data.coupons.map((c) => (
                  <li
                    key={c.id}
                    className="flex items-center justify-between gap-[var(--nx-space-2)]"
                  >
                    <div>
                      <p className="m-0 text-[13px] font-semibold tabular-nums">
                        {c.code}
                      </p>
                      <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                        {c.discountLabel}
                      </p>
                    </div>
                    <StatusBadge status={c.status} />
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="orders">
          <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)] mt-[var(--nx-space-3)]">
            <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
              {data.recentOrders.map((o) => (
                <li
                  key={o.id}
                  className="flex items-center justify-between gap-[var(--nx-space-2)] py-[var(--nx-space-2)] border-b border-[var(--nx-border-subtle)] last:border-0"
                >
                  <div>
                    <Link
                      href={`/orders/${o.id}`}
                      className="text-[13px] font-semibold tabular-nums text-[var(--nx-text-link)]"
                    >
                      {o.id}
                    </Link>
                    <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)]">
                      {new Date(o.createdAt).toLocaleString("tr-TR")}
                    </p>
                  </div>
                  <div className="flex items-center gap-[var(--nx-space-2)]">
                    <StatusBadge status={o.status.replace("_", " ")} />
                    <span className="tabular-nums text-[13px]">
                      {formatMinorUnits(o.totalMinor, o.currency)}
                    </span>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </TabsContent>

        <TabsContent value="wallet">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)] mt-[var(--nx-space-3)]">
            <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
              <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
                Wallet transactions
              </h3>
              <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
                {data.walletTxns.map((t) => (
                  <li
                    key={t.id}
                    className="flex items-center justify-between gap-[var(--nx-space-2)]"
                  >
                    <div>
                      <p className="m-0 text-[13px] font-medium">{t.note}</p>
                      <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)]">
                        {t.type} · {new Date(t.at).toLocaleDateString("tr-TR")}
                      </p>
                    </div>
                    <span className="tabular-nums text-[13px]">
                      {t.type === "debit" ? "−" : "+"}
                      {formatMinorUnits(t.amountMinor, t.currency)}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
            <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
              <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
                Loyalty
              </h3>
              <p className="m-0 text-[13px]">
                Tier <strong>{data.loyalty.tier}</strong>
              </p>
              <p className="m-0 text-[13px] tabular-nums">
                {data.loyalty.points.toLocaleString("tr-TR")} points
              </p>
              <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                {data.loyalty.pointsToNextTier > 0
                  ? `${data.loyalty.pointsToNextTier} to next tier`
                  : "Top tier reached"}
              </p>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="support">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)] mt-[var(--nx-space-3)]">
            <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
              <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
                Support history
              </h3>
              <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
                {data.supportHistory.map((t) => (
                  <li key={t.id}>
                    <div className="flex items-center gap-[var(--nx-space-2)]">
                      <span className="text-[13px] font-semibold tabular-nums">
                        {t.id}
                      </span>
                      <StatusBadge status={t.status} />
                    </div>
                    <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                      {t.subject}
                    </p>
                  </li>
                ))}
              </ul>
            </div>
            <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
              <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
                Notes
              </h3>
              <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
                {data.notes.map((n) => (
                  <li key={n.id}>
                    <p className="m-0 text-[13px]">{n.body}</p>
                    <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)]">
                      {n.author} ·{" "}
                      {new Date(n.createdAt).toLocaleDateString("tr-TR")}
                    </p>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="adjust">
          <PermissionGate
            allowed={canWrite}
            fallback={
              <EmptyState
                title="Permission required"
                description="customers:write is needed for manual adjustments."
              />
            }
          >
            <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)] mt-[var(--nx-space-3)] flex flex-col gap-[var(--nx-space-3)] max-w-xl">
              <div>
                <label className="block text-[12px] font-medium mb-[var(--nx-space-1)]">
                  Wallet credit (TRY)
                </label>
                <div className="flex gap-[var(--nx-space-2)]">
                  <Input
                    type="number"
                    value={walletAmount}
                    onChange={(e) => setWalletAmount(e.target.value)}
                  />
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => setPending("wallet")}
                  >
                    Adjust wallet
                  </Button>
                </div>
              </div>
              <div>
                <label className="block text-[12px] font-medium mb-[var(--nx-space-1)]">
                  Loyalty points
                </label>
                <div className="flex gap-[var(--nx-space-2)]">
                  <Input
                    type="number"
                    value={loyaltyPoints}
                    onChange={(e) => setLoyaltyPoints(e.target.value)}
                  />
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => setPending("loyalty")}
                  >
                    Adjust points
                  </Button>
                </div>
              </div>
              <div>
                <label className="block text-[12px] font-medium mb-[var(--nx-space-1)]">
                  Note / audit comment
                </label>
                <Textarea
                  value={noteBody}
                  onChange={(e) => setNoteBody(e.target.value)}
                  rows={3}
                />
                <Button
                  className="mt-[var(--nx-space-2)]"
                  variant="primary"
                  size="sm"
                  onClick={() => setPending("note")}
                >
                  Add note
                </Button>
              </div>
            </div>
          </PermissionGate>
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        open={pending != null}
        title={
          pending === "wallet"
            ? "Confirm wallet adjustment"
            : pending === "loyalty"
              ? "Confirm loyalty adjustment"
              : "Add customer note"
        }
        description={
          pending === "wallet"
            ? `Credit ${walletAmount} TRY to wallet.`
            : pending === "loyalty"
              ? `Add ${loyaltyPoints} loyalty points.`
              : "Save operator note to customer profile."
        }
        loading={adjust.isPending}
        onConfirm={() => void confirmAdjust()}
        onCancel={() => setPending(null)}
      />
    </div>
  );
}
