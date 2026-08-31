"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { RouteGuard } from "@nexora/web-core";
import { warehouseApi, useSession } from "@/shared/api/client";

type Task = {
  id?: string;
  taskId?: string;
  orderId?: string;
  order_id?: string;
  status?: string;
  orderStatus?: string;
};

function taskIdOf(t: Task): string {
  return (t.id ?? t.taskId ?? t.orderId ?? t.order_id ?? "").toString();
}

function unwrapTasks(data: unknown): Task[] {
  if (Array.isArray(data)) return data as Task[];
  if (data && typeof data === "object" && "items" in data) {
    const items = (data as { items?: unknown }).items;
    if (Array.isArray(items)) return items as Task[];
  }
  return [];
}

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [queue, setQueue] = useState<Task[]>([]);
  const [taskId, setTaskId] = useState(
    () => process.env.NEXT_PUBLIC_WAREHOUSE_TASK_ID?.trim() || "",
  );
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function loadQueue() {
    try {
      const raw = await warehouseApi().request<unknown>("/v1/warehouse/tasks");
      const items = unwrapTasks(raw);
      setQueue(items);
      setTaskId((current) => current || (items[0] ? taskIdOf(items[0]) : ""));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load the pick queue");
    }
  }

  useEffect(() => {
    const q = new URLSearchParams(window.location.search).get("orderId")?.trim();
    if (q) setTaskId(q);
    void loadQueue();
  }, []);

  async function advance(step: "pick" | "pack" | "ready") {
    const id = taskId.trim();
    if (!id) {
      setError("Select an order from the queue.");
      return;
    }
    setError("");
    setBusy(true);
    try {
      const res = await warehouseApi().request<Record<string, unknown>>(
        `/v1/warehouse/tasks/${id}/${step}`,
        { method: "POST", body: {} },
      );
      setStatus(typeof res.status === "string" ? res.status : step);
      await loadQueue();
    } catch (e) {
      setStatus("");
      setError(e instanceof Error ? e.message : "Transition failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <RouteGuard
      session={session}
      allow={["picker", "packer", "dispatcher", "admin", "super_admin"]}
      onDeny={logout}
    >
      <div className="mx-auto max-w-lg space-y-4 p-4">
        <div className="flex items-center justify-between gap-3">
          <h1 className="text-xl font-semibold">Fulfillment</h1>
          <button
            type="button"
            className="rounded-lg px-3 text-sm"
            style={{ minHeight: 44 }}
            onClick={() => {
              logout();
              router.push("/login");
            }}
          >
            Logout
          </button>
        </div>
        <section className="space-y-2">
          <h2 className="text-sm font-medium">Incoming orders</h2>
          {queue.length === 0 ? (
            <p className="text-sm text-neutral-600">No orders waiting in the warehouse queue.</p>
          ) : (
            <ul className="divide-y rounded-xl border">
              {queue.map((task) => {
                const id = taskIdOf(task);
                return (
                  <li key={id}>
                    <button
                      type="button"
                      className={`flex w-full items-center justify-between gap-3 p-4 text-left ${
                        taskId === id ? "bg-violet-50" : "hover:bg-neutral-50"
                      }`}
                      style={{ minHeight: 44 }}
                      onClick={() => setTaskId(id)}
                    >
                      <span className="truncate text-sm font-medium">{id.slice(0, 8)}…</span>
                      <span className="text-xs text-neutral-600">
                        {task.orderStatus ?? task.status ?? "queued"}
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </section>
        <p className="text-sm text-neutral-600" role="status" aria-live="polite">
          {busy ? "Working…" : status ? `Status: ${status}` : "Select an order, then pick, pack, or mark ready."}
        </p>
        {error ? (
          <p id="warehouse-error" className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
        <div className="grid gap-2">
          <button
            type="button"
            className="rounded-lg border py-3 disabled:opacity-50"
            style={{ minHeight: 44 }}
            onClick={() => advance("pick")}
            disabled={busy || !taskId.trim()}
          >
            Start pick
          </button>
          <button
            type="button"
            className="rounded-lg border py-3 disabled:opacity-50"
            style={{ minHeight: 44 }}
            onClick={() => advance("pack")}
            disabled={busy || !taskId.trim()}
          >
            Complete pack
          </button>
          <button
            type="button"
            className="rounded-lg bg-violet-600 py-3 text-white disabled:opacity-50"
            style={{ minHeight: 44 }}
            onClick={() => advance("ready")}
            disabled={busy || !taskId.trim()}
          >
            Mark ready
          </button>
        </div>
      </div>
    </RouteGuard>
  );
}
