"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Button,
  ConfirmDialog,
  DataGrid,
  FilterBar,
  Input,
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
import {
  useCreateTenant,
  useProposeTenantAction,
  useResolveTenantProposal,
  useTenants,
} from "../hooks";
import type {
  TenantDualControlProposal,
  TenantIsolationMode,
  TenantListItem,
  TenantStatus,
} from "../types";

function statusTone(
  status: TenantStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "active":
      return "success";
    case "pending":
    case "migrating":
      return "warning";
    case "suspended":
      return "danger";
    default:
      return "neutral";
  }
}

function isolationTone(
  mode: TenantIsolationMode,
): "info" | "success" | "warning" {
  if (mode === "shared") return "info";
  if (mode === "hybrid") return "warning";
  return "success";
}

export function TenantsView() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } = useTenants();
  const createMutation = useCreateTenant();
  const proposeMutation = useProposeTenantAction();
  const resolveMutation = useResolveTenantProposal();

  const [q, setQ] = useState("");
  const [status, setStatus] = useState("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [isolationMode, setIsolationMode] =
    useState<TenantIsolationMode>("shared");
  const [proposeTarget, setProposeTarget] = useState<{
    tenant: TenantListItem;
    action: "tenant_suspend" | "tenant_delete";
  } | null>(null);
  const [proposeReason, setProposeReason] = useState("");

  const filtered = useMemo(() => {
    const items = data?.items ?? [];
    return items.filter((t) => {
      if (status !== "all" && t.status !== status) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        t.name.toLowerCase().includes(needle) ||
        t.slug.toLowerCase().includes(needle) ||
        t.companyName.toLowerCase().includes(needle)
      );
    });
  }, [data?.items, q, status]);

  const pending = useMemo(
    () => (data?.proposals ?? []).filter((p) => p.status === "pending"),
    [data?.proposals],
  );

  const columns = useMemo<DataGridColumnDef<TenantListItem>[]>(
    () => [
      { id: "name", header: "Tenant", accessorKey: "name" },
      { id: "slug", header: "Slug", accessorKey: "slug", width: 120 },
      { id: "company", header: "Company", accessorKey: "companyName" },
      {
        id: "iso",
        header: "Isolation",
        cell: ({ row }) => (
          <StatusBadge
            status={row.isolationMode}
            tone={isolationTone(row.isolationMode)}
          />
        ),
        width: 110,
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={statusTone(row.status)} />
        ),
        width: 110,
      },
      { id: "region", header: "Region", accessorKey: "region", width: 130 },
      {
        id: "actions",
        header: "Dual-control",
        cell: ({ row }) => (
          <div
            className="flex gap-[var(--nx-space-1)]"
            onClick={(e) => e.stopPropagation()}
          >
            <PermissionGate allowed={can(session, "tenants:suspend")}>
              <Button
                size="sm"
                variant="secondary"
                disabled={row.status === "suspended"}
                onClick={() =>
                  setProposeTarget({ tenant: row, action: "tenant_suspend" })
                }
              >
                Propose suspend
              </Button>
            </PermissionGate>
            <PermissionGate allowed={can(session, "tenants:delete")}>
              <Button
                size="sm"
                variant="danger"
                onClick={() =>
                  setProposeTarget({ tenant: row, action: "tenant_delete" })
                }
              >
                Propose delete
              </Button>
            </PermissionGate>
          </div>
        ),
        width: 260,
      },
    ],
    [session],
  );

  const proposalCols = useMemo<DataGridColumnDef<TenantDualControlProposal>[]>(
    () => [
      { id: "tenant", header: "Tenant", accessorKey: "tenantName" },
      {
        id: "action",
        header: "Action",
        cell: ({ row }) => (
          <StatusBadge
            status={row.action.replace("tenant_", "")}
            tone={row.action === "tenant_delete" ? "danger" : "warning"}
          />
        ),
      },
      { id: "by", header: "Requester", accessorKey: "requesterEmail" },
      { id: "reason", header: "Reason", accessorKey: "reason" },
      {
        id: "when",
        header: "Proposed",
        cell: ({ row }) =>
          new Date(row.createdAt).toLocaleString("en-US"),
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
                  loading={resolveMutation.isPending}
                  onClick={() =>
                    void resolveMutation.mutateAsync({
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
                    void resolveMutation.mutateAsync({
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
    [session, resolveMutation],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={40} />
        <Skeleton height={280} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load tenants
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
        title="Tenants"
        description={`${data.total} tenants · isolation · dual-control suspend/delete${isFetching ? " · refreshing…" : ""}`}
        actions={
          <PermissionGate allowed={can(session, "tenants:write")}>
            <Button size="sm" onClick={() => setCreateOpen(true)}>
              Create tenant
            </Button>
          </PermissionGate>
        }
      />

      <FilterBar
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      >
        <Input
          placeholder="Search name, slug, company…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search tenants"
        />
        <Select
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          aria-label="Filter status"
        >
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="pending">Pending</option>
          <option value="migrating">Migrating</option>
          <option value="suspended">Suspended</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={filtered}
        getRowId={(r) => r.id}
        onRowClick={(row) => router.push(`/tenants/${row.id}`)}
      />

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Pending dual-control approvals ({pending.length})
        </h3>
        <DataGrid
          columns={proposalCols}
          data={pending}
          getRowId={(r) => r.id}
          emptyMessage="No pending tenant proposals"
        />
      </section>

      <ConfirmDialog
        open={createOpen}
        title="Create tenant"
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)] mt-[var(--nx-space-2)]">
            <Input
              placeholder="Display name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
            <Input
              placeholder="Slug"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
            />
            <Select
              value={isolationMode}
              onChange={(e) =>
                setIsolationMode(e.target.value as TenantIsolationMode)
              }
            >
              <option value="shared">Shared DB + RLS</option>
              <option value="separate">Separate database</option>
              <option value="hybrid">Hybrid</option>
            </Select>
          </div>
        }
        confirmLabel="Create"
        loading={createMutation.isPending}
        onCancel={() => setCreateOpen(false)}
        onConfirm={() => {
          if (!name.trim() || !slug.trim()) return;
          void createMutation
            .mutateAsync({
              name: name.trim(),
              slug: slug.trim(),
              companyId: "co_new",
              isolationMode,
              region: "eu-west-1",
            })
            .then(() => {
              setCreateOpen(false);
              setName("");
              setSlug("");
              setIsolationMode("shared");
            });
        }}
      />

      <ConfirmDialog
        open={proposeTarget != null}
        title={
          proposeTarget?.action === "tenant_delete"
            ? "Propose tenant delete"
            : "Propose tenant suspend"
        }
        danger={proposeTarget?.action === "tenant_delete"}
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)]">
            <p className="m-0 text-[13px]">
              Requires a second approver. Target:{" "}
              <strong>{proposeTarget?.tenant.name}</strong>
            </p>
            <Input
              placeholder="Reason for proposal"
              value={proposeReason}
              onChange={(e) => setProposeReason(e.target.value)}
            />
          </div>
        }
        confirmLabel="Submit proposal"
        loading={proposeMutation.isPending}
        onCancel={() => {
          setProposeTarget(null);
          setProposeReason("");
        }}
        onConfirm={() => {
          if (!proposeTarget || !session || !proposeReason.trim()) return;
          void proposeMutation
            .mutateAsync({
              tenantId: proposeTarget.tenant.id,
              action: proposeTarget.action,
              reason: proposeReason.trim(),
              requesterId: session.userId,
              requesterEmail: session.email,
            })
            .then(() => {
              setProposeTarget(null);
              setProposeReason("");
            });
        }}
      />
    </div>
  );
}
