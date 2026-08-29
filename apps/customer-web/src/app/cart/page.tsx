"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCart } from "@/shared/stores/cart-store";

export default function CartPage() {
  const router = useRouter();
  const { lines, updateQty, removeLine, totalMinor } = useCart();

  if (lines.length === 0) {
    return (
      <div className="space-y-4 text-center py-12">
        <p className="text-neutral-600">Your cart is empty.</p>
        <Link href="/home" className="text-[var(--nx-brand)] font-medium">
          Browse products
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Cart</h1>
      <ul className="divide-y rounded-xl border">
        {lines.map((line) => (
          <li key={line.sku} className="flex items-center justify-between gap-3 p-4">
            <div>
              <p className="font-medium">{line.name}</p>
              <p className="text-sm text-neutral-600">₺{(line.unitMinor / 100).toFixed(2)}</p>
            </div>
            <div className="flex items-center gap-2">
              <button
                type="button"
                aria-label="Decrease quantity"
                className="h-11 w-11 rounded border"
                onClick={() => updateQty(line.sku, line.qty - 1)}
              >
                −
              </button>
              <span className="w-6 text-center">{line.qty}</span>
              <button
                type="button"
                aria-label="Increase quantity"
                className="h-11 w-11 rounded border"
                onClick={() => updateQty(line.sku, line.qty + 1)}
              >
                +
              </button>
              <button
                type="button"
                className="text-xs text-red-600"
                onClick={() => removeLine(line.sku)}
              >
                Remove
              </button>
            </div>
          </li>
        ))}
      </ul>
      <div className="flex items-center justify-between font-semibold">
        <span>Total</span>
        <span>₺{(totalMinor() / 100).toFixed(2)}</span>
      </div>
      <button
        type="button"
        className="w-full rounded-lg bg-[var(--nx-brand)] py-3 font-semibold text-white"
        onClick={() => router.push("/checkout")}
      >
        Checkout
      </button>
    </div>
  );
}
