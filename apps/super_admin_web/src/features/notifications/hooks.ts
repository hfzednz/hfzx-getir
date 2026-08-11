"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchNotifications, setProviderStatus } from "./api";
import type { NotificationProvider } from "./types";

export const notificationKeys = {
  all: ["platform-notifications"] as const,
  snapshot: () => [...notificationKeys.all, "snapshot"] as const,
};

export function useNotifications() {
  return useQuery({
    queryKey: notificationKeys.snapshot(),
    queryFn: fetchNotifications,
  });
}

export function useSetProviderStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: {
      providerId: string;
      status: NotificationProvider["status"];
    }) => setProviderStatus(input),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: notificationKeys.all });
    },
  });
}
