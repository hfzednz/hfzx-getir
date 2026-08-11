"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
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
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { usePermission } from "@/shared/permissions/use-permission";
import { CAMPAIGN_STATUSES, CAMPAIGN_TYPES } from "../api";
import {
  useCampaigns,
  useCreateCampaign,
  useDuplicateCampaign,
} from "../hooks";
import type { Campaign, CampaignStatus, CampaignType } from "../types";

function typeLabel(t: CampaignType): string {
  return t.replaceAll("_", " ");
}

export function CampaignsView() {
  const router = useRouter();
  const canWrite = usePermission("campaigns:write");
  const [status, setStatus] = useState<CampaignStatus | "all">("all");
  const [type, setType] = useState<CampaignType | "all">("all");
  const [q, setQ] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [newName, setNewName] = useState("");
  const [newType, setNewType] = useState<CampaignType>("coupon");

  const { data, isLoading, isError, error, refetch } = useCampaigns({
    status,
    type,
    q,
  });
  const createMut = useCreateCampaign();
  const dupMut = useDuplicateCampaign();

  const columns = useMemo<DataGridColumnDef<Campaign & Record<string, unknown>>[]>(
    () => [
      {
        id: "name",
        header: "Campaign",
        cell: ({ row }) => (
          <div className="flex flex-col gap-0.5">
            <span className="font-semibold text-[13px]">{row.name}</span>
            <span className="text-[11px] text-[var(--nx-text-tertiary)] tabular-nums">
              {row.id}
            </span>
          </div>
        ),
      },
      {
        id: "type",
        header: "Type",
        cell: ({ row }) => (
          <span className="capitalize text-[12px]">{typeLabel(row.type)}</span>
        ),
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => <StatusBadge status={row.status} />,
      },
      {
        id: "budget",
        header: "Budget",
        align: "right",
        cell: ({ row }) => (
          <span className="tabular-nums text-[12px]">
            {formatMinorUnits(row.spentMinor, row.currency)} /{" "}
            {formatMinorUnits(row.budgetMinor, row.currency)}
          </span>
        ),
      },
      {
        id: "schedule",
        header: "Schedule",
        cell: ({ row }) => (
          <span className="text-[12px] text-[var(--nx-text-secondary)]">
            {row.startsAt
              ? new Date(row.startsAt).toLocaleString("tr-TR")
              : "—"}
          </span>
        ),
      },
      {
        id: "actions",
        header: "",
        cell: ({ row }) => (
          <PermissionGate allowed={canWrite}>
            <Button
              size="sm"
              variant="ghost"
              onClick={(e) => {
                e.stopPropagation();
                void dupMut.mutateAsync(row.id).then((c) => {
                  router.push(`/campaigns/${c.id}`);
                });
              }}
            >
              Duplicate
            </Button>
          </PermissionGate>
        ),
      },
    ],
    [canWrite, dupMut, router],
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
          Failed to load campaigns
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
        title="Campaigns"
        description="Coupons, bundles, flash sales, audiences, and personalized promos"
        actions={
          <PermissionGate allowed={canWrite}>
            <Button size="sm" onClick={() => setCreateOpen(true)}>
              Create campaign
            </Button>
          </PermissionGate>
        }
      />

      <FilterBar
        actions={
          <Button
            size="sm"
            variant="secondary"
            onClick={() => {
              setStatus("all");
              setType("all");
              setQ("");
            }}
          >
            Reset
          </Button>
        }
      >
        <Input
          placeholder="Search name, id, coupon…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search campaigns"
        />
        <Select
          value={status}
          onChange={(e) => setStatus(e.target.value as CampaignStatus | "all")}
          aria-label="Status"
        >
          <option value="all">All statuses</option>
          {CAMPAIGN_STATUSES.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </Select>
        <Select
          value={type}
          onChange={(e) => setType(e.target.value as CampaignType | "all")}
          aria-label="Type"
        >
          <option value="all">All types</option>
          {CAMPAIGN_TYPES.map((t) => (
            <option key={t} value={t}>
              {typeLabel(t)}
            </option>
          ))}
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={data.items as (Campaign & Record<string, unknown>)[]}
        getRowId={(row) => row.id}
        onRowClick={(row) => router.push(`/campaigns/${row.id}`)}
        emptyMessage="No campaigns match filters"
      />

      {createOpen ? (
        <div className="nx-confirm-root" role="presentation">
          <button
            type="button"
            className="nx-confirm-backdrop"
            aria-label="Close"
            onClick={() => setCreateOpen(false)}
          />
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="create-campaign-title"
            className="nx-confirm-dialog"
          >
            <h2 id="create-campaign-title" className="nx-confirm-dialog__title">
              Create campaign
            </h2>
            <div className="flex flex-col gap-2 mt-2">
              <Input
                placeholder="Campaign name"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
              />
              <Select
                value={newType}
                onChange={(e) => setNewType(e.target.value as CampaignType)}
              >
                {CAMPAIGN_TYPES.map((t) => (
                  <option key={t} value={t}>
                    {typeLabel(t)}
                  </option>
                ))}
              </Select>
            </div>
            <div className="nx-confirm-dialog__actions mt-3">
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setCreateOpen(false)}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                loading={createMut.isPending}
                onClick={() => {
                  if (!newName.trim()) return;
                  void createMut
                    .mutateAsync({
                      name: newName.trim(),
                      type: newType,
                      cityIds: ["city_ist"],
                      budgetMinor: 50_000_00,
                    })
                    .then((c) => {
                      setCreateOpen(false);
                      setNewName("");
                      router.push(`/campaigns/${c.id}`);
                    });
                }}
              >
                Create draft
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
