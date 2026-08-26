"use client";

import { useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { customerApi, useSession } from "@/shared/api/client";
import { useLocation } from "@/shared/stores/location-store";

export default function SearchInner() {
  const sp = useSearchParams();
  const q = sp.get("q") ?? "";
  const { lat, lng } = useLocation();
  const session = useSession((s) => s.session);

  const { data, isLoading } = useQuery({
    queryKey: ["search", q, lat, lng],
    queryFn: () =>
      customerApi().request<{ products?: Array<{ sku?: string; name?: string }> }>(
        `/v1/customer/home?lat=${lat}&lng=${lng}&q=${encodeURIComponent(q)}&customerId=${session?.principalId ?? ""}`,
      ),
    enabled: q.length > 0,
  });

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Search: {q || "—"}</h1>
      {isLoading ? <p>Searching…</p> : null}
      <ul className="divide-y rounded-xl border">
        {(data?.products ?? []).map((p) => (
          <li key={p.sku}>
            <Link
              href={`/product/${encodeURIComponent(p.sku ?? "")}`}
              className="block p-4"
            >
              {p.name ?? p.sku}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
