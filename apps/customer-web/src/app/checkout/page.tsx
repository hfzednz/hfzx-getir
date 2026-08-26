"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { customerApi, useSession } from "@/shared/api/client";
import { useCart } from "@/shared/stores/cart-store";
import { useLocation } from "@/shared/stores/location-store";

export default function CheckoutPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const { cartId, lines, setCartId, clear, totalMinor } = useCart();
  const { addressLabel } = useLocation();
  const [preview, setPreview] = useState<Record<string, unknown> | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [paymentMethod] = useState("card");
  const activeCartId = cartId ?? crypto.randomUUID();

  async function syncCart() {
    const api = customerApi();
    for (const line of lines) {
      await api.request("/v1/customer/cart/items", {
        method: "POST",
        body: {
          cartId: activeCartId,
          sku: line.sku,
          qty: line.qty,
          unitMinor: line.unitMinor,
        },
      });
    }
    setCartId(activeCartId);
    return activeCartId;
  }

  async function runPreview() {
    setLoading(true);
    setError("");
    try {
      const id = await syncCart();
      const api = customerApi();
      const res = await api.request<Record<string, unknown>>(
        "/v1/customer/checkout/preview",
        {
          method: "POST",
          body: { cartId: id, principalId: session?.principalId },
        },
      );
      setPreview(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Preview failed");
    } finally {
      setLoading(false);
    }
  }

  async function placeOrder() {
    setLoading(true);
    setError("");
    try {
      const id = cartId ?? activeCartId;
      const api = customerApi();
      const res = await api.request<{ orderId?: string }>(
        "/v1/customer/checkout/place",
        {
          method: "POST",
          idempotencyKey: crypto.randomUUID(),
          body: {
            cartId: id,
            paymentMethod,
            sessionId: crypto.randomUUID(),
            principalId: session?.principalId,
          },
        },
      );
      clear();
      const orderId = res.orderId ?? crypto.randomUUID();
      try {
        const raw = localStorage.getItem("nexora-customer-order-ids");
        const ids = raw ? (JSON.parse(raw) as string[]) : [];
        localStorage.setItem(
          "nexora-customer-order-ids",
          JSON.stringify([orderId, ...ids].slice(0, 20)),
        );
      } catch {
        /* ignore */
      }
      router.push(`/orders/${orderId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Checkout failed");
    } finally {
      setLoading(false);
    }
  }

  if (lines.length === 0) {
    return <p className="text-neutral-600">Cart is empty.</p>;
  }

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Checkout</h1>
      <section className="rounded-xl border p-4">
        <h2 className="font-medium">Delivery address</h2>
        <p className="text-sm text-neutral-600">{addressLabel}</p>
      </section>
      <section className="rounded-xl border p-4">
        <h2 className="font-medium">Payment</h2>
        <p className="text-sm text-neutral-600">Sandbox / MockPSP ({paymentMethod})</p>
      </section>
      <p className="font-semibold">Total ₺{(totalMinor() / 100).toFixed(2)}</p>
      {!preview ? (
        <button
          type="button"
          disabled={loading}
          className="w-full rounded-lg border py-3 font-medium"
          onClick={runPreview}
        >
          {loading ? "Loading…" : "Preview order"}
        </button>
      ) : (
        <pre className="overflow-auto rounded bg-neutral-50 p-3 text-xs">
          {JSON.stringify(preview, null, 2)}
        </pre>
      )}
      <button
        type="button"
        disabled={loading || !preview}
        className="w-full rounded-lg bg-[var(--nx-brand)] py-3 font-semibold text-white disabled:opacity-50"
        onClick={placeOrder}
      >
        {loading ? "Placing…" : "Place order"}
      </button>
      {error ? (
        <p className="text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
