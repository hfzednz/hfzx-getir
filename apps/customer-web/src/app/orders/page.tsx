"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

export default function OrdersPage() {
  const [ids, setIds] = useState<string[]>([]);

  useEffect(() => {
    try {
      const raw = localStorage.getItem("nexora-customer-order-ids");
      setIds(raw ? (JSON.parse(raw) as string[]) : []);
    } catch {
      setIds([]);
    }
  }, []);

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Orders</h1>
      {ids.length === 0 ? (
        <p className="text-sm text-neutral-600">No orders yet.</p>
      ) : (
        <ul className="divide-y rounded-xl border">
          {ids.map((id) => (
            <li key={id}>
              <Link href={`/orders/${id}`} className="block p-4 hover:bg-neutral-50">
                Order {id.slice(0, 8)}…
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
