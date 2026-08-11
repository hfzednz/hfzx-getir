"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Button,
  DataGrid,
  FilterBar,
  Input,
  PageHeader,
  Select,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { CUSTOMER_SEGMENT_OPTIONS } from "../api";
import { useCustomersList } from "../hooks";
import type { CustomerListItem, CustomerSegment } from "../types";

function segmentTone(
  segment: CustomerSegment,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (segment) {
    case "vip":
      return "info";
    case "loyal":
      return "success";
    case "fraud_watch":
      return "danger";
    case "churn_risk":
      return "warning";
    case "high_aov":
      return "info";
    default:
      return "neutral";
  }
}

function riskTone(score: number): "success" | "warning" | "danger" {
  if (score >= 60) return "danger";
  if (score >= 30) return "warning";
  return "success";
}

export function CustomersListView() {
  const router = useRouter();
  const [q, setQ] = useState("");
  const [segment, setSegment] = useState<CustomerSegment | "all">("all");

  const filters = useMemo(
    () => ({
      q: q || undefined,
      segment,
      page: 1,
      pageSize: 50,
    }),
    [q, segment],
  );

  const { data, isLoading, isError, error, refetch, isFetching } =
    useCustomersList(filters);

  const columns: DataGridColumnDef<CustomerListItem>[] = [
    {
      id: "customer",
      header: "Customer",
      cell: ({ row }) => (
        <div>
          <div className="font-semibold text-[13px]">{row.name}</div>
          <div className="text-[11px] text-[var(--nx-text-tertiary)] tabular-nums">
            {row.id}
          </div>
        </div>
      ),
    },
    {
      id: "contact",
      header: "Contact",
      cell: ({ row }) => (
        <div>
          <div className="text-[13px]">{row.email}</div>
          <div className="text-[11px] text-[var(--nx-text-tertiary)]">
            {row.phone}
          </div>
        </div>
      ),
    },
    {
      id: "segment",
      header: "Segment",
      cell: ({ row }) => (
        <StatusBadge
          status={row.segment.replace("_", " ")}
          tone={segmentTone(row.segment)}
        />
      ),
    },
    {
      id: "orders",
      header: "Orders",
      align: "right",
      accessorKey: "orderCount",
    },
    {
      id: "ltv",
      header: "LTV",
      align: "right",
      cell: ({ row }) => (
        <span className="tabular-nums">
          {formatMinorUnits(row.lifetimeValueMinor, row.currency)}
        </span>
      ),
    },
    {
      id: "loyalty",
      header: "Loyalty",
      accessorKey: "loyaltyTier",
    },
    {
      id: "wallet",
      header: "Wallet",
      align: "right",
      cell: ({ row }) => (
        <span className="tabular-nums">
          {formatMinorUnits(row.walletBalanceMinor, row.currency)}
        </span>
      ),
    },
    {
      id: "risk",
      header: "Risk / Fraud",
      cell: ({ row }) => (
        <div className="flex gap-[var(--nx-space-1)]">
          <StatusBadge
            status={`R ${row.riskScore}`}
            tone={riskTone(row.riskScore)}
          />
          <StatusBadge
            status={`F ${row.fraudScore}`}
            tone={riskTone(row.fraudScore)}
          />
        </div>
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
          Failed to load customers
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
        title="Customers"
        description={`${data.total} profiles${isFetching ? " · refreshing…" : ""}`}
      />

      <FilterBar
        actions={
          <>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => {
                setQ("");
                setSegment("all");
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
          placeholder="Search name, email, phone, id…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search customers"
        />
        <Select
          value={segment}
          onChange={(e) =>
            setSegment(e.target.value as CustomerSegment | "all")
          }
          aria-label="Segment"
        >
          {CUSTOMER_SEGMENT_OPTIONS.map((s) => (
            <option key={s} value={s}>
              {s === "all" ? "All segments" : s.replace("_", " ")}
            </option>
          ))}
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={data.items}
        getRowId={(row) => row.id}
        emptyMessage="No customers match filters"
        onRowClick={(row) => router.push(`/customers/${row.id}`)}
      />
    </div>
  );
}
