"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

export type ThemeMode = "light" | "dark";

interface ChromeState {
  sidebarCollapsed: boolean;
  theme: ThemeMode;
  /** Optional tenant filter context — not city-ops. */
  tenantContextId: string | null;
  commandPaletteOpen: boolean;
  setSidebarCollapsed: (collapsed: boolean) => void;
  toggleSidebar: () => void;
  setTheme: (theme: ThemeMode) => void;
  toggleTheme: () => void;
  setTenantContextId: (tenantContextId: string | null) => void;
  setCommandPaletteOpen: (open: boolean) => void;
  toggleCommandPalette: () => void;
}

export const useChromeStore = create<ChromeState>()(
  persist(
    (set, get) => ({
      sidebarCollapsed: false,
      theme: "light",
      tenantContextId: null,
      commandPaletteOpen: false,
      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
      toggleSidebar: () =>
        set({ sidebarCollapsed: !get().sidebarCollapsed }),
      setTheme: (theme) => set({ theme }),
      toggleTheme: () =>
        set({ theme: get().theme === "light" ? "dark" : "light" }),
      setTenantContextId: (tenantContextId) => set({ tenantContextId }),
      setCommandPaletteOpen: (open) => set({ commandPaletteOpen: open }),
      toggleCommandPalette: () =>
        set({ commandPaletteOpen: !get().commandPaletteOpen }),
    }),
    {
      name: "nexora-super-admin-chrome",
      partialize: (state) => ({
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
        tenantContextId: state.tenantContextId,
      }),
    },
  ),
);
