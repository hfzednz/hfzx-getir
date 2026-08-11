import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

export interface ChartFrameProps extends HTMLAttributes<HTMLDivElement> {
  title: string;
  description?: ReactNode;
  actions?: ReactNode;
  children?: ReactNode;
}

export function ChartFrame({
  title,
  description,
  actions,
  children,
  className,
  ...props
}: ChartFrameProps) {
  return (
    <section className={cn("nx-chart-frame", className)} {...props}>
      <div className="nx-chart-frame__header">
        <div className="nx-chart-frame__copy">
          <h3 className="nx-chart-frame__title">{title}</h3>
          {description != null ? (
            <p className="nx-chart-frame__desc">{description}</p>
          ) : null}
        </div>
        {actions != null ? (
          <div className="nx-chart-frame__actions">{actions}</div>
        ) : null}
      </div>
      <div className="nx-chart-frame__body">{children}</div>
    </section>
  );
}
