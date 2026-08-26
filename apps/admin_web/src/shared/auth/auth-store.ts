"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { WebSession } from "@nexora/web-core";
import type { AdminSession } from "@/shared/auth/session";
import { adminSessionFromOtp } from "@/shared/auth/otp-bridge";

interface AuthState {
  session: AdminSession | null;
  setSessionFromOtp: (web: WebSession, phone: string) => void;
  logout: () => void;
  setMfaVerified: (verified: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      session: null,
      setSessionFromOtp: (web, phone) => {
        if (!web.accessToken) {
          throw new Error("No access token in auth response");
        }
        set({ session: adminSessionFromOtp(web, phone) });
      },
      logout: () => set({ session: null }),
      setMfaVerified: (verified) =>
        set((state) =>
          state.session
            ? { session: { ...state.session, mfaVerified: verified } }
            : state,
        ),
    }),
    {
      name: "nexora-admin-session",
      partialize: (state) => ({ session: state.session }),
    },
  ),
);
