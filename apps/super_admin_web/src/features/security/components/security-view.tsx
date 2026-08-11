"use client";

import { useMemo, useState } from "react";
import {
  Button,
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
  useAcknowledgeThreat,
  useRevokeSession,
  useSecurity,
  useToggleGeoRule,
  useTogglePolicy,
  useToggleProvider,
} from "../hooks";
import type {
  AuthProvider,
  GeoIpRule,
  LoginEvent,
  RegisteredDevice,
  SecurityPolicy,
  SuspiciousSession,
  ThreatAlert,
  ThreatSeverity,
} from "../types";

function threatTone(
  severity: ThreatSeverity,
): "danger" | "warning" | "info" | "neutral" {
  switch (severity) {
    case "critical":
    case "high":
      return "danger";
    case "medium":
      return "warning";
    case "low":
      return "info";
    default:
      return "neutral";
  }
}

export function SecurityView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useSecurity();
  const ack = useAcknowledgeThreat();
  const revoke = useRevokeSession();
  const toggleRule = useToggleGeoRule();
  const toggleProv = useToggleProvider();
  const togglePol = useTogglePolicy();

  const [loginQ, setLoginQ] = useState("");
  const [loginFilter, setLoginFilter] = useState("all");

  const loginFiltered = useMemo(() => {
    const items = data?.loginEvents ?? [];
    return items.filter((e) => {
      if (loginFilter === "failed" && e.success) return false;
      if (loginFilter === "success" && !e.success) return false;
      if (!loginQ.trim()) return true;
      const needle = loginQ.trim().toLowerCase();
      return (
        e.userEmail.toLowerCase().includes(needle) ||
        e.ip.includes(needle) ||
        e.country.toLowerCase().includes(needle)
      );
    });
  }, [data?.loginEvents, loginQ, loginFilter]);

  const loginCols = useMemo<DataGridColumnDef<LoginEvent>[]>(
    () => [
      { id: "email", header: "User", accessorKey: "userEmail" },
      { id: "ip", header: "IP", accessorKey: "ip", width: 130 },
      {
        id: "geo",
        header: "Geo",
        cell: ({ row }) => `${row.city}, ${row.country}`,
        width: 140,
      },
      {
        id: "result",
        header: "Result",
        cell: ({ row }) => (
          <StatusBadge
            status={row.success ? "ok" : "fail"}
            tone={row.success ? "success" : "danger"}
          />
        ),
        width: 80,
      },
      {
        id: "mfa",
        header: "MFA",
        cell: ({ row }) => (row.mfaUsed ? "yes" : "no"),
        width: 60,
      },
      {
        id: "when",
        header: "When",
        cell: ({ row }) =>
          new Date(row.createdAt).toLocaleString("en-US"),
        width: 160,
      },
    ],
    [],
  );

  const threatCols = useMemo<DataGridColumnDef<ThreatAlert>[]>(
    () => [
      {
        id: "sev",
        header: "Severity",
        cell: ({ row }) => (
          <StatusBadge status={row.severity} tone={threatTone(row.severity)} />
        ),
        width: 100,
      },
      { id: "title", header: "Threat", accessorKey: "title" },
      { id: "detail", header: "Detail", accessorKey: "detail" },
      { id: "source", header: "Source", accessorKey: "source", width: 120 },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "open"
                ? "danger"
                : row.status === "acknowledged"
                  ? "warning"
                  : "success"
            }
          />
        ),
        width: 120,
      },
      {
        id: "actions",
        header: "",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "security:write")}>
            <Button
              size="sm"
              variant="secondary"
              disabled={row.status !== "open"}
              loading={ack.isPending}
              onClick={() => void ack.mutateAsync(row.id)}
            >
              Ack
            </Button>
          </PermissionGate>
        ),
        width: 90,
      },
    ],
    [session, ack],
  );

  const sessionCols = useMemo<DataGridColumnDef<SuspiciousSession>[]>(
    () => [
      { id: "email", header: "User", accessorKey: "userEmail" },
      { id: "device", header: "Device", accessorKey: "deviceLabel" },
      {
        id: "geo",
        header: "IP / geo",
        cell: ({ row }) => `${row.ip} · ${row.country}`,
      },
      {
        id: "risk",
        header: "Risk",
        cell: ({ row }) => (
          <StatusBadge
            status={row.risk}
            tone={
              row.risk === "blocked"
                ? "danger"
                : row.risk === "suspicious"
                  ? "warning"
                  : "success"
            }
          />
        ),
        width: 110,
      },
      { id: "reason", header: "Reason", accessorKey: "reason" },
      {
        id: "actions",
        header: "",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "security:write")}>
            <Button
              size="sm"
              variant="danger"
              disabled={row.risk === "blocked"}
              loading={revoke.isPending}
              onClick={() => void revoke.mutateAsync(row.id)}
            >
              Revoke
            </Button>
          </PermissionGate>
        ),
        width: 100,
      },
    ],
    [session, revoke],
  );

  const deviceCols = useMemo<DataGridColumnDef<RegisteredDevice>[]>(
    () => [
      { id: "email", header: "User", accessorKey: "userEmail" },
      { id: "label", header: "Device", accessorKey: "label" },
      { id: "platform", header: "Platform", accessorKey: "platform", width: 100 },
      {
        id: "trusted",
        header: "Trusted",
        cell: ({ row }) => (
          <StatusBadge
            status={row.trusted ? "yes" : "no"}
            tone={row.trusted ? "success" : "warning"}
          />
        ),
        width: 90,
      },
      {
        id: "seen",
        header: "Last seen",
        cell: ({ row }) =>
          new Date(row.lastSeenAt).toLocaleString("en-US"),
        width: 160,
      },
    ],
    [],
  );

  const ruleCols = useMemo<DataGridColumnDef<GeoIpRule>[]>(
    () => [
      {
        id: "type",
        header: "Type",
        cell: ({ row }) => (
          <StatusBadge
            status={row.type}
            tone={row.type === "allow" ? "success" : "danger"}
          />
        ),
        width: 90,
      },
      { id: "target", header: "Target", accessorKey: "target", width: 100 },
      { id: "value", header: "Value", accessorKey: "value" },
      { id: "label", header: "Label", accessorKey: "label" },
      {
        id: "enabled",
        header: "Enabled",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "security:write")}>
            <Button
              size="sm"
              variant="secondary"
              onClick={() => void toggleRule.mutateAsync(row.id)}
            >
              {row.enabled ? "On" : "Off"}
            </Button>
          </PermissionGate>
        ),
        width: 90,
      },
    ],
    [session, toggleRule],
  );

  const providerCols = useMemo<DataGridColumnDef<AuthProvider>[]>(
    () => [
      { id: "name", header: "Provider", accessorKey: "name" },
      {
        id: "kind",
        header: "Kind",
        cell: ({ row }) => <StatusBadge status={row.kind} tone="info" />,
        width: 100,
      },
      {
        id: "enforced",
        header: "Enforced",
        cell: ({ row }) => (row.enforced ? "yes" : "no"),
        width: 90,
      },
      {
        id: "scope",
        header: "Scope",
        cell: ({ row }) => row.tenantScope ?? "platform",
        width: 110,
      },
      {
        id: "enabled",
        header: "Enabled",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "security:write")}>
            <Button
              size="sm"
              variant="secondary"
              onClick={() => void toggleProv.mutateAsync(row.id)}
            >
              {row.enabled ? "On" : "Off"}
            </Button>
          </PermissionGate>
        ),
        width: 90,
      },
    ],
    [session, toggleProv],
  );

  const policyCols = useMemo<DataGridColumnDef<SecurityPolicy>[]>(
    () => [
      { id: "name", header: "Policy", accessorKey: "name" },
      { id: "desc", header: "Description", accessorKey: "description" },
      { id: "value", header: "Value", accessorKey: "value", width: 120 },
      {
        id: "enforced",
        header: "Enforced",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "security:write")}>
            <Button
              size="sm"
              variant="secondary"
              onClick={() => void togglePol.mutateAsync(row.id)}
            >
              {row.enforced ? "On" : "Off"}
            </Button>
          </PermissionGate>
        ),
        width: 90,
      },
    ],
    [session, togglePol],
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
          Failed to load security command center
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
        title="Security command center"
        description={`Login monitoring · threats · sessions · devices · IP/geo · 2FA/SSO/OAuth · policies${isFetching ? " · refreshing…" : ""}`}
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      />

      <section
        aria-label="Security KPIs"
        className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]"
      >
        <KpiCard
          title="Failed logins (24h)"
          value={data.failedLogins24h.toLocaleString("en-US")}
          tone="danger"
        />
        <KpiCard
          title="Open threats"
          value={String(data.openThreats)}
          tone="warning"
        />
        <KpiCard
          title="Suspicious sessions"
          value={String(
            data.sessions.filter((s) => s.risk !== "normal").length,
          )}
          tone="warning"
        />
        <KpiCard
          title="Trusted devices"
          value={String(data.devices.filter((d) => d.trusted).length)}
          tone="success"
        />
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Threat detection
        </h3>
        <DataGrid
          columns={threatCols}
          data={data.threats}
          getRowId={(r) => r.id}
        />
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          Login monitoring
        </h3>
        <FilterBar>
          <Input
            placeholder="Search email, IP, country…"
            value={loginQ}
            onChange={(e) => setLoginQ(e.target.value)}
          />
          <Select
            value={loginFilter}
            onChange={(e) => setLoginFilter(e.target.value)}
          >
            <option value="all">All</option>
            <option value="success">Success</option>
            <option value="failed">Failed</option>
          </Select>
        </FilterBar>
        <DataGrid
          columns={loginCols}
          data={loginFiltered}
          getRowId={(r) => r.id}
        />
      </section>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-[var(--nx-space-3)]">
        <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            Suspicious sessions
          </h3>
          <DataGrid
            columns={sessionCols}
            data={data.sessions}
            getRowId={(r) => r.id}
          />
        </section>
        <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            Devices
          </h3>
          <DataGrid
            columns={deviceCols}
            data={data.devices}
            getRowId={(r) => r.id}
          />
        </section>
      </div>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
          IP / geo restrictions
        </h3>
        <DataGrid
          columns={ruleCols}
          data={data.geoIpRules}
          getRowId={(r) => r.id}
        />
      </section>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-[var(--nx-space-3)]">
        <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            2FA / SSO / OAuth providers
          </h3>
          <DataGrid
            columns={providerCols}
            data={data.providers}
            getRowId={(r) => r.id}
          />
        </section>
        <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
          <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
            Security policies
          </h3>
          <DataGrid
            columns={policyCols}
            data={data.policies}
            getRowId={(r) => r.id}
          />
        </section>
      </div>
    </div>
  );
}
