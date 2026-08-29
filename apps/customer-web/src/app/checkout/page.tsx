"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { customerApi, useSession } from "@/shared/api/client";
import { useCart } from "@/shared/stores/cart-store";
import { useLocation } from "@/shared/stores/location-store";

export default function CheckoutPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const { cartId, lines, setCartId, clear, totalMinor } = useCart();
  const { lat, lng, addressLabel } = useLocation();
  const [line1, setLine1] = useState(addressLabel);
  const [preview, setPreview] = useState<Record<string, unknown> | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [paymentMethod] = useState("card");
  // One cart id and one idempotency key for the whole checkout attempt: regenerating
  // either on re-render or on retry would let a retried submit create a second order.
  const [draftCartId] = useState(() => crypto.randomUUID());
  const placeKeyRef = useRef<string>("");
  const syncedCartRef = useRef<string>("");
  const activeCartId = cartId ?? draftCartId;
  const deliveryLine = line1.trim() || addressLabel.trim();
  const addressReady = Boolean(deliveryLine) || (lat !== 0 && lng !== 0);

  async function syncCart() {
    const signature = `${activeCartId}|${lines
      .map((l) => `${l.sku}:${l.qty}:${l.unitMinor}`)
      .join(",")}`;
    if (syncedCartRef.current === signature) {
      return activeCartId;
    }
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
    syncedCartRef.current = signature;
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
    if (!addressReady) {
      setError("Delivery address required");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const id = cartId ?? activeCartId;
      const api = customerApi();
      const sessionId = String(
        preview?.sessionId ?? preview?.SessionID ?? preview?.id ?? "",
      );
      if (!placeKeyRef.current) {
        placeKeyRef.current = crypto.randomUUID();
      }
      const res = await api.request<{ orderId?: string }>(
        "/v1/customer/checkout/place",
        {
          method: "POST",
          idempotencyKey: placeKeyRef.current,
          body: {
            cartId: id,
            paymentMethod,
            ...(sessionId ? { sessionId } : {}),
            principalId: session?.principalId,
            address: {
              label: addressLabel,
              line1: deliveryLine || addressLabel,
              city: addressLabel,
              country: "TR",
              lat,
              lng,
            },
          },
        },
      );
      const orderId = res.orderId;
      if (!orderId) {
        setError("Place order did not return an order id");
        return;
      }
      clear();
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

  const money = (minor: unknown, currency: string) =>
    typeof minor === "number" ? `${currency} ${(minor / 100).toFixed(2)}` : "—";
  const currency = String(preview?.currency ?? "TRY");
  const previewRows = preview
    ? [
        { label: "Subtotal", value: money(preview.subtotalMinor, currency) },
        { label: "Discount", value: money(preview.discountMinor, currency) },
        { label: "Total", value: money(preview.totalMinor, currency) },
        {
          label: "Payment",
          value: preview.paymentReady ? "Ready" : "Not ready",
        },
      ]
    : [];

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Checkout</h1>
      <section className="rounded-xl border p-4 space-y-2">
        <h2 className="font-medium">Delivery address</h2>
        <p className="text-sm text-neutral-600">{addressLabel}</p>
        <label className="block text-sm">
          Street / building
          <input
            value={line1}
            onChange={(e) => setLine1(e.target.value)}
            className="mt-1 w-full rounded-lg border px-3 py-3"
            style={{ minHeight: 44 }}
            placeholder="Street and building"
            autoComplete="street-address"
            aria-invalid={error && !addressReady ? true : undefined}
            aria-describedby={error ? "checkout-error" : undefined}
          />
        </label>
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
        <section className="rounded-xl border p-4 text-sm" aria-label="Order summary">
          <h2 className="mb-2 font-medium">Order summary</h2>
          <dl className="space-y-1">
            {previewRows.map((row) => (
              <div key={row.label} className="flex justify-between gap-3">
                <dt className="text-neutral-600">{row.label}</dt>
                <dd className="font-medium">{row.value}</dd>
              </div>
            ))}
          </dl>
        </section>
      )}
      <button
        type="button"
        disabled={loading || !preview || !addressReady}
        className="w-full rounded-lg bg-[var(--nx-brand)] py-3 font-semibold text-white disabled:opacity-50"
        onClick={placeOrder}
      >
        {loading ? "Placing…" : "Place order"}
      </button>
      {error ? (
        <p id="checkout-error" className="text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
