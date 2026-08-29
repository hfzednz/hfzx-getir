"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { customerApi } from "@/shared/api/client";
import { useLocation } from "@/shared/stores/location-store";
import { useCart } from "@/shared/stores/cart-store";
import { useSession } from "@/shared/api/client";

export default function HomePage() {
  const { lat, lng, addressLabel } = useLocation();
  const addLine = useCart((s) => s.addLine);
  const session = useSession((s) => s.session);

  const { data, isLoading, error } = useQuery({
    queryKey: ["home", lat, lng, session?.principalId],
    queryFn: async () => {
      const api = customerApi();
      return api.request<{
        serviceable?: boolean;
        Serviceable?: boolean;
        products?: Array<{ id?: string; sku?: string; name?: string; unitMinor?: number }>;
        Products?: Array<{ id?: string; sku?: string; name?: string; unitMinor?: number }>;
        rails?: unknown[];
      }>(`/v1/customer/home?lat=${lat}&lng=${lng}&customerId=${session?.principalId ?? ""}`);
    },
  });

  const products = data?.products ?? data?.Products ?? [];
  const serviceable = data?.serviceable ?? data?.Serviceable;

  return (
    <div className="space-y-6">
      <section>
        <p className="text-xs uppercase tracking-wide text-neutral-500">Deliver to</p>
        <p className="font-medium">{addressLabel}</p>
        {data && serviceable === false ? (
          <p className="text-sm text-amber-700">Area may be outside service zone.</p>
        ) : null}
      </section>

      <form action="/search" className="flex gap-2">
        <input
          name="q"
          placeholder="Search products…"
          className="flex-1 rounded-lg border px-3 py-2 text-sm"
          style={{ minHeight: 44 }}
          aria-label="Search products"
        />
        <button
          type="submit"
          className="rounded-lg bg-neutral-900 px-4 text-sm text-white"
          style={{ minHeight: 44 }}
        >
          Search
        </button>
      </form>

      <section>
        <h2 className="mb-3 text-lg font-semibold">Popular</h2>
        {isLoading ? <p className="text-sm text-neutral-500">Loading…</p> : null}
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error instanceof Error ? error.message : "Failed to load home"}
          </p>
        ) : null}
        {!isLoading && !error && products.length === 0 ? (
          <p className="text-sm text-neutral-500">No products in this area. Try search or another location.</p>
        ) : null}
        <ul className="grid gap-3 sm:grid-cols-2">
          {products.map((p) => {
            const item = p as { id?: string; sku?: string; name?: string; unitMinor?: number };
            const sku = item.sku ?? item.id ?? "unknown";
            const name = item.name ?? "Product";
            const price = item.unitMinor;
            return (
              <li
                key={sku}
                className="rounded-xl border p-4 shadow-sm"
              >
                <Link href={`/product/${encodeURIComponent(sku)}`} className="block">
                  <p className="font-medium">{name}</p>
                  <p className="text-sm text-neutral-600">
                    {price != null ? `₺${(price / 100).toFixed(2)}` : "Price at checkout"}
                  </p>
                </Link>
                <button
                  type="button"
                  className="mt-3 w-full rounded-lg bg-[var(--nx-brand)] py-2 text-sm font-medium text-white"
                  style={{ minHeight: 44 }}
                  onClick={() => addLine({ sku, name, qty: 1, unitMinor: price ?? 0 })}
                >
                  Add to cart
                </button>
              </li>
            );
          })}
        </ul>
      </section>
    </div>
  );
}
