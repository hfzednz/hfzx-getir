"use client";
import { createSessionStore, createApiClient, serviceUrl, tenantId } from "@nexora/web-core";
export const useSession = createSessionStore("nexora-supplier-web-session");
export function supplierApi() {
  const session = useSession.getState().session;
  return createApiClient({
    baseUrl: typeof window === "undefined" ? serviceUrl("supplier") : "",
    tenantId: tenantId(),
    getToken: () => session?.accessToken ?? null,
    getUserId: () => session?.principalId ?? null,
  });
}
