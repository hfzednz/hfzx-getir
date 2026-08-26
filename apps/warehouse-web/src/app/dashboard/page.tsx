"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { warehouseApi, useSession } from "@/shared/api/client";

const DEMO_TASK = "task-demo-001";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [status, setStatus] = useState("queued");

  useEffect(() => { if (!session) router.replace("/login"); }, [session, router]);

  async function advance(step: "pick" | "pack" | "ready") {
    try {
      await warehouseApi().request(`/v1/warehouse/tasks/${DEMO_TASK}/${step}`, { method: "POST", body: {} });
      setStatus(step === "pick" ? "picking" : step === "pack" ? "packing" : "ready_for_dispatch");
    } catch (e) {
      setStatus(e instanceof Error ? e.message : "failed");
    }
  }

  return (
    <div className="mx-auto max-w-lg space-y-4 p-4">
      <div className="flex justify-between"><h1 className="text-xl font-semibold">Fulfillment</h1>
        <button type="button" className="text-sm" onClick={() => { logout(); router.push("/login"); }}>Logout</button></div>
      <p className="text-sm text-neutral-600">Task {DEMO_TASK} · {status}</p>
      <div className="grid gap-2">
        <button type="button" className="rounded-lg border py-3" onClick={() => advance("pick")}>Start pick</button>
        <button type="button" className="rounded-lg border py-3" onClick={() => advance("pack")}>Complete pack</button>
        <button type="button" className="rounded-lg bg-violet-600 py-3 text-white" onClick={() => advance("ready")}>Mark ready</button>
      </div>
    </div>
  );
}
