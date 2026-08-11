"use client";

import { useMemo, useState } from "react";
import dynamic from "next/dynamic";
import {
  Button,
  ChartFrame,
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
import { lineChartOption } from "@/shared/lib/charts";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/platform-permissions";
import { canApprove } from "@/shared/permissions/dual-control";
import {
  useDisasterRecovery,
  useProposeDrFailover,
  useResolveDrFailover,
  useRunSimulation,
  useStartRestore,
} from "../hooks";
import type {
  BackupStatus,
  DrBackup,
  DrFailoverProposal,
  DrRegion,
  DrSimulation,
  GeoReplicationLink,
  RecoveryTest,
  ReplicationLag,
  RestoreJob,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function backupTone(
  s: BackupStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  if (s === "ok") return "success";
  if (s === "running") return "info";
  if (s === "stale") return "warning";
  return "danger";
}

function lagTone(
  s: ReplicationLag,
): "success" | "warning" | "danger" | "neutral" {
  if (s === "synced") return "success";
  if (s === "lagging") return "warning";
  return "danger";
}

export function DisasterRecoveryView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useDisasterRecovery();
  const proposeMutation = useProposeDrFailover();
  const resolveMutation = useResolveDrFailover();
  const restoreMutation = useStartRestore();
  const simMutation = useRunSimulation();

  const [backupQ, setBackupQ] = useState("");
  const [failoverOpen, setFailoverOpen] = useState(false);
  const [fromRegion, setFromRegion] = useState("eu-west-1");
  const [toRegion, setToRegion] = useState("eu-south-1");
  const [reason, setReason] = useState("");
  const [restoreTarget, setRestoreTarget] = useState<DrBackup | null>(null);
  const [restoreEnv, setRestoreEnv] = useState("dr-sandbox");
  const [simOpen, setSimOpen] = useState(false);
  const [simName, setSimName] = useState("");

  const pending = useMemo(
    () => (data?.proposals ?? []).filter((p) => p.status === "pending"),
    [data?.proposals],
  );

  const filteredBackups = useMemo(() => {
    const items = data?.backups ?? [];
    if (!backupQ.trim()) return items;
    const n = backupQ.trim().toLowerCase();
    return items.filter(
      (b) =>
        b.label.toLowerCase().includes(n) ||
        b.scope.toLowerCase().includes(n) ||
        b.region.toLowerCase().includes(n),
    );
  }, [data?.backups, backupQ]);

  const backupCols = useMemo<DataGridColumnDef<DrBackup>[]>(
    () => [
      { id: "label", header: "Backup", accessorKey: "label" },
      { id: "scope", header: "Scope", accessorKey: "scope", width: 140 },
      { id: "region", header: "Region", accessorKey: "region", width: 120 },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={backupTone(row.status)} />
        ),
        width: 100,
      },
      {
        id: "size",
        header: "Size",
        cell: ({ row }) => `${row.sizeGb.toLocaleString("en-US")} GB`,
        width: 100,
      },
      {
        id: "taken",
        header: "Taken",
        cell: ({ row }) => new Date(row.takenAt).toLocaleString("en-US"),
        width: 160,
      },
      {
        id: "actions",
        header: "Restore",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "dr:execute")}>
            <Button
              size="sm"
              variant="secondary"
              disabled={row.status === "failed" || row.status === "running"}
              onClick={(e) => {
                e.stopPropagation();
                setRestoreTarget(row);
              }}
            >
              Start restore
            </Button>
          </PermissionGate>
        ),
        width: 130,
      },
    ],
    [session],
  );

  const repCols = useMemo<DataGridColumnDef<GeoReplicationLink>[]>(
    () => [
      { id: "src", header: "Source", accessorKey: "sourceRegion" },
      { id: "tgt", header: "Target", accessorKey: "targetRegion" },
      { id: "mode", header: "Mode", accessorKey: "mode", width: 100 },
      {
        id: "lag",
        header: "Lag (s)",
        cell: ({ row }) => String(row.lagSeconds),
        width: 80,
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={lagTone(row.status)} />
        ),
        width: 100,
      },
      {
        id: "tp",
        header: "MB/s",
        cell: ({ row }) => String(row.throughputMBps),
        width: 80,
      },
    ],
    [],
  );

  const regionCols = useMemo<DataGridColumnDef<DrRegion>[]>(
    () => [
      { id: "name", header: "Region", accessorKey: "name" },
      { id: "code", header: "Code", accessorKey: "code", width: 120 },
      {
        id: "role",
        header: "Role",
        cell: ({ row }) => (
          <StatusBadge
            status={row.role}
            tone={
              row.role === "primary"
                ? "success"
                : row.role === "warm"
                  ? "warning"
                  : "info"
            }
          />
        ),
        width: 100,
      },
      {
        id: "rpo",
        header: "RPO/RTO",
        cell: ({ row }) => `${row.rpoMinutes}m / ${row.rtoMinutes}m`,
        width: 110,
      },
      {
        id: "health",
        header: "Health",
        cell: ({ row }) => (
          <StatusBadge
            status={row.healthy ? "healthy" : "unhealthy"}
            tone={row.healthy ? "success" : "danger"}
          />
        ),
        width: 110,
      },
    ],
    [],
  );

  const restoreCols = useMemo<DataGridColumnDef<RestoreJob>[]>(
    () => [
      { id: "label", header: "Backup", accessorKey: "backupLabel" },
      { id: "env", header: "Target", accessorKey: "targetEnv", width: 120 },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "completed"
                ? "success"
                : row.status === "failed"
                  ? "danger"
                  : "info"
            }
          />
        ),
        width: 110,
      },
      {
        id: "pct",
        header: "%",
        cell: ({ row }) => `${row.progressPct}%`,
        width: 60,
      },
      { id: "by", header: "By", accessorKey: "requestedBy" },
    ],
    [],
  );

  const testCols = useMemo<DataGridColumnDef<RecoveryTest>[]>(
    () => [
      { id: "name", header: "Test", accessorKey: "name" },
      { id: "scenario", header: "Scenario", accessorKey: "scenario" },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "passed"
                ? "success"
                : row.status === "failed"
                  ? "danger"
                  : row.status === "running"
                    ? "info"
                    : "warning"
            }
          />
        ),
        width: 110,
      },
      { id: "owner", header: "Owner", accessorKey: "owner", width: 120 },
      {
        id: "next",
        header: "Next run",
        cell: ({ row }) => new Date(row.nextRunAt).toLocaleDateString("en-US"),
        width: 110,
      },
    ],
    [],
  );

  const simCols = useMemo<DataGridColumnDef<DrSimulation>[]>(
    () => [
      { id: "name", header: "Simulation", accessorKey: "name" },
      {
        id: "path",
        header: "Path",
        cell: ({ row }) => `${row.regionFrom} → ${row.regionTo}`,
      },
      { id: "blast", header: "Blast radius", accessorKey: "blastRadius" },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "completed"
                ? "success"
                : row.status === "running"
                  ? "info"
                  : row.status === "aborted"
                    ? "danger"
                    : "neutral"
            }
          />
        ),
        width: 110,
      },
      { id: "notes", header: "Notes", accessorKey: "notes" },
    ],
    [],
  );

  const proposalCols = useMemo<DataGridColumnDef<DrFailoverProposal>[]>(
    () => [
      {
        id: "path",
        header: "Failover",
        cell: ({ row }) => `${row.fromRegion} → ${row.toRegion}`,
      },
      { id: "by", header: "Requester", accessorKey: "requesterEmail" },
      { id: "reason", header: "Reason", accessorKey: "reason" },
      {
        id: "when",
        header: "Proposed",
        cell: ({ row }) => new Date(row.createdAt).toLocaleString("en-US"),
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
        <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} height={88} />
          ))}
        </div>
        <Skeleton height={240} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load disaster recovery
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

  const k = data.kpis;

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Disaster recovery"
        description={`Backups · geo replication · dual-control failover${isFetching ? " · refreshing…" : ""}`}
        actions={
          <div className="flex gap-[var(--nx-space-2)]">
            <Button size="sm" variant="ghost" onClick={() => void refetch()}>
              Refresh
            </Button>
            <PermissionGate allowed={can(session, "dr:execute")}>
              <Button size="sm" variant="secondary" onClick={() => setSimOpen(true)}>
                Run simulation
              </Button>
              <Button size="sm" onClick={() => setFailoverOpen(true)}>
                Propose failover
              </Button>
            </PermissionGate>
          </div>
        }
      />

      <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard title="Backups OK" value={String(k.backupsOk)} tone="success" />
        <KpiCard
          title="Backups failed"
          value={String(k.backupsFailed)}
          tone={k.backupsFailed ? "danger" : "neutral"}
        />
        <KpiCard title="Replication links" value={String(k.replicationLinks)} />
        <KpiCard
          title="Lagging"
          value={String(k.laggingLinks)}
          tone={k.laggingLinks ? "warning" : "success"}
        />
        <KpiCard
          title="Pending failovers"
          value={String(k.pendingFailovers)}
          tone={k.pendingFailovers ? "warning" : "neutral"}
        />
        <KpiCard
          title="Last test (days)"
          value={String(k.lastSuccessfulTestDaysAgo)}
          tone={k.lastSuccessfulTestDaysAgo > 30 ? "warning" : "success"}
        />
      </div>

      <ChartFrame title="Observed RPO (minutes)">
        <ReactECharts
          style={{ height: 220 }}
          option={lineChartOption(data.rpoSeries, "#0B6E6E", "min")}
        />
      </ChartFrame>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Dual-control failover proposals ({pending.length})
        </h3>
        <DataGrid
          columns={proposalCols}
          data={pending}
          getRowId={(r) => r.id}
          emptyMessage="No pending DR failover proposals"
        />
      </section>

      <FilterBar>
        <Input
          placeholder="Filter backups…"
          value={backupQ}
          onChange={(e) => setBackupQ(e.target.value)}
          aria-label="Filter backups"
        />
      </FilterBar>
      <DataGrid
        columns={backupCols}
        data={filteredBackups}
        getRowId={(r) => r.id}
        emptyMessage="No backups"
      />

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Geo replication
      </h3>
      <DataGrid columns={repCols} data={data.replication} getRowId={(r) => r.id} />

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Multi-region topology
      </h3>
      <DataGrid columns={regionCols} data={data.regions} getRowId={(r) => r.id} />

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Restore workflows
      </h3>
      <DataGrid columns={restoreCols} data={data.restores} getRowId={(r) => r.id} />

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Recovery testing
      </h3>
      <DataGrid columns={testCols} data={data.tests} getRowId={(r) => r.id} />

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Failover simulations
      </h3>
      <DataGrid
        columns={simCols}
        data={data.simulations}
        getRowId={(r) => r.id}
      />

      <ConfirmDialog
        open={failoverOpen}
        title="Propose multi-region failover"
        danger
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)] mt-[var(--nx-space-2)]">
            <p className="m-0 text-[13px]">
              Requires a second approver with dual_control:approve. Traffic
              cutover executes only after approval.
            </p>
            <Select
              value={fromRegion}
              onChange={(e) => setFromRegion(e.target.value)}
              aria-label="From region"
            >
              {data.regions.map((r) => (
                <option key={r.id} value={r.code}>
                  {r.name}
                </option>
              ))}
            </Select>
            <Select
              value={toRegion}
              onChange={(e) => setToRegion(e.target.value)}
              aria-label="To region"
            >
              {data.regions.map((r) => (
                <option key={r.id} value={r.code}>
                  {r.name}
                </option>
              ))}
            </Select>
            <Input
              placeholder="Reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </div>
        }
        confirmLabel="Submit proposal"
        loading={proposeMutation.isPending}
        onCancel={() => {
          setFailoverOpen(false);
          setReason("");
        }}
        onConfirm={() => {
          if (!session || !reason.trim() || fromRegion === toRegion) return;
          void proposeMutation
            .mutateAsync({
              fromRegion,
              toRegion,
              reason: reason.trim(),
              requesterId: session.userId,
              requesterEmail: session.email,
            })
            .then(() => {
              setFailoverOpen(false);
              setReason("");
            });
        }}
      />

      <ConfirmDialog
        open={restoreTarget != null}
        title="Start restore workflow"
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)]">
            <p className="m-0 text-[13px]">
              Restore <strong>{restoreTarget?.label}</strong> into target
              environment (never production without dual-control change window).
            </p>
            <Select
              value={restoreEnv}
              onChange={(e) => setRestoreEnv(e.target.value)}
            >
              <option value="dr-sandbox">dr-sandbox</option>
              <option value="staging">staging</option>
              <option value="forensics">forensics</option>
            </Select>
          </div>
        }
        confirmLabel="Queue restore"
        loading={restoreMutation.isPending}
        onCancel={() => setRestoreTarget(null)}
        onConfirm={() => {
          if (!restoreTarget || !session) return;
          void restoreMutation
            .mutateAsync({
              backupId: restoreTarget.id,
              backupLabel: restoreTarget.label,
              targetEnv: restoreEnv,
              requestedBy: session.email,
            })
            .then(() => setRestoreTarget(null));
        }}
      />

      <ConfirmDialog
        open={simOpen}
        title="Run failover simulation"
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)]">
            <Input
              placeholder="Simulation name"
              value={simName}
              onChange={(e) => setSimName(e.target.value)}
            />
            <p className="m-0 text-[12px] text-[var(--nx-text-tertiary)]">
              Dry-run path {fromRegion} → {toRegion}. No live traffic shift.
            </p>
          </div>
        }
        confirmLabel="Start simulation"
        loading={simMutation.isPending}
        onCancel={() => {
          setSimOpen(false);
          setSimName("");
        }}
        onConfirm={() => {
          if (!simName.trim()) return;
          void simMutation
            .mutateAsync({
              name: simName.trim(),
              regionFrom: fromRegion,
              regionTo: toRegion,
              blastRadius: "platform (simulated)",
            })
            .then(() => {
              setSimOpen(false);
              setSimName("");
            });
        }}
      />
    </div>
  );
}
