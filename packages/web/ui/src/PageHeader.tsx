import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

export interface PageHeaderProps extends HTMLAttributes<HTMLElement> {
  title: string;
  description?: ReactNode;
  actions?: ReactNode;
}

export function PageHeader({
  title,
  description,
  actions,
  className,
  ...props
}: PageHeaderProps) {
  return (
    <header className={cn("nx-page-header", className)} {...props}>
      <div className="nx-page-header__copy">
        <h1 className="nx-page-header__title">{title}</h1>
        {description != null ? (
          <p className="nx-page-header__desc">{description}</p>
        ) : null}
      </div>
      {actions != null ? <div className="nx-page-header__actions">{actions}</div> : null}
    </header>
  );
}
