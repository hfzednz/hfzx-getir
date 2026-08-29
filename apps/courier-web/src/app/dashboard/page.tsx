"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { courierApi, useSession } from "@/shared/api/client";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [onDuty, setOnDuty] = useState(false);
  const [orderId, setOrderId] = useState("");
  const [msg, setMsg] = useState("");

  useEffect(() => {
    if (!session) router.replace("/login");
  }, [session, router]);

  useEffect(() => {
    const q = new URLSearchParams(window.location.search).get("orderId")?.trim();
    if (q) setOrderId(q);
  }, []);

  async function toggleDuty() {
    setMsg("");
    try {
      await courierApi().request("/v1/courier/duty", {
        method: "POST",
        body: {
          courierId: session?.principalId ?? "",
          on: !onDuty,
        },
      });
      setOnDuty(!onDuty);
      setMsg(!onDuty ? "Online" : "Offline");
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Duty toggle failed");
    }
  }

  async function offer(accept: boolean) {
    const id = orderId.trim();
    if (!id) {
      setMsg("Enter the real order id");
      return;
    }
    setMsg("");
    try {
      const res = await courierApi().request<Record<string, unknown>>(
        `/v1/courier/offers/${id}`,
        {
          method: "POST",
          body: { courierId: session?.principalId ?? "", accept },
        },
      );
      setMsg(typeof res.status === "string" ? res.status : accept ? "accepted" : "rejected");
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Offer failed");
    }
  }

  async function enroute() {
    const id = orderId.trim();
    if (!id) return;
    try {
      const res = await courierApi().request<Record<string, unknown>>(
        `/v1/courier/offers/${id}/enroute`,
        { method: "POST", body: {} },
      );
      setMsg(typeof res.status === "string" ? res.status : "out_for_delivery");
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Enroute failed");
    }
  }

  async function complete() {
    const id = orderId.trim();
    if (!id) return;
    try {
      const res = await courierApi().request<Record<string, unknown>>(
        `/v1/courier/offers/${id}/complete`,
        { method: "POST", body: {} },
      );
      setMsg(typeof res.status === "string" ? res.status : "delivered");
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Complete failed");
    }
  }

  return (
    <div className="space-y-4 p-4">
      <div className="flex justify-between">
        <h1 className="text-xl font-semibold">Deliveries</h1>
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
      <button
        type="button"
        className={`w-full rounded-xl py-4 font-bold ${onDuty ? "bg-emerald-600" : "bg-slate-700"}`}
        onClick={toggleDuty}
      >
        {onDuty ? "Online — tap to go offline" : "Go online"}
      </button>
      <label className="block text-sm">
        Order id
        <input
          className="mt-1 w-full rounded-lg border px-3 py-3"
          value={orderId}
          onChange={(e) => setOrderId(e.target.value)}
          placeholder="uuid"
        />
      </label>
      <div className="grid gap-2">
        <button type="button" className="rounded-lg border py-3" onClick={() => offer(true)} disabled={!orderId.trim()}>
          Accept assignment
        </button>
        <button type="button" className="rounded-lg border py-3" onClick={enroute} disabled={!orderId.trim()}>
          Out for delivery
        </button>
        <button
          type="button"
          className="rounded-lg bg-violet-600 py-3 text-white disabled:opacity-50"
          onClick={complete}
          disabled={!orderId.trim()}
        >
          Mark delivered
        </button>
      </div>
      {msg ? <p className="text-sm" role="status">{msg}</p> : null}
    </div>
  );
}
