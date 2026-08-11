"use client";

import { useMemo, useState } from "react";
import {
  Button,
  ConfirmDialog,
  DataGrid,
  type DataGridColumnDef,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
} from "@nexora/ui";
import { usePermission } from "@/shared/permissions/use-permission";
import { useSystemSnapshot } from "../hooks";
import type { FeatureFlag } from "../types";

export function SystemFlagsView() {
  const { data, isLoading, isError, error, refetch } = useSystemSnapshot();
  const canFlags = usePermission("system:flags");
  const canWrite = usePermission("system:write");
  const [pendingKill, setPendingKill] = useState<FeatureFlag | null>(null);
  const [localFlags, setLocalFlags] = useState<FeatureFlag[] | null>(null);

  const flags = localFlags ?? data?.flags ?? [];

  const columns: DataGridColumnDef<FeatureFlag>[] = useMemo(
    () => [
      { id: "key", header: "Flag", accessorKey: "key" },
      { id: "desc", header: "Description", accessorKey: "description" },
      {
        id: "kill",
        header: "Type",
        cell: ({ row }) =>
          row.killSwitch ? (
            <StatusBadge status="kill switch" tone="danger" />
          ) : (
            <StatusBadge status="feature" tone="info" />
          ),
      },
      {
        id: "enabled",
        header: "State",
        cell: ({ row }) => (
          <StatusBadge
            status={row.enabled ? "on" : "off"}
            tone={row.enabled ? "success" : "neutral"}
          />
        ),
      },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => {
          if (row.killSwitch) {
            return (
              <PermissionGate
                allowed={canFlags}
                fallback={
                  <span className="text-[11px] text-[var(--nx-text-tertiary)]">
                    super_admin only
                  </span>
                }
              >
                <Button
                  size="sm"
                  variant={row.enabled ? "danger" : "secondary"}
                  onClick={() => setPendingKill(row)}
                >
                  {row.enabled ? "Disable kill" : "Arm kill"}
                </Button>
              </PermissionGate>
            );
          }
          return (
            <PermissionGate
              allowed={canWrite || canFlags}
              fallback={
                <span className="text-[11px] text-[var(--nx-text-tertiary)]">
                  read-only
                </span>
              }
            >
              <Button
                size="sm"
                variant="secondary"
                onClick={() => {
                  setLocalFlags((prev) => {
                    const base = prev ?? data?.flags ?? [];
                    return base.map((f) =>
                      f.id === row.id ? { ...f, enabled: !f.enabled } : f,
                    );
                  });
                }}
              >
                Toggle
              </Button>
            </PermissionGate>
          );
        },
      },
    ],
    [canFlags, canWrite, data?.flags],
  );

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
          Failed to load flags
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
        description="Feature toggles and kill switches. Kill switches require super_admin (system:flags)."
      />

      {!canFlags ? (
        <p className="m-0 text-[13px] text-[var(--nx-text-secondary)] rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-subtle)] bg-[var(--nx-bg-sunken)] p-[var(--nx-space-3)]">
          You can view flags. Mutating kill switches is gated to{" "}
          <code className="font-[family-name:var(--nx-font-mono)] text-[12px]">
            super_admin
          </code>
          .
        </p>
      ) : null}

      <DataGrid columns={columns} data={flags} getRowId={(r) => r.id} />

      <ConfirmDialog
        open={pendingKill != null}
        title={
          pendingKill?.enabled ? "Disable kill switch?" : "Arm kill switch?"
        }
        description={
          pendingKill
            ? `${pendingKill.key} — ${pendingKill.description}. This action is audited.`
            : undefined
        }
        confirmLabel={pendingKill?.enabled ? "Disable" : "Arm"}
        danger
        onCancel={() => setPendingKill(null)}
        onConfirm={() => {
          if (!pendingKill) return;
          setLocalFlags((prev) => {
            const base = prev ?? data.flags;
            return base.map((f) =>
              f.id === pendingKill.id ? { ...f, enabled: !f.enabled } : f,
            );
          });
          setPendingKill(null);
        }}
      />
    </div>
  );
}
