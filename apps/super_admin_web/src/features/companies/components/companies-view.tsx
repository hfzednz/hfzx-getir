"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Button,
  ConfirmDialog,
  DataGrid,
  FilterBar,
  Input,
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
  useCompanies,
  useCreateCompany,
  useDeleteCompany,
  useUpdateCompany,
} from "../hooks";
import type { CompanyListItem, CompanyStatus } from "../types";

function statusTone(
  status: CompanyStatus,
): "success" | "warning" | "danger" | "neutral" {
  if (status === "active") return "success";
  if (status === "draft") return "warning";
  return "danger";
}

export function CompaniesView() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useCompanies();
  const createMutation = useCreateCompany();
  const updateMutation = useUpdateCompany();
  const deleteMutation = useDeleteCompany();

  const [q, setQ] = useState("");
  const [status, setStatus] = useState("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [legalName, setLegalName] = useState("");
  const [tradeName, setTradeName] = useState("");
  const [countryCode, setCountryCode] = useState("TR");
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const filtered = useMemo(() => {
    const items = data?.items ?? [];
    return items.filter((c) => {
      if (status !== "all" && c.status !== status) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        c.legalName.toLowerCase().includes(needle) ||
        c.tradeName.toLowerCase().includes(needle) ||
        c.countryCode.toLowerCase().includes(needle)
      );
    });
  }, [data?.items, q, status]);

  const columns = useMemo<DataGridColumnDef<CompanyListItem>[]>(
    () => [
      { id: "legal", header: "Legal name", accessorKey: "legalName" },
      { id: "trade", header: "Trade", accessorKey: "tradeName", width: 100 },
      { id: "cc", header: "Country", accessorKey: "countryCode", width: 80 },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={statusTone(row.status)} />
        ),
        width: 110,
      },
      {
        id: "tenants",
        header: "Tenants",
        accessorKey: "tenantCount",
        align: "right",
        width: 80,
      },
      {
        id: "ccy",
        header: "Currency",
        accessorKey: "primaryCurrency",
        width: 90,
      },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => (
          <div
            className="flex gap-[var(--nx-space-1)]"
            onClick={(e) => e.stopPropagation()}
          >
            <PermissionGate allowed={can(session, "companies:write")}>
              <Button
                size="sm"
                variant="secondary"
                disabled={row.status === "suspended"}
                onClick={() =>
                  void updateMutation.mutateAsync({
                    id: row.id,
                    patch: { status: "suspended" },
                  })
                }
              >
                Suspend
              </Button>
              <Button
                size="sm"
                variant="danger"
                onClick={() => setDeleteId(row.id)}
              >
                Delete
              </Button>
            </PermissionGate>
          </div>
        ),
        width: 200,
      },
    ],
    [session, updateMutation],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={40} />
        <Skeleton height={280} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load companies
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
        title="Companies"
        description={`${data.total} legal entities · business / tax / branding${isFetching ? " · refreshing…" : ""}`}
        actions={
          <PermissionGate allowed={can(session, "companies:write")}>
            <Button size="sm" onClick={() => setCreateOpen(true)}>
              Create company
            </Button>
          </PermissionGate>
        }
      />

      <FilterBar
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      >
        <Input
          placeholder="Search legal or trade name…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search companies"
        />
        <Select
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          aria-label="Filter status"
        >
          <option value="all">All statuses</option>
          <option value="active">Active</option>
          <option value="draft">Draft</option>
          <option value="suspended">Suspended</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={filtered}
        getRowId={(r) => r.id}
        onRowClick={(row) => router.push(`/companies/${row.id}`)}
      />

      <ConfirmDialog
        open={createOpen}
        title="Create company"
        description={
          <div className="flex flex-col gap-[var(--nx-space-2)] mt-[var(--nx-space-2)]">
            <Input
              placeholder="Legal name"
              value={legalName}
              onChange={(e) => setLegalName(e.target.value)}
            />
            <Input
              placeholder="Trade name"
              value={tradeName}
              onChange={(e) => setTradeName(e.target.value)}
            />
            <Select
              value={countryCode}
              onChange={(e) => setCountryCode(e.target.value)}
            >
              <option value="TR">TR</option>
              <option value="DE">DE</option>
              <option value="US">US</option>
              <option value="SG">SG</option>
            </Select>
          </div>
        }
        confirmLabel="Create"
        loading={createMutation.isPending}
        onCancel={() => setCreateOpen(false)}
        onConfirm={() => {
          if (!legalName.trim() || !tradeName.trim()) return;
          void createMutation
            .mutateAsync({
              legalName: legalName.trim(),
              tradeName: tradeName.trim(),
              countryCode,
              primaryCurrency:
                countryCode === "TR"
                  ? "TRY"
                  : countryCode === "DE"
                    ? "EUR"
                    : countryCode === "SG"
                      ? "SGD"
                      : "USD",
            })
            .then(() => {
              setCreateOpen(false);
              setLegalName("");
              setTradeName("");
            });
        }}
      />

      <ConfirmDialog
        open={deleteId != null}
        title="Delete company"
        danger
        description="This removes the legal entity record. Tenants must be reassigned first in production."
        confirmLabel="Delete"
        loading={deleteMutation.isPending}
        onCancel={() => setDeleteId(null)}
        onConfirm={() => {
          if (!deleteId) return;
          void deleteMutation.mutateAsync(deleteId).then(() => setDeleteId(null));
        }}
      />
    </div>
  );
}
