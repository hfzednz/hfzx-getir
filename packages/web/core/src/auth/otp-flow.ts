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
  const accessToken = String(res.accessToken ?? res.AccessToken ?? "");
  const principalId = String(
    res.principalId ??
      res.PrincipalID ??
      res.customerId ??
      res.CustomerID ??
      "",
  );
  const expiresRaw = res.expiresIn ?? res.ExpiresIn;
  return {
    accessToken,
    refreshToken: res.refreshToken
      ? String(res.refreshToken)
      : res.RefreshToken
        ? String(res.RefreshToken)
        : undefined,
    principalId,
    roles: Array.isArray(res.roles)
      ? (res.roles as string[])
      : Array.isArray(res.Roles)
        ? (res.Roles as string[])
        : fallbackRoles,
    phone,
    expiresAt: expiresRaw
      ? new Date(Date.now() + Number(expiresRaw) * 1000).toISOString()
      : undefined,
  };
}
