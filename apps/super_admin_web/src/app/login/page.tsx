"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { Button, Checkbox, Input, Select } from "@nexora/ui";
import { useAuthStore } from "@/shared/auth/auth-store";
import type { Role } from "@/shared/permissions/platform-permissions";

const DEMO_ROLES: { value: Role; label: string }[] = [
  { value: "platform_owner", label: "Platform Owner" },
  { value: "platform_security", label: "Platform Security" },
  { value: "platform_sre", label: "Platform SRE" },
  { value: "platform_finops", label: "Platform FinOps" },
  { value: "platform_compliance", label: "Platform Compliance" },
  { value: "platform_viewer", label: "Platform Viewer" },
];

export default function LoginPage() {
  const router = useRouter();
  const login = useAuthStore((s) => s.login);
  const session = useAuthStore((s) => s.session);
  const [email, setEmail] = useState("platform@nexora.local");
  const [password, setPassword] = useState("demo");
  const [role, setRole] = useState<Role>("platform_owner");
  const [mfaVerified, setMfaVerified] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (session) {
      router.replace("/dashboard");
    }
  }, [session, router]);

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      login({ email, password, role, mfaVerified });
      router.replace("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center px-[var(--nx-space-4)] bg-[var(--nx-bg-canvas)]">
      <div className="w-full max-w-sm">
        <div className="mb-[var(--nx-space-6)] text-center">
          <div className="inline-flex size-10 items-center justify-center rounded-[var(--nx-radius-sm)] bg-[var(--nx-brand-600)] text-[var(--nx-text-on-brand)] text-[13px] font-bold mb-[var(--nx-space-3)]">
            NX
          </div>
          <h1 className="m-0 font-[family-name:var(--nx-font-display)] text-[var(--nx-font-size-page)] font-semibold tracking-[-0.02em] text-[var(--nx-text-primary)]">
            NEXORA Super Admin
          </h1>
          <p className="m-0 mt-[var(--nx-space-1)] text-[12px] text-[var(--nx-text-secondary)]">
            Global platform control · demo sign-in
          </p>
        </div>

        <form
          onSubmit={onSubmit}
          className="flex flex-col gap-[var(--nx-space-3)] bg-[var(--nx-bg-surface)] border border-[var(--nx-border-subtle)] rounded-[var(--nx-radius-sm)] p-[var(--nx-space-5)]"
        >
          <label className="flex flex-col gap-[var(--nx-space-1)]">
            <span className="text-[12px] font-medium text-[var(--nx-text-secondary)]">
              Email
            </span>
            <Input
              type="email"
              autoComplete="username"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </label>

          <label className="flex flex-col gap-[var(--nx-space-1)]">
            <span className="text-[12px] font-medium text-[var(--nx-text-secondary)]">
              Password
            </span>
            <Input
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </label>

          <label className="flex flex-col gap-[var(--nx-space-1)]">
            <span className="text-[12px] font-medium text-[var(--nx-text-secondary)]">
              Platform role
            </span>
            <Select
              value={role}
              onChange={(e) => setRole(e.target.value as Role)}
            >
              {DEMO_ROLES.map((r) => (
                <option key={r.value} value={r.value}>
                  {r.label}
                </option>
              ))}
            </Select>
          </label>

          <Checkbox
            label="MFA verified (mock)"
            checked={mfaVerified}
            onChange={(e) => setMfaVerified(e.target.checked)}
          />

          {error ? (
            <p className="m-0 text-[12px] text-[var(--nx-danger)]">{error}</p>
          ) : null}

          <Button
            type="submit"
            loading={submitting}
            className="w-full mt-[var(--nx-space-1)]"
          >
            Sign in
          </Button>
        </form>
      </div>
    </div>
  );
}
