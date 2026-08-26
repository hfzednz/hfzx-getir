"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { supplierApi, useSession } from "@/shared/api/client";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  const [suppliers, setSuppliers] = useState<unknown[]>([]);
  const [sellers, setSellers] = useState<unknown[]>([]);
  const [pos, setPos] = useState<unknown[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!session) router.replace("/login");
  }, [session, router]);

  useEffect(() => {
    if (!session) return;
    (async () => {
      try {
        const s = await supplierApi().request<unknown[]>("/v1/supplier/suppliers");
        const p = await supplierApi().request<unknown[]>("/v1/supplier/purchase-orders");
        let sellers: unknown[] = [];
        try {
          sellers = await supplierApi().request<unknown[]>("/v1/supplier/sellers");
        } catch {
          sellers = [];
        }
        setSuppliers(Array.isArray(s) ? s : []);
        setPos(Array.isArray(p) ? p : []);
        setSellers(Array.isArray(sellers) ? sellers : []);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Load failed");
      }
    })();
  }, [session]);

  return (
    <div className="space-y-4 p-4">
      <div className="flex justify-between">
        <h1 className="text-xl font-semibold">Supplier portal</h1>
        <button type="button" className="text-sm" onClick={() => { logout(); router.push("/login"); }}>Logout</button>
      </div>
      {error ? <p className="text-sm text-red-600" role="alert">{error}</p> : null}
      <section className="rounded-xl border p-4">
        <h2 className="font-medium">Suppliers ({suppliers.length})</h2>
      </section>
      <section className="rounded-xl border p-4">
        <h2 className="font-medium">Marketplace sellers ({sellers.length})</h2>
      </section>
      <section className="rounded-xl border p-4">
        <h2 className="font-medium">Purchase orders ({pos.length})</h2>
      </section>
    </div>
  );
}
