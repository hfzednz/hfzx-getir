import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

export interface TopBarProps extends HTMLAttributes<HTMLElement> {
  /** Leading slot (menu toggle, breadcrumbs) */
  leading?: ReactNode;
  /** Center / search slot */
  center?: ReactNode;
  /** Trailing slot (actions, avatar) */
  trailing?: ReactNode;
  children?: ReactNode;
}

export function TopBar({
  leading,
  center,
  trailing,
  children,
  className,
  ...props
}: TopBarProps) {
  return (
    <header className={cn("nx-topbar", className)} {...props}>
      {leading != null ? <div className="nx-topbar__leading">{leading}</div> : null}
      {center != null ? (
        <div className="nx-topbar__center">{center}</div>
      ) : (
        <div className="nx-topbar__spacer" />
      )}
      {children}
      {trailing != null ? <div className="nx-topbar__trailing">{trailing}</div> : null}
    </header>
  );
}
