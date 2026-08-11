"use client";

import { useMemo, useState } from "react";
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
  useEmergencyRollback,
  useFlags,
  useProposeKillSwitch,
  useResolveKillSwitchProposal,
  useToggleFeatureFlag,
  useUpdateFlagRollout,
  useUpsertFlag,
} from "../hooks";
import type {
  FeatureFlag,
  FlagKind,
  FlagScope,
  KillSwitchProposal,
} from "../types";

function kindTone(
  kind: FlagKind,
): "danger" | "warning" | "info" | "success" | "neutral" {
  switch (kind) {
    case "kill_switch":
      return "danger";
    case "ab_test":
      return "info";
    case "scheduled":
      return "warning";
    default:
      return "neutral";
  }
}

export function FlagsView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } = useFlags();
  const upsert = useUpsertFlag();
  const toggle = useToggleFeatureFlag();
  const rollout = useUpdateFlagRollout();
  const rollback = useEmergencyRollback();
  const propose = useProposeKillSwitch();
  const resolve = useResolveKillSwitchProposal();

  const [q, setQ] = useState("");
  const [scope, setScope] = useState("all");
  const [kind, setKind] = useState("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [newKey, setNewKey] = useState("");
  const [newName, setNewName] = useState("");
  const [newScope, setNewScope] = useState<FlagScope>("global");
  const [newKind, setNewKind] = useState<FlagKind>("feature");
  const [newRollout, setNewRollout] = useState("10");
  const [killTarget, setKillTarget] = useState<FeatureFlag | null>(null);
  const [killEnabled, setKillEnabled] = useState(true);
  const [killReason, setKillReason] = useState("");
  const [rolloutTarget, setRolloutTarget] = useState<FeatureFlag | null>(null);
  const [rolloutPct, setRolloutPct] = useState("50");

  const filtered = useMemo(() => {
    const items = data?.flags ?? [];
    return items.filter((f) => {
      if (scope !== "all" && f.scope !== scope) return false;
      if (kind !== "all" && f.kind !== kind) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        f.key.toLowerCase().includes(needle) ||
        f.name.toLowerCase().includes(needle) ||
        (f.scopeTarget?.toLowerCase().includes(needle) ?? false)
      );
    });
  }, [data?.flags, q, scope, kind]);

  const pending = useMemo(
    () => (data?.proposals ?? []).filter((p) => p.status === "pending"),
    [data?.proposals],
  );

  const columns = useMemo<DataGridColumnDef<FeatureFlag>[]>(
    () => [
      { id: "key", header: "Key", accessorKey: "key", width: 180 },
      { id: "name", header: "Name", accessorKey: "name" },
      {
        id: "scope",
        header: "Scope",
        cell: ({ row }) => (
          <span className="text-[12px]">
            {row.scope}
            {row.scopeTarget ? ` · ${row.scopeTarget}` : ""}
          </span>
        ),
        width: 140,
      },
      {
        id: "kind",
        header: "Kind",
        cell: ({ row }) => (
          <StatusBadge status={row.kind} tone={kindTone(row.kind)} />
        ),
        width: 120,
      },
      {
        id: "rollout",
        header: "Rollout",
        cell: ({ row }) => (
          <span className="tabular-nums text-[12px]">{row.rolloutPct}%</span>
        ),
        width: 80,
      },
      {
        id: "ab",
        header: "A/B",
        cell: ({ row }) =>
          row.variants.length > 0
            ? row.variants.map((v) => `${v.key}:${v.weightPct}%`).join(" / ")
            : "—",
        width: 140,
      },
      {
        id: "status",
        header: "State",
        cell: ({ row }) => (
          <StatusBadge
            status={row.enabled ? "on" : row.status}
            tone={
              row.enabled
                ? "success"
                : row.status === "rolling_back"
                  ? "warning"
                  : "neutral"
            }
          />
        ),
        width: 110,
      },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => (
          <div
            className="flex flex-wrap gap-[var(--nx-space-1)]"
            onClick={(e) => e.stopPropagation()}
          >
            {row.kind === "kill_switch" ? (
              <PermissionGate allowed={can(session, "flags:kill")}>
                <Button
                  size="sm"
                  variant={row.enabled ? "secondary" : "danger"}
                  onClick={() => {
                    setKillTarget(row);
                    setKillEnabled(!row.enabled);
                  }}
                >
                  Propose {row.enabled ? "disable" : "arm"}
                </Button>
              </PermissionGate>
            ) : (
              <PermissionGate allowed={can(session, "flags:write")}>
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() =>
                    void toggle.mutateAsync({
                      flagId: row.id,
                      enabled: !row.enabled,
                    })
                  }
                >
                  Toggle
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => {
                    setRolloutTarget(row);
                    setRolloutPct(String(row.rolloutPct));
                  }}
                >
                  Rollout
                </Button>
                <Button
                  size="sm"
                  variant="danger"
                  onClick={() => void rollback.mutateAsync(row.id)}
                >
                  Rollback
                </Button>
              </PermissionGate>
            )}
          </div>
        ),
        width: 280,
      },
    ],
    [session, toggle, rollback],
  );

  const proposalCols = useMemo<DataGridColumnDef<KillSwitchProposal>[]>(
    () => [
      { id: "flag", header: "Kill switch", accessorKey: "flagName" },
      {
        id: "target",
        header: "Target",
        cell: ({ row }) => (
          <StatusBadge
            status={row.targetEnabled ? "arm" : "disable"}
            tone={row.targetEnabled ? "danger" : "warning"}
          />
        ),
      },
      { id: "by", header: "Requester", accessorKey: "requesterEmail" },
      { id: "reason", header: "Reason", accessorKey: "reason" },
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
        <Skeleton height={40} />
        <Skeleton height={280} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load feature flags
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
        title="Feature flags"
        description={`Global / country / company / user · % rollout · A/B · scheduled · kill switches (dual-control)${isFetching ? " · refreshing…" : ""}`}
        actions={
          <PermissionGate allowed={can(session, "flags:write")}>
            <Button size="sm" onClick={() => setCreateOpen(true)}>
              Create flag
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
          placeholder="Search key, name, target…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search flags"
        />
        <Select
          value={scope}
          onChange={(e) => setScope(e.target.value)}
          aria-label="Filter scope"
        >
          <option value="all">All scopes</option>
          <option value="global">Global</option>
          <option value="country">Country</option>
          <option value="company">Company</option>
          <option value="user">User</option>
        </Select>
        <Select
          value={kind}
          onChange={(e) => setKind(e.target.value)}
          aria-label="Filter kind"
        >
          <option value="all">All kinds</option>
          <option value="feature">Feature</option>
          <option value="ab_test">A/B</option>
          <option value="scheduled">Scheduled</option>
          <option value="kill_switch">Kill switch</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={filtered}
        getRowId={(r) => r.id}
        emptyMessage="No flags match filters"
      />

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Kill-switch dual-control ({pending.length})
        </h3>
        <DataGrid
          columns={proposalCols}
          data={pending}
          getRowId={(r) => r.id}
          emptyMessage="No pending kill-switch proposals"
        />
      </section>

      <ConfirmDialog
        open={createOpen}
        title="Create feature flag"
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)] mt-[var(--nx-space-2)]">
            <Input
              placeholder="Key (e.g. checkout.v3)"
              value={newKey}
              onChange={(e) => setNewKey(e.target.value)}
            />
            <Input
              placeholder="Display name"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
            />
            <Select
              value={newScope}
              onChange={(e) => setNewScope(e.target.value as FlagScope)}
            >
              <option value="global">Global</option>
              <option value="country">Country</option>
              <option value="company">Company</option>
              <option value="user">User</option>
            </Select>
            <Select
              value={newKind}
              onChange={(e) => setNewKind(e.target.value as FlagKind)}
            >
              <option value="feature">Feature</option>
              <option value="ab_test">A/B test</option>
              <option value="scheduled">Scheduled</option>
            </Select>
            <Input
              placeholder="Initial rollout %"
              value={newRollout}
              onChange={(e) => setNewRollout(e.target.value)}
            />
          </div>
        }
        confirmLabel="Create"
        loading={upsert.isPending}
        onCancel={() => setCreateOpen(false)}
        onConfirm={() => {
          if (!newKey.trim() || !newName.trim()) return;
          void upsert
            .mutateAsync({
              key: newKey.trim(),
              name: newName.trim(),
              description: newName.trim(),
              scope: newScope,
              kind: newKind,
              rolloutPct: Number(newRollout) || 0,
              variants:
                newKind === "ab_test"
                  ? [
                      { key: "control", weightPct: 50 },
                      { key: "treatment", weightPct: 50 },
                    ]
                  : [],
            })
            .then(() => {
              setCreateOpen(false);
              setNewKey("");
              setNewName("");
              setNewScope("global");
              setNewKind("feature");
              setNewRollout("10");
            });
        }}
      />

      <ConfirmDialog
        open={killTarget != null}
        title={killEnabled ? "Propose arm kill switch" : "Propose disable kill switch"}
        danger={killEnabled}
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)]">
            <p className="m-0 text-[13px]">
              Dual-control required. Flag: <strong>{killTarget?.name}</strong>
            </p>
            <Input
              placeholder="Reason"
              value={killReason}
              onChange={(e) => setKillReason(e.target.value)}
            />
          </div>
        }
        confirmLabel="Submit proposal"
        loading={propose.isPending}
        onCancel={() => {
          setKillTarget(null);
          setKillReason("");
        }}
        onConfirm={() => {
          if (!killTarget || !session || !killReason.trim()) return;
          void propose
            .mutateAsync({
              flagId: killTarget.id,
              targetEnabled: killEnabled,
              reason: killReason.trim(),
              requesterId: session.userId,
              requesterEmail: session.email,
            })
            .then(() => {
              setKillTarget(null);
              setKillReason("");
            });
        }}
      />

      <ConfirmDialog
        open={rolloutTarget != null}
        title="Update rollout %"
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)]">
            <p className="m-0 text-[13px]">{rolloutTarget?.key}</p>
            <Input
              placeholder="Rollout percent 0–100"
              value={rolloutPct}
              onChange={(e) => setRolloutPct(e.target.value)}
            />
          </div>
        }
        confirmLabel="Save"
        loading={rollout.isPending}
        onCancel={() => setRolloutTarget(null)}
        onConfirm={() => {
          if (!rolloutTarget) return;
          const pct = Math.min(100, Math.max(0, Number(rolloutPct) || 0));
          void rollout
            .mutateAsync({ flagId: rolloutTarget.id, rolloutPct: pct })
            .then(() => setRolloutTarget(null));
        }}
      />
    </div>
  );
}
