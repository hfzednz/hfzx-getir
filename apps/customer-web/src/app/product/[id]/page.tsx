"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useCart } from "@/shared/stores/cart-store";

export default function ProductPage() {
  const params = useParams<{ id: string }>();
  const sku = decodeURIComponent(params.id);
  const addLine = useCart((s) => s.addLine);

  return (
    <div className="space-y-4">
      <Link href="/home" className="text-sm text-[var(--nx-brand)]">
        ← Back
      </Link>
      <h1 className="text-xl font-semibold">Product</h1>
      <p className="text-sm text-neutral-600">SKU: {sku}</p>
      <button
        type="button"
        className="w-full rounded-lg bg-[var(--nx-brand)] py-3 font-semibold text-white"
        onClick={() =>
          addLine({ sku, name: `Product ${sku.slice(0, 8)}`, qty: 1, unitMinor: 1500 })
        }
      >
        Add to cart — ₺15.00
      </button>
    </div>
  );
}
