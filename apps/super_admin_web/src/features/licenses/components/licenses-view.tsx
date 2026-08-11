"use client";

import { useMemo, useState } from "react";
import {
  Button,
  ConfirmDialog,
  DataGrid,
  FilterBar,
  Input,
  KpiCard,
  PageHeader,
  PermissionGate,
  Select,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/platform-permissions";
import { canApprove } from "@/shared/permissions/dual-control";
import { formatMinorUnits } from "@/shared/lib/money";
import {
  useLicenses,
  useProposeLicenseOverride,
  useRenewLicense,
  useResolveLicenseProposal,
} from "../hooks";
import type {
  LicenseOverrideProposal,
  LicensePlan,
  LicenseStatus,
  TenantLicense,
  UsageLimitRow,
} from "../types";

function statusTone(
  status: LicenseStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "active":
      return "success";
    case "trial":
      return "info";
    case "past_due":
      return "warning";
    case "cancelled":
    case "expired":
      return "danger";
    default:
      return "neutral";
  }
}

export function LicensesView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useLicenses();
  const renew = useRenewLicense();
  const propose = useProposeLicenseOverride();
  const resolve = useResolveLicenseProposal();

  const [q, setQ] = useState("");
  const [status, setStatus] = useState("all");
  const [overrideTarget, setOverrideTarget] = useState<TenantLicense | null>(
    null,
  );
  const [overrideSeats, setOverrideSeats] = useState("");
  const [overrideReason, setOverrideReason] = useState("");

  const renewalsSoon = useMemo(() => {
    const cutoff = Date.now() + 14 * 86400_000;
    return (data?.licenses ?? []).filter(
      (l) => new Date(l.renewsAt).getTime() <= cutoff,
    ).length;
  }, [data?.licenses]);

  const filtered = useMemo(() => {
    const items = data?.licenses ?? [];
    return items.filter((l) => {
      if (status !== "all" && l.status !== status) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        l.tenantName.toLowerCase().includes(needle) ||
        l.planName.toLowerCase().includes(needle) ||
        (l.enterpriseContractId?.toLowerCase().includes(needle) ?? false)
      );
    });
  }, [data?.licenses, q, status]);

  const pending = useMemo(
    () => (data?.proposals ?? []).filter((p) => p.status === "pending"),
    [data?.proposals],
  );

  const planCols = useMemo<DataGridColumnDef<LicensePlan>[]>(
    () => [
      { id: "name", header: "Plan", accessorKey: "name" },
      {
        id: "tier",
        header: "Tier",
        cell: ({ row }) => (
          <StatusBadge
            status={row.tier}
            tone={row.tier === "enterprise" || row.tier === "custom" ? "success" : "info"}
          />
        ),
        width: 110,
      },
      {
        id: "price",
        header: "Monthly",
        cell: ({ row }) =>
          row.monthlyPriceMinor === 0
            ? "Custom / contract"
            : formatMinorUnits(row.monthlyPriceMinor, row.currency),
        width: 140,
      },
      {
        id: "limits",
        header: "Limits",
        cell: ({ row }) =>
          `${row.limits.warehouses} WH · ${row.limits.couriers} couriers · ${row.limits.seats} seats`,
      },
      {
        id: "features",
        header: "Features",
        cell: ({ row }) => row.features.join(", "),
      },
    ],
    [],
  );

  const licenseCols = useMemo<DataGridColumnDef<TenantLicense>[]>(
    () => [
      { id: "tenant", header: "Tenant", accessorKey: "tenantName" },
      { id: "plan", header: "Plan", accessorKey: "planName", width: 120 },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={statusTone(row.status)} />
        ),
        width: 100,
      },
      {
        id: "seats",
        header: "Seats",
        cell: ({ row }) => (
          <span className="tabular-nums">
            {row.seatsUsed}/{row.seats}
          </span>
        ),
        width: 90,
      },
      {
        id: "renew",
        header: "Renews",
        cell: ({ row }) =>
          new Date(row.renewsAt).toLocaleDateString("en-US"),
        width: 110,
      },
      {
        id: "overage",
        header: "Overage",
        cell: ({ row }) =>
          row.overageEnabled
            ? formatMinorUnits(row.overageSpendMinor, row.currency)
            : "Off",
        width: 110,
      },
      {
        id: "ent",
        header: "Enterprise",
        cell: ({ row }) => row.enterpriseContractId ?? "—",
        width: 130,
      },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => (
          <div
            className="flex gap-[var(--nx-space-1)]"
            onClick={(e) => e.stopPropagation()}
          >
            <PermissionGate allowed={can(session, "licenses:write")}>
              <Button
                size="sm"
                variant="secondary"
                loading={renew.isPending}
                onClick={() => void renew.mutateAsync(row.id)}
              >
                Renew
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => {
                  setOverrideTarget(row);
                  setOverrideSeats(String(row.seats));
                }}
              >
                Override
              </Button>
            </PermissionGate>
          </div>
        ),
        width: 200,
      },
    ],
    [session, renew],
  );

  const usageCols = useMemo<DataGridColumnDef<UsageLimitRow>[]>(
    () => [
      { id: "tenant", header: "Tenant", accessorKey: "tenantName" },
      { id: "metric", header: "Metric", accessorKey: "metric" },
      {
        id: "used",
        header: "Used / limit",
        cell: ({ row }) => (
          <span className="tabular-nums">
            {row.used.toLocaleString("en-US")} / {row.limit.toLocaleString("en-US")}{" "}
            {row.unit}
          </span>
        ),
      },
      {
        id: "over",
        header: "Overage",
        cell: ({ row }) => (
          <StatusBadge
            status={row.overagePct > 0 ? `+${row.overagePct}%` : "ok"}
            tone={row.overagePct > 0 ? "warning" : "success"}
          />
        ),
        width: 100,
      },
    ],
    [],
  );

  const proposalCols = useMemo<DataGridColumnDef<LicenseOverrideProposal>[]>(
    () => [
      { id: "tenant", header: "Tenant", accessorKey: "tenantName" },
      { id: "reason", header: "Reason", accessorKey: "reason" },
      { id: "by", header: "Requester", accessorKey: "requesterEmail" },
      {
        id: "payload",
        header: "Override",
        cell: ({ row }) =>
          [
            row.payload.seats != null ? `seats=${row.payload.seats}` : null,
            row.payload.overageEnabled != null
              ? `overage=${row.payload.overageEnabled}`
              : null,
            row.payload.featureOverrides?.join(",") ?? null,
          ]
            .filter(Boolean)
            .join(" · ") || "—",
      },
      {
        id: "approve",
        header: "Approve",
        cell: ({ row }) => {
          const allowed = canApprove(session, row);
          return (
            <PermissionGate
              allowed={allowed}
              fallback={
                <span className="text-[11px] text-[var(--nx-text-tertiary)]">
                  {session?.userId === row.requesterId
                    ? "Cannot self-approve"
                    : "No approve right"}
                </span>
              }
            >
              <div className="flex gap-[var(--nx-space-1)]">
                <Button
                  size="sm"
                  variant="primary"
                  loading={resolve.isPending}
                  onClick={() =>
                    void resolve.mutateAsync({
                      proposalId: row.id,
                      decision: "approved",
                      approverId: session!.userId,
                    })
                  }
                >
                  Approve
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() =>
                    void resolve.mutateAsync({
                      proposalId: row.id,
                      decision: "rejected",
                      approverId: session!.userId,
                    })
                  }
                >
                  Reject
                </Button>
              </div>
            </PermissionGate>
          );
        },
      },
    ],
    [session, resolve],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={96} />
        <Skeleton height={280} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load licenses
        </p>
        <p className="m-0 mt-[var(--nx-space-1)] text-[var(--nx-text-secondary)]">
          {error instanceof Error ? error.message : "Unknown error"}
        </p>
        <button
          type="button"
          onClick={() => void refetch()}
          className="mt-[var(--nx-space-3)] text-[var(--nx-text-link)] underline cursor-pointer bg-transparent border-0"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Licenses"
        description={`Subscription plans · feature access · usage limits · renewals · overage · enterprise${isFetching ? " · refreshing…" : ""}`}
      />

      <section
        aria-label="License KPIs"
        className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]"
      >
        <KpiCard
          title="Active licenses"
          value={String(data.licenses.filter((l) => l.status === "active").length)}
          tone="success"
        />
        <KpiCard
          title="Past due"
          value={String(data.licenses.filter((l) => l.status === "past_due").length)}
          tone="warning"
        />
        <KpiCard
          title="Renewals ≤14d"
          value={String(renewalsSoon)}
          tone="brand"
        />
        <KpiCard
          title="Pending overrides"
          value={String(pending.length)}
          tone="neutral"
        />
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Subscription plans
        </h3>
        <DataGrid columns={planCols} data={data.plans} getRowId={(r) => r.id} />
      </section>

      <FilterBar
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      >
        <Input
          placeholder="Search tenant, plan, contract…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
        <Select value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="trial">Trial</option>
          <option value="past_due">Past due</option>
          <option value="cancelled">Cancelled</option>
          <option value="expired">Expired</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={licenseCols}
        data={filtered}
        getRowId={(r) => r.id}
        emptyMessage="No licenses"
      />

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Usage limits & overage
        </h3>
        <DataGrid columns={usageCols} data={data.usage} getRowId={(r) => r.id} />
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          License override dual-control ({pending.length})
        </h3>
        <DataGrid
          columns={proposalCols}
          data={pending}
          getRowId={(r) => r.id}
          emptyMessage="No pending license overrides"
        />
      </section>

      <ConfirmDialog
        open={overrideTarget != null}
        title="Propose license override"
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)]">
            <p className="m-0 text-[13px]">
              Dual-control required. Tenant:{" "}
              <strong>{overrideTarget?.tenantName}</strong>
            </p>
            <Input
              placeholder="Seat override"
              value={overrideSeats}
              onChange={(e) => setOverrideSeats(e.target.value)}
            />
            <Input
              placeholder="Reason"
              value={overrideReason}
              onChange={(e) => setOverrideReason(e.target.value)}
            />
          </div>
        }
        confirmLabel="Submit proposal"
        loading={propose.isPending}
        onCancel={() => {
          setOverrideTarget(null);
          setOverrideReason("");
        }}
        onConfirm={() => {
          if (!overrideTarget || !session || !overrideReason.trim()) return;
          void propose
            .mutateAsync({
              licenseId: overrideTarget.id,
              reason: overrideReason.trim(),
              requesterId: session.userId,
              requesterEmail: session.email,
              payload: {
                seats: Number(overrideSeats) || overrideTarget.seats,
                overageEnabled: true,
              },
            })
            .then(() => {
              setOverrideTarget(null);
              setOverrideReason("");
            });
        }}
      />
    </div>
  );
}
