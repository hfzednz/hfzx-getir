"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import {
  Button,
  DataGrid,
  type DataGridColumnDef,
  FilterBar,
  Input,
  PageHeader,
  PermissionGate,
  Select,
  Skeleton,
  StatusBadge,
  Textarea,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { usePermission } from "@/shared/permissions/use-permission";
import { useAddCrmNote, useCrmWorkspace } from "../hooks";
import type { CrmCustomer } from "../types";

export function CrmView() {
  const canWrite = usePermission("crm:write");
  const [q, setQ] = useState("");
  const [tag, setTag] = useState<string>("all");
  const [segment, setSegment] = useState<string>("all");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [note, setNote] = useState("");

  const { data, isLoading, isError, error, refetch } = useCrmWorkspace({
    q,
    tag,
    segment,
  });
  const addNote = useAddCrmNote();

  const selected = data?.customers.find((c) => c.id === selectedId) ?? null;

  const columns = useMemo<
    DataGridColumnDef<CrmCustomer & Record<string, unknown>>[]
  >(
    () => [
      {
        id: "name",
        header: "Customer",
        cell: ({ row }) => (
          <div className="flex flex-col gap-0.5">
            <span className="font-semibold text-[13px]">{row.name}</span>
            <span className="text-[11px] text-[var(--nx-text-tertiary)]">
              {row.email}
            </span>
          </div>
        ),
      },
      {
        id: "ltv",
        header: "LTV",
        align: "right",
        cell: ({ row }) =>
          formatMinorUnits(row.lifetimeValueMinor, row.currency),
      },
      {
        id: "orders",
        header: "Orders",
        align: "right",
        cell: ({ row }) => row.orderCount,
      },
      {
        id: "tags",
        header: "Tags",
        cell: ({ row }) => (
          <span className="text-[12px]">
            {row.tags.map((t) => t.label).join(", ") || "—"}
          </span>
        ),
      },
      {
        id: "risk",
        header: "Risk",
        align: "right",
        cell: ({ row }) => (
          <span className="tabular-nums text-[12px]">
            {(row.riskScore * 100).toFixed(0)}%
          </span>
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
          Failed to load CRM
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
        title="CRM"
        description="Customer 360 · tags, segments, notes, channel history"
      />

      <FilterBar>
        <Input
          placeholder="Search name, email, phone…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
        <Select value={tag} onChange={(e) => setTag(e.target.value)}>
          <option value="all">All tags</option>
          {data.tags.map((t) => (
            <option key={t.id} value={t.id}>
              {t.label}
            </option>
          ))}
        </Select>
        <Select value={segment} onChange={(e) => setSegment(e.target.value)}>
          <option value="all">All segments</option>
          {data.segments.map((s) => (
            <option key={s.id} value={s.name}>
              {s.name}
            </option>
          ))}
        </Select>
      </FilterBar>

      <div className="grid grid-cols-1 xl:grid-cols-[1.4fr_1fr] gap-[var(--nx-space-3)]">
        <DataGrid
          columns={columns}
          data={data.customers as (CrmCustomer & Record<string, unknown>)[]}
          getRowId={(r) => r.id}
          onRowClick={(row) => setSelectedId(row.id)}
          emptyMessage="No customers"
        />

        <aside className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)] min-h-[320px]">
          {!selected ? (
            <p className="m-0 text-[13px] text-[var(--nx-text-secondary)]">
              Select a customer for 360° view
            </p>
          ) : (
            <div className="flex flex-col gap-[var(--nx-space-3)]">
              <div>
                <h2 className="m-0 text-[15px] font-semibold">{selected.name}</h2>
                <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                  {selected.phone} · {selected.email}
                </p>
              </div>

              <div>
                <h3 className="m-0 mb-2 text-[12px] uppercase tracking-wide text-[var(--nx-text-tertiary)]">
                  Segments
                </h3>
                <p className="m-0 text-[13px]">
                  {selected.segments.join(", ") || "—"}
                </p>
              </div>

              <div>
                <h3 className="m-0 mb-2 text-[12px] uppercase tracking-wide text-[var(--nx-text-tertiary)]">
                  Channel history
                </h3>
                <ul className="m-0 p-0 list-none flex flex-col gap-2">
                  {selected.channelHistory.map((ev) => (
                    <li key={ev.id} className="text-[12px]">
                      <div className="flex items-center gap-2">
                        <StatusBadge status={ev.channel} tone="info" />
                        <span className="font-semibold">{ev.subject}</span>
                      </div>
                      <p className="m-0 text-[var(--nx-text-secondary)]">
                        {ev.preview}
                      </p>
                      {ev.campaignId ? (
                        <Link
                          href={`/campaigns/${ev.campaignId}`}
                          className="text-[var(--nx-text-link)]"
                        >
                          Campaign {ev.campaignId}
                        </Link>
                      ) : null}
                    </li>
                  ))}
                </ul>
              </div>

              <div>
                <h3 className="m-0 mb-2 text-[12px] uppercase tracking-wide text-[var(--nx-text-tertiary)]">
                  Notes
                </h3>
                <ul className="m-0 p-0 list-none flex flex-col gap-2 mb-2">
                  {selected.notes.map((n) => (
                    <li key={n.id} className="text-[12px]">
                      <span className="font-semibold">{n.author}</span>
                      <span className="text-[var(--nx-text-tertiary)]">
                        {" "}
                        · {new Date(n.createdAt).toLocaleString("tr-TR")}
                      </span>
                      <p className="m-0">{n.body}</p>
                    </li>
                  ))}
                </ul>
                <PermissionGate allowed={canWrite}>
                  <Textarea
                    rows={3}
                    value={note}
                    onChange={(e) => setNote(e.target.value)}
                    placeholder="Add note…"
                  />
                  <Button
                    className="mt-2"
                    size="sm"
                    loading={addNote.isPending}
                    onClick={() => {
                      if (!note.trim()) return;
                      void addNote
                        .mutateAsync({
                          customerId: selected.id,
                          body: note.trim(),
                        })
                        .then(() => setNote(""));
                    }}
                  >
                    Save note
                  </Button>
                </PermissionGate>
              </div>

              {selected.linkedCampaignIds.length > 0 ? (
                <div>
                  <h3 className="m-0 mb-2 text-[12px] uppercase tracking-wide text-[var(--nx-text-tertiary)]">
                    Linked campaigns
                  </h3>
                  <div className="flex flex-wrap gap-2">
                    {selected.linkedCampaignIds.map((cid) => (
                      <Link
                        key={cid}
                        href={`/campaigns/${cid}`}
                        className="text-[12px] text-[var(--nx-text-link)]"
                      >
                        {cid}
                      </Link>
                    ))}
                  </div>
                </div>
              ) : null}
            </div>
          )}
        </aside>
      </div>

      <section className="grid grid-cols-1 md:grid-cols-3 gap-[var(--nx-space-3)]">
        {data.segments.map((s) => (
          <div
            key={s.id}
            className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-3"
          >
            <p className="m-0 font-semibold text-[13px]">{s.name}</p>
            <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
              {s.rulesSummary}
            </p>
            <p className="m-0 mt-1 text-[12px] tabular-nums">
              {s.size.toLocaleString("tr-TR")} customers
            </p>
          </div>
        ))}
      </section>
    </div>
  );
}
