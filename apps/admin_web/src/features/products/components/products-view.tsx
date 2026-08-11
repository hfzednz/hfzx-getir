"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  BulkActionBar,
  Button,
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
import { formatMinorUnits } from "@/shared/lib/money";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/permissions";
import { useProducts } from "../hooks";
import type { ProductListItem, ProductStatus } from "../types";

function statusTone(
  status: ProductStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "active":
      return "success";
    case "draft":
      return "neutral";
    case "out_of_stock":
      return "warning";
    default:
      return "info";
  }
}

export function ProductsView() {
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch, isFetching } = useProducts();
  const [q, setQ] = useState("");
  const [brand, setBrand] = useState("all");
  const [category, setCategory] = useState("all");
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const filtered = useMemo(() => {
    const items = data?.items ?? [];
    return items.filter((p) => {
      if (brand !== "all" && p.brand !== brand) return false;
      if (category !== "all" && p.category !== category) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        p.name.toLowerCase().includes(needle) ||
        p.sku.toLowerCase().includes(needle) ||
        p.brand.toLowerCase().includes(needle)
      );
    });
  }, [data?.items, q, brand, category]);

  const columns = useMemo<DataGridColumnDef<ProductListItem>[]>(
    () => [
      {
        id: "sel",
        header: "",
        width: 40,
        cell: ({ row }) => (
          <input
            type="checkbox"
            checked={selected.has(row.id)}
            onChange={(e) => {
              e.stopPropagation();
              setSelected((prev) => {
                const next = new Set(prev);
                if (next.has(row.id)) next.delete(row.id);
                else next.add(row.id);
                return next;
              });
            }}
            onClick={(e) => e.stopPropagation()}
            aria-label={`Select ${row.sku}`}
          />
        ),
      },
      { id: "sku", header: "SKU", accessorKey: "sku", width: 130 },
      { id: "name", header: "Product", accessorKey: "name" },
      { id: "brand", header: "Brand", accessorKey: "brand", width: 100 },
      { id: "cat", header: "Category", accessorKey: "category", width: 100 },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge status={row.status} tone={statusTone(row.status)} />
        ),
        width: 120,
      },
      {
        id: "price",
        header: "Price",
        cell: ({ row }) => formatMinorUnits(row.price, row.currency),
        align: "right",
        width: 100,
      },
      {
        id: "variants",
        header: "Variants",
        accessorKey: "variantCount",
        align: "right",
        width: 80,
      },
      {
        id: "inv",
        header: "Inventory",
        cell: ({ row }) => (row.inventoryLinked ? "Linked" : "—"),
        width: 90,
      },
      {
        id: "bundle",
        header: "Bundle",
        cell: ({ row }) => (row.hasBundle ? "Yes" : "—"),
        width: 70,
      },
    ],
    [selected],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={40} />
        <Skeleton height={320} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load products
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
        title="Products"
        description={`${data.total} catalog items · ${data.brands.length} brands · ${data.categories.length} categories${isFetching ? " · refreshing…" : ""}`}
        actions={
          <div className="flex gap-[var(--nx-space-2)]">
            <PermissionGate allowed={can(session, "catalog:write")}>
              <Link href="/products/import" className="no-underline">
                <Button size="sm" variant="secondary">
                  CSV import
                </Button>
              </Link>
            </PermissionGate>
            <PermissionGate allowed={can(session, "catalog:write")}>
              <Button size="sm">New product</Button>
            </PermissionGate>
          </div>
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
          placeholder="Search SKU, name, brand…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search products"
        />
        <Select
          value={brand}
          onChange={(e) => setBrand(e.target.value)}
          aria-label="Brand"
        >
          <option value="all">All brands</option>
          {data.brands.map((b) => (
            <option key={b} value={b}>
              {b}
            </option>
          ))}
        </Select>
        <Select
          value={category}
          onChange={(e) => setCategory(e.target.value)}
          aria-label="Category"
        >
          <option value="all">All categories</option>
          {data.categories.map((c) => (
            <option key={c} value={c}>
              {c}
            </option>
          ))}
        </Select>
      </FilterBar>

      <PermissionGate allowed={can(session, "catalog:write")}>
        <BulkActionBar
          selectedCount={selected.size}
          onClear={() => setSelected(new Set())}
        >
          <Button size="sm" variant="secondary">
            Activate
          </Button>
          <Button size="sm" variant="secondary">
            Archive
          </Button>
          <Button size="sm" variant="ghost">
            Export
          </Button>
        </BulkActionBar>
      </PermissionGate>

      <DataGrid
        columns={columns}
        data={filtered}
        getRowId={(row) => row.id}
        onRowClick={(row) => router.push(`/products/${row.id}`)}
        emptyMessage="No products match filters"
      />
    </div>
  );
}
