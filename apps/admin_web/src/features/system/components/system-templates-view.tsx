"use client";

import {
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
import type { MessageTemplate } from "../types";

const templateCols: DataGridColumnDef<MessageTemplate>[] = [
  { id: "name", header: "Name", accessorKey: "name" },
  {
    id: "channel",
    header: "Channel",
    accessorKey: "channel",
    cell: ({ value }) => (
      <StatusBadge status={String(value)} tone="info" />
    ),
  },
  { id: "locale", header: "Locale", accessorKey: "locale" },
  {
    id: "subject",
    header: "Subject",
    accessorKey: "subject",
    cell: ({ value }) => (value ? String(value) : "—"),
  },
  { id: "body", header: "Preview", accessorKey: "bodyPreview" },
];

export function SystemTemplatesView() {
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
          Failed to load templates
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

  const byChannel = (ch: MessageTemplate["channel"]) =>
    data.templates.filter((t) => t.channel === ch);

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Notification templates"
        description="Email, SMS, push and in-app message templates"
      />

      <Tabs defaultValue="all">
        <TabsList>
          <TabsTrigger value="all">All</TabsTrigger>
          <TabsTrigger value="email">Email</TabsTrigger>
          <TabsTrigger value="sms">SMS</TabsTrigger>
          <TabsTrigger value="push">Push</TabsTrigger>
          <TabsTrigger value="in_app">In-app</TabsTrigger>
        </TabsList>
        <TabsContent value="all" className="mt-[var(--nx-space-3)]">
          <DataGrid
            columns={templateCols}
            data={data.templates}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        {(["email", "sms", "push", "in_app"] as const).map((ch) => (
          <TabsContent
            key={ch}
            value={ch}
            className="mt-[var(--nx-space-3)]"
          >
            <DataGrid
              columns={templateCols}
              data={byChannel(ch)}
              getRowId={(r) => r.id}
            />
          </TabsContent>
        ))}
      </Tabs>
    </div>
  );
}
