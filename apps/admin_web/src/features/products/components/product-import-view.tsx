"use client";

import { useRef, useState } from "react";
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
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/permissions";
import { useProductImportPreview } from "../hooks";
import type { ProductImportPreviewRow } from "../types";

const previewCols: DataGridColumnDef<ProductImportPreviewRow>[] = [
  { id: "row", header: "Row", accessorKey: "row", width: 60, align: "right" },
  { id: "sku", header: "SKU", accessorKey: "sku", width: 130 },
  { id: "name", header: "Name", accessorKey: "name" },
  { id: "brand", header: "Brand", accessorKey: "brand", width: 100 },
  { id: "cat", header: "Category", accessorKey: "category", width: 100 },
  { id: "price", header: "Price", accessorKey: "price", align: "right", width: 80 },
  {
    id: "status",
    header: "Status",
    cell: ({ row }) => (
      <StatusBadge
        status={row.status}
        tone={
          row.status === "ok"
            ? "success"
            : row.status === "warning"
              ? "warning"
              : "danger"
        }
      />
    ),
    width: 100,
  },
  { id: "msg", header: "Message", accessorKey: "message" },
];

export function ProductImportView() {
  const session = useAuthStore((s) => s.session);
  const fileRef = useRef<HTMLInputElement>(null);
  const [fileName, setFileName] = useState<string | null>(null);
  const importPreview = useProductImportPreview();

  const allowed = can(session, "catalog:write");

  function onFileChange(file: File | undefined) {
    if (!file) return;
    setFileName(file.name);
    importPreview.mutate(file.name);
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Product CSV import"
        description={
          <span>
            <Link href="/products" className="text-[var(--nx-text-link)]">
              Products
            </Link>
            {" · "}
            Bulk create / update from CSV
          </span>
        }
      />

      <PermissionGate
        allowed={allowed}
        fallback={
          <p className="m-0 text-[13px] text-[var(--nx-text-secondary)]">
            You need catalog:write to import products.
          </p>
        }
      >
        <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)] flex flex-col gap-[var(--nx-space-3)]">
          <p className="m-0 text-[13px] text-[var(--nx-text-secondary)]">
            Expected columns: sku, name, brand, category, price, status. Mock
            preview runs when BFF is offline.
          </p>
          <div className="flex flex-wrap items-center gap-[var(--nx-space-2)]">
            <input
              ref={fileRef}
              type="file"
              accept=".csv,text/csv"
              className="sr-only"
              onChange={(e) => onFileChange(e.target.files?.[0])}
            />
            <Button
              size="sm"
              variant="secondary"
              onClick={() => fileRef.current?.click()}
            >
              Choose CSV
            </Button>
            <Button
              size="sm"
              onClick={() => {
                setFileName("catalog-sample.csv");
                importPreview.mutate("catalog-sample.csv");
              }}
              loading={importPreview.isPending}
            >
              Preview sample
            </Button>
            {fileName ? (
              <span className="text-[12px] text-[var(--nx-text-tertiary)]">
                {fileName}
              </span>
            ) : null}
          </div>
        </section>

        {importPreview.isPending ? <Skeleton height={200} /> : null}

        {importPreview.isError ? (
          <p className="m-0 text-[var(--nx-danger)] text-[13px]">
            {importPreview.error instanceof Error
              ? importPreview.error.message
              : "Import preview failed"}
          </p>
        ) : null}

        {importPreview.data ? (
          <section className="flex flex-col gap-[var(--nx-space-3)]">
            <div className="flex flex-wrap gap-[var(--nx-space-3)] text-[13px]">
              <span>
                OK: <strong>{importPreview.data.okCount}</strong>
              </span>
              <span>
                Warnings: <strong>{importPreview.data.warningCount}</strong>
              </span>
              <span>
                Errors: <strong>{importPreview.data.errorCount}</strong>
              </span>
            </div>
            <DataGrid
              columns={previewCols}
              data={importPreview.data.rows}
              getRowId={(r) => String(r.row)}
            />
            <div className="flex gap-[var(--nx-space-2)]">
              <Button
                size="sm"
                disabled={importPreview.data.errorCount > 0}
              >
                Commit import
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => importPreview.reset()}
              >
                Clear
              </Button>
            </div>
          </section>
        ) : null}
      </PermissionGate>
    </div>
  );
}
