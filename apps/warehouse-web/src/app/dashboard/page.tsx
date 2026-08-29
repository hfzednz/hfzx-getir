"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { RouteGuard } from "@nexora/web-core";
import { warehouseApi, useSession } from "@/shared/api/client";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [taskId, setTaskId] = useState(
    () => process.env.NEXT_PUBLIC_WAREHOUSE_TASK_ID?.trim() || "",
  );
  // Empty until a transition actually returns one — never show a status we did not read.
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

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
    setBusy(true);
    try {
      const res = await warehouseApi().request<Record<string, unknown>>(
        `/v1/warehouse/tasks/${id}/${step}`,
        { method: "POST", body: {} },
      );
      setStatus(typeof res.status === "string" ? res.status : step);
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
        <label className="block text-sm">
          Order / task id
          <input
            className="mt-1 w-full rounded-lg border px-3 py-3"
            style={{ minHeight: 44 }}
            value={taskId}
            onChange={(e) => setTaskId(e.target.value)}
            placeholder="uuid"
            aria-invalid={error ? true : undefined}
            aria-describedby={error ? "warehouse-error" : undefined}
          />
        </label>
        <p className="text-sm text-neutral-600" role="status" aria-live="polite">
          {busy ? "Working…" : status ? `Status: ${status}` : "No transition run yet."}
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
