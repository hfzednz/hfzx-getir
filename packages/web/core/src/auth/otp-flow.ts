import { createApiClient } from "../api/create-client";
import { bffUrl, identityUrl, tenantId } from "../config/env";
import type { WebSession } from "./types";

export type OtpChannel = "customer-bff" | "identity";

export async function startOtp(
  phone: string,
  channel: OtpChannel = "identity",
): Promise<{ challengeId: string }> {
  const tid = tenantId();
  if (channel === "customer-bff") {
    const api = createApiClient({ baseUrl: bffUrl("customer"), tenantId: tid });
    return api.request("/v1/customer/auth/otp/start", {
      method: "POST",
      body: { phone },
    });
  }
  const api = createApiClient({ baseUrl: identityUrl(), tenantId: tid });
  return api.request("/v1/identity/auth/otp/start", {
    method: "POST",
    body: { phone, tenantId: tid },
  });
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
    const api = createApiClient({ baseUrl: bffUrl("customer"), tenantId: tid });
    const res = await api.request<Record<string, unknown>>(
      "/v1/customer/auth/otp/verify",
      { method: "POST", body: { challengeId, code } },
    );
    return sessionFromResponse(res, phone, expectedRoles.length ? expectedRoles : ["customer"]);
  }
  const api = createApiClient({ baseUrl: identityUrl(), tenantId: tid });
  const res = await api.request<Record<string, unknown>>(
    "/v1/identity/auth/otp/verify",
    { method: "POST", body: { challengeId, code, tenantId: tid },
    },
  );
  return sessionFromResponse(res, phone, expectedRoles);
}

function sessionFromResponse(
  res: Record<string, unknown>,
  phone: string,
  fallbackRoles: string[],
): WebSession {
  return {
    accessToken: String(res.accessToken ?? res.AccessToken ?? ""),
    refreshToken: res.refreshToken ? String(res.refreshToken) : undefined,
    principalId: String(res.principalId ?? res.PrincipalID ?? ""),
    roles: Array.isArray(res.roles)
      ? (res.roles as string[])
      : fallbackRoles,
    phone,
    expiresAt: res.expiresIn
      ? new Date(Date.now() + Number(res.expiresIn) * 1000).toISOString()
      : undefined,
  };
}
