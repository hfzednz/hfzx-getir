"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { PlatformSession } from "@/shared/auth/session";
import {
  permissionsForRoles,
  type Role,
} from "@/shared/permissions/platform-permissions";

export interface LoginInput {
  email: string;
  password: string;
  role?: Role;
  displayName?: string;
  mfaVerified?: boolean;
  webauthnVerified?: boolean;
}

interface AuthState {
  session: PlatformSession | null;
  login: (input: LoginInput) => void;
  logout: () => void;
  setMfaVerified: (verified: boolean) => void;
  setWebauthnVerified: (verified: boolean) => void;
}

function buildDemoSession(input: LoginInput): PlatformSession {
  const role: Role = input.role ?? "platform_owner";
  const roles: Role[] = [role];
  return {
    userId: `usr_${role}_demo`,
    email: input.email,
    displayName: input.displayName ?? input.email.split("@")[0] ?? "Platform",
    roles,
    permissions: permissionsForRoles(roles),
    mfaVerified: input.mfaVerified ?? role === "platform_owner",
    webauthnVerified: input.webauthnVerified ?? false,
  };
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      session: null,
      login: (input) => {
        // Demo: password presence is enough; real OIDC + WebAuthn replaces this.
        if (!input.email || !input.password) {
          throw new Error("Email and password are required");
        }
        set({ session: buildDemoSession(input) });
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
