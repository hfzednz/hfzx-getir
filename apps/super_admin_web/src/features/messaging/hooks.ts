"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchMessagingSnapshot } from "./api";

export const messagingKeys = {
  all: ["platform-messaging"] as const,
  snapshot: () => [...messagingKeys.all, "snapshot"] as const,
};

export function useMessagingSnapshot() {
  return useQuery({
    queryKey: messagingKeys.snapshot(),
    queryFn: fetchMessagingSnapshot,
  });
}
