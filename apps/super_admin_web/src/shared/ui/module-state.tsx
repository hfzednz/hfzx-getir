"use client";

import { Skeleton } from "@nexora/ui";

export function ModuleLoading({ cards = 6 }: { cards?: number }) {
  return (
    <div className="flex flex-col gap-[var(--nx-space-4)]">
      <Skeleton height={48} />
      <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-[var(--nx-space-3)]">
        {Array.from({ length: cards }).map((_, i) => (
          <Skeleton key={i} height={88} />
        ))}
      </div>
      <Skeleton height={280} />
    </div>
  );
}

export function ModuleError({
  title,
  message,
  onRetry,
}: {
  title: string;
  message: string;
  onRetry: () => void;
}) {
  return (
    <div className="rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-danger)] bg-[var(--nx-danger-surface)] p-[var(--nx-space-4)]">
      <p className="m-0 font-semibold text-[var(--nx-danger)]">{title}</p>
      <p className="m-0 mt-[var(--nx-space-1)] text-[var(--nx-text-secondary)]">
        {message}
      </p>
      <button
        type="button"
        onClick={onRetry}
        className="mt-[var(--nx-space-3)] text-[var(--nx-text-link)] underline cursor-pointer bg-transparent border-0"
      >
        Retry
      </button>
    </div>
  );
}

export function healthTone(
  status: string,
): "success" | "warning" | "danger" | "info" | "neutral" {
  const s = status.toLowerCase();
  if (
    ["healthy", "up", "active", "ok", "stable", "streaming", "completed", "valid", "serving", "enforced", "succeeded"].includes(
      s,
    )
  ) {
    return "success";
  }
  if (
    [
      "degraded",
      "warning",
      "warn",
      "expiring",
      "burn",
      "catching_up",
      "rebalancing",
      "throttled",
      "canary",
      "shadow",
      "acked",
      "staging",
      "scaling",
      "running",
      "draining",
      "mitigating",
      "investigating",
      "failover",
      "maintenance",
    ].includes(s)
  ) {
    return "warning";
  }
  if (
    [
      "down",
      "danger",
      "critical",
      "failed",
      "expired",
      "breached",
      "broken",
      "dead",
      "revoked",
      "backed_up",
      "sev1",
    ].includes(s)
  ) {
    return "danger";
  }
  return "info";
}
