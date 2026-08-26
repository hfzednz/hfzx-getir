"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { supplierApi, useSession } from "@/shared/api/client";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [msg, setMsg] = useState("");

  useEffect(() => { if (!session) router.replace("/login"); }, [session, router]);

  async function loadCatalog() {
    try {
      await supplierApi().request("/v1/supplier/catalog");
      setMsg("Catalog loaded");
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Catalog load failed");
    }
  }

  return (
    <div className="space-y-4 p-4">
      <div className="flex justify-between">
        <h1 className="text-xl font-semibold">Supplier portal</h1>
        <button type="button" className="text-sm" onClick={() => { logout(); router.push("/login"); }}>Logout</button>
      </div>
      <button type="button" className="w-full rounded-xl bg-violet-600 py-4 font-bold" onClick={loadCatalog}>
        Refresh catalog
      </button>
      <section className="rounded-xl bg-slate-800 p-4">
        <h2 className="font-medium">Purchase orders</h2>
        <p className="text-sm text-slate-400">Connect supplier-service for inbound PO queue.</p>
      </section>
      {msg ? <p className="text-sm">{msg}</p> : null}
    </div>
  );
}
