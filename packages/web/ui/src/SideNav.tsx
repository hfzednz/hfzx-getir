import type { HTMLAttributes, ReactNode, AnchorHTMLAttributes } from "react";
import { cn } from "./tokens/cn";

export interface SideNavItem {
  href: string;
  label: string;
  icon?: ReactNode;
  active?: boolean;
}

export interface SideNavProps extends HTMLAttributes<HTMLElement> {
  items: SideNavItem[];
  collapsed?: boolean;
  /** Optional brand / logo slot at top */
  brand?: ReactNode;
  /** Custom link render — defaults to <a> */
  renderLink?: (item: SideNavItem, className: string) => ReactNode;
}

export function SideNav({
  items,
  collapsed = false,
  brand,
  renderLink,
  className,
  ...props
}: SideNavProps) {
  return (
    <nav
      className={cn("nx-sidenav", collapsed && "nx-sidenav--collapsed", className)}
      aria-label="Side navigation"
      {...props}
    >
      {brand != null ? <div className="nx-sidenav__brand">{brand}</div> : null}
      <ul className="nx-sidenav__list">
        {items.map((item) => {
          const itemClass = cn(
            "nx-sidenav__item",
            item.active && "nx-sidenav__item--active",
          );

          return (
            <li key={item.href}>
              {renderLink ? (
                renderLink(item, itemClass)
              ) : (
                <DefaultLink item={item} className={itemClass} collapsed={collapsed} />
              )}
            </li>
          );
        })}
      </ul>
    </nav>
  );
}

function DefaultLink({
  item,
  className,
  collapsed,
}: {
  item: SideNavItem;
  className: string;
  collapsed: boolean;
}) {
  const anchorProps: AnchorHTMLAttributes<HTMLAnchorElement> = {
    href: item.href,
    className,
    "aria-current": item.active ? "page" : undefined,
    title: collapsed ? item.label : undefined,
  };

  return (
    <a {...anchorProps}>
      {item.icon != null ? (
        <span className="nx-sidenav__icon" aria-hidden>
          {item.icon}
        </span>
      ) : null}
      {!collapsed ? <span className="nx-sidenav__label">{item.label}</span> : null}
      {collapsed && !item.icon ? (
        <span className="nx-sidenav__initial">{item.label.slice(0, 1)}</span>
      ) : null}
    </a>
  );
}
