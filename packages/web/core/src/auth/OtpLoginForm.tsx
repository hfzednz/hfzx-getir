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
  defaultPhone = "",
  onSuccess,
  otpHint,
}: OtpLoginFormProps) {
  const [phone, setPhone] = useState(defaultPhone);
  const [challengeId, setChallengeId] = useState("");
  const [code, setCode] = useState("");
  const [step, setStep] = useState<"phone" | "otp">("phone");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function onStart(e?: React.FormEvent) {
    e?.preventDefault();
    e?.stopPropagation();
    const submitted = phone.trim();
    if (!submitted) {
      setError("Enter a phone number to receive a verification code.");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const res = await startOtp(submitted, channel);
      if (!res.challengeId) {
        throw new Error("Could not start verification. Please try again.");
      }
      setPhone(submitted);
      setChallengeId(res.challengeId);
      setStep("otp");
    } catch (err) {
      setError(
        err instanceof Error && err.message
          ? err.message
          : "Could not send the verification code. Please try again.",
      );
    } finally {
      setLoading(false);
    }
  }

  async function onVerify(e?: React.FormEvent) {
    e?.preventDefault();
    e?.stopPropagation();
    if (!challengeId) {
      setError("Verification expired. Please request a new code.");
      setStep("phone");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const session = await verifyOtp(challengeId, code, phone, channel, expectedRoles);
      if (!session.accessToken) {
        throw new Error("No access token in auth response");
      }
      onSuccess(session);
    } catch (err) {
      setError(err instanceof Error && err.message ? err.message : "Invalid OTP");
    } finally {
      setLoading(false);
    }
  }

  const errorAlert = error ? (
    <p
      id="otp-form-error"
      className="rounded-lg border border-red-200 bg-red-50 px-3 py-3 text-sm text-red-700"
      role="alert"
    >
      {error}
    </p>
  ) : null;

  return (
    <div className="flex min-h-dvh flex-col justify-center px-6 py-10">
      <h1 className="mb-2 text-2xl font-bold text-violet-700">{title}</h1>
      {step === "phone" ? (
        <form onSubmit={(e) => void onStart(e)} noValidate className="space-y-4">
          <label className="block text-sm font-medium">
            Phone
            <input
              type="tel"
              autoComplete="tel"
              className="mt-1 w-full rounded-lg border px-3 py-3"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              aria-invalid={error ? true : undefined}
              aria-describedby={error ? "otp-form-error" : undefined}
              required
            />
          </label>
          {errorAlert}
          <button
            type="button"
            disabled={loading}
            onClick={() => void onStart()}
            className="w-full rounded-lg bg-violet-600 px-4 font-semibold text-white disabled:opacity-60"
            style={{ minHeight: 44, paddingTop: 12, paddingBottom: 12 }}
          >
            {loading ? "Sending…" : "Send OTP"}
          </button>
        </form>
      ) : (
        <form onSubmit={(e) => void onVerify(e)} noValidate className="space-y-4">
          <label className="block text-sm font-medium">
            OTP code
            <input
              inputMode="numeric"
              autoComplete="one-time-code"
              className="mt-1 w-full rounded-lg border px-3 py-3 tracking-widest"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              aria-invalid={error ? true : undefined}
              aria-describedby={error ? "otp-form-error" : undefined}
              required
            />
          </label>
          {otpHint ? <p className="text-xs text-neutral-500">{otpHint}</p> : null}
          {errorAlert}
          <button
            type="button"
            disabled={loading}
            onClick={() => void onVerify()}
            className="w-full rounded-lg bg-violet-600 px-4 font-semibold text-white disabled:opacity-60"
            style={{ minHeight: 44, paddingTop: 12, paddingBottom: 12 }}
          >
            {loading ? "Verifying…" : "Verify & continue"}
          </button>
        </form>
      )}
    </div>
  );
}
