"use client";

import Link from "next/link";
import {
  Button,
  DataGrid,
  KpiCard,
  PageHeader,
  PermissionGate,
  Skeleton,
  StatusBadge,
  type DataGridColumnDef,
} from "@nexora/ui";
import { formatMinorUnits } from "@/shared/lib/money";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can } from "@/shared/permissions/permissions";
import { useCourierDetail } from "../hooks";
import type {
  CourierAssignment,
  CourierBonus,
  CourierDocument,
  CourierLiveStatus,
  CourierPayment,
  CourierPenalty,
  CourierRating,
  CourierScheduleSlot,
} from "../types";

function statusTone(
  status: CourierLiveStatus,
): "success" | "warning" | "danger" | "info" | "neutral" {
  switch (status) {
    case "available":
      return "success";
    case "busy":
      return "info";
    case "break":
      return "warning";
    case "emergency":
      return "danger";
    default:
      return "neutral";
  }
}

const assignmentCols: DataGridColumnDef<CourierAssignment>[] = [
  { id: "order", header: "Order", accessorKey: "orderId" },
  { id: "status", header: "Status", accessorKey: "status" },
  { id: "zone", header: "Zone", accessorKey: "zoneName" },
  {
    id: "eta",
    header: "ETA",
    cell: ({ row }) => `${row.etaMinutes} min`,
    align: "right",
  },
];

const scheduleCols: DataGridColumnDef<CourierScheduleSlot>[] = [
  { id: "day", header: "Day", accessorKey: "day", width: 70 },
  { id: "start", header: "Start", accessorKey: "start" },
  { id: "end", header: "End", accessorKey: "end" },
  { id: "zone", header: "Zone", accessorKey: "zoneName" },
];

const ratingCols: DataGridColumnDef<CourierRating>[] = [
  {
    id: "score",
    header: "Score",
    cell: ({ row }) => row.score.toFixed(1),
    align: "right",
    width: 70,
  },
  { id: "comment", header: "Comment", accessorKey: "comment" },
  { id: "order", header: "Order", accessorKey: "orderId", width: 110 },
];

const docCols: DataGridColumnDef<CourierDocument>[] = [
  { id: "type", header: "Document", accessorKey: "type" },
  {
    id: "status",
    header: "Status",
    cell: ({ row }) => (
      <StatusBadge
        status={row.status}
        tone={
          row.status === "valid"
            ? "success"
            : row.status === "expiring"
              ? "warning"
              : "danger"
        }
      />
    ),
  },
  {
    id: "exp",
    header: "Expires",
    cell: ({ row }) => row.expiresAt ?? "—",
  },
];

const paymentCols: DataGridColumnDef<CourierPayment>[] = [
  { id: "period", header: "Period", accessorKey: "period" },
  {
    id: "net",
    header: "Net",
    cell: ({ row }) => formatMinorUnits(row.netAmount, row.currency),
    align: "right",
  },
  {
    id: "status",
    header: "Status",
    cell: ({ row }) => (
      <StatusBadge
        status={row.status}
        tone={row.status === "paid" ? "success" : "warning"}
      />
    ),
  },
];

const bonusCols: DataGridColumnDef<CourierBonus>[] = [
  { id: "reason", header: "Reason", accessorKey: "reason" },
  {
    id: "amount",
    header: "Amount",
    cell: ({ row }) => formatMinorUnits(row.amount, row.currency),
    align: "right",
  },
];

const penaltyCols: DataGridColumnDef<CourierPenalty>[] = [
  { id: "reason", header: "Reason", accessorKey: "reason" },
  {
    id: "amount",
    header: "Amount",
    cell: ({ row }) => formatMinorUnits(row.amount, row.currency),
    align: "right",
  },
];

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

