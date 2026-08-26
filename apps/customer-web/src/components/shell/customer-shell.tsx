"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import type { ReactNode } from "react";
import { useSession } from "@/shared/api/client";

const NAV = [
  { href: "/home", label: "Home" },
  { href: "/cart", label: "Cart" },
  { href: "/orders", label: "Orders" },
  { href: "/account", label: "Account" },
];

export function CustomerShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);

  if (pathname === "/login") {
    return <div className="nx-customer-shell">{children}</div>;
  }

  if (!session && pathname !== "/login") {
    router.replace("/login");
    return null;
  }

  return (
    <div className="nx-customer-shell flex min-h-dvh flex-col">
      <header className="sticky top-0 z-10 bg-[var(--nx-customer-header)] px-4 py-3 text-white">
        <div className="flex items-center justify-between gap-2">
          <Link href="/home" className="text-lg font-semibold tracking-tight">
            NEXORA
          </Link>
          <button
            type="button"
            className="text-sm opacity-90"
            onClick={() => {
              logout();
              router.push("/login");
            }}
          >
            Logout
          </button>
        </div>
      </header>
      <main className="flex-1 px-4 py-4">{children}</main>
      <nav
        className="sticky bottom-0 grid grid-cols-4 border-t bg-white text-xs"
        aria-label="Primary"
      >
        {NAV.map((item) => {
          const active = pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex flex-col items-center gap-1 py-3 ${active ? "text-[var(--nx-brand)] font-semibold" : "text-neutral-600"}`}
              aria-current={active ? "page" : undefined}
            >
              {item.label}
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
