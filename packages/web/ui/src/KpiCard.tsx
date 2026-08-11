import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

export type KpiTone = "neutral" | "brand" | "success" | "warning" | "danger";

export interface KpiCardProps extends HTMLAttributes<HTMLDivElement> {
  title: string;
  value: ReactNode;
  delta?: ReactNode;
  tone?: KpiTone;
}

export function KpiCard({
  title,
  value,
  delta,
  tone = "neutral",
  className,
  ...props
}: KpiCardProps) {
  return (
    <div className={cn("nx-kpi-card", className)} {...props}>
      <p className="nx-kpi-card__title">{title}</p>
      <p className="nx-kpi-card__value">{value}</p>
      {delta != null ? (
        <p className={cn("nx-kpi-card__delta", `nx-kpi-card__delta--${tone}`)}>{delta}</p>
      ) : null}
    </div>
  );
}
