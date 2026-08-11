import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

export interface FilterBarProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
  /** Trailing actions (e.g. Reset, Apply) */
  actions?: ReactNode;
}

export function FilterBar({
  children,
  actions,
  className,
  ...props
}: FilterBarProps) {
  return (
    <div className={cn("nx-filter-bar", className)} role="search" {...props}>
      <div className="nx-filter-bar__fields">{children}</div>
      {actions != null ? <div className="nx-filter-bar__actions">{actions}</div> : null}
    </div>
  );
}
