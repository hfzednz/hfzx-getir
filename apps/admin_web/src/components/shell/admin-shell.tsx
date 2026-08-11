"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Button, SideNav, TopBar, type SideNavItem } from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import { can, type Permission } from "@/shared/permissions/permissions";
import { useChromeStore } from "@/stores/chrome-store";
import { CitySwitcher } from "./city-switcher";
import { CommandPalette } from "./command-palette";
import { NAV_ITEMS } from "./nav-config";

function BrandMark({ collapsed }: { collapsed: boolean }) {
  return (
    <Link
      href="/dashboard"
      className="flex items-center gap-[var(--nx-space-2)] no-underline text-[var(--nx-text-brand)]"
    >
      <span
        className="inline-flex size-7 items-center justify-center rounded-[var(--nx-radius-xs)] bg-[var(--nx-brand-600)] text-[var(--nx-text-on-brand)] text-[11px] font-bold tracking-tight"
        aria-hidden
      >
        NX
      </span>
      {!collapsed ? (
        <span className="font-[family-name:var(--nx-font-display)] text-[14px] font-semibold tracking-[-0.02em] text-[var(--nx-text-primary)]">
          NEXORA
        </span>
      ) : null}
    </Link>
  );
}

export function AdminShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const session = useAuthStore((s) => s.session);
  const logout = useAuthStore((s) => s.logout);
  const collapsed = useChromeStore((s) => s.sidebarCollapsed);
  const toggleSidebar = useChromeStore((s) => s.toggleSidebar);
  const theme = useChromeStore((s) => s.theme);
  const toggleTheme = useChromeStore((s) => s.toggleTheme);
  const setCommandPaletteOpen = useChromeStore((s) => s.setCommandPaletteOpen);

  const navItems: SideNavItem[] = NAV_ITEMS.filter((item) => {
    if (!item.permission) return true;
    return can(session, item.permission as Permission);
  }).map((item) => ({
    href: item.href,
    label: item.label,
    active:
      pathname === item.href ||
      (item.href !== "/dashboard" && pathname.startsWith(item.href)),
  }));

  return (
    <div className="nx-root flex min-h-screen bg-[var(--nx-bg-canvas)]">
      <aside className="sticky top-0 h-screen shrink-0">
        <SideNav
          items={navItems}
          collapsed={collapsed}
          brand={<BrandMark collapsed={collapsed} />}
          className="overflow-y-auto"
          renderLink={(item, className) => (
            <Link
              href={item.href}
              className={className}
              aria-current={item.active ? "page" : undefined}
              title={collapsed ? item.label : undefined}
            >
              {!collapsed ? (
                <span className="truncate">{item.label}</span>
              ) : (
                <span className="text-[10px] font-bold uppercase">
                  {item.label.slice(0, 1)}
                </span>
              )}
            </Link>
          )}
        />
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <TopBar
          leading={
            <>
              <Button
                variant="ghost"
                size="sm"
                onClick={toggleSidebar}
                aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
              >
                {collapsed ? "»" : "«"}
              </Button>
              <CitySwitcher />
            </>
          }
          center={
            <button
              type="button"
              onClick={() => setCommandPaletteOpen(true)}
              className="hidden sm:flex w-full max-w-md h-[var(--nx-control-height-sm)] items-center gap-[var(--nx-space-2)] px-[var(--nx-space-3)] rounded-[var(--nx-radius-sm)] border border-[var(--nx-border-default)] bg-[var(--nx-bg-sunken)] text-[12px] text-[var(--nx-text-tertiary)] cursor-pointer hover:border-[var(--nx-border-strong)]"
            >
              <span className="flex-1 text-left">Search or jump to…</span>
              <kbd className="font-[family-name:var(--nx-font-mono)] text-[10px] px-1.5 py-0.5 rounded border border-[var(--nx-border-subtle)] bg-[var(--nx-bg-surface)] text-[var(--nx-text-secondary)]">
                ⌘K
              </kbd>
            </button>
          }
          trailing={
            <>
              <Button variant="ghost" size="sm" onClick={toggleTheme}>
                {theme === "light" ? "Dark" : "Light"}
              </Button>
              <span className="hidden md:inline text-[12px] text-[var(--nx-text-secondary)] max-w-[160px] truncate">
                {session?.displayName ?? "—"}
              </span>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  logout();
                  router.replace("/login");
                }}
              >
                Sign out
              </Button>
            </>
          }
        />

        <main className="flex-1 min-w-0 px-[var(--nx-space-4)] py-[var(--nx-space-4)] overflow-auto">
          {children}
        </main>
      </div>

      <CommandPalette />
    </div>
  );
}
