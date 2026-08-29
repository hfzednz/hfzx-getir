"use client";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { RouteGuard } from "@nexora/web-core";
import { financeApi, useSession } from "@/shared/api/client";

type Journal = {
  id?: string;
  reference?: string;
  description?: string;
  status?: string;
  currency?: string;
  debitTotal?: number;
  creditTotal?: number;
  postedAt?: string;
};

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);

  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ["journals"],
    enabled: Boolean(session),
    queryFn: () =>
      financeApi().request<{ items?: Journal[]; total?: number }>(
        "/v1/ledger/journals?limit=25",
      ),
  });

  const journals = data?.items ?? [];

  return (
    <RouteGuard
      session={session}
      allow={["finance_analyst", "admin", "super_admin"]}
      onDeny={logout}
    >
      <div className="space-y-4 p-4">
        <div className="flex items-center justify-between gap-3">
          <h1 className="text-xl font-semibold">Finance</h1>
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

        <button
          type="button"
          className="w-full rounded-xl bg-violet-600 py-4 font-bold disabled:opacity-60"
          onClick={() => refetch()}
          disabled={isFetching}
        >
          {isFetching ? "Loading journals…" : "Refresh journals"}
        </button>

        <section className="rounded-xl bg-slate-800 p-4">
          <h2 className="font-medium">
            Ledger journals{data?.total != null ? ` (${data.total})` : ""}
          </h2>
          {isLoading ? (
            <p className="text-sm text-slate-400">Loading…</p>
          ) : null}
          {error ? (
            <p className="text-sm text-red-400" role="alert">
              {error instanceof Error ? error.message : "Journals load failed"}
            </p>
          ) : null}
          {!isLoading && !error && journals.length === 0 ? (
            <p className="text-sm text-slate-400">
              No journal entries have been posted for this tenant yet.
            </p>
          ) : null}
          {journals.length > 0 ? (
            <ul className="mt-2 divide-y divide-slate-700 text-sm">
              {journals.map((j, index) => (
                <li key={j.id ?? index} className="flex items-center justify-between gap-3 py-2">
                  <span className="min-w-0">
                    <span className="block truncate font-medium">
                      {j.reference || j.description || j.id || "Journal"}
                    </span>
                    <span className="block text-xs text-slate-400">
                      {j.status ?? "unknown"}
                      {j.postedAt ? ` · ${new Date(j.postedAt).toLocaleString()}` : ""}
                    </span>
                  </span>
                  {typeof j.debitTotal === "number" ? (
                    <span className="shrink-0 font-medium">
                      {j.currency ?? "TRY"} {(j.debitTotal / 100).toFixed(2)}
                    </span>
                  ) : null}
                </li>
              ))}
            </ul>
          ) : null}
        </section>
      </div>
    </RouteGuard>
  );
}
