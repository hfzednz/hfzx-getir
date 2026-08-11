"use client";

import { useMemo, useState } from "react";
import dynamic from "next/dynamic";
import {
  Button,
  ChartFrame,
  ConfirmDialog,
  DataGrid,
  Input,
  KpiCard,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { barChartOption } from "@/shared/lib/charts";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/platform-permissions";
import { canApprove } from "@/shared/permissions/dual-control";
import {
  useDeployments,
  usePromoteCanary,
  useProposeSecretRotate,
  useResolveSecretRotate,
  useRollbackDeployment,
} from "../hooks";
import type {
  CicdPipeline,
  DeployEnvironment,
  DeploymentRecord,
  PipelineStatus,
  SecretMeta,
  SecretRotateProposal,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function pipeTone(
  s: PipelineStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  if (s === "passed") return "success";
  if (s === "failed") return "danger";
  if (s === "running") return "info";
  if (s === "queued") return "warning";
  return "neutral";
}

export function DeploymentsView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useDeployments();
  const promoteMutation = usePromoteCanary();
  const rollbackMutation = useRollbackDeployment();
  const proposeSecret = useProposeSecretRotate();
  const resolveSecret = useResolveSecretRotate();

  const [rotateTarget, setRotateTarget] = useState<SecretMeta | null>(null);
  const [rotateReason, setRotateReason] = useState("");

  const pendingSecrets = useMemo(
    () => (data?.secretProposals ?? []).filter((p) => p.status === "pending"),
    [data?.secretProposals],
  );

  const pipeCols = useMemo<DataGridColumnDef<CicdPipeline>[]>(
    () => [
      { id: "name", header: "Pipeline", accessorKey: "name" },
      { id: "repo", header: "Repo", accessorKey: "repo" },
      { id: "branch", header: "Branch", accessorKey: "branch", width: 110 },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={pipeTone(row.status)} />
        ),
        width: 100,
      },
      { id: "sha", header: "SHA", accessorKey: "commitSha", width: 90 },
      { id: "by", header: "Triggered by", accessorKey: "triggeredBy" },
      {
        id: "dur",
        header: "Sec",
        cell: ({ row }) => String(row.durationSec),
        width: 60,
      },
    ],
    [],
  );

  const deployCols = useMemo<DataGridColumnDef<DeploymentRecord>[]>(
    () => [
      { id: "svc", header: "Service", accessorKey: "service" },
      { id: "env", header: "Env", accessorKey: "environment", width: 100 },
      {
        id: "strategy",
        header: "Strategy",
        cell: ({ row }) => (
          <StatusBadge
            status={row.strategy.replace("_", "-")}
            tone={
              row.strategy === "canary"
                ? "warning"
                : row.strategy === "blue_green"
                  ? "info"
                  : "neutral"
            }
          />
        ),
        width: 110,
      },
      {
        id: "ver",
        header: "Version",
        cell: ({ row }) => `${row.version} ← ${row.previousVersion}`,
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "healthy"
                ? "success"
                : row.status === "rolled_back" || row.status === "degraded"
                  ? "danger"
                  : "info"
            }
          />
        ),
        width: 120,
      },
      {
        id: "canary",
        header: "Canary %",
        cell: ({ row }) =>
          row.canaryPct != null ? `${row.canaryPct}%` : "—",
        width: 90,
      },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => (
          <div
            className="flex gap-[var(--nx-space-1)]"
            onClick={(e) => e.stopPropagation()}
          >
            <PermissionGate allowed={can(session, "deployments:write")}>
              {row.strategy === "canary" && row.status === "progressing" ? (
                <Button
                  size="sm"
                  variant="secondary"
                  loading={promoteMutation.isPending}
                  onClick={() =>
                    void promoteMutation.mutateAsync({
                      deploymentId: row.id,
                      targetPct: Math.min(100, (row.canaryPct ?? 0) + 25),
                    })
                  }
                >
                  Promote
                </Button>
              ) : null}
            </PermissionGate>
            <PermissionGate allowed={can(session, "deployments:rollback")}>
              <Button
                size="sm"
                variant="danger"
                disabled={row.status === "rolled_back" || row.status === "idle"}
                loading={rollbackMutation.isPending}
                onClick={() =>
                  void rollbackMutation.mutateAsync({ deploymentId: row.id })
                }
              >
                Rollback
              </Button>
            </PermissionGate>
          </div>
        ),
        width: 200,
      },
    ],
    [session, promoteMutation, rollbackMutation],
  );

  const envCols = useMemo<DataGridColumnDef<DeployEnvironment>[]>(
    () => [
      { id: "name", header: "Environment", accessorKey: "name" },
      {
        id: "kind",
        header: "Kind",
        cell: ({ row }) => (
          <StatusBadge
            status={row.kind}
            tone={row.kind === "production" ? "danger" : "info"}
          />
        ),
        width: 110,
      },
      { id: "region", header: "Region", accessorKey: "region", width: 120 },
      { id: "cluster", header: "Cluster", accessorKey: "cluster" },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "healthy"
                ? "success"
                : row.status === "maintenance"
                  ? "warning"
                  : "danger"
            }
          />
        ),
        width: 120,
      },
      {
        id: "secrets",
        header: "Secrets",
        cell: ({ row }) => String(row.secretCount),
        width: 80,
      },
    ],
    [],
  );

  const secretCols = useMemo<DataGridColumnDef<SecretMeta>[]>(
    () => [
      { id: "name", header: "Secret name", accessorKey: "name" },
      { id: "env", header: "Env", accessorKey: "environment", width: 110 },
      { id: "provider", header: "Provider", accessorKey: "provider", width: 90 },
      {
        id: "ver",
        header: "Ver",
        cell: ({ row }) => String(row.version),
        width: 60,
      },
      {
        id: "rotated",
        header: "Rotated",
        cell: ({ row }) => new Date(row.rotatedAt).toLocaleDateString("en-US"),
        width: 110,
      },
      { id: "owners", header: "Owners", accessorKey: "owners", width: 130 },
      {
        id: "value",
        header: "Value",
        cell: () => (
          <span className="font-mono text-[11px] text-[var(--nx-text-tertiary)]">
            •••••••• (hidden)
          </span>
        ),
        width: 120,
      },
      {
        id: "rotate",
        header: "Rotate",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "deployments:write")}>
            <Button
              size="sm"
              variant="secondary"
              onClick={(e) => {
                e.stopPropagation();
                setRotateTarget(row);
              }}
            >
              Propose rotate
            </Button>
          </PermissionGate>
        ),
        width: 130,
      },
    ],
    [session],
  );

  const proposalCols = useMemo<DataGridColumnDef<SecretRotateProposal>[]>(
    () => [
      { id: "name", header: "Secret", accessorKey: "secretName" },
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
                  loading={resolveSecret.isPending}
                  onClick={() =>
                    void resolveSecret.mutateAsync({
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
                    void resolveSecret.mutateAsync({
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
    [session, resolveSecret],
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
        <Skeleton height={220} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load deployments
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
        title="Deployments"
        description={`CI/CD · blue-green / canary / rolling · secrets metadata only${isFetching ? " · refreshing…" : ""}`}
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      />

      <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard title="Pipelines running" value={String(k.pipelinesRunning)} tone="brand" />
        <KpiCard title="Deploys 24h" value={String(k.deploys24h)} />
        <KpiCard
          title="Failed pipelines"
          value={String(k.failedPipelines)}
          tone={k.failedPipelines ? "danger" : "success"}
        />
        <KpiCard title="Active canaries" value={String(k.canariesActive)} tone="warning" />
        <KpiCard
          title="Rollbacks 24h"
          value={String(k.rollbacks24h)}
          tone={k.rollbacks24h ? "warning" : "neutral"}
        />
        <KpiCard
          title="Secrets due"
          value={String(k.secretsDueRotation)}
          tone={k.secretsDueRotation ? "warning" : "success"}
        />
      </div>

      <ChartFrame title="Deploy frequency (week)">
        <ReactECharts
          style={{ height: 200 }}
          option={barChartOption(data.deployFrequency, "#0B6E6E", "deploys")}
        />
      </ChartFrame>

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        CI/CD pipelines
      </h3>
      <DataGrid columns={pipeCols} data={data.pipelines} getRowId={(r) => r.id} />

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Active deployments
      </h3>
      <DataGrid
        columns={deployCols}
        data={data.deployments}
        getRowId={(r) => r.id}
      />

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Environments
      </h3>
      <DataGrid
        columns={envCols}
        data={data.environments}
        getRowId={(r) => r.id}
      />

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-1)] text-[var(--nx-font-size-title)] font-semibold">
          Secrets management
        </h3>
        <p className="m-0 mb-[var(--nx-space-3)] text-[12px] text-[var(--nx-text-tertiary)]">
          Values are never displayed. Rotation requires dual-control approval.
        </p>
        <DataGrid columns={secretCols} data={data.secrets} getRowId={(r) => r.id} />
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Pending secret rotations ({pendingSecrets.length})
        </h3>
        <DataGrid
          columns={proposalCols}
          data={pendingSecrets}
          getRowId={(r) => r.id}
          emptyMessage="No pending secret rotations"
        />
      </section>

      <ConfirmDialog
        open={rotateTarget != null}
        title="Propose secret rotation"
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)]">
            <p className="m-0 text-[13px]">
              Rotate <strong>{rotateTarget?.name}</strong> — value remains
              hidden; only version metadata updates after approval.
            </p>
            <Input
              placeholder="Reason"
              value={rotateReason}
              onChange={(e) => setRotateReason(e.target.value)}
            />
          </div>
        }
        confirmLabel="Submit proposal"
        loading={proposeSecret.isPending}
        onCancel={() => {
          setRotateTarget(null);
          setRotateReason("");
        }}
        onConfirm={() => {
          if (!rotateTarget || !session || !rotateReason.trim()) return;
          void proposeSecret
            .mutateAsync({
              secretId: rotateTarget.id,
              secretName: rotateTarget.name,
              reason: rotateReason.trim(),
              requesterId: session.userId,
              requesterEmail: session.email,
            })
            .then(() => {
              setRotateTarget(null);
              setRotateReason("");
            });
        }}
      />
    </div>
  );
}
