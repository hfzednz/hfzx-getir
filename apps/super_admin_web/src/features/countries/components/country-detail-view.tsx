"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import {
  Button,
  DataGrid,
  PageHeader,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useCountry } from "../hooks";
import type {
  CountryTaxRule,
  Holiday,
  LegalRule,
  RegionalPricingHook,
} from "../types";

function Panel({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
      <h3 className="m-0 mb-[var(--nx-space-3)] text-[var(--nx-font-size-title)] font-semibold">
        {title}
      </h3>
      {children}
    </section>
  );
}

const taxCols: DataGridColumnDef<CountryTaxRule>[] = [
  { id: "name", header: "Tax", accessorKey: "name" },
  {
    id: "rate",
    header: "Rate",
    cell: ({ row }) => `${row.ratePct}%`,
    align: "right",
  },
  { id: "applies", header: "Applies to", accessorKey: "appliesTo" },
];

const legalCols: DataGridColumnDef<LegalRule>[] = [
  { id: "fw", header: "Framework", accessorKey: "framework" },
  { id: "sum", header: "Summary", accessorKey: "summary" },
  {
    id: "st",
    header: "Status",
    cell: ({ row }) => (
      <StatusBadge
        status={row.status}
        tone={row.status === "active" ? "success" : "warning"}
      />
    ),
  },
];

const holidayCols: DataGridColumnDef<Holiday>[] = [
  { id: "name", header: "Holiday", accessorKey: "name" },
  { id: "date", header: "Date", accessorKey: "date" },
  {
    id: "del",
    header: "Affects delivery",
    cell: ({ row }) => (row.affectsDelivery ? "Yes" : "No"),
  },
];

const hookCols: DataGridColumnDef<RegionalPricingHook>[] = [
  { id: "key", header: "Hook", accessorKey: "key" },
  { id: "desc", header: "Description", accessorKey: "description" },
  {
    id: "en",
    header: "Enabled",
    cell: ({ row }) => (
      <StatusBadge
        status={row.enabled ? "on" : "off"}
        tone={row.enabled ? "success" : "neutral"}
      />
    ),
  },
];

export function CountryDetailView({ countryId }: { countryId: string }) {
  const { data, isLoading, isError, error, refetch } = useCountry(countryId);

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={240} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load country
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
        title={`${data.name} (${data.code})`}
        description={`${data.defaultLocale} · ${data.defaultCurrency} · ${data.regionCount} regions / ${data.cityCount} cities`}
        actions={
          <Link href="/countries">
            <Button size="sm" variant="ghost">
              Back
            </Button>
          </Link>
        }
      />

      <Tabs defaultValue="locale">
        <TabsList>
          <TabsTrigger value="locale">Languages / FX / TZ</TabsTrigger>
          <TabsTrigger value="tax">Taxes</TabsTrigger>
          <TabsTrigger value="delivery">Delivery rules</TabsTrigger>
          <TabsTrigger value="legal">Legal</TabsTrigger>
          <TabsTrigger value="holidays">Holidays</TabsTrigger>
          <TabsTrigger value="pricing">Regional pricing</TabsTrigger>
          <TabsTrigger value="geo">Regions / cities</TabsTrigger>
        </TabsList>

        <TabsContent value="locale">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-[var(--nx-space-3)]">
            <Panel title="Languages">
              <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
                {data.languages.map((l) => (
                  <li key={l.code} className="text-[13px] flex justify-between">
                    <span>
                      {l.label} ({l.code})
                    </span>
                    {l.primary ? (
                      <StatusBadge status="primary" tone="info" />
                    ) : null}
                  </li>
                ))}
              </ul>
            </Panel>
            <Panel title="Currencies">
              <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
                {data.currencies.map((c) => (
                  <li key={c.code} className="text-[13px] flex justify-between">
                    <span>
                      {c.code} · {c.label}
                    </span>
                    {c.primary ? (
                      <StatusBadge status="primary" tone="success" />
                    ) : null}
                  </li>
                ))}
              </ul>
            </Panel>
            <Panel title="Timezones">
              <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
                {data.timezones.map((t) => (
                  <li key={t.id} className="text-[13px]">
                    {t.label}{" "}
                    <span className="text-[var(--nx-text-tertiary)]">
                      ({t.offset})
                    </span>
                  </li>
                ))}
              </ul>
            </Panel>
          </div>
        </TabsContent>

        <TabsContent value="tax">
          <DataGrid columns={taxCols} data={data.taxes} getRowId={(r) => r.id} />
        </TabsContent>

        <TabsContent value="delivery">
          <Panel title="Delivery rules summary">
            <dl className="m-0 grid grid-cols-1 md:grid-cols-3 gap-[var(--nx-space-3)] text-[13px]">
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Max radius</dt>
                <dd className="m-0 font-medium">
                  {data.deliveryRules.maxRadiusKm} km
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Default SLA</dt>
                <dd className="m-0 font-medium">
                  {data.deliveryRules.defaultSlaMin} min
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Night surcharge</dt>
                <dd className="m-0 font-medium">
                  {data.deliveryRules.nightSurchargePct}%
                </dd>
              </div>
            </dl>
            <p className="m-0 mt-[var(--nx-space-3)] text-[12px] text-[var(--nx-text-secondary)]">
              {data.deliveryRules.zoneEditorNote}
            </p>
          </Panel>
        </TabsContent>

        <TabsContent value="legal">
          <DataGrid
            columns={legalCols}
            data={data.legalRules}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="holidays">
          <DataGrid
            columns={holidayCols}
            data={data.holidays}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="pricing">
          <DataGrid
            columns={hookCols}
            data={data.regionalPricingHooks}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="geo">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
            {data.regions.map((region) => (
              <Panel key={region.id} title={region.name}>
                <ul className="m-0 p-0 list-none flex flex-col gap-[var(--nx-space-2)]">
                  {region.cities.map((city) => (
                    <li
                      key={city.id}
                      className="flex items-center justify-between gap-[var(--nx-space-2)] text-[13px] border-b border-[var(--nx-border-subtle)] last:border-0 pb-[var(--nx-space-2)] last:pb-0"
                    >
                      <span className="font-medium">{city.name}</span>
                      <div className="flex items-center gap-[var(--nx-space-2)]">
                        <span className="text-[var(--nx-text-tertiary)] tabular-nums">
                          {city.warehouseCount} WH
                        </span>
                        <StatusBadge
                          status={city.status}
                          tone={
                            city.status === "active" ? "success" : "warning"
                          }
                        />
                      </div>
                    </li>
                  ))}
                </ul>
              </Panel>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
