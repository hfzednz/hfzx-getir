"use client";
import { createSessionStore, createApiClient, bffUrl, tenantId } from "@nexora/web-core";
export const useSession = createSessionStore("nexora-courier-web-session");
export function courierApi() {
  const session = useSession.getState().session;
  return createApiClient({
    baseUrl: bffUrl("courier"),
    tenantId: tenantId(),
    getToken: () => session?.accessToken ?? null,
    getUserId: () => session?.principalId ?? null,
  });
}
