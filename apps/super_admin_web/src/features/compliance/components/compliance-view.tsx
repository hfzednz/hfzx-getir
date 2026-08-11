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
import {
  useAdvancePrivacyRequest,
  useCompliance,
  useCreatePrivacyRequest,
  useUpdateRetention,
} from "../hooks";
import type {
  ConsentRecord,
  PrivacyRegime,
  PrivacyRequest,
  PrivacyRequestType,
  RegimePanel,
  RetentionPolicy,
} from "../types";

function regimeTone(
  status: RegimePanel["status"],
): "success" | "warning" | "danger" {
  if (status === "compliant") return "success";
  if (status === "gaps") return "warning";
  return "danger";
}

function requestTone(
  status: PrivacyRequest["status"],
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "completed":
      return "success";
    case "rejected":
      return "danger";
    case "in_progress":
    case "verifying":
      return "warning";
    case "received":
      return "info";
    default:
      return "neutral";
  }
}

export function ComplianceView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useCompliance();
  const advance = useAdvancePrivacyRequest();
  const createReq = useCreatePrivacyRequest();
  const updateRet = useUpdateRetention();

  const [reqFilter, setReqFilter] = useState("open");
  const [createOpen, setCreateOpen] = useState(false);
  const [reqType, setReqType] = useState<PrivacyRequestType>("export");
  const [reqRegime, setReqRegime] = useState<PrivacyRegime>("gdpr");
  const [reqEmail, setReqEmail] = useState("");
  const [reqTenant, setReqTenant] = useState("");

  const openCount = useMemo(
    () =>
      (data?.requests ?? []).filter(
        (r) => r.status !== "completed" && r.status !== "rejected",
      ).length,
    [data?.requests],
  );

  const filteredRequests = useMemo(() => {
    const items = data?.requests ?? [];
    if (reqFilter === "all") return items;
    if (reqFilter === "open") {
      return items.filter(
        (r) => r.status !== "completed" && r.status !== "rejected",
      );
    }
    return items.filter((r) => r.type === reqFilter);
  }, [data?.requests, reqFilter]);

  const regimeCols = useMemo<DataGridColumnDef<RegimePanel>[]>(
    () => [
      { id: "label", header: "Regime", accessorKey: "label" },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={regimeTone(row.status)} />
        ),
        width: 110,
      },
      {
        id: "open",
        header: "Open requests",
        cell: ({ row }) => (
          <span className="tabular-nums">{row.openRequests}</span>
        ),
        width: 120,
      },
      {
        id: "ret",
        header: "Retention",
        cell: ({ row }) => (
          <StatusBadge
            status={row.retentionAligned ? "aligned" : "gap"}
            tone={row.retentionAligned ? "success" : "warning"}
          />
        ),
        width: 110,
      },
      {
        id: "dpa",
        header: "DPA",
        cell: ({ row }) => (row.dpaSigned ? "signed" : "missing"),
        width: 90,
      },
      { id: "notes", header: "Notes", accessorKey: "notes" },
    ],
    [],
  );

  const retentionCols = useMemo<DataGridColumnDef<RetentionPolicy>[]>(
    () => [
      { id: "class", header: "Data class", accessorKey: "dataClass" },
      {
        id: "regime",
        header: "Regime",
        cell: ({ row }) => row.regime.toUpperCase(),
        width: 90,
      },
      {
        id: "days",
        header: "Retention days",
        cell: ({ row }) => (
          <span className="tabular-nums">{row.retentionDays}</span>
        ),
        width: 120,
      },
      {
        id: "auto",
        header: "Auto-delete",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "compliance:write")}>
            <Button
              size="sm"
              variant="secondary"
              loading={updateRet.isPending}
              onClick={() =>
                void updateRet.mutateAsync({
                  policyId: row.id,
                  patch: { autoDelete: !row.autoDelete },
                })
              }
            >
              {row.autoDelete ? "On" : "Off"}
            </Button>
          </PermissionGate>
        ),
        width: 110,
      },
      {
        id: "hold",
        header: "Legal hold exempt",
        cell: ({ row }) => (row.legalHoldExempt ? "yes" : "no"),
        width: 140,
      },
    ],
    [session, updateRet],
  );

  const consentCols = useMemo<DataGridColumnDef<ConsentRecord>[]>(
    () => [
      { id: "subject", header: "Subject", accessorKey: "subjectId", width: 120 },
      { id: "purpose", header: "Purpose", accessorKey: "purpose" },
      {
        id: "regime",
        header: "Regime",
        cell: ({ row }) => row.regime.toUpperCase(),
        width: 80,
      },
      {
        id: "status",
        header: "Consent",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "granted"
                ? "success"
                : row.status === "pending"
                  ? "info"
                  : "warning"
            }
          />
        ),
        width: 110,
      },
      { id: "channel", header: "Channel", accessorKey: "channel", width: 120 },
      {
        id: "when",
        header: "Updated",
        cell: ({ row }) =>
          new Date(row.updatedAt).toLocaleDateString("en-US"),
        width: 110,
      },
    ],
    [],
  );

  const requestCols = useMemo<DataGridColumnDef<PrivacyRequest>[]>(
    () => [
      {
        id: "type",
        header: "Type",
        cell: ({ row }) => (
          <StatusBadge
            status={row.type}
            tone={
              row.type === "delete"
                ? "danger"
                : row.type === "export"
                  ? "info"
                  : "neutral"
            }
          />
        ),
        width: 100,
      },
      {
        id: "regime",
        header: "Regime",
        cell: ({ row }) => row.regime.toUpperCase(),
        width: 80,
      },
      { id: "email", header: "Subject", accessorKey: "subjectEmail" },
      { id: "tenant", header: "Tenant", accessorKey: "tenantName" },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={requestTone(row.status)} />
        ),
        width: 120,
      },
      {
        id: "due",
        header: "Due",
        cell: ({ row }) =>
          new Date(row.dueAt).toLocaleDateString("en-US"),
        width: 110,
      },
      {
        id: "actions",
        header: "Workflow",
        cell: ({ row }) => {
          if (row.status === "completed" || row.status === "rejected") {
            return (
              <span className="text-[11px] text-[var(--nx-text-tertiary)]">
                closed
              </span>
            );
          }
          const next: PrivacyRequest["status"] =
            row.status === "received"
              ? "verifying"
              : row.status === "verifying"
                ? "in_progress"
                : "completed";
          return (
            <PermissionGate
              allowed={
                can(session, "compliance:write") ||
                (row.type === "export" && can(session, "compliance:export"))
              }
            >
              <div className="flex gap-[var(--nx-space-1)]">
                <Button
                  size="sm"
                  variant="secondary"
                  loading={advance.isPending}
                  onClick={() =>
                    void advance.mutateAsync({
                      requestId: row.id,
                      status: next,
                      assignee: session?.email ?? null,
                    })
                  }
                >
                  Advance → {next.replace("_", " ")}
                </Button>
                {row.type === "delete" || row.type === "export" ? (
                  <Button
                    size="sm"
                    variant={row.type === "delete" ? "danger" : "primary"}
                    onClick={() =>
                      void advance.mutateAsync({
                        requestId: row.id,
                        status: "completed",
                        assignee: session?.email ?? null,
                      })
                    }
                  >
                    {row.type === "delete" ? "Execute delete" : "Complete export"}
                  </Button>
                ) : null}
              </div>
            </PermissionGate>
          );
        },
        width: 280,
      },
    ],
    [session, advance],
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
          Failed to load compliance
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
        title="Compliance"
        description={`GDPR · KVKK · CCPA · retention · consent · privacy requests · right to delete/export${isFetching ? " · refreshing…" : ""}`}
        actions={
          <PermissionGate allowed={can(session, "compliance:write")}>
            <Button size="sm" onClick={() => setCreateOpen(true)}>
              New privacy request
            </Button>
          </PermissionGate>
        }
      />

      <section
        aria-label="Compliance KPIs"
        className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]"
      >
        <KpiCard title="Open DSRs" value={String(openCount)} tone="warning" />
        <KpiCard
          title="Regimes at risk"
          value={String(data.regimes.filter((r) => r.status === "at_risk").length)}
          tone="danger"
        />
        <KpiCard
          title="Retention policies"
          value={String(data.retention.length)}
          tone="neutral"
        />
        <KpiCard
          title="Consent records"
          value={String(data.consents.length)}
          tone="brand"
        />
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Privacy regimes
        </h3>
        <DataGrid
          columns={regimeCols}
          data={data.regimes}
          getRowId={(r) => r.regime}
        />
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Retention policies
        </h3>
        <DataGrid
          columns={retentionCols}
          data={data.retention}
          getRowId={(r) => r.id}
        />
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Consent ledger
        </h3>
        <DataGrid
          columns={consentCols}
          data={data.consents}
          getRowId={(r) => r.id}
        />
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Privacy requests · delete / export workflows
        </h3>
        <FilterBar
          actions={
            <Button size="sm" variant="ghost" onClick={() => void refetch()}>
              Refresh
            </Button>
          }
        >
          <Select
            value={reqFilter}
            onChange={(e) => setReqFilter(e.target.value)}
          >
            <option value="open">Open</option>
            <option value="all">All</option>
            <option value="export">Export</option>
            <option value="delete">Delete</option>
            <option value="access">Access</option>
            <option value="rectify">Rectify</option>
            <option value="restrict">Restrict</option>
          </Select>
        </FilterBar>
        <DataGrid
          columns={requestCols}
          data={filteredRequests}
          getRowId={(r) => r.id}
          emptyMessage="No privacy requests"
        />
      </section>

      <ConfirmDialog
        open={createOpen}
        title="Create privacy request"
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)] mt-[var(--nx-space-2)]">
            <Select
              value={reqType}
              onChange={(e) =>
                setReqType(e.target.value as PrivacyRequestType)
              }
            >
              <option value="export">Right to export</option>
              <option value="delete">Right to delete</option>
              <option value="access">Access</option>
              <option value="rectify">Rectify</option>
              <option value="restrict">Restrict</option>
            </Select>
            <Select
              value={reqRegime}
              onChange={(e) =>
                setReqRegime(e.target.value as PrivacyRegime)
              }
            >
              <option value="gdpr">GDPR</option>
              <option value="kvkk">KVKK</option>
              <option value="ccpa">CCPA</option>
            </Select>
            <Input
              placeholder="Subject email"
              value={reqEmail}
              onChange={(e) => setReqEmail(e.target.value)}
            />
            <Input
              placeholder="Tenant name"
              value={reqTenant}
              onChange={(e) => setReqTenant(e.target.value)}
            />
          </div>
        }
        confirmLabel="Create"
        loading={createReq.isPending}
        onCancel={() => setCreateOpen(false)}
        onConfirm={() => {
          if (!reqEmail.trim() || !reqTenant.trim()) return;
          void createReq
            .mutateAsync({
              type: reqType,
              regime: reqRegime,
              subjectEmail: reqEmail.trim(),
              tenantName: reqTenant.trim(),
            })
            .then(() => {
              setCreateOpen(false);
              setReqEmail("");
              setReqTenant("");
            });
        }}
      />
    </div>
  );
}
