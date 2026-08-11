"use client";

import Link from "next/link";
import {
  Button,
  DataGrid,
  type DataGridColumnDef,
  KpiCard,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { usePermission } from "@/shared/permissions/use-permission";
import {
  useApproveFinanceRefund,
  useApprovePayout,
  useFinanceSnapshot,
  useSettleCourier,
} from "../hooks";
import type {
  CourierSettlement,
  InvoiceRow,
  PaymentRow,
  PayoutRow,
  RefundRow,
  RevenueRow,
  SupplierPayment,
  TaxRow,
} from "../types";

export function FinanceView() {
  const canApprovePayout = usePermission("finance:payout:approve");
  const canSettlement = usePermission("finance:settlement");
  const canRefund = usePermission("orders:refund");
  const canExport = usePermission("finance:export");

  const { data, isLoading, isError, error, refetch } = useFinanceSnapshot();
  const approvePayout = useApprovePayout();
  const settleCourier = useSettleCourier();
  const approveRefund = useApproveFinanceRefund();

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
          Failed to load finance
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

  const revenueCols: DataGridColumnDef<RevenueRow & Record<string, unknown>>[] =
    [
      { id: "period", header: "Period", accessorKey: "period" },
      {
        id: "gmv",
        header: "GMV",
        align: "right",
        cell: ({ row }) => formatMinorUnits(row.gmvMinor, row.currency),
      },
      {
        id: "net",
        header: "Net revenue",
        align: "right",
        cell: ({ row }) => formatMinorUnits(row.netRevenueMinor, row.currency),
      },
    ];

  const refundCols: DataGridColumnDef<RefundRow & Record<string, unknown>>[] = [
    { id: "order", header: "Order", accessorKey: "orderId" },
    {
      id: "amount",
      header: "Amount",
      align: "right",
      cell: ({ row }) => formatMinorUnits(row.amountMinor, row.currency),
    },
    { id: "reason", header: "Reason", accessorKey: "reason" },
    {
      id: "status",
      header: "Status",
      cell: ({ row }) => <StatusBadge status={row.status} />,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <PermissionGate allowed={canRefund}>
          <Button
            size="sm"
            variant="ghost"
            disabled={row.status !== "pending"}
            loading={approveRefund.isPending}
            onClick={() => void approveRefund.mutateAsync(row.id)}
          >
            Approve
          </Button>
        </PermissionGate>
      ),
    },
  ];

  const taxCols: DataGridColumnDef<TaxRow & Record<string, unknown>>[] = [
    { id: "j", header: "Jurisdiction", accessorKey: "jurisdiction" },
    {
      id: "rate",
      header: "Rate",
      align: "right",
      cell: ({ row }) => `${row.ratePct}%`,
    },
    {
      id: "collected",
      header: "Collected",
      align: "right",
      cell: ({ row }) => formatMinorUnits(row.collectedMinor, row.currency),
    },
  ];

  const invoiceCols: DataGridColumnDef<InvoiceRow & Record<string, unknown>>[] =
    [
      { id: "num", header: "Invoice", accessorKey: "number" },
      { id: "cp", header: "Counterparty", accessorKey: "counterparty" },
      {
        id: "amt",
        header: "Amount",
        align: "right",
        cell: ({ row }) => formatMinorUnits(row.amountMinor, row.currency),
      },
      {
        id: "st",
        header: "Status",
        cell: ({ row }) => <StatusBadge status={row.status} />,
      },
    ];

  const paymentCols: DataGridColumnDef<PaymentRow & Record<string, unknown>>[] =
    [
      { id: "method", header: "Method", accessorKey: "method" },
      {
        id: "amt",
        header: "Amount",
        align: "right",
        cell: ({ row }) => formatMinorUnits(row.amountMinor, row.currency),
      },
      {
        id: "st",
        header: "Status",
        cell: ({ row }) => <StatusBadge status={row.status} />,
      },
      {
        id: "at",
        header: "At",
        cell: ({ row }) => new Date(row.at).toLocaleString("tr-TR"),
      },
    ];

  const payoutCols: DataGridColumnDef<PayoutRow & Record<string, unknown>>[] = [
    { id: "ben", header: "Beneficiary", accessorKey: "beneficiary" },
    {
      id: "amt",
      header: "Amount",
      align: "right",
      cell: ({ row }) => formatMinorUnits(row.amountMinor, row.currency),
    },
    {
      id: "st",
      header: "Status",
      cell: ({ row }) => <StatusBadge status={row.status} />,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <PermissionGate allowed={canApprovePayout}>
          <Button
            size="sm"
            variant="ghost"
            disabled={row.status !== "pending"}
            loading={approvePayout.isPending}
            onClick={() => void approvePayout.mutateAsync(row.id)}
          >
            Approve payout
          </Button>
        </PermissionGate>
      ),
    },
  ];

  const courierCols: DataGridColumnDef<
    CourierSettlement & Record<string, unknown>
  >[] = [
    { id: "name", header: "Courier", accessorKey: "courierName" },
    {
      id: "del",
      header: "Deliveries",
      align: "right",
      cell: ({ row }) => row.deliveries,
    },
    {
      id: "amt",
      header: "Amount",
      align: "right",
      cell: ({ row }) => formatMinorUnits(row.amountMinor, row.currency),
    },
    {
      id: "st",
      header: "Status",
      cell: ({ row }) => <StatusBadge status={row.status} />,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <PermissionGate allowed={canSettlement}>
          <Button
            size="sm"
            variant="ghost"
            disabled={row.status === "settled"}
            loading={settleCourier.isPending}
            onClick={() => void settleCourier.mutateAsync(row.id)}
          >
            Settle
          </Button>
        </PermissionGate>
      ),
    },
  ];

  const supplierCols: DataGridColumnDef<
    SupplierPayment & Record<string, unknown>
  >[] = [
    { id: "sup", header: "Supplier", accessorKey: "supplier" },
    {
      id: "amt",
      header: "Amount",
      align: "right",
      cell: ({ row }) => formatMinorUnits(row.amountMinor, row.currency),
    },
    {
      id: "st",
      header: "Status",
      cell: ({ row }) => <StatusBadge status={row.status} />,
    },
    {
      id: "due",
      header: "Due",
      cell: ({ row }) => new Date(row.dueAt).toLocaleDateString("tr-TR"),
    },
  ];

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title="Finance"
        description="Revenue, refunds, taxes, invoices, payments, payouts, settlements, profit"
        actions={
          <PermissionGate allowed={canExport}>
            <Link
              href="/reports"
              className="nx-btn nx-btn--secondary nx-btn--sm inline-flex items-center no-underline"
            >
              Report links
            </Link>
          </PermissionGate>
        }
      />

      <section className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
        {data.kpis.map((k) => (
          <KpiCard
            key={k.id}
            title={k.label}
            value={formatMinorUnits(k.valueMinor, k.currency)}
            delta={`${k.deltaPct > 0 ? "+" : ""}${k.deltaPct.toFixed(1)}%`}
            tone={k.id === "refunds" ? "danger" : "brand"}
          />
        ))}
      </section>

      <section className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)]">
        <h3 className="m-0 mb-3 text-[var(--nx-font-size-title)] font-semibold">
          Profit analysis
        </h3>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-3 text-[12px]">
          <ProfitCell
            label="GMV"
            value={formatMinorUnits(data.profit.gmvMinor, data.profit.currency)}
          />
          <ProfitCell
            label="COGS"
            value={formatMinorUnits(data.profit.cogsMinor, data.profit.currency)}
          />
          <ProfitCell
            label="Delivery"
            value={formatMinorUnits(
              data.profit.deliveryCostMinor,
              data.profit.currency,
            )}
          />
          <ProfitCell
            label="Promo"
            value={formatMinorUnits(
              data.profit.promoCostMinor,
              data.profit.currency,
            )}
          />
          <ProfitCell
            label="Contribution"
            value={formatMinorUnits(
              data.profit.contributionMinor,
              data.profit.currency,
            )}
          />
        </div>
      </section>

      <Tabs defaultValue="revenue">
        <TabsList>
          <TabsTrigger value="revenue">Revenue</TabsTrigger>
          <TabsTrigger value="refunds">Refunds</TabsTrigger>
          <TabsTrigger value="taxes">Taxes</TabsTrigger>
          <TabsTrigger value="invoices">Invoices</TabsTrigger>
          <TabsTrigger value="payments">Payments</TabsTrigger>
          <TabsTrigger value="payouts">Payouts</TabsTrigger>
          <TabsTrigger value="courier">Courier settlements</TabsTrigger>
          <TabsTrigger value="supplier">Supplier payments</TabsTrigger>
          <TabsTrigger value="reports">Reports</TabsTrigger>
        </TabsList>

        <TabsContent value="revenue">
          <DataGrid
            columns={revenueCols}
            data={data.revenue as (RevenueRow & Record<string, unknown>)[]}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="refunds">
          <DataGrid
            columns={refundCols}
            data={data.refunds as (RefundRow & Record<string, unknown>)[]}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="taxes">
          <DataGrid
            columns={taxCols}
            data={data.taxes as (TaxRow & Record<string, unknown>)[]}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="invoices">
          <DataGrid
            columns={invoiceCols}
            data={data.invoices as (InvoiceRow & Record<string, unknown>)[]}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="payments">
          <DataGrid
            columns={paymentCols}
            data={data.payments as (PaymentRow & Record<string, unknown>)[]}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="payouts">
          <DataGrid
            columns={payoutCols}
            data={data.payouts as (PayoutRow & Record<string, unknown>)[]}
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="courier">
          <DataGrid
            columns={courierCols}
            data={
              data.courierSettlements as (CourierSettlement &
                Record<string, unknown>)[]
            }
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="supplier">
          <DataGrid
            columns={supplierCols}
            data={
              data.supplierPayments as (SupplierPayment &
                Record<string, unknown>)[]
            }
            getRowId={(r) => r.id}
          />
        </TabsContent>
        <TabsContent value="reports">
          <ul className="m-0 p-0 list-none flex flex-col gap-3">
            {data.reports.map((r) => (
              <li
                key={r.id}
                className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-3"
              >
                <Link
                  href={r.href}
                  className="font-semibold text-[13px] text-[var(--nx-text-link)]"
                >
                  {r.title}
                </Link>
                <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                  {r.description}
                </p>
              </li>
            ))}
          </ul>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function ProfitCell({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)] uppercase">
        {label}
      </p>
      <p className="m-0 mt-1 font-semibold tabular-nums">{value}</p>
    </div>
  );
}
