"use client";

import Link from "next/link";
import {
  Button,
  DataGrid,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/permissions";
import { useProductDetail } from "../hooks";
import type {
  ProductBundleItem,
  ProductInventoryLink,
  ProductMedia,
  ProductSupplierMapping,
  ProductVariant,
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

const variantCols: DataGridColumnDef<ProductVariant>[] = [
  { id: "sku", header: "SKU", accessorKey: "sku" },
  { id: "name", header: "Name", accessorKey: "name" },
  {
    id: "attrs",
    header: "Attributes",
    cell: ({ row }) =>
      Object.entries(row.attributes)
        .map(([k, v]) => `${k}=${v}`)
        .join(", "),
  },
  {
    id: "price",
    header: "Price",
    cell: ({ row }) => formatMinorUnits(row.price, row.currency),
    align: "right",
  },
  { id: "barcode", header: "Barcode", accessorKey: "barcode" },
];

const bundleCols: DataGridColumnDef<ProductBundleItem>[] = [
  { id: "sku", header: "SKU", accessorKey: "sku" },
  { id: "name", header: "Component", accessorKey: "name" },
  { id: "qty", header: "Qty", accessorKey: "qty", align: "right", width: 60 },
];

const invCols: DataGridColumnDef<ProductInventoryLink>[] = [
  { id: "wh", header: "Warehouse", accessorKey: "warehouseCode" },
  { id: "oh", header: "On hand", accessorKey: "onHand", align: "right" },
  { id: "res", header: "Reserved", accessorKey: "reserved", align: "right" },
  {
    id: "safe",
    header: "Safety",
    accessorKey: "safetyStock",
    align: "right",
  },
];

const supplierCols: DataGridColumnDef<ProductSupplierMapping>[] = [
  { id: "name", header: "Supplier", accessorKey: "supplierName" },
  { id: "sku", header: "Supplier SKU", accessorKey: "supplierSku" },
  {
    id: "lead",
    header: "Lead days",
    accessorKey: "leadTimeDays",
    align: "right",
  },
];

const mediaCols: DataGridColumnDef<ProductMedia>[] = [
  { id: "alt", header: "Alt", accessorKey: "alt" },
  { id: "url", header: "URL", accessorKey: "url" },
  {
    id: "primary",
    header: "Primary",
    cell: ({ row }) => (row.primary ? "Yes" : "—"),
    width: 80,
  },
];

export function ProductDetailView({ productId }: { productId: string }) {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch } =
    useProductDetail(productId);

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
          Failed to load product
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
        title={`${data.name} · ${data.sku}`}
        description={
          <span>
            <Link href="/products" className="text-[var(--nx-text-link)]">
              Products
            </Link>
            {" · "}
            {data.brand} · {data.category}
          </span>
        }
        actions={
          <PermissionGate allowed={can(session, "catalog:write")}>
            <Button size="sm" variant="secondary">
              Edit product
            </Button>
          </PermissionGate>
        }
      />

      <div className="flex flex-wrap items-center gap-[var(--nx-space-2)]">
        <StatusBadge status={data.status} />
        <span className="text-[13px] text-[var(--nx-text-secondary)]">
          {data.description}
        </span>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <Panel title="Pricing">
          <dl className="m-0 grid grid-cols-2 gap-[var(--nx-space-2)] text-[13px]">
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Base</dt>
              <dd className="m-0 font-medium">
                {formatMinorUnits(data.pricing.basePrice, data.pricing.currency)}
              </dd>
            </div>
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Promo</dt>
              <dd className="m-0 font-medium">
                {data.pricing.promoPrice != null
                  ? formatMinorUnits(
                      data.pricing.promoPrice,
                      data.pricing.currency,
                    )
                  : "—"}
              </dd>
            </div>
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Tax</dt>
              <dd className="m-0 font-medium">{data.pricing.taxRatePct}%</dd>
            </div>
          </dl>
        </Panel>

        <Panel title="Attributes">
          <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-1)] text-[13px]">
            {data.attributes.map((a) => (
              <li key={a.key} className="flex justify-between gap-4 border-b border-[var(--nx-border-subtle)] py-1 last:border-0">
                <span className="text-[var(--nx-text-tertiary)]">{a.key}</span>
                <span className="font-medium">{a.value}</span>
              </li>
            ))}
          </ul>
        </Panel>

        <Panel title="Variants">
          <DataGrid
            columns={variantCols}
            data={data.variants}
            getRowId={(r) => r.id}
            emptyMessage="No variants"
          />
        </Panel>

        <Panel title="Bundles">
          <DataGrid
            columns={bundleCols}
            data={data.bundles}
            getRowId={(r) => r.productId}
            emptyMessage="Not a bundle"
          />
        </Panel>

        <Panel title="Media">
          <DataGrid
            columns={mediaCols}
            data={data.media}
            getRowId={(r) => r.id}
          />
        </Panel>

        <Panel title="Nutrition">
          {data.nutrition ? (
            <dl className="m-0 grid grid-cols-2 gap-[var(--nx-space-2)] text-[13px]">
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Serving</dt>
                <dd className="m-0 font-medium">{data.nutrition.servingSize}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Calories</dt>
                <dd className="m-0 font-medium">{data.nutrition.calories}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Protein</dt>
                <dd className="m-0 font-medium">{data.nutrition.proteinG}g</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Carbs</dt>
                <dd className="m-0 font-medium">{data.nutrition.carbsG}g</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Fat</dt>
                <dd className="m-0 font-medium">{data.nutrition.fatG}g</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Allergens</dt>
                <dd className="m-0 font-medium">
                  {data.nutrition.allergens.join(", ") || "—"}
                </dd>
              </div>
            </dl>
          ) : (
            <p className="m-0 text-[13px] text-[var(--nx-text-secondary)]">
              No nutrition data
            </p>
          )}
        </Panel>

        <Panel title="Inventory link">
          <DataGrid
            columns={invCols}
            data={data.inventoryLinks}
            getRowId={(r) => r.warehouseId}
            emptyMessage="Not linked to inventory"
          />
        </Panel>

        <Panel title="Supplier mapping">
          <DataGrid
            columns={supplierCols}
            data={data.suppliers}
            getRowId={(r) => r.supplierId}
          />
        </Panel>

        <Panel title="SEO">
          <dl className="m-0 flex flex-col gap-[var(--nx-space-2)] text-[13px]">
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Slug</dt>
              <dd className="m-0 font-medium">{data.seo.slug}</dd>
            </div>
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Meta title</dt>
              <dd className="m-0 font-medium">{data.seo.metaTitle}</dd>
            </div>
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Meta description</dt>
              <dd className="m-0 font-medium">{data.seo.metaDescription}</dd>
            </div>
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Keywords</dt>
              <dd className="m-0 font-medium">
                {data.seo.keywords.join(", ")}
              </dd>
            </div>
          </dl>
        </Panel>
      </div>
    </div>
  );
}
