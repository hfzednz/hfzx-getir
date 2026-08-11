import type { HTMLAttributes } from "react";
import { Badge, type BadgeTone } from "./Badge";
import { cn } from "./tokens/cn";

export type StatusTone = Exclude<BadgeTone, "accent">;

export interface StatusBadgeProps extends HTMLAttributes<HTMLSpanElement> {
  status: string;
  tone?: StatusTone;
  /** Show a leading status dot */
  withDot?: boolean;
}

const statusToneMap: Record<string, StatusTone> = {
  active: "success",
  online: "success",
  completed: "success",
  success: "success",
  delivered: "success",
  pending: "warning",
  processing: "warning",
  warning: "warning",
  scheduled: "info",
  draft: "neutral",
  inactive: "neutral",
  offline: "neutral",
  cancelled: "danger",
  failed: "danger",
  error: "danger",
  danger: "danger",
  info: "info",
};

function resolveTone(status: string, tone?: StatusTone): StatusTone {
  if (tone) return tone;
  const key = status.trim().toLowerCase();
  return statusToneMap[key] ?? "neutral";
}

export function StatusBadge({
  status,
  tone,
  withDot = true,
  className,
  ...props
}: StatusBadgeProps) {
  const resolved = resolveTone(status, tone);

  return (
    <Badge tone={resolved} className={cn("nx-status-badge", className)} {...props}>
      {withDot ? <span className="nx-status-badge__dot" aria-hidden /> : null}
      {status}
    </Badge>
  );
}
