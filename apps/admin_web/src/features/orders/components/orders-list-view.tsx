"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  BulkActionBar,
  Button,
  Checkbox,
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
import { formatMinorUnits } from "@/shared/lib/money";
import {
  canCancelOrder,
  canReassign,
} from "@/shared/permissions/order-rules";
import { ORDER_STATUS_OPTIONS } from "../api";
import { useBulkOrderAction, useOrdersList } from "../hooks";
import type { OrderListItem, OrderStatus } from "../types";

function statusTone(
  status: OrderStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "delivered":
      return "success";
    case "cancelled":
    case "failed":
    case "refunded":
      return "danger";
    case "en_route":
    case "assigned":
      return "info";
    case "picking":
    case "ready":
    case "confirmed":
      return "warning";
    default:
      return "neutral";
  }
}

export function OrdersListView() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const [q, setQ] = useState("");
  const [status, setStatus] = useState<OrderStatus | "all">("all");
  const [warehouseCode, setWarehouseCode] = useState("");
  const [zone, setZone] = useState("");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [bulkCourierId, setBulkCourierId] = useState("");

  const filters = useMemo(
    () => ({
      q: q || undefined,
      status,
      warehouseCode: warehouseCode || undefined,
      zone: zone || undefined,
      page: 1,
      pageSize: 50,
    }),
    [q, status, warehouseCode, zone],
  );

  const { data, isLoading, isError, error, refetch, isFetching } =
    useOrdersList(filters);
  const bulk = useBulkOrderAction();

  const toggle = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const toggleAll = (rows: OrderListItem[]) => {
    setSelected((prev) => {
      if (prev.size === rows.length) return new Set();
      return new Set(rows.map((r) => r.id));
    });
  };

  const columns: DataGridColumnDef<OrderListItem>[] = [
    {
      id: "select",
      header: (
        <Checkbox
          aria-label="Select all"
          checked={
            (data?.items.length ?? 0) > 0 &&
            selected.size === (data?.items.length ?? 0)
          }
          onChange={() => toggleAll(data?.items ?? [])}
          onClick={(e) => e.stopPropagation()}
        />
      ),
      width: 40,
      cell: ({ row }) => (
        <Checkbox
          aria-label={`Select ${row.id}`}
          checked={selected.has(row.id)}
          onChange={() => toggle(row.id)}
          onClick={(e) => e.stopPropagation()}
        />
      ),
    },
    {
      id: "id",
      header: "Order",
      cell: ({ row }) => (
        <div>
          <div className="font-semibold tabular-nums text-[13px]">{row.id}</div>
          <div className="text-[11px] text-[var(--nx-text-tertiary)]">
            {row.externalRef}
          </div>
        </div>
      ),
    },
    {
      id: "status",
      header: "Status",
      cell: ({ row }) => (
        <StatusBadge
          status={row.status.replace("_", " ")}
          tone={statusTone(row.status)}
        />
      ),
    },
    {
      id: "customer",
      header: "Customer",
      cell: ({ row }) => (
        <div>
          <div className="text-[13px]">{row.customerName}</div>
          <div className="text-[11px] text-[var(--nx-text-tertiary)]">
            {row.customerPhone}
          </div>
        </div>
      ),
    },
    {
      id: "warehouse",
      header: "Warehouse",
      accessorKey: "warehouseCode",
    },
    {
      id: "courier",
      header: "Courier",
      cell: ({ row }) => row.courierName ?? "—",
    },
    {
      id: "zone",
      header: "Zone",
      accessorKey: "zone",
    },
    {
      id: "total",
      header: "Total",
      align: "right",
      cell: ({ row }) => (
        <span className="tabular-nums">
          {formatMinorUnits(row.totalMinor, row.currency)}
        </span>
      ),
    },
    {
      id: "delay",
      header: "Delay",
      align: "right",
      cell: ({ row }) =>
        row.delayMinutes > 0 ? (
          <StatusBadge status={`+${row.delayMinutes}m`} tone="warning" />
        ) : (
          "—"
        ),
    },
    {
      id: "updated",
      header: "Updated",
      cell: ({ row }) => (
        <span className="tabular-nums text-[12px] text-[var(--nx-text-secondary)]">
          {new Date(row.updatedAt).toLocaleTimeString("tr-TR")}
        </span>
      ),
    },
  ];

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={56} />
        <Skeleton height={360} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load orders
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
        title="Orders"
        description={`${data.total} orders${isFetching ? " · refreshing…" : ""}`}
      />

      <FilterBar
        actions={
          <>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => {
                setQ("");
                setStatus("all");
                setWarehouseCode("");
                setZone("");
                setSelected(new Set());
              }}
            >
              Reset
            </Button>
            <Button variant="primary" size="sm" onClick={() => void refetch()}>
              Refresh
            </Button>
          </>
        }
      >
        <Input
          placeholder="Search id, name, phone…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search orders"
        />
        <Select
          value={status}
          onChange={(e) => setStatus(e.target.value as OrderStatus | "all")}
          aria-label="Status"
        >
          {ORDER_STATUS_OPTIONS.map((s) => (
            <option key={s} value={s}>
              {s === "all" ? "All statuses" : s.replace("_", " ")}
            </option>
          ))}
        </Select>
        <Select
          value={warehouseCode}
          onChange={(e) => setWarehouseCode(e.target.value)}
          aria-label="Warehouse"
        >
          <option value="">All warehouses</option>
          <option value="WH-07">WH-07</option>
          <option value="WH-14">WH-14</option>
          <option value="WH-03">WH-03</option>
          <option value="WH-11">WH-11</option>
        </Select>
        <Select
          value={zone}
          onChange={(e) => setZone(e.target.value)}
          aria-label="Zone"
        >
          <option value="">All zones</option>
          <option value="Kadıköy">Kadıköy</option>
          <option value="Beşiktaş">Beşiktaş</option>
          <option value="Şişli">Şişli</option>
          <option value="Üsküdar">Üsküdar</option>
          <option value="Bakırköy">Bakırköy</option>
        </Select>
      </FilterBar>

      <BulkActionBar
        selectedCount={selected.size}
        onClear={() => setSelected(new Set())}
      >
        <PermissionGate allowed={canReassign(session)}>
          <Input
            value={bulkCourierId}
            onChange={(e) => setBulkCourierId(e.target.value)}
            placeholder="Courier ID"
            aria-label="Bulk reassign courier ID"
            className="w-40"
          />
          <Button
            variant="secondary"
            size="sm"
            loading={bulk.isPending}
            disabled={!bulkCourierId.trim()}
            onClick={() =>
              void bulk
                .mutateAsync({
                  orderIds: [...selected],
                  action: "reassign",
                  reason: "Bulk reassign from orders list",
                  courierId: bulkCourierId.trim(),
                })
                .then(() => {
                  setSelected(new Set());
                  setBulkCourierId("");
                })
            }
          >
            Reassign
          </Button>
        </PermissionGate>
        <PermissionGate allowed={canCancelOrder(session)}>
          <Button
            variant="danger"
            size="sm"
            loading={bulk.isPending}
            onClick={() =>
              void bulk.mutateAsync({
                orderIds: [...selected],
                action: "cancel",
                reason: "Bulk cancel from orders list",
              }).then(() => setSelected(new Set()))
            }
          >
            Cancel
          </Button>
        </PermissionGate>
      </BulkActionBar>

      <DataGrid
        columns={columns}
        data={data.items}
        getRowId={(row) => row.id}
        emptyMessage="No orders match filters"
        onRowClick={(row) => router.push(`/orders/${row.id}`)}
      />
    </div>
  );
}
