"use client";

import { useState, type ReactNode } from "react";
import Link from "next/link";
import {
  Button,
  DataGrid,
  PageHeader,
  PermissionGate,
  Select,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/platform-permissions";
import { useTenant, useUpdateTenantIsolation } from "../hooks";
import type {
  TenantBackup,
  TenantIsolationMode,
  TenantMonitorMetric,
} from "../types";

function Panel({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
      <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
        {title}
      </h3>
      {children}
    </section>
  );
}

const backupCols: DataGridColumnDef<TenantBackup>[] = [
  { id: "label", header: "Backup", accessorKey: "label" },
  {
    id: "status",
    header: "Status",
    cell: ({ row }) => (
      <StatusBadge
        status={row.status}
        tone={
          row.status === "ok"
            ? "success"
            : row.status === "running"
              ? "info"
              : row.status === "stale"
                ? "warning"
                : "danger"
        }
      />
    ),
  },
  {
    id: "size",
    header: "Size",
    cell: ({ row }) => `${row.sizeGb.toFixed(1)} GB`,
    align: "right",
  },
  {
    id: "at",
    header: "Taken",
    cell: ({ row }) => new Date(row.takenAt).toLocaleString("en-US"),
  },
];

const monitorCols: DataGridColumnDef<TenantMonitorMetric>[] = [
  { id: "label", header: "Metric", accessorKey: "label" },
  { id: "value", header: "Value", accessorKey: "value" },
  {
    id: "tone",
    header: "Health",
    cell: ({ row }) => <StatusBadge status={row.tone} tone={row.tone} />,
  },
];

export function TenantDetailView({ tenantId }: { tenantId: string }) {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch } = useTenant(tenantId);
  const isolationMutation = useUpdateTenantIsolation();
  const [mode, setMode] = useState<TenantIsolationMode | "">("");

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={220} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load tenant
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

  const selectedMode = mode || data.isolationMode;

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title={data.name}
        description={`${data.slug} · ${data.companyName} · ${data.region}`}
        actions={
          <Link href="/tenants">
            <Button size="sm" variant="ghost">
              Back to tenants
            </Button>
          </Link>
        }
      />

      <div className="flex flex-wrap gap-[var(--nx-space-2)]">
        <StatusBadge status={data.status} tone="info" />
        <StatusBadge status={data.isolationMode} tone="success" />
        <StatusBadge
          status={`migration: ${data.migration.status}`}
          tone={data.migration.status === "in_progress" ? "warning" : "neutral"}
        />
      </div>

      <p className="m-0 text-[13px] text-[var(--nx-text-secondary)]">
        {data.description}
      </p>

      <Tabs defaultValue="config">
        <TabsList>
          <TabsTrigger value="config">Config</TabsTrigger>
          <TabsTrigger value="isolation">Isolation</TabsTrigger>
          <TabsTrigger value="custom">Customization</TabsTrigger>
          <TabsTrigger value="migration">Migration</TabsTrigger>
          <TabsTrigger value="backups">Backups</TabsTrigger>
          <TabsTrigger value="monitor">Monitoring</TabsTrigger>
        </TabsList>

        <TabsContent value="config">
          <Panel title="Tenant config">
            <dl className="m-0 grid grid-cols-1 md:grid-cols-2 gap-[var(--nx-space-3)] text-[13px]">
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Feature pack</dt>
                <dd className="m-0 font-medium">{data.config.featurePack}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Data residency</dt>
                <dd className="m-0 font-medium">{data.config.dataResidency}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Max warehouses</dt>
                <dd className="m-0 font-medium tabular-nums">
                  {data.config.maxWarehouses}
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Max users</dt>
                <dd className="m-0 font-medium tabular-nums">
                  {data.config.maxUsers}
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">RLS</dt>
                <dd className="m-0 font-medium">
                  {data.config.rlsEnabled ? "Enabled" : "N/A (separate DB)"}
                </dd>
              </div>
            </dl>
          </Panel>
        </TabsContent>

        <TabsContent value="isolation">
          <Panel title="Isolation mode">
            <p className="m-0 mb-[var(--nx-space-3)] text-[13px] text-[var(--nx-text-secondary)]">
              Shared = RLS · Separate = dedicated DB · Hybrid = shared catalog +
              isolated PII/ledger
            </p>
            <div className="flex flex-wrap items-end gap-[var(--nx-space-2)]">
              <Select
                value={selectedMode}
                onChange={(e) =>
                  setMode(e.target.value as TenantIsolationMode)
                }
                aria-label="Isolation mode"
              >
                <option value="shared">Shared</option>
                <option value="hybrid">Hybrid</option>
                <option value="separate">Separate</option>
              </Select>
              <PermissionGate allowed={can(session, "tenants:write")}>
                <Button
                  size="sm"
                  loading={isolationMutation.isPending}
                  disabled={selectedMode === data.isolationMode}
                  onClick={() =>
                    void isolationMutation.mutateAsync({
                      id: data.id,
                      isolationMode: selectedMode,
                    })
                  }
                >
                  Save isolation
                </Button>
              </PermissionGate>
            </div>
          </Panel>
        </TabsContent>

        <TabsContent value="custom">
          <Panel title="Customization">
            <dl className="m-0 grid grid-cols-1 md:grid-cols-2 gap-[var(--nx-space-3)] text-[13px]">
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Primary color</dt>
                <dd className="m-0 font-medium flex items-center gap-2">
                  <span
                    className="inline-block w-4 h-4 rounded-sm border border-[var(--nx-border-subtle)]"
                    style={{ background: data.customization.primaryColor }}
                  />
                  {data.customization.primaryColor}
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Logo</dt>
                <dd className="m-0 font-medium">{data.customization.logoUrl}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Custom domain</dt>
                <dd className="m-0 font-medium">
                  {data.customization.customDomain ?? "—"}
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">White label</dt>
                <dd className="m-0 font-medium">
                  {data.customization.whiteLabel ? "Yes" : "No"}
                </dd>
              </div>
            </dl>
          </Panel>
        </TabsContent>

        <TabsContent value="migration">
          <Panel title="Migration status">
            <dl className="m-0 grid grid-cols-1 md:grid-cols-2 gap-[var(--nx-space-3)] text-[13px]">
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Status</dt>
                <dd className="m-0 font-medium">{data.migration.status}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Target mode</dt>
                <dd className="m-0 font-medium">
                  {data.migration.targetMode ?? "—"}
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Progress</dt>
                <dd className="m-0 font-medium tabular-nums">
                  {data.migration.progressPct}%
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Updated</dt>
                <dd className="m-0 font-medium">
                  {new Date(data.migration.updatedAt).toLocaleString("en-US")}
                </dd>
              </div>
            </dl>
            <p className="m-0 mt-[var(--nx-space-3)] text-[13px] text-[var(--nx-text-secondary)]">
              {data.migration.lastMessage}
            </p>
          </Panel>
        </TabsContent>

        <TabsContent value="backups">
          <DataGrid
            columns={backupCols}
            data={data.backups}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="monitor">
          <DataGrid
            columns={monitorCols}
            data={data.monitoring}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
