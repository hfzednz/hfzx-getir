"use client";

import {
  DataGrid,
  PageHeader,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useRolesSnapshot } from "../hooks";
import type {
  ApprovalChain,
  PermissionTemplate,
  PlatformRoleTemplate,
  RoleInheritanceEdge,
  TemporaryPermission,
} from "../types";

const roleCols: DataGridColumnDef<PlatformRoleTemplate>[] = [
  { id: "key", header: "Key", accessorKey: "key" },
  { id: "label", header: "Label", accessorKey: "label" },
  {
    id: "scope",
    header: "Scope",
    cell: ({ row }) => (
      <StatusBadge
        status={row.scope}
        tone={
          row.scope === "global"
            ? "danger"
            : row.scope === "company"
              ? "warning"
              : "info"
        }
      />
    ),
  },
  {
    id: "members",
    header: "Members",
    accessorKey: "members",
    align: "right",
  },
  {
    id: "inherits",
    header: "Inherits",
    cell: ({ row }) => row.inheritsFrom ?? "—",
  },
  { id: "desc", header: "Description", accessorKey: "description" },
];

const tmplCols: DataGridColumnDef<PermissionTemplate>[] = [
  { id: "key", header: "Template", accessorKey: "key" },
  { id: "label", header: "Label", accessorKey: "label" },
  {
    id: "perms",
    header: "Permissions",
    cell: ({ row }) => row.permissions.join(", "),
  },
];

const chainCols: DataGridColumnDef<ApprovalChain>[] = [
  { id: "name", header: "Chain", accessorKey: "name" },
  {
    id: "steps",
    header: "Steps",
    cell: ({ row }) => row.steps.join(" → "),
  },
  {
    id: "pending",
    header: "Pending",
    accessorKey: "pending",
    align: "right",
  },
  { id: "desc", header: "Description", accessorKey: "description" },
];

const inheritCols: DataGridColumnDef<RoleInheritanceEdge>[] = [
  { id: "child", header: "Child role", accessorKey: "childRole" },
  { id: "parent", header: "Parent role", accessorKey: "parentRole" },
];

const tempCols: DataGridColumnDef<TemporaryPermission>[] = [
  { id: "subject", header: "Subject", accessorKey: "subject" },
  { id: "perm", header: "Permission", accessorKey: "permission" },
  { id: "scope", header: "Scope", accessorKey: "scope" },
  {
    id: "exp",
    header: "Expires",
    cell: ({ row }) => new Date(row.expiresAt).toLocaleString("en-US"),
  },
  { id: "reason", header: "Reason", accessorKey: "reason" },
  {
    id: "status",
    header: "Status",
    cell: ({ row }) => (
      <StatusBadge
        status={row.status}
        tone={
          row.status === "active"
            ? "success"
            : row.status === "expired"
              ? "warning"
              : "danger"
        }
      />
    ),
  },
];

export function RolesView() {
  const { data, isLoading, isError, error, refetch } = useRolesSnapshot();

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={300} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load roles
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
        title="Roles"
        description={`Global / company / department templates · updated ${new Date(data.generatedAt).toLocaleTimeString("en-US")}`}
      />

      <Tabs defaultValue="roles">
        <TabsList>
          <TabsTrigger value="roles">Roles</TabsTrigger>
          <TabsTrigger value="templates">Permission templates</TabsTrigger>
          <TabsTrigger value="chains">Approval chains</TabsTrigger>
          <TabsTrigger value="inheritance">Inheritance</TabsTrigger>
          <TabsTrigger value="temporary">Temporary permissions</TabsTrigger>
        </TabsList>

        <TabsContent value="roles">
          <DataGrid columns={roleCols} data={data.roles} getRowId={(r) => r.id} />
        </TabsContent>
        <TabsContent value="templates">
          <DataGrid
            columns={tmplCols}
            data={data.permissionTemplates}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="chains">
          <DataGrid
            columns={chainCols}
            data={data.approvalChains}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="inheritance">
          <DataGrid
            columns={inheritCols}
            data={data.inheritance}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="temporary">
          <DataGrid
            columns={tempCols}
            data={data.temporaryPermissions}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
