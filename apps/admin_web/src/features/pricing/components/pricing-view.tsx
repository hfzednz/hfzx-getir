"use client";

import { useMemo, useState } from "react";
import {
  Button,
  DataGrid,
  type DataGridColumnDef,
  FilterBar,
  Input,
  PageHeader,
  PermissionGate,
  Select,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { usePermission } from "@/shared/permissions/use-permission";
import { PRICING_KINDS } from "../api";
import {
  useCreatePricingRule,
  usePricingRules,
  useUpdatePricingStatus,
} from "../hooks";
import type { PricingKind, PricingRule, PricingRuleStatus } from "../types";

export function PricingView() {
  const canWrite = usePermission("pricing:write");
  const [kind, setKind] = useState<PricingKind | "all">("all");
  const [status, setStatus] = useState<PricingRuleStatus | "all">("all");
  const [q, setQ] = useState("");
  const [tab, setTab] = useState<PricingKind | "all">("all");

  const { data, isLoading, isError, error, refetch } = usePricingRules({
    kind: tab === "all" ? kind : tab,
    status,
    q,
  });
  const createMut = useCreatePricingRule();
  const statusMut = useUpdatePricingStatus();

  const columns = useMemo<
    DataGridColumnDef<PricingRule & Record<string, unknown>>[]
  >(
    () => [
      {
        id: "name",
        header: "Rule",
        cell: ({ row }) => (
          <div className="flex flex-col gap-0.5">
            <span className="font-semibold text-[13px]">{row.name}</span>
            <span className="text-[11px] text-[var(--nx-text-tertiary)]">
              {row.kind.replaceAll("_", " ")}
              {row.skuId ? ` · ${row.skuId}` : ""}
              {row.warehouseId ? ` · ${row.warehouseId}` : ""}
            </span>
          </div>
        ),
      },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => <StatusBadge status={row.status} />,
      },
      {
        id: "price",
        header: "Price / adj",
        align: "right",
        cell: ({ row }) => (
          <span className="tabular-nums text-[12px]">
            {row.overridePriceMinor != null
              ? formatMinorUnits(row.overridePriceMinor, row.currency)
              : row.adjustmentPct != null
                ? `${row.adjustmentPct > 0 ? "+" : ""}${row.adjustmentPct}%`
                : formatMinorUnits(row.basePriceMinor, row.currency)}
          </span>
        ),
      },
      {
        id: "scope",
        header: "Scope",
        cell: ({ row }) => (
          <span className="text-[12px] text-[var(--nx-text-secondary)]">
            {[row.cityId, row.categoryId, row.competitorRef]
              .filter(Boolean)
              .join(" · ") || "national"}
          </span>
        ),
      },
      {
        id: "ai",
        header: "AI",
        align: "right",
        cell: ({ row }) =>
          row.aiConfidence != null ? (
            <span className="tabular-nums text-[12px]">
              {(row.aiConfidence * 100).toFixed(0)}%
            </span>
          ) : (
            "—"
          ),
      },
      {
        id: "actions",
        header: "",
        cell: ({ row }) => (
          <PermissionGate allowed={canWrite}>
            <div className="flex gap-1">
              {row.status !== "active" ? (
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() =>
                    void statusMut.mutateAsync({ id: row.id, status: "active" })
                  }
                >
                  Activate
                </Button>
              ) : (
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() =>
                    void statusMut.mutateAsync({ id: row.id, status: "paused" })
                  }
                >
                  Pause
                </Button>
              )}
            </div>
          </PermissionGate>
        ),
      },
    ],
    [canWrite, statusMut],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={320} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load pricing
        </p>
        <p className="m-0 mt-1 text-[var(--nx-text-secondary)]">
          {error instanceof Error ? error.message : "Unknown error"}
        </p>
        <button
          type="button"
          onClick={() => void refetch()}
          className="mt-3 text-[var(--nx-text-link)] underline cursor-pointer bg-transparent border-0"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Pricing"
        description="Base, regional, warehouse, dynamic, competitor, scheduled, emergency, and AI-assisted rules"
        actions={
          <PermissionGate allowed={canWrite}>
            <Button
              size="sm"
              loading={createMut.isPending}
              onClick={() =>
                void createMut.mutateAsync({
                  name: "New draft rule",
                  kind: tab === "all" ? "base" : tab,
                  basePriceMinor: 1000,
                  notes: "Created from admin UI",
                })
              }
            >
              New rule
            </Button>
          </PermissionGate>
        }
      />

      <Tabs
        value={tab}
        onValueChange={(v) => setTab(v as PricingKind | "all")}
      >
        <TabsList>
          <TabsTrigger value="all">All</TabsTrigger>
          {PRICING_KINDS.map((k) => (
            <TabsTrigger key={k} value={k}>
              {k.replaceAll("_", " ")}
            </TabsTrigger>
          ))}
        </TabsList>
        <TabsContent value={tab}>
          <FilterBar>
            <Input
              placeholder="Search rules…"
              value={q}
              onChange={(e) => setQ(e.target.value)}
            />
            <Select
              value={status}
              onChange={(e) =>
                setStatus(e.target.value as PricingRuleStatus | "all")
              }
            >
              <option value="all">All statuses</option>
              <option value="draft">draft</option>
              <option value="active">active</option>
              <option value="scheduled">scheduled</option>
              <option value="paused">paused</option>
              <option value="expired">expired</option>
            </Select>
            {tab === "all" ? (
              <Select
                value={kind}
                onChange={(e) =>
                  setKind(e.target.value as PricingKind | "all")
                }
              >
                <option value="all">All kinds</option>
                {PRICING_KINDS.map((k) => (
                  <option key={k} value={k}>
                    {k}
                  </option>
                ))}
              </Select>
            ) : null}
          </FilterBar>
          <div className="mt-[var(--nx-space-3)]">
            <DataGrid
              columns={columns}
              data={data.items as (PricingRule & Record<string, unknown>)[]}
              getRowId={(row) => row.id}
              emptyMessage="No pricing rules"
            />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
