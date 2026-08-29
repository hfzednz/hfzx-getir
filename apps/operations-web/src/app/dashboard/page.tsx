"use client";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { RouteGuard } from "@nexora/web-core";
import { operationsApi, useSession } from "@/shared/api/client";

type Dashboard = {
  ordersLive?: number;
  couriersOnDuty?: number;
  openIncidents?: number;
  sloBurn?: number;
};

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);

  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ["ops-dashboard"],
    enabled: Boolean(session),
    queryFn: () => operationsApi().request<Dashboard>("/v1/admin/dashboard"),
  });

  const cards: Array<{ label: string; value: string }> = [
    { label: "Live orders", value: data?.ordersLive?.toString() ?? "—" },
    { label: "Couriers on duty", value: data?.couriersOnDuty?.toString() ?? "—" },
    { label: "Open incidents", value: data?.openIncidents?.toString() ?? "—" },
    {
      label: "SLO burn",
      value: typeof data?.sloBurn === "number" ? data.sloBurn.toFixed(2) : "—",
    },
  ];

  return (
    <RouteGuard session={session} allow={["city_ops", "admin", "super_admin"]} onDeny={logout}>
      <div className="space-y-4 p-4">
        <div className="flex items-center justify-between gap-3">
          <h1 className="text-xl font-semibold">City operations</h1>
          <button
            type="button"
            className="rounded-lg px-3 text-sm"
            style={{ minHeight: 44 }}
            onClick={() => {
              logout();
              router.push("/login");
            }}
          >
            Logout
          </button>
        </div>

        {isLoading ? <p className="text-sm text-neutral-500">Loading…</p> : null}
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error instanceof Error ? error.message : "Load failed"}
          </p>
        ) : null}

        {data ? (
          <ul className="grid grid-cols-2 gap-3">
            {cards.map((card) => (
              <li key={card.label} className="rounded-xl border p-4">
                <p className="text-xs uppercase tracking-wide text-neutral-500">{card.label}</p>
                <p className="text-2xl font-semibold">{card.value}</p>
              </li>
            ))}
          </ul>
        ) : null}

        <button
          type="button"
          className="w-full rounded-lg border py-3 text-sm disabled:opacity-60"
          style={{ minHeight: 44 }}
          onClick={() => refetch()}
          disabled={isFetching}
        >
          {isFetching ? "Refreshing…" : "Refresh"}
        </button>
      </div>
    </RouteGuard>
  );
}
