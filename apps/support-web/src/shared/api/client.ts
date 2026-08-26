"use client";
import { createSessionStore, createApiClient, bffUrl, tenantId } from "@nexora/web-core";
export const useSession = createSessionStore("nexora-support-web-session");
export function supportApi() {
  const session = useSession.getState().session;
  return createApiClient({
    baseUrl: bffUrl("admin"),
    tenantId: tenantId(),
    getToken: () => session?.accessToken ?? null,
    getUserId: () => session?.principalId ?? null,
  });
}
