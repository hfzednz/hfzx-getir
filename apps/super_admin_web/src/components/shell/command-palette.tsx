"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useChromeStore } from "@/stores/chrome-store";
import { COMMAND_ROUTES } from "./nav-config";

export function CommandPalette() {
  const open = useChromeStore((s) => s.commandPaletteOpen);
  const setOpen = useChromeStore((s) => s.setCommandPaletteOpen);
  const toggle = useChromeStore((s) => s.toggleCommandPalette);
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        toggle();
      }
      if (e.key === "Escape" && open) {
        setOpen(false);
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, setOpen, toggle]);

  useEffect(() => {
    if (!open) {
      setQuery("");
      setActiveIndex(0);
    }
  }, [open]);

  const results = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return COMMAND_ROUTES.slice(0, 12);
    return COMMAND_ROUTES.filter((r) => {
      const hay = `${r.label} ${r.href} ${r.keywords ?? ""}`.toLowerCase();
      return hay.includes(q);
    }).slice(0, 12);
  }, [query]);

  useEffect(() => {
    setActiveIndex(0);
  }, [query]);

  const go = useCallback(
    (href: string) => {
      setOpen(false);
      router.push(href);
    },
    [router, setOpen],
  );

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-[var(--nx-z-modal)] flex items-start justify-center pt-[12vh] px-[var(--nx-space-4)]"
      role="dialog"
      aria-modal="true"
      aria-label="Command palette"
    >
      <button
        type="button"
        className="absolute inset-0 bg-[var(--nx-bg-overlay)] border-0 cursor-default"
        aria-label="Close command palette"
        onClick={() => setOpen(false)}
      />
      <div className="relative w-full max-w-lg bg-[var(--nx-bg-surface)] border border-[var(--nx-border-default)] rounded-[var(--nx-radius-md)] shadow-[var(--nx-shadow-3)] overflow-hidden">
        <input
          autoFocus
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "ArrowDown") {
              e.preventDefault();
              setActiveIndex((i) => Math.min(i + 1, results.length - 1));
            } else if (e.key === "ArrowUp") {
              e.preventDefault();
              setActiveIndex((i) => Math.max(i - 1, 0));
            } else if (e.key === "Enter" && results[activeIndex]) {
              e.preventDefault();
              go(results[activeIndex].href);
            }
          }}
          placeholder="Search platform routes… (↑↓ Enter)"
          className="w-full h-12 px-[var(--nx-space-4)] border-0 border-b border-[var(--nx-border-subtle)] bg-transparent text-[var(--nx-font-size-body)] text-[var(--nx-text-primary)] placeholder:text-[var(--nx-text-tertiary)] outline-none"
        />
        <ul className="m-0 p-[var(--nx-space-2)] list-none max-h-72 overflow-auto">
          {results.length === 0 ? (
            <li className="px-[var(--nx-space-3)] py-[var(--nx-space-2)] text-[var(--nx-text-secondary)] text-[12px]">
              No matches
            </li>
          ) : (
            results.map((r, i) => (
              <li key={`${r.href}-${r.label}`} className="m-0">
                <button
                  type="button"
                  onClick={() => go(r.href)}
                  onMouseEnter={() => setActiveIndex(i)}
                  className={`w-full flex items-center justify-between gap-[var(--nx-space-3)] px-[var(--nx-space-3)] py-[var(--nx-space-2)] rounded-[var(--nx-radius-sm)] text-left border-0 cursor-pointer ${
                    i === activeIndex
                      ? "bg-[var(--nx-brand-50)] text-[var(--nx-nav-item-active)]"
                      : "bg-transparent text-[var(--nx-text-primary)] hover:bg-[var(--nx-bg-sunken)]"
                  }`}
                >
                  <span className="text-[13px] font-medium">{r.label}</span>
                  <span className="text-[11px] text-[var(--nx-text-tertiary)] font-[family-name:var(--nx-font-mono)]">
                    {r.href}
                  </span>
                </button>
              </li>
            ))
          )}
        </ul>
        <div className="px-[var(--nx-space-4)] py-[var(--nx-space-2)] border-t border-[var(--nx-border-subtle)] text-[11px] text-[var(--nx-text-tertiary)]">
          Esc to close · Ctrl/⌘ K to toggle
        </div>
      </div>
    </div>
  );
}
