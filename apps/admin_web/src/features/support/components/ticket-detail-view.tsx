"use client";

import Link from "next/link";
import {
  Button,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { usePermission } from "@/shared/permissions/use-permission";
import {
  useApproveTicketRefund,
  useEscalateTicket,
  useResolveTicket,
  useTicket,
} from "../hooks";

export function TicketDetailView({ id }: { id: string }) {
  const canEscalate = usePermission("support:escalate");
  const canWrite = usePermission("support:write");
  const canRefund = usePermission("orders:refund");

  const { data, isLoading, isError, error, refetch } = useTicket(id);
  const escalateMut = useEscalateTicket(id);
  const resolveMut = useResolveTicket(id);
  const refundMut = useApproveTicketRefund(id);

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={280} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load ticket
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

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title={data.subject}
        description={
          <span className="flex items-center gap-2 flex-wrap">
            <Link href="/support" className="text-[var(--nx-text-link)] no-underline">
              Support
            </Link>
            <span>·</span>
            <span className="tabular-nums">{data.id}</span>
            <StatusBadge status={data.status} />
            <StatusBadge status={data.priority} />
          </span>
        }
        actions={
          <div className="flex flex-wrap gap-2">
            <PermissionGate allowed={canEscalate}>
              <Button
                size="sm"
                variant="secondary"
                loading={escalateMut.isPending}
                disabled={data.escalated}
                onClick={() => void escalateMut.mutateAsync()}
              >
                Escalate
              </Button>
            </PermissionGate>
            <PermissionGate allowed={canWrite}>
              <Button
                size="sm"
                loading={resolveMut.isPending}
                onClick={() => void resolveMut.mutateAsync()}
              >
                Resolve
              </Button>
            </PermissionGate>
          </div>
        }
      />

      <section className="grid grid-cols-1 lg:grid-cols-[1.4fr_1fr] gap-[var(--nx-space-3)]">
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-3 text-[var(--nx-font-size-title)] font-semibold">
            Conversation
          </h3>
          <ul className="m-0 p-0 list-none flex flex-col gap-3">
            {data.messages.map((m) => (
              <li key={m.id} className="flex flex-col gap-1">
                <div className="flex items-center gap-2">
                  <span className="text-[12px] font-semibold">{m.author}</span>
                  <StatusBadge status={m.role} tone="info" withDot={false} />
                  <span className="text-[11px] text-[var(--nx-text-tertiary)]">
                    {new Date(m.at).toLocaleString("tr-TR")}
                  </span>
                </div>
                <p className="m-0 text-[13px]">{m.body}</p>
              </li>
            ))}
          </ul>
          {data.aiSuggestedReply ? (
            <div className="mt-4 p-3 rounded-[var(--nx-radius-sm)] bg-[var(--nx-bg-sunken)] border border-[var(--nx-border-subtle)]">
              <p className="m-0 text-[11px] uppercase tracking-wide text-[var(--nx-text-tertiary)]">
                AI suggested reply
              </p>
              <p className="m-0 mt-1 text-[13px]">{data.aiSuggestedReply}</p>
            </div>
          ) : null}
        </div>

        <div className="flex flex-col gap-[var(--nx-space-3)]">
          <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
            <h3 className="m-0 mb-2 text-[var(--nx-font-size-title)] font-semibold">
              Details
            </h3>
            <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-[12px]">
              <dt className="text-[var(--nx-text-tertiary)]">Customer</dt>
              <dd className="m-0">{data.customerName}</dd>
              <dt className="text-[var(--nx-text-tertiary)]">Order</dt>
              <dd className="m-0 tabular-nums">{data.orderId ?? "—"}</dd>
              <dt className="text-[var(--nx-text-tertiary)]">Category</dt>
              <dd className="m-0 capitalize">
                {data.category.replaceAll("_", " ")}
              </dd>
              <dt className="text-[var(--nx-text-tertiary)]">Assignee</dt>
              <dd className="m-0">{data.assignee ?? "Unassigned"}</dd>
            </dl>
          </div>

          <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
            <h3 className="m-0 mb-2 text-[var(--nx-font-size-title)] font-semibold">
              Refund workflow
            </h3>
            {!data.refund ? (
              <p className="m-0 text-[13px] text-[var(--nx-text-secondary)]">
                No refund requested
              </p>
            ) : (
              <div className="flex flex-col gap-2">
                <p className="m-0 text-[13px] tabular-nums font-semibold">
                  {formatMinorUnits(data.refund.amountMinor, data.refund.currency)}
                </p>
                <p className="m-0 text-[12px]">{data.refund.reason}</p>
                <StatusBadge status={data.refund.status} />
                <PermissionGate allowed={canRefund}>
                  <Button
                    size="sm"
                    loading={refundMut.isPending}
                    disabled={data.refund.status !== "pending"}
                    onClick={() => void refundMut.mutateAsync()}
                  >
                    Approve refund
                  </Button>
                </PermissionGate>
              </div>
            )}
          </div>

          <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
            <h3 className="m-0 mb-2 text-[var(--nx-font-size-title)] font-semibold">
              QC findings
            </h3>
            {data.qc.length === 0 ? (
              <p className="m-0 text-[13px] text-[var(--nx-text-secondary)]">
                No QC issues
              </p>
            ) : (
              <ul className="m-0 p-0 list-none flex flex-col gap-2">
                {data.qc.map((q) => (
                  <li key={q.id} className="text-[12px]">
                    <span className="font-semibold tabular-nums">{q.skuId}</span>
                    {" · "}
                    {q.issue}{" "}
                    <StatusBadge
                      status={q.severity}
                      tone={
                        q.severity === "high"
                          ? "danger"
                          : q.severity === "medium"
                            ? "warning"
                            : "neutral"
                      }
                    />
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </section>
    </div>
  );
}
