"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface CartLine {
  sku: string;
  name: string;
  qty: number;
  unitMinor: number;
}

interface CartState {
  cartId: string | null;
  lines: CartLine[];
  setCartId: (id: string) => void;
  addLine: (line: CartLine) => void;
  updateQty: (sku: string, qty: number) => void;
  removeLine: (sku: string) => void;
  clear: () => void;
  totalMinor: () => number;
}

export const useCart = create<CartState>()(
  persist(
    (set, get) => ({
      cartId: null,
      lines: [],
      setCartId: (id) => set({ cartId: id }),
      addLine: (line) =>
        set((s) => {
          const existing = s.lines.find((l) => l.sku === line.sku);
          if (existing) {
            return {
              lines: s.lines.map((l) =>
                l.sku === line.sku ? { ...l, qty: l.qty + line.qty } : l,
              ),
            };
          }
          return { lines: [...s.lines, line] };
        }),
      updateQty: (sku, qty) =>
        set((s) => ({
          lines: s.lines
            .map((l) => (l.sku === sku ? { ...l, qty } : l))
            .filter((l) => l.qty > 0),
        })),
      removeLine: (sku) =>
        set((s) => ({ lines: s.lines.filter((l) => l.sku !== sku) })),
      clear: () => set({ cartId: null, lines: [] }),
      totalMinor: () =>
        get().lines.reduce((sum, l) => sum + l.qty * l.unitMinor, 0),
    }),
    { name: "nexora-customer-cart" },
  ),
);
