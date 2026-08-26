"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { WebSession } from "./types";

interface SessionState {
  session: WebSession | null;
  setSession: (session: WebSession | null) => void;
  logout: () => void;
}

export function createSessionStore(storageKey: string) {
  return create<SessionState>()(
    persist(
      (set) => ({
        session: null,
        setSession: (session) => set({ session }),
        logout: () => set({ session: null }),
      }),
      { name: storageKey },
    ),
  );
}
