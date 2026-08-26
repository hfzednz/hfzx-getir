export function subscribeOrderSse(
  orderId: string,
  onEvent: (payload: Record<string, unknown>) => void,
  onDisconnect?: () => void,
): () => void {
  if (typeof EventSource === "undefined") {
    onDisconnect?.();
    return () => undefined;
  }
  const base = (process.env.NEXT_PUBLIC_REALTIME_URL ?? "http://localhost:8115").replace(/\/$/, "");
  const topic = encodeURIComponent(`order:${orderId}`);
  const es = new EventSource(`${base}/v1/realtime/sse?topic=${topic}`);
  es.onmessage = (ev) => {
    try {
      onEvent(JSON.parse(ev.data) as Record<string, unknown>);
    } catch {
      /* ignore malformed */
    }
  };
  es.onerror = () => {
    es.close();
    onDisconnect?.();
  };
  return () => es.close();
}
