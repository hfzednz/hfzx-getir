"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchGatewaySnapshot } from "./api";

export const gatewayKeys = {
  all: ["platform-gateway"] as const,
  snapshot: () => [...gatewayKeys.all, "snapshot"] as const,
};

export function useGatewaySnapshot() {
  return useQuery({
    queryKey: gatewayKeys.snapshot(),
    queryFn: fetchGatewaySnapshot,
  });
}
