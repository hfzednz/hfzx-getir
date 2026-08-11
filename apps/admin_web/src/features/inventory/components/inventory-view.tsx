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
import { can } from "@/shared/permissions/permissions";
import { useInventorySnapshot } from "../hooks";
import type {
  InventoryAdjustment,
  InventoryCycleCount,
  InventoryStockRow,
  InventoryTransfer,
} from "../types";

function Panel({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
      <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
        {title}
      </h3>
      {children}
    </section>
  );
}

const transferCols: DataGridColumnDef<InventoryTransfer>[] = [
  { id: "from", header: "From", accessorKey: "fromWarehouse" },
  { id: "to", header: "To", accessorKey: "toWarehouse" },
  { id: "skus", header: "SKUs", accessorKey: "skuCount", align: "right" },
  { id: "status", header: "Status", accessorKey: "status" },
];

const cycleCols: DataGridColumnDef<InventoryCycleCount>[] = [
  { id: "wh", header: "Warehouse", accessorKey: "warehouseCode" },
  { id: "zone", header: "Zone", accessorKey: "zone" },
  {
    id: "var",
    header: "Variance",
    accessorKey: "varianceUnits",
    align: "right",
  },
  { id: "status", header: "Status", accessorKey: "status" },
];

const adjCols: DataGridColumnDef<InventoryAdjustment>[] = [
  { id: "sku", header: "SKU", accessorKey: "sku" },
  { id: "wh", header: "WH", accessorKey: "warehouseCode", width: 80 },
  { id: "delta", header: "Δ", accessorKey: "delta", align: "right", width: 60 },
  { id: "reason", header: "Reason", accessorKey: "reason" },
];

export function InventoryView() {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } =
    useInventorySnapshot();
  const [q, setQ] = useState("");
  const [kind, setKind] = useState("all");

  const filtered = useMemo(() => {
    const items = data?.stock ?? [];
    return items.filter((r) => {
      if (kind !== "all" && r.kind !== kind) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        r.sku.toLowerCase().includes(needle) ||
        r.productName.toLowerCase().includes(needle) ||
        r.warehouseCode.toLowerCase().includes(needle)
      );
    });
  }, [data?.stock, q, kind]);

  const stockCols = useMemo<DataGridColumnDef<InventoryStockRow>[]>(
    () => [
      { id: "sku", header: "SKU", accessorKey: "sku", width: 120 },
      { id: "name", header: "Product", accessorKey: "productName" },
      { id: "wh", header: "WH", accessorKey: "warehouseCode", width: 80 },
      {
        id: "oh",
        header: "On hand",
        accessorKey: "onHand",
        align: "right",
        width: 80,
      },
      {
        id: "res",
        header: "Reserved",
        accessorKey: "reserved",
        align: "right",
        width: 80,
      },
      {
        id: "avl",
        header: "Available",
        accessorKey: "available",
        align: "right",
        width: 90,
      },
      {
        id: "safe",
        header: "Safety",
        cell: ({ row }) => (
          <span
            className="tabular-nums"
            style={{
              color:
                row.onHand < row.safetyStock
                  ? "var(--nx-danger)"
                  : "var(--nx-text-primary)",
            }}
          >
            {row.safetyStock}
          </span>
        ),
        align: "right",
        width: 70,
      },
      {
        id: "fc",
        header: "Forecast 7d",
        accessorKey: "forecast7d",
        align: "right",
        width: 100,
      },
      {
        id: "kind",
        header: "Flag",
        cell: ({ row }) =>
          row.kind === "normal" ? (
            <span className="text-[var(--nx-text-tertiary)]">—</span>
          ) : (
            <StatusBadge
              status={row.kind}
              tone={
                row.kind === "damaged" || row.kind === "expired"
                  ? "danger"
                  : row.kind === "shrinkage"
                    ? "warning"
                    : "info"
              }
            />
          ),
        width: 110,
      },
    ],
    [],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={96} />
        <Skeleton height={320} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load inventory
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

  const t = data.totals;

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Inventory"
        description={`Real-time stock · updated ${new Date(data.generatedAt).toLocaleTimeString("tr-TR")}${isFetching ? " · refreshing…" : ""}`}
        actions={
          <PermissionGate allowed={can(session, "inventory:transfer")}>
            <Button size="sm" variant="secondary">
              New transfer
            </Button>
          </PermissionGate>
        }
      />

      <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        <KpiCard title="SKUs" value={t.skus} />
        <KpiCard title="Units on hand" value={t.unitsOnHand.toLocaleString("tr-TR")} />
        <KpiCard title="Reserved" value={t.reserved} tone="brand" />
        <KpiCard title="Below safety" value={t.belowSafety} tone="danger" />
        <KpiCard title="Damaged lots" value={t.damaged} tone="warning" />
        <KpiCard title="Expired lots" value={t.expired} tone="danger" />
      </section>

      <FilterBar
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      >
        <Input
          placeholder="Search SKU, product, warehouse…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search inventory"
        />
        <Select
          value={kind}
          onChange={(e) => setKind(e.target.value)}
          aria-label="Flag filter"
        >
          <option value="all">All flags</option>
          <option value="normal">Normal</option>
          <option value="reserved">Reserved</option>
          <option value="damaged">Damaged</option>
          <option value="expired">Expired</option>
          <option value="shrinkage">Shrinkage</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={stockCols}
        data={filtered}
        getRowId={(row) => row.id}
        emptyMessage="No stock rows"
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-[var(--nx-space-3)]">
        <Panel title="Transfers">
          <DataGrid
            columns={transferCols}
            data={data.transfers}
            getRowId={(r) => r.id}
          />
        </Panel>
        <Panel title="Cycle counts">
          <DataGrid
            columns={cycleCols}
            data={data.cycleCounts}
            getRowId={(r) => r.id}
          />
        </Panel>
        <Panel title="Adjustments">
          <DataGrid
            columns={adjCols}
            data={data.adjustments}
            getRowId={(r) => r.id}
          />
        </Panel>
      </div>
    </div>
  );
}
