"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { warehouseApi, useSession } from "@/shared/api/client";

const taskId = () =>
  process.env.NEXT_PUBLIC_WAREHOUSE_TASK_ID?.trim() || "";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [status, setStatus] = useState("queued");
  const [error, setError] = useState("");

  useEffect(() => { if (!session) router.replace("/login"); }, [session, router]);

  async function advance(step: "pick" | "pack" | "ready") {
    const id = taskId();
    if (!id) {
      setError("Set NEXT_PUBLIC_WAREHOUSE_TASK_ID for warehouse task operations.");
      return;
    }
    setError("");
    try {
      await warehouseApi().request(`/v1/warehouse/tasks/${id}/${step}`, { method: "POST", body: {} });
      setStatus(step === "pick" ? "picking" : step === "pack" ? "packing" : "ready_for_dispatch");
    } catch (e) {
      setStatus("error");
      setError(e instanceof Error ? e.message : "Transition failed");
    }
  }

  const id = taskId();

  return (
    <div className="mx-auto max-w-lg space-y-4 p-4">
      <div className="flex justify-between"><h1 className="text-xl font-semibold">Fulfillment</h1>
        <button type="button" className="text-sm" onClick={() => { logout(); router.push("/login"); }}>Logout</button></div>
      {id ? (
        <p className="text-sm text-neutral-600">Task {id} · {status}</p>
      ) : (
        <p className="text-sm text-amber-700" role="alert">No warehouse task configured. Set NEXT_PUBLIC_WAREHOUSE_TASK_ID in staging.</p>
      )}
      {error ? <p className="text-sm text-red-600" role="alert">{error}</p> : null}
      <div className="grid gap-2">
        <button type="button" className="rounded-lg border py-3" onClick={() => advance("pick")} disabled={!id}>Start pick</button>
        <button type="button" className="rounded-lg border py-3" onClick={() => advance("pack")} disabled={!id}>Complete pack</button>
        <button type="button" className="rounded-lg bg-violet-600 py-3 text-white disabled:opacity-50" onClick={() => advance("ready")} disabled={!id}>Mark ready</button>
      </div>
    </div>
  );
}
