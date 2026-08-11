"use client";

import { useMemo, useState } from "react";
import dynamic from "next/dynamic";
import {
  Button,
  ChartFrame,
  DataGrid,
  FilterBar,
  KpiCard,
  PageHeader,
  PermissionGate,
  Select,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { barChartOption } from "@/shared/lib/charts";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/platform-permissions";
import { useNotifications, useSetProviderStatus } from "../hooks";
import type {
  DeliveryEvent,
  NotificationProvider,
  NotificationTemplate,
  ProviderStatus,
} from "../types";

const ReactECharts = dynamic(() => import("echarts-for-react"), { ssr: false });

function providerTone(
  s: ProviderStatus,
): "success" | "warning" | "danger" | "neutral" {
  if (s === "active") return "success";
  if (s === "degraded") return "warning";
  return "neutral";
}

export function NotificationsView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useNotifications();
  const statusMutation = useSetProviderStatus();
  const [channel, setChannel] = useState("all");

  const providers = useMemo(() => {
    const items = data?.providers ?? [];
    if (channel === "all") return items;
    return items.filter((p) => p.channel === channel);
  }, [data?.providers, channel]);

  const providerCols = useMemo<DataGridColumnDef<NotificationProvider>[]>(
    () => [
      { id: "name", header: "Provider", accessorKey: "name" },
      {
        id: "ch",
        header: "Channel",
        cell: ({ row }) => (
          <StatusBadge status={row.channel} tone="info" />
        ),
        width: 100,
      },
      { id: "vendor", header: "Vendor", accessorKey: "vendor", width: 110 },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={providerTone(row.status)} />
        ),
        width: 100,
      },
      {
        id: "sent",
        header: "Sent today",
        cell: ({ row }) => row.sentToday.toLocaleString("en-US"),
        width: 110,
      },
      {
        id: "ok",
        header: "Success %",
        cell: ({ row }) => `${row.successPct}%`,
        width: 90,
      },
      {
        id: "actions",
        header: "Control",
        cell: ({ row }) => (
          <PermissionGate allowed={can(session, "notifications:write")}>
            <div
              className="flex gap-[var(--nx-space-1)]"
              onClick={(e) => e.stopPropagation()}
            >
              {row.status !== "active" ? (
                <Button
                  size="sm"
                  variant="secondary"
                  loading={statusMutation.isPending}
                  onClick={() =>
                    void statusMutation.mutateAsync({
                      providerId: row.id,
                      status: "active",
                    })
                  }
                >
                  Enable
                </Button>
              ) : (
                <Button
                  size="sm"
                  variant="ghost"
                  loading={statusMutation.isPending}
                  onClick={() =>
                    void statusMutation.mutateAsync({
                      providerId: row.id,
                      status: "disabled",
                    })
                  }
                >
                  Disable
                </Button>
              )}
            </div>
          </PermissionGate>
        ),
        width: 110,
      },
    ],
    [session, statusMutation],
  );

  const templateCols = useMemo<DataGridColumnDef<NotificationTemplate>[]>(
    () => [
      { id: "name", header: "Template", accessorKey: "name" },
      {
        id: "ch",
        header: "Channel",
        cell: ({ row }) => (
          <StatusBadge status={row.channel} tone="info" />
        ),
        width: 100,
      },
      { id: "locale", header: "Locale", accessorKey: "locale", width: 80 },
      {
        id: "ver",
        header: "Ver",
        cell: ({ row }) => String(row.version),
        width: 60,
      },
      { id: "owner", header: "Owner", accessorKey: "owner", width: 140 },
    ],
    [],
  );

  const deliveryCols = useMemo<DataGridColumnDef<DeliveryEvent>[]>(
    () => [
      { id: "provider", header: "Provider", accessorKey: "providerName" },
      {
        id: "ch",
        header: "Channel",
        cell: ({ row }) => (
          <StatusBadge status={row.channel} tone="info" />
        ),
        width: 100,
      },
      { id: "tpl", header: "Template", accessorKey: "template" },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "delivered" || row.status === "sent"
                ? "success"
                : row.status === "failed" || row.status === "bounced"
                  ? "danger"
                  : "warning"
            }
          />
        ),
        width: 100,
      },
      {
        id: "rcpt",
        header: "Recipient",
        accessorKey: "recipientHash",
        width: 130,
      },
      {
        id: "lat",
        header: "ms",
        cell: ({ row }) => String(row.latencyMs),
        width: 60,
      },
      {
        id: "err",
        header: "Error",
        cell: ({ row }) => row.errorCode ?? "—",
        width: 140,
      },
    ],
    [],
  );

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
          Failed to load notifications hub
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
        title="Notifications"
        description={`Provider hub — email / SMS / push / WhatsApp / webhooks (not city-ops inbox)${isFetching ? " · refreshing…" : ""}`}
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      />

      <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard title="Providers active" value={String(k.providersActive)} tone="success" />
        <KpiCard title="Delivered 24h" value={k.delivered24h.toLocaleString("en-US")} />
        <KpiCard
          title="Failed 24h"
          value={k.failed24h.toLocaleString("en-US")}
          tone="warning"
        />
        <KpiCard title="Avg latency" value={`${k.avgLatencyMs} ms`} />
        <KpiCard title="Templates" value={String(k.templates)} />
        <KpiCard title="Webhook success" value={`${k.webhookSuccessPct}%`} tone="success" />
      </div>

      <ChartFrame title="Outbound volume">
        <ReactECharts
          style={{ height: 200 }}
          option={barChartOption(data.volumeSeries, "#0B6E6E", "msgs")}
        />
      </ChartFrame>

      <FilterBar>
        <Select
          value={channel}
          onChange={(e) => setChannel(e.target.value)}
          aria-label="Filter channel"
        >
          <option value="all">All channels</option>
          <option value="email">Email</option>
          <option value="sms">SMS</option>
          <option value="push">Push</option>
          <option value="whatsapp">WhatsApp</option>
          <option value="webhook">Webhook</option>
        </Select>
      </FilterBar>

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Providers
      </h3>
      <DataGrid columns={providerCols} data={providers} getRowId={(r) => r.id} />

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Templates
      </h3>
      <DataGrid
        columns={templateCols}
        data={data.templates}
        getRowId={(r) => r.id}
      />

      <h3 className="m-0 text-[var(--nx-font-size-title)] font-semibold">
        Delivery tracking
      </h3>
      <DataGrid
        columns={deliveryCols}
        data={data.deliveries}
        getRowId={(r) => r.id}
      />
    </div>
  );
}
