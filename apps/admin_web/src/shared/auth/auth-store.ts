"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { AdminSession } from "@/shared/auth/session";
import {
  permissionsForRoles,
  type Role,
} from "@/shared/permissions/permissions";

export interface LoginInput {
  email: string;
  password: string;
  role?: Role;
  displayName?: string;
  cityIds?: string[];
}

interface AuthState {
  session: AdminSession | null;
  login: (input: LoginInput) => void;
  logout: () => void;
  setMfaVerified: (verified: boolean) => void;
}

function buildDemoSession(input: LoginInput): AdminSession {
  const role: Role = input.role ?? "admin";
  const roles: Role[] = [role];
  return {
    userId: `usr_${role}_demo`,
    email: input.email,
    displayName: input.displayName ?? input.email.split("@")[0] ?? "Operator",
    roles,
    permissions: permissionsForRoles(roles),
    cityIds: input.cityIds ?? ["city_ist", "city_ank"],
    mfaVerified: role === "super_admin" || role === "admin",
  };
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      session: null,
      login: (input) => {
        // Demo: password presence is enough; real OIDC replaces this.
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
    }),
    {
      name: "nexora-admin-session",
      partialize: (state) => ({ session: state.session }),
    },
  ),
);
