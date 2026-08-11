"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchNotificationsSnapshot } from "./api";

export const notificationsKeys = {
  all: ["notifications"] as const,
  snapshot: () => [...notificationsKeys.all, "snapshot"] as const,
};

export function useNotificationsSnapshot() {
  return useQuery({
    queryKey: notificationsKeys.snapshot(),
    queryFn: () => fetchNotificationsSnapshot(),
  });
}
