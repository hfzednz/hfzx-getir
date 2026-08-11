"use client";

import Link from "next/link";
import {
  Button,
  DataGrid,
  KpiCard,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/permissions";
import { useWarehouseDetail } from "../hooks";
import type {
  AiOptimizationItem,
  WarehouseAudit,
  WarehouseStockAlert,
  WarehouseTransfer,
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

const transferCols: DataGridColumnDef<WarehouseTransfer>[] = [
  {
    id: "dir",
    header: "Dir",
    cell: ({ row }) => row.direction.toUpperCase(),
    width: 60,
  },
  { id: "cp", header: "Counterpart", accessorKey: "counterpart" },
  {
    id: "skus",
    header: "SKUs",
    accessorKey: "skuCount",
    align: "right",
    width: 70,
  },
  { id: "status", header: "Status", accessorKey: "status" },
  {
    id: "eta",
    header: "ETA",
    cell: ({ row }) => new Date(row.etaAt).toLocaleTimeString("tr-TR"),
  },
];

const auditCols: DataGridColumnDef<WarehouseAudit>[] = [
  { id: "type", header: "Type", accessorKey: "type" },
  { id: "result", header: "Result", accessorKey: "result" },
  { id: "auditor", header: "Auditor", accessorKey: "auditor" },
  {
    id: "at",
    header: "When",
    cell: ({ row }) => new Date(row.auditedAt).toLocaleDateString("tr-TR"),
  },
];

const alertCols: DataGridColumnDef<WarehouseStockAlert>[] = [
  { id: "sku", header: "SKU", accessorKey: "sku", width: 120 },
  { id: "name", header: "Product", accessorKey: "productName" },
  {
    id: "sev",
    header: "Severity",
    cell: ({ row }) => (
      <StatusBadge status={row.severity} tone={row.severity} />
    ),
    width: 100,
  },
  {
    id: "oh",
    header: "On hand",
    cell: ({ row }) => `${row.onHand} / ${row.safetyStock}`,
    align: "right",
  },
  { id: "msg", header: "Message", accessorKey: "message" },
];

export function WarehouseDetailView({ warehouseId }: { warehouseId: string }) {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch } =
    useWarehouseDetail(warehouseId);

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
          Failed to load warehouse
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
        title={`${data.name} · ${data.code}`}
        description={
          <span>
            <Link href="/warehouses" className="text-[var(--nx-text-link)]">
              Warehouses
            </Link>
            {" · "}
            {data.district} · {data.address}
          </span>
        }
        actions={
          <PermissionGate allowed={can(session, "warehouses:write")}>
            <Button size="sm" variant="secondary">
              Edit warehouse
            </Button>
          </PermissionGate>
        }
      />

      <div className="flex items-center gap-[var(--nx-space-2)]">
        <StatusBadge
          status={data.status}
          tone={
            data.status === "open"
              ? "success"
              : data.status === "busy"
                ? "warning"
                : "info"
          }
        />
        <span className="text-[12px] text-[var(--nx-text-secondary)]">
          Manager {data.managerName}
        </span>
      </div>

      <section className="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-8 gap-[var(--nx-space-3)]">
        <KpiCard title="Capacity" value={`${k.capacityPct}%`} tone={k.capacityPct >= 90 ? "danger" : "neutral"} />
        <KpiCard title="SKUs" value={k.skuCount.toLocaleString("tr-TR")} />
        <KpiCard title="Units" value={k.unitsOnHand.toLocaleString("tr-TR")} />
        <KpiCard title="Pick SLA" value={`${k.pickSlaPct.toFixed(1)}%`} tone="warning" />
        <KpiCard title="Pack SLA" value={`${k.packSlaPct.toFixed(1)}%`} tone="success" />
        <KpiCard title="Dispatch SLA" value={`${k.dispatchSlaPct.toFixed(1)}%`} tone="success" />
        <KpiCard title="Avg pick" value={`${k.avgPickMinutes} min`} />
        <KpiCard title="Stock alerts" value={k.stockAlerts} tone="danger" />
      </section>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <Panel title="Inventory summary">
          <DataGrid
            columns={[
              { id: "cat", header: "Category", accessorKey: "category" },
              {
                id: "sku",
                header: "SKUs",
                accessorKey: "skuCount",
                align: "right",
              },
              {
                id: "units",
                header: "Units",
                accessorKey: "units",
                align: "right",
              },
            ]}
            data={data.inventorySummary}
            getRowId={(r) => r.category}
          />
        </Panel>

        <Panel title="AI optimization">
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-3)]">
            {data.aiOptimization.map((item: AiOptimizationItem) => (
              <li key={item.id}>
                <p className="m-0 text-[13px] font-semibold">{item.title}</p>
                <p className="m-0 mt-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
                  {item.summary}
                </p>
                <p className="m-0 mt-[var(--nx-space-1)] text-[11px] text-[var(--nx-text-tertiary)] tabular-nums">
                  {item.impact} · Confidence {(item.confidence * 100).toFixed(0)}%
                </p>
              </li>
            ))}
          </ul>
        </Panel>

        <Panel title="Transfers">
          <DataGrid
            columns={transferCols}
            data={data.transfers}
            getRowId={(r) => r.id}
          />
        </Panel>

        <Panel title="Audits">
          <DataGrid
            columns={auditCols}
            data={data.audits}
            getRowId={(r) => r.id}
          />
        </Panel>

        <Panel title="Stock alerts">
          <DataGrid
            columns={alertCols}
            data={data.stockAlerts}
            getRowId={(r) => r.id}
            emptyMessage="No stock alerts"
          />
        </Panel>

        <Panel title="Picking / packing / dispatch">
          <ul className="m-0 p-0 list-none grid grid-cols-2 gap-[var(--nx-space-2)] text-[13px]">
            <li>
              Avg pick: <strong>{k.avgPickMinutes} min</strong>
            </li>
            <li>
              Avg pack: <strong>{k.avgPackMinutes} min</strong>
            </li>
            <li>
              Pick SLA: <strong>{k.pickSlaPct.toFixed(1)}%</strong>
            </li>
            <li>
              Pack SLA: <strong>{k.packSlaPct.toFixed(1)}%</strong>
            </li>
            <li>
              Dispatch SLA: <strong>{k.dispatchSlaPct.toFixed(1)}%</strong>
            </li>
            <li>
              Open transfers: <strong>{k.openTransfers}</strong>
            </li>
          </ul>
        </Panel>
      </div>
    </div>
  );
}
