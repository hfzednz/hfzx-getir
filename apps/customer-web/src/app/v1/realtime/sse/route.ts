import { type NextRequest } from "next/server";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

const realtimeInternal = (process.env.REALTIME_INTERNAL || "http://127.0.0.1:8115").replace(
  /\/$/,
  "",
);

export async function GET(req: NextRequest) {
  const topic = req.nextUrl.searchParams.get("topic") ?? "";
  const ticket = req.nextUrl.searchParams.get("ticket") ?? "";
  if (!topic) {
    return new Response("topic required", { status: 400 });
  }
  const qs = new URLSearchParams({ topic });
  if (ticket) qs.set("ticket", ticket);
  const up = await fetch(
    `${realtimeInternal}/v1/realtime/sse?${qs.toString()}`,
    {
      headers: { Accept: "text/event-stream" },
      cache: "no-store",
    },
  );
  if (!up.ok || !up.body) {
    return new Response(await up.text(), { status: up.status });
  }
  return new Response(up.body, {
    status: 200,
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache, no-transform",
      Connection: "keep-alive",
      "X-Accel-Buffering": "no",
    },
  });
}
