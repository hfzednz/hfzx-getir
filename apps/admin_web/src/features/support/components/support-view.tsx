"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Button,
  DataGrid,
  type DataGridColumnDef,
  FilterBar,
  Input,
  KpiCard,
  PageHeader,
  Select,
  Skeleton,
  StatusBadge,
} from "@nexora/ui";
import { useSupportWorkspace } from "../hooks";
import type { SupportTicket, TicketCategory, TicketStatus } from "../types";

export function SupportView() {
  const router = useRouter();
  const [status, setStatus] = useState<TicketStatus | "all">("all");
  const [category, setCategory] = useState<TicketCategory | "all">("all");
  const [q, setQ] = useState("");

  const { data, isLoading, isError, error, refetch } = useSupportWorkspace({
    status,
    category,
    q,
  });

  const columns = useMemo<
    DataGridColumnDef<SupportTicket & Record<string, unknown>>[]
  >(
    () => [
      {
        id: "subject",
        header: "Ticket",
        cell: ({ row }) => (
          <div className="flex flex-col gap-0.5">
            <span className="font-semibold text-[13px]">{row.subject}</span>
            <span className="text-[11px] text-[var(--nx-text-tertiary)] tabular-nums">
              {row.id} · {row.customerName}
            </span>
          </div>
        ),
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => <StatusBadge status={row.status} />,
      },
      {
        id: "priority",
        header: "Priority",
        cell: ({ row }) => (
          <StatusBadge
            status={row.priority}
            tone={
              row.priority === "urgent" || row.priority === "high"
                ? "danger"
                : row.priority === "medium"
                  ? "warning"
                  : "neutral"
            }
          />
        ),
      },
      {
        id: "category",
        header: "Category",
        cell: ({ row }) => (
          <span className="text-[12px] capitalize">
            {row.category.replaceAll("_", " ")}
          </span>
        ),
      },
      {
        id: "assignee",
        header: "Assignee",
        cell: ({ row }) => (
          <span className="text-[12px]">{row.assignee ?? "Unassigned"}</span>
        ),
      },
    ],
    [],
  );

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
          Failed to load support
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
        title="Support"
        description="Tickets, live chat, AI chatbot, escalations, complaints, refunds, QC"
      />

      <section className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
        <KpiCard
          title="Open tickets"
          value={String(
            data.tickets.filter((t) => t.status !== "closed" && t.status !== "resolved")
              .length,
          )}
          tone="warning"
        />
        <KpiCard
          title="Complaints"
          value={String(data.complaintCount)}
          tone="danger"
        />
        <KpiCard
          title="Open refunds"
          value={String(data.openRefunds)}
          tone="neutral"
        />
        <KpiCard
          title="Live chat queue"
          value={String(data.liveChat.queued)}
          delta={`${data.liveChat.agentsOnline} agents · ${data.liveChat.avgWaitSec}s wait`}
          tone="brand"
        />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-3 text-[var(--nx-font-size-title)] font-semibold">
            Live chat (stub)
          </h3>
          <p className="m-0 text-[13px]">
            Active sessions:{" "}
            <span className="tabular-nums font-semibold">
              {data.liveChat.activeSessions}
            </span>
          </p>
          <p className="m-0 mt-1 text-[12px] text-[var(--nx-text-secondary)]">
            Realtime panel wires to WS gateway when bff-admin chat stream is live.
          </p>
          <Button size="sm" variant="secondary" className="mt-3" disabled>
            Open console (coming soon)
          </Button>
        </div>
        <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-3 text-[var(--nx-font-size-title)] font-semibold">
            AI chatbot
          </h3>
          <div className="flex items-center gap-2 mb-2">
            <StatusBadge
              status={data.aiChatbot.enabled ? "active" : "inactive"}
            />
            <span className="text-[12px] tabular-nums">
              Containment {data.aiChatbot.containmentRatePct}% · Handoff{" "}
              {data.aiChatbot.handoffRatePct}%
            </span>
          </div>
          <ul className="m-0 p-0 list-none flex flex-col gap-1">
            {data.aiChatbot.topIntents.map((i) => (
              <li
                key={i.intent}
                className="flex justify-between text-[12px] border-b border-[var(--nx-border-subtle)] py-1 last:border-0"
              >
                <span>{i.intent}</span>
                <span className="tabular-nums">{i.count}</span>
              </li>
            ))}
          </ul>
        </div>
      </section>

      <FilterBar>
        <Input
          placeholder="Search tickets…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
        <Select
          value={status}
          onChange={(e) => setStatus(e.target.value as TicketStatus | "all")}
        >
          <option value="all">All statuses</option>
          <option value="open">open</option>
          <option value="pending">pending</option>
          <option value="escalated">escalated</option>
          <option value="resolved">resolved</option>
          <option value="closed">closed</option>
        </Select>
        <Select
          value={category}
          onChange={(e) =>
            setCategory(e.target.value as TicketCategory | "all")
          }
        >
          <option value="all">All categories</option>
          <option value="order">order</option>
          <option value="refund">refund</option>
          <option value="complaint">complaint</option>
          <option value="delivery">delivery</option>
          <option value="product_qc">product_qc</option>
          <option value="other">other</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={data.tickets as (SupportTicket & Record<string, unknown>)[]}
        getRowId={(r) => r.id}
        onRowClick={(row) => router.push(`/support/tickets/${row.id}`)}
        emptyMessage="No tickets"
      />
    </div>
  );
}
