"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { RouteGuard } from "@nexora/web-core";
import { courierApi, useSession } from "@/shared/api/client";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [onDuty, setOnDuty] = useState(false);
  const [orderId, setOrderId] = useState("");
  const [msg, setMsg] = useState("");
  const [busy, setBusy] = useState(false);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    const q = new URLSearchParams(window.location.search).get("orderId")?.trim();
    if (q) setOrderId(q);
  }, []);

  async function run(label: string, action: () => Promise<Record<string, unknown> | void>) {
    setBusy(true);
    setFailed(false);
    setMsg("");
    try {
      const res = (await action()) ?? {};
      const status = (res as Record<string, unknown>).status;
      setMsg(typeof status === "string" ? status : label);
    } catch (e) {
      setFailed(true);
      setMsg(e instanceof Error ? e.message : `${label} failed`);
    } finally {
      setBusy(false);
    }
  }

  async function toggleDuty() {
    const next = !onDuty;
    await run(next ? "Online" : "Offline", async () => {
      await courierApi().request("/v1/courier/duty", {
        method: "POST",
        body: { courierId: session?.principalId ?? "", on: next },
      });
      setOnDuty(next);
    });
  }

  async function offer(accept: boolean) {
    const id = orderId.trim();
    if (!id) {
      setFailed(true);
      setMsg("Enter the real order id");
      return;
    }
    await run(accept ? "accepted" : "rejected", () =>
      courierApi().request<Record<string, unknown>>(`/v1/courier/offers/${id}`, {
        method: "POST",
        body: { courierId: session?.principalId ?? "", accept },
      }),
    );
  }

  async function enroute() {
    const id = orderId.trim();
    if (!id) return;
    await run("out_for_delivery", () =>
      courierApi().request<Record<string, unknown>>(`/v1/courier/offers/${id}/enroute`, {
        method: "POST",
        body: {},
      }),
    );
  }

  async function complete() {
    const id = orderId.trim();
    if (!id) return;
    await run("delivered", () =>
      courierApi().request<Record<string, unknown>>(`/v1/courier/offers/${id}/complete`, {
        method: "POST",
        body: {},
      }),
    );
  }

  return (
    <RouteGuard session={session} allow={["courier"]} onDeny={logout}>
      <div className="space-y-4 p-4">
        <div className="flex items-center justify-between gap-3">
          <h1 className="text-xl font-semibold">Deliveries</h1>
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
        <button
          type="button"
          className={`w-full rounded-xl py-4 font-bold disabled:opacity-60 ${onDuty ? "bg-emerald-600" : "bg-slate-700"}`}
          onClick={toggleDuty}
          disabled={busy}
        >
          {onDuty ? "Online — tap to go offline" : "Go online"}
        </button>
        <label className="block text-sm">
          Order id
          <input
            className="mt-1 w-full rounded-lg border px-3 py-3"
            style={{ minHeight: 44 }}
            value={orderId}
            onChange={(e) => setOrderId(e.target.value)}
            placeholder="uuid"
            aria-invalid={failed ? true : undefined}
            aria-describedby={msg ? "courier-status" : undefined}
          />
        </label>
        <div className="grid gap-2">
          <button
            type="button"
            className="rounded-lg border py-3 disabled:opacity-50"
            style={{ minHeight: 44 }}
            onClick={() => offer(true)}
            disabled={busy || !orderId.trim()}
          >
            Accept assignment
          </button>
          <button
            type="button"
            className="rounded-lg border py-3 disabled:opacity-50"
            style={{ minHeight: 44 }}
            onClick={enroute}
            disabled={busy || !orderId.trim()}
          >
            Out for delivery
          </button>
          <button
            type="button"
            className="rounded-lg bg-violet-600 py-3 text-white disabled:opacity-50"
            style={{ minHeight: 44 }}
            onClick={complete}
            disabled={busy || !orderId.trim()}
          >
            Mark delivered
          </button>
        </div>
        {busy ? <p className="text-sm text-neutral-500">Working…</p> : null}
        {msg ? (
          <p
            id="courier-status"
            className={`text-sm ${failed ? "text-red-600" : ""}`}
            role={failed ? "alert" : "status"}
          >
            {msg}
          </p>
        ) : null}
      </div>
    </RouteGuard>
  );
}
