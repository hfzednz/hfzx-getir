"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import { buildMockLiveSnapshot, fetchLiveSnapshot } from "./api";
import type { LiveOpsSnapshot } from "./types";

export const liveKeys = {
  all: ["live"] as const,
  snapshot: (cityId: string | null) =>
    [...liveKeys.all, "snapshot", cityId] as const,
};

const POLL_MS = 8_000;

/**
 * WebSocket stub for ops realtime.
 * Attempts WS connection; on failure falls back to polling mock snapshots.
 */
export function useOpsSocket(enabled = true) {
  const cityId = useChromeStore((s) => s.cityId);
  const queryClient = useQueryClient();
  const [connection, setConnection] =
    useState<LiveOpsSnapshot["connection"]>("polling");
  const wsRef = useRef<WebSocket | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const pushSnapshot = useCallback(
    (snapshot: LiveOpsSnapshot) => {
      queryClient.setQueryData(liveKeys.snapshot(cityId), snapshot);
    },
    [cityId, queryClient],
  );

  const startPolling = useCallback(() => {
    if (pollRef.current) return;
    setConnection("polling");
    const tick = () => {
      pushSnapshot(buildMockLiveSnapshot(cityId, "polling"));
    };
    tick();
    pollRef.current = setInterval(tick, POLL_MS);
  }, [cityId, pushSnapshot]);

  const stopPolling = useCallback(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  }, []);

  useEffect(() => {
    if (!enabled) return;

    // Default to the same origin the rest of the console uses, so the socket follows
    // the Next rewrite to bff-admin instead of a stale hardcoded port.
    const sameOrigin =
      typeof window === "undefined"
        ? ""
        : `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/ws/admin/ops`;
    const wsBase = process.env.NEXT_PUBLIC_BFF_ADMIN_WS_URL ?? sameOrigin;
    if (!wsBase) return;
    const url = cityId
      ? `${wsBase}?cityId=${encodeURIComponent(cityId)}`
      : wsBase;

    let closed = false;

    try {
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        if (closed) return;
        stopPolling();
        setConnection("live");
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(String(event.data)) as LiveOpsSnapshot;
          pushSnapshot({ ...data, connection: "live" });
        } catch {
          // ignore malformed frames
        }
      };

      ws.onerror = () => {
        ws.close();
      };

      ws.onclose = () => {
        wsRef.current = null;
        if (!closed) startPolling();
      };
    } catch {
      startPolling();
    }

    // Most local envs have no WS — fall back quickly.
    const fallbackTimer = setTimeout(() => {
      if (wsRef.current?.readyState !== WebSocket.OPEN) {
        try {
          wsRef.current?.close();
        } catch {
          /* ignore */
        }
        startPolling();
      }
    }, 1200);

    return () => {
      closed = true;
      clearTimeout(fallbackTimer);
      stopPolling();
      try {
        wsRef.current?.close();
      } catch {
        /* ignore */
      }
      wsRef.current = null;
    };
  }, [cityId, enabled, pushSnapshot, startPolling, stopPolling]);

  return { connection };
}

export function useLiveSnapshot() {
  const cityId = useChromeStore((s) => s.cityId);
  const { connection } = useOpsSocket(true);

  const query = useQuery({
    queryKey: liveKeys.snapshot(cityId),
    queryFn: () => fetchLiveSnapshot(cityId),
    refetchInterval: false,
    staleTime: POLL_MS,
  });

  return {
    ...query,
    connection: query.data?.connection ?? connection,
  };
}
