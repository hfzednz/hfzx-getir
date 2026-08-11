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
import { useRbacSnapshot } from "../hooks";
import type {
  ApprovalWorkflow,
  CustomPermission,
  Department,
  RoleRow,
  TemporaryGrant,
} from "../types";

const deptCols: DataGridColumnDef<Department>[] = [
  { id: "name", header: "Department", accessorKey: "name" },
  {
    id: "hc",
    header: "Headcount",
    accessorKey: "headcount",
    align: "right",
  },
  {
    id: "roles",
    header: "Roles",
    cell: ({ row }) => row.roles.join(", "),
  },
];

const roleCols: DataGridColumnDef<RoleRow>[] = [
  { id: "key", header: "Key", accessorKey: "key" },
  { id: "label", header: "Label", accessorKey: "label" },
  {
    id: "members",
    header: "Members",
    accessorKey: "members",
    align: "right",
  },
  { id: "desc", header: "Description", accessorKey: "description" },
];

const customCols: DataGridColumnDef<CustomPermission>[] = [
  { id: "key", header: "Permission", accessorKey: "key" },
  { id: "desc", header: "Description", accessorKey: "description" },
  { id: "by", header: "Created by", accessorKey: "createdBy" },
];

const grantCols: DataGridColumnDef<TemporaryGrant>[] = [
  { id: "user", header: "User", accessorKey: "user" },
  { id: "perm", header: "Permission", accessorKey: "permission" },
  {
    id: "exp",
    header: "Expires",
    accessorKey: "expiresAt",
    cell: ({ value }) =>
      new Date(String(value)).toLocaleString("tr-TR"),
  },
  { id: "reason", header: "Reason", accessorKey: "reason" },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    cell: ({ value }) => {
      const s = String(value);
      return (
        <StatusBadge
          status={s}
          tone={
            s === "active" ? "success" : s === "expired" ? "warning" : "danger"
          }
        />
      );
    },
  },
];

const approvalCols: DataGridColumnDef<ApprovalWorkflow>[] = [
  { id: "name", header: "Workflow", accessorKey: "name" },
  {
    id: "steps",
    header: "Steps",
    accessorKey: "steps",
    align: "right",
  },
  {
    id: "pending",
    header: "Pending",
    accessorKey: "pending",
    align: "right",
  },
  { id: "desc", header: "Description", accessorKey: "description" },
];

export function RbacView() {
  const { data, isLoading, isError, error, refetch } = useRbacSnapshot();

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
          Failed to load RBAC
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
        title="RBAC"
        description="Departments, roles, permissions matrix, grants & approvals"
      />

      <Tabs defaultValue="departments">
        <TabsList>
          <TabsTrigger value="departments">Departments</TabsTrigger>
          <TabsTrigger value="roles">Roles</TabsTrigger>
          <TabsTrigger value="matrix">Permissions matrix</TabsTrigger>
          <TabsTrigger value="custom">Custom permissions</TabsTrigger>
          <TabsTrigger value="grants">Temporary grants</TabsTrigger>
          <TabsTrigger value="approvals">Approval workflows</TabsTrigger>
        </TabsList>

        <TabsContent value="departments" className="mt-[var(--nx-space-3)]">
          <DataGrid
            columns={deptCols}
            data={data.departments}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="roles" className="mt-[var(--nx-space-3)]">
          <DataGrid
            columns={roleCols}
            data={data.roles}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="matrix" className="mt-[var(--nx-space-3)]">
          <div className="nx-data-grid overflow-auto">
            <table className="nx-data-grid__table">
              <thead>
                <tr className="nx-data-grid__thead-row">
                  <th className="nx-data-grid__th nx-data-grid__th--sticky">
                    Permission
                  </th>
                  {data.matrixRoles.map((role) => (
                    <th
                      key={role}
                      className="nx-data-grid__th nx-data-grid__th--center"
                    >
                      {role}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {data.matrixPermissions.map((perm) => (
                  <tr key={perm} className="nx-data-grid__row">
                    <td className="nx-data-grid__td font-[family-name:var(--nx-font-mono)] text-[11px]">
                      {perm}
                    </td>
                    {data.matrixRoles.map((role) => {
                      const cell = data.matrix.find(
                        (m) => m.role === role && m.permission === perm,
                      );
                      return (
                        <td
                          key={`${role}-${perm}`}
                          className="nx-data-grid__td nx-data-grid__td--center"
                        >
                          {cell?.granted ? (
                            <StatusBadge status="✓" tone="success" />
                          ) : (
                            <span className="text-[var(--nx-text-tertiary)]">
                              —
                            </span>
                          )}
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </TabsContent>

        <TabsContent value="custom" className="mt-[var(--nx-space-3)]">
          <DataGrid
            columns={customCols}
            data={data.customPermissions}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="grants" className="mt-[var(--nx-space-3)]">
          <DataGrid
            columns={grantCols}
            data={data.temporaryGrants}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="approvals" className="mt-[var(--nx-space-3)]">
          <DataGrid
            columns={approvalCols}
            data={data.approvals}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
