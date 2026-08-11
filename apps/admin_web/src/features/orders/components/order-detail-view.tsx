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
  Textarea,
} from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { formatMinorUnits } from "@/shared/lib/money";
import {
  canCancelOrder,
  canForceCancel,
  canForceComplete,
  canRefund,
  canReassign,
  canReplaceOrder,
} from "@/shared/permissions/order-rules";
import { useOrderAction, useOrderDetail } from "../hooks";
import type { OrderActionType, OrderStatus } from "../types";

function statusTone(
  status: OrderStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "delivered":
      return "success";
    case "cancelled":
    case "failed":
    case "refunded":
      return "danger";
    case "en_route":
    case "assigned":
      return "info";
    case "picking":
    case "ready":
    case "confirmed":
      return "warning";
    default:
      return "neutral";
  }
}

interface PendingAction {
  action: OrderActionType;
  title: string;
  description: string;
  danger?: boolean;
}

export function OrderDetailView({ orderId }: { orderId: string }) {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch } = useOrderDetail(orderId);
  const actionMut = useOrderAction(orderId);
  const [pending, setPending] = useState<PendingAction | null>(null);
  const [reason, setReason] = useState("");
  const [courierId, setCourierId] = useState("");
  const [refundMinor, setRefundMinor] = useState("");
  const [flash, setFlash] = useState<string | null>(null);

  const runAction = async () => {
    if (!pending) return;
    if (pending.action === "reassign" && !courierId.trim()) {
      setFlash("Courier ID is required for reassign");
      return;
    }
    let refundAmount: number | undefined;
    if (pending.action === "refund") {
      const parsed = Number.parseInt(refundMinor, 10);
      if (!Number.isFinite(parsed) || parsed <= 0) {
        setFlash("Refund amount (minor units) is required");
        return;
      }
      refundAmount = parsed;
    }
    const result = await actionMut.mutateAsync({
      action: pending.action,
      reason: reason || pending.description,
      courierId: pending.action === "reassign" ? courierId.trim() : undefined,
      refundMinor: refundAmount,
    });
    setFlash(result.message);
    setPending(null);
    setReason("");
    setCourierId("");
    setRefundMinor("");
  };

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={200} />
        <Skeleton height={280} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <EmptyState
        title="Order not found"
        description={
          error instanceof Error ? error.message : `No order ${orderId}`
        }
        action={
          <div className="flex gap-[var(--nx-space-2)]">
            <Button variant="secondary" size="sm" onClick={() => void refetch()}>
              Retry
            </Button>
            <Link href="/orders">
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
        title={data.id}
        description={
          <span className="flex items-center gap-[var(--nx-space-2)] flex-wrap">
            <span>{data.externalRef}</span>
            <StatusBadge
              status={data.status.replace("_", " ")}
              tone={statusTone(data.status)}
            />
            <StatusBadge status={data.paymentStatus} />
          </span>
        }
        actions={
          <Link href="/orders">
            <Button variant="secondary" size="sm">
              Back
            </Button>
          </Link>
        }
      />

      {flash ? (
        <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-subtle)] bg-[var(--nx-success-surface,var(--nx-bg-surface))] p-[var(--nx-space-3)] text-[13px]">
          {flash}
        </div>
      ) : null}

      <section
        aria-label="Order summary"
        className="grid grid-cols-1 md:grid-cols-3 gap-[var(--nx-space-3)]"
      >
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
            Customer
          </h3>
          <p className="m-0 text-[13px] font-medium">{data.customerName}</p>
          <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
            {data.customerPhone}
          </p>
          <Link
            href={`/customers/${data.customerId}`}
            className="text-[12px] text-[var(--nx-text-link)]"
          >
            Open profile
          </Link>
          <p className="m-0 mt-[var(--nx-space-2)] text-[12px] text-[var(--nx-text-secondary)]">
            {data.address}
          </p>
        </div>

        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
            Fulfillment
          </h3>
          <p className="m-0 text-[13px]">
            {data.warehouseCode} · {data.zone}
          </p>
          <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
            Courier: {data.courierName ?? "Unassigned"}
          </p>
          <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
            Items: {data.itemCount}
            {data.delayMinutes > 0 ? ` · delay +${data.delayMinutes}m` : ""}
          </p>
          {data.notes ? (
            <p className="m-0 mt-[var(--nx-space-2)] text-[12px] text-[var(--nx-warning)]">
              {data.notes}
            </p>
          ) : null}
        </div>

        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-2)] text-[var(--nx-font-size-title)] font-semibold">
            Payment
          </h3>
          <p className="m-0 text-[18px] font-semibold tabular-nums">
            {formatMinorUnits(data.totalMinor, data.currency)}
          </p>
          <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
            Refunded{" "}
            {formatMinorUnits(data.refundedMinor, data.currency)}
          </p>
        </div>
      </section>

      <section
        aria-label="Actions"
        className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]"
      >
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Interventions
        </h3>
        <div className="flex flex-wrap gap-[var(--nx-space-2)]">
          <PermissionGate allowed={canReassign(session)}>
            <Button
              variant="secondary"
              size="sm"
              onClick={() =>
                setPending({
                  action: "reassign",
                  title: "Reassign courier",
                  description: `Reassign ${data.id} to a selected courier.`,
                })
              }
            >
              Reassign
            </Button>
          </PermissionGate>
          <PermissionGate allowed={canCancelOrder(session)}>
            <Button
              variant="danger"
              size="sm"
              onClick={() =>
                setPending({
                  action: "cancel",
                  title: "Cancel order",
                  description: `Cancel ${data.id} and notify customer.`,
                  danger: true,
                })
              }
            >
              Cancel
            </Button>
          </PermissionGate>
          <PermissionGate allowed={canRefund(session)}>
            <Button
              variant="secondary"
              size="sm"
              onClick={() =>
                setPending({
                  action: "refund",
                  title: "Issue refund",
                  description: `Refund part or all of ${formatMinorUnits(data.totalMinor, data.currency)} (enter minor units).`,
                  danger: true,
                })
              }
            >
              Refund
            </Button>
          </PermissionGate>
          <PermissionGate allowed={canReplaceOrder(session)}>
            <Button
              variant="secondary"
              size="sm"
              onClick={() =>
                setPending({
                  action: "replace",
                  title: "Replace / recreate",
                  description: `Create a replacement order from ${data.id}.`,
                })
              }
            >
              Replace
            </Button>
          </PermissionGate>
          <PermissionGate allowed={canForceComplete(session)}>
            <Button
              variant="accent"
              size="sm"
              onClick={() =>
                setPending({
                  action: "force_complete",
                  title: "Force complete",
                  description: `Force-complete ${data.id}. Audited admin action.`,
                  danger: true,
                })
              }
            >
              Force complete
            </Button>
          </PermissionGate>
          <PermissionGate allowed={canForceCancel(session)}>
            <Button
              variant="danger"
              size="sm"
              onClick={() =>
                setPending({
                  action: "force_cancel",
                  title: "Force cancel",
                  description: `Force-cancel ${data.id}. Audited admin action.`,
                  danger: true,
                })
              }
            >
              Force cancel
            </Button>
          </PermissionGate>
        </div>
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            Line items
          </h3>
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
            {data.lines.map((line) => (
              <li
                key={line.id}
                className="flex items-center justify-between gap-[var(--nx-space-2)] py-[var(--nx-space-1)] border-b border-[var(--nx-border-subtle)] last:border-0"
              >
                <div>
                  <p className="m-0 text-[13px] font-medium">{line.name}</p>
                  <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)]">
                    {line.sku} · ×{line.qty}
                  </p>
                </div>
                <span className="tabular-nums text-[13px]">
                  {formatMinorUnits(line.totalMinor, data.currency)}
                </span>
              </li>
            ))}
          </ul>
        </div>

        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            Timeline
          </h3>
          <ol className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)] relative">
            {data.timeline.map((evt, idx) => (
              <li key={evt.id} className="flex gap-[var(--nx-space-3)]">
                <div className="flex flex-col items-center">
                  <span className="w-2.5 h-2.5 rounded-full bg-[var(--nx-brand)] mt-1" />
                  {idx < data.timeline.length - 1 ? (
                    <span className="w-px flex-1 bg-[var(--nx-border-subtle)] my-1" />
                  ) : null}
                </div>
                <div className="pb-[var(--nx-space-2)]">
                  <p className="m-0 text-[13px] font-semibold">{evt.title}</p>
                  {evt.detail ? (
                    <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                      {evt.detail}
                    </p>
                  ) : null}
                  <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)] tabular-nums">
                    {new Date(evt.at).toLocaleString("tr-TR")}
                    {evt.actor ? ` · ${evt.actor}` : ""}
                  </p>
                </div>
              </li>
            ))}
          </ol>
        </div>
      </section>

      <ConfirmDialog
        open={pending != null}
        title={pending?.title ?? ""}
        description={pending?.description}
        confirmLabel="Confirm"
        danger={pending?.danger}
        loading={actionMut.isPending}
        onConfirm={() => void runAction()}
        onCancel={() => {
          setPending(null);
          setReason("");
        }}
      />
      {pending ? (
        <div className="fixed bottom-4 right-4 z-50 w-[min(360px,calc(100vw-2rem))] bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-3)] shadow-lg flex flex-col gap-[var(--nx-space-2)]">
          <label className="block text-[12px] font-medium">
            Audit reason
          </label>
          <Textarea
            placeholder="Reason / audit note"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            rows={2}
          />
          {pending.action === "reassign" ? (
            <>
              <label className="block text-[12px] font-medium">Courier ID</label>
              <Input
                value={courierId}
                onChange={(e) => setCourierId(e.target.value)}
                placeholder="courier opaque id"
                aria-label="Courier ID"
              />
            </>
          ) : null}
          {pending.action === "refund" ? (
            <>
              <label className="block text-[12px] font-medium">
                Refund amount (minor units)
              </label>
              <Input
                value={refundMinor}
                onChange={(e) => setRefundMinor(e.target.value)}
                placeholder={String(data.totalMinor)}
                inputMode="numeric"
                aria-label="Refund minor units"
              />
            </>
          ) : null}
        </div>
      ) : null}    </div>
  );
}
