import { createApiClient } from "../api/create-client";
import { identityUrl, tenantId } from "../config/env";
import type { WebSession } from "./types";

export type OtpChannel = "customer-bff" | "identity";

function challengeIdFrom(res: unknown): string {
  if (!res || typeof res !== "object") return "";
  const body = res as Record<string, unknown>;
  const raw = body.challengeId ?? body.ChallengeID ?? body.challenge_id;
  return raw == null ? "" : String(raw).trim();
}

export async function startOtp(
  phone: string,
  channel: OtpChannel = "identity",
): Promise<{ challengeId: string }> {
  const tid = tenantId();
  const trimmed = phone.trim();
  if (!trimmed) {
    throw new Error("Enter a phone number to receive a verification code.");
  }
  let res: unknown;
  if (channel === "customer-bff") {
    // Always same-origin so the Next rewrite reaches Codespace BFF.
    // Never use http://localhost:8111 — that is the phone itself on iOS.
    const api = createApiClient({
      baseUrl: "",
      tenantId: tid,
    });
    res = await api.request("/v1/customer/auth/otp/start", {
      method: "POST",
      body: { phone: trimmed },
    });
  } else {
    const api = createApiClient({ baseUrl: identityUrl(), tenantId: tid });
    res = await api.request("/v1/identity/auth/otp/start", {
      method: "POST",
      body: { phone: trimmed, tenantId: tid },
    });
  }
  const challengeId = challengeIdFrom(res);
  if (!challengeId) {
    throw new Error("Could not start verification. Please try again.");
  }
  return { challengeId };
}

export async function verifyOtp(
  challengeId: string,
  code: string,
  phone: string,
  channel: OtpChannel = "identity",
  expectedRoles: string[] = [],
): Promise<WebSession> {
  const tid = tenantId();
  if (channel === "customer-bff") {
    const api = createApiClient({
      baseUrl: "",
      tenantId: tid,
    });
    const res = await api.request<Record<string, unknown>>(
      "/v1/customer/auth/otp/verify",
      { method: "POST", body: { challengeId, code } },
    );
    return assertRoles(sessionFromResponse(res, phone), expectedRoles);
  }
  const api = createApiClient({ baseUrl: identityUrl(), tenantId: tid });
  const res = await api.request<Record<string, unknown>>(
    "/v1/identity/auth/otp/verify",
    { method: "POST", body: { challengeId, code } },
  );
  return assertRoles(sessionFromResponse(res, phone), expectedRoles);
}

export class RoleNotAllowedError extends Error {
  constructor() {
    super("This account is not allowed to use this application.");
    this.name = "RoleNotAllowedError";
  }
}

function assertRoles(session: WebSession, expectedRoles: string[]): WebSession {
  if (expectedRoles.length === 0) return session;
  if (session.roles.some((role) => expectedRoles.includes(role))) return session;
  throw new RoleNotAllowedError();
}

function rolesFromAccessToken(token: string): string[] {
  try {
    const part = token.split(".")[1];
    if (!part) return [];
    const json = atob(part.replace(/-/g, "+").replace(/_/g, "/"));
    const payload = JSON.parse(json) as { roles?: unknown };
    return Array.isArray(payload.roles) ? payload.roles.map(String) : [];
  } catch {
    return [];
  }
}

function sessionFromResponse(
  res: Record<string, unknown>,
  phone: string,
): WebSession {
  const accessToken = String(res.accessToken ?? res.AccessToken ?? "");
  const principalId = String(
    res.principalId ??
      res.PrincipalID ??
      res.customerId ??
      res.CustomerID ??
      "",
  );
  const expiresRaw = res.expiresIn ?? res.ExpiresIn;
  const jwtRoles = rolesFromAccessToken(accessToken);
  const bodyRoles = Array.isArray(res.roles)
    ? (res.roles as string[])
    : Array.isArray(res.Roles)
      ? (res.Roles as string[])
      : [];
  return {
    accessToken,
    refreshToken: res.refreshToken
      ? String(res.refreshToken)
      : res.RefreshToken
        ? String(res.RefreshToken)
        : undefined,
    principalId,
    roles: jwtRoles.length ? jwtRoles : bodyRoles,
    phone,
    expiresAt: expiresRaw
      ? new Date(Date.now() + Number(expiresRaw) * 1000).toISOString()
      : undefined,
  };
}
