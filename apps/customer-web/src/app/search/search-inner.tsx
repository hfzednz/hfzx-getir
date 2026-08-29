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

  const { data, isLoading, error } = useQuery({
    queryKey: ["search", q, lat, lng],
    queryFn: () =>
      customerApi().request<{ products?: Array<{ sku?: string; id?: string; name?: string }> }>(
        `/v1/customer/home?lat=${lat}&lng=${lng}&q=${encodeURIComponent(q)}&customerId=${session?.principalId ?? ""}`,
      ),
    enabled: q.length > 0,
  });

  const products = data?.products ?? [];

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Search: {q || "—"}</h1>
      {q.length === 0 ? (
        <p className="text-sm text-neutral-500">
          Type a product name in the search box on the home page.
        </p>
      ) : null}
      {isLoading ? <p>Searching…</p> : null}
      {error ? (
        <p className="text-sm text-red-600" role="alert">
          {error instanceof Error ? error.message : "Search failed"}
        </p>
      ) : null}
      {q.length > 0 && !isLoading && !error && products.length === 0 ? (
        <p className="text-sm text-neutral-500">No products matched “{q}”.</p>
      ) : null}
      {products.length > 0 ? (
        <ul className="divide-y rounded-xl border">
          {products.map((p) => {
            const sku = p.sku ?? p.id ?? "";
            return (
              <li key={sku}>
                <Link
                  href={`/product/${encodeURIComponent(sku)}`}
                  className="block p-4"
                  style={{ minHeight: 44 }}
                >
                  {p.name ?? sku}
                </Link>
              </li>
            );
          })}
        </ul>
      ) : null}
    </div>
  );
}
