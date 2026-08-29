"use client";

import { useState } from "react";
import { useSession } from "@/shared/api/client";
import { useLocation } from "@/shared/stores/location-store";

export default function AccountPage() {
  const session = useSession((s) => s.session);
  const { addressLabel, lat, lng, setLocation } = useLocation();
  const [label, setLabel] = useState(addressLabel);
  const [status, setStatus] = useState("");
  const [locating, setLocating] = useState(false);

  function useCurrentPosition() {
    if (!navigator.geolocation) {
      setStatus("This browser cannot share your location.");
      return;
    }
    setLocating(true);
    setStatus("");
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLocation(pos.coords.latitude, pos.coords.longitude, label.trim() || addressLabel);
        setStatus("Delivery location updated from your device.");
        setLocating(false);
      },
      () => {
        setStatus("Location permission was denied. You can still name your address below.");
        setLocating(false);
      },
      { timeout: 10000 },
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-semibold">Account</h1>
      <section className="rounded-xl border p-4 space-y-2">
        <p className="text-sm text-neutral-500">Phone</p>
        <p>{session?.phone ?? "—"}</p>
        <p className="text-sm text-neutral-500">Customer ID</p>
        <p className="break-all text-sm">{session?.principalId}</p>
      </section>
      <section className="rounded-xl border p-4 space-y-3">
        <h2 className="font-medium">Delivery location</h2>
        <label className="block text-sm">
          Address name
          <input
            className="mt-1 w-full rounded-lg border px-3 py-3"
            style={{ minHeight: 44 }}
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            placeholder="Home, office, …"
            autoComplete="address-level2"
          />
        </label>
        <button
          type="button"
          className="w-full rounded-lg border px-3 text-sm disabled:opacity-60"
          style={{ minHeight: 44 }}
          onClick={() => setLocation(lat, lng, label.trim() || addressLabel)}
          disabled={!label.trim() || label.trim() === addressLabel}
        >
          Save address name
        </button>
        <button
          type="button"
          className="w-full rounded-lg bg-[var(--nx-brand)] px-3 text-sm font-medium text-white disabled:opacity-60"
          style={{ minHeight: 44 }}
          onClick={useCurrentPosition}
          disabled={locating}
        >
          {locating ? "Locating…" : "Use my current location"}
        </button>
        <p className="text-xs text-neutral-500">
          Delivering to {addressLabel} ({lat.toFixed(4)}, {lng.toFixed(4)})
        </p>
        {status ? (
          <p className="text-xs text-neutral-600" role="status">
            {status}
          </p>
        ) : null}
      </section>
    </div>
  );
}
