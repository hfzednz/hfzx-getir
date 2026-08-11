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
import { useCountries } from "../hooks";
import type { CountryListItem } from "../types";

export function CountriesView() {
  const router = useRouter();
  const { data, isLoading, isError, error, refetch, isFetching } =
    useCountries();
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("all");

  const filtered = useMemo(() => {
    const items = data?.items ?? [];
    return items.filter((c) => {
      if (status !== "all" && c.status !== status) return false;
      if (!q.trim()) return true;
      const needle = q.trim().toLowerCase();
      return (
        c.name.toLowerCase().includes(needle) ||
        c.code.toLowerCase().includes(needle)
      );
    });
  }, [data?.items, q, status]);

  const columns = useMemo<DataGridColumnDef<CountryListItem>[]>(
    () => [
      { id: "code", header: "Code", accessorKey: "code", width: 70 },
      { id: "name", header: "Country", accessorKey: "name" },
      {
        id: "status",
        header: "Status",
        cell: ({ row }) => (
          <StatusBadge
            status={row.status}
            tone={
              row.status === "active"
                ? "success"
                : row.status === "pilot"
                  ? "warning"
                  : "neutral"
            }
          />
        ),
        width: 100,
      },
      {
        id: "regions",
        header: "Regions",
        accessorKey: "regionCount",
        align: "right",
        width: 80,
      },
      {
        id: "cities",
        header: "Cities",
        accessorKey: "cityCount",
        align: "right",
        width: 80,
      },
      {
        id: "ccy",
        header: "Currency",
        accessorKey: "defaultCurrency",
        width: 90,
      },
      {
        id: "locale",
        header: "Locale",
        accessorKey: "defaultLocale",
        width: 100,
      },
    ],
    [],
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
          Failed to load countries
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
        title="Countries"
        description={`${data.total} markets · languages, tax, legal, regions (not city-ops zones)${isFetching ? " · refreshing…" : ""}`}
      />

      <FilterBar
        actions={
          <Button size="sm" variant="ghost" onClick={() => void refetch()}>
            Refresh
          </Button>
        }
      >
        <Input
          placeholder="Search country…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Search countries"
        />
        <Select
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          aria-label="Filter status"
        >
          <option value="all">All</option>
          <option value="active">Active</option>
          <option value="pilot">Pilot</option>
          <option value="disabled">Disabled</option>
        </Select>
      </FilterBar>

      <DataGrid
        columns={columns}
        data={filtered}
        getRowId={(r) => r.id}
        onRowClick={(row) => router.push(`/countries/${row.id}`)}
      />
    </div>
  );
}
