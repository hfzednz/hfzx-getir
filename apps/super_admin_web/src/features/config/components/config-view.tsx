"use client";

import {
  Button,
  DataGrid,
  PageHeader,
  PermissionGate,
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
import { useConfigSnapshot } from "../hooks";
import type { ConfigSetting, NotificationProviderConfig } from "../types";

const settingCols: DataGridColumnDef<ConfigSetting>[] = [
  { id: "key", header: "Key", accessorKey: "key" },
  { id: "value", header: "Value", accessorKey: "value" },
  { id: "category", header: "Category", accessorKey: "category", width: 120 },
  { id: "desc", header: "Description", accessorKey: "description" },
];

const providerCols: DataGridColumnDef<NotificationProviderConfig>[] = [
  { id: "name", header: "Provider", accessorKey: "name" },
  { id: "channel", header: "Channel", accessorKey: "channel", width: 100 },
  {
    id: "status",
    header: "Status",
    cell: ({ row }) => (
      <StatusBadge
        status={row.status}
        tone={
          row.status === "configured"
            ? "success"
            : row.status === "error"
              ? "danger"
              : "neutral"
        }
      />
    ),
  },
  { id: "endpoint", header: "Endpoint", accessorKey: "endpoint" },
];

export function ConfigView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useConfigSnapshot();

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
          Failed to load config
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

  const byCategory = (cat: ConfigSetting["category"]) =>
    data.settings.filter((s) => s.category === cat);

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Platform config"
        description={`Settings, brand, API limits, localization & notification providers (config only)${isFetching ? " · refreshing…" : ""}`}
        actions={
          <PermissionGate allowed={can(session, "config:write")}>
            <Button size="sm" variant="secondary" onClick={() => void refetch()}>
              Refresh
            </Button>
          </PermissionGate>
        }
      />

      <div className="flex flex-wrap gap-[var(--nx-space-2)]">
        {data.locales.map((l) => (
          <StatusBadge key={l} status={l} tone="info" />
        ))}
        {data.currencies.map((c) => (
          <StatusBadge key={c} status={c} tone="success" />
        ))}
        {data.regions.map((r) => (
          <StatusBadge key={r} status={r} tone="neutral" />
        ))}
        {data.taxEngines.map((t) => (
          <StatusBadge key={t} status={t} tone="warning" />
        ))}
      </div>

      <Tabs defaultValue="platform">
        <TabsList>
          <TabsTrigger value="platform">Platform</TabsTrigger>
          <TabsTrigger value="brand">Brand</TabsTrigger>
          <TabsTrigger value="api">API / rate limits</TabsTrigger>
          <TabsTrigger value="locale">Localization</TabsTrigger>
          <TabsTrigger value="currency">Currencies</TabsTrigger>
          <TabsTrigger value="region">Regions</TabsTrigger>
          <TabsTrigger value="tax">Tax engines</TabsTrigger>
          <TabsTrigger value="notification">Notification providers</TabsTrigger>
        </TabsList>

        {(
          [
            ["platform", "platform"],
            ["brand", "brand"],
            ["api", "api"],
            ["locale", "locale"],
            ["currency", "currency"],
            ["region", "region"],
            ["tax", "tax"],
          ] as const
        ).map(([tab, cat]) => (
          <TabsContent key={tab} value={tab}>
            <DataGrid
              columns={settingCols}
              data={byCategory(cat)}
              getRowId={(r) => r.id}
            />
          </TabsContent>
        ))}

        <TabsContent value="notification">
          <p className="m-0 mb-[var(--nx-space-3)] text-[12px] text-[var(--nx-text-secondary)]">
            Provider credentials are configured here; message composition lives
            under Notifications hub.
          </p>
          <DataGrid
            columns={settingCols}
            data={byCategory("notification")}
            getRowId={(r) => r.id}
          />
          <div className="mt-[var(--nx-space-3)]">
            <DataGrid
              columns={providerCols}
              data={data.notificationProviders}
              getRowId={(r) => r.id}
            />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
