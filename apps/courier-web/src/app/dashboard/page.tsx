"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { courierApi, useSession } from "@/shared/api/client";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [onDuty, setOnDuty] = useState(false);
  const [msg, setMsg] = useState("");

  useEffect(() => {
    if (!session) router.replace("/login");
  }, [session, router]);

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
      {msg ? <p className="text-sm" role="status">{msg}</p> : null}
    </div>
  );
}
