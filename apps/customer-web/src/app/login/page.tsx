"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { customerApi, useSession } from "@/shared/api/client";

export default function LoginPage() {
  const router = useRouter();
  const setSession = useSession((s) => s.setSession);
  const [phone, setPhone] = useState("+905551112233");
  const [challengeId, setChallengeId] = useState("");
  const [code, setCode] = useState("");
  const [step, setStep] = useState<"phone" | "otp">("phone");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function startOtp(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const api = customerApi();
      const res = await api.request<{ challengeId: string }>(
        "/v1/customer/auth/otp/start",
        { method: "POST", body: { phone } },
      );
      setChallengeId(res.challengeId);
      setStep("otp");
    } catch (err) {
      setError(err instanceof Error ? err.message : "OTP start failed");
    } finally {
      setLoading(false);
    }
  }

  async function verifyOtp(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const api = customerApi();
      const res = await api.request<{
        accessToken?: string;
        AccessToken?: string;
        principalId?: string;
        PrincipalID?: string;
        roles?: string[];
      }>("/v1/customer/auth/otp/verify", {
        method: "POST",
        body: { challengeId, code },
      });
      setSession({
        accessToken: res.accessToken ?? res.AccessToken ?? "session",
        principalId:
          res.principalId ??
          res.PrincipalID ??
          "22222222-2222-2222-2222-222222222222",
        roles: res.roles ?? ["customer"],
        phone,
      });
      router.push("/home");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid OTP");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-dvh flex-col justify-center px-6 py-10">
      <h1 className="mb-2 text-2xl font-bold text-[var(--nx-brand)]">NEXORA</h1>
      <p className="mb-8 text-sm text-neutral-600">Sign in with your phone number</p>
      {step === "phone" ? (
        <form onSubmit={startOtp} className="space-y-4">
          <label className="block text-sm font-medium">
            Phone
            <input
              type="tel"
              autoComplete="tel"
              className="mt-1 w-full rounded-lg border px-3 py-3 text-base"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              required
            />
          </label>
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-lg bg-[var(--nx-brand)] py-3 font-semibold text-white disabled:opacity-60"
          >
            {loading ? "Sending…" : "Send OTP"}
          </button>
        </form>
      ) : (
        <form onSubmit={verifyOtp} className="space-y-4">
          <label className="block text-sm font-medium">
            OTP code
            <input
              inputMode="numeric"
              autoComplete="one-time-code"
              className="mt-1 w-full rounded-lg border px-3 py-3 text-base tracking-widest"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              required
            />
          </label>
          <p className="text-xs text-neutral-500">
            Staging: read OTP from identity-service logs when OTP_DEV_MODE=true.
          </p>
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-lg bg-[var(--nx-brand)] py-3 font-semibold text-white disabled:opacity-60"
          >
            {loading ? "Verifying…" : "Verify & continue"}
          </button>
        </form>
      )}
      {error ? (
        <p className="mt-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
