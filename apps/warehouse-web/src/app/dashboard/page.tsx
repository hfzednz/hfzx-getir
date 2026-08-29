"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { warehouseApi, useSession } from "@/shared/api/client";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [taskId, setTaskId] = useState(
    () => process.env.NEXT_PUBLIC_WAREHOUSE_TASK_ID?.trim() || "",
  );
  const [status, setStatus] = useState("queued");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!session) router.replace("/login");
  }, [session, router]);

  useEffect(() => {
    const q = new URLSearchParams(window.location.search).get("orderId")?.trim();
    if (q) setTaskId(q);
  }, []);

  async function advance(step: "pick" | "pack" | "ready") {
    const id = taskId.trim();
    if (!id) {
      setError("Enter the real order id (warehouse task id).");
      return;
    }
    setError("");
    try {
      const res = await warehouseApi().request<Record<string, unknown>>(
        `/v1/warehouse/tasks/${id}/${step}`,
        { method: "POST", body: {} },
      );
      const next = typeof res.status === "string" ? res.status : step;
      setStatus(next);
    } catch (e) {
      setStatus("error");
      setError(e instanceof Error ? e.message : "Transition failed");
    }
  }

  return (
    <div className="mx-auto max-w-lg space-y-4 p-4">
      <div className="flex justify-between">
        <h1 className="text-xl font-semibold">Fulfillment</h1>
        <button
          type="button"
          className="text-sm"
          onClick={() => {
            logout();
            router.push("/login");
          }}
        >
          Logout
        </button>
      </div>
      <label className="block text-sm">
        Order / task id
        <input
          className="mt-1 w-full rounded-lg border px-3 py-3"
          value={taskId}
          onChange={(e) => setTaskId(e.target.value)}
          placeholder="uuid"
        />
      </label>
      <p className="text-sm text-neutral-600">Status: {status}</p>
      {error ? (
        <p className="text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
      <div className="grid gap-2">
        <button
          type="button"
          className="rounded-lg border py-3"
          onClick={() => advance("pick")}
          disabled={!taskId.trim()}
        >
          Start pick
        </button>
        <button
          type="button"
          className="rounded-lg border py-3"
          onClick={() => advance("pack")}
          disabled={!taskId.trim()}
        >
          Complete pack
        </button>
        <button
          type="button"
          className="rounded-lg bg-violet-600 py-3 text-white disabled:opacity-50"
          onClick={() => advance("ready")}
          disabled={!taskId.trim()}
        >
          Mark ready
        </button>
      </div>
    </div>
  );
}
