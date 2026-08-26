#!/usr/bin/env bash
# Scaffold a role-specific Next.js web app from template.
set -euo pipefail
APP="$1"
PORT="$2"
TITLE="$3"
BFF_ENV="$4"
ROLE="$5"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
DIR="$ROOT/apps/$APP"

mkdir -p "$DIR/src/app/login" "$DIR/src/app/dashboard" "$DIR/src/components/providers" "$DIR/src/styles" "$DIR/src/shared/api"

cat > "$DIR/package.json" <<EOF
{
  "name": "$APP",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev --turbopack -p $PORT",
    "build": "next build",
    "start": "next start -p $PORT",
    "lint": "next lint"
  },
  "dependencies": {
    "@nexora/ui": "file:../../packages/web/ui",
    "@nexora/web-core": "file:../../packages/web/core",
    "@tanstack/react-query": "^5.101.4",
    "next": "^15.5.7",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "zustand": "^5.0.14"
  },
  "devDependencies": {
    "@tailwindcss/postcss": "^4",
    "@types/node": "^20",
    "@types/react": "^19",
    "@types/react-dom": "^19",
    "tailwindcss": "^4",
    "typescript": "^5"
  }
}
EOF

cp "$ROOT/apps/customer-web/next.config.ts" "$DIR/"
cp "$ROOT/apps/customer-web/tsconfig.json" "$DIR/"
cp "$ROOT/apps/customer-web/postcss.config.mjs" "$DIR/"

cat > "$DIR/src/styles/globals.css" <<'EOF'
@import "@nexora/ui/styles.css";
:root { --nx-brand: #5c35ff; }
body { min-height: 100dvh; background: #f4f4f5; }
EOF

cat > "$DIR/src/shared/api/client.ts" <<EOF
"use client";
import { createSessionStore, createApiClient, bffUrl, tenantId, serviceUrl } from "@nexora/web-core";
export const useSession = createSessionStore("nexora-$APP-session");
export function roleApi() {
  const session = useSession.getState().session;
  const base = "$BFF_ENV" === "supplier" ? serviceUrl("supplier") : "$BFF_ENV" === "finance" ? serviceUrl("finance") : "$BFF_ENV" === "settlement" ? serviceUrl("settlement") : bffUrl("$BFF_ENV");
  return createApiClient({
    baseUrl: base,
    tenantId: tenantId(),
    getToken: () => session?.accessToken ?? null,
    getUserId: () => session?.principalId ?? null,
  });
}
EOF

cat > "$DIR/src/components/providers/app-providers.tsx" <<'EOF'
"use client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState, type ReactNode } from "react";
export function AppProviders({ children }: { children: ReactNode }) {
  const [client] = useState(() => new QueryClient());
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
EOF

cat > "$DIR/src/app/layout.tsx" <<EOF
import { AppProviders } from "@/components/providers/app-providers";
import "@/styles/globals.css";
export const metadata = { title: "$TITLE" };
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en"><body><AppProviders>{children}</AppProviders></body></html>
  );
}
EOF

cat > "$DIR/src/app/page.tsx" <<'EOF'
import { redirect } from "next/navigation";
export default function Page() { redirect("/dashboard"); }
EOF

cat > "$DIR/src/app/login/page.tsx" <<EOF
"use client";
import { useRouter } from "next/navigation";
import { useSession } from "@/shared/api/client";
export default function LoginPage() {
  const router = useRouter();
  const setSession = useSession((s) => s.setSession);
  return (
    <div className="mx-auto flex min-h-dvh max-w-md flex-col justify-center p-6">
      <h1 className="mb-6 text-2xl font-bold text-[var(--nx-brand)]">$TITLE</h1>
      <button type="button" className="rounded-lg bg-[var(--nx-brand)] py-3 font-semibold text-white"
        onClick={() => { setSession({ accessToken: "demo", principalId: "demo-principal", roles: ["$ROLE"] }); router.push("/dashboard"); }}>
        Continue as $ROLE (demo)
      </button>
      <p className="mt-4 text-xs text-neutral-500">Wire OTP via identity-service for production staging.</p>
    </div>
  );
}
EOF

cat > "$DIR/src/app/dashboard/page.tsx" <<EOF
"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useSession, roleApi } from "@/shared/api/client";
export default function DashboardPage() {
  const router = useRouter();
  const session = useSession((s) => s.session);
  const logout = useSession((s) => s.logout);
  useEffect(() => { if (!session) router.replace("/login"); }, [session, router]);
  async function ping() {
    try {
      const api = roleApi();
      await api.request("/health");
      alert("BFF/API health OK");
    } catch (e) {
      alert(e instanceof Error ? e.message : "Health check failed");
    }
  }
  return (
    <div className="mx-auto max-w-3xl p-6 space-y-4">
      <div className="flex justify-between items-center">
        <h1 className="text-xl font-semibold">$TITLE</h1>
        <button type="button" className="text-sm" onClick={() => { logout(); router.push("/login"); }}>Logout</button>
      </div>
      <p className="text-sm text-neutral-600">Role: $ROLE · Principal: {session?.principalId}</p>
      <button type="button" className="rounded border px-4 py-2 text-sm" onClick={ping}>Check API health</button>
    </div>
  );
}
EOF

echo "OK scaffold $APP"
