export function subscribeOrderSse(
  orderId: string,
  onEvent: (payload: Record<string, unknown>) => void,
  onDisconnect?: () => void,
  onOpen?: () => void,
  getTicket?: () => Promise<string>,
): () => void {
  if (typeof EventSource === "undefined") {
    onDisconnect?.();
    return () => undefined;
  }
  const envUrl = process.env.NEXT_PUBLIC_REALTIME_URL;
  const base =
    envUrl === undefined
      ? typeof window !== "undefined"
        ? ""
        : "http://localhost:8115"
      : envUrl.replace(/\/$/, "");
  const topic = encodeURIComponent(`order:${orderId}`);
  let closed = false;
  let es: EventSource | null = null;

  void (async () => {
    let ticket = "";
    if (getTicket) {
      try {
        ticket = await getTicket();
      } catch {
        if (!closed) onDisconnect?.();
        return;
      }
    }
    if (closed) return;
    const q = ticket
      ? `${base}/v1/realtime/sse?topic=${topic}&ticket=${encodeURIComponent(ticket)}`
      : `${base}/v1/realtime/sse?topic=${topic}`;
    es = new EventSource(q);
    es.onopen = () => {
      onOpen?.();
    };
    es.onmessage = (ev) => {
      try {
        onEvent(JSON.parse(ev.data) as Record<string, unknown>);
      } catch {
        /* ignore malformed */
      }
    };
    es.onerror = () => {
      if (es && es.readyState === EventSource.CLOSED) {
        onDisconnect?.();
      }
    };
  })();

  return () => {
    closed = true;
    es?.close();
  };
}