export function CourierDetailView({ courierId }: { courierId: string }) {
  const session = useAuthStore((s) => s.session);
  const { data, isLoading, isError, error, refetch } = useCourierDetail(courierId);

  if (isLoading) {
    return (
      <div className="flex flex-col gap-[var(--nx-space-4)]">
        <Skeleton height={48} />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} height={88} />
          ))}
        </div>
        <Skeleton height={240} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
        <p className="m-0 font-semibold text-[var(--nx-danger)]">
          Failed to load courier
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
        title={`${data.fullName} · ${data.code}`}
        description={
          <span>
            <Link href="/couriers" className="text-[var(--nx-text-link)]">
              Couriers
            </Link>
            {" · "}
            {data.zoneName} · {data.phone}
          </span>
        }
        actions={
          <div className="flex gap-[var(--nx-space-2)]">
            <PermissionGate allowed={can(session, "couriers:intervene")}>
              <Button size="sm" variant="danger">
                Force offline
              </Button>
            </PermissionGate>
            <PermissionGate allowed={can(session, "couriers:write")}>
              <Button size="sm" variant="secondary">
                Edit profile
              </Button>
            </PermissionGate>
          </div>
        }
      />

      <div className="flex flex-wrap items-center gap-[var(--nx-space-2)]">
        <StatusBadge
          status={data.liveStatus}
          tone={statusTone(data.liveStatus)}
        />
        {data.emergency ? (
          <StatusBadge
            status={data.emergencyReason ?? "Emergency"}
            tone="danger"
          />
        ) : null}
        <span className="text-[12px] text-[var(--nx-text-tertiary)]">
          Last seen {new Date(data.lastSeenAt).toLocaleString("tr-TR")}
        </span>
      </div>

      <section className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
        <KpiCard
          title="Rating"
          value={`${data.rating.toFixed(1)} (${data.ratingCount})`}
          tone="brand"
        />
        <KpiCard
          title="Deliveries today"
          value={data.performance.deliveriesToday}
          tone="success"
        />
        <KpiCard
          title="On-time %"
          value={`${data.performance.onTimePct.toFixed(1)}%`}
          tone="neutral"
        />
        <KpiCard
          title="Avg delivery"
          value={`${data.performance.avgDeliveryMinutes} min`}
          tone="neutral"
        />
      </section>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-[var(--nx-space-3)]">
        <Panel title="Active assignments">
          <DataGrid
            columns={assignmentCols}
            data={data.assignments}
            getRowId={(r) => r.id}
            emptyMessage="No active assignments"
          />
        </Panel>
        <Panel title="Schedule">
          <DataGrid
            columns={scheduleCols}
            data={data.schedule}
            getRowId={(r) => r.id}
          />
        </Panel>
        <Panel title="Performance">
          <ul className="m-0 p-0 list-none grid grid-cols-2 gap-[var(--nx-space-2)] text-[13px]">
            <li>
              Week deliveries:{" "}
              <strong>{data.performance.deliveriesWeek}</strong>
            </li>
            <li>
              Acceptance:{" "}
              <strong>{data.performance.acceptanceRatePct}%</strong>
            </li>
            <li>
              Courier cancel:{" "}
              <strong>{data.performance.cancelByCourierPct}%</strong>
            </li>
            <li>
              Avg ETA:{" "}
              <strong>{data.performance.avgDeliveryMinutes} min</strong>
            </li>
          </ul>
        </Panel>
        <Panel title="Vehicle">
          <dl className="m-0 grid grid-cols-2 gap-[var(--nx-space-2)] text-[13px]">
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Plate</dt>
              <dd className="m-0 font-medium">{data.vehicle.plate}</dd>
            </div>
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Type</dt>
              <dd className="m-0 font-medium">{data.vehicle.type}</dd>
            </div>
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Model</dt>
              <dd className="m-0 font-medium">{data.vehicle.model}</dd>
            </div>
            <div>
              <dt className="text-[var(--nx-text-tertiary)]">Insurance</dt>
              <dd className="m-0 font-medium">
                {data.vehicle.insuranceExpiresAt}
              </dd>
            </div>
          </dl>
        </Panel>
        <Panel title="Ratings">
          <DataGrid
            columns={ratingCols}
            data={data.ratings}
            getRowId={(r) => r.id}
          />
        </Panel>
        <Panel title="Documents">
          <DataGrid
            columns={docCols}
            data={data.documents}
            getRowId={(r) => r.id}
          />
        </Panel>
        <Panel title="Payments">
          <DataGrid
            columns={paymentCols}
            data={data.payments}
            getRowId={(r) => r.id}
          />
        </Panel>
        <Panel title="Bonuses & penalties">
          <div className="flex flex-col gap-[var(--nx-space-3)]">
            <DataGrid
              columns={bonusCols}
              data={data.bonuses}
              getRowId={(r) => r.id}
              emptyMessage="No bonuses"
            />
            <DataGrid
              columns={penaltyCols}
              data={data.penalties}
              getRowId={(r) => r.id}
              emptyMessage="No penalties"
            />
          </div>
        </Panel>
      </div>
    </div>
  );
}
