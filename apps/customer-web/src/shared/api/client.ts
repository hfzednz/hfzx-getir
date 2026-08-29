"use client";

import { createSessionStore } from "@nexora/web-core";
import { createApiClient, bffUrl, tenantId } from "@nexora/web-core";

export const useSession = createSessionStore("nexora-customer-session");

export function customerApi() {
  const session = useSession.getState().session;
  return createApiClient({
    // Same-origin /v1 is rewritten by Next to the Codespace BFF (not localhost / LAN IP).
    baseUrl: typeof window === "undefined" ? bffUrl("customer") : "",
    tenantId: tenantId(),
    getToken: () => session?.accessToken ?? null,
    getUserId: () => session?.principalId ?? null,
  });
}
