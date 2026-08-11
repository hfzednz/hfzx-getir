"use client";

import Link from "next/link";
import {
  Button,
  DataGrid,
  type DataGridColumnDef,
  PageHeader,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@nexora/ui";
import { useSystemSnapshot } from "../hooks";
import type { DeliveryZoneLink, SystemSetting } from "../types";

const settingCols: DataGridColumnDef<SystemSetting>[] = [
  { id: "key", header: "Key", accessorKey: "key" },
  { id: "value", header: "Value", accessorKey: "value" },
  { id: "category", header: "Category", accessorKey: "category" },
  { id: "desc", header: "Description", accessorKey: "description" },
];

const zoneCols: DataGridColumnDef<DeliveryZoneLink>[] = [
  { id: "name", header: "Zone", accessorKey: "name" },
  { id: "city", header: "City", accessorKey: "city" },
  {
    id: "link",
    header: "Manage",
    cell: ({ row }) => (
      <Link
        href={row.href}
        className="text-[var(--nx-text-link)] text-[12px] underline"
      >
        Open zones
      </Link>
    ),
  },
];

export function SystemConfigView() {
  const { data, isLoading, isError, error, refetch } = useSystemSnapshot();

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
          Failed to load system config
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

  const byCategory = (cat: SystemSetting["category"]) =>
    data.settings.filter((s) => s.category === cat);

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="System"
        description="Config, localization, currencies, taxes & regions"
        actions={
          <div className="flex gap-[var(--nx-space-2)]">
            <Link href="/system/flags">
              <Button variant="secondary" size="sm">
                Feature flags
              </Button>
            </Link>
            <Link href="/system/templates">
              <Button variant="secondary" size="sm">
                Templates
              </Button>
            </Link>
          </div>
        }
      />

      <div className="flex flex-wrap gap-[var(--nx-space-2)]">
        {data.locales.map((l) => (
          <StatusBadge key={l} status={l} tone="info" />
        ))}
        {data.currencies.map((c) => (
          <StatusBadge key={c} status={c} tone="success" />
        ))}
      </div>

      <Tabs defaultValue="app">
        <TabsList>
          <TabsTrigger value="app">App settings</TabsTrigger>
          <TabsTrigger value="locale">Localization</TabsTrigger>
          <TabsTrigger value="currency">Currencies</TabsTrigger>
          <TabsTrigger value="tax">Taxes</TabsTrigger>
          <TabsTrigger value="region">Regions</TabsTrigger>
          <TabsTrigger value="zones">Delivery zones</TabsTrigger>
        </TabsList>
        {(
          [
            ["app", "app"],
            ["locale", "locale"],
            ["currency", "currency"],
            ["tax", "tax"],
            ["region", "region"],
          ] as const
        ).map(([tab, cat]) => (
          <TabsContent key={tab} value={tab} className="mt-[var(--nx-space-3)]">
            <DataGrid
              columns={settingCols}
              data={byCategory(cat)}
              getRowId={(r) => r.id}
            />
          </TabsContent>
        ))}
        <TabsContent value="zones" className="mt-[var(--nx-space-3)]">
          <DataGrid
            columns={zoneCols}
            data={data.zones}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
