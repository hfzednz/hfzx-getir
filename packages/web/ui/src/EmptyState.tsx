import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

export interface EmptyStateProps extends HTMLAttributes<HTMLDivElement> {
  title: string;
  description?: ReactNode;
  icon?: ReactNode;
  action?: ReactNode;
}

export function EmptyState({
  title,
  description,
  icon,
  action,
  className,
  ...props
}: EmptyStateProps) {
  return (
    <div className={cn("nx-empty-state", className)} role="status" {...props}>
      {icon != null ? (
        <div className="nx-empty-state__icon" aria-hidden>
          {icon}
        </div>
      ) : null}
      <div className="nx-empty-state__copy">
        <p className="nx-empty-state__title">{title}</p>
        {description != null ? (
          <p className="nx-empty-state__desc">{description}</p>
        ) : null}
      </div>
      {action != null ? <div className="nx-empty-state__action">{action}</div> : null}
    </div>
  );
}
