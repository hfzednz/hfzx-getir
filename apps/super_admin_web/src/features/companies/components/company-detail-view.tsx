"use client";

import { useState, type ReactNode } from "react";
import Link from "next/link";
import {
  Button,
  DataGrid,
  Input,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  type DataGridColumnDef,
} from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/platform-permissions";
import { useCompany, useUpdateCompany } from "../hooks";
import type { CompanyDomain } from "../types";

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

const domainCols: DataGridColumnDef<CompanyDomain>[] = [
  { id: "host", header: "Hostname", accessorKey: "hostname" },
  {
    id: "verified",
    header: "Verified",
    cell: ({ row }) => (
      <StatusBadge
        status={row.verified ? "verified" : "pending"}
        tone={row.verified ? "success" : "warning"}
      />
    ),
  },
  {
    id: "primary",
    header: "Primary",
    cell: ({ row }) => (row.primary ? "Yes" : "No"),
  },
];

export function CompanyDetailView({ companyId }: { companyId: string }) {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch } = useCompany(companyId);
  const updateMutation = useUpdateCompany();
  const [billingEmail, setBillingEmail] = useState("");
  const [primaryColor, setPrimaryColor] = useState("");

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <Skeleton height={220} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load company
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
        title={data.legalName}
        description={`${data.tradeName} · ${data.countryCode} · ${data.tenantCount} tenants`}
        actions={
          <div className="flex gap-[var(--nx-space-2)]">
            <PermissionGate allowed={can(session, "companies:write")}>
              <Button
                size="sm"
                variant="secondary"
                loading={updateMutation.isPending}
                onClick={() =>
                  void updateMutation.mutateAsync({
                    id: data.id,
                    patch: {
                      status: data.status === "suspended" ? "active" : "suspended",
                    },
                  })
                }
              >
                {data.status === "suspended" ? "Reactivate" : "Suspend"}
              </Button>
            </PermissionGate>
            <Link href="/companies">
              <Button size="sm" variant="ghost">
                Back
              </Button>
            </Link>
          </div>
        }
      />

      <div className="flex flex-wrap gap-[var(--nx-space-2)]">
        <StatusBadge status={data.status} tone="info" />
        <StatusBadge status={data.primaryCurrency} tone="success" />
        {data.locales.locales.map((l) => (
          <StatusBadge key={l} status={l} tone="neutral" />
        ))}
      </div>

      <Tabs defaultValue="business">
        <TabsList>
          <TabsTrigger value="business">Business</TabsTrigger>
          <TabsTrigger value="tax">Tax</TabsTrigger>
          <TabsTrigger value="locales">Currencies / locales</TabsTrigger>
          <TabsTrigger value="domains">Domains</TabsTrigger>
          <TabsTrigger value="branding">Branding</TabsTrigger>
        </TabsList>

        <TabsContent value="business">
          <Panel title="Business & tax identity">
            <dl className="m-0 grid grid-cols-1 md:grid-cols-2 gap-[var(--nx-space-3)] text-[13px]">
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Industry</dt>
                <dd className="m-0 font-medium">{data.business.industry}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Tax ID</dt>
                <dd className="m-0 font-medium">{data.business.taxId}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">VAT</dt>
                <dd className="m-0 font-medium">{data.business.vatNumber}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Address</dt>
                <dd className="m-0 font-medium">
                  {data.business.registeredAddress}
                </dd>
              </div>
            </dl>
            <PermissionGate allowed={can(session, "companies:write")}>
              <div className="mt-[var(--nx-space-3)] flex flex-wrap gap-[var(--nx-space-2)] items-end">
                <Input
                  placeholder={data.business.billingEmail}
                  value={billingEmail}
                  onChange={(e) => setBillingEmail(e.target.value)}
                  aria-label="Billing email"
                />
                <Button
                  size="sm"
                  loading={updateMutation.isPending}
                  disabled={!billingEmail.trim()}
                  onClick={() =>
                    void updateMutation
                      .mutateAsync({
                        id: data.id,
                        patch: {
                          business: { billingEmail: billingEmail.trim() },
                        },
                      })
                      .then(() => setBillingEmail(""))
                  }
                >
                  Update billing email
                </Button>
              </div>
            </PermissionGate>
          </Panel>
        </TabsContent>

        <TabsContent value="tax">
          <Panel title="Tax settings">
            <dl className="m-0 grid grid-cols-1 md:grid-cols-2 gap-[var(--nx-space-3)] text-[13px]">
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Tax engine</dt>
                <dd className="m-0 font-medium">{data.tax.defaultTaxEngine}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">VAT registered</dt>
                <dd className="m-0 font-medium">
                  {data.tax.vatRegistered ? "Yes" : "No"}
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Withholding</dt>
                <dd className="m-0 font-medium">
                  {data.tax.withholdingEnabled ? "Enabled" : "Disabled"}
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Fiscal year start</dt>
                <dd className="m-0 font-medium">
                  Month {data.tax.fiscalYearStartMonth}
                </dd>
              </div>
            </dl>
          </Panel>
        </TabsContent>

        <TabsContent value="locales">
          <Panel title="Locales & currencies">
            <dl className="m-0 grid grid-cols-1 md:grid-cols-2 gap-[var(--nx-space-3)] text-[13px]">
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Default locale</dt>
                <dd className="m-0 font-medium">{data.locales.defaultLocale}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Timezone</dt>
                <dd className="m-0 font-medium">{data.locales.timeZone}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Locales</dt>
                <dd className="m-0 font-medium">
                  {data.locales.locales.join(", ")}
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Currencies</dt>
                <dd className="m-0 font-medium">
                  {data.locales.currencies.join(", ")}
                </dd>
              </div>
            </dl>
          </Panel>
        </TabsContent>

        <TabsContent value="domains">
          <DataGrid
            columns={domainCols}
            data={data.domains}
            getRowId={(r) => r.id}
          />
        </TabsContent>

        <TabsContent value="branding">
          <Panel title="Branding">
            <dl className="m-0 grid grid-cols-1 md:grid-cols-2 gap-[var(--nx-space-3)] text-[13px] mb-[var(--nx-space-3)]">
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Primary</dt>
                <dd className="m-0 font-medium">{data.branding.primaryColor}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Secondary</dt>
                <dd className="m-0 font-medium">
                  {data.branding.secondaryColor}
                </dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Logo</dt>
                <dd className="m-0 font-medium">{data.branding.logoUrl}</dd>
              </div>
              <div>
                <dt className="text-[var(--nx-text-tertiary)]">Favicon</dt>
                <dd className="m-0 font-medium">{data.branding.faviconUrl}</dd>
              </div>
            </dl>
            <PermissionGate allowed={can(session, "companies:write")}>
              <div className="flex flex-wrap gap-[var(--nx-space-2)] items-end">
                <Input
                  placeholder={data.branding.primaryColor}
                  value={primaryColor}
                  onChange={(e) => setPrimaryColor(e.target.value)}
                  aria-label="Primary color"
                />
                <Button
                  size="sm"
                  loading={updateMutation.isPending}
                  disabled={!primaryColor.trim()}
                  onClick={() =>
                    void updateMutation
                      .mutateAsync({
                        id: data.id,
                        patch: {
                          branding: { primaryColor: primaryColor.trim() },
                        },
                      })
                      .then(() => setPrimaryColor(""))
                  }
                >
                  Save primary color
                </Button>
              </div>
            </PermissionGate>
          </Panel>
        </TabsContent>
      </Tabs>
    </div>
  );
}
