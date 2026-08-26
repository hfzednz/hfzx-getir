"use client";

import { useSession } from "@/shared/api/client";
import { useLocation } from "@/shared/stores/location-store";

export default function AccountPage() {
  const session = useSession((s) => s.session);
  const { addressLabel, lat, lng, setLocation } = useLocation();

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
        <p className="text-sm">{addressLabel}</p>
        <button
          type="button"
          className="rounded-lg border px-3 py-2 text-sm"
          onClick={() => setLocation(41.0082, 28.9784, "Kadıköy, Istanbul")}
        >
          Use Kadıköy
        </button>
        <p className="text-xs text-neutral-500">
          Coords: {lat.toFixed(4)}, {lng.toFixed(4)}
        </p>
      </section>
    </div>
  );
}
