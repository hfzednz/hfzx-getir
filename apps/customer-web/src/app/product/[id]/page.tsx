"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { customerApi } from "@/shared/api/client";
import { useLocation } from "@/shared/stores/location-store";
import { useCart } from "@/shared/stores/cart-store";

type CatalogProduct = {
  id?: string;
  sku?: string;
  name?: string;
  unitMinor?: number;
};

export default function ProductPage() {
  const params = useParams<{ id: string }>();
  const sku = decodeURIComponent(params.id);
  const { lat, lng } = useLocation();
  const addLine = useCart((s) => s.addLine);

  const { data, isLoading, error } = useQuery({
    queryKey: ["product", sku, lat, lng],
    queryFn: async () => {
      const feed = await customerApi().request<{ products?: CatalogProduct[] }>(
        `/v1/customer/home?lat=${lat}&lng=${lng}&q=${encodeURIComponent(sku)}`,
      );
      const products = feed.products ?? [];
      return (
        products.find((p) => p.sku === sku || p.id === sku) ?? products[0] ?? null
      );
    },
  });

  const name = data?.name;
  const unitMinor = data?.unitMinor;

  return (
    <div className="space-y-4">
      <Link href="/home" className="text-sm text-[var(--nx-brand)]">
        ← Back
      </Link>
      <h1 className="text-xl font-semibold">{name ?? "Product"}</h1>
      <p className="text-sm text-neutral-600">SKU: {sku}</p>
      {isLoading ? <p className="text-sm text-neutral-500">Loading…</p> : null}
      {error ? (
        <p className="text-sm text-red-600" role="alert">
          {error instanceof Error ? error.message : "Could not load this product"}
        </p>
      ) : null}
      {!isLoading && !error && !data ? (
        <p className="text-sm text-neutral-500">
          This product is not available in your area.
        </p>
      ) : null}
      {data ? (
        <>
          <p className="text-sm text-neutral-600">
            {unitMinor != null
              ? `₺${(unitMinor / 100).toFixed(2)}`
              : "Price is calculated at checkout."}
          </p>
          <button
            type="button"
            className="w-full rounded-lg bg-[var(--nx-brand)] py-3 font-semibold text-white"
            style={{ minHeight: 44 }}
            onClick={() =>
              addLine({ sku, name: name ?? sku, qty: 1, unitMinor: unitMinor ?? 0 })
            }
          >
            Add to cart
          </button>
        </>
      ) : null}
    </div>
  );
}
