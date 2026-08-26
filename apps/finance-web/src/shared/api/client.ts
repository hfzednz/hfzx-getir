"use client";
import { createSessionStore, createApiClient, serviceUrl, tenantId } from "@nexora/web-core";
export const useSession = createSessionStore("nexora-finance-web-session");
export function financeApi() {
  const session = useSession.getState().session;
  return createApiClient({
    baseUrl: serviceUrl("finance"),
    tenantId: tenantId(),
    getToken: () => session?.accessToken ?? null,
    getUserId: () => session?.principalId ?? null,
  });
}
