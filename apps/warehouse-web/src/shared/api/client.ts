"use client";
import { createSessionStore, createApiClient, bffUrl, tenantId } from "@nexora/web-core";
export const useSession = createSessionStore("nexora-warehouse-web-session");
export function warehouseApi() {
  const session = useSession.getState().session;
  return createApiClient({
    baseUrl: typeof window === "undefined" ? bffUrl("warehouse") : "",
    tenantId: tenantId(),
    getToken: () => session?.accessToken ?? null,
    getUserId: () => session?.principalId ?? null,
  });
}
