"use client";

import { useMemo, useState } from "react";
import {
  DataGrid,
  FilterBar,
  Input,
  PageHeader,
  Select,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useOrgSnapshot } from "../hooks";
import type { OrgUnit, PlatformPerson, PlatformPersonKind } from "../types";

const unitCols: DataGridColumnDef<OrgUnit>[] = [
  { id: "name", header: "Name", accessorKey: "name" },
  { id: "type", header: "Type", accessorKey: "type", width: 120 },
  {
    id: "hc",
    header: "Headcount",
    accessorKey: "headcount",
    align: "right",
    width: 100,
  },
  {
    id: "parent",
    header: "Parent",
    cell: ({ row }) => row.parentId ?? "—",
  },
];

function kindTone(
  kind: PlatformPersonKind,
): "info" | "success" | "warning" | "danger" | "neutral" {
  switch (kind) {
    case "platform_admin":
      return "danger";
    case "manager":
      return "warning";
    case "auditor":
      return "info";
    case "partner":
    case "supplier":
      return "success";
    default:
      return "neutral";
  }
}

export function OrgView() {
  const { data, isLoading, isError, error, refetch } = useOrgSnapshot();
  const [q, setQ] = useState("");
  const [kind, setKind] = useState("all");

  const people = useMemo(() => {
    const items = data?.people ?? [];
    return items.filter((p) => {
      if (kind !== "all" && p.kind !== kind) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        p.name.toLowerCase().includes(needle) ||
        p.email.toLowerCase().includes(needle)
      );
    });
  }, [data?.people, q, kind]);

  const peopleCols = useMemo<DataGridColumnDef<PlatformPerson>[]>(
    () => [
      { id: "name", header: "Name", accessorKey: "name" },
      { id: "email", header: "Email", accessorKey: "email" },
      {
        id: "kind",
        header: "Kind",
        cell: ({ row }) => (
          <StatusBadge status={row.kind} tone={kindTone(row.kind)} />
        ),
      },
      {
        id: "org",
        header: "Org unit",
        cell: ({ row }) => row.orgUnitName ?? "—",
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "active"
                ? "success"
                : row.status === "invited"
                  ? "warning"
                  : "neutral"
            }
          />
        ),
      },
    ],
    [],
  );

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
          Failed to load organization
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
        title="Organization"
        description={`Platform orgs, departments, teams & people · updated ${new Date(data.generatedAt).toLocaleTimeString("en-US")}`}
      />

      <Tabs defaultValue="people">
        <TabsList>
          <TabsTrigger value="people">People</TabsTrigger>
          <TabsTrigger value="orgs">Organizations</TabsTrigger>
          <TabsTrigger value="depts">Departments</TabsTrigger>
          <TabsTrigger value="teams">Teams</TabsTrigger>
        </TabsList>

        <TabsContent value="people">
          <FilterBar>
            <Input
              placeholder="Search name or email…"
              value={q}
              onChange={(e) => setQ(e.target.value)}
              aria-label="Search people"
            />
            <Select
              value={kind}
              onChange={(e) => setKind(e.target.value)}
              aria-label="Filter kind"
            >
              <option value="all">All kinds</option>
              <option value="platform_admin">Platform admins</option>
              <option value="employee">Employees</option>
              <option value="manager">Managers</option>
              <option value="external_user">External users</option>
              <option value="partner">Partners</option>
              <option value="supplier">Suppliers</option>
              <option value="auditor">Auditors</option>
            </Select>
          </FilterBar>
          <DataGrid
            columns={peopleCols}
            data={people}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="orgs">
          <DataGrid
            columns={unitCols}
            data={data.organizations}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="depts">
          <DataGrid
            columns={unitCols}
            data={data.departments}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="teams">
          <DataGrid
            columns={unitCols}
            data={data.teams}
            getRowId={(r) => r.id}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
