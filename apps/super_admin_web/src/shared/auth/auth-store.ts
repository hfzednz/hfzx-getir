"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { WebSession } from "@nexora/web-core";
import type { PlatformSession } from "@/shared/auth/session";
import { platformSessionFromOtp } from "@/shared/auth/otp-bridge";

interface AuthState {
  session: PlatformSession | null;
  setSessionFromOtp: (web: WebSession, phone: string) => void;
  logout: () => void;
  setMfaVerified: (verified: boolean) => void;
  setWebauthnVerified: (verified: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      session: null,
      setSessionFromOtp: (web, phone) => {
        if (!web.accessToken) {
          throw new Error("No access token in auth response");
        }
        if (!web.roles.includes("super_admin")) {
          throw new Error("Super admin role required");
        }
        set({ session: platformSessionFromOtp(web, phone) });
      },
      logout: () => set({ session: null }),
      setMfaVerified: (verified) =>
        set((state) =>
          state.session
            ? { session: { ...state.session, mfaVerified: verified } }
            : state,
        ),
      setWebauthnVerified: (verified) =>
        set((state) =>
          state.session
            ? { session: { ...state.session, webauthnVerified: verified } }
            : state,
        ),
    }),
    {
      name: "nexora-super-admin-session",
      partialize: (state) => ({ session: state.session }),
    },
  ),
);
