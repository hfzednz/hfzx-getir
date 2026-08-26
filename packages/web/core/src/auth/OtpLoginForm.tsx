"use client";

import { useState } from "react";
import { startOtp, verifyOtp, type OtpChannel } from "./otp-flow";
import type { WebSession } from "./types";

export interface OtpLoginFormProps {
  title: string;
  channel?: OtpChannel;
  expectedRoles?: string[];
  defaultPhone?: string;
  onSuccess: (session: WebSession) => void;
  otpHint?: string;
}

export function OtpLoginForm({
  title,
  channel = "identity",
  expectedRoles = [],
  defaultPhone = "+905551112233",
  onSuccess,
  otpHint = "Staging: read OTP from identity-service logs when OTP_DEV_MODE=true.",
}: OtpLoginFormProps) {
  const [phone, setPhone] = useState(defaultPhone);
  const [challengeId, setChallengeId] = useState("");
  const [code, setCode] = useState("");
  const [step, setStep] = useState<"phone" | "otp">("phone");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function onStart(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await startOtp(phone, channel);
      setChallengeId(res.challengeId);
      setStep("otp");
    } catch (err) {
      setError(err instanceof Error ? err.message : "OTP start failed");
    } finally {
      setLoading(false);
    }
  }

  async function onVerify(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const session = await verifyOtp(challengeId, code, phone, channel, expectedRoles);
      if (!session.accessToken) {
        throw new Error("No access token in auth response");
      }
      onSuccess(session);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid OTP");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-dvh flex-col justify-center px-6 py-10">
      <h1 className="mb-2 text-2xl font-bold text-violet-700">{title}</h1>
      {step === "phone" ? (
        <form onSubmit={onStart} className="space-y-4">
          <label className="block text-sm font-medium">
            Phone
            <input
              type="tel"
              autoComplete="tel"
              className="mt-1 w-full rounded-lg border px-3 py-3"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              required
            />
          </label>
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-lg bg-violet-600 py-3 font-semibold text-white disabled:opacity-60"
          >
            {loading ? "Sending…" : "Send OTP"}
          </button>
        </form>
      ) : (
        <form onSubmit={onVerify} className="space-y-4">
          <label className="block text-sm font-medium">
            OTP code
            <input
              inputMode="numeric"
              autoComplete="one-time-code"
              className="mt-1 w-full rounded-lg border px-3 py-3 tracking-widest"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              required
            />
          </label>
          <p className="text-xs text-neutral-500">{otpHint}</p>
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-lg bg-violet-600 py-3 font-semibold text-white disabled:opacity-60"
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
