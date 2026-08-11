"use client";

import { useState, type ReactNode } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Button,
  Input,
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
  useCampaign,
  useDuplicateCampaign,
  usePauseCampaign,
  useResumeCampaign,
  useScheduleCampaign,
  useUpdateCampaign,
} from "../hooks";

export function CampaignDetailView({ id }: { id: string }) {
  const router = useRouter();
  const canWrite = usePermission("campaigns:write");
  const { data, isLoading, isError, error, refetch } = useCampaign(id);
  const updateMut = useUpdateCampaign(id);
  const scheduleMut = useScheduleCampaign(id);
  const pauseMut = usePauseCampaign(id);
  const resumeMut = useResumeCampaign(id);
  const dupMut = useDuplicateCampaign();

  const [editName, setEditName] = useState<string | null>(null);
  const [scheduleOpen, setScheduleOpen] = useState(false);
  const [startsAt, setStartsAt] = useState("");
  const [endsAt, setEndsAt] = useState("");

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
          Failed to load campaign
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

  const name = editName ?? data.name;

  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <PageHeader
        title={data.name}
        description={
          <span className="flex items-center gap-2 flex-wrap">
            <Link
              href="/campaigns"
              className="text-[var(--nx-text-link)] no-underline"
            >
              Campaigns
            </Link>
            <span>·</span>
            <span className="tabular-nums">{data.id}</span>
            <StatusBadge status={data.status} />
          </span>
        }
        actions={
          <PermissionGate allowed={canWrite}>
            <div className="flex flex-wrap gap-2">
              {(data.status === "active" || data.status === "scheduled") && (
                <Button
                  size="sm"
                  variant="secondary"
                  loading={pauseMut.isPending}
                  onClick={() => void pauseMut.mutateAsync()}
                >
                  Pause
                </Button>
              )}
              {data.status === "paused" && (
                <Button
                  size="sm"
                  variant="secondary"
                  loading={resumeMut.isPending}
                  onClick={() => void resumeMut.mutateAsync()}
                >
                  Resume
                </Button>
              )}
              {(data.status === "draft" || data.status === "paused") && (
                <Button size="sm" variant="secondary" onClick={() => setScheduleOpen(true)}>
                  Schedule
                </Button>
              )}
              <Button
                size="sm"
                variant="ghost"
                loading={dupMut.isPending}
                onClick={() =>
                  void dupMut.mutateAsync(id).then((c) => router.push(`/campaigns/${c.id}`))
                }
              >
                Duplicate
              </Button>
              <Button
                size="sm"
                loading={updateMut.isPending}
                onClick={() => {
                  if (editName == null) {
                    setEditName(data.name);
                    return;
                  }
                  void updateMut.mutateAsync({ name: editName }).then(() => setEditName(null));
                }}
              >
                {editName == null ? "Edit" : "Save"}
              </Button>
            </div>
          </PermissionGate>
        }
      />

      {editName != null ? (
        <Input
          value={name}
          onChange={(e) => setEditName(e.target.value)}
          aria-label="Campaign name"
        />
      ) : null}

      <section className="grid grid-cols-2 md:grid-cols-4 gap-[var(--nx-space-3)]">
        <Meta label="Type" value={data.type.replaceAll("_", " ")} />
        <Meta
          label="Budget spent"
          value={`${formatMinorUnits(data.spentMinor, data.currency)} / ${formatMinorUnits(data.budgetMinor, data.currency)}`}
        />
        <Meta
          label="Starts"
          value={
            data.startsAt
              ? new Date(data.startsAt).toLocaleString("tr-TR")
              : "—"
          }
        />
        <Meta
          label="Ends"
          value={
            data.endsAt ? new Date(data.endsAt).toLocaleString("tr-TR") : "—"
          }
        />
      </section>

      <Tabs defaultValue="audience">
        <TabsList>
          <TabsTrigger value="audience">Audience</TabsTrigger>
          <TabsTrigger value="coupon">Coupon</TabsTrigger>
          <TabsTrigger value="bundle">Bundle</TabsTrigger>
          <TabsTrigger value="flash">Flash sale</TabsTrigger>
          <TabsTrigger value="personalized">Personalized</TabsTrigger>
        </TabsList>

        <TabsContent value="audience">
          {data.audience ? (
            <Panel>
              <p className="m-0 font-semibold">{data.audience.name}</p>
              <p className="m-0 text-[12px] text-[var(--nx-text-secondary)]">
                {data.audience.rulesSummary}
              </p>
              <p className="m-0 text-[12px] tabular-nums">
                Segment size: {data.audience.segmentSize.toLocaleString("tr-TR")}
              </p>
            </Panel>
          ) : (
            <EmptyPanel text="No audience attached" />
          )}
        </TabsContent>

        <TabsContent value="coupon">
          {data.coupon ? (
            <Panel>
              <p className="m-0 font-semibold tabular-nums">{data.coupon.code}</p>
              <p className="m-0 text-[12px]">
                {data.coupon.discountType === "percent"
                  ? `${data.coupon.discountValue}% off`
                  : formatMinorUnits(data.coupon.discountValue, data.currency)}
                {" · "}
                min order {formatMinorUnits(data.coupon.minOrderMinor, data.currency)}
              </p>
              <p className="m-0 text-[12px] tabular-nums">
                Redemptions {data.coupon.redemptions.toLocaleString("tr-TR")} /{" "}
                {data.coupon.maxRedemptions.toLocaleString("tr-TR")}
              </p>
            </Panel>
          ) : (
            <EmptyPanel text="No coupon configured" />
          )}
        </TabsContent>

        <TabsContent value="bundle">
          {data.bundle ? (
            <Panel>
              <p className="m-0 font-semibold">{data.bundle.label}</p>
              <p className="m-0 text-[12px]">
                SKUs: {data.bundle.skuIds.join(", ")}
              </p>
              <p className="m-0 text-[12px] tabular-nums">
                {formatMinorUnits(data.bundle.bundlePriceMinor, data.currency)} ·{" "}
                {data.bundle.savingsPct}% savings
              </p>
            </Panel>
          ) : (
            <EmptyPanel text="No bundle configured" />
          )}
        </TabsContent>

        <TabsContent value="flash">
          {data.flashSale ? (
            <Panel>
              <p className="m-0 font-semibold">
                {data.flashSale.discountPct}% flash
              </p>
              <p className="m-0 text-[12px] tabular-nums">
                Sold {data.flashSale.sold} / {data.flashSale.stockCap}
              </p>
              <p className="m-0 text-[12px]">
                {new Date(data.flashSale.startsAt).toLocaleString("tr-TR")} →{" "}
                {new Date(data.flashSale.endsAt).toLocaleString("tr-TR")}
              </p>
            </Panel>
          ) : (
            <EmptyPanel text="No flash sale window" />
          )}
        </TabsContent>

        <TabsContent value="personalized">
          {data.personalized ? (
            <Panel>
              <p className="m-0 font-semibold">{data.personalized.model}</p>
              <p className="m-0 text-[12px]">
                Uplift {data.personalized.upliftPct}% · channels{" "}
                {data.personalized.channels.join(", ")}
              </p>
            </Panel>
          ) : (
            <EmptyPanel text="No personalized promo" />
          )}
        </TabsContent>
      </Tabs>

      {scheduleOpen ? (
        <div className="nx-confirm-root" role="presentation">
          <button
            type="button"
            className="nx-confirm-backdrop"
            aria-label="Close"
            onClick={() => setScheduleOpen(false)}
          />
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="schedule-campaign-title"
            className="nx-confirm-dialog"
          >
            <h2
              id="schedule-campaign-title"
              className="nx-confirm-dialog__title"
            >
              Schedule campaign
            </h2>
            <div className="flex flex-col gap-2 mt-2">
              <label className="text-[12px] text-[var(--nx-text-secondary)]">
                Starts (ISO)
                <Input
                  className="mt-1"
                  value={startsAt}
                  onChange={(e) => setStartsAt(e.target.value)}
                  placeholder="2026-08-10T00:00:00Z"
                />
              </label>
              <label className="text-[12px] text-[var(--nx-text-secondary)]">
                Ends (ISO, optional)
                <Input
                  className="mt-1"
                  value={endsAt}
                  onChange={(e) => setEndsAt(e.target.value)}
                  placeholder="2026-09-10T00:00:00Z"
                />
              </label>
            </div>
            <div className="nx-confirm-dialog__actions mt-3">
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setScheduleOpen(false)}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                loading={scheduleMut.isPending}
                onClick={() => {
                  if (!startsAt.trim()) return;
                  void scheduleMut
                    .mutateAsync({
                      startsAt: startsAt.trim(),
                      endsAt: endsAt.trim() || null,
                    })
                    .then(() => setScheduleOpen(false));
                }}
              >
                Schedule
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-3)]">
      <p className="m-0 text-[11px] text-[var(--nx-text-tertiary)] uppercase tracking-wide">
        {label}
      </p>
      <p className="m-0 mt-1 text-[13px] font-semibold capitalize">{value}</p>
    </div>
  );
}

function Panel({ children }: { children: ReactNode }) {
  return (
    <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)] flex flex-col gap-2">
      {children}
    </div>
  );
}

function EmptyPanel({ text }: { text: string }) {
  return (
    <div className="bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-4)] text-[13px] text-[var(--nx-text-secondary)]">
      {text}
    </div>
  );
}
